package priceupdater

import (
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/stretchr/testify/assert"
)

func TestPriceUpdater(t *testing.T) {
	// Init DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	assert.NoError(t, err)
	historyDB := historydb.NewHistoryDB(db)
	// Clean DB
	assert.NoError(t, historyDB.Reorg(-1))
	// Populate DB
	// Gen blocks and add them to DB
	blocks := test.GenBlocks(1, 2)
	assert.NoError(t, historyDB.AddBlocks(blocks))
	// Gen tokens and add them to DB
	tokens := []common.Token{}
	tokens = append(tokens, common.Token{
		TokenID:     0,
		EthBlockNum: blocks[0].EthBlockNum,
		EthAddr:     ethCommon.BigToAddress(big.NewInt(1)),
		Name:        "Ether",
		Symbol:      "ETH",
		Decimals:    18,
	})
	tokens = append(tokens, common.Token{
		TokenID:     1,
		EthBlockNum: blocks[0].EthBlockNum,
		EthAddr:     ethCommon.BigToAddress(big.NewInt(2)),
		Name:        "DAI",
		Symbol:      "DAI",
		Decimals:    18,
	})
	assert.NoError(t, historyDB.AddTokens(tokens))
	// Init price updater
	pu := NewPriceUpdater("https://api-pub.bitfinex.com/v2/", historyDB)
	// Update token list
	assert.NoError(t, pu.UpdateTokenList())
	// Update prices
	pu.UpdatePrices()
	// Check that prices have been updated
	fetchedTokens, err := historyDB.GetTokens()
	assert.NoError(t, err)
	for _, token := range fetchedTokens {
		assert.NotNil(t, token.USD)
		assert.NotNil(t, token.USDUpdate)
	}
}
