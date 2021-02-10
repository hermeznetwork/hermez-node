package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getToken(c *gin.Context) {
	// Get TokenID
	tokenIDUint, err := parseParamUint("id", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if tokenIDUint == nil { // tokenID is required
		retBadReq(errors.New("Invalid tokenID"), c)
		return
	}
	tokenID := common.TokenID(*tokenIDUint)
	// Fetch token from historyDB
	token, err := a.h.GetTokenAPI(tokenID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, token)
}

func (a *API) getTokens(c *gin.Context) {
	// Account filters
	tokenIDs, symbols, name, err := parseTokenFilters(c)
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
	tokens, pendingItems, err := a.h.GetTokensAPI(
		tokenIDs, symbols, name, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type tokensResponse struct {
		Tokens       []historydb.TokenWithUSD `json:"tokens"`
		PendingItems uint64                   `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &tokensResponse{
		Tokens:       tokens,
		PendingItems: pendingItems,
	})
}
