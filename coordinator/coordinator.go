package coordinator

import (
	"context"
	"fmt"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txselector"
)

var errTODO = fmt.Errorf("TODO")

// ErrStop is returned when the function is stopped asynchronously via the stop
// channel.  It doesn't indicate an error.
var ErrStop = fmt.Errorf("Stopped")

// Config contains the Coordinator configuration
type Config struct {
	ForgerAddress ethCommon.Address
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	forging bool
	// rw         *sync.RWMutex
	// isForgeSeq bool // WIP just for testing while implementing

	config Config

	batchNum        common.BatchNum
	serverProofPool *ServerProofPool

	// synchronizer *synchronizer.Synchronizer
	hdb          *historydb.HistoryDB
	txsel        *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	ethClient eth.ClientInterface
	ethTxs    []*types.Transaction
	// ethTxStore kvdb.Storage
}

// NewCoordinator creates a new Coordinator
func NewCoordinator(conf Config,
	hdb *historydb.HistoryDB,
	txsel *txselector.TxSelector,
	bb *batchbuilder.BatchBuilder,
	serverProofs []ServerProofInterface,
	ethClient eth.ClientInterface) *Coordinator { // once synchronizer is ready, synchronizer.Synchronizer will be passed as parameter here
	serverProofPool := NewServerProofPool(len(serverProofs))
	for _, serverProof := range serverProofs {
		serverProofPool.Add(serverProof)
	}
	c := Coordinator{
		config:          conf,
		serverProofPool: serverProofPool,
		hdb:             hdb,
		txsel:           txsel,
		batchBuilder:    bb,
		ethClient:       ethClient,

		ethTxs: make([]*types.Transaction, 0),
		// ethTxStore:      memory.NewMemoryStorage(),
		// rw:              &sync.RWMutex{},
	}
	return &c
}

// TODO(Edu): Change the current design of the coordinator structur:
// - Move Start and Stop functions (from node/node.go) here
// - Add concept of StartPipeline, StopPipeline, that spawns and stops the goroutines
// - Add a Manager that calls StartPipeline and StopPipeline, checks when it's time to forge, schedules new batches, etc.
// - Add a TxMonitor that monitors successful ForgeBatch ethereum transactions and waits for N blocks of confirmation, and reports back errors to the Manager.

// ForgeLoopFn is the function ran in a loop that checks if it's time to forge
// and forges a batch if so and sends it to outBatchCh.  Returns true if it's
// the coordinator turn to forge.
func (c *Coordinator) ForgeLoopFn(outBatchCh chan *BatchInfo, stopCh chan bool) (forgetime bool, err error) {
	// TODO: Move the logic to check if it's forge time or not outside the pipeline
	isForgeSequence, err := c.isForgeSequence()
	if err != nil {
		return false, err
	}
	if !isForgeSequence {
		if c.forging {
			log.Info("ForgeLoopFn: forging state end")
			c.forging = false
		}
		log.Debug("ForgeLoopFn: not in forge time")
		return false, nil
	}
	log.Debug("ForgeLoopFn: forge time")
	if !c.forging {
		// Start pipeline from a batchNum state taken from synchronizer
		log.Info("ForgeLoopFn: forging state begin")
		// c.batchNum = c.hdb.GetLastBatchNum() // uncomment when HistoryDB is ready
		err := c.txsel.Reset(c.batchNum)
		if err != nil {
			log.Errorw("ForgeLoopFn: TxSelector.Reset", "error", err)
			return true, err
		}
		err = c.batchBuilder.Reset(c.batchNum, true)
		if err != nil {
			log.Errorw("ForgeLoopFn: BatchBuilder.Reset", "error", err)
			return true, err
		}
		// c.batchQueue = NewBatchQueue()
		c.forging = true
	}
	// TODO once synchronizer has this method ready:
	// If there's been a reorg, handle it
	// handleReorg() function decides if the reorg must restart the pipeline or not
	// if c.synchronizer.Reorg():
	_ = c.handleReorg()

	defer func() {
		if err == ErrStop {
			log.Info("ForgeLoopFn: forgeLoopFn stopped")
		}
	}()

	// 0. Wait for an available server proof
	// blocking call
	serverProof, err := c.serverProofPool.Get(stopCh)
	if err != nil {
		return true, err
	}
	defer func() {
		if !forgetime || err != nil {
			c.serverProofPool.Add(serverProof)
		}
	}()

	log.Debugw("ForgeLoopFn: using serverProof", "server", serverProof)
	log.Debugw("ForgeLoopFn: forge start")
	// forge for batchNum = batchNum + 1.
	batchInfo, err := c.forge(serverProof)
	if err != nil {
		log.Errorw("forge", "error", err)
		return true, err
	}
	log.Debugw("ForgeLoopFn: forge end", "batchNum", batchInfo.batchNum)
	outBatchCh <- batchInfo
	return true, nil
}

// GetProofCallForgeLoopFn is the function ran in a loop that gets a forged
// batch via inBatchCh, waits for the proof server to finish, calls the ForgeBatch
// function in the Rollup Smart Contract, and sends the batch to outBatchCh.
func (c *Coordinator) GetProofCallForgeLoopFn(inBatchCh, outBatchCh chan *BatchInfo, stopCh chan bool) (err error) {
	defer func() {
		if err == ErrStop {
			log.Info("GetProofCallForgeLoopFn: forgeLoopFn stopped")
		}
	}()
	select {
	case <-stopCh:
		return ErrStop
	case batchInfo := <-inBatchCh:
		log.Debugw("GetProofCallForgeLoopFn: getProofCallForge start", "batchNum", batchInfo.batchNum)
		if err := c.getProofCallForge(batchInfo, stopCh); err != nil {
			return err
		}
		log.Debugw("GetProofCallForgeLoopFn: getProofCallForge end", "batchNum", batchInfo.batchNum)
		outBatchCh <- batchInfo
	}
	return nil
}

// ForgeCallConfirmLoopFn is the function ran in a loop that gets a batch that
// has been sent to the Rollup Smart Contract via inBatchCh and waits for the
// ethereum transaction confirmation.
func (c *Coordinator) ForgeCallConfirmLoopFn(inBatchCh chan *BatchInfo, stopCh chan bool) (err error) {
	defer func() {
		if err == ErrStop {
			log.Info("ForgeCallConfirmLoopFn: forgeConfirmLoopFn stopped")
		}
	}()
	select {
	case <-stopCh:
		return ErrStop
	case batchInfo := <-inBatchCh:
		log.Debugw("ForgeCallConfirmLoopFn: forgeCallConfirm start", "batchNum", batchInfo.batchNum)
		if err := c.forgeCallConfirm(batchInfo, stopCh); err != nil {
			return err
		}
		log.Debugw("ForgeCallConfirmLoopFn: forgeCallConfirm  end", "batchNum", batchInfo.batchNum)
	}
	return nil
}

func (c *Coordinator) forge(serverProof ServerProofInterface) (*BatchInfo, error) {
	// remove transactions from the pool that have been there for too long
	err := c.purgeRemoveByTimeout()
	if err != nil {
		return nil, err
	}

	c.batchNum = c.batchNum + 1
	batchInfo := NewBatchInfo(c.batchNum, serverProof) // to accumulate metadata of the batch

	var poolL2Txs []common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1OperatorTxs []common.L1Tx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if c.shouldL1L2Batch() {
		// 2a: L1+L2 txs
		// l1UserTxs, toForgeL1TxsNumber := c.hdb.GetNextL1UserTxs() // TODO once HistoryDB is ready, uncomment
		var l1UserTxs []common.L1Tx = nil                                                                                 // tmp, depends on HistoryDB
		l1UserTxsExtra, l1OperatorTxs, poolL2Txs, err = c.txsel.GetL1L2TxSelection([]common.Idx{}, c.batchNum, l1UserTxs) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, err
		}
	} else {
		// 2b: only L2 txs
		poolL2Txs, err = c.txsel.GetL2TxSelection([]common.Idx{}, c.batchNum) // TODO once feesInfo is added to method return, add the var
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
	batchInfo.SetTxsInfo(l1UserTxsExtra, l1OperatorTxs, poolL2Txs) // TODO feesInfo

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress: c.config.ForgerAddress,
	}
	zkInputs, err := c.batchBuilder.BuildBatch([]common.Idx{}, configBatch, l1UserTxsExtra, l1OperatorTxs, poolL2Txs, nil) // TODO []common.TokenID --> feesInfo
	if err != nil {
		return nil, err
	}

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.SetZKInputs(zkInputs)

	// 6. Call an idle server proof with BatchBuilder output, save server proof info for batchNum
	err = batchInfo.serverProof.CalculateProof(zkInputs)
	if err != nil {
		return nil, err
	}

	return &batchInfo, nil
}

// getProofCallForge gets the generated zkProof & sends it to the SmartContract
func (c *Coordinator) getProofCallForge(batchInfo *BatchInfo, stopCh chan bool) error {
	serverProof := batchInfo.serverProof
	proof, err := serverProof.GetProof(stopCh) // blocking call, until not resolved don't continue. Returns when the proof server has calculated the proof
	c.serverProofPool.Add(serverProof)
	batchInfo.serverProof = nil
	if err != nil {
		return err
	}
	batchInfo.SetProof(proof)
	forgeBatchArgs := c.prepareForgeBatchArgs(batchInfo)
	ethTx, err := c.ethClient.RollupForgeBatch(forgeBatchArgs)
	if err != nil {
		return err
	}
	// TODO: Move this to the next step (forgeCallConfirm)
	log.Debugf("ethClient ForgeCall sent, batchNum: %d", c.batchNum)
	batchInfo.SetEthTx(ethTx)

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

func (c *Coordinator) forgeCallConfirm(batchInfo *BatchInfo, stopCh chan bool) error {
	// TODO strategy of this sequence TBD
	// confirm eth txs and mark them as accepted sequence
	// IDEA: Keep an array in Coordinator with the list of sent ethTx.
	// Here, loop over them and only delete them once the number of
	// confirmed blocks is over a configured value.  If the tx is rejected,
	// return error.
	// ethTx := ethTxStore.GetFirstPending()
	// waitForAccepted(ethTx) // blocking call, returns once the ethTx is mined
	// ethTxStore.MarkAccepted(ethTx)
	txID := batchInfo.ethTx.Hash()
	// TODO: Follow EthereumClient.waitReceipt logic
	count := 0
	// TODO: Define this waitTime in the config
	waitTime := 100 * time.Millisecond //nolint:gomnd
	select {
	case <-time.After(waitTime):
		receipt, err := c.ethClient.EthTransactionReceipt(context.TODO(), txID)
		if err != nil {
			return err
		}
		if receipt != nil {
			if receipt.Status == types.ReceiptStatusFailed {
				return fmt.Errorf("receipt status is failed")
			} else if receipt.Status == types.ReceiptStatusSuccessful {
				return nil
			}
		}
		// TODO: Call go-ethereum:
		// if err == nil && receipt == nil :
		// `func (ec *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {`
		count++
		if time.Duration(count)*waitTime > 60*time.Second {
			log.Warnw("Waiting for ethTx receipt for more than 60 seconds", "tx", batchInfo.ethTx)
			// TODO: Decide if we resend the Tx with higher gas price
		}
	case <-stopCh:
		return ErrStop
	}
	return fmt.Errorf("timeout")
}

func (c *Coordinator) handleReorg() error {
	return nil // TODO
}

// isForgeSequence returns true if the node is the Forger in the current ethereum block
func (c *Coordinator) isForgeSequence() (bool, error) {
	// TODO: Consider checking if we can forge by quering the Synchronizer instead of using ethClient
	blockNum, err := c.ethClient.EthCurrentBlock()
	if err != nil {
		return false, err
	}
	addr, err := c.ethClient.EthAddress()
	if err != nil {
		return false, err
	}
	return c.ethClient.AuctionCanForge(*addr, blockNum+1)
}

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
