package txselector

/*
   TODO update transactions generation
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
	txsel := initTest(t, transakcio.SetPool0, 5, 5, 10)
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

		// TODO once L2DB is updated to return error in case that AddTxTest
		// fails, and the Transakcio is updated, update this test, checking that the
		// selected PoolL2Tx are correctly sorted by Nonce



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
*/
