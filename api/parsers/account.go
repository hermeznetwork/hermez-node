package parsers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type accountFilter struct {
	AccountIndex string `uri:"accountIndex" binding:"required"`
}

func ParseAccountFilter(c *gin.Context) (*common.Idx, error) {
	var accountFilter accountFilter
	if err := c.ShouldBindUri(&accountFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return common.StringToIdx(accountFilter.AccountIndex, "accountIndex")
}

type accountsFilter struct {
	IDs  string `form:"tokenIds"`
	Addr string `form:"hezEthereumAddress"`
	Bjj  string `form:"BJJ"`

	Pagination
}

func ParseAccountsFilters(c *gin.Context) (historydb.GetAccountsAPIRequest, error) {
	var accountsFilter accountsFilter
	if err := c.BindQuery(&accountsFilter); err != nil {
		return historydb.GetAccountsAPIRequest{}, err
	}

	var tokenIDs []common.TokenID
	if accountsFilter.IDs != "" {
		ids := strings.Split(accountsFilter.IDs, ",")
		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return historydb.GetAccountsAPIRequest{}, err
			}
			tokenID := common.TokenID(idUint)
			tokenIDs = append(tokenIDs, tokenID)
		}
	}

	addr, err := common.HezStringToEthAddr(accountsFilter.Addr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetAccountsAPIRequest{}, err
	}

	bjj, err := common.HezStringToBJJ(accountsFilter.Bjj, "BJJ")
	if err != nil {
		return historydb.GetAccountsAPIRequest{}, err
	}

	if addr != nil && bjj != nil {
		return historydb.GetAccountsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	return historydb.GetAccountsAPIRequest{
		TokenIDs: tokenIDs,
		EthAddr:  addr,
		Bjj:      bjj,
		FromItem: accountsFilter.FromItem,
		Order:    *accountsFilter.Order,
		Limit:    accountsFilter.Limit,
	}, nil
}
