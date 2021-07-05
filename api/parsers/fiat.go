package parsers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type currencyFilter struct {
	Symbol string `uri:"symbol" binding:"required"`
}

func ParseCurrencyFilter(c *gin.Context) (string, error) {
	var currencyFilter currencyFilter
	if err := c.ShouldBindUri(&currencyFilter); err != nil {
		return "", err
	}
	return currencyFilter.Symbol, nil
}

type currenciesFilters struct {
	Symbols string `form:"symbols"`
}

func ParseCurrenciesFilters(c *gin.Context) ([]string, error) {
	var currenciesFilters currenciesFilters
	var symbols []string
	if err := c.BindQuery(&currenciesFilters); err != nil {
		return symbols, err
	}
	if currenciesFilters.Symbols != "" {
		symbols = strings.Split(currenciesFilters.Symbols, "|")
	}
	return symbols, nil
}
