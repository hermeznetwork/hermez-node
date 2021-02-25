package coordinator

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/tracerr"
)

// TxManager handles everything related to ethereum transactions:  It makes the
// call to forge, waits for transaction confirmation, and keeps checking them
// until a number of confirmed blocks have passed.
type TxManager struct {
	cfg       Config
	ethClient eth.ClientInterface
	l2DB      *l2db.L2DB   // Used only to mark forged txs as forged in the L2DB
	coord     *Coordinator // Used only to send messages to stop the pipeline
	batchCh   chan *BatchInfo
	chainID   *big.Int
	account   accounts.Account
	consts    synchronizer.SCConsts

	stats       synchronizer.Stats
	vars        synchronizer.SCVariables
	statsVarsCh chan statsVars

	discardPipelineCh chan int // int refers to the pipelineNum

	minPipelineNum int
	queue          Queue
	// lastSuccessBatch stores the last BatchNum that who's forge call was confirmed
	lastSuccessBatch common.BatchNum
	// lastPendingBatch common.BatchNum
	// accNonce is the account nonce in the last mined block (due to mined txs)
	accNonce uint64
	// accNextNonce is the nonce that we should use to send the next tx.
	// In some cases this will be a reused nonce of an already pending tx.
	accNextNonce uint64

	lastSentL1BatchBlockNum int64
}

// NewTxManager creates a new TxManager
func NewTxManager(ctx context.Context, cfg *Config, ethClient eth.ClientInterface, l2DB *l2db.L2DB,
	coord *Coordinator, scConsts *synchronizer.SCConsts, initSCVars *synchronizer.SCVariables) (*TxManager, error) {
	chainID, err := ethClient.EthChainID()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	address, err := ethClient.EthAddress()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	accNonce, err := ethClient.EthNonceAt(ctx, *address, nil)
	if err != nil {
		return nil, err
	}
	log.Infow("TxManager started", "nonce", accNonce)
	return &TxManager{
		cfg:               *cfg,
		ethClient:         ethClient,
		l2DB:              l2DB,
		coord:             coord,
		batchCh:           make(chan *BatchInfo, queueLen),
		statsVarsCh:       make(chan statsVars, queueLen),
		discardPipelineCh: make(chan int, queueLen),
		account: accounts.Account{
			Address: *address,
		},
		chainID: chainID,
		consts:  *scConsts,

		vars: *initSCVars,

		minPipelineNum: 0,
		queue:          NewQueue(),
		accNonce:       accNonce,
		accNextNonce:   accNonce,
	}, nil
}

// AddBatch is a thread safe method to pass a new batch TxManager to be sent to
// the smart contract via the forge call
func (t *TxManager) AddBatch(ctx context.Context, batchInfo *BatchInfo) {
	select {
	case t.batchCh <- batchInfo:
	case <-ctx.Done():
	}
}

// SetSyncStatsVars is a thread safe method to sets the synchronizer Stats
func (t *TxManager) SetSyncStatsVars(ctx context.Context, stats *synchronizer.Stats, vars *synchronizer.SCVariablesPtr) {
	select {
	case t.statsVarsCh <- statsVars{Stats: *stats, Vars: *vars}:
	case <-ctx.Done():
	}
}

// DiscardPipeline is a thread safe method to notify about a discarded pipeline
// due to a reorg
func (t *TxManager) DiscardPipeline(ctx context.Context, pipelineNum int) {
	select {
	case t.discardPipelineCh <- pipelineNum:
	case <-ctx.Done():
	}
}

func (t *TxManager) syncSCVars(vars synchronizer.SCVariablesPtr) {
	updateSCVars(&t.vars, vars)
}

// NewAuth generates a new auth object for an ethereum transaction
func (t *TxManager) NewAuth(ctx context.Context, batchInfo *BatchInfo) (*bind.TransactOpts, error) {
	gasPrice, err := t.ethClient.EthSuggestGasPrice(ctx)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if t.cfg.GasPriceIncPerc != 0 {
		inc := new(big.Int).Set(gasPrice)
		inc.Mul(inc, new(big.Int).SetInt64(t.cfg.GasPriceIncPerc))
		// nolint reason: to calculate percentages we use 100
		inc.Div(inc, new(big.Int).SetUint64(100)) //nolint:gomnd
		gasPrice.Add(gasPrice, inc)
	}

	// log.Debugw("TxManager: transaction metadata", "gasPrice", gasPrice)

	auth, err := bind.NewKeyStoreTransactorWithChainID(t.ethClient.EthKeyStore(), t.account, t.chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auth.Value = big.NewInt(0) // in wei

	gasLimit := t.cfg.ForgeBatchGasCost.Fixed +
		uint64(len(batchInfo.L1UserTxsExtra))*t.cfg.ForgeBatchGasCost.L1UserTx +
		uint64(len(batchInfo.L1CoordTxs))*t.cfg.ForgeBatchGasCost.L1CoordTx +
		uint64(len(batchInfo.L2Txs))*t.cfg.ForgeBatchGasCost.L2Tx
	auth.GasLimit = gasLimit
	auth.GasPrice = gasPrice
	auth.Nonce = nil

	return auth, nil
}

func (t *TxManager) shouldSendRollupForgeBatch(batchInfo *BatchInfo) error {
	nextBlock := t.stats.Eth.LastBlock.Num + 1
	if !t.canForgeAt(nextBlock) {
		return tracerr.Wrap(fmt.Errorf("can't forge in the next block: %v", nextBlock))
	}
	if t.mustL1L2Batch(nextBlock) && !batchInfo.L1Batch {
		return tracerr.Wrap(fmt.Errorf("can't forge non-L1Batch in the next block: %v", nextBlock))
	}
	margin := t.cfg.SendBatchBlocksMarginCheck
	if margin != 0 {
		if !t.canForgeAt(nextBlock + margin) {
			return tracerr.Wrap(fmt.Errorf("can't forge after %v blocks: %v",
				margin, nextBlock))
		}
		if t.mustL1L2Batch(nextBlock+margin) && !batchInfo.L1Batch {
			return tracerr.Wrap(fmt.Errorf("can't forge non-L1Batch after %v blocks: %v",
				margin, nextBlock))
		}
	}
	return nil
}

func addPerc(v *big.Int, p int64) *big.Int {
	r := new(big.Int).Set(v)
	r.Mul(r, big.NewInt(p))
	// nolint reason: to calculate percentages we divide by 100
	r.Div(r, big.NewInt(100)) //nolit:gomnd
	return r.Add(v, r)
}

func (t *TxManager) sendRollupForgeBatch(ctx context.Context, batchInfo *BatchInfo, resend bool) error {
	var ethTx *types.Transaction
	var err error
	auth, err := t.NewAuth(ctx, batchInfo)
	if err != nil {
		return tracerr.Wrap(err)
	}
	auth.Nonce = big.NewInt(int64(t.accNextNonce))
	if resend {
		auth.Nonce = big.NewInt(int64(batchInfo.EthTx.Nonce()))
	}
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		if auth.GasPrice.Cmp(t.cfg.MaxGasPrice) > 0 {
			return tracerr.Wrap(fmt.Errorf("calculated gasPrice (%v) > maxGasPrice (%v)",
				auth.GasPrice, t.cfg.MaxGasPrice))
		}
		// RollupForgeBatch() calls ethclient.SendTransaction()
		ethTx, err = t.ethClient.RollupForgeBatch(batchInfo.ForgeBatchArgs, auth)
		// We check the errors via strings because we match the
		// definition of the error from geth, with the string returned
		// via RPC obtained by the client.
		if err == nil {
			break
		} else if strings.Contains(err.Error(), core.ErrNonceTooLow.Error()) {
			log.Warnw("TxManager ethClient.RollupForgeBatch incrementing nonce",
				"err", err, "nonce", auth.Nonce, "batchNum", batchInfo.BatchNum)
			auth.Nonce.Add(auth.Nonce, big.NewInt(1))
			attempt--
		} else if strings.Contains(err.Error(), core.ErrNonceTooHigh.Error()) {
			log.Warnw("TxManager ethClient.RollupForgeBatch decrementing nonce",
				"err", err, "nonce", auth.Nonce, "batchNum", batchInfo.BatchNum)
			auth.Nonce.Sub(auth.Nonce, big.NewInt(1))
			attempt--
		} else if strings.Contains(err.Error(), core.ErrReplaceUnderpriced.Error()) {
			log.Warnw("TxManager ethClient.RollupForgeBatch incrementing gasPrice",
				"err", err, "gasPrice", auth.GasPrice, "batchNum", batchInfo.BatchNum)
			auth.GasPrice = addPerc(auth.GasPrice, 10)
			attempt--
		} else if strings.Contains(err.Error(), core.ErrUnderpriced.Error()) {
			log.Warnw("TxManager ethClient.RollupForgeBatch incrementing gasPrice",
				"err", err, "gasPrice", auth.GasPrice, "batchNum", batchInfo.BatchNum)
			auth.GasPrice = addPerc(auth.GasPrice, 10)
			attempt--
		} else {
			log.Errorw("TxManager ethClient.RollupForgeBatch",
				"attempt", attempt, "err", err, "block", t.stats.Eth.LastBlock.Num+1,
				"batchNum", batchInfo.BatchNum)
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
	if !resend {
		t.accNextNonce = auth.Nonce.Uint64() + 1
	}
	batchInfo.EthTx = ethTx
	log.Infow("TxManager ethClient.RollupForgeBatch", "batch", batchInfo.BatchNum, "tx", ethTx.Hash())
	now := time.Now()
	batchInfo.SendTimestamp = now

	if resend {
		batchInfo.Debug.ResendNum++
	}
	batchInfo.Debug.Status = StatusSent
	batchInfo.Debug.SendBlockNum = t.stats.Eth.LastBlock.Num + 1
	batchInfo.Debug.SendTimestamp = batchInfo.SendTimestamp
	batchInfo.Debug.StartToSendDelay = batchInfo.Debug.SendTimestamp.Sub(
		batchInfo.Debug.StartTimestamp).Seconds()
	t.cfg.debugBatchStore(batchInfo)

	if !resend {
		if batchInfo.L1Batch {
			t.lastSentL1BatchBlockNum = t.stats.Eth.LastBlock.Num + 1
		}
	}
	if err := t.l2DB.DoneForging(common.TxIDsFromL2Txs(batchInfo.L2Txs), batchInfo.BatchNum); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// checkEthTransactionReceipt takes the txHash from the BatchInfo and stores
// the corresponding receipt if found
func (t *TxManager) checkEthTransactionReceipt(ctx context.Context, batchInfo *BatchInfo) error {
	txHash := batchInfo.EthTx.Hash()
	var receipt *types.Receipt
	var err error
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		receipt, err = t.ethClient.EthTransactionReceipt(ctx, txHash)
		if ctx.Err() != nil {
			continue
		} else if tracerr.Unwrap(err) == ethereum.NotFound {
			err = nil
			break
		} else if err != nil {
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

func (t *TxManager) handleReceipt(ctx context.Context, batchInfo *BatchInfo) (*int64, error) {
	receipt := batchInfo.Receipt
	if receipt != nil {
		if batchInfo.EthTx.Nonce()+1 > t.accNonce {
			t.accNonce = batchInfo.EthTx.Nonce() + 1
		}
		if receipt.Status == types.ReceiptStatusFailed {
			batchInfo.Debug.Status = StatusFailed
			_, err := t.ethClient.EthCall(ctx, batchInfo.EthTx, receipt.BlockNumber)
			log.Warnw("TxManager receipt status is failed", "tx", receipt.TxHash,
				"batch", batchInfo.BatchNum, "block", receipt.BlockNumber.Int64(),
				"err", err)
			batchInfo.EthTxErr = err
			if batchInfo.BatchNum <= t.lastSuccessBatch {
				t.lastSuccessBatch = batchInfo.BatchNum - 1
			}
			t.cfg.debugBatchStore(batchInfo)
			return nil, tracerr.Wrap(fmt.Errorf(
				"ethereum transaction receipt status is failed: %w", err))
		} else if receipt.Status == types.ReceiptStatusSuccessful {
			batchInfo.Debug.Status = StatusMined
			batchInfo.Debug.MineBlockNum = receipt.BlockNumber.Int64()
			batchInfo.Debug.StartToMineBlocksDelay = batchInfo.Debug.MineBlockNum -
				batchInfo.Debug.StartBlockNum
			if batchInfo.Debug.StartToMineDelay == 0 {
				if block, err := t.ethClient.EthBlockByNumber(ctx,
					receipt.BlockNumber.Int64()); err != nil {
					log.Warnw("TxManager: ethClient.EthBlockByNumber", "err", err)
				} else {
					batchInfo.Debug.SendToMineDelay = block.Timestamp.Sub(
						batchInfo.Debug.SendTimestamp).Seconds()
					batchInfo.Debug.StartToMineDelay = block.Timestamp.Sub(
						batchInfo.Debug.StartTimestamp).Seconds()
				}
			}
			t.cfg.debugBatchStore(batchInfo)
			if batchInfo.BatchNum > t.lastSuccessBatch {
				t.lastSuccessBatch = batchInfo.BatchNum
			}
			confirm := t.stats.Eth.LastBlock.Num - receipt.BlockNumber.Int64()
			return &confirm, nil
		}
	}
	return nil, nil
}

// TODO:
// - After sending a message: CancelPipeline, stop all consecutive pending Batches (transactions)

// Queue of BatchInfos
type Queue struct {
	list []*BatchInfo
	// nonceByBatchNum map[common.BatchNum]uint64
	next int
}

// NewQueue returns a new queue
func NewQueue() Queue {
	return Queue{
		list: make([]*BatchInfo, 0),
		// nonceByBatchNum: make(map[common.BatchNum]uint64),
		next: 0,
	}
}

// Len is the length of the queue
func (q *Queue) Len() int {
	return len(q.list)
}

// At returns the BatchInfo at position (or nil if position is out of bounds)
func (q *Queue) At(position int) *BatchInfo {
	if position >= len(q.list) {
		return nil
	}
	return q.list[position]
}

// Next returns the next BatchInfo (or nil if queue is empty)
func (q *Queue) Next() (int, *BatchInfo) {
	if len(q.list) == 0 {
		return 0, nil
	}
	defer func() { q.next = (q.next + 1) % len(q.list) }()
	return q.next, q.list[q.next]
}

// Remove removes the BatchInfo at position
func (q *Queue) Remove(position int) {
	// batchInfo := q.list[position]
	// delete(q.nonceByBatchNum, batchInfo.BatchNum)
	q.list = append(q.list[:position], q.list[position+1:]...)
	if len(q.list) == 0 {
		q.next = 0
	} else {
		q.next = position % len(q.list)
	}
}

// Push adds a new BatchInfo
func (q *Queue) Push(batchInfo *BatchInfo) {
	q.list = append(q.list, batchInfo)
	// q.nonceByBatchNum[batchInfo.BatchNum] = batchInfo.EthTx.Nonce()
}

// func (q *Queue) NonceByBatchNum(batchNum common.BatchNum) (uint64, bool) {
// 	nonce, ok := q.nonceByBatchNum[batchNum]
// 	return nonce, ok
// }

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	var statsVars statsVars
	select {
	case statsVars = <-t.statsVarsCh:
	case <-ctx.Done():
	}
	t.stats = statsVars.Stats
	t.syncSCVars(statsVars.Vars)
	log.Infow("TxManager: received initial statsVars",
		"block", t.stats.Eth.LastBlock.Num, "batch", t.stats.Eth.LastBatchNum)

	timer := time.NewTimer(longWaitDuration)
	for {
		select {
		case <-ctx.Done():
			log.Info("TxManager done")
			return
		case statsVars := <-t.statsVarsCh:
			t.stats = statsVars.Stats
			t.syncSCVars(statsVars.Vars)
		case pipelineNum := <-t.discardPipelineCh:
			t.minPipelineNum = pipelineNum + 1
			if err := t.removeBadBatchInfos(ctx); ctx.Err() != nil {
				continue
			} else if err != nil {
				log.Errorw("TxManager: removeBadBatchInfos", "err", err)
				continue
			}
		case batchInfo := <-t.batchCh:
			if batchInfo.PipelineNum < t.minPipelineNum {
				log.Warnw("TxManager: batchInfo received pipelineNum < minPipelineNum",
					"num", batchInfo.PipelineNum, "minNum", t.minPipelineNum)
			}
			if err := t.shouldSendRollupForgeBatch(batchInfo); err != nil {
				log.Warnw("TxManager: shouldSend", "err", err,
					"batch", batchInfo.BatchNum)
				t.coord.SendMsg(ctx, MsgStopPipeline{
					Reason: fmt.Sprintf("forgeBatch shouldSend: %v", err)})
				continue
			}
			if err := t.sendRollupForgeBatch(ctx, batchInfo, false); ctx.Err() != nil {
				continue
			} else if err != nil {
				// If we reach here it's because our ethNode has
				// been unable to send the transaction to
				// ethereum.  This could be due to the ethNode
				// failure, or an invalid transaction (that
				// can't be mined)
				log.Warnw("TxManager: forgeBatch send failed", "err", err,
					"batch", batchInfo.BatchNum)
				t.coord.SendMsg(ctx, MsgStopPipeline{
					Reason: fmt.Sprintf("forgeBatch send: %v", err)})
				continue
			}
			t.queue.Push(batchInfo)
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(t.cfg.TxManagerCheckInterval)
		case <-timer.C:
			queuePosition, batchInfo := t.queue.Next()
			if batchInfo == nil {
				timer.Reset(longWaitDuration)
				continue
			}
			timer.Reset(t.cfg.TxManagerCheckInterval)
			if err := t.checkEthTransactionReceipt(ctx, batchInfo); ctx.Err() != nil {
				continue
			} else if err != nil { //nolint:staticcheck
				// Our ethNode is giving an error different
				// than "not found" when getting the receipt
				// for the transaction, so we can't figure out
				// if it was not mined, mined and succesfull or
				// mined and failed.  This could be due to the
				// ethNode failure.
				t.coord.SendMsg(ctx, MsgStopPipeline{
					Reason: fmt.Sprintf("forgeBatch receipt: %v", err)})
			}

			confirm, err := t.handleReceipt(ctx, batchInfo)
			if ctx.Err() != nil {
				continue
			} else if err != nil { //nolint:staticcheck
				// Transaction was rejected
				if err := t.removeBadBatchInfos(ctx); ctx.Err() != nil {
					continue
				} else if err != nil {
					log.Errorw("TxManager: removeBadBatchInfos", "err", err)
					continue
				}
				t.coord.SendMsg(ctx, MsgStopPipeline{
					Reason: fmt.Sprintf("forgeBatch reject: %v", err)})
				continue
			}
			now := time.Now()
			if !t.cfg.EthNoReuseNonce && confirm == nil &&
				now.Sub(batchInfo.SendTimestamp) > t.cfg.EthTxResendTimeout {
				log.Infow("TxManager: forgeBatch tx not been mined timeout, resending",
					"tx", batchInfo.EthTx.Hash(), "batch", batchInfo.BatchNum)
				if err := t.sendRollupForgeBatch(ctx, batchInfo, true); ctx.Err() != nil {
					continue
				} else if err != nil {
					// If we reach here it's because our ethNode has
					// been unable to send the transaction to
					// ethereum.  This could be due to the ethNode
					// failure, or an invalid transaction (that
					// can't be mined)
					log.Warnw("TxManager: forgeBatch resend failed", "err", err,
						"batch", batchInfo.BatchNum)
					t.coord.SendMsg(ctx, MsgStopPipeline{
						Reason: fmt.Sprintf("forgeBatch resend: %v", err)})
					continue
				}
			}

			if confirm != nil && *confirm >= t.cfg.ConfirmBlocks {
				log.Debugw("TxManager: forgeBatch tx confirmed",
					"tx", batchInfo.EthTx.Hash(), "batch", batchInfo.BatchNum)
				t.queue.Remove(queuePosition)
			}
		}
	}
}

func (t *TxManager) removeBadBatchInfos(ctx context.Context) error {
	next := 0
	for {
		batchInfo := t.queue.At(next)
		if batchInfo == nil {
			break
		}
		if err := t.checkEthTransactionReceipt(ctx, batchInfo); ctx.Err() != nil {
			return nil
		} else if err != nil {
			// Our ethNode is giving an error different
			// than "not found" when getting the receipt
			// for the transaction, so we can't figure out
			// if it was not mined, mined and succesfull or
			// mined and failed.  This could be due to the
			// ethNode failure.
			next++
			continue
		}
		confirm, err := t.handleReceipt(ctx, batchInfo)
		if ctx.Err() != nil {
			return nil
		} else if err != nil {
			// Transaction was rejected
			if t.minPipelineNum <= batchInfo.PipelineNum {
				t.minPipelineNum = batchInfo.PipelineNum + 1
			}
			t.queue.Remove(next)
			continue
		}
		// If tx is pending but is from a cancelled pipeline, remove it
		// from the queue
		if confirm == nil {
			if batchInfo.PipelineNum < t.minPipelineNum {
				t.queue.Remove(next)
				continue
			}
		}
		next++
	}
	accNonce, err := t.ethClient.EthNonceAt(ctx, t.account.Address, nil)
	if err != nil {
		return err
	}
	if !t.cfg.EthNoReuseNonce {
		t.accNextNonce = accNonce
	}
	return nil
}

func (t *TxManager) canForgeAt(blockNum int64) bool {
	return canForge(&t.consts.Auction, &t.vars.Auction,
		&t.stats.Sync.Auction.CurrentSlot, &t.stats.Sync.Auction.NextSlot,
		t.cfg.ForgerAddress, blockNum)
}

func (t *TxManager) mustL1L2Batch(blockNum int64) bool {
	lastL1BatchBlockNum := t.lastSentL1BatchBlockNum
	if t.stats.Sync.LastL1BatchBlock > lastL1BatchBlockNum {
		lastL1BatchBlockNum = t.stats.Sync.LastL1BatchBlock
	}
	return blockNum-lastL1BatchBlockNum >= t.vars.Rollup.ForgeL1L2BatchTimeout-1
}
