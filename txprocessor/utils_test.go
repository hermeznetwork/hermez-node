package txprocessor

import (
	"math/big"
	"testing"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
)

func TestBJJCompressedTo256BigInt(t *testing.T) {
	var pkComp babyjub.PublicKeyComp
	r := BJJCompressedTo256BigInts(pkComp)
	zero := big.NewInt(0)
	for i := 0; i < 256; i++ {
		assert.Equal(t, zero, r[i])
	}

	pkComp[0] = 3
	r = BJJCompressedTo256BigInts(pkComp)
	one := big.NewInt(1)
	for i := 0; i < 256; i++ {
		if i != 0 && i != 1 {
			assert.Equal(t, zero, r[i])
		} else {
			assert.Equal(t, one, r[i])
		}
	}

	pkComp[31] = 4
	r = BJJCompressedTo256BigInts(pkComp)
	for i := 0; i < 256; i++ {
		if i != 0 && i != 1 && i != 250 {
			assert.Equal(t, zero, r[i])
		} else {
			assert.Equal(t, one, r[i])
		}
	}
}
