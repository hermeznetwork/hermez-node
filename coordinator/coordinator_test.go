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

func TestCoordinator(t *testing.T) {
	txsel, bb := newTestModules(t)

	conf := Config{
		LoopInterval: 100 * time.Millisecond,
	}
	hdb := &historydb.HistoryDB{}
	c := NewCoordinator(conf, hdb, txsel, bb, &eth.Client{})
	c.Start()
	time.Sleep(1 * time.Second)

	// simulate forgeSequence time
	log.Debug("simulate entering in forge time")
	c.isForgeSeq = true
	time.Sleep(1 * time.Second)

	// simulate going out from forgeSequence
	log.Debug("simulate going out from forge time")
	c.isForgeSeq = false
	time.Sleep(1 * time.Second)

	// simulate entering forgeSequence time again
	log.Debug("simulate entering in forge time again")
	c.isForgeSeq = true
	time.Sleep(1 * time.Second)

	// simulate stopping forgerLoop by channel
	log.Debug("simulate stopping forgerLoop by closing coordinator stopch")
	c.Stop()
	time.Sleep(1 * time.Second)
}
