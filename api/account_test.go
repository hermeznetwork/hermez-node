package api

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
)

type testAccount struct {
	ItemID    int                    `json:"itemId"`
	Idx       apitypes.HezIdx        `json:"accountIndex"`
	BatchNum  common.BatchNum        `json:"batchNum"`
	PublicKey apitypes.HezBJJ        `json:"bjj"`
	EthAddr   apitypes.HezEthAddr    `json:"hezEthereumAddress"`
	Nonce     common.Nonce           `json:"nonce"`
	Balance   *apitypes.BigIntStr    `json:"balance"`
	Token     historydb.TokenWithUSD `json:"token"`
}

type testAccountsResponse struct {
	Accounts   []testAccount  `json:"accounts"`
	Pagination *db.Pagination `json:"pagination"`
}

func genTestAccounts(accounts []common.Account, tokens []historydb.TokenWithUSD) []testAccount {
	tAccounts := []testAccount{}
	for x, account := range accounts {
		token := getTokenByID(account.TokenID, tokens)
		tAccount := testAccount{
			ItemID:    x + 1,
			Idx:       apitypes.HezIdx(idxToHez(account.Idx, token.Symbol)),
			PublicKey: apitypes.NewHezBJJ(account.PublicKey),
			EthAddr:   apitypes.NewHezEthAddr(account.EthAddr),
			Nonce:     account.Nonce,
			Balance:   apitypes.NewBigIntStr(account.Balance),
			Token:     token,
		}
		tAccounts = append(tAccounts, tAccount)
	}
	return tAccounts
}

func (t *testAccountsResponse) GetPagination() *db.Pagination {
	if t.Accounts[0].ItemID < t.Accounts[len(t.Accounts)-1].ItemID {
		t.Pagination.FirstReturnedItem = t.Accounts[0].ItemID
		t.Pagination.LastReturnedItem = t.Accounts[len(t.Accounts)-1].ItemID
	} else {
		t.Pagination.LastReturnedItem = t.Accounts[0].ItemID
		t.Pagination.FirstReturnedItem = t.Accounts[len(t.Accounts)-1].ItemID
	}
	return t.Pagination
}

func (t *testAccountsResponse) Len() int { return len(t.Accounts) }

func TestGetAccounts(t *testing.T) {
	endpoint := apiURL + "accounts"
	fetchedAccounts := []testAccount{}

	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testAccountsResponse).Accounts); i++ {
			tmp, err := copystructure.Copy(intr.(*testAccountsResponse).Accounts[i])
			if err != nil {
				panic(err)
			}
			fetchedAccounts = append(fetchedAccounts, tmp.(testAccount))
		}
	}

	limit := 5
	stringIds := strconv.Itoa(int(tc.tokens[2].TokenID)) + "," + strconv.Itoa(int(tc.tokens[5].TokenID)) + "," + strconv.Itoa(int(tc.tokens[6].TokenID))

	// Filter by BJJ
	path := fmt.Sprintf("%s?BJJ=%s&limit=%d&fromItem=", endpoint, tc.accounts[0].PublicKey, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// Filter by ethAddr
	path = fmt.Sprintf("%s?hermezEthereumAddress=%s&limit=%d&fromItem=", endpoint, tc.accounts[0].EthAddr, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// both filters (incompatible)
	path = fmt.Sprintf("%s?hermezEthereumAddress=%s&BJJ=%s&limit=%d&fromItem=", endpoint, tc.accounts[0].EthAddr, tc.accounts[0].PublicKey, limit)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	fetchedAccounts = []testAccount{}
	// Filter by token IDs
	path = fmt.Sprintf("%s?tokenIds=%s&limit=%d&fromItem=", endpoint, stringIds, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// Token Ids + bjj
	path = fmt.Sprintf("%s?tokenIds=%s&BJJ=%s&limit=%d&fromItem=", endpoint, stringIds, tc.accounts[0].PublicKey, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// No filters (checks response content)
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Equal(t, len(tc.accounts), len(fetchedAccounts))
	for i := 0; i < len(fetchedAccounts); i++ {
		fetchedAccounts[i].Token.ItemID = 0
		if tc.accounts[i].Token.USDUpdate != nil {
			assert.Less(t, fetchedAccounts[i].Token.USDUpdate.Unix()-3, tc.accounts[i].Token.USDUpdate.Unix())
			fetchedAccounts[i].Token.USDUpdate = tc.accounts[i].Token.USDUpdate
		}
		assert.Equal(t, tc.accounts[i], fetchedAccounts[i])
	}

	// No filters  Reverse Order (checks response content)
	reversedAccounts := []testAccount{}
	appendIter = func(intr interface{}) {
		for i := 0; i < len(intr.(*testAccountsResponse).Accounts); i++ {
			tmp, err := copystructure.Copy(intr.(*testAccountsResponse).Accounts[i])
			if err != nil {
				panic(err)
			}
			reversedAccounts = append(reversedAccounts, tmp.(testAccount))
		}
	}
	err = doGoodReqPaginated(path, historydb.OrderDesc, &testAccountsResponse{}, appendIter)
	assert.NoError(t, err)
	assert.Equal(t, len(reversedAccounts), len(fetchedAccounts))
	for i := 0; i < len(fetchedAccounts); i++ {
		reversedAccounts[i].Token.ItemID = 0
		fetchedAccounts[len(fetchedAccounts)-1-i].Token.ItemID = 0
		if reversedAccounts[i].Token.USDUpdate != nil {
			assert.Less(t, fetchedAccounts[len(fetchedAccounts)-1-i].Token.USDUpdate.Unix()-3, reversedAccounts[i].Token.USDUpdate.Unix())
			fetchedAccounts[len(fetchedAccounts)-1-i].Token.USDUpdate = reversedAccounts[i].Token.USDUpdate
		}
		assert.Equal(t, reversedAccounts[i], fetchedAccounts[len(fetchedAccounts)-1-i])
	}

	// Test GetAccount
	path = fmt.Sprintf("%s/%s", endpoint, fetchedAccounts[2].Idx)
	account := testAccount{}
	assert.NoError(t, doGoodReq("GET", path, nil, &account))
	account.Token.ItemID = 0
	assert.Equal(t, fetchedAccounts[2], account)

	// 400
	path = fmt.Sprintf("%s/hez:12345", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// 404
	path = fmt.Sprintf("%s/hez:10:12345", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}
