package priceupdater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCon(t *testing.T) {

	var db MemoryDB = make(map[string]TokenInfo)

	config := ConfigPriceUpdater{

		RecommendedFee:              1,
		RecommendedCreateAccountFee: 1,
		TokensList:                  []string{"ETH", "NEC"},
	}

	pud, err := NewPriceUpdater(&db, config)
	assert.Equal(t, err, nil)
	assert.Equal(t, pud.Config.TokensList[0], "ETH")
	assert.Equal(t, pud.Config.TokensList[1], "NEC")

	err = pud.UpdatePrices()
	assert.Equal(t, err, nil)

	info, err := pud.Get("ETH")
	assert.NotZero(t, info.Value)

	info2, err := pud.Get("NEC")
	assert.NotZero(t, info2.Value)

	info3, err := pud.Get("INVENTED")
	assert.Equal(t, info3.Value, float64(0))

	prices, err := pud.DB.GetPrices()
	assert.Equal(t, prices["ETH"], info)
	assert.Equal(t, prices["NEC"], info2)

}
