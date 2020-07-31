package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversions(t *testing.T) {

	testVector := map[int64]string{
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

		fl := Fix2float(testVector[test])
		fx2 := Float2fix(fl.Int64())

		assert.Equal(t, fx2.String(), testVector[test])

	}

}

func TestFloorFix2Float(t *testing.T) {

	testVector := map[string]int64{
		"87999990000000000": 0x776f,
		"87950000000000001": 0x776f,
		"87950000000000000": 0x776f,
		"87949999999999999": 0x736f,
	}

	for test := range testVector {

		testFloat := FloorFix2Float(test)

		assert.Equal(t, testFloat.Int64(), testVector[test])

	}

}
