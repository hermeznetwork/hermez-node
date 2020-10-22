package common

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Block represents of an Ethereum block
type Block struct {
	EthBlockNum int64          `meddler:"eth_block_num"`
	Timestamp   time.Time      `meddler:"timestamp,utctime"`
	Hash        ethCommon.Hash `meddler:"hash"`
	ParentHash  ethCommon.Hash `meddler:"-"`
}

// BlockData contains the information of a Block
type BlockData struct {
	Block Block
	// Rollup
	// L1UserTxs that were submitted in the block
	L1UserTxs   []L1Tx
	Batches     []BatchData
	AddedTokens []Token
	RollupVars  *RollupVars
	// Auction
	Bids                []Bid
	Coordinators        []Coordinator
	AuctionVars         *AuctionVars
	WithdrawDelayerVars *WithdrawDelayerVars
	// TODO: enable when common.WithdrawalDelayerVars is Merged from Synchronizer PR
	// WithdrawalDelayerVars *common.WithdrawalDelayerVars
}

// BatchData contains the information of a Batch
type BatchData struct {
	// L1UserTxs that were forged in the batch
	L1Batch bool // TODO: Remove once Batch.ForgeL1TxsNum is a pointer
	// L1UserTxs        []common.L1Tx
	L1CoordinatorTxs []L1Tx
	L2Txs            []L2Tx
	CreatedAccounts  []Account
	ExitTree         []ExitInfo
	Batch            Batch
}

// NewBatchData creates an empty BatchData with the slices initialized.
func NewBatchData() *BatchData {
	return &BatchData{
		L1Batch: false,
		// L1UserTxs:        make([]common.L1Tx, 0),
		L1CoordinatorTxs: make([]L1Tx, 0),
		L2Txs:            make([]L2Tx, 0),
		CreatedAccounts:  make([]Account, 0),
		ExitTree:         make([]ExitInfo, 0),
		Batch:            Batch{},
	}
}
