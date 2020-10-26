package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHistoryTxs(c *gin.Context) {
	// Get query parameters
	tokenID, addr, bjj, idx, err := parseAccountFilters(c)
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
	apiTxs := historyTxsToAPI(txs)
	c.JSON(http.StatusOK, &historyTxsAPI{
		Txs:        apiTxs,
		Pagination: pagination,
	})
}
