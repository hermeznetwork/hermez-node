package eth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wdelayerClient *WDelayerClient

// var wdelayerClientKep *WDelayerClient

var initWithdrawalDelay = big.NewInt(60)
var newWithdrawalDelay = big.NewInt(79)

func TestWDelayerGetHermezGovernanceDAOAddress(t *testing.T) {
	governanceAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
	require.Nil(t, err)
	assert.Equal(t, &hermezGovernanceDAOAddressConst, governanceAddress)
}

func TestWDelayerSetHermezGovernanceDAOAddress(t *testing.T) {
	wdelayerClientGov := NewWDelayerClient(ethereumClientGovDAO, wdelayerAddressConst)
	_, err := wdelayerClientGov.WDelayerSetHermezGovernanceDAOAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	wdelayerClientAux := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	_, err = wdelayerClientAux.WDelayerSetHermezGovernanceDAOAddress(hermezGovernanceDAOAddressConst)
	require.Nil(t, err)
}

func TestWDelayerGetHermezKeeperAddress(t *testing.T) {
	keeperAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
	require.Nil(t, err)
	assert.Equal(t, &hermezKeeperAddressConst, keeperAddress)
}

func TestWDelayerSetHermezKeeperAddress(t *testing.T) {
	wdelayerClientKep := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	_, err := wdelayerClientKep.WDelayerSetHermezKeeperAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	wdelayerClientAux := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	_, err = wdelayerClientAux.WDelayerSetHermezKeeperAddress(hermezKeeperAddressConst)
	require.Nil(t, err)
}

func TestWDelayerGetWhiteHackGroupAddress(t *testing.T) {
	whiteHackGroupAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
	require.Nil(t, err)
	assert.Equal(t, &whiteHackGroupAddressConst, whiteHackGroupAddress)
}

func TestWDelayerSetWhiteHackGroupAddress(t *testing.T) {
	wdelayerClientWhite := NewWDelayerClient(ethereumClientWhite, wdelayerAddressConst)
	_, err := wdelayerClientWhite.WDelayerSetWhiteHackGroupAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	wdelayerClientAux := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	_, err = wdelayerClientAux.WDelayerSetWhiteHackGroupAddress(whiteHackGroupAddressConst)
	require.Nil(t, err)
}

func TestWDelayerIsEmergencyMode(t *testing.T) {
	emergencyMode, err := wdelayerClient.WDelayerIsEmergencyMode()
	require.Nil(t, err)
	assert.Equal(t, false, emergencyMode)
}

func TestWDelayerGetWithdrawalDelay(t *testing.T) {
	withdrawalDelay, err := wdelayerClient.WDelayerGetWithdrawalDelay()
	require.Nil(t, err)
	assert.Equal(t, initWithdrawalDelay, withdrawalDelay)
}

func TestWDelayerEnableEmergencyMode(t *testing.T) {
	wdelayerClientKep := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	_, err := wdelayerClientKep.WDelayerEnableEmergencyMode()
	require.Nil(t, err)
	emergencyMode, err := wdelayerClient.WDelayerIsEmergencyMode()
	require.Nil(t, err)
	assert.Equal(t, true, emergencyMode)
}

func TestWDelayerChangeWithdrawalDelay(t *testing.T) {
	wdelayerClientKep := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	_, err := wdelayerClientKep.WDelayerChangeWithdrawalDelay(newWithdrawalDelay.Uint64())
	require.Nil(t, err)
	withdrawalDelay, err := wdelayerClient.WDelayerGetWithdrawalDelay()
	require.Nil(t, err)
	assert.Equal(t, newWithdrawalDelay, withdrawalDelay)
}

func TestWDelayerGetEmergencyModeStartingTime(t *testing.T) {
	emergencyModeStartingTime, err := wdelayerClient.WDelayerGetEmergencyModeStartingTime()
	require.Nil(t, err)
	// `emergencyModeStartingTime` is initialized to 0 in the smart
	// contract construction.  Since we called WDelayerEnableEmergencyMode
	// previously, `emergencyModeStartingTime` is set to the time when the
	// call was made, so it's > 0.
	assert.True(t, emergencyModeStartingTime.Cmp(big.NewInt(0)) == 1)
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
