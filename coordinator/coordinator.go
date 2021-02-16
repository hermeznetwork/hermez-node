package coordinator

import (
	"context"
	"fmt"
	"math/big"
	"os"
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
	// StartSlotBlocksDelay is the number of blocks of delay to wait before
	// starting the pipeline when we reach a slot in which we can forge.
	StartSlotBlocksDelay int64
	// ScheduleBatchBlocksAheadCheck is the number of blocks ahead in which
	// the forger address is checked to be allowed to forge (apart from
	// checking the next block), used to decide when to stop scheduling new
	// batches (by stopping the pipeline).
	// For example, if we are at block 10 and ScheduleBatchBlocksAheadCheck
	// is 5, eventhough at block 11 we canForge, the pipeline will be
	// stopped if we can't forge at block 15.
	// This value should be the expected number of blocks it takes between
	// scheduling a batch and having it mined.
	ScheduleBatchBlocksAheadCheck int64
	// SendBatchBlocksMarginCheck is the number of margin blocks ahead in
	// which the coordinator is also checked to be allowed to forge, apart
	// from the next block; used to decide when to stop sending batches to
	// the smart contract.
	// For example, if we are at block 10 and SendBatchBlocksMarginCheck is
	// 5, eventhough at block 11 we canForge, the batch will be discarded
	// if we can't forge at block 15.
	// This value should be the expected number of blocks it takes between
	// sending a batch and having it mined.
	SendBatchBlocksMarginCheck int64
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
	// EthTxResendTimeout is the timeout after which a non-mined ethereum
	// transaction will be resent (reusing the nonce) with a newly
	// calculated gas price
	EthTxResendTimeout time.Duration
	// EthNoReuseNonce disables reusing nonces of pending transactions for
	// new replacement transactions
	EthNoReuseNonce bool
	// MaxGasPrice is the maximum gas price allowed for ethereum
	// transactions
	MaxGasPrice *big.Int
	// TxManagerCheckInterval is the waiting interval between receipt
	// checks of ethereum transactions in the TxManager
	TxManagerCheckInterval time.Duration
	// DebugBatchPath if set, specifies the path where batchInfo is stored
	// in JSON in every step/update of the pipeline
	DebugBatchPath string
	Purger         PurgerCfg
	// VerifierIdx is the index of the verifier contract registered in the
	// smart contract
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

type fromBatch struct {
	BatchNum   common.BatchNum
	ForgerAddr ethCommon.Address
	StateRoot  *big.Int
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	// State
	pipelineNum       int       // Pipeline sequential number.  The first pipeline is 1
	pipelineFromBatch fromBatch // batch from which we started the pipeline
	provers           []prover.Client
	consts            synchronizer.SCConsts
	vars              synchronizer.SCVariables
	stats             synchronizer.Stats
	started           bool

	cfg Config

	historyDB    *historydb.HistoryDB
	l2DB         *l2db.L2DB
	txSelector   *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	msgCh  chan interface{}
	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc

	pipeline              *Pipeline
	lastNonFailedBatchNum common.BatchNum

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
		pipelineNum: 0,
		pipelineFromBatch: fromBatch{
			BatchNum:   0,
			ForgerAddr: ethCommon.Address{},
			StateRoot:  big.NewInt(0),
		},
		provers: serverProofs,
		consts:  *scConsts,
		vars:    *initSCVars,

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
	c.pipelineNum++
	return NewPipeline(ctx, c.cfg, c.pipelineNum, c.historyDB, c.l2DB, c.txSelector,
		c.batchBuilder, c.purger, c, c.txManager, c.provers, &c.consts)
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
	// FailedBatchNum indicates the first batchNum that failed in the
	// pipeline.  If FailedBatchNum is 0, it should be ignored.
	FailedBatchNum common.BatchNum
}

// SendMsg is a thread safe method to pass a message to the Coordinator
func (c *Coordinator) SendMsg(ctx context.Context, msg interface{}) {
	select {
	case c.msgCh <- msg:
	case <-ctx.Done():
	}
}

func updateSCVars(vars *synchronizer.SCVariables, update synchronizer.SCVariablesPtr) {
	if update.Rollup != nil {
		vars.Rollup = *update.Rollup
	}
	if update.Auction != nil {
		vars.Auction = *update.Auction
	}
	if update.WDelayer != nil {
		vars.WDelayer = *update.WDelayer
	}
}

func (c *Coordinator) syncSCVars(vars synchronizer.SCVariablesPtr) {
	updateSCVars(&c.vars, vars)
}

func canForge(auctionConstants *common.AuctionConstants, auctionVars *common.AuctionVariables,
	currentSlot *common.Slot, nextSlot *common.Slot, addr ethCommon.Address, blockNum int64) bool {
	if blockNum < auctionConstants.GenesisBlockNum {
		log.Infow("canForge: requested blockNum is < genesis", "blockNum", blockNum,
			"genesis", auctionConstants.GenesisBlockNum)
		return false
	}
	var slot *common.Slot
	if currentSlot.StartBlock <= blockNum && blockNum <= currentSlot.EndBlock {
		slot = currentSlot
	} else if nextSlot.StartBlock <= blockNum && blockNum <= nextSlot.EndBlock {
		slot = nextSlot
	} else {
		log.Warnw("canForge: requested blockNum is outside current and next slot",
			"blockNum", blockNum, "currentSlot", currentSlot,
			"nextSlot", nextSlot,
		)
		return false
	}
	anyoneForge := false
	if !slot.ForgerCommitment &&
		auctionConstants.RelativeBlock(blockNum) >= int64(auctionVars.SlotDeadline) {
		log.Debugw("canForge: anyone can forge in the current slot (slotDeadline passed)",
			"block", blockNum)
		anyoneForge = true
	}
	if slot.Forger == addr || anyoneForge {
		return true
	}
	log.Debugw("canForge: can't forge", "slot.Forger", slot.Forger)
	return false
}

func (c *Coordinator) canForgeAt(blockNum int64) bool {
	return canForge(&c.consts.Auction, &c.vars.Auction,
		&c.stats.Sync.Auction.CurrentSlot, &c.stats.Sync.Auction.NextSlot,
		c.cfg.ForgerAddress, blockNum)
}

func (c *Coordinator) canForge() bool {
	blockNum := c.stats.Eth.LastBlock.Num + 1
	return canForge(&c.consts.Auction, &c.vars.Auction,
		&c.stats.Sync.Auction.CurrentSlot, &c.stats.Sync.Auction.NextSlot,
		c.cfg.ForgerAddress, blockNum)
}

func (c *Coordinator) syncStats(ctx context.Context, stats *synchronizer.Stats) error {
	nextBlock := c.stats.Eth.LastBlock.Num + 1
	canForge := c.canForgeAt(nextBlock)
	if c.cfg.ScheduleBatchBlocksAheadCheck != 0 && canForge {
		canForge = c.canForgeAt(nextBlock + c.cfg.ScheduleBatchBlocksAheadCheck)
	}
	if c.pipeline == nil {
		relativeBlock := c.consts.Auction.RelativeBlock(nextBlock)
		if canForge && relativeBlock < c.cfg.StartSlotBlocksDelay {
			log.Debugf("Coordinator: delaying pipeline start due to "+
				"relativeBlock (%v) < cfg.StartSlotBlocksDelay (%v)",
				relativeBlock, c.cfg.StartSlotBlocksDelay)
		} else if canForge {
			log.Infow("Coordinator: forging state begin", "block",
				stats.Eth.LastBlock.Num+1, "batch", stats.Sync.LastBatch.BatchNum)
			batchNum := stats.Sync.LastBatch.BatchNum
			if c.lastNonFailedBatchNum > batchNum {
				batchNum = c.lastNonFailedBatchNum
			}
			var err error
			if c.pipeline, err = c.newPipeline(ctx); err != nil {
				return tracerr.Wrap(err)
			}
			if err := c.pipeline.Start(batchNum, stats, &c.vars); err != nil {
				c.pipeline = nil
				return tracerr.Wrap(err)
			}
		}
	} else {
		if !canForge {
			log.Infow("Coordinator: forging state end", "block", stats.Eth.LastBlock.Num+1)
			c.pipeline.Stop(c.ctx)
			c.pipeline = nil
		}
	}
	if c.pipeline == nil {
		if _, err := c.purger.InvalidateMaybe(c.l2DB, c.txSelector.LocalAccountsDB(),
			stats.Sync.LastBlock.Num, int64(stats.Sync.LastBatch.BatchNum)); err != nil {
			return tracerr.Wrap(err)
		}
		if _, err := c.purger.PurgeMaybe(c.l2DB, stats.Sync.LastBlock.Num,
			int64(stats.Sync.LastBatch.BatchNum)); err != nil {
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
	if c.stats.Sync.LastBatch.ForgerAddr != c.cfg.ForgerAddress &&
		c.stats.Sync.LastBatch.StateRoot.Cmp(c.pipelineFromBatch.StateRoot) != 0 {
		// There's been a reorg and the batch state root from which the
		// pipeline was started has changed (probably because it was in
		// a block that was discarded), and it was sent by a different
		// coordinator than us.  That batch may never be in the main
		// chain, so we stop the pipeline  (it will be started again
		// once the node is in sync).
		log.Infow("Coordinator.handleReorg StopPipeline sync.LastBatch.ForgerAddr != cfg.ForgerAddr "+
			"& sync.LastBatch.StateRoot != pipelineFromBatch.StateRoot",
			"sync.LastBatch.StateRoot", c.stats.Sync.LastBatch.StateRoot,
			"pipelineFromBatch.StateRoot", c.pipelineFromBatch.StateRoot)
		c.txManager.DiscardPipeline(ctx, c.pipelineNum)
		if err := c.handleStopPipeline(ctx, "reorg", 0); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// handleStopPipeline handles stopping the pipeline.  If failedBatchNum is 0,
// the next pipeline will start from the last state of the synchronizer,
// otherwise, it will state from failedBatchNum-1.
func (c *Coordinator) handleStopPipeline(ctx context.Context, reason string, failedBatchNum common.BatchNum) error {
	batchNum := c.stats.Sync.LastBatch.BatchNum
	if failedBatchNum != 0 {
		batchNum = failedBatchNum - 1
	}
	if c.pipeline != nil {
		c.pipeline.Stop(c.ctx)
		c.pipeline = nil
	}
	if err := c.l2DB.Reorg(batchNum); err != nil {
		return tracerr.Wrap(err)
	}
	c.lastNonFailedBatchNum = batchNum
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
		if err := c.handleStopPipeline(ctx, msg.Reason, msg.FailedBatchNum); err != nil {
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
