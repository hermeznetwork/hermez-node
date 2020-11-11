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
	fetchedAccs, err := historyDB.GetAccounts()
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
	// TODO: Generate batches using til (and blocks for foreign key)
	set := `
		Type: Blockchain

		// Things to test:
		// One tx of each type
		// batches that forge user L1s
		// historic USD is not set if USDUpdate is too old (24h)
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

	/*

		OLD TEST


		const fromBlock int64 = 1
		const toBlock int64 = 5
		// Prepare blocks in the DB
		blocks := setTestBlocks(fromBlock, toBlock)
		// Generate fake tokens
		const nTokens = 500
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

			Uncomment once the transaction generation is fixed
			!! test that batches that forge user L1s !!
			!! Missing tests to check that  !!

			// Generate fake L1 txs
			const nL1s = 64
			_, l1txs := test.GenL1Txs(256, nL1s, 0, nil, accs, tokens, blocks, batches)
			err = historyDB.AddL1Txs(l1txs)
			assert.NoError(t, err)
			// Generate fake L2 txs
			const nL2s = 2048 - nL1s
			_, l2txs := test.GenL2Txs(256, nL2s, 0, nil, accs, tokens, blocks, batches)
			err = historyDB.AddL2Txs(l2txs)
			assert.NoError(t, err)
			// Compare fetched txs vs generated txs.
			fetchAndAssertTxs(t, l1txs, l2txs)
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
			// Test trigger: forge L1 txs
			// add next batch to DB
			batchNum, toForgeL1TxsNum := test.GetNextToForgeNumAndBatch(batches)
			batch := batches[0]
			batch.BatchNum = batchNum
			batch.ForgeL1TxsNum = toForgeL1TxsNum
			assert.NoError(t, historyDB.AddBatch(&batch)) // This should update nL1s / 2 rows
			// Set batch num in txs that should have been marked as forged in the DB
			for i := 0; i < len(l1txs); i++ {
				fetchedTx, err := historyDB.GetTx(l1txs[i].TxID)
				assert.NoError(t, err)
				if l1txs[i].ToForgeL1TxsNum == toForgeL1TxsNum {
					assert.Equal(t, batchNum, *fetchedTx.BatchNum)
				} else {
					if fetchedTx.BatchNum != nil {
						assert.NotEqual(t, batchNum, *fetchedTx.BatchNum)
					}
				}
			}

			// Test helper functions for Synchronizer
			// GetLastTxsPosition
			expectedPosition := -1
			var choosenToForgeL1TxsNum int64 = -1
			for _, tx := range l1txs {
				if choosenToForgeL1TxsNum == -1 && tx.ToForgeL1TxsNum > 0 {
					choosenToForgeL1TxsNum = tx.ToForgeL1TxsNum
					expectedPosition = tx.Position
				} else if choosenToForgeL1TxsNum == tx.ToForgeL1TxsNum && expectedPosition < tx.Position {
					expectedPosition = tx.Position
				}
			}
			position, err := historyDB.GetLastTxsPosition(choosenToForgeL1TxsNum)
			assert.NoError(t, err)
			assert.Equal(t, expectedPosition, position)

			// GetL1UserTxs: not needed? tests were broken
			// txs, err := historyDB.GetL1UserTxs(2)
			// assert.NoError(t, err)
			// assert.NotZero(t, len(txs))
			// assert.NoError(t, err)
			// assert.Equal(t, 22, position)
			// // Test Update L1 TX Batch_num
			// assert.Equal(t, common.BatchNum(0), txs[0].BatchNum)
			// txs[0].BatchNum = common.BatchNum(1)
			// txs, err = historyDB.GetL1UserTxs(2)
			// assert.NoError(t, err)
			// assert.NotZero(t, len(txs))
			// assert.Equal(t, common.BatchNum(1), txs[0].BatchNum)
	*/
}

/*
func fetchAndAssertTxs(t *testing.T, l1txs []common.L1Tx, l2txs []common.L2Tx) {
	for i := 0; i < len(l1txs); i++ {
		tx := l1txs[i].Tx()
		fmt.Println("ASDF", i, tx.TxID)
		fetchedTx, err := historyDB.GetTx(tx.TxID)
		require.NoError(t, err)
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
}
*/

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
	exitTree := test.GenExitTree(nBatches, batches, accs)
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
	// fmt.Printf("DBG Blocks: %+v\n", blocks)

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

func TestSetInitialSCVars(t *testing.T) {
	test.WipeDB(historyDB.DB())
	_, _, _, err := historyDB.GetSCVars()
	assert.Equal(t, sql.ErrNoRows, err)

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

		Exit(1) C: 50 (200)
		Exit(1) D: 30 (200)

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{3}
		> batchL1 // forge L1UserTxs{3}, freeze defined L1UserTxs{nil}
		> block // blockNum=3

		> block // blockNum=4 (empty block)
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

	// Add all blocks except for the last one
	for i := range blocks[:len(blocks)-1] {
		err = historyDB.AddBlockSCData(&blocks[i])
		require.Nil(t, err)
	}

	// Add withdraws to the last block, and insert block into the DB
	block := &blocks[len(blocks)-1]
	require.Equal(t, int64(4), block.Block.EthBlockNum)
	block.Rollup.Withdrawals = append(block.Rollup.Withdrawals,
		common.WithdrawInfo{Idx: 256, NumExitRoot: 4, InstantWithdraw: true},
		common.WithdrawInfo{Idx: 257, NumExitRoot: 4, InstantWithdraw: false},
		common.WithdrawInfo{Idx: 258, NumExitRoot: 3, InstantWithdraw: true},
		common.WithdrawInfo{Idx: 259, NumExitRoot: 3, InstantWithdraw: false},
	)
	err = historyDB.addBlock(historyDB.db, &block.Block)
	require.Nil(t, err)

	// update exit trees in DB
	instantWithdrawn := []exitID{}
	delayedWithdrawRequest := []exitID{}
	for _, withdraw := range block.Rollup.Withdrawals {
		exitID := exitID{
			batchNum: int64(withdraw.NumExitRoot),
			idx:      int64(withdraw.Idx),
		}
		if withdraw.InstantWithdraw {
			instantWithdrawn = append(instantWithdrawn, exitID)
		} else {
			delayedWithdrawRequest = append(delayedWithdrawRequest, exitID)
		}
	}
	err = historyDB.updateExitTree(historyDB.db, block.Block.EthBlockNum, instantWithdrawn, delayedWithdrawRequest)
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
