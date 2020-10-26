package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func getBatches(c *gin.Context) {
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
	batches, pagination, err := h.GetBatchesAPI(
		minBatchNum, maxBatchNum, slotNum, forgerAddr, fromItem, limit, order,
	)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type batchesResponse struct {
		Batches    []historydb.BatchAPI `json:"batches"`
		Pagination *db.Pagination       `json:"pagination"`
	}
	c.JSON(http.StatusOK, &batchesResponse{
		Batches:    batches,
		Pagination: pagination,
	})
}

func getBatch(c *gin.Context) {
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
	batch, err := h.GetBatchAPI(common.BatchNum(*batchNum))
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// JSON response
	c.JSON(http.StatusOK, batch)
}

type fullBatch struct {
	Batch *historydb.BatchAPI
	Txs   []historyTxAPI
}

func getFullBatch(c *gin.Context) {
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
	batch, err := h.GetBatchAPI(common.BatchNum(*batchNum))
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Fetch txs from historyDB
	// TODO
	txs := []historyTxAPI{}
	// JSON response
	c.JSON(http.StatusOK, fullBatch{
		Batch: batch,
		Txs:   txs,
	})
}
