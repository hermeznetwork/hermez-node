package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func getHistoryTxs(c *gin.Context) {
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
	txs, pagination, err := h.GetHistoryTxs(
		addr, bjj, tokenID, idx, batchNum, txType, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type txsResponse struct {
		Txs        []historydb.TxAPI `json:"transactions"`
		Pagination *db.Pagination    `json:"pagination"`
	}
	c.JSON(http.StatusOK, &txsResponse{
		Txs:        txs,
		Pagination: pagination,
	})
}

func getHistoryTx(c *gin.Context) {
	// Get TxID
	txID, err := parseParamTxID(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from historyDB
	tx, err := h.GetHistoryTx(txID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build succesfull response
	c.JSON(http.StatusOK, tx)
}
