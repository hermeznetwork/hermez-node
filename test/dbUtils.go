package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
