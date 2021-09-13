package api

import (
	"fmt"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testCVP struct {
	Root     string
	Siblings []string
	OldKey   string
	OldValue string
	IsOld0   bool
	Key      string
	Value    string
	Fnc      int
}

type testExit struct {
	ItemID                 uint64                 `json:"itemId"`
	BatchNum               common.BatchNum        `json:"batchNum"`
	AccountIdx             string                 `json:"accountIndex"`
	BJJ                    apitypes.HezBJJ        `json:"bjj"`
	EthAddr                apitypes.HezEthAddr    `json:"hezEthereumAddress"`
	MerkleProof            testCVP                `json:"merkleProof"`
	Balance                string                 `json:"balance"`
	InstantWithdrawn       *int64                 `json:"instantWithdraw"`
	DelayedWithdrawRequest *int64                 `json:"delayedWithdrawRequest"`
	DelayedWithdrawn       *int64                 `json:"delayedWithdraw"`
	Token                  historydb.TokenWithUSD `json:"token"`
}

type testExitsResponse struct {
	Exits        []testExit `json:"exits"`
	PendingItems uint64     `json:"pendingItems"`
}

func (t testExitsResponse) GetPending() (pendingItems, lastItemID uint64) {
	if len(t.Exits) == 0 {
		return 0, 0
	}
	pendingItems = t.PendingItems
	lastItemID = t.Exits[len(t.Exits)-1].ItemID
	return pendingItems, lastItemID
}

func (t testExitsResponse) New() Pendinger { return &testExitsResponse{} }

func (t *testExitsResponse) Len() int {
	return len(t.Exits)
}

func genTestExits(
	commonExits []common.ExitInfo,
	tokens []historydb.TokenWithUSD,
	accs []common.Account,
) []testExit {
	allExits := []testExit{}
	for _, exit := range commonExits {
		token := getTokenByIdx(exit.AccountIdx, tokens, accs)
		siblings := []string{}
		for i := 0; i < len(exit.MerkleProof.Siblings); i++ {
			siblings = append(siblings, exit.MerkleProof.Siblings[i].String())
		}
		acc := getAccountByIdx(exit.AccountIdx, accs)
		allExits = append(allExits, testExit{
			BatchNum:   exit.BatchNum,
			AccountIdx: common.IdxToHez(exit.AccountIdx, token.Symbol),
			BJJ:        apitypes.NewHezBJJ(acc.BJJ),
			EthAddr:    apitypes.NewHezEthAddr(acc.EthAddr),
			MerkleProof: testCVP{
				Root:     exit.MerkleProof.Root.String(),
				Siblings: siblings,
				OldKey:   exit.MerkleProof.OldKey.String(),
				OldValue: exit.MerkleProof.OldValue.String(),
				IsOld0:   exit.MerkleProof.IsOld0,
				Key:      exit.MerkleProof.Key.String(),
				Value:    exit.MerkleProof.Value.String(),
				Fnc:      exit.MerkleProof.Fnc,
			},
			Balance:                exit.Balance.String(),
			InstantWithdrawn:       exit.InstantWithdrawn,
			DelayedWithdrawRequest: exit.DelayedWithdrawRequest,
			DelayedWithdrawn:       exit.DelayedWithdrawn,
			Token:                  token,
		})
	}
	return allExits
}

func TestGetExits(t *testing.T) {
	endpoint := apiURL + "exits"
	fetchedExits := []testExit{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testExitsResponse).Exits); i++ {
			tmp, err := copystructure.Copy(intr.(*testExitsResponse).Exits[i])
			if err != nil {
				panic(err)
			}
			fetchedExits = append(fetchedExits, tmp.(testExit))
		}
	}
	// Get all (no filters)
	limit := 8
	path := fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err := doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, tc.exits, fetchedExits)

	// Get by ethAddr
	fetchedExits = []testExit{}
	limit = 7
	var account testAccount
	for _, tx := range tc.txs {
		found := false
		if tx.Type == common.TxTypeExit {
			for i := 0; i < len(tc.accounts); i++ {
				if tx.FromIdx != nil && string(tc.accounts[i].Idx) == *tx.FromIdx {
					account = tc.accounts[i]
					break
				}
			}
		}
		if found {
			break
		}
	}
	path = fmt.Sprintf(
		"%s?hezEthereumAddress=%s&limit=%d",
		endpoint, account.EthAddr, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	var accountExits []testExit
	for i := range tc.exits {
		for _, acc := range tc.accounts {
			if string(acc.Idx) == tc.exits[i].AccountIdx {
				if acc.EthAddr == account.EthAddr {
					accountExits = append(accountExits, tc.exits[i])
				}
			}
		}
	}
	assertExitAPIs(t, accountExits, fetchedExits)
	// Get by bjj
	fetchedExits = []testExit{}
	limit = 6
	path = fmt.Sprintf(
		"%s?BJJ=%s&limit=%d",
		endpoint, account.PublicKey, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, accountExits, fetchedExits)
	// Get by tokenID
	fetchedExits = []testExit{}
	limit = 5
	tokenID := tc.exits[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	tokenIDExits := []testExit{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].Token.TokenID == tokenID {
			tokenIDExits = append(tokenIDExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, tokenIDExits, fetchedExits)
	// idx
	fetchedExits = []testExit{}
	limit = 4
	idx := tc.exits[0].AccountIdx
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d",
		endpoint, idx, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	idxExits := []testExit{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].AccountIdx[6:] == idx[6:] {
			idxExits = append(idxExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, idxExits, fetchedExits)
	// batchNum
	fetchedExits = []testExit{}
	limit = 3
	batchNum := tc.exits[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d",
		endpoint, batchNum, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	batchNumExits := []testExit{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].BatchNum == batchNum {
			batchNumExits = append(batchNumExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, batchNumExits, fetchedExits)
	// OnlyPendingWithdraws
	fetchedExits = []testExit{}
	limit = 7
	path = fmt.Sprintf(
		"%s?&onlyPendingWithdraws=%t&limit=%d",
		endpoint, true, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	pendingExits := []testExit{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].InstantWithdrawn == nil && tc.exits[i].DelayedWithdrawn == nil {
			pendingExits = append(pendingExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, pendingExits, fetchedExits)
	// Multiple filters
	fetchedExits = []testExit{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokeId=%d&limit=%d",
		endpoint, batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	mixedExits := []testExit{}
	flipedExits := []testExit{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].BatchNum == batchNum && tc.exits[i].Token.TokenID == tokenID {
			mixedExits = append(mixedExits, tc.exits[i])
		}
		flipedExits = append(flipedExits, tc.exits[len(tc.exits)-1-i])
	}
	assertExitAPIs(t, mixedExits, fetchedExits)
	// All, in reverse order
	fetchedExits = []testExit{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err = doGoodReqPaginated(path, db.OrderDesc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, flipedExits, fetchedExits)
	// Empty array
	fetchedExits = []testExit{}
	path = fmt.Sprintf("%s?batchNum=999999", endpoint)
	err = doGoodReqPaginated(path, db.OrderDesc, &testExitsResponse{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, []testExit{}, fetchedExits)
	// 400
	path = fmt.Sprintf(
		"%s?accountIndex=%s&hezEthereumAddress=%s",
		endpoint, idx, account.EthAddr,
	)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	path = fmt.Sprintf("%s?tokenId=X", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
}

func TestGetExit(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "exits/"
	fetchedExits := []testExit{}
	for _, exit := range tc.exits {
		fetchedExit := testExit{}
		assert.NoError(
			t, doGoodReq(
				"GET",
				fmt.Sprintf("%s%d/%s", endpoint, exit.BatchNum, exit.AccountIdx),
				nil, &fetchedExit,
			),
		)
		fetchedExits = append(fetchedExits, fetchedExit)
	}
	assertExitAPIs(t, tc.exits, fetchedExits)
	// 400
	err := doBadReq("GET", endpoint+"1/haz:BOOM:1", nil, 400)
	assert.NoError(t, err)
	err = doBadReq("GET", endpoint+"-1/hez:BOOM:1", nil, 400)
	assert.NoError(t, err)
	// 404
	err = doBadReq("GET", endpoint+"494/hez:XXX:1", nil, 404)
	assert.NoError(t, err)
}

func assertExitAPIs(t *testing.T, expected, actual []testExit) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		actual[i].ItemID = 0
		actual[i].Token.ItemID = 0
		if expected[i].Token.USDUpdate == nil {
			assert.Equal(t, expected[i].Token.USDUpdate, actual[i].Token.USDUpdate)
		} else {
			assert.Equal(t, expected[i].Token.USDUpdate.Unix(), actual[i].Token.USDUpdate.Unix())
			expected[i].Token.USDUpdate = actual[i].Token.USDUpdate
		}
		assert.Equal(t, expected[i], actual[i])
	}
}
