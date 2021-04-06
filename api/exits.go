package api

import (
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type GetExitsAPIRequest struct {
	EthAddr              *ethCommon.Address
	Bjj                  *babyjub.PublicKeyComp
	TokenID              *common.TokenID
	Idx                  *common.Idx
	BatchNum             *uint
	OnlyPendingWithdraws *bool

	FromItem *uint
	Limit    *uint
	Order    string
}

func (a *API) getExits(c *gin.Context) {
	// Get query parameters
	// Account filters
	tokenID, addr, bjj, idx, err := parseExitFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// BatchNum
	batchNum, err := parseQueryUint("batchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// OnlyPendingWithdraws
	onlyPendingWithdraws, err := parseQueryBool("onlyPendingWithdraws", nil, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	request := GetExitsAPIRequest{
		EthAddr:              addr,
		Bjj:                  bjj,
		TokenID:              tokenID,
		Idx:                  idx,
		BatchNum:             batchNum,
		OnlyPendingWithdraws: onlyPendingWithdraws,
		FromItem:             fromItem,
		Limit:                limit,
		Order:                order,
	}
	// Fetch exits from historyDB
	exits, pendingItems, err := a.h.GetExitsAPI(request)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type exitsResponse struct {
		Exits        []historydb.ExitAPI `json:"exits"`
		PendingItems uint64              `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &exitsResponse{
		Exits:        exits,
		PendingItems: pendingItems,
	})
}

func (a *API) getExit(c *gin.Context) {
	// Get batchNum and accountIndex
	batchNum, err := parseParamUint("batchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	idx, err := parseParamIdx(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from historyDB
	exit, err := a.h.GetExitAPI(batchNum, idx)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, exit)
}
