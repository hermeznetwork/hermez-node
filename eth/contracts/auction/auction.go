// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package auction

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

// AuctionABI is the input ABI used to generate the binding from.
const AuctionABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"}],\"name\":\"HEZClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"donationAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"bootCoordinatorAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"bootCoordinatorURL\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"outbidding\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"slotDeadline\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"closedAuctionSlots\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"openAuctionSlots\",\"type\":\"uint16\"},{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"allocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"InitializeHermezAuctionProtocolEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16[3]\",\"name\":\"newAllocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"NewAllocationRatio\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"}],\"name\":\"NewBid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newBootCoordinator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"newBootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"NewBootCoordinator\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newClosedAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"NewClosedAuctionSlots\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"slotSet\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"newInitialMinBid\",\"type\":\"uint128\"}],\"name\":\"NewDefaultSlotSetBid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newDonationAddress\",\"type\":\"address\"}],\"name\":\"NewDonationAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slotToForge\",\"type\":\"uint128\"}],\"name\":\"NewForge\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"slotToForge\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"burnAmount\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"donationAmount\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"governanceAmount\",\"type\":\"uint128\"}],\"name\":\"NewForgeAllocated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newOpenAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"NewOpenAuctionSlots\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"newOutbidding\",\"type\":\"uint16\"}],\"name\":\"NewOutbidding\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"newSlotDeadline\",\"type\":\"uint8\"}],\"name\":\"NewSlotDeadline\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"name\":\"SetCoordinator\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BLOCKS_PER_SLOT\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"INITIAL_MINIMAL_BIDDING\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bootCoordinatorURL\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"canForge\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slotSet\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"newInitialMinBid\",\"type\":\"uint128\"}],\"name\":\"changeDefaultSlotSetBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimHEZ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"claimPendingHEZ\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"coordinators\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"}],\"name\":\"forge\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"genesisBlock\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAllocationRatio\",\"outputs\":[{\"internalType\":\"uint16[3]\",\"name\":\"\",\"type\":\"uint16[3]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBootCoordinator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"}],\"name\":\"getClaimableHEZ\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getClosedAuctionSlots\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentSlotNumber\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"slotSet\",\"type\":\"uint8\"}],\"name\":\"getDefaultSlotSetBid\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDonationAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"getMinBidBySlot\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOpenAuctionSlots\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getOutbidding\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSlotDeadline\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"blockNumber\",\"type\":\"uint128\"}],\"name\":\"getSlotNumber\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"}],\"name\":\"getSlotSet\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"uint128\",\"name\":\"genesis\",\"type\":\"uint128\"},{\"internalType\":\"address\",\"name\":\"hermezRollupAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_governanceAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"donationAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"bootCoordinatorAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_bootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"hermezAuctionProtocolInitializer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollup\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"pendingBalances\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"slot\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"processBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"amount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"startingSlot\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"endingSlot\",\"type\":\"uint128\"},{\"internalType\":\"bool[6]\",\"name\":\"slotSets\",\"type\":\"bool[6]\"},{\"internalType\":\"uint128\",\"name\":\"maxBid\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"minBid\",\"type\":\"uint128\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"processMultiBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16[3]\",\"name\":\"newAllocationRatio\",\"type\":\"uint16[3]\"}],\"name\":\"setAllocationRatio\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newBootCoordinator\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"newBootCoordinatorURL\",\"type\":\"string\"}],\"name\":\"setBootCoordinator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newClosedAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"setClosedAuctionSlots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"forger\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"coordinatorURL\",\"type\":\"string\"}],\"name\":\"setCoordinator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newDonationAddress\",\"type\":\"address\"}],\"name\":\"setDonationAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newOpenAuctionSlots\",\"type\":\"uint16\"}],\"name\":\"setOpenAuctionSlots\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"newOutbidding\",\"type\":\"uint16\"}],\"name\":\"setOutbidding\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newDeadline\",\"type\":\"uint8\"}],\"name\":\"setSlotDeadline\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"name\":\"slots\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"bidder\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"fulfilled\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"forgerCommitment\",\"type\":\"bool\"},{\"internalType\":\"uint128\",\"name\":\"bidAmount\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"closedMinBid\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenHEZ\",\"outputs\":[{\"internalType\":\"contractIHEZToken\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// Auction is an auto generated Go binding around an Ethereum contract.
type Auction struct {
	AuctionCaller     // Read-only binding to the contract
	AuctionTransactor // Write-only binding to the contract
	AuctionFilterer   // Log filterer for contract events
}

// AuctionCaller is an auto generated read-only Go binding around an Ethereum contract.
type AuctionCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuctionTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AuctionTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuctionFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AuctionFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AuctionSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AuctionSession struct {
	Contract     *Auction          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AuctionCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AuctionCallerSession struct {
	Contract *AuctionCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// AuctionTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AuctionTransactorSession struct {
	Contract     *AuctionTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// AuctionRaw is an auto generated low-level Go binding around an Ethereum contract.
type AuctionRaw struct {
	Contract *Auction // Generic contract binding to access the raw methods on
}

// AuctionCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AuctionCallerRaw struct {
	Contract *AuctionCaller // Generic read-only contract binding to access the raw methods on
}

// AuctionTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AuctionTransactorRaw struct {
	Contract *AuctionTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAuction creates a new instance of Auction, bound to a specific deployed contract.
func NewAuction(address common.Address, backend bind.ContractBackend) (*Auction, error) {
	contract, err := bindAuction(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Auction{AuctionCaller: AuctionCaller{contract: contract}, AuctionTransactor: AuctionTransactor{contract: contract}, AuctionFilterer: AuctionFilterer{contract: contract}}, nil
}

// NewAuctionCaller creates a new read-only instance of Auction, bound to a specific deployed contract.
func NewAuctionCaller(address common.Address, caller bind.ContractCaller) (*AuctionCaller, error) {
	contract, err := bindAuction(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AuctionCaller{contract: contract}, nil
}

// NewAuctionTransactor creates a new write-only instance of Auction, bound to a specific deployed contract.
func NewAuctionTransactor(address common.Address, transactor bind.ContractTransactor) (*AuctionTransactor, error) {
	contract, err := bindAuction(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AuctionTransactor{contract: contract}, nil
}

// NewAuctionFilterer creates a new log filterer instance of Auction, bound to a specific deployed contract.
func NewAuctionFilterer(address common.Address, filterer bind.ContractFilterer) (*AuctionFilterer, error) {
	contract, err := bindAuction(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AuctionFilterer{contract: contract}, nil
}

// bindAuction binds a generic wrapper to an already deployed contract.
func bindAuction(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AuctionABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Auction *AuctionRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Auction.Contract.AuctionCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Auction *AuctionRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Auction.Contract.AuctionTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Auction *AuctionRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Auction.Contract.AuctionTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Auction *AuctionCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Auction.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Auction *AuctionTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Auction.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Auction *AuctionTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Auction.Contract.contract.Transact(opts, method, params...)
}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_Auction *AuctionCaller) BLOCKSPERSLOT(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "BLOCKS_PER_SLOT")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_Auction *AuctionSession) BLOCKSPERSLOT() (uint8, error) {
	return _Auction.Contract.BLOCKSPERSLOT(&_Auction.CallOpts)
}

// BLOCKSPERSLOT is a free data retrieval call binding the contract method 0x2243de47.
//
// Solidity: function BLOCKS_PER_SLOT() view returns(uint8)
func (_Auction *AuctionCallerSession) BLOCKSPERSLOT() (uint8, error) {
	return _Auction.Contract.BLOCKSPERSLOT(&_Auction.CallOpts)
}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_Auction *AuctionCaller) INITIALMINIMALBIDDING(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "INITIAL_MINIMAL_BIDDING")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_Auction *AuctionSession) INITIALMINIMALBIDDING() (*big.Int, error) {
	return _Auction.Contract.INITIALMINIMALBIDDING(&_Auction.CallOpts)
}

// INITIALMINIMALBIDDING is a free data retrieval call binding the contract method 0xe6065914.
//
// Solidity: function INITIAL_MINIMAL_BIDDING() view returns(uint128)
func (_Auction *AuctionCallerSession) INITIALMINIMALBIDDING() (*big.Int, error) {
	return _Auction.Contract.INITIALMINIMALBIDDING(&_Auction.CallOpts)
}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_Auction *AuctionCaller) BootCoordinatorURL(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "bootCoordinatorURL")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_Auction *AuctionSession) BootCoordinatorURL() (string, error) {
	return _Auction.Contract.BootCoordinatorURL(&_Auction.CallOpts)
}

// BootCoordinatorURL is a free data retrieval call binding the contract method 0x72ca58a3.
//
// Solidity: function bootCoordinatorURL() view returns(string)
func (_Auction *AuctionCallerSession) BootCoordinatorURL() (string, error) {
	return _Auction.Contract.BootCoordinatorURL(&_Auction.CallOpts)
}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_Auction *AuctionCaller) CanForge(opts *bind.CallOpts, forger common.Address, blockNumber *big.Int) (bool, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "canForge", forger, blockNumber)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_Auction *AuctionSession) CanForge(forger common.Address, blockNumber *big.Int) (bool, error) {
	return _Auction.Contract.CanForge(&_Auction.CallOpts, forger, blockNumber)
}

// CanForge is a free data retrieval call binding the contract method 0x83b1f6a0.
//
// Solidity: function canForge(address forger, uint256 blockNumber) view returns(bool)
func (_Auction *AuctionCallerSession) CanForge(forger common.Address, blockNumber *big.Int) (bool, error) {
	return _Auction.Contract.CanForge(&_Auction.CallOpts, forger, blockNumber)
}

// Coordinators is a free data retrieval call binding the contract method 0xa48af096.
//
// Solidity: function coordinators(address ) view returns(address forger, string coordinatorURL)
func (_Auction *AuctionCaller) Coordinators(opts *bind.CallOpts, arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "coordinators", arg0)

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
func (_Auction *AuctionSession) Coordinators(arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	return _Auction.Contract.Coordinators(&_Auction.CallOpts, arg0)
}

// Coordinators is a free data retrieval call binding the contract method 0xa48af096.
//
// Solidity: function coordinators(address ) view returns(address forger, string coordinatorURL)
func (_Auction *AuctionCallerSession) Coordinators(arg0 common.Address) (struct {
	Forger         common.Address
	CoordinatorURL string
}, error) {
	return _Auction.Contract.Coordinators(&_Auction.CallOpts, arg0)
}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_Auction *AuctionCaller) GenesisBlock(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "genesisBlock")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_Auction *AuctionSession) GenesisBlock() (*big.Int, error) {
	return _Auction.Contract.GenesisBlock(&_Auction.CallOpts)
}

// GenesisBlock is a free data retrieval call binding the contract method 0x4cdc9c63.
//
// Solidity: function genesisBlock() view returns(uint128)
func (_Auction *AuctionCallerSession) GenesisBlock() (*big.Int, error) {
	return _Auction.Contract.GenesisBlock(&_Auction.CallOpts)
}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_Auction *AuctionCaller) GetAllocationRatio(opts *bind.CallOpts) ([3]uint16, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getAllocationRatio")

	if err != nil {
		return *new([3]uint16), err
	}

	out0 := *abi.ConvertType(out[0], new([3]uint16)).(*[3]uint16)

	return out0, err

}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_Auction *AuctionSession) GetAllocationRatio() ([3]uint16, error) {
	return _Auction.Contract.GetAllocationRatio(&_Auction.CallOpts)
}

// GetAllocationRatio is a free data retrieval call binding the contract method 0xec29159b.
//
// Solidity: function getAllocationRatio() view returns(uint16[3])
func (_Auction *AuctionCallerSession) GetAllocationRatio() ([3]uint16, error) {
	return _Auction.Contract.GetAllocationRatio(&_Auction.CallOpts)
}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_Auction *AuctionCaller) GetBootCoordinator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getBootCoordinator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_Auction *AuctionSession) GetBootCoordinator() (common.Address, error) {
	return _Auction.Contract.GetBootCoordinator(&_Auction.CallOpts)
}

// GetBootCoordinator is a free data retrieval call binding the contract method 0xb5f7f2f0.
//
// Solidity: function getBootCoordinator() view returns(address)
func (_Auction *AuctionCallerSession) GetBootCoordinator() (common.Address, error) {
	return _Auction.Contract.GetBootCoordinator(&_Auction.CallOpts)
}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_Auction *AuctionCaller) GetClaimableHEZ(opts *bind.CallOpts, bidder common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getClaimableHEZ", bidder)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_Auction *AuctionSession) GetClaimableHEZ(bidder common.Address) (*big.Int, error) {
	return _Auction.Contract.GetClaimableHEZ(&_Auction.CallOpts, bidder)
}

// GetClaimableHEZ is a free data retrieval call binding the contract method 0x5cca4903.
//
// Solidity: function getClaimableHEZ(address bidder) view returns(uint128)
func (_Auction *AuctionCallerSession) GetClaimableHEZ(bidder common.Address) (*big.Int, error) {
	return _Auction.Contract.GetClaimableHEZ(&_Auction.CallOpts, bidder)
}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_Auction *AuctionCaller) GetClosedAuctionSlots(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getClosedAuctionSlots")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_Auction *AuctionSession) GetClosedAuctionSlots() (uint16, error) {
	return _Auction.Contract.GetClosedAuctionSlots(&_Auction.CallOpts)
}

// GetClosedAuctionSlots is a free data retrieval call binding the contract method 0x4da9639d.
//
// Solidity: function getClosedAuctionSlots() view returns(uint16)
func (_Auction *AuctionCallerSession) GetClosedAuctionSlots() (uint16, error) {
	return _Auction.Contract.GetClosedAuctionSlots(&_Auction.CallOpts)
}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_Auction *AuctionCaller) GetCurrentSlotNumber(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getCurrentSlotNumber")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_Auction *AuctionSession) GetCurrentSlotNumber() (*big.Int, error) {
	return _Auction.Contract.GetCurrentSlotNumber(&_Auction.CallOpts)
}

// GetCurrentSlotNumber is a free data retrieval call binding the contract method 0x0c4da4f6.
//
// Solidity: function getCurrentSlotNumber() view returns(uint128)
func (_Auction *AuctionCallerSession) GetCurrentSlotNumber() (*big.Int, error) {
	return _Auction.Contract.GetCurrentSlotNumber(&_Auction.CallOpts)
}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_Auction *AuctionCaller) GetDefaultSlotSetBid(opts *bind.CallOpts, slotSet uint8) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getDefaultSlotSetBid", slotSet)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_Auction *AuctionSession) GetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	return _Auction.Contract.GetDefaultSlotSetBid(&_Auction.CallOpts, slotSet)
}

// GetDefaultSlotSetBid is a free data retrieval call binding the contract method 0x564e6a71.
//
// Solidity: function getDefaultSlotSetBid(uint8 slotSet) view returns(uint128)
func (_Auction *AuctionCallerSession) GetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	return _Auction.Contract.GetDefaultSlotSetBid(&_Auction.CallOpts, slotSet)
}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_Auction *AuctionCaller) GetDonationAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getDonationAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_Auction *AuctionSession) GetDonationAddress() (common.Address, error) {
	return _Auction.Contract.GetDonationAddress(&_Auction.CallOpts)
}

// GetDonationAddress is a free data retrieval call binding the contract method 0x54c03ab7.
//
// Solidity: function getDonationAddress() view returns(address)
func (_Auction *AuctionCallerSession) GetDonationAddress() (common.Address, error) {
	return _Auction.Contract.GetDonationAddress(&_Auction.CallOpts)
}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_Auction *AuctionCaller) GetMinBidBySlot(opts *bind.CallOpts, slot *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getMinBidBySlot", slot)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_Auction *AuctionSession) GetMinBidBySlot(slot *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetMinBidBySlot(&_Auction.CallOpts, slot)
}

// GetMinBidBySlot is a free data retrieval call binding the contract method 0x37d1bd0b.
//
// Solidity: function getMinBidBySlot(uint128 slot) view returns(uint128)
func (_Auction *AuctionCallerSession) GetMinBidBySlot(slot *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetMinBidBySlot(&_Auction.CallOpts, slot)
}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_Auction *AuctionCaller) GetOpenAuctionSlots(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getOpenAuctionSlots")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_Auction *AuctionSession) GetOpenAuctionSlots() (uint16, error) {
	return _Auction.Contract.GetOpenAuctionSlots(&_Auction.CallOpts)
}

// GetOpenAuctionSlots is a free data retrieval call binding the contract method 0xac4b9012.
//
// Solidity: function getOpenAuctionSlots() view returns(uint16)
func (_Auction *AuctionCallerSession) GetOpenAuctionSlots() (uint16, error) {
	return _Auction.Contract.GetOpenAuctionSlots(&_Auction.CallOpts)
}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_Auction *AuctionCaller) GetOutbidding(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getOutbidding")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_Auction *AuctionSession) GetOutbidding() (uint16, error) {
	return _Auction.Contract.GetOutbidding(&_Auction.CallOpts)
}

// GetOutbidding is a free data retrieval call binding the contract method 0x55b442e6.
//
// Solidity: function getOutbidding() view returns(uint16)
func (_Auction *AuctionCallerSession) GetOutbidding() (uint16, error) {
	return _Auction.Contract.GetOutbidding(&_Auction.CallOpts)
}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_Auction *AuctionCaller) GetSlotDeadline(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getSlotDeadline")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_Auction *AuctionSession) GetSlotDeadline() (uint8, error) {
	return _Auction.Contract.GetSlotDeadline(&_Auction.CallOpts)
}

// GetSlotDeadline is a free data retrieval call binding the contract method 0x13de9af2.
//
// Solidity: function getSlotDeadline() view returns(uint8)
func (_Auction *AuctionCallerSession) GetSlotDeadline() (uint8, error) {
	return _Auction.Contract.GetSlotDeadline(&_Auction.CallOpts)
}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_Auction *AuctionCaller) GetSlotNumber(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getSlotNumber", blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_Auction *AuctionSession) GetSlotNumber(blockNumber *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetSlotNumber(&_Auction.CallOpts, blockNumber)
}

// GetSlotNumber is a free data retrieval call binding the contract method 0xb3dc7bb1.
//
// Solidity: function getSlotNumber(uint128 blockNumber) view returns(uint128)
func (_Auction *AuctionCallerSession) GetSlotNumber(blockNumber *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetSlotNumber(&_Auction.CallOpts, blockNumber)
}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_Auction *AuctionCaller) GetSlotSet(opts *bind.CallOpts, slot *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "getSlotSet", slot)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_Auction *AuctionSession) GetSlotSet(slot *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetSlotSet(&_Auction.CallOpts, slot)
}

// GetSlotSet is a free data retrieval call binding the contract method 0xac5f658b.
//
// Solidity: function getSlotSet(uint128 slot) view returns(uint128)
func (_Auction *AuctionCallerSession) GetSlotSet(slot *big.Int) (*big.Int, error) {
	return _Auction.Contract.GetSlotSet(&_Auction.CallOpts, slot)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_Auction *AuctionCaller) GovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "governanceAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_Auction *AuctionSession) GovernanceAddress() (common.Address, error) {
	return _Auction.Contract.GovernanceAddress(&_Auction.CallOpts)
}

// GovernanceAddress is a free data retrieval call binding the contract method 0x795053d3.
//
// Solidity: function governanceAddress() view returns(address)
func (_Auction *AuctionCallerSession) GovernanceAddress() (common.Address, error) {
	return _Auction.Contract.GovernanceAddress(&_Auction.CallOpts)
}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_Auction *AuctionCaller) HermezRollup(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "hermezRollup")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_Auction *AuctionSession) HermezRollup() (common.Address, error) {
	return _Auction.Contract.HermezRollup(&_Auction.CallOpts)
}

// HermezRollup is a free data retrieval call binding the contract method 0xaebd6d98.
//
// Solidity: function hermezRollup() view returns(address)
func (_Auction *AuctionCallerSession) HermezRollup() (common.Address, error) {
	return _Auction.Contract.HermezRollup(&_Auction.CallOpts)
}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_Auction *AuctionCaller) PendingBalances(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "pendingBalances", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_Auction *AuctionSession) PendingBalances(arg0 common.Address) (*big.Int, error) {
	return _Auction.Contract.PendingBalances(&_Auction.CallOpts, arg0)
}

// PendingBalances is a free data retrieval call binding the contract method 0xecdae41b.
//
// Solidity: function pendingBalances(address ) view returns(uint128)
func (_Auction *AuctionCallerSession) PendingBalances(arg0 common.Address) (*big.Int, error) {
	return _Auction.Contract.PendingBalances(&_Auction.CallOpts, arg0)
}

// Slots is a free data retrieval call binding the contract method 0xbc415567.
//
// Solidity: function slots(uint128 ) view returns(address bidder, bool fulfilled, bool forgerCommitment, uint128 bidAmount, uint128 closedMinBid)
func (_Auction *AuctionCaller) Slots(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "slots", arg0)

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
func (_Auction *AuctionSession) Slots(arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	return _Auction.Contract.Slots(&_Auction.CallOpts, arg0)
}

// Slots is a free data retrieval call binding the contract method 0xbc415567.
//
// Solidity: function slots(uint128 ) view returns(address bidder, bool fulfilled, bool forgerCommitment, uint128 bidAmount, uint128 closedMinBid)
func (_Auction *AuctionCallerSession) Slots(arg0 *big.Int) (struct {
	Bidder           common.Address
	Fulfilled        bool
	ForgerCommitment bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}, error) {
	return _Auction.Contract.Slots(&_Auction.CallOpts, arg0)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Auction *AuctionCaller) TokenHEZ(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Auction.contract.Call(opts, &out, "tokenHEZ")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Auction *AuctionSession) TokenHEZ() (common.Address, error) {
	return _Auction.Contract.TokenHEZ(&_Auction.CallOpts)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Auction *AuctionCallerSession) TokenHEZ() (common.Address, error) {
	return _Auction.Contract.TokenHEZ(&_Auction.CallOpts)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_Auction *AuctionTransactor) ChangeDefaultSlotSetBid(opts *bind.TransactOpts, slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "changeDefaultSlotSetBid", slotSet, newInitialMinBid)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_Auction *AuctionSession) ChangeDefaultSlotSetBid(slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _Auction.Contract.ChangeDefaultSlotSetBid(&_Auction.TransactOpts, slotSet, newInitialMinBid)
}

// ChangeDefaultSlotSetBid is a paid mutator transaction binding the contract method 0x7c643b70.
//
// Solidity: function changeDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid) returns()
func (_Auction *AuctionTransactorSession) ChangeDefaultSlotSetBid(slotSet *big.Int, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return _Auction.Contract.ChangeDefaultSlotSetBid(&_Auction.TransactOpts, slotSet, newInitialMinBid)
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_Auction *AuctionTransactor) ClaimHEZ(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "claimHEZ")
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_Auction *AuctionSession) ClaimHEZ() (*types.Transaction, error) {
	return _Auction.Contract.ClaimHEZ(&_Auction.TransactOpts)
}

// ClaimHEZ is a paid mutator transaction binding the contract method 0x6dfe47c9.
//
// Solidity: function claimHEZ() returns()
func (_Auction *AuctionTransactorSession) ClaimHEZ() (*types.Transaction, error) {
	return _Auction.Contract.ClaimHEZ(&_Auction.TransactOpts)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_Auction *AuctionTransactor) ClaimPendingHEZ(opts *bind.TransactOpts, slot *big.Int) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "claimPendingHEZ", slot)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_Auction *AuctionSession) ClaimPendingHEZ(slot *big.Int) (*types.Transaction, error) {
	return _Auction.Contract.ClaimPendingHEZ(&_Auction.TransactOpts, slot)
}

// ClaimPendingHEZ is a paid mutator transaction binding the contract method 0x41d42c23.
//
// Solidity: function claimPendingHEZ(uint128 slot) returns()
func (_Auction *AuctionTransactorSession) ClaimPendingHEZ(slot *big.Int) (*types.Transaction, error) {
	return _Auction.Contract.ClaimPendingHEZ(&_Auction.TransactOpts, slot)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_Auction *AuctionTransactor) Forge(opts *bind.TransactOpts, forger common.Address) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "forge", forger)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_Auction *AuctionSession) Forge(forger common.Address) (*types.Transaction, error) {
	return _Auction.Contract.Forge(&_Auction.TransactOpts, forger)
}

// Forge is a paid mutator transaction binding the contract method 0x4e5a5178.
//
// Solidity: function forge(address forger) returns()
func (_Auction *AuctionTransactorSession) Forge(forger common.Address) (*types.Transaction, error) {
	return _Auction.Contract.Forge(&_Auction.TransactOpts, forger)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_Auction *AuctionTransactor) HermezAuctionProtocolInitializer(opts *bind.TransactOpts, token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "hermezAuctionProtocolInitializer", token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_Auction *AuctionSession) HermezAuctionProtocolInitializer(token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.HermezAuctionProtocolInitializer(&_Auction.TransactOpts, token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// HermezAuctionProtocolInitializer is a paid mutator transaction binding the contract method 0x5e73a67f.
//
// Solidity: function hermezAuctionProtocolInitializer(address token, uint128 genesis, address hermezRollupAddress, address _governanceAddress, address donationAddress, address bootCoordinatorAddress, string _bootCoordinatorURL) returns()
func (_Auction *AuctionTransactorSession) HermezAuctionProtocolInitializer(token common.Address, genesis *big.Int, hermezRollupAddress common.Address, _governanceAddress common.Address, donationAddress common.Address, bootCoordinatorAddress common.Address, _bootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.HermezAuctionProtocolInitializer(&_Auction.TransactOpts, token, genesis, hermezRollupAddress, _governanceAddress, donationAddress, bootCoordinatorAddress, _bootCoordinatorURL)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_Auction *AuctionTransactor) ProcessBid(opts *bind.TransactOpts, amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "processBid", amount, slot, bidAmount, permit)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_Auction *AuctionSession) ProcessBid(amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.Contract.ProcessBid(&_Auction.TransactOpts, amount, slot, bidAmount, permit)
}

// ProcessBid is a paid mutator transaction binding the contract method 0x4b93b7fa.
//
// Solidity: function processBid(uint128 amount, uint128 slot, uint128 bidAmount, bytes permit) returns()
func (_Auction *AuctionTransactorSession) ProcessBid(amount *big.Int, slot *big.Int, bidAmount *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.Contract.ProcessBid(&_Auction.TransactOpts, amount, slot, bidAmount, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_Auction *AuctionTransactor) ProcessMultiBid(opts *bind.TransactOpts, amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "processMultiBid", amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_Auction *AuctionSession) ProcessMultiBid(amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.Contract.ProcessMultiBid(&_Auction.TransactOpts, amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// ProcessMultiBid is a paid mutator transaction binding the contract method 0x583ad0dd.
//
// Solidity: function processMultiBid(uint128 amount, uint128 startingSlot, uint128 endingSlot, bool[6] slotSets, uint128 maxBid, uint128 minBid, bytes permit) returns()
func (_Auction *AuctionTransactorSession) ProcessMultiBid(amount *big.Int, startingSlot *big.Int, endingSlot *big.Int, slotSets [6]bool, maxBid *big.Int, minBid *big.Int, permit []byte) (*types.Transaction, error) {
	return _Auction.Contract.ProcessMultiBid(&_Auction.TransactOpts, amount, startingSlot, endingSlot, slotSets, maxBid, minBid, permit)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_Auction *AuctionTransactor) SetAllocationRatio(opts *bind.TransactOpts, newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setAllocationRatio", newAllocationRatio)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_Auction *AuctionSession) SetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetAllocationRatio(&_Auction.TransactOpts, newAllocationRatio)
}

// SetAllocationRatio is a paid mutator transaction binding the contract method 0x82787405.
//
// Solidity: function setAllocationRatio(uint16[3] newAllocationRatio) returns()
func (_Auction *AuctionTransactorSession) SetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetAllocationRatio(&_Auction.TransactOpts, newAllocationRatio)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_Auction *AuctionTransactor) SetBootCoordinator(opts *bind.TransactOpts, newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setBootCoordinator", newBootCoordinator, newBootCoordinatorURL)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_Auction *AuctionSession) SetBootCoordinator(newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.SetBootCoordinator(&_Auction.TransactOpts, newBootCoordinator, newBootCoordinatorURL)
}

// SetBootCoordinator is a paid mutator transaction binding the contract method 0x6cbdc3df.
//
// Solidity: function setBootCoordinator(address newBootCoordinator, string newBootCoordinatorURL) returns()
func (_Auction *AuctionTransactorSession) SetBootCoordinator(newBootCoordinator common.Address, newBootCoordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.SetBootCoordinator(&_Auction.TransactOpts, newBootCoordinator, newBootCoordinatorURL)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_Auction *AuctionTransactor) SetClosedAuctionSlots(opts *bind.TransactOpts, newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setClosedAuctionSlots", newClosedAuctionSlots)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_Auction *AuctionSession) SetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetClosedAuctionSlots(&_Auction.TransactOpts, newClosedAuctionSlots)
}

// SetClosedAuctionSlots is a paid mutator transaction binding the contract method 0xd92bdda3.
//
// Solidity: function setClosedAuctionSlots(uint16 newClosedAuctionSlots) returns()
func (_Auction *AuctionTransactorSession) SetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetClosedAuctionSlots(&_Auction.TransactOpts, newClosedAuctionSlots)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_Auction *AuctionTransactor) SetCoordinator(opts *bind.TransactOpts, forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setCoordinator", forger, coordinatorURL)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_Auction *AuctionSession) SetCoordinator(forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.SetCoordinator(&_Auction.TransactOpts, forger, coordinatorURL)
}

// SetCoordinator is a paid mutator transaction binding the contract method 0x0eeaf080.
//
// Solidity: function setCoordinator(address forger, string coordinatorURL) returns()
func (_Auction *AuctionTransactorSession) SetCoordinator(forger common.Address, coordinatorURL string) (*types.Transaction, error) {
	return _Auction.Contract.SetCoordinator(&_Auction.TransactOpts, forger, coordinatorURL)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_Auction *AuctionTransactor) SetDonationAddress(opts *bind.TransactOpts, newDonationAddress common.Address) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setDonationAddress", newDonationAddress)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_Auction *AuctionSession) SetDonationAddress(newDonationAddress common.Address) (*types.Transaction, error) {
	return _Auction.Contract.SetDonationAddress(&_Auction.TransactOpts, newDonationAddress)
}

// SetDonationAddress is a paid mutator transaction binding the contract method 0x6f48e79b.
//
// Solidity: function setDonationAddress(address newDonationAddress) returns()
func (_Auction *AuctionTransactorSession) SetDonationAddress(newDonationAddress common.Address) (*types.Transaction, error) {
	return _Auction.Contract.SetDonationAddress(&_Auction.TransactOpts, newDonationAddress)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_Auction *AuctionTransactor) SetOpenAuctionSlots(opts *bind.TransactOpts, newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setOpenAuctionSlots", newOpenAuctionSlots)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_Auction *AuctionSession) SetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetOpenAuctionSlots(&_Auction.TransactOpts, newOpenAuctionSlots)
}

// SetOpenAuctionSlots is a paid mutator transaction binding the contract method 0xc63de515.
//
// Solidity: function setOpenAuctionSlots(uint16 newOpenAuctionSlots) returns()
func (_Auction *AuctionTransactorSession) SetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetOpenAuctionSlots(&_Auction.TransactOpts, newOpenAuctionSlots)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_Auction *AuctionTransactor) SetOutbidding(opts *bind.TransactOpts, newOutbidding uint16) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setOutbidding", newOutbidding)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_Auction *AuctionSession) SetOutbidding(newOutbidding uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetOutbidding(&_Auction.TransactOpts, newOutbidding)
}

// SetOutbidding is a paid mutator transaction binding the contract method 0xdfd5281b.
//
// Solidity: function setOutbidding(uint16 newOutbidding) returns()
func (_Auction *AuctionTransactorSession) SetOutbidding(newOutbidding uint16) (*types.Transaction, error) {
	return _Auction.Contract.SetOutbidding(&_Auction.TransactOpts, newOutbidding)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_Auction *AuctionTransactor) SetSlotDeadline(opts *bind.TransactOpts, newDeadline uint8) (*types.Transaction, error) {
	return _Auction.contract.Transact(opts, "setSlotDeadline", newDeadline)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_Auction *AuctionSession) SetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return _Auction.Contract.SetSlotDeadline(&_Auction.TransactOpts, newDeadline)
}

// SetSlotDeadline is a paid mutator transaction binding the contract method 0x87e6b6bb.
//
// Solidity: function setSlotDeadline(uint8 newDeadline) returns()
func (_Auction *AuctionTransactorSession) SetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return _Auction.Contract.SetSlotDeadline(&_Auction.TransactOpts, newDeadline)
}

// AuctionHEZClaimedIterator is returned from FilterHEZClaimed and is used to iterate over the raw logs and unpacked data for HEZClaimed events raised by the Auction contract.
type AuctionHEZClaimedIterator struct {
	Event *AuctionHEZClaimed // Event containing the contract specifics and raw log

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
func (it *AuctionHEZClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionHEZClaimed)
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
		it.Event = new(AuctionHEZClaimed)
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
func (it *AuctionHEZClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionHEZClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionHEZClaimed represents a HEZClaimed event raised by the Auction contract.
type AuctionHEZClaimed struct {
	Owner  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterHEZClaimed is a free log retrieval operation binding the contract event 0x199ef0cb54d2b296ff6eaec2721bacf0ca3fd8344a43f5bdf4548b34dfa2594f.
//
// Solidity: event HEZClaimed(address indexed owner, uint128 amount)
func (_Auction *AuctionFilterer) FilterHEZClaimed(opts *bind.FilterOpts, owner []common.Address) (*AuctionHEZClaimedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "HEZClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return &AuctionHEZClaimedIterator{contract: _Auction.contract, event: "HEZClaimed", logs: logs, sub: sub}, nil
}

// WatchHEZClaimed is a free log subscription operation binding the contract event 0x199ef0cb54d2b296ff6eaec2721bacf0ca3fd8344a43f5bdf4548b34dfa2594f.
//
// Solidity: event HEZClaimed(address indexed owner, uint128 amount)
func (_Auction *AuctionFilterer) WatchHEZClaimed(opts *bind.WatchOpts, sink chan<- *AuctionHEZClaimed, owner []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "HEZClaimed", ownerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionHEZClaimed)
				if err := _Auction.contract.UnpackLog(event, "HEZClaimed", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseHEZClaimed(log types.Log) (*AuctionHEZClaimed, error) {
	event := new(AuctionHEZClaimed)
	if err := _Auction.contract.UnpackLog(event, "HEZClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionInitializeHermezAuctionProtocolEventIterator is returned from FilterInitializeHermezAuctionProtocolEvent and is used to iterate over the raw logs and unpacked data for InitializeHermezAuctionProtocolEvent events raised by the Auction contract.
type AuctionInitializeHermezAuctionProtocolEventIterator struct {
	Event *AuctionInitializeHermezAuctionProtocolEvent // Event containing the contract specifics and raw log

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
func (it *AuctionInitializeHermezAuctionProtocolEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionInitializeHermezAuctionProtocolEvent)
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
		it.Event = new(AuctionInitializeHermezAuctionProtocolEvent)
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
func (it *AuctionInitializeHermezAuctionProtocolEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionInitializeHermezAuctionProtocolEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionInitializeHermezAuctionProtocolEvent represents a InitializeHermezAuctionProtocolEvent event raised by the Auction contract.
type AuctionInitializeHermezAuctionProtocolEvent struct {
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
func (_Auction *AuctionFilterer) FilterInitializeHermezAuctionProtocolEvent(opts *bind.FilterOpts) (*AuctionInitializeHermezAuctionProtocolEventIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "InitializeHermezAuctionProtocolEvent")
	if err != nil {
		return nil, err
	}
	return &AuctionInitializeHermezAuctionProtocolEventIterator{contract: _Auction.contract, event: "InitializeHermezAuctionProtocolEvent", logs: logs, sub: sub}, nil
}

// WatchInitializeHermezAuctionProtocolEvent is a free log subscription operation binding the contract event 0x9717e4e04c13817c600463a7a450110c754fd78758cdd538603f30528a24ce4b.
//
// Solidity: event InitializeHermezAuctionProtocolEvent(address donationAddress, address bootCoordinatorAddress, string bootCoordinatorURL, uint16 outbidding, uint8 slotDeadline, uint16 closedAuctionSlots, uint16 openAuctionSlots, uint16[3] allocationRatio)
func (_Auction *AuctionFilterer) WatchInitializeHermezAuctionProtocolEvent(opts *bind.WatchOpts, sink chan<- *AuctionInitializeHermezAuctionProtocolEvent) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "InitializeHermezAuctionProtocolEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionInitializeHermezAuctionProtocolEvent)
				if err := _Auction.contract.UnpackLog(event, "InitializeHermezAuctionProtocolEvent", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseInitializeHermezAuctionProtocolEvent(log types.Log) (*AuctionInitializeHermezAuctionProtocolEvent, error) {
	event := new(AuctionInitializeHermezAuctionProtocolEvent)
	if err := _Auction.contract.UnpackLog(event, "InitializeHermezAuctionProtocolEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewAllocationRatioIterator is returned from FilterNewAllocationRatio and is used to iterate over the raw logs and unpacked data for NewAllocationRatio events raised by the Auction contract.
type AuctionNewAllocationRatioIterator struct {
	Event *AuctionNewAllocationRatio // Event containing the contract specifics and raw log

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
func (it *AuctionNewAllocationRatioIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewAllocationRatio)
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
		it.Event = new(AuctionNewAllocationRatio)
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
func (it *AuctionNewAllocationRatioIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewAllocationRatioIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewAllocationRatio represents a NewAllocationRatio event raised by the Auction contract.
type AuctionNewAllocationRatio struct {
	NewAllocationRatio [3]uint16
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNewAllocationRatio is a free log retrieval operation binding the contract event 0x0bb59eceb12f1bdb63e4a7d57c70d6473fefd7c3f51af5a3604f7e97197073e4.
//
// Solidity: event NewAllocationRatio(uint16[3] newAllocationRatio)
func (_Auction *AuctionFilterer) FilterNewAllocationRatio(opts *bind.FilterOpts) (*AuctionNewAllocationRatioIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewAllocationRatio")
	if err != nil {
		return nil, err
	}
	return &AuctionNewAllocationRatioIterator{contract: _Auction.contract, event: "NewAllocationRatio", logs: logs, sub: sub}, nil
}

// WatchNewAllocationRatio is a free log subscription operation binding the contract event 0x0bb59eceb12f1bdb63e4a7d57c70d6473fefd7c3f51af5a3604f7e97197073e4.
//
// Solidity: event NewAllocationRatio(uint16[3] newAllocationRatio)
func (_Auction *AuctionFilterer) WatchNewAllocationRatio(opts *bind.WatchOpts, sink chan<- *AuctionNewAllocationRatio) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewAllocationRatio")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewAllocationRatio)
				if err := _Auction.contract.UnpackLog(event, "NewAllocationRatio", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewAllocationRatio(log types.Log) (*AuctionNewAllocationRatio, error) {
	event := new(AuctionNewAllocationRatio)
	if err := _Auction.contract.UnpackLog(event, "NewAllocationRatio", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewBidIterator is returned from FilterNewBid and is used to iterate over the raw logs and unpacked data for NewBid events raised by the Auction contract.
type AuctionNewBidIterator struct {
	Event *AuctionNewBid // Event containing the contract specifics and raw log

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
func (it *AuctionNewBidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewBid)
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
		it.Event = new(AuctionNewBid)
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
func (it *AuctionNewBidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewBidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewBid represents a NewBid event raised by the Auction contract.
type AuctionNewBid struct {
	Slot      *big.Int
	BidAmount *big.Int
	Bidder    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterNewBid is a free log retrieval operation binding the contract event 0xd48e8329cdb2fb109b4fe445d7b681a74b256bff16e6f7f33b9d4fbe9038e433.
//
// Solidity: event NewBid(uint128 indexed slot, uint128 bidAmount, address indexed bidder)
func (_Auction *AuctionFilterer) FilterNewBid(opts *bind.FilterOpts, slot []*big.Int, bidder []common.Address) (*AuctionNewBidIterator, error) {

	var slotRule []interface{}
	for _, slotItem := range slot {
		slotRule = append(slotRule, slotItem)
	}

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewBid", slotRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return &AuctionNewBidIterator{contract: _Auction.contract, event: "NewBid", logs: logs, sub: sub}, nil
}

// WatchNewBid is a free log subscription operation binding the contract event 0xd48e8329cdb2fb109b4fe445d7b681a74b256bff16e6f7f33b9d4fbe9038e433.
//
// Solidity: event NewBid(uint128 indexed slot, uint128 bidAmount, address indexed bidder)
func (_Auction *AuctionFilterer) WatchNewBid(opts *bind.WatchOpts, sink chan<- *AuctionNewBid, slot []*big.Int, bidder []common.Address) (event.Subscription, error) {

	var slotRule []interface{}
	for _, slotItem := range slot {
		slotRule = append(slotRule, slotItem)
	}

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewBid", slotRule, bidderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewBid)
				if err := _Auction.contract.UnpackLog(event, "NewBid", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewBid(log types.Log) (*AuctionNewBid, error) {
	event := new(AuctionNewBid)
	if err := _Auction.contract.UnpackLog(event, "NewBid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewBootCoordinatorIterator is returned from FilterNewBootCoordinator and is used to iterate over the raw logs and unpacked data for NewBootCoordinator events raised by the Auction contract.
type AuctionNewBootCoordinatorIterator struct {
	Event *AuctionNewBootCoordinator // Event containing the contract specifics and raw log

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
func (it *AuctionNewBootCoordinatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewBootCoordinator)
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
		it.Event = new(AuctionNewBootCoordinator)
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
func (it *AuctionNewBootCoordinatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewBootCoordinatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewBootCoordinator represents a NewBootCoordinator event raised by the Auction contract.
type AuctionNewBootCoordinator struct {
	NewBootCoordinator    common.Address
	NewBootCoordinatorURL string
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewBootCoordinator is a free log retrieval operation binding the contract event 0x0487eab4c1da34bf653268e33bee8bfec7dacfd6f3226047197ebf872293cfd6.
//
// Solidity: event NewBootCoordinator(address indexed newBootCoordinator, string newBootCoordinatorURL)
func (_Auction *AuctionFilterer) FilterNewBootCoordinator(opts *bind.FilterOpts, newBootCoordinator []common.Address) (*AuctionNewBootCoordinatorIterator, error) {

	var newBootCoordinatorRule []interface{}
	for _, newBootCoordinatorItem := range newBootCoordinator {
		newBootCoordinatorRule = append(newBootCoordinatorRule, newBootCoordinatorItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewBootCoordinator", newBootCoordinatorRule)
	if err != nil {
		return nil, err
	}
	return &AuctionNewBootCoordinatorIterator{contract: _Auction.contract, event: "NewBootCoordinator", logs: logs, sub: sub}, nil
}

// WatchNewBootCoordinator is a free log subscription operation binding the contract event 0x0487eab4c1da34bf653268e33bee8bfec7dacfd6f3226047197ebf872293cfd6.
//
// Solidity: event NewBootCoordinator(address indexed newBootCoordinator, string newBootCoordinatorURL)
func (_Auction *AuctionFilterer) WatchNewBootCoordinator(opts *bind.WatchOpts, sink chan<- *AuctionNewBootCoordinator, newBootCoordinator []common.Address) (event.Subscription, error) {

	var newBootCoordinatorRule []interface{}
	for _, newBootCoordinatorItem := range newBootCoordinator {
		newBootCoordinatorRule = append(newBootCoordinatorRule, newBootCoordinatorItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewBootCoordinator", newBootCoordinatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewBootCoordinator)
				if err := _Auction.contract.UnpackLog(event, "NewBootCoordinator", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewBootCoordinator(log types.Log) (*AuctionNewBootCoordinator, error) {
	event := new(AuctionNewBootCoordinator)
	if err := _Auction.contract.UnpackLog(event, "NewBootCoordinator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewClosedAuctionSlotsIterator is returned from FilterNewClosedAuctionSlots and is used to iterate over the raw logs and unpacked data for NewClosedAuctionSlots events raised by the Auction contract.
type AuctionNewClosedAuctionSlotsIterator struct {
	Event *AuctionNewClosedAuctionSlots // Event containing the contract specifics and raw log

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
func (it *AuctionNewClosedAuctionSlotsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewClosedAuctionSlots)
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
		it.Event = new(AuctionNewClosedAuctionSlots)
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
func (it *AuctionNewClosedAuctionSlotsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewClosedAuctionSlotsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewClosedAuctionSlots represents a NewClosedAuctionSlots event raised by the Auction contract.
type AuctionNewClosedAuctionSlots struct {
	NewClosedAuctionSlots uint16
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterNewClosedAuctionSlots is a free log retrieval operation binding the contract event 0xc78051d3757db196b1e445f3a9a1380944518c69b5d7922ec747c54f0340a4ea.
//
// Solidity: event NewClosedAuctionSlots(uint16 newClosedAuctionSlots)
func (_Auction *AuctionFilterer) FilterNewClosedAuctionSlots(opts *bind.FilterOpts) (*AuctionNewClosedAuctionSlotsIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewClosedAuctionSlots")
	if err != nil {
		return nil, err
	}
	return &AuctionNewClosedAuctionSlotsIterator{contract: _Auction.contract, event: "NewClosedAuctionSlots", logs: logs, sub: sub}, nil
}

// WatchNewClosedAuctionSlots is a free log subscription operation binding the contract event 0xc78051d3757db196b1e445f3a9a1380944518c69b5d7922ec747c54f0340a4ea.
//
// Solidity: event NewClosedAuctionSlots(uint16 newClosedAuctionSlots)
func (_Auction *AuctionFilterer) WatchNewClosedAuctionSlots(opts *bind.WatchOpts, sink chan<- *AuctionNewClosedAuctionSlots) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewClosedAuctionSlots")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewClosedAuctionSlots)
				if err := _Auction.contract.UnpackLog(event, "NewClosedAuctionSlots", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewClosedAuctionSlots(log types.Log) (*AuctionNewClosedAuctionSlots, error) {
	event := new(AuctionNewClosedAuctionSlots)
	if err := _Auction.contract.UnpackLog(event, "NewClosedAuctionSlots", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewDefaultSlotSetBidIterator is returned from FilterNewDefaultSlotSetBid and is used to iterate over the raw logs and unpacked data for NewDefaultSlotSetBid events raised by the Auction contract.
type AuctionNewDefaultSlotSetBidIterator struct {
	Event *AuctionNewDefaultSlotSetBid // Event containing the contract specifics and raw log

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
func (it *AuctionNewDefaultSlotSetBidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewDefaultSlotSetBid)
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
		it.Event = new(AuctionNewDefaultSlotSetBid)
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
func (it *AuctionNewDefaultSlotSetBidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewDefaultSlotSetBidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewDefaultSlotSetBid represents a NewDefaultSlotSetBid event raised by the Auction contract.
type AuctionNewDefaultSlotSetBid struct {
	SlotSet          *big.Int
	NewInitialMinBid *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNewDefaultSlotSetBid is a free log retrieval operation binding the contract event 0xa922aa010d1ff8e70b2aa9247d891836795c3d3ba2a543c37c91a44dc4a50172.
//
// Solidity: event NewDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid)
func (_Auction *AuctionFilterer) FilterNewDefaultSlotSetBid(opts *bind.FilterOpts) (*AuctionNewDefaultSlotSetBidIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewDefaultSlotSetBid")
	if err != nil {
		return nil, err
	}
	return &AuctionNewDefaultSlotSetBidIterator{contract: _Auction.contract, event: "NewDefaultSlotSetBid", logs: logs, sub: sub}, nil
}

// WatchNewDefaultSlotSetBid is a free log subscription operation binding the contract event 0xa922aa010d1ff8e70b2aa9247d891836795c3d3ba2a543c37c91a44dc4a50172.
//
// Solidity: event NewDefaultSlotSetBid(uint128 slotSet, uint128 newInitialMinBid)
func (_Auction *AuctionFilterer) WatchNewDefaultSlotSetBid(opts *bind.WatchOpts, sink chan<- *AuctionNewDefaultSlotSetBid) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewDefaultSlotSetBid")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewDefaultSlotSetBid)
				if err := _Auction.contract.UnpackLog(event, "NewDefaultSlotSetBid", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewDefaultSlotSetBid(log types.Log) (*AuctionNewDefaultSlotSetBid, error) {
	event := new(AuctionNewDefaultSlotSetBid)
	if err := _Auction.contract.UnpackLog(event, "NewDefaultSlotSetBid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewDonationAddressIterator is returned from FilterNewDonationAddress and is used to iterate over the raw logs and unpacked data for NewDonationAddress events raised by the Auction contract.
type AuctionNewDonationAddressIterator struct {
	Event *AuctionNewDonationAddress // Event containing the contract specifics and raw log

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
func (it *AuctionNewDonationAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewDonationAddress)
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
		it.Event = new(AuctionNewDonationAddress)
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
func (it *AuctionNewDonationAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewDonationAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewDonationAddress represents a NewDonationAddress event raised by the Auction contract.
type AuctionNewDonationAddress struct {
	NewDonationAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterNewDonationAddress is a free log retrieval operation binding the contract event 0xa62863cbad1647a2855e9cd39d04fa6dfd32e1b9cfaff1aaf6523f4aaafeccd7.
//
// Solidity: event NewDonationAddress(address indexed newDonationAddress)
func (_Auction *AuctionFilterer) FilterNewDonationAddress(opts *bind.FilterOpts, newDonationAddress []common.Address) (*AuctionNewDonationAddressIterator, error) {

	var newDonationAddressRule []interface{}
	for _, newDonationAddressItem := range newDonationAddress {
		newDonationAddressRule = append(newDonationAddressRule, newDonationAddressItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewDonationAddress", newDonationAddressRule)
	if err != nil {
		return nil, err
	}
	return &AuctionNewDonationAddressIterator{contract: _Auction.contract, event: "NewDonationAddress", logs: logs, sub: sub}, nil
}

// WatchNewDonationAddress is a free log subscription operation binding the contract event 0xa62863cbad1647a2855e9cd39d04fa6dfd32e1b9cfaff1aaf6523f4aaafeccd7.
//
// Solidity: event NewDonationAddress(address indexed newDonationAddress)
func (_Auction *AuctionFilterer) WatchNewDonationAddress(opts *bind.WatchOpts, sink chan<- *AuctionNewDonationAddress, newDonationAddress []common.Address) (event.Subscription, error) {

	var newDonationAddressRule []interface{}
	for _, newDonationAddressItem := range newDonationAddress {
		newDonationAddressRule = append(newDonationAddressRule, newDonationAddressItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewDonationAddress", newDonationAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewDonationAddress)
				if err := _Auction.contract.UnpackLog(event, "NewDonationAddress", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewDonationAddress(log types.Log) (*AuctionNewDonationAddress, error) {
	event := new(AuctionNewDonationAddress)
	if err := _Auction.contract.UnpackLog(event, "NewDonationAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewForgeIterator is returned from FilterNewForge and is used to iterate over the raw logs and unpacked data for NewForge events raised by the Auction contract.
type AuctionNewForgeIterator struct {
	Event *AuctionNewForge // Event containing the contract specifics and raw log

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
func (it *AuctionNewForgeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewForge)
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
		it.Event = new(AuctionNewForge)
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
func (it *AuctionNewForgeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewForgeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewForge represents a NewForge event raised by the Auction contract.
type AuctionNewForge struct {
	Forger      common.Address
	SlotToForge *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterNewForge is a free log retrieval operation binding the contract event 0x7cae662d4cfa9d9c5575c65f0cc41a858c51ca14ebcbd02a802a62376c3ad238.
//
// Solidity: event NewForge(address indexed forger, uint128 indexed slotToForge)
func (_Auction *AuctionFilterer) FilterNewForge(opts *bind.FilterOpts, forger []common.Address, slotToForge []*big.Int) (*AuctionNewForgeIterator, error) {

	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewForge", forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return &AuctionNewForgeIterator{contract: _Auction.contract, event: "NewForge", logs: logs, sub: sub}, nil
}

// WatchNewForge is a free log subscription operation binding the contract event 0x7cae662d4cfa9d9c5575c65f0cc41a858c51ca14ebcbd02a802a62376c3ad238.
//
// Solidity: event NewForge(address indexed forger, uint128 indexed slotToForge)
func (_Auction *AuctionFilterer) WatchNewForge(opts *bind.WatchOpts, sink chan<- *AuctionNewForge, forger []common.Address, slotToForge []*big.Int) (event.Subscription, error) {

	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}
	var slotToForgeRule []interface{}
	for _, slotToForgeItem := range slotToForge {
		slotToForgeRule = append(slotToForgeRule, slotToForgeItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewForge", forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewForge)
				if err := _Auction.contract.UnpackLog(event, "NewForge", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewForge(log types.Log) (*AuctionNewForge, error) {
	event := new(AuctionNewForge)
	if err := _Auction.contract.UnpackLog(event, "NewForge", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewForgeAllocatedIterator is returned from FilterNewForgeAllocated and is used to iterate over the raw logs and unpacked data for NewForgeAllocated events raised by the Auction contract.
type AuctionNewForgeAllocatedIterator struct {
	Event *AuctionNewForgeAllocated // Event containing the contract specifics and raw log

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
func (it *AuctionNewForgeAllocatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewForgeAllocated)
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
		it.Event = new(AuctionNewForgeAllocated)
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
func (it *AuctionNewForgeAllocatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewForgeAllocatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewForgeAllocated represents a NewForgeAllocated event raised by the Auction contract.
type AuctionNewForgeAllocated struct {
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
func (_Auction *AuctionFilterer) FilterNewForgeAllocated(opts *bind.FilterOpts, bidder []common.Address, forger []common.Address, slotToForge []*big.Int) (*AuctionNewForgeAllocatedIterator, error) {

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

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewForgeAllocated", bidderRule, forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return &AuctionNewForgeAllocatedIterator{contract: _Auction.contract, event: "NewForgeAllocated", logs: logs, sub: sub}, nil
}

// WatchNewForgeAllocated is a free log subscription operation binding the contract event 0xd64ebb43f4c2b91022b97389834432f1027ef55586129ba05a3a3065b2304f05.
//
// Solidity: event NewForgeAllocated(address indexed bidder, address indexed forger, uint128 indexed slotToForge, uint128 burnAmount, uint128 donationAmount, uint128 governanceAmount)
func (_Auction *AuctionFilterer) WatchNewForgeAllocated(opts *bind.WatchOpts, sink chan<- *AuctionNewForgeAllocated, bidder []common.Address, forger []common.Address, slotToForge []*big.Int) (event.Subscription, error) {

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

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewForgeAllocated", bidderRule, forgerRule, slotToForgeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewForgeAllocated)
				if err := _Auction.contract.UnpackLog(event, "NewForgeAllocated", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewForgeAllocated(log types.Log) (*AuctionNewForgeAllocated, error) {
	event := new(AuctionNewForgeAllocated)
	if err := _Auction.contract.UnpackLog(event, "NewForgeAllocated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewOpenAuctionSlotsIterator is returned from FilterNewOpenAuctionSlots and is used to iterate over the raw logs and unpacked data for NewOpenAuctionSlots events raised by the Auction contract.
type AuctionNewOpenAuctionSlotsIterator struct {
	Event *AuctionNewOpenAuctionSlots // Event containing the contract specifics and raw log

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
func (it *AuctionNewOpenAuctionSlotsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewOpenAuctionSlots)
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
		it.Event = new(AuctionNewOpenAuctionSlots)
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
func (it *AuctionNewOpenAuctionSlotsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewOpenAuctionSlotsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewOpenAuctionSlots represents a NewOpenAuctionSlots event raised by the Auction contract.
type AuctionNewOpenAuctionSlots struct {
	NewOpenAuctionSlots uint16
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterNewOpenAuctionSlots is a free log retrieval operation binding the contract event 0x3da0492dea7298351bc14d1c0699905fd0657c33487449751af50fc0c8b593f1.
//
// Solidity: event NewOpenAuctionSlots(uint16 newOpenAuctionSlots)
func (_Auction *AuctionFilterer) FilterNewOpenAuctionSlots(opts *bind.FilterOpts) (*AuctionNewOpenAuctionSlotsIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewOpenAuctionSlots")
	if err != nil {
		return nil, err
	}
	return &AuctionNewOpenAuctionSlotsIterator{contract: _Auction.contract, event: "NewOpenAuctionSlots", logs: logs, sub: sub}, nil
}

// WatchNewOpenAuctionSlots is a free log subscription operation binding the contract event 0x3da0492dea7298351bc14d1c0699905fd0657c33487449751af50fc0c8b593f1.
//
// Solidity: event NewOpenAuctionSlots(uint16 newOpenAuctionSlots)
func (_Auction *AuctionFilterer) WatchNewOpenAuctionSlots(opts *bind.WatchOpts, sink chan<- *AuctionNewOpenAuctionSlots) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewOpenAuctionSlots")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewOpenAuctionSlots)
				if err := _Auction.contract.UnpackLog(event, "NewOpenAuctionSlots", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewOpenAuctionSlots(log types.Log) (*AuctionNewOpenAuctionSlots, error) {
	event := new(AuctionNewOpenAuctionSlots)
	if err := _Auction.contract.UnpackLog(event, "NewOpenAuctionSlots", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewOutbiddingIterator is returned from FilterNewOutbidding and is used to iterate over the raw logs and unpacked data for NewOutbidding events raised by the Auction contract.
type AuctionNewOutbiddingIterator struct {
	Event *AuctionNewOutbidding // Event containing the contract specifics and raw log

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
func (it *AuctionNewOutbiddingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewOutbidding)
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
		it.Event = new(AuctionNewOutbidding)
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
func (it *AuctionNewOutbiddingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewOutbiddingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewOutbidding represents a NewOutbidding event raised by the Auction contract.
type AuctionNewOutbidding struct {
	NewOutbidding uint16
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterNewOutbidding is a free log retrieval operation binding the contract event 0xd3748b8c326e93d12af934fbf87471e315a89bc3f7b8222343acf0210edf248e.
//
// Solidity: event NewOutbidding(uint16 newOutbidding)
func (_Auction *AuctionFilterer) FilterNewOutbidding(opts *bind.FilterOpts) (*AuctionNewOutbiddingIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewOutbidding")
	if err != nil {
		return nil, err
	}
	return &AuctionNewOutbiddingIterator{contract: _Auction.contract, event: "NewOutbidding", logs: logs, sub: sub}, nil
}

// WatchNewOutbidding is a free log subscription operation binding the contract event 0xd3748b8c326e93d12af934fbf87471e315a89bc3f7b8222343acf0210edf248e.
//
// Solidity: event NewOutbidding(uint16 newOutbidding)
func (_Auction *AuctionFilterer) WatchNewOutbidding(opts *bind.WatchOpts, sink chan<- *AuctionNewOutbidding) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewOutbidding")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewOutbidding)
				if err := _Auction.contract.UnpackLog(event, "NewOutbidding", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewOutbidding(log types.Log) (*AuctionNewOutbidding, error) {
	event := new(AuctionNewOutbidding)
	if err := _Auction.contract.UnpackLog(event, "NewOutbidding", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionNewSlotDeadlineIterator is returned from FilterNewSlotDeadline and is used to iterate over the raw logs and unpacked data for NewSlotDeadline events raised by the Auction contract.
type AuctionNewSlotDeadlineIterator struct {
	Event *AuctionNewSlotDeadline // Event containing the contract specifics and raw log

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
func (it *AuctionNewSlotDeadlineIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionNewSlotDeadline)
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
		it.Event = new(AuctionNewSlotDeadline)
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
func (it *AuctionNewSlotDeadlineIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionNewSlotDeadlineIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionNewSlotDeadline represents a NewSlotDeadline event raised by the Auction contract.
type AuctionNewSlotDeadline struct {
	NewSlotDeadline uint8
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterNewSlotDeadline is a free log retrieval operation binding the contract event 0x4a0d90b611c15e02dbf23b10f35b936cf2c77665f8c77822d3eca131f9d986d3.
//
// Solidity: event NewSlotDeadline(uint8 newSlotDeadline)
func (_Auction *AuctionFilterer) FilterNewSlotDeadline(opts *bind.FilterOpts) (*AuctionNewSlotDeadlineIterator, error) {

	logs, sub, err := _Auction.contract.FilterLogs(opts, "NewSlotDeadline")
	if err != nil {
		return nil, err
	}
	return &AuctionNewSlotDeadlineIterator{contract: _Auction.contract, event: "NewSlotDeadline", logs: logs, sub: sub}, nil
}

// WatchNewSlotDeadline is a free log subscription operation binding the contract event 0x4a0d90b611c15e02dbf23b10f35b936cf2c77665f8c77822d3eca131f9d986d3.
//
// Solidity: event NewSlotDeadline(uint8 newSlotDeadline)
func (_Auction *AuctionFilterer) WatchNewSlotDeadline(opts *bind.WatchOpts, sink chan<- *AuctionNewSlotDeadline) (event.Subscription, error) {

	logs, sub, err := _Auction.contract.WatchLogs(opts, "NewSlotDeadline")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionNewSlotDeadline)
				if err := _Auction.contract.UnpackLog(event, "NewSlotDeadline", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseNewSlotDeadline(log types.Log) (*AuctionNewSlotDeadline, error) {
	event := new(AuctionNewSlotDeadline)
	if err := _Auction.contract.UnpackLog(event, "NewSlotDeadline", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AuctionSetCoordinatorIterator is returned from FilterSetCoordinator and is used to iterate over the raw logs and unpacked data for SetCoordinator events raised by the Auction contract.
type AuctionSetCoordinatorIterator struct {
	Event *AuctionSetCoordinator // Event containing the contract specifics and raw log

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
func (it *AuctionSetCoordinatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AuctionSetCoordinator)
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
		it.Event = new(AuctionSetCoordinator)
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
func (it *AuctionSetCoordinatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AuctionSetCoordinatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AuctionSetCoordinator represents a SetCoordinator event raised by the Auction contract.
type AuctionSetCoordinator struct {
	Bidder         common.Address
	Forger         common.Address
	CoordinatorURL string
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterSetCoordinator is a free log retrieval operation binding the contract event 0x5246b2ac9ee77efe2e64af6df00055d97e2d6e1b277f5a8d17ba5bca1a573da0.
//
// Solidity: event SetCoordinator(address indexed bidder, address indexed forger, string coordinatorURL)
func (_Auction *AuctionFilterer) FilterSetCoordinator(opts *bind.FilterOpts, bidder []common.Address, forger []common.Address) (*AuctionSetCoordinatorIterator, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}

	logs, sub, err := _Auction.contract.FilterLogs(opts, "SetCoordinator", bidderRule, forgerRule)
	if err != nil {
		return nil, err
	}
	return &AuctionSetCoordinatorIterator{contract: _Auction.contract, event: "SetCoordinator", logs: logs, sub: sub}, nil
}

// WatchSetCoordinator is a free log subscription operation binding the contract event 0x5246b2ac9ee77efe2e64af6df00055d97e2d6e1b277f5a8d17ba5bca1a573da0.
//
// Solidity: event SetCoordinator(address indexed bidder, address indexed forger, string coordinatorURL)
func (_Auction *AuctionFilterer) WatchSetCoordinator(opts *bind.WatchOpts, sink chan<- *AuctionSetCoordinator, bidder []common.Address, forger []common.Address) (event.Subscription, error) {

	var bidderRule []interface{}
	for _, bidderItem := range bidder {
		bidderRule = append(bidderRule, bidderItem)
	}
	var forgerRule []interface{}
	for _, forgerItem := range forger {
		forgerRule = append(forgerRule, forgerItem)
	}

	logs, sub, err := _Auction.contract.WatchLogs(opts, "SetCoordinator", bidderRule, forgerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AuctionSetCoordinator)
				if err := _Auction.contract.UnpackLog(event, "SetCoordinator", log); err != nil {
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
func (_Auction *AuctionFilterer) ParseSetCoordinator(log types.Log) (*AuctionSetCoordinator, error) {
	event := new(AuctionSetCoordinator)
	if err := _Auction.contract.UnpackLog(event, "SetCoordinator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
