package parsers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type exitFilter struct {
	BatchNum     uint   `uri:"batchNum" binding:"required"`
	AccountIndex string `uri:"accountIndex" binding:"required"`
}

func ParseExitFilter(c *gin.Context) (*uint, *common.Idx, error) {
	var exitFilter exitFilter
	if err := c.ShouldBindUri(&exitFilter); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	idx, err := common.StringToIdx(exitFilter.AccountIndex, "accountIndex")
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	return &exitFilter.BatchNum, idx, nil
}

type exitsFilters struct {
	TokenID              *uint  `form:"tokenId"`
	Addr                 string `form:"hezEthereumAddress"`
	Bjj                  string `form:"BJJ"`
	AccountIndex         string `form:"accountIndex"`
	BatchNum             *uint  `form:"batchNum"`
	OnlyPendingWithdraws *bool  `form:"onlyPendingWithdraws"`

	Pagination
}

func ParseExitsFilters(c *gin.Context) (historydb.GetExitsAPIRequest, error) {
	var exitsFilters exitsFilters
	if err := c.ShouldBindQuery(&exitsFilters); err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	// Token ID
	var tokenID *common.TokenID
	if exitsFilters.TokenID != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*exitsFilters.TokenID)
	}

	addr, err := common.HezStringToEthAddr(exitsFilters.Addr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := common.HezStringToBJJ(exitsFilters.Bjj, "BJJ")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	if addr != nil && bjj != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	idx, err := common.StringToIdx(exitsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	if idx != nil && (addr != nil || bjj != nil || tokenID != nil) {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}

	return historydb.GetExitsAPIRequest{
		EthAddr:              addr,
		Bjj:                  bjj,
		TokenID:              tokenID,
		Idx:                  idx,
		BatchNum:             exitsFilters.BatchNum,
		OnlyPendingWithdraws: exitsFilters.OnlyPendingWithdraws,
		FromItem:             exitsFilters.FromItem,
		Limit:                exitsFilters.Limit,
		Order:                *exitsFilters.Order,
	}, nil
}
