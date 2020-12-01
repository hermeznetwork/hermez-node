package coordinator

import (
	"context"
	"fmt"
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
	// EthClientAttemptsDelay is delay between attempts do do an eth client
	// RPC call
	EthClientAttemptsDelay time.Duration
	// TxManagerCheckInterval is the waiting interval between receipt
	// checks of ethereum transactions in the TxManager
	TxManagerCheckInterval time.Duration
	// DebugBatchPath if set, specifies the path where batchInfo is stored
	// in JSON in every step/update of the pipeline
	DebugBatchPath string
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

func (c *Coordinator) newPipeline(ctx context.Context) (*Pipeline, error) {
	return NewPipeline(ctx, c.cfg, c.historyDB, c.l2DB,
		c.txSelector, c.batchBuilder, c.txManager, c.provers, &c.consts)
}

// MsgSyncStats indicates an update to the Synchronizer stats
type MsgSyncStats struct {
	Stats synchronizer.Stats
}

// MsgSyncSCVars indicates an update to Smart Contract Vars
type MsgSyncSCVars struct {
	Rollup   *common.RollupVariables
	Auction  *common.AuctionVariables
	WDelayer *common.WDelayerVariables
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

func (c *Coordinator) handleMsgSyncSCVars(msg *MsgSyncSCVars) {
	if msg.Rollup != nil {
		c.vars.Rollup = *msg.Rollup
	}
	if msg.Auction != nil {
		c.vars.Auction = *msg.Auction
	}
	if msg.WDelayer != nil {
		c.vars.WDelayer = *msg.WDelayer
	}
}

func (c *Coordinator) canForge(stats *synchronizer.Stats) bool {
	anyoneForge := false
	if stats.Sync.Auction.CurrentSlot.BatchesLen == 0 &&
		c.consts.Auction.RelativeBlock(stats.Eth.LastBlock.Num+1) > int64(c.vars.Auction.SlotDeadline) {
		log.Debug("Coordinator: anyone can forge in the current slot (slotDeadline passed)")
		anyoneForge = true
	}
	if stats.Sync.Auction.CurrentSlot.Forger == c.cfg.ForgerAddress || anyoneForge {
		return true
	}
	return false
}

func (c *Coordinator) handleMsgSyncStats(ctx context.Context, stats *synchronizer.Stats) error {
	if !stats.Synced() {
		return nil
	}
	c.txManager.SetLastBlock(stats.Eth.LastBlock.Num)

	canForge := c.canForge(stats)
	if c.pipeline == nil {
		if canForge {
			log.Infow("Coordinator: forging state begin", "block", stats.Eth.LastBlock.Num,
				"batch", stats.Sync.LastBatch)
			batchNum := common.BatchNum(stats.Sync.LastBatch)
			var err error
			if c.pipeline, err = c.newPipeline(ctx); err != nil {
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
	return nil
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
		for {
			select {
			case <-c.ctx.Done():
				log.Info("Coordinator done")
				c.wg.Done()
				return
			case msg := <-c.msgCh:
				switch msg := msg.(type) {
				case MsgSyncStats:
					stats := msg.Stats
					if err := c.handleMsgSyncStats(c.ctx, &stats); common.IsErrDone(err) {
						continue
					} else if err != nil {
						log.Errorw("Coordinator.handleMsgSyncStats error", "err", err)
						continue
					}
				case MsgSyncReorg:
					if err := c.handleReorg(c.ctx, &msg.Stats); common.IsErrDone(err) {
						continue
					} else if err != nil {
						log.Errorw("Coordinator.handleReorg error", "err", err)
						continue
					}
				case MsgStopPipeline:
					log.Infow("Coordinator received MsgStopPipeline", "reason", msg.Reason)
					if err := c.handleStopPipeline(c.ctx, msg.Reason); common.IsErrDone(err) {
						continue
					} else if err != nil {
						log.Errorw("Coordinator.handleStopPipeline", "err", err)
					}
				case MsgSyncSCVars:
					c.handleMsgSyncSCVars(&msg)
				default:
					log.Fatalw("Coordinator Unexpected Coordinator msg of type %T: %+v", msg, msg)
				}
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
	if common.BatchNum(stats.Sync.LastBatch) < c.pipelineBatchNum {
		// There's been a reorg and the batch from which the pipeline
		// was started was in a block that was discarded.  The batch
		// may not be in the main chain, so we stop the pipeline as a
		// precaution (it will be started again once the node is in
		// sync).
		log.Infow("Coordinator.handleReorg StopPipeline sync.LastBatch < c.pipelineBatchNum",
			"sync.LastBatch", stats.Sync.LastBatch,
			"c.pipelineBatchNum", c.pipelineBatchNum)
		if err := c.handleStopPipeline(ctx, "reorg"); err != nil {
			return tracerr.Wrap(err)
		}
		if err := c.l2DB.Reorg(common.BatchNum(stats.Sync.LastBatch)); err != nil {
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
				"attempt", attempt, "err", err, "block", t.lastBlock)
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
	if err := t.l2DB.DoneForging(l2TxsIDs(batchInfo.L2Txs), batchInfo.BatchNum); err != nil {
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

const longWaitTime = 999 * time.Hour

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	next := 0
	waitTime := time.Duration(longWaitTime)
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
			waitTime = t.cfg.TxManagerCheckInterval
		case <-time.After(waitTime):
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
					waitTime = longWaitTime
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
	txManager *TxManager,
	provers []prover.Client,
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
		txManager:    txManager,
		consts:       *scConsts,
		statsCh:      make(chan synchronizer.Stats, queueLen),
	}, nil
}

// SetSyncStats is a thread safe method to sets the synchronizer Stats
func (p *Pipeline) SetSyncStats(stats *synchronizer.Stats) {
	p.statsCh <- *stats
}

// Start the forging pipeline
func (p *Pipeline) Start(batchNum common.BatchNum, lastForgeL1TxsNum int64,
	syncStats *synchronizer.Stats, initSCVars *synchronizer.SCVariables) error {
	if p.started {
		log.Fatal("Pipeline already started")
	}
	p.started = true

	// Reset pipeline state
	p.batchNum = batchNum
	p.lastForgeL1TxsNum = lastForgeL1TxsNum
	p.vars = *initSCVars
	p.lastScheduledL1BatchBlockNum = 0

	p.ctx, p.cancel = context.WithCancel(context.Background())

	err := p.txSelector.Reset(p.batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = p.batchBuilder.Reset(p.batchNum, true)
	if err != nil {
		return tracerr.Wrap(err)
	}

	queueSize := 1
	batchChSentServerProof := make(chan *BatchInfo, queueSize)

	p.wg.Add(1)
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				log.Debug("Pipeline forgeSendServerProof loop done")
				p.wg.Done()
				return
			case syncStats := <-p.statsCh:
				p.stats = syncStats
			default:
				p.batchNum = p.batchNum + 1
				batchInfo, err := p.forgeSendServerProof(p.ctx, p.batchNum)
				if common.IsErrDone(err) {
					continue
				}
				if err != nil {
					log.Errorw("forgeSendServerProof", "err", err)
					continue
				}
				batchChSentServerProof <- batchInfo
			}
		}
	}()

	p.wg.Add(1)
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				log.Debug("Pipeline waitServerProofSendEth loop done")
				p.wg.Done()
				return
			case batchInfo := <-batchChSentServerProof:
				err := p.waitServerProof(p.ctx, batchInfo)
				if common.IsErrDone(err) {
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
	log.Debug("Stopping Pipeline...")
	p.cancel()
	p.wg.Wait()
	for _, prover := range p.provers {
		if err := prover.Cancel(ctx); err != nil {
			log.Errorw("prover.Cancel", "err", err)
		}
	}
}

func l2TxsIDs(txs []common.PoolL2Tx) []common.TxID {
	txIDs := make([]common.TxID, len(txs))
	for i, tx := range txs {
		txIDs[i] = tx.TxID
	}
	return txIDs
}

// forgeSendServerProof the next batch, wait for a proof server to be available and send the
// circuit inputs to the proof server.
func (p *Pipeline) forgeSendServerProof(ctx context.Context, batchNum common.BatchNum) (*BatchInfo, error) {
	// remove transactions from the pool that have been there for too long
	err := p.l2DB.Purge(common.BatchNum(p.stats.Sync.LastBatch))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	batchInfo := BatchInfo{BatchNum: batchNum} // to accumulate metadata of the batch

	var poolL2Txs []common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1CoordTxs []common.L1Tx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if p.shouldL1L2Batch() {
		p.lastScheduledL1BatchBlockNum = p.stats.Eth.LastBlock.Num
		// 2a: L1+L2 txs
		p.lastForgeL1TxsNum++
		l1UserTxs, err := p.historyDB.GetL1UserTxs(p.lastForgeL1TxsNum)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1UserTxsExtra, l1CoordTxs, poolL2Txs, err = p.txSelector.GetL1L2TxSelection([]common.Idx{}, batchNum, l1UserTxs) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	} else {
		// 2b: only L2 txs
		l1CoordTxs, poolL2Txs, err = p.txSelector.GetL2TxSelection([]common.Idx{}, batchNum)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1UserTxsExtra = nil
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	// TODO feesInfo
	batchInfo.L1UserTxsExtra = l1UserTxsExtra
	batchInfo.L1CoordTxs = l1CoordTxs
	batchInfo.L2Txs = poolL2Txs

	if err := p.l2DB.StartForging(l2TxsIDs(batchInfo.L2Txs), batchInfo.BatchNum); err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Run purger to invalidate transactions that become invalid beause of
	// the poolL2Txs selected.  Will mark as invalid the txs that have a
	// (fromIdx, nonce) which already appears in the selected txs (includes
	// all the nonces smaller than the current one)
	err = p.purgeInvalidDueToL2TxsSelection(poolL2Txs)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress: p.cfg.ForgerAddress,
	}
	zkInputs, err := p.batchBuilder.BuildBatch([]common.Idx{}, configBatch,
		l1UserTxsExtra, l1CoordTxs, poolL2Txs, nil) // TODO []common.TokenID --> feesInfo
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.ZKInputs = zkInputs
	p.cfg.debugBatchStore(&batchInfo)

	// 6. Wait for an available server proof blocking call
	serverProof, err := p.proversPool.Get(ctx)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	batchInfo.ServerProof = serverProof
	defer func() {
		// If there's an error further on, add the serverProof back to
		// the pool
		if err != nil {
			p.proversPool.Add(serverProof)
		}
	}()
	p.cfg.debugBatchStore(&batchInfo)

	// 7. Call the selected idle server proof with BatchBuilder output,
	// save server proof info for batchNum
	err = batchInfo.ServerProof.CalculateProof(zkInputs)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return &batchInfo, nil
}

// waitServerProof gets the generated zkProof & sends it to the SmartContract
func (p *Pipeline) waitServerProof(ctx context.Context, batchInfo *BatchInfo) error {
	proof, err := batchInfo.ServerProof.GetProof(ctx) // blocking call, until not resolved don't continue. Returns when the proof server has calculated the proof
	if err != nil {
		return tracerr.Wrap(err)
	}
	p.proversPool.Add(batchInfo.ServerProof)
	batchInfo.ServerProof = nil
	batchInfo.Proof = proof
	batchInfo.ForgeBatchArgs = p.prepareForgeBatchArgs(batchInfo)
	batchInfo.TxStatus = TxStatusPending
	p.cfg.debugBatchStore(batchInfo)
	return nil
}

func (p *Pipeline) purgeInvalidDueToL2TxsSelection(l2Txs []common.PoolL2Tx) error {
	return nil // TODO
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

func (p *Pipeline) prepareForgeBatchArgs(batchInfo *BatchInfo) *eth.RollupForgeBatchArgs {
	// TODO
	return &eth.RollupForgeBatchArgs{}
}
