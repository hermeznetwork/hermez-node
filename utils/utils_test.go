package utils

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversions(t *testing.T) {

	testVector := map[Float16]string{
		0x307B: "123000000",
		0x1DC6: "454500",
		0xFFFF: "10235000000000000000000000000000000",
		0x0000: "0",
		0x0400: "0",
		0x0001: "1",
		0x0401: "1",
		0x0800: "0",
		0x0c00: "5",
		0x0801: "10",
		0x0c01: "15",
	}

	for test := range testVector {

		fix := Float2Fix(test)

		assert.Equal(t, fix.String(), testVector[test])

		bi := big.NewInt(0)
		bi.SetString(testVector[test], 10)

		fl, err := Fix2Float(bi)
		assert.Equal(t, nil, err)

		fx2 := Float2Fix(fl)
		assert.Equal(t, fx2.String(), testVector[test])

	}

}

func TestFloorFix2Float(t *testing.T) {

	testVector := map[string]Float16{
		"87999990000000000": 0x776f,
		"87950000000000001": 0x776f,
		"87950000000000000": 0x776f,
		"87949999999999999": 0x736f,
	}

	for test := range testVector {

		bi := big.NewInt(0)
		bi.SetString(test, 10)

		testFloat := FloorFix2Float(bi)

		assert.Equal(t, testFloat, testVector[test])

	}

}

func TestConversionLosses(t *testing.T) {
	a := big.NewInt(1000)
	b, err := Fix2Float(a)
	assert.Equal(t, nil, err)
	c := Float2Fix(b)
	assert.Equal(t, c, a)

	a = big.NewInt(1024)
	b, err = Fix2Float(a)
	assert.Equal(t, ErrRoundingLoss, err)
	c = Float2Fix(b)
	assert.NotEqual(t, c, a)

	a = big.NewInt(32767)
	b, err = Fix2Float(a)
	assert.Equal(t, ErrRoundingLoss, err)
	c = Float2Fix(b)
	assert.NotEqual(t, c, a)

	a = big.NewInt(32768)
	b, err = Fix2Float(a)
	assert.Equal(t, ErrRoundingLoss, err)
	c = Float2Fix(b)
	assert.NotEqual(t, c, a)

	a = big.NewInt(65536000)
	b, err = Fix2Float(a)
	assert.Equal(t, ErrRoundingLoss, err)
	c = Float2Fix(b)
	assert.NotEqual(t, c, a)

}
