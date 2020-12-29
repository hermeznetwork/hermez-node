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
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTest(t *testing.T, chainID uint16, testSet string) *TxSelector {
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	require.NoError(t, err)
	l2DB := l2db.NewL2DB(db, 10, 100, 24*time.Hour)

	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))
	sdb, err := statedb.NewStateDB(dir, 128, statedb.TypeTxSelector, 0)
	require.NoError(t, err)

	txselDir, err := ioutil.TempDir("", "tmpTxSelDB")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	coordAccount := &CoordAccount{ // TODO TMP
		Addr:                ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
		BJJ:                 common.EmptyBJJComp,
		AccountCreationAuth: nil,
	}

	txsel, err := NewTxSelector(coordAccount, txselDir, sdb, l2DB)
	require.NoError(t, err)

	return txsel
}
func addL2Txs(t *testing.T, txsel *TxSelector, poolL2Txs []common.PoolL2Tx) {
	for i := 0; i < len(poolL2Txs); i++ {
		err := txsel.l2db.AddTxTest(&poolL2Txs[i])
		require.NoError(t, err)
	}
}

func addTokens(t *testing.T, tokens []common.Token, db *sqlx.DB) {
	hdb := historydb.NewHistoryDB(db)
	test.WipeDB(hdb.DB())
	assert.NoError(t, hdb.AddBlock(&common.Block{
		Num: 1,
	}))
	assert.NoError(t, hdb.AddTokens(tokens))
}

func TestCoordIdxsDB(t *testing.T) {
	chainID := uint16(0)
	txsel := initTest(t, chainID, til.SetPool0)
	test.WipeDB(txsel.l2db.DB())

	coordIdxs := make(map[common.TokenID]common.Idx)
	coordIdxs[common.TokenID(0)] = common.Idx(256)
	coordIdxs[common.TokenID(1)] = common.Idx(257)
	coordIdxs[common.TokenID(2)] = common.Idx(258)

	err := txsel.AddCoordIdxs(coordIdxs)
	assert.NoError(t, err)

	r, err := txsel.GetCoordIdxs()
	assert.NoError(t, err)
	assert.Equal(t, coordIdxs, r)
}

func TestGetL2TxSelection(t *testing.T) {
	chainID := uint16(0)
	txsel := initTest(t, chainID, til.SetPool0)
	test.WipeDB(txsel.l2db.DB())

	tc := til.NewContext(chainID, common.RollupConstMaxL1UserTx)
	// generate test transactions
	blocks, err := tc.GenerateBlocks(til.SetBlockchain0)
	assert.NoError(t, err)
	// poolL2Txs, err := tc.GeneratePoolL2Txs(til.SetPool0)
	// assert.NoError(t, err)

	coordIdxs := make(map[common.TokenID]common.Idx)
	coordIdxs[common.TokenID(0)] = common.Idx(256)
	coordIdxs[common.TokenID(1)] = common.Idx(257)
	coordIdxs[common.TokenID(2)] = common.Idx(258)
	coordIdxs[common.TokenID(3)] = common.Idx(259)
	err = txsel.AddCoordIdxs(coordIdxs)
	assert.NoError(t, err)

	// add tokens to HistoryDB to avoid breaking FK constrains
	var tokens []common.Token
	for i := 0; i < int(tc.LastRegisteredTokenID); i++ {
		tokens = append(tokens, common.Token{
			TokenID:     common.TokenID(i + 1),
			EthBlockNum: 1,
			EthAddr:     ethCommon.BytesToAddress([]byte{byte(i + 1)}),
			Name:        strconv.Itoa(i),
			Symbol:      strconv.Itoa(i),
			Decimals:    18,
		})
	}
	addTokens(t, tokens, txsel.l2db.DB())

	tpc := txprocessor.Config{
		NLevels:  32,
		MaxFeeTx: 64,
		MaxTx:    512,
		MaxL1Tx:  64,
		ChainID:  chainID,
	}
	selectionConfig := &SelectionConfig{
		MaxL1UserTxs:        32,
		MaxL1CoordinatorTxs: 32,
		TxProcessorConfig:   tpc,
	}
	txselStateDB := txsel.localAccountsDB.StateDB
	tp := txprocessor.NewTxProcessor(txselStateDB, selectionConfig.TxProcessorConfig)

	// Process the 1st batch, which contains the L1CoordinatorTxs necessary
	// to create the Coordinator accounts to receive the fees
	_, err = tp.ProcessTxs(nil, nil, blocks[0].Rollup.Batches[0].L1CoordinatorTxs, nil)
	require.NoError(t, err)

	// add the 1st batch of transactions to the TxSelector
	addL2Txs(t, txsel, common.L2TxsToPoolL2Txs(blocks[0].Rollup.Batches[0].L2Txs))

	_, _, l1CoordTxs, l2Txs, err := txsel.GetL2TxSelection(selectionConfig, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(l2Txs))
	assert.Equal(t, 0, len(l1CoordTxs))

	_, _, _, _, _, err = txsel.GetL1L2TxSelection(selectionConfig, 0, blocks[0].Rollup.L1UserTxs)
	assert.NoError(t, err)

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
		assert.NoError(t, err)
		for _, tx := range l2Txs {
			fmt.Println(tx.FromIdx, tx.ToIdx, tx.AbsoluteFee)
		}
		require.Equal(t, 10, len(l2Txs))
		assert.Equal(t, float64(0), l2Txs[0].AbsoluteFee)

		fmt.Println(l2Txs[len(l2Txs)-1].Amount)
		assert.Equal(t, float64(4), l2Txs[len(l2Txs)-1].AbsoluteFee)
	*/
}
