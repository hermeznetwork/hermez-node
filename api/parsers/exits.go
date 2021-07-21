package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// ExitFilter struct to hold exit filter
type ExitFilter struct {
	BatchNum     uint   `uri:"batchNum" binding:"required"`
	AccountIndex string `uri:"accountIndex" binding:"required"`
}

// ParseExitFilter func parsing exit filter from the /exits request to the accountIndex and batchNum
func ParseExitFilter(c *gin.Context) (*uint, *common.Idx, error) {
	var exitFilter ExitFilter
	if err := c.ShouldBindUri(&exitFilter); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	queryAccount, err := common.StringToIdx(exitFilter.AccountIndex, "accountIndex")
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	return &exitFilter.BatchNum, queryAccount.AccountIndex, nil
}

// ExitsFilters struct for holding exits filters
type ExitsFilters struct {
	TokenID              *uint  `form:"tokenId"`
	Addr                 string `form:"hezEthereumAddress"`
	Bjj                  string `form:"BJJ"`
	AccountIndex         string `form:"accountIndex"`
	BatchNum             *uint  `form:"batchNum"`
	OnlyPendingWithdraws *bool  `form:"onlyPendingWithdraws"`

	Pagination
}

// ExitsFiltersStructValidation func validates ExitsFilters
func ExitsFiltersStructValidation(sl validator.StructLevel) {
	ef := sl.Current().Interface().(ExitsFilters)

	if ef.Addr != "" && ef.Bjj != "" {
		sl.ReportError(ef.Addr, "hezEthereumAddress", "Addr", "hezethaddrorbjj", "")
		sl.ReportError(ef.Bjj, "BJJ", "Bjj", "hezethaddrorbjj", "")
	}

	if ef.AccountIndex != "" && (ef.Addr != "" || ef.Bjj != "" || ef.TokenID != nil) {
		sl.ReportError(ef.AccountIndex, "accountIndex", "AccountIndex", "onlyaccountindex", "")
		sl.ReportError(ef.Addr, "hezEthereumAddress", "Addr", "onlyaccountindex", "")
		sl.ReportError(ef.Bjj, "BJJ", "Bjj", "onlyaccountindex", "")
		sl.ReportError(ef.TokenID, "tokenId", "TokenID", "onlyaccountindex", "")
	}
}

// ParseExitsFilters func parsing exits filters
func ParseExitsFilters(c *gin.Context, v *validator.Validate) (historydb.GetExitsAPIRequest, error) {
	var exitsFilters ExitsFilters
	if err := c.ShouldBindQuery(&exitsFilters); err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	if err := v.Struct(exitsFilters); err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

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

	queryAccount, err := common.StringToIdx(exitsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetExitsAPIRequest{
		EthAddr:              addr,
		Bjj:                  bjj,
		TokenID:              tokenID,
		Idx:                  queryAccount.AccountIndex,
		BatchNum:             exitsFilters.BatchNum,
		OnlyPendingWithdraws: exitsFilters.OnlyPendingWithdraws,
		FromItem:             exitsFilters.FromItem,
		Limit:                exitsFilters.Limit,
		Order:                *exitsFilters.Order,
	}, nil
}
