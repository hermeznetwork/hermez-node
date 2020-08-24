package statedb

import (
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree/db"
)

// keyidx is used as key in the db to store the current Idx
var keyidx = []byte("idx")

// ProcessTxs process the given L1Txs & L2Txs applying the needed updates
// to the StateDB depending on the transaction Type. Returns the
// common.ZKInputs to generate the SnarkProof later used by the BatchBuilder,
// and returns common.ExitTreeLeaf that is later used by the Synchronizer to
// update the HistoryDB.
func (s *StateDB) ProcessTxs(l1usertxs, l1coordinatortxs []*common.L1Tx, l2txs []*common.L2Tx) (*common.ZKInputs, []*common.ExitTreeLeaf, error) {
	for _, tx := range l1usertxs {
		err := s.processL1Tx(tx)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, tx := range l1coordinatortxs {
		err := s.processL1Tx(tx)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, tx := range l2txs {
		err := s.processL2Tx(tx)
		if err != nil {
			return nil, nil, err
		}
	}

	return nil, nil, nil
}

// processL2Tx process the given L2Tx applying the needed updates to
// the StateDB depending on the transaction Type.
func (s *StateDB) processL2Tx(tx *common.L2Tx) error {
	switch tx.Type {
	case common.TxTypeTransfer:
		// go to the MT account of sender and receiver, and update
		// balance & nonce
		err := s.applyTransfer(tx.Tx())
		if err != nil {
			return err
		}
	case common.TxTypeExit:
		// execute exit flow
	default:
	}
	return nil
}

// processL1Tx process the given L1Tx applying the needed updates to the
// StateDB depending on the transaction Type.
func (s *StateDB) processL1Tx(tx *common.L1Tx) error {
	switch tx.Type {
	case common.TxTypeForceTransfer, common.TxTypeTransfer:
		// go to the MT account of sender and receiver, and update balance
		// & nonce
		err := s.applyTransfer(tx.Tx())
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new account to the MT, update balance of the MT account
		err := s.applyCreateAccount(tx)
		if err != nil {
			return err
		}
	case common.TxTypeDeposit:
		// update balance of the MT account
		err := s.applyDeposit(tx, false)
		if err != nil {
			return err
		}
	case common.TxTypeDepositTransfer:
		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := s.applyDeposit(tx, true)
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDepositTransfer:
		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := s.applyCreateAccount(tx)
		if err != nil {
			return err
		}
		err = s.applyTransfer(tx.Tx())
		if err != nil {
			return err
		}
	case common.TxTypeExit:
		// execute exit flow
	default:
	}

	return nil
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
		// substract amount to the sender
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

	// substract amount to the sender
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
	tx.Put(keyidx, idx.Bytes())
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
