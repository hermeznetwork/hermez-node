package coordinator

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
)

var (
	errLastL1BatchNotSynced = fmt.Errorf("last L1Batch not synced yet")
)

const (
	queueLen         = 16
	longWaitDuration = 999 * time.Hour
	zeroDuration     = 0 * time.Second
)

// Config contains the Coordinator configuration
type Config struct {
	// ForgerAddress is the address under which this coordinator is forging
	ForgerAddress ethCommon.Address
	// ConfirmBlocks is the number of confirmation blocks to wait for sent
	// ethereum transactions before forgetting about them
	ConfirmBlocks int64
	// L1BatchTimeoutPerc is the portion of the range before the L1Batch
	// timeout that will trigger a schedule to forge an L1Batch
	L1BatchTimeoutPerc float64
	// EthClientAttempts is the number of attempts to do an eth client RPC
	// call before giving up
	EthClientAttempts int
	// ForgeRetryInterval is the waiting interval between calls forge a
	// batch after an error
	ForgeRetryInterval time.Duration
	// SyncRetryInterval is the waiting interval between calls to the main
	// handler of a synced block after an error
	SyncRetryInterval time.Duration
	// EthClientAttemptsDelay is delay between attempts do do an eth client
	// RPC call
	EthClientAttemptsDelay time.Duration
	// TxManagerCheckInterval is the waiting interval between receipt
	// checks of ethereum transactions in the TxManager
	TxManagerCheckInterval time.Duration
	// DebugBatchPath if set, specifies the path where batchInfo is stored
	// in JSON in every step/update of the pipeline
	DebugBatchPath    string
	Purger            PurgerCfg
	VerifierIdx       uint8
	TxProcessorConfig txprocessor.Config
}

func (c *Config) debugBatchStore(batchInfo *BatchInfo) {
	if c.DebugBatchPath != "" {
		if err := batchInfo.DebugStore(c.DebugBatchPath); err != nil {
			log.Warnw("Error storing debug BatchInfo",
				"path", c.DebugBatchPath, "err", err)
		}
	}
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	// State
	pipelineBatchNum common.BatchNum // batchNum from which we started the pipeline
	provers          []prover.Client
	consts           synchronizer.SCConsts
	vars             synchronizer.SCVariables
	stats            synchronizer.Stats
	started          bool

	cfg Config

	historyDB    *historydb.HistoryDB
	l2DB         *l2db.L2DB
	txSelector   *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	msgCh  chan interface{}
	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc

	pipeline *Pipeline

	purger    *Purger
	txManager *TxManager
}

// NewCoordinator creates a new Coordinator
func NewCoordinator(cfg Config,
	historyDB *historydb.HistoryDB,
	l2DB *l2db.L2DB,
	txSelector *txselector.TxSelector,
	batchBuilder *batchbuilder.BatchBuilder,
	serverProofs []prover.Client,
	ethClient eth.ClientInterface,
	scConsts *synchronizer.SCConsts,
	initSCVars *synchronizer.SCVariables,
) (*Coordinator, error) {
	// nolint reason: hardcoded `1.0`, by design the percentage can't be over 100%
	if cfg.L1BatchTimeoutPerc >= 1.0 { //nolint:gomnd
		return nil, tracerr.Wrap(fmt.Errorf("invalid value for Config.L1BatchTimeoutPerc (%v >= 1.0)",
			cfg.L1BatchTimeoutPerc))
	}
	if cfg.EthClientAttempts < 1 {
		return nil, tracerr.Wrap(fmt.Errorf("invalid value for Config.EthClientAttempts (%v < 1)",
			cfg.EthClientAttempts))
	}

	if cfg.DebugBatchPath != "" {
		if err := os.MkdirAll(cfg.DebugBatchPath, 0744); err != nil {
			return nil, tracerr.Wrap(err)
		}
	}

	purger := Purger{
		cfg:                 cfg.Purger,
		lastPurgeBlock:      0,
		lastPurgeBatch:      0,
		lastInvalidateBlock: 0,
		lastInvalidateBatch: 0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := Coordinator{
		pipelineBatchNum: -1,
		provers:          serverProofs,
		consts:           *scConsts,
		vars:             *initSCVars,

		cfg: cfg,

		historyDB:    historyDB,
		l2DB:         l2DB,
		txSelector:   txSelector,
		batchBuilder: batchBuilder,

		purger: &purger,

		msgCh: make(chan interface{}),
		ctx:   ctx,
		// wg
		cancel: cancel,
	}
	ctxTimeout, ctxTimeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	defer ctxTimeoutCancel()
	txManager, err := NewTxManager(ctxTimeout, &cfg, ethClient, l2DB, &c,
		scConsts, initSCVars)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	c.txManager = txManager
	// Set Eth LastBlockNum to -1 in stats so that stats.Synced() is
	// guaranteed to return false before it's updated with a real stats
	c.stats.Eth.LastBlock.Num = -1
	return &c, nil
}

// TxSelector returns the inner TxSelector
func (c *Coordinator) TxSelector() *txselector.TxSelector {
	return c.txSelector
}

// BatchBuilder returns the inner BatchBuilder
func (c *Coordinator) BatchBuilder() *batchbuilder.BatchBuilder {
	return c.batchBuilder
}

func (c *Coordinator) newPipeline(ctx context.Context) (*Pipeline, error) {
	return NewPipeline(ctx, c.cfg, c.historyDB, c.l2DB, c.txSelector,
		c.batchBuilder, c.purger, c.txManager, c.provers, &c.consts)
}

// MsgSyncBlock indicates an update to the Synchronizer stats
type MsgSyncBlock struct {
	Stats   synchronizer.Stats
	Batches []common.BatchData
	// Vars contains each Smart Contract variables if they are updated, or
	// nil if they haven't changed.
	Vars synchronizer.SCVariablesPtr
}

// MsgSyncReorg indicates a reorg
type MsgSyncReorg struct {
	Stats synchronizer.Stats
	Vars  synchronizer.SCVariablesPtr
}

// MsgStopPipeline indicates a signal to reset the pipeline
type MsgStopPipeline struct {
	Reason string
}

// SendMsg is a thread safe method to pass a message to the Coordinator
func (c *Coordinator) SendMsg(ctx context.Context, msg interface{}) {
	select {
	case c.msgCh <- msg:
	case <-ctx.Done():
	}
}

func (c *Coordinator) syncSCVars(vars synchronizer.SCVariablesPtr) {
	if vars.Rollup != nil {
		c.vars.Rollup = *vars.Rollup
	}
	if vars.Auction != nil {
		c.vars.Auction = *vars.Auction
	}
	if vars.WDelayer != nil {
		c.vars.WDelayer = *vars.WDelayer
	}
}

func canForge(auctionConstants *common.AuctionConstants, auctionVars *common.AuctionVariables,
	currentSlot *common.Slot, nextSlot *common.Slot, addr ethCommon.Address, blockNum int64) bool {
	var slot *common.Slot
	if currentSlot.StartBlock <= blockNum && blockNum <= currentSlot.EndBlock {
		slot = currentSlot
	} else if nextSlot.StartBlock <= blockNum && blockNum <= nextSlot.EndBlock {
		slot = nextSlot
	} else {
		log.Warnw("Coordinator: requested blockNum for canForge is outside slot",
			"blockNum", blockNum, "currentSlot", currentSlot,
			"nextSlot", nextSlot,
		)
		return false
	}
	anyoneForge := false
	if !slot.ForgerCommitment &&
		auctionConstants.RelativeBlock(blockNum) >= int64(auctionVars.SlotDeadline) {
		log.Debugw("Coordinator: anyone can forge in the current slot (slotDeadline passed)",
			"block", blockNum)
		anyoneForge = true
	}
	if slot.Forger == addr || anyoneForge {
		return true
	}
	return false
}

func (c *Coordinator) canForge() bool {
	blockNum := c.stats.Eth.LastBlock.Num + 1
	return canForge(&c.consts.Auction, &c.vars.Auction,
		&c.stats.Sync.Auction.CurrentSlot, &c.stats.Sync.Auction.NextSlot,
		c.cfg.ForgerAddress, blockNum)
}

func (c *Coordinator) syncStats(ctx context.Context, stats *synchronizer.Stats) error {
	canForge := c.canForge()
	if c.pipeline == nil {
		if canForge {
			log.Infow("Coordinator: forging state begin", "block",
				stats.Eth.LastBlock.Num+1, "batch", stats.Sync.LastBatch)
			batchNum := common.BatchNum(stats.Sync.LastBatch)
			var err error
			if c.pipeline, err = c.newPipeline(ctx); err != nil {
				return tracerr.Wrap(err)
			}
			if err := c.pipeline.Start(batchNum, stats, &c.vars); err != nil {
				c.pipeline = nil
				return tracerr.Wrap(err)
			}
			c.pipelineBatchNum = batchNum
		}
	} else {
		if !canForge {
			log.Infow("Coordinator: forging state end", "block", stats.Eth.LastBlock.Num+1)
			c.pipeline.Stop(c.ctx)
			c.pipeline = nil
		}
	}
	if c.pipeline == nil {
		// Mark invalid in Pool due to forged L2Txs
		// for _, batch := range batches {
		// 	if err := c.l2DB.InvalidateOldNonces(
		// 		idxsNonceFromL2Txs(batch.L2Txs), batch.Batch.BatchNum); err != nil {
		// 		return err
		// 	}
		// }
		if c.purger.CanInvalidate(stats.Sync.LastBlock.Num, stats.Sync.LastBatch) {
			if err := c.txSelector.Reset(common.BatchNum(stats.Sync.LastBatch)); err != nil {
				return tracerr.Wrap(err)
			}
		}
		_, err := c.purger.InvalidateMaybe(c.l2DB, c.txSelector.LocalAccountsDB(),
			stats.Sync.LastBlock.Num, stats.Sync.LastBatch)
		if err != nil {
			return tracerr.Wrap(err)
		}
		_, err = c.purger.PurgeMaybe(c.l2DB, stats.Sync.LastBlock.Num, stats.Sync.LastBatch)
		if err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

func (c *Coordinator) handleMsgSyncBlock(ctx context.Context, msg *MsgSyncBlock) error {
	c.stats = msg.Stats
	c.syncSCVars(msg.Vars)
	c.txManager.SetSyncStatsVars(ctx, &msg.Stats, &msg.Vars)
	if c.pipeline != nil {
		c.pipeline.SetSyncStatsVars(ctx, &msg.Stats, &msg.Vars)
	}
	if !c.stats.Synced() {
		return nil
	}
	return c.syncStats(ctx, &c.stats)
}

func (c *Coordinator) handleReorg(ctx context.Context, msg *MsgSyncReorg) error {
	c.stats = msg.Stats
	c.syncSCVars(msg.Vars)
	c.txManager.SetSyncStatsVars(ctx, &msg.Stats, &msg.Vars)
	if c.pipeline != nil {
		c.pipeline.SetSyncStatsVars(ctx, &msg.Stats, &msg.Vars)
	}
	if common.BatchNum(c.stats.Sync.LastBatch) < c.pipelineBatchNum {
		// There's been a reorg and the batch from which the pipeline
		// was started was in a block that was discarded.  The batch
		// may not be in the main chain, so we stop the pipeline as a
		// precaution (it will be started again once the node is in
		// sync).
		log.Infow("Coordinator.handleReorg StopPipeline sync.LastBatch < c.pipelineBatchNum",
			"sync.LastBatch", c.stats.Sync.LastBatch,
			"c.pipelineBatchNum", c.pipelineBatchNum)
		if err := c.handleStopPipeline(ctx, "reorg"); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

func (c *Coordinator) handleStopPipeline(ctx context.Context, reason string) error {
	if c.pipeline != nil {
		c.pipeline.Stop(c.ctx)
		c.pipeline = nil
	}
	if err := c.l2DB.Reorg(common.BatchNum(c.stats.Sync.LastBatch)); err != nil {
		return tracerr.Wrap(err)
	}
	if strings.Contains(reason, common.AuctionErrMsgCannotForge) { //nolint:staticcheck
		// TODO: Check that we are in a slot in which we can't forge
	}
	return nil
}

func (c *Coordinator) handleMsg(ctx context.Context, msg interface{}) error {
	switch msg := msg.(type) {
	case MsgSyncBlock:
		if err := c.handleMsgSyncBlock(ctx, &msg); err != nil {
			return tracerr.Wrap(fmt.Errorf("Coordinator.handleMsgSyncBlock error: %w", err))
		}
	case MsgSyncReorg:
		if err := c.handleReorg(ctx, &msg); err != nil {
			return tracerr.Wrap(fmt.Errorf("Coordinator.handleReorg error: %w", err))
		}
	case MsgStopPipeline:
		log.Infow("Coordinator received MsgStopPipeline", "reason", msg.Reason)
		if err := c.handleStopPipeline(ctx, msg.Reason); err != nil {
			return tracerr.Wrap(fmt.Errorf("Coordinator.handleStopPipeline: %w", err))
		}
	default:
		log.Fatalw("Coordinator Unexpected Coordinator msg of type %T: %+v", msg, msg)
	}
	return nil
}

// Start the coordinator
func (c *Coordinator) Start() {
	if c.started {
		log.Fatal("Coordinator already started")
	}
	c.started = true
	c.wg.Add(1)
	go func() {
		c.txManager.Run(c.ctx)
		c.wg.Done()
	}()

	c.wg.Add(1)
	go func() {
		waitDuration := longWaitDuration
		for {
			select {
			case <-c.ctx.Done():
				log.Info("Coordinator done")
				c.wg.Done()
				return
			case msg := <-c.msgCh:
				if err := c.handleMsg(c.ctx, msg); c.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("Coordinator.handleMsg", "err", err)
					waitDuration = c.cfg.SyncRetryInterval
					continue
				}
				waitDuration = longWaitDuration
			case <-time.After(waitDuration):
				if !c.stats.Synced() {
					waitDuration = longWaitDuration
					continue
				}
				if err := c.syncStats(c.ctx, &c.stats); c.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("Coordinator.syncStats", "err", err)
					waitDuration = c.cfg.SyncRetryInterval
					continue
				}
				waitDuration = longWaitDuration
			}
		}
	}()
}

const stopCtxTimeout = 200 * time.Millisecond

// Stop the coordinator
func (c *Coordinator) Stop() {
	if !c.started {
		log.Fatal("Coordinator already stopped")
	}
	c.started = false
	log.Infow("Stopping Coordinator...")
	c.cancel()
	c.wg.Wait()
	if c.pipeline != nil {
		ctx, cancel := context.WithTimeout(context.Background(), stopCtxTimeout)
		defer cancel()
		c.pipeline.Stop(ctx)
		c.pipeline = nil
	}
}
