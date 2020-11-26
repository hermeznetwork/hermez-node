package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/ztrue/tracerr"
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

	// 2^64 /2 -1
	maxInt64 = 9223372036854775807
)

var (
	// ErrNillBidderAddr is used when a nil bidderAddr is received in the getCoordinator method
	ErrNillBidderAddr = errors.New("biderAddr can not be nil")
)

func retSQLErr(err error, c *gin.Context) {
	if tracerr.Unwrap(err) == sql.ErrNoRows {
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
