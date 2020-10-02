package txselector

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T, testSet string) *TxSelector {
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)

	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	sdb, err := statedb.NewStateDB(dir, statedb.TypeTxSelector, 0)
	require.Nil(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	txsel, err := NewTxSelector(txselDir, sdb, l2DB, 100, 100, 1000)
	require.Nil(t, err)

	return txsel
}
func addL2Txs(t *testing.T, txsel *TxSelector, poolL2Txs []common.PoolL2Tx) {
	for i := 0; i < len(poolL2Txs); i++ {
		err := txsel.l2db.AddTxTest(&poolL2Txs[i])
		require.Nil(t, err)
	}
}

func addTokens(t *testing.T, tokens []common.Token, db *sqlx.DB) {
	hdb := historydb.NewHistoryDB(db)
	assert.NoError(t, hdb.Reorg(-1))
	assert.NoError(t, hdb.AddBlock(&common.Block{
		EthBlockNum: 1,
	}))
	assert.NoError(t, hdb.AddTokens(tokens))
}

func TestGetL2TxSelection(t *testing.T) {
	txsel := initTest(t, test.SetTest0)
	test.CleanL2DB(txsel.l2db.DB())

	// generate test transactions
	l1Txs, _, poolL2Txs, tokens := test.GenerateTestTxsFromSet(t, test.SetTest0)

	// add tokens to HistoryDB to avoid breaking FK constrains
	addTokens(t, tokens, txsel.l2db.DB())
	// add the first batch of transactions to the TxSelector
	addL2Txs(t, txsel, poolL2Txs[0])

	_, err := txsel.GetL2TxSelection(0)
	assert.Nil(t, err)

	_, _, _, err = txsel.GetL1L2TxSelection(0, l1Txs[0])
	assert.Nil(t, err)

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
