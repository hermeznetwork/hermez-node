package common

import (
	"encoding/hex"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

func TestNonceParser(t *testing.T) {
	n := Nonce(1)
	nBytes, err := n.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(nBytes))
	assert.Equal(t, "0100000000", hex.EncodeToString(nBytes[:]))
	n2 := NonceFromBytes(nBytes)
	assert.Equal(t, n, n2)

	// value before overflow
	n = Nonce(1099511627775)
	nBytes, err = n.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, 5, len(nBytes))
	assert.Equal(t, "ffffffffff", hex.EncodeToString(nBytes[:]))
	n2 = NonceFromBytes(nBytes)
	assert.Equal(t, n, n2)

	// expect value overflow
	n = Nonce(1099511627776)
	nBytes, err = n.Bytes()
	assert.NotNil(t, err)
	assert.Equal(t, ErrNonceOverflow, err)
}

func TestTxCompressedData(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)

	tx := PoolL2Tx{
		FromIdx: 2,
		ToIdx:   3,
		Amount:  big.NewInt(4),
		TokenID: 5,
		Nonce:   6,
		ToBJJ:   sk.Public(),
	}
	txCompressedData, err := tx.TxCompressedData()
	assert.Nil(t, err)
	// test vector value generated from javascript implementation
	assert.Equal(t, "1766847064778421992193717128424891165872736891548909569553540449389241871", txCompressedData.String())
	assert.Equal(t, "10000000000060000000500040000000000030000000000020001c60be60f", hex.EncodeToString(txCompressedData.Bytes())[1:])

	tx = PoolL2Tx{
		FromIdx: 7,
		ToIdx:   8,
		Amount:  big.NewInt(9),
		TokenID: 10,
		Nonce:   11,
		Fee:     12,
		ToBJJ:   sk.Public(),
	}
	txCompressedData, err = tx.TxCompressedDataV2()
	assert.Nil(t, err)
	// test vector value generated from javascript implementation
	assert.Equal(t, "6571340879233176732837827812956721483162819083004853354503", txCompressedData.String())
	assert.Equal(t, "10c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes())[1:])
}

func TestHashToSign(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	ethAddr := ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370")

	tx := PoolL2Tx{
		FromIdx:     2,
		ToIdx:       3,
		Amount:      big.NewInt(4),
		TokenID:     5,
		Nonce:       6,
		ToBJJ:       sk.Public(),
		RqToEthAddr: ethAddr,
		RqToBJJ:     sk.Public(),
	}
	toSign, err := tx.HashToSign()
	assert.Nil(t, err)
	assert.Equal(t, "14526446928649310956370997581245770629723313742905751117262272426489782809503", toSign.String())
}

func TestVerifyTxSignature(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	ethAddr := ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370")

	tx := PoolL2Tx{
		FromIdx:     2,
		ToIdx:       3,
		Amount:      big.NewInt(4),
		TokenID:     5,
		Nonce:       6,
		ToBJJ:       sk.Public(),
		RqToEthAddr: ethAddr,
		RqToBJJ:     sk.Public(),
	}
	toSign, err := tx.HashToSign()
	assert.Nil(t, err)
	assert.Equal(t, "14526446928649310956370997581245770629723313742905751117262272426489782809503", toSign.String())

	sig := sk.SignPoseidon(toSign)
	tx.Signature = sig
	assert.True(t, tx.VerifySignature(sk.Public()))
}
