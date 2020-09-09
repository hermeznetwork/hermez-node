package statedb

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/memory"
)

var (
	// keyidx is used as key in the db to store the current Idx
	keyidx = []byte("idx")

	ffAddr = ethCommon.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")
)

func (s *StateDB) resetZKInputs() {
	s.zki = nil
	s.i = 0
}

// ProcessTxs process the given L1Txs & L2Txs applying the needed updates to
// the StateDB depending on the transaction Type. Returns the common.ZKInputs
// to generate the SnarkProof later used by the BatchBuilder, and if
// cmpExitTree is set to true, returns common.ExitTreeLeaf that is later used
// by the Synchronizer to update the HistoryDB.
func (s *StateDB) ProcessTxs(cmpExitTree, cmpZKInputs bool, l1usertxs, l1coordinatortxs []*common.L1Tx, l2txs []*common.PoolL2Tx) (*common.ZKInputs, []*common.ExitInfo, error) {
	var err error
	var exitTree *merkletree.MerkleTree
	exits := make(map[common.Idx]common.Account)

	if s.zki != nil {
		return nil, nil, errors.New("Expected StateDB.zki==nil, something went wrong ans is not empty")
	}
	defer s.resetZKInputs()

	nTx := len(l1usertxs) + len(l1coordinatortxs) + len(l2txs)
	if nTx == 0 {
		return nil, nil, nil // TBD if return an error in the case of no Txs to process
	}

	if cmpZKInputs {
		s.zki = common.NewZKInputs(nTx, 24, 32) // TODO this values will be parameters of the function
	}

	// TBD if ExitTree is only in memory or stored in disk, for the moment
	// only needed in memory
	exitTree, err = merkletree.NewMerkleTree(memory.NewMemoryStorage(), s.mt.MaxLevels())
	if err != nil {
		return nil, nil, err
	}

	// assumption: l1usertx are sorted by L1Tx.Position
	for _, tx := range l1usertxs {
		exitIdx, exitAccount, err := s.processL1Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil && cmpExitTree {
			exits[*exitIdx] = *exitAccount
		}
		if s.zki != nil {
			s.i++
		}
	}
	for _, tx := range l1coordinatortxs {
		exitIdx, exitAccount, err := s.processL1Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil && cmpExitTree {
			exits[*exitIdx] = *exitAccount
		}
		if s.zki != nil {
			s.i++
		}
	}
	for _, tx := range l2txs {
		exitIdx, exitAccount, err := s.processL2Tx(exitTree, tx)
		if err != nil {
			return nil, nil, err
		}
		if exitIdx != nil && cmpExitTree {
			exits[*exitIdx] = *exitAccount
		}
		if s.zki != nil {
			s.i++
		}
	}

	if !cmpExitTree && !cmpZKInputs {
		return nil, nil, nil
	}

	// once all txs processed (exitTree root frozen), for each leaf
	// generate common.ExitInfo data
	var exitInfos []*common.ExitInfo
	for exitIdx, exitAccount := range exits {
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
	}
	if !cmpZKInputs {
		return nil, exitInfos, nil
	}

	// compute last ZKInputs parameters
	s.zki.OldLastIdx = (s.idx - 1).BigInt()
	s.zki.OldStateRoot = s.mt.Root().BigInt()
	s.zki.GlobalChainID = big.NewInt(0) // TODO, 0: ethereum, get this from config file
	// zki.FeeIdxs = ? // TODO, this will be get from the config file
	tokenIDs, err := s.getTokenIDsBigInt(l1usertxs, l1coordinatortxs, l2txs)
	if err != nil {
		return nil, nil, err
	}
	s.zki.FeePlanTokens = tokenIDs

	// s.zki.ISInitStateRootFee = s.mt.Root().BigInt()
	// compute fees

	// once fees are computed

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
// StateDB depending on the transaction Type.
func (s *StateDB) processL1Tx(exitTree *merkletree.MerkleTree, tx *common.L1Tx) (*common.Idx, *common.Account, error) {
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
			s.zki.FromBJJCompressed[s.i] = common.BJJCompressedTo256BigInts(tx.FromBJJ.Compress())
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
			return nil, nil, err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new account to the MT, update balance of the MT account
		err := s.applyCreateAccount(tx)
		if err != nil {
			return nil, nil, err
		}

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeDeposit:
		// update balance of the MT account
		err := s.applyDeposit(tx, false)
		if err != nil {
			return nil, nil, err
		}
	case common.TxTypeDepositTransfer:
		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := s.applyDeposit(tx, true)
		if err != nil {
			return nil, nil, err
		}
	case common.TxTypeCreateAccountDepositTransfer:
		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := s.applyCreateAccount(tx)
		if err != nil {
			return nil, nil, err
		}
		err = s.applyTransfer(tx.Tx())
		if err != nil {
			return nil, nil, err
		}

		if s.zki != nil {
			s.zki.AuxFromIdx[s.i] = s.idx.BigInt() // last s.idx is the one used for creating the new account
			s.zki.NewAccount[s.i] = big.NewInt(1)
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, err := s.applyExit(exitTree, tx.Tx())
		if err != nil {
			return nil, nil, err
		}
		return &tx.FromIdx, exitAccount, nil
	default:
	}

	return nil, nil, nil
}

// processL2Tx process the given L2Tx applying the needed updates to
// the StateDB depending on the transaction Type.
func (s *StateDB) processL2Tx(exitTree *merkletree.MerkleTree, tx *common.PoolL2Tx) (*common.Idx, *common.Account, error) {
	// ZKInputs
	if s.zki != nil {
		// Txs
		// s.zki.TxCompressedData[s.i] = tx.TxCompressedData() // uncomment once L1Tx.TxCompressedData is ready
		// s.zki.TxCompressedDataV2[s.i] = tx.TxCompressedDataV2() // uncomment once L2Tx.TxCompressedDataV2 is ready
		s.zki.FromIdx[s.i] = tx.FromIdx.BigInt()
		s.zki.ToIdx[s.i] = tx.ToIdx.BigInt()

		// fill AuxToIdx if needed
		if tx.ToIdx == common.Idx(0) {
			// Idx not set in the Tx, get it from DB through ToEthAddr or ToBJJ
			var idx common.Idx
			if !bytes.Equal(tx.ToEthAddr.Bytes(), ffAddr.Bytes()) {
				idx = s.getIdxByEthAddr(tx.ToEthAddr)
				if idx == common.Idx(0) {
					return nil, nil, fmt.Errorf("Idx can not be found for given tx.FromEthAddr")
				}
			} else {
				idx = s.getIdxByBJJ(tx.ToBJJ)
				if idx == common.Idx(0) {
					return nil, nil, fmt.Errorf("Idx can not be found for given tx.FromBJJ")
				}
			}
			s.zki.AuxToIdx[s.i] = idx.BigInt()
		}
		s.zki.ToBJJAy[s.i] = tx.ToBJJ.Y
		s.zki.ToEthAddr[s.i] = common.EthAddrToBigInt(tx.ToEthAddr)

		s.zki.OnChain[s.i] = big.NewInt(0)
		s.zki.NewAccount[s.i] = big.NewInt(0)

		// L2Txs
		// s.zki.RqOffset[s.i] =  // TODO
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
		err := s.applyTransfer(tx.Tx())
		if err != nil {
			return nil, nil, err
		}
	case common.TxTypeExit:
		// execute exit flow
		exitAccount, err := s.applyExit(exitTree, tx.Tx())
		if err != nil {
			return nil, nil, err
		}
		return &tx.FromIdx, exitAccount, nil
	default:
	}
	return nil, nil, nil
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

	_, err := s.CreateAccount(common.Idx(s.idx+1), account)
	if err != nil {
		return err
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
	if transfer {
		accReceiver, err := s.GetAccount(tx.ToIdx)
		if err != nil {
			return err
		}
		// subtract amount to the sender
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
		// add amount to the receiver
		accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)
		// update receiver account in localStateDB
		_, err = s.UpdateAccount(tx.ToIdx, accReceiver)
		if err != nil {
			return err
		}
	}
	// update sender account in localStateDB
	_, err = s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return err
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

	// update receiver account in localStateDB
	_, err = s.UpdateAccount(tx.ToIdx, accReceiver)
	if err != nil {
		return err
	}
	// update sender account in localStateDB
	_, err = s.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return err
	}

	return nil
}

func (s *StateDB) applyExit(exitTree *merkletree.MerkleTree, tx *common.Tx) (*common.Account, error) {
	// 0. subtract tx.Amount from current Account in StateMT
	// add the tx.Amount into the Account (tx.FromIdx) in the ExitMT
	acc, err := s.GetAccount(tx.FromIdx)
	if err != nil {
		return nil, err
	}
	acc.Balance = new(big.Int).Sub(acc.Balance, tx.Amount)
	_, err = s.UpdateAccount(tx.FromIdx, acc)
	if err != nil {
		return nil, err
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
		return exitAccount, err
	} else if err != nil {
		return exitAccount, err
	}

	// 1b. if idx already exist in exitTree:
	// update account, where account.Balance += exitAmount
	exitAccount.Balance = new(big.Int).Add(exitAccount.Balance, tx.Amount)
	_, err = updateAccountInTreeDB(exitTree.DB(), exitTree, tx.FromIdx, exitAccount)
	return exitAccount, err
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
