package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdx(t *testing.T) {
	i := Idx(100)
	assert.Equal(t, big.NewInt(100), i.BigInt())

	i = Idx(uint32(4294967295))
	assert.Equal(t, "4294967295", i.BigInt().String())

	b := big.NewInt(4294967296)
	i, err := IdxFromBigInt(b)
	assert.NotNil(t, err)
	assert.Equal(t, ErrNumOverflow, err)
	assert.Equal(t, Idx(0), i)

}
