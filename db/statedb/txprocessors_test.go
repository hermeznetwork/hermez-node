package statedb

import (
	"encoding/json"
	"fmt"
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
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	assert.Equal(t, 31, len(blocks[0].L1UserTxs))
	assert.Equal(t, 4, len(blocks[0].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 0, len(blocks[0].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 22, len(blocks[0].Batches[2].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 62, len(blocks[1].Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[1].Batches[1].L2Txs))

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	ptOut, err := sdb.ProcessTxs(nil, nil, blocks[0].Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)
	assert.Equal(t, 4, len(ptOut.CreatedAccounts))
	assert.Equal(t, 0, len(ptOut.CollectedFees))

	log.Debug("block:0 batch:1")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(coordIdxs, blocks[0].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 31, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err := sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	ptOut, err = sdb.ProcessTxs(coordIdxs, nil, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "2", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:0")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[0].L2Txs)
	// before processing expect l2Txs[0:2].Nonce==0
	assert.Equal(t, common.Nonce(0), l2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[2].Nonce)

	ptOut, err = sdb.ProcessTxs(coordIdxs, nil, blocks[1].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	// after processing expect l2Txs[0:2].Nonce!=0 and has expected value
	assert.Equal(t, common.Nonce(6), l2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(7), l2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(8), l2Txs[2].Nonce)

	assert.Equal(t, 4, len(ptOut.ExitInfos)) // the 'ForceExit(1)' is not computed yet, as the batch is without L1UserTxs
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "1", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(coordIdxs, blocks[1].L1UserTxs, blocks[1].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	assert.Equal(t, 2, len(ptOut.ExitInfos)) // 2, as previous batch was without L1UserTxs, and has pending the 'ForceExit(1) A: 5'
	assert.Equal(t, 1, len(ptOut.CreatedAccounts))
	assert.Equal(t, 4, len(ptOut.CollectedFees))
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(0)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(1)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(2)].String())
	assert.Equal(t, "0", ptOut.CollectedFees[common.TokenID(3)].String())
	acc, err = sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "82", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = sdb.GetAccount(idxB0)
	require.Nil(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "2", acc.Balance.String())
}

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	ptOut, err := sdb.ProcessTxs(nil, nil, blocks[0].Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)
	// expect 0 at CreatedAccount, as is only computed when StateDB.Type==TypeSynchronizer
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))

	log.Debug("block:0 batch:1")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(coordIdxs, blocks[0].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err := sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs)
	ptOut, err = sdb.ProcessTxs(coordIdxs, nil, blocks[0].Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:0")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[0].L2Txs)
	_, err = sdb.ProcessTxs(coordIdxs, nil, blocks[1].Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Batches[1].L2Txs)
	_, err = sdb.ProcessTxs(coordIdxs, blocks[1].L1UserTxs, blocks[1].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	acc, err = sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "82", acc.Balance.String())

	idxB0 := tc.Users["C"].Accounts[common.TokenID(0)].Idx
	acc, err = sdb.GetAccount(idxB0)
	require.Nil(t, err)
	assert.Equal(t, "51", acc.Balance.String())

	// get balance of Coordinator account for TokenID==0
	acc, err = sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, common.TokenID(0), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.Nil(t, err)
	assert.Equal(t, common.TokenID(1), acc.TokenID)
	assert.Equal(t, "2", acc.Balance.String())
}

func TestZKInputsGeneration(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	_, err = sdb.ProcessTxs(nil, nil, blocks[0].Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)

	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Batches[1].L2Txs)
	ptOut, err := sdb.ProcessTxs(coordIdxs, blocks[0].L1UserTxs, blocks[0].Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	s, err := json.Marshal(ptOut.ZKInputs)
	require.Nil(t, err)
	debug := false
	if debug {
		fmt.Println(string(s))
	}
}
