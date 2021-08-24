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
	withdrawaldelayer "github.com/hermeznetwork/hermez-node/eth/contracts/withdrawaldelayer"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

// DepositState is the state of Deposit
type DepositState struct {
	Amount           *big.Int
	DepositTimestamp uint64
}

// WDelayerEventInitialize is the InitializeWithdrawalDelayerEvent event of the
// Smart Contract
type WDelayerEventInitialize struct {
	InitialWithdrawalDelay         uint64
	InitialHermezGovernanceAddress ethCommon.Address
	InitialEmergencyCouncil        ethCommon.Address
}

// WDelayerVariables returns the WDelayerVariables from the initialize event
func (ei *WDelayerEventInitialize) WDelayerVariables() *common.WDelayerVariables {
	return &common.WDelayerVariables{
		EthBlockNum:                0,
		HermezGovernanceAddress:    ei.InitialHermezGovernanceAddress,
		EmergencyCouncilAddress:    ei.InitialEmergencyCouncil,
		WithdrawalDelay:            ei.InitialWithdrawalDelay,
		EmergencyModeStartingBlock: 0,
		EmergencyMode:              false,
	}
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

// WDelayerEventNewEmergencyCouncil an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewEmergencyCouncil struct {
	NewEmergencyCouncil ethCommon.Address
}

// WDelayerEventNewHermezGovernanceAddress an event of the WithdrawalDelayer Smart Contract
type WDelayerEventNewHermezGovernanceAddress struct {
	NewHermezGovernanceAddress ethCommon.Address
}

// WDelayerEvents is the lis of events in a block of the WithdrawalDelayer Smart Contract
type WDelayerEvents struct {
	Deposit                    []WDelayerEventDeposit
	Withdraw                   []WDelayerEventWithdraw
	EmergencyModeEnabled       []WDelayerEventEmergencyModeEnabled
	NewWithdrawalDelay         []WDelayerEventNewWithdrawalDelay
	EscapeHatchWithdrawal      []WDelayerEventEscapeHatchWithdrawal
	NewEmergencyCouncil        []WDelayerEventNewEmergencyCouncil
	NewHermezGovernanceAddress []WDelayerEventNewHermezGovernanceAddress
}

// NewWDelayerEvents creates an empty WDelayerEvents with the slices initialized.
func NewWDelayerEvents() WDelayerEvents {
	return WDelayerEvents{
		Deposit:                    make([]WDelayerEventDeposit, 0),
		Withdraw:                   make([]WDelayerEventWithdraw, 0),
		EmergencyModeEnabled:       make([]WDelayerEventEmergencyModeEnabled, 0),
		NewWithdrawalDelay:         make([]WDelayerEventNewWithdrawalDelay, 0),
		EscapeHatchWithdrawal:      make([]WDelayerEventEscapeHatchWithdrawal, 0),
		NewEmergencyCouncil:        make([]WDelayerEventNewEmergencyCouncil, 0),
		NewHermezGovernanceAddress: make([]WDelayerEventNewHermezGovernanceAddress, 0),
	}
}

// WDelayerInterface is the inteface to WithdrawalDelayer Smart Contract
type WDelayerInterface interface {
	//
	// Smart Contract Methods
	//

	WDelayerGetHermezGovernanceAddress() (*ethCommon.Address, error)
	WDelayerTransferGovernance(newAddress ethCommon.Address) (*types.Transaction, error)
	WDelayerClaimGovernance() (*types.Transaction, error)
	WDelayerGetEmergencyCouncil() (*ethCommon.Address, error)
	WDelayerTransferEmergencyCouncil(newAddress ethCommon.Address) (*types.Transaction, error)
	WDelayerClaimEmergencyCouncil() (*types.Transaction, error)
	WDelayerIsEmergencyMode() (bool, error)
	WDelayerGetWithdrawalDelay() (int64, error)
	WDelayerGetEmergencyModeStartingTime() (int64, error)
	WDelayerEnableEmergencyMode() (*types.Transaction, error)
	WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error)
	WDelayerDepositInfo(owner, token ethCommon.Address) (depositInfo DepositState, err error)
	WDelayerDeposit(onwer, token ethCommon.Address, amount *big.Int) (*types.Transaction, error)
	WDelayerWithdrawal(owner, token ethCommon.Address) (*types.Transaction, error)
	WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address,
		amount *big.Int) (*types.Transaction, error)

	WDelayerEventsByBlock(blockNum int64, blockHash *ethCommon.Hash) (*WDelayerEvents, error)
	WDelayerConstants() (*common.WDelayerConstants, error)
	WDelayerEventInit(genesisBlockNum int64) (*WDelayerEventInitialize, int64, error)
}

//
// Implementation
//

// WDelayerClient is the implementation of the interface to the WithdrawDelayer
// Smart Contract in ethereum.
type WDelayerClient struct {
	client      *EthereumClient
	address     ethCommon.Address
	wdelayer    *withdrawaldelayer.Withdrawaldelayer
	contractAbi abi.ABI
	opts        *bind.CallOpts
}

// NewWDelayerClient creates a new WDelayerClient
func NewWDelayerClient(client *EthereumClient, address ethCommon.Address) (*WDelayerClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(withdrawaldelayer.WithdrawaldelayerABI)))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	wdelayer, err := withdrawaldelayer.NewWithdrawaldelayer(address, client.Client())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &WDelayerClient{
		client:      client,
		address:     address,
		wdelayer:    wdelayer,
		contractAbi: contractAbi,
		opts:        newCallOpts(),
	}, nil
}

// WDelayerGetHermezGovernanceAddress is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetHermezGovernanceAddress() (
	hermezGovernanceAddress *ethCommon.Address, err error) {
	var _hermezGovernanceAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_hermezGovernanceAddress, err = c.wdelayer.GetHermezGovernanceAddress(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_hermezGovernanceAddress, nil
}

// WDelayerTransferGovernance is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerTransferGovernance(newAddress ethCommon.Address) (
	tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.TransferGovernance(auth, newAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed transfer hermezGovernanceAddress: %w", err))
	}
	return tx, nil
}

// WDelayerClaimGovernance is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerClaimGovernance() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.ClaimGovernance(auth)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed claim hermezGovernanceAddress: %w", err))
	}
	return tx, nil
}

// WDelayerGetEmergencyCouncil is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetEmergencyCouncil() (emergencyCouncilAddress *ethCommon.Address,
	err error) {
	var _emergencyCouncilAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_emergencyCouncilAddress, err = c.wdelayer.GetEmergencyCouncil(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_emergencyCouncilAddress, nil
}

// WDelayerTransferEmergencyCouncil is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerTransferEmergencyCouncil(newAddress ethCommon.Address) (
	tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.TransferEmergencyCouncil(auth, newAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed transfer EmergencyCouncil: %w", err))
	}
	return tx, nil
}

// WDelayerClaimEmergencyCouncil is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerClaimEmergencyCouncil() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.wdelayer.ClaimEmergencyCouncil(auth)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed claim EmergencyCouncil: %w", err))
	}
	return tx, nil
}

// WDelayerIsEmergencyMode is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerIsEmergencyMode() (ermergencyMode bool, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		ermergencyMode, err = c.wdelayer.IsEmergencyMode(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return false, tracerr.Wrap(err)
	}
	return ermergencyMode, nil
}

// WDelayerGetWithdrawalDelay is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetWithdrawalDelay() (withdrawalDelay int64, err error) {
	var _withdrawalDelay uint64
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_withdrawalDelay, err = c.wdelayer.GetWithdrawalDelay(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return int64(_withdrawalDelay), nil
}

// WDelayerGetEmergencyModeStartingTime is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerGetEmergencyModeStartingTime() (emergencyModeStartingTime int64,
	err error) {
	var _emergencyModeStartingTime uint64
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_emergencyModeStartingTime, err = c.wdelayer.GetEmergencyModeStartingTime(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return int64(_emergencyModeStartingTime), nil
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
func (c *WDelayerClient) WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (
	tx *types.Transaction, err error) {
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
func (c *WDelayerClient) WDelayerDepositInfo(owner, token ethCommon.Address) (
	depositInfo DepositState, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		amount, depositTimestamp, err := c.wdelayer.DepositInfo(c.opts, owner, token)
		depositInfo.Amount = amount
		depositInfo.DepositTimestamp = depositTimestamp
		return tracerr.Wrap(err)
	}); err != nil {
		return depositInfo, tracerr.Wrap(err)
	}
	return depositInfo, nil
}

// WDelayerDeposit is the interface to call the smart contract function
func (c *WDelayerClient) WDelayerDeposit(owner, token ethCommon.Address, amount *big.Int) (
	tx *types.Transaction, err error) {
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
func (c *WDelayerClient) WDelayerWithdrawal(owner, token ethCommon.Address) (tx *types.Transaction,
	err error) {
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
func (c *WDelayerClient) WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address,
	amount *big.Int) (tx *types.Transaction, err error) {
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
		constants.MaxWithdrawalDelay, err = c.wdelayer.MAXWITHDRAWALDELAY(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		constants.MaxEmergencyModeTime, err = c.wdelayer.MAXEMERGENCYMODETIME(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		constants.HermezRollup, err = c.wdelayer.HermezRollupAddress(c.opts)
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
	logWDelayerDeposit = crypto.Keccak256Hash([]byte(
		"Deposit(address,address,uint192,uint64)"))
	logWDelayerWithdraw = crypto.Keccak256Hash([]byte(
		"Withdraw(address,address,uint192)"))
	logWDelayerEmergencyModeEnabled = crypto.Keccak256Hash([]byte(
		"EmergencyModeEnabled()"))
	logWDelayerNewWithdrawalDelay = crypto.Keccak256Hash([]byte(
		"NewWithdrawalDelay(uint64)"))
	logWDelayerEscapeHatchWithdrawal = crypto.Keccak256Hash([]byte(
		"EscapeHatchWithdrawal(address,address,address,uint256)"))
	logWDelayerNewEmergencyCouncil = crypto.Keccak256Hash([]byte(
		"NewEmergencyCouncil(address)"))
	logWDelayerNewHermezGovernanceAddress = crypto.Keccak256Hash([]byte(
		"NewHermezGovernanceAddress(address)"))
	logWDelayerInitialize = crypto.Keccak256Hash([]byte(
		"InitializeWithdrawalDelayerEvent(uint64,address,address)"))
)

// WDelayerEventInit returns the initialize event with its corresponding block number
func (c *WDelayerClient) WDelayerEventInit(genesisBlockNum int64) (*WDelayerEventInitialize, int64, error) {
	query := ethereum.FilterQuery{
		Addresses: []ethCommon.Address{
			c.address,
		},
		FromBlock: big.NewInt(max(0, genesisBlockNum-blocksPerDay)),
		ToBlock:   big.NewInt(genesisBlockNum),
		Topics:    [][]ethCommon.Hash{{logWDelayerInitialize}},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(logs) != 1 {
		return nil, 0, tracerr.Wrap(fmt.Errorf("no event of type InitializeWithdrawalDelayerEvent found"))
	}
	vLog := logs[0]
	if vLog.Topics[0] != logWDelayerInitialize {
		return nil, 0, tracerr.Wrap(fmt.Errorf("event is not InitializeWithdrawalDelayerEvent"))
	}

	var wDelayerInit WDelayerEventInitialize
	if err := c.contractAbi.UnpackIntoInterface(&wDelayerInit, "InitializeWithdrawalDelayerEvent",
		vLog.Data); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	return &wDelayerInit, int64(vLog.BlockNumber), tracerr.Wrap(err)
}

// WDelayerEventsByBlock returns the events in a block that happened in the
// WDelayer Smart Contract.
// To query by blockNum, set blockNum >= 0 and blockHash == nil.
// To query by blockHash set blockHash != nil, and blockNum will be ignored.
// If there are no events in that block the result is nil.
func (c *WDelayerClient) WDelayerEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*WDelayerEvents, error) {
	var wdelayerEvents WDelayerEvents

	var blockNumBigInt *big.Int
	if blockHash == nil {
		blockNumBigInt = big.NewInt(blockNum)
	}
	query := ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: blockNumBigInt,
		ToBlock:   blockNumBigInt,
		Addresses: []ethCommon.Address{
			c.address,
		},
		Topics: [][]ethCommon.Hash{},
	}

	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if len(logs) == 0 {
		return nil, nil
	}

	for _, vLog := range logs {
		if blockHash != nil && vLog.BlockHash != *blockHash {
			log.Errorw("Block hash mismatch", "expected", blockHash.String(), "got", vLog.BlockHash.String())
			return nil, tracerr.Wrap(ErrBlockHashMismatchEvent)
		}
		switch vLog.Topics[0] {
		case logWDelayerDeposit:
			var deposit WDelayerEventDeposit
			err := c.contractAbi.UnpackIntoInterface(&deposit, "Deposit", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			deposit.Owner = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			deposit.Token = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			deposit.TxHash = vLog.TxHash
			wdelayerEvents.Deposit = append(wdelayerEvents.Deposit, deposit)

		case logWDelayerWithdraw:
			var withdraw WDelayerEventWithdraw
			err := c.contractAbi.UnpackIntoInterface(&withdraw, "Withdraw", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			withdraw.Token = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			withdraw.Owner = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			wdelayerEvents.Withdraw = append(wdelayerEvents.Withdraw, withdraw)

		case logWDelayerEmergencyModeEnabled:
			var emergencyModeEnabled WDelayerEventEmergencyModeEnabled
			wdelayerEvents.EmergencyModeEnabled =
				append(wdelayerEvents.EmergencyModeEnabled, emergencyModeEnabled)

		case logWDelayerNewWithdrawalDelay:
			var withdrawalDelay WDelayerEventNewWithdrawalDelay
			err := c.contractAbi.UnpackIntoInterface(&withdrawalDelay,
				"NewWithdrawalDelay", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewWithdrawalDelay =
				append(wdelayerEvents.NewWithdrawalDelay, withdrawalDelay)

		case logWDelayerEscapeHatchWithdrawal:
			var escapeHatchWithdrawal WDelayerEventEscapeHatchWithdrawal
			err := c.contractAbi.UnpackIntoInterface(&escapeHatchWithdrawal,
				"EscapeHatchWithdrawal", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			escapeHatchWithdrawal.Who = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			escapeHatchWithdrawal.To = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			escapeHatchWithdrawal.Token = ethCommon.BytesToAddress(vLog.Topics[3].Bytes())
			wdelayerEvents.EscapeHatchWithdrawal =
				append(wdelayerEvents.EscapeHatchWithdrawal, escapeHatchWithdrawal)

		case logWDelayerNewEmergencyCouncil:
			var emergencyCouncil WDelayerEventNewEmergencyCouncil
			err := c.contractAbi.UnpackIntoInterface(&emergencyCouncil,
				"NewEmergencyCouncil", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewEmergencyCouncil =
				append(wdelayerEvents.NewEmergencyCouncil, emergencyCouncil)

		case logWDelayerNewHermezGovernanceAddress:
			var governanceAddress WDelayerEventNewHermezGovernanceAddress
			err := c.contractAbi.UnpackIntoInterface(&governanceAddress,
				"NewHermezGovernanceAddress", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			wdelayerEvents.NewHermezGovernanceAddress =
				append(wdelayerEvents.NewHermezGovernanceAddress, governanceAddress)
		}
	}
	return &wdelayerEvents, nil
}
