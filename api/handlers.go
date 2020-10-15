package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

const (
	// maxLimit is the max permited items to be returned in paginated responses
	maxLimit uint = 2049

	// dfltOrder indicates how paginated endpoints are ordered if not specified
	dfltOrder = historydb.OrderAsc

	// dfltLimit indicates the limit of returned items in paginated responses if the query param limit is not provided
	dfltLimit uint = 20

	// 2^32 -1
	maxUint32 = 4294967295
)

func postAccountCreationAuth(c *gin.Context) {

}

func getAccountCreationAuth(c *gin.Context) {

}

func postPoolTx(c *gin.Context) {

}

func getPoolTx(c *gin.Context) {

}

func getAccounts(c *gin.Context) {

}

func getAccount(c *gin.Context) {

}

func getExits(c *gin.Context) {
	// Get query parameters
	// Account filters
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
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Fetch exits from historyDB
	exits, pagination, err := h.GetExits(
		addr, bjj, tokenID, idx, batchNum, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	apiExits := historyExitsToAPI(exits)
	c.JSON(http.StatusOK, &exitsAPI{
		Exits:      apiExits,
		Pagination: pagination,
	})
}

func getExit(c *gin.Context) {
	// Get batchNum and accountIndex
	batchNum, err := parseParamUint("batchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	idx, err := parseParamIdx(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from historyDB
	exit, err := h.GetExit(batchNum, idx)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	apiExits := historyExitsToAPI([]historydb.HistoryExit{*exit})
	// Build succesfull response
	c.JSON(http.StatusOK, apiExits[0])
}

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
	apiTxs := historyTxsToAPI([]historydb.HistoryTx{*tx})
	// Build succesfull response
	c.JSON(http.StatusOK, apiTxs[0])
}

func getBatches(c *gin.Context) {

}

func getBatch(c *gin.Context) {

}

func getFullBatch(c *gin.Context) {

}

func getSlots(c *gin.Context) {

}

func getBids(c *gin.Context) {

}

func getNextForgers(c *gin.Context) {

}

func getState(c *gin.Context) {

}

func getConfig(c *gin.Context) {

}

func getTokens(c *gin.Context) {

}

func getToken(c *gin.Context) {
	// Get TokenID
	tokenIDUint, err := parseParamUint("id", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	tokenID := common.TokenID(*tokenIDUint)
	// Fetch token from historyDB
	token, err := h.GetToken(tokenID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, token)
}

func getRecommendedFee(c *gin.Context) {

}

func getCoordinators(c *gin.Context) {

}

func getCoordinator(c *gin.Context) {

}

func retSQLErr(err error, c *gin.Context) {
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorMsg{
			Message: err.Error(),
		})
	} else {
		c.JSON(http.StatusInternalServerError, errorMsg{
			Message: err.Error(),
		})
	}
}

func retBadReq(err error, c *gin.Context) {
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}
