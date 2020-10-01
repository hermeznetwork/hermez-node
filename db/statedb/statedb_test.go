package statedb

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAccount(t *testing.T, i int) *common.Account {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	require.Nil(t, err)
	pk := sk.Public()

	key, err := ethCrypto.GenerateKey()
	require.Nil(t, err)
	address := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &common.Account{
		TokenID:   common.TokenID(i),
		Nonce:     common.Nonce(i),
		Balance:   big.NewInt(1000),
		PublicKey: pk,
		EthAddr:   address,
	}
}

func TestNewStateDBIntermediateState(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeTxSelector, 0)
	assert.Nil(t, err)

	// test values
	k0 := []byte("testkey0")
	k1 := []byte("testkey1")
	v0 := []byte("testvalue0")
	v1 := []byte("testvalue1")

	// store some data
	tx, err := sdb.db.NewTx()
	assert.Nil(t, err)
	err = tx.Put(k0, v0)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)
	v, err := sdb.db.Get(k0)
	assert.Nil(t, err)
	assert.Equal(t, v0, v)

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey0'&'testvalue0' data)
	sdb, err = NewStateDB(dir, TypeTxSelector, 0)
	assert.Nil(t, err)
	v, err = sdb.db.Get(k0)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, err)
	assert.Nil(t, v)

	// store the same data from the beginning that has ben lost since last NewStateDB
	tx, err = sdb.db.NewTx()
	assert.Nil(t, err)
	err = tx.Put(k0, v0)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)
	v, err = sdb.db.Get(k0)
	assert.Nil(t, err)
	assert.Equal(t, v0, v)

	// make checkpoints with the current state
	bn, err := sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(0), bn)
	err = sdb.MakeCheckpoint()
	assert.Nil(t, err)
	bn, err = sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(1), bn)

	// write more data
	tx, err = sdb.db.NewTx()
	assert.Nil(t, err)
	err = tx.Put(k1, v1)
	assert.Nil(t, err)
	err = tx.Commit()
	assert.Nil(t, err)

	v, err = sdb.db.Get(k1)
	assert.Nil(t, err)
	assert.Equal(t, v1, v)

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey1'&'testvalue1' data)
	sdb, err = NewStateDB(dir, TypeTxSelector, 0)
	assert.Nil(t, err)

	v, err = sdb.db.Get(k0)
	assert.Nil(t, err)
	assert.Equal(t, v0, v)

	v, err = sdb.db.Get(k1)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, err)
	assert.Nil(t, v)
}

func TestStateDBWithoutMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeTxSelector, 0)
	assert.Nil(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 100; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// get non-existing account, expecting an error
	_, err = sdb.GetAccount(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, err)

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		accGetted, err := sdb.GetAccount(common.Idx(i))
		assert.Nil(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	_, err = sdb.GetAccount(common.Idx(1)) // check that exist
	assert.Nil(t, err)
	_, err = sdb.CreateAccount(common.Idx(1), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		_, err = sdb.UpdateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, err)
}

func TestStateDBWithMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 20; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// get non-existing account, expecting an error
	_, err = sdb.GetAccount(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, err)

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		accGetted, err := sdb.GetAccount(common.Idx(i))
		assert.Nil(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	_, err = sdb.GetAccount(common.Idx(1)) // check that exist
	assert.Nil(t, err)
	_, err = sdb.CreateAccount(common.Idx(1), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, err)

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.Nil(t, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		_, err = sdb.UpdateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}
	a, err := sdb.GetAccount(common.Idx(1)) // check that account value has been updated
	assert.Nil(t, err)
	assert.Equal(t, accounts[1].Nonce, a.Nonce)
}

func TestCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "sdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, TypeSynchronizer, 32)
	assert.Nil(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 10; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}

	// do checkpoints and check that currentBatch is correct
	err = sdb.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err := sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(1), cb)

	for i := 1; i < 10; i++ {
		err = sdb.MakeCheckpoint()
		assert.Nil(t, err)

		cb, err = sdb.GetCurrentBatch()
		assert.Nil(t, err)
		assert.Equal(t, common.BatchNum(i+1), cb)
	}

	// printCheckpoints(t, sdb.path)

	// reset checkpoint
	err = sdb.Reset(3)
	assert.Nil(t, err)

	// check that reset can be repeated (as there exist the 'current' and
	// 'BatchNum3', from where the 'current' is a copy)
	err = sdb.Reset(3)
	require.Nil(t, err)

	// check that currentBatch is as expected after Reset
	cb, err = sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(3), cb)

	// advance one checkpoint and check that currentBatch is fine
	err = sdb.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(4), cb)

	err = sdb.DeleteCheckpoint(common.BatchNum(9))
	assert.Nil(t, err)
	err = sdb.DeleteCheckpoint(common.BatchNum(10))
	assert.Nil(t, err)
	err = sdb.DeleteCheckpoint(common.BatchNum(9)) // does not exist, should return err
	assert.NotNil(t, err)
	err = sdb.DeleteCheckpoint(common.BatchNum(11)) // does not exist, should return err
	assert.NotNil(t, err)

	// Create a LocalStateDB from the initial StateDB
	dirLocal, err := ioutil.TempDir("", "ldb")
	require.Nil(t, err)
	ldb, err := NewLocalStateDB(dirLocal, sdb, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb.Reset(4, true)
	assert.Nil(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb
	err = ldb.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = ldb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	// Create a 2nd LocalStateDB from the initial StateDB
	dirLocal2, err := ioutil.TempDir("", "ldb2")
	require.Nil(t, err)
	ldb2, err := NewLocalStateDB(dirLocal2, sdb, TypeBatchBuilder, 32)
	assert.Nil(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb2.Reset(4, true)
	assert.Nil(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb2.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb2
	err = ldb2.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = ldb2.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	debug := false
	if debug {
		printCheckpoints(t, sdb.path)
		printCheckpoints(t, ldb.path)
		printCheckpoints(t, ldb2.path)
	}
}

func printCheckpoints(t *testing.T, path string) {
	files, err := ioutil.ReadDir(path)
	assert.Nil(t, err)

	fmt.Println(path)
	for _, f := range files {
		fmt.Println("	" + f.Name())
	}
}
