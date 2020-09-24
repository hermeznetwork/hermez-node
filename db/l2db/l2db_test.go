package l2db

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/jmoiron/sqlx"

	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/stretchr/testify/assert"
)

var l2DB *L2DB
var tokens []common.Token

func TestMain(m *testing.M) {
	// init DB
	pass := "yourpasswordhere" // os.Getenv("POSTGRES_PASS")
	db, err := dbUtils.InitSQLDB(5432, "localhost", "hermez", pass, "hermez")
	if err != nil {
		panic(err)
	}
	l2DB = NewL2DB(db, 10, 100, 24*time.Hour)
	tokens, err = prepareHistoryDB(db)
	if err != nil {
		log.Error("L2DB migration failed: " + err.Error())
		panic(err)
	} else {
		log.Debug("L2DB migration succed")
	}
	// Run tests
	result := m.Run()
	// Close DB
	if err := db.Close(); err != nil {
		fmt.Println("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func prepareHistoryDB(db *sqlx.DB) ([]common.Token, error) {
	historyDB := historydb.NewHistoryDB(db)
	const fromBlock int64 = 1
	const toBlock int64 = 5
	// Clean historyDB
	if err := historyDB.Reorg(0); err != nil {
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
	return tokens, historyDB.AddTokens(tokens)
}

func TestAddTx(t *testing.T) {
	// Gen poolTxs
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	for _, tx := range txs {
		err := l2DB.AddTx(tx)
		assert.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(tx.TxID)
		assert.NoError(t, err)
		assert.Equal(t, tx.Timestamp.Unix(), fetchedTx.Timestamp.Unix())
		tx.Timestamp = fetchedTx.Timestamp
		assert.Equal(t, tx.AbsoluteFeeUpdate.Unix(), fetchedTx.AbsoluteFeeUpdate.Unix())
		tx.Timestamp = fetchedTx.Timestamp
		tx.AbsoluteFeeUpdate = fetchedTx.AbsoluteFeeUpdate
		assert.Equal(t, tx, fetchedTx)
	}
}

func BenchmarkAddTx(b *testing.B) {
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	now := time.Now()
	for _, tx := range txs {
		_ = l2DB.AddTx(tx)
	}
	elapsedTime := time.Since(now)
	fmt.Println("Time to insert 2048 txs:", elapsedTime)
}

func TestGetPending(t *testing.T) {
	const nInserts = 20
	test.CleanL2DB(l2DB.DB())
	txs := test.GenPoolTxs(nInserts, tokens)
	var pendingTxs []*common.PoolL2Tx
	for _, tx := range txs {
		err := l2DB.AddTx(tx)
		assert.NoError(t, err)
		if tx.State == common.PoolL2TxStatePending {
			pendingTxs = append(pendingTxs, tx)
		}
	}
	fetchedTxs, err := l2DB.GetPendingTxs()
	assert.NoError(t, err)
	assert.Equal(t, len(pendingTxs), len(fetchedTxs))
	for i, fetchedTx := range fetchedTxs {
		assert.Equal(t, pendingTxs[i].Timestamp.Unix(), fetchedTx.Timestamp.Unix())
		pendingTxs[i].Timestamp = fetchedTx.Timestamp
		assert.Equal(t, pendingTxs[i].AbsoluteFeeUpdate.Unix(), fetchedTx.AbsoluteFeeUpdate.Unix())
		pendingTxs[i].AbsoluteFeeUpdate = fetchedTx.AbsoluteFeeUpdate
		assert.Equal(t, pendingTxs[i], fetchedTx)
	}
}

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
		err := l2DB.AddTx(tx)
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
		assert.Equal(t, fakeBatchNum, fetchedTx.BatchNum)
	}
}

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
		err := l2DB.AddTx(tx)
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
		assert.Equal(t, fakeBatchNum, fetchedTx.BatchNum)
	}
}

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
		err := l2DB.AddTx(tx)
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
		assert.Equal(t, fakeBatchNum, fetchedTx.BatchNum)
	}
}

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
		err := l2DB.AddTx(txs[i])
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
		assert.Equal(t, fakeBatchNum, fetchedTx.BatchNum)
	}
}

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
		if txs[i].State == common.PoolL2TxStateForged || txs[i].State == common.PoolL2TxStateInvalid {
			txs[i].BatchNum = reorgBatch
			reorgedTxIDs = append(reorgedTxIDs, txs[i].TxID)
		} else {
			txs[i].BatchNum = lastValidBatch
			nonReorgedTxIDs = append(nonReorgedTxIDs, txs[i].TxID)
		}
		err := l2DB.AddTx(txs[i])
		assert.NoError(t, err)
	}
	err := l2DB.Reorg(lastValidBatch)
	assert.NoError(t, err)
	var nullBatchNum common.BatchNum
	for _, id := range reorgedTxIDs {
		tx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, nullBatchNum, tx.BatchNum)
		assert.Equal(t, common.PoolL2TxStatePending, tx.State)
	}
	for _, id := range nonReorgedTxIDs {
		tx, err := l2DB.GetTx(id)
		assert.NoError(t, err)
		assert.Equal(t, lastValidBatch, tx.BatchNum)
	}
}

func TestPurge(t *testing.T) {
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
		if i%1 == 0 { // keep tx
			txs[i].BatchNum = safeBatchNum
			keepedIDs = append(keepedIDs, txs[i].TxID)
		} else if i%2 == 0 { // delete after safety period
			txs[i].BatchNum = toDeleteBatchNum
			if i%3 == 0 {
				txs[i].State = common.PoolL2TxStateForged
			} else {
				txs[i].State = common.PoolL2TxStateInvalid
			}
			deletedIDs = append(deletedIDs, txs[i].TxID)
		}
		err := l2DB.AddTx(txs[i])
		assert.NoError(t, err)
	}
	for i := int(l2DB.maxTxs); i < len(txs); i++ {
		// Delete after TTL
		txs[i].Timestamp = time.Unix(time.Now().UTC().Unix()-int64(l2DB.ttl.Seconds()+float64(4*time.Second)), 0)
		deletedIDs = append(deletedIDs, txs[i].TxID)
		err := l2DB.AddTx(txs[i])
		assert.NoError(t, err)
	}
	// Purge txs
	err := l2DB.Purge(safeBatchNum - 1)
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
