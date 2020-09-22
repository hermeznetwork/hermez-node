package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// maxLimit is the max permited items to be returned in paginated responses
const maxLimit uint = 2049

// dfltLast indicates how paginated endpoints use the query param last if not provided
const dfltLast = false

// dfltLimit indicates the limit of returned items in paginated responses if the query param limit is not provided
const dfltLimit uint = 20

// 2^32 -1
const maxUint32 = 4294967295

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

}

func getExit(c *gin.Context) {

}

func getHistoryTxs(c *gin.Context) {
	// Get query parameters
	// TokenID
	tokenID, err := parseQueryUint("tokenId", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Hez Eth addr
	addr, err := parseQueryHezEthAddr(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// BJJ
	bjj, err := parseQueryBJJ(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if addr != nil && bjj != nil {
		retBadReq(errors.New("bjj and hermezEthereumAddress params are incompatible"), c)
		return
	}
	// Idx
	idx, err := parseIdx(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if idx != nil && (addr != nil || bjj != nil || tokenID != nil) {
		retBadReq(errors.New("accountIndex is incompatible with BJJ, hermezEthereumAddress and tokenId"), c)
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
	offset, last, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Fetch txs from historyDB
	txs, totalItems, err := h.GetHistoryTxs(
		addr, bjj, tokenID, idx, batchNum, txType, offset, limit, *last,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	apiTxs := historyTxsToAPI(txs)
	lastRet := int(*offset) + len(apiTxs) - 1
	if *last {
		lastRet = totalItems - 1
	}
	c.JSON(http.StatusOK, &historyTxsAPI{
		Txs: apiTxs,
		Pagination: pagination{
			TotalItems:       totalItems,
			LastReturnedItem: lastRet,
		},
	})
}

func getHistoryTx(c *gin.Context) {

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
