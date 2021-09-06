package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getBids(c *gin.Context) {
	filters, err := parsers.ParseBidsFilters(c, a.validate)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	bids, pendingItems, err := a.historyDB.GetBidsAPI(historydb.GetBidsAPIRequest{
		SlotNum:    filters.SlotNum,
		BidderAddr: filters.BidderAddr,
		FromItem:   filters.FromItem,
		Limit:      filters.Limit,
		Order:      filters.Order,
	})

	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type bidsResponse struct {
		Bids         []historydb.BidAPI `json:"bids"`
		PendingItems uint64             `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &bidsResponse{
		Bids:         bids,
		PendingItems: pendingItems,
	})
}
