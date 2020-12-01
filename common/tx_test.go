package common

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureConstant(t *testing.T) {
	signatureConstant := uint32(3322668559)
	var signatureConstantBytes [4]byte
	binary.BigEndian.PutUint32(signatureConstantBytes[:], signatureConstant)
	assert.Equal(t, SignatureConstantBytes, signatureConstantBytes[:])
	assert.Equal(t, "c60be60f", hex.EncodeToString(SignatureConstantBytes))
}

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

func TestTxIDMarshalers(t *testing.T) {
	h := []byte("0x00000000000001e240004700")
	var txid TxID
	err := txid.UnmarshalText(h)
	assert.Nil(t, err)
	assert.Equal(t, h, []byte(txid.String()))

	h2, err := txid.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t, h, h2)

	var txid2 TxID
	err = txid2.UnmarshalText(h2)
	assert.Nil(t, err)
	assert.Equal(t, h2, []byte(txid2.String()))
	assert.Equal(t, h, h2)
}
