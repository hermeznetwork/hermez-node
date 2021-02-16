package coordinator

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddPerc(t *testing.T) {
	assert.Equal(t, "110", addPerc(big.NewInt(100), 10).String())
	assert.Equal(t, "101", addPerc(big.NewInt(100), 1).String())
	assert.Equal(t, "12", addPerc(big.NewInt(10), 20).String())
	assert.Equal(t, "1500", addPerc(big.NewInt(1000), 50).String())
}
