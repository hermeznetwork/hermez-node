package api

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAccount struct {
	ItemID    uint64                 `json:"itemId"`
	Idx       apitypes.HezIdx        `json:"accountIndex"`
	BatchNum  common.BatchNum        `json:"batchNum"`
	PublicKey apitypes.HezBJJ        `json:"bjj"`
	EthAddr   apitypes.HezEthAddr    `json:"hezEthereumAddress"`
	Nonce     nonce.Nonce            `json:"nonce"`
	Balance   *apitypes.BigIntStr    `json:"balance"`
	Token     historydb.TokenWithUSD `json:"token"`
}

type testAccountsResponse struct {
	Accounts     []testAccount `json:"accounts"`
	PendingItems uint64        `json:"pendingItems"`
}

func (t testAccountsResponse) GetPending() (pendingItems, lastItemID uint64) {
	pendingItems = t.PendingItems
	lastItemID = t.Accounts[len(t.Accounts)-1].ItemID
	return pendingItems, lastItemID
}

func (t *testAccountsResponse) Len() int { return len(t.Accounts) }

func (t testAccountsResponse) New() Pendinger { return &testAccountsResponse{} }

func genTestAccounts(accounts []common.Account, tokens []historydb.TokenWithUSD) []testAccount {
	tAccounts := []testAccount{}
	for x, account := range accounts {
		token := getTokenByID(account.TokenID, tokens)
		tAccount := testAccount{
			ItemID:    uint64(x + 1),
			Idx:       apitypes.HezIdx(common.IdxToHez(account.Idx, token.Symbol)),
			PublicKey: apitypes.NewHezBJJ(account.BJJ),
			EthAddr:   apitypes.NewHezEthAddr(account.EthAddr),
			Nonce:     account.Nonce,
			Balance:   apitypes.NewBigIntStr(account.Balance),
			Token:     token,
		}
		tAccounts = append(tAccounts, tAccount)
	}
	return tAccounts
}

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
	path := fmt.Sprintf("%s?BJJ=%s&limit=%d", endpoint, tc.accounts[0].PublicKey, limit)
	err := doGoodReqPaginated(path, db.OrderAsc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// Filter by ethAddr
	path = fmt.Sprintf("%s?hezEthereumAddress=%s&limit=%d", endpoint, tc.accounts[3].EthAddr, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// both filters (incompatible)
	path = fmt.Sprintf("%s?hezEthereumAddress=%s&BJJ=%s&limit=%d", endpoint, tc.accounts[0].EthAddr, tc.accounts[0].PublicKey, limit)
	err = doBadReq("GET", path, nil, 400)
	require.NoError(t, err)
	fetchedAccounts = []testAccount{}
	// Filter by token IDs
	path = fmt.Sprintf("%s?tokenIds=%s&limit=%d", endpoint, stringIds, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// Token Ids + bjj
	path = fmt.Sprintf("%s?tokenIds=%s&BJJ=%s&limit=%d", endpoint, stringIds, tc.accounts[10].PublicKey, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
	assert.Greater(t, len(fetchedAccounts), 0)
	assert.LessOrEqual(t, len(fetchedAccounts), len(tc.accounts))
	fetchedAccounts = []testAccount{}
	// No filters (checks response content)
	path = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
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
	err = doGoodReqPaginated(path, db.OrderDesc, &testAccountsResponse{}, appendIter)
	require.NoError(t, err)
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

	// 400
	path = fmt.Sprintf("%s?hezEthereumAddress=hez:0x123456", endpoint)
	err = doBadReq("GET", path, nil, 400)
	require.NoError(t, err)

	// Test GetAccount
	path = fmt.Sprintf("%s/%s", endpoint, fetchedAccounts[2].Idx)
	account := testAccount{}
	require.NoError(t, doGoodReq("GET", path, nil, &account))
	account.Token.ItemID = 0
	assert.Equal(t, fetchedAccounts[2], account)

	// Test invalid token symbol GetAccount
	path = fmt.Sprintf("%s/%s", endpoint, "hez:UNI:258")
	account = testAccount{}
	require.Error(t, doGoodReq("GET", path, nil, &account))

	// 400
	path = fmt.Sprintf("%s/hez:12345", endpoint)
	err = doBadReq("GET", path, nil, 400)
	require.NoError(t, err)

	// 404
	path = fmt.Sprintf("%s/hez:10:12345", endpoint)
	err = doBadReq("GET", path, nil, 404)
	require.NoError(t, err)

	// 400
	path = fmt.Sprintf("%s?hez:hez:25641", endpoint)
	err = doBadReq("GET", path, nil, 400)
	require.NoError(t, err)

	// 400
	path = fmt.Sprintf("%s?hez:hez:0xb4A2333993a70fD103b7cC39883797Aa209bAa21", endpoint)
	err = doBadReq("GET", path, nil, 400)
	require.NoError(t, err)
}
