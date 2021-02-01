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
	assert.Equal(t, "0x02fb52b5d0b9ef2626c11701bb751b2720c76d59946b9a48146ac153bb6e63bf6a", poolL2Tx.TxID.String())
}

func TestTxCompressedData(t *testing.T) {
	chainID := uint16(0)
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	tx := PoolL2Tx{
		FromIdx: 2,
		ToIdx:   3,
		Amount:  big.NewInt(4),
		TokenID: 5,
		Nonce:   6,
		ToBJJ:   sk.Public().Compress(),
	}
	txCompressedData, err := tx.TxCompressedData(chainID)
	assert.NoError(t, err)
	// test vector value generated from javascript implementation
	expectedStr := "1766847064778421992193717128424891165872736891548909569553540445094274575"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "010000000000060000000500040000000000030000000000020000c60be60f", hex.EncodeToString(txCompressedData.Bytes()))
	// using a different chainID
	txCompressedData, err = tx.TxCompressedData(uint16(100))
	assert.NoError(t, err)
	expectedStr = "1766847064778421992193717128424891165872736891548909569553540874591004175"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "010000000000060000000500040000000000030000000000020064c60be60f", hex.EncodeToString(txCompressedData.Bytes()))
	txCompressedData, err = tx.TxCompressedData(uint16(65535))
	assert.NoError(t, err)
	expectedStr = "1766847064778421992193717128424891165872736891548909569553821915776017935"
	assert.Equal(t, expectedStr, txCompressedData.String())
	assert.Equal(t, "01000000000006000000050004000000000003000000000002ffffc60be60f", hex.EncodeToString(txCompressedData.Bytes()))

	tx = PoolL2Tx{
		RqFromIdx: 7,
		RqToIdx:   8,
		RqAmount:  big.NewInt(9),
		RqTokenID: 10,
		RqNonce:   11,
		RqFee:     12,
		RqToBJJ:   sk.Public().Compress(),
	}
	rqTxCompressedData, err := tx.RqTxCompressedDataV2()
	assert.NoError(t, err)
	// test vector value generated from javascript implementation
	expectedStr = "6571340879233176732837827812956721483162819083004853354503"
	assert.Equal(t, expectedStr, rqTxCompressedData.String())
	assert.Equal(t, "010c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(rqTxCompressedData.Bytes()))
}

func TestTxCompressedDataV2(t *testing.T) {
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
	tx := PoolL2Tx{
		FromIdx: 7,
		ToIdx:   8,
		Amount:  big.NewInt(9),
		TokenID: 10,
		Nonce:   11,
		Fee:     12,
		ToBJJ:   sk.Public().Compress(),
	}
	txCompressedData, err := tx.TxCompressedDataV2()
	assert.NoError(t, err)
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
	expectedStr := "6571340879233176732837827812956721483162819083004853354503"
	assert.Equal(t, expectedStr, txCompressedData.String())
	expected, ok := new(big.Int).SetString(expectedStr, 10)
	assert.True(t, ok)
	assert.Equal(t, expected.Bytes(), txCompressedData.Bytes())
	assert.Equal(t, "010c000000000b0000000a0009000000000008000000000007", hex.EncodeToString(txCompressedData.Bytes()))
}

func TestHashToSign(t *testing.T) {
	chainID := uint16(0)
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:], []byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)
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
	assert.Equal(t, "1469900657138253851938022936440971384682713995864967090251961124784132925291", toSign.String())
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
	assert.Equal(t, "18645218094210271622244722988708640202588315450486586312909439859037906375295", toSign.String())

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
