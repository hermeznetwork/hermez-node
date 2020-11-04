package api

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testL1Info struct {
	ToForgeL1TxsNum       *int64   `json:"toForgeL1TransactionsNum"`
	UserOrigin            bool     `json:"userOrigin"`
	LoadAmount            string   `json:"loadAmount"`
	HistoricLoadAmountUSD *float64 `json:"historicLoadAmountUSD"`
	EthBlockNum           int64    `json:"ethereumBlockNum"`
}

type testL2Info struct {
	Fee            common.FeeSelector `json:"fee"`
	HistoricFeeUSD *float64           `json:"historicFeeUSD"`
	Nonce          common.Nonce       `json:"nonce"`
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

type testTxsResponse struct {
	Txs        []testTx       `json:"transactions"`
	Pagination *db.Pagination `json:"pagination"`
}

func (t testTxsResponse) GetPagination() *db.Pagination {
	if t.Txs[0].ItemID < t.Txs[len(t.Txs)-1].ItemID {
		t.Pagination.FirstReturnedItem = t.Txs[0].ItemID
		t.Pagination.LastReturnedItem = t.Txs[len(t.Txs)-1].ItemID
	} else {
		t.Pagination.LastReturnedItem = t.Txs[0].ItemID
		t.Pagination.FirstReturnedItem = t.Txs[len(t.Txs)-1].ItemID
	}
	return t.Pagination
}

func (t testTxsResponse) Len() int {
	return len(t.Txs)
}

// TxSortFields represents the fields needed to sort L1 and L2 transactions
type txSortFields struct {
	BatchNum *common.BatchNum
	Position int
}

// TxSortFielder is a interface that allows sorting L1 and L2 transactions in a combined way
type txSortFielder interface {
	SortFields() txSortFields
	L1() *common.L1Tx
	L2() *common.L2Tx
}

// TxsSort array of TxSortFielder
type txsSort []txSortFielder

func (t txsSort) Len() int      { return len(t) }
func (t txsSort) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t txsSort) Less(i, j int) bool {
	// i not forged yet
	isf := t[i].SortFields()
	jsf := t[j].SortFields()
	if isf.BatchNum == nil {
		if jsf.BatchNum != nil { // j is already forged
			return false
		}
		// Both aren't forged, is i in a smaller position?
		return isf.Position < jsf.Position
	}
	// i is forged
	if jsf.BatchNum == nil {
		return false // j is not forged
	}
	// Both are forged
	if *isf.BatchNum == *jsf.BatchNum {
		// At the same batch,  is i in a smaller position?
		return isf.Position < jsf.Position
	}
	// At different batches, is i in a smaller batch?
	return *isf.BatchNum < *jsf.BatchNum
}

type wrappedL1 common.L1Tx

// SortFields implements TxSortFielder
func (tx *wrappedL1) SortFields() txSortFields {
	return txSortFields{
		BatchNum: tx.BatchNum,
		Position: tx.Position,
	}
}

// L1 implements TxSortFielder
func (tx *wrappedL1) L1() *common.L1Tx {
	l1tx := common.L1Tx(*tx)
	return &l1tx
}

// L2 implements TxSortFielder
func (tx *wrappedL1) L2() *common.L2Tx { return nil }

type wrappedL2 common.L2Tx

// SortFields implements TxSortFielder
func (tx *wrappedL2) SortFields() txSortFields {
	return txSortFields{
		BatchNum: &tx.BatchNum,
		Position: tx.Position,
	}
}

// L1 implements TxSortFielder
func (tx *wrappedL2) L1() *common.L1Tx { return nil }

// L2 implements TxSortFielder
func (tx *wrappedL2) L2() *common.L2Tx {
	l2tx := common.L2Tx(*tx)
	return &l2tx
}

func genTestTxs(
	genericTxs []txSortFielder,
	usrIdxs []string,
	accs []common.Account,
	tokens []historydb.TokenWithUSD,
	blocks []common.Block,
) (usrTxs []testTx, allTxs []testTx) {
	usrTxs = []testTx{}
	allTxs = []testTx{}
	isUsrTx := func(tx testTx) bool {
		for _, idx := range usrIdxs {
			if tx.FromIdx != nil && *tx.FromIdx == idx {
				return true
			}
			if tx.ToIdx == idx {
				return true
			}
		}
		return false
	}
	for _, genericTx := range genericTxs {
		l1 := genericTx.L1()
		l2 := genericTx.L2()
		if l1 != nil { // L1Tx to testTx
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
				ToIdx:       idxToHez(l1.ToIdx, token.Symbol),
				Amount:      l1.Amount.String(),
				BatchNum:    l1.BatchNum,
				Timestamp:   getTimestamp(l1.EthBlockNum, blocks),
				L1Info: &testL1Info{
					ToForgeL1TxsNum: l1.ToForgeL1TxsNum,
					UserOrigin:      l1.UserOrigin,
					LoadAmount:      l1.LoadAmount.String(),
					EthBlockNum:     l1.EthBlockNum,
				},
				Token: token,
			}
			// If FromIdx is not nil
			if l1.FromIdx != 0 {
				idxStr := idxToHez(l1.FromIdx, token.Symbol)
				tx.FromIdx = &idxStr
			}
			// If tx has a normal ToIdx (>255), set FromEthAddr and FromBJJ
			if l1.ToIdx >= common.UserThreshold {
				// find account
				for _, acc := range accs {
					if l1.ToIdx == acc.Idx {
						toEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
						tx.ToEthAddr = &toEthAddr
						toBJJ := string(apitypes.NewHezBJJ(acc.PublicKey))
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
				tx.HistoricUSD = &usd
				laf := new(big.Float).SetInt(l1.LoadAmount)
				loadAmountFloat, _ := laf.Float64()
				loadUSD := *token.USD * loadAmountFloat / math.Pow(10, float64(token.Decimals))
				tx.L1Info.HistoricLoadAmountUSD = &loadUSD
			}
			allTxs = append(allTxs, tx)
			if isUsrTx(tx) {
				usrTxs = append(usrTxs, tx)
			}
		} else { // L2Tx to testTx
			token := getTokenByIdx(l2.FromIdx, tokens, accs)
			// l1.FromIdx can't be nil
			fromIdx := idxToHez(l2.FromIdx, token.Symbol)
			tx := testTx{
				IsL1:      "L2",
				TxID:      l2.TxID,
				Type:      l2.Type,
				Position:  l2.Position,
				ToIdx:     idxToHez(l2.ToIdx, token.Symbol),
				FromIdx:   &fromIdx,
				Amount:    l2.Amount.String(),
				BatchNum:  &l2.BatchNum,
				Timestamp: getTimestamp(l2.EthBlockNum, blocks),
				L2Info: &testL2Info{
					Nonce: l2.Nonce,
					Fee:   l2.Fee,
				},
				Token: token,
			}
			// If FromIdx is not nil
			if l2.FromIdx != 0 {
				idxStr := idxToHez(l2.FromIdx, token.Symbol)
				tx.FromIdx = &idxStr
			}
			// Set FromEthAddr and FromBJJ (FromIdx it's always >255)
			for _, acc := range accs {
				if l2.ToIdx == acc.Idx {
					fromEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
					tx.FromEthAddr = &fromEthAddr
					fromBJJ := string(apitypes.NewHezBJJ(acc.PublicKey))
					tx.FromBJJ = &fromBJJ
					break
				}
			}
			// If tx has a normal ToIdx (>255), set FromEthAddr and FromBJJ
			if l2.ToIdx >= common.UserThreshold {
				// find account
				for _, acc := range accs {
					if l2.ToIdx == acc.Idx {
						toEthAddr := string(apitypes.NewHezEthAddr(acc.EthAddr))
						tx.ToEthAddr = &toEthAddr
						toBJJ := string(apitypes.NewHezBJJ(acc.PublicKey))
						tx.ToBJJ = &toBJJ
						break
					}
				}
			}
			// If the token has USD value setted
			if token.USD != nil {
				af := new(big.Float).SetInt(l2.Amount)
				amountFloat, _ := af.Float64()
				usd := *token.USD * amountFloat / math.Pow(10, float64(token.Decimals))
				tx.HistoricUSD = &usd
				feeUSD := usd * l2.Fee.Percentage()
				tx.HistoricUSD = &usd
				tx.L2Info.HistoricFeeUSD = &feeUSD
			}
			allTxs = append(allTxs, tx)
			if isUsrTx(tx) {
				usrTxs = append(usrTxs, tx)
			}
		}
	}
	return usrTxs, allTxs
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
	// Get all (no filters)
	limit := 8
	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	assertTxs(t, tc.allTxs, fetchedTxs)
	// Uncomment once tx generation for tests is fixed
	// // Get by ethAddr
	// fetchedTxs = []testTx{}
	// limit = 7
	// path = fmt.Sprintf(
	// 	"%s?hermezEthereumAddress=%s&limit=%d&fromItem=",
	// 	endpoint, tc.usrAddr, limit,
	// )
	// err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	// assert.NoError(t, err)
	// assertTxs(t, tc.usrTxs, fetchedTxs)
	// // Get by bjj
	// fetchedTxs = []testTx{}
	// limit = 6
	// path = fmt.Sprintf(
	// 	"%s?BJJ=%s&limit=%d&fromItem=",
	// 	endpoint, tc.usrBjj, limit,
	// )
	// err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	// assert.NoError(t, err)
	// assertTxs(t, tc.usrTxs, fetchedTxs)
	// Get by tokenID
	fetchedTxs = []testTx{}
	limit = 5
	tokenID := tc.allTxs[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d&fromItem=",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	tokenIDTxs := []testTx{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].Token.TokenID == tokenID {
			tokenIDTxs = append(tokenIDTxs, tc.allTxs[i])
		}
	}
	assertTxs(t, tokenIDTxs, fetchedTxs)
	// // idx
	// fetchedTxs = []testTx{}
	// limit = 4
	idx := tc.allTxs[0].ToIdx
	// path = fmt.Sprintf(
	// 	"%s?accountIndex=%s&limit=%d&fromItem=",
	// 	endpoint, idx, limit,
	// )
	// err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	// assert.NoError(t, err)
	// idxTxs := []testTx{}
	// for i := 0; i < len(tc.allTxs); i++ {
	// 	if (tc.allTxs[i].FromIdx != nil && (*tc.allTxs[i].FromIdx)[6:] == idx[6:]) ||
	// 		tc.allTxs[i].ToIdx[6:] == idx[6:] {
	// 		idxTxs = append(idxTxs, tc.allTxs[i])
	// 	}
	// }
	// assertHistoryTxAPIs(t, idxTxs, fetchedTxs)
	// batchNum
	fetchedTxs = []testTx{}
	limit = 3
	batchNum := tc.allTxs[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d&fromItem=",
		endpoint, *batchNum, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	batchNumTxs := []testTx{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil &&
			*tc.allTxs[i].BatchNum == *batchNum {
			batchNumTxs = append(batchNumTxs, tc.allTxs[i])
		}
	}
	assertTxs(t, batchNumTxs, fetchedTxs)
	// type
	txTypes := []common.TxType{
		// Uncomment once test gen is fixed
		// common.TxTypeExit,
		// common.TxTypeTransfer,
		// common.TxTypeDeposit,
		common.TxTypeCreateAccountDeposit,
		// common.TxTypeCreateAccountDepositTransfer,
		// common.TxTypeDepositTransfer,
		common.TxTypeForceTransfer,
		// common.TxTypeForceExit,
		// common.TxTypeTransferToEthAddr,
		// common.TxTypeTransferToBJJ,
	}
	for _, txType := range txTypes {
		fetchedTxs = []testTx{}
		limit = 2
		path = fmt.Sprintf(
			"%s?type=%s&limit=%d&fromItem=",
			endpoint, txType, limit,
		)
		err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
		assert.NoError(t, err)
		txTypeTxs := []testTx{}
		for i := 0; i < len(tc.allTxs); i++ {
			if tc.allTxs[i].Type == txType {
				txTypeTxs = append(txTypeTxs, tc.allTxs[i])
			}
		}
		assertTxs(t, txTypeTxs, fetchedTxs)
	}
	// Multiple filters
	fetchedTxs = []testTx{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokenId=%d&limit=%d&fromItem=",
		endpoint, *batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	mixedTxs := []testTx{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil {
			if *tc.allTxs[i].BatchNum == *batchNum && tc.allTxs[i].Token.TokenID == tokenID {
				mixedTxs = append(mixedTxs, tc.allTxs[i])
			}
		}
	}
	assertTxs(t, mixedTxs, fetchedTxs)
	// All, in reverse order
	fetchedTxs = []testTx{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderDesc, &testTxsResponse{}, appendIter)
	assert.NoError(t, err)
	flipedTxs := []testTx{}
	for i := 0; i < len(tc.allTxs); i++ {
		flipedTxs = append(flipedTxs, tc.allTxs[len(tc.allTxs)-1-i])
	}
	assertTxs(t, flipedTxs, fetchedTxs)
	// 400
	path = fmt.Sprintf(
		"%s?accountIndex=%s&hermezEthereumAddress=%s",
		endpoint, idx, tc.usrAddr,
	)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	path = fmt.Sprintf("%s?tokenId=X", endpoint)
	err = doBadReq("GET", path, nil, 400)
	assert.NoError(t, err)
	// 404
	path = fmt.Sprintf("%s?batchNum=999999", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
	path = fmt.Sprintf("%s?limit=1000&fromItem=999999", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}

func TestGetHistoryTx(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "transactions-history/"
	fetchedTxs := []testTx{}
	for _, tx := range tc.allTxs {
		fetchedTx := testTx{}
		err := doGoodReq("GET", endpoint+tx.TxID.String(), nil, &fetchedTx)
		assert.NoError(t, err)
		fetchedTxs = append(fetchedTxs, fetchedTx)
	}
	assertTxs(t, tc.allTxs, fetchedTxs)
	// 400
	err := doBadReq("GET", endpoint+"0x001", nil, 400)
	assert.NoError(t, err)
	// 404
	err = doBadReq("GET", endpoint+"0x00000000000001e240004700", nil, 404)
	assert.NoError(t, err)
}

func assertTxs(t *testing.T, expected, actual []testTx) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		actual[i].ItemID = 0
		actual[i].Token.ItemID = 0
		assert.Equal(t, expected[i].Timestamp.Unix(), actual[i].Timestamp.Unix())
		expected[i].Timestamp = actual[i].Timestamp
		if expected[i].Token.USDUpdate == nil {
			assert.Equal(t, expected[i].Token.USDUpdate, actual[i].Token.USDUpdate)
		} else {
			assert.Equal(t, expected[i].Token.USDUpdate.Unix(), actual[i].Token.USDUpdate.Unix())
			expected[i].Token.USDUpdate = actual[i].Token.USDUpdate
		}
		test.AssertUSD(t, expected[i].HistoricUSD, actual[i].HistoricUSD)
		if expected[i].L2Info != nil {
			test.AssertUSD(t, expected[i].L2Info.HistoricFeeUSD, actual[i].L2Info.HistoricFeeUSD)
		} else {
			test.AssertUSD(t, expected[i].L1Info.HistoricLoadAmountUSD, actual[i].L1Info.HistoricLoadAmountUSD)
		}
		assert.Equal(t, expected[i], actual[i])
	}
}
