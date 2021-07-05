package parsers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
)

type poolTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

func ParsePoolTxFilter(c *gin.Context) (common.TxID, error) {
	var poolTxFilter poolTxFilter
	if err := c.ShouldBindUri(&poolTxFilter); err != nil {
		return common.TxID{}, tracerr.Wrap(err)
	}
	txID, err := common.NewTxIDFromString(poolTxFilter.TxID)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid txID"))
	}
	return txID, nil
}

type poolTxsFilter struct {
	TokenId                *uint  `form:"tokenId"`
	HezEthereumAddress     string `form:"hezEthereumAddress"`
	FromHezEthereumAddress string `form:"fromHezEthereumAddress"`
	ToHezEthereumAddress   string `form:"toHezEthereumAddress"`
	Bjj                    string `form:"BJJ"`
	FromBJJ                string `form:"fromBJJ"`
	ToBJJ                  string `form:"toBJJ"`
	AccountIndex           string `form:"accountIndex"`
	FromAccountIndex       string `form:"fromAccountIndex"`
	ToAccountIndex         string `form:"toAccountIndex"`
	TxType                 string `form:"type"`
	State                  string `form:"state"`

	Pagination
}

func ParsePoolTxsFilters(c *gin.Context) (l2db.GetPoolTxsAPIRequest, error) {
	var poolTxsFilter poolTxsFilter
	if err := c.BindQuery(&poolTxsFilter); err != nil {
		return l2db.GetPoolTxsAPIRequest{}, err
	}
	// TokenID
	var tokenID *common.TokenID
	if poolTxsFilter.TokenId != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*poolTxsFilter.TokenId)
	}

	addr, err := common.HezStringToEthAddr(poolTxsFilter.HezEthereumAddress, "hezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromAddr, err := common.HezStringToEthAddr(poolTxsFilter.FromHezEthereumAddress, "fromHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toAddr, err := common.HezStringToEthAddr(poolTxsFilter.ToHezEthereumAddress, "toHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := common.HezStringToBJJ(poolTxsFilter.Bjj, "BJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromBjj, err := common.HezStringToBJJ(poolTxsFilter.FromBJJ, "fromBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toBjj, err := common.HezStringToBJJ(poolTxsFilter.ToBJJ, "toBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	isAddrNotNil := addr != nil || toAddr != nil || fromAddr != nil
	isBjjNotNil := bjj != nil || toBjj != nil || fromBjj != nil

	if isAddrNotNil && isBjjNotNil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	// Idx
	idx, err := common.StringToIdx(poolTxsFilter.AccountIndex, "accountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromIdx, err := common.StringToIdx(poolTxsFilter.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toIdx, err := common.StringToIdx(poolTxsFilter.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// TODO: move to this https://github.com/go-playground/validator/releases/tag/v8.7
	isIdxNotNil := fromIdx != nil || toIdx != nil || idx != nil

	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || tokenID != nil) {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}

	txType, err := common.StringToTxType(poolTxsFilter.TxType)
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	txState, err := common.StringToL2TxState(poolTxsFilter.State)
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	return l2db.GetPoolTxsAPIRequest{
		EthAddr:     addr,
		FromEthAddr: fromAddr,
		ToEthAddr:   toAddr,
		Bjj:         bjj,
		FromBjj:     fromBjj,
		ToBjj:       toBjj,
		TxType:      txType,
		TokenID:     tokenID,
		Idx:         idx,
		FromIdx:     fromIdx,
		ToIdx:       toIdx,
		State:       txState,

		FromItem: poolTxsFilter.FromItem,
		Limit:    poolTxsFilter.Limit,
		Order:    *poolTxsFilter.Order,
	}, nil
}
