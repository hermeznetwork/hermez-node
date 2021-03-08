package priceupdater

import (
	"context"
	"os"
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

func TestMain(m *testing.M) {
	// Init DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	historyDB = historydb.NewHistoryDB(db, db, nil)
	// Clean DB
	test.WipeDB(historyDB.DB())
	// Populate DB
	// Gen blocks and add them to DB
	blocks := test.GenBlocks(1, 2)
	err = historyDB.AddBlocks(blocks)
	if err != nil {
		panic(err)
	}
	// Gen tokens and add them to DB
	tokens := []common.Token{}
	tokens = append(tokens, common.Token{
		TokenID:     1,
		EthBlockNum: blocks[0].Num,
		EthAddr:     ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f"),
		Name:        "DAI",
		Symbol:      "DAI",
		Decimals:    18,
	})
	err = historyDB.AddTokens(tokens)
	if err != nil {
		panic(err)
	}

	result := m.Run()
	os.Exit(result)
}

func TestPriceUpdaterBitfinex(t *testing.T) {
	// Init price updater
	pu, err := NewPriceUpdater("https://api-pub.bitfinex.com/v2/", APITypeBitFinexV2, historyDB)
	require.NoError(t, err)
	// Update token list
	assert.NoError(t, pu.UpdateTokenList())
	// Update prices
	pu.UpdatePrices(context.Background())
	assertTokenHasPriceAndClean(t)
}

func TestPriceUpdaterCoingecko(t *testing.T) {
	// Init price updater
	pu, err := NewPriceUpdater("https://api.coingecko.com/api/v3/", APITypeCoingeckoV3, historyDB)
	require.NoError(t, err)
	// Update token list
	assert.NoError(t, pu.UpdateTokenList())
	// Update prices
	pu.UpdatePrices(context.Background())
	assertTokenHasPriceAndClean(t)
}

func assertTokenHasPriceAndClean(t *testing.T) {
	// Check that prices have been updated
	fetchedTokens, err := historyDB.GetTokensTest()
	require.NoError(t, err)
	// TokenID 0 (ETH) is always on the DB
	assert.Equal(t, 2, len(fetchedTokens))
	for _, token := range fetchedTokens {
		require.NotNil(t, token.USD)
		require.NotNil(t, token.USDUpdate)
		assert.Greater(t, *token.USD, 0.0)
	}
}
