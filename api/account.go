package api

import (
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type GetAccountsAPIRequest struct {
	TokenIDs []common.TokenID
	EthAddr  *ethCommon.Address
	Bjj      *babyjub.PublicKeyComp

	FromItem *uint
	Limit    *uint
	Order    string
}

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

	request := GetAccountsAPIRequest{
		TokenIDs: tokenIDs,
		EthAddr:  addr,
		Bjj:      bjj,
		FromItem: fromItem,
		Limit:    limit,
		Order:    order,
	}
	// Fetch Accounts from historyDB
	apiAccounts, pendingItems, err := a.h.GetAccountsAPI(request)
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
