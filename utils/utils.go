package utils

import (
	"errors"
	"math/big"
)

var (
	// ErrRoundingLoss is used when converted big.Int to Float16 causes rounding loss
	ErrRoundingLoss = errors.New("input value causes rounding loss")
)

// Float16 represents a float in a 16 bit format
type Float16 uint16

// BigInt converts the Float16 to a big.Int integer
func (fl16 *Float16) BigInt() *big.Int {

	fl := int64(*fl16)

	m := big.NewInt(fl & 0x3FF)
	e := big.NewInt(fl >> 11)
	e5 := (fl >> 10) & 0x01

	exp := big.NewInt(0).Exp(big.NewInt(10), e, nil)
	res := m.Mul(m, exp)

	if e5 != 0 && e.Cmp(big.NewInt(0)) != 0 {

		res.Add(res, exp.Div(exp, big.NewInt(2)))

	}

	return res

}

// floorFix2Float converts a fix to a float, always rounding down
func floorFix2Float(_f *big.Int) Float16 {

	zero := big.NewInt(0)
	ten := big.NewInt(10)
	e := int64(0)

	m := big.NewInt(0)
	m.Set(_f)

	if m.Cmp(zero) == 0 {
		return 0
	}

	s := big.NewInt(0).Rsh(m, 10)

	for s.Cmp(zero) != 0 {

		m.Div(m, ten)
		s.Rsh(m, 10)
		e++

	}

	return Float16(m.Int64() | e<<11)

}

// NewFloat16 encodes a big.Int integer as a Float16, returning error in case
// of loss during the encoding.
func NewFloat16(f *big.Int) (Float16, error) {

	fl1 := floorFix2Float(f)
	fi1 := fl1.BigInt()
	fl2 := fl1 | 0x400
	fi2 := fl2.BigInt()

	m3 := (fl1 & 0x3FF) + 1
	e3 := fl1 >> 11

	if m3&0x400 == 0 {
		m3 = 0x66
		e3++
	}

	fl3 := m3 + e3<<11
	fi3 := fl3.BigInt()

	res := fl1

	d := big.NewInt(0).Abs(fi1.Sub(fi1, f))
	d2 := big.NewInt(0).Abs(fi2.Sub(fi2, f))

	if d.Cmp(d2) == 1 {
		res = fl2
		d = d2
	}

	d3 := big.NewInt(0).Abs(fi3.Sub(fi3, f))

	if d.Cmp(d3) == 1 {

		res = fl3
	}

	// Do rounding check

	if res.BigInt().Cmp(f) == 0 {

		return res, nil
	}

	return res, ErrRoundingLoss

}

// NewFloat16Floor encodes a big.Int integer as a Float16, rounding down in
// case of loss during the encoding.
func NewFloat16Floor(f *big.Int) Float16 {

	fl1 := floorFix2Float(f)
	fl2 := fl1 | 0x400
	fi2 := fl2.BigInt()

	if fi2.Cmp(f) < 1 {
		return fl2
	}
	return fl1

}
