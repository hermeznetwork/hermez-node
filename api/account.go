package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/tracerr"
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
	account, err := a.s.LastGetAccount(*idx)
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
	if err := a.s.LastRead(func(sdb *statedb.Last) error {
		for x, apiAccount := range apiAccounts {
			idx, err := stringToIdx(string(apiAccount.Idx), "Account Idx")
			if err != nil {
				return tracerr.Wrap(err)
			}
			account, err := sdb.GetAccount(*idx)
			if err != nil {
				return tracerr.Wrap(err)
			}
			apiAccounts[x].Balance = apitypes.NewBigIntStr(account.Balance)
			apiAccounts[x].Nonce = account.Nonce
		}
		return nil
	}); err != nil {
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
