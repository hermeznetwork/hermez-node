package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	WithdrawalDelayer "github.com/hermeznetwork/hermez-node/eth/contracts/withdrawdelayer"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

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
	TxHash           ethCommon.Hash // Hash of the transaction that generated this event
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
	Who    ethCommon.Address
	To     ethCommon.Address
	Token  ethCommon.Address
	Amount *big.Int
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
type WDelayerEvents struct {
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
	WDelayerDepositInfo(owner, token ethCommon.Address) (depositInfo DepositState, err error)
	WDelayerDeposit(onwer, token ethCommon.Address, amount *big.Int) (*types.Transaction, error)
	WDelayerWithdrawal(owner, token ethCommon.Address) (*types.Transaction, error)
	WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address, amount *big.Int) (*types.Transaction, error)

	WDelayerEventsByBlock(blockNum int64) (*WDelayerEvents, *ethCommon.Hash, error)
	WDelayerConstants() (*common.WDelayerConstants, error)
}

//
// Implementation
//

// WDelayerClient is the implementation of the interface to the WithdrawDelayer Smart Contract in ethereum.
type WDelayerClient struct {
	client      *EthereumClient
	address     ethCommon.Address
	wdelayer    *WithdrawalDelayer.WithdrawalDelayer
	contractAbi abi.ABI
}

// NewWDelayerClient creates a new WDelayerClient
func NewWDelayerClient(client *EthereumClient, address ethCommon.Address) (*WDelayerClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(WithdrawalDelayer.WithdrawalDelayerABI)))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	wdelayer, err := WithdrawalDelayer.NewWithdrawalDelayer(address, client.Client())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &WDelayerClient{
		client:      client,
		address:     address,
		wdelayer:    wdelayer,
		contractAbi: contractAbi,
	}, nil
}

// WDelayerGetHermezGovernanceDAOAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetHermezGovernanceDAOAddress() (hermezGovernanceDAOAddress *ethCommon.Address, err error) {
	var _hermezGovernanceDAOAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_hermezGovernanceDAOAddress, err = c.wdelayer.GetHermezGovernanceDAOAddress(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_hermezGovernanceDAOAddress, nil
}

// WDelayerSetHermezGovernanceDAOAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetHermezGovernanceDAOAddress(newAddress ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.SetHermezGovernanceDAOAddress(auth, newAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting hermezGovernanceDAOAddress: %w", err))
	}
	return tx, nil
}

// WDelayerGetHermezKeeperAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetHermezKeeperAddress() (hermezKeeperAddress *ethCommon.Address, err error) {
	var _hermezKeeperAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_hermezKeeperAddress, err = c.wdelayer.GetHermezKeeperAddress(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_hermezKeeperAddress, nil
}

// WDelayerSetHermezKeeperAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetHermezKeeperAddress(newAddress ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.SetHermezKeeperAddress(auth, newAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting hermezKeeperAddress: %w", err))
	}
	return tx, nil
}

// WDelayerGetWhiteHackGroupAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetWhiteHackGroupAddress() (whiteHackGroupAddress *ethCommon.Address, err error) {
	var _whiteHackGroupAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_whiteHackGroupAddress, err = c.wdelayer.GetWhiteHackGroupAddress(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_whiteHackGroupAddress, nil
}

// WDelayerSetWhiteHackGroupAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerSetWhiteHackGroupAddress(newAddress ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.SetWhiteHackGroupAddress(auth, newAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting whiteHackGroupAddress: %w", err))
	}
	return tx, nil
}

// WDelayerIsEmergencyMode is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerIsEmergencyMode() (ermergencyMode bool, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		ermergencyMode, err = c.wdelayer.IsEmergencyMode(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return false, tracerr.Wrap(err)
	}
	return ermergencyMode, nil
}

// WDelayerGetWithdrawalDelay is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetWithdrawalDelay() (withdrawalDelay *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		withdrawalDelay, err = c.wdelayer.GetWithdrawalDelay(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return withdrawalDelay, nil
}

// WDelayerGetEmergencyModeStartingTime is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetEmergencyModeStartingTime() (emergencyModeStartingTime *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		emergencyModeStartingTime, err = c.wdelayer.GetEmergencyModeStartingTime(nil)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return emergencyModeStartingTime, nil
}

// WDelayerEnableEmergencyMode is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerEnableEmergencyMode() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.EnableEmergencyMode(auth)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting enable emergency mode: %w", err))
	}
	return tx, nil
}

// WDelayerChangeWithdrawalDelay is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.ChangeWithdrawalDelay(auth, newWithdrawalDelay)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting withdrawal delay: %w", err))
	}
	return tx, nil
}

// WDelayerDepositInfo is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerDepositInfo(owner, token ethCommon.Address) (depositInfo DepositState, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		amount, depositTimestamp, err := c.wdelayer.DepositInfo(nil, owner, token)
		depositInfo.Amount = amount
		depositInfo.DepositTimestamp = depositTimestamp
		return tracerr.Wrap(err)
	}); err != nil {
		return depositInfo, tracerr.Wrap(err)
	}
	return depositInfo, nil
}

// WDelayerDeposit is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerDeposit(owner, token ethCommon.Address, amount *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.Deposit(auth, owner, token, amount)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed deposit: %w", err))
	}
	return tx, nil
}

// WDelayerWithdrawal is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerWithdrawal(owner, token ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.Withdrawal(auth, owner, token)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed withdrawal: %w", err))
	}
	return tx, nil
}

// WDelayerEscapeHatchWithdrawal is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address, amount *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.EscapeHatchWithdrawal(auth, to, token, amount)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed escapeHatchWithdrawal: %w", err))
	}
	return tx, nil
}

// WDelayerConstants returns the Constants of the WDelayer Smart Contract
func (c *WDelayerClient) WDelayerConstants() (constants *common.WDelayerConstants, err error) {
	constants = new(common.WDelayerConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		constants.MaxWithdrawalDelay, err = c.wdelayer.MAXWITHDRAWALDELAY(nil)
		if err != nil {
			return tracerr.Wrap(err)
		}
		constants.MaxEmergencyModeTime, err = c.wdelayer.MAXEMERGENCYMODETIME(nil)
		if err != nil {
			return tracerr.Wrap(err)
		}
		constants.HermezRollup, err = c.wdelayer.HermezRollupAddress(nil)
		if err != nil {
			return tracerr.Wrap(err)
		}
		return tracerr.Wrap(err)
	}); err != nil {
		return constants, tracerr.Wrap(err)
	}
	return constants, nil
}

var (
	logWDelayerDeposit                       = crypto.Keccak256Hash([]byte("Deposit(address,address,uint192,uint64)"))
	logWDelayerWithdraw                      = crypto.Keccak256Hash([]byte("Withdraw(address,address,uint192)"))
	logWDelayerEmergencyModeEnabled          = crypto.Keccak256Hash([]byte("EmergencyModeEnabled()"))
	logWDelayerNewWithdrawalDelay            = crypto.Keccak256Hash([]byte("NewWithdrawalDelay(uint64)"))
	logWDelayerEscapeHatchWithdrawal         = crypto.Keccak256Hash([]byte("EscapeHatchWithdrawal(address,address,address,uint256)"))
	logWDelayerNewHermezKeeperAddress        = crypto.Keccak256Hash([]byte("NewHermezKeeperAddress(address)"))
	logWDelayerNewWhiteHackGroupAddress      = crypto.Keccak256Hash([]byte("NewWhiteHackGroupAddress(address)"))
	logWDelayerNewHermezGovernanceDAOAddress = crypto.Keccak256Hash([]byte("NewHermezGovernanceDAOAddress(address)"))
)

// WDelayerEventsByBlock returns the events in a block that happened in the
// WDelayer Smart Contract and the blockHash where the eents happened.  If
// there are no events in that block, blockHash is nil.
func (c *WDelayerClient) WDelayerEventsByBlock(blockNum int64) (*WDelayerEvents, *ethCommon.Hash, error) {
	var wdelayerEvents WDelayerEvents
	var blockHash *ethCommon.Hash

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []ethCommon.Address{
			c.address,
		},
		BlockHash: nil,
		Topics:    [][]ethCommon.Hash{},
	}

	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	if len(logs) > 0 {
		blockHash = &logs[0].BlockHash
	}
	for _, vLog := range logs {
		if vLog.BlockHash != *blockHash {
			log.Errorw("Block hash mismatch", "expected", blockHash.String(), "got", vLog.BlockHash.String())
			return nil, nil, tracerr.Wrap(ErrBlockHashMismatchEvent)
		}
		switch vLog.Topics[0] {
		case logWDelayerDeposit:
			var deposit WDelayerEventDeposit
			err := c.contractAbi.Unpack(&deposit, "Deposit", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			deposit.Owner = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			deposit.Token = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			deposit.TxHash = vLog.TxHash
			wdelayerEvents.Deposit = append(wdelayerEvents.Deposit, deposit)

		case logWDelayerWithdraw:
			var withdraw WDelayerEventWithdraw
			err := c.contractAbi.Unpack(&withdraw, "Withdraw", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			withdraw.Token = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			withdraw.Owner = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			wdelayerEvents.Withdraw = append(wdelayerEvents.Withdraw, withdraw)

		case logWDelayerEmergencyModeEnabled:
			var emergencyModeEnabled WDelayerEventEmergencyModeEnabled
			wdelayerEvents.EmergencyModeEnabled = append(wdelayerEvents.EmergencyModeEnabled, emergencyModeEnabled)

		case logWDelayerNewWithdrawalDelay:
			var withdrawalDelay WDelayerEventNewWithdrawalDelay
			err := c.contractAbi.Unpack(&withdrawalDelay, "NewWithdrawalDelay", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewWithdrawalDelay = append(wdelayerEvents.NewWithdrawalDelay, withdrawalDelay)

		case logWDelayerEscapeHatchWithdrawal:
			var escapeHatchWithdrawal WDelayerEventEscapeHatchWithdrawal
			err := c.contractAbi.Unpack(&escapeHatchWithdrawal, "EscapeHatchWithdrawal", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			escapeHatchWithdrawal.Who = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			escapeHatchWithdrawal.To = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			escapeHatchWithdrawal.Token = ethCommon.BytesToAddress(vLog.Topics[3].Bytes())
			wdelayerEvents.EscapeHatchWithdrawal = append(wdelayerEvents.EscapeHatchWithdrawal, escapeHatchWithdrawal)

		case logWDelayerNewHermezKeeperAddress:
			var keeperAddress WDelayerEventNewHermezKeeperAddress
			err := c.contractAbi.Unpack(&keeperAddress, "NewHermezKeeperAddress", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewHermezKeeperAddress = append(wdelayerEvents.NewHermezKeeperAddress, keeperAddress)

		case logWDelayerNewWhiteHackGroupAddress:
			var whiteHackGroupAddress WDelayerEventNewWhiteHackGroupAddress
			err := c.contractAbi.Unpack(&whiteHackGroupAddress, "NewWhiteHackGroupAddress", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewWhiteHackGroupAddress = append(wdelayerEvents.NewWhiteHackGroupAddress, whiteHackGroupAddress)

		case logWDelayerNewHermezGovernanceDAOAddress:
			var governanceDAOAddress WDelayerEventNewHermezGovernanceDAOAddress
			err := c.contractAbi.Unpack(&governanceDAOAddress, "NewHermezGovernanceDAOAddress", vLog.Data)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewHermezGovernanceDAOAddress = append(wdelayerEvents.NewHermezGovernanceDAOAddress, governanceDAOAddress)
		}
	}
	return &wdelayerEvents, blockHash, nil
}
