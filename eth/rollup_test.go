package eth

import (
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
	ks := keystore.NewKeyStore(pathKs, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(key, password)
	ks.Unlock(account, password)
	// Init eth client
	ethClient, err := ethclient.Dial(ehtClientDialURL)
	require.Nil(t, err)
	ethereumClient := NewEthereumClient(ethClient, &account, ks, nil)
	if integration != "" {
		rollupClient = NewRollupClient(ethereumClient, HERMEZROLLUP)
	}
}

func TestRollupConstants(t *testing.T) {
	if rollupClient != nil {
		_, err := rollupClient.RollupConstants()
		require.Nil(t, err)
	}
}
