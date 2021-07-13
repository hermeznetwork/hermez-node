package parsers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// HistoryTxFilter struct to get history tx uri param from /transaction-history/:id request
type HistoryTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

// ParseHistoryTxFilter function for parsing history tx filter to the txID
func ParseHistoryTxFilter(c *gin.Context) (common.TxID, error) {
	var historyTxFilter HistoryTxFilter
	if err := c.ShouldBindUri(&historyTxFilter); err != nil {
		return common.TxID{}, tracerr.Wrap(err)
	}
	txID, err := common.NewTxIDFromString(historyTxFilter.TxID)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid txID"))
	}
	return txID, nil
}

// HistoryTxsFilters struct for holding filters from the /transaction-history request
type HistoryTxsFilters struct {
	TokenID             *uint  `form:"tokenId"`
	HezEthereumAddr     string `form:"hezEthereumAddress"`
	FromHezEthereumAddr string `form:"fromHezEthereumAddress"`
	ToHezEthereumAddr   string `form:"toHezEthereumAddress"`
	Bjj                 string `form:"BJJ"`
	ToBjj               string `form:"toBJJ"`
	FromBjj             string `form:"fromBJJ"`
	AccountIndex        string `form:"accountIndex"`
	FromAccountIndex    string `form:"fromAccountIndex"`
	ToAccountIndex      string `form:"toAccountIndex"`
	BatchNum            *uint  `form:"batchNum"`
	TxType              string `form:"type"`
	IncludePendingTxs   *bool  `form:"includePendingL1s"`

	Pagination
}

// HistoryTxsFiltersStructValidation func to validate history txs filters
func HistoryTxsFiltersStructValidation(sl validator.StructLevel) {
	ef := sl.Current().Interface().(HistoryTxsFilters)

	isAddrNotNil := ef.HezEthereumAddr != "" || ef.ToHezEthereumAddr != "" || ef.FromHezEthereumAddr != ""
	isBjjNotNil := ef.Bjj != "" || ef.ToBjj != "" || ef.FromBjj != ""

	if isAddrNotNil && isBjjNotNil {
		sl.ReportError(ef.HezEthereumAddr, "hezEthereumAddress", "HezEthereumAddr", "hezethaddrorbjj", "")
		sl.ReportError(ef.FromHezEthereumAddr, "fromHezEthereumAddress", "FromHezEthereumAddr", "hezethaddrorbjj", "")
		sl.ReportError(ef.ToHezEthereumAddr, "toHezEthereumAddress", "ToHezEthereumAddr", "hezethaddrorbjj", "")
		sl.ReportError(ef.Bjj, "BJJ", "Bjj", "hezethaddrorbjj", "")
		sl.ReportError(ef.ToBjj, "toBJJ", "ToBjj", "hezethaddrorbjj", "")
		sl.ReportError(ef.FromBjj, "fromBJJ", "FromBjj", "hezethaddrorbjj", "")
	}

	isIdxNotNil := ef.FromAccountIndex != "" || ef.ToAccountIndex != "" || ef.AccountIndex != ""
	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || ef.TokenID != nil) {
		sl.ReportError(ef.HezEthereumAddr, "hezEthereumAddress", "HezEthereumAddr", "onlyaccountindex", "")
		sl.ReportError(ef.FromHezEthereumAddr, "fromHezEthereumAddress", "FromHezEthereumAddr", "onlyaccountindex", "")
		sl.ReportError(ef.ToHezEthereumAddr, "toHezEthereumAddress", "ToHezEthereumAddr", "onlyaccountindex", "")
		sl.ReportError(ef.Bjj, "BJJ", "Bjj", "onlyaccountindex", "")
		sl.ReportError(ef.ToBjj, "toBJJ", "ToBjj", "onlyaccountindex", "")
		sl.ReportError(ef.FromBjj, "fromBJJ", "FromBjj", "onlyaccountindex", "")
		sl.ReportError(ef.AccountIndex, "accountIndex", "AccountIndex", "onlyaccountindex", "")
		sl.ReportError(ef.FromAccountIndex, "fromAccountIndex", "FromAccountIndex", "onlyaccountindex", "")
		sl.ReportError(ef.ToAccountIndex, "toAccountIndex", "ToAccountIndex", "onlyaccountindex", "")
	}
}

// ParseHistoryTxsFilters func to parse history txs filters from query to the GetTxsAPIRequest
func ParseHistoryTxsFilters(c *gin.Context, v *validator.Validate) (historydb.GetTxsAPIRequest, error) {
	var historyTxsFilters HistoryTxsFilters
	if err := c.ShouldBindQuery(&historyTxsFilters); err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	if err := v.Struct(historyTxsFilters); err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}
	// TokenID
	var tokenID *common.TokenID
	if historyTxsFilters.TokenID != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*historyTxsFilters.TokenID)
	}

	addr, err := common.HezStringToEthAddr(historyTxsFilters.HezEthereumAddr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromAddr, err := common.HezStringToEthAddr(historyTxsFilters.FromHezEthereumAddr, "fromHezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toAddr, err := common.HezStringToEthAddr(historyTxsFilters.ToHezEthereumAddr, "toHezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := common.HezStringToBJJ(historyTxsFilters.Bjj, "BJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromBjj, err := common.HezStringToBJJ(historyTxsFilters.FromBjj, "fromBJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toBjj, err := common.HezStringToBJJ(historyTxsFilters.ToBjj, "toBJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// Idx
	queryAccount, err := common.StringToIdx(historyTxsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromQueryAccount, err := common.StringToIdx(historyTxsFilters.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toQueryAccount, err := common.StringToIdx(historyTxsFilters.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	txType, err := common.StringToTxType(historyTxsFilters.TxType)
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetTxsAPIRequest{
		EthAddr:           addr,
		FromEthAddr:       fromAddr,
		ToEthAddr:         toAddr,
		Bjj:               bjj,
		FromBjj:           fromBjj,
		ToBjj:             toBjj,
		TokenID:           tokenID,
		Idx:               queryAccount.AccountIndex,
		FromIdx:           fromQueryAccount.AccountIndex,
		ToIdx:             toQueryAccount.AccountIndex,
		BatchNum:          historyTxsFilters.BatchNum,
		TxType:            txType,
		IncludePendingL1s: historyTxsFilters.IncludePendingTxs,
		FromItem:          historyTxsFilters.FromItem,
		Limit:             historyTxsFilters.Limit,
		Order:             *historyTxsFilters.Order,
	}, nil
}
