package coordinator

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestModules(t *testing.T) (*txselector.TxSelector, *batchbuilder.BatchBuilder) { // FUTURE once Synchronizer is ready, should return it also
	nLevels := 32

	synchDB, err := ioutil.TempDir("", "tmpSynchDB")
	require.Nil(t, err)
	sdb, err := statedb.NewStateDB(synchDB, true, nLevels)
	assert.Nil(t, err)

	pass := os.Getenv("POSTGRES_PASS")
	l2DB, err := l2db.NewL2DB(5432, "localhost", "hermez", pass, "l2", 10, 512, 24*time.Hour)
	require.Nil(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	txsel, err := txselector.NewTxSelector(txselDir, sdb, l2DB, 10, 10, 10)
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	bb, err := batchbuilder.NewBatchBuilder(bbDir, sdb, nil, 0, uint64(nLevels))
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
	batchCh0 := make(chan *BatchInfo)
	batchCh1 := make(chan *BatchInfo)

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
				} else if !forge {
					time.Sleep(500 * time.Millisecond)
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

func TestCoordinator(t *testing.T) {
	txsel, bb := newTestModules(t)

	conf := Config{}
	hdb := &historydb.HistoryDB{}
	c := NewCoordinator(conf, hdb, txsel, bb, &eth.Client{})
	cn := NewCoordNode(c)
	cn.Start()
	time.Sleep(1 * time.Second)

	// simulate forgeSequence time
	log.Info("simulate entering in forge time")
	c.isForgeSeq = true
	time.Sleep(1 * time.Second)

	// simulate going out from forgeSequence
	log.Info("simulate going out from forge time")
	c.isForgeSeq = false
	time.Sleep(1 * time.Second)

	// simulate entering forgeSequence time again
	log.Info("simulate entering in forge time again")
	c.isForgeSeq = true
	time.Sleep(1 * time.Second)

	// simulate stopping forgerLoop by channel
	log.Info("simulate stopping forgerLoop by closing coordinator stopch")
	cn.Stop()
	time.Sleep(1 * time.Second)
}
