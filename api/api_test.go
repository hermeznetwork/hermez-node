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
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
)

const apiPort = ":4010"
const apiURL = "http://localhost" + apiPort + "/"

type testCommon struct {
	blocks  []common.Block
	tokens  []common.Token
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
	pass := os.Getenv("POSTGRES_PASS")
	hdb, err := historydb.NewHistoryDB(5432, "localhost", "hermez", pass, "history")
	if err != nil {
		panic(err)
	}
	// Reset DB
	err = hdb.Reorg(-1)
	if err != nil {
		panic(err)
	}
	dir, err := ioutil.TempDir("", "tmpdb")
	if err != nil {
		panic(err)
	}
	sdb, err := statedb.NewStateDB(dir, false, 0)
	if err != nil {
		panic(err)
	}
	l2db, err := l2db.NewL2DB(5432, "localhost", "hermez", pass, "l2", 10, 512, 24*time.Hour)
	if err != nil {
		panic(err)
	}
	test.CleanL2DB(l2db.DB())
	// Init API
	api := gin.Default()
	if err := SetAPIEndpoints(
		true,
		true,
		api,
		hdb,
		sdb,
		l2db,
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
	usrL1Txs, othrL1Txs := test.GenL1Txs(0, totalL1Txs, userL1Txs, &usrAddr, accs, tokens, blocks, batches)
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
	usrL2Txs, othrL2Txs := test.GenL2Txs(totalL1Txs, totalL2Txs, userL2Txs, &usrAddr, accs, tokens, blocks, batches)
	var l2Txs []common.L2Tx
	l2Txs = append(l2Txs, usrL2Txs...)
	l2Txs = append(l2Txs, othrL2Txs...)
	err = h.AddL2Txs(l2Txs)
	if err != nil {
		panic(err)
	}

	// Set test commons
	txsToAPITxs := func(l1Txs []common.L1Tx, l2Txs []common.L2Tx, blocks []common.Block, tokens []common.Token) historyTxAPIs {
		// Transform L1Txs and L2Txs to generic Txs
		genericTxs := []*common.Tx{}
		for _, l1tx := range l1Txs {
			genericTxs = append(genericTxs, l1tx.Tx())
		}
		for _, l2tx := range l2Txs {
			genericTxs = append(genericTxs, l2tx.Tx())
		}
		// Transform generic Txs to HistoryTx
		historyTxs := []*historydb.HistoryTx{}
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
			token := common.Token{}
			for i := 0; i < len(tokens); i++ {
				if tokens[i].TokenID == genericTx.TokenID {
					token = tokens[i]
					break
				}
			}
			historyTxs = append(historyTxs, &historydb.HistoryTx{
				IsL1:            genericTx.IsL1,
				TxID:            genericTx.TxID,
				Type:            genericTx.Type,
				Position:        genericTx.Position,
				FromIdx:         genericTx.FromIdx,
				ToIdx:           genericTx.ToIdx,
				Amount:          genericTx.Amount,
				AmountFloat:     genericTx.AmountFloat,
				TokenID:         genericTx.TokenID,
				USD:             token.USD * genericTx.AmountFloat,
				BatchNum:        genericTx.BatchNum,
				EthBlockNum:     genericTx.EthBlockNum,
				ToForgeL1TxsNum: genericTx.ToForgeL1TxsNum,
				UserOrigin:      genericTx.UserOrigin,
				FromEthAddr:     genericTx.FromEthAddr,
				FromBJJ:         genericTx.FromBJJ,
				LoadAmount:      genericTx.LoadAmount,
				LoadAmountFloat: genericTx.LoadAmountFloat,
				LoadAmountUSD:   token.USD * genericTx.LoadAmountFloat,
				Fee:             genericTx.Fee,
				FeeUSD:          genericTx.Fee.Percentage() * token.USD * genericTx.AmountFloat,
				Nonce:           genericTx.Nonce,
				Timestamp:       timestamp,
				TokenSymbol:     token.Symbol,
				CurrentUSD:      token.USD * genericTx.AmountFloat,
				USDUpdate:       token.USDUpdate,
			})
		}
		return historyTxAPIs(historyTxsToAPI(historyTxs))
	}
	usrTxs := txsToAPITxs(usrL1Txs, usrL2Txs, blocks, tokens)
	sort.Sort(usrTxs)
	othrTxs := txsToAPITxs(othrL1Txs, othrL2Txs, blocks, tokens)
	sort.Sort(othrTxs)
	allTxs := append(usrTxs, othrTxs...)
	sort.Sort(allTxs)
	tc = testCommon{
		blocks:  blocks,
		tokens:  tokens,
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
	if err := h.Close(); err != nil {
		panic(err)
	}
	if err := l2.Close(); err != nil {
		panic(err)
	}
	os.Exit(result)
}

func TestGetHistoryTxs(t *testing.T) {
	endpoint := apiURL + "transactions-history"
	fetchedTxs := historyTxAPIs{}
	appendIter := func(intr interface{}) {
		for i := 0; i < len(intr.(*historyTxsAPI).Txs); i++ {
			tmp := &historyTxAPI{}
			if err := copier.Copy(tmp, &intr.(*historyTxsAPI).Txs[i]); err != nil {
				panic(err)
			}
			fetchedTxs = append(fetchedTxs, *tmp)
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
	tokenID := tc.allTxs[0].TokenID
	path = fmt.Sprintf(
		"%s?tokenId=%d&limit=%d&offset=",
		endpoint, tokenID, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	tokenIDTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].TokenID == tokenID {
			tokenIDTxs = append(tokenIDTxs, tc.allTxs[i])
		}
	}
	assertHistoryTxAPIs(t, tokenIDTxs, fetchedTxs)
	// idx
	fetchedTxs = historyTxAPIs{}
	limit = 4
	idx := tc.allTxs[0].FromIdx
	path = fmt.Sprintf(
		"%s?accountIndex=%s&limit=%d&offset=",
		endpoint, idx, limit,
	)
	err = doGoodReqPaginated(path, &historyTxsAPI{}, appendIter)
	assert.NoError(t, err)
	idxTxs := historyTxAPIs{}
	for i := 0; i < len(tc.allTxs); i++ {
		if tc.allTxs[i].FromIdx == idx {
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
		common.TxTypeWithdrawn,
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
			if *tc.allTxs[i].BatchNum == *batchNum && tc.allTxs[i].TokenID == tokenID {
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
			tmpItem := &historyTxAPI{}
			if err := copier.Copy(tmpItem, &intr.(*historyTxsAPI).Txs[i]); err != nil {
				panic(err)
			}
			tmpAll = append(tmpAll, *tmpItem)
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

func assertHistoryTxAPIs(t *testing.T, expected, actual historyTxAPIs) {
	assert.Equal(t, len(expected), len(actual))
	for i := 0; i < len(actual); i++ { //nolint len(actual) won't change within the loop
		assert.Equal(t, expected[i].Timestamp.Unix(), actual[i].Timestamp.Unix())
		expected[i].Timestamp = actual[i].Timestamp
		assert.Equal(t, expected[i].USDUpdate.Unix(), actual[i].USDUpdate.Unix())
		expected[i].USDUpdate = actual[i].USDUpdate
		if expected[i].L2Info != nil {
			if expected[i].L2Info.FeeUSD > actual[i].L2Info.FeeUSD {
				assert.Less(t, 0.999, actual[i].L2Info.FeeUSD/expected[i].L2Info.FeeUSD)
			} else if expected[i].L2Info.FeeUSD < actual[i].L2Info.FeeUSD {
				assert.Less(t, 0.999, expected[i].L2Info.FeeUSD/actual[i].L2Info.FeeUSD)
			}
			expected[i].L2Info.FeeUSD = actual[i].L2Info.FeeUSD
		}
		assert.Equal(t, expected[i], actual[i])
	}
}

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
