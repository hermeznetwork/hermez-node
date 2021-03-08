package common

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/hermeznetwork/tracerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionsFloat40(t *testing.T) {
	testVector := map[Float40]string{
		6*0x800000000 + 123:    "123000000",
		2*0x800000000 + 4545:   "454500",
		30*0x800000000 + 10235: "10235000000000000000000000000000000",
		0x000000000:            "0",
		0x800000000:            "0",
		0x0001:                 "1",
		0x0401:                 "1025",
		0x800000000 + 1:        "10",
		0xFFFFFFFFFF:           "343597383670000000000000000000000000000000",
	}

	for test := range testVector {
		fix, err := test.BigInt()
		require.NoError(t, err)
		assert.Equal(t, fix.String(), testVector[test])

		bi, ok := new(big.Int).SetString(testVector[test], 10)
		require.True(t, ok)

		fl, err := NewFloat40(bi)
		assert.NoError(t, err)

		fx2, err := fl.BigInt()
		require.NoError(t, err)
		assert.Equal(t, fx2.String(), testVector[test])
	}
}

func TestExpectError(t *testing.T) {
	testVector := map[string]error{
		"9922334455000000000000000000000000000000":   nil,
		"9922334455000000000000000000000000000001":   ErrFloat40NotEnoughPrecission,
		"9922334454999999999999999999999999999999":   ErrFloat40NotEnoughPrecission,
		"42949672950000000000000000000000000000000":  nil,
		"99223344556573838487575":                    ErrFloat40NotEnoughPrecission,
		"992233445500000000000000000000000000000000": ErrFloat40E31,
		"343597383670000000000000000000000000000000": nil,
		"343597383680000000000000000000000000000000": ErrFloat40NotEnoughPrecission,
		"343597383690000000000000000000000000000000": ErrFloat40NotEnoughPrecission,
		"343597383700000000000000000000000000000000": ErrFloat40E31,
	}
	for test := range testVector {
		bi, ok := new(big.Int).SetString(test, 10)
		require.True(t, ok)
		_, err := NewFloat40(bi)
		assert.Equal(t, testVector[test], tracerr.Unwrap(err))
	}
}

func TestNewFloat40Floor(t *testing.T) {
	testVector := map[string][]string{
		// []int contains [Float40 value, Flot40 Floor value], when
		// Float40 value is expected to be 0, is because is expected to
		// be an error
		"9922334455000000000000000000000000000000": {
			"1040714485495", "1040714485495", "9922334455000000000000000000000000000000"},
		"9922334455000000000000000000000000000001": { // Floor [2] will be same as prev line
			"0", "1040714485495", "9922334455000000000000000000000000000000"},
		"9922334454999999999999999999999999999999": {
			"0", "1040714485494", "9922334454000000000000000000000000000000"},
		"42949672950000000000000000000000000000000": {
			"1069446856703", "1069446856703", "42949672950000000000000000000000000000000"},
		"99223344556573838487575": {
			"0", "456598933239", "99223344550000000000000"},
		"992233445500000000000000000000000000000000": {
			"0", "0", "0"}, // e>31, returns 0, err
		"343597383670000000000000000000000000000000": {
			"1099511627775", "1099511627775", "343597383670000000000000000000000000000000"},
		"343597383680000000000000000000000000000000": {
			"0", "0", "0"}, // e>31, returns 0, err
		"1157073197879933027": {
			"0", "286448638922", "1157073197800000000"},
	}
	for test := range testVector {
		bi, ok := new(big.Int).SetString(test, 10)
		require.True(t, ok)
		f40, err := NewFloat40(bi)
		if f40 == 0 {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, testVector[test][0], fmt.Sprint(uint64(f40)))

		f40, err = NewFloat40Floor(bi)
		if f40 == 0 {
			assert.Equal(t, ErrFloat40E31, tracerr.Unwrap(err))
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, testVector[test][1], fmt.Sprint(uint64(f40)))

		bi2, err := f40.BigInt()
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprint(testVector[test][2]), bi2.String())
	}
}

func BenchmarkFloat40(b *testing.B) {
	newBigInt := func(s string) *big.Int {
		bigInt, ok := new(big.Int).SetString(s, 10)
		if !ok {
			panic("Can not convert string to *big.Int")
		}
		return bigInt
	}
	type pair struct {
		Float40 Float40
		BigInt  *big.Int
	}
	testVector := []pair{
		{6*0x800000000 + 123, newBigInt("123000000")},
		{2*0x800000000 + 4545, newBigInt("454500")},
		{30*0x800000000 + 10235, newBigInt("10235000000000000000000000000000000")},
		{0x000000000, newBigInt("0")},
		{0x800000000, newBigInt("0")},
		{0x0001, newBigInt("1")},
		{0x0401, newBigInt("1025")},
		{0x800000000 + 1, newBigInt("10")},
		{0xFFFFFFFFFF, newBigInt("343597383670000000000000000000000000000000")},
	}
	b.Run("NewFloat40()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = NewFloat40(testVector[i%len(testVector)].BigInt)
		}
	})
	b.Run("Float40.BigInt()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = testVector[i%len(testVector)].Float40.BigInt()
		}
	})
}
