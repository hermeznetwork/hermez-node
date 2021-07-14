// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package tokenhez

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

// TokenhezABI is the input ABI used to generate the binding from.
const TokenhezABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialHolder\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"authorizer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"}],\"name\":\"AuthorizationUsed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"EIP712DOMAIN_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NAME_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PERMIT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFER_WITH_AUTHORIZATION_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"authorizationState\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"permit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validAfter\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validBefore\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"transferWithAuthorization\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// Tokenhez is an auto generated Go binding around an Ethereum contract.
type Tokenhez struct {
	TokenhezCaller     // Read-only binding to the contract
	TokenhezTransactor // Write-only binding to the contract
	TokenhezFilterer   // Log filterer for contract events
}

// TokenhezCaller is an auto generated read-only Go binding around an Ethereum contract.
type TokenhezCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenhezTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TokenhezTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenhezFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TokenhezFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TokenhezSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TokenhezSession struct {
	Contract     *Tokenhez         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TokenhezCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TokenhezCallerSession struct {
	Contract *TokenhezCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// TokenhezTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TokenhezTransactorSession struct {
	Contract     *TokenhezTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// TokenhezRaw is an auto generated low-level Go binding around an Ethereum contract.
type TokenhezRaw struct {
	Contract *Tokenhez // Generic contract binding to access the raw methods on
}

// TokenhezCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TokenhezCallerRaw struct {
	Contract *TokenhezCaller // Generic read-only contract binding to access the raw methods on
}

// TokenhezTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TokenhezTransactorRaw struct {
	Contract *TokenhezTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTokenhez creates a new instance of Tokenhez, bound to a specific deployed contract.
func NewTokenhez(address common.Address, backend bind.ContractBackend) (*Tokenhez, error) {
	contract, err := bindTokenhez(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Tokenhez{TokenhezCaller: TokenhezCaller{contract: contract}, TokenhezTransactor: TokenhezTransactor{contract: contract}, TokenhezFilterer: TokenhezFilterer{contract: contract}}, nil
}

// NewTokenhezCaller creates a new read-only instance of Tokenhez, bound to a specific deployed contract.
func NewTokenhezCaller(address common.Address, caller bind.ContractCaller) (*TokenhezCaller, error) {
	contract, err := bindTokenhez(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TokenhezCaller{contract: contract}, nil
}

// NewTokenhezTransactor creates a new write-only instance of Tokenhez, bound to a specific deployed contract.
func NewTokenhezTransactor(address common.Address, transactor bind.ContractTransactor) (*TokenhezTransactor, error) {
	contract, err := bindTokenhez(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TokenhezTransactor{contract: contract}, nil
}

// NewTokenhezFilterer creates a new log filterer instance of Tokenhez, bound to a specific deployed contract.
func NewTokenhezFilterer(address common.Address, filterer bind.ContractFilterer) (*TokenhezFilterer, error) {
	contract, err := bindTokenhez(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TokenhezFilterer{contract: contract}, nil
}

// bindTokenhez binds a generic wrapper to an already deployed contract.
func bindTokenhez(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TokenhezABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Tokenhez *TokenhezRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Tokenhez.Contract.TokenhezCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Tokenhez *TokenhezRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Tokenhez.Contract.TokenhezTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Tokenhez *TokenhezRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Tokenhez.Contract.TokenhezTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Tokenhez *TokenhezCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Tokenhez.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Tokenhez *TokenhezTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Tokenhez.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Tokenhez *TokenhezTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Tokenhez.Contract.contract.Transact(opts, method, params...)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCaller) EIP712DOMAINHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "EIP712DOMAIN_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _Tokenhez.Contract.EIP712DOMAINHASH(&_Tokenhez.CallOpts)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCallerSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _Tokenhez.Contract.EIP712DOMAINHASH(&_Tokenhez.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCaller) NAMEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "NAME_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezSession) NAMEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.NAMEHASH(&_Tokenhez.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCallerSession) NAMEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.NAMEHASH(&_Tokenhez.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezCaller) PERMITTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "PERMIT_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezSession) PERMITTYPEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.PERMITTYPEHASH(&_Tokenhez.CallOpts)
}

// PERMITTYPEHASH is a free data retrieval call binding the contract method 0x30adf81f.
//
// Solidity: function PERMIT_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezCallerSession) PERMITTYPEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.PERMITTYPEHASH(&_Tokenhez.CallOpts)
}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezCaller) TRANSFERWITHAUTHORIZATIONTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "TRANSFER_WITH_AUTHORIZATION_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezSession) TRANSFERWITHAUTHORIZATIONTYPEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.TRANSFERWITHAUTHORIZATIONTYPEHASH(&_Tokenhez.CallOpts)
}

// TRANSFERWITHAUTHORIZATIONTYPEHASH is a free data retrieval call binding the contract method 0xa0cc6a68.
//
// Solidity: function TRANSFER_WITH_AUTHORIZATION_TYPEHASH() view returns(bytes32)
func (_Tokenhez *TokenhezCallerSession) TRANSFERWITHAUTHORIZATIONTYPEHASH() ([32]byte, error) {
	return _Tokenhez.Contract.TRANSFERWITHAUTHORIZATIONTYPEHASH(&_Tokenhez.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCaller) VERSIONHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "VERSION_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezSession) VERSIONHASH() ([32]byte, error) {
	return _Tokenhez.Contract.VERSIONHASH(&_Tokenhez.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Tokenhez *TokenhezCallerSession) VERSIONHASH() ([32]byte, error) {
	return _Tokenhez.Contract.VERSIONHASH(&_Tokenhez.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Tokenhez *TokenhezCaller) Allowance(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "allowance", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Tokenhez *TokenhezSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.Allowance(&_Tokenhez.CallOpts, arg0, arg1)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address , address ) view returns(uint256)
func (_Tokenhez *TokenhezCallerSession) Allowance(arg0 common.Address, arg1 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.Allowance(&_Tokenhez.CallOpts, arg0, arg1)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_Tokenhez *TokenhezCaller) AuthorizationState(opts *bind.CallOpts, arg0 common.Address, arg1 [32]byte) (bool, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "authorizationState", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_Tokenhez *TokenhezSession) AuthorizationState(arg0 common.Address, arg1 [32]byte) (bool, error) {
	return _Tokenhez.Contract.AuthorizationState(&_Tokenhez.CallOpts, arg0, arg1)
}

// AuthorizationState is a free data retrieval call binding the contract method 0xe94a0102.
//
// Solidity: function authorizationState(address , bytes32 ) view returns(bool)
func (_Tokenhez *TokenhezCallerSession) AuthorizationState(arg0 common.Address, arg1 [32]byte) (bool, error) {
	return _Tokenhez.Contract.AuthorizationState(&_Tokenhez.CallOpts, arg0, arg1)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Tokenhez *TokenhezCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Tokenhez *TokenhezSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.BalanceOf(&_Tokenhez.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_Tokenhez *TokenhezCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.BalanceOf(&_Tokenhez.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Tokenhez *TokenhezCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Tokenhez *TokenhezSession) Decimals() (uint8, error) {
	return _Tokenhez.Contract.Decimals(&_Tokenhez.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Tokenhez *TokenhezCallerSession) Decimals() (uint8, error) {
	return _Tokenhez.Contract.Decimals(&_Tokenhez.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Tokenhez *TokenhezCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Tokenhez *TokenhezSession) GetChainId() (*big.Int, error) {
	return _Tokenhez.Contract.GetChainId(&_Tokenhez.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Tokenhez *TokenhezCallerSession) GetChainId() (*big.Int, error) {
	return _Tokenhez.Contract.GetChainId(&_Tokenhez.CallOpts)
}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_Tokenhez *TokenhezCaller) InitialBalance(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "initialBalance")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_Tokenhez *TokenhezSession) InitialBalance() (*big.Int, error) {
	return _Tokenhez.Contract.InitialBalance(&_Tokenhez.CallOpts)
}

// InitialBalance is a free data retrieval call binding the contract method 0x18369a2a.
//
// Solidity: function initialBalance() view returns(uint256)
func (_Tokenhez *TokenhezCallerSession) InitialBalance() (*big.Int, error) {
	return _Tokenhez.Contract.InitialBalance(&_Tokenhez.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Tokenhez *TokenhezCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Tokenhez *TokenhezSession) Name() (string, error) {
	return _Tokenhez.Contract.Name(&_Tokenhez.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_Tokenhez *TokenhezCallerSession) Name() (string, error) {
	return _Tokenhez.Contract.Name(&_Tokenhez.CallOpts)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_Tokenhez *TokenhezCaller) Nonces(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "nonces", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_Tokenhez *TokenhezSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.Nonces(&_Tokenhez.CallOpts, arg0)
}

// Nonces is a free data retrieval call binding the contract method 0x7ecebe00.
//
// Solidity: function nonces(address ) view returns(uint256)
func (_Tokenhez *TokenhezCallerSession) Nonces(arg0 common.Address) (*big.Int, error) {
	return _Tokenhez.Contract.Nonces(&_Tokenhez.CallOpts, arg0)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Tokenhez *TokenhezCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Tokenhez *TokenhezSession) Symbol() (string, error) {
	return _Tokenhez.Contract.Symbol(&_Tokenhez.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_Tokenhez *TokenhezCallerSession) Symbol() (string, error) {
	return _Tokenhez.Contract.Symbol(&_Tokenhez.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Tokenhez *TokenhezCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Tokenhez.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Tokenhez *TokenhezSession) TotalSupply() (*big.Int, error) {
	return _Tokenhez.Contract.TotalSupply(&_Tokenhez.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_Tokenhez *TokenhezCallerSession) TotalSupply() (*big.Int, error) {
	return _Tokenhez.Contract.TotalSupply(&_Tokenhez.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_Tokenhez *TokenhezSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Approve(&_Tokenhez.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Approve(&_Tokenhez.TransactOpts, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactor) Burn(opts *bind.TransactOpts, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "burn", value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_Tokenhez *TokenhezSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Burn(&_Tokenhez.TransactOpts, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactorSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Burn(&_Tokenhez.TransactOpts, value)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezTransactor) Permit(opts *bind.TransactOpts, owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "permit", owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.Contract.Permit(&_Tokenhez.TransactOpts, owner, spender, value, deadline, v, r, s)
}

// Permit is a paid mutator transaction binding the contract method 0xd505accf.
//
// Solidity: function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezTransactorSession) Permit(owner common.Address, spender common.Address, value *big.Int, deadline *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.Contract.Permit(&_Tokenhez.TransactOpts, owner, spender, value, deadline, v, r, s)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Transfer(&_Tokenhez.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.Transfer(&_Tokenhez.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.TransferFrom(&_Tokenhez.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_Tokenhez *TokenhezTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _Tokenhez.Contract.TransferFrom(&_Tokenhez.TransactOpts, from, to, value)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezTransactor) TransferWithAuthorization(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.contract.Transact(opts, "transferWithAuthorization", from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.Contract.TransferWithAuthorization(&_Tokenhez.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TransferWithAuthorization is a paid mutator transaction binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address from, address to, uint256 value, uint256 validAfter, uint256 validBefore, bytes32 nonce, uint8 v, bytes32 r, bytes32 s) returns()
func (_Tokenhez *TokenhezTransactorSession) TransferWithAuthorization(from common.Address, to common.Address, value *big.Int, validAfter *big.Int, validBefore *big.Int, nonce [32]byte, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Tokenhez.Contract.TransferWithAuthorization(&_Tokenhez.TransactOpts, from, to, value, validAfter, validBefore, nonce, v, r, s)
}

// TokenhezApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the Tokenhez contract.
type TokenhezApprovalIterator struct {
	Event *TokenhezApproval // Event containing the contract specifics and raw log

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
func (it *TokenhezApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TokenhezApproval)
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
		it.Event = new(TokenhezApproval)
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
func (it *TokenhezApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TokenhezApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TokenhezApproval represents a Approval event raised by the Tokenhez contract.
type TokenhezApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Tokenhez *TokenhezFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*TokenhezApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Tokenhez.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &TokenhezApprovalIterator{contract: _Tokenhez.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_Tokenhez *TokenhezFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *TokenhezApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _Tokenhez.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TokenhezApproval)
				if err := _Tokenhez.contract.UnpackLog(event, "Approval", log); err != nil {
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
func (_Tokenhez *TokenhezFilterer) ParseApproval(log types.Log) (*TokenhezApproval, error) {
	event := new(TokenhezApproval)
	if err := _Tokenhez.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TokenhezAuthorizationUsedIterator is returned from FilterAuthorizationUsed and is used to iterate over the raw logs and unpacked data for AuthorizationUsed events raised by the Tokenhez contract.
type TokenhezAuthorizationUsedIterator struct {
	Event *TokenhezAuthorizationUsed // Event containing the contract specifics and raw log

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
func (it *TokenhezAuthorizationUsedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TokenhezAuthorizationUsed)
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
		it.Event = new(TokenhezAuthorizationUsed)
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
func (it *TokenhezAuthorizationUsedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TokenhezAuthorizationUsedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TokenhezAuthorizationUsed represents a AuthorizationUsed event raised by the Tokenhez contract.
type TokenhezAuthorizationUsed struct {
	Authorizer common.Address
	Nonce      [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAuthorizationUsed is a free log retrieval operation binding the contract event 0x98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a5.
//
// Solidity: event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce)
func (_Tokenhez *TokenhezFilterer) FilterAuthorizationUsed(opts *bind.FilterOpts, authorizer []common.Address, nonce [][32]byte) (*TokenhezAuthorizationUsedIterator, error) {

	var authorizerRule []interface{}
	for _, authorizerItem := range authorizer {
		authorizerRule = append(authorizerRule, authorizerItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Tokenhez.contract.FilterLogs(opts, "AuthorizationUsed", authorizerRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return &TokenhezAuthorizationUsedIterator{contract: _Tokenhez.contract, event: "AuthorizationUsed", logs: logs, sub: sub}, nil
}

// WatchAuthorizationUsed is a free log subscription operation binding the contract event 0x98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a5.
//
// Solidity: event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce)
func (_Tokenhez *TokenhezFilterer) WatchAuthorizationUsed(opts *bind.WatchOpts, sink chan<- *TokenhezAuthorizationUsed, authorizer []common.Address, nonce [][32]byte) (event.Subscription, error) {

	var authorizerRule []interface{}
	for _, authorizerItem := range authorizer {
		authorizerRule = append(authorizerRule, authorizerItem)
	}
	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}

	logs, sub, err := _Tokenhez.contract.WatchLogs(opts, "AuthorizationUsed", authorizerRule, nonceRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TokenhezAuthorizationUsed)
				if err := _Tokenhez.contract.UnpackLog(event, "AuthorizationUsed", log); err != nil {
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
func (_Tokenhez *TokenhezFilterer) ParseAuthorizationUsed(log types.Log) (*TokenhezAuthorizationUsed, error) {
	event := new(TokenhezAuthorizationUsed)
	if err := _Tokenhez.contract.UnpackLog(event, "AuthorizationUsed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TokenhezTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the Tokenhez contract.
type TokenhezTransferIterator struct {
	Event *TokenhezTransfer // Event containing the contract specifics and raw log

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
func (it *TokenhezTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TokenhezTransfer)
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
		it.Event = new(TokenhezTransfer)
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
func (it *TokenhezTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TokenhezTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TokenhezTransfer represents a Transfer event raised by the Tokenhez contract.
type TokenhezTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Tokenhez *TokenhezFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*TokenhezTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Tokenhez.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &TokenhezTransferIterator{contract: _Tokenhez.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_Tokenhez *TokenhezFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *TokenhezTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _Tokenhez.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TokenhezTransfer)
				if err := _Tokenhez.contract.UnpackLog(event, "Transfer", log); err != nil {
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
func (_Tokenhez *TokenhezFilterer) ParseTransfer(log types.Log) (*TokenhezTransfer, error) {
	event := new(TokenhezTransfer)
	if err := _Tokenhez.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
