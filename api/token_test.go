package api

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTokensResponse struct {
	Tokens     []historydb.TokenWithUSD `json:"tokens"`
	Pagination *db.Pagination           `json:"pagination"`
}

func (t *testTokensResponse) GetPagination() *db.Pagination {
	if t.Tokens[0].ItemID < t.Tokens[len(t.Tokens)-1].ItemID {
		t.Pagination.FirstReturnedItem = t.Tokens[0].ItemID
		t.Pagination.LastReturnedItem = t.Tokens[len(t.Tokens)-1].ItemID
	} else {
		t.Pagination.LastReturnedItem = t.Tokens[0].ItemID
		t.Pagination.FirstReturnedItem = t.Tokens[len(t.Tokens)-1].ItemID
	}
	return t.Pagination
}

func (t *testTokensResponse) Len() int {
	return len(t.Tokens)
}

func TestGetToken(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "tokens/"
	fetchedTokens := []historydb.TokenWithUSD{}
	for _, token := range tc.tokens {
		fetchedToken := historydb.TokenWithUSD{}
		assert.NoError(t, doGoodReq("GET", endpoint+strconv.Itoa(int(token.TokenID)), nil, &fetchedToken))
		fetchedTokens = append(fetchedTokens, fetchedToken)
	}
	assertTokensAPIs(t, tc.tokens, fetchedTokens)
}

func TestGetTokens(t *testing.T) {
	endpoint := apiURL + "tokens"
	fetchedTokens := []historydb.TokenWithUSD{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testTokensResponse).Tokens); i++ {
			tmp, err := copystructure.Copy(intr.(*testTokensResponse).Tokens[i])
			if err != nil {
				panic(err)
			}
			fetchedTokens = append(fetchedTokens, tmp.(historydb.TokenWithUSD))
		}
	}
	// Get all (no filters)
	limit := 8
	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)
	assertTokensAPIs(t, tc.tokens, fetchedTokens)

	// Get by tokenIds
	fetchedTokens = []historydb.TokenWithUSD{}
	limit = 7
	stringIds := strconv.Itoa(int(tc.tokens[2].TokenID)) + "," + strconv.Itoa(int(tc.tokens[5].TokenID)) + "," + strconv.Itoa(int(tc.tokens[6].TokenID))
	path = fmt.Sprintf(
		"%s?ids=%s&limit=%d&fromItem=",
		endpoint, stringIds, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)
	var tokensFiltered []historydb.TokenWithUSD
	tokensFiltered = append(tokensFiltered, tc.tokens[2])
	tokensFiltered = append(tokensFiltered, tc.tokens[5])
	tokensFiltered = append(tokensFiltered, tc.tokens[6])
	assertTokensAPIs(t, tokensFiltered, fetchedTokens)

	// Get by symbols
	fetchedTokens = []historydb.TokenWithUSD{}
	limit = 7
	stringSymbols := tc.tokens[1].Symbol + "," + tc.tokens[3].Symbol
	path = fmt.Sprintf(
		"%s?symbols=%s&limit=%d&fromItem=",
		endpoint, stringSymbols, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)
	tokensFiltered = nil
	tokensFiltered = append(tokensFiltered, tc.tokens[1])
	tokensFiltered = append(tokensFiltered, tc.tokens[3])
	assertTokensAPIs(t, tokensFiltered, fetchedTokens)

	// Get by name
	fetchedTokens = []historydb.TokenWithUSD{}
	limit = 5
	stringName := tc.tokens[8].Name[4:5]
	path = fmt.Sprintf(
		"%s?name=%s&limit=%d&fromItem=",
		endpoint, stringName, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)
	tokensFiltered = nil
	tokensFiltered = append(tokensFiltered, tc.tokens[8])
	assertTokensAPIs(t, tokensFiltered, fetchedTokens)

	// Multiple filters
	fetchedTokens = []historydb.TokenWithUSD{}
	limit = 5
	stringSymbols = tc.tokens[2].Symbol + "," + tc.tokens[6].Symbol
	stringIds = strconv.Itoa(int(tc.tokens[2].TokenID)) + "," + strconv.Itoa(int(tc.tokens[5].TokenID)) + "," + strconv.Itoa(int(tc.tokens[6].TokenID))
	path = fmt.Sprintf(
		"%s?symbols=%s&ids=%s&limit=%d&fromItem=",
		endpoint, stringSymbols, stringIds, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)

	tokensFiltered = nil
	tokensFiltered = append(tokensFiltered, tc.tokens[2])
	tokensFiltered = append(tokensFiltered, tc.tokens[6])
	assertTokensAPIs(t, tokensFiltered, fetchedTokens)

	// All, in reverse order
	fetchedTokens = []historydb.TokenWithUSD{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderDesc, &testTokensResponse{}, appendIter)
	assert.NoError(t, err)
	flipedTokens := []historydb.TokenWithUSD{}
	for i := 0; i < len(tc.tokens); i++ {
		flipedTokens = append(flipedTokens, tc.tokens[len(tc.tokens)-1-i])
	}
	assertTokensAPIs(t, flipedTokens, fetchedTokens)
}

func assertTokensAPIs(t *testing.T, expected, actual []historydb.TokenWithUSD) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		actual[i].ItemID = 0
		if expected[i].USDUpdate == nil {
			assert.Equal(t, expected[i].USDUpdate, actual[i].USDUpdate)
		} else {
			assert.Equal(t, expected[i].USDUpdate.Unix(), actual[i].USDUpdate.Unix())
			expected[i].USDUpdate = actual[i].USDUpdate
		}
		assert.Equal(t, expected[i], actual[i])
	}
}
