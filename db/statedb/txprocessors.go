package statedb

import (
	"encoding/binary"
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree/db"
)

// KEYIDX is used as key in the db to store the current Idx
var KEYIDX = []byte("idx")

// ProcessPoolL2Tx process the given PoolL2Tx applying the needed updates to
// the StateDB depending on the transaction Type.
func (s *StateDB) ProcessPoolL2Tx(tx *common.PoolL2Tx) error {
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

// ProcessL1Tx process the given L1Tx applying the needed updates to the
// StateDB depending on the transaction Type.
func (s *StateDB) ProcessL1Tx(tx *common.L1Tx) error {
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
func (s *StateDB) getIdx() (uint64, error) {
	idxBytes, err := s.DB().Get(KEYIDX)
	if err == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	idx := binary.LittleEndian.Uint64(idxBytes[:8])
	return idx, nil
}

// setIdx stores Idx in the localStateDB
func (s *StateDB) setIdx(idx uint64) error {
	tx, err := s.DB().NewTx()
	if err != nil {
		return err
	}
	var idxBytes [8]byte
	binary.LittleEndian.PutUint64(idxBytes[:], idx)
	tx.Put(KEYIDX, idxBytes[:])
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
