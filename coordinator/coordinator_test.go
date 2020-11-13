package coordinator

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestModules(t *testing.T) (*txselector.TxSelector, *batchbuilder.BatchBuilder) { // FUTURE once Synchronizer is ready, should return it also
	nLevels := 32

	synchDBPath, err := ioutil.TempDir("", "tmpSynchDB")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(synchDBPath))
	synchSdb, err := statedb.NewStateDB(synchDBPath, statedb.TypeSynchronizer, nLevels)
	assert.Nil(t, err)

	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(txselDir))
	txsel, err := txselector.NewTxSelector(txselDir, synchSdb, l2DB, 10, 10, 10)
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(bbDir))
	bb, err := batchbuilder.NewBatchBuilder(bbDir, synchSdb, nil, 0, uint64(nLevels))
	assert.Nil(t, err)

	// l1Txs, coordinatorL1Txs, poolL2Txs := test.GenerateTestTxsFromSet(t, test.SetTest0)

	return txsel, bb
}

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

var forger ethCommon.Address
var bidder ethCommon.Address

func waitForSlot(t *testing.T, coord *Coordinator, c *test.Client, slot int64) {
	for {
		blockNum, err := c.EthLastBlock()
		require.Nil(t, err)
		nextBlockSlot, err := c.AuctionGetSlotNumber(blockNum + 1)
		require.Nil(t, err)
		if nextBlockSlot == slot {
			break
		}
		c.CtlMineBlock()
		time.Sleep(100 * time.Millisecond)
		var stats synchronizer.Stats
		stats.Eth.LastBlock = c.CtlLastBlock()
		stats.Sync.LastBlock = c.CtlLastBlock()
		canForge, err := c.AuctionCanForge(forger, blockNum+1)
		require.Nil(t, err)
		if canForge {
			// fmt.Println("DBG canForge")
			stats.Sync.Auction.CurrentSlot.Forger = forger
		}
		coord.SendMsg(MsgSyncStats{
			Stats: stats,
		})
	}
}

func TestCoordinator(t *testing.T) {
	txsel, bb := newTestModules(t)
	bidder = ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	forger = ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")

	conf := Config{
		ForgerAddress: forger,
	}
	hdb := &historydb.HistoryDB{}
	serverProofs := []ServerProofInterface{&ServerProofMock{}, &ServerProofMock{}}

	var timer timer
	ethClientSetup := test.NewClientSetupExample()
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)

	// Bid for slot 2 and 4
	_, err := ethClient.AuctionSetCoordinator(forger, "https://foo.bar")
	require.Nil(t, err)
	_, err = ethClient.AuctionBidSimple(2, big.NewInt(9999))
	require.Nil(t, err)
	_, err = ethClient.AuctionBidSimple(4, big.NewInt(9999))
	require.Nil(t, err)

	scConsts := &synchronizer.SCConsts{
		Rollup:   *ethClientSetup.RollupConstants,
		Auction:  *ethClientSetup.AuctionConstants,
		WDelayer: *ethClientSetup.WDelayerConstants,
	}
	initSCVars := &synchronizer.SCVariables{
		Rollup:   *ethClientSetup.RollupVariables,
		Auction:  *ethClientSetup.AuctionVariables,
		WDelayer: *ethClientSetup.WDelayerVariables,
	}
	c := NewCoordinator(conf, hdb, txsel, bb, serverProofs, ethClient, scConsts, initSCVars)
	c.Start()
	time.Sleep(1 * time.Second)

	// NOTE: With the current test, the coordinator will enter in forge
	// time before the bidded slot because no one else is forging in the
	// other slots before the slot deadline.
	// simulate forgeSequence time
	waitForSlot(t, c, ethClient, 2)
	log.Info("~~~ simulate entering in forge time")
	time.Sleep(1 * time.Second)

	// simulate going out from forgeSequence
	waitForSlot(t, c, ethClient, 3)
	log.Info("~~~ simulate going out from forge time")
	time.Sleep(1 * time.Second)

	// simulate entering forgeSequence time again
	waitForSlot(t, c, ethClient, 4)
	log.Info("~~~ simulate entering in forge time again")
	time.Sleep(2 * time.Second)

	// simulate stopping forgerLoop by channel
	log.Info("~~~ simulate stopping forgerLoop by closing coordinator stopch")
	c.Stop()
	time.Sleep(1 * time.Second)
}
