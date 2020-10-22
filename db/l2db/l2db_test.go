package l2db

import (
	"math/big"
	"os"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var l2DB *L2DB
var tokens []common.Token
var tokensUSD []historydb.TokenRead

func TestMain(m *testing.M) {
	// init DB
	pass := os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	l2DB = NewL2DB(db, 10, 100, 24*time.Hour)
	tokens, tokensUSD = prepareHistoryDB(db)
	if err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		log.Error("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func prepareHistoryDB(db *sqlx.DB) ([]common.Token, []historydb.TokenRead) {
	historyDB := historydb.NewHistoryDB(db)
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Clean historyDB
	if err := historyDB.Reorg(-1); err != nil {
		panic(err)
	}
	// Store blocks to historyDB
	blocks := test.GenBlocks(fromBlock, toBlock)
	if err := historyDB.AddBlocks(blocks); err != nil {
		panic(err)
	}
	// Store tokens to historyDB
	const nTokens = 5
	tokens := test.GenTokens(nTokens, blocks)
	if err := historyDB.AddTokens(tokens); err != nil {
		panic(err)
	}
	readTokens := []historydb.TokenRead{}
	for i, token := range tokens {
		readToken := historydb.TokenRead{
			TokenID:     token.TokenID,
			EthBlockNum: token.EthBlockNum,
			EthAddr:     token.EthAddr,
			Name:        token.Name,
			Symbol:      token.Symbol,
			Decimals:    token.Decimals,
		}
		if i%2 != 0 {
			value := float64(i) * 5.4321
			if err := historyDB.UpdateTokenValue(token.Symbol, value); err != nil {
				panic(err)
			}
			now := time.Now().UTC()
			readToken.USDUpdate = &now
			readToken.USD = &value
		}
		readTokens = append(readTokens, readToken)
	}
	return tokens, readTokens
}

func TestAddTxTest(t *testing.T) {
	// Gen poolTxs
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	for _, tx := range txs {
		err := l2DB.AddTxTest(tx)
		assert.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(tx.TxID)
		assert.NoError(t, err)
		assertReadTx(t, commonToRead(tx, tokens), fetchedTx)
	}
}

func commonToRead(commonTx *common.PoolL2Tx, tokens []common.Token) *PoolL2TxRead {
	readTx := &PoolL2TxRead{
		TxID:      commonTx.TxID,
		FromIdx:   commonTx.FromIdx,
		ToBJJ:     commonTx.ToBJJ,
		Amount:    commonTx.Amount,
		Fee:       commonTx.Fee,
		Nonce:     commonTx.Nonce,
		State:     commonTx.State,
		Signature: commonTx.Signature,
		RqToBJJ:   commonTx.RqToBJJ,
		RqAmount:  commonTx.RqAmount,
		Type:      commonTx.Type,
		Timestamp: commonTx.Timestamp,
		TokenID:   commonTx.TokenID,
	}
	// token related fields
	// find token
	token := historydb.TokenRead{}
	for _, tkn := range tokensUSD {
		if tkn.TokenID == readTx.TokenID {
			token = tkn
			break
		}
	}
	// set token related fields
	readTx.TokenEthBlockNum = token.EthBlockNum
	readTx.TokenEthAddr = token.EthAddr
	readTx.TokenName = token.Name
	readTx.TokenSymbol = token.Symbol
	readTx.TokenDecimals = token.Decimals
	readTx.TokenUSD = token.USD
	readTx.TokenUSDUpdate = token.USDUpdate
	// nullable fields
	if commonTx.ToIdx != 0 {
		readTx.ToIdx = &commonTx.ToIdx
	}
	nilAddr := ethCommon.BigToAddress(big.NewInt(0))
	if commonTx.ToEthAddr != nilAddr {
		readTx.ToEthAddr = &commonTx.ToEthAddr
	}
	if commonTx.RqFromIdx != 0 {
		readTx.RqFromIdx = &commonTx.RqFromIdx
	}
	if commonTx.RqToIdx != 0 { // if true, all Rq... fields must be different to nil
		readTx.RqToIdx = &commonTx.RqToIdx
		readTx.RqTokenID = &commonTx.RqTokenID
		readTx.RqFee = &commonTx.RqFee
		readTx.RqNonce = &commonTx.RqNonce
	}
	if commonTx.RqToEthAddr != nilAddr {
		readTx.RqToEthAddr = &commonTx.RqToEthAddr
	}
	return readTx
}

func assertReadTx(t *testing.T, expected, actual *PoolL2TxRead) {
	// Check that timestamp has been set within the last 3 seconds
	assert.Less(t, time.Now().UTC().Unix()-3, actual.Timestamp.Unix())
	assert.GreaterOrEqual(t, time.Now().UTC().Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	// Check token related stuff
	if expected.TokenUSDUpdate != nil {
		// Check that TokenUSDUpdate has been set within the last 3 seconds
		assert.Less(t, time.Now().UTC().Unix()-3, actual.TokenUSDUpdate.Unix())
		assert.GreaterOrEqual(t, time.Now().UTC().Unix(), actual.TokenUSDUpdate.Unix())
		expected.TokenUSDUpdate = actual.TokenUSDUpdate
	}
	assert.Equal(t, expected, actual)
}

func assertTx(t *testing.T, expected, actual *common.PoolL2Tx) {
	// Check that timestamp has been set within the last 3 seconds
	assert.Less(t, time.Now().UTC().Unix()-3, actual.Timestamp.Unix())
	assert.GreaterOrEqual(t, time.Now().UTC().Unix(), actual.Timestamp.Unix())
	expected.Timestamp = actual.Timestamp
	// Check absolute fee
	// find token
	token := historydb.TokenRead{}
	for _, tkn := range tokensUSD {
		if expected.TokenID == tkn.TokenID {
			token = tkn
			break
		}
	}
	// If the token has value in USD setted
	if token.USDUpdate != nil {
		assert.Equal(t, token.USDUpdate.Unix(), actual.AbsoluteFeeUpdate.Unix())
		expected.AbsoluteFeeUpdate = actual.AbsoluteFeeUpdate
		// Set expected fee
		f := new(big.Float).SetInt(expected.Amount)
		amountF, _ := f.Float64()
		expected.AbsoluteFee = *token.USD * amountF * expected.Fee.Percentage()
		test.AssertUSD(t, &expected.AbsoluteFee, &actual.AbsoluteFee)
	}
	assert.Equal(t, expected, actual)
}

func BenchmarkAddTxTest(b *testing.B) {
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	now := time.Now()
	for _, tx := range txs {
		_ = l2DB.AddTxTest(tx)
	}
	elapsedTime := time.Since(now)
	log.Info("Time to insert 2048 txs:", elapsedTime)
}

func TestGetPending(t *testing.T) {
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var pendingTxs []*common.PoolL2Tx
	for _, tx := range txs {
		err := l2DB.AddTxTest(tx)
		assert.NoError(t, err)
		if tx.State == common.PoolL2TxStatePending {
			pendingTxs = append(pendingTxs, tx)
		}
	}
	fetchedTxs, err := l2DB.GetPendingTxs()
	assert.NoError(t, err)
	assert.Equal(t, len(pendingTxs), len(fetchedTxs))
	for i := range fetchedTxs {
		assertTx(t, pendingTxs[i], &fetchedTxs[i])
	}
}

/*
	WARNING: this should be fixed once transaktio is ready
func TestStartForging(t *testing.T) {
	// Generate txs
	const nInserts = 60
	const fakeBatchNum common.BatchNum = 33
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var startForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for _, tx := range txs {
		err := l2DB.AddTxTest(tx)
		assert.NoError(t, err)
		if tx.State == common.PoolL2TxStatePending && randomizer%2 == 0 {
			randomizer++
			startForgingTxIDs = append(startForgingTxIDs, tx.TxID)
		}
	}
	// Start forging txs
	err := l2DB.StartForging(startForgingTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range startForgingTxIDs {
		fetchedTx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateForging, fetchedTx.State)
		assert.Equal(t, fakeBatchNum, *fetchedTx.BatchNum)
	}
}
*/

/*
	WARNING: this should be fixed once transaktio is ready
func TestDoneForging(t *testing.T) {
	// Generate txs
	const nInserts = 60
	const fakeBatchNum common.BatchNum = 33
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var doneForgingTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for _, tx := range txs {
		err := l2DB.AddTxTest(tx)
		assert.NoError(t, err)
		if tx.State == common.PoolL2TxStateForging && randomizer%2 == 0 {
			randomizer++
			doneForgingTxIDs = append(doneForgingTxIDs, tx.TxID)
		}
	}
	// Start forging txs
	err := l2DB.DoneForging(doneForgingTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range doneForgingTxIDs {
		fetchedTx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateForged, fetchedTx.State)
		assert.Equal(t, fakeBatchNum, *fetchedTx.BatchNum)
	}
}
*/

/*
	WARNING: this should be fixed once transaktio is ready
func TestInvalidate(t *testing.T) {
	// Generate txs
	const nInserts = 60
	const fakeBatchNum common.BatchNum = 33
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var invalidTxIDs []common.TxID
	randomizer := 0
	// Add txs to DB
	for _, tx := range txs {
		err := l2DB.AddTxTest(tx)
		assert.NoError(t, err)
		if tx.State != common.PoolL2TxStateInvalid && randomizer%2 == 0 {
			randomizer++
			invalidTxIDs = append(invalidTxIDs, tx.TxID)
		}
	}
	// Start forging txs
	err := l2DB.InvalidateTxs(invalidTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateInvalid, fetchedTx.State)
		assert.Equal(t, fakeBatchNum, *fetchedTx.BatchNum)
	}
}
*/

/*
	WARNING: this should be fixed once transaktio is ready
func TestCheckNonces(t *testing.T) {
	// Generate txs
	const nInserts = 60
	const fakeBatchNum common.BatchNum = 33
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var invalidTxIDs []common.TxID
	// Generate accounts
	const nAccoutns = 2
	const currentNonce = 2
	accs := []common.Account{}
	for i := 0; i < nAccoutns; i++ {
		accs = append(accs, common.Account{
			Idx:   common.Idx(i),
			Nonce: currentNonce,
		})
	}
	// Add txs to DB
	for i := 0; i < len(txs); i++ {
		if txs[i].State != common.PoolL2TxStateInvalid {
			if i%2 == 0 { // Ensure transaction will be marked as invalid due to old nonce
				txs[i].Nonce = accs[i%len(accs)].Nonce
				txs[i].FromIdx = accs[i%len(accs)].Idx
				invalidTxIDs = append(invalidTxIDs, txs[i].TxID)
			} else { // Ensure transaction will NOT be marked as invalid due to old nonce
				txs[i].Nonce = currentNonce + 1
			}
		}
		err := l2DB.AddTxTest(txs[i])
		assert.NoError(t, err)
	}
	// Start forging txs
	err := l2DB.InvalidateTxs(invalidTxIDs, fakeBatchNum)
	assert.NoError(t, err)
	// Fetch txs and check that they've been updated correctly
	for _, id := range invalidTxIDs {
		fetchedTx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, common.PoolL2TxStateInvalid, fetchedTx.State)
		assert.Equal(t, fakeBatchNum, *fetchedTx.BatchNum)
	}
}
*/

func TestReorg(t *testing.T) {
	// Generate txs
	const nInserts = 20
	const lastValidBatch common.BatchNum = 20
	const reorgBatch common.BatchNum = lastValidBatch + 1
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	// Add txs to the DB
	reorgedTxIDs := []common.TxID{}
	nonReorgedTxIDs := []common.TxID{}
	for i := 0; i < len(txs); i++ {
		err := l2DB.AddTxTest(txs[i])
		assert.NoError(t, err)
		var batchNum common.BatchNum
		if txs[i].State == common.PoolL2TxStateForged || txs[i].State == common.PoolL2TxStateInvalid {
			reorgedTxIDs = append(reorgedTxIDs, txs[i].TxID)
			batchNum = reorgBatch
		} else {
			nonReorgedTxIDs = append(nonReorgedTxIDs, txs[i].TxID)
			batchNum = lastValidBatch
		}
		_, err = l2DB.db.Exec(
			"UPDATE tx_pool SET batch_num = $1 WHERE tx_id = $2;",
			batchNum, txs[i].TxID,
		)
		assert.NoError(t, err)
	}
	err := l2DB.Reorg(lastValidBatch)
	assert.NoError(t, err)
	for _, id := range reorgedTxIDs {
		tx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Nil(t, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		fetchedTx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, lastValidBatch, *fetchedTx.BatchNum)
	}
}

func TestPurge(t *testing.T) {
	/*
		WARNING: this should be fixed once transaktio is ready
		// Generate txs
		nInserts := l2DB.maxTxs + 20
		test.CleanL2DB(l2DB.DB())
		txs := test.GenPoolTxs(int(nInserts), tokens)
		deletedIDs := []common.TxID{}
		keepedIDs := []common.TxID{}
		const toDeleteBatchNum common.BatchNum = 30
		safeBatchNum := toDeleteBatchNum + l2DB.safetyPeriod + 1
		// Add txs to the DB
		for i := 0; i < int(l2DB.maxTxs); i++ {
			var batchNum common.BatchNum
			if i%2 == 0 { // keep tx
				batchNum = safeBatchNum
				keepedIDs = append(keepedIDs, txs[i].TxID)
			} else { // delete after safety period
				batchNum = toDeleteBatchNum
				if i%3 == 0 {
					txs[i].State = common.PoolL2TxStateForged
				} else {
					txs[i].State = common.PoolL2TxStateInvalid
				}
				deletedIDs = append(deletedIDs, txs[i].TxID)
			}
			err := l2DB.AddTxTest(txs[i])
			assert.NoError(t, err)
			// Set batchNum
			_, err = l2DB.db.Exec(
				"UPDATE tx_pool SET batch_num = $1 WHERE tx_id = $2;",
				batchNum, txs[i].TxID,
			)
			assert.NoError(t, err)
		}
		for i := int(l2DB.maxTxs); i < len(txs); i++ {
			// Delete after TTL
			deletedIDs = append(deletedIDs, txs[i].TxID)
			err := l2DB.AddTxTest(txs[i])
			assert.NoError(t, err)
			// Set timestamp
			deleteTimestamp := time.Unix(time.Now().UTC().Unix()-int64(l2DB.ttl.Seconds()+float64(4*time.Second)), 0)
			_, err = l2DB.db.Exec(
				"UPDATE tx_pool SET timestamp = $1 WHERE tx_id = $2;",
				deleteTimestamp, txs[i].TxID,
			)
			assert.NoError(t, err)
		}
		// Purge txs
		err := l2DB.Purge(safeBatchNum)
		assert.NoError(t, err)
		// Check results
		for _, id := range deletedIDs {
			tx, err := l2DB.GetTx(id)
			if err == nil {
				log.Debug(tx)
			}
			assert.Error(t, err)
		}
		for _, id := range keepedIDs {
			_, err := l2DB.GetTx(id)
			assert.NoError(t, err)
		}
	*/
}

func TestAuth(t *testing.T) {
	test.CleanL2DB(l2DB.DB())
	const nAuths = 5
	// Generate authorizations
	auths := test.GenAuths(nAuths)
	for i := 0; i < len(auths); i++ {
		// Add to the DB
		err := l2DB.AddAccountCreationAuth(auths[i])
		assert.NoError(t, err)
		// Fetch from DB
		auth, err := l2DB.GetAccountCreationAuth(auths[i].EthAddr)
		assert.NoError(t, err)
		// Check fetched vs generated
		assert.Equal(t, auths[i].EthAddr, auth.EthAddr)
		assert.Equal(t, auths[i].BJJ, auth.BJJ)
		assert.Equal(t, auths[i].Signature, auth.Signature)
		assert.Equal(t, auths[i].Timestamp.Unix(), auths[i].Timestamp.Unix())
	}
}
