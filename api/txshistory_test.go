package api

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testL1Info struct {
	ToForgeL1TxsNum          *int64   `json:"toForgeL1TransactionsNum"`
	UserOrigin               bool     `json:"userOrigin"`
	DepositAmount            string   `json:"depositAmount"`
	AmountSuccess            bool     `json:"amountSuccess"`
	DepositAmountSuccess     bool     `json:"depositAmountSuccess"`
	HistoricDepositAmountUSD *float64 `json:"historicDepositAmountUSD"`
	EthBlockNum              int64    `json:"ethereumBlockNum"`
}

type testL2Info struct {
	Fee            common.FeeSelector `json:"fee"`
	HistoricFeeUSD *float64           `json:"historicFeeUSD"`
	Nonce          nonce.Nonce        `json:"nonce"`
}

type testTx struct {
	IsL1        string                 `json:"L1orL2"`
	TxID        common.TxID            `json:"id"`
	ItemID      uint64                 `json:"itemId"`
	Type        common.TxType          `json:"type"`
	Position    int                    `json:"position"`
	FromIdx     *string                `json:"fromAccountIndex"`
	FromEthAddr *string                `json:"fromHezEthereumAddress"`
	FromBJJ     *string                `json:"fromBJJ"`
	ToIdx       string                 `json:"toAccountIndex"`
	ToEthAddr   *string                `json:"toHezEthereumAddress"`
	ToBJJ       *string                `json:"toBJJ"`
	Amount      string                 `json:"amount"`
	BatchNum    *common.BatchNum       `json:"batchNum"`
	HistoricUSD *float64               `json:"historicUSD"`
	Timestamp   time.Time              `json:"timestamp"`
	L1Info      *testL1Info            `json:"L1Info"`
	L2Info      *testL2Info            `json:"L2Info"`
	Token       historydb.TokenWithUSD `json:"token"`
}

type txsSort []testTx

func (t txsSort) Len() int      { return len(t) }
func (t txsSort) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t txsSort) Less(i, j int) bool {
	// i not forged yet
	isf := t[i]
	jsf := t[j]
	if isf.BatchNum == nil {
		if jsf.BatchNum != nil { // j is already forged
			return false
		}
		// Both aren't forged, is i in a smaller position?
		return isf.Position < jsf.Position
	}
	// i is forged
	if jsf.BatchNum == nil {
		return true // j is not forged
	}
	// Both are forged
	if *isf.BatchNum == *jsf.BatchNum {
		// At the same batch,  is i in a smaller position?
		return isf.Position < jsf.Position
	}
	// At different batches, is i in a smaller batch?
	return *isf.BatchNum < *jsf.BatchNum
}

type testTxsResponse struct {
	Txs          []testTx `json:"transactions"`
	PendingItems uint64   `json:"pendingItems"`
}

func (t testTxsResponse) GetPending() (pendingItems, lastItemID uint64) {
	if len(t.Txs) == 0 {
		return 0, 0
	}
	pendingItems = t.PendingItems
	lastItemID = t.Txs[len(t.Txs)-1].ItemID
	return pendingItems, lastItemID
}

func (t testTxsResponse) Len() int {
	return len(t.Txs)
}

func (t testTxsResponse) New() Pendinger { return &testTxsResponse{} }

func genTestTxs(
	l1s []common.L1Tx,
	l2s []common.L2Tx,
	accs []common.Account,
	tokens []historydb.TokenWithUSD,
	blocks []common.Block,
) []testTx {
	txs := []testTx{}
	// common.L1Tx ==> testTx
	for i, l1 := range l1s {
		token := getTokenByID(l1.TokenID, tokens)
		// l1.FromEthAddr and l1.FromBJJ can't be nil
		fromEthAddr := string(apitypes.NewHezEthAddr(l1.FromEthAddr))
		fromBJJ := string(apitypes.NewHezBJJ(l1.FromBJJ))
		tx := testTx{
			IsL1:        "L1",
			TxID:        l1.TxID,
			Type:        l1.Type,
			Position:    l1.Position,
			FromEthAddr: &fromEthAddr,
			FromBJJ:     &fromBJJ,
			ToIdx:       common.IdxToHez(l1.ToIdx, token.Symbol),
			Amount:      l1.Amount.String(),
			BatchNum:    l1.BatchNum,
			Timestamp:   getTimestamp(l1.EthBlockNum, blocks),
			L1Info: &testL1Info{
				ToForgeL1TxsNum:      l1.ToForgeL1TxsNum,
				UserOrigin:           l1.UserOrigin,
				DepositAmount:        l1.DepositAmount.String(),
				AmountSuccess:        true,
				DepositAmountSuccess: true,
				EthBlockNum:          l1.EthBlockNum,
			},
			Token: token,
		}

		// set BatchNum for user txs
		if tx.L1Info.ToForgeL1TxsNum != nil {
			// WARNING: this works just because the way "common" txs are generated using til
			// any change on the test set could break this
			bn := common.BatchNum(*tx.L1Info.ToForgeL1TxsNum + 2)
			tx.BatchNum = &bn
		}
		// If FromIdx is not nil
		idxStr := common.IdxToHez(l1.EffectiveFromIdx, token.Symbol)
		tx.FromIdx = &idxStr
		if i == len(l1s)-1 {
			// Last tx of the L1 set is supposed to be unforged as per the til set.
			// Unforged txs have some special propperties
			tx.L1Info.DepositAmountSuccess = false
			tx.L1Info.AmountSuccess = false
			tx.BatchNum = nil
			idxStrUnforged := common.IdxToHez(l1.FromIdx, token.Symbol)
			tx.FromIdx = &idxStrUnforged
		}
		// If tx has a normal ToIdx (>255), set FromEthAddr and FromBJJ
		if l1.ToIdx >= common.UserThreshold {
			// find account
			for _, acc := range accs {
				if l1.ToIdx == acc.Idx {
					toEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
					tx.ToEthAddr = &toEthAddr
					toBJJ := string(apitypes.NewHezBJJ(acc.BJJ))
					tx.ToBJJ = &toBJJ
					break
				}
			}
		}
		// If the token has USD value setted
		if token.USD != nil {
			af := new(big.Float).SetInt(l1.Amount)
			amountFloat, _ := af.Float64()
			usd := *token.USD * amountFloat / math.Pow(10, float64(token.Decimals))
			if usd != 0 {
				tx.HistoricUSD = &usd
			}
			laf := new(big.Float).SetInt(l1.DepositAmount)
			depositAmountFloat, _ := laf.Float64()
			depositUSD := *token.USD * depositAmountFloat / math.Pow(10, float64(token.Decimals))
			if depositAmountFloat != 0 {
				tx.L1Info.HistoricDepositAmountUSD = &depositUSD
			}
		}
		txs = append(txs, tx)
	}

	// common.L2Tx ==> testTx
	for i := 0; i < len(l2s); i++ {
		token := getTokenByIdx(l2s[i].FromIdx, tokens, accs)
		// l1.FromIdx can't be nil
		fromIdx := common.IdxToHez(l2s[i].FromIdx, token.Symbol)
		tx := testTx{
			IsL1:      "L2",
			TxID:      l2s[i].TxID,
			Type:      l2s[i].Type,
			Position:  l2s[i].Position,
			ToIdx:     common.IdxToHez(l2s[i].ToIdx, token.Symbol),
			FromIdx:   &fromIdx,
			Amount:    l2s[i].Amount.String(),
			BatchNum:  &l2s[i].BatchNum,
			Timestamp: getTimestamp(l2s[i].EthBlockNum, blocks),
			L2Info: &testL2Info{
				Nonce: l2s[i].Nonce,
				Fee:   l2s[i].Fee,
			},
			Token: token,
		}
		// If FromIdx is not nil
		if l2s[i].FromIdx != 0 {
			idxStr := common.IdxToHez(l2s[i].FromIdx, token.Symbol)
			tx.FromIdx = &idxStr
		}
		// Set FromEthAddr and FromBJJ (FromIdx it's always >255)
		for _, acc := range accs {
			if l2s[i].FromIdx == acc.Idx {
				fromEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
				tx.FromEthAddr = &fromEthAddr
				fromBJJ := string(apitypes.NewHezBJJ(acc.BJJ))
				tx.FromBJJ = &fromBJJ
				break
			}
		}
		// If tx has a normal ToIdx (>255), set FromEthAddr and FromBJJ
		if l2s[i].ToIdx >= common.UserThreshold {
			// find account
			for _, acc := range accs {
				if l2s[i].ToIdx == acc.Idx {
					toEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
					tx.ToEthAddr = &toEthAddr
					toBJJ := string(apitypes.NewHezBJJ(acc.BJJ))
					tx.ToBJJ = &toBJJ
					break
				}
			}
		}
		// If the token has USD value setted
		if token.USD != nil {
			af := new(big.Float).SetInt(l2s[i].Amount)
			amountFloat, _ := af.Float64()
			usd := *token.USD * amountFloat / math.Pow(10, float64(token.Decimals))
			if usd != 0 {
				tx.HistoricUSD = &usd
				feeUSD := usd * l2s[i].Fee.Percentage()
				if feeUSD != 0 {
					tx.L2Info.HistoricFeeUSD = &feeUSD
				}
			}
		}
		txs = append(txs, tx)
	}

	// Sort txs
	sortedTxs := txsSort(txs)
	sort.Sort(sortedTxs)

	return []testTx(sortedTxs)
}

func TestGetHistoryTxs(t *testing.T) {
	endpoint := apiURL + "transactions-history"
	fetchedTxs := []testTx{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*testTxsResponse).Txs); i++ {
			tmp, err := copystructure.Copy(intr.(*testTxsResponse).Txs[i])
			if err != nil {
				panic(err)
			}
			fetchedTxs = append(fetchedTxs, tmp.(testTx))
		}
	}
	// Get all (no filters, excluding unforged txs)
	limit := 20
	path := fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err := doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	forgedTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[i].BatchNum != nil {
			forgedTxs = append(forgedTxs, tc.txs[i])
		}
	}
	assertTxs(t, forgedTxs, fetchedTxs)

	// Get all, including unforged txs
	fetchedTxs = []testTx{}
	path = fmt.Sprintf("%s?limit=%d&includePendingL1s=true", endpoint, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	assertTxs(t, tc.txs, fetchedTxs)

	// Get by ethAddr
	account := tc.accounts[2]
	fetchedTxs = []testTx{}
	limit = 7
	path = fmt.Sprintf(
		"%s?hezEthereumAddress=%s&limit=%d",
		endpoint, account.EthAddr, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	accountTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		tx := tc.txs[i]
		if (tx.FromIdx != nil && *tx.FromIdx == string(account.Idx)) ||
			tx.ToIdx == string(account.Idx) ||
			(tx.FromEthAddr != nil && *tx.FromEthAddr == string(account.EthAddr)) ||
			(tx.ToEthAddr != nil && *tx.ToEthAddr == string(account.EthAddr)) ||
			(tx.FromBJJ != nil && *tx.FromBJJ == string(account.PublicKey)) ||
			(tx.ToBJJ != nil && *tx.ToBJJ == string(account.PublicKey)) && tx.BatchNum != nil {
			accountTxs = append(accountTxs, tx)
		}
	}
	assertTxs(t, accountTxs, fetchedTxs)
	// Get by bjj
	fetchedTxs = []testTx{}
	limit = 6
	path = fmt.Sprintf(
		"%s?BJJ=%s&limit=%d",
		endpoint, account.PublicKey, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	assertTxs(t, accountTxs, fetchedTxs)
	// Get by tokenID
	fetchedTxs = []testTx{}
	limit = 5
	tokenID := tc.txs[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	tokenIDTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[i].BatchNum != nil && tc.txs[i].Token.TokenID == tokenID {
			tokenIDTxs = append(tokenIDTxs, tc.txs[i])
		}
	}
	assertTxs(t, tokenIDTxs, fetchedTxs)
	// idx
	fetchedTxs = []testTx{}
	limit = 4
	idxStr := tc.txs[0].ToIdx
	queryAccount, err := common.StringToIdx(idxStr, "")
	assert.NoError(t, err)
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d",
		endpoint, idxStr, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	idxTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[i].BatchNum == nil {
			continue
		}
		var fromQueryAccount common.QueryAccount
		if tc.txs[i].FromIdx != nil {
			fromQueryAccount, err = common.StringToIdx(*tc.txs[i].FromIdx, "")
			assert.NoError(t, err)
			if *fromQueryAccount.AccountIndex == *queryAccount.AccountIndex {
				idxTxs = append(idxTxs, tc.txs[i])
				continue
			}
		}
		toQueryAccount, err := common.StringToIdx(tc.txs[i].ToIdx, "")
		assert.NoError(t, err)
		if *toQueryAccount.AccountIndex == *queryAccount.AccountIndex {
			idxTxs = append(idxTxs, tc.txs[i])
		}
	}
	assertTxs(t, idxTxs, fetchedTxs)
	// from idx
	fetchedTxs = []testTx{}
	idxTxs = []testTx{}
	path = fmt.Sprintf("%s?fromAccountIndex=%s&limit=%d", endpoint, idxStr, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	for i := 0; i < len(tc.txs); i++ {
		var fromQueryAccount common.QueryAccount
		if tc.txs[i].FromIdx != nil {
			fromQueryAccount, err = common.StringToIdx(*tc.txs[i].FromIdx, "")
			assert.NoError(t, err)
			if *fromQueryAccount.AccountIndex == *queryAccount.AccountIndex {
				idxTxs = append(idxTxs, tc.txs[i])
				continue
			}
		}
	}
	assertTxs(t, idxTxs, fetchedTxs)
	// to idx
	fetchedTxs = []testTx{}
	path = fmt.Sprintf("%s?toAccountIndex=%s&limit=%d", endpoint, idxStr, limit)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	idxTxs = []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		toQueryAccount, err := common.StringToIdx(tc.txs[i].ToIdx, "")
		assert.NoError(t, err)
		if *toQueryAccount.AccountIndex == *queryAccount.AccountIndex {
			idxTxs = append(idxTxs, tc.txs[i])
		}
	}
	assertTxs(t, idxTxs, fetchedTxs)
	// batchNum
	fetchedTxs = []testTx{}
	limit = 3
	batchNum := tc.txs[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d",
		endpoint, *batchNum, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	batchNumTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[i].BatchNum != nil &&
			*tc.txs[i].BatchNum == *batchNum {
			batchNumTxs = append(batchNumTxs, tc.txs[i])
		}
	}
	assertTxs(t, batchNumTxs, fetchedTxs)
	// type
	txTypes := []common.TxType{
		// Uncomment once test gen is fixed
		common.TxTypeExit,
		common.TxTypeTransfer,
		common.TxTypeDeposit,
		common.TxTypeCreateAccountDeposit,
		common.TxTypeCreateAccountDepositTransfer,
		common.TxTypeDepositTransfer,
		common.TxTypeForceTransfer,
		common.TxTypeForceExit,
	}
	for _, txType := range txTypes {
		fetchedTxs = []testTx{}
		limit = 2
		path = fmt.Sprintf(
			"%s?type=%s&limit=%d",
			endpoint, txType, limit,
		)
		err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
		assert.NoError(t, err)
		txTypeTxs := []testTx{}
		for i := 0; i < len(tc.txs); i++ {
			if tc.txs[i].Type == txType && tc.txs[i].BatchNum != nil {
				txTypeTxs = append(txTypeTxs, tc.txs[i])
			}
		}
		assertTxs(t, txTypeTxs, fetchedTxs)
	}
	// Multiple filters
	fetchedTxs = []testTx{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokenId=%d&limit=%d",
		endpoint, *batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, db.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	mixedTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[i].BatchNum != nil {
			if *tc.txs[i].BatchNum == *batchNum && tc.txs[i].Token.TokenID == tokenID {
				mixedTxs = append(mixedTxs, tc.txs[i])
			}
		}
	}
	assertTxs(t, mixedTxs, fetchedTxs)
	// All, in reverse order
	fetchedTxs = []testTx{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	err = doGoodReqPaginated(path, db.OrderDesc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	flipedTxs := []testTx{}
	for i := 0; i < len(tc.txs); i++ {
		if tc.txs[len(tc.txs)-1-i].BatchNum != nil {
			flipedTxs = append(flipedTxs, tc.txs[len(tc.txs)-1-i])
		}
	}
	assertTxs(t, flipedTxs, fetchedTxs)
	// Empty array
	fetchedTxs = []testTx{}
	path = fmt.Sprintf("%s?batchNum=999999", endpoint)
	err = doGoodReqPaginated(path, db.OrderDesc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	assertTxs(t, []testTx{}, fetchedTxs)
	// 400
	path = fmt.Sprintf(
		"%s?accountIndex=%s&hezEthereumAddress=%s",
		endpoint, queryAccount.AccountIndex, account.EthAddr,
	)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	path = fmt.Sprintf("%s?tokenId=X", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
}

func TestGetHistoryTx(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "transactions-history/"
	fetchedTxs := []testTx{}
	for _, tx := range tc.txs {
		fetchedTx := testTx{}
		err := doGoodReq("GET", endpoint+tx.TxID.String(), nil, &fetchedTx)
		assert.NoError(t, err)
		fetchedTxs = append(fetchedTxs, fetchedTx)
	}
	assertTxs(t, tc.txs, fetchedTxs)
	// 400, due invalid TxID
	err := doBadReq("GET", endpoint+"0x001", nil, 400)
	assert.NoError(t, err)
	// 404, due nonexistent TxID in DB
	err = doBadReq("GET", endpoint+"0x00eb5e95e1ce5e9f6c4ed402d415e8d0bdd7664769cfd2064d28da04a2c76be432", nil, 404)
	assert.NoError(t, err)
}

func assertTxs(t *testing.T, expected, actual []testTx) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		assert.Equal(t, expected[i].BatchNum, actual[i].BatchNum)
		assert.Equal(t, expected[i].Position, actual[i].Position)
		actual[i].ItemID = 0
		actual[i].Token.ItemID = 0
		assert.Equal(t, expected[i].Timestamp.Unix(), actual[i].Timestamp.Unix())
		expected[i].Timestamp = actual[i].Timestamp
		if expected[i].Token.USDUpdate == nil {
			assert.Equal(t, expected[i].Token.USDUpdate, actual[i].Token.USDUpdate)
			expected[i].Token.USDUpdate = actual[i].Token.USDUpdate
		} else {
			assert.Equal(t, expected[i].Token.USDUpdate.Unix(), actual[i].Token.USDUpdate.Unix())
			expected[i].Token.USDUpdate = actual[i].Token.USDUpdate
		}
		test.AssertUSD(t, expected[i].HistoricUSD, actual[i].HistoricUSD)
		if expected[i].L2Info != nil {
			test.AssertUSD(t, expected[i].L2Info.HistoricFeeUSD, actual[i].L2Info.HistoricFeeUSD)
		} else {
			test.AssertUSD(t, expected[i].L1Info.HistoricDepositAmountUSD, actual[i].L1Info.HistoricDepositAmountUSD)
		}
		assert.Equal(t, expected[i], actual[i])
	}
}
