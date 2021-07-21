package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// SlotFilter struct to get slot filter uri param from /slots/:slotNum request
type SlotFilter struct {
	SlotNum *uint `uri:"slotNum" binding:"required"`
}

// ParseSlotFilter func to parse slot filter from uri to the slot number
func ParseSlotFilter(c *gin.Context) (*uint, error) {
	var slotFilter SlotFilter
	if err := c.ShouldBindUri(&slotFilter); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return slotFilter.SlotNum, nil
}

// SlotsFilters struct to get slots filters from query params from /slots request
type SlotsFilters struct {
	MinSlotNum           *int64 `form:"minSlotNum" binding:"omitempty,min=0"`
	MaxSlotNum           *int64 `form:"maxSlotNum" binding:"omitempty,min=0"`
	WonByEthereumAddress string `form:"wonByEthereumAddress"`
	FinishedAuction      *bool  `form:"finishedAuction"`

	Pagination
}

// SlotsFiltersStructValidation func validating filters struct
func SlotsFiltersStructValidation(sl validator.StructLevel) {
	ef := sl.Current().Interface().(SlotsFilters)

	if ef.MaxSlotNum == nil && ef.FinishedAuction == nil {
		sl.ReportError(ef.MaxSlotNum, "maxSlotNum", "MaxSlotNum", "maxslotnumrequired", "")
		sl.ReportError(ef.FinishedAuction, "finishedAuction", "FinishedAuction", "maxslotnumrequired", "")
	} else if ef.FinishedAuction != nil {
		if ef.MaxSlotNum == nil && !*ef.FinishedAuction {
			sl.ReportError(ef.MaxSlotNum, "maxSlotNum", "MaxSlotNum", "maxslotnumrequired", "")
			sl.ReportError(ef.FinishedAuction, "finishedAuction", "FinishedAuction", "maxslotnumrequired", "")
		}
	} else if ef.MaxSlotNum != nil && ef.MinSlotNum != nil {
		if *ef.MinSlotNum > *ef.MaxSlotNum {
			sl.ReportError(ef.MaxSlotNum, "maxSlotNum", "MaxSlotNum", "maxslotlessthanminslot", "")
			sl.ReportError(ef.MinSlotNum, "minSlotNum", "MinSlotNum", "maxslotlessthanminslot", "")
		}
	}
}

// ParseSlotsFilters func for parsing slots filters to the GetBestBidsAPIRequest
func ParseSlotsFilters(c *gin.Context, v *validator.Validate) (historydb.GetBestBidsAPIRequest, error) {
	var slotsFilters SlotsFilters
	if err := c.ShouldBindQuery(&slotsFilters); err != nil {
		return historydb.GetBestBidsAPIRequest{}, err
	}

	if err := v.Struct(slotsFilters); err != nil {
		return historydb.GetBestBidsAPIRequest{}, tracerr.Wrap(err)
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
