package eth

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// WDelayerConstants are the constants of the Rollup Smart Contract
type WDelayerConstants struct {
	// Max Withdrawal Delay
	MaxWithdrawalDelay uint64
	// Max Emergency mode time
	MaxEmergencyModeTime uint64
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address
}

// DepositState is the state of Deposit
type DepositState struct {
	Amount           *big.Int
	DepositTimestamp uint64
}

// WDelayerEventDeposit is an event of the WithdrawalDelayer Smart Contract
type WDelayerEventDeposit struct {
	Owner            ethCommon.Address
	Token            ethCommon.Address
	Amount           *big.Int
	DepositTimestamp uint64
}

// WDelayerEventWithdraw is an event of the WithdrawalDelayer Smart Contract
type WDelayerEventWithdraw struct {
	Owner  ethCommon.Address
	Token  ethCommon.Address
	Amount *big.Int
}

// WDelayerEventEmergencyModeEnabled an event of the WithdrawalDelayer Smart Contract
type WDelayerEventEmergencyModeEnabled struct {
}

// WDelayerEventNewWithdrawalDelay an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewWithdrawalDelay struct {
	WithdrawalDelay uint64
}

// WDelayerEventEscapeHatchWithdrawal an event of the WithdrawalDelayer Smart Contract
type WDelayerEventEscapeHatchWithdrawal struct {
	Who   ethCommon.Address
	To    ethCommon.Address
	Token ethCommon.Address
}

// WDelayerEventNewHermezKeeperAddress an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewHermezKeeperAddress struct {
	NewHermezKeeperAddress ethCommon.Address
}

// WDelayerEventNewWhiteHackGroupAddress an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewWhiteHackGroupAddress struct {
	NewWhiteHackGroupAddress ethCommon.Address
}

// WDelayerEventNewHermezGovernanceDAOAddress an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewHermezGovernanceDAOAddress struct {
	NewHermezGovernanceDAOAddress ethCommon.Address
}

// WDelayerEvents is the lis of events in a block of the WithdrawalDelayer Smart Contract
type WDelayerEvents struct { //nolint:structcheck
	Deposit                       []WDelayerEventDeposit
	Withdraw                      []WDelayerEventWithdraw
	EmergencyModeEnabled          []WDelayerEventEmergencyModeEnabled
	NewWithdrawalDelay            []WDelayerEventNewWithdrawalDelay
	EscapeHatchWithdrawal         []WDelayerEventEscapeHatchWithdrawal
	NewHermezKeeperAddress        []WDelayerEventNewHermezKeeperAddress
	NewWhiteHackGroupAddress      []WDelayerEventNewWhiteHackGroupAddress
	NewHermezGovernanceDAOAddress []WDelayerEventNewHermezGovernanceDAOAddress
}

// NewWDelayerEvents creates an empty WDelayerEvents with the slices initialized.
func NewWDelayerEvents() WDelayerEvents {
	return WDelayerEvents{
		Deposit:                       make([]WDelayerEventDeposit, 0),
		Withdraw:                      make([]WDelayerEventWithdraw, 0),
		EmergencyModeEnabled:          make([]WDelayerEventEmergencyModeEnabled, 0),
		NewWithdrawalDelay:            make([]WDelayerEventNewWithdrawalDelay, 0),
		EscapeHatchWithdrawal:         make([]WDelayerEventEscapeHatchWithdrawal, 0),
		NewHermezKeeperAddress:        make([]WDelayerEventNewHermezKeeperAddress, 0),
		NewWhiteHackGroupAddress:      make([]WDelayerEventNewWhiteHackGroupAddress, 0),
		NewHermezGovernanceDAOAddress: make([]WDelayerEventNewHermezGovernanceDAOAddress, 0),
	}
}

// WDelayerInterface is the inteface to WithdrawalDelayer Smart Contract
type WDelayerInterface interface {
	//
	// Smart Contract Methods
	//

	WDelayerGetHermezGovernanceDAOAddress() (*ethCommon.Address, error)
	WDelayerSetHermezGovernanceDAOAddress(newAddress ethCommon.Address) (*types.Transaction, error)
	WDelayerGetHermezKeeperAddress() (*ethCommon.Address, error)
	WDelayerSetHermezKeeperAddress(newAddress ethCommon.Address) (*types.Transaction, error)
	WDelayerGetWhiteHackGroupAddress() (*ethCommon.Address, error)
	WDelayerSetWhiteHackGroupAddress(newAddress ethCommon.Address) (*types.Transaction, error)
	WDelayerIsEmergencyMode() (bool, error)
	WDelayerGetWithdrawalDelay() (*big.Int, error)
	WDelayerGetEmergencyModeStartingTime() (*big.Int, error)
	WDelayerEnableEmergencyMode() (*types.Transaction, error)
	WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error)
	WDelayerDepositInfo(owner, token ethCommon.Address) (*big.Int, uint64)
	WDelayerDeposit(onwer, token ethCommon.Address, amount *big.Int) (*types.Transaction, error)
	WDelayerWithdrawal(owner, token ethCommon.Address) (*types.Transaction, error)
	WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address) (*types.Transaction, error)
}

//
// Implementation
//

// WDelayerClient is the implementation of the interface to the WithdrawDelayer Smart Contract in ethereum.
type WDelayerClient struct {
	client   *EthereumClient
	address  ethCommon.Address
	gasLimit uint64
}

// NewWDelayerClient creates a new WDelayerClient
func NewWDelayerClient(client *EthereumClient, address ethCommon.Address) *WDelayerClient {
	return &WDelayerClient{
		client:   client,
		address:  address,
		gasLimit: 1000000, //nolint:gomnd
	}
}

// WDelayerGetHermezGovernanceDAOAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetHermezGovernanceDAOAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// WDelayerSetHermezGovernanceDAOAddress is the interface to call the smart contract function
func WDelayerSetHermezGovernanceDAOAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerGetHermezKeeperAddress is the interface to call the smart contract function
func WDelayerGetHermezKeeperAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// WDelayerSetHermezKeeperAddress is the interface to call the smart contract function
func WDelayerSetHermezKeeperAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerGetWhiteHackGroupAddress is the interface to call the smart contract function
func WDelayerGetWhiteHackGroupAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// WDelayerSetWhiteHackGroupAddress is the interface to call the smart contract function
func WDelayerSetWhiteHackGroupAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerIsEmergencyMode is the interface to call the smart contract function
func WDelayerIsEmergencyMode() (bool, error) {
	return false, errTODO
}

// WDelayerGetWithdrawalDelay is the interface to call the smart contract function
func WDelayerGetWithdrawalDelay() (*big.Int, error) {
	return nil, errTODO
}

// WDelayerGetEmergencyModeStartingTime is the interface to call the smart contract function
func WDelayerGetEmergencyModeStartingTime() (*big.Int, error) {
	return nil, errTODO
}

// WDelayerEnableEmergencyMode is the interface to call the smart contract function
func WDelayerEnableEmergencyMode() (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerChangeWithdrawalDelay is the interface to call the smart contract function
func WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerDepositInfo is the interface to call the smart contract function
func WDelayerDepositInfo(owner, token ethCommon.Address) (*big.Int, uint64, error) {
	return big.NewInt(0), 0, errTODO
}

// WDelayerDeposit is the interface to call the smart contract function
func WDelayerDeposit(onwer, token ethCommon.Address, amount *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerWithdrawal is the interface to call the smart contract function
func WDelayerWithdrawal(owner, token ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// WDelayerEscapeHatchWithdrawal is the interface to call the smart contract function
func WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}
