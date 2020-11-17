package statedb

import (
	"errors"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

var (
	// keyidx is used as key in the db to store the current Idx
	keyidx = []byte("k:idx")
)

func (s *StateDB) resetZKInputs() {
	s.zki = nil
	s.i = 0 // initialize current transaction index in the ZKInputs generation
}

type processedExit struct {
	exit    bool
	newExit bool
	idx     common.Idx
	acc     common.Account
}

// ProcessTxOutput contains the output of the ProcessTxs method
type ProcessTxOutput struct {
	ZKInputs           *common.ZKInputs
	ExitInfos          []common.ExitInfo
	CreatedAccounts    []common.Account
	CoordinatorIdxsMap map[common.TokenID]common.Idx
	CollectedFees      map[common.TokenID]*big.Int
}

// ProcessTxsConfig contains the config for ProcessTxs
type ProcessTxsConfig struct {
	NLevels  uint32
	MaxFeeTx uint32
	MaxTx    uint32
	MaxL1Tx  uint32
}

// ProcessTxs process the given L1Txs & L2Txs applying the needed updates to
// the StateDB depending on the transaction Type.  If StateDB
// type==TypeBatchBuilder, returns the common.ZKInputs to generate the
// SnarkProof later used by the BatchBuilder.  If StateDB
// type==TypeSynchronizer, assumes that the call is done from the Synchronizer,
// returns common.ExitTreeLeaf that is later used by the Synchronizer to update
// the HistoryDB, and adds Nonce & TokenID to the L2Txs.
// And if TypeSynchronizer returns an array of common.Account with all the
// created accounts.
func (s *StateDB) ProcessTxs(ptc ProcessTxsConfig, coordIdxs []common.Idx, l1usertxs, l1coordinatortxs []common.L1Tx, l2txs []common.PoolL2Tx) (ptOut *ProcessTxOutput, err error) {
	defer func() {
		if err == nil {
			err = s.MakeCheckpoint()
		}
	}()

	var exitTree *merkletree.MerkleTree
	var createdAccounts []common.Account

	if s.zki != nil {
		return nil, errors.New("Expected StateDB.zki==nil, something went wrong and it's not empty")
	}
	defer s.resetZKInputs()

	s.accumulatedFees = make(map[common.Idx]*big.Int)

	nTx := len(l1usertxs) + len(l1coordinatortxs) + len(l2txs)
	if nTx == 0 {
		// TODO return ZKInputs of batch without txs
		return &ProcessTxOutput{
			ZKInputs:           nil,
			ExitInfos:          nil,
			CreatedAccounts:    nil,
			CoordinatorIdxsMap: nil,
			CollectedFees:      nil,
		}, nil
	}
	exits := make([]processedExit, nTx)

	if s.typ == TypeBatchBuilder {
		s.zki = common.NewZKInputs(uint32(nTx), ptc.MaxL1Tx, ptc.MaxTx, ptc.MaxFeeTx, ptc.NLevels)
		s.zki.OldLastIdx = s.idx.BigInt()
		s.zki.OldStateRoot = s.mt.Root().BigInt()
	}

	// TBD if ExitTree is only in memory or stored in disk, for the moment
	// only needed in memory
	if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
		tmpDir, err := ioutil.TempDir("", "hermez-statedb-exittree")
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Errorw("Deleting statedb temp exit tree", "err", err)
			}
		}()
		sto, err := pebble.NewPebbleStorage(tmpDir, false)
		if err != nil {
			return nil, err
		}
		exitTree, err = merkletree.NewMerkleTree(sto, s.mt.MaxLevels())
		if err != nil {
			return nil, err
		}
	}

	// Process L1UserTxs
	for i := 0; i < len(l1usertxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		exitIdx, exitAccount, newExit, createdAccount, err := s.processL1Tx(exitTree, &l1usertxs[i])
		if err != nil {
			return nil, err
		}
		if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
			if exitIdx != nil && exitTree != nil {
				exits[s.i] = processedExit{
					exit:    true,
					newExit: newExit,
					idx:     *exitIdx,
					acc:     *exitAccount,
				}
			}
			s.i++
		}
		if s.typ == TypeSynchronizer && createdAccount != nil {
			createdAccounts = append(createdAccounts, *createdAccount)
		}

		if s.zki != nil {
			l1TxData, err := l1usertxs[i].BytesGeneric()
			if err != nil {
				return nil, err
			}
			s.zki.Metadata.L1TxsData = append(s.zki.Metadata.L1TxsData, l1TxData)
		}
	}

	// Process L1CoordinatorTxs
	for i := 0; i < len(l1coordinatortxs); i++ {
		exitIdx, _, _, createdAccount, err := s.processL1Tx(exitTree, &l1coordinatortxs[i])
		if err != nil {
			return nil, err
		}
		if exitIdx != nil {
			log.Error("Unexpected Exit in L1CoordinatorTx")
		}
		if s.typ == TypeSynchronizer && createdAccount != nil {
			createdAccounts = append(createdAccounts, *createdAccount)
		}
		if s.zki != nil {
			l1TxData, err := l1coordinatortxs[i].BytesGeneric()
			if err != nil {
				return nil, err
			}
			s.zki.Metadata.L1TxsData = append(s.zki.Metadata.L1TxsData, l1TxData)
		}
	}

	s.accumulatedFees = make(map[common.Idx]*big.Int)
	for _, idx := range coordIdxs {
		s.accumulatedFees[idx] = big.NewInt(0)
	}

	// once L1UserTxs & L1CoordinatorTxs are processed, get TokenIDs of
	// coordIdxs. In this way, if a coordIdx uses an Idx that is being
	// created in the current batch, at this point the Idx will be created
	coordIdxsMap, err := s.getTokenIDsFromIdxs(coordIdxs)
	if err != nil {
		return nil, err
	}
	var collectedFees map[common.TokenID]*big.Int
	if s.typ == TypeSynchronizer {
		collectedFees = make(map[common.TokenID]*big.Int)
		for tokenID := range coordIdxsMap {
			collectedFees[tokenID] = big.NewInt(0)
		}
	}

	// Process L2Txs
	for i := 0; i < len(l2txs); i++ {
		exitIdx, exitAccount, newExit, err := s.processL2Tx(coordIdxsMap, collectedFees, exitTree, &l2txs[i])
		if err != nil {
			return nil, err
		}
		if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
			if exitIdx != nil && exitTree != nil {
				exits[s.i] = processedExit{
					exit:    true,
					newExit: newExit,
					idx:     *exitIdx,
					acc:     *exitAccount,
				}
			}
			s.i++
		}
		if s.zki != nil {
			l2TxData, err := l2txs[i].L2Tx().Bytes(s.zki.Metadata.NLevels)
			if err != nil {
				return nil, err
			}
			s.zki.Metadata.L2TxsData = append(s.zki.Metadata.L2TxsData, l2TxData)
		}
	}

	// distribute the AccumulatedFees from the processed L2Txs into the
	// Coordinator Idxs
	iFee := 0
	for idx, accumulatedFee := range s.accumulatedFees {
		// send the fee to the Idx of the Coordinator for the TokenID
		accCoord, err := s.GetAccount(idx)
		if err != nil {
			log.Errorw("Can not distribute accumulated fees to coordinator account: No coord Idx to receive fee", "idx", idx)
			return nil, err
		}
		accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
		pFee, err := s.UpdateAccount(idx, accCoord)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		if s.zki != nil {
			s.zki.TokenID3[iFee] = accCoord.TokenID.BigInt()
			s.zki.Nonce3[iFee] = accCoord.Nonce.BigInt()
			if babyjub.PointCoordSign(accCoord.PublicKey.X) {
				s.zki.Sign3[iFee] = big.NewInt(1)
			}
			s.zki.Ay3[iFee] = accCoord.PublicKey.Y
			s.zki.Balance3[iFee] = accCoord.Balance
			s.zki.EthAddr3[iFee] = common.EthAddrToBigInt(accCoord.EthAddr)
			s.zki.Siblings3[iFee] = siblingsToZKInputFormat(pFee.Siblings)

			// add Coord Idx to ZKInputs.FeeTxsData
			s.zki.FeeIdxs[iFee] = idx.BigInt()
		}
		iFee++
	}

	if s.typ == TypeTxSelector {
		return nil, nil
	}

	// once all txs processed (exitTree root frozen), for each Exit,
	// generate common.ExitInfo data
	var exitInfos []common.ExitInfo
	for i := 0; i < nTx; i++ {
		if !exits[i].exit {
			continue
		}
		exitIdx := exits[i].idx
		exitAccount := exits[i].acc

		// 0. generate MerkleProof
		p, err := exitTree.GenerateCircomVerifierProof(exitIdx.BigInt(), nil)
		if err != nil {
			return nil, err
		}
		// 1. generate common.ExitInfo
		ei := common.ExitInfo{
			AccountIdx:  exitIdx,
			MerkleProof: p,
			Balance:     exitAccount.Balance,
		}
		exitInfos = append(exitInfos, ei)

		if s.zki != nil {
			s.zki.TokenID2[i] = exitAccount.TokenID.BigInt()
			s.zki.Nonce2[i] = exitAccount.Nonce.BigInt()
			if babyjub.PointCoordSign(exitAccount.PublicKey.X) {
				s.zki.Sign2[i] = big.NewInt(1)
			}
			s.zki.Ay2[i] = exitAccount.PublicKey.Y
			s.zki.Balance2[i] = exitAccount.Balance
			s.zki.EthAddr2[i] = common.EthAddrToBigInt(exitAccount.EthAddr)
			for j := 0; j < len(p.Siblings); j++ {
				s.zki.Siblings2[i][j] = p.Siblings[j].BigInt()
			}
			if exits[i].newExit {
				s.zki.NewExit[i] = big.NewInt(1)
			}
			if p.IsOld0 {
				s.zki.IsOld0_2[i] = big.NewInt(1)
			}
			s.zki.OldKey2[i] = p.OldKey.BigInt()
			s.zki.OldValue2[i] = p.OldValue.BigInt()
		}
	}
	if s.typ == TypeSynchronizer {
		// return exitInfos, createdAccounts and collectedFees, so Synchronizer will
		// be able to store it into HistoryDB for the concrete BatchNum
		return &ProcessTxOutput{
			ZKInputs:           nil,
			ExitInfos:          exitInfos,
			CreatedAccounts:    createdAccounts,
			CoordinatorIdxsMap: coordIdxsMap,
			CollectedFees:      collectedFees,
		}, nil
	}

	// compute last ZKInputs parameters
	s.zki.GlobalChainID = big.NewInt(0) // TODO, 0: ethereum, this will be get from config file
	// zki.FeeIdxs = ? // TODO, this will be get from the config file
	tokenIDs, err := s.getTokenIDsBigInt(l1usertxs, l1coordinatortxs, l2txs)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	s.zki.FeePlanTokens = tokenIDs
	s.zki.Metadata.NewStateRootRaw = s.mt.Root()
	s.zki.Metadata.NewExitRootRaw = exitTree.Root()

	// s.zki.ISInitStateRootFee = s.mt.Root().BigInt()

	// return ZKInputs as the BatchBuilder will return it to forge the Batch
	return &ProcessTxOutput{
		ZKInputs:           s.zki,
		ExitInfos:          nil,
		CreatedAccounts:    nil,
		CoordinatorIdxsMap: coordIdxsMap,
		CollectedFees:      nil,
	}, nil
}

// getTokenIDsBigInt returns the list of TokenIDs in *big.Int format
func (s *StateDB) getTokenIDsBigInt(l1usertxs, l1coordinatortxs []common.L1Tx, l2txs []common.PoolL2Tx) ([]*big.Int, error) {
	tokenIDs := make(map[common.TokenID]bool)
	for i := 0; i < len(l1usertxs); i++ {
		tokenIDs[l1usertxs[i].TokenID] = true
	}
	for i := 0; i < len(l1coordinatortxs); i++ {
		tokenIDs[l1coordinatortxs[i].TokenID] = true
	}
	for i := 0; i < len(l2txs); i++ {
		// as L2Tx does not have parameter TokenID, get it from the
		// AccountsDB (in the StateDB)
		acc, err := s.GetAccount(l2txs[i].FromIdx)
		if err != nil {
			log.Errorf("could not get account to determine TokenID of L2Tx: FromIdx %d not found: %s", l2txs[i].FromIdx, err.Error())
			return nil, err
		}
		tokenIDs[acc.TokenID] = true
	}
	var tBI []*big.Int
	for t := range tokenIDs {
		tBI = append(tBI, t.BigInt())
	}
	return tBI, nil
}

// processL1Tx process the given L1Tx applying the needed updates to the
// StateDB depending on the transaction Type. It returns the 3 parameters
// related to the Exit (in case of): Idx, ExitAccount, boolean determining if
// the Exit created a new Leaf in the ExitTree.
// And another *common.Account parameter which contains the created account in
// case that has been a new created account and that the StateDB is of type
// TypeSynchronizer.
func (s *StateDB) processL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx, *common.Account, bool, *common.Account, error) {
	// ZKInputs
	if s.zki != nil {
		// Txs
		// s.zki.TxCompressedData[s.i] = tx.TxCompressedData() // uncomment once L1Tx.TxCompressedData is ready
		s.zki.FromIdx[s.i] = tx.FromIdx.BigInt()
		s.zki.ToIdx[s.i] = tx.ToIdx.BigInt()
		s.zki.OnChain[s.i] = big.NewInt(1)

		// L1Txs
		s.zki.LoadAmountF[s.i] = tx.LoadAmount
		s.zki.FromEthAddr[s.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		if tx.FromBJJ != nil {
			s.zki.FromBJJCompressed[s.i] = BJJCompressedTo256BigInts(tx.FromBJJ.Compress())
		}

		// Intermediate States, for all the transactions except for the last one
		if s.i < len(s.zki.ISOnChain) { // len(s.zki.ISOnChain) == nTx
			s.zki.ISOnChain[s.i] = big.NewInt(1)
		}
	}

	switch tx.Type {
	case common.TxTypeForceTransfer:
		// go to the MT account of sender and receiver, and update balance
		// & nonce

		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees
		// 0 for the parameter toIdx, as at L1Tx ToIdx can only be 0 in the Deposit type case.
		err := s.applyTransfer(nil, nil, tx.Tx(), 0)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new account to the MT, update balance of the MT account
		err := s.applyCreateAccount(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
		// TODO applyCreateAccount will return the created account,
		// which in the case type==TypeSynchronizer will be added to an
		// array of created accounts that will be returned

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeDeposit:
		// update balance of the MT account
		err := s.applyDeposit(tx, false)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
	case common.TxTypeDepositTransfer:
		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := s.applyDeposit(tx, true)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
	case common.TxTypeCreateAccountDepositTransfer:
		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := s.applyCreateAccountDepositTransfer(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeForceExit:
		// execute exit flow
		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees
		exitAccount, newExit, err := s.applyExit(nil, nil, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
		return &tx.FromIdx, exitAccount, newExit, nil, nil
	default:
	}

	var createdAccount *common.Account
	if s.typ == TypeSynchronizer && (tx.Type == common.TxTypeCreateAccountDeposit || tx.Type == common.TxTypeCreateAccountDepositTransfer) {
		var err error
		createdAccount, err = s.GetAccount(s.idx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, err
		}
	}

	return nil, nil, false, createdAccount, nil
}

// processL2Tx process the given L2Tx applying the needed updates to the
// StateDB depending on the transaction Type. It returns the 3 parameters
// related to the Exit (in case of): Idx, ExitAccount, boolean determining if
// the Exit created a new Leaf in the ExitTree.
func (s *StateDB) processL2Tx(coordIdxsMap map[common.TokenID]common.Idx, collectedFees map[common.TokenID]*big.Int,
	exitTree *merkletree.MerkleTree, tx *common.PoolL2Tx) (*common.Idx, *common.Account, bool, error) {
	var err error
	// if tx.ToIdx==0, get toIdx by ToEthAddr or ToBJJ
	if tx.ToIdx == common.Idx(0) && tx.AuxToIdx == common.Idx(0) {
		// case when tx.Type== common.TxTypeTransferToEthAddr or common.TxTypeTransferToBJJ
		tx.AuxToIdx, err = s.GetIdxByEthAddrBJJ(tx.ToEthAddr, tx.ToBJJ, tx.TokenID)
		if err != nil {
			return nil, nil, false, err
		}
	}

	// ZKInputs
	if s.zki != nil {
		// Txs
		// s.zki.TxCompressedData[s.i] = tx.TxCompressedData() // uncomment once L1Tx.TxCompressedData is ready
		// s.zki.TxCompressedDataV2[s.i] = tx.TxCompressedDataV2() // uncomment once L2Tx.TxCompressedDataV2 is ready
		s.zki.FromIdx[s.i] = tx.FromIdx.BigInt()
		s.zki.ToIdx[s.i] = tx.ToIdx.BigInt()

		// fill AuxToIdx if needed
		if tx.ToIdx == 0 {
			// use toIdx that can have been filled by tx.ToIdx or
			// if tx.Idx==0 (this case), toIdx is filled by the Idx
			// from db by ToEthAddr&ToBJJ
			s.zki.AuxToIdx[s.i] = tx.AuxToIdx.BigInt()
		}

		if tx.ToBJJ != nil {
			s.zki.ToBJJAy[s.i] = tx.ToBJJ.Y
		}
		s.zki.ToEthAddr[s.i] = common.EthAddrToBigInt(tx.ToEthAddr)

		s.zki.OnChain[s.i] = big.NewInt(0)
		s.zki.NewAccount[s.i] = big.NewInt(0)

		// L2Txs
		// s.zki.RqOffset[s.i] =  // TODO Rq once TxSelector is ready
		// s.zki.RqTxCompressedDataV2[s.i] = // TODO
		// s.zki.RqToEthAddr[s.i] = common.EthAddrToBigInt(tx.RqToEthAddr) // TODO
		// s.zki.RqToBJJAy[s.i] = tx.ToBJJ.Y // TODO
		signature, err := tx.Signature.Decompress()
		if err != nil {
			log.Error(err)
			return nil, nil, false, err
		}
		s.zki.S[s.i] = signature.S
		s.zki.R8x[s.i] = signature.R8.X
		s.zki.R8y[s.i] = signature.R8.Y
	}

	// if StateDB type==TypeSynchronizer, will need to add Nonce
	if s.typ == TypeSynchronizer {
		// as type==TypeSynchronizer, always tx.ToIdx!=0
		acc, err := s.GetAccount(tx.FromIdx)
		if err != nil {
			log.Errorw("GetAccount", "fromIdx", tx.FromIdx, "err", err)
			return nil, nil, false, err
		}
		tx.Nonce = acc.Nonce + 1
		tx.TokenID = acc.TokenID
	}

	switch tx.Type {
	case common.TxTypeTransfer, common.TxTypeTransferToEthAddr, common.TxTypeTransferToBJJ:
		// go to the MT account of sender and receiver, and update
		// balance & nonce
		err = s.applyTransfer(coordIdxsMap, collectedFees, tx.Tx(), tx.AuxToIdx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, err
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, newExit, err := s.applyExit(coordIdxsMap, collectedFees, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, err
		}
		return &tx.FromIdx, exitAccount, newExit, nil
	default:
	}
	return nil, nil, false, nil
}

// applyCreateAccount creates a new account in the account of the depositer, it
// stores the deposit value
func (s *StateDB) applyCreateAccount(tx *common.L1Tx) error {
	account := &common.Account{
		TokenID:   tx.TokenID,
		Nonce:     0,
		Balance:   tx.LoadAmount,
		PublicKey: tx.FromBJJ,
		EthAddr:   tx.FromEthAddr,
	}

	p, err := s.CreateAccount(common.Idx(s.idx+1), account)
	if err != nil {
		return err
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = tx.TokenID.BigInt()
		s.zki.Nonce1[s.i] = big.NewInt(0)
		if babyjub.PointCoordSign(tx.FromBJJ.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = tx.FromBJJ.Y
		s.zki.Balance1[s.i] = tx.LoadAmount
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			s.zki.IsOld0_1[s.i] = big.NewInt(1)
		}
		s.zki.OldKey1[s.i] = p.OldKey.BigInt()
		s.zki.OldValue1[s.i] = p.OldValue.BigInt()

		s.zki.Metadata.NewLastIdxRaw = s.idx + 1
	}

	s.idx = s.idx + 1
	return s.setIdx(s.idx)
}

// applyDeposit updates the balance in the account of the depositer, if
// andTransfer parameter is set to true, the method will also apply the
// Transfer of the L1Tx/DepositTransfer
func (s *StateDB) applyDeposit(tx *common.L1Tx, transfer bool) error {
	// deposit the tx.LoadAmount into the sender account
	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return err
	}
	accSender.Balance = new(big.Int).Add(accSender.Balance, tx.LoadAmount)

	// in case that the tx is a L1Tx>DepositTransfer
	var accReceiver *common.Account
	if transfer {
		accReceiver, err = s.GetAccount(tx.ToIdx)
		if err != nil {
			return err
		}
		// subtract amount to the sender
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
		// add amount to the receiver
		accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)
	}
	// update sender account in localStateDB
	p, err := s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return err
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = accSender.TokenID.BigInt()
		s.zki.Nonce1[s.i] = accSender.Nonce.BigInt()
		if babyjub.PointCoordSign(accSender.PublicKey.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = accSender.PublicKey.Y
		s.zki.Balance1[s.i] = accSender.Balance
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(accSender.EthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		// IsOld0_1, OldKey1, OldValue1 not needed as this is not an insert
	}

	// this is done after updating Sender Account (depositer)
	if transfer {
		// update receiver account in localStateDB
		p, err := s.UpdateAccount(tx.ToIdx, accReceiver)
		if err != nil {
			return err
		}
		if s.zki != nil {
			s.zki.TokenID2[s.i] = accReceiver.TokenID.BigInt()
			s.zki.Nonce2[s.i] = accReceiver.Nonce.BigInt()
			if babyjub.PointCoordSign(accReceiver.PublicKey.X) {
				s.zki.Sign2[s.i] = big.NewInt(1)
			}
			s.zki.Ay2[s.i] = accReceiver.PublicKey.Y
			s.zki.Balance2[s.i] = accReceiver.Balance
			s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
			s.zki.Siblings2[s.i] = siblingsToZKInputFormat(p.Siblings)
			// IsOld0_2, OldKey2, OldValue2 not needed as this is not an insert
		}
	}

	return nil
}

// applyTransfer updates the balance & nonce in the account of the sender, and
// the balance in the account of the receiver.
// Parameter 'toIdx' should be at 0 if the tx already has tx.ToIdx!=0, if
// tx.ToIdx==0, then toIdx!=0, and will be used the toIdx parameter as Idx of
// the receiver. This parameter is used when the tx.ToIdx is not specified and
// the real ToIdx is found trhrough the ToEthAddr or ToBJJ.
func (s *StateDB) applyTransfer(coordIdxsMap map[common.TokenID]common.Idx, collectedFees map[common.TokenID]*big.Int,
	tx common.Tx, auxToIdx common.Idx) error {
	if auxToIdx == common.Idx(0) {
		auxToIdx = tx.ToIdx
	}
	// get sender and receiver accounts from localStateDB
	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		log.Error(err)
		return err
	}
	if !tx.IsL1 {
		// increment nonce
		accSender.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return err
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		accSender.Balance = new(big.Int).Sub(accSender.Balance, feeAndAmount)

		accCoord, err := s.GetAccount(coordIdxsMap[accSender.TokenID])
		if err != nil {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		} else {
			// accumulate the fee for the Coord account
			accumulated := s.accumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if s.typ == TypeSynchronizer {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		}
	} else {
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
	}

	var accReceiver *common.Account
	if tx.FromIdx == auxToIdx {
		// if Sender is the Receiver, reuse 'accSender' pointer,
		// because in the DB the account for 'auxToIdx' won't be
		// updated yet
		accReceiver = accSender
	} else {
		accReceiver, err = s.GetAccount(auxToIdx)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	// add amount-feeAmount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// update sender account in localStateDB
	pSender, err := s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		log.Error(err)
		return err
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = accSender.TokenID.BigInt()
		s.zki.Nonce1[s.i] = accSender.Nonce.BigInt()
		if babyjub.PointCoordSign(accSender.PublicKey.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = accSender.PublicKey.Y
		s.zki.Balance1[s.i] = accSender.Balance
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(accSender.EthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(pSender.Siblings)
	}

	// update receiver account in localStateDB
	pReceiver, err := s.UpdateAccount(auxToIdx, accReceiver)
	if err != nil {
		return err
	}
	if s.zki != nil {
		s.zki.TokenID2[s.i] = accReceiver.TokenID.BigInt()
		s.zki.Nonce2[s.i] = accReceiver.Nonce.BigInt()
		if babyjub.PointCoordSign(accReceiver.PublicKey.X) {
			s.zki.Sign2[s.i] = big.NewInt(1)
		}
		s.zki.Ay2[s.i] = accReceiver.PublicKey.Y
		s.zki.Balance2[s.i] = accReceiver.Balance
		s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
		s.zki.Siblings2[s.i] = siblingsToZKInputFormat(pReceiver.Siblings)
	}

	return nil
}

// applyCreateAccountDepositTransfer, in a single tx, creates a new account,
// makes a deposit, and performs a transfer to another account
func (s *StateDB) applyCreateAccountDepositTransfer(tx *common.L1Tx) error {
	accSender := &common.Account{
		TokenID:   tx.TokenID,
		Nonce:     0,
		Balance:   tx.LoadAmount,
		PublicKey: tx.FromBJJ,
		EthAddr:   tx.FromEthAddr,
	}
	accReceiver, err := s.GetAccount(tx.ToIdx)
	if err != nil {
		return err
	}
	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
	// add amount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// create Account of the Sender
	p, err := s.CreateAccount(common.Idx(s.idx+1), accSender)
	if err != nil {
		return err
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = tx.TokenID.BigInt()
		s.zki.Nonce1[s.i] = big.NewInt(0)
		if babyjub.PointCoordSign(tx.FromBJJ.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = tx.FromBJJ.Y
		s.zki.Balance1[s.i] = tx.LoadAmount
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			s.zki.IsOld0_1[s.i] = big.NewInt(1)
		}
		s.zki.OldKey1[s.i] = p.OldKey.BigInt()
		s.zki.OldValue1[s.i] = p.OldValue.BigInt()

		s.zki.Metadata.NewLastIdxRaw = s.idx + 1
	}

	// update receiver account in localStateDB
	p, err = s.UpdateAccount(tx.ToIdx, accReceiver)
	if err != nil {
		return err
	}
	if s.zki != nil {
		s.zki.TokenID2[s.i] = accReceiver.TokenID.BigInt()
		s.zki.Nonce2[s.i] = accReceiver.Nonce.BigInt()
		if babyjub.PointCoordSign(accReceiver.PublicKey.X) {
			s.zki.Sign2[s.i] = big.NewInt(1)
		}
		s.zki.Ay2[s.i] = accReceiver.PublicKey.Y
		s.zki.Balance2[s.i] = accReceiver.Balance
		s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
		s.zki.Siblings2[s.i] = siblingsToZKInputFormat(p.Siblings)
	}

	s.idx = s.idx + 1
	return s.setIdx(s.idx)
}

// It returns the ExitAccount and a boolean determining if the Exit created a
// new Leaf in the ExitTree.
func (s *StateDB) applyExit(coordIdxsMap map[common.TokenID]common.Idx, collectedFees map[common.TokenID]*big.Int,
	exitTree *merkletree.MerkleTree, tx common.Tx) (*common.Account, bool, error) {
	// 0. subtract tx.Amount from current Account in StateMT
	// add the tx.Amount into the Account (tx.FromIdx) in the ExitMT
	acc, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return nil, false, err
	}

	if !tx.IsL1 {
		// increment nonce
		acc.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return nil, false, err
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		acc.Balance = new(big.Int).Sub(acc.Balance, feeAndAmount)

		accCoord, err := s.GetAccount(coordIdxsMap[acc.TokenID])
		if err != nil {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		} else {
			// accumulate the fee for the Coord account
			accumulated := s.accumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if s.typ == TypeSynchronizer {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		}
	} else {
		acc.Balance = new(big.Int).Sub(acc.Balance, tx.Amount)
	}

	p, err := s.UpdateAccount(tx.FromIdx, acc)
	if err != nil {
		return nil, false, err
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = acc.TokenID.BigInt()
		s.zki.Nonce1[s.i] = acc.Nonce.BigInt()
		if babyjub.PointCoordSign(acc.PublicKey.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = acc.PublicKey.Y
		s.zki.Balance1[s.i] = acc.Balance
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(acc.EthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
	}

	if exitTree == nil {
		return nil, false, nil
	}
	exitAccount, err := getAccountInTreeDB(exitTree.DB(), tx.FromIdx)
	if err == db.ErrNotFound {
		// 1a. if idx does not exist in exitTree:
		// add new leaf 'ExitTreeLeaf', where ExitTreeLeaf.Balance = exitAmount (exitAmount=tx.Amount)
		exitAccount := &common.Account{
			TokenID:   acc.TokenID,
			Nonce:     common.Nonce(1),
			Balance:   tx.Amount,
			PublicKey: acc.PublicKey,
			EthAddr:   acc.EthAddr,
		}
		_, err = createAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
		return exitAccount, true, err
	} else if err != nil {
		return exitAccount, false, err
	}

	// 1b. if idx already exist in exitTree:
	// update account, where account.Balance += exitAmount
	exitAccount.Balance = new(big.Int).Add(exitAccount.Balance, tx.Amount)
	_, err = updateAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
	return exitAccount, false, err
}

// getIdx returns the stored Idx from the localStateDB, which is the last Idx
// used for an Account in the localStateDB.
func (s *StateDB) getIdx() (common.Idx, error) {
	idxBytes, err := s.DB().Get(keyidx)
	if err == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return common.IdxFromBytes(idxBytes[:])
}

// setIdx stores Idx in the localStateDB
func (s *StateDB) setIdx(idx common.Idx) error {
	tx, err := s.DB().NewTx()
	if err != nil {
		return err
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return err
	}
	err = tx.Put(keyidx, idxBytes[:])
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
