package coordinator

import (
	"fmt"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txselector"
	kvdb "github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/memory"
)

// ErrStop is returned when the function is stopped asynchronously via the stop
// channel.  It doesn't indicate an error.
var ErrStop = fmt.Errorf("Stopped")

// Config contains the Coordinator configuration
type Config struct {
	ForgerAddress ethCommon.Address
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	forging    bool
	isForgeSeq bool // WIP just for testing while implementing

	config Config

	batchNum        common.BatchNum
	serverProofPool *ServerProofPool

	// synchronizer *synchronizer.Synchronizer
	hdb          *historydb.HistoryDB
	txsel        *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	ethClient  eth.ClientInterface
	ethTxStore kvdb.Storage
}

// NewCoordinator creates a new Coordinator
func NewCoordinator(conf Config,
	hdb *historydb.HistoryDB,
	txsel *txselector.TxSelector,
	bb *batchbuilder.BatchBuilder,
	serverProofs []ServerProofInterface,
	ethClient *eth.Client) *Coordinator { // once synchronizer is ready, synchronizer.Synchronizer will be passed as parameter here
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
		ethTxStore:      memory.NewMemoryStorage(),
	}
	return &c
}

// ForgeLoopFn is the function ran in a loop that checks if it's time to forge
// and forges a batch if so and sends it to outBatchCh.  Returns true if it's
// the coordinator turn to forge.
func (c *Coordinator) ForgeLoopFn(outBatchCh chan *BatchInfo, stopCh chan bool) (bool, error) {
	if !c.isForgeSequence() {
		if c.forging {
			log.Info("stop forging")
			c.forging = false
		}
		log.Debug("not in forge time")
		return false, nil
	}
	log.Debug("forge time")
	if !c.forging {
		log.Info("start forging")
		// c.batchNum = c.hdb.GetLastBatchNum() // uncomment when HistoryDB is ready
		err := c.txsel.Reset(c.batchNum)
		if err != nil {
			log.Errorw("TxSelector.Reset", "error", err)
			return true, err
		}
		err = c.batchBuilder.Reset(c.batchNum, true)
		if err != nil {
			log.Errorw("BatchBuilder.Reset", "error", err)
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

	// 0. If there's an available server proof: Start pipeline for batchNum = batchNum + 1.
	// non-blocking call, returns nil if a server proof is
	// not available, or non-nil otherwise.
	serverProof, err := c.serverProofPool.Get(stopCh)
	if err != nil {
		return true, err
	}
	log.Debugw("got serverProof", "server", serverProof)

	log.Debugw("start forge")
	batchInfo, err := c.forge(serverProof)
	if err != nil {
		log.Errorw("forge", "error", err)
		return true, err
	}
	log.Debugw("end forge", "batchNum", batchInfo.batchNum)
	outBatchCh <- batchInfo
	return true, nil
}

// GetProofCallForgeLoopFn is the function ran in a loop that gets a forged
// batch via inBatchCh, waits for the proof server to finish, calls the ForgeBatch
// function in the Rollup Smart Contract, and sends the batch to outBatchCh.
func (c *Coordinator) GetProofCallForgeLoopFn(inBatchCh, outBatchCh chan *BatchInfo, stopCh chan bool) error {
	select {
	case <-stopCh:
		log.Info("forgeLoopFn stopped")
		return ErrStop
	case batchInfo := <-inBatchCh:
		log.Debugw("start getProofCallForge", "batchNum", batchInfo.batchNum)
		if err := c.getProofCallForge(batchInfo, stopCh); err != nil {
			return err
		}
		log.Debugw("end getProofCallForge", "batchNum", batchInfo.batchNum)
		outBatchCh <- batchInfo
	}
	return nil
}

// ForgeCallConfirmLoopFn is the function ran in a loop that gets a batch that
// has been sent to the Rollup Smart Contract via inBatchCh and waits for the
// ethereum transaction confirmation.
func (c *Coordinator) ForgeCallConfirmLoopFn(inBatchCh chan *BatchInfo, stopCh chan bool) error {
	select {
	case <-stopCh:
		log.Info("forgeConfirmLoopFn stopped")
		return ErrStop
	case batchInfo := <-inBatchCh:
		log.Debugw("start forgeCallConfirm", "batchNum", batchInfo.batchNum)
		if err := c.forgeCallConfirm(batchInfo); err != nil {
			return err
		}
		log.Debugw("end forgeCallConfirm", "batchNum", batchInfo.batchNum)
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

	var poolL2Txs []*common.PoolL2Tx
	// var feesInfo
	var l1UserTxsExtra, l1OperatorTxs []*common.L1Tx
	// 1. Decide if we forge L2Tx or L1+L2Tx
	if c.shouldL1L2Batch() {
		// 2a: L1+L2 txs
		// l1UserTxs, toForgeL1TxsNumber := c.hdb.GetNextL1UserTxs() // TODO once HistoryDB is ready, uncomment
		var l1UserTxs []*common.L1Tx = nil                                                                // tmp, depends on HistoryDB
		l1UserTxsExtra, l1OperatorTxs, poolL2Txs, err = c.txsel.GetL1L2TxSelection(c.batchNum, l1UserTxs) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return nil, err
		}
	} else {
		// 2b: only L2 txs
		poolL2Txs, err = c.txsel.GetL2TxSelection(c.batchNum) // TODO once feesInfo is added to method return, add the var
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
	zkInputs, err := c.batchBuilder.BuildBatch(configBatch, l1UserTxsExtra, l1OperatorTxs, poolL2Txs, nil) // TODO []common.TokenID --> feesInfo
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
	if err != nil {
		return err
	}
	batchInfo.SetProof(proof)
	forgeBatchArgs := c.prepareForgeBatchArgs(batchInfo)
	_, err = c.ethClient.RollupForgeBatch(forgeBatchArgs)
	if err != nil {
		return err
	}
	log.Debugf("ethClient ForgeCall sent, batchNum: %d", c.batchNum)

	// TODO once tx data type is defined, store ethTx (returned by ForgeCall)
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

func (c *Coordinator) forgeCallConfirm(batchInfo *BatchInfo) error {
	// TODO strategy of this sequence TBD
	// confirm eth txs and mark them as accepted sequence
	// ethTx := ethTxStore.GetFirstPending()
	// waitForAccepted(ethTx) // blocking call, returns once the ethTx is mined
	// ethTxStore.MarkAccepted(ethTx)
	return nil
}

func (c *Coordinator) handleReorg() error {
	return nil // TODO
}

// isForgeSequence returns true if the node is the Forger in the current ethereum block
func (c *Coordinator) isForgeSequence() bool {
	return c.isForgeSeq // TODO
}

func (c *Coordinator) purgeRemoveByTimeout() error {
	return nil // TODO
}

func (c *Coordinator) purgeInvalidDueToL2TxsSelection(l2Txs []*common.PoolL2Tx) error {
	return nil // TODO
}

func (c *Coordinator) shouldL1L2Batch() bool {
	return false // TODO
}

func (c *Coordinator) prepareForgeBatchArgs(batchInfo *BatchInfo) *eth.RollupForgeBatchArgs {
	return nil // TODO
}
