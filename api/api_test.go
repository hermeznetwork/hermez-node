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
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

const apiPort = ":4010"
const apiURL = "http://localhost" + apiPort + "/"

type testCommon struct {
	blocks           []common.Block
	tokens           []historydb.TokenWithUSD
	batches          []testBatch
	fullBatches      []testFullBatch
	coordinators     []historydb.CoordinatorAPI
	usrAddr          string
	usrBjj           string
	accs             []common.Account
	usrTxs           []testTx
	allTxs           []testTx
	exits            []testExit
	usrExits         []testExit
	poolTxsToSend    []testPoolTxSend
	poolTxsToReceive []testPoolTxReceive
	auths            []testAuth
	router           *swagger.Router
	bids             []testBid
}

var tc testCommon
var config configAPI

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
	config.RollupConstants.ExchangeMultiplier = eth.RollupConstExchangeMultiplier
	config.RollupConstants.ExitIdx = eth.RollupConstExitIDx
	config.RollupConstants.ReservedIdx = eth.RollupConstReservedIDx
	config.RollupConstants.LimitLoadAmount, _ = new(big.Int).SetString("340282366920938463463374607431768211456", 10)
	config.RollupConstants.LimitL2TransferAmount, _ = new(big.Int).SetString("6277101735386680763835789423207666416102355444464034512896", 10)
	config.RollupConstants.LimitTokens = eth.RollupConstLimitTokens
	config.RollupConstants.L1CoordinatorTotalBytes = eth.RollupConstL1CoordinatorTotalBytes
	config.RollupConstants.L1UserTotalBytes = eth.RollupConstL1UserTotalBytes
	config.RollupConstants.MaxL1UserTx = eth.RollupConstMaxL1UserTx
	config.RollupConstants.MaxL1Tx = eth.RollupConstMaxL1Tx
	config.RollupConstants.InputSHAConstantBytes = eth.RollupConstInputSHAConstantBytes
	config.RollupConstants.NumBuckets = eth.RollupConstNumBuckets
	config.RollupConstants.MaxWithdrawalDelay = eth.RollupConstMaxWithdrawalDelay
	var rollupPublicConstants eth.RollupPublicConstants
	rollupPublicConstants.AbsoluteMaxL1L2BatchTimeout = 240
	rollupPublicConstants.HermezAuctionContract = ethCommon.HexToAddress("0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e")
	rollupPublicConstants.HermezGovernanceDAOAddress = ethCommon.HexToAddress("0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4")
	rollupPublicConstants.SafetyAddress = ethCommon.HexToAddress("0xE5904695748fe4A84b40b3fc79De2277660BD1D3")
	rollupPublicConstants.TokenHEZ = ethCommon.HexToAddress("0xf784709d2317D872237C4bC22f867d1BAe2913AB")
	rollupPublicConstants.WithdrawDelayerContract = ethCommon.HexToAddress("0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe")
	var verifier eth.RollupVerifierStruct
	verifier.MaxTx = 512
	verifier.NLevels = 32
	rollupPublicConstants.Verifiers = append(rollupPublicConstants.Verifiers, verifier)

	var auctionConstants eth.AuctionConstants
	auctionConstants.BlocksPerSlot = 40
	auctionConstants.GenesisBlockNum = 100
	auctionConstants.GovernanceAddress = ethCommon.HexToAddress("0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4")
	auctionConstants.InitialMinimalBidding, _ = new(big.Int).SetString("10000000000000000000", 10)
	auctionConstants.HermezRollup = ethCommon.HexToAddress("0xEa960515F8b4C237730F028cBAcF0a28E7F45dE0")
	auctionConstants.TokenHEZ = ethCommon.HexToAddress("0xf784709d2317D872237C4bC22f867d1BAe2913AB")

	var wdelayerConstants eth.WDelayerConstants
	wdelayerConstants.HermezRollup = ethCommon.HexToAddress("0xEa960515F8b4C237730F028cBAcF0a28E7F45dE0")
	wdelayerConstants.MaxEmergencyModeTime = uint64(1000000)
	wdelayerConstants.MaxWithdrawalDelay = uint64(10000000)

	config.RollupConstants.PublicConstants = rollupPublicConstants
	config.AuctionConstants = auctionConstants
	config.WDelayerConstants = wdelayerConstants

	// API
	api := gin.Default()
	if err := SetAPIEndpoints(
		true,
		true,
		api,
		hdb,
		sdb,
		l2DB,
		&config,
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

	// Fill HistoryDB and StateDB with fake data
	// Gen blocks and add them to DB
	const nBlocks = 5
	blocks := test.GenBlocks(1, nBlocks+1)
	err = h.AddBlocks(blocks)
	if err != nil {
		panic(err)
	}
	// Gen tokens and add them to DB
	const nTokens = 10
	tokens, ethToken := test.GenTokens(nTokens, blocks)
	err = h.AddTokens(tokens)
	if err != nil {
		panic(err)
	}
	tokens = append([]common.Token{ethToken}, tokens...)
	// Set token value
	tokensUSD := []historydb.TokenWithUSD{}
	for i, tkn := range tokens {
		token := historydb.TokenWithUSD{
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
	// Gen accounts and add them to HistoryDB and StateDB
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
	for i := 0; i < len(accs); i++ {
		if _, err := s.CreateAccount(accs[i].Idx, &accs[i]); err != nil {
			panic(err)
		}
	}
	// helper to vinculate user related resources
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
	// Gen exits and add them to DB
	const totalExits = 40
	exits := test.GenExitTree(totalExits, batches, accs)
	err = h.AddExitTree(exits)
	if err != nil {
		panic(err)
	}

	// L1 and L2 txs need to be sorted in a combined way
	// Gen L1Txs
	const totalL1Txs = 40
	const userL1Txs = 4
	usrL1Txs, othrL1Txs := test.GenL1Txs(256, totalL1Txs, userL1Txs, &usrAddr, accs, tokens, blocks, batches)
	// Gen L2Txs
	const totalL2Txs = 20
	const userL2Txs = 4
	usrL2Txs, othrL2Txs := test.GenL2Txs(256+totalL1Txs, totalL2Txs, userL2Txs, &usrAddr, accs, tokens, blocks, batches)
	// Sort txs
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
	// Store txs to DB
	for _, genericTx := range sortedTxs {
		l1 := genericTx.L1()
		l2 := genericTx.L2()
		if l1 != nil {
			err = h.AddL1Txs([]common.L1Tx{*l1})
			if err != nil {
				panic(err)
			}
		} else if l2 != nil {
			err = h.AddL2Txs([]common.L2Tx{*l2})
			if err != nil {
				panic(err)
			}
		} else {
			panic("should be l1 or l2")
		}
	}

	// Coordinators
	const nCoords = 10
	coords := test.GenCoordinators(nCoords, blocks)
	err = hdb.AddCoordinators(coords)
	if err != nil {
		panic(err)
	}
	fromItem := uint(0)
	limit := uint(99999)
	coordinators, _, err := hdb.GetCoordinatorsAPI(&fromItem, &limit, historydb.OrderAsc)
	if err != nil {
		panic(err)
	}

	// Bids
	const nBids = 10
	bids := test.GenBids(nBids, blocks, coords)
	err = hdb.AddBids(bids)
	if err != nil {
		panic(err)
	}

	// Set testCommon
	usrTxs, allTxs := genTestTxs(sortedTxs, usrIdxs, accs, tokensUSD, blocks)
	poolTxsToSend, poolTxsToReceive := genTestPoolTx(accs, []babyjub.PrivateKey{privK}, tokensUSD) // NOTE: pool txs are not inserted to the DB here. In the test they will be posted and getted.
	testBatches, fullBatches := genTestBatches(blocks, batches, allTxs)
	usrExits, allExits := genTestExits(exits, tokensUSD, accs, usrIdxs)
	tc = testCommon{
		blocks:           blocks,
		tokens:           tokensUSD,
		batches:          testBatches,
		fullBatches:      fullBatches,
		coordinators:     coordinators,
		usrAddr:          ethAddrToHez(usrAddr),
		usrBjj:           bjjToString(usrBjj),
		accs:             accs,
		usrTxs:           usrTxs,
		allTxs:           allTxs,
		exits:            allExits,
		usrExits:         usrExits,
		poolTxsToSend:    poolTxsToSend,
		poolTxsToReceive: poolTxsToReceive,
		auths:            genTestAuths(test.GenAuths(5)),
		router:           router,
		bids:             genTestBids(blocks, coordinators, bids),
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

func TestGetConfig(t *testing.T) {
	endpoint := apiURL + "config"
	var configTest configAPI
	assert.NoError(t, doGoodReq("GET", endpoint, nil, &configTest))
	assert.Equal(t, config, configTest)
	assert.Equal(t, cg, &configTest)
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
	httpReq, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		return err
	}
	if reqBody != nil {
		httpReq.Header.Add("Content-Type", "application/json")
	}
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
	if resp.Body == nil && returnStruct != nil {
		return errors.New("Nil body")
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%d response. Body: %s", resp.StatusCode, string(body))
	}
	if returnStruct == nil {
		return nil
	}
	// Unmarshal body into return struct
	if err := json.Unmarshal(body, returnStruct); err != nil {
		log.Error("invalid json: " + string(body))
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
		return fmt.Errorf("Unexpected response code: %d. Body: %s", resp.StatusCode, string(body))
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
		if blocks[i].EthBlockNum == blockNum {
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
		if b.EthBlockNum == ethBlockNum {
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
