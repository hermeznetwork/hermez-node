package txselector

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetL2TxSelection(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	sdb, err := statedb.NewStateDB(dir, false, 0)
	assert.Nil(t, err)
	testL2DB := &l2db.L2DB{}
	// initTestDB(testL2DB, sdb)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	txsel, err := NewTxSelector(txselDir, sdb, testL2DB, 3, 3, 3)
	assert.Nil(t, err)
	fmt.Println(txsel)

	// txs, err := txsel.GetL2TxSelection(0)
	// assert.Nil(t, err)
	// for _, tx := range txs {
	//         fmt.Println(tx.FromIdx, tx.ToIdx, tx.AbsoluteFee)
	// }
	// assert.Equal(t, 3, len(txs))
	// assert.Equal(t, uint64(6), txs[0].AbsoluteFee)
	// assert.Equal(t, uint64(5), txs[1].AbsoluteFee)
	// assert.Equal(t, uint64(4), txs[2].AbsoluteFee)
}
