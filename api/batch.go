package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

func (a *API) getBatches(c *gin.Context) {
	// Get query parameters
	filter, err := parsers.ParseBatchesFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch batches from historyDB
	batches, pendingItems, err := a.historyDB.GetBatchesAPI(filter)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type batchesResponse struct {
		Batches      []historydb.BatchAPI `json:"batches"`
		PendingItems uint64               `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &batchesResponse{
		Batches:      batches,
		PendingItems: pendingItems,
	})
}

func (a *API) getBatch(c *gin.Context) {
	// Get batchNum
	batchNum, err := parsers.ParseBatchFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch batch from historyDB
	batch, err := a.historyDB.GetBatchAPI(common.BatchNum(*batchNum))
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// JSON response
	c.JSON(http.StatusOK, batch)
}

type fullBatch struct {
	Batch *historydb.BatchAPI `json:"batch"`
	Txs   []historydb.TxAPI   `json:"transactions"`
}

func (a *API) getFullBatch(c *gin.Context) {
	// Get batchNum
	batchNum, err := parsers.ParseBatchFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch batch from historyDB
	batch, err := a.historyDB.GetBatchAPI(common.BatchNum(*batchNum))
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Fetch txs forged in the batch from historyDB
	maxTxsPerBatch := uint(2048) //nolint:gomnd
	txs, _, err := a.historyDB.GetTxsAPI(historydb.GetTxsAPIRequest{
		BatchNum: batchNum,
		Limit:    &maxTxsPerBatch,
		Order:    db.OrderAsc,
	})
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		retSQLErr(err, c)
		return
	}
	// JSON response
	c.JSON(http.StatusOK, fullBatch{
		Batch: batch,
		Txs:   txs,
	})
}
