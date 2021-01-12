package coordinator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

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

func (t *TxManager) callRollupForgeBatch(ctx context.Context, batchInfo *BatchInfo) error {
	batchInfo.Debug.Status = StatusSent
	batchInfo.Debug.SendBlockNum = t.lastBlock + 1
	batchInfo.Debug.SendTimestamp = time.Now()
	batchInfo.Debug.StartToSendDelay = batchInfo.Debug.SendTimestamp.Sub(
		batchInfo.Debug.StartTimestamp).Seconds()
	var ethTx *types.Transaction
	var err error
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		ethTx, err = t.ethClient.RollupForgeBatch(batchInfo.ForgeBatchArgs)
		if err != nil {
			if strings.Contains(err.Error(), common.AuctionErrMsgCannotForge) {
				log.Debugw("TxManager ethClient.RollupForgeBatch", "err", err,
					"block", t.lastBlock+1)
				return tracerr.Wrap(err)
			}
			log.Errorw("TxManager ethClient.RollupForgeBatch",
				"attempt", attempt, "err", err, "block", t.lastBlock+1,
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

func (t *TxManager) checkEthTransactionReceipt(ctx context.Context, batchInfo *BatchInfo) error {
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
			batchInfo.Debug.Status = StatusFailed
			t.cfg.debugBatchStore(batchInfo)
			log.Errorw("TxManager receipt status is failed", "receipt", receipt)
			return nil, tracerr.Wrap(fmt.Errorf("ethereum transaction receipt statis is failed"))
		} else if receipt.Status == types.ReceiptStatusSuccessful {
			batchInfo.Debug.Status = StatusMined
			batchInfo.Debug.MineBlockNum = receipt.BlockNumber.Int64()
			batchInfo.Debug.StartToMineBlocksDelay = batchInfo.Debug.MineBlockNum -
				batchInfo.Debug.StartBlockNum
			t.cfg.debugBatchStore(batchInfo)
			if batchInfo.BatchNum > t.lastConfirmedBatch {
				t.lastConfirmedBatch = batchInfo.BatchNum
			}
			confirm := t.lastBlock - receipt.BlockNumber.Int64()
			return &confirm, nil
		}
	}
	return nil, nil
}

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	next := 0
	waitDuration := longWaitDuration

	for {
		select {
		case <-ctx.Done():
			log.Info("TxManager done")
			return
		case lastBlock := <-t.lastBlockCh:
			t.lastBlock = lastBlock
		case batchInfo := <-t.batchCh:
			if err := t.callRollupForgeBatch(ctx, batchInfo); ctx.Err() != nil {
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
			if err := t.checkEthTransactionReceipt(ctx, batchInfo); ctx.Err() != nil {
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
