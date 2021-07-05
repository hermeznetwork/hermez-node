package parsers

import (
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

type coordinatorsFilters struct {
	BidderAddr string `form:"bidderAddr"`
	ForgerAddr string `form:"forgerAddr"`

	Pagination
}

func ParseCoordinatorsFilters(c *gin.Context) (historydb.GetCoordinatorsAPIRequest, error) {
	var coordinatorsFilters coordinatorsFilters
	if err := c.BindQuery(&coordinatorsFilters); err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}
	bidderAddr, err := common.StringToEthAddr(coordinatorsFilters.BidderAddr)
	if err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}
	forgerAddr, err := common.StringToEthAddr(coordinatorsFilters.ForgerAddr)
	if err != nil {
		return historydb.GetCoordinatorsAPIRequest{}, tracerr.Wrap(err)
	}

	return historydb.GetCoordinatorsAPIRequest{
		BidderAddr: bidderAddr,
		ForgerAddr: forgerAddr,
		FromItem:   coordinatorsFilters.FromItem,
		Limit:      coordinatorsFilters.Limit,
		Order:      *coordinatorsFilters.Order,
	}, nil
}
