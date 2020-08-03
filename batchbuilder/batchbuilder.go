package batchbuilder

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/memory"
)

type ConfigCircuit struct {
	TxsMax       uint64
	L1TxsMax     uint64
	SMTLevelsMax uint64
}

type BatchBuilder struct {
	StateDB        db.Storage // where the MTs will be stored by the Synchronizer
	idx            uint64
	mt             *merkletree.MerkleTree
	configCircuits []ConfigCircuit
}

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

	bb.Reset(batchNum, idx, true)

	return &bb, nil
}

// Reset tells the BatchBuilder to reset it's internal state to the required
// `batchNum`.  If `fromSynchronizer` is true, the BatchBuilder must take a
// copy of the rollup state from the Synchronizer at that `batchNum`, otherwise
// it can just roll back the internal copy.
func (bb *BatchBuilder) Reset(batchNum int, idx uint64, fromSynchronizer bool) error {

	return nil
}

func (bb *BatchBuilder) BuildBatch(configBatch ConfigBatch, l1usertxs, l1coordinatortxs []common.L1Tx, l2txs []common.L2Tx, tokenIDs []common.TokenID) (*common.ZKInputs, error) {

	for _, tx := range l1usertxs {
		bb.processL1Tx(tx)
	}
	for _, tx := range l1coordinatortxs {
		bb.processL1Tx(tx)
	}
	for _, tx := range l2txs {
		switch tx.Type {
		case common.TxTypeTransfer:
			// go to the MT leaf of sender and receiver, and update
			// balance & nonce
			bb.applyTransfer(tx.Tx)
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
		bb.applyTransfer(tx.Tx)
	case common.TxTypeCreateAccountDeposit:
		// add new leaf to the MT, update balance of the MT leaf
		bb.applyCreateLeaf(tx)
	case common.TxTypeDeposit:
		// update balance of the MT leaf
		bb.applyDeposit(tx)
	case common.TxTypeDepositAndTransfer:
		// update balance in MT leaf, update balance & nonce of sender
		// & receiver
		bb.applyDeposit(tx) // this after v0, can be done by bb.applyDepositAndTransfer in a single step
		bb.applyTransfer(tx.Tx)
	case common.TxTypeCreateAccountDepositAndTransfer:
		// add new leaf to the merkletree, update balance in MT leaf,
		// update balance & nonce of sender & receiver
		bb.applyCreateLeaf(tx)
		bb.applyTransfer(tx.Tx)
	case common.TxTypeExit:
		// execute exit flow
	default:
	}

	return nil
}

// applyCreateLeaf creates a new leaf in the leaf of the depositer, it stores
// the deposit value
func (bb *BatchBuilder) applyCreateLeaf(tx common.L1Tx) error {
	k := big.NewInt(int64(bb.idx + 1))

	leaf := common.Leaf{
		TokenID: tx.TokenID,
		Nonce:   0, // TODO check always that a new leaf is created nonce is at 0
		Balance: tx.LoadAmount,
		Ax:      tx.FromBJJ.X,
		Ay:      tx.FromBJJ.Y,
		EthAddr: tx.FromEthAddr,
	}

	v, err := leaf.Value()
	if err != nil {
		return err
	}

	// store at the DB the key: v, and value: leaf.Bytes()
	dbTx, err := bb.mt.DB().NewTx()
	if err != nil {
		return err
	}
	leafBytes := leaf.Bytes()
	dbTx.Put(v.Bytes(), leafBytes[:])

	// Add k & v into the MT
	err = bb.mt.Add(k, v)
	if err != nil {
		return err
	}

	// if everything is fine, increment idx
	bb.idx = bb.idx + 1
	return nil
}

// applyDeposit updates the balance in the leaf of the depositer
func (bb *BatchBuilder) applyDeposit(tx common.L1Tx) error {

	return nil
}

// applyTransfer updates the balance & nonce in the leaf of the sender, and the
// balance in the leaf of the receiver
func (bb *BatchBuilder) applyTransfer(tx common.Tx) error {

	return nil
}
