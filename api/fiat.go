package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getFiatCurrency(c *gin.Context) {
	// Get TokenID
	symbol := c.Param("symbol")
	if symbol == "" { // tokenID is required
		retBadReq(errors.New("Invalid Symbol"), c)
		return
	}
	// Fetch currency from historyDB
	currency, err := a.h.GetCurrencyAPI(symbol)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, currency)
}

func (a *API) getFiatCurrencies(c *gin.Context) {
	// Currency filters
	symbols, err := parseCurrencyFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	order := c.Query("order")

	// Fetch exits from historyDB
	currencies, pendingItems, err := a.h.GetCurrenciesAPI(historydb.GetCurrencyAPIRequest{
		Symbols: symbols,
		Order:   order,
	})
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type CurrenciesResponse struct {
		Currencies   []historydb.FiatCurrency `json:"currencies"`
		PendingItems uint64                   `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &CurrenciesResponse{
		Currencies:   currencies,
		PendingItems: pendingItems,
	})
}
