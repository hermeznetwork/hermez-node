package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/tracerr"
	"github.com/lib/pq"
	"github.com/russross/meddler"
)

type errorMsg struct {
	Message string
}

func retSQLErr(err error, c *gin.Context) {
	log.Warnw("HTTP API SQL request error", "err", err)
	unwrapErr := tracerr.Unwrap(err)
	metric.CollectError(unwrapErr)
	errMsg := unwrapErr.Error()
	retDupKey := func(errCode pq.ErrorCode) {
		// https://www.postgresql.org/docs/current/errcodes-appendix.html
		if errCode == "23505" {
			c.JSON(http.StatusConflict, apiErrorResponse{
				Message: ErrDuplicatedKey,
				Code:    ErrDuplicatedKeyCode,
				Type:    ErrDuplicatedKeyType,
			})
		} else {
			c.JSON(http.StatusInternalServerError, errorMsg{
				Message: errMsg,
			})
		}
	}
	if errMsg == errCtxTimeout {
		c.JSON(http.StatusServiceUnavailable, apiErrorResponse{
			Message: ErrSQLTimeout,
			Code:    ErrSQLTimeoutCode,
			Type:    ErrSQLTimeoutType,
		})
	} else if sqlErr, ok := tracerr.Unwrap(err).(*pq.Error); ok {
		retDupKey(sqlErr.Code)
	} else if sqlErr, ok := meddler.DriverErr(tracerr.Unwrap(err)); ok {
		retDupKey(sqlErr.(*pq.Error).Code)
	} else if tracerr.Unwrap(err) == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, apiErrorResponse{
			Message: ErrSQLNoRows,
			Code:    ErrSQLNoRowsCode,
			Type:    ErrSQLNoRowsType,
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
	if err, ok := err.(*apiError); ok {
		unwrapError := tracerr.Unwrap(err.Err)
		errMsg := unwrapError.Error()
		c.JSON(http.StatusBadRequest, apiErrorResponse{
			Message: errMsg,
			Code:    err.Code,
			Type:    err.Type,
		})
		return
	}
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}
