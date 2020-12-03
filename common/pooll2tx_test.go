package common

import (
	"encoding/hex"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

func TestNewPoolL2Tx(t *testing.T) {
	poolL2Tx := &PoolL2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		Amount:  big.NewInt(4),
		TokenID: 5,
		Nonce:   144,
	}
	poolL2Tx, err := NewPoolL2Tx(poolL2Tx)
	assert.Nil(t, err)
	assert.Equal(t, "0x020000000156660000000090", poolL2Tx.TxID.String())
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
	expectedStr := "1766847064778421992193717128424891165872736891548909569553540445094274575"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "010000000000060000000500040000000000030000000000020000c60be60f", hex.EncodeToString(txCompressedData.Bytes()))
	tx = PoolL2Tx{
		RqFromIdx: 7,
		RqToIdx:   8,
		RqAmount:  big.NewInt(9),
		RqTokenID: 10,
		RqNonce:   11,
		RqFee:     12,
		RqToBJJ:   sk.Public(),
	}
	txCompressedData, err = tx.RqTxCompressedDataV2()
	assert.Nil(t, err)
	// test vector value generated from javascript implementation
	expectedStr = "6571340879233176732837827812956721483162819083004853354503"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "010c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestTxCompressedDataV2(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	tx := PoolL2Tx{
		FromIdx: 7,
		ToIdx:   8,
		Amount:  big.NewInt(9),
		TokenID: 10,
		Nonce:   11,
		Fee:     12,
		ToBJJ:   sk.Public(),
	}
	txCompressedData, err := tx.TxCompressedDataV2()
	assert.Nil(t, err)
	// test vector value generated from javascript implementation
	expectedStr := "6571340879233176732837827812956721483162819083004853354503"
	assert.Equal(t, expectedStr, txCompressedData.String())
	expected, ok := new(big.Int).SetString(expectedStr, 10)
	assert.True(t, ok)

	assert.Equal(t, expected.Bytes(), txCompressedData.Bytes())
	assert.Equal(t, "010c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestRqTxCompressedDataV2(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	tx := PoolL2Tx{
		RqFromIdx: 7,
		RqToIdx:   8,
		RqAmount:  big.NewInt(9),
		RqTokenID: 10,
		RqNonce:   11,
		RqFee:     12,
		RqToBJJ:   sk.Public(),
	}
	txCompressedData, err := tx.RqTxCompressedDataV2()
	assert.Nil(t, err)
	// test vector value generated from javascript implementation
	expectedStr := "6571340879233176732837827812956721483162819083004853354503"
	assert.Equal(t, expectedStr, txCompressedData.String())
	expected, ok := new(big.Int).SetString(expectedStr, 10)
	assert.True(t, ok)
	assert.Equal(t, expected.Bytes(), txCompressedData.Bytes())
	assert.Equal(t, "010c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestHashToSign(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	tx := PoolL2Tx{
		FromIdx:   2,
		ToIdx:     3,
		Amount:    big.NewInt(4),
		TokenID:   5,
		Nonce:     6,
		ToEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	toSign, err := tx.HashToSign()
	assert.Nil(t, err)
	assert.Equal(t, "1469900657138253851938022936440971384682713995864967090251961124784132925291", toSign.String())
}

func TestVerifyTxSignature(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.Nil(t, err)
	tx := PoolL2Tx{
		FromIdx:     2,
		ToIdx:       3,
		Amount:      big.NewInt(4),
		TokenID:     5,
		Nonce:       6,
		ToBJJ:       sk.Public(),
		RqToEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
		RqToBJJ:     sk.Public(),
	}
	toSign, err := tx.HashToSign()
	assert.Nil(t, err)
	assert.Equal(t, "18645218094210271622244722988708640202588315450486586312909439859037906375295", toSign.String())

	sig := sk.SignPoseidon(toSign)
	tx.Signature = sig.Compress()
	assert.True(t, tx.VerifySignature(sk.Public()))
}
