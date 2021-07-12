// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package HEZ

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

// HEZABI is the input ABI used to generate the binding from.
const HEZABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialHolder\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"authorizer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"}],\"name\":\"AuthorizationUsed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"EIP712DOMAIN_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NAME_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PERMIT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFER_WITH_AUTHORIZATION_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"authorizationState\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"defaultMintAmount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"permit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validAfter\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validBefore\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"transferWithAuthorization\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// HEZ is an auto generated Go binding around an Ethereum contract.
type HEZ struct {
	HEZCaller     // Read-only binding to the contract
	HEZTransactor // Write-only binding to the contract
	HEZFilterer   // Log filterer for contract events
}

// HEZCaller is an auto generated read-only Go binding around an Ethereum contract.
type HEZCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HEZTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HEZTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HEZFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HEZFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HEZSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HEZSession struct {
	Contract     *HEZ              // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HEZCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HEZCallerSession struct {
	Contract *HEZCaller    // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// HEZTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HEZTransactorSession struct {
	Contract     *HEZTransactor    // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HEZRaw is an auto generated low-level Go binding around an Ethereum contract.
type HEZRaw struct {
	Contract *HEZ // Generic contract binding to access the raw methods on
}

// HEZCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HEZCallerRaw struct {
	Contract *HEZCaller // Generic read-only contract binding to access the raw methods on
}

// HEZTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HEZTransactorRaw struct {
	Contract *HEZTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHEZ creates a new instance of HEZ, bound to a specific deployed contract.
func NewHEZ(address common.Address, backend bind.ContractBackend) (*HEZ, error) {
	contract, err := bindHEZ(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HEZ{HEZCaller: HEZCaller{contract: contract}, HEZTransactor: HEZTransactor{contract: contract}, HEZFilterer: HEZFilterer{contract: contract}}, nil
}

// NewHEZCaller creates a new read-only instance of HEZ, bound to a specific deployed contract.
func NewHEZCaller(address common.Address, caller bind.ContractCaller) (*HEZCaller, error) {
	contract, err := bindHEZ(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HEZCaller{contract: contract}, nil
}

// NewHEZTransactor creates a new write-only instance of HEZ, bound to a specific deployed contract.
func NewHEZTransactor(address common.Address, transactor bind.ContractTransactor) (*HEZTransactor, error) {
	contract, err := bindHEZ(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HEZTransactor{contract: contract}, nil
}

// NewHEZFilterer creates a new log filterer instance of HEZ, bound to a specific deployed contract.
func NewHEZFilterer(address common.Address, filterer bind.ContractFilterer) (*HEZFilterer, error) {
	contract, err := bindHEZ(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HEZFilterer{contract: contract}, nil
}

// bindHEZ binds a generic wrapper to an already deployed contract.
func bindHEZ(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HEZABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HEZ *HEZRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HEZ.Contract.HEZCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HEZ *HEZRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HEZ.Contract.HEZTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HEZ *HEZRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HEZ.Contract.HEZTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HEZ *HEZCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HEZ.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HEZ *HEZTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HEZ.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HEZ *HEZTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HEZ.Contract.contract.Transact(opts, method, params...)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_HEZ *HEZCaller) EIP712DOMAINHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "EIP712DOMAIN_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_HEZ *HEZSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _HEZ.Contract.EIP712DOMAINHASH(&_HEZ.CallOpts)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_HEZ *HEZCallerSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _HEZ.Contract.EIP712DOMAINHASH(&_HEZ.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_HEZ *HEZCaller) NAMEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "NAME_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_HEZ *HEZSession) NAMEHASH() ([32]byte, error) {
	return _HEZ.Contract.NAMEHASH(&_HEZ.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_HEZ *HEZCallerSession) NAMEHASH() ([32]byte, error) {
	return _HEZ.Contract.NAMEHASH(&_HEZ.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZCaller) PERMITTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "PERMIT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZSession) PERMITTYPEHASH() ([32]byte, error) {
	return _HEZ.Contract.PERMITTYPEHASH(&_HEZ.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZCallerSession) PERMITTYPEHASH() ([32]byte, error) {
	return _HEZ.Contract.PERMITTYPEHASH(&_HEZ.CallOpts)
}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZCaller) TRANSFERWITHAUTHORIZATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "TRANSFER_WITH_AUTHORIZATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZSession) TRANSFERWITHAUTHORIZATIONTYPEHASH() ([32]byte, error) {
	return _HEZ.Contract.TRANSFERWITHAUTHORIZATIONTYPEHASH(&_HEZ.CallOpts)
}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_HEZ *HEZCallerSession) TRANSFERWITHAUTHORIZATIONTYPEHASH() ([32]byte, error) {
	return _HEZ.Contract.TRANSFERWITHAUTHORIZATIONTYPEHASH(&_HEZ.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_HEZ *HEZCaller) VERSIONHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "VERSION_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_HEZ *HEZSession) VERSIONHASH() ([32]byte, error) {
	return _HEZ.Contract.VERSIONHASH(&_HEZ.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_HEZ *HEZCallerSession) VERSIONHASH() ([32]byte, error) {
	return _HEZ.Contract.VERSIONHASH(&_HEZ.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_HEZ *HEZCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_HEZ *HEZSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _HEZ.Contract.Allowance(&_HEZ.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_HEZ *HEZCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _HEZ.Contract.Allowance(&_HEZ.CallOpts, arg0, arg1)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_HEZ *HEZCaller) AuthorizationState(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (bool, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "authorizationState", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_HEZ *HEZSession) AuthorizationState(arg0 common.Address, arg1 [32]byte) (bool, error) {
	return _HEZ.Contract.AuthorizationState(&_HEZ.CallOpts, arg0, arg1)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_HEZ *HEZCallerSession) AuthorizationState(arg0 common.Address, arg1 [32]byte) (bool, error) {
	return _HEZ.Contract.AuthorizationState(&_HEZ.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_HEZ *HEZCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_HEZ *HEZSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _HEZ.Contract.BalanceOf(&_HEZ.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_HEZ *HEZCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _HEZ.Contract.BalanceOf(&_HEZ.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_HEZ *HEZCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_HEZ *HEZSession) Decimals() (uint8, error) {
	return _HEZ.Contract.Decimals(&_HEZ.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_HEZ *HEZCallerSession) Decimals() (uint8, error) {
	return _HEZ.Contract.Decimals(&_HEZ.CallOpts)
}

// DefaultMintAmount is a free data retrieval call binding the contract method 0x05b23f2f.
//
// Solidity: function defaultMintAmount() view returns(uint256)
func (_HEZ *HEZCaller) DefaultMintAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "defaultMintAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DefaultMintAmount is a free data retrieval call binding the contract method 0x05b23f2f.
//
// Solidity: function defaultMintAmount() view returns(uint256)
func (_HEZ *HEZSession) DefaultMintAmount() (*big.Int, error) {
	return _HEZ.Contract.DefaultMintAmount(&_HEZ.CallOpts)
}

// DefaultMintAmount is a free data retrieval call binding the contract method 0x05b23f2f.
//
// Solidity: function defaultMintAmount() view returns(uint256)
func (_HEZ *HEZCallerSession) DefaultMintAmount() (*big.Int, error) {
	return _HEZ.Contract.DefaultMintAmount(&_HEZ.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_HEZ *HEZCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_HEZ *HEZSession) GetChainId() (*big.Int, error) {
	return _HEZ.Contract.GetChainId(&_HEZ.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_HEZ *HEZCallerSession) GetChainId() (*big.Int, error) {
	return _HEZ.Contract.GetChainId(&_HEZ.CallOpts)
}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_HEZ *HEZCaller) InitialBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "initialBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_HEZ *HEZSession) InitialBalance() (*big.Int, error) {
	return _HEZ.Contract.InitialBalance(&_HEZ.CallOpts)
}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_HEZ *HEZCallerSession) InitialBalance() (*big.Int, error) {
	return _HEZ.Contract.InitialBalance(&_HEZ.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_HEZ *HEZCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_HEZ *HEZSession) Name() (string, error) {
	return _HEZ.Contract.Name(&_HEZ.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_HEZ *HEZCallerSession) Name() (string, error) {
	return _HEZ.Contract.Name(&_HEZ.CallOpts)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_HEZ *HEZCaller) Nonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "nonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_HEZ *HEZSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _HEZ.Contract.Nonces(&_HEZ.CallOpts, arg0)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_HEZ *HEZCallerSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _HEZ.Contract.Nonces(&_HEZ.CallOpts, arg0)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_HEZ *HEZCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_HEZ *HEZSession) Symbol() (string, error) {
	return _HEZ.Contract.Symbol(&_HEZ.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_HEZ *HEZCallerSession) Symbol() (string, error) {
	return _HEZ.Contract.Symbol(&_HEZ.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_HEZ *HEZCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HEZ.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_HEZ *HEZSession) TotalSupply() (*big.Int, error) {
	return _HEZ.Contract.TotalSupply(&_HEZ.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_HEZ *HEZCallerSession) TotalSupply() (*big.Int, error) {
	return _HEZ.Contract.TotalSupply(&_HEZ.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_HEZ *HEZTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_HEZ *HEZSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Approve(&_HEZ.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_HEZ *HEZTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Approve(&_HEZ.TransactOpts, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_HEZ *HEZTransactor) Burn(opts *bind.TransactOpts, value *big.Int) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "burn", value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_HEZ *HEZSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Burn(&_HEZ.TransactOpts, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_HEZ *HEZTransactorSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Burn(&_HEZ.TransactOpts, value)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZTransactor) Permit(opts *bind.TransactOpts, owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "permit", owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.Contract.Permit(&_HEZ.TransactOpts, owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZTransactorSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.Contract.Permit(&_HEZ.TransactOpts, owner, spender, value, deadline, v, r, s)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_HEZ *HEZTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_HEZ *HEZSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Transfer(&_HEZ.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_HEZ *HEZTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.Transfer(&_HEZ.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_HEZ *HEZTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_HEZ *HEZSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.TransferFrom(&_HEZ.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_HEZ *HEZTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _HEZ.Contract.TransferFrom(&_HEZ.TransactOpts, from, to, value)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZTransactor) TransferWithAuthorization(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.contract.Transact(opts, "transferWithAuthorization", from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.Contract.TransferWithAuthorization(&_HEZ.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_HEZ *HEZTransactorSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _HEZ.Contract.TransferWithAuthorization(&_HEZ.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HEZ *HEZTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HEZ.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HEZ *HEZSession) Receive() (*types.Transaction, error) {
	return _HEZ.Contract.Receive(&_HEZ.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_HEZ *HEZTransactorSession) Receive() (*types.Transaction, error) {
	return _HEZ.Contract.Receive(&_HEZ.TransactOpts)
}

// HEZApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the HEZ contract.
type HEZApprovalIterator struct {
	Event *HEZApproval // Event containing the contract specifics and raw log

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
func (it *HEZApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HEZApproval)
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
		it.Event = new(HEZApproval)
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
func (it *HEZApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HEZApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HEZApproval represents a Approval event raised by the HEZ contract.
type HEZApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_HEZ *HEZFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*HEZApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _HEZ.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &HEZApprovalIterator{contract: _HEZ.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_HEZ *HEZFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *HEZApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _HEZ.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZApproval)
				if err := _HEZ.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_HEZ *HEZFilterer) ParseApproval(log types.Log) (*HEZApproval, error) {
	event := new(HEZApproval)
	if err := _HEZ.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HEZAuthorizationUsedIterator is returned from FilterAuthorizationUsed and is used to iterate over the raw logs and unpacked data for AuthorizationUsed events raised by the HEZ contract.
type HEZAuthorizationUsedIterator struct {
	Event *HEZAuthorizationUsed // Event containing the contract specifics and raw log

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
func (it *HEZAuthorizationUsedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HEZAuthorizationUsed)
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
		it.Event = new(HEZAuthorizationUsed)
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
func (it *HEZAuthorizationUsedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HEZAuthorizationUsedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HEZAuthorizationUsed represents a AuthorizationUsed event raised by the HEZ contract.
type HEZAuthorizationUsed struct {
	Authorizer common.Address
	Nonce      [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAuthorizationUsed is a free log retrieval operation binding the contract event 0x98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a5.
//
// Solidity: event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce)
func (_HEZ *HEZFilterer) FilterAuthorizationUsed(opts *bind.FilterOpts, authorizer []common.Address, nonce [][32]byte) (*HEZAuthorizationUsedIterator, error) {

	var authorizerRule []interface{}
	for _, authorizerItem := range authorizer {
		authorizerRule = append(authorizerRule, authorizerItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _HEZ.contract.FilterLogs(opts, "AuthorizationUsed", authorizerRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &HEZAuthorizationUsedIterator{contract: _HEZ.contract, event: "AuthorizationUsed", logs: logs, sub: sub}, nil
}

// WatchAuthorizationUsed is a free log subscription operation binding the contract event 0x98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a5.
//
// Solidity: event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce)
func (_HEZ *HEZFilterer) WatchAuthorizationUsed(opts *bind.WatchOpts, sink chan<- *HEZAuthorizationUsed, authorizer []common.Address, nonce [][32]byte) (event.Subscription, error) {

	var authorizerRule []interface{}
	for _, authorizerItem := range authorizer {
		authorizerRule = append(authorizerRule, authorizerItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _HEZ.contract.WatchLogs(opts, "AuthorizationUsed", authorizerRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZAuthorizationUsed)
				if err := _HEZ.contract.UnpackLog(event, "AuthorizationUsed", log); err != nil {
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

// ParseAuthorizationUsed is a log parse operation binding the contract event 0x98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a5.
//
// Solidity: event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce)
func (_HEZ *HEZFilterer) ParseAuthorizationUsed(log types.Log) (*HEZAuthorizationUsed, error) {
	event := new(HEZAuthorizationUsed)
	if err := _HEZ.contract.UnpackLog(event, "AuthorizationUsed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HEZTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the HEZ contract.
type HEZTransferIterator struct {
	Event *HEZTransfer // Event containing the contract specifics and raw log

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
func (it *HEZTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HEZTransfer)
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
		it.Event = new(HEZTransfer)
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
func (it *HEZTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HEZTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HEZTransfer represents a Transfer event raised by the HEZ contract.
type HEZTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_HEZ *HEZFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*HEZTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _HEZ.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &HEZTransferIterator{contract: _HEZ.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_HEZ *HEZFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *HEZTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _HEZ.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZTransfer)
				if err := _HEZ.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_HEZ *HEZFilterer) ParseTransfer(log types.Log) (*HEZTransfer, error) {
	event := new(HEZTransfer)
	if err := _HEZ.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
