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
	client := NewEthereumClient(ethClient, accountAux, ks, nil)

	consts, err := client.EthERC20Consts(tokenERC20AddressConst)
	require.Nil(t, err)
	assert.Equal(t, "ERC20_0", consts.Name)
	assert.Equal(t, "20_0", consts.Symbol)
	assert.Equal(t, uint64(18), consts.Decimals)
}
