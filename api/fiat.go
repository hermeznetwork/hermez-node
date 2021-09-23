package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

func (a *API) getFiatCurrency(c *gin.Context) {
	// Get symbol
	symbol, err := parsers.ParseCurrencyFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch currency from historyDB
	currency, err := a.historyDB.GetCurrencyAPI(symbol)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	c.JSON(http.StatusOK, currency)
}

// CurrenciesResponse is the response object for multiple fiat prices
type CurrenciesResponse struct {
	Currencies []historydb.FiatCurrency `json:"currencies"`
}

func (a *API) getFiatCurrencies(c *gin.Context) {
	// Currency filters
	symbols, err := parsers.ParseCurrenciesFilters(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	// Fetch exits from historyDB
	currencies, err := a.historyDB.GetCurrenciesAPI(symbols)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	c.JSON(http.StatusOK, &CurrenciesResponse{
		Currencies: currencies,
	})
}
