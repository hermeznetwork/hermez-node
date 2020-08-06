package batchbuilder

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/memory"
)

// ConfigCircuit contains the circuit configuration
type ConfigCircuit struct {
	TxsMax       uint64
	L1TxsMax     uint64
	SMTLevelsMax uint64
}

// BatchBuilder implements the batch builder type, which contains the functionallities
type BatchBuilder struct {
	StateDB        db.Storage // where the MTs will be stored by the Synchronizer
	idx            uint64
	mt             *merkletree.MerkleTree
	configCircuits []ConfigCircuit
}

// ConfigBatch contains the batch configuration
type ConfigBatch struct {
	CoordinatorAddress ethCommon.Address
}

// NewBatchBuilder constructs a new BatchBuilder, and executes the bb.Reset
// method
func NewBatchBuilder(stateDB db.Storage, configCircuits []ConfigCircuit, batchNum int, idx, nLevels uint64) (*BatchBuilder, error) {
	localMt, err := merkletree.NewMerkleTree(memory.NewMemoryStorage(), int(nLevels))
	if err != nil {
		return nil, err
	}
	bb := BatchBuilder{
		StateDB:        stateDB,
		idx:            idx,
		mt:             localMt,
		configCircuits: configCircuits,
	}

	err = bb.Reset(batchNum, idx, true)
	return &bb, err
}

// Reset tells the BatchBuilder to reset it's internal state to the required
// `batchNum`.  If `fromSynchronizer` is true, the BatchBuilder must take a
// copy of the rollup state from the Synchronizer at that `batchNum`, otherwise
// it can just roll back the internal copy.
func (bb *BatchBuilder) Reset(batchNum int, idx uint64, fromSynchronizer bool) error {
	// TODO
	return nil
}

// BuildBatch takes the transactions and returns the common.ZKInputs of the next batch
func (bb *BatchBuilder) BuildBatch(configBatch ConfigBatch, l1usertxs, l1coordinatortxs []common.L1Tx, l2txs []common.L2Tx, tokenIDs []common.TokenID) (*common.ZKInputs, error) {
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
			// go to the MT leaf of sender and receiver, and update
			// balance & nonce
			err := bb.applyTransfer(tx.Tx)
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
		// go to the MT leaf of sender and receiver, and update balance
		// & nonce
		err := bb.applyTransfer(tx.Tx)
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDeposit:
		// add new leaf to the MT, update balance of the MT leaf
		err := bb.applyCreateLeaf(tx)
		if err != nil {
			return err
		}
	case common.TxTypeDeposit:
		// update balance of the MT leaf
		err := bb.applyDeposit(tx, false)
		if err != nil {
			return err
		}
	case common.TxTypeDepositAndTransfer:
		// update balance in MT leaf, update balance & nonce of sender
		// & receiver
		err := bb.applyDeposit(tx, true)
		if err != nil {
			return err
		}
	case common.TxTypeCreateAccountDepositAndTransfer:
		// add new leaf to the merkletree, update balance in MT leaf,
		// update balance & nonce of sender & receiver
		err := bb.applyCreateLeaf(tx)
		if err != nil {
			return err
		}
		err = bb.applyTransfer(tx.Tx)
		if err != nil {
			return err
		}
	case common.TxTypeExit:
		// execute exit flow
	default:
	}

	return nil
}

// applyCreateLeaf creates a new leaf in the leaf of the depositer, it stores
// the deposit value
func (bb *BatchBuilder) applyCreateLeaf(tx common.L1Tx) error {
	leaf := common.Leaf{
		TokenID: tx.TokenID,
		Nonce:   0, // TODO check w spec: always that a new leaf is created nonce is at 0
		Balance: tx.LoadAmount,
		Sign:    babyjub.PointCoordSign(tx.FromBJJ.X),
		Ay:      tx.FromBJJ.Y,
		EthAddr: tx.FromEthAddr,
	}

	v, err := leaf.HashValue()
	if err != nil {
		return err
	}
	dbTx, err := bb.mt.DB().NewTx()
	if err != nil {
		return err
	}

	err = bb.CreateBalance(dbTx, common.Idx(bb.idx+1), leaf)
	if err != nil {
		return err
	}
	leafBytes, err := leaf.Bytes()
	if err != nil {
		return err
	}
	dbTx.Put(v.Bytes(), leafBytes[:])

	// if everything is fine, do dbTx & increment idx
	if err := dbTx.Commit(); err != nil {
		return err
	}
	bb.idx = bb.idx + 1
	return nil
}

// applyDeposit updates the balance in the leaf of the depositer, if andTransfer parameter is set to true, the method will also apply the Transfer of the L1Tx/DepositAndTransfer
func (bb *BatchBuilder) applyDeposit(tx common.L1Tx, andTransfer bool) error {
	dbTx, err := bb.mt.DB().NewTx()
	if err != nil {
		return err
	}

	// deposit
	err = bb.UpdateBalance(dbTx, tx.FromIdx, tx.LoadAmount, false)
	if err != nil {
		return err
	}

	// in case that the tx is a L1Tx>DepositAndTransfer
	if andTransfer {
		// transact
		err = bb.UpdateBalance(dbTx, tx.FromIdx, tx.Tx.Amount, true)
		if err != nil {
			return err
		}
		err = bb.UpdateBalance(dbTx, tx.ToIdx, tx.Tx.Amount, false)
		if err != nil {
			return err
		}
	}

	if err := dbTx.Commit(); err != nil {
		return err
	}
	return nil
}

// applyTransfer updates the balance & nonce in the leaf of the sender, and the
// balance in the leaf of the receiver
func (bb *BatchBuilder) applyTransfer(tx common.Tx) error {
	dbTx, err := bb.mt.DB().NewTx()
	if err != nil {
		return err
	}

	// transact
	err = bb.UpdateBalance(dbTx, tx.FromIdx, tx.Amount, true)
	if err != nil {
		return err
	}
	err = bb.UpdateBalance(dbTx, tx.ToIdx, tx.Amount, false)
	if err != nil {
		return err
	}

	if err := dbTx.Commit(); err != nil {
		return err
	}
	return nil
}
