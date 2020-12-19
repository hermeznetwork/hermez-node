package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/apitypes"
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

	// Get balance from stateDB
	account, err := a.s.GetAccount(*idx)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	apiAccount.Balance = apitypes.NewBigIntStr(account.Balance)
	apiAccount.Nonce = account.Nonce

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

	// Get balances from stateDB
	for x, apiAccount := range apiAccounts {
		idx, err := stringToIdx(string(apiAccount.Idx), "Account Idx")
		if err != nil {
			retSQLErr(err, c)
			return
		}
		account, err := a.s.GetAccount(*idx)
		if err != nil {
			retSQLErr(err, c)
			return
		}
		apiAccounts[x].Balance = apitypes.NewBigIntStr(account.Balance)
		apiAccounts[x].Nonce = account.Nonce
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
