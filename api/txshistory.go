package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getHistoryTxs(c *gin.Context) {
	// Get query parameters
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
	// TxType
	txType, err := parseQueryTxType(c)
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

	// Fetch txs from historyDB
	txs, pendingItems, err := a.h.GetTxsAPI(
		addr, bjj, tokenID, idx, batchNum, txType, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type txsResponse struct {
		Txs          []historydb.TxAPI `json:"transactions"`
		PendingItems uint64            `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &txsResponse{
		Txs:          txs,
		PendingItems: pendingItems,
	})
}

func (a *API) getHistoryTx(c *gin.Context) {
	// Get TxID
	txID, err := parseParamTxID(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from historyDB
	tx, err := a.h.GetTxAPI(txID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build succesfull response
	c.JSON(http.StatusOK, tx)
}
