package synchronizer

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

type tokenData struct {
	TokenID common.TokenID
	Addr    ethCommon.Address
	Consts  eth.ERC20Consts
}

func TestSync(t *testing.T) {
	ctx := context.Background()
	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	stateDB, err := statedb.NewStateDB(dir, statedb.TypeSynchronizer, 32)
	assert.Nil(t, err)

	// Init History DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	historyDB := historydb.NewHistoryDB(db)
	// Clear DB
	err = historyDB.Reorg(-1)
	assert.Nil(t, err)

	// Init eth client
	var timer timer
	clientSetup := test.NewClientSetupExample()
	client := test.NewClient(true, &timer, &ethCommon.Address{}, clientSetup)

	// Create Synchronizer
	s, err := NewSynchronizer(client, historyDB, stateDB)
	require.Nil(t, err)

	// Test Sync for rollup genesis block
	blockData, _, err := s.Sync2(ctx, nil)
	require.Nil(t, err)
	require.NotNil(t, blockData)
	assert.Equal(t, int64(1), blockData.Block.EthBlockNum)
	blocks, err := s.historyDB.GetBlocks(0, 9999)
	require.Nil(t, err)
	assert.Equal(t, 1, len(blocks))
	assert.Equal(t, int64(1), blocks[0].EthBlockNum)

	/*
		// Test Sync for a block with new Tokens and L1UserTxs
		// accounts := test.GenerateKeys(t, []string{"A", "B", "C", "D"})
		l1UserTxs, _, _, _ := test.GenerateTestTxsFromSet(t, `
		A (1): 10
		A (2): 20
		B (1): 5
		C (1): 8
		D (3): 15
		> advance batch
			`)
		require.Greater(t, len(l1UserTxs[0]), 0)
			// require.Greater(t, len(tokens), 0)

			for i := 1; i <= 3; i++ {
				_, err := client.RollupAddToken(ethCommon.BigToAddress(big.NewInt(int64(i*10000))),
					clientSetup.RollupVariables.FeeAddToken)
				require.Nil(t, err)
			}

			for i := range l1UserTxs[0] {
				client.CtlAddL1TxUser(&l1UserTxs[0][i])
			}
			client.CtlMineBlock()

			err = s.Sync(context.Background())
			require.Nil(t, err)

			getTokens, err := s.historyDB.GetTokens()
			require.Nil(t, err)
			assert.Equal(t, 3, len(getTokens))
	*/

	// Generate tokens vector
	numTokens := 3
	tokens := make([]tokenData, numTokens)
	for i := 1; i <= numTokens; i++ {
		addr := ethCommon.BigToAddress(big.NewInt(int64(i * 10000)))
		consts := eth.ERC20Consts{
			Name:     fmt.Sprintf("Token %d", i),
			Symbol:   fmt.Sprintf("TK%d", i),
			Decimals: uint64(i * 2),
		}
		tokens[i-1] = tokenData{common.TokenID(i), addr, consts}
	}

	numUsers := 4
	keys := make([]*userKeys, numUsers)
	for i := range keys {
		keys[i] = genKeys(i)
	}

	// Generate some L1UserTxs of type deposit
	l1UserTxs := make([]*common.L1Tx, 5)
	for i := range l1UserTxs {
		l1UserTxs[i] = &common.L1Tx{
			FromIdx:     common.Idx(0),
			FromEthAddr: keys[i%numUsers].Addr,
			FromBJJ:     keys[i%numUsers].BJJPK,
			Amount:      big.NewInt(0),
			LoadAmount:  big.NewInt((int64(i) + 1) * 1000),
			TokenID:     common.TokenID(i%numTokens + 1),
		}
	}

	// Add tokens to ethereum, and to rollup
	for _, token := range tokens {
		client.CtlAddERC20(token.Addr, token.Consts)
		_, err := client.RollupAddToken(token.Addr, clientSetup.RollupVariables.FeeAddToken)
		require.Nil(t, err)
	}

	// Add L1Txs to rollup
	for i := range l1UserTxs {
		tx := l1UserTxs[i]
		_, err := client.RollupL1UserTxERC20ETH(tx.FromBJJ, int64(tx.FromIdx), tx.LoadAmount, tx.Amount,
			uint32(tx.TokenID), int64(tx.ToIdx))
		require.Nil(t, err)
	}

	// Mine block and sync
	client.CtlMineBlock()

	blockData, _, err = s.Sync2(ctx, nil)
	require.Nil(t, err)
	require.NotNil(t, blockData)
	assert.Equal(t, int64(2), blockData.Block.EthBlockNum)

	// Check tokens in DB
	dbTokens, err := s.historyDB.GetAllTokens()
	require.Nil(t, err)
	assert.Equal(t, len(tokens), len(dbTokens))
	assert.Equal(t, len(tokens), len(blockData.AddedTokens))
	for i := range tokens {
		token := tokens[i]
		addToken := blockData.AddedTokens[i]
		dbToken := dbTokens[i]

		assert.Equal(t, int64(2), addToken.EthBlockNum)
		assert.Equal(t, token.TokenID, addToken.TokenID)
		assert.Equal(t, token.Addr, addToken.EthAddr)
		assert.Equal(t, token.Consts.Name, addToken.Name)
		assert.Equal(t, token.Consts.Symbol, addToken.Symbol)
		assert.Equal(t, token.Consts.Decimals, addToken.Decimals)

		var addTokenCpy historydb.TokenRead
		require.Nil(t, copier.Copy(&addTokenCpy, &addToken)) // copy common.Token to historydb.TokenRead
		addTokenCpy.ItemID = dbToken.ItemID                  // we don't care about ItemID
		assert.Equal(t, addTokenCpy, dbToken)
	}

	// Check L1UserTxs in DB

	// TODO: Reorg will be properly tested once we have the mock ethClient implemented
	/*
		// Force a Reorg
		lastSavedBlock, err := historyDB.GetLastBlock()
		require.Nil(t, err)

		lastSavedBlock.EthBlockNum++
		err = historyDB.AddBlock(lastSavedBlock)
		require.Nil(t, err)

		lastSavedBlock.EthBlockNum++
		err = historyDB.AddBlock(lastSavedBlock)
		require.Nil(t, err)

		log.Debugf("Wait for the blockchain to generate some blocks...")
		time.Sleep(40 * time.Second)


		err = s.Sync()
		require.Nil(t, err)
	*/
}

type userKeys struct {
	BJJSK *babyjub.PrivateKey
	BJJPK *babyjub.PublicKey
	Addr  ethCommon.Address
}

func genKeys(i int) *userKeys {
	i++ // i = 0 doesn't work for the ecdsa key generation
	var sk babyjub.PrivateKey
	binary.LittleEndian.PutUint64(sk[:], uint64(i))

	// eth address
	var key ecdsa.PrivateKey
	key.D = big.NewInt(int64(i)) // only for testing
	key.PublicKey.X, key.PublicKey.Y = ethCrypto.S256().ScalarBaseMult(key.D.Bytes())
	key.Curve = ethCrypto.S256()
	addr := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &userKeys{
		BJJSK: &sk,
		BJJPK: sk.Public(),
		Addr:  addr,
	}
}
