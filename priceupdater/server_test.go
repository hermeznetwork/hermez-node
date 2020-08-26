package priceupdater

import (
	"testing"

	"github.com/dghubble/sling"
	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost"
const port = "3002"

var pud *PriceUpdater

// ListJSON tokens list
type ListJSON struct {
	ListTokens map[string]TokenInfo `json:"tokensList,omitempty"`
}

func TestConServer(t *testing.T) {
	config := ConfigPriceUpdater{
		RecommendedFee:              1,
		RecommendedCreateAccountFee: 1,
		TokensList:                  []string{"ETH", "NEC"},
		APIURL:                      "https://api-pub.bitfinex.com/v2/",
	}
	pud = NewPriceUpdater(config)
	err := pud.UpdatePrices()
	assert.Equal(t, err, nil)
	pud.startServerPrices()
	go func() {
		errServer := pud.server.Run(":" + port)
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

	paramsRec := new(ListJSON)
	var err error
	path := "prices"
	_, errReceive := sling.New().Get(baseURL+":"+port).Path(path).Receive(paramsRec, err)
	if errReceive != nil {
		panic(errReceive)
	}
	assert.Equal(t, err, nil)

	assert.Equal(t, paramsRec.ListTokens["ETH"].Value, info.Value)
	assert.Equal(t, paramsRec.ListTokens["NEC"].Value, info2.Value)

	assert.Equal(t, paramsRec.ListTokens["ETH"].Symbol, info.Symbol)
	assert.Equal(t, paramsRec.ListTokens["NEC"].Symbol, info2.Symbol)

	lastUpdatedETH, _ := paramsRec.ListTokens["ETH"].LastUpdated.MarshalText()
	lastUpdatedNEC, _ := paramsRec.ListTokens["NEC"].LastUpdated.MarshalText()

	infoLastUpdatedETH, _ := info.LastUpdated.MarshalText()
	infoLastUpdatedNEC, _ := info2.LastUpdated.MarshalText()

	assert.Equal(t, lastUpdatedETH, infoLastUpdatedETH)
	assert.Equal(t, lastUpdatedNEC, infoLastUpdatedNEC)
}
