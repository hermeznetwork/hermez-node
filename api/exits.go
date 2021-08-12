package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getExits(c *gin.Context) {
	// Get query parameters
	exitsFilters, err := parsers.ParseExitsFilters(c, a.validate)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	// Fetch exits from historyDB
	exits, pendingItems, err := a.historyDB.GetExitsAPI(exitsFilters)
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
	batchNum, idx, err := parsers.ParseExitFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch tx from historyDB
	exit, err := a.historyDB.GetExitAPI(batchNum, idx)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, exit)
}
