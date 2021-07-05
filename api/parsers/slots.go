package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type slotFilter struct {
	SlotNum *uint `uri:"slotNum" binding:"required"`
}

func ParseSlotFilter(c *gin.Context) (*uint, error) {
	var slotFilter slotFilter
	if err := c.ShouldBindUri(&slotFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return slotFilter.SlotNum, nil
}

type slotsFilters struct {
	MinSlotNum           *int64 `form:"minSlotNum" binding:"omitempty,min=0"`
	MaxSlotNum           *int64 `form:"maxSlotNum" binding:"omitempty,min=0"`
	WonByEthereumAddress string `form:"wonByEthereumAddress"`
	FinishedAuction      *bool  `form:"finishedAuction"`

	Pagination
}

func ParseSlotsFilters(c *gin.Context) (historydb.GetBestBidsAPIRequest, error) {
	var slotsFilters slotsFilters
	if err := c.ShouldBindQuery(&slotsFilters); err != nil {
		return historydb.GetBestBidsAPIRequest{}, err
	}

	wonByEthereumAddress, err := common.StringToEthAddr(slotsFilters.WonByEthereumAddress)
	if err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetBestBidsAPIRequest{
		MinSlotNum:      slotsFilters.MinSlotNum,
		MaxSlotNum:      slotsFilters.MaxSlotNum,
		BidderAddr:      wonByEthereumAddress,
		FinishedAuction: slotsFilters.FinishedAuction,
		FromItem:        slotsFilters.FromItem,
		Order:           *slotsFilters.Order,
		Limit:           slotsFilters.Limit,
	}, nil
}
