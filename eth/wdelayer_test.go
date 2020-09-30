package eth

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wdelayerClient *WDelayerClient

// var wdelayerClientKep *WDelayerClient

var initWithdrawalDelay = big.NewInt(60)
var newWithdrawalDelay = big.NewInt(79)
var maxEmergencyModeTime = time.Hour * 24 * 7 * 26

func TestWDelayerGetHermezGovernanceDAOAddress(t *testing.T) {
	governanceAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
	require.Nil(t, err)
	assert.Equal(t, &hermezGovernanceDAOAddressConst, governanceAddress)
}

func TestWDelayerSetHermezGovernanceDAOAddress(t *testing.T) {
	wdelayerClientGov, err := NewWDelayerClient(ethereumClientGovDAO, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientGov.WDelayerSetHermezGovernanceDAOAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetHermezGovernanceDAOAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, auxAddressConst, wdelayerEvents.NewHermezGovernanceDAOAddress[0].NewHermezGovernanceDAOAddress)
	wdelayerClientAux, err := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientAux.WDelayerSetHermezGovernanceDAOAddress(hermezGovernanceDAOAddressConst)
	require.Nil(t, err)
}

func TestWDelayerGetHermezKeeperAddress(t *testing.T) {
	keeperAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
	require.Nil(t, err)
	assert.Equal(t, &hermezKeeperAddressConst, keeperAddress)
}

func TestWDelayerSetHermezKeeperAddress(t *testing.T) {
	wdelayerClientKep, err := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientKep.WDelayerSetHermezKeeperAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetHermezKeeperAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, auxAddressConst, wdelayerEvents.NewHermezKeeperAddress[0].NewHermezKeeperAddress)
	wdelayerClientAux, err := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientAux.WDelayerSetHermezKeeperAddress(hermezKeeperAddressConst)
	require.Nil(t, err)
}

func TestWDelayerGetWhiteHackGroupAddress(t *testing.T) {
	whiteHackGroupAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
	require.Nil(t, err)
	assert.Equal(t, &whiteHackGroupAddressConst, whiteHackGroupAddress)
}

func TestWDelayerSetWhiteHackGroupAddress(t *testing.T) {
	wdelayerClientWhite, err := NewWDelayerClient(ethereumClientWhite, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientWhite.WDelayerSetWhiteHackGroupAddress(auxAddressConst)
	require.Nil(t, err)
	auxAddress, err := wdelayerClient.WDelayerGetWhiteHackGroupAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, auxAddressConst, wdelayerEvents.NewWhiteHackGroupAddress[0].NewWhiteHackGroupAddress)
	wdelayerClientAux, err := NewWDelayerClient(ethereumClientAux, wdelayerAddressConst)
	require.Nil(t, err)
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

func TestWDelayerChangeWithdrawalDelay(t *testing.T) {
	wdelayerClientKep, err := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientKep.WDelayerChangeWithdrawalDelay(newWithdrawalDelay.Uint64())
	require.Nil(t, err)
	withdrawalDelay, err := wdelayerClient.WDelayerGetWithdrawalDelay()
	require.Nil(t, err)
	assert.Equal(t, newWithdrawalDelay, withdrawalDelay)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, newWithdrawalDelay.Uint64(), wdelayerEvents.NewWithdrawalDelay[0].WithdrawalDelay)
}

func TestWDelayerDeposit(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	wdelayerClientHermez, err := NewWDelayerClient(ethereumClientHermez, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientHermez.WDelayerDeposit(auxAddressConst, tokenHezAddressConst, amount)
	require.Nil(t, err)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, amount, wdelayerEvents.Deposit[0].Amount)
	assert.Equal(t, auxAddressConst, wdelayerEvents.Deposit[0].Owner)
	assert.Equal(t, tokenHezAddressConst, wdelayerEvents.Deposit[0].Token)
}

func TestWDelayerDepositInfo(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	state, err := wdelayerClient.WDelayerDepositInfo(auxAddressConst, tokenHezAddressConst)
	require.Nil(t, err)
	assert.Equal(t, state.Amount, amount)
}

func TestWDelayerWithdrawal(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	_, err := wdelayerClient.WDelayerWithdrawal(auxAddressConst, tokenHezAddressConst)
	require.Contains(t, err.Error(), "Withdrawal not allowed yet")
	addBlocks(newWithdrawalDelay.Int64(), ethClientDialURL)
	_, err = wdelayerClient.WDelayerWithdrawal(auxAddressConst, tokenHezAddressConst)
	require.Nil(t, err)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, amount, wdelayerEvents.Withdraw[0].Amount)
	assert.Equal(t, auxAddressConst, wdelayerEvents.Withdraw[0].Owner)
	assert.Equal(t, tokenHezAddressConst, wdelayerEvents.Withdraw[0].Token)
}

func TestWDelayerEnableEmergencyMode(t *testing.T) {
	wdelayerClientKep, err := NewWDelayerClient(ethereumClientKep, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientKep.WDelayerEnableEmergencyMode()
	require.Nil(t, err)
	emergencyMode, err := wdelayerClient.WDelayerIsEmergencyMode()
	require.Nil(t, err)
	assert.Equal(t, true, emergencyMode)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	auxEvent := new(WDelayerEventEmergencyModeEnabled)
	assert.Equal(t, auxEvent, &wdelayerEvents.EmergencyModeEnabled[0])
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

func TestWDelayerEscapeHatchWithdrawal(t *testing.T) {
	wdelayerClientWhite, err := NewWDelayerClient(ethereumClientWhite, wdelayerAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientWhite.WDelayerEscapeHatchWithdrawal(governanceAddressConst, tokenHezAddressConst)
	require.Contains(t, err.Error(), "NO MAX_EMERGENCY_MODE_TIME")
	seconds := maxEmergencyModeTime.Seconds()
	addTime(seconds, ethClientDialURL)
	_, err = wdelayerClientWhite.WDelayerEscapeHatchWithdrawal(governanceAddressConst, tokenHezAddressConst)
	require.Nil(t, err)
	currentBlockNum, _ := wdelayerClient.client.EthCurrentBlock()
	wdelayerEvents, _, _ := wdelayerClient.WDelayerEventsByBlock(currentBlockNum)
	assert.Equal(t, tokenHezAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].Token)
	assert.Equal(t, governanceAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].To)
	assert.Equal(t, whiteHackGroupAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].Who)
}
