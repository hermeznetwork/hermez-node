package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func getCoordinator(c *gin.Context) {
	// Get bidderAddr
	const name = "bidderAddr"
	bidderAddr, err := parseParamEthAddr(name, c)

	if err != nil {
		retBadReq(err, c)
		return
	} else if bidderAddr == nil {
		retBadReq(ErrNillBidderAddr, c)
		return
	}

	coordinator, err := h.GetCoordinatorAPI(*bidderAddr)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	c.JSON(http.StatusOK, coordinator)
}

func getCoordinators(c *gin.Context) {
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}

	// Fetch coordinators from historyDB
	coordinators, pagination, err := h.GetCoordinatorsAPI(fromItem, limit, order)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build succesfull response
	type coordinatorsResponse struct {
		Coordinators []historydb.CoordinatorAPI `json:"coordinators"`
		Pagination   *db.Pagination             `json:"pagination"`
	}
	c.JSON(http.StatusOK, &coordinatorsResponse{
		Coordinators: coordinators,
		Pagination:   pagination,
	})
}
