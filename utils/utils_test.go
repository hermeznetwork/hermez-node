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

		fix := Float2fix(test)

		assert.Equal(t, fix.String(), testVector[test])

		bi := big.NewInt(0)
		bi.SetString(testVector[test], 10)

		fl := Fix2float(bi)
		fx2 := Float2fix(fl)

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

/*
func TestConversions2(t *testing.T) {
	assert.Equal(t, "10", Float2fix(10).String())
	assert.Equal(t, "1000", Float2fix(1000).String())
	assert.Equal(t, "65535", Float2fix(65535).String())
	assert.Equal(t, "65536", Float2fix(65536).String())
	assert.Equal(t, "100000", Float2fix(100000).String()) // should this return an error?
	assert.Equal(t, "10000000", Float2fix(10000000).String())

	assert.Equal(t, "10", floorFix2Float("10").String())
	assert.Equal(t, "10", floorFix2Float("10.004").String())
	assert.Equal(t, "10000", floorFix2Float("10000").String())

	assert.Equal(t, "10", Fix2float("10").String())
	assert.Equal(t, "10", Fix2float("10.004").String())
	assert.Equal(t, "32767", Fix2float("32767").String())
	assert.Equal(t, "32768", Fix2float("32768").String())
	assert.Equal(t, "65535", Fix2float("65535").String())
	assert.Equal(t, "100000", Fix2float("100000").String())

	assert.Equal(t, "10", FloorFix2Float("10").String())
	assert.Equal(t, "10", FloorFix2Float("10.04").String())
	assert.Equal(t, "1000", FloorFix2Float("1000").String())
}
*/
