package eth

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestClientInterface(t *testing.T) {
	ethClient, err := ethclient.Dial(ethClientDialURL)
	require.Nil(t, err)
	var c ClientInterface
	client, _ := NewClient(ethClient, nil, nil, &ClientConfig{})
	c = client
	require.NotNil(t, c)
}
