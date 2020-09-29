package synchronizer

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := statedb.NewStateDB(dir, true, 32)
	assert.Nil(t, err)

	// Init History DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	historyDB := historydb.NewHistoryDB(db)
	err = historyDB.Reorg(0)
	assert.Nil(t, err)

	// Init eth client
	ehtClientDialURL := os.Getenv("ETHCLIENT_DIAL_URL")
	ethClient, err := ethclient.Dial(ehtClientDialURL)
	require.Nil(t, err)

	client := eth.NewClient(ethClient, nil, nil, nil)

	// Create Synchronizer
	s := NewSynchronizer(client, historyDB, sdb)
	require.NotNil(t, s)

	// Test Sync
	// err = s.Sync()
	// require.Nil(t, err)

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
