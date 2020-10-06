package eth

import (
	"os"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEthERC20(t *testing.T) {
	address := ethCommon.HexToAddress("0x44021007485550008e0f9f1f7b506c7d970ad8ce")
	ethClient, err := ethclient.Dial(os.Getenv("ETHCLIENT_DIAL_URL"))
	require.Nil(t, err)
	client := NewEthereumClient(ethClient, accountAux, ks, nil)

	consts, err := client.EthERC20Consts(address)
	require.Nil(t, err)
	assert.Equal(t, "Golem Network Token", consts.Name)
	assert.Equal(t, "GNT", consts.Symbol)
	assert.Equal(t, uint64(18), consts.Decimals)
}
