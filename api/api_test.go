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
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/tracerr"
)

// Pendinger is an interface that allows getting last returned item ID and PendingItems to be used for building fromItem
// when testing paginated endpoints.
type Pendinger interface {
	GetPending() (pendingItems, lastItemID uint64)
	Len() int
	New() Pendinger
}

const apiPort = ":4010"
const apiURL = "http://localhost" + apiPort + "/"

var SetBlockchain = `
	Type: Blockchain

	AddToken(1)
	AddToken(2)
	AddToken(3)
	AddToken(4)
	AddToken(5)
	AddToken(6)
	AddToken(7)
	AddToken(8)
	> block

	// Coordinator accounts, Idxs: 256, 257
	CreateAccountCoordinator(0) Coord
	CreateAccountCoordinator(1) Coord

	// close Block:0, Batch:0
	> batch

	CreateAccountDeposit(0) A: 11111111100000000000
	CreateAccountDeposit(1) C: 22222222200000000000
	CreateAccountCoordinator(0) C

	// close Block:0, Batch:1
	> batchL1
	// Expected balances:
	//     Coord(0): 0, Coord(1): 0
	//     C(0): 0

	CreateAccountDeposit(1) A: 33333333300000000000

	// close Block:0, Batch:2
	> batchL1

	// close Block:0, Batch:3
	> batchL1

	CreateAccountDepositTransfer(0) B-A: 44444444400000000000, 123444444400000000000

	// close Block:0, Batch:4
	> batchL1
	CreateAccountDeposit(0) D: 55555555500000000000

	// close Block:0, Batch:5
	> batchL1

	CreateAccountCoordinator(1) B

	Transfer(1) A-B: 11111100000000000 (2)
	Transfer(0) B-C: 22222200000000000 (3)

	// close Block:0, Batch:6
	> batchL1 // forge L1User{1}, forge L1Coord{2}, forge L2{2}

	Deposit(0) C: 66666666600000000000
	DepositTransfer(0) C-D: 77777777700000000000, 12377777700000000000

	Transfer(0) A-B: 33333300000000000 (111)
	Transfer(0) C-A: 44444400000000000 (222)
	Transfer(1) B-C: 55555500000000000 (123)
	Exit(0) A: 66666600000000000 (44)

	ForceTransfer(0) D-B: 77777700000000000
	ForceExit(0) B: 88888800000000000

	// close Block:0, Batch:7
	> batchL1
	> block

	Transfer(0) D-A: 99999900000000000 (77)
	Transfer(0) B-D: 12312300000000000 (55)

	// close Block:1, Batch:0
	> batchL1

	CreateAccountCoordinator(0) F

	CreateAccountCoordinator(0) G
	CreateAccountCoordinator(0) H
	CreateAccountCoordinator(0) I
	CreateAccountCoordinator(0) J
	CreateAccountCoordinator(0) K
	CreateAccountCoordinator(0) L
	CreateAccountCoordinator(0) M
	CreateAccountCoordinator(0) N
	CreateAccountCoordinator(0) O
	CreateAccountCoordinator(0) P

	CreateAccountCoordinator(5) G
	CreateAccountCoordinator(5) H
	CreateAccountCoordinator(5) I
	CreateAccountCoordinator(5) J
	CreateAccountCoordinator(5) K
	CreateAccountCoordinator(5) L
	CreateAccountCoordinator(5) M
	CreateAccountCoordinator(5) N
	CreateAccountCoordinator(5) O
	CreateAccountCoordinator(5) P

	CreateAccountCoordinator(2) G
	CreateAccountCoordinator(2) H
	CreateAccountCoordinator(2) I
	CreateAccountCoordinator(2) J
	CreateAccountCoordinator(2) K
	CreateAccountCoordinator(2) L
	CreateAccountCoordinator(2) M
	CreateAccountCoordinator(2) N
	CreateAccountCoordinator(2) O
	CreateAccountCoordinator(2) P


	> batch
	> block
	> batch
	> block
	> batch
	> block
`

type testCommon struct {
	blocks           []common.Block
	tokens           []historydb.TokenWithUSD
	batches          []testBatch
	fullBatches      []testFullBatch
	coordinators     []historydb.CoordinatorAPI
	accounts         []testAccount
	txs              []testTx
	exits            []testExit
	poolTxsToSend    []testPoolTxSend
	poolTxsToReceive []testPoolTxReceive
	auths            []testAuth
	router           *swagger.Router
	bids             []testBid
	slots            []testSlot
	auctionVars      common.AuctionVariables
	rollupVars       common.RollupVariables
	wdelayerVars     common.WDelayerVariables
}

var tc testCommon
var config configAPI
var api *API

// TestMain initializes the API server, and fill HistoryDB and StateDB with fake data,
// emulating the task of the synchronizer in order to have data to be returned
// by the API endpoints that will be tested
func TestMain(m *testing.M) {
	// Initializations
	// Swagger
	router := swagger.NewRouter().WithSwaggerFromFile("./swagger.yml")
	// HistoryDB
	pass := os.Getenv("POSTGRES_PASS")

	database, err := db.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	hdb := historydb.NewHistoryDB(database)
	if err != nil {
		panic(err)
	}
	// StateDB
	dir, err := ioutil.TempDir("", "tmpdb")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()
	sdb, err := statedb.NewStateDB(dir, statedb.TypeTxSelector, 0)
	if err != nil {
		panic(err)
	}
	// L2DB
	l2DB := l2db.NewL2DB(database, 10, 100, 24*time.Hour)
	test.WipeDB(l2DB.DB()) // this will clean HistoryDB and L2DB

	// Config (smart contract constants)
	_config := getConfigTest()
	config = configAPI{
		RollupConstants:   *newRollupConstants(_config.RollupConstants),
		AuctionConstants:  _config.AuctionConstants,
		WDelayerConstants: _config.WDelayerConstants,
	}

	// API
	apiGin := gin.Default()
	api, err = NewAPI(
		true,
		true,
		apiGin,
		hdb,
		sdb,
		l2DB,
		&_config,
	)
	if err != nil {
		panic(err)
	}
	// Start server
	server := &http.Server{Addr: apiPort, Handler: apiGin}
	go func() {
		if err := server.ListenAndServe(); err != nil && tracerr.Unwrap(err) != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Reset DB
	test.WipeDB(api.h.DB())

	// Genratre blockchain data with til
	tcc := til.NewContext(common.RollupConstMaxL1UserTx)
	tilCfgExtra := til.ConfigExtra{
		BootCoordAddr: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		CoordUser:     "Coord",
	}
	blocksData, err := tcc.GenerateBlocks(SetBlockchain)
	if err != nil {
		panic(err)
	}
	err = tcc.FillBlocksExtra(blocksData, &tilCfgExtra)
	if err != nil {
		panic(err)
	}
	AddAditionalInformation(blocksData)
	// Generate L2 Txs with til
	commonPoolTxs, err := tcc.GeneratePoolL2Txs(til.SetPoolL2MinimumFlow0)
	if err != nil {
		panic(err)
	}

	// Extract til generated data, and add it to HistoryDB
	var commonBlocks []common.Block
	var commonBatches []common.Batch
	var commonAccounts []common.Account
	var commonExitTree []common.ExitInfo
	var commonL1Txs []common.L1Tx
	var commonL2Txs []common.L2Tx
	// Add ETH token at the beginning of the array
	testTokens := []historydb.TokenWithUSD{}
	ethUSD := float64(500)
	ethNow := time.Now()
	testTokens = append(testTokens, historydb.TokenWithUSD{
		TokenID:     test.EthToken.TokenID,
		EthBlockNum: test.EthToken.EthBlockNum,
		EthAddr:     test.EthToken.EthAddr,
		Name:        test.EthToken.Name,
		Symbol:      test.EthToken.Symbol,
		Decimals:    test.EthToken.Decimals,
		USD:         &ethUSD,
		USDUpdate:   &ethNow,
	})
	err = api.h.UpdateTokenValue(test.EthToken.Symbol, ethUSD)
	if err != nil {
		panic(err)
	}
	for _, block := range blocksData {
		// Insert block into HistoryDB
		// nolint reason: block is used as read only in the function
		if err := api.h.AddBlockSCData(&block); err != nil { //nolint:gosec
			panic(err)
		}
		// Extract data
		commonBlocks = append(commonBlocks, block.Block)
		for i, tkn := range block.Rollup.AddedTokens {
			token := historydb.TokenWithUSD{
				TokenID:     tkn.TokenID,
				EthBlockNum: tkn.EthBlockNum,
				EthAddr:     tkn.EthAddr,
				Name:        tkn.Name,
				Symbol:      tkn.Symbol,
				Decimals:    tkn.Decimals,
			}
			value := float64(i + 423)
			now := time.Now().UTC()
			token.USD = &value
			token.USDUpdate = &now
			// Set value in DB
			err = api.h.UpdateTokenValue(token.Symbol, value)
			if err != nil {
				panic(err)
			}
			testTokens = append(testTokens, token)
		}
		// Set USD value for tokens in DB
		commonL1Txs = append(commonL1Txs, block.Rollup.L1UserTxs...)
		for _, batch := range block.Rollup.Batches {
			commonL2Txs = append(commonL2Txs, batch.L2Txs...)
			for i := range batch.CreatedAccounts {
				batch.CreatedAccounts[i].Nonce = common.Nonce(i)
				commonAccounts = append(commonAccounts, batch.CreatedAccounts[i])
			}
			commonBatches = append(commonBatches, batch.Batch)
			commonExitTree = append(commonExitTree, batch.ExitTree...)
			commonL1Txs = append(commonL1Txs, batch.L1CoordinatorTxs...)
		}
	}

	// lastBlockNum2 := blocksData[len(blocksData)-1].Block.EthBlockNum

	// Add accounts to StateDB
	for i := 0; i < len(commonAccounts); i++ {
		if _, err := api.s.CreateAccount(commonAccounts[i].Idx, &commonAccounts[i]); err != nil {
			panic(err)
		}
	}

	// Generate Coordinators and add them to HistoryDB
	const nCoords = 10
	commonCoords := test.GenCoordinators(nCoords, commonBlocks)
	if err := api.h.AddCoordinators(commonCoords); err != nil {
		panic(err)
	}

	// Generate Bids and add them to HistoryDB
	const nBids = 20
	commonBids := test.GenBids(nBids, commonBlocks, commonCoords)
	if err = api.h.AddBids(commonBids); err != nil {
		panic(err)
	}

	// Generate SC vars and add them to HistoryDB (if needed)
	var defaultSlotSetBid [6]*big.Int = [6]*big.Int{big.NewInt(10), big.NewInt(10), big.NewInt(10), big.NewInt(10), big.NewInt(10), big.NewInt(10)}
	auctionVars := common.AuctionVariables{
		EthBlockNum:        int64(2),
		DonationAddress:    ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		DefaultSlotSetBid:  defaultSlotSetBid,
		Outbidding:         uint16(1),
		SlotDeadline:       uint8(20),
		BootCoordinator:    ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		BootCoordinatorURL: "https://boot.coordinator.io",
		ClosedAuctionSlots: uint16(2),
		OpenAuctionSlots:   uint16(5),
	}

	var buckets [common.RollupConstNumBuckets]common.BucketParams
	for i := range buckets {
		buckets[i].CeilUSD = big.NewInt(int64(i) * 10)
		buckets[i].Withdrawals = big.NewInt(int64(i) * 100)
		buckets[i].BlockWithdrawalRate = big.NewInt(int64(i) * 1000)
		buckets[i].MaxWithdrawals = big.NewInt(int64(i) * 10000)
	}

	rollupVars := common.RollupVariables{
		EthBlockNum:           int64(3),
		FeeAddToken:           big.NewInt(100),
		ForgeL1L2BatchTimeout: int64(44),
		WithdrawalDelay:       uint64(3000),
		Buckets:               buckets,
		SafeMode:              false,
	}

	wdelayerVars := common.WDelayerVariables{
		WithdrawalDelay: uint64(3000),
	}

	err = api.h.AddAuctionVars(&auctionVars)
	if err != nil {
		panic(err)
	}

	// Generate test data, as expected to be received/sended from/to the API
	testCoords := genTestCoordinators(commonCoords)
	testBids := genTestBids(commonBlocks, testCoords, commonBids)
	testExits := genTestExits(commonExitTree, testTokens, commonAccounts)
	testTxs := genTestTxs(commonL1Txs, commonL2Txs, commonAccounts, testTokens, commonBlocks)
	testBatches, testFullBatches := genTestBatches(commonBlocks, commonBatches, testTxs)
	poolTxsToSend, poolTxsToReceive := genTestPoolTxs(commonPoolTxs, testTokens, commonAccounts)
	tc = testCommon{
		blocks:           commonBlocks,
		tokens:           testTokens,
		batches:          testBatches,
		fullBatches:      testFullBatches,
		coordinators:     testCoords,
		accounts:         genTestAccounts(commonAccounts, testTokens),
		txs:              testTxs,
		exits:            testExits,
		poolTxsToSend:    poolTxsToSend,
		poolTxsToReceive: poolTxsToReceive,
		auths:            genTestAuths(test.GenAuths(5)),
		router:           router,
		bids:             testBids,
		slots: api.genTestSlots(
			20,
			commonBlocks[len(commonBlocks)-1].Num,
			testBids,
			auctionVars,
		),
		auctionVars:  auctionVars,
		rollupVars:   rollupVars,
		wdelayerVars: wdelayerVars,
	}

	// Fake server
	if os.Getenv("FAKE_SERVER") == "yes" {
		for {
			log.Info("Running fake server at " + apiURL + " until ^C is received")
			time.Sleep(30 * time.Second)
		}
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
	if err := os.RemoveAll(dir); err != nil {
		panic(err)
	}
	os.Exit(result)
}

func doGoodReqPaginated(
	path, order string,
	iterStruct Pendinger,
	appendIter func(res interface{}),
) error {
	var next uint64
	firstIte := true
	expectedTotal := 0
	totalReceived := 0
	for {
		// Calculate fromItem
		iterPath := path
		if !firstIte {
			iterPath += "&fromItem=" + strconv.Itoa(int(next))
		}
		// Call API to get this iteration items
		iterStruct = iterStruct.New()
		if err := doGoodReq(
			"GET", iterPath+"&order="+order, nil,
			iterStruct,
		); err != nil {
			return tracerr.Wrap(err)
		}
		appendIter(iterStruct)
		// Keep iterating?
		remaining, lastID := iterStruct.GetPending()
		if remaining == 0 {
			break
		}
		if order == historydb.OrderDesc {
			next = lastID - 1
		} else {
			next = lastID + 1
		}
		// Check that the expected amount of items is consistent across iterations
		totalReceived += iterStruct.Len()
		if firstIte {
			firstIte = false
			expectedTotal = totalReceived + int(remaining)
		}
		if expectedTotal != totalReceived+int(remaining) {
			panic(fmt.Sprintf(
				"pagination error, totalReceived + remaining should be %d, but is %d",
				expectedTotal, totalReceived+int(remaining),
			))
		}
	}
	return nil
}

func doGoodReq(method, path string, reqBody io.Reader, returnStruct interface{}) error {
	ctx := context.Background()
	client := &http.Client{}
	httpReq, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if reqBody != nil {
		httpReq.Header.Add("Content-Type", "application/json")
	}
	route, pathParams, err := tc.router.FindRoute(httpReq.Method, httpReq.URL)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Validate request against swagger spec
	requestValidationInput := &swagger.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := swagger.ValidateRequest(ctx, requestValidationInput); err != nil {
		return tracerr.Wrap(err)
	}
	// Do API call
	resp, err := client.Do(httpReq)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if resp.Body == nil && returnStruct != nil {
		return tracerr.Wrap(errors.New("Nil body"))
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if resp.StatusCode != 200 {
		return tracerr.Wrap(fmt.Errorf("%d response. Body: %s", resp.StatusCode, string(body)))
	}
	if returnStruct == nil {
		return nil
	}
	// Unmarshal body into return struct
	if err := json.Unmarshal(body, returnStruct); err != nil {
		log.Error("invalid json: " + string(body))
		log.Error(err)
		return tracerr.Wrap(err)
	}
	// log.Info(string(body))
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
		return tracerr.Wrap(err)
	}
	// Validate request against swagger spec
	requestValidationInput := &swagger.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := swagger.ValidateRequest(ctx, requestValidationInput); err != nil {
		if expectedResponseCode != 400 {
			return tracerr.Wrap(err)
		}
		log.Warn("The request does not match the API spec")
	}
	// Do API call
	resp, err := client.Do(httpReq)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if resp.Body == nil {
		return tracerr.Wrap(errors.New("Nil body"))
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if resp.StatusCode != expectedResponseCode {
		return tracerr.Wrap(fmt.Errorf("Unexpected response code: %d. Body: %s", resp.StatusCode, string(body)))
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

// test helpers

func getTimestamp(blockNum int64, blocks []common.Block) time.Time {
	for i := 0; i < len(blocks); i++ {
		if blocks[i].Num == blockNum {
			return blocks[i].Timestamp
		}
	}
	panic("timesamp not found")
}

func getTokenByID(id common.TokenID, tokens []historydb.TokenWithUSD) historydb.TokenWithUSD {
	for i := 0; i < len(tokens); i++ {
		if tokens[i].TokenID == id {
			return tokens[i]
		}
	}
	panic("token not found")
}

func getTokenByIdx(idx common.Idx, tokens []historydb.TokenWithUSD, accs []common.Account) historydb.TokenWithUSD {
	for _, acc := range accs {
		if idx == acc.Idx {
			return getTokenByID(acc.TokenID, tokens)
		}
	}
	panic("token not found")
}

func getAccountByIdx(idx common.Idx, accs []common.Account) *common.Account {
	for _, acc := range accs {
		if acc.Idx == idx {
			return &acc
		}
	}
	panic("account not found")
}

func getBlockByNum(ethBlockNum int64, blocks []common.Block) common.Block {
	for _, b := range blocks {
		if b.Num == ethBlockNum {
			return b
		}
	}
	panic("block not found")
}

func getCoordinatorByBidder(bidder ethCommon.Address, coordinators []historydb.CoordinatorAPI) historydb.CoordinatorAPI {
	for _, c := range coordinators {
		if c.Bidder == bidder {
			return c
		}
	}
	panic("coordinator not found")
}
