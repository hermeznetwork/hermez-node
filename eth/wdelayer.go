package eth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	WithdrawalDelayer "github.com/hermeznetwork/hermez-node/eth/contracts/withdrawdelayer"
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
	var hermezGovernanceDAOAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		hermezGovernanceDAOAddress, err = wdelayer.GetHermezGovernanceDAOAddress(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &hermezGovernanceDAOAddress, nil
}

// WDelayerSetHermezGovernanceDAOAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetHermezGovernanceDAOAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.SetHermezGovernanceDAOAddress(auth, newAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting hermezGovernanceDAOAddress: %w", err)
	}
	return tx, nil
}

// WDelayerGetHermezKeeperAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetHermezKeeperAddress() (*ethCommon.Address, error) {
	var hermezKeeperAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		hermezKeeperAddress, err = wdelayer.GetHermezKeeperAddress(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &hermezKeeperAddress, nil
}

// WDelayerSetHermezKeeperAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetHermezKeeperAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.SetHermezKeeperAddress(auth, newAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting hermezKeeperAddress: %w", err)
	}
	return tx, nil
}

// WDelayerGetWhiteHackGroupAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetWhiteHackGroupAddress() (*ethCommon.Address, error) {
	var whiteHackGroupAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		whiteHackGroupAddress, err = wdelayer.GetWhiteHackGroupAddress(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &whiteHackGroupAddress, nil
}

// WDelayerSetWhiteHackGroupAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetWhiteHackGroupAddress(newAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.SetWhiteHackGroupAddress(auth, newAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting whiteHackGroupAddress: %w", err)
	}
	return tx, nil
}

// WDelayerIsEmergencyMode is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerIsEmergencyMode() (bool, error) {
	var ermergencyMode bool
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		ermergencyMode, err = wdelayer.IsEmergencyMode(nil)
		return err
	}); err != nil {
		return false, err
	}
	return ermergencyMode, nil
}

// WDelayerGetWithdrawalDelay is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetWithdrawalDelay() (*big.Int, error) {
	var withdrawalDelay *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		withdrawalDelay, err = wdelayer.GetWithdrawalDelay(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return withdrawalDelay, nil
}

// WDelayerGetEmergencyModeStartingTime is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetEmergencyModeStartingTime() (*big.Int, error) {
	var emergencyModeStartingTime *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		emergencyModeStartingTime, err = wdelayer.GetEmergencyModeStartingTime(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return emergencyModeStartingTime, nil
}

// WDelayerEnableEmergencyMode is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerEnableEmergencyMode() (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.EnableEmergencyMode(auth)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting enable emergency mode: %w", err)
	}
	return tx, nil
}

// WDelayerChangeWithdrawalDelay is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.ChangeWithdrawalDelay(auth, newWithdrawalDelay)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting withdrawal delay: %w", err)
	}
	return tx, nil
}

// WDelayerDepositInfo is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerDepositInfo(owner, token ethCommon.Address) (DepositState, error) {
	var depositInfo DepositState
	if err := c.client.Call(func(ec *ethclient.Client) error {
		wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
		if err != nil {
			return err
		}
		amount, depositTimestamp, err := wdelayer.DepositInfo(nil, owner, token)
		depositInfo.Amount = amount
		depositInfo.DepositTimestamp = depositTimestamp
		return err
	}); err != nil {
		return depositInfo, err
	}
	return depositInfo, nil
}

// WDelayerDeposit is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerDeposit(owner, token ethCommon.Address, amount *big.Int) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.Deposit(auth, owner, token, amount)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed deposit: %w", err)
	}
	return tx, nil
}

// WDelayerWithdrawal is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerWithdrawal(owner, token ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.Withdrawal(auth, owner, token)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed withdrawal: %w", err)
	}
	return tx, nil
}

// WDelayerEscapeHatchWithdrawal is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(c.address, ec)
			if err != nil {
				return nil, err
			}
			return wdelayer.EscapeHatchWithdrawal(auth, to, token)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed escapeHatchWithdrawal: %w", err)
	}
	return tx, nil
}
