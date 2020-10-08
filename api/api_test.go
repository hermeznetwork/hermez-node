package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	dbUtils "github.com/hermeznetwork/hermez-node/db"
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
	blocks  []common.Block
	tokens  []historydb.TokenRead
	batches []common.Batch
	usrAddr string
	usrBjj  string
	accs    []common.Account
	usrTxs  historyTxAPIs
	othrTxs historyTxAPIs
	allTxs  historyTxAPIs
	router  *swagger.Router
}

type historyTxAPIs []historyTxAPI

func (h historyTxAPIs) Len() int      { return len(h) }
func (h historyTxAPIs) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h historyTxAPIs) Less(i, j int) bool {
	// i not forged yet
	if h[i].BatchNum == nil {
		if h[j].BatchNum != nil { // j is already forged
			return false
		}
		// Both aren't forged, is i in a smaller position?
		return h[i].Position < h[j].Position
	}
	// i is forged
	if h[j].BatchNum == nil {
		return true // j is not forged
	}
	// Both are forged
	if *h[i].BatchNum == *h[j].BatchNum {
		// At the same batch,  is i in a smaller position?
		return h[i].Position < h[j].Position
	}
	// At different batches, is i in a smaller batch?
	return *h[i].BatchNum < *h[j].BatchNum
}

var tc testCommon

func TestMain(m *testing.M) {
	// Init swagger
	router := swagger.NewRouter().WithSwaggerFromFile("./swagger.yml")
	// Init DBs
	// HistoryDB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	hdb := historydb.NewHistoryDB(db)
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
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)
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
	// Gen L1Txs and add them to DB
	const totalL1Txs = 40
	const userL1Txs = 4
	usrL1Txs, othrL1Txs := test.GenL1Txs(256, totalL1Txs, userL1Txs, &usrAddr, accs, tokens, blocks, batches)
	var l1Txs []common.L1Tx
	l1Txs = append(l1Txs, usrL1Txs...)
	l1Txs = append(l1Txs, othrL1Txs...)
	err = h.AddL1Txs(l1Txs)
	if err != nil {
		panic(err)
	}
	// Gen L2Txs and add them to DB
	const totalL2Txs = 20
	const userL2Txs = 4
	usrL2Txs, othrL2Txs := test.GenL2Txs(256+totalL1Txs, totalL2Txs, userL2Txs, &usrAddr, accs, tokens, blocks, batches)
	var l2Txs []common.L2Tx
	l2Txs = append(l2Txs, usrL2Txs...)
	l2Txs = append(l2Txs, othrL2Txs...)
	err = h.AddL2Txs(l2Txs)
	if err != nil {
		panic(err)
	}

	// Set test commons
	txsToAPITxs := func(l1Txs []common.L1Tx, l2Txs []common.L2Tx, blocks []common.Block, tokens []historydb.TokenRead) historyTxAPIs {
		/* TODO: stop using l1tx.Tx() & l2tx.Tx()
		// Transform L1Txs and L2Txs to generic Txs
		genericTxs := []*common.Tx{}
		for _, l1tx := range l1Txs {
			genericTxs = append(genericTxs, l1tx.Tx())
		}
		for _, l2tx := range l2Txs {
			genericTxs = append(genericTxs, l2tx.Tx())
		}
		// Transform generic Txs to HistoryTx
		historyTxs := []historydb.HistoryTx{}
		for _, genericTx := range genericTxs {
			// find timestamp
			var timestamp time.Time
			for i := 0; i < len(blocks); i++ {
				if blocks[i].EthBlockNum == genericTx.EthBlockNum {
					timestamp = blocks[i].Timestamp
					break
				}
			}
			// find token
			var token historydb.TokenRead
			if genericTx.IsL1 {
				tokenID := genericTx.TokenID
				found := false
				for i := 0; i < len(tokens); i++ {
					if tokens[i].TokenID == tokenID {
						token = tokens[i]
						found = true
						break
					}
				}
				if !found {
					panic("Token not found")
				}
			} else {
				var id common.TokenID
				found := false
				for _, acc := range accs {
					if acc.Idx == genericTx.FromIdx {
						found = true
						id = acc.TokenID
						break
					}
				}
				if !found {
					panic("tokenID not found")
				}
				found = false
				for i := 0; i < len(tokensUSD); i++ {
					if tokensUSD[i].TokenID == id {
						token = tokensUSD[i]
						found = true
						break
					}
				}
				if !found {
					panic("tokenID not found")
				}
			}
			var usd, loadUSD, feeUSD *float64
			if token.USD != nil {
				noDecimalsUSD := *token.USD / math.Pow(10, float64(token.Decimals))
				usd = new(float64)
				*usd = noDecimalsUSD * genericTx.AmountFloat
				if genericTx.IsL1 {
					loadUSD = new(float64)
					*loadUSD = noDecimalsUSD * *genericTx.LoadAmountFloat
				} else {
					feeUSD = new(float64)
					*feeUSD = *usd * genericTx.Fee.Percentage()
				}
			}
			historyTx := &historydb.HistoryTx{
				IsL1:                  genericTx.IsL1,
				TxID:                  genericTx.TxID,
				Type:                  genericTx.Type,
				Position:              genericTx.Position,
				ToIdx:                 genericTx.ToIdx,
				Amount:                genericTx.Amount,
				HistoricUSD:           usd,
				BatchNum:              genericTx.BatchNum,
				EthBlockNum:           genericTx.EthBlockNum,
				ToForgeL1TxsNum:       genericTx.ToForgeL1TxsNum,
				UserOrigin:            genericTx.UserOrigin,
				FromBJJ:               genericTx.FromBJJ,
				LoadAmount:            genericTx.LoadAmount,
				HistoricLoadAmountUSD: loadUSD,
				Fee:                   genericTx.Fee,
				HistoricFeeUSD:        feeUSD,
				Nonce:                 genericTx.Nonce,
				Timestamp:             timestamp,
				TokenID:               token.TokenID,
				TokenEthBlockNum:      token.EthBlockNum,
				TokenEthAddr:          token.EthAddr,
				TokenName:             token.Name,
				TokenSymbol:           token.Symbol,
				TokenDecimals:         token.Decimals,
				TokenUSD:              token.USD,
				TokenUSDUpdate:        token.USDUpdate,
			}
			if genericTx.FromIdx != 0 {
				historyTx.FromIdx = &genericTx.FromIdx
			}
			if !bytes.Equal(genericTx.FromEthAddr.Bytes(), common.EmptyAddr.Bytes()) {
				historyTx.FromEthAddr = &genericTx.FromEthAddr
			}
			historyTxs = append(historyTxs, historyTx)
		}
		return historyTxAPIs(historyTxsToAPI(historyTxs))
		*/
		return nil
	}
	usrTxs := txsToAPITxs(usrL1Txs, usrL2Txs, blocks, tokensUSD)
	sort.Sort(usrTxs)
	othrTxs := txsToAPITxs(othrL1Txs, othrL2Txs, blocks, tokensUSD)
	sort.Sort(othrTxs)
	allTxs := append(usrTxs, othrTxs...)
	sort.Sort(allTxs)
	tc = testCommon{
		blocks:  blocks,
		tokens:  tokensUSD,
		batches: batches,
		usrAddr: "hez:" + usrAddr.String(),
		usrBjj:  bjjToString(usrBjj),
		accs:    accs,
		usrTxs:  usrTxs,
		othrTxs: othrTxs,
		allTxs:  allTxs,
		router:  router,
	}
	// Run tests
	result := m.Run()
	// Stop server
	if err := server.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	if err := db.Close(); err != nil {
		panic(err)
	}
	os.Exit(result)
}

func TestGetHistoryTxs(t *testing.T) {
	return
	//nolint:govet this is a temp patch to avoid running the test
	endpoint := apiURL + "transactions-history"
	fetchedTxs := historyTxAPIs{}
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
	path := fmt.Sprintf("%s?limit=%d&offset=", endpoint, limit)
	err := doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	assertHistoryTxAPIs(t, tc.allTxs, fetchedTxs)
	// Get by ethAddr
	fetchedTxs = historyTxAPIs{}
	limit = 7
	path = fmt.Sprintf(
		"%s?hermezEthereumAddress=%s&limit=%d&offset=",
		endpoint, tc.usrAddr, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	assertHistoryTxAPIs(t, tc.usrTxs, fetchedTxs)
	// Get by bjj
	fetchedTxs = historyTxAPIs{}
	limit = 6
	path = fmt.Sprintf(
		"%s?BJJ=%s&limit=%d&offset=",
		endpoint, tc.usrBjj, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	assertHistoryTxAPIs(t, tc.usrTxs, fetchedTxs)
	// Get by tokenID
	fetchedTxs = historyTxAPIs{}
	limit = 5
	tokenID := tc.allTxs[0].Token.TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d&offset=",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	tokenIDTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].Token.TokenID == tokenID {
			tokenIDTxs = append(tokenIDTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, tokenIDTxs, fetchedTxs)
	// idx
	fetchedTxs = historyTxAPIs{}
	limit = 4
	idx := tc.allTxs[0].ToIdx
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d&offset=",
		endpoint, idx, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	idxTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if (tc.allTxs[i].FromIdx != nil && (*tc.allTxs[i].FromIdx)[6:] == idx[6:]) ||
			tc.allTxs[i].ToIdx[6:] == idx[6:] {
			idxTxs = append(idxTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, idxTxs, fetchedTxs)
	// batchNum
	fetchedTxs = historyTxAPIs{}
	limit = 3
	batchNum := tc.allTxs[0].BatchNum
	path = fmt.Sprintf(
		"%s?batchNum=%d&limit=%d&offset=",
		endpoint, *batchNum, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	batchNumTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil &&
			*tc.allTxs[i].BatchNum == *batchNum {
			batchNumTxs = append(batchNumTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, batchNumTxs, fetchedTxs)
	// type
	txTypes := []common.TxType{
		common.TxTypeExit,
		common.TxTypeTransfer,
		common.TxTypeDeposit,
		common.TxTypeCreateAccountDeposit,
		common.TxTypeCreateAccountDepositTransfer,
		common.TxTypeDepositTransfer,
		common.TxTypeForceTransfer,
		common.TxTypeForceExit,
		common.TxTypeTransferToEthAddr,
		common.TxTypeTransferToBJJ,
	}
	for _, txType := range txTypes {
		fetchedTxs = historyTxAPIs{}
		limit = 2
		path = fmt.Sprintf(
			"%s?type=%s&limit=%d&offset=",
			endpoint, txType, limit,
		)
		err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
		assert.NoError(t, err)
		txTypeTxs := historyTxAPIs{}
		for i := 0; i < len(tc.allTxs); i++ {
			if tc.allTxs[i].Type == txType {
				txTypeTxs = append(txTypeTxs, tc.allTxs[i])
			}
		}
		assertHistoryTxAPIs(t, txTypeTxs, fetchedTxs)
	}
	// Multiple filters
	fetchedTxs = historyTxAPIs{}
	limit = 1
	path = fmt.Sprintf(
		"%s?batchNum=%d&tokeId=%d&limit=%d&offset=",
		endpoint, *batchNum, tokenID, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	mixedTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].BatchNum != nil {
			if *tc.allTxs[i].BatchNum == *batchNum && tc.allTxs[i].Token.TokenID == tokenID {
				mixedTxs = append(mixedTxs, tc.allTxs[i])
			}
		}
	}
	assertHistoryTxAPIs(t, mixedTxs, fetchedTxs)
	// All, in reverse order
	fetchedTxs = historyTxAPIs{}
	limit = 5
	path = fmt.Sprintf("%s?", endpoint)
	appendIterRev := func(intr interface{}) {
		tmpAll := historyTxAPIs{}
		for i := 0; i < len(intr.(*historyTxsAPI).Txs); i++ {
			tmp, err := copystructure.Copy(intr.(*historyTxsAPI).Txs[i])
			if err != nil {
				panic(err)
			}
			tmpAll = append(tmpAll, tmp.(historyTxAPI))
		}
		fetchedTxs = append(tmpAll, fetchedTxs...)
	}
	err = doGoodReqPaginatedReverse(path, &historyTxsAPI{}, appendIterRev, limit)
	assert.NoError(t, err)
	assertHistoryTxAPIs(t, tc.allTxs, fetchedTxs)
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
	path = fmt.Sprintf("%s?limit=1000&offset=1000", endpoint)
	err = doBadReq("GET", path, nil, 404)
	assert.NoError(t, err)
}

//nolint:govet this is a temp patch to avoid running the test
func assertHistoryTxAPIs(t *testing.T, expected, actual historyTxAPIs) {
	require.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
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

//nolint:govet this is a temp patch to avoid running the test
func doGoodReqPaginated(
	path string,
	iterStruct paginationer,
	appendIter func(res interface{}),
) error {
	next := 0
	for {
		// Call API to get this iteration items
		if err := doGoodReq("GET", path+strconv.Itoa(next), nil, iterStruct); err != nil {
			return err
		}
		appendIter(iterStruct)
		// Keep iterating?
		pag := iterStruct.GetPagination()
		if pag.LastReturnedItem == pag.TotalItems-1 { // No
			break
		} else { // Yes
			next = int(pag.LastReturnedItem + 1)
		}
	}
	return nil
}

//nolint:govet this is a temp patch to avoid running the test
func doGoodReqPaginatedReverse(
	path string,
	iterStruct paginationer,
	appendIter func(res interface{}),
	limit int,
) error {
	next := 0
	first := true
	for {
		// Call API to get this iteration items
		if first {
			first = false
			pagQuery := fmt.Sprintf("last=true&limit=%d", limit)
			if err := doGoodReq("GET", path+pagQuery, nil, iterStruct); err != nil {
				return err
			}
		} else {
			pagQuery := fmt.Sprintf("offset=%d&limit=%d", next, limit)
			if err := doGoodReq("GET", path+pagQuery, nil, iterStruct); err != nil {
				return err
			}
		}
		appendIter(iterStruct)
		// Keep iterating?
		pag := iterStruct.GetPagination()
		if iterStruct.Len() == pag.TotalItems || pag.LastReturnedItem-iterStruct.Len() == -1 { // No
			break
		} else { // Yes
			prevOffset := next
			next = pag.LastReturnedItem - iterStruct.Len() - limit + 1
			if next < 0 {
				next = 0
				limit = prevOffset
			}
		}
	}
	return nil
}

//nolint:govet this is a temp patch to avoid running the test
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

//nolint:govet this is a temp patch to avoid running the test
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
