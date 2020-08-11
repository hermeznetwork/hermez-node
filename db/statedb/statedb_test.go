package statedb

import (
	"encoding/hex"
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

	sdb, err := NewStateDB(dir, false, false, 0)
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
		err = sdb.CreateAccount(common.Idx(i), accounts[i])
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
	err = sdb.CreateAccount(common.Idx(1), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		err = sdb.UpdateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}

	// check that can not call MerkleTree methods of the StateDB
	_, err = sdb.MTCreateAccount(common.Idx(1), accounts[1])
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, err)

	_, err = sdb.MTUpdateAccount(common.Idx(1), accounts[1])
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, err)

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.NotNil(t, err)
	assert.Equal(t, ErrStateDBWithoutMT, err)
}

func TestStateDBWithMT(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	sdb, err := NewStateDB(dir, false, true, 32)
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
		_, err = sdb.MTCreateAccount(common.Idx(i), accounts[i])
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
	_, err = sdb.MTCreateAccount(common.Idx(1), accounts[1]) // check that can not be created twice
	assert.NotNil(t, err)
	assert.Equal(t, ErrAccountAlreadyExists, err)

	_, err = sdb.MTGetProof(common.Idx(1))
	assert.Nil(t, err)

	// update accounts
	for i := 0; i < len(accounts); i++ {
		accounts[i].Nonce = accounts[i].Nonce + 1
		_, err = sdb.MTUpdateAccount(common.Idx(i), accounts[i])
		assert.Nil(t, err)
	}
	a, err := sdb.GetAccount(common.Idx(1)) // check that account value has been updated
	assert.Nil(t, err)
	assert.Equal(t, accounts[1].Nonce, a.Nonce)
}
