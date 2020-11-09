package common

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZKInputs(t *testing.T) {
	zki := NewZKInputs(100, 16, 512, 24, 32)
	_, err := json.Marshal(zki)
	require.Nil(t, err)
	// fmt.Println(string(s))
}
