package parsers

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
)

type getAccountCreationAuthFilter struct {
	Addr string `uri:"hezEthereumAddress" binding:"required"`
}

func ParseGetAccountCreationAuthFilter(c *gin.Context) (*ethCommon.Address, error) {
	var getAccountCreationAuthFilter getAccountCreationAuthFilter
	if err := c.ShouldBindUri(&getAccountCreationAuthFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return common.HezStringToEthAddr(getAccountCreationAuthFilter.Addr, "hezEthereumAddress")
}
