package requests

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

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
