package txselector

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T, testSet string, maxL1UserTxs, maxL1OperatorTxs, maxTxs uint64) *TxSelector {
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.Nil(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)

	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))
	sdb, err := statedb.NewStateDB(dir, statedb.TypeTxSelector, 0)
	require.Nil(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))
	txsel, err := NewTxSelector(txselDir, sdb, l2DB, maxL1UserTxs, maxL1OperatorTxs, maxTxs)
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
	assert.Nil(t, hdb.Reorg(-1))
	assert.Nil(t, hdb.AddBlock(&common.Block{
		EthBlockNum: 1,
	}))
	assert.Nil(t, hdb.AddTokens(tokens))
}

func TestGetL2TxSelection(t *testing.T) {
	txsel := initTest(t, til.SetPool0, 5, 5, 10)
	test.CleanL2DB(txsel.l2db.DB())

	tc := til.NewContext(eth.RollupConstMaxL1UserTx)
	// generate test transactions
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	assert.Nil(t, err)
	// poolL2Txs, err := tc.GeneratePoolL2Txs(til.SetPool0)
	// assert.Nil(t, err)

	coordIdxs := []common.Idx{256, 257, 258, 259}

	// add tokens to HistoryDB to avoid breaking FK constrains
	var tokens []common.Token
	for i := 0; i < int(tc.LastRegisteredTokenID); i++ {
		tokens = append(tokens, common.Token{
			TokenID:     common.TokenID(i),
			EthBlockNum: 1,
			EthAddr:     ethCommon.BytesToAddress([]byte{byte(i)}),
			Name:        strconv.Itoa(i),
			Symbol:      strconv.Itoa(i),
			Decimals:    18,
		})
	}
	addTokens(t, tokens, txsel.l2db.DB())

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	_, err = txsel.localAccountsDB.ProcessTxs(nil, nil, blocks[0].Batches[0].L1CoordinatorTxs, nil)
	require.Nil(t, err)

	// add the 1st batch of transactions to the TxSelector
	addL2Txs(t, txsel, common.L2TxsToPoolL2Txs(blocks[0].Batches[0].L2Txs))

	l1CoordTxs, l2Txs, err := txsel.GetL2TxSelection(coordIdxs, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(l2Txs))
	assert.Equal(t, 0, len(l1CoordTxs))

	_, _, _, err = txsel.GetL1L2TxSelection(coordIdxs, 0, blocks[0].L1UserTxs)
	assert.Nil(t, err)

	// TODO once L2DB is updated to return error in case that AddTxTest
	// fails, and the Til is updated, update this test, checking that the
	// selected PoolL2Tx are correctly sorted by Nonce

	// TODO once L2DB is updated to store the parameter AbsoluteFee (which
	// is used by TxSelector to sort L2Txs), uncomment this next lines of
	// test, and put the expected value for
	// l2Txs[len(l2Txs)-1].AbsoluteFee, which is the Tx which has the
	// Fee==192.
	/*
		// add the 3rd batch of transactions to the TxSelector
		addL2Txs(t, txsel, common.L2TxsToPoolL2Txs(blocks[0].Batches[2].L2Txs))

		_, l2Txs, err = txsel.GetL2TxSelection(coordIdxs, 0)
		assert.Nil(t, err)
		for _, tx := range l2Txs {
			fmt.Println(tx.FromIdx, tx.ToIdx, tx.AbsoluteFee)
		}
		require.Equal(t, 10, len(l2Txs))
		assert.Equal(t, float64(0), l2Txs[0].AbsoluteFee)

		fmt.Println(l2Txs[len(l2Txs)-1].Amount)
		assert.Equal(t, float64(4), l2Txs[len(l2Txs)-1].AbsoluteFee)
	*/
}
