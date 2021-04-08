package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getBids(c *gin.Context) {
	slotNum, bidderAddr, err := parseBidFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	if slotNum == nil && bidderAddr == nil {
		retBadReq(errors.New("It is necessary to add at least one filter: slotNum or/and bidderAddr"), c)
		return
	}
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	bids, pendingItems, err := a.h.GetBidsAPI(historydb.GetBidsAPIRequest{
		SlotNum:    slotNum,
		BidderAddr: bidderAddr,
		FromItem:   fromItem,
		Limit:      limit,
		Order:      order,
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
