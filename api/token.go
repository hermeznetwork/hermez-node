package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getToken(c *gin.Context) {
	// Get TokenID
	tokenIDUint, err := parsers.ParseTokenFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	tokenID := common.TokenID(*tokenIDUint)
	// Fetch token from historyDB
	token, err := a.historyDB.GetTokenAPI(tokenID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, token)
}

func (a *API) getTokens(c *gin.Context) {
	// Account filters
	filters, err := parsers.ParseTokensFilters(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch exits from historyDB
	tokens, pendingItems, err := a.historyDB.GetTokensAPI(filters)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type tokensResponse struct {
		Tokens       []historydb.TokenWithUSD `json:"tokens"`
		PendingItems uint64                   `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &tokensResponse{
		Tokens:       tokens,
		PendingItems: pendingItems,
	})
}
