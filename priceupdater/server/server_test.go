package server

import (
	"testing"

	"github.com/dghubble/sling"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/priceupdater"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost"
const port = "3002"
const path = "prices"

var pud *priceupdater.PriceUpdater

func TestConServer(t *testing.T) {
	config := priceupdater.Config{
		RecommendedFee:              1,
		RecommendedCreateAccountFee: 1,
		TokensList:                  []string{"ETH", "NEC"},
		APIURL:                      "https://api-pub.bitfinex.com/v2/",
	}
	// Init PriceUpdater
	pud = priceupdater.NewPriceUpdater(config)
	// Update Prices
	err := pud.UpdatePrices()
	assert.Equal(t, err, nil)
	// Init Server
	server := gin.Default()
	InitializeServerPrices(pud, server)
	// Start Server
	go func() {
		errServer := server.Run(":" + port)
		if errServer != nil {
			panic(errServer)
		}
	}()
}

func TestGetPricesServer(t *testing.T) {
	info, _ := pud.Get("ETH")
	assert.NotZero(t, info.Value)

	info2, _ := pud.Get("NEC")
	assert.NotZero(t, info2.Value)

	paramsRec := new(map[string]priceupdater.TokenInfo)

	var err error
	_, errReceive := sling.New().Get(baseURL+":"+port).Path(path).Receive(paramsRec, err)
	if errReceive != nil {
		panic(errReceive)
	}
	assert.Equal(t, err, nil)

	assert.Equal(t, (*paramsRec)["ETH"].Value, info.Value)
	assert.Equal(t, (*paramsRec)["NEC"].Value, info2.Value)

	assert.Equal(t, (*paramsRec)["ETH"].Symbol, info.Symbol)
	assert.Equal(t, (*paramsRec)["NEC"].Symbol, info2.Symbol)

	lastUpdatedETH, _ := (*paramsRec)["ETH"].LastUpdated.MarshalText()
	lastUpdatedNEC, _ := (*paramsRec)["NEC"].LastUpdated.MarshalText()

	infoLastUpdatedETH, _ := info.LastUpdated.MarshalText()
	infoLastUpdatedNEC, _ := info2.LastUpdated.MarshalText()

	assert.Equal(t, lastUpdatedETH, infoLastUpdatedETH)
	assert.Equal(t, lastUpdatedNEC, infoLastUpdatedNEC)
}
