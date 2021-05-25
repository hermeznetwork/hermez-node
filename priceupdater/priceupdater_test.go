package priceupdater

import (
	"context"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var historyDB *historydb.HistoryDB

const usdtAddr = "0xdac17f958d2ee523a2206206994597c13d831ec7"

func TestPriceUpdater(t *testing.T) {
	// Init DB
	db, err := dbUtils.InitTestSQLDB()
	if err != nil {
		panic(err)
	}
	historyDB = historydb.NewHistoryDB(db, db, nil)
	// Clean DB
	test.WipeDB(historyDB.DB())
	// Populate DB
	// Gen blocks and add them to DB
	blocks := test.GenBlocks(1, 2)
	require.NoError(t, historyDB.AddBlocks(blocks))
	// Gen tokens and add them to DB
	tokens := []common.Token{
		{
			TokenID:     1,
			EthBlockNum: blocks[0].Num,
			EthAddr:     ethCommon.HexToAddress("0x1"),
			Name:        "DAI",
			Symbol:      "DAI",
			Decimals:    18,
		}, // Used to test get by token symbol
		{
			TokenID:     2,
			EthBlockNum: blocks[0].Num,
			EthAddr:     ethCommon.HexToAddress(usdtAddr),
			Name:        "Tether",
			Symbol:      "USDT",
			Decimals:    18,
		}, // Used to test get by SC addr
		{
			TokenID:     3,
			EthBlockNum: blocks[0].Num,
			EthAddr:     ethCommon.HexToAddress("0x2"),
			Name:        "FOO",
			Symbol:      "FOO",
			Decimals:    18,
		}, // Used to test ignore
		{
			TokenID:     4,
			EthBlockNum: blocks[0].Num,
			EthAddr:     ethCommon.HexToAddress("0x3"),
			Name:        "BAR",
			Symbol:      "BAR",
			Decimals:    18,
		}, // Used to test static
		{
			TokenID:     5,
			EthBlockNum: blocks[0].Num,
			EthAddr:     ethCommon.HexToAddress("0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"),
			Name:        "Uniswap",
			Symbol:      "UNI",
			Decimals:    18,
		}, // Used to test default
	}
	require.NoError(t, historyDB.AddTokens(tokens)) // ETH token exist in DB by default
	// Update token price used to test ignore
	ignoreValue := 44.44
	require.NoError(t, historyDB.UpdateTokenValue(tokens[2].EthAddr, ignoreValue))

	// Prepare token config
	tc := []Provider{
		{
			Provider:       "bitfinexV2",
			BASEURL:        "https://api-pub.bitfinex.com/v2/",
			URL:            "ticker/t",
			URLExtraParams: "USD",
			Symbols:        "1=DAI,2=USDT,3=ignore,5=UNI",
		},
		{
			Provider:       "CoinGeckoV3",
			BASEURL:        "https://api.coingecko.com/api/v3/",
			URL:            "simple/token_price/ethereum?contract_addresses=",
			URLExtraParams: "&vs_currencies=usd",
			Addresses:      "1=0x1,2=" + usdtAddr + ",3=ignore,5=0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
		},
	}
	var priority Priority = "bitfinexV2,CoinGeckoV3"
	var staticTokens = "4=30.02"
	var fiat = Fiat{
		APIKey:       "FFFFFFFFFF",
		URL:          "https://api.exchangeratesapi.io/v1/",
		BaseCurrency: "USD",
		Currencies:   "CNY,EUR,JPY,GBP",
	}

	// Init price updater
	pu, err := NewPriceUpdater(
		priority,
		tc,
		staticTokens,
		fiat,
		historyDB,
	)
	require.NoError(t, err)

	// Update token list
	require.NoError(t, pu.UpdateTokenList())

	// Update prices
	pu.UpdatePrices(context.Background())

	// Check results: get tokens from DB
	fetchedTokens, err := historyDB.GetTokensTest()
	require.NoError(t, err)
	// Check that tokens that are updated via API have value:
	// ETH
	require.NotNil(t, fetchedTokens[0].USDUpdate)
	assert.Greater(t, *fetchedTokens[0].USD, 0.0)
	// DAI
	require.NotNil(t, fetchedTokens[1].USDUpdate)
	assert.Greater(t, *fetchedTokens[1].USD, 0.0)
	// USDT
	require.NotNil(t, fetchedTokens[2].USDUpdate)
	assert.Greater(t, *fetchedTokens[2].USD, 0.0)
	// UNI
	require.NotNil(t, fetchedTokens[5].USDUpdate)
	assert.Greater(t, *fetchedTokens[5].USD, 0.0)
	// Check ignored token
	assert.Equal(t, ignoreValue, *fetchedTokens[3].USD)
	// Check static value
	assert.Equal(t, 30.02, *fetchedTokens[4].USD)

	// Get fiat currencies: get currencies from DB
	fetchedcurrencies, err := historyDB.GetAllFiatPrice("USD")
	require.NoError(t, err)
	assert.Equal(t, 0, len(fetchedcurrencies))
	//Update fiat currencies
	err = pu.UpdateFiatPricesMock(context.Background())
	require.NoError(t, err)

	// Check results: get fiat currencies from DB
	fetchedcurrencies, err = historyDB.GetAllFiatPrice("USD")
	require.NoError(t, err)
	assert.Greater(t, fetchedcurrencies[0].Price, 0.0)
	assert.Greater(t, fetchedcurrencies[1].Price, 0.0)
	assert.Greater(t, fetchedcurrencies[2].Price, 0.0)
	assert.Greater(t, fetchedcurrencies[3].Price, 0.0)
}
