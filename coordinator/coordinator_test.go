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
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test"
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

func newTestModules(t *testing.T) (*historydb.HistoryDB, *l2db.L2DB,
	*txselector.TxSelector, *batchbuilder.BatchBuilder) { // FUTURE once Synchronizer is ready, should return it also
	nLevels := 32

	var err error
	syncDBPath, err = ioutil.TempDir("", "tmpSyncDB")
	require.NoError(t, err)
	deleteme = append(deleteme, syncDBPath)
	syncSdb, err := statedb.NewStateDB(syncDBPath, statedb.TypeSynchronizer, nLevels)
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
	txsel, err := txselector.NewTxSelector(txSelDBPath, syncSdb, l2DB, 10, 10, 10)
	assert.NoError(t, err)

	batchBuilderDBPath, err = ioutil.TempDir("", "tmpBatchBuilderDB")
	require.NoError(t, err)
	deleteme = append(deleteme, batchBuilderDBPath)
	bb, err := batchbuilder.NewBatchBuilder(batchBuilderDBPath, syncSdb, nil, 0, uint64(nLevels))
	assert.NoError(t, err)

	// l1Txs, coordinatorL1Txs, poolL2Txs := test.GenerateTestTxsFromSet(t, test.SetTest0)

	return historyDB, l2DB, txsel, bb
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

func newTestCoordinator(t *testing.T, forgerAddr ethCommon.Address, ethClient *test.Client, ethClientSetup *test.ClientSetup) *Coordinator {
	historyDB, l2DB, txsel, bb := newTestModules(t)

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
	coord, err := NewCoordinator(conf, historyDB, l2DB, txsel, bb, serverProofs,
		ethClient, scConsts, initSCVars)
	require.NoError(t, err)
	return coord
}

func TestCoordinatorFlow(t *testing.T) {
	if os.Getenv("TEST_COORD_FLOW") == "" {
		return
	}
	ethClientSetup := test.NewClientSetupExample()
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup)

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
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup)
	coord.Start()
	coord.Stop()
}

func TestCoordCanForge(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	_, err = ethClient.AuctionBidSimple(2, big.NewInt(9999))
	require.NoError(t, err)

	bootCoord := newTestCoordinator(t, bootForger, ethClient, ethClientSetup)

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

func TestCoordHandleMsgSyncStats(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	_, err = ethClient.AuctionBidSimple(2, big.NewInt(9999))
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
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup)
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

// TODO: Test Reorg
// TODO: Test Pipeline
// TODO: Test TxMonitor
// TODO: Test forgeSendServerProof
// TODO: Test waitServerProof
// TODO: Test handleReorg
