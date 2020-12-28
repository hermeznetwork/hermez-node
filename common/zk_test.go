package common

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZKInputs(t *testing.T) {
	chainID := uint16(0)
	zki := NewZKInputs(chainID, 100, 24, 512, 32, big.NewInt(1))
	_, err := json.Marshal(zki)
	require.NoError(t, err)
	// fmt.Println(string(s))
}
