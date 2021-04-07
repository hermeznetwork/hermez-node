package requests

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// GetAccountsAPIRequest is an API request struct for getting accounts
type GetAccountsAPIRequest struct {
	TokenIDs []common.TokenID
	EthAddr  *ethCommon.Address
	Bjj      *babyjub.PublicKeyComp

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetTxsAPIRequest is an API request struct for getting txs
type GetTxsAPIRequest struct {
	EthAddr           *ethCommon.Address
	Bjj               *babyjub.PublicKeyComp
	TokenID           *common.TokenID
	Idx               *common.Idx
	BatchNum          *uint
	TxType            *common.TxType
	IncludePendingL1s *bool

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetBatchesAPIRequest is an API request struct for getting batches
type GetBatchesAPIRequest struct {
	MinBatchNum *uint
	MaxBatchNum *uint
	SlotNum     *uint
	ForgerAddr  *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetBestBidsAPIRequest is an API request struct for getting best bids
type GetBestBidsAPIRequest struct {
	MinSlotNum *int64
	MaxSlotNum *int64
	BidderAddr *ethCommon.Address

	Limit *uint
	Order string
}

// GetTokensAPIRequest is an API request struct for getting tokens
type GetTokensAPIRequest struct {
	Ids     []common.TokenID
	Symbols []string
	Name    string

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetExitsAPIRequest is an API request struct for getting exits
type GetExitsAPIRequest struct {
	EthAddr              *ethCommon.Address
	Bjj                  *babyjub.PublicKeyComp
	TokenID              *common.TokenID
	Idx                  *common.Idx
	BatchNum             *uint
	OnlyPendingWithdraws *bool

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetCoordinatorsAPIRequest is an API request struct for getting coordinators
type GetCoordinatorsAPIRequest struct {
	BidderAddr *ethCommon.Address
	ForgerAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetBidsAPIRequest is an API request struct for getting bids
type GetBidsAPIRequest struct {
	SlotNum    *int64
	BidderAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
