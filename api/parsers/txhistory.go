package parsers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type historyTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

func ParseHistoryTxFilter(c *gin.Context) (common.TxID, error) {
	var historyTxFilter historyTxFilter
	if err := c.ShouldBindUri(&historyTxFilter); err != nil {
		return common.TxID{}, err
	}
	txID, err := common.NewTxIDFromString(historyTxFilter.TxID)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid %s", err))
	}
	return txID, nil
}

type historyTxsFilters struct {
	TokenID             *uint  `form:"tokenId"`
	HezEthereumAddr     string `form:"hezEthereumAddress"`
	FromHezEthereumAddr string `form:"fromHezEthereumAddress"`
	ToHezEthereumAddr   string `form:"toHezEthereumAddress"`
	Bjj                 string `form:"BJJ"`
	ToBjj               string `form:"toBJJ"`
	FromBjj             string `form:"toBJJ"`
	AccountIndex        string `form:"accountIndex"`
	FromAccountIndex    string `form:"fromAccountIndex"`
	ToAccountIndex      string `form:"toAccountIndex"`
	BatchNum            *uint  `form:"batchNum"`
	TxType              string `form:"type"`
	IncludePendingTxs   *bool  `form:"includePendingL1s"`

	Pagination
}

func ParseHistoryTxsFilters(c *gin.Context) (historydb.GetTxsAPIRequest, error) {
	var historyTxsFilters historyTxsFilters
	if err := c.ShouldBindQuery(&historyTxsFilters); err != nil {
		return historydb.GetTxsAPIRequest{}, err
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

	isAddrNotNil := addr != nil || toAddr != nil || fromAddr != nil
	isBjjNotNil := bjj != nil || toBjj != nil || fromBjj != nil

	if isAddrNotNil && isBjjNotNil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	// Idx
	idx, err := common.StringToIdx(historyTxsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromIdx, err := common.StringToIdx(historyTxsFilters.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toIdx, err := common.StringToIdx(historyTxsFilters.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// TODO: move to this https://github.com/go-playground/validator/releases/tag/v8.7
	isIdxNotNil := fromIdx != nil || toIdx != nil || idx != nil

	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || tokenID != nil) {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
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
		Idx:               idx,
		FromIdx:           fromIdx,
		ToIdx:             toIdx,
		BatchNum:          historyTxsFilters.BatchNum,
		TxType:            txType,
		IncludePendingL1s: historyTxsFilters.IncludePendingTxs,
		FromItem:          historyTxsFilters.FromItem,
		Limit:             historyTxsFilters.Limit,
		Order:             *historyTxsFilters.Order,
	}, nil
}
