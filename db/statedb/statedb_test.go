package statedb

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAccount(t *testing.T, i int) *common.Account {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	require.NoError(t, err)
	pk := sk.Public()

	key, err := ethCrypto.GenerateKey()
	require.NoError(t, err)
	address := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &common.Account{
		Idx:       common.Idx(256 + i),
		TokenID:   common.TokenID(i),
		Nonce:     common.Nonce(i),
		Balance:   big.NewInt(1000),
		PublicKey: pk.Compress(),
		EthAddr:   address,
	}
}

func TestNewStateDBIntermediateState(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeTxSelector, 0)
	assert.NoError(t, err)

	// test values
	k0 := []byte("testkey0")
	k1 := []byte("testkey1")
	v0 := []byte("testvalue0")
	v1 := []byte("testvalue1")

	// store some data
	tx, err := sdb.db.DB().NewTx()
	assert.NoError(t, err)
	err = tx.Put(k0, v0)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)
	v, err := sdb.db.DB().Get(k0)
	assert.NoError(t, err)
	assert.Equal(t, v0, v)

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey0'&'testvalue0' data)
	sdb, err = NewStateDB(dir, 128, TypeTxSelector, 0)
	assert.NoError(t, err)
	v, err = sdb.db.DB().Get(k0)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
	assert.Nil(t, v)

	// store the same data from the beginning that has ben lost since last NewStateDB
	tx, err = sdb.db.DB().NewTx()
	assert.NoError(t, err)
	err = tx.Put(k0, v0)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)
	v, err = sdb.db.DB().Get(k0)
	assert.NoError(t, err)
	assert.Equal(t, v0, v)

	// make checkpoints with the current state
	bn, err := sdb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(0), bn)
	err = sdb.db.MakeCheckpoint()
	assert.NoError(t, err)
	bn, err = sdb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), bn)

	// write more data
	tx, err = sdb.db.DB().NewTx()
	assert.NoError(t, err)
	err = tx.Put(k1, v1)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)

	v, err = sdb.db.DB().Get(k1)
	assert.NoError(t, err)
	assert.Equal(t, v1, v)

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey1'&'testvalue1' data)
	sdb, err = NewStateDB(dir, 128, TypeTxSelector, 0)
	assert.NoError(t, err)

	v, err = sdb.db.DB().Get(k0)
	assert.NoError(t, err)
	assert.Equal(t, v0, v)

	v, err = sdb.db.DB().Get(k1)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
	assert.Nil(t, v)
}

func TestStateDBWithoutMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeTxSelector, 0)
	assert.NoError(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 4; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// get non-existing account, expecting an error
	unexistingAccount := common.Idx(1)
	_, err = sdb.GetAccount(unexistingAccount)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		assert.NoError(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		existingAccount := accounts[i].Idx
		accGetted, err := sdb.GetAccount(existingAccount)
		assert.NoError(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	existingAccount := common.Idx(256)
	_, err = sdb.GetAccount(existingAccount) // check that exist
	assert.NoError(t, err)
	_, err = sdb.CreateAccount(common.Idx(256), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, tracerr.Unwrap(err))

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		existingAccount = common.Idx(i)
		_, err = sdb.UpdateAccount(existingAccount, accounts[i])
		assert.NoError(t, err)
	}

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, tracerr.Unwrap(err))
}

func TestStateDBWithMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeSynchronizer, 32)
	assert.NoError(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 20; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// get non-existing account, expecting an error
	_, err = sdb.GetAccount(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		assert.NoError(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		accGetted, err := sdb.GetAccount(accounts[i].Idx)
		assert.NoError(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	_, err = sdb.GetAccount(common.Idx(256)) // check that exist
	assert.NoError(t, err)
	_, err = sdb.CreateAccount(common.Idx(256), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, tracerr.Unwrap(err))

	_, err = sdb.MTGetProof(common.Idx(256))
	assert.NoError(t, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		_, err = sdb.UpdateAccount(accounts[i].Idx, accounts[i])
		assert.NoError(t, err)
	}
	a, err := sdb.GetAccount(common.Idx(256)) // check that account value has been updated
	assert.NoError(t, err)
	assert.Equal(t, accounts[0].Nonce, a.Nonce)
}

// TestCheckpoints performs almost the same test than kvdb/kvdb_test.go
// TestCheckpoints, but over the StateDB
func TestCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "sdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeSynchronizer, 32)
	assert.NoError(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 10; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		assert.NoError(t, err)
	}

	// do checkpoints and check that currentBatch is correct
	err = sdb.db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err := sdb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), cb)

	for i := 1; i < 10; i++ {
		err = sdb.db.MakeCheckpoint()
		assert.NoError(t, err)

		cb, err = sdb.db.GetCurrentBatch()
		assert.NoError(t, err)
		assert.Equal(t, common.BatchNum(i+1), cb)
	}

	// printCheckpoints(t, sdb.path)

	// reset checkpoint
	err = sdb.Reset(3)
	assert.NoError(t, err)

	// check that reset can be repeated (as there exist the 'current' and
	// 'BatchNum3', from where the 'current' is a copy)
	err = sdb.Reset(3)
	require.NoError(t, err)

	// check that currentBatch is as expected after Reset
	cb, err = sdb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(3), cb)

	// advance one checkpoint and check that currentBatch is fine
	err = sdb.db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = sdb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)

	err = sdb.db.DeleteCheckpoint(common.BatchNum(1))
	assert.NoError(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(2))
	assert.NoError(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(1)) // does not exist, should return err
	assert.NotNil(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(2)) // does not exist, should return err
	assert.NotNil(t, err)

	// Create a LocalStateDB from the initial StateDB
	dirLocal, err := ioutil.TempDir("", "ldb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dirLocal))
	ldb, err := NewLocalStateDB(dirLocal, 128, sdb, TypeBatchBuilder, 32)
	assert.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb.Reset(4, true)
	assert.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb
	err = ldb.db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = ldb.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	// Create a 2nd LocalStateDB from the initial StateDB
	dirLocal2, err := ioutil.TempDir("", "ldb2")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dirLocal2))
	ldb2, err := NewLocalStateDB(dirLocal2, 128, sdb, TypeBatchBuilder, 32)
	assert.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb2.Reset(4, true)
	assert.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb2.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb2
	err = ldb2.db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = ldb2.db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	debug := false
	if debug {
		printCheckpoints(t, sdb.path)
		printCheckpoints(t, ldb.path)
		printCheckpoints(t, ldb2.path)
	}
}

func TestStateDBGetAccounts(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)

	sdb, err := NewStateDB(dir, 128, TypeTxSelector, 0)
	assert.NoError(t, err)

	// create test accounts
	var accounts []common.Account
	for i := 0; i < 16; i++ {
		account := newAccount(t, i)
		accounts = append(accounts, *account)
	}

	// add test accounts
	for i := range accounts {
		_, err = sdb.CreateAccount(accounts[i].Idx, &accounts[i])
		require.NoError(t, err)
	}

	dbAccounts, err := sdb.GetAccounts()
	require.NoError(t, err)
	assert.Equal(t, accounts, dbAccounts)
}

func printCheckpoints(t *testing.T, path string) {
	files, err := ioutil.ReadDir(path)
	assert.NoError(t, err)

	fmt.Println(path)
	for _, f := range files {
		fmt.Println("	" + f.Name())
	}
}

func bigFromStr(h string, u int) *big.Int {
	if u == 16 {
		h = strings.TrimPrefix(h, "0x")
	}
	b, ok := new(big.Int).SetString(h, u)
	if !ok {
		panic("bigFromStr err")
	}
	return b
}

func TestCheckAccountsTreeTestVectors(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeSynchronizer, 32)
	require.NoError(t, err)

	ay0 := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(253), nil), big.NewInt(1))
	// test value from js version (compatibility-canary)
	assert.Equal(t, "1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", (hex.EncodeToString(ay0.Bytes())))
	bjjPoint0Comp := babyjub.PackSignY(true, ay0)
	bjj0 := babyjub.PublicKeyComp(bjjPoint0Comp)

	ay1 := bigFromStr("00", 16)
	bjjPoint1Comp := babyjub.PackSignY(false, ay1)
	bjj1 := babyjub.PublicKeyComp(bjjPoint1Comp)
	ay2 := bigFromStr("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7", 16)
	bjjPoint2Comp := babyjub.PackSignY(false, ay2)
	bjj2 := babyjub.PublicKeyComp(bjjPoint2Comp)

	ay3 := bigFromStr("0x10", 16) // 0x10=16
	bjjPoint3Comp := babyjub.PackSignY(false, ay3)
	require.NoError(t, err)
	bjj3 := babyjub.PublicKeyComp(bjjPoint3Comp)
	accounts := []*common.Account{
		{
			Idx:       1,
			TokenID:   0xFFFFFFFF,
			PublicKey: bjj0,
			EthAddr:   ethCommon.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
			Nonce:     common.Nonce(0xFFFFFFFFFF),
			Balance:   bigFromStr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16),
		},
		{
			Idx:       100,
			TokenID:   0,
			PublicKey: bjj1,
			EthAddr:   ethCommon.HexToAddress("0x00"),
			Nonce:     common.Nonce(0),
			Balance:   bigFromStr("0", 10),
		},
		{
			Idx:       0xFFFFFFFFFFFF,
			TokenID:   3,
			PublicKey: bjj2,
			EthAddr:   ethCommon.HexToAddress("0xA3C88ac39A76789437AED31B9608da72e1bbfBF9"),
			Nonce:     common.Nonce(129),
			Balance:   bigFromStr("42000000000000000000", 10),
		},
		{
			Idx:       10000,
			TokenID:   1000,
			PublicKey: bjj3,
			EthAddr:   ethCommon.HexToAddress("0x64"),
			Nonce:     common.Nonce(1900),
			Balance:   bigFromStr("14000000000000000000", 10),
		},
	}
	for i := 0; i < len(accounts); i++ {
		_, err = accounts[i].HashValue()
		require.NoError(t, err)
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		if err != nil {
			log.Error(err)
		}
		require.NoError(t, err)
	}
	// root value generated by js version:
	assert.Equal(t, "17298264051379321456969039521810887093935433569451713402227686942080129181291", sdb.MT.Root().BigInt().String())
}

// TestListCheckpoints performs almost the same test than kvdb/kvdb_test.go
// TestListCheckpoints, but over the StateDB
func TestListCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	sdb, err := NewStateDB(dir, 128, TypeSynchronizer, 32)
	require.NoError(t, err)

	numCheckpoints := 16
	// do checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = sdb.db.MakeCheckpoint()
		require.NoError(t, err)
	}
	list, err := sdb.db.ListCheckpoints()
	require.NoError(t, err)
	assert.Equal(t, numCheckpoints, len(list))
	assert.Equal(t, 1, list[0])
	assert.Equal(t, numCheckpoints, list[len(list)-1])

	numReset := 10
	err = sdb.Reset(common.BatchNum(numReset))
	require.NoError(t, err)
	list, err = sdb.db.ListCheckpoints()
	require.NoError(t, err)
	assert.Equal(t, numReset, len(list))
	assert.Equal(t, 1, list[0])
	assert.Equal(t, numReset, list[len(list)-1])
}

// TestDeleteOldCheckpoints performs almost the same test than
// kvdb/kvdb_test.go TestDeleteOldCheckpoints, but over the StateDB
func TestDeleteOldCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	keep := 16
	sdb, err := NewStateDB(dir, keep, TypeSynchronizer, 32)
	require.NoError(t, err)

	numCheckpoints := 32
	// do checkpoints and check that we never have more than `keep`
	// checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = sdb.db.MakeCheckpoint()
		require.NoError(t, err)
		checkpoints, err := sdb.db.ListCheckpoints()
		require.NoError(t, err)
		assert.LessOrEqual(t, len(checkpoints), keep)
	}
}
