package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func getBids(c *gin.Context) {
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

	bids, pagination, err := h.GetBidsAPI(
		slotNum, bidderAddr, fromItem, limit, order,
	)

	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type bidsResponse struct {
		Bids       []historydb.BidAPI `json:"bids"`
		Pagination *db.Pagination     `json:"pagination"`
	}
	c.JSON(http.StatusOK, &bidsResponse{
		Bids:       bids,
		Pagination: pagination,
	})
}
