package coordinator

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/etherscan"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
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

var maxL1Txs uint64 = 256
var maxTxs uint64 = 376
var nLevels uint32 = 32   //nolint:deadcode,unused
var maxFeeTxs uint32 = 64 //nolint:deadcode,varcheck
var chainID uint16 = 0

func newTestModules(t *testing.T) modules {
	var err error
	syncDBPath, err = ioutil.TempDir("", "tmpSyncDB")
	require.NoError(t, err)
	deleteme = append(deleteme, syncDBPath)
	syncStateDB, err := statedb.NewStateDB(statedb.Config{Path: syncDBPath, Keep: 128,
		Type: statedb.TypeSynchronizer, NLevels: 48})
	assert.NoError(t, err)

	db, err := dbUtils.InitTestSQLDB()
	require.NoError(t, err)
	test.WipeDB(db)
	l2DB := l2db.NewL2DB(db, db, 10, 100, 0.0, 1000.0, 24*time.Hour, nil)
	historyDB := historydb.NewHistoryDB(db, db, nil)

	txSelDBPath, err = ioutil.TempDir("", "tmpTxSelDB")
	require.NoError(t, err)
	deleteme = append(deleteme, txSelDBPath)

	var bjj babyjub.PublicKeyComp
	err = bjj.UnmarshalText([]byte("c433f7a696b7aa3a5224efb3993baf0ccd9e92eecee0c29a3f6c8208a9e81d9e"))
	require.NoError(t, err)
	coordAccount := txselector.CoordAccount{
		Addr:                ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
		BJJ:                 bjj,
		AccountCreationAuth: nil,
	}
	txSelector, err := txselector.NewTxSelector(&coordAccount, txSelDBPath, syncStateDB, l2DB)
	assert.NoError(t, err)

	batchBuilderDBPath, err = ioutil.TempDir("", "tmpBatchBuilderDB")
	require.NoError(t, err)
	deleteme = append(deleteme, batchBuilderDBPath)
	batchBuilder, err := batchbuilder.NewBatchBuilder(batchBuilderDBPath, syncStateDB, 0,
		uint64(nLevels))
	assert.NoError(t, err)

	return modules{
		historyDB:    historyDB,
		l2DB:         l2DB,
		txSelector:   txSelector,
		batchBuilder: batchBuilder,
		stateDB:      syncStateDB,
	}
}

func closeTestModules(t *testing.T, modules modules) {
	_ = modules.l2DB.DB().Close()
	modules.txSelector.LocalAccountsDB().Close()
	modules.batchBuilder.LocalStateDB().Close()
	modules.stateDB.Close()
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
	ethClientSetup *test.ClientSetup, modules modules, etherscanService *etherscan.Service) *Coordinator {
	debugBatchPath, err := ioutil.TempDir("", "tmpDebugBatch")
	require.NoError(t, err)
	deleteme = append(deleteme, debugBatchPath)

	conf := Config{
		ForgerAddress:           forgerAddr,
		ConfirmBlocks:           5,
		L1BatchTimeoutPerc:      0.5,
		EthClientAttempts:       5,
		SyncRetryInterval:       400 * time.Microsecond,
		EthClientAttemptsDelay:  100 * time.Millisecond,
		TxManagerCheckInterval:  300 * time.Millisecond,
		ProverReadTimeout:       20 * time.Second,
		DebugBatchPath:          debugBatchPath,
		MustForgeAtSlotDeadline: true,
		Purger: PurgerCfg{
			PurgeBatchDelay:      10,
			PurgeBlockDelay:      10,
			InvalidateBatchDelay: 4,
			InvalidateBlockDelay: 4,
		},
		TxProcessorConfig: txprocessor.Config{
			NLevels:  nLevels,
			MaxFeeTx: maxFeeTxs,
			MaxTx:    uint32(maxTxs),
			MaxL1Tx:  uint32(maxL1Txs),
			ChainID:  chainID,
		},
		VerifierIdx: 0,
	}

	serverProofs := []prover.Client{
		&prover.MockClient{Delay: 300 * time.Millisecond},
		&prover.MockClient{Delay: 400 * time.Millisecond},
	}

	scConsts := &common.SCConsts{
		Rollup:   *ethClientSetup.RollupConstants,
		Auction:  *ethClientSetup.AuctionConstants,
		WDelayer: *ethClientSetup.WDelayerConstants,
	}
	initSCVars := &common.SCVariables{
		Rollup:   *ethClientSetup.RollupVariables,
		Auction:  *ethClientSetup.AuctionVariables,
		WDelayer: *ethClientSetup.WDelayerVariables,
	}
	coord, err := NewCoordinator(conf, modules.historyDB, modules.l2DB, modules.txSelector,
		modules.batchBuilder, serverProofs, ethClient, scConsts, initSCVars, etherscanService)
	require.NoError(t, err)
	return coord
}

func newTestSynchronizer(t *testing.T, ethClient *test.Client, ethClientSetup *test.ClientSetup,
	modules modules) *synchronizer.Synchronizer {
	sync, err := synchronizer.NewSynchronizer(ethClient, modules.historyDB, modules.l2DB, modules.stateDB,
		synchronizer.Config{
			StatsUpdateBlockNumDiffThreshold: 100,
			StatsUpdateFrequencyDivider:      100,
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
	ethClientSetup.ChainID = big.NewInt(int64(chainID))
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)

	// Bid for slot 2 and 4
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	bid, ok := new(big.Int).SetString("11000000000000000000", 10)
	if !ok {
		panic("bad bid")
	}
	_, err = ethClient.AuctionBidSimple(3, bid)
	require.NoError(t, err)
	_, err = ethClient.AuctionBidSimple(5, bid)
	require.NoError(t, err)

	coord.Start()
	ctx := context.Background()
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
			stats.Eth.LastBatchNum = ethClient.CtlLastForgedBatch()
			stats.Sync.LastBatch.BatchNum = common.BatchNum(stats.Eth.LastBatchNum)
			canForge, err := ethClient.AuctionCanForge(forger, blockNum+1)
			require.NoError(t, err)
			var slot common.Slot
			slotNum := ethClientSetup.AuctionConstants.SlotNum(blockNum + 1)
			slot.StartBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
				(slotNum)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
			slot.EndBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
				(slotNum+1)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
			if canForge {
				slot.Forger = forger
			}
			stats.Sync.Auction.CurrentSlot = slot

			// Copy stateDB to synchronizer if there was a new batch
			source := fmt.Sprintf("%v/BatchNum%v", batchBuilderDBPath, stats.Sync.LastBatch)
			dest := fmt.Sprintf("%v/BatchNum%v", syncDBPath, stats.Sync.LastBatch)
			if stats.Sync.LastBatch.BatchNum != 0 {
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					log.Infow("Making pebble checkpoint for sync",
						"source", source, "dest", dest)
					err = pebbleMakeCheckpoint(source, dest)
					require.NoError(t, err)
				}
			}
			coord.SendMsg(ctx, MsgSyncBlock{
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

	closeTestModules(t, modules)
}

func TestCoordinatorStartStop(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	ethClientSetup.ChainID = big.NewInt(int64(chainID))
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)
	coord.Start()
	coord.Stop()

	closeTestModules(t, modules)
}

func TestCoordCanForge(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	ethClientSetup.ChainID = big.NewInt(int64(chainID))
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	bid, ok := new(big.Int).SetString("12000000000000000000", 10)
	if !ok {
		panic("bad bid")
	}
	_, err = ethClient.AuctionBidSimple(3, bid)
	require.NoError(t, err)

	modules2 := newTestModules(t)
	bootCoord := newTestCoordinator(t, bootForger, ethClient, ethClientSetup, modules2, etherScanService)

	assert.Equal(t, forger, coord.cfg.ForgerAddress)
	assert.Equal(t, bootForger, bootCoord.cfg.ForgerAddress)
	ethBootCoord, err := ethClient.AuctionGetBootCoordinator()
	require.NoError(t, err)
	assert.Equal(t, &bootForger, ethBootCoord)

	var stats synchronizer.Stats

	slots := [4]common.Slot{}
	for i := 0; i < 4; i++ {
		slots[i].StartBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
			int64(i)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
		slots[i].EndBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
			int64(i+1)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
		// Only slot 3 has Coordinator winner, the rest are BootCoordinator
		if i == 3 {
			slots[i].Forger = forger
		} else {
			slots[i].Forger = bootForger
		}
	}

	// Slot 0.  No bid, so the winner is the boot coordinator
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot = slots[0]
	coord.stats = stats
	bootCoord.stats = stats
	assert.Equal(t, false, coord.canForge())
	assert.Equal(t, true, bootCoord.canForge())

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot = slots[0]
	coord.stats = stats
	bootCoord.stats = stats
	assert.Equal(t, true, coord.canForge())
	assert.Equal(t, true, bootCoord.canForge())

	// Anyone can forge but the node MustForgeAtSlotDeadline as set as false
	coord.cfg.MustForgeAtSlotDeadline = false
	assert.Equal(t, false, coord.canForge())

	// Slot 3. coordinator bid, so the winner is the coordinator
	stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		3*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.Auction.CurrentSlot = slots[3]
	coord.stats = stats
	bootCoord.stats = stats
	assert.Equal(t, true, coord.canForge())
	assert.Equal(t, false, bootCoord.canForge())

	closeTestModules(t, modules)
	closeTestModules(t, modules2)
}

func TestCoordHandleMsgSyncBlock(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	ethClientSetup.ChainID = big.NewInt(int64(chainID))
	bootForger := ethClientSetup.AuctionVariables.BootCoordinator

	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.NoError(t, err)
	bid, ok := new(big.Int).SetString("11000000000000000000", 10)
	if !ok {
		panic("bad bid")
	}
	_, err = ethClient.AuctionBidSimple(3, bid)
	require.NoError(t, err)

	slots := [4]common.Slot{}
	for i := 0; i < 4; i++ {
		slots[i].StartBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
			int64(i)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
		slots[i].EndBlock = ethClientSetup.AuctionConstants.GenesisBlockNum +
			int64(i+1)*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
		// Only slot 3 has Coordinator winner, the rest are BootCoordinator
		if i == 3 {
			slots[i].Forger = forger
		} else {
			slots[i].Forger = bootForger
		}
	}

	var msg MsgSyncBlock
	coord.stats = msg.Stats
	ctx := context.Background()

	// Slot 0.  No bid, so the winner is the boot coordinator
	// pipelineStarted: false -> false
	coord.stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum
	coord.stats.Sync.LastBlock = coord.stats.Eth.LastBlock
	coord.stats.Sync.Auction.CurrentSlot = slots[0]
	assert.Equal(t, false, coord.canForge())
	msg.Stats = coord.stats
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.Nil(t, coord.pipeline)

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	// pipelineStarted: false -> true
	coord.stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline)
	coord.stats.Sync.LastBlock = coord.stats.Eth.LastBlock
	coord.stats.Sync.Auction.CurrentSlot = slots[0]
	assert.Equal(t, true, coord.canForge())
	msg.Stats = coord.stats
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.NotNil(t, coord.pipeline)

	// Slot 0.  No bid, and we reach the deadline, so anyone can forge
	// pipelineStarted: true -> true
	coord.stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		int64(ethClientSetup.AuctionVariables.SlotDeadline) + 1
	coord.stats.Sync.LastBlock = coord.stats.Eth.LastBlock
	coord.stats.Sync.Auction.CurrentSlot = slots[0]
	assert.Equal(t, true, coord.canForge())
	msg.Stats = coord.stats
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.NotNil(t, coord.pipeline)

	// Slot 1. No bid, so the winner is the boot coordinator
	// pipelineStarted: true -> false
	coord.stats.Eth.LastBlock.Num = ethClientSetup.AuctionConstants.GenesisBlockNum +
		1*int64(ethClientSetup.AuctionConstants.BlocksPerSlot)
	coord.stats.Sync.LastBlock = coord.stats.Eth.LastBlock
	coord.stats.Sync.Auction.CurrentSlot = slots[1]
	assert.Equal(t, false, coord.canForge())
	msg.Stats = coord.stats
	require.NoError(t, coord.handleMsgSyncBlock(ctx, &msg))
	assert.Nil(t, coord.pipeline)

	closeTestModules(t, modules)
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

func TestCoordinatorStress(t *testing.T) {
	if os.Getenv("TEST_COORD_STRESS") == "" {
		return
	}
	log.Info("Begin Test Coord Stress")
	ethClientSetup := test.NewClientSetupExample()
	ethClientSetup.ChainID = big.NewInt(int64(chainID))
	var timer timer
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)
	syn := newTestSynchronizer(t, ethClient, ethClientSetup, modules)

	coord.Start()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Synchronizer loop
	wg.Add(1)
	go func() {
		for {
			blockData, _, err := syn.Sync(ctx, nil)
			if ctx.Err() != nil {
				wg.Done()
				return
			}
			require.NoError(t, err)
			if blockData != nil {
				stats := syn.Stats()
				coord.SendMsg(ctx, MsgSyncBlock{
					Stats:   *stats,
					Batches: blockData.Rollup.Batches,
					Vars: common.SCVariablesPtr{
						Rollup:   blockData.Rollup.Vars,
						Auction:  blockData.Auction.Vars,
						WDelayer: blockData.WDelayer.Vars,
					},
				})
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Blockchain mining loop
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				wg.Done()
				return
			case <-time.After(1 * time.Second):
				ethClient.CtlMineBlock()
			}
		}
	}()

	time.Sleep(600 * time.Second)

	cancel()
	wg.Wait()
	coord.Stop()

	closeTestModules(t, modules)
}
