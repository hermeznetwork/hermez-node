package common

import (
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxIDScannerValue(t *testing.T) {
	txid0 := &TxID{}
	txid1 := &TxID{}
	txid0B := [12]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	txid1B := [12]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	copy(txid0[:], txid0B[:])
	copy(txid1[:], txid1B[:])

	var value driver.Valuer
	var scan sql.Scanner
	value = txid0
	scan = txid1
	fromDB, err := value.Value()
	assert.NoError(t, err)
	assert.NoError(t, scan.Scan(fromDB))
	assert.Equal(t, value, scan)
}
