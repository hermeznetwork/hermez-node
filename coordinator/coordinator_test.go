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
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestModules(t *testing.T) (*txselector.TxSelector, *batchbuilder.BatchBuilder) { // FUTURE once Synchronizer is ready, should return it also
	nLevels := 32

	synchDBPath, err := ioutil.TempDir("", "tmpSynchDB")
	require.Nil(t, err)
	synchSdb, err := statedb.NewStateDB(synchDBPath, statedb.TypeSynchronizer, nLevels)
	assert.Nil(t, err)

	// txselDBPath, err := ioutil.TempDir("", "tmpTxSelDB")
	// require.Nil(t, err)
	// bbDBPath, err := ioutil.TempDir("", "tmpBBDB")
	// require.Nil(t, err)
	// txselSdb, err := statedb.NewLocalStateDB(txselDBPath, synchSdb, statedb.TypeTxSelector, nLevels)
	// assert.Nil(t, err)
	// bbSdb, err := statedb.NewLocalStateDB(bbDBPath, synchSdb, statedb.TypeBatchBuilder, nLevels)
	// assert.Nil(t, err)

	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	txsel, err := txselector.NewTxSelector(txselDir, synchSdb, l2DB, 10, 10, 10)
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	bb, err := batchbuilder.NewBatchBuilder(bbDir, synchSdb, nil, 0, uint64(nLevels))
	assert.Nil(t, err)

	// l1Txs, coordinatorL1Txs, poolL2Txs := test.GenerateTestTxsFromSet(t, test.SetTest0)

	return txsel, bb
}

// CoordNode is an example of a Node that handles the goroutines for the coordinator
type CoordNode struct {
	c                     *Coordinator
	stopForge             chan bool
	stopGetProofCallForge chan bool
	stopForgeCallConfirm  chan bool
}

func NewCoordNode(c *Coordinator) *CoordNode {
	return &CoordNode{
		c: c,
	}
}

func (cn *CoordNode) Start() {
	log.Debugw("Starting CoordNode...")
	cn.stopForge = make(chan bool)
	cn.stopGetProofCallForge = make(chan bool)
	cn.stopForgeCallConfirm = make(chan bool)
	queueSize := 8
	batchCh0 := make(chan *BatchInfo, queueSize)
	batchCh1 := make(chan *BatchInfo, queueSize)

	go func() {
		for {
			select {
			case <-cn.stopForge:
				return
			default:
				if forge, err := cn.c.ForgeLoopFn(batchCh0, cn.stopForge); err == ErrStop {
					return
				} else if err != nil {
					log.Errorw("CoordNode ForgeLoopFn", "error", err)
					time.Sleep(200 * time.Millisecond) // Avoid overflowing log with errors
				} else if !forge {
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-cn.stopGetProofCallForge:
				return
			default:
				if err := cn.c.GetProofCallForgeLoopFn(
					batchCh0, batchCh1, cn.stopGetProofCallForge); err == ErrStop {
					return
				} else if err != nil {
					log.Errorw("CoordNode GetProofCallForgeLoopFn", "error", err)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-cn.stopForgeCallConfirm:
				return
			default:
				if err := cn.c.ForgeCallConfirmLoopFn(
					batchCh1, cn.stopForgeCallConfirm); err == ErrStop {
					return
				} else if err != nil {
					log.Errorw("CoordNode ForgeCallConfirmLoopFn", "error", err)
				}
			}
		}
	}()
}

func (cn *CoordNode) Stop() {
	log.Debugw("Stopping CoordNode...")
	cn.stopForge <- true
	cn.stopGetProofCallForge <- true
	cn.stopForgeCallConfirm <- true
}

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

func waitForSlot(t *testing.T, c *test.Client, slot int64) {
	for {
		blockNum, err := c.EthCurrentBlock()
		require.Nil(t, err)
		nextBlockSlot, err := c.AuctionGetSlotNumber(blockNum + 1)
		require.Nil(t, err)
		if nextBlockSlot == slot {
			break
		}
		c.CtlMineBlock()
	}
}

func TestCoordinator(t *testing.T) {
	txsel, bb := newTestModules(t)

	conf := Config{}
	hdb := &historydb.HistoryDB{}
	serverProofs := []ServerProofInterface{&ServerProofMock{}, &ServerProofMock{}}

	var timer timer
	ethClientSetup := test.NewClientSetupExample()
	addr := ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")
	ethClient := test.NewClient(true, &timer, &addr, ethClientSetup)

	// Bid for slot 2 and 4
	_, err := ethClient.AuctionRegisterCoordinator(addr, "https://foo.bar")
	require.Nil(t, err)
	_, err = ethClient.AuctionBid(2, big.NewInt(9999), addr)
	require.Nil(t, err)
	_, err = ethClient.AuctionBid(4, big.NewInt(9999), addr)
	require.Nil(t, err)

	c := NewCoordinator(conf, hdb, txsel, bb, serverProofs, ethClient)
	cn := NewCoordNode(c)
	cn.Start()
	time.Sleep(1 * time.Second)

	// simulate forgeSequence time
	waitForSlot(t, ethClient, 2)
	log.Info("simulate entering in forge time")
	time.Sleep(1 * time.Second)

	// simulate going out from forgeSequence
	waitForSlot(t, ethClient, 3)
	log.Info("simulate going out from forge time")
	time.Sleep(1 * time.Second)

	// simulate entering forgeSequence time again
	waitForSlot(t, ethClient, 4)
	log.Info("simulate entering in forge time again")
	time.Sleep(2 * time.Second)

	// simulate stopping forgerLoop by channel
	log.Info("simulate stopping forgerLoop by closing coordinator stopch")
	cn.Stop()
	time.Sleep(1 * time.Second)
}
