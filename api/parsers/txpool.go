package parsers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// PoolTxFilter struct to get uri param from /transactions-pool/:id request
type PoolTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

// ParsePoolTxFilter func for parsing tx filter to the transaction id
func ParsePoolTxFilter(c *gin.Context) (common.TxID, error) {
	var poolTxFilter PoolTxFilter
	if err := c.ShouldBindUri(&poolTxFilter); err != nil {
		return common.TxID{}, tracerr.Wrap(err)
	}
	txID, err := common.NewTxIDFromString(poolTxFilter.TxID)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid txID"))
	}
	return txID, nil
}

// PoolTxUpdateByIdxAndNonceFilter struct to get uri param from /transactions-pool/accounts/:accountIndex/nonces/:nonce request
type PoolTxUpdateByIdxAndNonceFilter struct {
	AccountIndex string `uri:"accountIndex" binding:"required"`
	Nonce        *uint  `uri:"nonce" binding:"required"`
}

// ParsePoolTxUpdateByIdxAndNonceFilter func for parsing pool tx update by idx and nonce filter to the account index and nonce
func ParsePoolTxUpdateByIdxAndNonceFilter(c *gin.Context) (common.Idx, nonce.Nonce, error) {
	var poolTxUpdateByIdxAndNonceFilter PoolTxUpdateByIdxAndNonceFilter
	if err := c.ShouldBindUri(&poolTxUpdateByIdxAndNonceFilter); err != nil {
		return common.Idx(0), 0, tracerr.Wrap(err)
	}
	queryAccount, err := common.StringToIdx(poolTxUpdateByIdxAndNonceFilter.AccountIndex, "accountIndex")
	if err != nil {
		return common.Idx(0), 0, tracerr.Wrap(err)
	}

	queryNonce := nonce.Nonce(*poolTxUpdateByIdxAndNonceFilter.Nonce)
	return *queryAccount.AccountIndex, queryNonce, nil
}

// PoolTxsFilters struct for holding query params from /transactions-pool request
type PoolTxsFilters struct {
	TokenID             *uint  `form:"tokenId"`
	HezEthereumAddr     string `form:"hezEthereumAddress"`
	FromHezEthereumAddr string `form:"fromHezEthereumAddress"`
	ToHezEthereumAddr   string `form:"toHezEthereumAddress"`
	Bjj                 string `form:"BJJ"`
	FromBjj             string `form:"fromBJJ"`
	ToBjj               string `form:"toBJJ"`
	AccountIndex        string `form:"accountIndex"`
	FromAccountIndex    string `form:"fromAccountIndex"`
	ToAccountIndex      string `form:"toAccountIndex"`
	TxType              string `form:"type"`
	State               string `form:"state"`

	Pagination
}

// PoolTxsTxsFiltersStructValidation func for pool txs query params validation
func PoolTxsTxsFiltersStructValidation(sl validator.StructLevel) {
	ef := sl.Current().Interface().(PoolTxsFilters)

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
		sl.ReportError(ef.TokenID, "tokenId", "TokenID", "onlyaccountindex", "")
	}
}

// ParsePoolTxsFilters func to parse pool txs filters from the /transactions-pool request to the GetPoolTxsAPIRequest
func ParsePoolTxsFilters(c *gin.Context, v *validator.Validate) (l2db.GetPoolTxsAPIRequest, error) {
	var poolTxsFilter PoolTxsFilters
	if err := c.BindQuery(&poolTxsFilter); err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	if err := v.Struct(poolTxsFilter); err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// TokenID
	var tokenID *common.TokenID
	if poolTxsFilter.TokenID != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*poolTxsFilter.TokenID)
	}

	addr, err := common.HezStringToEthAddr(poolTxsFilter.HezEthereumAddr, "hezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromAddr, err := common.HezStringToEthAddr(poolTxsFilter.FromHezEthereumAddr, "fromHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toAddr, err := common.HezStringToEthAddr(poolTxsFilter.ToHezEthereumAddr, "toHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := common.HezStringToBJJ(poolTxsFilter.Bjj, "BJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromBjj, err := common.HezStringToBJJ(poolTxsFilter.FromBjj, "fromBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toBjj, err := common.HezStringToBJJ(poolTxsFilter.ToBjj, "toBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// Idx
	queryAccount, err := common.StringToIdx(poolTxsFilter.AccountIndex, "accountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromQueryAccount, err := common.StringToIdx(poolTxsFilter.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toQueryAccount, err := common.StringToIdx(poolTxsFilter.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
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
		Idx:         queryAccount.AccountIndex,
		FromIdx:     fromQueryAccount.AccountIndex,
		ToIdx:       toQueryAccount.AccountIndex,
		State:       txState,

		FromItem: poolTxsFilter.FromItem,
		Limit:    poolTxsFilter.Limit,
		Order:    *poolTxsFilter.Order,
	}, nil
}
