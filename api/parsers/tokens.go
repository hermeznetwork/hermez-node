package parsers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

// TokenFilter struct to get token uri param from /tokens/:id request
type TokenFilter struct {
	ID *uint `uri:"id" binding:"required"`
}

// ParseTokenFilter for parsing token filter from uri to the id
func ParseTokenFilter(c *gin.Context) (*uint, error) {
	var tokenFilter TokenFilter
	if err := c.ShouldBindUri(&tokenFilter); err != nil {
		return nil, err
	}
	return tokenFilter.ID, nil
}

// TokensFilters struct to get token query params from /tokens request
type TokensFilters struct {
	IDs     string `form:"ids"`
	Symbols string `form:"symbols"`
	Name    string `form:"name"`

	Pagination
}

// ParseTokensFilters function for parsing tokens filters to the GetTokensAPIRequest
func ParseTokensFilters(c *gin.Context) (historydb.GetTokensAPIRequest, error) {
	var tokensFilters TokensFilters
	if err := c.BindQuery(&tokensFilters); err != nil {
		return historydb.GetTokensAPIRequest{}, err
	}
	var tokensIDs []common.TokenID
	if tokensFilters.IDs != "" {
		ids := strings.Split(tokensFilters.IDs, ",")

		for _, id := range ids {
			idUint, err := strconv.Atoi(id)
			if err != nil {
				return historydb.GetTokensAPIRequest{}, tracerr.Wrap(err)
			}
			tokenID := common.TokenID(idUint)
			tokensIDs = append(tokensIDs, tokenID)
		}
	}

	var symbols []string
	if tokensFilters.Symbols != "" {
		symbols = strings.Split(tokensFilters.Symbols, ",")
	}

	return historydb.GetTokensAPIRequest{
		Ids:      tokensIDs,
		Symbols:  symbols,
		Name:     tokensFilters.Name,
		FromItem: tokensFilters.FromItem,
		Limit:    tokensFilters.Limit,
		Order:    *tokensFilters.Order,
	}, nil
}
