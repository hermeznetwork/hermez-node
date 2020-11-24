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
)

var errTODO = fmt.Errorf("TODO")

// ErrDone is returned when the function is stopped asynchronously via a done
// (terminated) context. It doesn't indicate an error.
var ErrDone = fmt.Errorf("done")

// Config contains the Coordinator configuration
type Config struct {
	ForgerAddress ethCommon.Address
	ConfirmBlocks int64
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	// State
	forging         bool
	batchNum        common.BatchNum
	serverProofPool *ServerProofPool
	consts          synchronizer.SCConsts
	vars            synchronizer.SCVariables

	cfg Config

	hdb          *historydb.HistoryDB
	txsel        *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	ethClient eth.ClientInterface

	msgCh  chan interface{}
	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc

	pipelineCtx    context.Context
	pipelineWg     sync.WaitGroup
	pipelineCancel context.CancelFunc

	txManager *TxManager
}

// NewCoordinator creates a new Coordinator
func NewCoordinator(cfg Config,
	hdb *historydb.HistoryDB,
	txsel *txselector.TxSelector,
	bb *batchbuilder.BatchBuilder,
	serverProofs []ServerProofInterface,
	ethClient eth.ClientInterface,
	scConsts *synchronizer.SCConsts,
	initSCVars *synchronizer.SCVariables,
) *Coordinator { // once synchronizer is ready, synchronizer.Synchronizer will be passed as parameter here
	serverProofPool := NewServerProofPool(len(serverProofs))
	for _, serverProof := range serverProofs {
		serverProofPool.Add(serverProof)
	}

	txManager := NewTxManager(ethClient, cfg.ConfirmBlocks)

	ctx, cancel := context.WithCancel(context.Background())
	c := Coordinator{
		forging:         false,
		batchNum:        -1,
		serverProofPool: serverProofPool,
		consts:          *scConsts,
		vars:            *initSCVars,

		cfg: cfg,

		hdb:          hdb,
		txsel:        txsel,
		batchBuilder: bb,

		ethClient: ethClient,

		msgCh: make(chan interface{}),
		ctx:   ctx,
		// wg
		cancel: cancel,

		txManager: txManager,
	}
	return &c
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

func (c *Coordinator) handleMsgSyncStats(stats *synchronizer.Stats) error {
	if !stats.Synced() {
		return nil
	}
	c.txManager.SetLastBlock(stats.Eth.LastBlock)

	anyoneForge := false
	if stats.Sync.Auction.CurrentSlot.BatchesLen == 0 &&
		c.consts.Auction.RelativeBlock(stats.Eth.LastBlock) > int64(c.vars.Auction.SlotDeadline) {
		log.Debug("Coordinator: anyone can forge in the current slot (slotDeadline passed)")
		anyoneForge = true
	}
	if stats.Sync.Auction.CurrentSlot.Forger != c.cfg.ForgerAddress && !anyoneForge {
		if c.forging {
			log.Info("Coordinator: forging state end")
			c.forging = false
			c.PipelineStop()
		}
		// log.Debug("Coordinator: not in forge time") // DBG
		return nil
	}
	// log.Debug("Coordinator: forge time") // DBG
	if !c.forging {
		// Start pipeline from a batchNum state taken from synchronizer
		log.Info("Coordinator: forging state begin")
		c.batchNum = common.BatchNum(stats.Sync.LastBatch)
		err := c.txsel.Reset(c.batchNum)
		if err != nil {
			log.Errorw("Coordinator: TxSelector.Reset", "error", err)
			return err
		}
		err = c.batchBuilder.Reset(c.batchNum, true)
		if err != nil {
			log.Errorw("Coordinator: BatchBuilder.Reset", "error", err)
			return err
		}
		c.forging = true
		c.PipelineStart()
	}
	return nil
}

// Start the coordinator
func (c *Coordinator) Start() {
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
					}
				case MsgSyncReorg:
					if err := c.handleReorg(); err != nil {
						log.Errorw("Coordinator.handleReorg error", "err", err)
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
	log.Infow("Stopping coordinator...")
	c.cancel()
	c.wg.Wait()
	if c.forging {
		c.forging = false
		c.PipelineStop()
	}
}

// PipelineStart starts the forging pipeline
func (c *Coordinator) PipelineStart() {
	c.pipelineCtx, c.pipelineCancel = context.WithCancel(context.Background())

	queueSize := 1
	batchChSentServerProof := make(chan *BatchInfo, queueSize)

	c.pipelineWg.Add(1)
	go func() {
		for {
			select {
			case <-c.pipelineCtx.Done():
				log.Debug("Pipeline forgeSendServerProof loop done")
				c.pipelineWg.Done()
				return
			default:
				c.batchNum = c.batchNum + 1
				batchInfo, err := c.forgeSendServerProof(c.pipelineCtx, c.batchNum)
				if err == ErrDone {
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

	c.pipelineWg.Add(1)
	go func() {
		for {
			select {
			case <-c.pipelineCtx.Done():
				log.Debug("Pipeline waitServerProofSendEth loop done")
				c.pipelineWg.Done()
				return
			case batchInfo := <-batchChSentServerProof:
				err := c.waitServerProof(c.pipelineCtx, batchInfo)
				if err == ErrDone {
					continue
				}
				if err != nil {
					log.Errorw("waitServerProof", "err", err)
					continue
				}
				c.txManager.AddBatch(batchInfo)
			}
		}
	}()
}

// PipelineStop stops the forging pipeline
func (c *Coordinator) PipelineStop() {
	log.Debug("Stopping pipeline...")
	c.pipelineCancel()
	c.pipelineWg.Wait()
}

// TxManager handles everything related to ethereum transactions:  It makes the
// call to forge, waits for transaction confirmation, and keeps checking them
// until a number of confirmed blocks have passed.
type TxManager struct {
	ethClient    eth.ClientInterface
	batchCh      chan *BatchInfo
	lastBlockCh  chan int64
	queue        []*BatchInfo
	confirmation int64
	lastBlock    int64
}

// NewTxManager creates a new TxManager
func NewTxManager(ethClient eth.ClientInterface, confirmation int64) *TxManager {
	return &TxManager{
		ethClient: ethClient,
		// TODO: Find best queue size
		batchCh: make(chan *BatchInfo, 16), //nolint:gomnd
		// TODO: Find best queue size
		lastBlockCh:  make(chan int64, 16), //nolint:gomnd
		confirmation: confirmation,
		lastBlock:    -1,
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

const waitTime = 200 * time.Millisecond
const longWaitTime = 999 * time.Hour

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	next := 0
	d := time.Duration(longWaitTime)
	for {
		select {
		case <-ctx.Done():
			log.Info("TxManager done")
			return
		case lastBlock := <-t.lastBlockCh:
			t.lastBlock = lastBlock
		case batchInfo := <-t.batchCh:
			ethTx, err := t.ethClient.RollupForgeBatch(batchInfo.ForgeBatchArgs)
			if err != nil {
				// TODO: Figure out different error cases and handle them properly
				log.Errorw("TxManager ethClient.RollupForgeBatch", "err", err)
				continue
			}
			log.Debugf("ethClient ForgeCall sent, batchNum: %d", batchInfo.BatchNum)
			batchInfo.EthTx = ethTx
			t.queue = append(t.queue, batchInfo)
			d = waitTime
		case <-time.After(d):
			if len(t.queue) == 0 {
				continue
			}
			batchInfo := t.queue[next]
			txID := batchInfo.EthTx.Hash()
			receipt, err := t.ethClient.EthTransactionReceipt(ctx, txID)
			if err != nil {
				log.Errorw("TxManager ethClient.EthTransactionReceipt", "err", err)
				// TODO: Figure out different error cases and handle them properly
				// TODO: Notify the Coordinator to maybe reset the pipeline
				continue
			}

			if receipt != nil {
				if receipt.Status == types.ReceiptStatusFailed {
					log.Errorw("TxManager receipt status is failed", "receipt", receipt)
				} else if receipt.Status == types.ReceiptStatusSuccessful {
					if t.lastBlock-receipt.BlockNumber.Int64() >= t.confirmation {
						log.Debugw("TxManager tx for RollupForgeBatch confirmed", "batchNum", batchInfo.BatchNum)
						t.queue = t.queue[1:]
						if len(t.queue) == 0 {
							d = longWaitTime
						}
					}
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

// forgeSendServerProof the next batch, wait for a proof server to be available and send the
// circuit inputs to the proof server.
func (c *Coordinator) forgeSendServerProof(ctx context.Context, batchNum common.BatchNum) (*BatchInfo, error) {
	// remove transactions from the pool that have been there for too long
	err := c.purgeRemoveByTimeout()
	if err != nil {
		return nil, err
	}

	batchInfo := BatchInfo{BatchNum: batchNum} // to accumulate metadata of the batch

	var poolL2Txs []common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1OperatorTxs []common.L1Tx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if c.shouldL1L2Batch() {
		// 2a: L1+L2 txs
		// l1UserTxs, toForgeL1TxsNumber := c.hdb.GetNextL1UserTxs() // TODO once HistoryDB is ready, uncomment
		var l1UserTxs []common.L1Tx = nil                                                                               // tmp, depends on HistoryDB
		l1UserTxsExtra, l1OperatorTxs, poolL2Txs, err = c.txsel.GetL1L2TxSelection([]common.Idx{}, batchNum, l1UserTxs) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, err
		}
	} else {
		// 2b: only L2 txs
		_, poolL2Txs, err = c.txsel.GetL2TxSelection([]common.Idx{}, batchNum) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, err
		}
		l1UserTxsExtra = nil
		l1OperatorTxs = nil
	}

	// Run purger to invalidate transactions that become invalid beause of
	// the poolL2Txs selected.  Will mark as invalid the txs that have a
	// (fromIdx, nonce) which already appears in the selected txs (includes
	// all the nonces smaller than the current one)
	err = c.purgeInvalidDueToL2TxsSelection(poolL2Txs)
	if err != nil {
		return nil, err
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	// batchInfo.SetTxsInfo(l1UserTxsExtra, l1OperatorTxs, poolL2Txs) // TODO feesInfo
	batchInfo.L1UserTxsExtra = l1UserTxsExtra
	batchInfo.L1OperatorTxs = l1OperatorTxs
	batchInfo.L2Txs = poolL2Txs

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress: c.cfg.ForgerAddress,
	}
	zkInputs, err := c.batchBuilder.BuildBatch([]common.Idx{}, configBatch, l1UserTxsExtra, l1OperatorTxs, poolL2Txs, nil) // TODO []common.TokenID --> feesInfo
	if err != nil {
		return nil, err
	}

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.ZKInputs = zkInputs

	// 6. Wait for an available server proof blocking call
	serverProof, err := c.serverProofPool.Get(ctx)
	if err != nil {
		return nil, err
	}
	batchInfo.ServerProof = serverProof
	defer func() {
		// If there's an error further on, add the serverProof back to
		// the pool
		if err != nil {
			c.serverProofPool.Add(serverProof)
		}
	}()

	// 7. Call the selected idle server proof with BatchBuilder output,
	// save server proof info for batchNum
	err = batchInfo.ServerProof.CalculateProof(zkInputs)
	if err != nil {
		return nil, err
	}

	return &batchInfo, nil
}

// waitServerProof gets the generated zkProof & sends it to the SmartContract
func (c *Coordinator) waitServerProof(ctx context.Context, batchInfo *BatchInfo) error {
	proof, err := batchInfo.ServerProof.GetProof(ctx) // blocking call, until not resolved don't continue. Returns when the proof server has calculated the proof
	if err != nil {
		return err
	}
	c.serverProofPool.Add(batchInfo.ServerProof)
	batchInfo.ServerProof = nil
	batchInfo.Proof = proof
	batchInfo.ForgeBatchArgs = c.prepareForgeBatchArgs(batchInfo)
	batchInfo.TxStatus = TxStatusPending

	// TODO(FUTURE) once tx data type is defined, store ethTx (returned by ForgeCall)
	// TBD if use ethTxStore as a disk k-v database, or use a Queue
	// tx, err := c.ethTxStore.NewTx()
	// if err != nil {
	//         return err
	// }
	// tx.Put(ethTx.Hash(), ethTx.Bytes())
	// if err := tx.Commit(); err!=nil {
	//         return nil
	// }

	return nil
}

func (c *Coordinator) handleReorg() error {
	return nil // TODO
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

func (c *Coordinator) purgeRemoveByTimeout() error {
	return nil // TODO
}

func (c *Coordinator) purgeInvalidDueToL2TxsSelection(l2Txs []common.PoolL2Tx) error {
	return nil // TODO
}

func (c *Coordinator) shouldL1L2Batch() bool {
	return false // TODO
}

func (c *Coordinator) prepareForgeBatchArgs(batchInfo *BatchInfo) *eth.RollupForgeBatchArgs {
	// TODO
	return &eth.RollupForgeBatchArgs{}
}
