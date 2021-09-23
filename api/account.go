package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getAccount(c *gin.Context) {
	// Get Addr
	account, err := parsers.ParseAccountFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	apiAccount, err := a.historyDB.GetAccountAPI(*account.AccountIndex)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	//Check if symbol is correct
	if apiAccount.TokenSymbol != account.Symbol {
		retBadReq(&apiError{
			Err:  fmt.Errorf("invalid token symbol"),
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	c.JSON(http.StatusOK, apiAccount)
}

func (a *API) getAccounts(c *gin.Context) {
	for id := range c.Request.URL.Query() {
		if id != "tokenIds" && id != "hezEthereumAddress" && id != "BJJ" &&
			id != "fromItem" && id != "order" && id != "limit" {
			retBadReq(&apiError{
				Err:  fmt.Errorf("invalid Param: %s", id),
				Code: ErrParamValidationFailedCode,
				Type: ErrParamValidationFailedType,
			}, c)
			return
		}
	}

	accountsFilter, err := parsers.ParseAccountsFilters(c, a.validate)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	// Fetch Accounts from historyDB
	apiAccounts, pendingItems, err := a.historyDB.GetAccountsAPI(accountsFilter)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type accountResponse struct {
		Accounts     []historydb.AccountAPI `json:"accounts"`
		PendingItems uint64                 `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &accountResponse{
		Accounts:     apiAccounts,
		PendingItems: pendingItems,
	})
}
