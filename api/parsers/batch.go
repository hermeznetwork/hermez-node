package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

// BatchFilter struct to hold batch num from request /batches/:batchNum
type BatchFilter struct {
	BatchNum uint `uri:"batchNum" binding:"required"`
}

// ParseBatchFilter parsing /batches request to the batch num
func ParseBatchFilter(c *gin.Context) (*uint, error) {
	var batchFilter BatchFilter
	if err := c.ShouldBindUri(&batchFilter); err != nil {
		return nil, err
	}
	return &batchFilter.BatchNum, nil
}

// BatchesFilters struct to hold batch num from request /batches/:batchNum
type BatchesFilters struct {
	MinBatchNum *uint  `form:"minBatchNum"`
	MaxBatchNum *uint  `form:"maxBatchNum"`
	SlotNum     *uint  `form:"slotNum"`
	ForgerAddr  string `form:"forgerAddr"`

	Pagination
}

// ParseBatchesFilter parsing batches filter to the GetBatchesAPIRequest
func ParseBatchesFilter(c *gin.Context) (historydb.GetBatchesAPIRequest, error) {
	var batchesFilters BatchesFilters
	if err := c.ShouldBindQuery(&batchesFilters); err != nil {
		return historydb.GetBatchesAPIRequest{}, err
	}

	addr, err := common.StringToEthAddr(batchesFilters.ForgerAddr)
	if err != nil {
		return historydb.GetBatchesAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetBatchesAPIRequest{
		MinBatchNum: batchesFilters.MinBatchNum,
		MaxBatchNum: batchesFilters.MaxBatchNum,
		SlotNum:     batchesFilters.SlotNum,
		ForgerAddr:  addr,
		FromItem:    batchesFilters.FromItem,
		Limit:       batchesFilters.Limit,
		Order:       *batchesFilters.Order,
	}, nil
}
