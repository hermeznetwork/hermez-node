package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getAccount(c *gin.Context) {
	// Get Addr
	idx, err := parseParamIdx(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	apiAccount, err := a.h.GetAccountAPI(*idx)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	c.JSON(http.StatusOK, apiAccount)
}

func (a *API) getAccounts(c *gin.Context) {
	// Account filters
	tokenIDs, addr, bjj, err := parseAccountFilters(c)
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

	// Fetch Accounts from historyDB
	apiAccounts, pendingItems, err := a.h.GetAccountsAPI(tokenIDs, addr, bjj, fromItem, limit, order)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type accountResponse struct {
		Accounts     []historydb.AccountAPI `json:"accounts"`
		PendingItems uint64                 `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &accountResponse{
		Accounts:     apiAccounts,
		PendingItems: pendingItems,
	})
}
