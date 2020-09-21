package eth

import (
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wdelayerClient *WDelayerClient
var wdelayerClientKep *WDelayerClient

var wdelayerAddressStr = "0x1A1FEe7EeD918BD762173e4dc5EfDB8a78C924A8"
var hermezGovernanceDAOAddressStr = "0x84Fae3d3Cba24A97817b2a18c2421d462dbBCe9f"
var hermezGovernanceDAOAddressPK = "2a8aede924268f84156a00761de73998dac7bf703408754b776ff3f873bcec60"
var whiteHackGroupAddressPK = "8b24fd94f1ce869d81a34b95351e7f97b2cd88a891d5c00abc33d0ec9501902e"
var hermezKeeperAddressPK = "7f307c41137d1ed409f0a7b028f6c7596f12734b1d289b58099b99d60a96efff"
var hermezGovernanceDAOAddressConst = common.HexToAddress(hermezGovernanceDAOAddressStr)
var whiteHackGroupAddressStr = "0xfa3BdC8709226Da0dA13A4d904c8b66f16c3c8BA"
var whiteHackGroupAddressConst = common.HexToAddress(whiteHackGroupAddressStr)
var hermezKeeperAddressStr = "0xFbC51a9582D031f2ceaaD3959256596C5D3a5468"
var hermezKeeperAddressConst = common.HexToAddress(hermezKeeperAddressStr)

var initWithdrawalDelay = big.NewInt(60)
var newWithdrawalDelay = big.NewInt(79)

func TestNewWDelayer(t *testing.T) {
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
	wdelayerAddress := common.HexToAddress(wdelayerAddressStr)
	if integration != "" {
		wdelayerClient = NewWDelayerClient(ethereumClient, wdelayerAddress)
	}
}

func TestWDelayerGetHermezGovernanceDAOAddress(t *testing.T) {
	if wdelayerClient != nil {
		governanceAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
		require.Nil(t, err)
		assert.Equal(t, &hermezGovernanceDAOAddressConst, governanceAddress)
	}
}

func TestWDelayerSetHermezGovernanceDAOAddress(t *testing.T) {
	key, err := crypto.HexToECDSA(hermezGovernanceDAOAddressPK)
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
	wdelayerAddressGov := common.HexToAddress(wdelayerAddressStr)
	if integration != "" {
		wdelayerClientGov := NewWDelayerClient(ethereumClient, wdelayerAddressGov)
		_, err := wdelayerClientGov.WDelayerSetHermezGovernanceDAOAddress(hermezKeeperAddressConst)
		require.Nil(t, err)
		keeperAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
		require.Nil(t, err)
		assert.Equal(t, &hermezKeeperAddressConst, keeperAddress)
	}
}

func TestWDelayerGetHermezKeeperAddress(t *testing.T) {
	if wdelayerClient != nil {
		keeperAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
		require.Nil(t, err)
		assert.Equal(t, &hermezKeeperAddressConst, keeperAddress)
	}
}

func TestWDelayerSetHermezKeeperAddress(t *testing.T) {
	key, err := crypto.HexToECDSA(hermezKeeperAddressPK)
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
	wdelayerAddressKep := common.HexToAddress(wdelayerAddressStr)
	if integration != "" {
		wdelayerClientKep = NewWDelayerClient(ethereumClient, wdelayerAddressKep)
		_, err := wdelayerClientKep.WDelayerSetHermezKeeperAddress(whiteHackGroupAddressConst)
		require.Nil(t, err)
		whiteHackGroupAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
		require.Nil(t, err)
		assert.Equal(t, &whiteHackGroupAddressConst, whiteHackGroupAddress)
	}
}

func TestWDelayerGetWhiteHackGroupAddress(t *testing.T) {
	if wdelayerClient != nil {
		whiteHackGroupAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
		require.Nil(t, err)
		assert.Equal(t, &whiteHackGroupAddressConst, whiteHackGroupAddress)
	}
}

func TestWDelayerSetWhiteHackGroupAddress(t *testing.T) {
	key, err := crypto.HexToECDSA(whiteHackGroupAddressPK)
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
	wdelayerAddressWhite := common.HexToAddress(wdelayerAddressStr)
	if integration != "" {
		wdelayerClientWhite := NewWDelayerClient(ethereumClient, wdelayerAddressWhite)
		_, err := wdelayerClientWhite.WDelayerSetWhiteHackGroupAddress(governanceAddressConst)
		require.Nil(t, err)
		governanceAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
		require.Nil(t, err)
		assert.Equal(t, &governanceAddressConst, governanceAddress)
		_, err = wdelayerClientWhite.WDelayerSetHermezKeeperAddress(hermezKeeperAddressConst)
		require.Nil(t, err)
	}
}

func TestWDelayerIsEmergencyMode(t *testing.T) {
	if wdelayerClient != nil {
		emergencyMode, err := wdelayerClient.WDelayerIsEmergencyMode()
		require.Nil(t, err)
		assert.Equal(t, false, emergencyMode)
	}
}

func TestWDelayerGetWithdrawalDelay(t *testing.T) {
	if wdelayerClient != nil {
		withdrawalDelay, err := wdelayerClient.WDelayerGetWithdrawalDelay()
		require.Nil(t, err)
		assert.Equal(t, initWithdrawalDelay, withdrawalDelay)
	}
}

func TestWDelayerEnableEmergencyMode(t *testing.T) {
	if wdelayerClientKep != nil {
		_, err := wdelayerClientKep.WDelayerEnableEmergencyMode()
		require.Nil(t, err)
		emergencyMode, err := wdelayerClient.WDelayerIsEmergencyMode()
		require.Nil(t, err)
		assert.Equal(t, true, emergencyMode)
	}
}

func TestWDelayerGetEmergencyModeStartingTime(t *testing.T) {
	if wdelayerClient != nil {
		emergencyModeStartingTime, err := wdelayerClient.WDelayerGetEmergencyModeStartingTime()
		require.Nil(t, err)
		assert.True(t, emergencyModeStartingTime.Cmp(big.NewInt(0)) == 1)
	}
}

func TestWDelayerChangeWithdrawalDelay(t *testing.T) {
	if wdelayerClientKep != nil {
		_, err := wdelayerClientKep.WDelayerChangeWithdrawalDelay(newWithdrawalDelay.Uint64())
		require.Nil(t, err)
		withdrawalDelay, err := wdelayerClient.WDelayerGetWithdrawalDelay()
		require.Nil(t, err)
		assert.Equal(t, newWithdrawalDelay, withdrawalDelay)
	}
}

/* func TestWDelayerDeposit(t *testing.T) {
	if wdelayerClient != nil {

	}
}

func TestWDelayerDepositInfo(t *testing.T) {
	if wdelayerClient != nil {

	}
}

func TestWDelayerWithdrawal(t *testing.T) {
	if wdelayerClient != nil {

	}
}

func TestWDelayerEscapeHatchWithdrawal(t *testing.T) {
	if wdelayerClient != nil {

	}
} */
