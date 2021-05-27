package api

import (
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/stretchr/testify/assert"
)

func genFiatPrices(db *historydb.HistoryDB) error {
	err := db.CreateFiatPrice("EUR", "USD", 0.82)
	if err != nil {
		return err
	}
	return nil
}

func TestFiat(t *testing.T) {
	// HistoryDB
	database, err := db.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	apiConnCon := db.NewAPIConnectionController(1, time.Second)
	db := historydb.NewHistoryDB(database, database, apiConnCon)
	if err != nil {
		panic(err)
	}
	err = genFiatPrices(db)
	assert.NoError(t, err)

	//Get all fiat currencies
	endpoint := apiURL + "currencies/"
	type responseTest struct {
		Currencies []historydb.FiatCurrency
	}
	var response responseTest
	err = doGoodReq("GET", endpoint, nil, &response)
	assert.NoError(t, err)
	assert.Equal(t, response.Currencies[0].BaseCurrency, "USD")
	assert.Equal(t, response.Currencies[0].Currency, "EUR")
	assert.Equal(t, response.Currencies[0].Price, 0.82)

	//Get some fiat currencies
	endpoint = endpoint + "?symbols=EUR"
	err = doGoodReq("GET", endpoint, nil, &response)
	assert.NoError(t, err)
	assert.Equal(t, response.Currencies[0].BaseCurrency, "USD")
	assert.Equal(t, response.Currencies[0].Currency, "EUR")
	assert.Equal(t, response.Currencies[0].Price, 0.82)

	//Get EUR fiat currency
	endpoint = apiURL + "currencies/EUR"
	var singleItemResp historydb.FiatCurrency
	err = doGoodReq("GET", endpoint, nil, &singleItemResp)
	assert.NoError(t, err)
	assert.Equal(t, singleItemResp.BaseCurrency, "USD")
	assert.Equal(t, singleItemResp.Currency, "EUR")
	assert.Equal(t, singleItemResp.Price, 0.82)
}
