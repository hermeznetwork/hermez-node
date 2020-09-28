package eth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var rollupClient *RollupClient

func TestRollupConstants(t *testing.T) {
	if rollupClient != nil {
		_, err := rollupClient.RollupConstants()
		require.Nil(t, err)
	}
}
