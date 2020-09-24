package historydb

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

var historyDB *HistoryDB

// In order to run the test you need to run a Posgres DB with
// a database named "history" that is accessible by
// user: "hermez"
// pass: set it using the env var POSTGRES_PASS
// This can be achieved by running: POSTGRES_PASS=your_strong_pass && sudo docker run --rm --name hermez-db-test -p 5432:5432 -e POSTGRES_DB=history -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD=$POSTGRES_PASS -d postgres && sleep 2s && sudo docker exec -it hermez-db-test psql -a history -U hermez -c "CREATE DATABASE l2;"
// After running the test you can stop the container by running: sudo docker kill hermez-db-test
// If you already did that for the L2DB you don't have to do it again

func TestMain(m *testing.M) {
	// init DB
	pass := "yourpasswordhere" // os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	historyDB = NewHistoryDB(db)
	if err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		fmt.Println("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func TestBlocks(t *testing.T) {
	var fromBlock, toBlock int64
	fromBlock = 1
	toBlock = 5
	// Delete peviously created rows (clean previous test execs)
	assert.NoError(t, historyDB.Reorg(fromBlock-1))
	// Generate fake blocks
	blocks := test.GenBlocks(fromBlock, toBlock)
	// Insert blocks into DB
	for i := 0; i < len(blocks); i++ {
		err := historyDB.AddBlock(&blocks[i])
		assert.NoError(t, err)
	}
	// Get all blocks from DB
	fetchedBlocks, err := historyDB.GetBlocks(fromBlock, toBlock)
	assert.Equal(t, len(blocks), len(fetchedBlocks))
	// Compare generated vs getted blocks
	assert.NoError(t, err)
	for i, fetchedBlock := range fetchedBlocks {
		assertEqualBlock(t, &blocks[i], fetchedBlock)
	}
	// Get blocks from the DB one by one
	for i := fromBlock; i < toBlock; i++ {
		fetchedBlock, err := historyDB.GetBlock(i)
		assert.NoError(t, err)
		assertEqualBlock(t, &blocks[i-1], fetchedBlock)
	}
	// Get last block
	lastBlock, err := historyDB.GetLastBlock()
	assert.NoError(t, err)
	assertEqualBlock(t, &blocks[len(blocks)-1], lastBlock)
}

func assertEqualBlock(t *testing.T, expected *common.Block, actual *common.Block) {
	assert.Equal(t, expected.EthBlockNum, actual.EthBlockNum)
	assert.Equal(t, expected.Hash, actual.Hash)
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
}

func TestBatches(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 3
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake batches
	const nBatches = 9
	batches := test.GenBatches(nBatches, blocks)
	// Test GetLastL1TxsNum with no batches
	fetchedLastL1TxsNum, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Nil(t, fetchedLastL1TxsNum)
	// Add batches to the DB
	err = historyDB.AddBatches(batches)
	assert.NoError(t, err)
	// Get batches from the DB
	fetchedBatches, err := historyDB.GetBatches(0, common.BatchNum(nBatches))
	assert.NoError(t, err)
	for i, fetchedBatch := range fetchedBatches {
		assert.Equal(t, batches[i], *fetchedBatch)
	}
	// Test GetLastBatchNum
	fetchedLastBatchNum, err := historyDB.GetLastBatchNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[len(batches)-1].BatchNum, fetchedLastBatchNum)
	// Test GetLastL1TxsNum
	fetchedLastL1TxsNum, err = historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[nBatches-1].ForgeL1TxsNum, *fetchedLastL1TxsNum)
}

func TestBids(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake coordinators
	const nCoords = 5
	coords := test.GenCoordinators(nCoords, blocks)
	err := historyDB.AddCoordinators(coords)
	assert.NoError(t, err)
	// Generate fake bids
	const nBids = 20
	bids := test.GenBids(nBids, blocks, coords)
	err = historyDB.AddBids(bids)
	assert.NoError(t, err)
	// Fetch bids
	fetchedBids, err := historyDB.GetBids()
	assert.NoError(t, err)
	// Compare fetched bids vs generated bids
	for i, bid := range fetchedBids {
		assert.Equal(t, bids[i], *bid)
	}
}

func TestTokens(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens := test.GenTokens(nTokens, blocks)
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	// Fetch tokens
	fetchedTokens, err := historyDB.GetTokens()
	assert.NoError(t, err)
	// Compare fetched tokens vs generated tokens
	// All the tokens should have USDUpdate setted by the DB trigger
	for i, token := range fetchedTokens {
		assert.Equal(t, tokens[i].TokenID, token.TokenID)
		assert.Equal(t, tokens[i].EthBlockNum, token.EthBlockNum)
		assert.Equal(t, tokens[i].EthAddr, token.EthAddr)
		assert.Equal(t, tokens[i].Name, token.Name)
		assert.Equal(t, tokens[i].Symbol, token.Symbol)
		assert.Equal(t, tokens[i].USD, token.USD)
		if token.USDUpdate != nil {
			assert.Greater(t, int64(1*time.Second), int64(time.Since(*token.USDUpdate)))
		} else {
			assert.Equal(t, tokens[i].USDUpdate, token.USDUpdate)
		}
	}
}

func TestAccounts(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens := test.GenTokens(nTokens, blocks)
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	// Generate fake batches
	const nBatches = 10
	batches := test.GenBatches(nBatches, blocks)
	err = historyDB.AddBatches(batches)
	assert.NoError(t, err)
	// Generate fake accounts
	const nAccounts = 3
	accs := test.GenAccounts(nAccounts, 0, tokens, nil, nil, batches)
	err = historyDB.AddAccounts(accs)
	assert.NoError(t, err)
	// Fetch accounts
	fetchedAccs, err := historyDB.GetAccounts()
	assert.NoError(t, err)
	// Compare fetched accounts vs generated accounts
	for i, acc := range fetchedAccs {
		assert.Equal(t, accs[i], *acc)
	}
}

func TestTxs(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens := test.GenTokens(nTokens, blocks)
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	// Generate fake batches
	const nBatches = 10
	batches := test.GenBatches(nBatches, blocks)
	err = historyDB.AddBatches(batches)
	assert.NoError(t, err)
	// Generate fake accounts
	const nAccounts = 3
	accs := test.GenAccounts(nAccounts, 0, tokens, nil, nil, batches)
	err = historyDB.AddAccounts(accs)
	assert.NoError(t, err)
	// Generate fake L1 txs
	const nL1s = 30
	_, l1txs := test.GenL1Txs(0, nL1s, 0, nil, accs, tokens, blocks, batches)
	err = historyDB.AddL1Txs(l1txs)
	fmt.Println(err)
	assert.NoError(t, err)
	// Generate fake L2 txs
	const nL2s = 20
	_, l2txs := test.GenL2Txs(0, nL2s, 0, nil, accs, tokens, blocks, batches)
	err = historyDB.AddL2Txs(l2txs)
	assert.NoError(t, err)
	// Compare fetched txs vs generated txs.
	for i := 0; i < len(l1txs); i++ {
		tx := l1txs[i].Tx()
		fetchedTx, err := historyDB.GetTx(tx.TxID)
		assert.NoError(t, err)
		test.AssertUSD(t, tx.USD, fetchedTx.USD)
		test.AssertUSD(t, tx.LoadAmountUSD, fetchedTx.LoadAmountUSD)
		assert.Equal(t, tx, fetchedTx)
	}
	for i := 0; i < len(l2txs); i++ {
		tx := l2txs[i].Tx()
		fetchedTx, err := historyDB.GetTx(tx.TxID)
		tx.TokenID = fetchedTx.TokenID
		assert.NoError(t, err)
		test.AssertUSD(t, fetchedTx.USD, tx.USD)
		test.AssertUSD(t, fetchedTx.FeeUSD, tx.FeeUSD)
		assert.Equal(t, tx, fetchedTx)
	}
	// Test trigger: L1 integrity
	// from_eth_addr can't be null
	l1txs[0].FromEthAddr = ethCommon.Address{}
	err = historyDB.AddL1Txs(l1txs)
	assert.Error(t, err)
	l1txs[0].FromEthAddr = ethCommon.BigToAddress(big.NewInt(int64(5)))
	// from_bjj can't be null
	l1txs[0].FromBJJ = nil
	err = historyDB.AddL1Txs(l1txs)
	assert.Error(t, err)
	privK := babyjub.NewRandPrivKey()
	l1txs[0].FromBJJ = privK.Public()
	// load_amount can't be null
	l1txs[0].LoadAmount = nil
	err = historyDB.AddL1Txs(l1txs)
	assert.Error(t, err)
	// Test trigger: L2 integrity
	// batch_num can't be null
	l2txs[0].BatchNum = 0
	err = historyDB.AddL2Txs(l2txs)
	assert.Error(t, err)
	l2txs[0].BatchNum = 1
	// nonce can't be null
	l2txs[0].Nonce = 0
	err = historyDB.AddL2Txs(l2txs)
	assert.Error(t, err)
	// Test helper functions for Synchronizer
	txs, err := historyDB.GetL1UserTxs(2)
	assert.NoError(t, err)
	assert.NotZero(t, len(txs))
	position, err := historyDB.GetLastTxsPosition(2)
	assert.NoError(t, err)
	assert.Equal(t, 22, position)
	// Test Update L1 TX Batch_num
	assert.Equal(t, common.BatchNum(0), txs[0].BatchNum)
	txs[0].BatchNum = common.BatchNum(1)
	// err = historyDB.UpdateTxsBatchNum(txs)
	err = historyDB.SetBatchNumL1UserTxs(2, 1)
	assert.NoError(t, err)
	txs, err = historyDB.GetL1UserTxs(2)
	assert.NoError(t, err)
	assert.NotZero(t, len(txs))
	assert.Equal(t, common.BatchNum(1), txs[0].BatchNum)
}

func TestExitTree(t *testing.T) {
	nBatches := 17
	blocks := setTestBlocks(0, 10)
	batches := test.GenBatches(nBatches, blocks)
	err := historyDB.AddBatches(batches)
	assert.NoError(t, err)

	exitTree := test.GenExitTree(nBatches)
	err = historyDB.AddExitTree(exitTree)
	assert.NoError(t, err)
}

// setTestBlocks WARNING: this will delete the blocks and recreate them
func setTestBlocks(from, to int64) []common.Block {
	if err := cleanHistoryDB(); err != nil {
		panic(err)
	}
	blocks := test.GenBlocks(from, to)
	if err := historyDB.AddBlocks(blocks); err != nil {
		panic(err)
	}
	return blocks
}

func cleanHistoryDB() error {
	return historyDB.Reorg(-1)
}
