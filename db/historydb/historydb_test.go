package historydb

import (
	"database/sql"
	"math"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	pass := os.Getenv("POSTGRES_PASS")
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
		log.Error("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func TestBlocks(t *testing.T) {
	var fromBlock, toBlock int64
	fromBlock = 0
	toBlock = 7
	// Reset DB
	test.WipeDB(historyDB.DB())
	// Generate blocks using til
	set1 := `
		Type: Blockchain
		// block 0 is stored as default in the DB
		// block 1 does not exist
		> block // blockNum=2
		> block // blockNum=3 
		> block // blockNum=4
		> block // blockNum=5
		> block // blockNum=6
	`
	tc := til.NewContext(1)
	blocks, err := tc.GenerateBlocks(set1)
	require.NoError(t, err)
	// Save timestamp of a block with UTC and change it without UTC
	timestamp := time.Now().Add(time.Second * 13)
	blocks[fromBlock].Block.Timestamp = timestamp
	// Insert blocks into DB
	for i := 0; i < len(blocks); i++ {
		err := historyDB.AddBlock(&blocks[i].Block)
		assert.NoError(t, err)
	}
	// Add block 0 to the generated blocks
	blocks = append(
		[]common.BlockData{common.BlockData{Block: test.Block0}}, //nolint:gofmt
		blocks...,
	)
	// Get all blocks from DB
	fetchedBlocks, err := historyDB.GetBlocks(fromBlock, toBlock)
	assert.Equal(t, len(blocks), len(fetchedBlocks))
	// Compare generated vs getted blocks
	assert.NoError(t, err)
	for i := range fetchedBlocks {
		assertEqualBlock(t, &blocks[i].Block, &fetchedBlocks[i])
	}
	// Compare saved timestamp vs getted
	nameZoneUTC, offsetUTC := timestamp.UTC().Zone()
	zoneFetchedBlock, offsetFetchedBlock := fetchedBlocks[fromBlock].Timestamp.Zone()
	assert.Equal(t, nameZoneUTC, zoneFetchedBlock)
	assert.Equal(t, offsetUTC, offsetFetchedBlock)
	// Get blocks from the DB one by one
	for i := int64(2); i < toBlock; i++ { // avoid block 0 for simplicity
		fetchedBlock, err := historyDB.GetBlock(i)
		assert.NoError(t, err)
		assertEqualBlock(t, &blocks[i-1].Block, fetchedBlock)
	}
	// Get last block
	lastBlock, err := historyDB.GetLastBlock()
	assert.NoError(t, err)
	assertEqualBlock(t, &blocks[len(blocks)-1].Block, lastBlock)
}

func assertEqualBlock(t *testing.T, expected *common.Block, actual *common.Block) {
	assert.Equal(t, expected.EthBlockNum, actual.EthBlockNum)
	assert.Equal(t, expected.Hash, actual.Hash)
	assert.Equal(t, expected.Timestamp.Unix(), actual.Timestamp.Unix())
}

func TestBatches(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())
	// Generate batches using til (and blocks for foreign key)
	set := `
		Type: Blockchain
		
		AddToken(1) // Will have value in USD
		AddToken(2) // Will NOT have value in USD
		CreateAccountDeposit(1) A: 2000
		CreateAccountDeposit(2) A: 2000
		CreateAccountDeposit(1) B: 1000
		CreateAccountDeposit(2) B: 1000
		> batchL1 
		> batchL1 
		Transfer(1) A-B: 100 (5)
		Transfer(2) B-A: 100 (199)
		> batch   // batchNum=2, L2 only batch, forges transfers (mixed case of with(out) USD value)
		> block
		Transfer(1) A-B: 100 (5)
		> batch   // batchNum=3, L2 only batch, forges transfer (with USD value)
		Transfer(2) B-A: 100 (199)
		> batch   // batchNum=4, L2 only batch, forges transfer (without USD value)
		> block
	`
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.Nil(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	assert.Nil(t, err)
	// Insert to DB
	batches := []common.Batch{}
	tokensValue := make(map[common.TokenID]float64)
	lastL1TxsNum := new(int64)
	for _, block := range blocks {
		// Insert block
		assert.NoError(t, historyDB.AddBlock(&block.Block))
		// Insert tokens
		for i, token := range block.Rollup.AddedTokens {
			assert.NoError(t, historyDB.AddToken(&token)) //nolint:gosec
			if i%2 != 0 {
				// Set value to the token
				value := (float64(i) + 5) * 5.389329
				assert.NoError(t, historyDB.UpdateTokenValue(token.Symbol, value))
				tokensValue[token.TokenID] = value / math.Pow(10, float64(token.Decimals))
			}
		}
		// Combine all generated batches into single array
		for _, batch := range block.Rollup.Batches {
			batches = append(batches, batch.Batch)
			forgeTxsNum := batch.Batch.ForgeL1TxsNum
			if forgeTxsNum != nil && (lastL1TxsNum == nil || *lastL1TxsNum < *forgeTxsNum) {
				*lastL1TxsNum = *forgeTxsNum
			}
		}
	}
	// Insert batches
	assert.NoError(t, historyDB.AddBatches(batches))
	// Set expected total fee
	for _, batch := range batches {
		total := .0
		for tokenID, amount := range batch.CollectedFees {
			af := new(big.Float).SetInt(amount)
			amountFloat, _ := af.Float64()
			total += tokensValue[tokenID] * amountFloat
		}
		batch.TotalFeesUSD = &total
	}
	// Get batches from the DB
	fetchedBatches, err := historyDB.GetBatches(0, common.BatchNum(len(batches)+1))
	assert.NoError(t, err)
	assert.Equal(t, len(batches), len(fetchedBatches))
	for i, fetchedBatch := range fetchedBatches {
		assert.Equal(t, batches[i], fetchedBatch)
	}
	// Test GetLastBatchNum
	fetchedLastBatchNum, err := historyDB.GetLastBatchNum()
	assert.NoError(t, err)
	assert.Equal(t, batches[len(batches)-1].BatchNum, fetchedLastBatchNum)
	// Test GetLastL1TxsNum
	fetchedLastL1TxsNum, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Equal(t, lastL1TxsNum, fetchedLastL1TxsNum)
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
	fetchedBids, err := historyDB.GetAllBids()
	assert.NoError(t, err)
	// Compare fetched bids vs generated bids
	for i, bid := range fetchedBids {
		assert.Equal(t, bids[i], bid)
	}
}

func TestTokens(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens, ethToken := test.GenTokens(nTokens, blocks)
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	tokens = append([]common.Token{ethToken}, tokens...)
	limit := uint(10)
	// Fetch tokens
	fetchedTokens, _, err := historyDB.GetTokens(nil, nil, "", nil, &limit, OrderAsc)
	assert.NoError(t, err)
	// Compare fetched tokens vs generated tokens
	// All the tokens should have USDUpdate setted by the DB trigger
	for i, token := range fetchedTokens {
		assert.Equal(t, tokens[i].TokenID, token.TokenID)
		assert.Equal(t, tokens[i].EthBlockNum, token.EthBlockNum)
		assert.Equal(t, tokens[i].EthAddr, token.EthAddr)
		assert.Equal(t, tokens[i].Name, token.Name)
		assert.Equal(t, tokens[i].Symbol, token.Symbol)
		assert.Nil(t, token.USD)
		assert.Nil(t, token.USDUpdate)
	}

	// Update token value
	for i, token := range tokens {
		value := 1.01 * float64(i)
		assert.NoError(t, historyDB.UpdateTokenValue(token.Symbol, value))
	}
	// Fetch tokens
	fetchedTokens, _, err = historyDB.GetTokens(nil, nil, "", nil, &limit, OrderAsc)
	assert.NoError(t, err)
	// Compare fetched tokens vs generated tokens
	// All the tokens should have USDUpdate setted by the DB trigger
	for i, token := range fetchedTokens {
		value := 1.01 * float64(i)
		assert.Equal(t, value, *token.USD)
		nameZone, offset := token.USDUpdate.Zone()
		assert.Equal(t, "UTC", nameZone)
		assert.Equal(t, 0, offset)
	}
}

func TestAccounts(t *testing.T) {
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens, ethToken := test.GenTokens(nTokens, blocks)
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	tokens = append([]common.Token{ethToken}, tokens...)
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
	fetchedAccs, err := historyDB.GetAllAccounts()
	assert.NoError(t, err)
	// Compare fetched accounts vs generated accounts
	for i, acc := range fetchedAccs {
		accs[i].Balance = nil
		assert.Equal(t, accs[i], acc)
	}
}

func TestTxs(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain
		
		AddToken(1)
		AddToken(2)
		CreateAccountDeposit(1) A: 10
		CreateAccountDeposit(1) B: 10
		> batchL1
		> batchL1
		> block
				
		CreateAccountDepositTransfer(1) C-A: 20, 10
		CreateAccountCoordinator(1) User0
		> batchL1
		> batchL1
		> block

		Deposit(1) B: 10
		Deposit(1) C: 10
		Transfer(1) C-A : 10 (1)
		Transfer(1) B-C : 10 (1)
		Transfer(1) A-B : 10 (1)
		Exit(1) A: 10 (1)
		> batch
		> block
		
		DepositTransfer(1) A-B: 10, 10
		> batchL1
		> block

		ForceTransfer(1) A-B: 10
		ForceExit(1) A: 5
		> batchL1
		> batchL1
		> block

		CreateAccountDeposit(2) D: 10
		> batchL1
		> block

		CreateAccountDeposit(2) E: 10
		> batchL1
		> batchL1
		> block

	`
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.Nil(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	assert.Nil(t, err)

	// Sanity check
	require.Equal(t, 7, len(blocks))
	require.Equal(t, 2, len(blocks[0].Rollup.L1UserTxs))
	require.Equal(t, 1, len(blocks[1].Rollup.L1UserTxs))
	require.Equal(t, 2, len(blocks[2].Rollup.L1UserTxs))
	require.Equal(t, 1, len(blocks[3].Rollup.L1UserTxs))
	require.Equal(t, 2, len(blocks[4].Rollup.L1UserTxs))
	require.Equal(t, 1, len(blocks[5].Rollup.L1UserTxs))
	require.Equal(t, 1, len(blocks[6].Rollup.L1UserTxs))

	var null *common.BatchNum = nil
	var txID common.TxID

	// Insert blocks into DB
	for i := range blocks {
		if i == len(blocks)-1 {
			blocks[i].Block.Timestamp = time.Now()
			dbL1Txs, err := historyDB.GetAllL1UserTxs()
			assert.NoError(t, err)
			// Check batch_num is nil before forging
			assert.Equal(t, null, dbL1Txs[len(dbL1Txs)-1].BatchNum)
			// Save this TxId
			txID = dbL1Txs[len(dbL1Txs)-1].TxID
		}
		err = historyDB.AddBlockSCData(&blocks[i])
		assert.NoError(t, err)
	}

	// Check blocks
	dbBlocks, err := historyDB.GetAllBlocks()
	assert.NoError(t, err)
	assert.Equal(t, len(blocks)+1, len(dbBlocks))

	// Check batches
	batches, err := historyDB.GetAllBatches()
	assert.NoError(t, err)
	assert.Equal(t, 11, len(batches))

	// Check L1 Transactions
	dbL1Txs, err := historyDB.GetAllL1UserTxs()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(dbL1Txs))

	// Tx Type
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[0].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[1].Type)
	assert.Equal(t, common.TxTypeCreateAccountDepositTransfer, dbL1Txs[2].Type)
	assert.Equal(t, common.TxTypeDeposit, dbL1Txs[3].Type)
	assert.Equal(t, common.TxTypeDeposit, dbL1Txs[4].Type)
	assert.Equal(t, common.TxTypeDepositTransfer, dbL1Txs[5].Type)
	assert.Equal(t, common.TxTypeForceTransfer, dbL1Txs[6].Type)
	assert.Equal(t, common.TxTypeForceExit, dbL1Txs[7].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[8].Type)
	assert.Equal(t, common.TxTypeCreateAccountDeposit, dbL1Txs[9].Type)

	// Tx ID
	assert.Equal(t, "0x000000000000000001000000", dbL1Txs[0].TxID.String())
	assert.Equal(t, "0x000000000000000001000100", dbL1Txs[1].TxID.String())
	assert.Equal(t, "0x000000000000000003000000", dbL1Txs[2].TxID.String())
	assert.Equal(t, "0x000000000000000005000000", dbL1Txs[3].TxID.String())
	assert.Equal(t, "0x000000000000000005000100", dbL1Txs[4].TxID.String())
	assert.Equal(t, "0x000000000000000005000200", dbL1Txs[5].TxID.String())
	assert.Equal(t, "0x000000000000000006000000", dbL1Txs[6].TxID.String())
	assert.Equal(t, "0x000000000000000006000100", dbL1Txs[7].TxID.String())
	assert.Equal(t, "0x000000000000000008000000", dbL1Txs[8].TxID.String())
	assert.Equal(t, "0x000000000000000009000000", dbL1Txs[9].TxID.String())

	// Tx From IDx
	assert.Equal(t, common.Idx(0), dbL1Txs[0].FromIdx)
	assert.Equal(t, common.Idx(0), dbL1Txs[1].FromIdx)
	assert.Equal(t, common.Idx(0), dbL1Txs[2].FromIdx)
	assert.NotEqual(t, common.Idx(0), dbL1Txs[3].FromIdx)
	assert.NotEqual(t, common.Idx(0), dbL1Txs[4].FromIdx)
	assert.NotEqual(t, common.Idx(0), dbL1Txs[5].FromIdx)
	assert.NotEqual(t, common.Idx(0), dbL1Txs[6].FromIdx)
	assert.NotEqual(t, common.Idx(0), dbL1Txs[7].FromIdx)
	assert.Equal(t, common.Idx(0), dbL1Txs[8].FromIdx)
	assert.Equal(t, common.Idx(0), dbL1Txs[9].FromIdx)
	assert.Equal(t, common.Idx(0), dbL1Txs[9].FromIdx)
	assert.Equal(t, dbL1Txs[5].FromIdx, dbL1Txs[6].FromIdx)
	assert.Equal(t, dbL1Txs[5].FromIdx, dbL1Txs[7].FromIdx)

	// Tx to IDx
	assert.Equal(t, dbL1Txs[2].ToIdx, dbL1Txs[5].FromIdx)
	assert.Equal(t, dbL1Txs[5].ToIdx, dbL1Txs[3].FromIdx)
	assert.Equal(t, dbL1Txs[6].ToIdx, dbL1Txs[3].FromIdx)

	// Token ID
	assert.Equal(t, common.TokenID(1), dbL1Txs[0].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[1].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[2].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[3].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[4].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[5].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[6].TokenID)
	assert.Equal(t, common.TokenID(1), dbL1Txs[7].TokenID)
	assert.Equal(t, common.TokenID(2), dbL1Txs[8].TokenID)
	assert.Equal(t, common.TokenID(2), dbL1Txs[9].TokenID)

	// Batch Number
	var bn common.BatchNum = common.BatchNum(2)

	assert.Equal(t, &bn, dbL1Txs[0].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[1].BatchNum)

	bn = common.BatchNum(4)
	assert.Equal(t, &bn, dbL1Txs[2].BatchNum)

	bn = common.BatchNum(7)
	assert.Equal(t, &bn, dbL1Txs[3].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[4].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[5].BatchNum)

	bn = common.BatchNum(8)
	assert.Equal(t, &bn, dbL1Txs[6].BatchNum)
	assert.Equal(t, &bn, dbL1Txs[7].BatchNum)

	bn = common.BatchNum(10)
	assert.Equal(t, &bn, dbL1Txs[8].BatchNum)

	bn = common.BatchNum(11)
	assert.Equal(t, &bn, dbL1Txs[9].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(2), dbL1Txs[0].EthBlockNum)
	assert.Equal(t, int64(2), dbL1Txs[1].EthBlockNum)
	assert.Equal(t, int64(3), dbL1Txs[2].EthBlockNum)
	assert.Equal(t, int64(4), dbL1Txs[3].EthBlockNum)
	assert.Equal(t, int64(4), dbL1Txs[4].EthBlockNum)
	assert.Equal(t, int64(5), dbL1Txs[5].EthBlockNum)
	assert.Equal(t, int64(6), dbL1Txs[6].EthBlockNum)
	assert.Equal(t, int64(6), dbL1Txs[7].EthBlockNum)
	assert.Equal(t, int64(7), dbL1Txs[8].EthBlockNum)
	assert.Equal(t, int64(8), dbL1Txs[9].EthBlockNum)

	// User Origin
	assert.Equal(t, true, dbL1Txs[0].UserOrigin)
	assert.Equal(t, true, dbL1Txs[1].UserOrigin)
	assert.Equal(t, true, dbL1Txs[2].UserOrigin)
	assert.Equal(t, true, dbL1Txs[3].UserOrigin)
	assert.Equal(t, true, dbL1Txs[4].UserOrigin)
	assert.Equal(t, true, dbL1Txs[5].UserOrigin)
	assert.Equal(t, true, dbL1Txs[6].UserOrigin)
	assert.Equal(t, true, dbL1Txs[7].UserOrigin)
	assert.Equal(t, true, dbL1Txs[8].UserOrigin)
	assert.Equal(t, true, dbL1Txs[9].UserOrigin)

	// Load Amount
	assert.Equal(t, big.NewInt(10), dbL1Txs[0].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[1].LoadAmount)
	assert.Equal(t, big.NewInt(20), dbL1Txs[2].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[3].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[4].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[5].LoadAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[6].LoadAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[7].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[8].LoadAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[9].LoadAmount)

	// Check saved txID's batch_num is not nil
	assert.Equal(t, txID, dbL1Txs[len(dbL1Txs)-2].TxID)
	assert.NotEqual(t, null, dbL1Txs[len(dbL1Txs)-2].BatchNum)

	// Check Coordinator TXs
	coordTxs, err := historyDB.GetAllL1CoordinatorTxs()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(coordTxs))
	assert.Equal(t, common.TxTypeCreateAccountDeposit, coordTxs[0].Type)
	assert.Equal(t, false, coordTxs[0].UserOrigin)

	// Check L2 TXs
	dbL2Txs, err := historyDB.GetAllL2Txs()
	assert.NoError(t, err)
	assert.Equal(t, 4, len(dbL2Txs))

	// Tx Type
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[0].Type)
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[1].Type)
	assert.Equal(t, common.TxTypeTransfer, dbL2Txs[2].Type)
	assert.Equal(t, common.TxTypeExit, dbL2Txs[3].Type)

	// Tx ID
	assert.Equal(t, "0x020000000001030000000001", dbL2Txs[0].TxID.String())
	assert.Equal(t, "0x020000000001010000000001", dbL2Txs[1].TxID.String())
	assert.Equal(t, "0x020000000001000000000001", dbL2Txs[2].TxID.String())
	assert.Equal(t, "0x020000000001000000000002", dbL2Txs[3].TxID.String())

	// Tx From and To IDx
	assert.Equal(t, dbL2Txs[0].ToIdx, dbL2Txs[2].FromIdx)
	assert.Equal(t, dbL2Txs[1].ToIdx, dbL2Txs[0].FromIdx)
	assert.Equal(t, dbL2Txs[2].ToIdx, dbL2Txs[1].FromIdx)

	// Batch Number
	assert.Equal(t, common.BatchNum(5), dbL2Txs[0].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[1].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[2].BatchNum)
	assert.Equal(t, common.BatchNum(5), dbL2Txs[3].BatchNum)

	// eth_block_num
	assert.Equal(t, int64(4), dbL2Txs[0].EthBlockNum)
	assert.Equal(t, int64(4), dbL2Txs[1].EthBlockNum)
	assert.Equal(t, int64(4), dbL2Txs[2].EthBlockNum)

	// Amount
	assert.Equal(t, big.NewInt(10), dbL2Txs[0].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[1].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[2].Amount)
	assert.Equal(t, big.NewInt(10), dbL2Txs[3].Amount)
}

func TestExitTree(t *testing.T) {
	nBatches := 17
	blocks := setTestBlocks(1, 10)
	batches := test.GenBatches(nBatches, blocks)
	err := historyDB.AddBatches(batches)
	assert.NoError(t, err)
	const nTokens = 50
	tokens, ethToken := test.GenTokens(nTokens, blocks)
	err = historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	tokens = append([]common.Token{ethToken}, tokens...)
	const nAccounts = 3
	accs := test.GenAccounts(nAccounts, 0, tokens, nil, nil, batches)
	assert.NoError(t, historyDB.AddAccounts(accs))
	exitTree := test.GenExitTree(nBatches, batches, accs, blocks)
	err = historyDB.AddExitTree(exitTree)
	assert.NoError(t, err)
}

func TestGetL1UserTxs(t *testing.T) {
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain
		AddToken(1)
		AddToken(2)
		AddToken(3)

		CreateAccountDeposit(1) A: 20
		CreateAccountDeposit(2) A: 20
		CreateAccountDeposit(1) B: 5
		CreateAccountDeposit(1) C: 5
		CreateAccountDeposit(1) D: 5

		> block
	`
	tc := til.NewContext(128)
	blocks, err := tc.GenerateBlocks(set)
	require.Nil(t, err)
	// Sanity check
	require.Equal(t, 1, len(blocks))
	require.Equal(t, 5, len(blocks[0].Rollup.L1UserTxs))

	toForgeL1TxsNum := int64(1)

	for i := range blocks {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.Nil(t, err)
	}

	l1UserTxs, err := historyDB.GetL1UserTxs(toForgeL1TxsNum)
	require.Nil(t, err)
	assert.Equal(t, 5, len(l1UserTxs))
	assert.Equal(t, blocks[0].Rollup.L1UserTxs, l1UserTxs)

	// No l1UserTxs for this toForgeL1TxsNum
	l1UserTxs, err = historyDB.GetL1UserTxs(2)
	require.Nil(t, err)
	assert.Equal(t, 0, len(l1UserTxs))
}

func exampleInitSCVars() (*common.RollupVariables, *common.AuctionVariables, *common.WDelayerVariables) {
	//nolint:govet
	rollup := &common.RollupVariables{
		0,
		big.NewInt(10),
		12,
		13,
		[5]common.Bucket{},
	}
	//nolint:govet
	auction := &common.AuctionVariables{
		0,
		ethCommon.BigToAddress(big.NewInt(2)),
		ethCommon.BigToAddress(big.NewInt(3)),
		[6]*big.Int{
			big.NewInt(1), big.NewInt(2), big.NewInt(3),
			big.NewInt(4), big.NewInt(5), big.NewInt(6),
		},
		0,
		2,
		4320,
		[3]uint16{10, 11, 12},
		1000,
		20,
	}
	//nolint:govet
	wDelayer := &common.WDelayerVariables{
		0,
		ethCommon.BigToAddress(big.NewInt(2)),
		ethCommon.BigToAddress(big.NewInt(3)),
		ethCommon.BigToAddress(big.NewInt(4)),
		13,
		14,
		false,
	}
	return rollup, auction, wDelayer
}

func TestSetInitialSCVars(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, _, _, err := historyDB.GetSCVars()
	assert.Equal(t, sql.ErrNoRows, err)
	rollup, auction, wDelayer := exampleInitSCVars()
	err = historyDB.SetInitialSCVars(rollup, auction, wDelayer)
	require.Nil(t, err)
	dbRollup, dbAuction, dbWDelayer, err := historyDB.GetSCVars()
	assert.Nil(t, err)
	require.Equal(t, rollup, dbRollup)
	require.Equal(t, auction, dbAuction)
	require.Equal(t, wDelayer, dbWDelayer)
}

func TestUpdateExitTree(t *testing.T) {
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain

		AddToken(1)

		CreateAccountDeposit(1) C: 2000 // Idx=256+2=258
		CreateAccountDeposit(1) D: 500  // Idx=256+3=259

		CreateAccountCoordinator(1) A // Idx=256+0=256
		CreateAccountCoordinator(1) B // Idx=256+1=257

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{5}
		> batchL1 // forge defined L1UserTxs{5}, freeze L1UserTxs{nil}
		> block // blockNum=2

		ForceExit(1) A: 100
		ForceExit(1) B: 80

		Exit(1) C: 50 (172)
		Exit(1) D: 30 (172)

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{3}
		> batchL1 // forge L1UserTxs{3}, freeze defined L1UserTxs{nil}
		> block // blockNum=3

		> block // blockNum=4 (empty block)
		> block // blockNum=5 (empty block)
	`

	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.Nil(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	assert.Nil(t, err)

	// Add all blocks except for the last two
	for i := range blocks[:len(blocks)-2] {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.Nil(t, err)
	}

	// Add withdraws to the second-to-last block, and insert block into the DB
	block := &blocks[len(blocks)-2]
	require.Equal(t, int64(4), block.Block.EthBlockNum)
	tokenAddr := blocks[0].Rollup.AddedTokens[0].EthAddr
	// block.WDelayer.Deposits = append(block.WDelayer.Deposits,
	// 	common.WDelayerTransfer{Owner: tc.UsersByIdx[257].Addr, Token: tokenAddr, Amount: big.NewInt(80)}, // 257
	// 	common.WDelayerTransfer{Owner: tc.UsersByIdx[259].Addr, Token: tokenAddr, Amount: big.NewInt(15)}, // 259
	// )
	block.Rollup.Withdrawals = append(block.Rollup.Withdrawals,
		common.WithdrawInfo{Idx: 256, NumExitRoot: 4, InstantWithdraw: true},
		common.WithdrawInfo{Idx: 257, NumExitRoot: 4, InstantWithdraw: false,
			Owner: tc.UsersByIdx[257].Addr, Token: tokenAddr},
		common.WithdrawInfo{Idx: 258, NumExitRoot: 3, InstantWithdraw: true},
		common.WithdrawInfo{Idx: 259, NumExitRoot: 3, InstantWithdraw: false,
			Owner: tc.UsersByIdx[259].Addr, Token: tokenAddr},
	)
	err = historyDB.addBlock(historyDB.db, &block.Block)
	require.Nil(t, err)

	err = historyDB.updateExitTree(historyDB.db, block.Block.EthBlockNum,
		block.Rollup.Withdrawals, block.WDelayer.Withdrawals)
	require.Nil(t, err)

	// Check that exits in DB match with the expected values
	dbExits, err := historyDB.GetAllExits()
	require.Nil(t, err)
	assert.Equal(t, 4, len(dbExits))
	dbExitsByIdx := make(map[common.Idx]common.ExitInfo)
	for _, dbExit := range dbExits {
		dbExitsByIdx[dbExit.AccountIdx] = dbExit
	}
	for _, withdraw := range block.Rollup.Withdrawals {
		assert.Equal(t, withdraw.NumExitRoot, dbExitsByIdx[withdraw.Idx].BatchNum)
		if withdraw.InstantWithdraw {
			assert.Equal(t, &block.Block.EthBlockNum, dbExitsByIdx[withdraw.Idx].InstantWithdrawn)
		} else {
			assert.Equal(t, &block.Block.EthBlockNum, dbExitsByIdx[withdraw.Idx].DelayedWithdrawRequest)
		}
	}

	// Add delayed withdraw to the last block, and insert block into the DB
	block = &blocks[len(blocks)-1]
	require.Equal(t, int64(5), block.Block.EthBlockNum)
	block.WDelayer.Withdrawals = append(block.WDelayer.Withdrawals,
		common.WDelayerTransfer{
			Owner:  tc.UsersByIdx[257].Addr,
			Token:  tokenAddr,
			Amount: big.NewInt(80),
		})
	err = historyDB.addBlock(historyDB.db, &block.Block)
	require.Nil(t, err)

	err = historyDB.updateExitTree(historyDB.db, block.Block.EthBlockNum,
		block.Rollup.Withdrawals, block.WDelayer.Withdrawals)
	require.Nil(t, err)

	// Check that delayed withdrawn has been set
	dbExits, err = historyDB.GetAllExits()
	require.Nil(t, err)
	for _, dbExit := range dbExits {
		dbExitsByIdx[dbExit.AccountIdx] = dbExit
	}
	require.Equal(t, &block.Block.EthBlockNum, dbExitsByIdx[257].DelayedWithdrawn)
}

func TestGetBestBidCoordinator(t *testing.T) {
	test.WipeDB(historyDB.DB())

	rollup, auction, wDelayer := exampleInitSCVars()
	err := historyDB.SetInitialSCVars(rollup, auction, wDelayer)
	require.Nil(t, err)

	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(`
		Type: Blockchain
		> block // blockNum=2
	`)
	require.Nil(t, err)
	err = historyDB.AddBlockSCData(&blocks[0])
	require.Nil(t, err)

	coords := []common.Coordinator{
		{
			Bidder:      ethCommon.BigToAddress(big.NewInt(1)),
			Forger:      ethCommon.BigToAddress(big.NewInt(2)),
			EthBlockNum: 2,
			URL:         "foo",
		},
		{
			Bidder:      ethCommon.BigToAddress(big.NewInt(3)),
			Forger:      ethCommon.BigToAddress(big.NewInt(4)),
			EthBlockNum: 2,
			URL:         "bar",
		},
	}
	err = historyDB.addCoordinators(historyDB.db, coords)
	require.Nil(t, err)
	err = historyDB.addBids(historyDB.db, []common.Bid{
		{
			SlotNum:     10,
			BidValue:    big.NewInt(10),
			EthBlockNum: 2,
			Bidder:      coords[0].Bidder,
		},
		{
			SlotNum:     10,
			BidValue:    big.NewInt(20),
			EthBlockNum: 2,
			Bidder:      coords[1].Bidder,
		},
	})
	require.Nil(t, err)

	forger10, err := historyDB.GetBestBidCoordinator(10)
	require.Nil(t, err)
	require.Equal(t, coords[1].Forger, forger10.Forger)
	require.Equal(t, coords[1].Bidder, forger10.Bidder)
	require.Equal(t, coords[1].URL, forger10.URL)

	_, err = historyDB.GetBestBidCoordinator(11)
	require.Equal(t, sql.ErrNoRows, err)
}

// setTestBlocks WARNING: this will delete the blocks and recreate them
func setTestBlocks(from, to int64) []common.Block {
	test.WipeDB(historyDB.DB())
	blocks := test.GenBlocks(from, to)
	if err := historyDB.AddBlocks(blocks); err != nil {
		panic(err)
	}
	return blocks
}
