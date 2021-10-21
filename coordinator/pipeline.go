package coordinator

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
)

type statsVars struct {
	Stats synchronizer.Stats
	Vars  common.SCVariablesPtr
}

type state struct {
	batchNum                     common.BatchNum
	lastScheduledL1BatchBlockNum int64
	lastForgeL1TxsNum            int64
	lastSlotForged               int64
}

// Pipeline manages the forging of batches with parallel server proofs
type Pipeline struct {
	num    int
	cfg    Config
	consts common.SCConsts

	// state
	state         state
	started       bool
	rw            sync.RWMutex
	errAtBatchNum common.BatchNum
	lastForgeTime time.Time

	proversPool           *ProversPool
	provers               []prover.Client
	coord                 *Coordinator
	txManager             *TxManager
	historyDB             *historydb.HistoryDB
	l2DB                  *l2db.L2DB
	txSelector            *txselector.TxSelector
	batchBuilder          *batchbuilder.BatchBuilder
	mutexL2DBUpdateDelete *sync.Mutex
	purger                *Purger

	stats       synchronizer.Stats
	vars        common.SCVariables
	statsVarsCh chan statsVars

	ctx    context.Context
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func (p *Pipeline) setErrAtBatchNum(batchNum common.BatchNum) {
	p.rw.Lock()
	defer p.rw.Unlock()
	p.errAtBatchNum = batchNum
}

func (p *Pipeline) getErrAtBatchNum() common.BatchNum {
	p.rw.RLock()
	defer p.rw.RUnlock()
	return p.errAtBatchNum
}

// NewPipeline creates a new Pipeline
func NewPipeline(ctx context.Context,
	cfg Config,
	num int, // Pipeline sequential number
	historyDB *historydb.HistoryDB,
	l2DB *l2db.L2DB,
	txSelector *txselector.TxSelector,
	batchBuilder *batchbuilder.BatchBuilder,
	mutexL2DBUpdateDelete *sync.Mutex,
	purger *Purger,
	coord *Coordinator,
	txManager *TxManager,
	provers []prover.Client,
	scConsts *common.SCConsts,
) (*Pipeline, error) {
	proversPool := NewProversPool(len(provers))
	proversPoolSize := 0
	for _, prover := range provers {
		ctxTimeout, ctxTimeoutCancel := context.WithTimeout(ctx, cfg.ProverReadTimeout)
		defer ctxTimeoutCancel()
		if err := prover.WaitReady(ctxTimeout); err != nil {
			log.Errorw("prover.WaitReady", "err", err)
		} else {
			proversPool.Add(ctx, prover)
			proversPoolSize++
		}
	}
	if proversPoolSize == 0 {
		return nil, tracerr.Wrap(fmt.Errorf("no provers in the pool"))
	}
	return &Pipeline{
		num:                   num,
		cfg:                   cfg,
		historyDB:             historyDB,
		l2DB:                  l2DB,
		txSelector:            txSelector,
		batchBuilder:          batchBuilder,
		provers:               provers,
		proversPool:           proversPool,
		mutexL2DBUpdateDelete: mutexL2DBUpdateDelete,
		purger:                purger,
		coord:                 coord,
		txManager:             txManager,
		consts:                *scConsts,
		statsVarsCh:           make(chan statsVars, queueLen),
	}, nil
}

// SetSyncStatsVars is a thread safe method to sets the synchronizer Stats
func (p *Pipeline) SetSyncStatsVars(ctx context.Context, stats *synchronizer.Stats,
	vars *common.SCVariablesPtr) {
	select {
	case p.statsVarsCh <- statsVars{Stats: *stats, Vars: *vars}:
	case <-ctx.Done():
	}
}

// reset pipeline state
func (p *Pipeline) reset(batchNum common.BatchNum,
	stats *synchronizer.Stats, vars *common.SCVariables) error {
	p.state = state{
		batchNum:                     batchNum,
		lastForgeL1TxsNum:            stats.Sync.LastForgeL1TxsNum,
		lastScheduledL1BatchBlockNum: 0,
		lastSlotForged:               -1,
	}
	p.stats = *stats
	p.vars = *vars

	// Reset the StateDB in TxSelector and BatchBuilder from the
	// synchronizer only if the checkpoint we reset from either:
	// a. Doesn't exist in the TxSelector/BatchBuilder
	// b. The batch has already been synced by the synchronizer and has a
	//    different MTRoot than the BatchBuilder
	// Otherwise, reset from the local checkpoint.

	// First attempt to reset from local checkpoint if such checkpoint exists
	existsTxSelector, err := p.txSelector.LocalAccountsDB().CheckpointExists(p.state.batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	fromSynchronizerTxSelector := !existsTxSelector
	if err := p.txSelector.Reset(p.state.batchNum, fromSynchronizerTxSelector); err != nil {
		return tracerr.Wrap(err)
	}
	existsBatchBuilder, err := p.batchBuilder.LocalStateDB().CheckpointExists(p.state.batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	fromSynchronizerBatchBuilder := !existsBatchBuilder
	if err := p.batchBuilder.Reset(p.state.batchNum, fromSynchronizerBatchBuilder); err != nil {
		return tracerr.Wrap(err)
	}

	// After reset, check that if the batch exists in the historyDB, the
	// stateRoot matches with the local one, if not, force a reset from
	// synchronizer
	batch, err := p.historyDB.GetBatch(p.state.batchNum)
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		// nothing to do
	} else if err != nil {
		return tracerr.Wrap(err)
	} else {
		localStateRoot := p.batchBuilder.LocalStateDB().MT.Root().BigInt()
		if batch.StateRoot.Cmp(localStateRoot) != 0 {
			log.Debugw("localStateRoot (%v) != historyDB stateRoot (%v).  "+
				"Forcing reset from Synchronizer", localStateRoot, batch.StateRoot)
			// StateRoot from synchronizer doesn't match StateRoot
			// from batchBuilder, force a reset from synchronizer
			if err := p.txSelector.Reset(p.state.batchNum, true); err != nil {
				return tracerr.Wrap(err)
			}
			if err := p.batchBuilder.Reset(p.state.batchNum, true); err != nil {
				return tracerr.Wrap(err)
			}
		}
	}
	return nil
}

func (p *Pipeline) syncSCVars(vars common.SCVariablesPtr) {
	updateSCVars(&p.vars, vars)
}

// handleForgeBatch waits for an available proof server, calls p.forgeBatch to
// forge the batch and get the zkInputs, and then  sends the zkInputs to the
// selected proof server so that the proof computation begins.
func (p *Pipeline) handleForgeBatch(ctx context.Context,
	batchNum common.BatchNum) (batchInfo *BatchInfo, err error) {
	// 1. Wait for an available serverProof (blocking call)
	serverProof, err := p.proversPool.Get(ctx)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if err != nil {
		log.Errorw("proversPool.Get", "err", err)
		return nil, tracerr.Wrap(err)
	}
	defer func() {
		// If we encounter any error (notice that this function returns
		// errors to notify that a batch is not forged not only because
		// of unexpected errors but also due to benign causes), add the
		// serverProof back to the pool
		if err != nil {
			p.proversPool.Add(ctx, serverProof)
		}
	}()

	// 2. Forge the batch internally (make a selection of txs and prepare
	// all the smart contract arguments)
	var skipReason *string
	p.mutexL2DBUpdateDelete.Lock()
	batchInfo, skipReason, err = p.forgeBatch(batchNum)
	p.mutexL2DBUpdateDelete.Unlock()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if err != nil {
		log.Errorw("forgeBatch", "err", err)
		return nil, tracerr.Wrap(err)
	} else if skipReason != nil {
		log.Debugw("skipping batch", "batch", batchNum, "reason", *skipReason)
		return nil, tracerr.Wrap(errSkipBatchByPolicy)
	}

	// 3. Send the ZKInputs to the proof server
	batchInfo.ServerProof = serverProof
	batchInfo.ProofStart = time.Now()
	if err := p.sendServerProof(ctx, batchInfo); ctx.Err() != nil {
		return nil, ctx.Err()
	} else if err != nil {
		log.Errorw("sendServerProof", "err", err)
		return nil, tracerr.Wrap(err)
	}
	return batchInfo, nil
}

// Start the forging pipeline
func (p *Pipeline) Start(batchNum common.BatchNum,
	stats *synchronizer.Stats, vars *common.SCVariables) error {
	if p.started {
		log.Fatal("Pipeline already started")
	}
	p.started = true

	if err := p.reset(batchNum, stats, vars); err != nil {
		return tracerr.Wrap(err)
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())

	queueSize := 1
	batchChSentServerProof := make(chan *BatchInfo, queueSize)

	p.wg.Add(1)
	go func() {
		timer := time.NewTimer(zeroDuration)
		for {
			select {
			case <-p.ctx.Done():
				log.Info("Pipeline forgeBatch loop done")
				p.wg.Done()
				return
			case statsVars := <-p.statsVarsCh:
				p.stats = statsVars.Stats
				p.syncSCVars(statsVars.Vars)
			case <-timer.C:
				timer.Reset(p.cfg.ForgeRetryInterval)
				// Once errAtBatchNum != 0, we stop forging
				// batches because there's been an error and we
				// wait for the pipeline to be stopped.
				if p.getErrAtBatchNum() != 0 {
					continue
				}
				batchNum = p.state.batchNum + 1
				batchInfo, err := p.handleForgeBatch(p.ctx, batchNum)
				if p.ctx.Err() != nil {
					p.revertPoolChanges(batchNum)
					continue
				} else if tracerr.Unwrap(err) == errSkipBatchByPolicy {
					p.revertPoolChanges(batchNum)
					continue
				} else if err != nil {
					p.setErrAtBatchNum(batchNum)
					p.coord.SendMsg(p.ctx, MsgStopPipeline{
						Reason: fmt.Sprintf(
							"Pipeline.handleForgBatch: %v", err),
						FailedBatchNum: batchNum,
					})
					p.revertPoolChanges(batchNum)
					continue
				}
				p.lastForgeTime = time.Now()

				p.state.batchNum = batchNum
				select {
				case batchChSentServerProof <- batchInfo:
				case <-p.ctx.Done():
				}
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(zeroDuration)
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
				go func(p *Pipeline, batchInfo *BatchInfo, batchNum common.BatchNum) {
					// Once errAtBatchNum != 0, we stop forging
					// batches because there's been an error and we
					// wait for the pipeline to be stopped.
					if p.getErrAtBatchNum() != 0 {
						p.revertPoolChanges(batchNum)
						return
					}
					err := p.waitServerProof(p.ctx, batchInfo)
					if p.ctx.Err() != nil {
						p.revertPoolChanges(batchNum)
						return
					} else if err != nil {
						log.Errorw("waitServerProof", "err", err)
						p.setErrAtBatchNum(batchInfo.BatchNum)
						p.coord.SendMsg(p.ctx, MsgStopPipeline{
							Reason: fmt.Sprintf(
								"Pipeline.waitServerProof: %v", err),
							FailedBatchNum: batchInfo.BatchNum,
						})
						p.revertPoolChanges(batchNum)
						return
					}
					// We are done with this serverProof, add it back to the pool
					p.proversPool.Add(p.ctx, batchInfo.ServerProof)
					p.txManager.AddBatch(p.ctx, batchInfo)
				}(p, batchInfo, batchNum)
			}
		}
	}()
	return nil
}

// revertPoolChanges will undo changes made to the pool while trying to forge failedBatch.
// Call this function only if the porcess of forging a batch fails
func (p *Pipeline) revertPoolChanges(failedBatch common.BatchNum) {
	if err := p.l2DB.Reorg(failedBatch - 1); err != nil {
		// NOTE: the reason why this error si not returned is that this function is used in a error handling situation
		// and at this point the flow shouldn't change (handling the error of handling an error), things could get really meesy
		log.Error("Error trying to revert changes on the pool after the porcess of forging a batch failed: ", err)
	}
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
		if err := prover.Cancel(ctx); ctx.Err() != nil {
			continue
		} else if err != nil {
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

// slotCommitted returns true if the current slot has already been committed
func (p *Pipeline) slotCommitted() bool {
	// Synchronizer has synchronized a batch in the current slot (setting
	// CurrentSlot.ForgerCommitment) or the pipeline has already
	// internally-forged a batch in the current slot
	return p.stats.Sync.Auction.CurrentSlot.ForgerCommitment ||
		p.stats.Sync.Auction.CurrentSlot.SlotNum == p.state.lastSlotForged
}

// forgePolicySkipPreSelection is called before doing a tx selection in a batch to
// determine by policy if we should forge the batch or not.  Returns true and
// the reason when the forging of the batch must be skipped.
func (p *Pipeline) forgePolicySkipPreSelection(now time.Time) (bool, string) {
	// Check if the slot is not yet fulfilled
	slotCommitted := p.slotCommitted()
	if p.cfg.ForgeOncePerSlotIfTxs {
		if slotCommitted {
			return true, "cfg.ForgeOncePerSlotIfTxs = true and slot already committed"
		}
		return false, ""
	}
	// Determine if we must commit the slot
	if !p.cfg.IgnoreSlotCommitment && !slotCommitted {
		return false, ""
	}

	// If we haven't reached the ForgeDelay, skip forging the batch
	if now.Sub(p.lastForgeTime) < p.cfg.ForgeDelay {
		return true, "we haven't reached the forge delay"
	}
	return false, ""
}

// forgePolicySkipPostSelection is called after doing a tx selection in a batch to
// determine by policy if we should forge the batch or not.  Returns true and
// the reason when the forging of the batch must be skipped.
func (p *Pipeline) forgePolicySkipPostSelection(now time.Time, l1UserTxsExtra, l1CoordTxs []common.L1Tx,
	poolL2Txs []common.PoolL2Tx, batchInfo *BatchInfo) (bool, string, error) {
	// Check if the slot is not yet fulfilled
	slotCommitted := p.slotCommitted()

	pendingTxs := true
	if len(l1UserTxsExtra) == 0 && len(l1CoordTxs) == 0 && len(poolL2Txs) == 0 {
		if batchInfo.L1Batch {
			// Query the number of unforged L1UserTxs
			// (either in a open queue or in a frozen
			// not-yet-forged queue).
			count, err := p.historyDB.GetUnforgedL1UserTxsCount()
			if err != nil {
				return false, "", err
			}
			// If there are future L1UserTxs, we forge a
			// batch to advance the queues to be able to
			// forge the L1UserTxs in the future.
			// Otherwise, skip.
			if count == 0 {
				pendingTxs = false
			}
		} else {
			pendingTxs = false
		}
	}

	if p.cfg.ForgeOncePerSlotIfTxs {
		if slotCommitted {
			return true, "cfg.ForgeOncePerSlotIfTxs = true and slot already committed",
				nil
		}
		if pendingTxs {
			return false, "", nil
		}
		return true, "cfg.ForgeOncePerSlotIfTxs = true and no pending txs",
			nil
	}

	// Determine if we must commit the slot
	if !p.cfg.IgnoreSlotCommitment && !slotCommitted {
		return false, "", nil
	}

	// check if there is no txs to forge, no l1UserTxs in the open queue to
	// freeze and we haven't reached the ForgeNoTxsDelay
	if now.Sub(p.lastForgeTime) < p.cfg.ForgeNoTxsDelay {
		if !pendingTxs {
			return true, "no txs to forge and we haven't reached the forge no txs delay",
				nil
		}
	}
	return false, "", nil
}

// forgeBatch forges the batchNum batch.
func (p *Pipeline) forgeBatch(batchNum common.BatchNum) (batchInfo *BatchInfo,
	skipReason *string, err error) {
	// remove transactions from the pool that have been there for too long
	_, err = p.purger.InvalidateMaybe(p.l2DB, p.txSelector.LocalAccountsDB(),
		p.stats.Sync.LastBlock.Num, int64(batchNum))
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	_, err = p.purger.PurgeMaybe(p.l2DB, p.stats.Sync.LastBlock.Num, int64(batchNum))
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	// Structure to accumulate data and metadata of the batch
	now := time.Now()
	batchInfo = &BatchInfo{PipelineNum: p.num, BatchNum: batchNum}
	batchInfo.Debug.StartTimestamp = now
	batchInfo.Debug.StartBlockNum = p.stats.Eth.LastBlock.Num + 1

	var poolL2Txs []common.PoolL2Tx
	var discardedL2Txs []common.PoolL2Tx
	var l1UserTxs, l1CoordTxs []common.L1Tx
	var auths [][]byte
	var coordIdxs []common.Idx

	if skip, reason := p.forgePolicySkipPreSelection(now); skip {
		return nil, &reason, nil
	}

	// 1. Decide if we forge L2Tx or L1+L2Tx
	if p.shouldL1L2Batch(batchInfo) {
		batchInfo.L1Batch = true
		// 2a: L1+L2 txs
		_l1UserTxs, err := p.historyDB.GetUnforgedL1UserTxs(p.state.lastForgeL1TxsNum + 1)
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		// l1UserFutureTxs are the l1UserTxs that are not being forged
		// in the next batch, but that are also in the queue for the
		// future batches
		l1UserFutureTxs, err := p.historyDB.GetUnforgedL1UserFutureTxs(p.state.lastForgeL1TxsNum + 1)
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}

		coordIdxs, auths, l1UserTxs, l1CoordTxs, poolL2Txs, discardedL2Txs, err =
			p.txSelector.GetL1L2TxSelection(p.cfg.TxProcessorConfig, _l1UserTxs, l1UserFutureTxs)
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
	} else {
		// get l1UserFutureTxs which are all the l1 pending in all the
		// queues
		l1UserFutureTxs, err := p.historyDB.GetUnforgedL1UserFutureTxs(p.state.lastForgeL1TxsNum) //nolint:gomnd
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}

		// 2b: only L2 txs
		coordIdxs, auths, l1CoordTxs, poolL2Txs, discardedL2Txs, err =
			p.txSelector.GetL2TxSelection(p.cfg.TxProcessorConfig, l1UserFutureTxs)
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		l1UserTxs = nil
	}

	if skip, reason, err := p.forgePolicySkipPostSelection(now,
		l1UserTxs, l1CoordTxs, poolL2Txs, batchInfo); err != nil {
		return nil, nil, tracerr.Wrap(err)
	} else if skip {
		if err := p.txSelector.Reset(batchInfo.BatchNum-1, false); err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		return nil, &reason, tracerr.Wrap(err)
	}

	if batchInfo.L1Batch {
		p.state.lastScheduledL1BatchBlockNum = p.stats.Eth.LastBlock.Num + 1
		p.state.lastForgeL1TxsNum++
	}

	// 3.  Save metadata from TxSelector output for BatchNum
	batchInfo.L1UserTxs = l1UserTxs
	batchInfo.L1CoordTxs = l1CoordTxs
	batchInfo.L1CoordinatorTxsAuths = auths
	batchInfo.CoordIdxs = coordIdxs
	batchInfo.VerifierIdx = p.cfg.VerifierIdx

	if err := p.l2DB.StartForging(common.TxIDsFromPoolL2Txs(poolL2Txs),
		batchInfo.BatchNum); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	if err := p.l2DB.UpdateTxsInfo(discardedL2Txs, batchInfo.BatchNum); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	// Invalidate transactions that become invalid because of
	// the poolL2Txs selected.  Will mark as invalid the txs that have a
	// (fromIdx, nonce) which already appears in the selected txs (includes
	// all the nonces smaller than the current one)
	err = p.l2DB.InvalidateOldNonces(idxsNonceFromPoolL2Txs(poolL2Txs), batchInfo.BatchNum)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	// 4. Call BatchBuilder with TxSelector output
	configBatch := &batchbuilder.ConfigBatch{
		TxProcessorConfig: p.cfg.TxProcessorConfig,
	}
	zkInputs, err := p.batchBuilder.BuildBatch(coordIdxs, configBatch, l1UserTxs,
		l1CoordTxs, poolL2Txs)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	l2Txs, err := common.PoolL2TxsToL2Txs(poolL2Txs) // NOTE: This is a big uggly, find a better way
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	batchInfo.L2Txs = l2Txs

	// 5. Save metadata from BatchBuilder output for BatchNum
	batchInfo.ZKInputs = zkInputs
	batchInfo.Debug.Status = StatusForged
	p.cfg.debugBatchStore(batchInfo)
	log.Infow("Pipeline: batch forged internally", "batch", batchInfo.BatchNum)

	p.state.lastSlotForged = p.stats.Sync.Auction.CurrentSlot.SlotNum

	return batchInfo, nil, nil
}

// waitServerProof gets the generated zkProof & sends it to the SmartContract
func (p *Pipeline) waitServerProof(ctx context.Context, batchInfo *BatchInfo) error {
	defer metric.MeasureDuration(metric.WaitServerProof, batchInfo.ProofStart,
		batchInfo.BatchNum.BigInt().String(), strconv.Itoa(batchInfo.PipelineNum))

	proof, pubInputs, err := batchInfo.ServerProof.GetProof(ctx) // blocking call,
	// until not resolved don't continue. Returns when the proof server has calculated the proof
	if err != nil {
		return tracerr.Wrap(err)
	}
	batchInfo.Proof = proof
	batchInfo.PublicInputs = pubInputs
	batchInfo.ForgeBatchArgs = prepareForgeBatchArgs(batchInfo)
	batchInfo.Debug.Status = StatusProof
	p.cfg.debugBatchStore(batchInfo)
	log.Infow("Pipeline: batch proof calculated", "batch", batchInfo.BatchNum)
	return nil
}

func (p *Pipeline) shouldL1L2Batch(batchInfo *BatchInfo) bool {
	// Take the lastL1BatchBlockNum as the biggest between the last
	// scheduled one, and the synchronized one.
	lastL1BatchBlockNum := p.state.lastScheduledL1BatchBlockNum
	if p.stats.Sync.LastL1BatchBlock > lastL1BatchBlockNum {
		lastL1BatchBlockNum = p.stats.Sync.LastL1BatchBlock
	}
	// Set Debug information
	batchInfo.Debug.LastScheduledL1BatchBlockNum = p.state.lastScheduledL1BatchBlockNum
	batchInfo.Debug.LastL1BatchBlock = p.stats.Sync.LastL1BatchBlock
	batchInfo.Debug.LastL1BatchBlockDelta = p.stats.Eth.LastBlock.Num + 1 - lastL1BatchBlockNum
	batchInfo.Debug.L1BatchBlockScheduleDeadline =
		int64(float64(p.vars.Rollup.ForgeL1L2BatchTimeout-1) * p.cfg.L1BatchTimeoutPerc)
	// Return true if we have passed the l1BatchTimeoutPerc portion of the
	// range before the l1batch timeout.
	return p.stats.Eth.LastBlock.Num+1-lastL1BatchBlockNum >=
		int64(float64(p.vars.Rollup.ForgeL1L2BatchTimeout-1)*p.cfg.L1BatchTimeoutPerc)
}

func prepareForgeBatchArgs(batchInfo *BatchInfo) *eth.RollupForgeBatchArgs {
	proof := batchInfo.Proof
	zki := batchInfo.ZKInputs
	return &eth.RollupForgeBatchArgs{
		NewLastIdx:            int64(zki.Metadata.NewLastIdxRaw),
		NewStRoot:             zki.Metadata.NewStateRootRaw.BigInt(),
		NewExitRoot:           zki.Metadata.NewExitRootRaw.BigInt(),
		L1UserTxs:             batchInfo.L1UserTxs,
		L1CoordinatorTxs:      batchInfo.L1CoordTxs,
		L1CoordinatorTxsAuths: batchInfo.L1CoordinatorTxsAuths,
		L2TxsData:             batchInfo.L2Txs,
		FeeIdxCoordinator:     batchInfo.CoordIdxs,
		// Circuit selector
		VerifierIdx: batchInfo.VerifierIdx,
		L1Batch:     batchInfo.L1Batch,
		ProofA:      [2]*big.Int{proof.PiA[0], proof.PiA[1]},
		// Implementation of the verifier need a swap on the proofB vector
		ProofB: [2][2]*big.Int{
			{proof.PiB[0][1], proof.PiB[0][0]},
			{proof.PiB[1][1], proof.PiB[1][0]},
		},
		ProofC: [2]*big.Int{proof.PiC[0], proof.PiC[1]},
	}
}
