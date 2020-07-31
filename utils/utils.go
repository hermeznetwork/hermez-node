package utils

import (
	"math/big"
)

// Float2fix converts a float to a fix
func Float2fix(fl int64) *big.Int {

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
func floorFix2Float(_f string) *big.Int {

	zero := big.NewInt(0)
	ten := big.NewInt(10)
	e := 0

	m := big.NewInt(0)
	m.SetString(_f, 10)

	if m.Cmp(zero) == 0 {
		return zero
	}

	s := big.NewInt(0).Rsh(m, 10)

	for s.Cmp(zero) != 0 {

		m.Div(m, ten)
		s.Rsh(m, 10)
		e++

	}

	m.Add(m, big.NewInt(int64(e<<11)))

	return m

}

// Fix2float converts a fix to a float
func Fix2float(_f string) *big.Int {

	f := big.NewInt(0)
	f.SetString(_f, 10)

	fl1 := floorFix2Float(_f)
	fi1 := Float2fix(fl1.Int64())
	fl2 := big.NewInt(int64(fl1.Int64() | 0x400))
	fi2 := Float2fix(fl2.Int64())

	m3 := big.NewInt((fl1.Int64() & 0x3FF) + 1)
	e3 := fl1.Int64() >> 11

	if m3.Cmp(big.NewInt(0x400)) == 0 {

		m3.SetInt64(0x66) // 0x400 / 10
		e3++
	}

	fl3 := m3.Add(m3, big.NewInt(e3<<11))
	fi3 := Float2fix(fl3.Int64())

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

	return res

}

// FloorFix2Float Converts a float to a fix, always rounding down
func FloorFix2Float(_f string) *big.Int {

	f := big.NewInt(0)
	f.SetString(_f, 10)

	fl1 := floorFix2Float(_f)
	fl2 := big.NewInt(int64(fl1.Int64() | 0x400))
	fi2 := Float2fix(fl2.Int64())

	if fi2.Cmp(f) < 1 {
		return fl2
	} else {
		return fl1
	}
}
