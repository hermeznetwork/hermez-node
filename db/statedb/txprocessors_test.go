package statedb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessTxsSynchronizer(t *testing.T) {
	// TODO once TTGL is updated, use the blockchain L2Tx (not PoolL2Tx) for
	// the Synchronizer tests

	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	assert.Equal(t, 31, len(blocks[0].L1UserTxs))
	assert.Equal(t, 4, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 22, len(blocks[0].Batches[2].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 59, len(blocks[1].Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[1].Batches[1].L2Txs))

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	log.Debug("1st batch, 1st block, only L1CoordinatorTxs")
	_, _, createdAccounts, err := sdb.ProcessTxs(nil, nil, blocks[0].Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)
	assert.Equal(t, 4, len(createdAccounts))

	log.Debug("2nd batch, 1st block")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	_, exitInfos, createdAccounts, err := sdb.ProcessTxs(coordIdxs, blocks[0].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(exitInfos))
	assert.Equal(t, 31, len(createdAccounts))
	acc, err := sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("3rd batch, 1st block")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	_, exitInfos, createdAccounts, err = sdb.ProcessTxs(coordIdxs, nil, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	// TODO once TTGL is updated, add a check that a input poolL2Tx with
	// Nonce & TokenID =0, after ProcessTxs call has the expected value

	assert.Equal(t, 0, len(exitInfos))
	assert.Equal(t, 0, len(createdAccounts))
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	log.Debug("1st batch, 2nd block")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[0].L2Txs)
	_, exitInfos, createdAccounts, err = sdb.ProcessTxs(coordIdxs, nil, blocks[1].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 4, len(exitInfos)) // the 'ForceExit(1)' is not computed yet, as the batch is without L1UserTxs
	assert.Equal(t, 1, len(createdAccounts))
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "53", acc.Balance.String())

	log.Debug("2nd batch, 2nd block")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[1].L2Txs)
	_, exitInfos, createdAccounts, err = sdb.ProcessTxs(coordIdxs, blocks[1].L1UserTxs, blocks[1].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	assert.Equal(t, 2, len(exitInfos)) // 2, as previous batch was without L1UserTxs, and has pending the 'ForceExit(1) A: 5'
	assert.Equal(t, 1, len(createdAccounts))
	acc, err = sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "78", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = sdb.GetAccount(idxB0)
	require.Nil(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "2", acc.Balance.String())
}

/*
WIP

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := til.NewContext()
	blocks := tc.GenerateBlocks(til.SetBlockchain0)

	assert.Equal(t, 29, len(blocks[0].Batches[0].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 21, len(blocks[0].Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1UserTxs))
	assert.Equal(t, 1, len(blocks[0].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 59, len(blocks[0].Batches[1].L2Txs))
	assert.Equal(t, 9, len(blocks[0].Batches[2].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[2].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[0].Batches[2].L2Txs))

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	// use first batch
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs)
	_, exitInfos, err := sdb.ProcessTxs(coordIdxs, blocks[0].Batches[0].L1UserTxs, blocks[0].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(exitInfos))
	acc, err := sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	// use second batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(coordIdxs, blocks[0].Batches[1].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 5, len(exitInfos))
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "48", acc.Balance.String())

	// use third batch
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	_, exitInfos, err = sdb.ProcessTxs(coordIdxs, blocks[0].Batches[2].L1UserTxs, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 1, len(exitInfos))
	acc, err = sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "23", acc.Balance.String())
}

func TestZKInputsGeneration(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	tc := til.NewContext()
	blocks := tc.GenerateBlocks(til.SetBlockchain0)
	assert.Equal(t, 29, len(blocks[0].Batches[0].L1UserTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 21, len(blocks[0].Batches[0].L2Txs))

	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs)
	zki, _, err := sdb.ProcessTxs(coordIdxs, blocks[0].Batches[0].L1UserTxs, blocks[0].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	s, err := json.Marshal(zki)
	require.Nil(t, err)
	debug:=true
	if debug {
		fmt.Println(string(s))
	}
}
*/
