package synchronizer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tokenConsts = map[common.TokenID]eth.ERC20Consts{}

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

func accountsCmp(accounts []common.Account) func(i, j int) bool {
	return func(i, j int) bool { return accounts[i].Idx < accounts[j].Idx }
}

// Check Sync output and HistoryDB state against expected values generated by
// til
func checkSyncBlock(t *testing.T, s *Synchronizer, blockNum int, block, syncBlock *common.BlockData) {
	// Check Blocks
	dbBlocks, err := s.historyDB.GetAllBlocks()
	require.NoError(t, err)
	dbBlocks = dbBlocks[1:] // ignore block 0, added by default in the DB
	assert.Equal(t, blockNum, len(dbBlocks))
	assert.Equal(t, int64(blockNum), dbBlocks[blockNum-1].Num)
	assert.NotEqual(t, dbBlocks[blockNum-1].Hash, dbBlocks[blockNum-2].Hash)
	assert.Greater(t, dbBlocks[blockNum-1].Timestamp.Unix(), dbBlocks[blockNum-2].Timestamp.Unix())

	// Check Tokens
	assert.Equal(t, len(block.Rollup.AddedTokens), len(syncBlock.Rollup.AddedTokens))
	dbTokens, err := s.historyDB.GetAllTokens()
	require.NoError(t, err)
	dbTokens = dbTokens[1:] // ignore token 0, added by default in the DB
	for i, token := range block.Rollup.AddedTokens {
		dbToken := dbTokens[i]
		syncToken := syncBlock.Rollup.AddedTokens[i]

		assert.Equal(t, block.Block.Num, syncToken.EthBlockNum)
		assert.Equal(t, token.TokenID, syncToken.TokenID)
		assert.Equal(t, token.EthAddr, syncToken.EthAddr)
		tokenConst := tokenConsts[token.TokenID]
		assert.Equal(t, tokenConst.Name, syncToken.Name)
		assert.Equal(t, tokenConst.Symbol, syncToken.Symbol)
		assert.Equal(t, tokenConst.Decimals, syncToken.Decimals)

		var tokenCpy historydb.TokenWithUSD
		//nolint:gosec
		require.Nil(t, copier.Copy(&tokenCpy, &token))      // copy common.Token to historydb.TokenWithUSD
		require.Nil(t, copier.Copy(&tokenCpy, &tokenConst)) // copy common.Token to historydb.TokenWithUSD
		tokenCpy.ItemID = dbToken.ItemID                    // we don't care about ItemID
		assert.Equal(t, tokenCpy, dbToken)
	}

	// Check submitted L1UserTxs
	assert.Equal(t, len(block.Rollup.L1UserTxs), len(syncBlock.Rollup.L1UserTxs))
	dbL1UserTxs, err := s.historyDB.GetAllL1UserTxs()
	require.NoError(t, err)
	// Ignore BatchNum in syncBlock.L1UserTxs because this value is set by
	// the HistoryDB. Also ignore EffectiveAmount & EffectiveDepositAmount
	// because this value is set by StateDB.ProcessTxs.
	for i := range syncBlock.Rollup.L1UserTxs {
		syncBlock.Rollup.L1UserTxs[i].BatchNum = block.Rollup.L1UserTxs[i].BatchNum
		assert.Nil(t, syncBlock.Rollup.L1UserTxs[i].EffectiveDepositAmount)
		assert.Nil(t, syncBlock.Rollup.L1UserTxs[i].EffectiveAmount)
	}
	assert.Equal(t, block.Rollup.L1UserTxs, syncBlock.Rollup.L1UserTxs)
	for _, tx := range block.Rollup.L1UserTxs {
		var dbTx *common.L1Tx
		// Find tx in DB output
		for _, _dbTx := range dbL1UserTxs {
			if *tx.ToForgeL1TxsNum == *_dbTx.ToForgeL1TxsNum &&
				tx.Position == _dbTx.Position {
				dbTx = new(common.L1Tx)
				*dbTx = _dbTx
				break
			}
		}
		// If the tx has been forged in this block, this will be
		// reflected in the DB, and so the Effective values will be
		// already set
		if dbTx.BatchNum != nil {
			tx.EffectiveAmount = tx.Amount
			tx.EffectiveDepositAmount = tx.DepositAmount
		}
		assert.Equal(t, &tx, dbTx) //nolint:gosec
	}

	// Check Batches
	assert.Equal(t, len(block.Rollup.Batches), len(syncBlock.Rollup.Batches))
	dbBatches, err := s.historyDB.GetAllBatches()
	require.NoError(t, err)

	dbL1CoordinatorTxs, err := s.historyDB.GetAllL1CoordinatorTxs()
	require.NoError(t, err)
	dbL2Txs, err := s.historyDB.GetAllL2Txs()
	require.NoError(t, err)
	dbExits, err := s.historyDB.GetAllExits()
	require.NoError(t, err)
	// dbL1CoordinatorTxs := []common.L1Tx{}
	for i, batch := range block.Rollup.Batches {
		var dbBatch *common.Batch
		// Find batch in DB output
		for _, _dbBatch := range dbBatches {
			if batch.Batch.BatchNum == _dbBatch.BatchNum {
				dbBatch = new(common.Batch)
				*dbBatch = _dbBatch
				break
			}
		}
		syncBatch := syncBlock.Rollup.Batches[i]

		// We don't care about TotalFeesUSD.  Use the syncBatch that
		// has a TotalFeesUSD inserted by the HistoryDB
		batch.Batch.TotalFeesUSD = syncBatch.Batch.TotalFeesUSD
		assert.Equal(t, batch.CreatedAccounts, syncBatch.CreatedAccounts)
		batch.Batch.NumAccounts = len(batch.CreatedAccounts)

		// Test field by field to facilitate debugging of errors
		assert.Equal(t, batch.L1UserTxs, syncBatch.L1UserTxs)
		assert.Equal(t, batch.L1CoordinatorTxs, syncBatch.L1CoordinatorTxs)
		assert.Equal(t, batch.L2Txs, syncBatch.L2Txs)
		// In exit tree, we only check AccountIdx and Balance, because
		// it's what we have precomputed before.
		require.Equal(t, len(batch.ExitTree), len(syncBatch.ExitTree))
		for j := range batch.ExitTree {
			exit := &batch.ExitTree[j]
			assert.Equal(t, exit.AccountIdx, syncBatch.ExitTree[j].AccountIdx)
			assert.Equal(t, exit.Balance, syncBatch.ExitTree[j].Balance)
			*exit = syncBatch.ExitTree[j]
		}
		assert.Equal(t, batch.Batch, syncBatch.Batch)
		assert.Equal(t, batch, syncBatch)
		assert.Equal(t, &batch.Batch, dbBatch) //nolint:gosec

		// Check forged L1UserTxs from DB, and check effective values
		// in sync output
		for j, tx := range batch.L1UserTxs {
			var dbTx *common.L1Tx
			// Find tx in DB output
			for _, _dbTx := range dbL1UserTxs {
				if *tx.BatchNum == *_dbTx.BatchNum &&
					tx.Position == _dbTx.Position {
					dbTx = new(common.L1Tx)
					*dbTx = _dbTx
					break
				}
			}
			assert.Equal(t, &tx, dbTx) //nolint:gosec

			syncTx := &syncBlock.Rollup.Batches[i].L1UserTxs[j]
			assert.Equal(t, syncTx.DepositAmount, syncTx.EffectiveDepositAmount)
			assert.Equal(t, syncTx.Amount, syncTx.EffectiveAmount)
		}

		// Check L1CoordinatorTxs from DB
		for _, tx := range batch.L1CoordinatorTxs {
			var dbTx *common.L1Tx
			// Find tx in DB output
			for _, _dbTx := range dbL1CoordinatorTxs {
				if *tx.BatchNum == *_dbTx.BatchNum &&
					tx.Position == _dbTx.Position {
					dbTx = new(common.L1Tx)
					*dbTx = _dbTx
					break
				}
			}
			assert.Equal(t, &tx, dbTx) //nolint:gosec
		}

		// Check L2Txs from DB
		for _, tx := range batch.L2Txs {
			var dbTx *common.L2Tx
			// Find tx in DB output
			for _, _dbTx := range dbL2Txs {
				if tx.BatchNum == _dbTx.BatchNum &&
					tx.Position == _dbTx.Position {
					dbTx = new(common.L2Tx)
					*dbTx = _dbTx
					break
				}
			}
			assert.Equal(t, &tx, dbTx) //nolint:gosec
		}

		// Check Exits from DB
		for _, exit := range batch.ExitTree {
			var dbExit *common.ExitInfo
			// Find exit in DB output
			for _, _dbExit := range dbExits {
				if exit.BatchNum == _dbExit.BatchNum &&
					exit.AccountIdx == _dbExit.AccountIdx {
					dbExit = new(common.ExitInfo)
					*dbExit = _dbExit
					break
				}
			}
			// Compare MerkleProof in JSON because unmarshaled 0
			// big.Int leaves the internal big.Int array at nil,
			// and gives trouble when comparing big.Int with
			// internal big.Int array != nil but empty.
			mtp, err := json.Marshal(exit.MerkleProof)
			require.NoError(t, err)
			dbMtp, err := json.Marshal(dbExit.MerkleProof)
			require.NoError(t, err)
			assert.Equal(t, mtp, dbMtp)
			dbExit.MerkleProof = exit.MerkleProof
			assert.Equal(t, &exit, dbExit) //nolint:gosec
		}
	}

	// Compare accounts from HistoryDB with StateDB (they should match)
	dbAccounts, err := s.historyDB.GetAllAccounts()
	require.NoError(t, err)
	sdbAccounts, err := s.stateDB.GetAccounts()
	require.NoError(t, err)
	assertEqualAccountsHistoryDBStateDB(t, dbAccounts, sdbAccounts)
}

func assertEqualAccountsHistoryDBStateDB(t *testing.T, hdbAccs, sdbAccs []common.Account) {
	assert.Equal(t, len(hdbAccs), len(sdbAccs))
	sort.SliceStable(hdbAccs, accountsCmp(hdbAccs))
	sort.SliceStable(sdbAccs, accountsCmp(sdbAccs))
	for i := range hdbAccs {
		hdbAcc := hdbAccs[i]
		sdbAcc := sdbAccs[i]
		assert.Equal(t, hdbAcc.Idx, sdbAcc.Idx)
		assert.Equal(t, hdbAcc.TokenID, sdbAcc.TokenID)
		assert.Equal(t, hdbAcc.EthAddr, sdbAcc.EthAddr)
		assert.Equal(t, hdbAcc.PublicKey, sdbAcc.PublicKey)
	}
}

// ethAddTokens adds the tokens from the blocks to the blockchain
func ethAddTokens(blocks []common.BlockData, client *test.Client) {
	for _, block := range blocks {
		for _, token := range block.Rollup.AddedTokens {
			consts := eth.ERC20Consts{
				Name:     fmt.Sprintf("Token %d", token.TokenID),
				Symbol:   fmt.Sprintf("TK%d", token.TokenID),
				Decimals: 18,
			}
			tokenConsts[token.TokenID] = consts
			client.CtlAddERC20(token.EthAddr, consts)
		}
	}
}

func TestSync(t *testing.T) {
	//
	// Setup
	//

	ctx := context.Background()
	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	stateDB, err := statedb.NewStateDB(dir, statedb.TypeSynchronizer, 32)
	require.NoError(t, err)

	// Init History DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.NoError(t, err)
	historyDB := historydb.NewHistoryDB(db)
	// Clear DB
	test.WipeDB(historyDB.DB())

	// Init eth client
	var timer timer
	clientSetup := test.NewClientSetupExample()
	bootCoordAddr := clientSetup.AuctionVariables.BootCoordinator
	client := test.NewClient(true, &timer, &ethCommon.Address{}, clientSetup)

	// Create Synchronizer
	s, err := NewSynchronizer(client, historyDB, stateDB, Config{
		StartBlockNum: ConfigStartBlockNum{
			Rollup:   1,
			Auction:  1,
			WDelayer: 1,
		},
		InitialVariables: SCVariables{
			Rollup:   *clientSetup.RollupVariables,
			Auction:  *clientSetup.AuctionVariables,
			WDelayer: *clientSetup.WDelayerVariables,
		},
	})
	require.NoError(t, err)

	//
	// First Sync from an initial state
	//
	stats := s.Stats()
	assert.Equal(t, false, stats.Synced())

	// Test Sync for rollup genesis block
	syncBlock, discards, err := s.Sync2(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	require.Nil(t, syncBlock.Rollup.Vars)
	require.Nil(t, syncBlock.Auction.Vars)
	require.Nil(t, syncBlock.WDelayer.Vars)
	assert.Equal(t, int64(1), syncBlock.Block.Num)
	stats = s.Stats()
	assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
	assert.Equal(t, int64(1), stats.Eth.LastBlock.Num)
	assert.Equal(t, int64(1), stats.Sync.LastBlock.Num)
	vars := s.SCVars()
	assert.Equal(t, clientSetup.RollupVariables, vars.Rollup)
	assert.Equal(t, clientSetup.AuctionVariables, vars.Auction)
	assert.Equal(t, clientSetup.WDelayerVariables, vars.WDelayer)

	dbBlocks, err := s.historyDB.GetAllBlocks()
	require.NoError(t, err)
	assert.Equal(t, 2, len(dbBlocks))
	assert.Equal(t, int64(1), dbBlocks[1].Num)

	// Sync again and expect no new blocks
	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.Nil(t, syncBlock)

	//
	// Generate blockchain and smart contract data, and fill the test smart contracts
	//

	// Generate blockchain data with til
	set1 := `
		Type: Blockchain

		AddToken(1)
		AddToken(2)
		AddToken(3)

		CreateAccountDeposit(1) C: 2000 // Idx=256+2=258
		CreateAccountDeposit(2) A: 2000 // Idx=256+3=259
		CreateAccountDeposit(1) D: 500  // Idx=256+4=260
		CreateAccountDeposit(2) B: 500  // Idx=256+5=261
		CreateAccountDeposit(2) C: 500  // Idx=256+6=262

		CreateAccountCoordinator(1) A // Idx=256+0=256
		CreateAccountCoordinator(1) B // Idx=256+1=257

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{5}
		> batchL1 // forge defined L1UserTxs{5}, freeze L1UserTxs{nil}
		> block // blockNum=2

		CreateAccountDepositTransfer(1) E-A: 1000, 200 // Idx=256+7=263
		ForceTransfer(1) C-B: 80
		ForceExit(1) A: 100
		ForceExit(1) B: 80
		ForceTransfer(1) A-D: 100

		Transfer(1) C-A: 100 (200)
		Exit(1) C: 50 (200)
		Exit(1) D: 30 (200)

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{3}
		> batchL1 // forge L1UserTxs{3}, freeze defined L1UserTxs{nil}
		> block // blockNum=3
	`
	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: bootCoordAddr,
		CoordUser:     "A",
	}
	blocks, err := tc.GenerateBlocks(set1)
	require.NoError(t, err)
	// Sanity check
	require.Equal(t, 2, len(blocks))
	// blocks 0 (blockNum=2)
	i := 0
	require.Equal(t, 2, int(blocks[i].Block.Num))
	require.Equal(t, 3, len(blocks[i].Rollup.AddedTokens))
	require.Equal(t, 5, len(blocks[i].Rollup.L1UserTxs))
	require.Equal(t, 2, len(blocks[i].Rollup.Batches))
	require.Equal(t, 2, len(blocks[i].Rollup.Batches[0].L1CoordinatorTxs))
	// blocks 1 (blockNum=3)
	i = 1
	require.Equal(t, 3, int(blocks[i].Block.Num))
	require.Equal(t, 5, len(blocks[i].Rollup.L1UserTxs))
	require.Equal(t, 2, len(blocks[i].Rollup.Batches))
	require.Equal(t, 3, len(blocks[i].Rollup.Batches[0].L2Txs))

	// Generate extra required data
	ethAddTokens(blocks, client)

	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	tc.FillBlocksL1UserTxsBatchNum(blocks)
	err = tc.FillBlocksForgedL1UserTxs(blocks)
	require.NoError(t, err)

	// Add block data to the smart contracts
	err = client.CtlAddBlocks(blocks)
	require.NoError(t, err)

	//
	// Sync to synchronize the current state from the test smart contracts,
	// and check the outcome
	//

	// Block 2

	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.Nil(t, syncBlock.Rollup.Vars)
	assert.Nil(t, syncBlock.Auction.Vars)
	assert.Nil(t, syncBlock.WDelayer.Vars)
	assert.Equal(t, int64(2), syncBlock.Block.Num)
	stats = s.Stats()
	assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
	assert.Equal(t, int64(3), stats.Eth.LastBlock.Num)
	assert.Equal(t, int64(2), stats.Sync.LastBlock.Num)

	checkSyncBlock(t, s, 2, &blocks[0], syncBlock)

	// Block 3

	syncBlock, discards, err = s.Sync2(ctx, nil)
	assert.NoError(t, err)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.Nil(t, syncBlock.Rollup.Vars)
	assert.Nil(t, syncBlock.Auction.Vars)
	assert.Nil(t, syncBlock.WDelayer.Vars)
	assert.Equal(t, int64(3), syncBlock.Block.Num)
	stats = s.Stats()
	assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
	assert.Equal(t, int64(3), stats.Eth.LastBlock.Num)
	assert.Equal(t, int64(3), stats.Sync.LastBlock.Num)

	checkSyncBlock(t, s, 3, &blocks[1], syncBlock)

	// Block 4
	// Generate 2 withdraws manually
	_, err = client.RollupWithdrawMerkleProof(tc.Users["A"].BJJ.Public(), 1, 4, 256, big.NewInt(100), []*big.Int{}, true)
	require.NoError(t, err)
	_, err = client.RollupWithdrawMerkleProof(tc.Users["C"].BJJ.Public(), 1, 3, 258, big.NewInt(50), []*big.Int{}, false)
	require.NoError(t, err)
	client.CtlMineBlock()

	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.Nil(t, syncBlock.Rollup.Vars)
	assert.Nil(t, syncBlock.Auction.Vars)
	assert.Nil(t, syncBlock.WDelayer.Vars)
	assert.Equal(t, int64(4), syncBlock.Block.Num)
	stats = s.Stats()
	assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
	assert.Equal(t, int64(4), stats.Eth.LastBlock.Num)
	assert.Equal(t, int64(4), stats.Sync.LastBlock.Num)
	vars = s.SCVars()
	assert.Equal(t, clientSetup.RollupVariables, vars.Rollup)
	assert.Equal(t, clientSetup.AuctionVariables, vars.Auction)
	assert.Equal(t, clientSetup.WDelayerVariables, vars.WDelayer)

	dbExits, err := s.historyDB.GetAllExits()
	require.NoError(t, err)
	foundA1, foundC1 := false, false
	for _, exit := range dbExits {
		if exit.AccountIdx == 256 && exit.BatchNum == 4 {
			foundA1 = true
			assert.Equal(t, int64(4), *exit.InstantWithdrawn)
		}
		if exit.AccountIdx == 258 && exit.BatchNum == 3 {
			foundC1 = true
			assert.Equal(t, int64(4), *exit.DelayedWithdrawRequest)
		}
	}
	assert.True(t, foundA1)
	assert.True(t, foundC1)

	// Block 5
	// Update variables manually
	rollupVars, auctionVars, wDelayerVars, err := s.historyDB.GetSCVars()
	require.NoError(t, err)
	rollupVars.ForgeL1L2BatchTimeout = 42
	_, err = client.RollupUpdateForgeL1L2BatchTimeout(rollupVars.ForgeL1L2BatchTimeout)
	require.NoError(t, err)

	auctionVars.OpenAuctionSlots = 17
	_, err = client.AuctionSetOpenAuctionSlots(auctionVars.OpenAuctionSlots)
	require.NoError(t, err)

	wDelayerVars.WithdrawalDelay = 99
	_, err = client.WDelayerChangeWithdrawalDelay(wDelayerVars.WithdrawalDelay)
	require.NoError(t, err)

	client.CtlMineBlock()

	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.NoError(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.NotNil(t, syncBlock.Rollup.Vars)
	assert.NotNil(t, syncBlock.Auction.Vars)
	assert.NotNil(t, syncBlock.WDelayer.Vars)
	assert.Equal(t, int64(5), syncBlock.Block.Num)
	stats = s.Stats()
	assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
	assert.Equal(t, int64(5), stats.Eth.LastBlock.Num)
	assert.Equal(t, int64(5), stats.Sync.LastBlock.Num)
	vars = s.SCVars()
	assert.NotEqual(t, clientSetup.RollupVariables, vars.Rollup)
	assert.NotEqual(t, clientSetup.AuctionVariables, vars.Auction)
	assert.NotEqual(t, clientSetup.WDelayerVariables, vars.WDelayer)

	dbRollupVars, dbAuctionVars, dbWDelayerVars, err := s.historyDB.GetSCVars()
	require.NoError(t, err)
	// Set EthBlockNum for Vars to the blockNum in which they were updated (should be 5)
	rollupVars.EthBlockNum = syncBlock.Block.Num
	auctionVars.EthBlockNum = syncBlock.Block.Num
	wDelayerVars.EthBlockNum = syncBlock.Block.Num
	assert.Equal(t, rollupVars, dbRollupVars)
	assert.Equal(t, auctionVars, dbAuctionVars)
	assert.Equal(t, wDelayerVars, dbWDelayerVars)

	//
	// Reorg test
	//

	// Redo blocks 2-5 (as a reorg) only leaving:
	// - 2 create account transactions
	// - 2 add tokens
	// We add a 6th block so that the synchronizer can detect the reorg
	set2 := `
		Type: Blockchain

		AddToken(1)
		AddToken(2)

		CreateAccountDeposit(1) C: 2000 // Idx=256+1=257

		CreateAccountCoordinator(1) A // Idx=256+0=256

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs{1}
		> batchL1 // forge defined L1UserTxs{1}, freeze L1UserTxs{nil}
		> block // blockNum=2
		> block // blockNum=3
		> block // blockNum=4
		> block // blockNum=5
		> block // blockNum=6
	`
	tc = til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra = til.ConfigExtra{
		BootCoordAddr: bootCoordAddr,
		CoordUser:     "A",
	}
	blocks, err = tc.GenerateBlocks(set2)
	require.NoError(t, err)

	for i := 0; i < 4; i++ {
		client.CtlRollback()
	}
	block := client.CtlLastBlock()
	require.Equal(t, int64(1), block.Num)

	// Generate extra required data
	ethAddTokens(blocks, client)

	err = tc.FillBlocksExtra(blocks, &tilCfgExtra)
	require.NoError(t, err)
	tc.FillBlocksL1UserTxsBatchNum(blocks)

	// Add block data to the smart contracts
	err = client.CtlAddBlocks(blocks)
	require.NoError(t, err)

	// First sync detects the reorg and discards 4 blocks
	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.NoError(t, err)
	expetedDiscards := int64(4)
	require.Equal(t, &expetedDiscards, discards)
	require.Nil(t, syncBlock)
	stats = s.Stats()
	assert.Equal(t, false, stats.Synced())
	assert.Equal(t, int64(6), stats.Eth.LastBlock.Num)
	vars = s.SCVars()
	assert.Equal(t, clientSetup.RollupVariables, vars.Rollup)
	assert.Equal(t, clientSetup.AuctionVariables, vars.Auction)
	assert.Equal(t, clientSetup.WDelayerVariables, vars.WDelayer)

	// At this point, the DB only has data up to block 1
	dbBlock, err := s.historyDB.GetLastBlock()
	require.NoError(t, err)
	assert.Equal(t, int64(1), dbBlock.Num)

	// Accounts in HistoryDB and StateDB must be empty
	dbAccounts, err := s.historyDB.GetAllAccounts()
	require.NoError(t, err)
	sdbAccounts, err := s.stateDB.GetAccounts()
	require.NoError(t, err)
	assert.Equal(t, 0, len(dbAccounts))
	assertEqualAccountsHistoryDBStateDB(t, dbAccounts, sdbAccounts)

	// Sync blocks 2-6
	for i := 0; i < 5; i++ {
		syncBlock, discards, err = s.Sync2(ctx, nil)
		require.NoError(t, err)
		require.Nil(t, discards)
		require.NotNil(t, syncBlock)
		assert.Nil(t, syncBlock.Rollup.Vars)
		assert.Nil(t, syncBlock.Auction.Vars)
		assert.Nil(t, syncBlock.WDelayer.Vars)
		assert.Equal(t, int64(2+i), syncBlock.Block.Num)

		stats = s.Stats()
		assert.Equal(t, int64(1), stats.Eth.FirstBlockNum)
		assert.Equal(t, int64(6), stats.Eth.LastBlock.Num)
		assert.Equal(t, int64(2+i), stats.Sync.LastBlock.Num)
		if i == 4 {
			assert.Equal(t, true, stats.Synced())
		} else {
			assert.Equal(t, false, stats.Synced())
		}

		vars = s.SCVars()
		assert.Equal(t, clientSetup.RollupVariables, vars.Rollup)
		assert.Equal(t, clientSetup.AuctionVariables, vars.Auction)
		assert.Equal(t, clientSetup.WDelayerVariables, vars.WDelayer)
	}

	dbBlock, err = s.historyDB.GetLastBlock()
	require.NoError(t, err)
	assert.Equal(t, int64(6), dbBlock.Num)

	// Accounts in HistoryDB and StateDB is only 2 entries
	dbAccounts, err = s.historyDB.GetAllAccounts()
	require.NoError(t, err)
	sdbAccounts, err = s.stateDB.GetAccounts()
	require.NoError(t, err)
	assert.Equal(t, 2, len(dbAccounts))
	assertEqualAccountsHistoryDBStateDB(t, dbAccounts, sdbAccounts)
}
