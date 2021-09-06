package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	swagger "github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/stateapiupdater"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/test/txsets"
	"github.com/hermeznetwork/tracerr"
	"github.com/stretchr/testify/require"
)

// Pendinger is an interface that allows getting last returned item ID and PendingItems to be used for building fromItem
// when testing paginated endpoints.
type Pendinger interface {
	GetPending() (pendingItems, lastItemID uint64)
	Len() int
	New() Pendinger
}

const (
	apiPort = "4010"
	apiIP   = "http://localhost:"
	apiURL  = apiIP + apiPort + "/v1/"
)

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

	// close Block:0, Batch:1
	> batch

	CreateAccountDeposit(0) A: 11100000000000000
	CreateAccountDeposit(1) C: 22222222200000000000
	CreateAccountCoordinator(0) C

	// close Block:0, Batch:2
	> batchL1
	// Expected balances:
	//     Coord(0): 0, Coord(1): 0
	//     C(0): 0

	CreateAccountDeposit(1) A: 33333333300000000000

	// close Block:0, Batch:3
	> batchL1

	// close Block:0, Batch:4
	> batchL1

	CreateAccountDepositTransfer(0) B-A: 44444444400000000000, 123444444400000000000

	// close Block:0, Batch:5
	> batchL1
	CreateAccountDeposit(0) D: 55555555500000000000

	// close Block:0, Batch:6
	> batchL1

	CreateAccountCoordinator(1) B

	Transfer(1) A-B: 11100000000000000 (2)
	Transfer(0) B-C: 22200000000000000 (3)

	// close Block:0, Batch:7
	> batchL1 // forge L1User{1}, forge L1Coord{2}, forge L2{2}

	Deposit(0) C: 66666666600000000000
	DepositTransfer(0) C-D: 77777777700000000000, 12377777700000000000

	Transfer(0) A-B: 33350000000000000 (111)
	Transfer(0) C-A: 44450000000000000 (222)
	Transfer(1) B-C: 55550000000000000 (123)
	Exit(0) A: 66650000000000000 (44)

	ForceTransfer(0) D-B: 77777700000000000
	ForceExit(0) B: 88888800000000000

	// close Block:0, Batch:8
	> batchL1
	> block

	Transfer(0) D-A: 99950000000000000 (77)
	Transfer(0) B-D: 12300000000000000 (55)

	// close Block:1, Batch:1
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
	ForceTransfer(0) D-B: 77777700000000000
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
	poolTxsToSend    []common.PoolL2Tx
	poolTxsToReceive []testPoolTxReceive
	auths            []testAuth
	router           *swagger.Router
	bids             []testBid
	slots            []testSlot
	auctionVars      common.AuctionVariables
	rollupVars       common.RollupVariables
	wdelayerVars     common.WDelayerVariables
	nextForgers      []historydb.NextForgerAPI
}

var tc testCommon
var config configAPI
var api *API
var stateAPIUpdater *stateapiupdater.Updater

// TestMain initializes the API server, and fill HistoryDB and StateDB with fake data,
// emulating the task of the synchronizer in order to have data to be returned
// by the API endpoints that will be tested
func TestMain(m *testing.M) {
	// Initializations
	// Swagger
	router := swagger.NewRouter().WithSwaggerFromFile("./swagger.yml")
	// HistoryDB
	database, err := db.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	apiConnCon := db.NewAPIConnectionController(1, time.Second)
	hdb := historydb.NewHistoryDB(database, database, apiConnCon)
	if err != nil {
		panic(err)
	}
	// L2DB
	nodeConfig := &historydb.NodeConfig{
		MaxPoolTxs: 10,
		MinFeeUSD:  0.000000000000001,
		MaxFeeUSD:  10000000000,
	}
	l2DB := l2db.NewL2DB(database, database, 10, 1000, nodeConfig.MinFeeUSD, nodeConfig.MaxFeeUSD, 24*time.Hour, apiConnCon)
	test.WipeDB(l2DB.DB()) // this will clean HistoryDB and L2DB
	// Config (smart contract constants)
	chainID := uint16(0)
	_config := getConfigTest(chainID)
	config = configAPI{
		ChainID:           chainID,
		RollupConstants:   *newRollupConstants(_config.RollupConstants),
		AuctionConstants:  _config.AuctionConstants,
		WDelayerConstants: _config.WDelayerConstants,
	}

	// API
	apiGin := gin.Default()
	// Reset DB
	test.WipeDB(hdb.DB())

	constants := &historydb.Constants{
		SCConsts: common.SCConsts{
			Rollup:   _config.RollupConstants,
			Auction:  _config.AuctionConstants,
			WDelayer: _config.WDelayerConstants,
		},
		ChainID:       chainID,
		HermezAddress: _config.HermezAddress,
	}
	if err := hdb.SetConstants(constants); err != nil {
		panic(err)
	}
	if err := hdb.SetNodeConfig(nodeConfig); err != nil {
		panic(err)
	}

	api, err = NewAPI(Config{
		Version:                  "test",
		ExplorerEndpoints:        true,
		CoordinatorEndpoints:     true,
		Server:                   apiGin,
		HistoryDB:                hdb,
		L2DB:                     l2DB,
		StateDB:                  nil,
		EthClient:                nil,
		ForgerAddress:            nil,
		CoordinatorNetworkConfig: nil,
	})
	if err != nil {
		log.Error(err)
		panic(err)
	}
	// Start server
	listener, err := net.Listen("tcp", ":"+apiPort) //nolint:gosec
	if err != nil {
		panic(err)
	}
	server := &http.Server{Handler: apiGin}
	go func() {
		if err := server.Serve(listener); err != nil &&
			tracerr.Unwrap(err) != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Generate blockchain data with til
	tcc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
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
	err = tcc.FillBlocksForgedL1UserTxs(blocksData)
	if err != nil {
		panic(err)
	}
	AddAdditionalInformation(blocksData)
	// Generate L2 Txs with til
	commonPoolTxs, err := tcc.GeneratePoolL2Txs(txsets.SetPoolL2MinimumFlow0)
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
	err = api.historyDB.UpdateTokenValue(common.EmptyAddr, ethUSD)
	if err != nil {
		panic(err)
	}
	for _, block := range blocksData {
		// Insert block into HistoryDB
		// nolint reason: block is used as read only in the function
		if err := api.historyDB.AddBlockSCData(&block); err != nil { //nolint:gosec
			log.Error(err)
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
			err = api.historyDB.UpdateTokenValue(token.EthAddr, value)
			if err != nil {
				panic(err)
			}
			testTokens = append(testTokens, token)
		}
		// Set USD value for tokens in DB
		for _, batch := range block.Rollup.Batches {
			commonL2Txs = append(commonL2Txs, batch.L2Txs...)
			for i := range batch.CreatedAccounts {
				batch.CreatedAccounts[i].Nonce = nonce.Nonce(i)
				commonAccounts = append(commonAccounts, batch.CreatedAccounts[i])
			}
			commonBatches = append(commonBatches, batch.Batch)
			commonExitTree = append(commonExitTree, batch.ExitTree...)
			commonL1Txs = append(commonL1Txs, batch.L1UserTxs...)
			commonL1Txs = append(commonL1Txs, batch.L1CoordinatorTxs...)
		}
	}
	// Add unforged L1 tx
	unforgedTx := blocksData[len(blocksData)-1].Rollup.L1UserTxs[0]
	if unforgedTx.BatchNum != nil {
		panic("Unforged tx batch num should be nil")
	}
	commonL1Txs = append(commonL1Txs, unforgedTx)

	// Generate Coordinators and add them to HistoryDB
	const nCoords = 10
	commonCoords := test.GenCoordinators(nCoords, commonBlocks)
	// Update one coordinator to test behaviour when bidder address is repeated
	updatedCoordBlock := commonCoords[len(commonCoords)-1].EthBlockNum
	commonCoords = append(commonCoords, common.Coordinator{
		Bidder:      commonCoords[0].Bidder,
		Forger:      commonCoords[0].Forger,
		EthBlockNum: updatedCoordBlock,
		URL:         commonCoords[0].URL + ".new",
	})
	if err := api.historyDB.AddCoordinators(commonCoords); err != nil {
		panic(err)
	}

	// Test next forgers
	// Set auction vars
	// Slots 3 and 6 will have bids that will be invalidated because of minBid update
	// Slots 4 and 7 will have valid bids, the rest will be cordinator slots
	var slot3MinBid int64 = 3
	var slot4MinBid int64 = 4
	var slot6MinBid int64 = 6
	var slot7MinBid int64 = 7
	// First update will indicate how things behave from slot 0
	var defaultSlotSetBid [6]*big.Int = [6]*big.Int{
		big.NewInt(10),          // Slot 0 min bid
		big.NewInt(10),          // Slot 1 min bid
		big.NewInt(10),          // Slot 2 min bid
		big.NewInt(slot3MinBid), // Slot 3 min bid
		big.NewInt(slot4MinBid), // Slot 4 min bid
		big.NewInt(10),          // Slot 5 min bid
	}
	auctionVars := common.AuctionVariables{
		EthBlockNum:              int64(2),
		DonationAddress:          ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		DefaultSlotSetBid:        defaultSlotSetBid,
		DefaultSlotSetBidSlotNum: 0,
		Outbidding:               uint16(1),
		SlotDeadline:             uint8(20),
		BootCoordinator:          ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		BootCoordinatorURL:       "https://boot.coordinator.io",
		ClosedAuctionSlots:       uint16(10),
		OpenAuctionSlots:         uint16(20),
	}
	if err := api.historyDB.AddAuctionVars(&auctionVars); err != nil {
		panic(err)
	}
	// Last update in auction vars will indicate how things will behave from slot 5
	defaultSlotSetBid = [6]*big.Int{
		big.NewInt(10),          // Slot 5 min bid
		big.NewInt(slot6MinBid), // Slot 6 min bid
		big.NewInt(slot7MinBid), // Slot 7 min bid
		big.NewInt(10),          // Slot 8 min bid
		big.NewInt(10),          // Slot 9 min bid
		big.NewInt(10),          // Slot 10 min bid
	}
	auctionVars = common.AuctionVariables{
		EthBlockNum:              int64(3),
		DonationAddress:          ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		DefaultSlotSetBid:        defaultSlotSetBid,
		DefaultSlotSetBidSlotNum: 5,
		Outbidding:               uint16(1),
		SlotDeadline:             uint8(20),
		BootCoordinator:          ethCommon.HexToAddress("0x1111111111111111111111111111111111111111"),
		BootCoordinatorURL:       "https://boot.coordinator.io",
		ClosedAuctionSlots:       uint16(10),
		OpenAuctionSlots:         uint16(20),
	}
	if err := api.historyDB.AddAuctionVars(&auctionVars); err != nil {
		panic(err)
	}

	// Generate Bids and add them to HistoryDB
	bids := []common.Bid{}
	// Slot 1 and 2, no bids, wins boot coordinator
	// Slot 3, below what's going to be the minimum (wins boot coordinator)
	bids = append(bids, common.Bid{
		SlotNum:     3,
		BidValue:    big.NewInt(slot3MinBid - 1),
		EthBlockNum: commonBlocks[0].Num,
		Bidder:      commonCoords[0].Bidder,
	})
	// Slot 4, valid bid (wins bidder)
	bids = append(bids, common.Bid{
		SlotNum:     4,
		BidValue:    big.NewInt(slot4MinBid),
		EthBlockNum: commonBlocks[0].Num,
		Bidder:      commonCoords[0].Bidder,
	})
	// Slot 5 no bids, wins boot coordinator
	// Slot 6, below what's going to be the minimum (wins boot coordinator)
	bids = append(bids, common.Bid{
		SlotNum:     6,
		BidValue:    big.NewInt(slot6MinBid - 1),
		EthBlockNum: commonBlocks[0].Num,
		Bidder:      commonCoords[0].Bidder,
	})
	// Slot 7, valid bid (wins bidder)
	bids = append(bids, common.Bid{
		SlotNum:     7,
		BidValue:    big.NewInt(slot7MinBid),
		EthBlockNum: commonBlocks[0].Num,
		Bidder:      commonCoords[0].Bidder,
	})
	if err = api.historyDB.AddBids(bids); err != nil {
		panic(err)
	}
	bootForger := historydb.NextForgerAPI{
		Coordinator: historydb.CoordinatorAPI{
			Forger: auctionVars.BootCoordinator,
			URL:    auctionVars.BootCoordinatorURL,
		},
	}
	// Set next forgers: set all as boot coordinator then replace the non boot coordinators
	nextForgers := []historydb.NextForgerAPI{}
	var initBlock int64 = 140
	var deltaBlocks int64 = 40
	for i := 1; i < int(auctionVars.ClosedAuctionSlots)+2; i++ {
		fromBlock := initBlock + deltaBlocks*int64(i-1)
		bootForger.Period = historydb.Period{
			SlotNum:   int64(i),
			FromBlock: fromBlock,
			ToBlock:   fromBlock + deltaBlocks - 1,
		}
		nextForgers = append(nextForgers, bootForger)
	}
	// Set next forgers that aren't the boot coordinator
	nonBootForger := historydb.CoordinatorAPI{
		Bidder: commonCoords[0].Bidder,
		Forger: commonCoords[0].Forger,
		URL:    commonCoords[0].URL + ".new",
	}
	// Slot 4
	nextForgers[3].Coordinator = nonBootForger
	// Slot 7
	nextForgers[6].Coordinator = nonBootForger

	buckets := make([]common.BucketParams, 5)
	for i := range buckets {
		buckets[i].CeilUSD = big.NewInt(int64(i) * 10)
		buckets[i].BlockStamp = big.NewInt(int64(i) * 100)
		buckets[i].Withdrawals = big.NewInt(int64(i) * 1000)
		buckets[i].RateBlocks = big.NewInt(int64(i) * 10000)
		buckets[i].RateWithdrawals = big.NewInt(int64(i) * 100000)
		buckets[i].MaxWithdrawals = big.NewInt(int64(i) * 1000000)
	}

	// Generate SC vars and add them to HistoryDB (if needed)
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

	stateAPIUpdater, err = stateapiupdater.NewUpdater(hdb, nodeConfig, &common.SCVariables{
		Rollup:   rollupVars,
		Auction:  auctionVars,
		WDelayer: wdelayerVars,
	}, constants, &stateapiupdater.RecommendedFeePolicy{
		PolicyType: stateapiupdater.RecommendedFeePolicyTypeAvgLastHour,
	}, 400)
	if err != nil {
		panic(err)
	}

	// Generate test data, as expected to be received/sended from/to the API
	testCoords := genTestCoordinators(commonCoords)
	testBids := genTestBids(commonBlocks, testCoords, bids)
	testExits := genTestExits(commonExitTree, testTokens, commonAccounts)
	testTxs := genTestTxs(commonL1Txs, commonL2Txs, commonAccounts, testTokens, commonBlocks)
	testBatches, testFullBatches := genTestBatches(commonBlocks, commonBatches, testTxs)
	poolTxsToSend, poolTxsToReceive := genTestPoolTxs(commonPoolTxs, testTokens, commonAccounts)
	// Add balance and nonce to historyDB
	accounts := genTestAccounts(commonAccounts, testTokens)
	accUpdates := []common.AccountUpdate{}
	for i := 0; i < len(accounts); i++ {
		balance := new(big.Int)
		balance.SetString(string(*accounts[i].Balance), 10)
		queryAccount, err := common.StringToIdx(string(accounts[i].Idx), "foo")
		if err != nil {
			panic(err)
		}
		accUpdates = append(accUpdates, common.AccountUpdate{
			EthBlockNum: 0,
			BatchNum:    1,
			Idx:         *queryAccount.AccountIndex,
			Nonce:       0,
			Balance:     balance,
		})
		accUpdates = append(accUpdates, common.AccountUpdate{
			EthBlockNum: 0,
			BatchNum:    1,
			Idx:         *queryAccount.AccountIndex,
			Nonce:       accounts[i].Nonce,
			Balance:     balance,
		})
	}
	if err := api.historyDB.AddAccountUpdates(accUpdates); err != nil {
		panic(err)
	}
	tc = testCommon{
		blocks:           commonBlocks,
		tokens:           testTokens,
		batches:          testBatches,
		fullBatches:      testFullBatches,
		coordinators:     testCoords,
		accounts:         accounts,
		txs:              testTxs,
		exits:            testExits,
		poolTxsToSend:    poolTxsToSend,
		poolTxsToReceive: poolTxsToReceive,
		auths:            genTestAuths(test.GenAuths(5, _config.ChainID, _config.HermezAddress)),
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
		nextForgers:  nextForgers,
	}

	// Run tests
	result := m.Run()
	// Fake server
	if os.Getenv("FAKE_SERVER") == "yes" {
		for {
			log.Info("Running fake server at " + apiURL + " until ^C is received")
			time.Sleep(30 * time.Second)
		}
	}
	// Stop server
	if err := server.Shutdown(context.Background()); err != nil {
		panic(err)
	}
	if err := database.Close(); err != nil {
		panic(err)
	}
	os.Exit(result)
}

func TestTimeout(t *testing.T) {
	databaseTO, err := db.InitTestSQLDB()
	require.NoError(t, err)
	apiConnConTO := db.NewAPIConnectionController(1, 100*time.Millisecond)
	hdbTO := historydb.NewHistoryDB(databaseTO, databaseTO, apiConnConTO)
	require.NoError(t, err)
	// L2DB
	l2DBTO := l2db.NewL2DB(databaseTO, databaseTO, 10, 1000, 1.0, 1000.0, 24*time.Hour, apiConnConTO)

	// API
	apiGinTO := gin.Default()
	finishWait := make(chan interface{})
	startWait := make(chan interface{})
	apiGinTO.GET("/v1/wait", func(c *gin.Context) {
		cancel, err := apiConnConTO.Acquire()
		defer cancel()
		require.NoError(t, err)
		defer apiConnConTO.Release()
		startWait <- nil
		<-finishWait
	})
	// Start server
	serverTO := &http.Server{Handler: apiGinTO}
	listener, err := net.Listen("tcp", ":4444") //nolint:gosec
	require.NoError(t, err)
	go func() {
		if err := serverTO.Serve(listener); err != nil &&
			tracerr.Unwrap(err) != http.ErrServerClosed {
			require.NoError(t, err)
		}
	}()
	_, err = NewAPI(Config{
		Version:                  "test",
		ExplorerEndpoints:        true,
		CoordinatorEndpoints:     true,
		Server:                   apiGinTO,
		HistoryDB:                hdbTO,
		L2DB:                     l2DBTO,
		StateDB:                  nil,
		EthClient:                nil,
		ForgerAddress:            nil,
		CoordinatorNetworkConfig: nil,
	})
	require.NoError(t, err)

	client := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:4444/v1/tokens", nil)
	require.NoError(t, err)
	httpReqWait, err := http.NewRequest("GET", "http://localhost:4444/v1/wait", nil)
	require.NoError(t, err)
	// Request that will get timed out
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Request that will make the API busy
		_, err = client.Do(httpReqWait)
		require.NoError(t, err)
		wg.Done()
	}()
	<-startWait
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	defer resp.Body.Close() //nolint
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	// Unmarshal body into return struct
	msg := &errorMsg{}
	err = json.Unmarshal(body, msg)
	require.NoError(t, err)
	// Check that the error was the expected down
	require.Equal(t, ErrSQLTimeout, msg.Message)
	finishWait <- nil

	// Stop server
	wg.Wait()
	require.NoError(t, serverTO.Shutdown(context.Background()))
	require.NoError(t, databaseTO.Close())
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
		if order == db.OrderDesc {
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
	httpReq.Header.Add("Content-Type", "application/json")
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

func doSimpleReq(method, endpoint string) (string, error) {
	client := &http.Client{}
	httpReq, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	//nolint
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", tracerr.Wrap(err)
	}
	return string(body), nil
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
	var coordLastUpdate historydb.CoordinatorAPI
	found := false
	for _, c := range coordinators {
		if c.Bidder == bidder {
			coordLastUpdate = c
			found = true
		}
	}
	if !found {
		panic("coordinator not found")
	}
	return coordLastUpdate
}
