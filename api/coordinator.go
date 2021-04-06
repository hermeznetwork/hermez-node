package api

import (
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

type GetCoordinatorsAPIRequest struct {
	BidderAddr *ethCommon.Address
	ForgerAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}

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

	request := GetCoordinatorsAPIRequest{
		BidderAddr: bidderAddr,
		ForgerAddr: forgerAddr,
		FromItem:   fromItem,
		Limit:      limit,
		Order:      order,
	}
	// Fetch coordinators from historyDB
	coordinators, pendingItems, err := a.h.GetCoordinatorsAPI(request)
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
