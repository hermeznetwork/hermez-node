package coordinator

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

	queue []*BatchInfo
	// lastSuccessBatch stores the last BatchNum that who's forge call was confirmed
	lastSuccessBatch common.BatchNum
	lastPendingBatch common.BatchNum
	lastSuccessNonce uint64
	lastPendingNonce uint64
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
	lastSuccessNonce, err := ethClient.EthNonceAt(ctx, *address, nil)
	if err != nil {
		return nil, err
	}
	lastPendingNonce, err := ethClient.EthPendingNonceAt(ctx, *address)
	if err != nil {
		return nil, err
	}
	if lastSuccessNonce != lastPendingNonce {
		return nil, tracerr.Wrap(fmt.Errorf("lastSuccessNonce (%v) != lastPendingNonce (%v)",
			lastSuccessNonce, lastPendingNonce))
	}
	log.Infow("TxManager started", "nonce", lastSuccessNonce)
	return &TxManager{
		cfg:         *cfg,
		ethClient:   ethClient,
		l2DB:        l2DB,
		coord:       coord,
		batchCh:     make(chan *BatchInfo, queueLen),
		statsVarsCh: make(chan statsVars, queueLen),
		account: accounts.Account{
			Address: *address,
		},
		chainID: chainID,
		consts:  *scConsts,

		vars: *initSCVars,

		lastSuccessNonce: lastSuccessNonce,
		lastPendingNonce: lastPendingNonce,
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

func (t *TxManager) syncSCVars(vars synchronizer.SCVariablesPtr) {
	if vars.Rollup != nil {
		t.vars.Rollup = *vars.Rollup
	}
	if vars.Auction != nil {
		t.vars.Auction = *vars.Auction
	}
	if vars.WDelayer != nil {
		t.vars.WDelayer = *vars.WDelayer
	}
}

// NewAuth generates a new auth object for an ethereum transaction
func (t *TxManager) NewAuth(ctx context.Context) (*bind.TransactOpts, error) {
	gasPrice, err := t.ethClient.EthSuggestGasPrice(ctx)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	inc := new(big.Int).Set(gasPrice)
	const gasPriceDiv = 100
	inc.Div(inc, new(big.Int).SetUint64(gasPriceDiv))
	gasPrice.Add(gasPrice, inc)
	// log.Debugw("TxManager: transaction metadata", "gasPrice", gasPrice)

	auth, err := bind.NewKeyStoreTransactorWithChainID(t.ethClient.EthKeyStore(), t.account, t.chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auth.Value = big.NewInt(0) // in wei
	// TODO: Calculate GasLimit based on the contents of the ForgeBatchArgs
	auth.GasLimit = 1000000
	auth.GasPrice = gasPrice
	auth.Nonce = nil

	return auth, nil
}

func (t *TxManager) sendRollupForgeBatch(ctx context.Context, batchInfo *BatchInfo) error {
	// TODO: Check if we can forge in the next blockNum, abort if we can't
	batchInfo.Debug.Status = StatusSent
	batchInfo.Debug.SendBlockNum = t.stats.Eth.LastBlock.Num + 1
	batchInfo.Debug.SendTimestamp = time.Now()
	batchInfo.Debug.StartToSendDelay = batchInfo.Debug.SendTimestamp.Sub(
		batchInfo.Debug.StartTimestamp).Seconds()
	var ethTx *types.Transaction
	var err error
	auth, err := t.NewAuth(ctx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	auth.Nonce = big.NewInt(int64(t.lastPendingNonce))
	t.lastPendingNonce++
	for attempt := 0; attempt < t.cfg.EthClientAttempts; attempt++ {
		ethTx, err = t.ethClient.RollupForgeBatch(batchInfo.ForgeBatchArgs, auth)
		if err != nil {
			// if strings.Contains(err.Error(), common.AuctionErrMsgCannotForge) {
			// 	log.Errorw("TxManager ethClient.RollupForgeBatch", "err", err,
			// 		"block", t.stats.Eth.LastBlock.Num+1)
			// 	return tracerr.Wrap(err)
			// }
			log.Errorw("TxManager ethClient.RollupForgeBatch",
				"attempt", attempt, "err", err, "block", t.stats.Eth.LastBlock.Num+1,
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
	t.lastPendingBatch = batchInfo.BatchNum
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
		if receipt.Status == types.ReceiptStatusFailed {
			batchInfo.Debug.Status = StatusFailed
			t.cfg.debugBatchStore(batchInfo)
			_, err := t.ethClient.EthCall(ctx, batchInfo.EthTx, receipt.BlockNumber)
			log.Warnw("TxManager receipt status is failed", "tx", receipt.TxHash.Hex(),
				"batch", batchInfo.BatchNum, "block", receipt.BlockNumber.Int64(),
				"err", err)
			return nil, tracerr.Wrap(fmt.Errorf(
				"ethereum transaction receipt status is failed: %w", err))
		} else if receipt.Status == types.ReceiptStatusSuccessful {
			batchInfo.Debug.Status = StatusMined
			batchInfo.Debug.MineBlockNum = receipt.BlockNumber.Int64()
			batchInfo.Debug.StartToMineBlocksDelay = batchInfo.Debug.MineBlockNum -
				batchInfo.Debug.StartBlockNum
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

// Run the TxManager
func (t *TxManager) Run(ctx context.Context) {
	next := 0
	waitDuration := longWaitDuration

	var statsVars statsVars
	select {
	case statsVars = <-t.statsVarsCh:
	case <-ctx.Done():
	}
	t.stats = statsVars.Stats
	t.syncSCVars(statsVars.Vars)
	log.Infow("TxManager: received initial statsVars",
		"block", t.stats.Eth.LastBlock.Num, "batch", t.stats.Eth.LastBatch)

	for {
		select {
		case <-ctx.Done():
			log.Info("TxManager done")
			return
		case statsVars := <-t.statsVarsCh:
			t.stats = statsVars.Stats
			t.syncSCVars(statsVars.Vars)
		case batchInfo := <-t.batchCh:
			if err := t.sendRollupForgeBatch(ctx, batchInfo); ctx.Err() != nil {
				continue
			} else if err != nil {
				// If we reach here it's because our ethNode has
				// been unable to send the transaction to
				// ethereum.  This could be due to the ethNode
				// failure, or an invalid transaction (that
				// can't be mined)
				t.coord.SendMsg(ctx, MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch send: %v", err)})
				continue
			}
			t.queue = append(t.queue, batchInfo)
			waitDuration = t.cfg.TxManagerCheckInterval
		case <-time.After(waitDuration):
			if len(t.queue) == 0 {
				waitDuration = longWaitDuration
				continue
			}
			current := next
			next = (current + 1) % len(t.queue)
			batchInfo := t.queue[current]
			if err := t.checkEthTransactionReceipt(ctx, batchInfo); ctx.Err() != nil {
				continue
			} else if err != nil { //nolint:staticcheck
				// Our ethNode is giving an error different
				// than "not found" when getting the receipt
				// for the transaction, so we can't figure out
				// if it was not mined, mined and succesfull or
				// mined and failed.  This could be due to the
				// ethNode failure.
				t.coord.SendMsg(ctx, MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch receipt: %v", err)})
			}

			confirm, err := t.handleReceipt(ctx, batchInfo)
			if ctx.Err() != nil {
				continue
			} else if err != nil { //nolint:staticcheck
				// Transaction was rejected
				t.queue = append(t.queue[:current], t.queue[current+1:]...)
				if len(t.queue) == 0 {
					next = 0
				} else {
					next = current % len(t.queue)
				}
				t.coord.SendMsg(ctx, MsgStopPipeline{Reason: fmt.Sprintf("forgeBatch reject: %v", err)})
			}
			if confirm != nil && *confirm >= t.cfg.ConfirmBlocks {
				log.Debugw("TxManager tx for RollupForgeBatch confirmed",
					"batch", batchInfo.BatchNum)
				t.queue = append(t.queue[:current], t.queue[current+1:]...)
				if len(t.queue) == 0 {
					next = 0
				} else {
					next = current % len(t.queue)
				}
			}
		}
	}
}

// nolint reason: this function will be used in the future
//nolint:unused
func (t *TxManager) canForge(stats *synchronizer.Stats, blockNum int64) bool {
	return canForge(&t.consts.Auction, &t.vars.Auction,
		&stats.Sync.Auction.CurrentSlot, &stats.Sync.Auction.NextSlot,
		t.cfg.ForgerAddress, blockNum)
}
