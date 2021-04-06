// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package HermezAuctionProtocol

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

// HermezAuctionProtocolABI is the input ABI used to generate the binding from.
const HermezAuctionProtocolABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"}],\"name\":\"HEZClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"donationAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"bootCoordinatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"bootCoordinatorURL\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"outbidding\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"slotDeadline\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"closedAuctionSlots\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"openAuctionSlots\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"allocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"InitializeHermezAuctionProtocolEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"newAllocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"NewAllocationRatio\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"}],\"name\":\"NewBid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newBootCoordinator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"newBootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"NewBootCoordinator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newClosedAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"NewClosedAuctionSlots\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"slotSet\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"newInitialMinBid\",\"type\":\"uint128\"}],\"name\":\"NewDefaultSlotSetBid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newDonationAddress\",\"type\":\"address\"}],\"name\":\"NewDonationAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slotToForge\",\"type\":\"uint128\"}],\"name\":\"NewForge\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slotToForge\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"burnAmount\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"donationAmount\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"governanceAmount\",\"type\":\"uint128\"}],\"name\":\"NewForgeAllocated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newOpenAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"NewOpenAuctionSlots\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newOutbidding\",\"type\":\"uint16\"}],\"name\":\"NewOutbidding\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"newSlotDeadline\",\"type\":\"uint8\"}],\"name\":\"NewSlotDeadline\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"name\":\"SetCoordinator\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BLOCKS_PER_SLOT\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"INITIAL_MINIMAL_BIDDING\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bootCoordinatorURL\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"canForge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slotSet\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"newInitialMinBid\",\"type\":\"uint128\"}],\"name\":\"changeDefaultSlotSetBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimHEZ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"claimPendingHEZ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"coordinators\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"}],\"name\":\"forge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"genesisBlock\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllocationRatio\",\"outputs\":[{\"internalType\":\"uint16[3]\",\"name\":\"\",\"type\":\"uint16[3]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBootCoordinator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"}],\"name\":\"getClaimableHEZ\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getClosedAuctionSlots\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentSlotNumber\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"slotSet\",\"type\":\"uint8\"}],\"name\":\"getDefaultSlotSetBid\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDonationAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"getMinBidBySlot\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOpenAuctionSlots\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOutbidding\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSlotDeadline\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"blockNumber\",\"type\":\"uint128\"}],\"name\":\"getSlotNumber\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"getSlotSet\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint128\",\"name\":\"genesis\",\"type\":\"uint128\"},{\"internalType\":\"address\",\"name\":\"hermezRollupAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_governanceAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"donationAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"bootCoordinatorAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_bootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"hermezAuctionProtocolInitializer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollup\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"pendingBalances\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"processBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"startingSlot\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"endingSlot\",\"type\":\"uint128\"},{\"internalType\":\"bool[6]\",\"name\":\"slotSets\",\"type\":\"bool[6]\"},{\"internalType\":\"uint128\",\"name\":\"maxBid\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"minBid\",\"type\":\"uint128\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"processMultiBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16[3]\",\"name\":\"newAllocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"setAllocationRatio\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newBootCoordinator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"newBootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"setBootCoordinator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newClosedAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"setClosedAuctionSlots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"name\":\"setCoordinator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newDonationAddress\",\"type\":\"address\"}],\"name\":\"setDonationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newOpenAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"setOpenAuctionSlots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newOutbidding\",\"type\":\"uint16\"}],\"name\":\"setOutbidding\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newDeadline\",\"type\":\"uint8\"}],\"name\":\"setSlotDeadline\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"name\":\"slots\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"fulfilled\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"forgerCommitment\",\"type\":\"bool\"},{\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"closedMinBid\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenHEZ\",\"outputs\":[{\"internalType\":\"contractIHEZToken\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// HermezAuctionProtocol is an auto generated Go binding around an Ethereum contract.
type HermezAuctionProtocol struct {
	HermezAuctionProtocolCaller     // Read-only binding to the contract
	HermezAuctionProtocolTransactor // Write-only binding to the contract
	HermezAuctionProtocolFilterer   // Log filterer for contract events
}

// HermezAuctionProtocolCaller is an auto generated read-only Go binding around an Ethereum contract.
type HermezAuctionProtocolCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezAuctionProtocolTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HermezAuctionProtocolTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezAuctionProtocolFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HermezAuctionProtocolFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezAuctionProtocolSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HermezAuctionProtocolSession struct {
	Contract     *HermezAuctionProtocol // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// HermezAuctionProtocolCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HermezAuctionProtocolCallerSession struct {
	Contract *HermezAuctionProtocolCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// HermezAuctionProtocolTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HermezAuctionProtocolTransactorSession struct {
	Contract     *HermezAuctionProtocolTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// HermezAuctionProtocolRaw is an auto generated low-level Go binding around an Ethereum contract.
type HermezAuctionProtocolRaw struct {
	Contract *HermezAuctionProtocol // Generic contract binding to access the raw methods on
}

// HermezAuctionProtocolCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HermezAuctionProtocolCallerRaw struct {
	Contract *HermezAuctionProtocolCaller // Generic read-only contract binding to access the raw methods on
}

// HermezAuctionProtocolTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HermezAuctionProtocolTransactorRaw struct {
	Contract *HermezAuctionProtocolTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHermezAuctionProtocol creates a new instance of HermezAuctionProtocol, bound to a specific deployed contract.
func NewHermezAuctionProtocol(address common.Address, backend bind.ContractBackend) (*HermezAuctionProtocol, error) {
	contract, err := bindHermezAuctionProtocol(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocol{HermezAuctionProtocolCaller: HermezAuctionProtocolCaller{contract: contract}, HermezAuctionProtocolTransactor: HermezAuctionProtocolTransactor{contract: contract}, HermezAuctionProtocolFilterer: HermezAuctionProtocolFilterer{contract: contract}}, nil
}

// NewHermezAuctionProtocolCaller creates a new read-only instance of HermezAuctionProtocol, bound to a specific deployed contract.
func NewHermezAuctionProtocolCaller(address common.Address, caller bind.ContractCaller) (*HermezAuctionProtocolCaller, error) {
	contract, err := bindHermezAuctionProtocol(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolCaller{contract: contract}, nil
}

// NewHermezAuctionProtocolTransactor creates a new write-only instance of HermezAuctionProtocol, bound to a specific deployed contract.
func NewHermezAuctionProtocolTransactor(address common.Address, transactor bind.ContractTransactor) (*HermezAuctionProtocolTransactor, error) {
	contract, err := bindHermezAuctionProtocol(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolTransactor{contract: contract}, nil
}

// NewHermezAuctionProtocolFilterer creates a new log filterer instance of HermezAuctionProtocol, bound to a specific deployed contract.
func NewHermezAuctionProtocolFilterer(address common.Address, filterer bind.ContractFilterer) (*HermezAuctionProtocolFilterer, error) {
	contract, err := bindHermezAuctionProtocol(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolFilterer{contract: contract}, nil
}

// bindHermezAuctionProtocol binds a generic wrapper to an already deployed contract.
func bindHermezAuctionProtocol(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HermezAuctionProtocolABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HermezAuctionProtocol *HermezAuctionProtocolRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HermezAuctionProtocol.Contract.HermezAuctionProtocolCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HermezAuctionProtocol *HermezAuctionProtocolRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.HermezAuctionProtocolTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HermezAuctionProtocol *HermezAuctionProtocolRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.HermezAuctionProtocolTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _HermezAuctionProtocol.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.contract.Transact(opts, method, params...)
}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) BLOCKSPERSLOT(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "BLOCKS_PER_SLOT")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) BLOCKSPERSLOT() (uint8, error) {
	return _HermezAuctionProtocol.Contract.BLOCKSPERSLOT(&_HermezAuctionProtocol.CallOpts)
}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) BLOCKSPERSLOT() (uint8, error) {
	return _HermezAuctionProtocol.Contract.BLOCKSPERSLOT(&_HermezAuctionProtocol.CallOpts)
}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) INITIALMINIMALBIDDING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "INITIAL_MINIMAL_BIDDING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) INITIALMINIMALBIDDING() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.INITIALMINIMALBIDDING(&_HermezAuctionProtocol.CallOpts)
}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) INITIALMINIMALBIDDING() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.INITIALMINIMALBIDDING(&_HermezAuctionProtocol.CallOpts)
}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) BootCoordinatorURL(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "bootCoordinatorURL")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) BootCoordinatorURL() (string, error) {
	return _HermezAuctionProtocol.Contract.BootCoordinatorURL(&_HermezAuctionProtocol.CallOpts)
}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) BootCoordinatorURL() (string, error) {
	return _HermezAuctionProtocol.Contract.BootCoordinatorURL(&_HermezAuctionProtocol.CallOpts)
}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) CanForge(opts *bind.CallOpts, forger common.Address, blockNumber *big.Int) (bool, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "canForge", forger, blockNumber)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) CanForge(forger common.Address, blockNumber *big.Int) (bool, error) {
	return _HermezAuctionProtocol.Contract.CanForge(&_HermezAuctionProtocol.CallOpts, forger, blockNumber)
}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) CanForge(forger common.Address, blockNumber *big.Int) (bool, error) {
	return _HermezAuctionProtocol.Contract.CanForge(&_HermezAuctionProtocol.CallOpts, forger, blockNumber)
}

// Coordinators is a free data retrieval call binding the contract method 0xa48af096.
//
// Solidity: function coordinators(address ) view returns(address forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) Coordinators(opts *bind.CallOpts, arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "coordinators", arg0)

	outstruct := new(struct {
		Forger         common.Address
		CoordinatorURL string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Forger = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.CoordinatorURL = *abi.ConvertType(out[1], new(string)).(*string)

	return *outstruct, err

}

// Coordinators is a free data retrieval call binding the contract method 0xa48af096.
//
// Solidity: function coordinators(address ) view returns(address forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) Coordinators(arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	return _HermezAuctionProtocol.Contract.Coordinators(&_HermezAuctionProtocol.CallOpts, arg0)
}

// Coordinators is a free data retrieval call binding the contract method 0xa48af096.
//
// Solidity: function coordinators(address ) view returns(address forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) Coordinators(arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	return _HermezAuctionProtocol.Contract.Coordinators(&_HermezAuctionProtocol.CallOpts, arg0)
}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GenesisBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "genesisBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GenesisBlock() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GenesisBlock(&_HermezAuctionProtocol.CallOpts)
}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GenesisBlock() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GenesisBlock(&_HermezAuctionProtocol.CallOpts)
}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetAllocationRatio(opts *bind.CallOpts) ([3]uint16, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getAllocationRatio")

	if err != nil {
		return *new([3]uint16), err
	}

	out0 := *abi.ConvertType(out[0], new([3]uint16)).(*[3]uint16)

	return out0, err

}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetAllocationRatio() ([3]uint16, error) {
	return _HermezAuctionProtocol.Contract.GetAllocationRatio(&_HermezAuctionProtocol.CallOpts)
}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetAllocationRatio() ([3]uint16, error) {
	return _HermezAuctionProtocol.Contract.GetAllocationRatio(&_HermezAuctionProtocol.CallOpts)
}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetBootCoordinator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getBootCoordinator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetBootCoordinator() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GetBootCoordinator(&_HermezAuctionProtocol.CallOpts)
}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetBootCoordinator() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GetBootCoordinator(&_HermezAuctionProtocol.CallOpts)
}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetClaimableHEZ(opts *bind.CallOpts, bidder common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getClaimableHEZ", bidder)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetClaimableHEZ(bidder common.Address) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetClaimableHEZ(&_HermezAuctionProtocol.CallOpts, bidder)
}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetClaimableHEZ(bidder common.Address) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetClaimableHEZ(&_HermezAuctionProtocol.CallOpts, bidder)
}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetClosedAuctionSlots(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getClosedAuctionSlots")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetClosedAuctionSlots() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetClosedAuctionSlots(&_HermezAuctionProtocol.CallOpts)
}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetClosedAuctionSlots() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetClosedAuctionSlots(&_HermezAuctionProtocol.CallOpts)
}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetCurrentSlotNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getCurrentSlotNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetCurrentSlotNumber() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetCurrentSlotNumber(&_HermezAuctionProtocol.CallOpts)
}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetCurrentSlotNumber() (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetCurrentSlotNumber(&_HermezAuctionProtocol.CallOpts)
}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetDefaultSlotSetBid(opts *bind.CallOpts, slotSet uint8) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getDefaultSlotSetBid", slotSet)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetDefaultSlotSetBid(&_HermezAuctionProtocol.CallOpts, slotSet)
}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetDefaultSlotSetBid(&_HermezAuctionProtocol.CallOpts, slotSet)
}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetDonationAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getDonationAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetDonationAddress() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GetDonationAddress(&_HermezAuctionProtocol.CallOpts)
}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetDonationAddress() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GetDonationAddress(&_HermezAuctionProtocol.CallOpts)
}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetMinBidBySlot(opts *bind.CallOpts, slot *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getMinBidBySlot", slot)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetMinBidBySlot(slot *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetMinBidBySlot(&_HermezAuctionProtocol.CallOpts, slot)
}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetMinBidBySlot(slot *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetMinBidBySlot(&_HermezAuctionProtocol.CallOpts, slot)
}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetOpenAuctionSlots(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getOpenAuctionSlots")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetOpenAuctionSlots() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetOpenAuctionSlots(&_HermezAuctionProtocol.CallOpts)
}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetOpenAuctionSlots() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetOpenAuctionSlots(&_HermezAuctionProtocol.CallOpts)
}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetOutbidding(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getOutbidding")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetOutbidding() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetOutbidding(&_HermezAuctionProtocol.CallOpts)
}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetOutbidding() (uint16, error) {
	return _HermezAuctionProtocol.Contract.GetOutbidding(&_HermezAuctionProtocol.CallOpts)
}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetSlotDeadline(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getSlotDeadline")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetSlotDeadline() (uint8, error) {
	return _HermezAuctionProtocol.Contract.GetSlotDeadline(&_HermezAuctionProtocol.CallOpts)
}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetSlotDeadline() (uint8, error) {
	return _HermezAuctionProtocol.Contract.GetSlotDeadline(&_HermezAuctionProtocol.CallOpts)
}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetSlotNumber(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getSlotNumber", blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetSlotNumber(blockNumber *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetSlotNumber(&_HermezAuctionProtocol.CallOpts, blockNumber)
}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetSlotNumber(blockNumber *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetSlotNumber(&_HermezAuctionProtocol.CallOpts, blockNumber)
}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GetSlotSet(opts *bind.CallOpts, slot *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "getSlotSet", slot)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GetSlotSet(slot *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetSlotSet(&_HermezAuctionProtocol.CallOpts, slot)
}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GetSlotSet(slot *big.Int) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.GetSlotSet(&_HermezAuctionProtocol.CallOpts, slot)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) GovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "governanceAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) GovernanceAddress() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GovernanceAddress(&_HermezAuctionProtocol.CallOpts)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) GovernanceAddress() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.GovernanceAddress(&_HermezAuctionProtocol.CallOpts)
}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) HermezRollup(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "hermezRollup")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) HermezRollup() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.HermezRollup(&_HermezAuctionProtocol.CallOpts)
}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) HermezRollup() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.HermezRollup(&_HermezAuctionProtocol.CallOpts)
}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) PendingBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "pendingBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) PendingBalances(arg0 common.Address) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.PendingBalances(&_HermezAuctionProtocol.CallOpts, arg0)
}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) PendingBalances(arg0 common.Address) (*big.Int, error) {
	return _HermezAuctionProtocol.Contract.PendingBalances(&_HermezAuctionProtocol.CallOpts, arg0)
}

// Slots is a free data retrieval call binding the contract method 0xbc415567.
//
// Solidity: function slots(uint128 ) view returns(address bidder, bool fulfilled, bool forgerCommitment, uint128 bidAmount, uint128 closedMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) Slots(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "slots", arg0)

	outstruct := new(struct {
		Bidder           common.Address
		Fulfilled        bool
		ForgerCommitment bool
		BidAmount        *big.Int
		ClosedMinBid     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Bidder = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Fulfilled = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.ForgerCommitment = *abi.ConvertType(out[2], new(bool)).(*bool)
	outstruct.BidAmount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.ClosedMinBid = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Slots is a free data retrieval call binding the contract method 0xbc415567.
//
// Solidity: function slots(uint128 ) view returns(address bidder, bool fulfilled, bool forgerCommitment, uint128 bidAmount, uint128 closedMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) Slots(arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	return _HermezAuctionProtocol.Contract.Slots(&_HermezAuctionProtocol.CallOpts, arg0)
}

// Slots is a free data retrieval call binding the contract method 0xbc415567.
//
// Solidity: function slots(uint128 ) view returns(address bidder, bool fulfilled, bool forgerCommitment, uint128 bidAmount, uint128 closedMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) Slots(arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	return _HermezAuctionProtocol.Contract.Slots(&_HermezAuctionProtocol.CallOpts, arg0)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCaller) TokenHEZ(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _HermezAuctionProtocol.contract.Call(opts, &out, "tokenHEZ")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) TokenHEZ() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.TokenHEZ(&_HermezAuctionProtocol.CallOpts)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_HermezAuctionProtocol *HermezAuctionProtocolCallerSession) TokenHEZ() (common.Address, error) {
	return _HermezAuctionProtocol.Contract.TokenHEZ(&_HermezAuctionProtocol.CallOpts)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) ChangeDefaultSlotSetBid(opts *bind.TransactOpts, slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "changeDefaultSlotSetBid", slotSet, newInitialMinBid)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) ChangeDefaultSlotSetBid(slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ChangeDefaultSlotSetBid(&_HermezAuctionProtocol.TransactOpts, slotSet, newInitialMinBid)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) ChangeDefaultSlotSetBid(slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ChangeDefaultSlotSetBid(&_HermezAuctionProtocol.TransactOpts, slotSet, newInitialMinBid)
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) ClaimHEZ(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "claimHEZ")
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) ClaimHEZ() (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ClaimHEZ(&_HermezAuctionProtocol.TransactOpts)
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) ClaimHEZ() (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ClaimHEZ(&_HermezAuctionProtocol.TransactOpts)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) ClaimPendingHEZ(opts *bind.TransactOpts, slot *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "claimPendingHEZ", slot)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) ClaimPendingHEZ(slot *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ClaimPendingHEZ(&_HermezAuctionProtocol.TransactOpts, slot)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) ClaimPendingHEZ(slot *big.Int) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ClaimPendingHEZ(&_HermezAuctionProtocol.TransactOpts, slot)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) Forge(opts *bind.TransactOpts, forger common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "forge", forger)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) Forge(forger common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.Forge(&_HermezAuctionProtocol.TransactOpts, forger)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) Forge(forger common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.Forge(&_HermezAuctionProtocol.TransactOpts, forger)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) HermezAuctionProtocolInitializer(opts *bind.TransactOpts, token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "hermezAuctionProtocolInitializer", token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) HermezAuctionProtocolInitializer(token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.HermezAuctionProtocolInitializer(&_HermezAuctionProtocol.TransactOpts, token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) HermezAuctionProtocolInitializer(token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.HermezAuctionProtocolInitializer(&_HermezAuctionProtocol.TransactOpts, token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) ProcessBid(opts *bind.TransactOpts, amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "processBid", amount, slot, bidAmount, permit)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) ProcessBid(amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ProcessBid(&_HermezAuctionProtocol.TransactOpts, amount, slot, bidAmount, permit)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) ProcessBid(amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ProcessBid(&_HermezAuctionProtocol.TransactOpts, amount, slot, bidAmount, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) ProcessMultiBid(opts *bind.TransactOpts, amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "processMultiBid", amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) ProcessMultiBid(amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ProcessMultiBid(&_HermezAuctionProtocol.TransactOpts, amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) ProcessMultiBid(amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.ProcessMultiBid(&_HermezAuctionProtocol.TransactOpts, amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetAllocationRatio(opts *bind.TransactOpts, newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setAllocationRatio", newAllocationRatio)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetAllocationRatio(&_HermezAuctionProtocol.TransactOpts, newAllocationRatio)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetAllocationRatio(&_HermezAuctionProtocol.TransactOpts, newAllocationRatio)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetBootCoordinator(opts *bind.TransactOpts, newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setBootCoordinator", newBootCoordinator, newBootCoordinatorURL)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetBootCoordinator(newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetBootCoordinator(&_HermezAuctionProtocol.TransactOpts, newBootCoordinator, newBootCoordinatorURL)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetBootCoordinator(newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetBootCoordinator(&_HermezAuctionProtocol.TransactOpts, newBootCoordinator, newBootCoordinatorURL)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetClosedAuctionSlots(opts *bind.TransactOpts, newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setClosedAuctionSlots", newClosedAuctionSlots)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetClosedAuctionSlots(&_HermezAuctionProtocol.TransactOpts, newClosedAuctionSlots)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetClosedAuctionSlots(&_HermezAuctionProtocol.TransactOpts, newClosedAuctionSlots)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetCoordinator(opts *bind.TransactOpts, forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setCoordinator", forger, coordinatorURL)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetCoordinator(forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetCoordinator(&_HermezAuctionProtocol.TransactOpts, forger, coordinatorURL)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetCoordinator(forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetCoordinator(&_HermezAuctionProtocol.TransactOpts, forger, coordinatorURL)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetDonationAddress(opts *bind.TransactOpts, newDonationAddress common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setDonationAddress", newDonationAddress)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetDonationAddress(newDonationAddress common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetDonationAddress(&_HermezAuctionProtocol.TransactOpts, newDonationAddress)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetDonationAddress(newDonationAddress common.Address) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetDonationAddress(&_HermezAuctionProtocol.TransactOpts, newDonationAddress)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetOpenAuctionSlots(opts *bind.TransactOpts, newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setOpenAuctionSlots", newOpenAuctionSlots)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetOpenAuctionSlots(&_HermezAuctionProtocol.TransactOpts, newOpenAuctionSlots)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetOpenAuctionSlots(&_HermezAuctionProtocol.TransactOpts, newOpenAuctionSlots)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetOutbidding(opts *bind.TransactOpts, newOutbidding uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setOutbidding", newOutbidding)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetOutbidding(newOutbidding uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetOutbidding(&_HermezAuctionProtocol.TransactOpts, newOutbidding)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetOutbidding(newOutbidding uint16) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetOutbidding(&_HermezAuctionProtocol.TransactOpts, newOutbidding)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactor) SetSlotDeadline(opts *bind.TransactOpts, newDeadline uint8) (*types.Transaction, error) {
	return _HermezAuctionProtocol.contract.Transact(opts, "setSlotDeadline", newDeadline)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolSession) SetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetSlotDeadline(&_HermezAuctionProtocol.TransactOpts, newDeadline)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_HermezAuctionProtocol *HermezAuctionProtocolTransactorSession) SetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return _HermezAuctionProtocol.Contract.SetSlotDeadline(&_HermezAuctionProtocol.TransactOpts, newDeadline)
}

// HermezAuctionProtocolHEZClaimedIterator is returned from FilterHEZClaimed and is used to iterate over the raw logs and unpacked data for HEZClaimed events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolHEZClaimedIterator struct {
	Event *HermezAuctionProtocolHEZClaimed // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolHEZClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolHEZClaimed)
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
		it.Event = new(HermezAuctionProtocolHEZClaimed)
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
func (it *HermezAuctionProtocolHEZClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolHEZClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolHEZClaimed represents a HEZClaimed event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolHEZClaimed struct {
	Owner  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHEZClaimed is a free log retrieval operation binding the contract event 0x199ef0cb54d2b296ff6eaec2721bacf0ca3fd8344a43f5bdf4548b34dfa2594f.
//
// Solidity: event HEZClaimed(address indexed owner, uint128 amount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterHEZClaimed(opts *bind.FilterOpts, owner []common.Address) (*HermezAuctionProtocolHEZClaimedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "HEZClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolHEZClaimedIterator{contract: _HermezAuctionProtocol.contract, event: "HEZClaimed", logs: logs, sub: sub}, nil
}

// WatchHEZClaimed is a free log subscription operation binding the contract event 0x199ef0cb54d2b296ff6eaec2721bacf0ca3fd8344a43f5bdf4548b34dfa2594f.
//
// Solidity: event HEZClaimed(address indexed owner, uint128 amount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchHEZClaimed(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolHEZClaimed, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "HEZClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolHEZClaimed)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "HEZClaimed", log); err != nil {
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

// ParseHEZClaimed is a log parse operation binding the contract event 0x199ef0cb54d2b296ff6eaec2721bacf0ca3fd8344a43f5bdf4548b34dfa2594f.
//
// Solidity: event HEZClaimed(address indexed owner, uint128 amount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseHEZClaimed(log types.Log) (*HermezAuctionProtocolHEZClaimed, error) {
	event := new(HermezAuctionProtocolHEZClaimed)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "HEZClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator is returned from FilterInitializeHermezAuctionProtocolEvent and is used to iterate over the raw logs and unpacked data for InitializeHermezAuctionProtocolEvent events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator struct {
	Event *HermezAuctionProtocolInitializeHermezAuctionProtocolEvent // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolInitializeHermezAuctionProtocolEvent)
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
		it.Event = new(HermezAuctionProtocolInitializeHermezAuctionProtocolEvent)
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
func (it *HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolInitializeHermezAuctionProtocolEvent represents a InitializeHermezAuctionProtocolEvent event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolInitializeHermezAuctionProtocolEvent struct {
	DonationAddress        common.Address
	BootCoordinatorAddress common.Address
	BootCoordinatorURL     string
	Outbidding             uint16
	SlotDeadline           uint8
	ClosedAuctionSlots     uint16
	OpenAuctionSlots       uint16
	AllocationRatio        [3]uint16
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterInitializeHermezAuctionProtocolEvent is a free log retrieval operation binding the contract event 0x9717e4e04c13817c600463a7a450110c754fd78758cdd538603f30528a24ce4b.
//
// Solidity: event InitializeHermezAuctionProtocolEvent(address donationAddress, address bootCoordinatorAddress, string bootCoordinatorURL, uint16 outbidding, uint8 slotDeadline, uint16 closedAuctionSlots, uint16 openAuctionSlots, uint16[3] allocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterInitializeHermezAuctionProtocolEvent(opts *bind.FilterOpts) (*HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "InitializeHermezAuctionProtocolEvent")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolInitializeHermezAuctionProtocolEventIterator{contract: _HermezAuctionProtocol.contract, event: "InitializeHermezAuctionProtocolEvent", logs: logs, sub: sub}, nil
}

// WatchInitializeHermezAuctionProtocolEvent is a free log subscription operation binding the contract event 0x9717e4e04c13817c600463a7a450110c754fd78758cdd538603f30528a24ce4b.
//
// Solidity: event InitializeHermezAuctionProtocolEvent(address donationAddress, address bootCoordinatorAddress, string bootCoordinatorURL, uint16 outbidding, uint8 slotDeadline, uint16 closedAuctionSlots, uint16 openAuctionSlots, uint16[3] allocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchInitializeHermezAuctionProtocolEvent(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolInitializeHermezAuctionProtocolEvent) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "InitializeHermezAuctionProtocolEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolInitializeHermezAuctionProtocolEvent)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "InitializeHermezAuctionProtocolEvent", log); err != nil {
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

// ParseInitializeHermezAuctionProtocolEvent is a log parse operation binding the contract event 0x9717e4e04c13817c600463a7a450110c754fd78758cdd538603f30528a24ce4b.
//
// Solidity: event InitializeHermezAuctionProtocolEvent(address donationAddress, address bootCoordinatorAddress, string bootCoordinatorURL, uint16 outbidding, uint8 slotDeadline, uint16 closedAuctionSlots, uint16 openAuctionSlots, uint16[3] allocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseInitializeHermezAuctionProtocolEvent(log types.Log) (*HermezAuctionProtocolInitializeHermezAuctionProtocolEvent, error) {
	event := new(HermezAuctionProtocolInitializeHermezAuctionProtocolEvent)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "InitializeHermezAuctionProtocolEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewAllocationRatioIterator is returned from FilterNewAllocationRatio and is used to iterate over the raw logs and unpacked data for NewAllocationRatio events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewAllocationRatioIterator struct {
	Event *HermezAuctionProtocolNewAllocationRatio // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewAllocationRatioIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewAllocationRatio)
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
		it.Event = new(HermezAuctionProtocolNewAllocationRatio)
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
func (it *HermezAuctionProtocolNewAllocationRatioIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewAllocationRatioIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewAllocationRatio represents a NewAllocationRatio event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewAllocationRatio struct {
	NewAllocationRatio [3]uint16
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNewAllocationRatio is a free log retrieval operation binding the contract event 0x0bb59eceb12f1bdb63e4a7d57c70d6473fefd7c3f51af5a3604f7e97197073e4.
//
// Solidity: event NewAllocationRatio(uint16[3] newAllocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewAllocationRatio(opts *bind.FilterOpts) (*HermezAuctionProtocolNewAllocationRatioIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewAllocationRatio")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewAllocationRatioIterator{contract: _HermezAuctionProtocol.contract, event: "NewAllocationRatio", logs: logs, sub: sub}, nil
}

// WatchNewAllocationRatio is a free log subscription operation binding the contract event 0x0bb59eceb12f1bdb63e4a7d57c70d6473fefd7c3f51af5a3604f7e97197073e4.
//
// Solidity: event NewAllocationRatio(uint16[3] newAllocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewAllocationRatio(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewAllocationRatio) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewAllocationRatio")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewAllocationRatio)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewAllocationRatio", log); err != nil {
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

// ParseNewAllocationRatio is a log parse operation binding the contract event 0x0bb59eceb12f1bdb63e4a7d57c70d6473fefd7c3f51af5a3604f7e97197073e4.
//
// Solidity: event NewAllocationRatio(uint16[3] newAllocationRatio)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewAllocationRatio(log types.Log) (*HermezAuctionProtocolNewAllocationRatio, error) {
	event := new(HermezAuctionProtocolNewAllocationRatio)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewAllocationRatio", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewBidIterator is returned from FilterNewBid and is used to iterate over the raw logs and unpacked data for NewBid events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewBidIterator struct {
	Event *HermezAuctionProtocolNewBid // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewBidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewBid)
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
		it.Event = new(HermezAuctionProtocolNewBid)
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
func (it *HermezAuctionProtocolNewBidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewBidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewBid represents a NewBid event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewBid struct {
	Slot      *big.Int
	BidAmount *big.Int
	Bidder    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNewBid is a free log retrieval operation binding the contract event 0xd48e8329cdb2fb109b4fe445d7b681a74b256bff16e6f7f33b9d4fbe9038e433.
//
// Solidity: event NewBid(uint128 indexed slot, uint128 bidAmount, address indexed bidder)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewBid(opts *bind.FilterOpts, slot []*big.Int, bidder []common.Address) (*HermezAuctionProtocolNewBidIterator, error) {

	var slotRule []interface{}
	for _, slotItem := range slot {
		slotRule = append(slotRule, slotItem)
	}

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewBid", slotRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewBidIterator{contract: _HermezAuctionProtocol.contract, event: "NewBid", logs: logs, sub: sub}, nil
}

// WatchNewBid is a free log subscription operation binding the contract event 0xd48e8329cdb2fb109b4fe445d7b681a74b256bff16e6f7f33b9d4fbe9038e433.
//
// Solidity: event NewBid(uint128 indexed slot, uint128 bidAmount, address indexed bidder)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewBid(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewBid, slot []*big.Int, bidder []common.Address) (event.Subscription, error) {

	var slotRule []interface{}
	for _, slotItem := range slot {
		slotRule = append(slotRule, slotItem)
	}

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewBid", slotRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewBid)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewBid", log); err != nil {
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

// ParseNewBid is a log parse operation binding the contract event 0xd48e8329cdb2fb109b4fe445d7b681a74b256bff16e6f7f33b9d4fbe9038e433.
//
// Solidity: event NewBid(uint128 indexed slot, uint128 bidAmount, address indexed bidder)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewBid(log types.Log) (*HermezAuctionProtocolNewBid, error) {
	event := new(HermezAuctionProtocolNewBid)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewBid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewBootCoordinatorIterator is returned from FilterNewBootCoordinator and is used to iterate over the raw logs and unpacked data for NewBootCoordinator events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewBootCoordinatorIterator struct {
	Event *HermezAuctionProtocolNewBootCoordinator // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewBootCoordinatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewBootCoordinator)
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
		it.Event = new(HermezAuctionProtocolNewBootCoordinator)
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
func (it *HermezAuctionProtocolNewBootCoordinatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewBootCoordinatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewBootCoordinator represents a NewBootCoordinator event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewBootCoordinator struct {
	NewBootCoordinator    common.Address
	NewBootCoordinatorURL string
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewBootCoordinator is a free log retrieval operation binding the contract event 0x0487eab4c1da34bf653268e33bee8bfec7dacfd6f3226047197ebf872293cfd6.
//
// Solidity: event NewBootCoordinator(address indexed newBootCoordinator, string newBootCoordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewBootCoordinator(opts *bind.FilterOpts, newBootCoordinator []common.Address) (*HermezAuctionProtocolNewBootCoordinatorIterator, error) {

	var newBootCoordinatorRule []interface{}
	for _, newBootCoordinatorItem := range newBootCoordinator {
		newBootCoordinatorRule = append(newBootCoordinatorRule, newBootCoordinatorItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewBootCoordinator", newBootCoordinatorRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewBootCoordinatorIterator{contract: _HermezAuctionProtocol.contract, event: "NewBootCoordinator", logs: logs, sub: sub}, nil
}

// WatchNewBootCoordinator is a free log subscription operation binding the contract event 0x0487eab4c1da34bf653268e33bee8bfec7dacfd6f3226047197ebf872293cfd6.
//
// Solidity: event NewBootCoordinator(address indexed newBootCoordinator, string newBootCoordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewBootCoordinator(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewBootCoordinator, newBootCoordinator []common.Address) (event.Subscription, error) {

	var newBootCoordinatorRule []interface{}
	for _, newBootCoordinatorItem := range newBootCoordinator {
		newBootCoordinatorRule = append(newBootCoordinatorRule, newBootCoordinatorItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewBootCoordinator", newBootCoordinatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewBootCoordinator)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewBootCoordinator", log); err != nil {
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

// ParseNewBootCoordinator is a log parse operation binding the contract event 0x0487eab4c1da34bf653268e33bee8bfec7dacfd6f3226047197ebf872293cfd6.
//
// Solidity: event NewBootCoordinator(address indexed newBootCoordinator, string newBootCoordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewBootCoordinator(log types.Log) (*HermezAuctionProtocolNewBootCoordinator, error) {
	event := new(HermezAuctionProtocolNewBootCoordinator)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewBootCoordinator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewClosedAuctionSlotsIterator is returned from FilterNewClosedAuctionSlots and is used to iterate over the raw logs and unpacked data for NewClosedAuctionSlots events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewClosedAuctionSlotsIterator struct {
	Event *HermezAuctionProtocolNewClosedAuctionSlots // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewClosedAuctionSlotsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewClosedAuctionSlots)
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
		it.Event = new(HermezAuctionProtocolNewClosedAuctionSlots)
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
func (it *HermezAuctionProtocolNewClosedAuctionSlotsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewClosedAuctionSlotsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewClosedAuctionSlots represents a NewClosedAuctionSlots event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewClosedAuctionSlots struct {
	NewClosedAuctionSlots uint16
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewClosedAuctionSlots is a free log retrieval operation binding the contract event 0xc78051d3757db196b1e445f3a9a1380944518c69b5d7922ec747c54f0340a4ea.
//
// Solidity: event NewClosedAuctionSlots(uint16 newClosedAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewClosedAuctionSlots(opts *bind.FilterOpts) (*HermezAuctionProtocolNewClosedAuctionSlotsIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewClosedAuctionSlots")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewClosedAuctionSlotsIterator{contract: _HermezAuctionProtocol.contract, event: "NewClosedAuctionSlots", logs: logs, sub: sub}, nil
}

// WatchNewClosedAuctionSlots is a free log subscription operation binding the contract event 0xc78051d3757db196b1e445f3a9a1380944518c69b5d7922ec747c54f0340a4ea.
//
// Solidity: event NewClosedAuctionSlots(uint16 newClosedAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewClosedAuctionSlots(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewClosedAuctionSlots) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewClosedAuctionSlots")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewClosedAuctionSlots)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewClosedAuctionSlots", log); err != nil {
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

// ParseNewClosedAuctionSlots is a log parse operation binding the contract event 0xc78051d3757db196b1e445f3a9a1380944518c69b5d7922ec747c54f0340a4ea.
//
// Solidity: event NewClosedAuctionSlots(uint16 newClosedAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewClosedAuctionSlots(log types.Log) (*HermezAuctionProtocolNewClosedAuctionSlots, error) {
	event := new(HermezAuctionProtocolNewClosedAuctionSlots)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewClosedAuctionSlots", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewDefaultSlotSetBidIterator is returned from FilterNewDefaultSlotSetBid and is used to iterate over the raw logs and unpacked data for NewDefaultSlotSetBid events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewDefaultSlotSetBidIterator struct {
	Event *HermezAuctionProtocolNewDefaultSlotSetBid // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewDefaultSlotSetBidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewDefaultSlotSetBid)
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
		it.Event = new(HermezAuctionProtocolNewDefaultSlotSetBid)
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
func (it *HermezAuctionProtocolNewDefaultSlotSetBidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewDefaultSlotSetBidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewDefaultSlotSetBid represents a NewDefaultSlotSetBid event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewDefaultSlotSetBid struct {
	SlotSet          *big.Int
	NewInitialMinBid *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNewDefaultSlotSetBid is a free log retrieval operation binding the contract event 0xa922aa010d1ff8e70b2aa9247d891836795c3d3ba2a543c37c91a44dc4a50172.
//
// Solidity: event NewDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewDefaultSlotSetBid(opts *bind.FilterOpts) (*HermezAuctionProtocolNewDefaultSlotSetBidIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewDefaultSlotSetBid")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewDefaultSlotSetBidIterator{contract: _HermezAuctionProtocol.contract, event: "NewDefaultSlotSetBid", logs: logs, sub: sub}, nil
}

// WatchNewDefaultSlotSetBid is a free log subscription operation binding the contract event 0xa922aa010d1ff8e70b2aa9247d891836795c3d3ba2a543c37c91a44dc4a50172.
//
// Solidity: event NewDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewDefaultSlotSetBid(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewDefaultSlotSetBid) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewDefaultSlotSetBid")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewDefaultSlotSetBid)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewDefaultSlotSetBid", log); err != nil {
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

// ParseNewDefaultSlotSetBid is a log parse operation binding the contract event 0xa922aa010d1ff8e70b2aa9247d891836795c3d3ba2a543c37c91a44dc4a50172.
//
// Solidity: event NewDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewDefaultSlotSetBid(log types.Log) (*HermezAuctionProtocolNewDefaultSlotSetBid, error) {
	event := new(HermezAuctionProtocolNewDefaultSlotSetBid)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewDefaultSlotSetBid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewDonationAddressIterator is returned from FilterNewDonationAddress and is used to iterate over the raw logs and unpacked data for NewDonationAddress events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewDonationAddressIterator struct {
	Event *HermezAuctionProtocolNewDonationAddress // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewDonationAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewDonationAddress)
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
		it.Event = new(HermezAuctionProtocolNewDonationAddress)
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
func (it *HermezAuctionProtocolNewDonationAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewDonationAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewDonationAddress represents a NewDonationAddress event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewDonationAddress struct {
	NewDonationAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNewDonationAddress is a free log retrieval operation binding the contract event 0xa62863cbad1647a2855e9cd39d04fa6dfd32e1b9cfaff1aaf6523f4aaafeccd7.
//
// Solidity: event NewDonationAddress(address indexed newDonationAddress)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewDonationAddress(opts *bind.FilterOpts, newDonationAddress []common.Address) (*HermezAuctionProtocolNewDonationAddressIterator, error) {

	var newDonationAddressRule []interface{}
	for _, newDonationAddressItem := range newDonationAddress {
		newDonationAddressRule = append(newDonationAddressRule, newDonationAddressItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewDonationAddress", newDonationAddressRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewDonationAddressIterator{contract: _HermezAuctionProtocol.contract, event: "NewDonationAddress", logs: logs, sub: sub}, nil
}

// WatchNewDonationAddress is a free log subscription operation binding the contract event 0xa62863cbad1647a2855e9cd39d04fa6dfd32e1b9cfaff1aaf6523f4aaafeccd7.
//
// Solidity: event NewDonationAddress(address indexed newDonationAddress)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewDonationAddress(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewDonationAddress, newDonationAddress []common.Address) (event.Subscription, error) {

	var newDonationAddressRule []interface{}
	for _, newDonationAddressItem := range newDonationAddress {
		newDonationAddressRule = append(newDonationAddressRule, newDonationAddressItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewDonationAddress", newDonationAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewDonationAddress)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewDonationAddress", log); err != nil {
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

// ParseNewDonationAddress is a log parse operation binding the contract event 0xa62863cbad1647a2855e9cd39d04fa6dfd32e1b9cfaff1aaf6523f4aaafeccd7.
//
// Solidity: event NewDonationAddress(address indexed newDonationAddress)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewDonationAddress(log types.Log) (*HermezAuctionProtocolNewDonationAddress, error) {
	event := new(HermezAuctionProtocolNewDonationAddress)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewDonationAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewForgeIterator is returned from FilterNewForge and is used to iterate over the raw logs and unpacked data for NewForge events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewForgeIterator struct {
	Event *HermezAuctionProtocolNewForge // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewForgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewForge)
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
		it.Event = new(HermezAuctionProtocolNewForge)
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
func (it *HermezAuctionProtocolNewForgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewForgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewForge represents a NewForge event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewForge struct {
	Forger      common.Address
	SlotToForge *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterNewForge is a free log retrieval operation binding the contract event 0x7cae662d4cfa9d9c5575c65f0cc41a858c51ca14ebcbd02a802a62376c3ad238.
//
// Solidity: event NewForge(address indexed forger, uint128 indexed slotToForge)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewForge(opts *bind.FilterOpts, forger []common.Address, slotToForge []*big.Int) (*HermezAuctionProtocolNewForgeIterator, error) {

	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewForge", forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewForgeIterator{contract: _HermezAuctionProtocol.contract, event: "NewForge", logs: logs, sub: sub}, nil
}

// WatchNewForge is a free log subscription operation binding the contract event 0x7cae662d4cfa9d9c5575c65f0cc41a858c51ca14ebcbd02a802a62376c3ad238.
//
// Solidity: event NewForge(address indexed forger, uint128 indexed slotToForge)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewForge(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewForge, forger []common.Address, slotToForge []*big.Int) (event.Subscription, error) {

	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewForge", forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewForge)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewForge", log); err != nil {
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

// ParseNewForge is a log parse operation binding the contract event 0x7cae662d4cfa9d9c5575c65f0cc41a858c51ca14ebcbd02a802a62376c3ad238.
//
// Solidity: event NewForge(address indexed forger, uint128 indexed slotToForge)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewForge(log types.Log) (*HermezAuctionProtocolNewForge, error) {
	event := new(HermezAuctionProtocolNewForge)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewForge", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewForgeAllocatedIterator is returned from FilterNewForgeAllocated and is used to iterate over the raw logs and unpacked data for NewForgeAllocated events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewForgeAllocatedIterator struct {
	Event *HermezAuctionProtocolNewForgeAllocated // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewForgeAllocatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewForgeAllocated)
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
		it.Event = new(HermezAuctionProtocolNewForgeAllocated)
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
func (it *HermezAuctionProtocolNewForgeAllocatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewForgeAllocatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewForgeAllocated represents a NewForgeAllocated event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewForgeAllocated struct {
	Bidder           common.Address
	Forger           common.Address
	SlotToForge      *big.Int
	BurnAmount       *big.Int
	DonationAmount   *big.Int
	GovernanceAmount *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNewForgeAllocated is a free log retrieval operation binding the contract event 0xd64ebb43f4c2b91022b97389834432f1027ef55586129ba05a3a3065b2304f05.
//
// Solidity: event NewForgeAllocated(address indexed bidder, address indexed forger, uint128 indexed slotToForge, uint128 burnAmount, uint128 donationAmount, uint128 governanceAmount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewForgeAllocated(opts *bind.FilterOpts, bidder []common.Address, forger []common.Address, slotToForge []*big.Int) (*HermezAuctionProtocolNewForgeAllocatedIterator, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewForgeAllocated", bidderRule, forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewForgeAllocatedIterator{contract: _HermezAuctionProtocol.contract, event: "NewForgeAllocated", logs: logs, sub: sub}, nil
}

// WatchNewForgeAllocated is a free log subscription operation binding the contract event 0xd64ebb43f4c2b91022b97389834432f1027ef55586129ba05a3a3065b2304f05.
//
// Solidity: event NewForgeAllocated(address indexed bidder, address indexed forger, uint128 indexed slotToForge, uint128 burnAmount, uint128 donationAmount, uint128 governanceAmount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewForgeAllocated(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewForgeAllocated, bidder []common.Address, forger []common.Address, slotToForge []*big.Int) (event.Subscription, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewForgeAllocated", bidderRule, forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewForgeAllocated)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewForgeAllocated", log); err != nil {
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

// ParseNewForgeAllocated is a log parse operation binding the contract event 0xd64ebb43f4c2b91022b97389834432f1027ef55586129ba05a3a3065b2304f05.
//
// Solidity: event NewForgeAllocated(address indexed bidder, address indexed forger, uint128 indexed slotToForge, uint128 burnAmount, uint128 donationAmount, uint128 governanceAmount)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewForgeAllocated(log types.Log) (*HermezAuctionProtocolNewForgeAllocated, error) {
	event := new(HermezAuctionProtocolNewForgeAllocated)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewForgeAllocated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewOpenAuctionSlotsIterator is returned from FilterNewOpenAuctionSlots and is used to iterate over the raw logs and unpacked data for NewOpenAuctionSlots events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewOpenAuctionSlotsIterator struct {
	Event *HermezAuctionProtocolNewOpenAuctionSlots // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewOpenAuctionSlotsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewOpenAuctionSlots)
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
		it.Event = new(HermezAuctionProtocolNewOpenAuctionSlots)
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
func (it *HermezAuctionProtocolNewOpenAuctionSlotsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewOpenAuctionSlotsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewOpenAuctionSlots represents a NewOpenAuctionSlots event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewOpenAuctionSlots struct {
	NewOpenAuctionSlots uint16
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterNewOpenAuctionSlots is a free log retrieval operation binding the contract event 0x3da0492dea7298351bc14d1c0699905fd0657c33487449751af50fc0c8b593f1.
//
// Solidity: event NewOpenAuctionSlots(uint16 newOpenAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewOpenAuctionSlots(opts *bind.FilterOpts) (*HermezAuctionProtocolNewOpenAuctionSlotsIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewOpenAuctionSlots")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewOpenAuctionSlotsIterator{contract: _HermezAuctionProtocol.contract, event: "NewOpenAuctionSlots", logs: logs, sub: sub}, nil
}

// WatchNewOpenAuctionSlots is a free log subscription operation binding the contract event 0x3da0492dea7298351bc14d1c0699905fd0657c33487449751af50fc0c8b593f1.
//
// Solidity: event NewOpenAuctionSlots(uint16 newOpenAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewOpenAuctionSlots(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewOpenAuctionSlots) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewOpenAuctionSlots")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewOpenAuctionSlots)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewOpenAuctionSlots", log); err != nil {
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

// ParseNewOpenAuctionSlots is a log parse operation binding the contract event 0x3da0492dea7298351bc14d1c0699905fd0657c33487449751af50fc0c8b593f1.
//
// Solidity: event NewOpenAuctionSlots(uint16 newOpenAuctionSlots)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewOpenAuctionSlots(log types.Log) (*HermezAuctionProtocolNewOpenAuctionSlots, error) {
	event := new(HermezAuctionProtocolNewOpenAuctionSlots)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewOpenAuctionSlots", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewOutbiddingIterator is returned from FilterNewOutbidding and is used to iterate over the raw logs and unpacked data for NewOutbidding events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewOutbiddingIterator struct {
	Event *HermezAuctionProtocolNewOutbidding // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewOutbiddingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewOutbidding)
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
		it.Event = new(HermezAuctionProtocolNewOutbidding)
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
func (it *HermezAuctionProtocolNewOutbiddingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewOutbiddingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewOutbidding represents a NewOutbidding event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewOutbidding struct {
	NewOutbidding uint16
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewOutbidding is a free log retrieval operation binding the contract event 0xd3748b8c326e93d12af934fbf87471e315a89bc3f7b8222343acf0210edf248e.
//
// Solidity: event NewOutbidding(uint16 newOutbidding)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewOutbidding(opts *bind.FilterOpts) (*HermezAuctionProtocolNewOutbiddingIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewOutbidding")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewOutbiddingIterator{contract: _HermezAuctionProtocol.contract, event: "NewOutbidding", logs: logs, sub: sub}, nil
}

// WatchNewOutbidding is a free log subscription operation binding the contract event 0xd3748b8c326e93d12af934fbf87471e315a89bc3f7b8222343acf0210edf248e.
//
// Solidity: event NewOutbidding(uint16 newOutbidding)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewOutbidding(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewOutbidding) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewOutbidding")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewOutbidding)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewOutbidding", log); err != nil {
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

// ParseNewOutbidding is a log parse operation binding the contract event 0xd3748b8c326e93d12af934fbf87471e315a89bc3f7b8222343acf0210edf248e.
//
// Solidity: event NewOutbidding(uint16 newOutbidding)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewOutbidding(log types.Log) (*HermezAuctionProtocolNewOutbidding, error) {
	event := new(HermezAuctionProtocolNewOutbidding)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewOutbidding", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolNewSlotDeadlineIterator is returned from FilterNewSlotDeadline and is used to iterate over the raw logs and unpacked data for NewSlotDeadline events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewSlotDeadlineIterator struct {
	Event *HermezAuctionProtocolNewSlotDeadline // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolNewSlotDeadlineIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolNewSlotDeadline)
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
		it.Event = new(HermezAuctionProtocolNewSlotDeadline)
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
func (it *HermezAuctionProtocolNewSlotDeadlineIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolNewSlotDeadlineIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolNewSlotDeadline represents a NewSlotDeadline event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolNewSlotDeadline struct {
	NewSlotDeadline uint8
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNewSlotDeadline is a free log retrieval operation binding the contract event 0x4a0d90b611c15e02dbf23b10f35b936cf2c77665f8c77822d3eca131f9d986d3.
//
// Solidity: event NewSlotDeadline(uint8 newSlotDeadline)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterNewSlotDeadline(opts *bind.FilterOpts) (*HermezAuctionProtocolNewSlotDeadlineIterator, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "NewSlotDeadline")
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolNewSlotDeadlineIterator{contract: _HermezAuctionProtocol.contract, event: "NewSlotDeadline", logs: logs, sub: sub}, nil
}

// WatchNewSlotDeadline is a free log subscription operation binding the contract event 0x4a0d90b611c15e02dbf23b10f35b936cf2c77665f8c77822d3eca131f9d986d3.
//
// Solidity: event NewSlotDeadline(uint8 newSlotDeadline)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchNewSlotDeadline(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolNewSlotDeadline) (event.Subscription, error) {

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "NewSlotDeadline")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolNewSlotDeadline)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewSlotDeadline", log); err != nil {
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

// ParseNewSlotDeadline is a log parse operation binding the contract event 0x4a0d90b611c15e02dbf23b10f35b936cf2c77665f8c77822d3eca131f9d986d3.
//
// Solidity: event NewSlotDeadline(uint8 newSlotDeadline)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseNewSlotDeadline(log types.Log) (*HermezAuctionProtocolNewSlotDeadline, error) {
	event := new(HermezAuctionProtocolNewSlotDeadline)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "NewSlotDeadline", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezAuctionProtocolSetCoordinatorIterator is returned from FilterSetCoordinator and is used to iterate over the raw logs and unpacked data for SetCoordinator events raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolSetCoordinatorIterator struct {
	Event *HermezAuctionProtocolSetCoordinator // Event containing the contract specifics and raw log

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
func (it *HermezAuctionProtocolSetCoordinatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAuctionProtocolSetCoordinator)
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
		it.Event = new(HermezAuctionProtocolSetCoordinator)
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
func (it *HermezAuctionProtocolSetCoordinatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAuctionProtocolSetCoordinatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAuctionProtocolSetCoordinator represents a SetCoordinator event raised by the HermezAuctionProtocol contract.
type HermezAuctionProtocolSetCoordinator struct {
	Bidder         common.Address
	Forger         common.Address
	CoordinatorURL string
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetCoordinator is a free log retrieval operation binding the contract event 0x5246b2ac9ee77efe2e64af6df00055d97e2d6e1b277f5a8d17ba5bca1a573da0.
//
// Solidity: event SetCoordinator(address indexed bidder, address indexed forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) FilterSetCoordinator(opts *bind.FilterOpts, bidder []common.Address, forger []common.Address) (*HermezAuctionProtocolSetCoordinatorIterator, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.FilterLogs(opts, "SetCoordinator", bidderRule, forgerRule)
	if err != nil {
		return nil, err
	}
	return &HermezAuctionProtocolSetCoordinatorIterator{contract: _HermezAuctionProtocol.contract, event: "SetCoordinator", logs: logs, sub: sub}, nil
}

// WatchSetCoordinator is a free log subscription operation binding the contract event 0x5246b2ac9ee77efe2e64af6df00055d97e2d6e1b277f5a8d17ba5bca1a573da0.
//
// Solidity: event SetCoordinator(address indexed bidder, address indexed forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) WatchSetCoordinator(opts *bind.WatchOpts, sink chan<- *HermezAuctionProtocolSetCoordinator, bidder []common.Address, forger []common.Address) (event.Subscription, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}

	logs, sub, err := _HermezAuctionProtocol.contract.WatchLogs(opts, "SetCoordinator", bidderRule, forgerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAuctionProtocolSetCoordinator)
				if err := _HermezAuctionProtocol.contract.UnpackLog(event, "SetCoordinator", log); err != nil {
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

// ParseSetCoordinator is a log parse operation binding the contract event 0x5246b2ac9ee77efe2e64af6df00055d97e2d6e1b277f5a8d17ba5bca1a573da0.
//
// Solidity: event SetCoordinator(address indexed bidder, address indexed forger, string coordinatorURL)
func (_HermezAuctionProtocol *HermezAuctionProtocolFilterer) ParseSetCoordinator(log types.Log) (*HermezAuctionProtocolSetCoordinator, error) {
	event := new(HermezAuctionProtocolSetCoordinator)
	if err := _HermezAuctionProtocol.contract.UnpackLog(event, "SetCoordinator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
