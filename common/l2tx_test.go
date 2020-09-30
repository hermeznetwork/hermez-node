package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewL2Tx(t *testing.T) {
	l2Tx := &L2Tx{
		FromIdx: 87654,
		ToIdx:   300,
		Amount:  big.NewInt(4),
		Nonce:   144,
	}
	l2Tx, err := NewL2Tx(l2Tx)
	assert.Nil(t, err)
	assert.Equal(t, "0x020000000156660000000090", l2Tx.TxID.String())
}
