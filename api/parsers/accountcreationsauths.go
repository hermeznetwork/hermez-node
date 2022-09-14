package parsers

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
)

// GetAccountCreationAuthFilter struct for parsing hezEthereumAddress from /account-creation-authorization/:hezEthereumAddress request
type GetAccountCreationAuthFilter struct {
	Addr string `uri:"hez:0xef2d4ea4f3c485bb47059b01b894a6d433504d9f" binding:"required"`
}

// ParseGetAccountCreationAuthFilter parsing uri request to the eth address
func ParseGetAccountCreationAuthFilter(c *gin.Context) (*ethCommon.Address, error) {
	var getAccountCreationAuthFilter GetAccountCreationAuthFilter
	if err := c.ShouldBindUri(&getAccountCreationAuthFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return common.HezStringToEthAddr(getAccountCreationAuthFilter.Addr, "hez:0xef2d4ea4f3c485bb47059b01b894a6d433504d9f")
}
