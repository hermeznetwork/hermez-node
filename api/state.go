package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) getState(c *gin.Context) {
	c.JSON(http.StatusOK, a.status)
}
