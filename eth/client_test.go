package eth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientInterface(t *testing.T) {
	var c ClientInterface
	client, _ := NewClient(nil, nil, nil, &ClientConfig{})
	c = client
	require.NotNil(t, c)
}
