package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getFiatCurrency(c *gin.Context) {
	// Get symbol
	symbol := c.Param("symbol")
	if symbol == "" { // symbol is required
		retBadReq(errors.New(ErrInvalidSymbol), c)
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

	// Fetch exits from historyDB
	currencies, err := a.h.GetCurrenciesAPI(symbols)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type CurrenciesResponse struct {
		Currencies []historydb.FiatCurrency `json:"currencies"`
	}
	c.JSON(http.StatusOK, &CurrenciesResponse{
		Currencies: currencies,
	})
}
