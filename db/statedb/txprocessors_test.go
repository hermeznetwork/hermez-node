package statedb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkBalance(t *testing.T, tc *til.Context, sdb *StateDB, username string, tokenid int, expected string) {
	idx := tc.Users[username].Accounts[common.TokenID(tokenid)].Idx
	acc, err := sdb.GetAccount(idx)
	require.Nil(t, err)
	assert.Equal(t, expected, acc.Balance.String())
}

func TestProcessTxsBalances(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchainMinimumFlow0)
	require.Nil(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257}
	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
	}

	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	_, err = sdb.ProcessTxs(ptc, nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)

	log.Debug("block:0 batch:1")
	l1UserTxs := []common.L1Tx{}
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	log.Debug("block:0 batch:2")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[2].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")

	log.Debug("block:0 batch:3")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[3].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[3].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[3].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")
	checkBalance(t, tc, sdb, "A", 1, "500")

	log.Debug("block:0 batch:4")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[4].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[4].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[4].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "A", 0, "500")
	checkBalance(t, tc, sdb, "A", 1, "500")

	log.Debug("block:0 batch:5")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[5].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[5].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[5].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "A", 0, "600")
	checkBalance(t, tc, sdb, "A", 1, "500")
	checkBalance(t, tc, sdb, "B", 0, "400")

	log.Debug("block:0 batch:6")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[6].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[6].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[6].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "10")
	checkBalance(t, tc, sdb, "Coord", 1, "20")
	checkBalance(t, tc, sdb, "A", 0, "600")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "290")
	checkBalance(t, tc, sdb, "B", 1, "200")
	checkBalance(t, tc, sdb, "C", 0, "100")
	checkBalance(t, tc, sdb, "D", 0, "800")

	log.Debug("block:0 batch:7")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[0].Rollup.Batches[7].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[7].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[0].Rollup.Batches[7].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "35")
	checkBalance(t, tc, sdb, "Coord", 1, "30")
	checkBalance(t, tc, sdb, "A", 0, "430")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "390")
	checkBalance(t, tc, sdb, "B", 1, "90")
	checkBalance(t, tc, sdb, "C", 0, "45")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "800")

	log.Debug("block:1 batch:0")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[0].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "75")
	checkBalance(t, tc, sdb, "Coord", 1, "30")
	checkBalance(t, tc, sdb, "A", 0, "730")
	checkBalance(t, tc, sdb, "A", 1, "280")
	checkBalance(t, tc, sdb, "B", 0, "380")
	checkBalance(t, tc, sdb, "B", 1, "90")
	checkBalance(t, tc, sdb, "C", 0, "845")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "470")

	log.Debug("block:1 batch:1")
	l1UserTxs = til.L1TxsToCommonL1Txs(tc.Queues[*blocks[1].Rollup.Batches[1].Batch.ForgeL1TxsNum])
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, l1UserTxs, blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)

	// use Set of PoolL2 txs
	poolL2Txs, err := tc.GeneratePoolL2Txs(til.SetPoolL2MinimumFlow0)
	assert.Nil(t, err)

	_, err = sdb.ProcessTxs(ptc, coordIdxs, []common.L1Tx{}, []common.L1Tx{}, poolL2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "Coord", 0, "105")
	checkBalance(t, tc, sdb, "Coord", 1, "40")
	checkBalance(t, tc, sdb, "A", 0, "510")
	checkBalance(t, tc, sdb, "A", 1, "170")
	checkBalance(t, tc, sdb, "B", 0, "480")
	checkBalance(t, tc, sdb, "B", 1, "190")
	checkBalance(t, tc, sdb, "C", 0, "845")
	checkBalance(t, tc, sdb, "C", 1, "100")
	checkBalance(t, tc, sdb, "D", 0, "360")
	checkBalance(t, tc, sdb, "F", 0, "100")
}

func TestProcessTxsSynchronizer(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	assert.Equal(t, 31, len(blocks[0].Rollup.L1UserTxs))
	assert.Equal(t, 4, len(blocks[0].Rollup.Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 0, len(blocks[0].Rollup.Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 22, len(blocks[0].Rollup.Batches[2].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches[0].L1CoordinatorTxs))
	assert.Equal(t, 62, len(blocks[1].Rollup.Batches[0].L2Txs))
	assert.Equal(t, 1, len(blocks[1].Rollup.Batches[1].L1CoordinatorTxs))
	assert.Equal(t, 8, len(blocks[1].Rollup.Batches[1].L2Txs))

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  16,
	}

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	ptOut, err := sdb.ProcessTxs(ptc, nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)
	assert.Equal(t, 4, len(ptOut.CreatedAccounts))
	assert.Equal(t, 0, len(ptOut.CollectedFees))

	log.Debug("block:0 batch:1")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, blocks[0].Rollup.L1UserTxs,
		blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
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
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
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
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	// before processing expect l2Txs[0:2].Nonce==0
	assert.Equal(t, common.Nonce(0), l2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(0), l2Txs[2].Nonce)

	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
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
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, blocks[1].Rollup.L1UserTxs,
		blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
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
	assert.Equal(t, "77", acc.Balance.String())

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
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	require.Nil(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256, 257, 258, 259}

	// Idx of user 'A'
	idxA1 := tc.Users["A"].Accounts[common.TokenID(1)].Idx

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  32,
	}

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	log.Debug("block:0 batch:0, only L1CoordinatorTxs")
	ptOut, err := sdb.ProcessTxs(ptc, nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)
	// expect 0 at CreatedAccount, as is only computed when StateDB.Type==TypeSynchronizer
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))

	log.Debug("block:0 batch:1")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[1].L2Txs)
	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, blocks[0].Rollup.L1UserTxs, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err := sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "50", acc.Balance.String())

	log.Debug("block:0 batch:2")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err = sdb.ProcessTxs(ptc, coordIdxs, nil, blocks[0].Rollup.Batches[2].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, 0, len(ptOut.ExitInfos))
	assert.Equal(t, 0, len(ptOut.CreatedAccounts))
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "35", acc.Balance.String())

	log.Debug("block:1 batch:0")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[0].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, nil, blocks[1].Rollup.Batches[0].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	acc, err = sdb.GetAccount(idxA1)
	require.Nil(t, err)
	assert.Equal(t, "57", acc.Balance.String())

	log.Debug("block:1 batch:1")
	l2Txs = common.L2TxsToPoolL2Txs(blocks[1].Rollup.Batches[1].L2Txs)
	_, err = sdb.ProcessTxs(ptc, coordIdxs, blocks[1].Rollup.L1UserTxs, blocks[1].Rollup.Batches[1].L1CoordinatorTxs, l2Txs)
	require.Nil(t, err)
	acc, err = sdb.GetAccount(idxA1)
	assert.Nil(t, err)
	assert.Equal(t, "77", acc.Balance.String())

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

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 4)
	assert.Nil(t, err)

	set := `
		Type: Blockchain
		AddToken(1)
		CreateAccountDeposit(1) A: 10
		> batchL1
		CreateAccountCoordinator(1) B
		CreateAccountCoordinator(1) C
		> batchL1
		// idxs: A:258, B:256, C:257

		Transfer(1) A-B: 6 (1)
		Transfer(1) A-C: 2 (1)
		> batch
		> block
	`
	// generate test transactions from test.SetBlockchain0 code
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	require.Nil(t, err)

	// Coordinator Idx where to send the fees
	coordIdxs := []common.Idx{256}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}

	log.Debug("block:0 batch:0, only L1UserTx")
	_, err = sdb.ProcessTxs(ptc, nil, blocks[0].Rollup.L1UserTxs, nil, nil)
	require.Nil(t, err)

	log.Debug("block:0 batch:1, only L1CoordinatorTxs")
	_, err = sdb.ProcessTxs(ptc, nil, nil, blocks[0].Rollup.Batches[1].L1CoordinatorTxs, nil)
	require.Nil(t, err)

	log.Debug("block:0 batch:2, only L2Txs")
	l2Txs := common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[2].L2Txs)
	ptOut, err := sdb.ProcessTxs(ptc, coordIdxs, nil, nil, l2Txs)
	require.Nil(t, err)
	checkBalance(t, tc, sdb, "A", 1, "2")

	s, err := json.Marshal(ptOut.ZKInputs)
	require.Nil(t, err)
	debug := false
	debug = true
	if debug {
		fmt.Println(string(s))
	}
}

func TestProcessTxsRootTestVectors(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// same values than in the js test
	bjj0, err := common.BJJFromStringWithChecksum("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
	assert.Nil(t, err)
	l1Txs := []common.L1Tx{
		{
			FromIdx:     0,
			LoadAmount:  big.NewInt(16000000),
			Amount:      big.NewInt(0),
			TokenID:     1,
			FromBJJ:     bjj0,
			FromEthAddr: ethCommon.HexToAddress("0x7e5f4552091a69125d5dfcb7b8c2659029395bdf"),
			ToIdx:       0,
			Type:        common.TxTypeCreateAccountDeposit,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}
	_, err = sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)
	assert.Equal(t, "9827704113668630072730115158977131501210702363656902211840117643154933433410", sdb.mt.Root().BigInt().String())
}

func TestZKInputsHashTestVector0(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// same values than in the js test
	bjj0, err := common.BJJFromStringWithChecksum("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
	assert.Nil(t, err)
	l1Txs := []common.L1Tx{
		{
			FromIdx: 0,
			// LoadAmount:  big.NewInt(10400),
			LoadAmount:  big.NewInt(16000000),
			Amount:      big.NewInt(0),
			TokenID:     1,
			FromBJJ:     bjj0,
			FromEthAddr: ethCommon.HexToAddress("0x7e5f4552091a69125d5dfcb7b8c2659029395bdf"),
			ToIdx:       0,
			Type:        common.TxTypeCreateAccountDeposit,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 256,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     126,
			Type:    common.TxTypeTransfer,
		},
	}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}
	ptOut, err := sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "d746824f7d0ac5044a573f51b278acb56d823bec39551d1d7bf7378b68a1b021", acc.PublicKey.Compress().String())
	assert.Equal(t, "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf", acc.EthAddr.Hex())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(257))
	require.NotNil(t, err)
	ptOut.ZKInputs.FeeIdxs[0] = common.Idx(256).BigInt()

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	assert.Nil(t, err)
	// value from js test vector
	expectedToHash := "0000000000ff000000000100000000000000000000000000000000000000000000000000000000000000000015ba488d749f6b891d29d0bf3a72481ec812e4d4ecef2bf7a3fc64f3c010444200000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000010003e87e00000100000000000000000000000000000000000000000000000000000000000000"
	// checks are splitted to find the difference easier
	assert.Equal(t, expectedToHash[:1000], hex.EncodeToString(toHash)[:1000])
	assert.Equal(t, expectedToHash[1000:2000], hex.EncodeToString(toHash)[1000:2000])
	assert.Equal(t, expectedToHash[2000:], hex.EncodeToString(toHash)[2000:])

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	// value from js test vector
	assert.Equal(t, "80757288244566854497474223360206077562032050734432637237701187686677568506", h.String())
}

func TestZKInputsHashTestVector1(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// same values than in the js test
	bjj0, err := common.BJJFromStringWithChecksum("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7")
	assert.Nil(t, err)
	bjj1, err := common.BJJFromStringWithChecksum("093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d")
	assert.Nil(t, err)
	l1Txs := []common.L1Tx{
		{
			FromIdx: 0,
			// LoadAmount:  big.NewInt(10400),
			LoadAmount:  big.NewInt(16000000),
			Amount:      big.NewInt(0),
			TokenID:     1,
			FromBJJ:     bjj0,
			FromEthAddr: ethCommon.HexToAddress("0x7e5f4552091a69125d5dfcb7b8c2659029395bdf"),
			ToIdx:       0,
			Type:        common.TxTypeCreateAccountDeposit,
		},
		{
			FromIdx:     0,
			LoadAmount:  big.NewInt(16000000),
			Amount:      big.NewInt(0),
			TokenID:     1,
			FromBJJ:     bjj1,
			FromEthAddr: ethCommon.HexToAddress("0x2b5ad5c4795c026514f8317c7a215e218dccd6cf"),
			ToIdx:       0,
			Type:        common.TxTypeCreateAccountDeposit,
		},
	}
	l2Txs := []common.PoolL2Tx{
		{
			FromIdx: 257,
			ToIdx:   256,
			TokenID: 1,
			Amount:  big.NewInt(1000),
			Nonce:   0,
			Fee:     137,
			Type:    common.TxTypeTransfer,
		},
	}

	ptc := ProcessTxsConfig{
		NLevels:  32,
		MaxFeeTx: 8,
		MaxTx:    32,
		MaxL1Tx:  16,
	}
	ptOut, err := sdb.ProcessTxs(ptc, nil, l1Txs, nil, l2Txs)
	require.Nil(t, err)

	// check expected account keys values from tx inputs
	acc, err := sdb.GetAccount(common.Idx(256))
	require.Nil(t, err)
	assert.Equal(t, "d746824f7d0ac5044a573f51b278acb56d823bec39551d1d7bf7378b68a1b021", acc.PublicKey.Compress().String())
	assert.Equal(t, "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf", acc.EthAddr.Hex())
	acc, err = sdb.GetAccount(common.Idx(257))
	require.Nil(t, err)
	assert.Equal(t, "4d05c307400c65795f02db96b1b81c60386fd53e947d9d3f749f3d99b1853909", acc.PublicKey.Compress().String())
	assert.Equal(t, "0x2B5AD5c4795c026514f8317c7a215E218DcCD6cF", acc.EthAddr.Hex())

	// check that there no exist more accounts
	_, err = sdb.GetAccount(common.Idx(258))
	require.NotNil(t, err)
	ptOut.ZKInputs.FeeIdxs[0] = common.Idx(257).BigInt()

	toHash, err := ptOut.ZKInputs.ToHashGlobalData()
	assert.Nil(t, err)
	// value from js test vector
	expectedToHash := "0000000000ff0000000001010000000000000000000000000000000000000000000000000000000000000000304a3f3aef4f416cca887aab7265227449077627138345c2eb25bf8ff946b09500000000000000000000000000000000000000000000000000000000000000007e5f4552091a69125d5dfcb7b8c2659029395bdf21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d700000000000028a00000000000010000000000002b5ad5c4795c026514f8317c7a215e218dccd6cf093985b1993d9f743f9d7d943ed56f38601cb8b196db025f79650c4007c3054d00000000000028a00000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001010000010003e88900000101000000000000000000000000000000000000000000000000000000000000"
	// checks are splitted to find the difference easier
	assert.Equal(t, expectedToHash[:1000], hex.EncodeToString(toHash)[:1000])
	assert.Equal(t, expectedToHash[1000:2000], hex.EncodeToString(toHash)[1000:2000])
	assert.Equal(t, expectedToHash[2000:], hex.EncodeToString(toHash)[2000:])

	h, err := ptOut.ZKInputs.HashGlobalData()
	require.Nil(t, err)
	// value from js test vector
	assert.Equal(t, "10900521462378877053056992084240080637954406133884857263674494661625916419481", h.String())
}
