package coordinator

import (
	"time"

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

// Config contains the Coordinator configuration
type Config struct {
	ForgerAddress ethCommon.Address
	LoopInterval  time.Duration
}

// Coordinator implements the Coordinator type
type Coordinator struct {
	// m            sync.Mutex
	stopch       chan bool
	stopforgerch chan bool

	forging    bool
	isForgeSeq bool // WIP just for testing while implementing

	config Config

	batchNum        common.BatchNum
	batchQueue      *BatchQueue
	serverProofPool ServerProofPool

	// synchronizer *synchronizer.Synchronizer
	hdb          *historydb.HistoryDB
	txsel        *txselector.TxSelector
	batchBuilder *batchbuilder.BatchBuilder

	ethClient  *eth.Client
	ethTxStore kvdb.Storage
}

// NewCoordinator creates a new Coordinator
func NewCoordinator(conf Config,
	hdb *historydb.HistoryDB,
	txsel *txselector.TxSelector,
	bb *batchbuilder.BatchBuilder,
	ethClient *eth.Client) *Coordinator { // once synchronizer is ready, synchronizer.Synchronizer will be passed as parameter here
	c := Coordinator{
		config:       conf,
		hdb:          hdb,
		txsel:        txsel,
		batchBuilder: bb,
		ethClient:    ethClient,
		ethTxStore:   memory.NewMemoryStorage(),
	}
	return &c
}

// Stop stops the Coordinator
func (c *Coordinator) Stop() {
	log.Info("Stopping Coordinator")
	c.stopch <- true
}

// Start starts the Coordinator service
func (c *Coordinator) Start() {
	c.stopch = make(chan bool) // initialize channel
	go func() {
		log.Info("Starting Coordinator")
		for {
			select {
			case <-c.stopch:
				close(c.stopforgerch)
				log.Info("Coordinator stopped")
				return
			case <-time.After(c.config.LoopInterval):
				if !c.isForgeSequence() {
					if c.forging {
						log.Info("forging stopped")
						c.forging = false
						close(c.stopforgerch)
					}
					log.Debug("not in forge time")
					continue
				}
				if !c.forging {
					log.Info("Start Forging")
					// c.batchNum = c.hdb.GetLastBatchNum() // uncomment when HistoryDB is ready
					err := c.txsel.Reset(c.batchNum)
					if err != nil {
						log.Error("forging err: ", err)
					}
					err = c.batchBuilder.Reset(c.batchNum, true)
					if err != nil {
						log.Error("forging err: ", err)
					}
					c.batchQueue = NewBatchQueue()

					c.forgerLoop()
					c.forging = true
				}
			}
		}
	}()
}

// forgerLoop trigers goroutines for:
// - forgeSequence
// - proveSequence
// - forgeConfirmationSequence
func (c *Coordinator) forgerLoop() {
	c.stopforgerch = make(chan bool) // initialize channel

	go func() {
		log.Info("forgeSequence started")
		for {
			select {
			case <-c.stopforgerch:
				log.Info("forgeSequence stopped")
				return
			case <-time.After(c.config.LoopInterval):
				if err := c.forgeSequence(); err != nil {
					log.Error("forgeSequence err: ", err)
				}
			}
		}
	}()
	go func() {
		log.Info("proveSequence started")
		for {
			select {
			case <-c.stopforgerch:
				log.Info("proveSequence stopped")
				return
			case <-time.After(c.config.LoopInterval):
				if err := c.proveSequence(); err != nil && err != common.ErrBatchQueueEmpty {
					log.Error("proveSequence err: ", err)
				}
			}
		}
	}()
	go func() {
		log.Info("forgeConfirmationSequence started")
		for {
			select {
			case <-c.stopforgerch:
				log.Info("forgeConfirmationSequence stopped")
				return
			case <-time.After(c.config.LoopInterval):
				if err := c.forgeConfirmationSequence(); err != nil {
					log.Error("forgeConfirmationSequence err: ", err)
				}
			}
		}
	}()
}

// forgeSequence
func (c *Coordinator) forgeSequence() error {
	// TODO once synchronizer has this method ready:
	// If there's been a reorg, handle it
	// handleReorg() function decides if the reorg must restart the pipeline or not
	// if c.synchronizer.Reorg():
	_ = c.handleReorg()

	// 0. If there's an available server proof: Start pipeline for batchNum = batchNum + 1
	serverProofInfo, err := c.serverProofPool.GetNextAvailable() // blocking call, returns when a server proof is available
	if err != nil {
		return err
	}

	// remove transactions from the pool that have been there for too long
	err = c.purgeRemoveByTimeout()
	if err != nil {
		return err
	}

	c.batchNum = c.batchNum + 1
	batchInfo := NewBatchInfo(c.batchNum, serverProofInfo) // to accumulate metadata of the batch

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
			return err
		}
	} else {
		// 2b: only L2 txs
		poolL2Txs, err = c.txsel.GetL2TxSelection(c.batchNum) // TODO once feesInfo is added to method return, add the var
		if err != nil {
			return err
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
		return err
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	batchInfo.SetTxsInfo(l1UserTxsExtra, l1OperatorTxs, poolL2Txs) // TODO feesInfo

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		ForgerAddress: c.config.ForgerAddress,
	}
	l2Txs := common.PoolL2TxsToL2Txs(poolL2Txs)
	zkInputs, err := c.batchBuilder.BuildBatch(configBatch, l1UserTxsExtra, l1OperatorTxs, l2Txs, nil) // TODO []common.TokenID --> feesInfo
	if err != nil {
		return err
	}

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.SetZKInputs(zkInputs)
	log.Debugf("Batch builded, batchNum: %d ", c.batchNum)

	// 6. Call an idle server proof with BatchBuilder output, save server proof info for batchNum
	err = batchInfo.serverProof.CalculateProof(zkInputs)
	if err != nil {
		return err
	}
	c.batchQueue.Push(&batchInfo)

	return nil
}

// proveSequence gets the generated zkProof & sends it to the SmartContract
func (c *Coordinator) proveSequence() error {
	batchInfo := c.batchQueue.Pop()
	if batchInfo == nil {
		// no batches in queue, return
		log.Debug("not batch to prove yet")
		return common.ErrBatchQueueEmpty
	}
	serverProofInfo := batchInfo.serverProof
	proof, err := serverProofInfo.GetProof() // blocking call, until not resolved don't continue. Returns when the proof server has calculated the proof
	if err != nil {
		return err
	}
	batchInfo.SetProof(proof)
	callData := c.prepareCallDataForge(batchInfo)
	_, err = c.ethClient.ForgeCall(callData)
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

func (c *Coordinator) forgeConfirmationSequence() error {
	// TODO strategy of this sequence TBD
	// confirm eth txs and mark them as accepted sequence
	// ethTx := ethTxStore.GetFirstPending()
	// waitForAccepted(ethTx) // blocking call, returns once the ethTx is mined
	// ethTxStore.MarkAccepted(ethTx)
	return nil
}

func (c *Coordinator) handleReorg() error {
	return nil
}

// isForgeSequence returns true if the node is the Forger in the current ethereum block
func (c *Coordinator) isForgeSequence() bool {
	return c.isForgeSeq
}

func (c *Coordinator) purgeRemoveByTimeout() error {
	return nil
}

func (c *Coordinator) purgeInvalidDueToL2TxsSelection(l2Txs []*common.PoolL2Tx) error {
	return nil
}

func (c *Coordinator) shouldL1L2Batch() bool {
	return false
}

func (c *Coordinator) prepareCallDataForge(batchInfo *BatchInfo) *common.CallDataForge {
	return nil
}
