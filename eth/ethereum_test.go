package eth

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEthERC20(t *testing.T) {
	ethClient, err := ethclient.Dial(ethClientDialURL)
	require.Nil(t, err)
	client, err := NewEthereumClient(ethClient, auxAccount, ks, nil)
	require.Nil(t, err)

	consts, err := client.EthERC20Consts(tokenHEZAddressConst)
	require.Nil(t, err)
	assert.Equal(t, "Hermez Network Token", consts.Name)
	assert.Equal(t, "HEZ", consts.Symbol)
	assert.Equal(t, uint64(18), consts.Decimals)
}
