package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertUSD asserts pointers to float64, and checks that they are equal
// with a tolerance of 0.01%. After that, the actual value is setted to the expected value
// in order to be able to perform further assertions using the standar assert functions.
func AssertUSD(t *testing.T, expected, actual *float64) {
	if actual == nil {
		assert.Equal(t, expected, actual)
		return
	}
	if *expected < *actual {
		assert.InEpsilon(t, *actual, *expected, 0.0001)
	} else if *expected > *actual {
		assert.InEpsilon(t, *expected, *actual, 0.0001)
	}
	*expected = *actual
}
