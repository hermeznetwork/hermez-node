package priceupdater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCon(t *testing.T) {

	config := ConfigPriceUpdater{

		RecommendedFee:              1,
		RecommendedCreateAccountFee: 1,
		TokensList:                  []string{"ETH", "NEC"},
		ApiUrl:                      "https://api-pub.bitfinex.com/v2/",
	}

	pud := NewPriceUpdater(config)

	assert.Equal(t, pud.Config.TokensList[0], "ETH")
	assert.Equal(t, pud.Config.TokensList[1], "NEC")

	err := pud.UpdatePrices()
	assert.Equal(t, err, nil)

	info, _ := pud.Get("ETH")
	assert.NotZero(t, info.Value)

	info2, _ := pud.Get("NEC")
	assert.NotZero(t, info2.Value)

	info3, err := pud.Get("INVENTED")
	if assert.Error(t, err) {
		assert.Equal(t, ErrSymbolDoesNotExistInDatabase, err)
	}
	assert.Equal(t, info3.Value, float64(0))

	prices := pud.GetPrices()
	assert.Equal(t, prices["ETH"], info)
	assert.Equal(t, prices["NEC"], info2)

}
