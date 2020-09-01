package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/priceupdater"
)

// PriceUpdaterWithServer PriceUpdater + server
type PriceUpdaterWithServer struct {
	PU     *priceupdater.PriceUpdater
	Server *gin.Engine
}

// InitializeServerPrices initialize the priceupdater server and return server
func InitializeServerPrices(p *priceupdater.PriceUpdater, server *gin.Engine) {
	// gin.SetMode(gin.ReleaseMode)
	PUServer := new(PriceUpdaterWithServer)
	PUServer.PU = p
	PUServer.Server = server
	PUServer.initialitzeRoutes()
}

func (PUServer *PriceUpdaterWithServer) initialitzeRoutes() {
	PUServer.Server.GET("/prices", func(c *gin.Context) {
		c.JSON(http.StatusOK, PUServer.PU.DB)
	})
}
