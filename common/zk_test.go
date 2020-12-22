package common

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZKInputs(t *testing.T) {
	zki := NewZKInputs(100, 16, 512, 24, 32, big.NewInt(1))
	_, err := json.Marshal(zki)
	require.NoError(t, err)
	// fmt.Println(string(s))
}
