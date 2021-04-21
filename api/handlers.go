package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/tracerr"
	"github.com/lib/pq"
	"github.com/russross/meddler"
)

const (
	// maxLimit is the max permitted items to be returned in paginated responses
	maxLimit uint = 2049

	// dfltOrder indicates how paginated endpoints are ordered if not specified
	dfltOrder = db.OrderAsc

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
	// ErrNilBidderAddr is used when a nil bidderAddr is received in the getCoordinator method
	ErrNilBidderAddr = errors.New("biderAddr can not be nil")
)

func retSQLErr(err error, c *gin.Context) {
	log.Warnw("HTTP API SQL request error", "err", err)
	unwrapErr := tracerr.Unwrap(err)
	metric.CollectError(unwrapErr)
	errMsg := unwrapErr.Error()
	retDupKey := func(errCode pq.ErrorCode) {
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		if errCode == "23505" {
			c.JSON(http.StatusInternalServerError, errorMsg{
				Message: errDuplicatedKey,
			})
		} else {
			c.JSON(http.StatusInternalServerError, errorMsg{
				Message: errMsg,
			})
		}
	}
	if errMsg == errCtxTimeout {
		c.JSON(http.StatusServiceUnavailable, errorMsg{
			Message: errSQLTimeout,
		})
	} else if sqlErr, ok := tracerr.Unwrap(err).(*pq.Error); ok {
		retDupKey(sqlErr.Code)
	} else if sqlErr, ok := meddler.DriverErr(tracerr.Unwrap(err)); ok {
		retDupKey(sqlErr.(*pq.Error).Code)
	} else if tracerr.Unwrap(err) == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, errorMsg{
			Message: errMsg,
		})
	} else {
		c.JSON(http.StatusInternalServerError, errorMsg{
			Message: errMsg,
		})
	}
}

func retBadReq(err error, c *gin.Context) {
	log.Warnw("HTTP API Bad request error", "err", err)
	metric.CollectError(err)
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}
