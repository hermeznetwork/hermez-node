package synchronizer

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In order to run the test you need to run a Posgres DB with
// a database named "history" that is accessible by
// user: "hermez"
// pass: set it using the env var POSTGRES_PASS
// This can be achieved by running: POSTGRES_PASS=your_strong_pass && sudo docker run --rm --name hermez-db-test -p 5432:5432 -e POSTGRES_DB=history -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD=$POSTGRES_PASS -d postgres && sleep 2s && sudo docker exec -it hermez-db-test psql -a history -U hermez -c "CREATE DATABASE l2;"
// After running the test you can stop the container by running: sudo docker kill hermez-db-test
// If you already did that for the L2DB you don't have to do it again

// A env var called INFURA_PROJECT_ID must be set with a valid project id code

func Test(t *testing.T) {

	syncConfig := SyncConfig{
		LoopInterval: 5,
	}

	// Int State DB
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := statedb.NewStateDB(dir, true, 32)
	assert.Nil(t, err)

	// Init History DB
	pass := os.Getenv("POSTGRES_PASS")
	historyDB, err := historydb.NewHistoryDB(5432, "localhost", "hermez", pass, "history")
	require.Nil(t, err)
	err = historyDB.Reorg(0)
	assert.Nil(t, err)

	// Init eth client
	projectID := os.Getenv("INFURA_PROJECT_ID")
	ethClient, err := ethclient.Dial("https://mainnet.infura.io/v3/" + projectID)
	require.Nil(t, err)

	client := eth.NewClient(ethClient, nil, nil)

	// Let's update the FirstSavedBlock to have a faster testing

	latestBlock, err := client.BlockByNumber(context.Background(), nil)
	require.Nil(t, err)

	latestBlock, err = client.BlockByNumber(context.Background(), big.NewInt(int64(latestBlock.EthBlockNum-20)))
	require.Nil(t, err)

	syncConfig.FirstSavedBlock = *latestBlock

	// Create Synchronizer

	s := NewSynchronizer(&syncConfig, client, historyDB, sdb)

	// Test Sync
	err = s.sync()
	require.Nil(t, err)

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

	err = s.sync()
	require.Nil(t, err)

	// Close History DB
	if err := historyDB.Close(); err != nil {
		fmt.Println("Error closing the history DB:", err)
	}
}
