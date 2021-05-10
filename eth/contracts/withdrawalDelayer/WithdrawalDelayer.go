// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package WithdrawalDelayer

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// WithdrawalDelayerABI is the input ABI used to generate the binding from.
const WithdrawalDelayerABI = "[{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_initialWithdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_initialHermezRollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezGovernanceAddress\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_initialEmergencyCouncil\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EmergencyModeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EscapeHatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"initialWithdrawalDelay\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialHermezGovernanceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"InitializeWithdrawalDelayerEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"NewEmergencyCouncil\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezGovernanceAddress\",\"type\":\"address\"}],\"name\":\"NewHermezGovernanceAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"NewWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_EMERGENCY_MODE_TIME\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_WITHDRAWAL_DELAY\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"changeWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"_amount\",\"type\":\"uint192\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"depositInfo\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enableEmergencyMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"escapeHatchWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyCouncil\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyModeStartingTime\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezGovernanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWithdrawalDelay\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isEmergencyMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingEmergencyCouncil\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"transferEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newGovernance\",\"type\":\"address\"}],\"name\":\"transferGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// WithdrawalDelayer is an auto generated Go binding around an Ethereum contract.
type WithdrawalDelayer struct {
	WithdrawalDelayerCaller     // Read-only binding to the contract
	WithdrawalDelayerTransactor // Write-only binding to the contract
	WithdrawalDelayerFilterer   // Log filterer for contract events
}

// WithdrawalDelayerCaller is an auto generated read-only Go binding around an Ethereum contract.
type WithdrawalDelayerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalDelayerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WithdrawalDelayerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalDelayerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WithdrawalDelayerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawalDelayerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WithdrawalDelayerSession struct {
	Contract     *WithdrawalDelayer // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// WithdrawalDelayerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WithdrawalDelayerCallerSession struct {
	Contract *WithdrawalDelayerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// WithdrawalDelayerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WithdrawalDelayerTransactorSession struct {
	Contract     *WithdrawalDelayerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// WithdrawalDelayerRaw is an auto generated low-level Go binding around an Ethereum contract.
type WithdrawalDelayerRaw struct {
	Contract *WithdrawalDelayer // Generic contract binding to access the raw methods on
}

// WithdrawalDelayerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WithdrawalDelayerCallerRaw struct {
	Contract *WithdrawalDelayerCaller // Generic read-only contract binding to access the raw methods on
}

// WithdrawalDelayerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WithdrawalDelayerTransactorRaw struct {
	Contract *WithdrawalDelayerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWithdrawalDelayer creates a new instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayer(address common.Address, backend bind.ContractBackend) (*WithdrawalDelayer, error) {
	contract, err := bindWithdrawalDelayer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayer{WithdrawalDelayerCaller: WithdrawalDelayerCaller{contract: contract}, WithdrawalDelayerTransactor: WithdrawalDelayerTransactor{contract: contract}, WithdrawalDelayerFilterer: WithdrawalDelayerFilterer{contract: contract}}, nil
}

// NewWithdrawalDelayerCaller creates a new read-only instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerCaller(address common.Address, caller bind.ContractCaller) (*WithdrawalDelayerCaller, error) {
	contract, err := bindWithdrawalDelayer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerCaller{contract: contract}, nil
}

// NewWithdrawalDelayerTransactor creates a new write-only instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerTransactor(address common.Address, transactor bind.ContractTransactor) (*WithdrawalDelayerTransactor, error) {
	contract, err := bindWithdrawalDelayer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerTransactor{contract: contract}, nil
}

// NewWithdrawalDelayerFilterer creates a new log filterer instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerFilterer(address common.Address, filterer bind.ContractFilterer) (*WithdrawalDelayerFilterer, error) {
	contract, err := bindWithdrawalDelayer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerFilterer{contract: contract}, nil
}

// bindWithdrawalDelayer binds a generic wrapper to an already deployed contract.
func bindWithdrawalDelayer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WithdrawalDelayerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WithdrawalDelayer *WithdrawalDelayerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WithdrawalDelayer *WithdrawalDelayerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WithdrawalDelayer *WithdrawalDelayerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_WithdrawalDelayer *WithdrawalDelayerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _WithdrawalDelayer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_WithdrawalDelayer *WithdrawalDelayerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_WithdrawalDelayer *WithdrawalDelayerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.contract.Transact(opts, method, params...)
}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) MAXEMERGENCYMODETIME(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "MAX_EMERGENCY_MODE_TIME")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerSession) MAXEMERGENCYMODETIME() (uint64, error) {
	return _WithdrawalDelayer.Contract.MAXEMERGENCYMODETIME(&_WithdrawalDelayer.CallOpts)
}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) MAXEMERGENCYMODETIME() (uint64, error) {
	return _WithdrawalDelayer.Contract.MAXEMERGENCYMODETIME(&_WithdrawalDelayer.CallOpts)
}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) MAXWITHDRAWALDELAY(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "MAX_WITHDRAWAL_DELAY")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerSession) MAXWITHDRAWALDELAY() (uint64, error) {
	return _WithdrawalDelayer.Contract.MAXWITHDRAWALDELAY(&_WithdrawalDelayer.CallOpts)
}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) MAXWITHDRAWALDELAY() (uint64, error) {
	return _WithdrawalDelayer.Contract.MAXWITHDRAWALDELAY(&_WithdrawalDelayer.CallOpts)
}

// DepositInfo is a free data retrieval call binding the contract method 0x493b0170.
//
// Solidity: function depositInfo(address _owner, address _token) view returns(uint192, uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) DepositInfo(opts *bind.CallOpts, _owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "depositInfo", _owner, _token)

	if err != nil {
		return *new(*big.Int), *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return out0, out1, err

}

// DepositInfo is a free data retrieval call binding the contract method 0x493b0170.
//
// Solidity: function depositInfo(address _owner, address _token) view returns(uint192, uint64)
func (_WithdrawalDelayer *WithdrawalDelayerSession) DepositInfo(_owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	return _WithdrawalDelayer.Contract.DepositInfo(&_WithdrawalDelayer.CallOpts, _owner, _token)
}

// DepositInfo is a free data retrieval call binding the contract method 0x493b0170.
//
// Solidity: function depositInfo(address _owner, address _token) view returns(uint192, uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) DepositInfo(_owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	return _WithdrawalDelayer.Contract.DepositInfo(&_WithdrawalDelayer.CallOpts, _owner, _token)
}

// Deposits is a free data retrieval call binding the contract method 0x3d4dff7b.
//
// Solidity: function deposits(bytes32 ) view returns(uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) Deposits(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "deposits", arg0)

	outstruct := new(struct {
		Amount           *big.Int
		DepositTimestamp uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.DepositTimestamp = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// Deposits is a free data retrieval call binding the contract method 0x3d4dff7b.
//
// Solidity: function deposits(bytes32 ) view returns(uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerSession) Deposits(arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	return _WithdrawalDelayer.Contract.Deposits(&_WithdrawalDelayer.CallOpts, arg0)
}

// Deposits is a free data retrieval call binding the contract method 0x3d4dff7b.
//
// Solidity: function deposits(bytes32 ) view returns(uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) Deposits(arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	return _WithdrawalDelayer.Contract.Deposits(&_WithdrawalDelayer.CallOpts, arg0)
}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetEmergencyCouncil(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "getEmergencyCouncil")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetEmergencyCouncil() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyCouncil(&_WithdrawalDelayer.CallOpts)
}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetEmergencyCouncil() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyCouncil(&_WithdrawalDelayer.CallOpts)
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetEmergencyModeStartingTime(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "getEmergencyModeStartingTime")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetEmergencyModeStartingTime() (uint64, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyModeStartingTime(&_WithdrawalDelayer.CallOpts)
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetEmergencyModeStartingTime() (uint64, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyModeStartingTime(&_WithdrawalDelayer.CallOpts)
}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetHermezGovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "getHermezGovernanceAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetHermezGovernanceAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezGovernanceAddress(&_WithdrawalDelayer.CallOpts)
}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetHermezGovernanceAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezGovernanceAddress(&_WithdrawalDelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetWithdrawalDelay(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "getWithdrawalDelay")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetWithdrawalDelay() (uint64, error) {
	return _WithdrawalDelayer.Contract.GetWithdrawalDelay(&_WithdrawalDelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetWithdrawalDelay() (uint64, error) {
	return _WithdrawalDelayer.Contract.GetWithdrawalDelay(&_WithdrawalDelayer.CallOpts)
}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) HermezRollupAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "hermezRollupAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) HermezRollupAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.HermezRollupAddress(&_WithdrawalDelayer.CallOpts)
}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) HermezRollupAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.HermezRollupAddress(&_WithdrawalDelayer.CallOpts)
}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) IsEmergencyMode(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "isEmergencyMode")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_WithdrawalDelayer *WithdrawalDelayerSession) IsEmergencyMode() (bool, error) {
	return _WithdrawalDelayer.Contract.IsEmergencyMode(&_WithdrawalDelayer.CallOpts)
}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) IsEmergencyMode() (bool, error) {
	return _WithdrawalDelayer.Contract.IsEmergencyMode(&_WithdrawalDelayer.CallOpts)
}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) PendingEmergencyCouncil(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "pendingEmergencyCouncil")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) PendingEmergencyCouncil() (common.Address, error) {
	return _WithdrawalDelayer.Contract.PendingEmergencyCouncil(&_WithdrawalDelayer.CallOpts)
}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) PendingEmergencyCouncil() (common.Address, error) {
	return _WithdrawalDelayer.Contract.PendingEmergencyCouncil(&_WithdrawalDelayer.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) PendingGovernance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _WithdrawalDelayer.contract.Call(opts, &out, "pendingGovernance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) PendingGovernance() (common.Address, error) {
	return _WithdrawalDelayer.Contract.PendingGovernance(&_WithdrawalDelayer.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) PendingGovernance() (common.Address, error) {
	return _WithdrawalDelayer.Contract.PendingGovernance(&_WithdrawalDelayer.CallOpts)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) ChangeWithdrawalDelay(opts *bind.TransactOpts, _newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "changeWithdrawalDelay", _newWithdrawalDelay)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) ChangeWithdrawalDelay(_newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ChangeWithdrawalDelay(&_WithdrawalDelayer.TransactOpts, _newWithdrawalDelay)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) ChangeWithdrawalDelay(_newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ChangeWithdrawalDelay(&_WithdrawalDelayer.TransactOpts, _newWithdrawalDelay)
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) ClaimEmergencyCouncil(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "claimEmergencyCouncil")
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) ClaimEmergencyCouncil() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ClaimEmergencyCouncil(&_WithdrawalDelayer.TransactOpts)
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) ClaimEmergencyCouncil() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ClaimEmergencyCouncil(&_WithdrawalDelayer.TransactOpts)
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) ClaimGovernance(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "claimGovernance")
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) ClaimGovernance() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ClaimGovernance(&_WithdrawalDelayer.TransactOpts)
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) ClaimGovernance() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.ClaimGovernance(&_WithdrawalDelayer.TransactOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) Deposit(opts *bind.TransactOpts, _owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "deposit", _owner, _token, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) Deposit(_owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Deposit(&_WithdrawalDelayer.TransactOpts, _owner, _token, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) Deposit(_owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Deposit(&_WithdrawalDelayer.TransactOpts, _owner, _token, _amount)
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) EnableEmergencyMode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "enableEmergencyMode")
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) EnableEmergencyMode() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EnableEmergencyMode(&_WithdrawalDelayer.TransactOpts)
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) EnableEmergencyMode() (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EnableEmergencyMode(&_WithdrawalDelayer.TransactOpts)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) EscapeHatchWithdrawal(opts *bind.TransactOpts, _to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "escapeHatchWithdrawal", _to, _token, _amount)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EscapeHatchWithdrawal(&_WithdrawalDelayer.TransactOpts, _to, _token, _amount)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EscapeHatchWithdrawal(&_WithdrawalDelayer.TransactOpts, _to, _token, _amount)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) TransferEmergencyCouncil(opts *bind.TransactOpts, newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "transferEmergencyCouncil", newEmergencyCouncil)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) TransferEmergencyCouncil(newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.TransferEmergencyCouncil(&_WithdrawalDelayer.TransactOpts, newEmergencyCouncil)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) TransferEmergencyCouncil(newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.TransferEmergencyCouncil(&_WithdrawalDelayer.TransactOpts, newEmergencyCouncil)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) TransferGovernance(opts *bind.TransactOpts, newGovernance common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "transferGovernance", newGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) TransferGovernance(newGovernance common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.TransferGovernance(&_WithdrawalDelayer.TransactOpts, newGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) TransferGovernance(newGovernance common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.TransferGovernance(&_WithdrawalDelayer.TransactOpts, newGovernance)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) Withdrawal(opts *bind.TransactOpts, _owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "withdrawal", _owner, _token)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) Withdrawal(_owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Withdrawal(&_WithdrawalDelayer.TransactOpts, _owner, _token)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) Withdrawal(_owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Withdrawal(&_WithdrawalDelayer.TransactOpts, _owner, _token)
}

// WithdrawalDelayerDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerDepositIterator struct {
	Event *WithdrawalDelayerDeposit // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerDeposit)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerDeposit)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerDeposit represents a Deposit event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerDeposit struct {
	Owner            common.Address
	Token            common.Address
	Amount           *big.Int
	DepositTimestamp uint64
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterDeposit(opts *bind.FilterOpts, owner []common.Address, token []common.Address) (*WithdrawalDelayerDepositIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "Deposit", ownerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerDepositIterator{contract: _WithdrawalDelayer.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerDeposit, owner []common.Address, token []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "Deposit", ownerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerDeposit)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposit is a log parse operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseDeposit(log types.Log) (*WithdrawalDelayerDeposit, error) {
	event := new(WithdrawalDelayerDeposit)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerEmergencyModeEnabledIterator is returned from FilterEmergencyModeEnabled and is used to iterate over the raw logs and unpacked data for EmergencyModeEnabled events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerEmergencyModeEnabledIterator struct {
	Event *WithdrawalDelayerEmergencyModeEnabled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerEmergencyModeEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerEmergencyModeEnabled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerEmergencyModeEnabled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerEmergencyModeEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerEmergencyModeEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerEmergencyModeEnabled represents a EmergencyModeEnabled event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerEmergencyModeEnabled struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEmergencyModeEnabled is a free log retrieval operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterEmergencyModeEnabled(opts *bind.FilterOpts) (*WithdrawalDelayerEmergencyModeEnabledIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "EmergencyModeEnabled")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerEmergencyModeEnabledIterator{contract: _WithdrawalDelayer.contract, event: "EmergencyModeEnabled", logs: logs, sub: sub}, nil
}

// WatchEmergencyModeEnabled is a free log subscription operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchEmergencyModeEnabled(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerEmergencyModeEnabled) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "EmergencyModeEnabled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerEmergencyModeEnabled)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEmergencyModeEnabled is a log parse operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseEmergencyModeEnabled(log types.Log) (*WithdrawalDelayerEmergencyModeEnabled, error) {
	event := new(WithdrawalDelayerEmergencyModeEnabled)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerEscapeHatchWithdrawalIterator is returned from FilterEscapeHatchWithdrawal and is used to iterate over the raw logs and unpacked data for EscapeHatchWithdrawal events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerEscapeHatchWithdrawalIterator struct {
	Event *WithdrawalDelayerEscapeHatchWithdrawal // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerEscapeHatchWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerEscapeHatchWithdrawal)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerEscapeHatchWithdrawal)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerEscapeHatchWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerEscapeHatchWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerEscapeHatchWithdrawal represents a EscapeHatchWithdrawal event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerEscapeHatchWithdrawal struct {
	Who    common.Address
	To     common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEscapeHatchWithdrawal is a free log retrieval operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterEscapeHatchWithdrawal(opts *bind.FilterOpts, who []common.Address, to []common.Address, token []common.Address) (*WithdrawalDelayerEscapeHatchWithdrawalIterator, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "EscapeHatchWithdrawal", whoRule, toRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerEscapeHatchWithdrawalIterator{contract: _WithdrawalDelayer.contract, event: "EscapeHatchWithdrawal", logs: logs, sub: sub}, nil
}

// WatchEscapeHatchWithdrawal is a free log subscription operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchEscapeHatchWithdrawal(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerEscapeHatchWithdrawal, who []common.Address, to []common.Address, token []common.Address) (event.Subscription, error) {

	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "EscapeHatchWithdrawal", whoRule, toRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerEscapeHatchWithdrawal)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscapeHatchWithdrawal is a log parse operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseEscapeHatchWithdrawal(log types.Log) (*WithdrawalDelayerEscapeHatchWithdrawal, error) {
	event := new(WithdrawalDelayerEscapeHatchWithdrawal)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerInitializeWithdrawalDelayerEventIterator is returned from FilterInitializeWithdrawalDelayerEvent and is used to iterate over the raw logs and unpacked data for InitializeWithdrawalDelayerEvent events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerInitializeWithdrawalDelayerEventIterator struct {
	Event *WithdrawalDelayerInitializeWithdrawalDelayerEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerInitializeWithdrawalDelayerEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerInitializeWithdrawalDelayerEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerInitializeWithdrawalDelayerEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerInitializeWithdrawalDelayerEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerInitializeWithdrawalDelayerEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerInitializeWithdrawalDelayerEvent represents a InitializeWithdrawalDelayerEvent event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerInitializeWithdrawalDelayerEvent struct {
	InitialWithdrawalDelay         uint64
	InitialHermezGovernanceAddress common.Address
	InitialEmergencyCouncil        common.Address
	Raw                            types.Log // Blockchain specific contextual infos
}

// FilterInitializeWithdrawalDelayerEvent is a free log retrieval operation binding the contract event 0x8b81dca4c96ae06989fa8aa1baa4ccc05dfb42e0948c7d5b7505b68ccde41eec.
//
// Solidity: event InitializeWithdrawalDelayerEvent(uint64 initialWithdrawalDelay, address initialHermezGovernanceAddress, address initialEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterInitializeWithdrawalDelayerEvent(opts *bind.FilterOpts) (*WithdrawalDelayerInitializeWithdrawalDelayerEventIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "InitializeWithdrawalDelayerEvent")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerInitializeWithdrawalDelayerEventIterator{contract: _WithdrawalDelayer.contract, event: "InitializeWithdrawalDelayerEvent", logs: logs, sub: sub}, nil
}

// WatchInitializeWithdrawalDelayerEvent is a free log subscription operation binding the contract event 0x8b81dca4c96ae06989fa8aa1baa4ccc05dfb42e0948c7d5b7505b68ccde41eec.
//
// Solidity: event InitializeWithdrawalDelayerEvent(uint64 initialWithdrawalDelay, address initialHermezGovernanceAddress, address initialEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchInitializeWithdrawalDelayerEvent(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerInitializeWithdrawalDelayerEvent) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "InitializeWithdrawalDelayerEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerInitializeWithdrawalDelayerEvent)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "InitializeWithdrawalDelayerEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitializeWithdrawalDelayerEvent is a log parse operation binding the contract event 0x8b81dca4c96ae06989fa8aa1baa4ccc05dfb42e0948c7d5b7505b68ccde41eec.
//
// Solidity: event InitializeWithdrawalDelayerEvent(uint64 initialWithdrawalDelay, address initialHermezGovernanceAddress, address initialEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseInitializeWithdrawalDelayerEvent(log types.Log) (*WithdrawalDelayerInitializeWithdrawalDelayerEvent, error) {
	event := new(WithdrawalDelayerInitializeWithdrawalDelayerEvent)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "InitializeWithdrawalDelayerEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerNewEmergencyCouncilIterator is returned from FilterNewEmergencyCouncil and is used to iterate over the raw logs and unpacked data for NewEmergencyCouncil events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewEmergencyCouncilIterator struct {
	Event *WithdrawalDelayerNewEmergencyCouncil // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerNewEmergencyCouncilIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewEmergencyCouncil)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerNewEmergencyCouncil)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerNewEmergencyCouncilIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewEmergencyCouncilIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewEmergencyCouncil represents a NewEmergencyCouncil event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewEmergencyCouncil struct {
	NewEmergencyCouncil common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterNewEmergencyCouncil is a free log retrieval operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewEmergencyCouncil(opts *bind.FilterOpts) (*WithdrawalDelayerNewEmergencyCouncilIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewEmergencyCouncil")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewEmergencyCouncilIterator{contract: _WithdrawalDelayer.contract, event: "NewEmergencyCouncil", logs: logs, sub: sub}, nil
}

// WatchNewEmergencyCouncil is a free log subscription operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewEmergencyCouncil(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewEmergencyCouncil) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewEmergencyCouncil")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewEmergencyCouncil)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewEmergencyCouncil is a log parse operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewEmergencyCouncil(log types.Log) (*WithdrawalDelayerNewEmergencyCouncil, error) {
	event := new(WithdrawalDelayerNewEmergencyCouncil)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerNewHermezGovernanceAddressIterator is returned from FilterNewHermezGovernanceAddress and is used to iterate over the raw logs and unpacked data for NewHermezGovernanceAddress events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezGovernanceAddressIterator struct {
	Event *WithdrawalDelayerNewHermezGovernanceAddress // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerNewHermezGovernanceAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewHermezGovernanceAddress)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerNewHermezGovernanceAddress)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerNewHermezGovernanceAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewHermezGovernanceAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewHermezGovernanceAddress represents a NewHermezGovernanceAddress event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezGovernanceAddress struct {
	NewHermezGovernanceAddress common.Address
	Raw                        types.Log // Blockchain specific contextual infos
}

// FilterNewHermezGovernanceAddress is a free log retrieval operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewHermezGovernanceAddress(opts *bind.FilterOpts) (*WithdrawalDelayerNewHermezGovernanceAddressIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewHermezGovernanceAddress")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewHermezGovernanceAddressIterator{contract: _WithdrawalDelayer.contract, event: "NewHermezGovernanceAddress", logs: logs, sub: sub}, nil
}

// WatchNewHermezGovernanceAddress is a free log subscription operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewHermezGovernanceAddress(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewHermezGovernanceAddress) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewHermezGovernanceAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewHermezGovernanceAddress)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewHermezGovernanceAddress is a log parse operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewHermezGovernanceAddress(log types.Log) (*WithdrawalDelayerNewHermezGovernanceAddress, error) {
	event := new(WithdrawalDelayerNewHermezGovernanceAddress)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerNewWithdrawalDelayIterator is returned from FilterNewWithdrawalDelay and is used to iterate over the raw logs and unpacked data for NewWithdrawalDelay events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewWithdrawalDelayIterator struct {
	Event *WithdrawalDelayerNewWithdrawalDelay // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerNewWithdrawalDelayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewWithdrawalDelay)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerNewWithdrawalDelay)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerNewWithdrawalDelayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewWithdrawalDelayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewWithdrawalDelay represents a NewWithdrawalDelay event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewWithdrawalDelay struct {
	WithdrawalDelay uint64
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNewWithdrawalDelay is a free log retrieval operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewWithdrawalDelay(opts *bind.FilterOpts) (*WithdrawalDelayerNewWithdrawalDelayIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewWithdrawalDelayIterator{contract: _WithdrawalDelayer.contract, event: "NewWithdrawalDelay", logs: logs, sub: sub}, nil
}

// WatchNewWithdrawalDelay is a free log subscription operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewWithdrawalDelay(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewWithdrawalDelay) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewWithdrawalDelay)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewWithdrawalDelay is a log parse operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewWithdrawalDelay(log types.Log) (*WithdrawalDelayerNewWithdrawalDelay, error) {
	event := new(WithdrawalDelayerNewWithdrawalDelay)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawalDelayerWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerWithdrawIterator struct {
	Event *WithdrawalDelayerWithdraw // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *WithdrawalDelayerWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerWithdraw)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(WithdrawalDelayerWithdraw)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *WithdrawalDelayerWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerWithdraw represents a Withdraw event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerWithdraw struct {
	Token  common.Address
	Owner  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterWithdraw(opts *bind.FilterOpts, token []common.Address, owner []common.Address) (*WithdrawalDelayerWithdrawIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "Withdraw", tokenRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerWithdrawIterator{contract: _WithdrawalDelayer.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerWithdraw, token []common.Address, owner []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "Withdraw", tokenRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerWithdraw)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdraw is a log parse operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseWithdraw(log types.Log) (*WithdrawalDelayerWithdraw, error) {
	event := new(WithdrawalDelayerWithdraw)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
