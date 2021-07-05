package parsers

import (
	"errors"
	"github.com/hermeznetwork/hermez-node/common"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type bidsFilters struct {
	SlotNum    *int64 `form:"slotNum" binding:"omitempty,min=0"`
	BidderAddr string `form:"bidderAddr"`

	Pagination
}

func ParseBidsFilters(c *gin.Context) (historydb.GetBidsAPIRequest, error) {
	var bidsFilters bidsFilters
	if err := c.ShouldBindQuery(&bidsFilters); err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}
	bidderAddress, err := common.StringToEthAddr(bidsFilters.BidderAddr)
	if err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}

	if bidsFilters.SlotNum == nil && bidderAddress == nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(errors.New("It is necessary to add at least one filter: slotNum or/and bidderAddr"))
	}

	return historydb.GetBidsAPIRequest{
		SlotNum:    bidsFilters.SlotNum,
		BidderAddr: bidderAddress,
		FromItem:   bidsFilters.FromItem,
		Order:      *bidsFilters.Order,
		Limit:      bidsFilters.Limit,
	}, nil
}
