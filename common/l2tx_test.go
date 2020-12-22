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
		Amount:  big.NewInt(4),
		Nonce:   144,
	}
	l2Tx, err := NewL2Tx(l2Tx)
	assert.NoError(t, err)
	assert.Equal(t, "0x020000000156660000000090", l2Tx.TxID.String())
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
