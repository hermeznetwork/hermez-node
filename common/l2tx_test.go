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
	assert.Equal(t, "0x02fb52b5d0b9ef2626c11701bb751b2720c76d59946b9a48146ac153bb6e63bf6a", l2Tx.TxID.String())

	l2Tx = &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		TokenID: 5,
		Amount:  big.NewInt(4),
		Nonce:   1,
	}
	l2Tx, err = NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x0276114a8f666fa1ff7dbf34b4a9da577808dc501e3b2760d01fe3ef5473f5737f", l2Tx.TxID.String())

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
	assert.Equal(t, "0x025afb63126d3067f61f633d13e5a51da0551af3a4567a9af2db5321ed04214ff4", l2Tx.TxID.String())

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
	assert.Equal(t, "0x02cf390157041c3b1b59f0aaed4da464f0d0d48f1d026e46fd89c7fe1e5aed7fcf", l2Tx.TxID.String())

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
	assert.Equal(t, "0x020ec18eaae67fcd545998841a9c4be09ee3083e12db6ae5e5213a2ecaaa52d5cf", l2Tx.TxID.String())

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
	assert.Equal(t, "0x02f036223e79fac776de107f50822552cc964ee9fc4caa304613285f6976bcc940", l2Tx.TxID.String())

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
	assert.Equal(t, "0x029c8aef9ef24531e4cf84e78cbab1018ba1626a5a10afb6b7c356be1b5c28e92c", l2Tx.TxID.String())
}

func TestL2TxByteParsers(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("79000000", 10)
	l2Tx := &L2Tx{
		ToIdx:   256,
		Amount:  amount,
		FromIdx: 257,
		Fee:     201,
	}
	// Data from the compatibility test
	expected := "00000101000001002b16c9"
	encodedData, err := l2Tx.BytesDataAvailability(32)
	require.NoError(t, err)
	assert.Equal(t, expected, hex.EncodeToString(encodedData))

	decodedData, err := L2TxFromBytesDataAvailability(encodedData, 32)
	require.NoError(t, err)
	assert.Equal(t, l2Tx, decodedData)
}
