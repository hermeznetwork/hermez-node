package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

func (a *API) getBatches(c *gin.Context) {
	// Get query parameters
	// minBatchNum
	minBatchNum, err := parseQueryUint("minBatchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// maxBatchNum
	maxBatchNum, err := parseQueryUint("maxBatchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// slotNum
	slotNum, err := parseQueryUint("slotNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// forgerAddr
	forgerAddr, err := parseQueryEthAddr("forgerAddr", c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch batches from historyDB
	batches, pendingItems, err := a.h.GetBatchesAPI(
		minBatchNum, maxBatchNum, slotNum, forgerAddr, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
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
	batchNum, err := parseParamUint("batchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if batchNum == nil { // batchNum is required
		retBadReq(errors.New("Invalid batchNum"), c)
		return
	}
	// Fetch batch from historyDB
	batch, err := a.h.GetBatchAPI(common.BatchNum(*batchNum))
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
	batchNum, err := parseParamUint("batchNum", nil, 0, maxUint32, c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if batchNum == nil {
		retBadReq(errors.New("Invalid batchNum"), c)
		return
	}
	// Fetch batch from historyDB
	batch, err := a.h.GetBatchAPI(common.BatchNum(*batchNum))
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Fetch txs forged in the batch from historyDB
	maxTxsPerBatch := uint(2048) //nolint:gomnd
	txs, _, err := a.h.GetTxsAPI(
		nil, nil, nil, nil, batchNum, nil, nil, &maxTxsPerBatch, historydb.OrderAsc,
	)
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
