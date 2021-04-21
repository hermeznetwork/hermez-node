package api

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

func (a *API) noRoute(c *gin.Context) {
	matched, _ := regexp.MatchString(`^/v[0-9]+/`, c.Request.URL.Path)
	if !matched {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Version not provided, please provide a valid version in the path such as v1",
		})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "404 page not found",
	})
}
