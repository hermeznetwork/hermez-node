package eth

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var wdelayerClient *WDelayerClient
var wdelayerClientTest *WDelayerClient

var initWithdrawalDelay int64 = 60
var newWithdrawalDelay int64 = 79
var maxEmergencyModeTime = time.Hour * 24 * 7 * 26
var maxWithdrawalDelay = time.Hour * 24 * 7 * 2

func TestWDelayerInit(t *testing.T) {
	wDelayerInit, blockNum, err := wdelayerClientTest.WDelayerEventInit(genesisBlock)
	require.NoError(t, err)
	assert.Equal(t, int64(16), blockNum)
	assert.Equal(t, uint64(initWithdrawalDelay), wDelayerInit.InitialWithdrawalDelay)
	assert.Equal(t, governanceAddressConst, wDelayerInit.InitialHermezGovernanceAddress)
	assert.Equal(t, emergencyCouncilAddressConst, wDelayerInit.InitialEmergencyCouncil)
}

func TestWDelayerConstants(t *testing.T) {
	wDelayerConstants, err := wdelayerClientTest.WDelayerConstants()
	require.Nil(t, err)
	assert.Equal(t, uint64(maxWithdrawalDelay.Seconds()), wDelayerConstants.MaxWithdrawalDelay)
	assert.Equal(t, uint64(maxEmergencyModeTime.Seconds()), wDelayerConstants.MaxEmergencyModeTime)
	assert.Equal(t, hermezRollupTestAddressConst, wDelayerConstants.HermezRollup)
}

func TestWDelayerGetHermezGovernanceAddress(t *testing.T) {
	governanceAddress, err := wdelayerClientTest.WDelayerGetHermezGovernanceAddress()
	require.Nil(t, err)
	assert.Equal(t, &governanceAddressConst, governanceAddress)
}

func TestWDelayerSetHermezGovernanceAddress(t *testing.T) {
	wdelayerClientAux, err := NewWDelayerClient(ethereumClientAux, wdelayerTestAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientTest.WDelayerTransferGovernance(auxAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientAux.WDelayerClaimGovernance()
	require.Nil(t, err)
	auxAddress, err := wdelayerClientTest.WDelayerGetHermezGovernanceAddress()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, auxAddressConst,
		wdelayerEvents.NewHermezGovernanceAddress[0].NewHermezGovernanceAddress)
	_, err = wdelayerClientAux.WDelayerTransferGovernance(governanceAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientTest.WDelayerClaimGovernance()
	require.Nil(t, err)
}

func TestWDelayerGetEmergencyCouncil(t *testing.T) {
	emergencyCouncil, err := wdelayerClientTest.WDelayerGetEmergencyCouncil()
	require.Nil(t, err)
	assert.Equal(t, &emergencyCouncilAddressConst, emergencyCouncil)
}

func TestWDelayerSetEmergencyCouncil(t *testing.T) {
	wdelayerClientEmergencyCouncil, err := NewWDelayerClient(ethereumClientEmergencyCouncil,
		wdelayerTestAddressConst)
	require.Nil(t, err)
	wdelayerClientAux, err := NewWDelayerClient(ethereumClientAux, wdelayerTestAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientEmergencyCouncil.WDelayerTransferEmergencyCouncil(auxAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientAux.WDelayerClaimEmergencyCouncil()
	require.Nil(t, err)
	auxAddress, err := wdelayerClientTest.WDelayerGetEmergencyCouncil()
	require.Nil(t, err)
	assert.Equal(t, &auxAddressConst, auxAddress)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, auxAddressConst, wdelayerEvents.NewEmergencyCouncil[0].NewEmergencyCouncil)
	_, err = wdelayerClientAux.WDelayerTransferEmergencyCouncil(emergencyCouncilAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientEmergencyCouncil.WDelayerClaimEmergencyCouncil()
	require.Nil(t, err)
}

func TestWDelayerIsEmergencyMode(t *testing.T) {
	emergencyMode, err := wdelayerClientTest.WDelayerIsEmergencyMode()
	require.Nil(t, err)
	assert.Equal(t, false, emergencyMode)
}

func TestWDelayerGetWithdrawalDelay(t *testing.T) {
	withdrawalDelay, err := wdelayerClientTest.WDelayerGetWithdrawalDelay()
	require.Nil(t, err)
	assert.Equal(t, initWithdrawalDelay, withdrawalDelay)
}

func TestWDelayerChangeWithdrawalDelay(t *testing.T) {
	_, err := wdelayerClientTest.WDelayerChangeWithdrawalDelay(uint64(newWithdrawalDelay))
	require.Nil(t, err)
	withdrawalDelay, err := wdelayerClientTest.WDelayerGetWithdrawalDelay()
	require.Nil(t, err)
	assert.Equal(t, newWithdrawalDelay, withdrawalDelay)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, uint64(newWithdrawalDelay), wdelayerEvents.NewWithdrawalDelay[0].WithdrawalDelay)
}

func TestWDelayerDeposit(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	wdelayerClientHermez, err := NewWDelayerClient(ethereumClientHermez, wdelayerTestAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientHermez.WDelayerDeposit(auxAddressConst, tokenHEZAddressConst, amount)
	require.Nil(t, err)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, amount, wdelayerEvents.Deposit[0].Amount)
	assert.Equal(t, auxAddressConst, wdelayerEvents.Deposit[0].Owner)
	assert.Equal(t, tokenHEZAddressConst, wdelayerEvents.Deposit[0].Token)
}

func TestWDelayerDepositInfo(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	state, err := wdelayerClientTest.WDelayerDepositInfo(auxAddressConst, tokenHEZAddressConst)
	require.Nil(t, err)
	assert.Equal(t, state.Amount, amount)
}

func TestWDelayerWithdrawal(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	_, err := wdelayerClientTest.WDelayerWithdrawal(auxAddressConst, tokenHEZAddressConst)
	require.Contains(t, err.Error(), "WITHDRAWAL_NOT_ALLOWED")
	addTime(float64(newWithdrawalDelay), ethClientDialURL)
	addBlock(ethClientDialURL)
	_, err = wdelayerClientTest.WDelayerWithdrawal(auxAddressConst, tokenHEZAddressConst)
	require.Nil(t, err)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, amount, wdelayerEvents.Withdraw[0].Amount)
	assert.Equal(t, auxAddressConst, wdelayerEvents.Withdraw[0].Owner)
	assert.Equal(t, tokenHEZAddressConst, wdelayerEvents.Withdraw[0].Token)
}

func TestWDelayerSecondDeposit(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("1100000000000000000", 10)
	wdelayerClientHermez, err := NewWDelayerClient(ethereumClientHermez, wdelayerTestAddressConst)
	require.Nil(t, err)
	_, err = wdelayerClientHermez.WDelayerDeposit(auxAddressConst, tokenHEZAddressConst, amount)
	require.Nil(t, err)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, amount, wdelayerEvents.Deposit[0].Amount)
	assert.Equal(t, auxAddressConst, wdelayerEvents.Deposit[0].Owner)
	assert.Equal(t, tokenHEZAddressConst, wdelayerEvents.Deposit[0].Token)
}

func TestWDelayerEnableEmergencyMode(t *testing.T) {
	_, err := wdelayerClientTest.WDelayerEnableEmergencyMode()
	require.Nil(t, err)
	emergencyMode, err := wdelayerClientTest.WDelayerIsEmergencyMode()
	require.Nil(t, err)
	assert.Equal(t, true, emergencyMode)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	auxEvent := new(WDelayerEventEmergencyModeEnabled)
	assert.Equal(t, auxEvent, &wdelayerEvents.EmergencyModeEnabled[0])
}

func TestWDelayerGetEmergencyModeStartingTime(t *testing.T) {
	emergencyModeStartingTime, err := wdelayerClientTest.WDelayerGetEmergencyModeStartingTime()
	require.Nil(t, err)
	// `emergencyModeStartingTime` is initialized to 0 in the smart
	// contract construction.  Since we called WDelayerEnableEmergencyMode
	// previously, `emergencyModeStartingTime` is set to the time when the
	// call was made, so it's > 0.
	assert.Greater(t, emergencyModeStartingTime, int64(0))
}

func TestWDelayerEscapeHatchWithdrawal(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("10000000000000000", 10)
	wdelayerClientEmergencyCouncil, err := NewWDelayerClient(ethereumClientEmergencyCouncil,
		wdelayerTestAddressConst)
	require.Nil(t, err)
	_, err =
		wdelayerClientEmergencyCouncil.WDelayerEscapeHatchWithdrawal(governanceAddressConst,
			tokenHEZAddressConst, amount)
	require.Contains(t, err.Error(), "NO_MAX_EMERGENCY_MODE_TIME")
	seconds := maxEmergencyModeTime.Seconds()
	addTime(seconds, ethClientDialURL)
	_, err =
		wdelayerClientEmergencyCouncil.WDelayerEscapeHatchWithdrawal(governanceAddressConst,
			tokenHEZAddressConst, amount)
	require.Nil(t, err)
	currentBlockNum, err := wdelayerClientTest.client.EthLastBlock()
	require.Nil(t, err)
	wdelayerEvents, err := wdelayerClientTest.WDelayerEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, tokenHEZAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].Token)
	assert.Equal(t, governanceAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].To)
	assert.Equal(t, emergencyCouncilAddressConst, wdelayerEvents.EscapeHatchWithdrawal[0].Who)
	assert.Equal(t, amount, wdelayerEvents.EscapeHatchWithdrawal[0].Amount)
}
