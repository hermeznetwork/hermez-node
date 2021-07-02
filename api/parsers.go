package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// Query parsers

type querier interface {
	Query(string) string
}

// nolint reason: res may be not overwritten
func parseQueryUint(name string, dflt *uint, min, max uint, c querier) (*uint, error) { //nolint:SA4009
	str := c.Query(name)
	return stringToUint(str, name, dflt, min, max)
}

// nolint reason: res may be not overwritten
func parseQueryInt64(name string, dflt *int64, min, max int64, c querier) (*int64, error) { //nolint:SA4009
	str := c.Query(name)
	return stringToInt64(str, name, dflt, min, max)
}

// nolint reason: res may be not overwritten
func parseQueryBool(name string, dflt *bool, c querier) (*bool, error) { //nolint:SA4009
	str := c.Query(name)
	return stringToBool(name, str, dflt)
}

func stringToBool(name, str string, dflt *bool) (*bool, error) {
	if str == "" {
		return dflt, nil
	}
	if str == "true" {
		res := new(bool)
		*res = true
		return res, nil
	}
	if str == "false" {
		res := new(bool)
		*res = false
		return res, nil
	}
	return nil, tracerr.Wrap(fmt.Errorf("invalid %s. Must be either true or false", name))
}

func parseQueryHezEthAddr(c querier) (*ethCommon.Address, error) {
	const name = "hezEthereumAddress"
	addrStr := c.Query(name)
	return hezStringToEthAddr(addrStr, name)
}

func parseQueryFromHezEthAddr(c querier) (*ethCommon.Address, error) {
	const name = "fromHezEthereumAddress"
	addrStr := c.Query(name)
	return hezStringToEthAddr(addrStr, name)
}

func parseQueryToHezEth(c querier) (*ethCommon.Address, error) {
	const name = "toHezEthereumAddress"
	addrStr := c.Query(name)
	return hezStringToEthAddr(addrStr, name)
}

func parseQueryBJJ(c querier) (*babyjub.PublicKeyComp, error) {
	const name = "BJJ"
	bjjStr := c.Query(name)
	return hezStringToBJJ(bjjStr, name)
}

func parseQueryFromBJJ(c querier) (*babyjub.PublicKeyComp, error) {
	const name = "fromBJJ"
	bjjStr := c.Query(name)
	return hezStringToBJJ(bjjStr, name)
}

func parseQueryToBJJ(c querier) (*babyjub.PublicKeyComp, error) {
	const name = "toBJJ"
	bjjStr := c.Query(name)
	return hezStringToBJJ(bjjStr, name)
}

func parseQueryPoolL2TxState(c querier) (*common.PoolL2TxState, error) {
	const name = "state"
	stateStr := c.Query(name)
	if stateStr == "" {
		return nil, nil
	}
	switch common.PoolL2TxState(stateStr) {
	case common.PoolL2TxStatePending:
		ret := common.PoolL2TxStatePending
		return &ret, nil
	case common.PoolL2TxStateForged:
		ret := common.PoolL2TxStateForged
		return &ret, nil
	case common.PoolL2TxStateForging:
		ret := common.PoolL2TxStateForging
		return &ret, nil
	case common.PoolL2TxStateInvalid:
		ret := common.PoolL2TxStateInvalid
		return &ret, nil
	}
	return nil, tracerr.Wrap(fmt.Errorf(
		"invalid %s, %s is not a valid option. Check the valid options in the documentation",
		name, stateStr,
	))
}

func parseQueryTxType(c querier) (*common.TxType, error) {
	const name = "type"
	typeStr := c.Query(name)
	if typeStr == "" {
		return nil, nil
	}
	switch common.TxType(typeStr) {
	case common.TxTypeExit:
		ret := common.TxTypeExit
		return &ret, nil
	case common.TxTypeTransfer:
		ret := common.TxTypeTransfer
		return &ret, nil
	case common.TxTypeDeposit:
		ret := common.TxTypeDeposit
		return &ret, nil
	case common.TxTypeCreateAccountDeposit:
		ret := common.TxTypeCreateAccountDeposit
		return &ret, nil
	case common.TxTypeCreateAccountDepositTransfer:
		ret := common.TxTypeCreateAccountDepositTransfer
		return &ret, nil
	case common.TxTypeDepositTransfer:
		ret := common.TxTypeDepositTransfer
		return &ret, nil
	case common.TxTypeForceTransfer:
		ret := common.TxTypeForceTransfer
		return &ret, nil
	case common.TxTypeForceExit:
		ret := common.TxTypeForceExit
		return &ret, nil
	case common.TxTypeTransferToEthAddr:
		ret := common.TxTypeTransferToEthAddr
		return &ret, nil
	case common.TxTypeTransferToBJJ:
		ret := common.TxTypeTransferToBJJ
		return &ret, nil
	}
	return nil, tracerr.Wrap(fmt.Errorf(
		"invalid %s, %s is not a valid option. Check the valid options in the documentation",
		name, typeStr,
	))
}

func stringToTxType(txType string) (*common.TxType, error) {
	if txType == "" {
		return nil, nil
	}
	txTypeCasted := common.TxType(txType)
	switch txTypeCasted {
	case common.TxTypeExit, common.TxTypeTransfer, common.TxTypeDeposit, common.TxTypeCreateAccountDeposit,
		common.TxTypeCreateAccountDepositTransfer, common.TxTypeDepositTransfer, common.TxTypeForceTransfer,
		common.TxTypeForceExit, common.TxTypeTransferToEthAddr, common.TxTypeTransferToBJJ:
		return &txTypeCasted, nil
	default:
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, %s is not a valid option. Check the valid options in the documentation",
			"type", txType,
		))
	}
}

func stringToL2TxState(txState string) (*common.PoolL2TxState, error) {
	if txState == "" {
		return nil, nil
	}
	txStateCasted := common.PoolL2TxState(txState)
	switch txStateCasted {
	case common.PoolL2TxStatePending, common.PoolL2TxStateForged, common.PoolL2TxStateForging, common.PoolL2TxStateInvalid:
		return &txStateCasted, nil
	default:
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, %s is not a valid option. Check the valid options in the documentation",
			"state", txState,
		))
	}
}
func parseIdx(c querier) (*common.Idx, error) {
	const name = "accountIndex"
	idxStr := c.Query(name)
	return stringToIdx(idxStr, name)
}

func parseFromIdx(c querier) (*common.Idx, error) {
	const name = "fromAccountIndex"
	idxStr := c.Query(name)
	return stringToIdx(idxStr, name)
}

func parseToIdx(c querier) (*common.Idx, error) {
	const name = "toAccountIndex"
	idxStr := c.Query(name)
	return stringToIdx(idxStr, name)
}

type exitFilter struct {
	BatchNum     uint   `uri:"batchNum" binding:"required"`
	AccountIndex string `uri:"accountIndex" binding:"required"`
}

func parseExitFilter(c *gin.Context) (*uint, *common.Idx, error) {
	var exitFilter exitFilter
	if err := c.ShouldBindUri(&exitFilter); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	idx, err := stringToIdx(exitFilter.AccountIndex, "accountIndex")
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	return &exitFilter.BatchNum, idx, nil
}

type exitsFilters struct {
	TokenID              string `form:"tokenId"`
	Addr                 string `form:"hezEthereumAddress"`
	Bjj                  string `form:"BJJ"`
	AccountIndex         string `form:"accountIndex"`
	BatchNum             string `form:"batchNum"`
	OnlyPendingWithdraws string `form:"onlyPendingWithdraws"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseExitsFilters(c *gin.Context) (historydb.GetExitsAPIRequest, error) {
	var exitsFilters exitsFilters
	if err := c.BindQuery(&exitsFilters); err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	// Token ID
	tid, err := stringToUint(exitsFilters.TokenID, "tokenId", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	var tokenID *common.TokenID
	if tid != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*tid)
	}

	addr, err := hezStringToEthAddr(exitsFilters.Addr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := hezStringToBJJ(exitsFilters.Bjj, "BJJ")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	if addr != nil && bjj != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	idx, err := stringToIdx(exitsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	if idx != nil && (addr != nil || bjj != nil || tokenID != nil) {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}

	batchNum, err := stringToUint(exitsFilters.BatchNum, "batchNum", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	onlyPendingWithdraws, err := stringToBool("onlyPendingWithdraws", exitsFilters.OnlyPendingWithdraws, nil)
	if err != nil {
		return historydb.GetExitsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetExitsAPIRequest{
		EthAddr:              addr,
		Bjj:                  bjj,
		TokenID:              tokenID,
		Idx:                  idx,
		BatchNum:             batchNum,
		OnlyPendingWithdraws: onlyPendingWithdraws,
		FromItem:             exitsFilters.FromItem,
		Limit:                exitsFilters.Limit,
		Order:                *exitsFilters.Order,
	}, nil
}

type txsFilters struct {
	tokenID  *common.TokenID
	addr     *ethCommon.Address
	fromAddr *ethCommon.Address
	toAddr   *ethCommon.Address
	bjj      *babyjub.PublicKeyComp
	fromBjj  *babyjub.PublicKeyComp
	toBjj    *babyjub.PublicKeyComp
	idx      *common.Idx
	fromIdx  *common.Idx
	toIdx    *common.Idx
}

type poolTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

func parsePoolTxFilter(c *gin.Context) (common.TxID, error) {
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
	TokenId                string `form:"tokenId"`
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

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parsePoolTxsFilters(c *gin.Context) (l2db.GetPoolTxsAPIRequest, error) {
	var poolTxsFilter poolTxsFilter
	if err := c.BindQuery(&poolTxsFilter); err != nil {
		return l2db.GetPoolTxsAPIRequest{}, err
	}
	// TokenID
	tid, err := stringToUint(poolTxsFilter.TokenId, "tokenId", nil, 0, maxUint32)
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}
	var tokenID *common.TokenID
	if tid != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*tid)
	}

	addr, err := hezStringToEthAddr(poolTxsFilter.HezEthereumAddress, "hezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromAddr, err := hezStringToEthAddr(poolTxsFilter.FromHezEthereumAddress, "fromHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toAddr, err := hezStringToEthAddr(poolTxsFilter.ToHezEthereumAddress, "toHezEthereumAddress")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := hezStringToBJJ(poolTxsFilter.Bjj, "BJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromBjj, err := hezStringToBJJ(poolTxsFilter.FromBJJ, "fromBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toBjj, err := hezStringToBJJ(poolTxsFilter.ToBJJ, "toBJJ")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	isAddrNotNil := addr != nil || toAddr != nil || fromAddr != nil
	isBjjNotNil := bjj != nil || toBjj != nil || fromBjj != nil

	if isAddrNotNil && isBjjNotNil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	// Idx
	idx, err := stringToIdx(poolTxsFilter.AccountIndex, "accountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromIdx, err := stringToIdx(poolTxsFilter.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toIdx, err := stringToIdx(poolTxsFilter.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// TODO: move to this https://github.com/go-playground/validator/releases/tag/v8.7
	isIdxNotNil := fromIdx != nil || toIdx != nil || idx != nil

	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || tokenID != nil) {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}

	txType, err := stringToTxType(poolTxsFilter.TxType)
	if err != nil {
		return l2db.GetPoolTxsAPIRequest{}, tracerr.Wrap(err)
	}

	txState, err := stringToL2TxState(poolTxsFilter.State)
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

type historyTxFilter struct {
	TxID string `uri:"id" binding:"required"`
}

func parseTxIDParam(c *gin.Context) (common.TxID, error) {
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
	TokenID             string `form:"tokenId"`
	HezEthereumAddr     string `form:"hezEthereumAddress"`
	FromHezEthereumAddr string `form:"fromHezEthereumAddress"`
	ToHezEthereumAddr   string `form:"toHezEthereumAddress"`
	Bjj                 string `form:"BJJ"`
	ToBjj               string `form:"toBJJ"`
	FromBjj             string `form:"toBJJ"`
	AccountIndex        string `form:"accountIndex"`
	FromAccountIndex    string `form:"fromAccountIndex"`
	ToAccountIndex      string `form:"toAccountIndex"`
	BatchNum            string `form:"batchNum"`
	TxType              string `form:"type"`
	IncludePendingTxs   string `form:"includePendingL1s"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseHistoryTxsFilters(c *gin.Context) (historydb.GetTxsAPIRequest, error) {
	var historyTxsFilters historyTxsFilters
	if err := c.BindQuery(&historyTxsFilters); err != nil {
		return historydb.GetTxsAPIRequest{}, err
	}

	// TokenID
	tid, err := stringToUint(historyTxsFilters.TokenID, "tokenId", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}
	var tokenID *common.TokenID
	if tid != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*tid)
	}

	addr, err := hezStringToEthAddr(historyTxsFilters.HezEthereumAddr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromAddr, err := hezStringToEthAddr(historyTxsFilters.FromHezEthereumAddr, "fromHezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toAddr, err := hezStringToEthAddr(historyTxsFilters.ToHezEthereumAddr, "toHezEthereumAddress")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	bjj, err := hezStringToBJJ(historyTxsFilters.Bjj, "BJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromBjj, err := hezStringToBJJ(historyTxsFilters.FromBjj, "fromBJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toBjj, err := hezStringToBJJ(historyTxsFilters.ToBjj, "toBJJ")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	isAddrNotNil := addr != nil || toAddr != nil || fromAddr != nil
	isBjjNotNil := bjj != nil || toBjj != nil || fromBjj != nil

	if isAddrNotNil && isBjjNotNil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	// Idx
	idx, err := stringToIdx(historyTxsFilters.AccountIndex, "accountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	fromIdx, err := stringToIdx(historyTxsFilters.FromAccountIndex, "fromAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	toIdx, err := stringToIdx(historyTxsFilters.ToAccountIndex, "toAccountIndex")
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	// TODO: move to this https://github.com/go-playground/validator/releases/tag/v8.7
	isIdxNotNil := fromIdx != nil || toIdx != nil || idx != nil

	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || tokenID != nil) {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}

	txType, err := stringToTxType(historyTxsFilters.TxType)
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	batchNum, err := stringToUint(historyTxsFilters.BatchNum, "batchNum", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetTxsAPIRequest{}, tracerr.Wrap(err)
	}

	includePendingL1s := new(bool)
	*includePendingL1s = false
	includePendingL1s, err = stringToBool("includePendingL1s", historyTxsFilters.IncludePendingTxs, includePendingL1s)
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
		BatchNum:          batchNum,
		TxType:            txType,
		IncludePendingL1s: includePendingL1s,
		FromItem:          historyTxsFilters.FromItem,
		Limit:             historyTxsFilters.Limit,
		Order:             *historyTxsFilters.Order,
	}, nil
}

type batchFilter struct {
	BatchNum uint `uri:"batchNum" binding:"required"`
}

func parseBatchFilter(c *gin.Context) (*uint, error) {
	var batchFilter batchFilter
	if err := c.ShouldBindUri(&batchFilter); err != nil {
		return nil, err
	}
	return &batchFilter.BatchNum, nil
}

type batchesFilters struct {
	MinBatchNum string `form:"minBatchNum"`
	MaxBatchNum string `form:"maxBatchNum"`
	SlotNum     string `form:"slotNum"`
	ForgerAddr  string `form:"forgerAddr"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseBatchesFilter(c *gin.Context) (historydb.GetBatchesAPIRequest, error) {
	var batchesFilters batchesFilters
	if err := c.BindQuery(&batchesFilters); err != nil {
		return historydb.GetBatchesAPIRequest{}, err
	}

	minBatchNum, err := stringToUint(batchesFilters.MinBatchNum, "minBatchNum", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetBatchesAPIRequest{}, tracerr.Wrap(err)
	}

	maxBatchNum, err := stringToUint(batchesFilters.MaxBatchNum, "maxBatchNum", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetBatchesAPIRequest{}, tracerr.Wrap(err)
	}

	slotNum, err := stringToUint(batchesFilters.SlotNum, "slotNum", nil, 0, maxUint32)
	if err != nil {
		return historydb.GetBatchesAPIRequest{}, tracerr.Wrap(err)
	}

	addr, err := parseEthAddr(batchesFilters.ForgerAddr)
	if err != nil {
		return historydb.GetBatchesAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetBatchesAPIRequest{
		MinBatchNum: minBatchNum,
		MaxBatchNum: maxBatchNum,
		SlotNum:     slotNum,
		ForgerAddr:  addr,
		FromItem:    batchesFilters.FromItem,
		Limit:       batchesFilters.Limit,
		Order:       *batchesFilters.Order,
	}, nil
}

func parseTxsFilters(c querier) (txsFilters, error) {
	// TokenID
	tid, err := parseQueryUint("tokenId", nil, 0, maxUint32, c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	var tokenID *common.TokenID
	if tid != nil {
		tokenID = new(common.TokenID)
		*tokenID = common.TokenID(*tid)
	}
	// Hez Eth addr
	addr, err := parseQueryHezEthAddr(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	fromAddr, err := parseQueryFromHezEthAddr(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	toAddr, err := parseQueryToHezEth(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	// BJJ
	bjj, err := parseQueryBJJ(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	fromBjj, err := parseQueryFromBJJ(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	toBjj, err := parseQueryToBJJ(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	isAddrNotNil := addr != nil || toAddr != nil || fromAddr != nil
	isBjjNotNil := bjj != nil || toBjj != nil || fromBjj != nil

	if isAddrNotNil && isBjjNotNil {
		return txsFilters{}, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}
	// Idx
	idx, err := parseIdx(c)
	if err != nil {
		return txsFilters{}, err
	}
	// from Idx
	fromIdx, err := parseFromIdx(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	// to Idx
	toIdx, err := parseToIdx(c)
	if err != nil {
		return txsFilters{}, tracerr.Wrap(err)
	}
	if (fromIdx != nil || toIdx != nil) && idx != nil {
		return txsFilters{}, tracerr.Wrap(errors.New("accountIndex is incompatible with fromAccountIndex or toAccountIndex"))
	}
	isIdxNotNil := fromIdx != nil || toIdx != nil || idx != nil

	if isIdxNotNil &&
		(isAddrNotNil || isBjjNotNil || tokenID != nil) {
		return txsFilters{}, tracerr.Wrap(errors.New("accountIndex is incompatible with BJJ, hezEthereumAddress and tokenId"))
	}
	return txsFilters{
		tokenID:  tokenID,
		addr:     addr,
		fromAddr: fromAddr,
		toAddr:   toAddr,
		bjj:      bjj,
		fromBjj:  fromBjj,
		toBjj:    toBjj,
		idx:      idx,
		fromIdx:  fromIdx,
		toIdx:    toIdx,
	}, err
}

type tokenFilter struct {
	ID *uint `uri:"id" binding:"required"`
}

func parseTokenFilter(c *gin.Context) (*uint, error) {
	var tokenFilter tokenFilter
	if err := c.ShouldBindUri(&tokenFilter); err != nil {
		return nil, err
	}
	return tokenFilter.ID, nil
}

type tokensFilters struct {
	IDs     string `form:"ids"`
	Symbols string `form:"symbols"`
	Name    string `form:"name"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseTokensFilters(c *gin.Context) (historydb.GetTokensAPIRequest, error) {
	var tokensFilters tokensFilters
	if err := c.BindQuery(&tokensFilters); err != nil {
		return historydb.GetTokensAPIRequest{}, err
	}
	var tokensIDs []common.TokenID
	if tokensFilters.IDs != "" {
		ids := strings.Split(tokensFilters.IDs, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return historydb.GetTokensAPIRequest{}, tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokensIDs = append(tokensIDs, tokenID)
		}
	}

	var symbols []string
	if tokensFilters.Symbols != "" {
		symbols = strings.Split(tokensFilters.Symbols, ",")
	}

	return historydb.GetTokensAPIRequest{
		Ids:      tokensIDs,
		Symbols:  symbols,
		Name:     tokensFilters.Name,
		FromItem: tokensFilters.FromItem,
		Limit:    tokensFilters.Limit,
		Order:    *tokensFilters.Order,
	}, nil
}

func parseTokenFilters(c querier) ([]common.TokenID, []string, string, error) {
	idsStr := c.Query("ids")
	symbolsStr := c.Query("symbols")
	nameStr := c.Query("name")
	var tokensIDs []common.TokenID
	if idsStr != "" {
		ids := strings.Split(idsStr, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return nil, nil, "", tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokensIDs = append(tokensIDs, tokenID)
		}
	}
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, ",")
	}
	return tokensIDs, symbols, nameStr, nil
}

type currencyFilter struct {
	Symbol string `uri:"symbol" binding:"required"`
}

func parseCurrencyFilter(c *gin.Context) (string, error) {
	var currencyFilter currencyFilter
	if err := c.ShouldBindUri(&currencyFilter); err != nil {
		return "", err
	}
	return currencyFilter.Symbol, nil
}

type currenciesFilters struct {
	Symbols string `form:"symbols"`
}

func parseCurrenciesFilters(c *gin.Context) ([]string, error) {
	var currenciesFilters currenciesFilters
	var symbols []string
	if err := c.BindQuery(&currenciesFilters); err != nil {
		return symbols, err
	}
	if currenciesFilters.Symbols != "" {
		symbols = strings.Split(currenciesFilters.Symbols, "|")
	}
	return symbols, nil
}

func parseCurrencyFilters(c querier) ([]string, error) {
	symbolsStr := c.Query("symbols")
	var symbols []string
	if symbolsStr != "" {
		symbols = strings.Split(symbolsStr, "|")
	}
	return symbols, nil
}

type bidsFilters struct {
	SlotNum    string `form:"slotNum"`
	BidderAddr string `form:"bidderAddr"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseBidsFilters(c *gin.Context) (historydb.GetBidsAPIRequest, error) {
	var bidsFilters bidsFilters
	if err := c.BindQuery(&bidsFilters); err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}
	slotNum, err := stringToInt64(bidsFilters.SlotNum, "slotNum", nil, 0, maxInt64)
	if err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}
	bidderAddress, err := parseEthAddr(bidsFilters.BidderAddr)
	if err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}

	if slotNum == nil && bidderAddress == nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(errors.New("It is necessary to add at least one filter: slotNum or/and bidderAddr"))
	}

	return historydb.GetBidsAPIRequest{
		SlotNum:    slotNum,
		BidderAddr: bidderAddress,
		FromItem:   bidsFilters.FromItem,
		Order:      *bidsFilters.Order,
		Limit:      bidsFilters.Limit,
	}, nil
}

func parseBidFilters(c querier) (*int64, *ethCommon.Address, error) {
	slotNum, err := parseQueryInt64("slotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	bidderAddr, err := parseQueryEthAddr("bidderAddr", c)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	return slotNum, bidderAddr, nil
}

type slotFilter struct {
	SlotNum *uint `uri:"slotNum" binding:"required"`
}

func parseSlotFilter(c *gin.Context) (*uint, error) {
	var slotFilter slotFilter
	if err := c.ShouldBindUri(&slotFilter); err != nil {
		return nil, err
	}
	return slotFilter.SlotNum, nil
}

type slotsFilters struct {
	MinSlotNum           string `form:"minSlotNum"`
	MaxSlotNum           string `form:"maxSlotNum"`
	WonByEthereumAddress string `form:"wonByEthereumAddress"`
	FinishedAuction      string `form:"finishedAuction"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseSlotsFilters(c *gin.Context) (historydb.GetBestBidsAPIRequest, error) {
	var slotsFilters slotsFilters
	if err := c.BindQuery(&slotsFilters); err != nil {
		return historydb.GetBestBidsAPIRequest{}, err
	}

	minSlotNum, err := stringToInt64(slotsFilters.MinSlotNum, "minSlotNum", nil, 0, maxInt64)
	if err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
	}

	maxSlotNum, err := stringToInt64(slotsFilters.MaxSlotNum, "maxSlotNum", nil, 0, maxInt64)
	if err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
	}

	wonByEthereumAddress, err := parseEthAddr(slotsFilters.WonByEthereumAddress)
	if err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
	}

	finishedAuction, err := stringToBool("finishedAuction", slotsFilters.FinishedAuction, nil)
	if err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetBestBidsAPIRequest{
		MinSlotNum:      minSlotNum,
		MaxSlotNum:      maxSlotNum,
		BidderAddr:      wonByEthereumAddress,
		FinishedAuction: finishedAuction,
		FromItem:        slotsFilters.FromItem,
		Order:           *slotsFilters.Order,
		Limit:           slotsFilters.Limit,
	}, nil
}
func parseSlotFilters(c querier) (*int64, *int64, *ethCommon.Address, *bool, error) {
	minSlotNum, err := parseQueryInt64("minSlotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	maxSlotNum, err := parseQueryInt64("maxSlotNum", nil, 0, maxInt64, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	wonByEthereumAddress, err := parseQueryEthAddr("wonByEthereumAddress", c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	finishedAuction, err := parseQueryBool("finishedAuction", nil, c)
	if err != nil {
		return nil, nil, nil, nil, tracerr.Wrap(err)
	}
	return minSlotNum, maxSlotNum, wonByEthereumAddress, finishedAuction, nil
}

type coordinatorsFilters struct {
	BidderAddr string `form:"bidderAddr"`
	ForgerAddr string `form:"forgerAddr"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseCoordinatorsFilters(c *gin.Context) (historydb.GetCoordinatorsAPIRequest, error) {
	var coordinatorsFilters coordinatorsFilters
	if err := c.BindQuery(&coordinatorsFilters); err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}
	bidderAddr, err := parseEthAddr(coordinatorsFilters.BidderAddr)
	if err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}
	forgerAddr, err := parseEthAddr(coordinatorsFilters.ForgerAddr)
	if err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetCoordinatorsAPIRequest{
		BidderAddr: bidderAddr,
		ForgerAddr: forgerAddr,
		FromItem:   coordinatorsFilters.FromItem,
		Limit:      coordinatorsFilters.Limit,
		Order:      *coordinatorsFilters.Order,
	}, nil
}

type accountFilter struct {
	AccountIndex string `uri:"accountIndex" binding:"required"`
}

func parseAccountFilter(c *gin.Context) (*common.Idx, error) {
	var accountFilter accountFilter
	if err := c.ShouldBindUri(&accountFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return stringToIdx(accountFilter.AccountIndex, "accountIndex")
}

type accountsFilter struct {
	IDs  string `form:"tokenIds"`
	Addr string `form:"hezEthereumAddress"`
	Bjj  string `form:"BJJ"`

	FromItem *uint   `form:"fromItem"`
	Order    *string `form:"order,default=ASC" binding:"omitempty,oneof=ASC DESC"`
	Limit    *uint   `form:"limit,default=20" binding:"omitempty,min=1,max=2049"`
}

func parseAccountsFilters(c *gin.Context) (historydb.GetAccountsAPIRequest, error) {
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

	addr, err := hezStringToEthAddr(accountsFilter.Addr, "hezEthereumAddress")
	if err != nil {
		return historydb.GetAccountsAPIRequest{}, err
	}

	bjj, err := hezStringToBJJ(accountsFilter.Bjj, "BJJ")
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

func parseAccountFilters(c querier) ([]common.TokenID, *ethCommon.Address, *babyjub.PublicKeyComp, error) {
	// TokenID
	idsStr := c.Query("tokenIds")
	var tokenIDs []common.TokenID
	if idsStr != "" {
		ids := strings.Split(idsStr, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return nil, nil, nil, tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokenIDs = append(tokenIDs, tokenID)
		}
	}
	// Hez Eth addr
	addr, err := parseQueryHezEthAddr(c)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	// BJJ
	bjj, err := parseQueryBJJ(c)
	if err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if addr != nil && bjj != nil {
		return nil, nil, nil, tracerr.Wrap(errors.New("bjj and hezEthereumAddress params are incompatible"))
	}

	return tokenIDs, addr, bjj, nil
}

// Param parsers

type paramer interface {
	Param(string) string
}

func parseParamTxID(c paramer) (common.TxID, error) {
	const name = "id"
	txIDStr := c.Param(name)
	if txIDStr == "" {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("%s is required", name))
	}
	txID, err := common.NewTxIDFromString(txIDStr)
	if err != nil {
		return common.TxID{}, tracerr.Wrap(fmt.Errorf("invalid %s", name))
	}
	return txID, nil
}

func parseParamIdx(c paramer) (*common.Idx, error) {
	const name = "accountIndex"
	idxStr := c.Param(name)
	return stringToIdx(idxStr, name)
}

// nolint reason: res may be not overwritten
func parseParamUint(name string, dflt *uint, min, max uint, c paramer) (*uint, error) { //nolint:SA4009
	str := c.Param(name)
	return stringToUint(str, name, dflt, min, max)
}

// nolint reason: res may be not overwritten
func parseParamInt64(name string, dflt *int64, min, max int64, c paramer) (*int64, error) { //nolint:SA4009
	str := c.Param(name)
	return stringToInt64(str, name, dflt, min, max)
}

func stringToIdx(idxStr, name string) (*common.Idx, error) {
	if idxStr == "" {
		return nil, nil
	}
	splitted := strings.Split(idxStr, ":")
	const expectedLen = 3
	if len(splitted) != expectedLen || splitted[0] != "hez" {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, must follow this: hez:<tokenSymbol>:index", name))
	}
	// TODO: check that the tokenSymbol match the token related to the account index
	idxInt, err := strconv.Atoi(splitted[2])
	idx := common.Idx(idxInt)
	return &idx, tracerr.Wrap(err)
}

func stringToUint(uintStr, name string, dflt *uint, min, max uint) (*uint, error) {
	if uintStr != "" {
		resInt, err := strconv.Atoi(uintStr)
		if err != nil || resInt < 0 || resInt < int(min) || resInt > int(max) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"Invalid %s. Must be an integer within the range [%d, %d]",
				name, min, max))
		}
		res := uint(resInt)
		return &res, nil
	}
	return dflt, nil
}

func stringToInt64(uintStr, name string, dflt *int64, min, max int64) (*int64, error) {
	if uintStr != "" {
		resInt, err := strconv.Atoi(uintStr)
		if err != nil || resInt < 0 || resInt < int(min) || resInt > int(max) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"Invalid %s. Must be an integer within the range [%d, %d]",
				name, min, max))
		}
		res := int64(resInt)
		return &res, nil
	}
	return dflt, nil
}

func hezStringToEthAddr(addrStr, name string) (*ethCommon.Address, error) {
	if addrStr == "" {
		return nil, nil
	}
	splitted := strings.Split(addrStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 42 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:0x[a-fA-F0-9]{40}$", name))
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(splitted[1]))
	return &addr, tracerr.Wrap(err)
}

func hezStringToBJJ(bjjStr, name string) (*babyjub.PublicKeyComp, error) {
	if bjjStr == "" {
		return nil, nil
	}
	const decodedLen = 33
	splitted := strings.Split(bjjStr, "hez:")
	if len(splitted) != 2 || len(splitted[1]) != 44 {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, must follow this regex: ^hez:[A-Za-z0-9+/=]{44}$",
			name))
	}
	decoded, err := base64.RawURLEncoding.DecodeString(splitted[1])
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf(
			"Invalid %s, error decoding base64 string: %s",
			name, err.Error()))
	}
	if len(decoded) != decodedLen {
		return nil, tracerr.Wrap(fmt.Errorf(
			"invalid %s, error decoding base64 string: unexpected byte array length",
			name))
	}
	bjjBytes := [decodedLen - 1]byte{}
	copy(bjjBytes[:decodedLen-1], decoded[:decodedLen-1])
	sum := bjjBytes[0]
	for i := 1; i < len(bjjBytes); i++ {
		sum += bjjBytes[i]
	}
	if decoded[decodedLen-1] != sum {
		return nil, tracerr.Wrap(fmt.Errorf("invalid %s, checksum failed",
			name))
	}
	bjjComp := babyjub.PublicKeyComp(bjjBytes)
	return &bjjComp, nil
}

func parseQueryEthAddr(name string, c querier) (*ethCommon.Address, error) {
	addrStr := c.Query(name)
	return parseEthAddr(addrStr)
}

func parseEthAddr(ethAddrStr string) (*ethCommon.Address, error) {
	if ethAddrStr == "" {
		return nil, nil
	}
	var addr ethCommon.Address
	err := addr.UnmarshalText([]byte(ethAddrStr))
	return &addr, tracerr.Wrap(err)
}

func parseParamHezEthAddr(c paramer) (*ethCommon.Address, error) {
	const name = "hezEthereumAddress"
	addrStr := c.Param(name)
	return hezStringToEthAddr(addrStr, name)
}

type errorMsg struct {
	Message string
}

func bjjToString(bjj babyjub.PublicKeyComp) string {
	pkComp := [32]byte(bjj)
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

func ethAddrToHez(addr ethCommon.Address) string {
	return "hez:" + addr.String()
}

func idxToHez(idx common.Idx, tokenSymbol string) string {
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}
