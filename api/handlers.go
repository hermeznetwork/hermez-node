package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/lib/pq"
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

	// Error for duplicated key
	errDuplicatedKey = "Item already exists"

	// Error for timeout due to SQL connection
	errSQLTimeout = "The node is under heavy preasure, please try again later"

	// Error message returned when context reaches timeout
	errCtxTimeout = "context deadline exceeded"
)

var (
	// ErrNillBidderAddr is used when a nil bidderAddr is received in the getCoordinator method
	ErrNillBidderAddr = errors.New("biderAddr can not be nil")
)

func retSQLErr(err error, c *gin.Context) {
	log.Warnw("HTTP API SQL request error", "err", err)
	errMsg := tracerr.Unwrap(err).Error()
	if errMsg == errCtxTimeout {
		c.JSON(http.StatusServiceUnavailable, errorMsg{
			Message: errSQLTimeout,
		})
	} else if sqlErr, ok := tracerr.Unwrap(err).(*pq.Error); ok {
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		if sqlErr.Code == "23505" {
			c.JSON(http.StatusInternalServerError, errorMsg{
				Message: errDuplicatedKey,
			})
		}
	} else if tracerr.Unwrap(err) == sql.ErrNoRows {
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
	log.Warnw("HTTP API Bad request error", "err", err)
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}
