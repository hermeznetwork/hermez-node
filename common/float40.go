// Package common float40.go provides methods to work with Hermez custom half
// float precision, 40 bits, codification internally called Float40 has been
// adopted to encode large integers. This is done in order to save bits when L2
// transactions are published.
//nolint:gomnd
package common

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/hermeznetwork/tracerr"
)

const (
	// maxFloat40Value is the maximum value that the Float40 can have
	// (40 bits: maxFloat40Value=2**40-1)
	maxFloat40Value = 0xffffffffff
	// Float40BytesLength defines the length of the Float40 values
	// represented as byte arrays
	Float40BytesLength = 5
)

var (
	// ErrFloat40Overflow is used when a given Float40 overflows the
	// maximum capacity of the Float40 (2**40-1)
	ErrFloat40Overflow = errors.New("Float40 overflow, max value: 2**40 -1")
	// ErrFloat40E31 is used when the e > 31 when trying to convert a
	// *big.Int to Float40
	ErrFloat40E31 = errors.New("Float40 error, e > 31")
	// ErrFloat40NotEnoughPrecission is used when the given *big.Int can
	// not be represented as Float40 due not enough precission
	ErrFloat40NotEnoughPrecission = errors.New("Float40 error, not enough precission")

	thres = big.NewInt(0x08_00_00_00_00)
)

// Float40 represents a float in a 64 bit format
type Float40 uint64

// Bytes return a byte array of length 5 with the Float40 value encoded in
// BigEndian
func (f40 Float40) Bytes() ([]byte, error) {
	if f40 > maxFloat40Value {
		return []byte{}, tracerr.Wrap(ErrFloat40Overflow)
	}

	var f40Bytes [8]byte
	binary.BigEndian.PutUint64(f40Bytes[:], uint64(f40))
	var b [5]byte
	copy(b[:], f40Bytes[3:])
	return b[:], nil
}

// Float40FromBytes returns a Float40 from a byte array of 5 bytes in Bigendian
// representation.
func Float40FromBytes(b []byte) Float40 {
	var f40Bytes [8]byte
	copy(f40Bytes[3:], b[:])
	f40 := binary.BigEndian.Uint64(f40Bytes[:])
	return Float40(f40)
}

// BigInt converts the Float40 to a *big.Int v, where v = m * 10^e, being:
// [    e   |    m    ]
// [ 5 bits | 35 bits ]
func (f40 Float40) BigInt() (*big.Int, error) {
	// take the 5 used bytes (FF * 5)
	var f40Uint64 uint64 = uint64(f40) & 0x00_00_00_FF_FF_FF_FF_FF
	f40Bytes, err := f40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	e := f40Bytes[0] & 0xF8 >> 3      // take first 5 bits
	m := f40Uint64 & 0x07_FF_FF_FF_FF // take the others 35 bits

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(e)), nil)
	r := new(big.Int).Mul(big.NewInt(int64(m)), exp)
	return r, nil
}

// NewFloat40 encodes a *big.Int integer as a Float40, returning error in case
// of loss during the encoding.
func NewFloat40(f *big.Int) (Float40, error) {
	m := f
	e := big.NewInt(0)
	zero := big.NewInt(0)
	ten := big.NewInt(10)
	for new(big.Int).Mod(m, ten).Cmp(zero) == 0 && m.Cmp(thres) >= 0 {
		m = new(big.Int).Div(m, ten)
		e = new(big.Int).Add(e, big.NewInt(1))
	}
	if e.Int64() > 31 {
		return 0, tracerr.Wrap(ErrFloat40E31)
	}
	if m.Cmp(thres) >= 0 {
		return 0, tracerr.Wrap(ErrFloat40NotEnoughPrecission)
	}
	r := new(big.Int).Add(m,
		new(big.Int).Mul(e, thres))
	return Float40(r.Uint64()), nil
}

// NewFloat40Floor encodes a *big.Int integer as a Float40, rounding down in
// case of loss during the encoding. It returns an error in case that the number
// is too big (e>31). Warning: this method should not be used inside the
// hermez-node, it's a helper for external usage to generate valid Float40
// values.
func NewFloat40Floor(f *big.Int) (Float40, error) {
	m := f
	e := big.NewInt(0)
	// zero := big.NewInt(0)
	ten := big.NewInt(10)
	for m.Cmp(thres) >= 0 {
		m = new(big.Int).Div(m, ten)
		e = new(big.Int).Add(e, big.NewInt(1))
	}
	if e.Int64() > 31 {
		return 0, tracerr.Wrap(ErrFloat40E31)
	}

	r := new(big.Int).Add(m,
		new(big.Int).Mul(e, thres))

	return Float40(r.Uint64()), nil
}
