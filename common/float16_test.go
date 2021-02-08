package common

import (
	"math/big"
	"testing"

	"github.com/hermeznetwork/tracerr"
	"github.com/stretchr/testify/assert"
)

func TestConversionsFloat16(t *testing.T) {
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
		fix := test.BigInt()

		assert.Equal(t, fix.String(), testVector[test])

		bi := big.NewInt(0)
		bi.SetString(testVector[test], 10)

		fl, err := NewFloat16(bi)
		assert.NoError(t, err)

		fx2 := fl.BigInt()
		assert.Equal(t, fx2.String(), testVector[test])
	}
}

func TestFloorFix2FloatFloat16(t *testing.T) {
	testVector := map[string]Float16{
		"87999990000000000": 0x776f,
		"87950000000000001": 0x776f,
		"87950000000000000": 0x776f,
		"87949999999999999": 0x736f,
	}

	for test := range testVector {
		bi := big.NewInt(0)
		bi.SetString(test, 10)

		testFloat := NewFloat16Floor(bi)

		assert.Equal(t, testFloat, testVector[test])
	}
}

func TestConversionLossesFloat16(t *testing.T) {
	a := big.NewInt(1000)
	b, err := NewFloat16(a)
	assert.NoError(t, err)
	c := b.BigInt()
	assert.Equal(t, c, a)

	a = big.NewInt(1024)
	b, err = NewFloat16(a)
	assert.Equal(t, ErrRoundingLoss, tracerr.Unwrap(err))
	c = b.BigInt()
	assert.NotEqual(t, c, a)

	a = big.NewInt(32767)
	b, err = NewFloat16(a)
	assert.Equal(t, ErrRoundingLoss, tracerr.Unwrap(err))
	c = b.BigInt()
	assert.NotEqual(t, c, a)

	a = big.NewInt(32768)
	b, err = NewFloat16(a)
	assert.Equal(t, ErrRoundingLoss, tracerr.Unwrap(err))
	c = b.BigInt()
	assert.NotEqual(t, c, a)

	a = big.NewInt(65536000)
	b, err = NewFloat16(a)
	assert.Equal(t, ErrRoundingLoss, tracerr.Unwrap(err))
	c = b.BigInt()
	assert.NotEqual(t, c, a)
}

func BenchmarkFloat16(b *testing.B) {
	newBigInt := func(s string) *big.Int {
		bigInt, ok := new(big.Int).SetString(s, 10)
		if !ok {
			panic("Bad big int")
		}
		return bigInt
	}
	type pair struct {
		Float16 Float16
		BigInt  *big.Int
	}
	testVector := []pair{
		{0x307B, newBigInt("123000000")},
		{0x1DC6, newBigInt("454500")},
		{0xFFFF, newBigInt("10235000000000000000000000000000000")},
		{0x0000, newBigInt("0")},
		{0x0400, newBigInt("0")},
		{0x0001, newBigInt("1")},
		{0x0401, newBigInt("1")},
		{0x0800, newBigInt("0")},
		{0x0c00, newBigInt("5")},
		{0x0801, newBigInt("10")},
		{0x0c01, newBigInt("15")},
	}
	b.Run("floorFix2Float()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewFloat16Floor(testVector[i%len(testVector)].BigInt)
		}
	})
	b.Run("NewFloat16()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = NewFloat16(testVector[i%len(testVector)].BigInt)
		}
	})
	b.Run("Float16.BigInt()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testVector[i%len(testVector)].Float16.BigInt()
		}
	})
}
