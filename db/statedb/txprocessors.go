package statedb

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/memory"
)

var (
	// keyidx is used as key in the db to store the current Idx
	keyidx = []byte("idx")
)

func (s *StateDB) resetZKInputs() {
	s.zki = nil
	s.i = 0
}

type processedExit struct {
	exit    bool
	newExit bool
	idx     common.Idx
	acc     common.Account
}

// ProcessTxs process the given L1Txs & L2Txs applying the needed updates to
// the StateDB depending on the transaction Type. Returns the common.ZKInputs
// to generate the SnarkProof later used by the BatchBuilder, and if
// cmpExitTree is set to true, returns common.ExitTreeLeaf that is later used
// by the Synchronizer to update the HistoryDB.
func (s *StateDB) ProcessTxs(cmpExitTree, cmpZKInputs bool, l1usertxs, l1coordinatortxs []*common.L1Tx, l2txs []*common.PoolL2Tx) (*common.ZKInputs, []*common.ExitInfo, error) {
	var err error
	var exitTree *merkletree.MerkleTree

	if s.zki != nil {
		return nil, nil, errors.New("Expected StateDB.zki==nil, something went wrong and it's not empty")
	}
	defer s.resetZKInputs()

	nTx := len(l1usertxs) + len(l1coordinatortxs) + len(l2txs)
	if nTx == 0 {
		// TODO return ZKInputs of batch without txs
		return nil, nil, nil
	}
	exits := make([]processedExit, nTx)

	if cmpZKInputs {
		s.zki = common.NewZKInputs(nTx, 24, 32) // TODO this values will be parameters of the function, taken from config file/coordinator call
		s.zki.OldLastIdx = (s.idx - 1).BigInt()
		s.zki.OldStateRoot = s.mt.Root().BigInt()
	}

	// TBD if ExitTree is only in memory or stored in disk, for the moment
	// only needed in memory
	if cmpExitTree {
		exitTree, err = merkletree.NewMerkleTree(memory.NewMemoryStorage(), s.mt.MaxLevels())
		if err != nil {
			return nil, nil, err
		}
	}

	// assumption: l1usertx are sorted by L1Tx.Position
	for _, tx := range l1usertxs {
		exitIdx, exitAccount, newExit, err := s.processL1Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil && cmpExitTree {
			exits[s.i] = processedExit{
				exit:    true,
				newExit: newExit,
				idx:     *exitIdx,
				acc:     *exitAccount,
			}
		}
		if s.zki != nil {
			s.i++
		}
	}
	for _, tx := range l1coordinatortxs {
		exitIdx, exitAccount, newExit, err := s.processL1Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil {
			log.Error("Unexpected Exit in L1CoordinatorTx")
		}
		if exitIdx != nil && cmpExitTree {
			exits[s.i] = processedExit{
				exit:    true,
				newExit: newExit,
				idx:     *exitIdx,
				acc:     *exitAccount,
			}
		}
		if s.zki != nil {
			s.i++
		}
	}
	for _, tx := range l2txs {
		exitIdx, exitAccount, newExit, err := s.processL2Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil && cmpExitTree {
			exits[s.i] = processedExit{
				exit:    true,
				newExit: newExit,
				idx:     *exitIdx,
				acc:     *exitAccount,
			}
		}
		if s.zki != nil {
			s.i++
		}
	}

	if !cmpExitTree && !cmpZKInputs {
		return nil, nil, nil
	}

	// once all txs processed (exitTree root frozen), for each Exit,
	// generate common.ExitInfo data
	var exitInfos []*common.ExitInfo
	for i := 0; i < nTx; i++ {
		if !exits[i].exit {
			continue
		}
		exitIdx := exits[i].idx
		exitAccount := exits[i].acc

		// 0. generate MerkleProof
		p, err := exitTree.GenerateCircomVerifierProof(exitIdx.BigInt(), nil)
		if err != nil {
			return nil, nil, err
		}
		// 1. compute nullifier
		exitAccStateValue, err := exitAccount.HashValue()
		if err != nil {
			return nil, nil, err
		}
		nullifier, err := poseidon.Hash([]*big.Int{
			exitAccStateValue,
			big.NewInt(int64(s.currentBatch)),
			exitTree.Root().BigInt(),
		})
		if err != nil {
			return nil, nil, err
		}
		// 2. generate common.ExitInfo
		ei := &common.ExitInfo{
			AccountIdx:  exitIdx,
			MerkleProof: p,
			Nullifier:   nullifier,
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
			s.zki.Siblings2[i] = p.Siblings
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
	if !cmpZKInputs {
		return nil, exitInfos, nil
	}

	// compute last ZKInputs parameters
	s.zki.GlobalChainID = big.NewInt(0) // TODO, 0: ethereum, this will be get from config file
	// zki.FeeIdxs = ? // TODO, this will be get from the config file
	tokenIDs, err := s.getTokenIDsBigInt(l1usertxs, l1coordinatortxs, l2txs)
	if err != nil {
		return nil, nil, err
	}
	s.zki.FeePlanTokens = tokenIDs

	// s.zki.ISInitStateRootFee = s.mt.Root().BigInt()

	// TODO once the Node Config sets the Accounts where to send the Fees
	// compute fees & update ZKInputs

	// return exitInfos, so Synchronizer will be able to store it into
	// HistoryDB for the concrete BatchNum
	return s.zki, exitInfos, nil
}

// getTokenIDsBigInt returns the list of TokenIDs in *big.Int format
func (s *StateDB) getTokenIDsBigInt(l1usertxs, l1coordinatortxs []*common.L1Tx, l2txs []*common.PoolL2Tx) ([]*big.Int, error) {
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
		acc, err := s.GetAccount(l2txs[i].ToIdx)
		if err != nil {
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
func (s *StateDB) processL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx, *common.Account, bool, error) {
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

		// Intermediate States
		s.zki.ISOnChain[s.i] = big.NewInt(1)
	}

	switch tx.Type {
	case common.TxTypeForceTransfer, common.TxTypeTransfer:
		// go to the MT account of sender and receiver, and update balance
		// & nonce
		err := s.applyTransfer(tx.Tx())
		if err != nil {
			return nil, nil, false, err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new account to the MT, update balance of the MT account
		err := s.applyCreateAccount(tx)
		if err != nil {
			return nil, nil, false, err
		}

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeDeposit:
		// update balance of the MT account
		err := s.applyDeposit(tx, false)
		if err != nil {
			return nil, nil, false, err
		}
	case common.TxTypeDepositTransfer:
		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := s.applyDeposit(tx, true)
		if err != nil {
			return nil, nil, false, err
		}
	case common.TxTypeCreateAccountDepositTransfer:
		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := s.applyCreateAccountDepositTransfer(tx)
		if err != nil {
			return nil, nil, false, err
		}

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, newExit, err := s.applyExit(exitTree, tx.Tx())
		if err != nil {
			return nil, nil, false, err
		}
		return &tx.FromIdx, exitAccount, newExit, nil
	default:
	}

	return nil, nil, false, nil
}

// processL2Tx process the given L2Tx applying the needed updates to the
// StateDB depending on the transaction Type. It returns the 3 parameters
// related to the Exit (in case of): Idx, ExitAccount, boolean determining if
// the Exit created a new Leaf in the ExitTree.
func (s *StateDB) processL2Tx(exitTree *merkletree.MerkleTree, tx *common.PoolL2Tx) (*common.Idx, *common.Account, bool, error) {
	var err error
	// ZKInputs
	if s.zki != nil {
		// Txs
		// s.zki.TxCompressedData[s.i] = tx.TxCompressedData() // uncomment once L1Tx.TxCompressedData is ready
		// s.zki.TxCompressedDataV2[s.i] = tx.TxCompressedDataV2() // uncomment once L2Tx.TxCompressedDataV2 is ready
		s.zki.FromIdx[s.i] = tx.FromIdx.BigInt()
		s.zki.ToIdx[s.i] = tx.ToIdx.BigInt()

		// fill AuxToIdx if needed
		if tx.ToIdx == common.Idx(0) {
			var idx common.Idx
			if !bytes.Equal(tx.ToEthAddr.Bytes(), common.EmptyAddr.Bytes()) && tx.ToBJJ == nil {
				// case ToEthAddr!=0 && ToBJJ=0
				idx, err = s.GetIdxByEthAddr(tx.ToEthAddr)
				if err != nil {
					return nil, nil, false, ErrToIdxNotFound
				}
			} else if !bytes.Equal(tx.ToEthAddr.Bytes(), common.EmptyAddr.Bytes()) && tx.ToBJJ != nil {
				// case ToEthAddr!=0 && ToBJJ!=0
				idx, err = s.GetIdxByEthAddrBJJ(tx.ToEthAddr, tx.ToBJJ)
				if err != nil {
					return nil, nil, false, ErrToIdxNotFound
				}
			} else {
				// rest of cases (included case ToEthAddr==0) are not possible
				return nil, nil, false, ErrToIdxNotFound
			}
			s.zki.AuxToIdx[s.i] = idx.BigInt()
		}
		s.zki.ToBJJAy[s.i] = tx.ToBJJ.Y
		s.zki.ToEthAddr[s.i] = common.EthAddrToBigInt(tx.ToEthAddr)

		s.zki.OnChain[s.i] = big.NewInt(0)
		s.zki.NewAccount[s.i] = big.NewInt(0)

		// L2Txs
		// s.zki.RqOffset[s.i] =  // TODO Rq once TxSelector is ready
		// s.zki.RqTxCompressedDataV2[s.i] = // TODO
		// s.zki.RqToEthAddr[s.i] = common.EthAddrToBigInt(tx.RqToEthAddr) // TODO
		// s.zki.RqToBJJAy[s.i] = tx.ToBJJ.Y // TODO
		s.zki.S[s.i] = tx.Signature.S
		s.zki.R8x[s.i] = tx.Signature.R8.X
		s.zki.R8y[s.i] = tx.Signature.R8.Y
	}

	switch tx.Type {
	case common.TxTypeTransfer:
		// go to the MT account of sender and receiver, and update
		// balance & nonce
		err = s.applyTransfer(tx.Tx())
		if err != nil {
			return nil, nil, false, err
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, newExit, err := s.applyExit(exitTree, tx.Tx())
		if err != nil {
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
// the balance in the account of the receiver
func (s *StateDB) applyTransfer(tx *common.Tx) error {
	// get sender and receiver accounts from localStateDB
	accSender, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return err
	}
	accReceiver, err := s.GetAccount(tx.ToIdx)
	if err != nil {
		return err
	}

	// increment nonce
	accSender.Nonce++

	// subtract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
	// add amount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// update sender account in localStateDB
	pSender, err := s.UpdateAccount(tx.FromIdx, accSender)
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
		s.zki.Siblings1[s.i] = siblingsToZKInputFormat(pSender.Siblings)
	}

	// update receiver account in localStateDB
	pReceiver, err := s.UpdateAccount(tx.ToIdx, accReceiver)
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
	accSender.Balance = new(big.Int).Add(accSender.Balance, tx.LoadAmount)
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
func (s *StateDB) applyExit(exitTree *merkletree.MerkleTree, tx *common.Tx) (*common.Account, bool, error) {
	// 0. subtract tx.Amount from current Account in StateMT
	// add the tx.Amount into the Account (tx.FromIdx) in the ExitMT
	acc, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return nil, false, err
	}
	acc.Balance = new(big.Int).Sub(acc.Balance, tx.Amount)
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
	return common.IdxFromBytes(idxBytes[:4])
}

// setIdx stores Idx in the localStateDB
func (s *StateDB) setIdx(idx common.Idx) error {
	tx, err := s.DB().NewTx()
	if err != nil {
		return err
	}
	err = tx.Put(keyidx, idx.Bytes())
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
