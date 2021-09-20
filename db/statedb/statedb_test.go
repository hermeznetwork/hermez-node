package statedb

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"sync"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme []string

func init() {
	log.Init("debug", []string{"stdout"})
}
func TestMain(m *testing.M) {
	exitVal := 0
	exitVal = m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

func newAccount(t *testing.T, i int) *common.Account {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	require.NoError(t, err)
	pk := sk.Public()

	key, err := ethCrypto.GenerateKey()
	require.NoError(t, err)
	address := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &common.Account{
		Idx:     common.Idx(256 + i),
		TokenID: common.TokenID(i),
		Nonce:   nonce.Nonce(i),
		Balance: big.NewInt(1000),
		BJJ:     pk.Compress(),
		EthAddr: address,
	}
}

func TestNewStateDBIntermediateState(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	require.NoError(t, err)

	// test values
	k0 := []byte("testkey0")
	k1 := []byte("testkey1")
	v0 := []byte("testvalue0")
	v1 := []byte("testvalue1")

	// store some data
	tx, err := sdb.db.DB().NewTx()
	require.NoError(t, err)
	err = tx.Put(k0, v0)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)
	v, err := sdb.db.DB().Get(k0)
	require.NoError(t, err)
	assert.Equal(t, v0, v)

	// k0 not yet in last
	err = sdb.LastRead(func(sdb *Last) error {
		_, err := sdb.DB().Get(k0)
		assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
		return nil
	})
	require.NoError(t, err)

	// Close PebbleDB before creating a new StateDB
	sdb.Close()

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey0'&'testvalue0' data)
	sdb, err = NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	require.NoError(t, err)
	v, err = sdb.db.DB().Get(k0)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
	assert.Nil(t, v)

	// k0 not in last
	err = sdb.LastRead(func(sdb *Last) error {
		_, err := sdb.DB().Get(k0)
		assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
		return nil
	})
	require.NoError(t, err)

	// store the same data from the beginning that has ben lost since last NewStateDB
	tx, err = sdb.db.DB().NewTx()
	require.NoError(t, err)
	err = tx.Put(k0, v0)
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)
	v, err = sdb.db.DB().Get(k0)
	require.NoError(t, err)
	assert.Equal(t, v0, v)

	// k0 yet not in last
	err = sdb.LastRead(func(sdb *Last) error {
		_, err := sdb.DB().Get(k0)
		assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
		return nil
	})
	require.NoError(t, err)

	// make checkpoints with the current state
	bn, err := sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(0), bn)
	err = sdb.MakeCheckpoint()
	require.NoError(t, err)
	bn, err = sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), bn)

	// k0 in last
	err = sdb.LastRead(func(sdb *Last) error {
		v, err := sdb.DB().Get(k0)
		require.NoError(t, err)
		assert.Equal(t, v0, v)
		return nil
	})
	require.NoError(t, err)

	// write more data
	tx, err = sdb.db.DB().NewTx()
	require.NoError(t, err)
	err = tx.Put(k1, v1)
	require.NoError(t, err)
	err = tx.Put(k0, v1) // overwrite k0 with v1
	require.NoError(t, err)
	err = tx.Commit()
	require.NoError(t, err)

	v, err = sdb.db.DB().Get(k1)
	require.NoError(t, err)
	assert.Equal(t, v1, v)

	err = sdb.LastRead(func(sdb *Last) error {
		v, err := sdb.DB().Get(k0)
		require.NoError(t, err)
		assert.Equal(t, v0, v)
		return nil
	})
	require.NoError(t, err)

	// Close PebbleDB before creating a new StateDB
	sdb.Close()

	// call NewStateDB which should get the db at the last checkpoint state
	// executing a Reset (discarding the last 'testkey1'&'testvalue1' data)
	sdb, err = NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	require.NoError(t, err)

	bn, err = sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), bn)

	// we closed the db without doing a checkpoint after overwriting k0, so
	// it's back to v0
	v, err = sdb.db.DB().Get(k0)
	require.NoError(t, err)
	assert.Equal(t, v0, v)

	v, err = sdb.db.DB().Get(k1)
	assert.NotNil(t, err)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))
	assert.Nil(t, v)

	sdb.Close()
}

func TestStateDBWithoutMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	require.NoError(t, err)

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
		require.NoError(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		existingAccount := accounts[i].Idx
		accGetted, err := sdb.GetAccount(existingAccount)
		require.NoError(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	existingAccount := common.Idx(256)
	_, err = sdb.GetAccount(existingAccount) // check that exist
	require.NoError(t, err)
	_, err = sdb.CreateAccount(common.Idx(256), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, tracerr.Unwrap(err))

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		existingAccount = common.Idx(i)
		_, err = sdb.UpdateAccount(existingAccount, accounts[i])
		require.NoError(t, err)
	}

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, tracerr.Unwrap(err))

	sdb.Close()
}

func TestStateDBWithMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

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
		require.NoError(t, err)
	}

	for i := 0; i < len(accounts); i++ {
		accGetted, err := sdb.GetAccount(accounts[i].Idx)
		require.NoError(t, err)
		assert.Equal(t, accounts[i], accGetted)
	}

	// try already existing idx and get error
	_, err = sdb.GetAccount(common.Idx(256)) // check that exist
	require.NoError(t, err)
	_, err = sdb.CreateAccount(common.Idx(256), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, tracerr.Unwrap(err))

	_, err = sdb.MTGetProof(common.Idx(256))
	require.NoError(t, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		_, err = sdb.UpdateAccount(accounts[i].Idx, accounts[i])
		require.NoError(t, err)
	}
	a, err := sdb.GetAccount(common.Idx(256)) // check that account value has been updated
	require.NoError(t, err)
	assert.Equal(t, accounts[0].Nonce, a.Nonce)

	sdb.Close()
}

// TestCheckpoints performs almost the same test than kvdb/kvdb_test.go
// TestCheckpoints, but over the StateDB
func TestCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "sdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	err = sdb.Reset(0)
	require.NoError(t, err)

	// create test accounts
	var accounts []*common.Account
	for i := 0; i < 10; i++ {
		accounts = append(accounts, newAccount(t, i))
	}

	// add test accounts
	for i := 0; i < len(accounts); i++ {
		_, err = sdb.CreateAccount(accounts[i].Idx, accounts[i])
		require.NoError(t, err)
	}
	// account doesn't exist in Last checkpoint
	_, err = sdb.LastGetAccount(accounts[0].Idx)
	assert.Equal(t, db.ErrNotFound, tracerr.Unwrap(err))

	// do checkpoints and check that currentBatch is correct
	err = sdb.MakeCheckpoint()
	require.NoError(t, err)
	cb, err := sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), cb)

	// account exists in Last checkpoint
	accCur, err := sdb.GetAccount(accounts[0].Idx)
	require.NoError(t, err)
	accLast, err := sdb.LastGetAccount(accounts[0].Idx)
	require.NoError(t, err)
	assert.Equal(t, accounts[0], accLast)
	assert.Equal(t, accCur, accLast)

	for i := 1; i < 10; i++ {
		err = sdb.MakeCheckpoint()
		require.NoError(t, err)

		cb, err = sdb.getCurrentBatch()
		require.NoError(t, err)
		assert.Equal(t, common.BatchNum(i+1), cb)
	}

	// printCheckpoints(t, sdb.cfg.Path)

	// reset checkpoint
	err = sdb.Reset(3)
	require.NoError(t, err)

	// check that reset can be repeated (as there exist the 'current' and
	// 'BatchNum3', from where the 'current' is a copy)
	err = sdb.Reset(3)
	require.NoError(t, err)

	// check that currentBatch is as expected after Reset
	cb, err = sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(3), cb)

	// advance one checkpoint and check that currentBatch is fine
	err = sdb.MakeCheckpoint()
	require.NoError(t, err)
	cb, err = sdb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)

	err = sdb.db.DeleteCheckpoint(common.BatchNum(1))
	require.NoError(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(2))
	require.NoError(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(1)) // does not exist, should return err
	assert.NotNil(t, err)
	err = sdb.db.DeleteCheckpoint(common.BatchNum(2)) // does not exist, should return err
	assert.NotNil(t, err)

	// Create a LocalStateDB from the initial StateDB
	dirLocal, err := ioutil.TempDir("", "ldb")
	require.NoError(t, err)
	deleteme = append(deleteme, dirLocal)
	ldb, err := NewLocalStateDB(Config{Path: dirLocal, Keep: 128, Type: TypeBatchBuilder,
		NLevels: 32}, sdb)
	require.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb.Reset(4, true)
	require.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb
	err = ldb.MakeCheckpoint()
	require.NoError(t, err)
	cb, err = ldb.getCurrentBatch()
	require.NoError(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	// Create a 2nd LocalStateDB from the initial StateDB
	dirLocal2, err := ioutil.TempDir("", "ldb2")
	require.NoError(t, err)
	deleteme = append(deleteme, dirLocal2)
	ldb2, err := NewLocalStateDB(Config{Path: dirLocal2, Keep: 128, Type: TypeBatchBuilder,
		NLevels: 32}, sdb)
	require.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb2.Reset(4, true)
	require.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb = ldb2.CurrentBatch()
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb2
	err = ldb2.MakeCheckpoint()
	require.NoError(t, err)
	cb = ldb2.CurrentBatch()
	assert.Equal(t, common.BatchNum(5), cb)

	debug := false
	if debug {
		printCheckpoints(t, sdb.cfg.Path)
		printCheckpoints(t, ldb.cfg.Path)
		printCheckpoints(t, ldb2.cfg.Path)
	}

	ldb2.Close()
	ldb.Close()
	sdb.Close()
}

func TestStateDBGetAccounts(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeTxSelector, NLevels: 0})
	require.NoError(t, err)

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

	dbAccounts, err := sdb.TestGetAccounts()
	require.NoError(t, err)
	assert.Equal(t, accounts, dbAccounts)

	sdb.Close()
}

func printCheckpoints(t *testing.T, path string) {
	files, err := ioutil.ReadDir(path)
	require.NoError(t, err)

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
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	ay0 := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(253), nil), big.NewInt(1))
	// test value from js version (compatibility-canary)
	assert.Equal(t, "1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		(hex.EncodeToString(ay0.Bytes())))
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
			Idx:     1,
			TokenID: 0xFFFFFFFF,
			BJJ:     bjj0,
			EthAddr: ethCommon.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
			Nonce:   nonce.Nonce(0xFFFFFFFFFF),
			Balance: bigFromStr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16),
		},
		{
			Idx:     100,
			TokenID: 0,
			BJJ:     bjj1,
			EthAddr: ethCommon.HexToAddress("0x00"),
			Nonce:   nonce.Nonce(0),
			Balance: bigFromStr("0", 10),
		},
		{
			Idx:     0xFFFFFFFFFFFF,
			TokenID: 3,
			BJJ:     bjj2,
			EthAddr: ethCommon.HexToAddress("0xA3C88ac39A76789437AED31B9608da72e1bbfBF9"),
			Nonce:   nonce.Nonce(129),
			Balance: bigFromStr("42000000000000000000", 10),
		},
		{
			Idx:     10000,
			TokenID: 1000,
			BJJ:     bjj3,
			EthAddr: ethCommon.HexToAddress("0x64"),
			Nonce:   nonce.Nonce(1900),
			Balance: bigFromStr("14000000000000000000", 10),
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
	assert.Equal(t,
		"13174362770971232417413036794215823584762073355951212910715422236001731746065",
		sdb.MT.Root().BigInt().String())

	sdb.Close()
}

// TestListCheckpoints performs almost the same test than kvdb/kvdb_test.go
// TestListCheckpoints, but over the StateDB
func TestListCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	sdb, err := NewStateDB(Config{Path: dir, Keep: 128, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	numCheckpoints := 16
	// do checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = sdb.MakeCheckpoint()
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

	sdb.Close()
}

// TestDeleteOldCheckpoints performs almost the same test than
// kvdb/kvdb_test.go TestDeleteOldCheckpoints, but over the StateDB
func TestDeleteOldCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	keep := 16
	sdb, err := NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	numCheckpoints := 32
	// do checkpoints and check that we never have more than `keep`
	// checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = sdb.MakeCheckpoint()
		require.NoError(t, err)
		err := sdb.DeleteOldCheckpoints()
		require.NoError(t, err)
		checkpoints, err := sdb.db.ListCheckpoints()
		require.NoError(t, err)
		assert.LessOrEqual(t, len(checkpoints), keep)
	}

	sdb.Close()
}

// TestConcurrentDeleteOldCheckpoints performs almost the same test than
// kvdb/kvdb_test.go TestConcurrentDeleteOldCheckpoints, but over the StateDB
func TestConcurrentDeleteOldCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	keep := 16
	sdb, err := NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	numCheckpoints := 32
	// do checkpoints and check that we never have more than `keep`
	// checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = sdb.MakeCheckpoint()
		require.NoError(t, err)
		wg := sync.WaitGroup{}
		n := 10
		wg.Add(n)
		for j := 0; j < n; j++ {
			go func() {
				err := sdb.DeleteOldCheckpoints()
				require.NoError(t, err)
				checkpoints, err := sdb.db.ListCheckpoints()
				require.NoError(t, err)
				assert.LessOrEqual(t, len(checkpoints), keep)
				wg.Done()
			}()
			_, err := sdb.db.ListCheckpoints()
			// only checking here for absence of errors, not the count of checkpoints
			require.NoError(t, err)
		}
		wg.Wait()
		checkpoints, err := sdb.db.ListCheckpoints()
		require.NoError(t, err)
		assert.LessOrEqual(t, len(checkpoints), keep)
	}

	sdb.Close()
}

func TestCurrentIdx(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	keep := 16
	sdb, err := NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	idx := sdb.CurrentIdx()
	assert.Equal(t, common.Idx(255), idx)

	sdb.Close()

	sdb, err = NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	idx = sdb.CurrentIdx()
	assert.Equal(t, common.Idx(255), idx)

	err = sdb.MakeCheckpoint()
	require.NoError(t, err)

	idx = sdb.CurrentIdx()
	assert.Equal(t, common.Idx(255), idx)

	sdb.Close()

	sdb, err = NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	idx = sdb.CurrentIdx()
	assert.Equal(t, common.Idx(255), idx)

	sdb.Close()
}

func TestResetFromBadCheckpoint(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	deleteme = append(deleteme, dir)

	keep := 16
	sdb, err := NewStateDB(Config{Path: dir, Keep: keep, Type: TypeSynchronizer, NLevels: 32})
	require.NoError(t, err)

	err = sdb.MakeCheckpoint()
	require.NoError(t, err)
	err = sdb.MakeCheckpoint()
	require.NoError(t, err)
	err = sdb.MakeCheckpoint()
	require.NoError(t, err)

	// reset from a checkpoint that doesn't exist
	err = sdb.Reset(10)
	require.Error(t, err)

	sdb.Close()
}
