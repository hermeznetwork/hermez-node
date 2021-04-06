package requests

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// API request struct for getting txs
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
