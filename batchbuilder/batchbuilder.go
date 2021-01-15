package batchbuilder

import (
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
)

// ConfigCircuit contains the circuit configuration
type ConfigCircuit struct {
	TxsMax       uint64
	L1TxsMax     uint64
	SMTLevelsMax uint64
}

// BatchBuilder implements the batch builder type, which contains the
// functionalities
type BatchBuilder struct {
	localStateDB *statedb.LocalStateDB
}

// ConfigBatch contains the batch configuration
type ConfigBatch struct {
	TxProcessorConfig txprocessor.Config
}

// NewBatchBuilder constructs a new BatchBuilder, and executes the bb.Reset
// method
func NewBatchBuilder(dbpath string, synchronizerStateDB *statedb.StateDB, batchNum common.BatchNum, nLevels uint64) (*BatchBuilder, error) {
	localStateDB, err := statedb.NewLocalStateDB(dbpath, 128, synchronizerStateDB,
		statedb.TypeBatchBuilder, int(nLevels))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	bb := BatchBuilder{
		localStateDB: localStateDB,
	}

	err = bb.Reset(batchNum, true)
	return &bb, tracerr.Wrap(err)
}

// Reset tells the BatchBuilder to reset it's internal state to the required
// `batchNum`.  If `fromSynchronizer` is true, the BatchBuilder must take a
// copy of the rollup state from the Synchronizer at that `batchNum`, otherwise
// it can just roll back the internal copy.
func (bb *BatchBuilder) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	return bb.localStateDB.Reset(batchNum, fromSynchronizer)
}

// BuildBatch takes the transactions and returns the common.ZKInputs of the next batch
func (bb *BatchBuilder) BuildBatch(coordIdxs []common.Idx, configBatch *ConfigBatch, l1usertxs,
	l1coordinatortxs []common.L1Tx, pooll2txs []common.PoolL2Tx) (*common.ZKInputs, error) {
	bbStateDB := bb.localStateDB.StateDB
	tp := txprocessor.NewTxProcessor(bbStateDB, configBatch.TxProcessorConfig)

	ptOut, err := tp.ProcessTxs(coordIdxs, l1usertxs, l1coordinatortxs, pooll2txs)
	return ptOut.ZKInputs, tracerr.Wrap(err)
}

// LocalStateDB returns the underlying LocalStateDB
func (bb *BatchBuilder) LocalStateDB() *statedb.LocalStateDB {
	return bb.localStateDB
}
