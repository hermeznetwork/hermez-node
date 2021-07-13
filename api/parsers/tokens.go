package parsers

import (
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
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
	IDs       string `form:"ids"`
	Symbols   string `form:"symbols"`
	Name      string `form:"name"`
	Addresses string `form:"addresses"`

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
		ids := strings.Split(tokensFilters.IDs, "|")

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
		symbols = strings.Split(tokensFilters.Symbols, "|")
	}

	var tokenAddresses []ethCommon.Address
	if tokensFilters.Addresses != "" {
		addrs := strings.Split(tokensFilters.Addresses, "|")
		for _, addr := range addrs {
			address := ethCommon.HexToAddress(addr)
			tokenAddresses = append(tokenAddresses, address)
		}
	}

	return historydb.GetTokensAPIRequest{
		Ids:       tokensIDs,
		Symbols:   symbols,
		Name:      tokensFilters.Name,
		Addresses: tokenAddresses,
		FromItem:  tokensFilters.FromItem,
		Limit:     tokensFilters.Limit,
		Order:     *tokensFilters.Order,
	}, nil
}
