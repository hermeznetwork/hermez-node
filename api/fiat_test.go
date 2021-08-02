package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/stretchr/testify/assert"
)

func genFiatPrices(db *historydb.HistoryDB) error {
	_, err := db.DB().Exec(
		"INSERT INTO fiat(currency, base_currency, price) VALUES ($1, $2, $3);",
		"EUR", "USD", 0.82,
	)
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
	client := &http.Client{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	var response CurrenciesResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Error message: %v", response)
	}
	assert.NoError(t, err)
	assert.Equal(t, response.Currencies[0].BaseCurrency, "USD")
	assert.Equal(t, response.Currencies[0].Currency, "EUR")
	assert.Equal(t, response.Currencies[0].Price, 0.82)

	//Get some fiat currencies
	endpoint = endpoint + "?symbols=EUR"
	req, err = http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	//nolint
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Fatalf("Error message: %v", response)
	}
	assert.NoError(t, err)
	assert.Equal(t, response.Currencies[0].BaseCurrency, "USD")
	assert.Equal(t, response.Currencies[0].Currency, "EUR")
	assert.Equal(t, response.Currencies[0].Price, 0.82)

	//Get EUR fiat currency
	endpoint = apiURL + "currencies/EUR"
	var singleItemResp historydb.FiatCurrency
	req, err = http.NewRequest("GET", endpoint, nil)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	//nolint
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	err = json.Unmarshal(body, &singleItemResp)
	if err != nil {
		t.Fatalf("Error message: %v", singleItemResp)
	}
	assert.NoError(t, err)
	assert.Equal(t, singleItemResp.BaseCurrency, "USD")
	assert.Equal(t, singleItemResp.Currency, "EUR")
	assert.Equal(t, singleItemResp.Price, 0.82)
}
