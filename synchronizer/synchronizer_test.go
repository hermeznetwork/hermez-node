package synchronizer

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
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

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

func TestSync(t *testing.T) {
	//
	// Setup
	//

	ctx := context.Background()
	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	stateDB, err := statedb.NewStateDB(dir, statedb.TypeSynchronizer, 32)
	assert.Nil(t, err)

	// Init History DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	historyDB := historydb.NewHistoryDB(db)
	// Clear DB
	err = historyDB.Reorg(-1)
	assert.Nil(t, err)

	// Init eth client
	var timer timer
	clientSetup := test.NewClientSetupExample()
	bootCoordAddr := clientSetup.AuctionVariables.BootCoordinator
	client := test.NewClient(true, &timer, &ethCommon.Address{}, clientSetup)

	// Create Synchronizer
	s, err := NewSynchronizer(client, historyDB, stateDB)
	require.Nil(t, err)

	//
	// First Sync from an initial state
	//

	// Test Sync for rollup genesis block
	syncBlock, discards, err := s.Sync2(ctx, nil)
	require.Nil(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.Equal(t, int64(1), syncBlock.Block.EthBlockNum)
	dbBlocks, err := s.historyDB.GetAllBlocks()
	require.Nil(t, err)
	assert.Equal(t, 1, len(dbBlocks))
	assert.Equal(t, int64(1), dbBlocks[0].EthBlockNum)

	// Sync again and expect no new blocks
	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.Nil(t, err)
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

		CreateAccountDeposit(1) A: 20 // Idx=256+1
		CreateAccountDeposit(2) A: 20 // Idx=256+2
		CreateAccountDeposit(1) B: 5  // Idx=256+3
		CreateAccountDeposit(1) C: 5  // Idx=256+4
		CreateAccountDeposit(1) D: 5  // Idx=256+5

		CreateAccountDepositCoordinator(2) B // Idx=256+0

		> batchL1 // forge L1UserTxs{nil}, freeze defined L1UserTxs
		> batchL1 // forge defined L1UserTxs, freeze L1UserTxs{nil}
		> block

	`
	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocks(set1)
	require.Nil(t, err)
	require.Equal(t, 1, len(blocks))
	require.Equal(t, 3, len(blocks[0].AddedTokens))
	require.Equal(t, 5, len(blocks[0].L1UserTxs))
	require.Equal(t, 2, len(blocks[0].Batches))

	tokenConsts := map[common.TokenID]eth.ERC20Consts{}
	// Generate extra required data
	for _, block := range blocks {
		for _, token := range block.AddedTokens {
			consts := eth.ERC20Consts{
				Name:     fmt.Sprintf("Token %d", token.TokenID),
				Symbol:   fmt.Sprintf("TK%d", token.TokenID),
				Decimals: 18,
			}
			tokenConsts[token.TokenID] = consts
			client.CtlAddERC20(token.EthAddr, consts)
		}
	}

	// Add block data to the smart contracts
	for _, block := range blocks {
		for _, token := range block.AddedTokens {
			_, err := client.RollupAddTokenSimple(token.EthAddr, clientSetup.RollupVariables.FeeAddToken)
			require.Nil(t, err)
		}
		for _, tx := range block.L1UserTxs {
			client.CtlSetAddr(tx.FromEthAddr)
			_, err := client.RollupL1UserTxERC20ETH(tx.FromBJJ, int64(tx.FromIdx), tx.LoadAmount, tx.Amount,
				uint32(tx.TokenID), int64(tx.ToIdx))
			require.Nil(t, err)
		}
		client.CtlSetAddr(bootCoordAddr)
		for _, batch := range block.Batches {
			_, err := client.RollupForgeBatch(&eth.RollupForgeBatchArgs{
				NewLastIdx:            batch.Batch.LastIdx,
				NewStRoot:             batch.Batch.StateRoot,
				NewExitRoot:           batch.Batch.ExitRoot,
				L1CoordinatorTxs:      batch.L1CoordinatorTxs,
				L1CoordinatorTxsAuths: [][]byte{}, // Intentionally empty
				L2TxsData:             batch.L2Txs,
				FeeIdxCoordinator:     []common.Idx{}, // TODO
				// Circuit selector
				VerifierIdx: 0, // Intentionally empty
				L1Batch:     batch.L1Batch,
				ProofA:      [2]*big.Int{},    // Intentionally empty
				ProofB:      [2][2]*big.Int{}, // Intentionally empty
				ProofC:      [2]*big.Int{},    // Intentionally empty
			})
			require.Nil(t, err)
		}
		// Mine block and sync
		client.CtlMineBlock()
	}

	//
	// Sync to synchronize the current state from the test smart contracts
	//

	syncBlock, discards, err = s.Sync2(ctx, nil)
	require.Nil(t, err)
	require.Nil(t, discards)
	require.NotNil(t, syncBlock)
	assert.Equal(t, int64(2), syncBlock.Block.EthBlockNum)

	// Fill extra fields not generated by til in til block
	openToForge := int64(0)
	toForgeL1TxsNum := int64(0)
	for i := range blocks {
		block := &blocks[i]
		for j := range block.Batches {
			batch := &block.Batches[j]
			if batch.L1Batch {
				// Set BatchNum for forged L1UserTxs to til blocks
				bn := batch.Batch.BatchNum
				for k := range blocks {
					block := &blocks[k]
					for l := range block.L1UserTxs {
						tx := &block.L1UserTxs[l]
						if *tx.ToForgeL1TxsNum == openToForge {
							tx.BatchNum = &bn
						}
					}
				}
				openToForge++
			}

			batch.Batch.EthBlockNum = block.Block.EthBlockNum
			batch.Batch.ForgerAddr = bootCoordAddr // til doesn't fill the batch forger addr
			if batch.L1Batch {
				toForgeL1TxsNumCpy := toForgeL1TxsNum
				batch.Batch.ForgeL1TxsNum = &toForgeL1TxsNumCpy // til doesn't fill the ForgeL1TxsNum
				toForgeL1TxsNum++
			}

			batchNum := batch.Batch.BatchNum
			for j := range batch.L1CoordinatorTxs {
				tx := &batch.L1CoordinatorTxs[j]
				tx.BatchNum = &batchNum
				tx.EthBlockNum = batch.Batch.EthBlockNum
				nTx, err := common.NewL1Tx(tx)
				require.Nil(t, err)
				*tx = *nTx
			}
		}
	}

	block := blocks[0]

	//
	// Check Sync output and HistoryDB state against expected values
	// generated by til
	//

	// Check Blocks
	dbBlocks, err = s.historyDB.GetAllBlocks()
	require.Nil(t, err)
	assert.Equal(t, 2, len(dbBlocks))
	assert.Equal(t, int64(2), dbBlocks[1].EthBlockNum)
	assert.NotEqual(t, dbBlocks[1].Hash, dbBlocks[0].Hash)
	assert.Greater(t, dbBlocks[1].Timestamp.Unix(), dbBlocks[0].Timestamp.Unix())

	// Check Tokens
	assert.Equal(t, len(block.AddedTokens), len(syncBlock.AddedTokens))
	dbTokens, err := s.historyDB.GetAllTokens()
	require.Nil(t, err)
	assert.Equal(t, len(block.AddedTokens), len(dbTokens))
	for i, token := range block.AddedTokens {
		dbToken := dbTokens[i]
		syncToken := syncBlock.AddedTokens[i]

		assert.Equal(t, block.Block.EthBlockNum, syncToken.EthBlockNum)
		assert.Equal(t, token.TokenID, syncToken.TokenID)
		assert.Equal(t, token.EthAddr, syncToken.EthAddr)
		tokenConst := tokenConsts[token.TokenID]
		assert.Equal(t, tokenConst.Name, syncToken.Name)
		assert.Equal(t, tokenConst.Symbol, syncToken.Symbol)
		assert.Equal(t, tokenConst.Decimals, syncToken.Decimals)

		var tokenCpy historydb.TokenRead
		//nolint:gosec
		require.Nil(t, copier.Copy(&tokenCpy, &token))      // copy common.Token to historydb.TokenRead
		require.Nil(t, copier.Copy(&tokenCpy, &tokenConst)) // copy common.Token to historydb.TokenRead
		tokenCpy.ItemID = dbToken.ItemID                    // we don't care about ItemID
		assert.Equal(t, tokenCpy, dbToken)
	}

	// Check L1UserTxs
	assert.Equal(t, len(block.L1UserTxs), len(syncBlock.L1UserTxs))
	dbL1UserTxs, err := s.historyDB.GetAllL1UserTxs()
	require.Nil(t, err)
	assert.Equal(t, len(block.L1UserTxs), len(dbL1UserTxs))
	// Ignore BatchNum in syncBlock.L1UserTxs because this value is set by the HistoryDB
	for i := range syncBlock.L1UserTxs {
		syncBlock.L1UserTxs[i].BatchNum = block.L1UserTxs[i].BatchNum
	}
	assert.Equal(t, block.L1UserTxs, syncBlock.L1UserTxs)
	assert.Equal(t, block.L1UserTxs, dbL1UserTxs)

	// Check Batches
	assert.Equal(t, len(block.Batches), len(syncBlock.Batches))
	dbBatches, err := s.historyDB.GetAllBatches()
	require.Nil(t, err)
	assert.Equal(t, len(block.Batches), len(dbBatches))

	for i, batch := range block.Batches {
		batchNum := batch.Batch.BatchNum
		dbBatch := dbBatches[i]
		syncBatch := syncBlock.Batches[i]

		// We don't care about TotalFeesUSD.  Use the syncBatch that
		// has a TotalFeesUSD inserted by the HistoryDB
		batch.Batch.TotalFeesUSD = syncBatch.Batch.TotalFeesUSD
		batch.CreatedAccounts = syncBatch.CreatedAccounts // til doesn't output CreatedAccounts

		// fmt.Printf("DBG Batch %d %+v\n", i, batch)
		// fmt.Printf("DBG Batch Sync %d %+v\n", i, syncBatch)
		// assert.Equal(t, batch.L1CoordinatorTxs, syncBatch.L1CoordinatorTxs)
		fmt.Printf("DBG BatchNum: %d, LastIdx: %d\n", batchNum, batch.Batch.LastIdx)
		assert.Equal(t, batch, syncBatch)
		assert.Equal(t, batch.Batch, dbBatch)
	}

	// Check L1UserTxs in DB

	// TODO: Reorg will be properly tested once we have the mock ethClient implemented
	/*
		// Force a Reorg
		lastSavedBlock, err := historyDB.GetLastBlock()
		require.Nil(t, err)

		lastSavedBlock.EthBlockNum++
		err = historyDB.AddBlock(lastSavedBlock)
		require.Nil(t, err)

		lastSavedBlock.EthBlockNum++
		err = historyDB.AddBlock(lastSavedBlock)
		require.Nil(t, err)

		log.Debugf("Wait for the blockchain to generate some blocks...")
		time.Sleep(40 * time.Second)


		err = s.Sync()
		require.Nil(t, err)
	*/
}
