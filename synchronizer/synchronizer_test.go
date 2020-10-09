package synchronizer

import (
	"context"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/test"
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

func TestSync(t *testing.T) {
	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

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

	// Test Sync for ethereum genesis block
	err = s.Sync(context.Background())
	require.Nil(t, err)

	blocks, err := s.historyDB.GetBlocks(0, 9999)
	require.Nil(t, err)
	assert.Equal(t, int64(0), blocks[0].EthBlockNum)

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
