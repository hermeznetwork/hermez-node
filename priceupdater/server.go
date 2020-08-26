package priceupdater

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *PriceUpdater) startServerPrices() {
	// gin.SetMode(gin.ReleaseMode)
	p.server = gin.Default()
	p.initialitzeRoutes()
}

func (p *PriceUpdater) initialitzeRoutes() {
	p.server.GET("/prices", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tokensList": p.db})
	})
}
