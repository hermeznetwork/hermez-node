package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

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

	// Fetch exits from historyDB
	exits, pagination, err := a.h.GetExitsAPI(
		addr, bjj, tokenID, idx, batchNum, onlyPendingWithdraws, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type exitsResponse struct {
		Exits      []historydb.ExitAPI `json:"exits"`
		Pagination *db.Pagination      `json:"pagination"`
	}
	c.JSON(http.StatusOK, &exitsResponse{
		Exits:      exits,
		Pagination: pagination,
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
	// Build succesfull response
	c.JSON(http.StatusOK, exit)
}
