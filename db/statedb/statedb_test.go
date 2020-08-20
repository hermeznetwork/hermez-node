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
		Nonce:     uint64(i),
		Balance:   big.NewInt(1000),
		PublicKey: pk,
		EthAddr:   address,
	}

}

func TestStateDBWithoutMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, false, 0)
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

	sdb, err := NewStateDB(dir, true, 32)
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

	sdb, err := NewStateDB(dir, true, 32)
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
	assert.Equal(t, uint64(1), cb)

	for i := 1; i < 10; i++ {
		err = sdb.MakeCheckpoint()
		assert.Nil(t, err)

		cb, err = sdb.GetCurrentBatch()
		assert.Nil(t, err)
		assert.Equal(t, uint64(i+1), cb)
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
	assert.Equal(t, uint64(3), cb)

	// advance one checkpoint and check that currentBatch is fine
	err = sdb.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = sdb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), cb)

	err = sdb.DeleteCheckpoint(uint64(9))
	assert.Nil(t, err)
	err = sdb.DeleteCheckpoint(uint64(10))
	assert.Nil(t, err)
	err = sdb.DeleteCheckpoint(uint64(9)) // does not exist, should return err
	assert.NotNil(t, err)
	err = sdb.DeleteCheckpoint(uint64(11)) // does not exist, should return err
	assert.NotNil(t, err)

	// Create a LocalStateDB from the initial StateDB
	dirLocal, err := ioutil.TempDir("", "ldb")
	require.Nil(t, err)
	ldb, err := NewLocalStateDB(dirLocal, sdb, true, 32)
	assert.Nil(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb.Reset(4, true)
	assert.Nil(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), cb)
	// advance one checkpoint in ldb
	err = ldb.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = ldb.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), cb)

	// Create a 2nd LocalStateDB from the initial StateDB
	dirLocal2, err := ioutil.TempDir("", "ldb2")
	require.Nil(t, err)
	ldb2, err := NewLocalStateDB(dirLocal2, sdb, true, 32)
	assert.Nil(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb2.Reset(4, true)
	assert.Nil(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb2.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, uint64(4), cb)
	// advance one checkpoint in ldb2
	err = ldb2.MakeCheckpoint()
	assert.Nil(t, err)
	cb, err = ldb2.GetCurrentBatch()
	assert.Nil(t, err)
	assert.Equal(t, uint64(5), cb)

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
