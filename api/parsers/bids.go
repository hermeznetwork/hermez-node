package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// BidsFilters struct to hold bids filters
type BidsFilters struct {
	SlotNum    *int64 `form:"slotNum" binding:"omitempty,min=0"`
	BidderAddr string `form:"bidderAddr"`

	Pagination
}

// BidsFiltersStructValidation func for bids filters validation
func BidsFiltersStructValidation(sl validator.StructLevel) {
	ef := sl.Current().Interface().(BidsFilters)

	if ef.SlotNum == nil && ef.BidderAddr == "" {
		sl.ReportError(ef.SlotNum, "slotNum", "SlotNum", "slotnumorbidderaddress", "")
		sl.ReportError(ef.BidderAddr, "bidderAddr", "BidderAddr", "slotnumorbidderaddress", "")
	}
}

// ParseBidsFilters function for parsing bids filters from the request /bids to the GetBidsAPIRequest
func ParseBidsFilters(c *gin.Context, v *validator.Validate) (historydb.GetBidsAPIRequest, error) {
	var bidsFilters BidsFilters
	if err := c.ShouldBindQuery(&bidsFilters); err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}

	if err := v.Struct(bidsFilters); err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}

	bidderAddress, err := common.StringToEthAddr(bidsFilters.BidderAddr)
	if err != nil {
		return historydb.GetBidsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetBidsAPIRequest{
		SlotNum:    bidsFilters.SlotNum,
		BidderAddr: bidderAddress,
		FromItem:   bidsFilters.FromItem,
		Order:      *bidsFilters.Order,
		Limit:      bidsFilters.Limit,
	}, nil
}
