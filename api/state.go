package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) getState(c *gin.Context) {
	stateAPI, err := a.h.GetStateAPI()
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, stateAPI)
}
