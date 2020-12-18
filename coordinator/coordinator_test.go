package coordinator

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree/db/pebble"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme = []string{}

func pebbleMakeCheckpoint(source, dest string) error {
	// Remove dest folder (if it exists) before doing the checkpoint
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		err := os.RemoveAll(dest)
		if err != nil {
			return tracerr.Wrap(err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return tracerr.Wrap(err)
	}

	sto, err := pebble.NewPebbleStorage(source, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer func() {
		errClose := sto.Pebble().Close()
		if errClose != nil {
			log.Errorw("Pebble.Close", "err", errClose)
		}
	}()

	// execute Checkpoint
	err = sto.Pebble().Checkpoint(dest)
	if err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

func TestMain(m *testing.M) {
	exitVal := m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

var syncDBPath string
var txSelDBPath string
var batchBuilderDBPath string

type modules struct {
	historyDB    *historydb.HistoryDB
	l2DB         *l2db.L2DB
	txSelector   *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder
	stateDB      *statedb.StateDB
}

var maxL1UserTxs uint64 = 128
var maxL1Txs uint64 = 256
var maxL1CoordinatorTxs uint64 = maxL1Txs - maxL1UserTxs
var maxTxs uint64 = 376
var nLevels uint32 = 32   //nolint:deadcode,unused
var maxFeeTxs uint32 = 64 //nolint:deadcode,varcheck

func newTestModules(t *testing.T) modules {
	nLevels := 32

	var err error
	syncDBPath, err = ioutil.TempDir("", "tmpSyncDB")
	require.NoError(t, err)
	deleteme = append(deleteme, syncDBPath)
	syncStateDB, err := statedb.NewStateDB(syncDBPath, statedb.TypeSynchronizer, nLevels)
	assert.NoError(t, err)

	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.NoError(t, err)
	test.WipeDB(db)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)
	historyDB := historydb.NewHistoryDB(db)

	txSelDBPath, err = ioutil.TempDir("", "tmpTxSelDB")
	require.NoError(t, err)
	deleteme = append(deleteme, txSelDBPath)

	coordAccount := &txselector.CoordAccount{ // TODO TMP
		Addr:                ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
		BJJ:                 nil,
		AccountCreationAuth: nil,
	}
	txSelector, err := txselector.NewTxSelector(coordAccount, txSelDBPath, syncStateDB, l2DB)
	assert.NoError(t, err)

	batchBuilderDBPath, err = ioutil.TempDir("", "tmpBatchBuilderDB")
	require.NoError(t, err)
	deleteme = append(deleteme, batchBuilderDBPath)
	batchBuilder, err := batchbuilder.NewBatchBuilder(batchBuilderDBPath, syncStateDB, nil, 0, uint64(nLevels))
	assert.NoError(t, err)

	return modules{
		historyDB:    historyDB,
		l2DB:         l2DB,
		txSelector:   txSelector,
		batchBuilder: batchBuilder,
		stateDB:      syncStateDB,
	}
}

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

var bidder = ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
var forger = ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")

func newTestCoordinator(t *testing.T, forgerAddr ethCommon.Address, ethClient *test.Client,
	ethClientSetup *test.ClientSetup, modules modules) *Coordinator {
	debugBatchPath, err := ioutil.TempDir("", "tmpDebugBatch")
	require.NoError(t, err)
	deleteme = append(deleteme, debugBatchPath)

	conf := Config{
		ForgerAddress:          forgerAddr,
		ConfirmBlocks:          5,
		L1BatchTimeoutPerc:     0.5,
		EthClientAttempts:      5,
		EthClientAttemptsDelay: 100 * time.Millisecond,
		TxManagerCheckInterval: 300 * time.Millisecond,
		DebugBatchPath:         debugBatchPath,
	}
	serverProofs := []prover.Client{&prover.MockClient{}, &prover.MockClient{}}

	scConsts := &synchronizer.SCConsts{
		Rollup:   *ethClientSetup.RollupConstants,
		Auction:  *ethClientSetup.AuctionConstants,
		WDelayer: *ethClientSetup.WDelayerConstants,
	}
	initSCVars := &synchronizer.SCVariables{
		Rollup:   *ethClientSetup.RollupVariables,
		Auction:  *ethClientSetup.AuctionVariables,
		WDelayer: *ethClientSetup.WDelayerVariables,
	}
	coord, err := NewCoordinator(conf, modules.historyDB, modules.l2DB, modules.txSelector,
		modules.batchBuilder, serverProofs, ethClient, scConsts, initSCVars)
	require.NoError(t, err)
	return coord
}

func newTestSynchronizer(t *testing.T, ethClient *test.Client, ethClientSetup *test.ClientSetup,
	modules modules) *synchronizer.Synchronizer {
	sync, err := synchronizer.NewSynchronizer(ethClient, modules.historyDB, modules.stateDB,
		synchronizer.Config{
			StatsRefreshPeriod: 0 * time.Second,
		})
	require.NoError(t, err)
	return sync
}

// TestCoordinatorFlow is a test where the coordinator is stared (which means
// that goroutines are spawned), and ethereum blocks are mined via the
// test.Client to simulate starting and stopping forging times.  This test
// works without a synchronizer, and no l2txs are inserted in the pool, so all
// the batches are forged empty.  The purpose of this test is to manually
// observe via the logs that nothing crashes and that the coordinator starts
// and stops forging at the right blocks.
func TestCoordinatorFlow(t *testing.T) {
	if os.Getenv("TEST_COORD_FLOW") == "" {
		return
	}
	ethClientSetup := test.NewClientSetupExample()
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)

	// Bid for slot 2 and 4
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	_, err = ethClient.AuctionBidSimple(2, big.NewInt(9999))
	require.NoError(t, err)
	_, err = ethClient.AuctionBidSimple(4, big.NewInt(9999))
	require.NoError(t, err)

	coord.Start()
	time.Sleep(1 * time.Second)

	waitForSlot := func(slot int64) {
		for {
			blockNum, err := ethClient.EthLastBlock()
			require.NoError(t, err)
			nextBlockSlot, err := ethClient.AuctionGetSlotNumber(blockNum + 1)
			require.NoError(t, err)
			if nextBlockSlot == slot {
				break
			}
			ethClient.CtlMineBlock()
			time.Sleep(100 * time.Millisecond)
			var stats synchronizer.Stats
			stats.Eth.LastBlock = *ethClient.CtlLastBlock()
			stats.Sync.LastBlock = stats.Eth.LastBlock
			stats.Eth.LastBatch = ethClient.CtlLastForgedBatch()
			stats.Sync.LastBatch = stats.Eth.LastBatch
			canForge, err := ethClient.AuctionCanForge(forger, blockNum+1)
			require.NoError(t, err)
			if canForge {
				// fmt.Println("DBG canForge")
				stats.Sync.Auction.CurrentSlot.Forger = forger
			}
			// Copy stateDB to synchronizer if there was a new batch
			source := fmt.Sprintf("%v/BatchNum%v", batchBuilderDBPath, stats.Sync.LastBatch)
			dest := fmt.Sprintf("%v/BatchNum%v", syncDBPath, stats.Sync.LastBatch)
			if stats.Sync.LastBatch != 0 {
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					log.Infow("Making pebble checkpoint for sync",
						"source", source, "dest", dest)
					err = pebbleMakeCheckpoint(source, dest)
					require.NoError(t, err)
				}
			}
			coord.SendMsg(MsgSyncBlock{
				Stats: stats,
			})
		}
	}

	// NOTE: With the current test, the coordinator will enter in forge
	// time before the bidded slot because no one else is forging in the
	// other slots before the slot deadline.
	// simulate forgeSequence time
	waitForSlot(2)
	log.Info("~~~ simulate entering in forge time")
	time.Sleep(1 * time.Second)

	// simulate going out from forgeSequence
	waitForSlot(3)
	log.Info("~~~ simulate going out from forge time")
	time.Sleep(1 * time.Second)

	// simulate entering forgeSequence time again
	waitForSlot(4)
	log.Info("~~~ simulate entering in forge time again")
	time.Sleep(2 * time.Second)

	// simulate stopping forgerLoop by channel
	log.Info("~~~ simulate stopping forgerLoop by closing coordinator stopch")
	coord.Stop()
	time.Sleep(1 * time.Second)
}

func TestCoordinatorStartStop(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)
	coord.Start()
	coord.Stop()
}

func TestCoordCanForge(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	bid, ok := new(big.Int).SetString("12000000000000000000", 10)
	if !ok {
		panic("bad bid")
	}
	_, err = ethClient.AuctionBidSimple(2, bid)
	require.NoError(t, err)

	modules2 := newTestModules(t)
	bootCoord := newTestCoordinator(t, bootForger, ethClient, ethClientSetup, modules2)

	assert.Equal(t, forger, coord.cfg.ForgerAddress)
	assert.Equal(t, bootForger, bootCoord.cfg.ForgerAddress)
	ethBootCoord, err := ethClient.AuctionGetBootCoordinator()
	require.NoError(t, err)
	assert.Equal(t, &bootForger, ethBootCoord)

	var stats synchronizer.Stats

	// Slot 0.  No bid, so the winner is the boot coordinator
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, false, coord.canForge(&stats))
	assert.Equal(t, true, bootCoord.canForge(&stats))

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, true, coord.canForge(&stats))
	assert.Equal(t, true, bootCoord.canForge(&stats))

	// Slot 1. coordinator bid, so the winner is the coordinator
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		1*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = forger
	assert.Equal(t, true, coord.canForge(&stats))
	assert.Equal(t, false, bootCoord.canForge(&stats))
}

func TestCoordHandleMsgSyncBlock(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	bid, ok := new(big.Int).SetString("11000000000000000000", 10)
	if !ok {
		panic("bad bid")
	}
	_, err = ethClient.AuctionBidSimple(2, bid)
	require.NoError(t, err)

	var msg MsgSyncBlock
	stats := &msg.Stats
	ctx := context.Background()

	// Slot 0.  No bid, so the winner is the boot coordinator
	// pipelineStarted: false -> false
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, false, coord.canForge(stats))
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.Nil(t, coord.pipeline)

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	// pipelineStarted: false -> true
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, true, coord.canForge(stats))
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.NotNil(t, coord.pipeline)

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	// pipelineStarted: true -> true
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline) + 1
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, true, coord.canForge(stats))
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.NotNil(t, coord.pipeline)

	// Slot 0. No bid, so the winner is the boot coordinator
	// pipelineStarted: true -> false
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		1*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot.Forger = bootForger
	assert.Equal(t, false, coord.canForge(stats))
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.Nil(t, coord.pipeline)
}

func TestPipelineShouldL1L2Batch(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()

	var timer timer
	ctx := context.Background()
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)
	pipeline, err := coord.newPipeline(ctx)
	require.NoError(t, err)
	pipeline.vars = coord.vars

	// Check that the parameters are the ones we expect and use in this test
	require.Equal(t, 0.5, pipeline.cfg.L1BatchTimeoutPerc)
	require.Equal(t, int64(9), ethClientSetup.RollupVariables.ForgeL1L2BatchTimeout)
	l1BatchTimeoutPerc := pipeline.cfg.L1BatchTimeoutPerc
	l1BatchTimeout := ethClientSetup.RollupVariables.ForgeL1L2BatchTimeout

	var stats synchronizer.Stats

	startBlock := int64(100)

	//
	// No scheduled L1Batch
	//

	// Last L1Batch was a long time ago
	stats.Eth.LastBlock.Num = startBlock
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.LastL1BatchBlock = 0
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch())

	stats.Sync.LastL1BatchBlock = startBlock

	// We are are one block before the timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock + int64(float64(l1BatchTimeout)*l1BatchTimeoutPerc) - 1
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, false, pipeline.shouldL1L2Batch())

	// We are are at timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock + int64(float64(l1BatchTimeout)*l1BatchTimeoutPerc)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch())

	//
	// Scheduled L1Batch
	//
	pipeline.lastScheduledL1BatchBlockNum = startBlock
	stats.Sync.LastL1BatchBlock = startBlock - 10

	// We are are one block before the timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock + int64(float64(l1BatchTimeout)*l1BatchTimeoutPerc) - 1
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, false, pipeline.shouldL1L2Batch())

	// We are are at timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock + int64(float64(l1BatchTimeout)*l1BatchTimeoutPerc)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch())
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
			// tokenConsts[token.TokenID] = consts
			client.CtlAddERC20(token.EthAddr, consts)
		}
	}
}

const testTokensLen = 3
const testUsersLen = 4

func preloadSync(t *testing.T, ethClient *test.Client, sync *synchronizer.Synchronizer,
	historyDB *historydb.HistoryDB, stateDB *statedb.StateDB) *til.Context {
	// Create a set with `testTokensLen` tokens and for each token
	// `testUsersLen` accounts.
	var set []til.Instruction
	// set = append(set, til.Instruction{Typ: "Blockchain"})
	for tokenID := 1; tokenID < testTokensLen; tokenID++ {
		set = append(set, til.Instruction{
			Typ:     til.TypeAddToken,
			TokenID: common.TokenID(tokenID),
		})
	}
	depositAmount, ok := new(big.Int).SetString("10225000000000000000000000000000000", 10)
	require.True(t, ok)
	for tokenID := 0; tokenID < testTokensLen; tokenID++ {
		for user := 0; user < testUsersLen; user++ {
			set = append(set, til.Instruction{
				Typ:           common.TxTypeCreateAccountDeposit,
				TokenID:       common.TokenID(tokenID),
				DepositAmount: depositAmount,
				From:          fmt.Sprintf("User%d", user),
			})
		}
	}
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBatchL1})
	set = append(set, til.Instruction{Typ: til.TypeNewBlock})

	tc := til.NewContext(common.RollupConstMaxL1UserTx)
	blocks, err := tc.GenerateBlocksFromInstructions(set)
	require.NoError(t, err)
	require.NotNil(t, blocks)

	ethAddTokens(blocks, ethClient)
	err = ethClient.CtlAddBlocks(blocks)
	require.NoError(t, err)

	ctx := context.Background()
	for {
		syncBlock, discards, err := sync.Sync2(ctx, nil)
		require.NoError(t, err)
		require.Nil(t, discards)
		if syncBlock == nil {
			break
		}
	}
	dbTokens, err := historyDB.GetAllTokens()
	require.Nil(t, err)
	require.Equal(t, testTokensLen, len(dbTokens))

	dbAccounts, err := historyDB.GetAllAccounts()
	require.Nil(t, err)
	require.Equal(t, testTokensLen*testUsersLen, len(dbAccounts))

	sdbAccounts, err := stateDB.GetAccounts()
	require.Nil(t, err)
	require.Equal(t, testTokensLen*testUsersLen, len(sdbAccounts))

	return tc
}

func TestPipeline1(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()

	var timer timer
	ctx := context.Background()
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules)
	sync := newTestSynchronizer(t, ethClient, ethClientSetup, modules)
	pipeline, err := coord.newPipeline(ctx)
	require.NoError(t, err)

	require.NotNil(t, sync)
	require.NotNil(t, pipeline)

	// preload the synchronier (via the test ethClient) some tokens and
	// users with positive balances
	tilCtx := preloadSync(t, ethClient, sync, modules.historyDB, modules.stateDB)
	syncStats := sync.Stats()
	batchNum := common.BatchNum(syncStats.Sync.LastBatch)
	syncSCVars := sync.SCVars()

	// Insert some l2txs in the Pool
	setPool := `
Type: PoolL2

PoolTransfer(0) User0-User1: 100 (126)
PoolTransfer(0) User1-User2: 200 (126)
PoolTransfer(0) User2-User3: 300 (126)
	`
	l2txs, err := tilCtx.GeneratePoolL2Txs(setPool)
	require.NoError(t, err)
	for _, tx := range l2txs {
		err := modules.l2DB.AddTxTest(&tx) //nolint:gosec
		require.NoError(t, err)
	}

	err = pipeline.reset(batchNum, syncStats.Sync.LastForgeL1TxsNum, &synchronizer.SCVariables{
		Rollup:   *syncSCVars.Rollup,
		Auction:  *syncSCVars.Auction,
		WDelayer: *syncSCVars.WDelayer,
	})
	require.NoError(t, err)
	// Sanity check
	sdbAccounts, err := pipeline.txSelector.LocalAccountsDB().GetAccounts()
	require.Nil(t, err)
	require.Equal(t, testTokensLen*testUsersLen, len(sdbAccounts))

	// Sanity check
	sdbAccounts, err = pipeline.batchBuilder.LocalStateDB().GetAccounts()
	require.Nil(t, err)
	require.Equal(t, testTokensLen*testUsersLen, len(sdbAccounts))

	// Sanity check
	require.Equal(t, modules.stateDB.MerkleTree().Root(),
		pipeline.batchBuilder.LocalStateDB().MerkleTree().Root())

	batchNum++

	selectionConfig := &txselector.SelectionConfig{
		MaxL1UserTxs:        maxL1UserTxs,
		MaxL1CoordinatorTxs: maxL1CoordinatorTxs,
		ProcessTxsConfig: statedb.ProcessTxsConfig{
			NLevels:  nLevels,
			MaxFeeTx: maxFeeTxs,
			MaxTx:    uint32(maxTxs),
			MaxL1Tx:  uint32(maxL1Txs),
		},
	}

	batchInfo, err := pipeline.forgeBatch(ctx, batchNum, selectionConfig)
	require.NoError(t, err)
	assert.Equal(t, 3, len(batchInfo.L2Txs))

	batchNum++
	batchInfo, err = pipeline.forgeBatch(ctx, batchNum, selectionConfig)
	require.NoError(t, err)
	assert.Equal(t, 0, len(batchInfo.L2Txs))
}

// TODO: Test Reorg
// TODO: Test Pipeline
// TODO: Test TxMonitor
// TODO: Test forgeBatch
// TODO: Test waitServerProof
// TODO: Test handleReorg
