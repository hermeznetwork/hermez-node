package statedb

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
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
		return nil, tracerr.Wrap(errors.New("Expected StateDB.zki==nil, something went wrong and it's not empty"))
	}
	defer s.resetZKInputs()

	if len(coordIdxs) >= int(ptc.MaxFeeTx) {
		return nil, tracerr.Wrap(fmt.Errorf("CoordIdxs (%d) length must be smaller than MaxFeeTx (%d)", len(coordIdxs), ptc.MaxFeeTx))
	}

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

	if nTx > int(ptc.MaxTx) {
		return nil, tracerr.Wrap(fmt.Errorf("L1UserTx + L1CoordinatorTx + L2Tx (%d) can not be bigger than MaxTx (%d)", nTx, ptc.MaxTx))
	}
	if len(l1usertxs)+len(l1coordinatortxs) > int(ptc.MaxL1Tx) {
		return nil, tracerr.Wrap(fmt.Errorf("L1UserTx + L1CoordinatorTx (%d) can not be bigger than MaxL1Tx (%d)", len(l1usertxs)+len(l1coordinatortxs), ptc.MaxTx))
	}

	exits := make([]processedExit, nTx)

	if s.typ == TypeBatchBuilder {
		s.zki = common.NewZKInputs(ptc.MaxTx, ptc.MaxL1Tx, ptc.MaxTx, ptc.MaxFeeTx, ptc.NLevels, s.currentBatch.BigInt())
		s.zki.OldLastIdx = s.idx.BigInt()
		s.zki.OldStateRoot = s.mt.Root().BigInt()
	}

	// TBD if ExitTree is only in memory or stored in disk, for the moment
	// is only needed in memory
	if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
		tmpDir, err := ioutil.TempDir("", "hermez-statedb-exittree")
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				log.Errorw("Deleting statedb temp exit tree", "err", err)
			}
		}()
		sto, err := pebble.NewPebbleStorage(tmpDir, false)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		exitTree, err = merkletree.NewMerkleTree(sto, s.mt.MaxLevels())
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}

	// Process L1UserTxs
	for i := 0; i < len(l1usertxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		exitIdx, exitAccount, newExit, createdAccount, err := s.processL1Tx(exitTree, &l1usertxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if s.typ == TypeSynchronizer && createdAccount != nil {
			createdAccounts = append(createdAccounts, *createdAccount)
		}

		if s.zki != nil {
			l1TxData, err := l1usertxs[i].BytesGeneric()
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			s.zki.Metadata.L1TxsData = append(s.zki.Metadata.L1TxsData, l1TxData)

			l1TxDataAvailability, err := l1usertxs[i].BytesDataAvailability(s.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			s.zki.Metadata.L1TxsDataAvailability = append(s.zki.Metadata.L1TxsDataAvailability, l1TxDataAvailability)

			s.zki.ISOutIdx[s.i] = s.idx.BigInt()
			s.zki.ISStateRoot[s.i] = s.mt.Root().BigInt()
			if exitIdx == nil {
				s.zki.ISExitRoot[s.i] = exitTree.Root().BigInt()
			}
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
	}

	// Process L1CoordinatorTxs
	for i := 0; i < len(l1coordinatortxs); i++ {
		exitIdx, _, _, createdAccount, err := s.processL1Tx(exitTree, &l1coordinatortxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
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
				return nil, tracerr.Wrap(err)
			}
			s.zki.Metadata.L1TxsData = append(s.zki.Metadata.L1TxsData, l1TxData)
			l1TxDataAvailability, err := l1coordinatortxs[i].BytesDataAvailability(s.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			s.zki.Metadata.L1TxsDataAvailability = append(s.zki.Metadata.L1TxsDataAvailability, l1TxDataAvailability)

			s.zki.ISOutIdx[s.i] = s.idx.BigInt()
			s.zki.ISStateRoot[s.i] = s.mt.Root().BigInt()
			s.i++
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
		return nil, tracerr.Wrap(err)
	}
	// collectedFees will contain the amount of fee collected for each
	// TokenID
	var collectedFees map[common.TokenID]*big.Int
	if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
		collectedFees = make(map[common.TokenID]*big.Int)
		for tokenID := range coordIdxsMap {
			collectedFees[tokenID] = big.NewInt(0)
		}
	}

	if s.zki != nil {
		// get the feePlanTokens
		feePlanTokens, err := s.getFeePlanTokens(coordIdxs)
		if err != nil {
			log.Error(err)
			return nil, tracerr.Wrap(err)
		}
		copy(s.zki.FeePlanTokens, feePlanTokens)
	}

	// Process L2Txs
	for i := 0; i < len(l2txs); i++ {
		exitIdx, exitAccount, newExit, err := s.processL2Tx(coordIdxsMap, collectedFees, exitTree, &l2txs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if s.zki != nil {
			l2TxData, err := l2txs[i].L2Tx().BytesDataAvailability(s.zki.Metadata.NLevels)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			s.zki.Metadata.L2TxsData = append(s.zki.Metadata.L2TxsData, l2TxData)

			// Intermediate States
			if s.i < nTx-1 {
				s.zki.ISOutIdx[s.i] = s.idx.BigInt()
				s.zki.ISStateRoot[s.i] = s.mt.Root().BigInt()
				s.zki.ISAccFeeOut[s.i] = formatAccumulatedFees(collectedFees, s.zki.FeePlanTokens)
				if exitIdx == nil {
					s.zki.ISExitRoot[s.i] = exitTree.Root().BigInt()
				}
			}
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
	}

	if s.zki != nil {
		for i := s.i - 1; i < int(ptc.MaxTx); i++ {
			if i < int(ptc.MaxTx)-1 {
				s.zki.ISOutIdx[i] = s.idx.BigInt()
				s.zki.ISStateRoot[i] = s.mt.Root().BigInt()
				s.zki.ISAccFeeOut[i] = formatAccumulatedFees(collectedFees, s.zki.FeePlanTokens)
				s.zki.ISExitRoot[i] = exitTree.Root().BigInt()
			}
			if i >= s.i {
				s.zki.TxCompressedData[i] = new(big.Int).SetBytes(common.SignatureConstantBytes)
			}
		}
		isFinalAccFee := formatAccumulatedFees(collectedFees, s.zki.FeePlanTokens)
		copy(s.zki.ISFinalAccFee, isFinalAccFee)
		// before computing the Fees txs, set the ISInitStateRootFee
		s.zki.ISInitStateRootFee = s.mt.Root().BigInt()
	}

	// distribute the AccumulatedFees from the processed L2Txs into the
	// Coordinator Idxs
	iFee := 0
	for idx, accumulatedFee := range s.accumulatedFees {
		cmp := accumulatedFee.Cmp(big.NewInt(0))
		if cmp == 1 { // accumulatedFee>0
			// send the fee to the Idx of the Coordinator for the TokenID
			accCoord, err := s.GetAccount(idx)
			if err != nil {
				log.Errorw("Can not distribute accumulated fees to coordinator account: No coord Idx to receive fee", "idx", idx)
				return nil, tracerr.Wrap(err)
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
			}
			accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
			pFee, err := s.UpdateAccount(idx, accCoord)
			if err != nil {
				log.Error(err)
				return nil, tracerr.Wrap(err)
			}
			if s.zki != nil {
				s.zki.Siblings3[iFee] = siblingsToZKInputFormat(pFee.Siblings)
				s.zki.ISStateRootFee[iFee] = s.mt.Root().BigInt()
			}
		}
		iFee++
	}
	if s.zki != nil {
		for i := len(s.accumulatedFees); i < int(ptc.MaxFeeTx)-1; i++ {
			s.zki.ISStateRootFee[i] = s.mt.Root().BigInt()
		}
		// add Coord Idx to ZKInputs.FeeTxsData
		for i := 0; i < len(coordIdxs); i++ {
			s.zki.FeeIdxs[i] = coordIdxs[i].BigInt()
		}
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
			return nil, tracerr.Wrap(err)
		}
		// 1. generate common.ExitInfo
		ei := common.ExitInfo{
			AccountIdx:  exitIdx,
			MerkleProof: p,
			Balance:     exitAccount.Balance,
		}
		exitInfos = append(exitInfos, ei)
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
	s.zki.Metadata.NewStateRootRaw = s.mt.Root()
	s.zki.Metadata.NewExitRootRaw = exitTree.Root()

	// return ZKInputs as the BatchBuilder will return it to forge the Batch
	return &ProcessTxOutput{
		ZKInputs:           s.zki,
		ExitInfos:          nil,
		CreatedAccounts:    nil,
		CoordinatorIdxsMap: coordIdxsMap,
		CollectedFees:      nil,
	}, nil
}

// getFeePlanTokens returns an array of *big.Int containing a list of tokenIDs
// corresponding to the given CoordIdxs and the processed L2Txs
func (s *StateDB) getFeePlanTokens(coordIdxs []common.Idx) ([]*big.Int, error) {
	var tBI []*big.Int
	for i := 0; i < len(coordIdxs); i++ {
		acc, err := s.GetAccount(coordIdxs[i])
		if err != nil {
			log.Errorf("could not get account to determine TokenID of CoordIdx %d not found: %s", coordIdxs[i], err.Error())
			return nil, tracerr.Wrap(err)
		}
		tBI = append(tBI, acc.TokenID.BigInt())
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
		var err error
		s.zki.TxCompressedData[s.i], err = tx.TxCompressedData()
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		s.zki.FromIdx[s.i] = tx.FromIdx.BigInt()
		s.zki.ToIdx[s.i] = tx.ToIdx.BigInt()
		s.zki.OnChain[s.i] = big.NewInt(1)

		// L1Txs
		depositAmountF16, err := common.NewFloat16(tx.DepositAmount)
		if err != nil {
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		s.zki.DepositAmountF[s.i] = big.NewInt(int64(depositAmountF16))
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
		s.computeEffectiveAmounts(tx)

		// go to the MT account of sender and receiver, and update balance
		// & nonce

		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees
		// 0 for the parameter toIdx, as at L1Tx ToIdx can only be 0 in the Deposit type case.
		err := s.applyTransfer(nil, nil, tx.Tx(), 0)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeCreateAccountDeposit:
		s.computeEffectiveAmounts(tx)

		// add new account to the MT, update balance of the MT account
		err := s.applyCreateAccount(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
		// TODO applyCreateAccount will return the created account,
		// which in the case type==TypeSynchronizer will be added to an
		// array of created accounts that will be returned
	case common.TxTypeDeposit:
		s.computeEffectiveAmounts(tx)

		// update balance of the MT account
		err := s.applyDeposit(tx, false)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeDepositTransfer:
		s.computeEffectiveAmounts(tx)

		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := s.applyDeposit(tx, true)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeCreateAccountDepositTransfer:
		s.computeEffectiveAmounts(tx)

		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := s.applyCreateAccountDepositTransfer(tx)
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
		}
	case common.TxTypeForceExit:
		s.computeEffectiveAmounts(tx)

		// execute exit flow
		// coordIdxsMap is 'nil', as at L1Txs there is no L2 fees
		exitAccount, newExit, err := s.applyExit(nil, nil, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, nil, tracerr.Wrap(err)
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
			return nil, nil, false, nil, tracerr.Wrap(err)
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
		if s.typ == TypeSynchronizer {
			// this should never be reached
			log.Error("WARNING: In StateDB with Synchronizer mode L2.ToIdx can't be 0")
			return nil, nil, false, tracerr.Wrap(fmt.Errorf("In StateDB with Synchronizer mode L2.ToIdx can't be 0"))
		}
		// case when tx.Type== common.TxTypeTransferToEthAddr or common.TxTypeTransferToBJJ
		tx.AuxToIdx, err = s.GetIdxByEthAddrBJJ(tx.ToEthAddr, tx.ToBJJ, tx.TokenID)
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
	}

	// ZKInputs
	if s.zki != nil {
		// Txs
		s.zki.TxCompressedData[s.i], err = tx.TxCompressedData()
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
		s.zki.TxCompressedDataV2[s.i], err = tx.TxCompressedDataV2()
		if err != nil {
			return nil, nil, false, tracerr.Wrap(err)
		}
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
			return nil, nil, false, tracerr.Wrap(err)
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
			return nil, nil, false, tracerr.Wrap(err)
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
			return nil, nil, false, tracerr.Wrap(err)
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, newExit, err := s.applyExit(coordIdxsMap, collectedFees, exitTree, tx.Tx())
		if err != nil {
			log.Error(err)
			return nil, nil, false, tracerr.Wrap(err)
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
		Balance:   tx.EffectiveDepositAmount,
		PublicKey: tx.FromBJJ,
		EthAddr:   tx.FromEthAddr,
	}

	p, err := s.CreateAccount(common.Idx(s.idx+1), account)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.TokenID1[s.i] = tx.TokenID.BigInt()
		s.zki.Nonce1[s.i] = big.NewInt(0)
		if babyjub.PointCoordSign(tx.FromBJJ.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = tx.FromBJJ.Y
		s.zki.Balance1[s.i] = tx.EffectiveDepositAmount
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(tx.FromEthAddr)
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			s.zki.IsOld0_1[s.i] = big.NewInt(1)
		}
		s.zki.OldKey1[s.i] = p.OldKey.BigInt()
		s.zki.OldValue1[s.i] = p.OldValue.BigInt()

		s.zki.Metadata.NewLastIdxRaw = s.idx + 1

		s.zki.AuxFromIdx[s.i] = common.Idx(s.idx + 1).BigInt()
		s.zki.NewAccount[s.i] = big.NewInt(1)

		if s.i < len(s.zki.ISOnChain) { // len(s.zki.ISOnChain) == nTx
			// intermediate states
			s.zki.ISOnChain[s.i] = big.NewInt(1)
		}
	}

	s.idx = s.idx + 1
	return s.setIdx(s.idx)
}

// applyDeposit updates the balance in the account of the depositer, if
// andTransfer parameter is set to true, the method will also apply the
// Transfer of the L1Tx/DepositTransfer
func (s *StateDB) applyDeposit(tx *common.L1Tx, transfer bool) error {
	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return tracerr.Wrap(err)
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
	}

	// add the deposit to the sender
	accSender.Balance = new(big.Int).Add(accSender.Balance, tx.EffectiveDepositAmount)
	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.EffectiveAmount)

	// update sender account in localStateDB
	p, err := s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		// IsOld0_1, OldKey1, OldValue1 not needed as this is not an insert
	}

	// in case that the tx is a L1Tx>DepositTransfer
	var accReceiver *common.Account
	if transfer {
		if tx.ToIdx == tx.FromIdx {
			accReceiver = accSender
		} else {
			accReceiver, err = s.GetAccount(tx.ToIdx)
			if err != nil {
				return tracerr.Wrap(err)
			}
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
		}

		// add amount to the receiver
		accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.EffectiveAmount)

		// update receiver account in localStateDB
		p, err := s.UpdateAccount(tx.ToIdx, accReceiver)
		if err != nil {
			return tracerr.Wrap(err)
		}
		if s.zki != nil {
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
func (s *StateDB) applyTransfer(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int,
	tx common.Tx, auxToIdx common.Idx) error {
	if auxToIdx == common.Idx(0) {
		auxToIdx = tx.ToIdx
	}
	// get sender and receiver accounts from localStateDB
	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		log.Error(err)
		return tracerr.Wrap(err)
	}

	if s.zki != nil {
		// Set the State1 before updating the Sender leaf
		s.zki.TokenID1[s.i] = accSender.TokenID.BigInt()
		s.zki.Nonce1[s.i] = accSender.Nonce.BigInt()
		if babyjub.PointCoordSign(accSender.PublicKey.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = accSender.PublicKey.Y
		s.zki.Balance1[s.i] = accSender.Balance
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(accSender.EthAddr)
	}
	if !tx.IsL1 {
		// increment nonce
		accSender.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return tracerr.Wrap(err)
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		accSender.Balance = new(big.Int).Sub(accSender.Balance, feeAndAmount)

		if _, ok := coordIdxsMap[accSender.TokenID]; ok {
			accCoord, err := s.GetAccount(coordIdxsMap[accSender.TokenID])
			if err != nil {
				return tracerr.Wrap(fmt.Errorf("Can not use CoordIdx that does not exist in the tree. TokenID: %d, CoordIdx: %d", accSender.TokenID, coordIdxsMap[accSender.TokenID]))
			}
			// accumulate the fee for the Coord account
			accumulated := s.accumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		} else {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		}
	} else {
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
	}

	// update sender account in localStateDB
	pSender, err := s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		log.Error(err)
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(pSender.Siblings)
	}

	var accReceiver *common.Account
	if auxToIdx == tx.FromIdx {
		// if Sender is the Receiver, reuse 'accSender' pointer,
		// because in the DB the account for 'auxToIdx' won't be
		// updated yet
		accReceiver = accSender
	} else {
		accReceiver, err = s.GetAccount(auxToIdx)
		if err != nil {
			log.Error(err)
			return tracerr.Wrap(err)
		}
	}
	if s.zki != nil {
		// Set the State2 before updating the Receiver leaf
		s.zki.TokenID2[s.i] = accReceiver.TokenID.BigInt()
		s.zki.Nonce2[s.i] = accReceiver.Nonce.BigInt()
		if babyjub.PointCoordSign(accReceiver.PublicKey.X) {
			s.zki.Sign2[s.i] = big.NewInt(1)
		}
		s.zki.Ay2[s.i] = accReceiver.PublicKey.Y
		s.zki.Balance2[s.i] = accReceiver.Balance
		s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
	}

	// add amount-feeAmount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// update receiver account in localStateDB
	pReceiver, err := s.UpdateAccount(auxToIdx, accReceiver)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings2[s.i] = siblingsToZKInputFormat(pReceiver.Siblings)
	}

	return nil
}

// applyCreateAccountDepositTransfer, in a single tx, creates a new account,
// makes a deposit, and performs a transfer to another account
func (s *StateDB) applyCreateAccountDepositTransfer(tx *common.L1Tx) error {
	auxFromIdx := common.Idx(s.idx + 1)
	accSender := &common.Account{
		TokenID:   tx.TokenID,
		Nonce:     0,
		Balance:   tx.EffectiveDepositAmount,
		PublicKey: tx.FromBJJ,
		EthAddr:   tx.FromEthAddr,
	}

	if s.zki != nil {
		// Set the State1 before updating the Sender leaf
		s.zki.TokenID1[s.i] = tx.TokenID.BigInt()
		s.zki.Nonce1[s.i] = big.NewInt(0)
		if babyjub.PointCoordSign(tx.FromBJJ.X) {
			s.zki.Sign1[s.i] = big.NewInt(1)
		}
		s.zki.Ay1[s.i] = tx.FromBJJ.Y
		s.zki.Balance1[s.i] = tx.EffectiveDepositAmount
		s.zki.EthAddr1[s.i] = common.EthAddrToBigInt(tx.FromEthAddr)
	}

	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.EffectiveAmount)

	// create Account of the Sender
	p, err := s.CreateAccount(common.Idx(s.idx+1), accSender)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			s.zki.IsOld0_1[s.i] = big.NewInt(1)
		}
		s.zki.OldKey1[s.i] = p.OldKey.BigInt()
		s.zki.OldValue1[s.i] = p.OldValue.BigInt()

		s.zki.Metadata.NewLastIdxRaw = s.idx + 1

		s.zki.AuxFromIdx[s.i] = auxFromIdx.BigInt()
		s.zki.NewAccount[s.i] = big.NewInt(1)

		// intermediate states
		s.zki.ISOnChain[s.i] = big.NewInt(1)
	}
	var accReceiver *common.Account
	if tx.ToIdx == auxFromIdx {
		accReceiver = accSender
	} else {
		accReceiver, err = s.GetAccount(tx.ToIdx)
		if err != nil {
			log.Error(err)
			return tracerr.Wrap(err)
		}
	}

	if s.zki != nil {
		// Set the State2 before updating the Receiver leaf
		s.zki.TokenID2[s.i] = accReceiver.TokenID.BigInt()
		s.zki.Nonce2[s.i] = accReceiver.Nonce.BigInt()
		if babyjub.PointCoordSign(accReceiver.PublicKey.X) {
			s.zki.Sign2[s.i] = big.NewInt(1)
		}
		s.zki.Ay2[s.i] = accReceiver.PublicKey.Y
		s.zki.Balance2[s.i] = accReceiver.Balance
		s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(accReceiver.EthAddr)
	}

	// add amount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.EffectiveAmount)

	// update receiver account in localStateDB
	p, err = s.UpdateAccount(tx.ToIdx, accReceiver)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings2[s.i] = siblingsToZKInputFormat(p.Siblings)
	}

	s.idx = s.idx + 1
	return s.setIdx(s.idx)
}

// It returns the ExitAccount and a boolean determining if the Exit created a
// new Leaf in the ExitTree.
func (s *StateDB) applyExit(coordIdxsMap map[common.TokenID]common.Idx,
	collectedFees map[common.TokenID]*big.Int, exitTree *merkletree.MerkleTree,
	tx common.Tx) (*common.Account, bool, error) {
	// 0. subtract tx.Amount from current Account in StateMT
	// add the tx.Amount into the Account (tx.FromIdx) in the ExitMT
	acc, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
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

		s.zki.NewExit[s.i] = big.NewInt(1)
	}

	if !tx.IsL1 {
		// increment nonce
		acc.Nonce++

		// compute fee and subtract it from the accSender
		fee, err := common.CalcFeeAmount(tx.Amount, *tx.Fee)
		if err != nil {
			return nil, false, tracerr.Wrap(err)
		}
		feeAndAmount := new(big.Int).Add(tx.Amount, fee)
		acc.Balance = new(big.Int).Sub(acc.Balance, feeAndAmount)

		if _, ok := coordIdxsMap[acc.TokenID]; ok {
			accCoord, err := s.GetAccount(coordIdxsMap[acc.TokenID])
			if err != nil {
				return nil, false, tracerr.Wrap(fmt.Errorf("Can not use CoordIdx that does not exist in the tree. TokenID: %d, CoordIdx: %d", acc.TokenID, coordIdxsMap[acc.TokenID]))
			}
			// accumulate the fee for the Coord account
			accumulated := s.accumulatedFees[accCoord.Idx]
			accumulated.Add(accumulated, fee)

			if s.typ == TypeSynchronizer || s.typ == TypeBatchBuilder {
				collected := collectedFees[accCoord.TokenID]
				collected.Add(collected, fee)
			}
		} else {
			log.Debugw("No coord Idx to receive fee", "tx", tx)
		}
	} else {
		acc.Balance = new(big.Int).Sub(acc.Balance, tx.Amount)
	}

	p, err := s.UpdateAccount(tx.FromIdx, acc)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
	}
	if s.zki != nil {
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(p.Siblings)
	}

	if exitTree == nil {
		return nil, false, nil
	}
	exitAccount, err := getAccountInTreeDB(exitTree.DB(), tx.FromIdx)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		// 1a. if idx does not exist in exitTree:
		// add new leaf 'ExitTreeLeaf', where ExitTreeLeaf.Balance = exitAmount (exitAmount=tx.Amount)
		exitAccount := &common.Account{
			TokenID:   acc.TokenID,
			Nonce:     common.Nonce(0),
			Balance:   tx.Amount,
			PublicKey: acc.PublicKey,
			EthAddr:   acc.EthAddr,
		}
		if s.zki != nil {
			// Set the State2 before creating the Exit leaf
			s.zki.TokenID2[s.i] = acc.TokenID.BigInt()
			s.zki.Nonce2[s.i] = big.NewInt(0)
			if babyjub.PointCoordSign(acc.PublicKey.X) {
				s.zki.Sign2[s.i] = big.NewInt(1)
			}
			s.zki.Ay2[s.i] = acc.PublicKey.Y
			s.zki.Balance2[s.i] = tx.Amount
			s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(acc.EthAddr)
		}
		p, err = createAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
		if err != nil {
			return nil, false, tracerr.Wrap(err)
		}
		if s.zki != nil {
			s.zki.Siblings2[s.i] = siblingsToZKInputFormat(p.Siblings)
			if p.IsOld0 {
				s.zki.IsOld0_2[s.i] = big.NewInt(1)
			}
			s.zki.OldKey2[s.i] = p.OldKey.BigInt()
			s.zki.OldValue2[s.i] = p.OldValue.BigInt()
			s.zki.ISExitRoot[s.i] = exitTree.Root().BigInt()
		}
		return exitAccount, true, nil
	} else if err != nil {
		return exitAccount, false, tracerr.Wrap(err)
	}

	// 1b. if idx already exist in exitTree:
	if s.zki != nil {
		// Set the State2 before updating the Exit leaf
		s.zki.TokenID2[s.i] = acc.TokenID.BigInt()
		s.zki.Nonce2[s.i] = big.NewInt(0)
		if babyjub.PointCoordSign(acc.PublicKey.X) {
			s.zki.Sign2[s.i] = big.NewInt(1)
		}
		s.zki.Ay2[s.i] = acc.PublicKey.Y
		s.zki.Balance2[s.i] = tx.Amount
		s.zki.EthAddr2[s.i] = common.EthAddrToBigInt(acc.EthAddr)
	}

	// update account, where account.Balance += exitAmount
	exitAccount.Balance = new(big.Int).Add(exitAccount.Balance, tx.Amount)
	p, err = updateAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
	if err != nil {
		return nil, false, tracerr.Wrap(err)
	}

	if s.zki != nil {
		s.zki.Siblings2[s.i] = siblingsToZKInputFormat(p.Siblings)
		if p.IsOld0 {
			s.zki.IsOld0_2[s.i] = big.NewInt(1)
		}
		s.zki.OldKey2[s.i] = p.OldKey.BigInt()
		s.zki.OldValue2[s.i] = p.OldValue.BigInt()
	}

	return exitAccount, false, nil
}

// computeEffectiveAmounts checks that the L1Tx data is correct
func (s *StateDB) computeEffectiveAmounts(tx *common.L1Tx) {
	tx.EffectiveAmount = tx.Amount
	tx.EffectiveDepositAmount = tx.DepositAmount

	if !tx.UserOrigin {
		// case where the L1Tx is generated by the Coordinator
		tx.EffectiveAmount = big.NewInt(0)
		tx.EffectiveDepositAmount = big.NewInt(0)
		return
	}

	if tx.Type == common.TxTypeCreateAccountDeposit {
		return
	}

	if tx.ToIdx >= common.UserThreshold && tx.FromIdx == common.Idx(0) {
		// CreateAccountDepositTransfer case
		cmp := tx.DepositAmount.Cmp(tx.Amount)
		if cmp == -1 { // DepositAmount<Amount
			tx.EffectiveAmount = big.NewInt(0)
			return
		}
		return
	}

	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: can not get account for tx.FromIdx: %d", tx.FromIdx)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that tx.TokenID corresponds to the Sender account TokenID
	if tx.TokenID != accSender.TokenID {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: tx.TokenID (%d) !=sender account TokenID (%d)", tx.TokenID, accSender.TokenID)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that Sender has enough balance
	bal := accSender.Balance
	if tx.DepositAmount != nil {
		bal = new(big.Int).Add(bal, tx.EffectiveDepositAmount)
	}
	cmp := bal.Cmp(tx.Amount)
	if cmp == -1 {
		log.Debugf("EffectiveAmount = 0: Not enough funds (%s<%s)", bal.String(), tx.Amount.String())
		tx.EffectiveAmount = big.NewInt(0)
		return
	}

	// check that the tx.FromEthAddr is the same than the EthAddress of the
	// Sender
	if !bytes.Equal(tx.FromEthAddr.Bytes(), accSender.EthAddr.Bytes()) {
		log.Debugf("EffectiveAmount = 0: tx.FromEthAddr (%s) must be the same EthAddr of the sender account by the Idx (%s)", tx.FromEthAddr.Hex(), accSender.EthAddr.Hex())
		tx.EffectiveAmount = big.NewInt(0)
	}

	if tx.ToIdx == common.Idx(1) || tx.ToIdx == common.Idx(0) {
		// if transfer is Exit type, there are no more checks
		return
	}

	// check that TokenID is the same for Sender & Receiver account
	accReceiver, err := s.GetAccount(tx.ToIdx)
	if err != nil {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: can not get account for tx.ToIdx: %d", tx.ToIdx)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}
	if accSender.TokenID != accReceiver.TokenID {
		log.Debugf("EffectiveAmount & EffectiveDepositAmount = 0: sender account TokenID (%d) != receiver account TokenID (%d)", accSender.TokenID, accReceiver.TokenID)
		tx.EffectiveDepositAmount = big.NewInt(0)
		tx.EffectiveAmount = big.NewInt(0)
		return
	}
}

// GetIdx returns the stored Idx from the localStateDB, which is the last Idx
// used for an Account in the localStateDB.
func (s *StateDB) GetIdx() (common.Idx, error) {
	idxBytes, err := s.DB().Get(keyidx)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return common.IdxFromBytes(idxBytes[:])
}

// setIdx stores Idx in the localStateDB
func (s *StateDB) setIdx(idx common.Idx) error {
	tx, err := s.DB().NewTx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = tx.Put(keyidx, idxBytes[:])
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}
