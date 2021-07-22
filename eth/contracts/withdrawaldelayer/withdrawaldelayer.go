// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package withdrawaldelayer

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

// WithdrawaldelayerABI is the input ABI used to generate the binding from.
const WithdrawaldelayerABI = "[{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_initialWithdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_initialHermezRollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezGovernanceAddress\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_initialEmergencyCouncil\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EmergencyModeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EscapeHatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"initialWithdrawalDelay\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialHermezGovernanceAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"initialEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"InitializeWithdrawalDelayerEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"NewEmergencyCouncil\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezGovernanceAddress\",\"type\":\"address\"}],\"name\":\"NewHermezGovernanceAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"NewWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_EMERGENCY_MODE_TIME\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_WITHDRAWAL_DELAY\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"changeWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"_amount\",\"type\":\"uint192\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"depositInfo\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enableEmergencyMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"escapeHatchWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyCouncil\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyModeStartingTime\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezGovernanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWithdrawalDelay\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isEmergencyMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingEmergencyCouncil\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"transferEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newGovernance\",\"type\":\"address\"}],\"name\":\"transferGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Withdrawaldelayer is an auto generated Go binding around an Ethereum contract.
type Withdrawaldelayer struct {
	WithdrawaldelayerCaller     // Read-only binding to the contract
	WithdrawaldelayerTransactor // Write-only binding to the contract
	WithdrawaldelayerFilterer   // Log filterer for contract events
}

// WithdrawaldelayerCaller is an auto generated read-only Go binding around an Ethereum contract.
type WithdrawaldelayerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawaldelayerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type WithdrawaldelayerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawaldelayerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type WithdrawaldelayerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// WithdrawaldelayerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type WithdrawaldelayerSession struct {
	Contract     *Withdrawaldelayer // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// WithdrawaldelayerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type WithdrawaldelayerCallerSession struct {
	Contract *WithdrawaldelayerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// WithdrawaldelayerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type WithdrawaldelayerTransactorSession struct {
	Contract     *WithdrawaldelayerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// WithdrawaldelayerRaw is an auto generated low-level Go binding around an Ethereum contract.
type WithdrawaldelayerRaw struct {
	Contract *Withdrawaldelayer // Generic contract binding to access the raw methods on
}

// WithdrawaldelayerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type WithdrawaldelayerCallerRaw struct {
	Contract *WithdrawaldelayerCaller // Generic read-only contract binding to access the raw methods on
}

// WithdrawaldelayerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type WithdrawaldelayerTransactorRaw struct {
	Contract *WithdrawaldelayerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewWithdrawaldelayer creates a new instance of Withdrawaldelayer, bound to a specific deployed contract.
func NewWithdrawaldelayer(address common.Address, backend bind.ContractBackend) (*Withdrawaldelayer, error) {
	contract, err := bindWithdrawaldelayer(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Withdrawaldelayer{WithdrawaldelayerCaller: WithdrawaldelayerCaller{contract: contract}, WithdrawaldelayerTransactor: WithdrawaldelayerTransactor{contract: contract}, WithdrawaldelayerFilterer: WithdrawaldelayerFilterer{contract: contract}}, nil
}

// NewWithdrawaldelayerCaller creates a new read-only instance of Withdrawaldelayer, bound to a specific deployed contract.
func NewWithdrawaldelayerCaller(address common.Address, caller bind.ContractCaller) (*WithdrawaldelayerCaller, error) {
	contract, err := bindWithdrawaldelayer(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerCaller{contract: contract}, nil
}

// NewWithdrawaldelayerTransactor creates a new write-only instance of Withdrawaldelayer, bound to a specific deployed contract.
func NewWithdrawaldelayerTransactor(address common.Address, transactor bind.ContractTransactor) (*WithdrawaldelayerTransactor, error) {
	contract, err := bindWithdrawaldelayer(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerTransactor{contract: contract}, nil
}

// NewWithdrawaldelayerFilterer creates a new log filterer instance of Withdrawaldelayer, bound to a specific deployed contract.
func NewWithdrawaldelayerFilterer(address common.Address, filterer bind.ContractFilterer) (*WithdrawaldelayerFilterer, error) {
	contract, err := bindWithdrawaldelayer(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerFilterer{contract: contract}, nil
}

// bindWithdrawaldelayer binds a generic wrapper to an already deployed contract.
func bindWithdrawaldelayer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WithdrawaldelayerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Withdrawaldelayer *WithdrawaldelayerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Withdrawaldelayer.Contract.WithdrawaldelayerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Withdrawaldelayer *WithdrawaldelayerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.WithdrawaldelayerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Withdrawaldelayer *WithdrawaldelayerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.WithdrawaldelayerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Withdrawaldelayer *WithdrawaldelayerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Withdrawaldelayer.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Withdrawaldelayer *WithdrawaldelayerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Withdrawaldelayer *WithdrawaldelayerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.contract.Transact(opts, method, params...)
}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) MAXEMERGENCYMODETIME(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "MAX_EMERGENCY_MODE_TIME")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerSession) MAXEMERGENCYMODETIME() (uint64, error) {
	return _Withdrawaldelayer.Contract.MAXEMERGENCYMODETIME(&_Withdrawaldelayer.CallOpts)
}

// MAXEMERGENCYMODETIME is a free data retrieval call binding the contract method 0xb4b8e39d.
//
// Solidity: function MAX_EMERGENCY_MODE_TIME() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) MAXEMERGENCYMODETIME() (uint64, error) {
	return _Withdrawaldelayer.Contract.MAXEMERGENCYMODETIME(&_Withdrawaldelayer.CallOpts)
}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) MAXWITHDRAWALDELAY(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "MAX_WITHDRAWAL_DELAY")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerSession) MAXWITHDRAWALDELAY() (uint64, error) {
	return _Withdrawaldelayer.Contract.MAXWITHDRAWALDELAY(&_Withdrawaldelayer.CallOpts)
}

// MAXWITHDRAWALDELAY is a free data retrieval call binding the contract method 0xa238f9df.
//
// Solidity: function MAX_WITHDRAWAL_DELAY() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) MAXWITHDRAWALDELAY() (uint64, error) {
	return _Withdrawaldelayer.Contract.MAXWITHDRAWALDELAY(&_Withdrawaldelayer.CallOpts)
}

// DepositInfo is a free data retrieval call binding the contract method 0x493b0170.
//
// Solidity: function depositInfo(address _owner, address _token) view returns(uint192, uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) DepositInfo(opts *bind.CallOpts, _owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "depositInfo", _owner, _token)

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
func (_Withdrawaldelayer *WithdrawaldelayerSession) DepositInfo(_owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	return _Withdrawaldelayer.Contract.DepositInfo(&_Withdrawaldelayer.CallOpts, _owner, _token)
}

// DepositInfo is a free data retrieval call binding the contract method 0x493b0170.
//
// Solidity: function depositInfo(address _owner, address _token) view returns(uint192, uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) DepositInfo(_owner common.Address, _token common.Address) (*big.Int, uint64, error) {
	return _Withdrawaldelayer.Contract.DepositInfo(&_Withdrawaldelayer.CallOpts, _owner, _token)
}

// Deposits is a free data retrieval call binding the contract method 0x3d4dff7b.
//
// Solidity: function deposits(bytes32 ) view returns(uint192 amount, uint64 depositTimestamp)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) Deposits(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "deposits", arg0)

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
func (_Withdrawaldelayer *WithdrawaldelayerSession) Deposits(arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	return _Withdrawaldelayer.Contract.Deposits(&_Withdrawaldelayer.CallOpts, arg0)
}

// Deposits is a free data retrieval call binding the contract method 0x3d4dff7b.
//
// Solidity: function deposits(bytes32 ) view returns(uint192 amount, uint64 depositTimestamp)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) Deposits(arg0 [32]byte) (struct {
	Amount           *big.Int
	DepositTimestamp uint64
}, error) {
	return _Withdrawaldelayer.Contract.Deposits(&_Withdrawaldelayer.CallOpts, arg0)
}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) GetEmergencyCouncil(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "getEmergencyCouncil")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerSession) GetEmergencyCouncil() (common.Address, error) {
	return _Withdrawaldelayer.Contract.GetEmergencyCouncil(&_Withdrawaldelayer.CallOpts)
}

// GetEmergencyCouncil is a free data retrieval call binding the contract method 0x99ef11c5.
//
// Solidity: function getEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) GetEmergencyCouncil() (common.Address, error) {
	return _Withdrawaldelayer.Contract.GetEmergencyCouncil(&_Withdrawaldelayer.CallOpts)
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) GetEmergencyModeStartingTime(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "getEmergencyModeStartingTime")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerSession) GetEmergencyModeStartingTime() (uint64, error) {
	return _Withdrawaldelayer.Contract.GetEmergencyModeStartingTime(&_Withdrawaldelayer.CallOpts)
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) GetEmergencyModeStartingTime() (uint64, error) {
	return _Withdrawaldelayer.Contract.GetEmergencyModeStartingTime(&_Withdrawaldelayer.CallOpts)
}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) GetHermezGovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "getHermezGovernanceAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerSession) GetHermezGovernanceAddress() (common.Address, error) {
	return _Withdrawaldelayer.Contract.GetHermezGovernanceAddress(&_Withdrawaldelayer.CallOpts)
}

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) GetHermezGovernanceAddress() (common.Address, error) {
	return _Withdrawaldelayer.Contract.GetHermezGovernanceAddress(&_Withdrawaldelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) GetWithdrawalDelay(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "getWithdrawalDelay")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerSession) GetWithdrawalDelay() (uint64, error) {
	return _Withdrawaldelayer.Contract.GetWithdrawalDelay(&_Withdrawaldelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint64)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) GetWithdrawalDelay() (uint64, error) {
	return _Withdrawaldelayer.Contract.GetWithdrawalDelay(&_Withdrawaldelayer.CallOpts)
}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) HermezRollupAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "hermezRollupAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerSession) HermezRollupAddress() (common.Address, error) {
	return _Withdrawaldelayer.Contract.HermezRollupAddress(&_Withdrawaldelayer.CallOpts)
}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) HermezRollupAddress() (common.Address, error) {
	return _Withdrawaldelayer.Contract.HermezRollupAddress(&_Withdrawaldelayer.CallOpts)
}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) IsEmergencyMode(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "isEmergencyMode")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_Withdrawaldelayer *WithdrawaldelayerSession) IsEmergencyMode() (bool, error) {
	return _Withdrawaldelayer.Contract.IsEmergencyMode(&_Withdrawaldelayer.CallOpts)
}

// IsEmergencyMode is a free data retrieval call binding the contract method 0x20a194b8.
//
// Solidity: function isEmergencyMode() view returns(bool)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) IsEmergencyMode() (bool, error) {
	return _Withdrawaldelayer.Contract.IsEmergencyMode(&_Withdrawaldelayer.CallOpts)
}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) PendingEmergencyCouncil(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "pendingEmergencyCouncil")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerSession) PendingEmergencyCouncil() (common.Address, error) {
	return _Withdrawaldelayer.Contract.PendingEmergencyCouncil(&_Withdrawaldelayer.CallOpts)
}

// PendingEmergencyCouncil is a free data retrieval call binding the contract method 0x67fa2403.
//
// Solidity: function pendingEmergencyCouncil() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) PendingEmergencyCouncil() (common.Address, error) {
	return _Withdrawaldelayer.Contract.PendingEmergencyCouncil(&_Withdrawaldelayer.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCaller) PendingGovernance(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Withdrawaldelayer.contract.Call(opts, &out, "pendingGovernance")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerSession) PendingGovernance() (common.Address, error) {
	return _Withdrawaldelayer.Contract.PendingGovernance(&_Withdrawaldelayer.CallOpts)
}

// PendingGovernance is a free data retrieval call binding the contract method 0xf39c38a0.
//
// Solidity: function pendingGovernance() view returns(address)
func (_Withdrawaldelayer *WithdrawaldelayerCallerSession) PendingGovernance() (common.Address, error) {
	return _Withdrawaldelayer.Contract.PendingGovernance(&_Withdrawaldelayer.CallOpts)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) ChangeWithdrawalDelay(opts *bind.TransactOpts, _newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "changeWithdrawalDelay", _newWithdrawalDelay)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) ChangeWithdrawalDelay(_newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ChangeWithdrawalDelay(&_Withdrawaldelayer.TransactOpts, _newWithdrawalDelay)
}

// ChangeWithdrawalDelay is a paid mutator transaction binding the contract method 0x0e670af5.
//
// Solidity: function changeWithdrawalDelay(uint64 _newWithdrawalDelay) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) ChangeWithdrawalDelay(_newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ChangeWithdrawalDelay(&_Withdrawaldelayer.TransactOpts, _newWithdrawalDelay)
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) ClaimEmergencyCouncil(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "claimEmergencyCouncil")
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) ClaimEmergencyCouncil() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ClaimEmergencyCouncil(&_Withdrawaldelayer.TransactOpts)
}

// ClaimEmergencyCouncil is a paid mutator transaction binding the contract method 0xca79033f.
//
// Solidity: function claimEmergencyCouncil() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) ClaimEmergencyCouncil() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ClaimEmergencyCouncil(&_Withdrawaldelayer.TransactOpts)
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) ClaimGovernance(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "claimGovernance")
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) ClaimGovernance() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ClaimGovernance(&_Withdrawaldelayer.TransactOpts)
}

// ClaimGovernance is a paid mutator transaction binding the contract method 0x5d36b190.
//
// Solidity: function claimGovernance() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) ClaimGovernance() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.ClaimGovernance(&_Withdrawaldelayer.TransactOpts)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) Deposit(opts *bind.TransactOpts, _owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "deposit", _owner, _token, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) Deposit(_owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.Deposit(&_Withdrawaldelayer.TransactOpts, _owner, _token, _amount)
}

// Deposit is a paid mutator transaction binding the contract method 0xcfc0b641.
//
// Solidity: function deposit(address _owner, address _token, uint192 _amount) payable returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) Deposit(_owner common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.Deposit(&_Withdrawaldelayer.TransactOpts, _owner, _token, _amount)
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) EnableEmergencyMode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "enableEmergencyMode")
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) EnableEmergencyMode() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.EnableEmergencyMode(&_Withdrawaldelayer.TransactOpts)
}

// EnableEmergencyMode is a paid mutator transaction binding the contract method 0xc5b1c7d0.
//
// Solidity: function enableEmergencyMode() returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) EnableEmergencyMode() (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.EnableEmergencyMode(&_Withdrawaldelayer.TransactOpts)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) EscapeHatchWithdrawal(opts *bind.TransactOpts, _to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "escapeHatchWithdrawal", _to, _token, _amount)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.EscapeHatchWithdrawal(&_Withdrawaldelayer.TransactOpts, _to, _token, _amount)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0x7fd6b102.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token, uint256 _amount) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.EscapeHatchWithdrawal(&_Withdrawaldelayer.TransactOpts, _to, _token, _amount)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) TransferEmergencyCouncil(opts *bind.TransactOpts, newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "transferEmergencyCouncil", newEmergencyCouncil)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) TransferEmergencyCouncil(newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.TransferEmergencyCouncil(&_Withdrawaldelayer.TransactOpts, newEmergencyCouncil)
}

// TransferEmergencyCouncil is a paid mutator transaction binding the contract method 0xdb2a1a81.
//
// Solidity: function transferEmergencyCouncil(address newEmergencyCouncil) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) TransferEmergencyCouncil(newEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.TransferEmergencyCouncil(&_Withdrawaldelayer.TransactOpts, newEmergencyCouncil)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) TransferGovernance(opts *bind.TransactOpts, newGovernance common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "transferGovernance", newGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) TransferGovernance(newGovernance common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.TransferGovernance(&_Withdrawaldelayer.TransactOpts, newGovernance)
}

// TransferGovernance is a paid mutator transaction binding the contract method 0xd38bfff4.
//
// Solidity: function transferGovernance(address newGovernance) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) TransferGovernance(newGovernance common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.TransferGovernance(&_Withdrawaldelayer.TransactOpts, newGovernance)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactor) Withdrawal(opts *bind.TransactOpts, _owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.contract.Transact(opts, "withdrawal", _owner, _token)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_Withdrawaldelayer *WithdrawaldelayerSession) Withdrawal(_owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.Withdrawal(&_Withdrawaldelayer.TransactOpts, _owner, _token)
}

// Withdrawal is a paid mutator transaction binding the contract method 0xde35f282.
//
// Solidity: function withdrawal(address _owner, address _token) returns()
func (_Withdrawaldelayer *WithdrawaldelayerTransactorSession) Withdrawal(_owner common.Address, _token common.Address) (*types.Transaction, error) {
	return _Withdrawaldelayer.Contract.Withdrawal(&_Withdrawaldelayer.TransactOpts, _owner, _token)
}

// WithdrawaldelayerDepositIterator is returned from FilterDeposit and is used to iterate over the raw logs and unpacked data for Deposit events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerDepositIterator struct {
	Event *WithdrawaldelayerDeposit // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerDepositIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerDeposit)
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
		it.Event = new(WithdrawaldelayerDeposit)
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
func (it *WithdrawaldelayerDepositIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerDepositIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerDeposit represents a Deposit event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerDeposit struct {
	Owner            common.Address
	Token            common.Address
	Amount           *big.Int
	DepositTimestamp uint64
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterDeposit is a free log retrieval operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterDeposit(opts *bind.FilterOpts, owner []common.Address, token []common.Address) (*WithdrawaldelayerDepositIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "Deposit", ownerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerDepositIterator{contract: _Withdrawaldelayer.contract, event: "Deposit", logs: logs, sub: sub}, nil
}

// WatchDeposit is a free log subscription operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchDeposit(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerDeposit, owner []common.Address, token []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "Deposit", ownerRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerDeposit)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseDeposit(log types.Log) (*WithdrawaldelayerDeposit, error) {
	event := new(WithdrawaldelayerDeposit)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerEmergencyModeEnabledIterator is returned from FilterEmergencyModeEnabled and is used to iterate over the raw logs and unpacked data for EmergencyModeEnabled events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerEmergencyModeEnabledIterator struct {
	Event *WithdrawaldelayerEmergencyModeEnabled // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerEmergencyModeEnabledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerEmergencyModeEnabled)
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
		it.Event = new(WithdrawaldelayerEmergencyModeEnabled)
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
func (it *WithdrawaldelayerEmergencyModeEnabledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerEmergencyModeEnabledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerEmergencyModeEnabled represents a EmergencyModeEnabled event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerEmergencyModeEnabled struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEmergencyModeEnabled is a free log retrieval operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterEmergencyModeEnabled(opts *bind.FilterOpts) (*WithdrawaldelayerEmergencyModeEnabledIterator, error) {

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "EmergencyModeEnabled")
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerEmergencyModeEnabledIterator{contract: _Withdrawaldelayer.contract, event: "EmergencyModeEnabled", logs: logs, sub: sub}, nil
}

// WatchEmergencyModeEnabled is a free log subscription operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchEmergencyModeEnabled(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerEmergencyModeEnabled) (event.Subscription, error) {

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "EmergencyModeEnabled")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerEmergencyModeEnabled)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseEmergencyModeEnabled(log types.Log) (*WithdrawaldelayerEmergencyModeEnabled, error) {
	event := new(WithdrawaldelayerEmergencyModeEnabled)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerEscapeHatchWithdrawalIterator is returned from FilterEscapeHatchWithdrawal and is used to iterate over the raw logs and unpacked data for EscapeHatchWithdrawal events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerEscapeHatchWithdrawalIterator struct {
	Event *WithdrawaldelayerEscapeHatchWithdrawal // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerEscapeHatchWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerEscapeHatchWithdrawal)
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
		it.Event = new(WithdrawaldelayerEscapeHatchWithdrawal)
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
func (it *WithdrawaldelayerEscapeHatchWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerEscapeHatchWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerEscapeHatchWithdrawal represents a EscapeHatchWithdrawal event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerEscapeHatchWithdrawal struct {
	Who    common.Address
	To     common.Address
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterEscapeHatchWithdrawal is a free log retrieval operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterEscapeHatchWithdrawal(opts *bind.FilterOpts, who []common.Address, to []common.Address, token []common.Address) (*WithdrawaldelayerEscapeHatchWithdrawalIterator, error) {

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

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "EscapeHatchWithdrawal", whoRule, toRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerEscapeHatchWithdrawalIterator{contract: _Withdrawaldelayer.contract, event: "EscapeHatchWithdrawal", logs: logs, sub: sub}, nil
}

// WatchEscapeHatchWithdrawal is a free log subscription operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchEscapeHatchWithdrawal(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerEscapeHatchWithdrawal, who []common.Address, to []common.Address, token []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "EscapeHatchWithdrawal", whoRule, toRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerEscapeHatchWithdrawal)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseEscapeHatchWithdrawal(log types.Log) (*WithdrawaldelayerEscapeHatchWithdrawal, error) {
	event := new(WithdrawaldelayerEscapeHatchWithdrawal)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerInitializeWithdrawalDelayerEventIterator is returned from FilterInitializeWithdrawalDelayerEvent and is used to iterate over the raw logs and unpacked data for InitializeWithdrawalDelayerEvent events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerInitializeWithdrawalDelayerEventIterator struct {
	Event *WithdrawaldelayerInitializeWithdrawalDelayerEvent // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerInitializeWithdrawalDelayerEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerInitializeWithdrawalDelayerEvent)
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
		it.Event = new(WithdrawaldelayerInitializeWithdrawalDelayerEvent)
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
func (it *WithdrawaldelayerInitializeWithdrawalDelayerEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerInitializeWithdrawalDelayerEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerInitializeWithdrawalDelayerEvent represents a InitializeWithdrawalDelayerEvent event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerInitializeWithdrawalDelayerEvent struct {
	InitialWithdrawalDelay         uint64
	InitialHermezGovernanceAddress common.Address
	InitialEmergencyCouncil        common.Address
	Raw                            types.Log // Blockchain specific contextual infos
}

// FilterInitializeWithdrawalDelayerEvent is a free log retrieval operation binding the contract event 0x8b81dca4c96ae06989fa8aa1baa4ccc05dfb42e0948c7d5b7505b68ccde41eec.
//
// Solidity: event InitializeWithdrawalDelayerEvent(uint64 initialWithdrawalDelay, address initialHermezGovernanceAddress, address initialEmergencyCouncil)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterInitializeWithdrawalDelayerEvent(opts *bind.FilterOpts) (*WithdrawaldelayerInitializeWithdrawalDelayerEventIterator, error) {

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "InitializeWithdrawalDelayerEvent")
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerInitializeWithdrawalDelayerEventIterator{contract: _Withdrawaldelayer.contract, event: "InitializeWithdrawalDelayerEvent", logs: logs, sub: sub}, nil
}

// WatchInitializeWithdrawalDelayerEvent is a free log subscription operation binding the contract event 0x8b81dca4c96ae06989fa8aa1baa4ccc05dfb42e0948c7d5b7505b68ccde41eec.
//
// Solidity: event InitializeWithdrawalDelayerEvent(uint64 initialWithdrawalDelay, address initialHermezGovernanceAddress, address initialEmergencyCouncil)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchInitializeWithdrawalDelayerEvent(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerInitializeWithdrawalDelayerEvent) (event.Subscription, error) {

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "InitializeWithdrawalDelayerEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerInitializeWithdrawalDelayerEvent)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "InitializeWithdrawalDelayerEvent", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseInitializeWithdrawalDelayerEvent(log types.Log) (*WithdrawaldelayerInitializeWithdrawalDelayerEvent, error) {
	event := new(WithdrawaldelayerInitializeWithdrawalDelayerEvent)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "InitializeWithdrawalDelayerEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerNewEmergencyCouncilIterator is returned from FilterNewEmergencyCouncil and is used to iterate over the raw logs and unpacked data for NewEmergencyCouncil events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewEmergencyCouncilIterator struct {
	Event *WithdrawaldelayerNewEmergencyCouncil // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerNewEmergencyCouncilIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerNewEmergencyCouncil)
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
		it.Event = new(WithdrawaldelayerNewEmergencyCouncil)
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
func (it *WithdrawaldelayerNewEmergencyCouncilIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerNewEmergencyCouncilIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerNewEmergencyCouncil represents a NewEmergencyCouncil event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewEmergencyCouncil struct {
	NewEmergencyCouncil common.Address
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterNewEmergencyCouncil is a free log retrieval operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterNewEmergencyCouncil(opts *bind.FilterOpts) (*WithdrawaldelayerNewEmergencyCouncilIterator, error) {

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "NewEmergencyCouncil")
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerNewEmergencyCouncilIterator{contract: _Withdrawaldelayer.contract, event: "NewEmergencyCouncil", logs: logs, sub: sub}, nil
}

// WatchNewEmergencyCouncil is a free log subscription operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchNewEmergencyCouncil(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerNewEmergencyCouncil) (event.Subscription, error) {

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "NewEmergencyCouncil")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerNewEmergencyCouncil)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseNewEmergencyCouncil(log types.Log) (*WithdrawaldelayerNewEmergencyCouncil, error) {
	event := new(WithdrawaldelayerNewEmergencyCouncil)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerNewHermezGovernanceAddressIterator is returned from FilterNewHermezGovernanceAddress and is used to iterate over the raw logs and unpacked data for NewHermezGovernanceAddress events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewHermezGovernanceAddressIterator struct {
	Event *WithdrawaldelayerNewHermezGovernanceAddress // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerNewHermezGovernanceAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerNewHermezGovernanceAddress)
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
		it.Event = new(WithdrawaldelayerNewHermezGovernanceAddress)
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
func (it *WithdrawaldelayerNewHermezGovernanceAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerNewHermezGovernanceAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerNewHermezGovernanceAddress represents a NewHermezGovernanceAddress event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewHermezGovernanceAddress struct {
	NewHermezGovernanceAddress common.Address
	Raw                        types.Log // Blockchain specific contextual infos
}

// FilterNewHermezGovernanceAddress is a free log retrieval operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterNewHermezGovernanceAddress(opts *bind.FilterOpts) (*WithdrawaldelayerNewHermezGovernanceAddressIterator, error) {

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "NewHermezGovernanceAddress")
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerNewHermezGovernanceAddressIterator{contract: _Withdrawaldelayer.contract, event: "NewHermezGovernanceAddress", logs: logs, sub: sub}, nil
}

// WatchNewHermezGovernanceAddress is a free log subscription operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchNewHermezGovernanceAddress(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerNewHermezGovernanceAddress) (event.Subscription, error) {

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "NewHermezGovernanceAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerNewHermezGovernanceAddress)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseNewHermezGovernanceAddress(log types.Log) (*WithdrawaldelayerNewHermezGovernanceAddress, error) {
	event := new(WithdrawaldelayerNewHermezGovernanceAddress)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerNewWithdrawalDelayIterator is returned from FilterNewWithdrawalDelay and is used to iterate over the raw logs and unpacked data for NewWithdrawalDelay events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewWithdrawalDelayIterator struct {
	Event *WithdrawaldelayerNewWithdrawalDelay // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerNewWithdrawalDelayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerNewWithdrawalDelay)
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
		it.Event = new(WithdrawaldelayerNewWithdrawalDelay)
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
func (it *WithdrawaldelayerNewWithdrawalDelayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerNewWithdrawalDelayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerNewWithdrawalDelay represents a NewWithdrawalDelay event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerNewWithdrawalDelay struct {
	WithdrawalDelay uint64
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNewWithdrawalDelay is a free log retrieval operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterNewWithdrawalDelay(opts *bind.FilterOpts) (*WithdrawaldelayerNewWithdrawalDelayIterator, error) {

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "NewWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerNewWithdrawalDelayIterator{contract: _Withdrawaldelayer.contract, event: "NewWithdrawalDelay", logs: logs, sub: sub}, nil
}

// WatchNewWithdrawalDelay is a free log subscription operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchNewWithdrawalDelay(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerNewWithdrawalDelay) (event.Subscription, error) {

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "NewWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerNewWithdrawalDelay)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseNewWithdrawalDelay(log types.Log) (*WithdrawaldelayerNewWithdrawalDelay, error) {
	event := new(WithdrawaldelayerNewWithdrawalDelay)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// WithdrawaldelayerWithdrawIterator is returned from FilterWithdraw and is used to iterate over the raw logs and unpacked data for Withdraw events raised by the Withdrawaldelayer contract.
type WithdrawaldelayerWithdrawIterator struct {
	Event *WithdrawaldelayerWithdraw // Event containing the contract specifics and raw log

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
func (it *WithdrawaldelayerWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawaldelayerWithdraw)
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
		it.Event = new(WithdrawaldelayerWithdraw)
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
func (it *WithdrawaldelayerWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawaldelayerWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawaldelayerWithdraw represents a Withdraw event raised by the Withdrawaldelayer contract.
type WithdrawaldelayerWithdraw struct {
	Token  common.Address
	Owner  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdraw is a free log retrieval operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) FilterWithdraw(opts *bind.FilterOpts, token []common.Address, owner []common.Address) (*WithdrawaldelayerWithdrawIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Withdrawaldelayer.contract.FilterLogs(opts, "Withdraw", tokenRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return &WithdrawaldelayerWithdrawIterator{contract: _Withdrawaldelayer.contract, event: "Withdraw", logs: logs, sub: sub}, nil
}

// WatchWithdraw is a free log subscription operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) WatchWithdraw(opts *bind.WatchOpts, sink chan<- *WithdrawaldelayerWithdraw, token []common.Address, owner []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Withdrawaldelayer.contract.WatchLogs(opts, "Withdraw", tokenRule, ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawaldelayerWithdraw)
				if err := _Withdrawaldelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
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
func (_Withdrawaldelayer *WithdrawaldelayerFilterer) ParseWithdraw(log types.Log) (*WithdrawaldelayerWithdraw, error) {
	event := new(WithdrawaldelayerWithdraw)
	if err := _Withdrawaldelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
