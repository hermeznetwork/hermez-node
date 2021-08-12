package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getCoordinators(c *gin.Context) {
	filters, err := parsers.ParseCoordinatorsFilters(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	// Fetch coordinators from historyDB
	coordinators, pendingItems, err := a.historyDB.GetCoordinatorsAPI(filters)
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
