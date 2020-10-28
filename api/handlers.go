package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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

var (
	// ErrNillBidderAddr is used when a nil bidderAddr is received in the getCoordinator method
	ErrNillBidderAddr = errors.New("biderAddr can not be nil")
)

func postAccountCreationAuth(c *gin.Context) {
	// Parse body
	var apiAuth accountCreationAuthAPI
	if err := c.ShouldBindJSON(&apiAuth); err != nil {
		retBadReq(err, c)
		return
	}
	// API to common + verify signature
	dbAuth, err := accountCreationAuthAPIToCommon(&apiAuth)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Insert to DB
	if err := l2.AddAccountCreationAuth(dbAuth); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return OK
	c.Status(http.StatusOK)
}

func getAccountCreationAuth(c *gin.Context) {
	// Get hezEthereumAddress
	addr, err := parseParamHezEthAddr(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch auth from l2DB
	dbAuth, err := l2.GetAccountCreationAuth(*addr)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	apiAuth := accountCreationAuthToAPI(dbAuth)
	// Build succesfull response
	c.JSON(http.StatusOK, apiAuth)
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

func getSlots(c *gin.Context) {

}

func getBids(c *gin.Context) {

}

func getNextForgers(c *gin.Context) {

}

func getState(c *gin.Context) {

}

func getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, cg)
}

func getRecommendedFee(c *gin.Context) {

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
