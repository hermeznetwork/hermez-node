package parsers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CurrencyFilter struct to get uri param from /currencies/:symbol request
type CurrencyFilter struct {
	Symbol string `uri:"symbol" binding:"required"`
}

// ParseCurrencyFilter func for parsing currency filter from uri to the symbol
func ParseCurrencyFilter(c *gin.Context) (string, error) {
	var currencyFilter CurrencyFilter
	if err := c.ShouldBindUri(&currencyFilter); err != nil {
		return "", err
	}
	return currencyFilter.Symbol, nil
}

// CurrenciesFilters struct to get query params from /currencies request
type CurrenciesFilters struct {
	Symbols string `form:"symbols"`
}

// ParseCurrenciesFilters func for parsing currencies filters from query to the symbols
func ParseCurrenciesFilters(c *gin.Context) ([]string, error) {
	var currenciesFilters CurrenciesFilters
	var symbols []string
	if err := c.BindQuery(&currenciesFilters); err != nil {
		return symbols, err
	}
	if currenciesFilters.Symbols != "" {
		symbols = strings.Split(currenciesFilters.Symbols, "|")
	}
	return symbols, nil
}
