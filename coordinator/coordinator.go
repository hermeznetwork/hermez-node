package coordinator

import (
	"context"
	"fmt"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
)

var errTODO = fmt.Errorf("TODO")

// ErrDone is returned when the function is stopped asynchronously via a done
// (terminated) context. It doesn't indicate an error.
var ErrDone = fmt.Errorf("done")

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
	batchNum     common.BatchNum
	serverProofs []ServerProofInterface
	consts       synchronizer.SCConsts
	vars         synchronizer.SCVariables
	started      bool

	cfg Config

	historyDB    *historydb.HistoryDB
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
	txSelector *txselector.TxSelector,
	batchBuilder *batchbuilder.BatchBuilder,
	serverProofs []ServerProofInterface,
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

	txManager := NewTxManager(&cfg, ethClient)

	ctx, cancel := context.WithCancel(context.Background())
	c := Coordinator{
		batchNum:     -1,
		serverProofs: serverProofs,
		consts:       *scConsts,
		vars:         *initSCVars,

		cfg: cfg,

		historyDB:    historyDB,
		txSelector:   txSelector,
		batchBuilder: batchBuilder,

		// ethClient: ethClient,

		msgCh: make(chan interface{}),
		ctx:   ctx,
		// wg
		cancel: cancel,

		txManager: txManager,
	}
	return &c, nil
}

func (c *Coordinator) newPipeline() *Pipeline {
	return NewPipeline(c.cfg, c.historyDB, c.txSelector, c.batchBuilder,
		c.txManager, c.serverProofs, &c.consts)
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

func (c *Coordinator) handleMsgSyncStats(stats *synchronizer.Stats) error {
	if !stats.Synced() {
		return nil
	}
	c.txManager.SetLastBlock(stats.Eth.LastBlock.Num)

	canForge := c.canForge(stats)
	if c.pipeline == nil {
		if canForge {
			log.Info("Coordinator: forging state begin")
			batchNum := common.BatchNum(stats.Sync.LastBatch)
			c.pipeline = c.newPipeline()
			if err := c.pipeline.Start(batchNum, stats, &c.vars); err != nil {
				return tracerr.Wrap(err)
			}
		}
	} else {
		if canForge {
			c.pipeline.SetSyncStats(stats)
		} else {
			log.Info("Coordinator: forging state end")
			c.pipeline.Stop()
			c.pipeline = nil
		}
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
					if err := c.handleMsgSyncStats(&stats); err != nil {
						log.Errorw("Coordinator.handleMsgSyncStats error", "err", err)
						continue
					}
				case MsgSyncReorg:
					if err := c.handleReorg(); err != nil {
						log.Errorw("Coordinator.handleReorg error", "err", err)
						continue
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
		c.pipeline.Stop()
		c.pipeline = nil
	}
}

func (c *Coordinator) handleReorg() error {
	return nil // TODO
}

// TxManager handles everything related to ethereum transactions:  It makes the
// call to forge, waits for transaction confirmation, and keeps checking them
// until a number of confirmed blocks have passed.
type TxManager struct {
	cfg         Config
	ethClient   eth.ClientInterface
	batchCh     chan *BatchInfo
	lastBlockCh chan int64
	queue       []*BatchInfo
	lastBlock   int64
}

// NewTxManager creates a new TxManager
func NewTxManager(cfg *Config, ethClient eth.ClientInterface) *TxManager {
	return &TxManager{
		cfg:       *cfg,
		ethClient: ethClient,
		// TODO: Find best queue size
		batchCh: make(chan *BatchInfo, 16), //nolint:gomnd
		// TODO: Find best queue size
		lastBlockCh: make(chan int64, 16), //nolint:gomnd
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
			log.Errorw("TxManager ethClient.RollupForgeBatch",
				"attempt", attempt, "err", err)
		} else {
			break
		}
		select {
		case <-ctx.Done():
			return tracerr.Wrap(ErrDone)
		case <-time.After(t.cfg.EthClientAttemptsDelay):
		}
	}
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("reached max attempts for ethClient.RollupForgeBatch: %w", err))
	}
	batchInfo.EthTx = ethTx
	t.cfg.debugBatchStore(batchInfo)
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
			return tracerr.Wrap(ErrDone)
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
			if err := t.rollupForgeBatch(ctx, batchInfo); tracerr.Unwrap(err) == ErrDone {
				continue
			} else if err != nil {
				// TODO: Reset pipeline
				continue
			}
			log.Debugf("ethClient ForgeCall sent, batchNum: %d", batchInfo.BatchNum)
			t.queue = append(t.queue, batchInfo)
			waitTime = t.cfg.TxManagerCheckInterval
		case <-time.After(waitTime):
			if len(t.queue) == 0 {
				continue
			}
			batchInfo := t.queue[next]
			err := t.ethTransactionReceipt(ctx, batchInfo)
			if tracerr.Unwrap(err) == ErrDone {
				continue
			} else if err != nil { //nolint:staticcheck
				// We can't get the receipt for the
				// transaction, so we can't confirm if it was
				// mined
				// TODO: Reset pipeline
			}

			confirm, err := t.handleReceipt(batchInfo)
			if err != nil { //nolint:staticcheck
				// Transaction was rejected
				// TODO: Reset pipeline
			}
			if confirm != nil && *confirm >= t.cfg.ConfirmBlocks {
				log.Debugw("TxManager tx for RollupForgeBatch confirmed",
					"batchNum", batchInfo.BatchNum)
				t.queue = t.queue[1:]
				if len(t.queue) == 0 {
					waitTime = longWaitTime
				}
			}
			if len(t.queue) == 0 {
				next = 0
			} else {
				next = (next + 1) % len(t.queue)
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
	started                      bool

	serverProofPool *ServerProofPool
	txManager       *TxManager
	historyDB       *historydb.HistoryDB
	txSelector      *txselector.TxSelector
	batchBuilder    *batchbuilder.BatchBuilder

	stats   synchronizer.Stats
	statsCh chan synchronizer.Stats

	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

// NewPipeline creates a new Pipeline
func NewPipeline(cfg Config,
	historyDB *historydb.HistoryDB,
	txSelector *txselector.TxSelector,
	batchBuilder *batchbuilder.BatchBuilder,
	txManager *TxManager,
	serverProofs []ServerProofInterface,
	scConsts *synchronizer.SCConsts,
) *Pipeline {
	serverProofPool := NewServerProofPool(len(serverProofs))
	for _, serverProof := range serverProofs {
		serverProofPool.Add(serverProof)
	}
	return &Pipeline{
		cfg:             cfg,
		historyDB:       historyDB,
		txSelector:      txSelector,
		batchBuilder:    batchBuilder,
		serverProofPool: serverProofPool,
		txManager:       txManager,
		consts:          *scConsts,
		// TODO: Find best queue size
		statsCh: make(chan synchronizer.Stats, 16), //nolint:gomnd
	}
}

// SetSyncStats is a thread safe method to sets the synchronizer Stats
func (p *Pipeline) SetSyncStats(stats *synchronizer.Stats) {
	p.statsCh <- *stats
}

// Start the forging pipeline
func (p *Pipeline) Start(batchNum common.BatchNum,
	syncStats *synchronizer.Stats, initSCVars *synchronizer.SCVariables) error {
	if p.started {
		log.Fatal("Pipeline already started")
	}
	p.started = true

	// Reset pipeline state
	p.batchNum = batchNum
	p.vars = *initSCVars
	p.lastScheduledL1BatchBlockNum = 0

	p.ctx, p.cancel = context.WithCancel(context.Background())

	err := p.txSelector.Reset(p.batchNum)
	if err != nil {
		log.Errorw("Pipeline: TxSelector.Reset", "error", err)
		return tracerr.Wrap(err)
	}
	err = p.batchBuilder.Reset(p.batchNum, true)
	if err != nil {
		log.Errorw("Pipeline: BatchBuilder.Reset", "error", err)
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
				if tracerr.Unwrap(err) == ErrDone {
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
				if tracerr.Unwrap(err) == ErrDone {
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
func (p *Pipeline) Stop() {
	if !p.started {
		log.Fatal("Pipeline already stopped")
	}
	p.started = false
	log.Debug("Stopping Pipeline...")
	p.cancel()
	p.wg.Wait()
	// TODO: Cancel all proofServers with pending proofs
}

// forgeSendServerProof the next batch, wait for a proof server to be available and send the
// circuit inputs to the proof server.
func (p *Pipeline) forgeSendServerProof(ctx context.Context, batchNum common.BatchNum) (*BatchInfo, error) {
	// remove transactions from the pool that have been there for too long
	err := p.purgeRemoveByTimeout()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	batchInfo := BatchInfo{BatchNum: batchNum} // to accumulate metadata of the batch

	var poolL2Txs []common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1OperatorTxs []common.L1Tx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if p.shouldL1L2Batch() {
		p.lastScheduledL1BatchBlockNum = p.stats.Eth.LastBatch
		// 2a: L1+L2 txs
		// l1UserTxs, toForgeL1TxsNumber := c.historyDB.GetNextL1UserTxs() // TODO once HistoryDB is ready, uncomment
		var l1UserTxs []common.L1Tx = nil                                                                                    // tmp, depends on HistoryDB
		l1UserTxsExtra, l1OperatorTxs, poolL2Txs, err = p.txSelector.GetL1L2TxSelection([]common.Idx{}, batchNum, l1UserTxs) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	} else {
		// 2b: only L2 txs
		_, poolL2Txs, err = p.txSelector.GetL2TxSelection([]common.Idx{}, batchNum) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1UserTxsExtra = nil
		l1OperatorTxs = nil
	}

	// Run purger to invalidate transactions that become invalid beause of
	// the poolL2Txs selected.  Will mark as invalid the txs that have a
	// (fromIdx, nonce) which already appears in the selected txs (includes
	// all the nonces smaller than the current one)
	err = p.purgeInvalidDueToL2TxsSelection(poolL2Txs)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	// batchInfo.SetTxsInfo(l1UserTxsExtra, l1OperatorTxs, poolL2Txs) // TODO feesInfo
	batchInfo.L1UserTxsExtra = l1UserTxsExtra
	batchInfo.L1OperatorTxs = l1OperatorTxs
	batchInfo.L2Txs = poolL2Txs

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress: p.cfg.ForgerAddress,
	}
	zkInputs, err := p.batchBuilder.BuildBatch([]common.Idx{}, configBatch, l1UserTxsExtra, l1OperatorTxs, poolL2Txs, nil) // TODO []common.TokenID --> feesInfo
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.ZKInputs = zkInputs
	p.cfg.debugBatchStore(&batchInfo)

	// 6. Wait for an available server proof blocking call
	serverProof, err := p.serverProofPool.Get(ctx)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	batchInfo.ServerProof = serverProof
	defer func() {
		// If there's an error further on, add the serverProof back to
		// the pool
		if err != nil {
			p.serverProofPool.Add(serverProof)
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
	p.serverProofPool.Add(batchInfo.ServerProof)
	batchInfo.ServerProof = nil
	batchInfo.Proof = proof
	batchInfo.ForgeBatchArgs = p.prepareForgeBatchArgs(batchInfo)
	batchInfo.TxStatus = TxStatusPending
	p.cfg.debugBatchStore(batchInfo)
	return nil
}

// isForgeSequence returns true if the node is the Forger in the current ethereum block
// func (c *Coordinator) isForgeSequence() (bool, error) {
// 	// TODO: Consider checking if we can forge by quering the Synchronizer instead of using ethClient
// 	blockNum, err := c.ethClient.EthLastBlock()
// 	if err != nil {
// 		return false, err
// 	}
// 	addr, err := c.ethClient.EthAddress()
// 	if err != nil {
// 		return false, err
// 	}
// 	return c.ethClient.AuctionCanForge(*addr, blockNum+1)
// }

func (p *Pipeline) purgeRemoveByTimeout() error {
	return nil // TODO
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
