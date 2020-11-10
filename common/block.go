package common

import (
	"math/big"
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

// RollupData contains information returned by the Rollup smart contract
type RollupData struct {
	// L1UserTxs that were submitted in the block
	L1UserTxs   []L1Tx
	Batches     []BatchData
	AddedTokens []Token
	Withdrawals []WithdrawInfo
	Vars        *RollupVariables
}

// NewRollupData creates an empty RollupData with the slices initialized.
func NewRollupData() RollupData {
	return RollupData{
		L1UserTxs:   make([]L1Tx, 0),
		Batches:     make([]BatchData, 0),
		AddedTokens: make([]Token, 0),
		Withdrawals: make([]WithdrawInfo, 0),
		Vars:        nil,
	}
}

// AuctionData contains information returned by the Action smart contract
type AuctionData struct {
	Bids         []Bid
	Coordinators []Coordinator
	Vars         *AuctionVariables
}

// NewAuctionData creates an empty AuctionData with the slices initialized.
func NewAuctionData() AuctionData {
	return AuctionData{
		Bids:         make([]Bid, 0),
		Coordinators: make([]Coordinator, 0),
		Vars:         nil,
	}
}

// WDelayerTransfer represents a transfer (either deposit or withdrawal) in the
// WDelayer smart contract
type WDelayerTransfer struct {
	Owner  ethCommon.Address
	Token  ethCommon.Address
	Amount *big.Int
	// TxHash ethCommon.Hash // hash of the transaction in which the wdelayer transfer happened
}

// WDelayerData contains information returned by the WDelayer smart contract
type WDelayerData struct {
	Vars             *WDelayerVariables
	Deposits         []WDelayerTransfer
	DepositsByTxHash map[ethCommon.Hash]*WDelayerTransfer
	Withdrawals      []WDelayerTransfer
}

// NewWDelayerData creates an empty WDelayerData.
func NewWDelayerData() WDelayerData {
	return WDelayerData{
		Vars:             nil,
		Deposits:         make([]WDelayerTransfer, 0),
		DepositsByTxHash: make(map[ethCommon.Hash]*WDelayerTransfer),
		Withdrawals:      make([]WDelayerTransfer, 0),
	}
}

// BlockData contains the information of a Block
type BlockData struct {
	Block    Block
	Rollup   RollupData
	Auction  AuctionData
	WDelayer WDelayerData
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
