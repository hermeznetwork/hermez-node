package requests

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type GetAccountsAPIRequest struct {
	TokenIDs []common.TokenID
	EthAddr  *ethCommon.Address
	Bjj      *babyjub.PublicKeyComp

	FromItem *uint
	Limit    *uint
	Order    string
}
