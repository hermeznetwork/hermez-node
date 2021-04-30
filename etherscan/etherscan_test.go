package etherscan

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var etherscanURL = "http://localhost:3000/api"
var apiKey = "FFFFFFFFFFFFFFFFFFF"

var etherScanMockService *MockEtherscanClient
var etherScanService *Service

func TestMain(m *testing.M) {
	exitVal := 0
	_etherscanURL := os.Getenv("ETHERSCAN_API_URL")
	if _etherscanURL != "" {
		etherscanURL = _etherscanURL
	}
	_apiKey := os.Getenv("ETHERSCAN_API_KEY")
	if _apiKey != "" {
		apiKey = _apiKey
	}
	var err error
	etherScanService, err = NewEtherscanService(etherscanURL, apiKey)
	if err != nil {
		panic(err)
	}
	exitVal = m.Run()
	os.Exit(exitVal)
}

func TestEtherscanApiServer(t *testing.T) {
	t.Run("testGetGasPrice", testGetGasPrice)
}

func testGetGasPrice(t *testing.T) {
	gasPrice, err := etherScanMockService.GetGasPrice(context.Background())
	require.NoError(t, err)
	assert.NotEqual(t, 0, gasPrice.LastBlock)
	assert.NotEqual(t, 90, gasPrice.SafeGasPrice)
	assert.NotEqual(t, 100, gasPrice.ProposeGasPrice)
	assert.NotEqual(t, 110, gasPrice.FastGasPrice)
}
