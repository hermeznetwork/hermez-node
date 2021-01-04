package coordinator

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

const queueLen = 16

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
	stats            *synchronizer.Stats
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

		// ethClient: ethClient,

		msgCh: make(chan interface{}),
		ctx:   ctx,
		// wg
		cancel: cancel,
	}
	txManager := NewTxManager(&cfg, ethClient, l2DB, &c)
	c.txManager = txManager
	return &c, nil
}

func (c *Coordinator) newPipeline(ctx context.Context,
	stats *synchronizer.Stats) (*Pipeline, error) {
	return NewPipeline(ctx, c.cfg, c.historyDB, c.l2DB, c.txSelector,
		c.batchBuilder, c.purger, c.txManager, c.provers, stats, &c.consts)
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
}

// MsgStopPipeline indicates a signal to reset the pipeline
type MsgStopPipeline struct {
	Reason string
}

// SendMsg is a thread safe method to pass a message to the Coordinator
func (c *Coordinator) SendMsg(msg interface{}) {
	c.msgCh <- msg
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

func (c *Coordinator) canForge(stats *synchronizer.Stats) bool {
	anyoneForge := false
	if !stats.Sync.Auction.CurrentSlot.ForgerCommitment &&
		c.consts.Auction.RelativeBlock(stats.Eth.LastBlock.Num+1) >= int64(c.vars.Auction.SlotDeadline) {
		log.Debugw("Coordinator: anyone can forge in the current slot (slotDeadline passed)",
			"block", stats.Eth.LastBlock.Num)
		anyoneForge = true
	}
	if stats.Sync.Auction.CurrentSlot.Forger == c.cfg.ForgerAddress || anyoneForge {
		return true
	}
	return false
}

func (c *Coordinator) syncStats(ctx context.Context, stats *synchronizer.Stats) error {
	c.txManager.SetLastBlock(stats.Eth.LastBlock.Num)

	canForge := c.canForge(stats)
	if c.pipeline == nil {
		if canForge {
			log.Infow("Coordinator: forging state begin", "block",
				stats.Eth.LastBlock.Num, "batch", stats.Sync.LastBatch)
			batchNum := common.BatchNum(stats.Sync.LastBatch)
			var err error
			if c.pipeline, err = c.newPipeline(ctx, stats); err != nil {
				return tracerr.Wrap(err)
			}
			if err := c.pipeline.Start(batchNum, stats.Sync.LastForgeL1TxsNum,
				stats, &c.vars); err != nil {
				c.pipeline = nil
				return tracerr.Wrap(err)
			}
			c.pipelineBatchNum = batchNum
		}
	} else {
		if canForge {
			c.pipeline.SetSyncStats(stats)
		} else {
			log.Infow("Coordinator: forging state end", "block", stats.Eth.LastBlock.Num)
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
	c.stats = &msg.Stats
	// batches := msg.Batches
	if !c.stats.Synced() {
		return nil
	}
	c.syncSCVars(msg.Vars)
	return c.syncStats(ctx, c.stats)
}

func (c *Coordinator) handleStopPipeline(ctx context.Context, reason string) error {
	if c.pipeline != nil {
		c.pipeline.Stop(c.ctx)
		c.pipeline = nil
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
		if err := c.handleReorg(ctx, &msg.Stats); err != nil {
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
		waitDuration := time.Duration(longWaitDuration)
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
					waitDuration = time.Duration(c.cfg.SyncRetryInterval)
					continue
				}
				waitDuration = time.Duration(longWaitDuration)
			case <-time.After(waitDuration):
				if c.stats == nil {
					waitDuration = time.Duration(longWaitDuration)
					continue
				}
				if err := c.syncStats(c.ctx, c.stats); c.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("Coordinator.syncStats", "err", err)
					waitDuration = time.Duration(c.cfg.SyncRetryInterval)
					continue
				}
				waitDuration = time.Duration(longWaitDuration)
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

func (c *Coordinator) handleReorg(ctx context.Context, stats *synchronizer.Stats) error {
	c.stats = stats
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
		if err := c.l2DB.Reorg(common.BatchNum(c.stats.Sync.LastBatch)); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// TxManager handles everything related to ethereum transactions:  It makes the
// call to forge, waits for transaction confirmation, and keeps checking them
// until a number of confirmed blocks have passed.
type TxManager struct {
	cfg         Config
	ethClient   eth.ClientInterface
	l2DB        *l2db.L2DB   // Used only to mark forged txs as forged in the L2DB
	coord       *Coordinator // Used only to send messages to stop the pipeline
	batchCh     chan *BatchInfo
	lastBlockCh chan int64
	queue       []*BatchInfo
	lastBlock   int64
	// lastConfirmedBatch stores the last BatchNum that who's forge call was confirmed
	lastConfirmedBatch common.BatchNum
}

// NewTxManager creates a new TxManager
func NewTxManager(cfg *Config, ethClient eth.ClientInterface, l2DB *l2db.L2DB,
	coord *Coordinator) *TxManager {
	return &TxManager{
		cfg:         *cfg,
		ethClient:   ethClient,
		l2DB:        l2DB,
		coord:       coord,
		batchCh:     make(chan *BatchInfo, queueLen),
		lastBlockCh: make(chan int64, queueLen),
		lastBlock:   -1,
	}
}

// AddBatch is a thread safe method to pass a new batch TxManager to be sent to
// the smart contract via the forge call
func (t *TxManager) AddBatch(batchInfo *BatchInfo) {
	t.batchCh <- batchInfo
}

// SetLastBlock is a thread safe method to pass the lastBlock to the TxManager
func (t *TxManager) SetLastBlock(lastBlock int64) {
	t.lastBlockCh <- lastBlock
}

func (t *TxManager) rollupForgeBatch(ctx context.Context, batchInfo *BatchInfo) error {
	var ethTx *types.Transaction
	var err error
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		ethTx, err = t.ethClient.RollupForgeBatch(batchInfo.ForgeBatchArgs)
		if err != nil {
			if strings.Contains(err.Error(), common.AuctionErrMsgCannotForge) {
				log.Debugw("TxManager ethClient.RollupForgeBatch", "err", err,
					"block", t.lastBlock)
				return tracerr.Wrap(err)
			}
			log.Errorw("TxManager ethClient.RollupForgeBatch",
				"attempt", attempt, "err", err, "block", t.lastBlock,
				"batchNum", batchInfo.BatchNum)
		} else {
			break
		}
		select {
		case <-ctx.Done():
			return tracerr.Wrap(common.ErrDone)
		case <-time.After(t.cfg.EthClientAttemptsDelay):
		}
	}
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("reached max attempts for ethClient.RollupForgeBatch: %w", err))
	}
	batchInfo.EthTx = ethTx
	log.Infow("TxManager ethClient.RollupForgeBatch", "batch", batchInfo.BatchNum, "tx", ethTx.Hash().Hex())
	t.cfg.debugBatchStore(batchInfo)
	if err := t.l2DB.DoneForging(common.TxIDsFromL2Txs(batchInfo.L2Txs), batchInfo.BatchNum); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (t *TxManager) ethTransactionReceipt(ctx context.Context, batchInfo *BatchInfo) error {
	txHash := batchInfo.EthTx.Hash()
	var receipt *types.Receipt
	var err error
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		receipt, err = t.ethClient.EthTransactionReceipt(ctx, txHash)
		if ctx.Err() != nil {
			continue
		}
		if err != nil {
			log.Errorw("TxManager ethClient.EthTransactionReceipt",
				"attempt", attempt, "err", err)
		} else {
			break
		}
		select {
		case <-ctx.Done():
			return tracerr.Wrap(common.ErrDone)
		case <-time.After(t.cfg.EthClientAttemptsDelay):
		}
	}
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("reached max attempts for ethClient.EthTransactionReceipt: %w", err))
	}
	batchInfo.Receipt = receipt
	t.cfg.debugBatchStore(batchInfo)
	return nil
}

func (t *TxManager) handleReceipt(batchInfo *BatchInfo) (*int64, error) {
	receipt := batchInfo.Receipt
	if receipt != nil {
		if receipt.Status == types.ReceiptStatusFailed {
			log.Errorw("TxManager receipt status is failed", "receipt", receipt)
			return nil, tracerr.Wrap(fmt.Errorf("ethereum transaction receipt statis is failed"))
		} else if receipt.Status == types.ReceiptStatusSuccessful {
			if batchInfo.BatchNum > t.lastConfirmedBatch {
				t.lastConfirmedBatch = batchInfo.BatchNum
			}
			confirm := t.lastBlock - receipt.BlockNumber.Int64()
			return &confirm, nil
		}
	}
	return nil, nil
}

const longWaitDuration = 999 * time.Hour

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	next := 0
	waitDuration := time.Duration(longWaitDuration)

	for {
		select {
		case <-ctx.Done():
			log.Info("TxManager done")
			return
		case lastBlock := <-t.lastBlockCh:
			t.lastBlock = lastBlock
		case batchInfo := <-t.batchCh:
			if err := t.rollupForgeBatch(ctx, batchInfo); common.IsErrDone(err) {
				continue
			} else if err != nil {
				t.coord.SendMsg(MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch call: %v", err)})
				continue
			}
			log.Debugf("ethClient ForgeCall sent, batchNum: %d", batchInfo.BatchNum)
			t.queue = append(t.queue, batchInfo)
			waitDuration = t.cfg.TxManagerCheckInterval
		case <-time.After(waitDuration):
			if len(t.queue) == 0 {
				continue
			}
			current := next
			next = (current + 1) % len(t.queue)
			batchInfo := t.queue[current]
			err := t.ethTransactionReceipt(ctx, batchInfo)
			if common.IsErrDone(err) {
				continue
			} else if err != nil { //nolint:staticcheck
				// We can't get the receipt for the
				// transaction, so we can't confirm if it was
				// mined
				t.coord.SendMsg(MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch receipt: %v", err)})
			}

			confirm, err := t.handleReceipt(batchInfo)
			if err != nil { //nolint:staticcheck
				// Transaction was rejected
				t.coord.SendMsg(MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch reject: %v", err)})
			}
			if confirm != nil && *confirm >= t.cfg.ConfirmBlocks {
				log.Debugw("TxManager tx for RollupForgeBatch confirmed",
					"batch", batchInfo.BatchNum)
				t.queue = append(t.queue[:current], t.queue[current+1:]...)
				if len(t.queue) == 0 {
					waitDuration = longWaitDuration
					next = 0
				} else {
					next = current % len(t.queue)
				}
			}
		}
	}
}

// Pipeline manages the forging of batches with parallel server proofs
type Pipeline struct {
	cfg    Config
	consts synchronizer.SCConsts

	// state
	batchNum                     common.BatchNum
	vars                         synchronizer.SCVariables
	lastScheduledL1BatchBlockNum int64
	lastForgeL1TxsNum            int64
	started                      bool

	proversPool  *ProversPool
	provers      []prover.Client
	txManager    *TxManager
	historyDB    *historydb.HistoryDB
	l2DB         *l2db.L2DB
	txSelector   *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder
	purger       *Purger

	stats   synchronizer.Stats
	statsCh chan synchronizer.Stats

	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewPipeline creates a new Pipeline
func NewPipeline(ctx context.Context,
	cfg Config,
	historyDB *historydb.HistoryDB,
	l2DB *l2db.L2DB,
	txSelector *txselector.TxSelector,
	batchBuilder *batchbuilder.BatchBuilder,
	purger *Purger,
	txManager *TxManager,
	provers []prover.Client,
	stats *synchronizer.Stats,
	scConsts *synchronizer.SCConsts,
) (*Pipeline, error) {
	proversPool := NewProversPool(len(provers))
	proversPoolSize := 0
	for _, prover := range provers {
		if err := prover.WaitReady(ctx); err != nil {
			log.Errorw("prover.WaitReady", "err", err)
		} else {
			proversPool.Add(prover)
			proversPoolSize++
		}
	}
	if proversPoolSize == 0 {
		return nil, tracerr.Wrap(fmt.Errorf("no provers in the pool"))
	}
	return &Pipeline{
		cfg:          cfg,
		historyDB:    historyDB,
		l2DB:         l2DB,
		txSelector:   txSelector,
		batchBuilder: batchBuilder,
		provers:      provers,
		proversPool:  proversPool,
		purger:       purger,
		txManager:    txManager,
		consts:       *scConsts,
		stats:        *stats,
		statsCh:      make(chan synchronizer.Stats, queueLen),
	}, nil
}

// SetSyncStats is a thread safe method to sets the synchronizer Stats
func (p *Pipeline) SetSyncStats(stats *synchronizer.Stats) {
	p.statsCh <- *stats
}

// reset pipeline state
func (p *Pipeline) reset(batchNum common.BatchNum, lastForgeL1TxsNum int64,
	initSCVars *synchronizer.SCVariables) error {
	p.batchNum = batchNum
	p.lastForgeL1TxsNum = lastForgeL1TxsNum
	p.vars = *initSCVars
	p.lastScheduledL1BatchBlockNum = 0

	err := p.txSelector.Reset(p.batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = p.batchBuilder.Reset(p.batchNum, true)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// Start the forging pipeline
func (p *Pipeline) Start(batchNum common.BatchNum, lastForgeL1TxsNum int64,
	syncStats *synchronizer.Stats, initSCVars *synchronizer.SCVariables) error {
	if p.started {
		log.Fatal("Pipeline already started")
	}
	p.started = true

	if err := p.reset(batchNum, lastForgeL1TxsNum, initSCVars); err != nil {
		return tracerr.Wrap(err)
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())

	queueSize := 1
	batchChSentServerProof := make(chan *BatchInfo, queueSize)

	p.wg.Add(1)
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				log.Info("Pipeline forgeBatch loop done")
				p.wg.Done()
				return
			case syncStats := <-p.statsCh:
				p.stats = syncStats
			default:
				batchNum = p.batchNum + 1
				batchInfo, err := p.forgeBatch(batchNum)
				if p.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("forgeBatch", "err", err)
					continue
				}
				// 6. Wait for an available server proof (blocking call)
				serverProof, err := p.proversPool.Get(p.ctx)
				if p.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("proversPool.Get", "err", err)
					continue
				}
				batchInfo.ServerProof = serverProof
				if err := p.sendServerProof(p.ctx, batchInfo); p.ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("sendServerProof", "err", err)
					batchInfo.ServerProof = nil
					p.proversPool.Add(serverProof)
					continue
				}
				p.batchNum = batchNum
				batchChSentServerProof <- batchInfo
			}
		}
	}()

	p.wg.Add(1)
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				log.Info("Pipeline waitServerProofSendEth loop done")
				p.wg.Done()
				return
			case batchInfo := <-batchChSentServerProof:
				err := p.waitServerProof(p.ctx, batchInfo)
				// We are done with this serverProof, add it back to the pool
				p.proversPool.Add(batchInfo.ServerProof)
				batchInfo.ServerProof = nil
				if p.ctx.Err() != nil {
					continue
				}
				if err != nil {
					log.Errorw("waitServerProof", "err", err)
					continue
				}
				p.txManager.AddBatch(batchInfo)
			}
		}
	}()
	return nil
}

// Stop the forging pipeline
func (p *Pipeline) Stop(ctx context.Context) {
	if !p.started {
		log.Fatal("Pipeline already stopped")
	}
	p.started = false
	log.Info("Stopping Pipeline...")
	p.cancel()
	p.wg.Wait()
	for _, prover := range p.provers {
		if err := prover.Cancel(ctx); err != nil {
			log.Errorw("prover.Cancel", "err", err)
		}
	}
}

// sendServerProof sends the circuit inputs to the proof server
func (p *Pipeline) sendServerProof(ctx context.Context, batchInfo *BatchInfo) error {
	p.cfg.debugBatchStore(batchInfo)

	// 7. Call the selected idle server proof with BatchBuilder output,
	// save server proof info for batchNum
	if err := batchInfo.ServerProof.CalculateProof(ctx, batchInfo.ZKInputs); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// forgeBatch the next batch.
func (p *Pipeline) forgeBatch(batchNum common.BatchNum) (*BatchInfo, error) {
	// remove transactions from the pool that have been there for too long
	_, err := p.purger.InvalidateMaybe(p.l2DB, p.txSelector.LocalAccountsDB(),
		p.stats.Sync.LastBlock.Num, int64(batchNum))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	_, err = p.purger.PurgeMaybe(p.l2DB, p.stats.Sync.LastBlock.Num, int64(batchNum))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	batchInfo := BatchInfo{BatchNum: batchNum} // to accumulate metadata of the batch

	selectionCfg := &txselector.SelectionConfig{
		MaxL1UserTxs:      common.RollupConstMaxL1UserTx,
		TxProcessorConfig: p.cfg.TxProcessorConfig,
	}

	var poolL2Txs []common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1CoordTxs []common.L1Tx
	var auths [][]byte
	var coordIdxs []common.Idx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if p.shouldL1L2Batch() {
		batchInfo.L1Batch = true
		p.lastScheduledL1BatchBlockNum = p.stats.Eth.LastBlock.Num
		// 2a: L1+L2 txs
		p.lastForgeL1TxsNum++
		l1UserTxs, err := p.historyDB.GetUnforgedL1UserTxs(p.lastForgeL1TxsNum)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		coordIdxs, auths, l1UserTxsExtra, l1CoordTxs, poolL2Txs, err =
			p.txSelector.GetL1L2TxSelection(selectionCfg, batchNum, l1UserTxs)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	} else {
		// 2b: only L2 txs
		coordIdxs, auths, l1CoordTxs, poolL2Txs, err =
			p.txSelector.GetL2TxSelection(selectionCfg, batchNum)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1UserTxsExtra = nil
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	batchInfo.L1UserTxsExtra = l1UserTxsExtra
	batchInfo.L1CoordTxs = l1CoordTxs
	batchInfo.L1CoordinatorTxsAuths = auths
	batchInfo.CoordIdxs = coordIdxs
	batchInfo.VerifierIdx = p.cfg.VerifierIdx

	if err := p.l2DB.StartForging(common.TxIDsFromPoolL2Txs(poolL2Txs), batchInfo.BatchNum); err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Invalidate transactions that become invalid beause of
	// the poolL2Txs selected.  Will mark as invalid the txs that have a
	// (fromIdx, nonce) which already appears in the selected txs (includes
	// all the nonces smaller than the current one)
	err = p.l2DB.InvalidateOldNonces(idxsNonceFromPoolL2Txs(poolL2Txs), batchInfo.BatchNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress:     p.cfg.ForgerAddress,
		TxProcessorConfig: p.cfg.TxProcessorConfig,
	}
	zkInputs, err := p.batchBuilder.BuildBatch(coordIdxs, configBatch, l1UserTxsExtra,
		l1CoordTxs, poolL2Txs, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	l2Txs, err := common.PoolL2TxsToL2Txs(poolL2Txs) // NOTE: This is a big uggly, find a better way
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	batchInfo.L2Txs = l2Txs

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.ZKInputs = zkInputs
	p.cfg.debugBatchStore(&batchInfo)

	return &batchInfo, nil
}

// waitServerProof gets the generated zkProof & sends it to the SmartContract
func (p *Pipeline) waitServerProof(ctx context.Context, batchInfo *BatchInfo) error {
	proof, pubInputs, err := batchInfo.ServerProof.GetProof(ctx) // blocking call, until not resolved don't continue. Returns when the proof server has calculated the proof
	if err != nil {
		return tracerr.Wrap(err)
	}
	batchInfo.Proof = proof
	batchInfo.PublicInputs = pubInputs
	batchInfo.ForgeBatchArgs = prepareForgeBatchArgs(batchInfo)
	batchInfo.TxStatus = TxStatusPending
	p.cfg.debugBatchStore(batchInfo)
	return nil
}

func (p *Pipeline) shouldL1L2Batch() bool {
	// Take the lastL1BatchBlockNum as the biggest between the last
	// scheduled one, and the synchronized one.
	lastL1BatchBlockNum := p.lastScheduledL1BatchBlockNum
	if p.stats.Sync.LastL1BatchBlock > lastL1BatchBlockNum {
		lastL1BatchBlockNum = p.stats.Sync.LastL1BatchBlock
	}
	// Return true if we have passed the l1BatchTimeoutPerc portion of the
	// range before the l1batch timeout.
	if p.stats.Eth.LastBlock.Num-lastL1BatchBlockNum >=
		int64(float64(p.vars.Rollup.ForgeL1L2BatchTimeout)*p.cfg.L1BatchTimeoutPerc) {
		return true
	}
	return false
}

func prepareForgeBatchArgs(batchInfo *BatchInfo) *eth.RollupForgeBatchArgs {
	proof := batchInfo.Proof
	zki := batchInfo.ZKInputs
	return &eth.RollupForgeBatchArgs{
		NewLastIdx:            int64(zki.Metadata.NewLastIdxRaw),
		NewStRoot:             zki.Metadata.NewStateRootRaw.BigInt(),
		NewExitRoot:           zki.Metadata.NewExitRootRaw.BigInt(),
		L1UserTxs:             batchInfo.L1UserTxsExtra,
		L1CoordinatorTxs:      batchInfo.L1CoordTxs,
		L1CoordinatorTxsAuths: batchInfo.L1CoordinatorTxsAuths,
		L2TxsData:             batchInfo.L2Txs,
		FeeIdxCoordinator:     batchInfo.CoordIdxs,
		// Circuit selector
		VerifierIdx: batchInfo.VerifierIdx,
		L1Batch:     batchInfo.L1Batch,
		ProofA:      [2]*big.Int{proof.PiA[0], proof.PiA[1]},
		ProofB: [2][2]*big.Int{
			{proof.PiB[0][0], proof.PiB[0][1]},
			{proof.PiB[1][0], proof.PiB[1][1]},
		},
		ProofC: [2]*big.Int{proof.PiC[0], proof.PiC[1]},
	}
}
