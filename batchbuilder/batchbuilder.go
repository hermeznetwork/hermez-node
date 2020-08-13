package batchbuilder

import (
	"encoding/binary"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/iden3/go-merkletree/db"
)

// KEYIDX is used as key in the db to store the current Idx
var KEYIDX = []byte("idx")

// ConfigCircuit contains the circuit configuration
type ConfigCircuit struct {
	TxsMax       uint64
	L1TxsMax     uint64
	SMTLevelsMax uint64
}

// BatchBuilder implements the batch builder type, which contains the
// functionalities
type BatchBuilder struct {
	// idx holds the current Idx that the BatchBuilder is using
	idx            uint64
	localStateDB   *statedb.LocalStateDB
	configCircuits []ConfigCircuit
}

// ConfigBatch contains the batch configuration
type ConfigBatch struct {
	ForgerAddress ethCommon.Address
}

// NewBatchBuilder constructs a new BatchBuilder, and executes the bb.Reset
// method
func NewBatchBuilder(dbpath string, synchronizerStateDB *statedb.StateDB, configCircuits []ConfigCircuit, batchNum uint64, nLevels uint64) (*BatchBuilder, error) {
	localStateDB, err := statedb.NewLocalStateDB(dbpath, synchronizerStateDB, true, int(nLevels))
	if err != nil {
		return nil, err
	}

	bb := BatchBuilder{
		localStateDB:   localStateDB,
		configCircuits: configCircuits,
	}

	err = bb.Reset(batchNum, true)
	return &bb, err
}

// Reset tells the BatchBuilder to reset it's internal state to the required
// `batchNum`.  If `fromSynchronizer` is true, the BatchBuilder must take a
// copy of the rollup state from the Synchronizer at that `batchNum`, otherwise
// it can just roll back the internal copy.
func (bb *BatchBuilder) Reset(batchNum uint64, fromSynchronizer bool) error {
	if batchNum == 0 {
		bb.idx = 0
		return nil
	}
	err := bb.localStateDB.Reset(batchNum, fromSynchronizer)
	if err != nil {
		return err
	}
	// idx is obtained from the statedb reset
	bb.idx, err = bb.getIdx()
	return err
}

// BuildBatch takes the transactions and returns the common.ZKInputs of the next batch
func (bb *BatchBuilder) BuildBatch(configBatch ConfigBatch, l1usertxs, l1coordinatortxs []common.L1Tx, l2txs []common.PoolL2Tx, tokenIDs []common.TokenID) (*common.ZKInputs, error) {
	for _, tx := range l1usertxs {
		err := bb.processL1Tx(tx)
		if err != nil {
			return nil, err
		}
	}
	for _, tx := range l1coordinatortxs {
		err := bb.processL1Tx(tx)
		if err != nil {
			return nil, err
		}
	}
	for _, tx := range l2txs {
		switch tx.Type {
		case common.TxTypeTransfer:
			// go to the MT account of sender and receiver, and update
			// balance & nonce
			err := bb.applyTransfer(tx.Tx())
			if err != nil {
				return nil, err
			}
		case common.TxTypeExit:
			// execute exit flow
		default:
		}
	}

	return nil, nil
}

func (bb *BatchBuilder) processL1Tx(tx common.L1Tx) error {
	switch tx.Type {
	case common.TxTypeForceTransfer, common.TxTypeTransfer:
		// go to the MT account of sender and receiver, and update balance
		// & nonce
		err := bb.applyTransfer(tx.Tx())
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new account to the MT, update balance of the MT account
		err := bb.applyCreateAccount(tx)
		if err != nil {
			return err
		}
	case common.TxTypeDeposit: // TODO check if this type will ever exist, or will be TxTypeDepositAndTransfer with transfer 0 value
		// update balance of the MT account
		err := bb.applyDeposit(tx, false)
		if err != nil {
			return err
		}
	case common.TxTypeDepositAndTransfer:
		// update balance in MT account, update balance & nonce of sender
		// & receiver
		err := bb.applyDeposit(tx, true)
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDepositAndTransfer:
		// add new account to the merkletree, update balance in MT account,
		// update balance & nonce of sender & receiver
		err := bb.applyCreateAccount(tx)
		if err != nil {
			return err
		}
		err = bb.applyTransfer(tx.Tx())
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
func (bb *BatchBuilder) applyCreateAccount(tx common.L1Tx) error {
	account := &common.Account{
		TokenID:   tx.TokenID,
		Nonce:     0,
		Balance:   tx.LoadAmount,
		PublicKey: &tx.FromBJJ,
		EthAddr:   tx.FromEthAddr,
	}

	err := bb.localStateDB.CreateAccount(common.Idx(bb.idx+1), account)
	if err != nil {
		return err
	}

	bb.idx = bb.idx + 1
	return bb.setIdx(bb.idx)
}

// applyDeposit updates the balance in the account of the depositer, if
// andTransfer parameter is set to true, the method will also apply the
// Transfer of the L1Tx/DepositAndTransfer
func (bb *BatchBuilder) applyDeposit(tx common.L1Tx, transfer bool) error {
	// deposit the tx.LoadAmount into the sender account
	accSender, err := bb.localStateDB.GetAccount(tx.FromIdx)
	if err != nil {
		return err
	}
	accSender.Balance = new(big.Int).Add(accSender.Balance, tx.LoadAmount)

	// in case that the tx is a L1Tx>DepositAndTransfer
	if transfer {
		accReceiver, err := bb.localStateDB.GetAccount(tx.ToIdx)
		if err != nil {
			return err
		}
		// substract amount to the sender
		accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
		// add amount to the receiver
		accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)
		// update receiver account in localStateDB
		err = bb.localStateDB.UpdateAccount(tx.ToIdx, accReceiver)
		if err != nil {
			return err
		}
	}
	// update sender account in localStateDB
	err = bb.localStateDB.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return err
	}
	return nil
}

// applyTransfer updates the balance & nonce in the account of the sender, and
// the balance in the account of the receiver
func (bb *BatchBuilder) applyTransfer(tx common.Tx) error {
	// get sender and receiver accounts from localStateDB
	accSender, err := bb.localStateDB.GetAccount(tx.FromIdx)
	if err != nil {
		return err
	}
	accReceiver, err := bb.localStateDB.GetAccount(tx.ToIdx)
	if err != nil {
		return err
	}

	// substract amount to the sender
	accSender.Balance = new(big.Int).Sub(accSender.Balance, tx.Amount)
	// add amount to the receiver
	accReceiver.Balance = new(big.Int).Add(accReceiver.Balance, tx.Amount)

	// update receiver account in localStateDB
	err = bb.localStateDB.UpdateAccount(tx.ToIdx, accReceiver)
	if err != nil {
		return err
	}
	// update sender account in localStateDB
	err = bb.localStateDB.UpdateAccount(tx.FromIdx, accSender)
	if err != nil {
		return err
	}

	return nil
}

// getIdx returns the stored Idx from the localStateDB, which is the last Idx used for an Account in the localStateDB.
func (bb *BatchBuilder) getIdx() (uint64, error) {
	idxBytes, err := bb.localStateDB.DB().Get(KEYIDX)
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
func (bb *BatchBuilder) setIdx(idx uint64) error {
	tx, err := bb.localStateDB.DB().NewTx()
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
