package statedb

import (
	"io/ioutil"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/test/transakcio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessTxsSynchronizer(t *testing.T) {
	// TODO once TTGL is updated, use the blockchain L2Tx (not PoolL2Tx) for
	// the Synchronizer tests

	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := transakcio.NewTestContext()
	blocks := tc.GenerateBlocks(transakcio.SetBlockchain0)

	assert.Equal(t, 29, len(blocks[0].Batches[0].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 21, len(blocks[0].Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1UserTxs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 59, len(blocks[0].Batches[1].L2Txs))
	assert.Equal(t, 9, len(blocks[0].Batches[2].L1UserTxs))
	assert.Equal(t, 1, len(blocks[0].Batches[2].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[0].Batches[2].L2Txs))

	// use first batch
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs)
	_, exitInfos, err := sdb.ProcessTxs(blocks[0].Batches[0].L1UserTxs, blocks[0].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	// TODO once TTGL is updated, add a check that a input poolL2Tx with
	// Nonce & TokenID =0, after ProcessTxs call has the expected value

	assert.Equal(t, 0, len(exitInfos))
	acc, err := sdb.GetAccount(common.Idx(256))
	assert.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	// use second batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(blocks[0].Batches[1].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 5, len(exitInfos))
	acc, err = sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "48", acc.Balance.String())

	// use third batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(blocks[0].Batches[2].L1UserTxs, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 1, len(exitInfos))
	acc, err = sdb.GetAccount(common.Idx(256))
	assert.Nil(t, err)
	assert.Equal(t, "73", acc.Balance.String())
}

/*
WIP

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := transakcio.NewTestContext()
	blocks := tc.GenerateBlocks(transakcio.SetBlockchain0)

	assert.Equal(t, 29, len(blocks[0].Batches[0].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 21, len(blocks[0].Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1UserTxs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 59, len(blocks[0].Batches[1].L2Txs))
	assert.Equal(t, 9, len(blocks[0].Batches[2].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[2].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[0].Batches[2].L2Txs))

	// use first batch
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs)
	_, exitInfos, err := sdb.ProcessTxs(blocks[0].Batches[0].L1UserTxs, blocks[0].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(exitInfos))
	acc, err := sdb.GetAccount(common.Idx(256))
	assert.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	// use second batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(blocks[0].Batches[1].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 5, len(exitInfos))
	acc, err = sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "48", acc.Balance.String())

	// use third batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(blocks[0].Batches[2].L1UserTxs, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 1, len(exitInfos))
	acc, err = sdb.GetAccount(common.Idx(256))
	assert.Nil(t, err)
	assert.Equal(t, "23", acc.Balance.String())
}

func TestZKInputsGeneration(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := transakcio.NewTestContext()
	blocks := tc.GenerateBlocks(transakcio.SetBlockchain0)
	assert.Equal(t, 29, len(blocks[0].Batches[0].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 21, len(blocks[0].Batches[0].L2Txs))

	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs)
	zki, _, err := sdb.ProcessTxs(blocks[0].Batches[0].L1UserTxs, blocks[0].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	s, err := json.Marshal(zki)
	require.Nil(t, err)
	debug:=true
	if debug {
		fmt.Println(string(s))
	}
}
*/
