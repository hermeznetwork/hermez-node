package historydb

import (
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/tracerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var historyDB *HistoryDB
var historyDBWithACC *HistoryDB

func TestMain(m *testing.M) {
	// init DB
	db, err := dbUtils.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	historyDB = NewHistoryDB(db, db, nil)
	if err != nil {
		panic(err)
	}
	apiConnCon := dbUtils.NewAPIConnectionController(1, time.Second)
	historyDBWithACC = NewHistoryDB(db, db, apiConnCon)
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
	tc := til.NewContext(uint16(0), 1)
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
		[]common.BlockData{{Block: test.Block0}}, //nolint:gofmt
		blocks...,
	)
	// Get all blocks from DB
	fetchedBlocks, err := historyDB.getBlocks(fromBlock, toBlock)
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
	assert.Equal(t, expected.Num, actual.Num)
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
	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	// Insert to DB
	batches := []common.Batch{}
	tokensValue := make(map[common.TokenID]float64)
	lastL1TxsNum := new(int64)
	lastL1BatchBlockNum := int64(0)
	for _, block := range blocks {
		// Insert block
		assert.NoError(t, historyDB.AddBlock(&block.Block))
		// Insert tokens
		for i, token := range block.Rollup.AddedTokens {
			assert.NoError(t, historyDB.AddToken(&token)) //nolint:gosec
			if i%2 != 0 {
				// Set value to the token
				value := (float64(i) + 5) * 5.389329
				assert.NoError(t, historyDB.UpdateTokenValue(token.EthAddr, value))
				tokensValue[token.TokenID] = value / math.Pow(10, float64(token.Decimals))
			}
		}
		// Combine all generated batches into single array
		for _, batch := range block.Rollup.Batches {
			batch.Batch.GasPrice = big.NewInt(0)
			batches = append(batches, batch.Batch)
			forgeTxsNum := batch.Batch.ForgeL1TxsNum
			if forgeTxsNum != nil && (lastL1TxsNum == nil || *lastL1TxsNum < *forgeTxsNum) {
				*lastL1TxsNum = *forgeTxsNum
				lastL1BatchBlockNum = batch.Batch.EthBlockNum
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
	// Test GetLastBatch
	fetchedLastBatch, err := historyDB.GetLastBatch()
	assert.NoError(t, err)
	assert.Equal(t, &batches[len(batches)-1], fetchedLastBatch)
	// Test GetLastL1TxsNum
	fetchedLastL1TxsNum, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
	assert.Equal(t, lastL1TxsNum, fetchedLastL1TxsNum)
	// Test GetLastL1BatchBlockNum
	fetchedLastL1BatchBlockNum, err := historyDB.GetLastL1BatchBlockNum()
	assert.NoError(t, err)
	assert.Equal(t, lastL1BatchBlockNum, fetchedLastL1BatchBlockNum)
	// Test GetBatch
	fetchedBatch, err := historyDB.GetBatch(1)
	require.NoError(t, err)
	assert.Equal(t, &batches[0], fetchedBatch)
	_, err = historyDB.GetBatch(common.BatchNum(len(batches) + 1))
	assert.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))
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
	// Fetch tokens
	fetchedTokens, err := historyDB.GetTokensTest()
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
		assert.NoError(t, historyDB.UpdateTokenValue(token.EthAddr, value))
	}
	// Fetch tokens
	fetchedTokens, err = historyDB.GetTokensTest()
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

func TestTokensUTF8(t *testing.T) {
	// Reset DB
	test.WipeDB(historyDB.DB())
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Prepare blocks in the DB
	blocks := setTestBlocks(fromBlock, toBlock)
	// Generate fake tokens
	const nTokens = 5
	tokens, ethToken := test.GenTokens(nTokens, blocks)
	nonUTFTokens := make([]common.Token, len(tokens))
	// Force token.name and token.symbol to be non UTF-8 Strings
	for i, token := range tokens {
		token.Name = fmt.Sprint("NON-UTF8-NAME-\xc5-", i)
		token.Symbol = fmt.Sprint("S-\xc5-", i)
		tokens[i] = token
		nonUTFTokens[i] = token
	}
	err := historyDB.AddTokens(tokens)
	assert.NoError(t, err)
	// Work with nonUTFTokens as tokens one gets updated and non UTF-8 characters are lost
	nonUTFTokens = append([]common.Token{ethToken}, nonUTFTokens...)
	// Fetch tokens
	fetchedTokens, err := historyDB.GetTokensTest()
	assert.NoError(t, err)
	// Compare fetched tokens vs generated tokens
	// All the tokens should have USDUpdate setted by the DB trigger
	for i, token := range fetchedTokens {
		assert.Equal(t, nonUTFTokens[i].TokenID, token.TokenID)
		assert.Equal(t, nonUTFTokens[i].EthBlockNum, token.EthBlockNum)
		assert.Equal(t, nonUTFTokens[i].EthAddr, token.EthAddr)
		assert.Equal(t, strings.ToValidUTF8(nonUTFTokens[i].Name, " "), token.Name)
		assert.Equal(t, strings.ToValidUTF8(nonUTFTokens[i].Symbol, " "), token.Symbol)
		assert.Nil(t, token.USD)
		assert.Nil(t, token.USDUpdate)
	}

	// Update token value
	for i, token := range nonUTFTokens {
		value := 1.01 * float64(i)
		assert.NoError(t, historyDB.UpdateTokenValue(token.EthAddr, value))
	}
	// Fetch tokens
	fetchedTokens, err = historyDB.GetTokensTest()
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
	// Test AccountBalances
	accUpdates := make([]common.AccountUpdate, len(accs))
	for i, acc := range accs {
		accUpdates[i] = common.AccountUpdate{
			EthBlockNum: batches[acc.BatchNum-1].EthBlockNum,
			BatchNum:    acc.BatchNum,
			Idx:         acc.Idx,
			Nonce:       nonce.Nonce(i),
			Balance:     big.NewInt(int64(i)),
		}
	}
	err = historyDB.AddAccountUpdates(accUpdates)
	require.NoError(t, err)
	fetchedAccBalances, err := historyDB.GetAllAccountUpdates()
	require.NoError(t, err)
	assert.Equal(t, accUpdates, fetchedAccBalances)
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
	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

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
	assert.Equal(t, "0x00e979da4b80d60a17ce56fa19278c6f3a7e1b43359fb8a8ea46d0264de7d653ab", dbL1Txs[0].TxID.String())
	assert.Equal(t, "0x00af9bf96eb60f2d618519402a2f6b07057a034fa2baefd379fe8e1c969f1c5cf4", dbL1Txs[1].TxID.String())
	assert.Equal(t, "0x00a256ee191905243320ea830840fd666a73c7b4e6f89ce4bd47ddf998dfee627a", dbL1Txs[2].TxID.String())
	assert.Equal(t, "0x00930696d03ae0a1e6150b6ccb88043cb539a4e06a7f8baf213029ce9a0600197e", dbL1Txs[3].TxID.String())
	assert.Equal(t, "0x00de8e41d49f23832f66364e8702c4b78237eb0c95542a94d34188e51696e74fc8", dbL1Txs[4].TxID.String())
	assert.Equal(t, "0x007a44d6d60b15f3789d4ff49d62377a70255bf13a8d42e41ef49bf4c7b77d2c1b", dbL1Txs[5].TxID.String())
	assert.Equal(t, "0x00c33f316240f8d33a973db2d0e901e4ac1c96de30b185fcc6b63dac4d0e147bd4", dbL1Txs[6].TxID.String())
	assert.Equal(t, "0x00b55f0882c5229d1be3d9d3c1a076290f249cd0bae5ae6e609234606befb91233", dbL1Txs[7].TxID.String())
	assert.Equal(t, "0x009133d4c8a412ca45f50bccdbcfdb8393b0dd8efe953d0cc3bcc82796b7a581b6", dbL1Txs[8].TxID.String())
	assert.Equal(t, "0x00f5e8ab141ac16d673e654ba7747c2f12e93ea2c50ba6c05563752ca531968c62", dbL1Txs[9].TxID.String())

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
	var bn = common.BatchNum(2)

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

	// Deposit Amount
	assert.Equal(t, big.NewInt(10), dbL1Txs[0].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[1].DepositAmount)
	assert.Equal(t, big.NewInt(20), dbL1Txs[2].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[3].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[4].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[5].DepositAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[6].DepositAmount)
	assert.Equal(t, big.NewInt(0), dbL1Txs[7].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[8].DepositAmount)
	assert.Equal(t, big.NewInt(10), dbL1Txs[9].DepositAmount)

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
	assert.Equal(t, "0x024e555248100b69a8aabf6d31719b9fe8a60dcc6c3407904a93c8d2d9ade18ee5", dbL2Txs[0].TxID.String())
	assert.Equal(t, "0x021ae87ca34d50ff35d98dfc0d7c95f2bf2e4ffeebb82ea71f43a8b0dfa5d36d89", dbL2Txs[1].TxID.String())
	assert.Equal(t, "0x024abce7f3f2382dc520ed557593f11dea1ee197e55b60402e664facc27aa19774", dbL2Txs[2].TxID.String())
	assert.Equal(t, "0x02f921ad9e7a6e59606570fe12a7dde0e36014197de0363b9b45e5097d6f2b1dd0", dbL2Txs[3].TxID.String())

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

func TestGetUnforgedL1UserTxs(t *testing.T) {
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

		> batchL1
		> block

		CreateAccountDeposit(1) E: 5
		CreateAccountDeposit(1) F: 5
		> block

	`
	tc := til.NewContext(uint16(0), 128)
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	// Sanity check
	require.Equal(t, 3, len(blocks))
	require.Equal(t, 5, len(blocks[0].Rollup.L1UserTxs))

	for i := range blocks {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}

	l1UserTxs, err := historyDB.GetUnforgedL1UserFutureTxs(0)
	require.NoError(t, err)
	assert.Equal(t, 7, len(l1UserTxs))

	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(1)
	require.NoError(t, err)
	assert.Equal(t, 5, len(l1UserTxs))
	assert.Equal(t, blocks[0].Rollup.L1UserTxs, l1UserTxs)

	l1UserTxs, err = historyDB.GetUnforgedL1UserFutureTxs(1)
	require.NoError(t, err)
	assert.Equal(t, 2, len(l1UserTxs))

	count, err := historyDB.GetUnforgedL1UserTxsCount()
	require.NoError(t, err)
	assert.Equal(t, 7, count)

	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(2)
	require.NoError(t, err)
	assert.Equal(t, 2, len(l1UserTxs))

	l1UserTxs, err = historyDB.GetUnforgedL1UserFutureTxs(2)
	require.NoError(t, err)
	assert.Equal(t, 0, len(l1UserTxs))

	// No l1UserTxs for this toForgeL1TxsNum
	l1UserTxs, err = historyDB.GetUnforgedL1UserTxs(3)
	require.NoError(t, err)
	assert.Equal(t, 0, len(l1UserTxs))
}

func exampleInitSCVars() (*common.RollupVariables, *common.AuctionVariables, *common.WDelayerVariables) {
	rollup := &common.RollupVariables{
		EthBlockNum:           0,
		FeeAddToken:           big.NewInt(10),
		ForgeL1L2BatchTimeout: 12,
		WithdrawalDelay:       13,
		Buckets:               []common.BucketParams{},
		SafeMode:              false,
	}
	auction := &common.AuctionVariables{
		EthBlockNum:        0,
		DonationAddress:    ethCommon.BigToAddress(big.NewInt(2)),
		BootCoordinator:    ethCommon.BigToAddress(big.NewInt(3)),
		BootCoordinatorURL: "https://boot.coord.com",
		DefaultSlotSetBid: [6]*big.Int{
			big.NewInt(1), big.NewInt(2), big.NewInt(3),
			big.NewInt(4), big.NewInt(5), big.NewInt(6),
		},
		DefaultSlotSetBidSlotNum: 0,
		ClosedAuctionSlots:       2,
		OpenAuctionSlots:         4320,
		AllocationRatio:          [3]uint16{10, 11, 12},
		Outbidding:               1000,
		SlotDeadline:             20,
	}
	wDelayer := &common.WDelayerVariables{
		EthBlockNum:                0,
		HermezGovernanceAddress:    ethCommon.BigToAddress(big.NewInt(2)),
		EmergencyCouncilAddress:    ethCommon.BigToAddress(big.NewInt(3)),
		WithdrawalDelay:            13,
		EmergencyModeStartingBlock: 14,
		EmergencyMode:              false,
	}
	return rollup, auction, wDelayer
}

func TestSetInitialSCVars(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, _, _, err := historyDB.GetSCVars()
	assert.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))
	rollup, auction, wDelayer := exampleInitSCVars()
	err = historyDB.SetInitialSCVars(rollup, auction, wDelayer)
	require.NoError(t, err)
	dbRollup, dbAuction, dbWDelayer, err := historyDB.GetSCVars()
	require.NoError(t, err)
	require.Equal(t, rollup, dbRollup)
	require.Equal(t, auction, dbAuction)
	require.Equal(t, wDelayer, dbWDelayer)
}

func TestSetExtraInfoForgedL1UserTxs(t *testing.T) {
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain

		AddToken(1)

		CreateAccountDeposit(1) A: 2000
		CreateAccountDeposit(1) B: 500
		CreateAccountDeposit(1) C: 500

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{*}
		> block // blockNum=2

		> batchL1 // forge defined L1UserTxs{*}
		> block // blockNum=3
	`

	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	// Add only first block so that the L1UserTxs are not marked as forged
	for i := range blocks[:1] {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}
	// Add second batch to trigger the update of the batch_num,
	// while avoiding the implicit call of setExtraInfoForgedL1UserTxs
	err = historyDB.addBlock(historyDB.dbWrite, &blocks[1].Block)
	require.NoError(t, err)
	err = historyDB.addBatch(historyDB.dbWrite, &blocks[1].Rollup.Batches[0].Batch)
	require.NoError(t, err)
	err = historyDB.addAccounts(historyDB.dbWrite, blocks[1].Rollup.Batches[0].CreatedAccounts)
	require.NoError(t, err)

	// Set the Effective{Amount,DepositAmount} of the L1UserTxs that are forged in the second block
	l1Txs := blocks[1].Rollup.Batches[0].L1UserTxs
	require.Equal(t, 3, len(l1Txs))
	// Change some values to test all cases
	l1Txs[1].EffectiveAmount = big.NewInt(0)
	l1Txs[2].EffectiveDepositAmount = big.NewInt(0)
	l1Txs[2].EffectiveAmount = big.NewInt(0)
	err = historyDB.setExtraInfoForgedL1UserTxs(historyDB.dbWrite, l1Txs)
	require.NoError(t, err)

	dbL1Txs, err := historyDB.GetAllL1UserTxs()
	require.NoError(t, err)
	for i, tx := range dbL1Txs {
		log.Infof("%d %v %v", i, tx.EffectiveAmount, tx.EffectiveDepositAmount)
		assert.NotNil(t, tx.EffectiveAmount)
		assert.NotNil(t, tx.EffectiveDepositAmount)
		switch tx.TxID {
		case l1Txs[0].TxID:
			assert.Equal(t, l1Txs[0].DepositAmount, tx.EffectiveDepositAmount)
			assert.Equal(t, l1Txs[0].Amount, tx.EffectiveAmount)
		case l1Txs[1].TxID:
			assert.Equal(t, l1Txs[1].DepositAmount, tx.EffectiveDepositAmount)
			assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)
		case l1Txs[2].TxID:
			assert.Equal(t, big.NewInt(0), tx.EffectiveDepositAmount)
			assert.Equal(t, big.NewInt(0), tx.EffectiveAmount)
		}
	}
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

	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

	// Add all blocks except for the last two
	for i := range blocks[:len(blocks)-2] {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}

	// Add withdraws to the second-to-last block, and insert block into the DB
	block := &blocks[len(blocks)-2]
	require.Equal(t, int64(4), block.Block.Num)
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
	err = historyDB.addBlock(historyDB.dbWrite, &block.Block)
	require.NoError(t, err)

	err = historyDB.updateExitTree(historyDB.dbWrite, block.Block.Num,
		block.Rollup.Withdrawals, block.WDelayer.Withdrawals)
	require.NoError(t, err)

	// Check that exits in DB match with the expected values
	dbExits, err := historyDB.GetAllExits()
	require.NoError(t, err)
	assert.Equal(t, 4, len(dbExits))
	dbExitsByIdx := make(map[common.Idx]common.ExitInfo)
	for _, dbExit := range dbExits {
		dbExitsByIdx[dbExit.AccountIdx] = dbExit
	}
	for _, withdraw := range block.Rollup.Withdrawals {
		assert.Equal(t, withdraw.NumExitRoot, dbExitsByIdx[withdraw.Idx].BatchNum)
		if withdraw.InstantWithdraw {
			assert.Equal(t, &block.Block.Num, dbExitsByIdx[withdraw.Idx].InstantWithdrawn)
		} else {
			assert.Equal(t, &block.Block.Num, dbExitsByIdx[withdraw.Idx].DelayedWithdrawRequest)
		}
	}

	// Add delayed withdraw to the last block, and insert block into the DB
	block = &blocks[len(blocks)-1]
	require.Equal(t, int64(5), block.Block.Num)
	block.WDelayer.Withdrawals = append(block.WDelayer.Withdrawals,
		common.WDelayerTransfer{
			Owner:  tc.UsersByIdx[257].Addr,
			Token:  tokenAddr,
			Amount: big.NewInt(80),
		})
	err = historyDB.addBlock(historyDB.dbWrite, &block.Block)
	require.NoError(t, err)

	err = historyDB.updateExitTree(historyDB.dbWrite, block.Block.Num,
		block.Rollup.Withdrawals, block.WDelayer.Withdrawals)
	require.NoError(t, err)

	// Check that delayed withdrawn has been set
	dbExits, err = historyDB.GetAllExits()
	require.NoError(t, err)
	for _, dbExit := range dbExits {
		dbExitsByIdx[dbExit.AccountIdx] = dbExit
	}
	require.Equal(t, &block.Block.Num, dbExitsByIdx[257].DelayedWithdrawn)
}

func TestGetBestBidCoordinator(t *testing.T) {
	test.WipeDB(historyDB.DB())

	rollup, auction, wDelayer := exampleInitSCVars()
	err := historyDB.SetInitialSCVars(rollup, auction, wDelayer)
	require.NoError(t, err)

	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(`
		Type: Blockchain
		> block // blockNum=2
	`)
	require.NoError(t, err)
	err = historyDB.AddBlockSCData(&blocks[0])
	require.NoError(t, err)

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
	err = historyDB.addCoordinators(historyDB.dbWrite, coords)
	require.NoError(t, err)

	bids := []common.Bid{
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
	}

	err = historyDB.addBids(historyDB.dbWrite, bids)
	require.NoError(t, err)

	forger10, err := historyDB.GetBestBidCoordinator(10)
	require.NoError(t, err)
	require.Equal(t, coords[1].Forger, forger10.Forger)
	require.Equal(t, coords[1].Bidder, forger10.Bidder)
	require.Equal(t, coords[1].URL, forger10.URL)
	require.Equal(t, bids[1].SlotNum, forger10.SlotNum)
	require.Equal(t, bids[1].BidValue, forger10.BidValue)
	for i := range forger10.DefaultSlotSetBid {
		require.Equal(t, auction.DefaultSlotSetBid[i], forger10.DefaultSlotSetBid[i])
	}

	_, err = historyDB.GetBestBidCoordinator(11)
	require.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))
}

func TestAddBucketUpdates(t *testing.T) {
	test.WipeDB(historyDB.DB())
	const fromBlock int64 = 1
	const toBlock int64 = 5 + 1
	setTestBlocks(fromBlock, toBlock)

	bucketUpdates := []common.BucketUpdate{
		{
			EthBlockNum: 4,
			NumBucket:   0,
			BlockStamp:  4,
			Withdrawals: big.NewInt(123),
		},
		{
			EthBlockNum: 5,
			NumBucket:   2,
			BlockStamp:  5,
			Withdrawals: big.NewInt(42),
		},
	}
	err := historyDB.addBucketUpdates(historyDB.dbWrite, bucketUpdates)
	require.NoError(t, err)
	dbBucketUpdates, err := historyDB.GetAllBucketUpdates()
	require.NoError(t, err)
	assert.Equal(t, bucketUpdates, dbBucketUpdates)
}

func TestAddTokenExchanges(t *testing.T) {
	test.WipeDB(historyDB.DB())
	const fromBlock int64 = 1
	const toBlock int64 = 5 + 1
	setTestBlocks(fromBlock, toBlock)

	tokenExchanges := []common.TokenExchange{
		{
			EthBlockNum: 4,
			Address:     ethCommon.BigToAddress(big.NewInt(111)),
			ValueUSD:    12345,
		},
		{
			EthBlockNum: 5,
			Address:     ethCommon.BigToAddress(big.NewInt(222)),
			ValueUSD:    67890,
		},
	}
	err := historyDB.addTokenExchanges(historyDB.dbWrite, tokenExchanges)
	require.NoError(t, err)
	dbTokenExchanges, err := historyDB.GetAllTokenExchanges()
	require.NoError(t, err)
	assert.Equal(t, tokenExchanges, dbTokenExchanges)
}

func TestAddEscapeHatchWithdrawals(t *testing.T) {
	test.WipeDB(historyDB.DB())
	const fromBlock int64 = 1
	const toBlock int64 = 5 + 1
	setTestBlocks(fromBlock, toBlock)

	escapeHatchWithdrawals := []common.WDelayerEscapeHatchWithdrawal{
		{
			EthBlockNum: 4,
			Who:         ethCommon.BigToAddress(big.NewInt(111)),
			To:          ethCommon.BigToAddress(big.NewInt(222)),
			TokenAddr:   ethCommon.BigToAddress(big.NewInt(333)),
			Amount:      big.NewInt(10002),
		},
		{
			EthBlockNum: 5,
			Who:         ethCommon.BigToAddress(big.NewInt(444)),
			To:          ethCommon.BigToAddress(big.NewInt(555)),
			TokenAddr:   ethCommon.BigToAddress(big.NewInt(666)),
			Amount:      big.NewInt(20003),
		},
	}
	err := historyDB.addEscapeHatchWithdrawals(historyDB.dbWrite, escapeHatchWithdrawals)
	require.NoError(t, err)
	dbEscapeHatchWithdrawals, err := historyDB.GetAllEscapeHatchWithdrawals()
	require.NoError(t, err)
	assert.Equal(t, escapeHatchWithdrawals, dbEscapeHatchWithdrawals)
}

func TestGetMetricsAPI(t *testing.T) {
	test.WipeDB(historyDB.DB())
	set := `
		Type: Blockchain

		AddToken(1)

		CreateAccountDeposit(1) A: 1000 // numTx=1
		CreateAccountDeposit(1) B: 2000 // numTx=2
		CreateAccountDeposit(1) C: 3000 //numTx=3

		// block 0 is stored as default in the DB
		// block 1 does not exist
		> batchL1 // numBatches=1
		> batchL1 // numBatches=2
		> block // blockNum=2

		Transfer(1) C-A : 10 (1) // numTx=4
		> batch // numBatches=3
		> block // blockNum=3
		Transfer(1) B-C : 10 (1) // numTx=5
		> batch // numBatches=5
		> block // blockNum=4
		Transfer(1) A-B : 10 (1) // numTx=6
		> batch // numBatches=5
		> block // blockNum=5
		Transfer(1) A-B : 10 (1) // numTx=7
		> batch // numBatches=6
		> block // blockNum=6
	`

	const numBatches int = 6
	const numTx int = 7
	const blockNum = 6 - 1

	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set)
	require.NoError(t, err)
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

	// Sanity check
	require.Equal(t, blockNum, len(blocks))

	// Adding one batch per block
	// batch frequency can be chosen
	const frequency int = 15

	for i := range blocks {
		blocks[i].Block.Timestamp = time.Now().Add(-time.Second * time.Duration(frequency*(len(blocks)-i)))
		err = historyDB.AddBlockSCData(&blocks[i])
		assert.NoError(t, err)
	}

	res, _, err := historyDB.GetMetricsInternalAPI(common.BatchNum(numBatches))
	assert.NoError(t, err)

	assert.Equal(t, float64(numTx)/float64(numBatches), res.TransactionsPerBatch)

	// Frequency is not exactly the desired one, some decimals may appear
	// There is a -2 as time for first and last batch is not taken into account
	assert.InEpsilon(t, float64(frequency)*float64(numBatches-2)/float64(numBatches), res.BatchFrequency, 0.01)
	assert.InEpsilon(t, float64(numTx)/float64(frequency*blockNum-frequency), res.TransactionsPerSecond, 0.01)
	assert.Equal(t, int64(3), res.TokenAccounts)
	assert.Equal(t, int64(3), res.Wallets)
	// Til does not set fees
	assert.Equal(t, float64(0), res.AvgTransactionFee)
}

func TestGetMetricsAPIMoreThan24Hours(t *testing.T) {
	test.WipeDB(historyDB.DB())

	testUsersLen := 3
	var set []til.Instruction
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeCreateAccountDeposit,
			TokenID:       common.TokenID(0),
			DepositAmount: big.NewInt(1000000),
			Amount:        big.NewInt(0),
			From:          fmt.Sprintf("User%02d", user),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})

	// Transfers
	const numBlocks int = 30
	for x := 0; x < numBlocks; x++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeTransfer,
			TokenID:       common.TokenID(0),
			DepositAmount: big.NewInt(1),
			Amount:        big.NewInt(0),
			From:          "User00",
			To:            "User01",
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBatch})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}

	var chainID uint16 = 0
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocksFromInstructions(set)
	assert.NoError(t, err)

	tilCfgExtra := til.ConfigExtra{
		CoordUser: "A",
	}
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

	const numBatches = 2 + numBlocks
	const blockNum = 4 + numBlocks

	// Sanity check
	require.Equal(t, blockNum, len(blocks))

	// Adding one batch per block
	// batch frequency can be chosen
	const blockTime = 3600 * time.Second
	now := time.Now()
	require.NoError(t, err)

	for i := range blocks {
		blocks[i].Block.Timestamp = now.Add(-time.Duration(len(blocks)-1-i) * blockTime)
		err = historyDB.AddBlockSCData(&blocks[i])
		assert.NoError(t, err)
	}

	res, _, err := historyDBWithACC.GetMetricsInternalAPI(common.BatchNum(numBatches))
	assert.NoError(t, err)

	assert.InEpsilon(t, 1.0, res.TransactionsPerBatch, 0.1)

	assert.InEpsilon(t, res.BatchFrequency, float64(blockTime/time.Second), 0.1)
	assert.InEpsilon(t, 1.0/float64(blockTime/time.Second), res.TransactionsPerSecond, 0.1)
	assert.Equal(t, int64(3), res.TokenAccounts)
	assert.Equal(t, int64(3), res.Wallets)
	// Til does not set fees
	assert.Equal(t, float64(0), res.AvgTransactionFee)
}

func TestGetMetricsAPIEmpty(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, _, err := historyDBWithACC.GetMetricsInternalAPI(0)
	assert.NoError(t, err)
}

func TestGetLastL1TxsNum(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, err := historyDB.GetLastL1TxsNum()
	assert.NoError(t, err)
}

func TestGetLastTxsPosition(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, err := historyDB.GetLastTxsPosition(0)
	assert.Equal(t, sql.ErrNoRows.Error(), err.Error())
}

func TestGetFirstBatchBlockNumBySlot(t *testing.T) {
	test.WipeDB(historyDB.DB())

	set := `
		Type: Blockchain

		// Slot = 0

		> block // 2
		> block // 3
		> block // 4
		> block // 5

		// Slot = 1

		> block // 6
		> block // 7
		> batch
		> block // 8
		> block // 9

		// Slot = 2

		> batch
		> block // 10
		> block // 11
		> block // 12
		> block // 13

	`
	tc := til.NewContext(uint16(0), common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set)
	assert.NoError(t, err)

	tilCfgExtra := til.ConfigExtra{
		CoordUser: "A",
	}
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

	for i := range blocks {
		for j := range blocks[i].Rollup.Batches {
			blocks[i].Rollup.Batches[j].Batch.SlotNum = int64(i) / 4
		}
	}

	// Add all blocks
	for i := range blocks {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}

	_, err = historyDB.GetFirstBatchBlockNumBySlot(0)
	require.Equal(t, sql.ErrNoRows, tracerr.Unwrap(err))

	bn1, err := historyDB.GetFirstBatchBlockNumBySlot(1)
	require.NoError(t, err)
	assert.Equal(t, int64(8), bn1)

	bn2, err := historyDB.GetFirstBatchBlockNumBySlot(2)
	require.NoError(t, err)
	assert.Equal(t, int64(10), bn2)
}

func TestTxItemID(t *testing.T) {
	test.WipeDB(historyDB.DB())
	testUsersLen := 10
	var set []til.Instruction
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeCreateAccountDeposit,
			TokenID:       common.TokenID(0),
			DepositAmount: big.NewInt(1000000),
			Amount:        big.NewInt(0),
			From:          fmt.Sprintf("User%02d", user),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeDeposit,
			TokenID:       common.TokenID(0),
			DepositAmount: big.NewInt(100000),
			Amount:        big.NewInt(0),
			From:          fmt.Sprintf("User%02d", user),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeDepositTransfer,
			TokenID:       common.TokenID(0),
			DepositAmount: big.NewInt(10000 * int64(user+1)),
			Amount:        big.NewInt(1000 * int64(user+1)),
			From:          fmt.Sprintf("User%02d", user),
			To:            fmt.Sprintf("User%02d", (user+1)%testUsersLen),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeForceTransfer,
			TokenID:       common.TokenID(0),
			Amount:        big.NewInt(100 * int64(user+1)),
			DepositAmount: big.NewInt(0),
			From:          fmt.Sprintf("User%02d", user),
			To:            fmt.Sprintf("User%02d", (user+1)%testUsersLen),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	for user := 0; user < testUsersLen; user++ {
		set = append(set, til.Instruction{
			Typ:           common.TxTypeForceExit,
			TokenID:       common.TokenID(0),
			Amount:        big.NewInt(10 * int64(user+1)),
			DepositAmount: big.NewInt(0),
			From:          fmt.Sprintf("User%02d", user),
		})
		set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})
	var chainID uint16 = 0
	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocksFromInstructions(set)
	assert.NoError(t, err)

	tilCfgExtra := til.ConfigExtra{
		CoordUser: "A",
	}
	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)

	// Add all blocks
	for i := range blocks {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.NoError(t, err)
	}

	txs, err := historyDB.GetAllL1UserTxs()
	require.NoError(t, err)
	position := 0
	for _, tx := range txs {
		if tx.Position == 0 {
			position = 0
		}
		assert.Equal(t, position, tx.Position)
		position++
	}
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

func TestNodeInfo(t *testing.T) {
	test.WipeDB(historyDB.DB())

	err := historyDB.SetStateInternalAPI(&StateAPI{})
	require.NoError(t, err)

	clientSetup := test.NewClientSetupExample()
	constants := &Constants{
		SCConsts: common.SCConsts{
			Rollup:   *clientSetup.RollupConstants,
			Auction:  *clientSetup.AuctionConstants,
			WDelayer: *clientSetup.WDelayerConstants,
		},
		ChainID:       42,
		HermezAddress: clientSetup.AuctionConstants.HermezRollup,
	}
	err = historyDB.SetConstants(constants)
	require.NoError(t, err)

	// Test parameters
	var f64 = 1.2
	var i64 int64 = 8888
	addr := ethCommon.HexToAddress("0x1234")
	hash := ethCommon.HexToHash("0x5678")
	stateAPI := &StateAPI{
		NodePublicInfo: NodePublicInfo{
			ForgeDelay: 3.1,
		},
		Network: NetworkAPI{
			LastEthBlock:  12,
			LastSyncBlock: 34,
			LastBatch: &BatchAPI{
				ItemID:       123,
				BatchNum:     456,
				EthBlockNum:  789,
				EthBlockHash: hash,
				Timestamp:    time.Now(),
				ForgerAddr:   addr,
				// CollectedFeesDB: map[common.TokenID]*big.Int{
				// 	0: big.NewInt(11111),
				// 	1: big.NewInt(21111),
				// 	2: big.NewInt(31111),
				// },
				CollectedFeesAPI: apitypes.CollectedFeesAPI(map[common.TokenID]apitypes.BigIntStr{
					0: apitypes.BigIntStr("11111"),
					1: apitypes.BigIntStr("21111"),
					2: apitypes.BigIntStr("31111"),
				}),
				TotalFeesUSD:  &f64,
				StateRoot:     apitypes.BigIntStr("1234"),
				NumAccounts:   11,
				ExitRoot:      apitypes.BigIntStr("5678"),
				ForgeL1TxsNum: &i64,
				SlotNum:       44,
				ForgedTxs:     23,
				TotalItems:    0,
				FirstItem:     0,
				LastItem:      0,
			},
			CurrentSlot: 22,
			NextForgers: []NextForgerAPI{
				{
					Coordinator: CoordinatorAPI{
						ItemID:      111,
						Bidder:      addr,
						Forger:      addr,
						EthBlockNum: 566,
						URL:         "asd",
						TotalItems:  0,
						FirstItem:   0,
						LastItem:    0,
					},
					Period: Period{
						SlotNum:       33,
						FromBlock:     55,
						ToBlock:       66,
						FromTimestamp: time.Now(),
						ToTimestamp:   time.Now(),
					},
				},
			},
		},
		Metrics: MetricsAPI{
			TransactionsPerBatch: 1.1,
			TokenAccounts:        42,
		},
		Rollup:            *NewRollupVariablesAPI(clientSetup.RollupVariables),
		Auction:           *NewAuctionVariablesAPI(clientSetup.AuctionVariables),
		WithdrawalDelayer: *clientSetup.WDelayerVariables,
		RecommendedFee: common.RecommendedFee{
			ExistingAccount: 0.15,
		},
	}
	err = historyDB.SetStateInternalAPI(stateAPI)
	require.NoError(t, err)

	nodeConfig := &NodeConfig{
		MaxPoolTxs: 123,
		MinFeeUSD:  0.5,
	}
	err = historyDB.SetNodeConfig(nodeConfig)
	require.NoError(t, err)

	dbConstants, err := historyDB.GetConstants()
	require.NoError(t, err)
	assert.Equal(t, constants, dbConstants)

	dbNodeConfig, err := historyDB.GetNodeConfig()
	require.NoError(t, err)
	assert.Equal(t, nodeConfig, dbNodeConfig)

	dbStateAPI, err := historyDB.getStateAPI(historyDB.dbRead)
	require.NoError(t, err)
	assert.Equal(t, stateAPI.Network.LastBatch.Timestamp.Unix(),
		dbStateAPI.Network.LastBatch.Timestamp.Unix())
	dbStateAPI.Network.LastBatch.Timestamp = stateAPI.Network.LastBatch.Timestamp
	assert.Equal(t, stateAPI.Network.NextForgers[0].Period.FromTimestamp.Unix(),
		dbStateAPI.Network.NextForgers[0].Period.FromTimestamp.Unix())
	dbStateAPI.Network.NextForgers[0].Period.FromTimestamp = stateAPI.Network.NextForgers[0].Period.FromTimestamp
	assert.Equal(t, stateAPI.Network.NextForgers[0].Period.ToTimestamp.Unix(),
		dbStateAPI.Network.NextForgers[0].Period.ToTimestamp.Unix())
	dbStateAPI.Network.NextForgers[0].Period.ToTimestamp = stateAPI.Network.NextForgers[0].Period.ToTimestamp
	assert.Equal(t, stateAPI, dbStateAPI)
}
