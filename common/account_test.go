package common

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
	cryptoConstants "github.com/iden3/go-iden3-crypto/constants"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
)

func TestAccount(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	pk := sk.Public()

	account := &Account{
		TokenID:   TokenID(1),
		Nonce:     Nonce(1234),
		Balance:   big.NewInt(1000),
		PublicKey: pk,
		EthAddr:   ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	b, err := account.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, byte(1), b[10])
	a1, err := AccountFromBytes(b)
	assert.Nil(t, err)
	assert.Equal(t, account, a1)

	e, err := account.BigInts()
	assert.Nil(t, err)
	assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))

	assert.Equal(t, "1000", e[1].String())
	assert.Equal(t, pk.Y.String(), e[2].String())
	assert.Equal(t, new(big.Int).SetBytes(SwapEndianness(account.EthAddr.Bytes())).String(), e[3].String())

	a2, err := AccountFromBigInts(e)
	assert.Nil(t, err)
	assert.Equal(t, account, a2)
	assert.Equal(t, a1, a2)
}

func TestAccountLoop(t *testing.T) {
	// check that for different Address there is no problem
	for i := 0; i < 256; i++ {
		var sk babyjub.PrivateKey
		_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
		assert.Nil(t, err)
		pk := sk.Public()

		key, err := ethCrypto.GenerateKey()
		assert.Nil(t, err)
		address := ethCrypto.PubkeyToAddress(key.PublicKey)

		account := &Account{
			TokenID:   TokenID(i),
			Nonce:     Nonce(i),
			Balance:   big.NewInt(1000),
			PublicKey: pk,
			EthAddr:   address,
		}
		b, err := account.Bytes()
		assert.Nil(t, err)
		a1, err := AccountFromBytes(b)
		assert.Nil(t, err)
		assert.Equal(t, account, a1)

		e, err := account.BigInts()
		assert.Nil(t, err)
		assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))

		a2, err := AccountFromBigInts(e)
		assert.Nil(t, err)
		assert.Equal(t, account, a2)
	}
}

func TestAccountHashValue(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	pk := sk.Public()

	account := &Account{
		TokenID:   TokenID(1),
		Nonce:     Nonce(1234),
		Balance:   big.NewInt(1000),
		PublicKey: pk,
		EthAddr:   ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}

	v, err := account.HashValue()
	assert.Nil(t, err)
	assert.Equal(t, "16085711911723375585301279875451049849443101031421093098714359651259271023730", v.String())
}

func TestAccountErrNotInFF(t *testing.T) {
	z := big.NewInt(0)

	// Q-1 should not give error
	r := new(big.Int).Sub(cryptoConstants.Q, big.NewInt(1))
	e := [NLeafElems]*big.Int{z, z, r, r}
	_, err := AccountFromBigInts(e)
	assert.Nil(t, err)

	// Q should give error
	r = cryptoConstants.Q
	e = [NLeafElems]*big.Int{z, z, r, r}
	_, err = AccountFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, err)

	// Q+1 should give error
	r = new(big.Int).Add(cryptoConstants.Q, big.NewInt(1))
	e = [NLeafElems]*big.Int{z, z, r, r}
	_, err = AccountFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, err)
}

func TestAccountErrNumOverflowNonce(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	pk := sk.Public()

	// check limit
	account := &Account{
		TokenID:   TokenID(1),
		Nonce:     Nonce(math.Pow(2, 40) - 1),
		Balance:   big.NewInt(1000),
		PublicKey: pk,
		EthAddr:   ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	_, err = account.Bytes()
	assert.Nil(t, err)

	// force value overflow
	account.Nonce = Nonce(math.Pow(2, 40))
	b, err := account.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Nonce", ErrNumOverflow), err)

	_, err = AccountFromBytes(b)
	assert.Nil(t, err)
}

func TestAccountErrNumOverflowBalance(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	pk := sk.Public()

	// check limit
	account := &Account{
		TokenID:   TokenID(1),
		Nonce:     Nonce(math.Pow(2, 40) - 1),
		Balance:   new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil), big.NewInt(1)),
		PublicKey: pk,
		EthAddr:   ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512895", account.Balance.String())

	_, err = account.Bytes()
	assert.Nil(t, err)

	// force value overflow
	account.Balance = new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil)
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512896", account.Balance.String())
	b, err := account.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), err)

	_, err = AccountFromBytes(b)
	assert.Nil(t, err)

	b[56] = 1
	_, err = AccountFromBytes(b)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), err)
}
