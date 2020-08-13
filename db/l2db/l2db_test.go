package l2db

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

var l2DB *L2DB

// In order to run the test you need to run a Posgres DB with
// a database named "l2" that is accessible by
// user: "hermez"
// pass: set it using the env var POSTGRES_PASS
// This can be achieved by running: POSTGRES_PASS=your_strong_pass && sudo docker run --rm --name hermez-db-test -p 5432:5432 -e POSTGRES_DB=history -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD=$POSTGRES_PASS -d postgres && sleep 2s && sudo docker exec -it hermez-db-test psql -a history -U hermez -c "CREATE DATABASE l2;"
// After running the test you can stop the container by running: sudo docker kill hermez-ydb-test
// If you already did that for the HistoryDB you don't have to do it again
func TestMain(m *testing.M) {
	// init DB
	var err error
	pass := "your_strong_pass" // os.Getenv("POSTGRES_PASS")
	l2DB, err = NewL2DB(5432, "localhost", "hermez", pass, "l2", 10, 512, 24*time.Hour)
	if err != nil {
		panic(err)
	}
	// Run tests
	result := m.Run()
	// Close DB
	if err := l2DB.Close(); err != nil {
		fmt.Println("Error closing the history DB:", err)
	}
	os.Exit(result)
}

func TestAddTx(t *testing.T) {
	const nInserts = 10
	cleanDB()
	txs := genTxs(nInserts)
	for _, tx := range txs {
		err := l2DB.AddTx(tx)
		assert.NoError(t, err)
		fetchedTx, err := l2DB.GetTx(tx.TxID)
		assert.NoError(t, err)
		assert.Equal(t, tx.Timestamp.Unix(), fetchedTx.Timestamp.Unix())
		tx.Timestamp = fetchedTx.Timestamp
		assert.Equal(t, tx, fetchedTx)
	}
}

func BenchmarkAddTx(b *testing.B) {
	const nInserts = 20
	cleanDB()
	txs := genTxs(nInserts)
	now := time.Now()
	for _, tx := range txs {
		l2DB.AddTx(tx)
	}
	elapsedTime := time.Since(now)
	fmt.Println("Time to insert 2048 txs:", elapsedTime)
}

func TestGetPending(t *testing.T) {
	const nInserts = 20
	cleanDB()
	txs := genTxs(nInserts)
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
		assert.Equal(t, pendingTxs[i], fetchedTx)
	}
}

func TestStartForging(t *testing.T) {
	const nInserts = 80
	const fakeBlockNum = 33
	cleanDB()
	txs := genTxs(nInserts)
	var startForgingTxs []*common.PoolL2Tx
	var startForgingTxIDs []common.TxID
	for i, tx := range txs {
		err := l2DB.AddTx(tx)
		assert.NoError(t, err)
		if tx.State == common.PoolL2TxStatePending && i%2 == 0 {
			startForgingTxs = append(startForgingTxs, tx)
			startForgingTxIDs = append(startForgingTxIDs, tx.TxID)
		}
	}
	err := l2DB.StartForging(startForgingTxIDs, fakeBlockNum)
	assert.NoError(t, err)
}

func genTxs(n int) []*common.PoolL2Tx {
	// WARNING: this is just to test geting/seting from/to the DB.
	// This tx doesn't follow the protocol (signature, txID, ...)!!
	// TODO:
	// - ENABLE SIGNATURE !!!!
	// - make it compliant with the protocol once common.Tx is more advanced
	txs := make([]*common.PoolL2Tx, 0, n)
	privK := babyjub.NewRandPrivKey()
	for i := 0; i < n; i++ {
		var state common.PoolL2TxState
		includeRq := i%2 == 0
		if i%4 == 0 {
			state = common.PoolL2TxStatePending
		} else if i%4 == 1 {
			state = common.PoolL2TxStateInvalid
		} else if i%4 == 2 {
			state = common.PoolL2TxStateForging
		} else if i%4 == 3 {
			state = common.PoolL2TxStateForged
		}
		tx := &common.PoolL2Tx{
			TxID:    common.TxID(common.Hash([]byte(strconv.Itoa(i)))),
			FromIdx: 47,
			ToIdx:   96,
			TokenID: 73,
			Amount:  big.NewInt(3487762374627846747),
			Nonce:   28,
			Fee:     99,
			//	Type:          common.TxTypeTransfer,
			ToBJJ:     *privK.Public(),
			State:     state,
			Timestamp: time.Now().UTC(),
			// Signature:     *privK.SignPoseidon(big.NewInt(674238462)),
			ToEthAddr: eth.BigToAddress(big.NewInt(234523534)),
		}
		if includeRq {
			tx.RqAmount = big.NewInt(3487762374627846747)
			tx.RqFee = 11
			tx.RqToEthAddr = eth.BigToAddress(big.NewInt(239457111187))
			tx.RqTokenID = 222
			tx.RqNonce = 78
		}
		txs = append(txs, tx)
	}
	return txs
}

func cleanDB() {
	if _, err := l2DB.db.Exec("DELETE FROM tx_pool"); err != nil {
		panic(err)
	}
	if _, err := l2DB.db.Exec("DELETE FROM account_creation_auth"); err != nil {
		panic(err)
	}
}
