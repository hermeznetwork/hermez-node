package eth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientInterface(t *testing.T) {
	var c ClientInterface
	client := NewClient(nil, nil, nil, nil)
	c = client
	require.NotNil(t, c)
}
