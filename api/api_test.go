package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	swagger "github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/mitchellh/copystructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const apiPort = ":4010"
const apiURL = "http://localhost" + apiPort + "/"

type testCommon struct {
	blocks   []common.Block
	tokens   []historydb.TokenRead
	batches  []common.Batch
	usrAddr  string
	usrBjj   string
	accs     []common.Account
	usrTxs   []historyTxAPI
	allTxs   []historyTxAPI
	exits    []exitAPI
	usrExits []exitAPI
	router   *swagger.Router
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

var tc testCommon

func TestMain(m *testing.M) {
	// Init swagger
	router := swagger.NewRouter().WithSwaggerFromFile("./swagger.yml")
	// Init DBs
	// HistoryDB
	pass := os.Getenv("POSTGRES_PASS")
	database, err := db.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	hdb := historydb.NewHistoryDB(database)
	err = hdb.Reorg(-1)
	if err != nil {
		panic(err)
	}
	// StateDB
	dir, err := ioutil.TempDir("", "tmpdb")
	if err != nil {
		panic(err)
	}
	sdb, err := statedb.NewStateDB(dir, statedb.TypeTxSelector, 0)
	if err != nil {
		panic(err)
	}
	// L2DB
	l2DB := l2db.NewL2DB(database, 10, 100, 24*time.Hour)
	test.CleanL2DB(l2DB.DB())

	// Init API
	api := gin.Default()
	if err := SetAPIEndpoints(
		true,
		true,
		api,
		hdb,
		sdb,
		l2DB,
	); err != nil {
		panic(err)
	}
	// Start server
	server := &http.Server{Addr: apiPort, Handler: api}
	go func() {
		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Populate DBs
	// Clean DB
	err = h.Reorg(0)
	if err != nil {
		panic(err)
	}
	// Gen blocks and add them to DB
	const nBlocks = 5
	blocks := test.GenBlocks(1, nBlocks+1)
	err = h.AddBlocks(blocks)
	if err != nil {
		panic(err)
	}
	// Gen tokens and add them to DB
	const nTokens = 10
	tokens := test.GenTokens(nTokens, blocks)
	err = h.AddTokens(tokens)
	if err != nil {
		panic(err)
	}
	// Set token value
	tokensUSD := []historydb.TokenRead{}
	for i, tkn := range tokens {
		token := historydb.TokenRead{
			TokenID:     tkn.TokenID,
			EthBlockNum: tkn.EthBlockNum,
			EthAddr:     tkn.EthAddr,
			Name:        tkn.Name,
			Symbol:      tkn.Symbol,
			Decimals:    tkn.Decimals,
		}
		// Set value of 50% of the tokens
		if i%2 != 0 {
			value := float64(i) * 1.234567
			now := time.Now().UTC()
			token.USD = &value
			token.USDUpdate = &now
			err = h.UpdateTokenValue(token.Symbol, value)
			if err != nil {
				panic(err)
			}
		}
		tokensUSD = append(tokensUSD, token)
	}
	// Gen batches and add them to DB
	const nBatches = 10
	batches := test.GenBatches(nBatches, blocks)
	err = h.AddBatches(batches)
	if err != nil {
		panic(err)
	}
	// Gen accounts and add them to DB
	const totalAccounts = 40
	const userAccounts = 4
	usrAddr := ethCommon.BigToAddress(big.NewInt(4896847))
	privK := babyjub.NewRandPrivKey()
	usrBjj := privK.Public()
	accs := test.GenAccounts(totalAccounts, userAccounts, tokens, &usrAddr, usrBjj, batches)
	err = h.AddAccounts(accs)
	if err != nil {
		panic(err)
	}
	// Gen exits and add them to DB
	const totalExits = 40
	exits := test.GenExitTree(totalExits, batches, accs)
	err = h.AddExitTree(exits)
	if err != nil {
		panic(err)
	}
	// Gen L1Txs and add them to DB
	const totalL1Txs = 40
	const userL1Txs = 4
	usrL1Txs, othrL1Txs := test.GenL1Txs(256, totalL1Txs, userL1Txs, &usrAddr, accs, tokens, blocks, batches)
	// Gen L2Txs and add them to DB
	const totalL2Txs = 20
	const userL2Txs = 4
	usrL2Txs, othrL2Txs := test.GenL2Txs(256+totalL1Txs, totalL2Txs, userL2Txs, &usrAddr, accs, tokens, blocks, batches)
	// Order txs
	sortedTxs := []txSortFielder{}
	for i := 0; i < len(usrL1Txs); i++ {
		wL1 := wrappedL1(usrL1Txs[i])
		sortedTxs = append(sortedTxs, &wL1)
	}
	for i := 0; i < len(othrL1Txs); i++ {
		wL1 := wrappedL1(othrL1Txs[i])
		sortedTxs = append(sortedTxs, &wL1)
	}
	for i := 0; i < len(usrL2Txs); i++ {
		wL2 := wrappedL2(usrL2Txs[i])
		sortedTxs = append(sortedTxs, &wL2)
	}
	for i := 0; i < len(othrL2Txs); i++ {
		wL2 := wrappedL2(othrL2Txs[i])
		sortedTxs = append(sortedTxs, &wL2)
	}
	sort.Sort(txsSort(sortedTxs))
	// Add txs to DB and prepare them for test commons
	usrTxs := []historyTxAPI{}
	allTxs := []historyTxAPI{}
	getTimestamp := func(blockNum int64) time.Time {
		for i := 0; i < len(blocks); i++ {
			if blocks[i].EthBlockNum == blockNum {
				return blocks[i].Timestamp
			}
		}
		panic("timesamp not found")
	}
	getToken := func(id common.TokenID) historydb.TokenRead {
		for i := 0; i < len(tokensUSD); i++ {
			if tokensUSD[i].TokenID == id {
				return tokensUSD[i]
			}
		}
		panic("token not found")
	}
	getTokenByIdx := func(idx common.Idx) historydb.TokenRead {
		for _, acc := range accs {
			if idx == acc.Idx {
				return getToken(acc.TokenID)
			}
		}
		panic("token not found")
	}
	usrIdxs := []string{}
	for _, acc := range accs {
		if acc.EthAddr == usrAddr || acc.PublicKey == usrBjj {
			for _, token := range tokens {
				if token.TokenID == acc.TokenID {
					usrIdxs = append(usrIdxs, idxToHez(acc.Idx, token.Symbol))
				}
			}
		}
	}
	isUsrTx := func(tx historyTxAPI) bool {
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
	for _, genericTx := range sortedTxs {
		l1 := genericTx.L1()
		l2 := genericTx.L2()
		if l1 != nil {
			// Add L1 tx to DB
			err = h.AddL1Txs([]common.L1Tx{*l1})
			if err != nil {
				panic(err)
			}
			// L1Tx ==> historyTxAPI
			token := getToken(l1.TokenID)
			tx := historyTxAPI{
				IsL1:      "L1",
				TxID:      l1.TxID.String(),
				Type:      l1.Type,
				Position:  l1.Position,
				ToIdx:     idxToHez(l1.ToIdx, token.Symbol),
				Amount:    l1.Amount.String(),
				BatchNum:  l1.BatchNum,
				Timestamp: getTimestamp(l1.EthBlockNum),
				L1Info: &l1Info{
					ToForgeL1TxsNum: l1.ToForgeL1TxsNum,
					UserOrigin:      l1.UserOrigin,
					FromEthAddr:     ethAddrToHez(l1.FromEthAddr),
					FromBJJ:         bjjToString(l1.FromBJJ),
					LoadAmount:      l1.LoadAmount.String(),
					EthBlockNum:     l1.EthBlockNum,
				},
				Token: token,
			}
			if l1.FromIdx != 0 {
				idxStr := idxToHez(l1.FromIdx, token.Symbol)
				tx.FromIdx = &idxStr
			}
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
		} else {
			// Add L2 tx to DB
			err = h.AddL2Txs([]common.L2Tx{*l2})
			if err != nil {
				panic(err)
			}
			// L2Tx ==> historyTxAPI
			var tokenID common.TokenID
			found := false
			for _, acc := range accs {
				if acc.Idx == l2.FromIdx {
					found = true
					tokenID = acc.TokenID
					break
				}
			}
			if !found {
				panic("tokenID not found")
			}
			token := getToken(tokenID)
			tx := historyTxAPI{
				IsL1:      "L2",
				TxID:      l2.TxID.String(),
				Type:      l2.Type,
				Position:  l2.Position,
				ToIdx:     idxToHez(l2.ToIdx, token.Symbol),
				Amount:    l2.Amount.String(),
				BatchNum:  &l2.BatchNum,
				Timestamp: getTimestamp(l2.EthBlockNum),
				L2Info: &l2Info{
					Nonce: l2.Nonce,
					Fee:   l2.Fee,
				},
				Token: token,
			}
			if l2.FromIdx != 0 {
				idxStr := idxToHez(l2.FromIdx, token.Symbol)
				tx.FromIdx = &idxStr
			}
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
	// Transform exits to API
	exitsToAPIExits := func(exits []common.ExitInfo, accs []common.Account, tokens []common.Token) []exitAPI {
		historyExits := []historydb.HistoryExit{}
		for _, exit := range exits {
			token := getTokenByIdx(exit.AccountIdx)
			historyExits = append(historyExits, historydb.HistoryExit{
				BatchNum:               exit.BatchNum,
				AccountIdx:             exit.AccountIdx,
				MerkleProof:            exit.MerkleProof,
				Balance:                exit.Balance,
				InstantWithdrawn:       exit.InstantWithdrawn,
				DelayedWithdrawRequest: exit.DelayedWithdrawRequest,
				DelayedWithdrawn:       exit.DelayedWithdrawn,
				TokenID:                token.TokenID,
				TokenEthBlockNum:       token.EthBlockNum,
				TokenEthAddr:           token.EthAddr,
				TokenName:              token.Name,
				TokenSymbol:            token.Symbol,
				TokenDecimals:          token.Decimals,
				TokenUSD:               token.USD,
				TokenUSDUpdate:         token.USDUpdate,
			})
		}
		return historyExitsToAPI(historyExits)
	}
	apiExits := exitsToAPIExits(exits, accs, tokens)
	// sort.Sort(apiExits)
	usrExits := []exitAPI{}
	for _, exit := range apiExits {
		for _, idx := range usrIdxs {
			if idx == exit.AccountIdx {
				usrExits = append(usrExits, exit)
			}
		}
	}
	tc = testCommon{
		blocks:   blocks,
		tokens:   tokensUSD,
		batches:  batches,
		usrAddr:  ethAddrToHez(usrAddr),
		usrBjj:   bjjToString(usrBjj),
		accs:     accs,
		usrTxs:   usrTxs,
		allTxs:   allTxs,
		exits:    apiExits,
		usrExits: usrExits,
		router:   router,
	}
	// Run tests
	result := m.Run()
	// Stop server
	if err := server.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	if err := database.Close(); err != nil {
		panic(err)
	}
	os.Exit(result)
}

func TestGetHistoryTxs(t *testing.T) {
	endpoint := apiURL + "transactions-history"
	fetchedTxs := []historyTxAPI{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*historyTxsAPI).Txs); i++ {
			tmp, err := copystructure.Copy(intr.(*historyTxsAPI).Txs[i])
			if err != nil {
				panic(err)
			}
			fetchedTxs = append(fetchedTxs, tmp.(historyTxAPI))
		}
	}
	// Get all (no filters)
	limit := 8
	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	assertHistoryTxAPIs(t, tc.allTxs, fetchedTxs)
	// Uncomment once tx generation for tests is fixed
	// // Get by ethAddr
	// fetchedTxs = []historyTxAPI{}
	// limit = 7
	// path = fmt.Sprintf(
	// 	"%s?hermezEthereumAddress=%s&limit=%d&fromItem=",
	// 	endpoint, tc.usrAddr, limit,
	// )
	// err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	// assert.NoError(t, err)
	// assertHistoryTxAPIs(t, tc.usrTxs, fetchedTxs)
	// // Get by bjj
	// fetchedTxs = []historyTxAPI{}
	// limit = 6
	// path = fmt.Sprintf(
	// 	"%s?BJJ=%s&limit=%d&fromItem=",
	// 	endpoint, tc.usrBjj, limit,
	// )
	// err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	// assert.NoError(t, err)
	// assertHistoryTxAPIs(t, tc.usrTxs, fetchedTxs)
	// Get by tokenID
	fetchedTxs = []historyTxAPI{}
	limit = 5
	tokenID := tc.allTxs[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d&fromItem=",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	tokenIDTxs := []historyTxAPI{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].Token.TokenID == tokenID {
			tokenIDTxs = append(tokenIDTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, tokenIDTxs, fetchedTxs)
	// idx
	fetchedTxs = []historyTxAPI{}
	limit = 4
	idx := tc.allTxs[0].ToIdx
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d&fromItem=",
		endpoint, idx, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	idxTxs := []historyTxAPI{}
	for i := 0; i < len(tc.allTxs); i++ {
		if (tc.allTxs[i].FromIdx != nil && (*tc.allTxs[i].FromIdx)[6:] == idx[6:]) ||
			tc.allTxs[i].ToIdx[6:] == idx[6:] {
			idxTxs = append(idxTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, idxTxs, fetchedTxs)
	// batchNum
	fetchedTxs = []historyTxAPI{}
	limit = 3
	batchNum := tc.allTxs[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d&fromItem=",
		endpoint, *batchNum, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	batchNumTxs := []historyTxAPI{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil &&
			*tc.allTxs[i].BatchNum == *batchNum {
			batchNumTxs = append(batchNumTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, batchNumTxs, fetchedTxs)
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
		fetchedTxs = []historyTxAPI{}
		limit = 2
		path = fmt.Sprintf(
			"%s?type=%s&limit=%d&fromItem=",
			endpoint, txType, limit,
		)
		err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
		assert.NoError(t, err)
		txTypeTxs := []historyTxAPI{}
		for i := 0; i < len(tc.allTxs); i++ {
			if tc.allTxs[i].Type == txType {
				txTypeTxs = append(txTypeTxs, tc.allTxs[i])
			}
		}
		assertHistoryTxAPIs(t, txTypeTxs, fetchedTxs)
	}
	// Multiple filters
	fetchedTxs = []historyTxAPI{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokenId=%d&limit=%d&fromItem=",
		endpoint, *batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	mixedTxs := []historyTxAPI{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil {
			if *tc.allTxs[i].BatchNum == *batchNum && tc.allTxs[i].Token.TokenID == tokenID {
				mixedTxs = append(mixedTxs, tc.allTxs[i])
			}
		}
	}
	assertHistoryTxAPIs(t, mixedTxs, fetchedTxs)
	// All, in reverse order
	fetchedTxs = []historyTxAPI{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderDesc, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	flipedTxs := []historyTxAPI{}
	for i := 0; i < len(tc.allTxs); i++ {
		flipedTxs = append(flipedTxs, tc.allTxs[len(tc.allTxs)-1-i])
	}
	assertHistoryTxAPIs(t, flipedTxs, fetchedTxs)
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
	fetchedTxs := []historyTxAPI{}
	for _, tx := range tc.allTxs {
		fetchedTx := historyTxAPI{}
		assert.NoError(t, doGoodReq("GET", endpoint+tx.TxID, nil, &fetchedTx))
		fetchedTxs = append(fetchedTxs, fetchedTx)
	}
	assertHistoryTxAPIs(t, tc.allTxs, fetchedTxs)
	// 400
	err := doBadReq("GET", endpoint+"0x001", nil, 400)
	assert.NoError(t, err)
	// 404
	err = doBadReq("GET", endpoint+"0x00000000000001e240004700", nil, 404)
	assert.NoError(t, err)
}

func assertHistoryTxAPIs(t *testing.T, expected, actual []historyTxAPI) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		actual[i].ItemID = 0
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

func TestGetExits(t *testing.T) {
	endpoint := apiURL + "exits"
	fetchedExits := []exitAPI{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*exitsAPI).Exits); i++ {
			tmp, err := copystructure.Copy(intr.(*exitsAPI).Exits[i])
			if err != nil {
				panic(err)
			}
			fetchedExits = append(fetchedExits, tmp.(exitAPI))
		}
	}
	// Get all (no filters)
	limit := 8
	path := fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err := doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, tc.exits, fetchedExits)

	// Get by ethAddr
	fetchedExits = []exitAPI{}
	limit = 7
	path = fmt.Sprintf(
		"%s?hermezEthereumAddress=%s&limit=%d&fromItem=",
		endpoint, tc.usrAddr, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, tc.usrExits, fetchedExits)
	// Get by bjj
	fetchedExits = []exitAPI{}
	limit = 6
	path = fmt.Sprintf(
		"%s?BJJ=%s&limit=%d&fromItem=",
		endpoint, tc.usrBjj, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, tc.usrExits, fetchedExits)
	// Get by tokenID
	fetchedExits = []exitAPI{}
	limit = 5
	tokenID := tc.exits[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d&fromItem=",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	tokenIDExits := []exitAPI{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].Token.TokenID == tokenID {
			tokenIDExits = append(tokenIDExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, tokenIDExits, fetchedExits)
	// idx
	fetchedExits = []exitAPI{}
	limit = 4
	idx := tc.exits[0].AccountIdx
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d&fromItem=",
		endpoint, idx, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	idxExits := []exitAPI{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].AccountIdx[6:] == idx[6:] {
			idxExits = append(idxExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, idxExits, fetchedExits)
	// batchNum
	fetchedExits = []exitAPI{}
	limit = 3
	batchNum := tc.exits[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d&fromItem=",
		endpoint, batchNum, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	batchNumExits := []exitAPI{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].BatchNum == batchNum {
			batchNumExits = append(batchNumExits, tc.exits[i])
		}
	}
	assertExitAPIs(t, batchNumExits, fetchedExits)
	// Multiple filters
	fetchedExits = []exitAPI{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokeId=%d&limit=%d&fromItem=",
		endpoint, batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, historydb.OrderAsc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	mixedExits := []exitAPI{}
	flipedExits := []exitAPI{}
	for i := 0; i < len(tc.exits); i++ {
		if tc.exits[i].BatchNum == batchNum && tc.exits[i].Token.TokenID == tokenID {
			mixedExits = append(mixedExits, tc.exits[i])
		}
		flipedExits = append(flipedExits, tc.exits[len(tc.exits)-1-i])
	}
	assertExitAPIs(t, mixedExits, fetchedExits)
	// All, in reverse order
	fetchedExits = []exitAPI{}
	limit = 5
	path = fmt.Sprintf("%s?limit=%d&fromItem=", endpoint, limit)
	err = doGoodReqPaginated(path, historydb.OrderDesc, &exitsAPI{}, appendIter)
	assert.NoError(t, err)
	assertExitAPIs(t, flipedExits, fetchedExits)
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
	path = fmt.Sprintf("%s?limit=1000&fromItem=1000", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}

func TestGetExit(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "exits/"
	fetchedExits := []exitAPI{}
	for _, exit := range tc.exits {
		fetchedExit := exitAPI{}
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

func assertExitAPIs(t *testing.T, expected, actual []exitAPI) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		actual[i].ItemID = 0
		if expected[i].Token.USDUpdate == nil {
			assert.Equal(t, expected[i].Token.USDUpdate, actual[i].Token.USDUpdate)
		} else {
			assert.Equal(t, expected[i].Token.USDUpdate.Unix(), actual[i].Token.USDUpdate.Unix())
			expected[i].Token.USDUpdate = actual[i].Token.USDUpdate
		}
		assert.Equal(t, expected[i], actual[i])
	}
}

func TestGetToken(t *testing.T) {
	// Get all txs by their ID
	endpoint := apiURL + "tokens/"
	fetchedTokens := []historydb.TokenRead{}
	for _, token := range tc.tokens {
		fetchedToken := historydb.TokenRead{}
		assert.NoError(t, doGoodReq("GET", endpoint+strconv.Itoa(int(token.TokenID)), nil, &fetchedToken))
		fetchedTokens = append(fetchedTokens, fetchedToken)
	}
	assertTokensAPIs(t, tc.tokens, fetchedTokens)
}

func assertTokensAPIs(t *testing.T, expected, actual []historydb.TokenRead) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		if expected[i].USDUpdate == nil {
			assert.Equal(t, expected[i].USDUpdate, actual[i].USDUpdate)
		} else {
			assert.Equal(t, expected[i].USDUpdate.Unix(), actual[i].USDUpdate.Unix())
			expected[i].USDUpdate = actual[i].USDUpdate
		}
		assert.Equal(t, expected[i], actual[i])
	}
}

func doGoodReqPaginated(
	path, order string,
	iterStruct db.Paginationer,
	appendIter func(res interface{}),
) error {
	next := 0
	for {
		// Call API to get this iteration items
		iterPath := path
		if next == 0 && order == historydb.OrderDesc {
			// Fetch first item in reverse order
			iterPath += "99999"
		} else {
			// Fetch from next item or 0 if it's ascending order
			iterPath += strconv.Itoa(next)
		}
		if err := doGoodReq("GET", iterPath+"&order="+order, nil, iterStruct); err != nil {
			return err
		}
		appendIter(iterStruct)
		// Keep iterating?
		pag := iterStruct.GetPagination()
		if order == historydb.OrderAsc {
			if pag.LastReturnedItem == pag.LastItem { // No
				break
			} else { // Yes
				next = pag.LastReturnedItem + 1
			}
		} else {
			if pag.FirstReturnedItem == pag.FirstItem { // No
				break
			} else { // Yes
				next = pag.FirstReturnedItem - 1
			}
		}
	}
	return nil
}

func doGoodReq(method, path string, reqBody io.Reader, returnStruct interface{}) error {
	ctx := context.Background()
	client := &http.Client{}
	httpReq, _ := http.NewRequest(method, path, reqBody)
	route, pathParams, err := tc.router.FindRoute(httpReq.Method, httpReq.URL)
	if err != nil {
		return err
	}
	// Validate request against swagger spec
	requestValidationInput := &swagger.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := swagger.ValidateRequest(ctx, requestValidationInput); err != nil {
		return err
	}
	// Do API call
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return errors.New("Nil body")
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%d response: %s", resp.StatusCode, string(body))
	}
	// Unmarshal body into return struct
	if err := json.Unmarshal(body, returnStruct); err != nil {
		return err
	}
	// Validate response against swagger spec
	responseValidationInput := &swagger.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 resp.StatusCode,
		Header:                 resp.Header,
	}
	responseValidationInput = responseValidationInput.SetBodyBytes(body)
	return swagger.ValidateResponse(ctx, responseValidationInput)
}

func doBadReq(method, path string, reqBody io.Reader, expectedResponseCode int) error {
	ctx := context.Background()
	client := &http.Client{}
	httpReq, _ := http.NewRequest(method, path, reqBody)
	route, pathParams, err := tc.router.FindRoute(httpReq.Method, httpReq.URL)
	if err != nil {
		return err
	}
	// Validate request against swagger spec
	requestValidationInput := &swagger.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := swagger.ValidateRequest(ctx, requestValidationInput); err != nil {
		if expectedResponseCode != 400 {
			return err
		}
		log.Warn("The request does not match the API spec")
	}
	// Do API call
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	if resp.Body == nil {
		return errors.New("Nil body")
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectedResponseCode {
		return fmt.Errorf("Unexpected response code: %d", resp.StatusCode)
	}
	// Validate response against swagger spec
	responseValidationInput := &swagger.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 resp.StatusCode,
		Header:                 resp.Header,
	}
	responseValidationInput = responseValidationInput.SetBodyBytes(body)
	return swagger.ValidateResponse(ctx, responseValidationInput)
}
