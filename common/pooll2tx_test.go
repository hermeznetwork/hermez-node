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
	assert.Equal(t, "0x022669acda59b827d20ef5354a3eebd1dffb3972b0a6bf89d18bfd2efa0ab9f41e",
		poolL2Tx.TxID.String())
}

func TestTxCompressedDataAndTxCompressedDataV2JSVectors(t *testing.T) {
	// test vectors values generated from javascript implementation
	var skPositive babyjub.PrivateKey // 'Positive' refers to the sign
	_, err := hex.Decode(skPositive[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
	assert.NoError(t, err)

	var skNegative babyjub.PrivateKey // 'Negative' refers to the sign
	_, err = hex.Decode(skNegative[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090002"))
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
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
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
	assert.Equal(t, "010c000000000b0000000a0000000009000000000008000000000007",
		hex.EncodeToString(txCompressedData.Bytes()))
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
	assert.Equal(t, "0b8abaf6b7933464e4450df2514da8b72606c02bf7f89bf6e54816fbda9d9d57",
		hex.EncodeToString(toSign.Bytes()))
}

func TestVerifyTxSignature(t *testing.T) {
	chainID := uint16(0)
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("0001020304050607080900010203040506070809000102030405060708090001"))
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
	assert.Equal(t,
		"3144939470626721092564692894890580265754250231349521601298746071096761507003",
		toSign.String())

	sig := sk.SignPoseidon(toSign)
	tx.Signature = sig.Compress()
	assert.True(t, tx.VerifySignature(chainID, sk.Public().Compress()))
}

func TestVerifyTxSignatureEthAddrWith0(t *testing.T) {
	chainID := uint16(5)
	var sk babyjub.PrivateKey
	_, err := hex.Decode(sk[:],
		[]byte("02f0b4f87065af3797aaaf934e8b5c31563c17f2272fa71bd0146535bfbb4184"))
	assert.NoError(t, err)
	tx := PoolL2Tx{
		FromIdx:   10659,
		ToIdx:     0,
		ToEthAddr: ethCommon.HexToAddress("0x0004308BD15Ead4F1173624dC289DBdcC806a309"),
		Amount:    big.NewInt(5000),
		TokenID:   0,
		Nonce:     946,
		Fee:       231,
	}
	toSign, err := tx.HashToSign(chainID)
	assert.NoError(t, err)

	sig := sk.SignPoseidon(toSign)
	assert.Equal(t,
		"f208b8298d5f37148ac3c0c03703272ea47b9f836851bcf8dd5f7e4e3b336ca1d2f6e92ad85dc25f174daf7a0abfd5f71dead3f059b783f4c4b2f56a18a47000",
		sig.Compress().String(),
	)
	tx.Signature = sig.Compress()
	assert.True(t, tx.VerifySignature(chainID, sk.Public().Compress()))
}

func TestDecompressEmptyBJJComp(t *testing.T) {
	pkComp := EmptyBJJComp
	pk, err := pkComp.Decompress()
	require.NoError(t, err)
	assert.Equal(t,
		"2957874849018779266517920829765869116077630550401372566248359756137677864698",
		pk.X.String())
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

func TestPoolL2Tx_SetType(t *testing.T) {
	bjjAddr := [32]byte{
		212, 229, 103, 64, 248, 118, 174, 248,
		192, 16, 184, 106, 64, 213, 245, 103,
		69, 161, 24, 208, 144, 106, 52, 230,
		154, 236, 140, 13, 177, 203, 143, 163,
	}
	ethAddr := ethCommon.HexToAddress("0x7ffC57839B00206D1ad20c69A1981b489f772031")
	tests := []struct {
		name    string
		tx      *PoolL2Tx
		want    TxType
		wantErr bool
	}{
		{
			"Send to bjj address",
			&PoolL2Tx{ToBJJ: bjjAddr, ToEthAddr: FFAddr, ToIdx: Idx(0)},
			TxTypeTransferToBJJ,
			false,
		}, {
			"Send to eth address",
			&PoolL2Tx{ToBJJ: EmptyBJJComp, ToEthAddr: ethAddr, ToIdx: Idx(0)},
			TxTypeTransferToEthAddr,
			false,
		}, {
			"Send to eth FFAddr address",
			&PoolL2Tx{ToBJJ: EmptyBJJComp, ToEthAddr: FFAddr, ToIdx: Idx(0)},
			TxTypeTransferToEthAddr,
			true,
		}, {
			"Send to idx",
			&PoolL2Tx{ToBJJ: EmptyBJJComp, ToEthAddr: EmptyAddr, ToIdx: Idx(400)},
			TxTypeTransfer,
			false,
		}, {
			"Empty transfer",
			&PoolL2Tx{ToBJJ: EmptyBJJComp, ToEthAddr: EmptyAddr, ToIdx: Idx(0)},
			TxType(""),
			true,
		}, {
			"Empty transfer and FFAddr",
			&PoolL2Tx{ToBJJ: EmptyBJJComp, ToEthAddr: FFAddr, ToIdx: Idx(0)},
			TxType(""),
			true,
		}, {
			"Send to eth and bjj addresses and idx",
			&PoolL2Tx{ToBJJ: bjjAddr, ToEthAddr: ethAddr, ToIdx: Idx(400)},
			TxTypeTransfer,
			false,
		}, {
			"Send to FFAddr eth and bjj addresses and idx",
			&PoolL2Tx{ToBJJ: bjjAddr, ToEthAddr: FFAddr, ToIdx: Idx(400)},
			TxTypeTransfer,
			false,
		}, {
			"Send to FFAddr eth and bjj addresses",
			&PoolL2Tx{ToBJJ: bjjAddr, ToEthAddr: FFAddr, ToIdx: Idx(0)},
			TxTypeTransferToBJJ,
			false,
		}, {
			"Send to eth and bjj addresses",
			&PoolL2Tx{ToBJJ: bjjAddr, ToEthAddr: ethAddr, ToIdx: Idx(0)},
			TxType(""),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.SetType()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, tt.tx.Type)
		})
	}
}

func TestHashMaxNumBatch(t *testing.T) {
	toBJJ, err := HezStringToBJJ("hez:xut2umeShR_Lmquf3wjnbT7j_p-5T9qZ24Iewr4KUR8W", "")
	require.NoError(t, err)
	rqToBJJ, err := HezStringToBJJ("hez:JjTLlgy5ZcPIudKBNX_ejAJMT3jA-dYqC1FhHgvlyQsH", "")
	require.NoError(t, err)
	amount := big.NewInt(0)
	amount.SetString("5300000000000000000", 10)
	tx := PoolL2Tx{
		FromIdx:     439,
		ToIdx:       825429,
		ToEthAddr:   ethCommon.HexToAddress("0xf4e2b0fcbd0dc4b326d8a52b718a7bb43bdbd072"),
		ToBJJ:       *toBJJ,
		Amount:      amount,
		TokenID:     1,
		Nonce:       6,
		Fee:         226,
		RqFromIdx:   227051877307886,
		RqToIdx:     350,
		RqToEthAddr: ethCommon.HexToAddress("0x4a4547136a017c665fcedcdddca9dfd6d7dbc77f"),
		RqToBJJ:     *rqToBJJ,
		MaxNumBatch: 16385,
	}
	chainID := uint16(1)
	toSign, err := tx.HashToSign(chainID)
	assert.NoError(t, err)
	assert.Equal(t,
		"06226f6b16dc853fa8225e82e5fd675f51858e7cd6b3a951c169ac01d7125c71",
		hex.EncodeToString(toSign.Bytes()))
}
