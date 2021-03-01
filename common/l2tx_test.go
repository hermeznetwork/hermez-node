package common

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewL2Tx(t *testing.T) {
	l2Tx := &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		TokenID: 5,
		Amount:  big.NewInt(4),
		Nonce:   144,
	}
	l2Tx, err := NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x022669acda59b827d20ef5354a3eebd1dffb3972b0a6bf89d18bfd2efa0ab9f41e",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		TokenID: 5,
		Amount:  big.NewInt(4),
		Nonce:   1,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x029e7499a830f8f5eb17c07da48cf91415710f1bcbe0169d363ff91e81faf92fc2",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		TokenID: 5,
		Amount:  big.NewInt(4),
		Fee:     126,
		Nonce:   3,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x0255c70ed20e1b8935232e1b9c5884dbcc88a6e1a3454d24f2d77252eb2bb0b64e",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		TokenID: 5,
		Amount:  big.NewInt(4),
		Nonce:   1003,
		Fee:     144,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x0206b372f967061d1148bbcff679de38120e075141a80a07326d0f514c2efc6ca9",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 1,
		ToIdx:   1,
		TokenID: 1,
		Amount:  big.NewInt(1),
		Nonce:   1,
		Fee:     1,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x0236f7ea5bccf78ba60baf56c058d235a844f9b09259fd0efa4f5f72a7d4a26618",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 999,
		ToIdx:   999,
		TokenID: 999,
		Amount:  big.NewInt(999),
		Nonce:   999,
		Fee:     255,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x02ac122f5b709ce190129fecbbe35bfd30c70e6433dbd85a8eb743d110906a1dc1",
		l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 4444,
		ToIdx:   300,
		TokenID: 0,
		Amount:  big.NewInt(3400000000),
		Nonce:   2,
		Fee:     25,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x02c674951a81881b7bc50db3b9e5efd97ac88550c7426ac548720e5057cfba515a",
		l2Tx.TxID.String())
}

func TestL2TxByteParsers(t *testing.T) {
	// test vectors values generated from javascript implementation
	amount, ok := new(big.Int).SetString("343597383670000000000000000000000000000000", 10)
	require.True(t, ok)
	l2Tx := &L2Tx{
		ToIdx:   (1 << 16) - 1,
		FromIdx: (1 << 16) - 1,
		Amount:  amount,
		Fee:     (1 << 8) - 1,
	}
	expected := "ffffffffffffffffffff"
	encodedData, err := l2Tx.BytesDataAvailability(16)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err := L2TxFromBytesDataAvailability(encodedData, 16)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)

	l2Tx = &L2Tx{
		ToIdx:   (1 << 32) - 1,
		FromIdx: (1 << 32) - 1,
		Amount:  amount,
		Fee:     (1 << 8) - 1,
	}
	expected = "ffffffffffffffffffffffffffff"
	encodedData, err = l2Tx.BytesDataAvailability(32)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err = L2TxFromBytesDataAvailability(encodedData, 32)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)

	l2Tx = &L2Tx{
		ToIdx:   0,
		FromIdx: 0,
		Amount:  big.NewInt(0),
		Fee:     0,
	}
	expected = "0000000000000000000000000000"
	encodedData, err = l2Tx.BytesDataAvailability(32)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err = L2TxFromBytesDataAvailability(encodedData, 32)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)

	l2Tx = &L2Tx{
		ToIdx:   0,
		FromIdx: 1061,
		Amount:  big.NewInt(420000000000),
		Fee:     127,
	}
	expected = "000004250000000010fa56ea007f"
	encodedData, err = l2Tx.BytesDataAvailability(32)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err = L2TxFromBytesDataAvailability(encodedData, 32)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)

	l2Tx = &L2Tx{
		ToIdx:   256,
		FromIdx: 257,
		Amount:  big.NewInt(79000000),
		Fee:     201,
	}
	expected = "00000101000001000004b571c0c9"
	encodedData, err = l2Tx.BytesDataAvailability(32)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err = L2TxFromBytesDataAvailability(encodedData, 32)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)
}
