package common

import (
	"encoding/hex"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.NoError(t, err)
	assert.Equal(t, "0x022669acda59b827d20ef5354a3eebd1dffb3972b0a6bf89d18bfd2efa0ab9f41e", poolL2Tx.TxID.String())
}

func TestTxCompressedDataAndTxCompressedDataV2JSVectors(t *testing.T) {
	// test vectors values generated from javascript implementation
	var skPositive babyjub.PrivateKey // 'Positive' refers to the sign
	_, err := hex.Decode(skPositive[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)

	var skNegative babyjub.PrivateKey // 'Negative' refers to the sign
	_, err = hex.Decode(skNegative[:], []byte("0001020304050607080900010203040506070809000102030405060708090002"))
	assert.NoError(t, err)

	amount, ok := new(big.Int).SetString("343597383670000000000000000000000000000000", 10)
	require.True(t, ok)
	tx := PoolL2Tx{
		FromIdx: (1 << 48) - 1,
		ToIdx:   (1 << 48) - 1,
		Amount:  amount,
		TokenID: (1 << 32) - 1,
		Nonce:   (1 << 40) - 1,
		Fee:     (1 << 3) - 1,
		ToBJJ:   skPositive.Public().Compress(),
	}
	txCompressedData, err := tx.TxCompressedData(uint16((1 << 16) - 1))
	require.NoError(t, err)
	expectedStr := "0107ffffffffffffffffffffffffffffffffffffffffffffffc60be60f"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedData.Bytes()))

	txCompressedDataV2, err := tx.TxCompressedDataV2()
	require.NoError(t, err)
	expectedStr = "0107ffffffffffffffffffffffffffffffffffffffffffffffffffff"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedDataV2.Bytes()))

	tx = PoolL2Tx{
		FromIdx: 0,
		ToIdx:   0,
		Amount:  big.NewInt(0),
		TokenID: 0,
		Nonce:   0,
		Fee:     0,
		ToBJJ:   skNegative.Public().Compress(),
	}
	txCompressedData, err = tx.TxCompressedData(uint16(0))
	require.NoError(t, err)
	expectedStr = "c60be60f"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedData.Bytes()))

	txCompressedDataV2, err = tx.TxCompressedDataV2()
	require.NoError(t, err)
	assert.Equal(t, "0", txCompressedDataV2.String())

	amount, ok = new(big.Int).SetString("63000000000000000", 10)
	require.True(t, ok)
	tx = PoolL2Tx{
		FromIdx: 324,
		ToIdx:   256,
		Amount:  amount,
		TokenID: 123,
		Nonce:   76,
		Fee:     214,
		ToBJJ:   skNegative.Public().Compress(),
	}
	txCompressedData, err = tx.TxCompressedData(uint16(1))
	require.NoError(t, err)
	expectedStr = "d6000000004c0000007b0000000001000000000001440001c60be60f"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedData.Bytes()))

	txCompressedDataV2, err = tx.TxCompressedDataV2()
	require.NoError(t, err)
	expectedStr = "d6000000004c0000007b3977825f00000000000100000000000144"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedDataV2.Bytes()))

	tx = PoolL2Tx{
		FromIdx: 1,
		ToIdx:   2,
		TokenID: 3,
		Nonce:   4,
		Fee:     5,
		ToBJJ:   skNegative.Public().Compress(),
	}
	txCompressedData, err = tx.TxCompressedData(uint16(0))
	require.NoError(t, err)
	expectedStr = "050000000004000000030000000000020000000000010000c60be60f"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedData.Bytes()))

	tx = PoolL2Tx{
		FromIdx: 2,
		ToIdx:   3,
		TokenID: 4,
		Nonce:   5,
		Fee:     6,
		ToBJJ:   skPositive.Public().Compress(),
	}
	txCompressedData, err = tx.TxCompressedData(uint16(0))
	require.NoError(t, err)
	expectedStr = "01060000000005000000040000000000030000000000020000c60be60f"
	assert.Equal(t, expectedStr, hex.EncodeToString(txCompressedData.Bytes()))
}

func TestRqTxCompressedDataV2(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	tx := PoolL2Tx{
		RqFromIdx: 7,
		RqToIdx:   8,
		RqAmount:  big.NewInt(9),
		RqTokenID: 10,
		RqNonce:   11,
		RqFee:     12,
		RqToBJJ:   sk.Public().Compress(),
	}
	txCompressedData, err := tx.RqTxCompressedDataV2()
	assert.NoError(t, err)
	// test vector value generated from javascript implementation
	expectedStr := "110248805340524920412994530176819463725852160917809517418728390663"
	assert.Equal(t, expectedStr, txCompressedData.String())
	expected, ok := new(big.Int).SetString(expectedStr, 10)
	assert.True(t, ok)
	assert.Equal(t, expected.Bytes(), txCompressedData.Bytes())
	assert.Equal(t, "010c000000000b0000000a0000000009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestHashToSign(t *testing.T) {
	chainID := uint16(0)
	tx := PoolL2Tx{
		FromIdx:   2,
		ToIdx:     3,
		Amount:    big.NewInt(4),
		TokenID:   5,
		Nonce:     6,
		ToEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
	}
	toSign, err := tx.HashToSign(chainID)
	assert.NoError(t, err)
	assert.Equal(t, "2d49ce1d4136e06f64e3eb1f79a346e6ee3e93ceeac909a57806a8d87005c263", hex.EncodeToString(toSign.Bytes()))
}

func TestVerifyTxSignature(t *testing.T) {
	chainID := uint16(0)
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	tx := PoolL2Tx{
		FromIdx:     2,
		ToIdx:       3,
		Amount:      big.NewInt(4),
		TokenID:     5,
		Nonce:       6,
		ToBJJ:       sk.Public().Compress(),
		RqToEthAddr: ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
		RqToBJJ:     sk.Public().Compress(),
	}
	toSign, err := tx.HashToSign(chainID)
	assert.NoError(t, err)
	assert.Equal(t, "1571327027383224465388301747239444557034990637650927918405777653988509342917", toSign.String())

	sig := sk.SignPoseidon(toSign)
	tx.Signature = sig.Compress()
	assert.True(t, tx.VerifySignature(chainID, sk.Public().Compress()))
}

func TestDecompressEmptyBJJComp(t *testing.T) {
	pkComp := EmptyBJJComp
	pk, err := pkComp.Decompress()
	require.NoError(t, err)
	assert.Equal(t, "2957874849018779266517920829765869116077630550401372566248359756137677864698", pk.X.String())
	assert.Equal(t, "0", pk.Y.String())
}

func TestPoolL2TxID(t *testing.T) {
	tx0 := PoolL2Tx{
		FromIdx: 5,
		ToIdx:   5,
		Amount:  big.NewInt(5),
		Fee:     126,
		TokenID: 5,
		Nonce:   5,
	}
	err := tx0.SetID()
	require.NoError(t, err)

	// differ TokenID
	tx1 := PoolL2Tx{
		FromIdx: 5,
		ToIdx:   5,
		Amount:  big.NewInt(5),
		Fee:     126,
		TokenID: 4,
		Nonce:   5,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)
	// differ Nonce
	tx1 = PoolL2Tx{
		FromIdx: 5,
		ToIdx:   5,
		Amount:  big.NewInt(5),
		Fee:     126,
		TokenID: 5,
		Nonce:   4,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)
	// differ Fee
	tx1 = PoolL2Tx{
		FromIdx: 5,
		ToIdx:   5,
		Amount:  big.NewInt(5),
		Fee:     124,
		TokenID: 5,
		Nonce:   5,
	}
	err = tx1.SetID()
	require.NoError(t, err)
	assert.NotEqual(t, tx0.TxID, tx1.TxID)
}
