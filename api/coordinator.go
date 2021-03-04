package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getCoordinators(c *gin.Context) {
	bidderAddr, err := parseQueryEthAddr("bidderAddr", c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	forgerAddr, err := parseQueryEthAddr("forgerAddr", c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Fetch coordinators from historyDB
	coordinators, pendingItems, err := a.h.GetCoordinatorsAPI(bidderAddr, forgerAddr, fromItem, limit, order)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type coordinatorsResponse struct {
		Coordinators []historydb.CoordinatorAPI `json:"coordinators"`
		PendingItems uint64                     `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &coordinatorsResponse{
		Coordinators: coordinators,
		PendingItems: pendingItems,
	})
}
