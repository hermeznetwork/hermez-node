package eth

import (
	"io/ioutil"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var rollupClient *RollupClient

func TestNewRollupClient(t *testing.T) {
	key, err := crypto.HexToECDSA(governancePrivateKey)
	require.Nil(t, err)
	dir, err := ioutil.TempDir("", "tmpks")
	require.Nil(t, err)
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(key, password)
	require.Nil(t, err)
	err = ks.Unlock(account, password)
	require.Nil(t, err)
	// Init eth client
	ethClient, err := ethclient.Dial(ehtClientDialURL)
	require.Nil(t, err)
	ethereumClient := NewEthereumClient(ethClient, &account, ks, nil)
	if integration != "" {
		rollupClient = NewRollupClient(ethereumClient, hermezRollupAddressConst)
	}
}

func TestRollupConstants(t *testing.T) {
	if rollupClient != nil {
		_, err := rollupClient.RollupConstants()
		require.Nil(t, err)
	}
}
