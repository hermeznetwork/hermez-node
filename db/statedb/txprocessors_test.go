package statedb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var debug = false

func TestProcessTxs(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	parser := test.NewParser(strings.NewReader(test.SetTest0))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1Txs, coordinatorL1Txs, poolL2Txs, _ := test.GenerateTestTxs(t, instructions)
	assert.Equal(t, 29, len(l1Txs[0]))
	assert.Equal(t, 0, len(coordinatorL1Txs[0]))
	assert.Equal(t, 21, len(poolL2Txs[0]))

	// iterate for each batch
	for i := 0; i < len(l1Txs); i++ {
		// l2Txs := common.PoolL2TxsToL2Txs(poolL2Txs[i])

		_, _, err := sdb.ProcessTxs(l1Txs[i], coordinatorL1Txs[i], poolL2Txs[i])
		require.Nil(t, err)
	}

	accountIdx := common.Idx(256)
	acc, err := sdb.GetAccount(&accountIdx)
	assert.Nil(t, err)
	assert.Equal(t, "23", acc.Balance.String())
}

func TestProcessTxsSynchronizer(t *testing.T) {
	// TODO once TTGL is updated, use the blockchain L2Tx (not PoolL2Tx) for
	// the Synchronizer tests

	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	parser := test.NewParser(strings.NewReader(test.SetTest0))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1Txs, coordinatorL1Txs, poolL2Txs, _ := test.GenerateTestTxs(t, instructions)
	assert.Equal(t, 29, len(l1Txs[0]))
	assert.Equal(t, 0, len(coordinatorL1Txs[0]))
	assert.Equal(t, 21, len(poolL2Txs[0]))
	assert.Equal(t, 5, len(l1Txs[1]))
	assert.Equal(t, 1, len(coordinatorL1Txs[1]))
	assert.Equal(t, 55, len(poolL2Txs[1]))
	assert.Equal(t, 10, len(l1Txs[2]))
	assert.Equal(t, 0, len(coordinatorL1Txs[2]))
	assert.Equal(t, 7, len(poolL2Txs[2]))

	// use first batch
	// l2txs := common.PoolL2TxsToL2Txs(poolL2Txs[0])
	_, exitInfos, err := sdb.ProcessTxs(l1Txs[0], coordinatorL1Txs[0], poolL2Txs[0])
	require.Nil(t, err)
	// TODO once TTGL is updated, add a check that a input poolL2Tx with
	// Nonce & TokenID =0, after ProcessTxs call has the expected value

	assert.Equal(t, 0, len(exitInfos))
	accountIdx := common.Idx(256)
	acc, err := sdb.GetAccount(&accountIdx)
	assert.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	// use second batch
	// l2txs = common.PoolL2TxsToL2Txs(poolL2Txs[1])
	_, exitInfos, err = sdb.ProcessTxs(l1Txs[1], coordinatorL1Txs[1], poolL2Txs[1])
	require.Nil(t, err)
	assert.Equal(t, 5, len(exitInfos))
	acc, err = sdb.GetAccount(&accountIdx)
	require.Nil(t, err)
	assert.Equal(t, "48", acc.Balance.String())

	// use third batch
	// l2txs = common.PoolL2TxsToL2Txs(poolL2Txs[2])
	_, exitInfos, err = sdb.ProcessTxs(l1Txs[2], coordinatorL1Txs[2], poolL2Txs[2])
	require.Nil(t, err)
	assert.Equal(t, 1, len(exitInfos))
	acc, err = sdb.GetAccount(&accountIdx)
	assert.Nil(t, err)
	assert.Equal(t, "23", acc.Balance.String())
}

func TestProcessTxsBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	parser := test.NewParser(strings.NewReader(test.SetTest0))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1Txs, coordinatorL1Txs, poolL2Txs, _ := test.GenerateTestTxs(t, instructions)
	assert.Equal(t, 29, len(l1Txs[0]))
	assert.Equal(t, 0, len(coordinatorL1Txs[0]))
	assert.Equal(t, 21, len(poolL2Txs[0]))
	assert.Equal(t, 5, len(l1Txs[1]))
	assert.Equal(t, 1, len(coordinatorL1Txs[1]))
	assert.Equal(t, 55, len(poolL2Txs[1]))
	assert.Equal(t, 10, len(l1Txs[2]))
	assert.Equal(t, 0, len(coordinatorL1Txs[2]))
	assert.Equal(t, 7, len(poolL2Txs[2]))

	// use first batch
	// l2txs := common.PoolL2TxsToL2Txs(poolL2Txs[0])
	_, exitInfos, err := sdb.ProcessTxs(l1Txs[0], coordinatorL1Txs[0], poolL2Txs[0])
	require.Nil(t, err)
	assert.Equal(t, 0, len(exitInfos))
	accountIdx := common.Idx(256)
	acc, err := sdb.GetAccount(&accountIdx)
	assert.Nil(t, err)
	assert.Equal(t, "28", acc.Balance.String())

	// use second batch
	// l2txs = common.PoolL2TxsToL2Txs(poolL2Txs[1])
	_, exitInfos, err = sdb.ProcessTxs(l1Txs[1], coordinatorL1Txs[1], poolL2Txs[1])
	require.Nil(t, err)
	assert.Equal(t, 5, len(exitInfos))
	acc, err = sdb.GetAccount(&accountIdx)
	require.Nil(t, err)
	assert.Equal(t, "48", acc.Balance.String())

	// use third batch
	// l2txs = common.PoolL2TxsToL2Txs(poolL2Txs[2])
	_, exitInfos, err = sdb.ProcessTxs(l1Txs[2], coordinatorL1Txs[2], poolL2Txs[2])
	require.Nil(t, err)
	assert.Equal(t, 1, len(exitInfos))
	acc, err = sdb.GetAccount(&accountIdx)
	assert.Nil(t, err)
	assert.Equal(t, "23", acc.Balance.String())
}

func TestZKInputsGeneration(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// generate test transactions from test.SetTest0 code
	parser := test.NewParser(strings.NewReader(test.SetTest0))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1Txs, coordinatorL1Txs, poolL2Txs, _ := test.GenerateTestTxs(t, instructions)
	assert.Equal(t, 29, len(l1Txs[0]))
	assert.Equal(t, 0, len(coordinatorL1Txs[0]))
	assert.Equal(t, 21, len(poolL2Txs[0]))

	zki, _, err := sdb.ProcessTxs(l1Txs[0], coordinatorL1Txs[0], poolL2Txs[0])
	require.Nil(t, err)

	s, err := json.Marshal(zki)
	require.Nil(t, err)
	if debug {
		fmt.Println(string(s))
	}
}
