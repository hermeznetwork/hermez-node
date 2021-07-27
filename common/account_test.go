package common

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	cryptoConstants "github.com/iden3/go-iden3-crypto/constants"
	"github.com/iden3/go-iden3-crypto/poseidon"
	cryptoUtils "github.com/iden3/go-iden3-crypto/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdxParser(t *testing.T) {
	i := Idx(1)
	iBytes, err := i.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, 6, len(iBytes))
	assert.Equal(t, "000000000001", hex.EncodeToString(iBytes[:]))
	i2, err := IdxFromBytes(iBytes[:])
	assert.NoError(t, err)
	assert.Equal(t, i, i2)

	i = Idx(100)
	assert.Equal(t, big.NewInt(100), i.BigInt())

	// value before overflow
	i = Idx(281474976710655)
	iBytes, err = i.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, 6, len(iBytes))
	assert.Equal(t, "ffffffffffff", hex.EncodeToString(iBytes[:]))
	i2, err = IdxFromBytes(iBytes[:])
	assert.NoError(t, err)
	assert.Equal(t, i, i2)

	// expect value overflow
	i = Idx(281474976710656)
	iBytes, err = i.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, ErrIdxOverflow, tracerr.Unwrap(err))
}

func TestNonceParser(t *testing.T) {
	n := nonce.Nonce(1)
	nBytes, err := n.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, 5, len(nBytes))
	assert.Equal(t, "0000000001", hex.EncodeToString(nBytes[:]))
	n2 := nonce.FromBytes(nBytes)
	assert.Equal(t, n, n2)

	// value before overflow
	n = nonce.Nonce(1099511627775)
	nBytes, err = n.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, 5, len(nBytes))
	assert.Equal(t, "ffffffffff", hex.EncodeToString(nBytes[:]))
	n2 = nonce.FromBytes(nBytes)
	assert.Equal(t, n, n2)

	// expect value overflow
	n = nonce.Nonce(1099511627776)
	nBytes, err = n.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, nonce.ErrNonceOverflow, tracerr.Unwrap(err))
}

func TestAccount(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	pk := sk.Public()

	account := &Account{
		TokenID: TokenID(1),
		Nonce:   nonce.Nonce(1234),
		Balance: big.NewInt(1000),
		BJJ:     pk.Compress(),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	b, err := account.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, byte(1), b[22])
	a1, err := AccountFromBytes(b)
	assert.NoError(t, err)
	assert.Equal(t, account, a1)

	e, err := account.BigInts()
	assert.NoError(t, err)
	assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
	assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))

	assert.Equal(t, "1000", e[1].String())
	assert.Equal(t, pk.Y.String(), e[2].String())
	assert.Equal(t, new(big.Int).SetBytes(account.EthAddr.Bytes()).String(), e[3].String())

	a2, err := AccountFromBigInts(e)
	assert.NoError(t, err)
	assert.Equal(t, account, a2)
	assert.Equal(t, a1, a2)
}

func TestAccountLoop(t *testing.T) {
	// check that for different deterministic BabyJubJub keys & random Address there is no problem
	for i := 0; i < 256; i++ {
		var sk babyjub.PrivateKey
		_, err := hex.Decode(sk[:],
			[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
		assert.NoError(t, err)
		pk := sk.Public()

		key, err := ethCrypto.GenerateKey()
		assert.NoError(t, err)
		address := ethCrypto.PubkeyToAddress(key.PublicKey)

		account := &Account{
			TokenID: TokenID(i),
			Nonce:   nonce.Nonce(i),
			Balance: big.NewInt(1000),
			BJJ:     pk.Compress(),
			EthAddr: address,
		}
		b, err := account.Bytes()
		assert.NoError(t, err)
		a1, err := AccountFromBytes(b)
		assert.NoError(t, err)
		assert.Equal(t, account, a1)

		e, err := account.BigInts()
		assert.NoError(t, err)
		assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))

		a2, err := AccountFromBigInts(e)
		assert.NoError(t, err)
		assert.Equal(t, account, a2)
	}
}

func TestAccountLoopRandom(t *testing.T) {
	// check that for different random Address & BabyJubJub keys there is
	// no problem
	for i := 0; i < 256; i++ {
		sk := babyjub.NewRandPrivKey()
		pk := sk.Public()

		key, err := ethCrypto.GenerateKey()
		assert.NoError(t, err)
		address := ethCrypto.PubkeyToAddress(key.PublicKey)

		account := &Account{
			TokenID: TokenID(i),
			Nonce:   nonce.Nonce(i),
			Balance: big.NewInt(1000),
			BJJ:     pk.Compress(),
			EthAddr: address,
		}
		b, err := account.Bytes()
		assert.NoError(t, err)
		a1, err := AccountFromBytes(b)
		assert.NoError(t, err)
		assert.Equal(t, account, a1)

		e, err := account.BigInts()
		assert.NoError(t, err)
		assert.True(t, cryptoUtils.CheckBigIntInField(e[0]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[1]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[2]))
		assert.True(t, cryptoUtils.CheckBigIntInField(e[3]))

		a2, err := AccountFromBigInts(e)
		assert.NoError(t, err)
		assert.Equal(t, account, a2)
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

func TestAccountHashValue(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	pk := sk.Public()

	account := &Account{
		TokenID: TokenID(1),
		Nonce:   nonce.Nonce(1234),
		Balance: big.NewInt(1000),
		BJJ:     pk.Compress(),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	v, err := account.HashValue()
	assert.NoError(t, err)
	assert.Equal(t,
		"447675324273474410516096114710387312413478475468606444107594732044698919451",
		v.String())
}

func TestAccountHashValueTestVectors(t *testing.T) {
	// values from js test vectors
	ay := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(253), nil), big.NewInt(1))
	assert.Equal(t, "1fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		(hex.EncodeToString(ay.Bytes())))
	bjjPoint, err := babyjub.PointFromSignAndY(true, ay)
	require.NoError(t, err)
	bjj := babyjub.PublicKey(*bjjPoint)

	account := &Account{
		Idx:     1,
		TokenID: 0xFFFFFFFF,
		BJJ:     bjj.Compress(),
		EthAddr: ethCommon.HexToAddress("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
		Nonce:   nonce.Nonce(0xFFFFFFFFFF),
		Balance: bigFromStr("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16),
	}

	e, err := account.BigInts()
	assert.NoError(t, err)
	assert.Equal(t, "9444732965739290427391", e[0].String())
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512895", e[1].String())
	assert.Equal(t,
		"14474011154664524427946373126085988481658748083205070504932198000989141204991",
		e[2].String())
	assert.Equal(t, "1461501637330902918203684832716283019655932542975", e[3].String())

	h, err := poseidon.Hash(e[:])
	assert.NoError(t, err)
	assert.Equal(t,
		"13265203488631320682117942952393454767418777767637549409684833552016769103047",
		h.String())

	v, err := account.HashValue()
	assert.NoError(t, err)
	assert.Equal(t,
		"13265203488631320682117942952393454767418777767637549409684833552016769103047",
		v.String())

	// second account
	ay = big.NewInt(0)
	bjjPoint, err = babyjub.PointFromSignAndY(false, ay)
	require.NoError(t, err)
	bjj = babyjub.PublicKey(*bjjPoint)
	account = &Account{
		TokenID: 0,
		BJJ:     bjj.Compress(),
		EthAddr: ethCommon.HexToAddress("0x00"),
		Nonce:   nonce.Nonce(0),
		Balance: big.NewInt(0),
	}
	v, err = account.HashValue()
	assert.NoError(t, err)
	assert.Equal(t,
		"2351654555892372227640888372176282444150254868378439619268573230312091195718",
		v.String())

	// third account
	ay = bigFromStr("21b0a1688b37f77b1d1d5539ec3b826db5ac78b2513f574a04c50a7d4f8246d7", 16)
	bjjPoint, err = babyjub.PointFromSignAndY(false, ay)
	require.NoError(t, err)
	bjj = babyjub.PublicKey(*bjjPoint)
	account = &Account{
		TokenID: 3,
		BJJ:     bjj.Compress(),
		EthAddr: ethCommon.HexToAddress("0xA3C88ac39A76789437AED31B9608da72e1bbfBF9"),
		Nonce:   nonce.Nonce(129),
		Balance: bigFromStr("42000000000000000000", 10),
	}
	e, err = account.BigInts()
	assert.NoError(t, err)
	assert.Equal(t, "554050781187", e[0].String())
	assert.Equal(t, "42000000000000000000", e[1].String())
	assert.Equal(t,
		"15238403086306505038849621710779816852318505119327426213168494964113886299863",
		e[2].String())
	assert.Equal(t, "935037732739828347587684875151694054123613453305", e[3].String())
	v, err = account.HashValue()
	assert.NoError(t, err)
	assert.Equal(t,
		"15036148928138382129196903417666258171042923749783835283230591475172197254845",
		v.String())
}

func TestAccountErrNotInFF(t *testing.T) {
	z := big.NewInt(0)

	// Q-1 should not give error
	r := new(big.Int).Sub(cryptoConstants.Q, big.NewInt(1))
	e := [NLeafElems]*big.Int{z, z, r, r}
	_, err := AccountFromBigInts(e)
	assert.NoError(t, err)

	// Q should give error
	r = cryptoConstants.Q
	e = [NLeafElems]*big.Int{z, z, r, r}
	_, err = AccountFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, tracerr.Unwrap(err))

	// Q+1 should give error
	r = new(big.Int).Add(cryptoConstants.Q, big.NewInt(1))
	e = [NLeafElems]*big.Int{z, z, r, r}
	_, err = AccountFromBigInts(e)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotInFF, tracerr.Unwrap(err))
}

func TestAccountErrNumOverflowNonce(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	pk := sk.Public()

	// check limit
	account := &Account{
		TokenID: TokenID(1),
		Nonce:   nonce.Nonce(math.Pow(2, 40) - 1),
		Balance: big.NewInt(1000),
		BJJ:     pk.Compress(),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	_, err = account.Bytes()
	assert.NoError(t, err)

	// force value overflow
	account.Nonce = nonce.Nonce(math.Pow(2, 40))
	b, err := account.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Nonce", ErrNumOverflow), tracerr.Unwrap(err))

	_, err = AccountFromBytes(b)
	assert.NoError(t, err)
}

func TestAccountErrNumOverflowBalance(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	pk := sk.Public()

	// check limit
	account := &Account{
		TokenID: TokenID(1),
		Nonce:   nonce.Nonce(math.Pow(2, 40) - 1),
		Balance: new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil), big.NewInt(1)),
		BJJ:     pk.Compress(),
		EthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512895",
		account.Balance.String())

	_, err = account.Bytes()
	assert.NoError(t, err)

	// force value overflow
	account.Balance = new(big.Int).Exp(big.NewInt(2), big.NewInt(192), nil)
	assert.Equal(t, "6277101735386680763835789423207666416102355444464034512896",
		account.Balance.String())
	b, err := account.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), tracerr.Unwrap(err))

	_, err = AccountFromBytes(b)
	assert.NoError(t, err)

	b[39] = 1
	_, err = AccountFromBytes(b)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("%s Balance", ErrNumOverflow), tracerr.Unwrap(err))
}
