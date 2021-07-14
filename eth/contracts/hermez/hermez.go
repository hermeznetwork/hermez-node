// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package hermez

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

// HermezABI is the input ABI used to generate the binding from.
const HermezABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tokenID\",\"type\":\"uint32\"}],\"name\":\"AddToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"batchNum\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"l1UserTxsLen\",\"type\":\"uint16\"}],\"name\":\"ForgeBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"forgeL1L2BatchTimeout\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAddToken\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"InitializeHermezEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"queueIndex\",\"type\":\"uint32\"},{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"position\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"l1UserTx\",\"type\":\"bytes\"}],\"name\":\"L1UserTxEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"SafeMode\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint8\",\"name\":\"numBucket\",\"type\":\"uint8\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"blockStamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"withdrawals\",\"type\":\"uint256\"}],\"name\":\"UpdateBucketWithdraw\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"arrayBuckets\",\"type\":\"uint256[]\"}],\"name\":\"UpdateBucketsParameters\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newFeeAddToken\",\"type\":\"uint256\"}],\"name\":\"UpdateFeeAddToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"newForgeL1L2BatchTimeout\",\"type\":\"uint8\"}],\"name\":\"UpdateForgeL1L2BatchTimeout\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"addressArray\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"uint64[]\",\"name\":\"valueArray\",\"type\":\"uint64[]\"}],\"name\":\"UpdateTokenExchange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"UpdateWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint48\",\"name\":\"idx\",\"type\":\"uint48\"},{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"numExitRoot\",\"type\":\"uint32\"},{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"instantWithdraw\",\"type\":\"bool\"}],\"name\":\"WithdrawEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"hermezV2\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ABSOLUTE_MAX_L1L2BATCHTIMEOUT\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ACCOUNT_CREATION_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"AUTHORISE_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"domainSeparator\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EIP712DOMAIN_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"HERMEZ_NETWORK_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NAME_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"babyPubKey\",\"type\":\"uint256\"},{\"internalType\":\"uint48\",\"name\":\"fromIdx\",\"type\":\"uint48\"},{\"internalType\":\"uint40\",\"name\":\"loadAmountF\",\"type\":\"uint40\"},{\"internalType\":\"uint40\",\"name\":\"amountF\",\"type\":\"uint40\"},{\"internalType\":\"uint32\",\"name\":\"tokenID\",\"type\":\"uint32\"},{\"internalType\":\"uint48\",\"name\":\"toIdx\",\"type\":\"uint48\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"addL1Transaction\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"permit\",\"type\":\"bytes\"}],\"name\":\"addToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"}],\"name\":\"buckets\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"},{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"name\":\"exitNullifierMap\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"exitRootsMap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feeAddToken\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint48\",\"name\":\"newLastIdx\",\"type\":\"uint48\"},{\"internalType\":\"uint256\",\"name\":\"newStRoot\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"newExitRoot\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"encodedL1CoordinatorTx\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"l1L2TxsData\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"feeIdxCoordinator\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"verifierIdx\",\"type\":\"uint8\"},{\"internalType\":\"bool\",\"name\":\"l1Batch\",\"type\":\"bool\"},{\"internalType\":\"uint256[2]\",\"name\":\"proofA\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2][2]\",\"name\":\"proofB\",\"type\":\"uint256[2][2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"proofC\",\"type\":\"uint256[2]\"}],\"name\":\"forgeBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"forgeL1L2BatchTimeout\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezAuctionContract\",\"outputs\":[{\"internalType\":\"contractIHermezAuctionProtocol\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezGovernanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_verifiers\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"_verifiersParams\",\"type\":\"uint256[]\"},{\"internalType\":\"address\",\"name\":\"_withdrawVerifier\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_hermezAuctionContract\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_tokenHEZ\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"_forgeL1L2BatchTimeout\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"_feeAddToken\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_poseidon2Elements\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_poseidon3Elements\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_poseidon4Elements\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_hermezGovernanceAddress\",\"type\":\"address\"},{\"internalType\":\"uint64\",\"name\":\"_withdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_withdrawDelayerContract\",\"type\":\"address\"}],\"name\":\"initializeHermez\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"tokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"instantWithdrawalViewer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"l1L2TxsDataHashMap\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastForgedBatch\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastIdx\",\"outputs\":[{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastL1L2Batch\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"mapL1TxQueue\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nBuckets\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextL1FillingQueue\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextL1ToForgeQueue\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"ceilUSD\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blockStamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rateBlocks\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rateWithdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxWithdrawals\",\"type\":\"uint256\"}],\"name\":\"packBucket\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"ret\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registerTokensCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"rollupVerifiers\",\"outputs\":[{\"internalType\":\"contractVerifierRollupInterface\",\"name\":\"verifierInterface\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"maxTx\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nLevels\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"rollupVerifiersLength\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"safeMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"stateRootMap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenExchange\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"tokenHEZ\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenList\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"tokenMap\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bucket\",\"type\":\"uint256\"}],\"name\":\"unpackBucket\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"ceilUSD\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"blockStamp\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"withdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rateBlocks\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rateWithdrawals\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxWithdrawals\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"newBuckets\",\"type\":\"uint256[]\"}],\"name\":\"updateBucketsParameters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newFeeAddToken\",\"type\":\"uint256\"}],\"name\":\"updateFeeAddToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"newForgeL1L2BatchTimeout\",\"type\":\"uint8\"}],\"name\":\"updateForgeL1L2BatchTimeout\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addressArray\",\"type\":\"address[]\"},{\"internalType\":\"uint64[]\",\"name\":\"valueArray\",\"type\":\"uint64[]\"}],\"name\":\"updateTokenExchange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"updateVerifiers\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"updateWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[2]\",\"name\":\"proofA\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2][2]\",\"name\":\"proofB\",\"type\":\"uint256[2][2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"proofC\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint32\",\"name\":\"tokenID\",\"type\":\"uint32\"},{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint32\",\"name\":\"numExitRoot\",\"type\":\"uint32\"},{\"internalType\":\"uint48\",\"name\":\"idx\",\"type\":\"uint48\"},{\"internalType\":\"bool\",\"name\":\"instantWithdraw\",\"type\":\"bool\"}],\"name\":\"withdrawCircuit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawDelayerContract\",\"outputs\":[{\"internalType\":\"contractIWithdrawalDelayer\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"tokenID\",\"type\":\"uint32\"},{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint256\",\"name\":\"babyPubKey\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"numExitRoot\",\"type\":\"uint32\"},{\"internalType\":\"uint256[]\",\"name\":\"siblings\",\"type\":\"uint256[]\"},{\"internalType\":\"uint48\",\"name\":\"idx\",\"type\":\"uint48\"},{\"internalType\":\"bool\",\"name\":\"instantWithdraw\",\"type\":\"bool\"}],\"name\":\"withdrawMerkleProof\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawVerifier\",\"outputs\":[{\"internalType\":\"contractVerifierWithdrawInterface\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdrawalDelay\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// Hermez is an auto generated Go binding around an Ethereum contract.
type Hermez struct {
	HermezCaller     // Read-only binding to the contract
	HermezTransactor // Write-only binding to the contract
	HermezFilterer   // Log filterer for contract events
}

// HermezCaller is an auto generated read-only Go binding around an Ethereum contract.
type HermezCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezTransactor is an auto generated write-only Go binding around an Ethereum contract.
type HermezTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type HermezFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HermezSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type HermezSession struct {
	Contract     *Hermez           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HermezCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type HermezCallerSession struct {
	Contract *HermezCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// HermezTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type HermezTransactorSession struct {
	Contract     *HermezTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HermezRaw is an auto generated low-level Go binding around an Ethereum contract.
type HermezRaw struct {
	Contract *Hermez // Generic contract binding to access the raw methods on
}

// HermezCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type HermezCallerRaw struct {
	Contract *HermezCaller // Generic read-only contract binding to access the raw methods on
}

// HermezTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type HermezTransactorRaw struct {
	Contract *HermezTransactor // Generic write-only contract binding to access the raw methods on
}

// NewHermez creates a new instance of Hermez, bound to a specific deployed contract.
func NewHermez(address common.Address, backend bind.ContractBackend) (*Hermez, error) {
	contract, err := bindHermez(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Hermez{HermezCaller: HermezCaller{contract: contract}, HermezTransactor: HermezTransactor{contract: contract}, HermezFilterer: HermezFilterer{contract: contract}}, nil
}

// NewHermezCaller creates a new read-only instance of Hermez, bound to a specific deployed contract.
func NewHermezCaller(address common.Address, caller bind.ContractCaller) (*HermezCaller, error) {
	contract, err := bindHermez(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HermezCaller{contract: contract}, nil
}

// NewHermezTransactor creates a new write-only instance of Hermez, bound to a specific deployed contract.
func NewHermezTransactor(address common.Address, transactor bind.ContractTransactor) (*HermezTransactor, error) {
	contract, err := bindHermez(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HermezTransactor{contract: contract}, nil
}

// NewHermezFilterer creates a new log filterer instance of Hermez, bound to a specific deployed contract.
func NewHermezFilterer(address common.Address, filterer bind.ContractFilterer) (*HermezFilterer, error) {
	contract, err := bindHermez(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HermezFilterer{contract: contract}, nil
}

// bindHermez binds a generic wrapper to an already deployed contract.
func bindHermez(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HermezABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hermez *HermezRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hermez.Contract.HermezCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hermez *HermezRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hermez.Contract.HermezTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hermez *HermezRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hermez.Contract.HermezTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Hermez *HermezCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Hermez.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Hermez *HermezTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hermez.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Hermez *HermezTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Hermez.Contract.contract.Transact(opts, method, params...)
}

// ABSOLUTEMAXL1L2BATCHTIMEOUT is a free data retrieval call binding the contract method 0x95a09f2a.
//
// Solidity: function ABSOLUTE_MAX_L1L2BATCHTIMEOUT() view returns(uint8)
func (_Hermez *HermezCaller) ABSOLUTEMAXL1L2BATCHTIMEOUT(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "ABSOLUTE_MAX_L1L2BATCHTIMEOUT")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// ABSOLUTEMAXL1L2BATCHTIMEOUT is a free data retrieval call binding the contract method 0x95a09f2a.
//
// Solidity: function ABSOLUTE_MAX_L1L2BATCHTIMEOUT() view returns(uint8)
func (_Hermez *HermezSession) ABSOLUTEMAXL1L2BATCHTIMEOUT() (uint8, error) {
	return _Hermez.Contract.ABSOLUTEMAXL1L2BATCHTIMEOUT(&_Hermez.CallOpts)
}

// ABSOLUTEMAXL1L2BATCHTIMEOUT is a free data retrieval call binding the contract method 0x95a09f2a.
//
// Solidity: function ABSOLUTE_MAX_L1L2BATCHTIMEOUT() view returns(uint8)
func (_Hermez *HermezCallerSession) ABSOLUTEMAXL1L2BATCHTIMEOUT() (uint8, error) {
	return _Hermez.Contract.ABSOLUTEMAXL1L2BATCHTIMEOUT(&_Hermez.CallOpts)
}

// ACCOUNTCREATIONHASH is a free data retrieval call binding the contract method 0x1300aff0.
//
// Solidity: function ACCOUNT_CREATION_HASH() view returns(bytes32)
func (_Hermez *HermezCaller) ACCOUNTCREATIONHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "ACCOUNT_CREATION_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ACCOUNTCREATIONHASH is a free data retrieval call binding the contract method 0x1300aff0.
//
// Solidity: function ACCOUNT_CREATION_HASH() view returns(bytes32)
func (_Hermez *HermezSession) ACCOUNTCREATIONHASH() ([32]byte, error) {
	return _Hermez.Contract.ACCOUNTCREATIONHASH(&_Hermez.CallOpts)
}

// ACCOUNTCREATIONHASH is a free data retrieval call binding the contract method 0x1300aff0.
//
// Solidity: function ACCOUNT_CREATION_HASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) ACCOUNTCREATIONHASH() ([32]byte, error) {
	return _Hermez.Contract.ACCOUNTCREATIONHASH(&_Hermez.CallOpts)
}

// AUTHORISETYPEHASH is a free data retrieval call binding the contract method 0xe62f6b92.
//
// Solidity: function AUTHORISE_TYPEHASH() view returns(bytes32)
func (_Hermez *HermezCaller) AUTHORISETYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "AUTHORISE_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// AUTHORISETYPEHASH is a free data retrieval call binding the contract method 0xe62f6b92.
//
// Solidity: function AUTHORISE_TYPEHASH() view returns(bytes32)
func (_Hermez *HermezSession) AUTHORISETYPEHASH() ([32]byte, error) {
	return _Hermez.Contract.AUTHORISETYPEHASH(&_Hermez.CallOpts)
}

// AUTHORISETYPEHASH is a free data retrieval call binding the contract method 0xe62f6b92.
//
// Solidity: function AUTHORISE_TYPEHASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) AUTHORISETYPEHASH() ([32]byte, error) {
	return _Hermez.Contract.AUTHORISETYPEHASH(&_Hermez.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32 domainSeparator)
func (_Hermez *HermezCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32 domainSeparator)
func (_Hermez *HermezSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Hermez.Contract.DOMAINSEPARATOR(&_Hermez.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32 domainSeparator)
func (_Hermez *HermezCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Hermez.Contract.DOMAINSEPARATOR(&_Hermez.CallOpts)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Hermez *HermezCaller) EIP712DOMAINHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "EIP712DOMAIN_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Hermez *HermezSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _Hermez.Contract.EIP712DOMAINHASH(&_Hermez.CallOpts)
}

// EIP712DOMAINHASH is a free data retrieval call binding the contract method 0xc473af33.
//
// Solidity: function EIP712DOMAIN_HASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) EIP712DOMAINHASH() ([32]byte, error) {
	return _Hermez.Contract.EIP712DOMAINHASH(&_Hermez.CallOpts)
}

// HERMEZNETWORKHASH is a free data retrieval call binding the contract method 0xf1f2fcab.
//
// Solidity: function HERMEZ_NETWORK_HASH() view returns(bytes32)
func (_Hermez *HermezCaller) HERMEZNETWORKHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "HERMEZ_NETWORK_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HERMEZNETWORKHASH is a free data retrieval call binding the contract method 0xf1f2fcab.
//
// Solidity: function HERMEZ_NETWORK_HASH() view returns(bytes32)
func (_Hermez *HermezSession) HERMEZNETWORKHASH() ([32]byte, error) {
	return _Hermez.Contract.HERMEZNETWORKHASH(&_Hermez.CallOpts)
}

// HERMEZNETWORKHASH is a free data retrieval call binding the contract method 0xf1f2fcab.
//
// Solidity: function HERMEZ_NETWORK_HASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) HERMEZNETWORKHASH() ([32]byte, error) {
	return _Hermez.Contract.HERMEZNETWORKHASH(&_Hermez.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Hermez *HermezCaller) NAMEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "NAME_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Hermez *HermezSession) NAMEHASH() ([32]byte, error) {
	return _Hermez.Contract.NAMEHASH(&_Hermez.CallOpts)
}

// NAMEHASH is a free data retrieval call binding the contract method 0x04622c2e.
//
// Solidity: function NAME_HASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) NAMEHASH() ([32]byte, error) {
	return _Hermez.Contract.NAMEHASH(&_Hermez.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Hermez *HermezCaller) VERSIONHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "VERSION_HASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Hermez *HermezSession) VERSIONHASH() ([32]byte, error) {
	return _Hermez.Contract.VERSIONHASH(&_Hermez.CallOpts)
}

// VERSIONHASH is a free data retrieval call binding the contract method 0x9e4e7318.
//
// Solidity: function VERSION_HASH() view returns(bytes32)
func (_Hermez *HermezCallerSession) VERSIONHASH() ([32]byte, error) {
	return _Hermez.Contract.VERSIONHASH(&_Hermez.CallOpts)
}

// Buckets is a free data retrieval call binding the contract method 0x061d0964.
//
// Solidity: function buckets(int256 ) view returns(uint256)
func (_Hermez *HermezCaller) Buckets(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "buckets", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Buckets is a free data retrieval call binding the contract method 0x061d0964.
//
// Solidity: function buckets(int256 ) view returns(uint256)
func (_Hermez *HermezSession) Buckets(arg0 *big.Int) (*big.Int, error) {
	return _Hermez.Contract.Buckets(&_Hermez.CallOpts, arg0)
}

// Buckets is a free data retrieval call binding the contract method 0x061d0964.
//
// Solidity: function buckets(int256 ) view returns(uint256)
func (_Hermez *HermezCallerSession) Buckets(arg0 *big.Int) (*big.Int, error) {
	return _Hermez.Contract.Buckets(&_Hermez.CallOpts, arg0)
}

// ExitNullifierMap is a free data retrieval call binding the contract method 0xf84f92ee.
//
// Solidity: function exitNullifierMap(uint32 , uint48 ) view returns(bool)
func (_Hermez *HermezCaller) ExitNullifierMap(opts *bind.CallOpts, arg0 uint32, arg1 *big.Int) (bool, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "exitNullifierMap", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// ExitNullifierMap is a free data retrieval call binding the contract method 0xf84f92ee.
//
// Solidity: function exitNullifierMap(uint32 , uint48 ) view returns(bool)
func (_Hermez *HermezSession) ExitNullifierMap(arg0 uint32, arg1 *big.Int) (bool, error) {
	return _Hermez.Contract.ExitNullifierMap(&_Hermez.CallOpts, arg0, arg1)
}

// ExitNullifierMap is a free data retrieval call binding the contract method 0xf84f92ee.
//
// Solidity: function exitNullifierMap(uint32 , uint48 ) view returns(bool)
func (_Hermez *HermezCallerSession) ExitNullifierMap(arg0 uint32, arg1 *big.Int) (bool, error) {
	return _Hermez.Contract.ExitNullifierMap(&_Hermez.CallOpts, arg0, arg1)
}

// ExitRootsMap is a free data retrieval call binding the contract method 0x3ee641ea.
//
// Solidity: function exitRootsMap(uint32 ) view returns(uint256)
func (_Hermez *HermezCaller) ExitRootsMap(opts *bind.CallOpts, arg0 uint32) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "exitRootsMap", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ExitRootsMap is a free data retrieval call binding the contract method 0x3ee641ea.
//
// Solidity: function exitRootsMap(uint32 ) view returns(uint256)
func (_Hermez *HermezSession) ExitRootsMap(arg0 uint32) (*big.Int, error) {
	return _Hermez.Contract.ExitRootsMap(&_Hermez.CallOpts, arg0)
}

// ExitRootsMap is a free data retrieval call binding the contract method 0x3ee641ea.
//
// Solidity: function exitRootsMap(uint32 ) view returns(uint256)
func (_Hermez *HermezCallerSession) ExitRootsMap(arg0 uint32) (*big.Int, error) {
	return _Hermez.Contract.ExitRootsMap(&_Hermez.CallOpts, arg0)
}

// FeeAddToken is a free data retrieval call binding the contract method 0xbded9bb8.
//
// Solidity: function feeAddToken() view returns(uint256)
func (_Hermez *HermezCaller) FeeAddToken(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "feeAddToken")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeeAddToken is a free data retrieval call binding the contract method 0xbded9bb8.
//
// Solidity: function feeAddToken() view returns(uint256)
func (_Hermez *HermezSession) FeeAddToken() (*big.Int, error) {
	return _Hermez.Contract.FeeAddToken(&_Hermez.CallOpts)
}

// FeeAddToken is a free data retrieval call binding the contract method 0xbded9bb8.
//
// Solidity: function feeAddToken() view returns(uint256)
func (_Hermez *HermezCallerSession) FeeAddToken() (*big.Int, error) {
	return _Hermez.Contract.FeeAddToken(&_Hermez.CallOpts)
}

// ForgeL1L2BatchTimeout is a free data retrieval call binding the contract method 0xa3275838.
//
// Solidity: function forgeL1L2BatchTimeout() view returns(uint8)
func (_Hermez *HermezCaller) ForgeL1L2BatchTimeout(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "forgeL1L2BatchTimeout")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// ForgeL1L2BatchTimeout is a free data retrieval call binding the contract method 0xa3275838.
//
// Solidity: function forgeL1L2BatchTimeout() view returns(uint8)
func (_Hermez *HermezSession) ForgeL1L2BatchTimeout() (uint8, error) {
	return _Hermez.Contract.ForgeL1L2BatchTimeout(&_Hermez.CallOpts)
}

// ForgeL1L2BatchTimeout is a free data retrieval call binding the contract method 0xa3275838.
//
// Solidity: function forgeL1L2BatchTimeout() view returns(uint8)
func (_Hermez *HermezCallerSession) ForgeL1L2BatchTimeout() (uint8, error) {
	return _Hermez.Contract.ForgeL1L2BatchTimeout(&_Hermez.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Hermez *HermezCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "getChainId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Hermez *HermezSession) GetChainId() (*big.Int, error) {
	return _Hermez.Contract.GetChainId(&_Hermez.CallOpts)
}

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_Hermez *HermezCallerSession) GetChainId() (*big.Int, error) {
	return _Hermez.Contract.GetChainId(&_Hermez.CallOpts)
}

// HermezAuctionContract is a free data retrieval call binding the contract method 0x2bd83626.
//
// Solidity: function hermezAuctionContract() view returns(address)
func (_Hermez *HermezCaller) HermezAuctionContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "hermezAuctionContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezAuctionContract is a free data retrieval call binding the contract method 0x2bd83626.
//
// Solidity: function hermezAuctionContract() view returns(address)
func (_Hermez *HermezSession) HermezAuctionContract() (common.Address, error) {
	return _Hermez.Contract.HermezAuctionContract(&_Hermez.CallOpts)
}

// HermezAuctionContract is a free data retrieval call binding the contract method 0x2bd83626.
//
// Solidity: function hermezAuctionContract() view returns(address)
func (_Hermez *HermezCallerSession) HermezAuctionContract() (common.Address, error) {
	return _Hermez.Contract.HermezAuctionContract(&_Hermez.CallOpts)
}

// HermezGovernanceAddress is a free data retrieval call binding the contract method 0x013f7852.
//
// Solidity: function hermezGovernanceAddress() view returns(address)
func (_Hermez *HermezCaller) HermezGovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "hermezGovernanceAddress")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// HermezGovernanceAddress is a free data retrieval call binding the contract method 0x013f7852.
//
// Solidity: function hermezGovernanceAddress() view returns(address)
func (_Hermez *HermezSession) HermezGovernanceAddress() (common.Address, error) {
	return _Hermez.Contract.HermezGovernanceAddress(&_Hermez.CallOpts)
}

// HermezGovernanceAddress is a free data retrieval call binding the contract method 0x013f7852.
//
// Solidity: function hermezGovernanceAddress() view returns(address)
func (_Hermez *HermezCallerSession) HermezGovernanceAddress() (common.Address, error) {
	return _Hermez.Contract.HermezGovernanceAddress(&_Hermez.CallOpts)
}

// InstantWithdrawalViewer is a free data retrieval call binding the contract method 0x375110aa.
//
// Solidity: function instantWithdrawalViewer(address tokenAddress, uint192 amount) view returns(bool)
func (_Hermez *HermezCaller) InstantWithdrawalViewer(opts *bind.CallOpts, tokenAddress common.Address, amount *big.Int) (bool, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "instantWithdrawalViewer", tokenAddress, amount)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// InstantWithdrawalViewer is a free data retrieval call binding the contract method 0x375110aa.
//
// Solidity: function instantWithdrawalViewer(address tokenAddress, uint192 amount) view returns(bool)
func (_Hermez *HermezSession) InstantWithdrawalViewer(tokenAddress common.Address, amount *big.Int) (bool, error) {
	return _Hermez.Contract.InstantWithdrawalViewer(&_Hermez.CallOpts, tokenAddress, amount)
}

// InstantWithdrawalViewer is a free data retrieval call binding the contract method 0x375110aa.
//
// Solidity: function instantWithdrawalViewer(address tokenAddress, uint192 amount) view returns(bool)
func (_Hermez *HermezCallerSession) InstantWithdrawalViewer(tokenAddress common.Address, amount *big.Int) (bool, error) {
	return _Hermez.Contract.InstantWithdrawalViewer(&_Hermez.CallOpts, tokenAddress, amount)
}

// L1L2TxsDataHashMap is a free data retrieval call binding the contract method 0xce5ec65a.
//
// Solidity: function l1L2TxsDataHashMap(uint32 ) view returns(bytes32)
func (_Hermez *HermezCaller) L1L2TxsDataHashMap(opts *bind.CallOpts, arg0 uint32) ([32]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "l1L2TxsDataHashMap", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// L1L2TxsDataHashMap is a free data retrieval call binding the contract method 0xce5ec65a.
//
// Solidity: function l1L2TxsDataHashMap(uint32 ) view returns(bytes32)
func (_Hermez *HermezSession) L1L2TxsDataHashMap(arg0 uint32) ([32]byte, error) {
	return _Hermez.Contract.L1L2TxsDataHashMap(&_Hermez.CallOpts, arg0)
}

// L1L2TxsDataHashMap is a free data retrieval call binding the contract method 0xce5ec65a.
//
// Solidity: function l1L2TxsDataHashMap(uint32 ) view returns(bytes32)
func (_Hermez *HermezCallerSession) L1L2TxsDataHashMap(arg0 uint32) ([32]byte, error) {
	return _Hermez.Contract.L1L2TxsDataHashMap(&_Hermez.CallOpts, arg0)
}

// LastForgedBatch is a free data retrieval call binding the contract method 0x44e0b2ce.
//
// Solidity: function lastForgedBatch() view returns(uint32)
func (_Hermez *HermezCaller) LastForgedBatch(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "lastForgedBatch")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LastForgedBatch is a free data retrieval call binding the contract method 0x44e0b2ce.
//
// Solidity: function lastForgedBatch() view returns(uint32)
func (_Hermez *HermezSession) LastForgedBatch() (uint32, error) {
	return _Hermez.Contract.LastForgedBatch(&_Hermez.CallOpts)
}

// LastForgedBatch is a free data retrieval call binding the contract method 0x44e0b2ce.
//
// Solidity: function lastForgedBatch() view returns(uint32)
func (_Hermez *HermezCallerSession) LastForgedBatch() (uint32, error) {
	return _Hermez.Contract.LastForgedBatch(&_Hermez.CallOpts)
}

// LastIdx is a free data retrieval call binding the contract method 0xd486645c.
//
// Solidity: function lastIdx() view returns(uint48)
func (_Hermez *HermezCaller) LastIdx(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "lastIdx")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LastIdx is a free data retrieval call binding the contract method 0xd486645c.
//
// Solidity: function lastIdx() view returns(uint48)
func (_Hermez *HermezSession) LastIdx() (*big.Int, error) {
	return _Hermez.Contract.LastIdx(&_Hermez.CallOpts)
}

// LastIdx is a free data retrieval call binding the contract method 0xd486645c.
//
// Solidity: function lastIdx() view returns(uint48)
func (_Hermez *HermezCallerSession) LastIdx() (*big.Int, error) {
	return _Hermez.Contract.LastIdx(&_Hermez.CallOpts)
}

// LastL1L2Batch is a free data retrieval call binding the contract method 0x84ef9ed4.
//
// Solidity: function lastL1L2Batch() view returns(uint64)
func (_Hermez *HermezCaller) LastL1L2Batch(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "lastL1L2Batch")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// LastL1L2Batch is a free data retrieval call binding the contract method 0x84ef9ed4.
//
// Solidity: function lastL1L2Batch() view returns(uint64)
func (_Hermez *HermezSession) LastL1L2Batch() (uint64, error) {
	return _Hermez.Contract.LastL1L2Batch(&_Hermez.CallOpts)
}

// LastL1L2Batch is a free data retrieval call binding the contract method 0x84ef9ed4.
//
// Solidity: function lastL1L2Batch() view returns(uint64)
func (_Hermez *HermezCallerSession) LastL1L2Batch() (uint64, error) {
	return _Hermez.Contract.LastL1L2Batch(&_Hermez.CallOpts)
}

// MapL1TxQueue is a free data retrieval call binding the contract method 0xdc3e718e.
//
// Solidity: function mapL1TxQueue(uint32 ) view returns(bytes)
func (_Hermez *HermezCaller) MapL1TxQueue(opts *bind.CallOpts, arg0 uint32) ([]byte, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "mapL1TxQueue", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// MapL1TxQueue is a free data retrieval call binding the contract method 0xdc3e718e.
//
// Solidity: function mapL1TxQueue(uint32 ) view returns(bytes)
func (_Hermez *HermezSession) MapL1TxQueue(arg0 uint32) ([]byte, error) {
	return _Hermez.Contract.MapL1TxQueue(&_Hermez.CallOpts, arg0)
}

// MapL1TxQueue is a free data retrieval call binding the contract method 0xdc3e718e.
//
// Solidity: function mapL1TxQueue(uint32 ) view returns(bytes)
func (_Hermez *HermezCallerSession) MapL1TxQueue(arg0 uint32) ([]byte, error) {
	return _Hermez.Contract.MapL1TxQueue(&_Hermez.CallOpts, arg0)
}

// NBuckets is a free data retrieval call binding the contract method 0x07feef6e.
//
// Solidity: function nBuckets() view returns(uint256)
func (_Hermez *HermezCaller) NBuckets(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "nBuckets")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NBuckets is a free data retrieval call binding the contract method 0x07feef6e.
//
// Solidity: function nBuckets() view returns(uint256)
func (_Hermez *HermezSession) NBuckets() (*big.Int, error) {
	return _Hermez.Contract.NBuckets(&_Hermez.CallOpts)
}

// NBuckets is a free data retrieval call binding the contract method 0x07feef6e.
//
// Solidity: function nBuckets() view returns(uint256)
func (_Hermez *HermezCallerSession) NBuckets() (*big.Int, error) {
	return _Hermez.Contract.NBuckets(&_Hermez.CallOpts)
}

// NextL1FillingQueue is a free data retrieval call binding the contract method 0x0ee8e52b.
//
// Solidity: function nextL1FillingQueue() view returns(uint32)
func (_Hermez *HermezCaller) NextL1FillingQueue(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "nextL1FillingQueue")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// NextL1FillingQueue is a free data retrieval call binding the contract method 0x0ee8e52b.
//
// Solidity: function nextL1FillingQueue() view returns(uint32)
func (_Hermez *HermezSession) NextL1FillingQueue() (uint32, error) {
	return _Hermez.Contract.NextL1FillingQueue(&_Hermez.CallOpts)
}

// NextL1FillingQueue is a free data retrieval call binding the contract method 0x0ee8e52b.
//
// Solidity: function nextL1FillingQueue() view returns(uint32)
func (_Hermez *HermezCallerSession) NextL1FillingQueue() (uint32, error) {
	return _Hermez.Contract.NextL1FillingQueue(&_Hermez.CallOpts)
}

// NextL1ToForgeQueue is a free data retrieval call binding the contract method 0xd0f32e67.
//
// Solidity: function nextL1ToForgeQueue() view returns(uint32)
func (_Hermez *HermezCaller) NextL1ToForgeQueue(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "nextL1ToForgeQueue")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// NextL1ToForgeQueue is a free data retrieval call binding the contract method 0xd0f32e67.
//
// Solidity: function nextL1ToForgeQueue() view returns(uint32)
func (_Hermez *HermezSession) NextL1ToForgeQueue() (uint32, error) {
	return _Hermez.Contract.NextL1ToForgeQueue(&_Hermez.CallOpts)
}

// NextL1ToForgeQueue is a free data retrieval call binding the contract method 0xd0f32e67.
//
// Solidity: function nextL1ToForgeQueue() view returns(uint32)
func (_Hermez *HermezCallerSession) NextL1ToForgeQueue() (uint32, error) {
	return _Hermez.Contract.NextL1ToForgeQueue(&_Hermez.CallOpts)
}

// PackBucket is a free data retrieval call binding the contract method 0xccd226a7.
//
// Solidity: function packBucket(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals) pure returns(uint256 ret)
func (_Hermez *HermezCaller) PackBucket(opts *bind.CallOpts, ceilUSD *big.Int, blockStamp *big.Int, withdrawals *big.Int, rateBlocks *big.Int, rateWithdrawals *big.Int, maxWithdrawals *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "packBucket", ceilUSD, blockStamp, withdrawals, rateBlocks, rateWithdrawals, maxWithdrawals)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// PackBucket is a free data retrieval call binding the contract method 0xccd226a7.
//
// Solidity: function packBucket(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals) pure returns(uint256 ret)
func (_Hermez *HermezSession) PackBucket(ceilUSD *big.Int, blockStamp *big.Int, withdrawals *big.Int, rateBlocks *big.Int, rateWithdrawals *big.Int, maxWithdrawals *big.Int) (*big.Int, error) {
	return _Hermez.Contract.PackBucket(&_Hermez.CallOpts, ceilUSD, blockStamp, withdrawals, rateBlocks, rateWithdrawals, maxWithdrawals)
}

// PackBucket is a free data retrieval call binding the contract method 0xccd226a7.
//
// Solidity: function packBucket(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals) pure returns(uint256 ret)
func (_Hermez *HermezCallerSession) PackBucket(ceilUSD *big.Int, blockStamp *big.Int, withdrawals *big.Int, rateBlocks *big.Int, rateWithdrawals *big.Int, maxWithdrawals *big.Int) (*big.Int, error) {
	return _Hermez.Contract.PackBucket(&_Hermez.CallOpts, ceilUSD, blockStamp, withdrawals, rateBlocks, rateWithdrawals, maxWithdrawals)
}

// RegisterTokensCount is a free data retrieval call binding the contract method 0x9f34e9a3.
//
// Solidity: function registerTokensCount() view returns(uint256)
func (_Hermez *HermezCaller) RegisterTokensCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "registerTokensCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RegisterTokensCount is a free data retrieval call binding the contract method 0x9f34e9a3.
//
// Solidity: function registerTokensCount() view returns(uint256)
func (_Hermez *HermezSession) RegisterTokensCount() (*big.Int, error) {
	return _Hermez.Contract.RegisterTokensCount(&_Hermez.CallOpts)
}

// RegisterTokensCount is a free data retrieval call binding the contract method 0x9f34e9a3.
//
// Solidity: function registerTokensCount() view returns(uint256)
func (_Hermez *HermezCallerSession) RegisterTokensCount() (*big.Int, error) {
	return _Hermez.Contract.RegisterTokensCount(&_Hermez.CallOpts)
}

// RollupVerifiers is a free data retrieval call binding the contract method 0x38330200.
//
// Solidity: function rollupVerifiers(uint256 ) view returns(address verifierInterface, uint256 maxTx, uint256 nLevels)
func (_Hermez *HermezCaller) RollupVerifiers(opts *bind.CallOpts, arg0 *big.Int) (struct {
	VerifierInterface common.Address
	MaxTx             *big.Int
	NLevels           *big.Int
}, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "rollupVerifiers", arg0)

	outstruct := new(struct {
		VerifierInterface common.Address
		MaxTx             *big.Int
		NLevels           *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.VerifierInterface = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.MaxTx = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.NLevels = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// RollupVerifiers is a free data retrieval call binding the contract method 0x38330200.
//
// Solidity: function rollupVerifiers(uint256 ) view returns(address verifierInterface, uint256 maxTx, uint256 nLevels)
func (_Hermez *HermezSession) RollupVerifiers(arg0 *big.Int) (struct {
	VerifierInterface common.Address
	MaxTx             *big.Int
	NLevels           *big.Int
}, error) {
	return _Hermez.Contract.RollupVerifiers(&_Hermez.CallOpts, arg0)
}

// RollupVerifiers is a free data retrieval call binding the contract method 0x38330200.
//
// Solidity: function rollupVerifiers(uint256 ) view returns(address verifierInterface, uint256 maxTx, uint256 nLevels)
func (_Hermez *HermezCallerSession) RollupVerifiers(arg0 *big.Int) (struct {
	VerifierInterface common.Address
	MaxTx             *big.Int
	NLevels           *big.Int
}, error) {
	return _Hermez.Contract.RollupVerifiers(&_Hermez.CallOpts, arg0)
}

// RollupVerifiersLength is a free data retrieval call binding the contract method 0x7ba3a5e0.
//
// Solidity: function rollupVerifiersLength() view returns(uint256)
func (_Hermez *HermezCaller) RollupVerifiersLength(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "rollupVerifiersLength")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RollupVerifiersLength is a free data retrieval call binding the contract method 0x7ba3a5e0.
//
// Solidity: function rollupVerifiersLength() view returns(uint256)
func (_Hermez *HermezSession) RollupVerifiersLength() (*big.Int, error) {
	return _Hermez.Contract.RollupVerifiersLength(&_Hermez.CallOpts)
}

// RollupVerifiersLength is a free data retrieval call binding the contract method 0x7ba3a5e0.
//
// Solidity: function rollupVerifiersLength() view returns(uint256)
func (_Hermez *HermezCallerSession) RollupVerifiersLength() (*big.Int, error) {
	return _Hermez.Contract.RollupVerifiersLength(&_Hermez.CallOpts)
}

// StateRootMap is a free data retrieval call binding the contract method 0x9e00d7ea.
//
// Solidity: function stateRootMap(uint32 ) view returns(uint256)
func (_Hermez *HermezCaller) StateRootMap(opts *bind.CallOpts, arg0 uint32) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "stateRootMap", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateRootMap is a free data retrieval call binding the contract method 0x9e00d7ea.
//
// Solidity: function stateRootMap(uint32 ) view returns(uint256)
func (_Hermez *HermezSession) StateRootMap(arg0 uint32) (*big.Int, error) {
	return _Hermez.Contract.StateRootMap(&_Hermez.CallOpts, arg0)
}

// StateRootMap is a free data retrieval call binding the contract method 0x9e00d7ea.
//
// Solidity: function stateRootMap(uint32 ) view returns(uint256)
func (_Hermez *HermezCallerSession) StateRootMap(arg0 uint32) (*big.Int, error) {
	return _Hermez.Contract.StateRootMap(&_Hermez.CallOpts, arg0)
}

// TokenExchange is a free data retrieval call binding the contract method 0x0dd94b96.
//
// Solidity: function tokenExchange(address ) view returns(uint64)
func (_Hermez *HermezCaller) TokenExchange(opts *bind.CallOpts, arg0 common.Address) (uint64, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "tokenExchange", arg0)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// TokenExchange is a free data retrieval call binding the contract method 0x0dd94b96.
//
// Solidity: function tokenExchange(address ) view returns(uint64)
func (_Hermez *HermezSession) TokenExchange(arg0 common.Address) (uint64, error) {
	return _Hermez.Contract.TokenExchange(&_Hermez.CallOpts, arg0)
}

// TokenExchange is a free data retrieval call binding the contract method 0x0dd94b96.
//
// Solidity: function tokenExchange(address ) view returns(uint64)
func (_Hermez *HermezCallerSession) TokenExchange(arg0 common.Address) (uint64, error) {
	return _Hermez.Contract.TokenExchange(&_Hermez.CallOpts, arg0)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Hermez *HermezCaller) TokenHEZ(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "tokenHEZ")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Hermez *HermezSession) TokenHEZ() (common.Address, error) {
	return _Hermez.Contract.TokenHEZ(&_Hermez.CallOpts)
}

// TokenHEZ is a free data retrieval call binding the contract method 0x79a135e3.
//
// Solidity: function tokenHEZ() view returns(address)
func (_Hermez *HermezCallerSession) TokenHEZ() (common.Address, error) {
	return _Hermez.Contract.TokenHEZ(&_Hermez.CallOpts)
}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Hermez *HermezCaller) TokenList(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "tokenList", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Hermez *HermezSession) TokenList(arg0 *big.Int) (common.Address, error) {
	return _Hermez.Contract.TokenList(&_Hermez.CallOpts, arg0)
}

// TokenList is a free data retrieval call binding the contract method 0x9ead7222.
//
// Solidity: function tokenList(uint256 ) view returns(address)
func (_Hermez *HermezCallerSession) TokenList(arg0 *big.Int) (common.Address, error) {
	return _Hermez.Contract.TokenList(&_Hermez.CallOpts, arg0)
}

// TokenMap is a free data retrieval call binding the contract method 0x004aca6e.
//
// Solidity: function tokenMap(address ) view returns(uint256)
func (_Hermez *HermezCaller) TokenMap(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "tokenMap", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenMap is a free data retrieval call binding the contract method 0x004aca6e.
//
// Solidity: function tokenMap(address ) view returns(uint256)
func (_Hermez *HermezSession) TokenMap(arg0 common.Address) (*big.Int, error) {
	return _Hermez.Contract.TokenMap(&_Hermez.CallOpts, arg0)
}

// TokenMap is a free data retrieval call binding the contract method 0x004aca6e.
//
// Solidity: function tokenMap(address ) view returns(uint256)
func (_Hermez *HermezCallerSession) TokenMap(arg0 common.Address) (*big.Int, error) {
	return _Hermez.Contract.TokenMap(&_Hermez.CallOpts, arg0)
}

// UnpackBucket is a free data retrieval call binding the contract method 0x3f267155.
//
// Solidity: function unpackBucket(uint256 bucket) pure returns(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals)
func (_Hermez *HermezCaller) UnpackBucket(opts *bind.CallOpts, bucket *big.Int) (struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "unpackBucket", bucket)

	outstruct := new(struct {
		CeilUSD         *big.Int
		BlockStamp      *big.Int
		Withdrawals     *big.Int
		RateBlocks      *big.Int
		RateWithdrawals *big.Int
		MaxWithdrawals  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.CeilUSD = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.BlockStamp = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Withdrawals = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.RateBlocks = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.RateWithdrawals = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.MaxWithdrawals = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// UnpackBucket is a free data retrieval call binding the contract method 0x3f267155.
//
// Solidity: function unpackBucket(uint256 bucket) pure returns(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals)
func (_Hermez *HermezSession) UnpackBucket(bucket *big.Int) (struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}, error) {
	return _Hermez.Contract.UnpackBucket(&_Hermez.CallOpts, bucket)
}

// UnpackBucket is a free data retrieval call binding the contract method 0x3f267155.
//
// Solidity: function unpackBucket(uint256 bucket) pure returns(uint256 ceilUSD, uint256 blockStamp, uint256 withdrawals, uint256 rateBlocks, uint256 rateWithdrawals, uint256 maxWithdrawals)
func (_Hermez *HermezCallerSession) UnpackBucket(bucket *big.Int) (struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}, error) {
	return _Hermez.Contract.UnpackBucket(&_Hermez.CallOpts, bucket)
}

// WithdrawDelayerContract is a free data retrieval call binding the contract method 0x1b0a8223.
//
// Solidity: function withdrawDelayerContract() view returns(address)
func (_Hermez *HermezCaller) WithdrawDelayerContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "withdrawDelayerContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WithdrawDelayerContract is a free data retrieval call binding the contract method 0x1b0a8223.
//
// Solidity: function withdrawDelayerContract() view returns(address)
func (_Hermez *HermezSession) WithdrawDelayerContract() (common.Address, error) {
	return _Hermez.Contract.WithdrawDelayerContract(&_Hermez.CallOpts)
}

// WithdrawDelayerContract is a free data retrieval call binding the contract method 0x1b0a8223.
//
// Solidity: function withdrawDelayerContract() view returns(address)
func (_Hermez *HermezCallerSession) WithdrawDelayerContract() (common.Address, error) {
	return _Hermez.Contract.WithdrawDelayerContract(&_Hermez.CallOpts)
}

// WithdrawVerifier is a free data retrieval call binding the contract method 0x864eb164.
//
// Solidity: function withdrawVerifier() view returns(address)
func (_Hermez *HermezCaller) WithdrawVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "withdrawVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// WithdrawVerifier is a free data retrieval call binding the contract method 0x864eb164.
//
// Solidity: function withdrawVerifier() view returns(address)
func (_Hermez *HermezSession) WithdrawVerifier() (common.Address, error) {
	return _Hermez.Contract.WithdrawVerifier(&_Hermez.CallOpts)
}

// WithdrawVerifier is a free data retrieval call binding the contract method 0x864eb164.
//
// Solidity: function withdrawVerifier() view returns(address)
func (_Hermez *HermezCallerSession) WithdrawVerifier() (common.Address, error) {
	return _Hermez.Contract.WithdrawVerifier(&_Hermez.CallOpts)
}

// WithdrawalDelay is a free data retrieval call binding the contract method 0xa7ab6961.
//
// Solidity: function withdrawalDelay() view returns(uint64)
func (_Hermez *HermezCaller) WithdrawalDelay(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _Hermez.contract.Call(opts, &out, "withdrawalDelay")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// WithdrawalDelay is a free data retrieval call binding the contract method 0xa7ab6961.
//
// Solidity: function withdrawalDelay() view returns(uint64)
func (_Hermez *HermezSession) WithdrawalDelay() (uint64, error) {
	return _Hermez.Contract.WithdrawalDelay(&_Hermez.CallOpts)
}

// WithdrawalDelay is a free data retrieval call binding the contract method 0xa7ab6961.
//
// Solidity: function withdrawalDelay() view returns(uint64)
func (_Hermez *HermezCallerSession) WithdrawalDelay() (uint64, error) {
	return _Hermez.Contract.WithdrawalDelay(&_Hermez.CallOpts)
}

// AddL1Transaction is a paid mutator transaction binding the contract method 0xc7273053.
//
// Solidity: function addL1Transaction(uint256 babyPubKey, uint48 fromIdx, uint40 loadAmountF, uint40 amountF, uint32 tokenID, uint48 toIdx, bytes permit) payable returns()
func (_Hermez *HermezTransactor) AddL1Transaction(opts *bind.TransactOpts, babyPubKey *big.Int, fromIdx *big.Int, loadAmountF *big.Int, amountF *big.Int, tokenID uint32, toIdx *big.Int, permit []byte) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "addL1Transaction", babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx, permit)
}

// AddL1Transaction is a paid mutator transaction binding the contract method 0xc7273053.
//
// Solidity: function addL1Transaction(uint256 babyPubKey, uint48 fromIdx, uint40 loadAmountF, uint40 amountF, uint32 tokenID, uint48 toIdx, bytes permit) payable returns()
func (_Hermez *HermezSession) AddL1Transaction(babyPubKey *big.Int, fromIdx *big.Int, loadAmountF *big.Int, amountF *big.Int, tokenID uint32, toIdx *big.Int, permit []byte) (*types.Transaction, error) {
	return _Hermez.Contract.AddL1Transaction(&_Hermez.TransactOpts, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx, permit)
}

// AddL1Transaction is a paid mutator transaction binding the contract method 0xc7273053.
//
// Solidity: function addL1Transaction(uint256 babyPubKey, uint48 fromIdx, uint40 loadAmountF, uint40 amountF, uint32 tokenID, uint48 toIdx, bytes permit) payable returns()
func (_Hermez *HermezTransactorSession) AddL1Transaction(babyPubKey *big.Int, fromIdx *big.Int, loadAmountF *big.Int, amountF *big.Int, tokenID uint32, toIdx *big.Int, permit []byte) (*types.Transaction, error) {
	return _Hermez.Contract.AddL1Transaction(&_Hermez.TransactOpts, babyPubKey, fromIdx, loadAmountF, amountF, tokenID, toIdx, permit)
}

// AddToken is a paid mutator transaction binding the contract method 0x70c2f1c0.
//
// Solidity: function addToken(address tokenAddress, bytes permit) returns()
func (_Hermez *HermezTransactor) AddToken(opts *bind.TransactOpts, tokenAddress common.Address, permit []byte) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "addToken", tokenAddress, permit)
}

// AddToken is a paid mutator transaction binding the contract method 0x70c2f1c0.
//
// Solidity: function addToken(address tokenAddress, bytes permit) returns()
func (_Hermez *HermezSession) AddToken(tokenAddress common.Address, permit []byte) (*types.Transaction, error) {
	return _Hermez.Contract.AddToken(&_Hermez.TransactOpts, tokenAddress, permit)
}

// AddToken is a paid mutator transaction binding the contract method 0x70c2f1c0.
//
// Solidity: function addToken(address tokenAddress, bytes permit) returns()
func (_Hermez *HermezTransactorSession) AddToken(tokenAddress common.Address, permit []byte) (*types.Transaction, error) {
	return _Hermez.Contract.AddToken(&_Hermez.TransactOpts, tokenAddress, permit)
}

// ForgeBatch is a paid mutator transaction binding the contract method 0x6e7e1365.
//
// Solidity: function forgeBatch(uint48 newLastIdx, uint256 newStRoot, uint256 newExitRoot, bytes encodedL1CoordinatorTx, bytes l1L2TxsData, bytes feeIdxCoordinator, uint8 verifierIdx, bool l1Batch, uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC) returns()
func (_Hermez *HermezTransactor) ForgeBatch(opts *bind.TransactOpts, newLastIdx *big.Int, newStRoot *big.Int, newExitRoot *big.Int, encodedL1CoordinatorTx []byte, l1L2TxsData []byte, feeIdxCoordinator []byte, verifierIdx uint8, l1Batch bool, proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "forgeBatch", newLastIdx, newStRoot, newExitRoot, encodedL1CoordinatorTx, l1L2TxsData, feeIdxCoordinator, verifierIdx, l1Batch, proofA, proofB, proofC)
}

// ForgeBatch is a paid mutator transaction binding the contract method 0x6e7e1365.
//
// Solidity: function forgeBatch(uint48 newLastIdx, uint256 newStRoot, uint256 newExitRoot, bytes encodedL1CoordinatorTx, bytes l1L2TxsData, bytes feeIdxCoordinator, uint8 verifierIdx, bool l1Batch, uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC) returns()
func (_Hermez *HermezSession) ForgeBatch(newLastIdx *big.Int, newStRoot *big.Int, newExitRoot *big.Int, encodedL1CoordinatorTx []byte, l1L2TxsData []byte, feeIdxCoordinator []byte, verifierIdx uint8, l1Batch bool, proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.ForgeBatch(&_Hermez.TransactOpts, newLastIdx, newStRoot, newExitRoot, encodedL1CoordinatorTx, l1L2TxsData, feeIdxCoordinator, verifierIdx, l1Batch, proofA, proofB, proofC)
}

// ForgeBatch is a paid mutator transaction binding the contract method 0x6e7e1365.
//
// Solidity: function forgeBatch(uint48 newLastIdx, uint256 newStRoot, uint256 newExitRoot, bytes encodedL1CoordinatorTx, bytes l1L2TxsData, bytes feeIdxCoordinator, uint8 verifierIdx, bool l1Batch, uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC) returns()
func (_Hermez *HermezTransactorSession) ForgeBatch(newLastIdx *big.Int, newStRoot *big.Int, newExitRoot *big.Int, encodedL1CoordinatorTx []byte, l1L2TxsData []byte, feeIdxCoordinator []byte, verifierIdx uint8, l1Batch bool, proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.ForgeBatch(&_Hermez.TransactOpts, newLastIdx, newStRoot, newExitRoot, encodedL1CoordinatorTx, l1L2TxsData, feeIdxCoordinator, verifierIdx, l1Batch, proofA, proofB, proofC)
}

// InitializeHermez is a paid mutator transaction binding the contract method 0x599897e3.
//
// Solidity: function initializeHermez(address[] _verifiers, uint256[] _verifiersParams, address _withdrawVerifier, address _hermezAuctionContract, address _tokenHEZ, uint8 _forgeL1L2BatchTimeout, uint256 _feeAddToken, address _poseidon2Elements, address _poseidon3Elements, address _poseidon4Elements, address _hermezGovernanceAddress, uint64 _withdrawalDelay, address _withdrawDelayerContract) returns()
func (_Hermez *HermezTransactor) InitializeHermez(opts *bind.TransactOpts, _verifiers []common.Address, _verifiersParams []*big.Int, _withdrawVerifier common.Address, _hermezAuctionContract common.Address, _tokenHEZ common.Address, _forgeL1L2BatchTimeout uint8, _feeAddToken *big.Int, _poseidon2Elements common.Address, _poseidon3Elements common.Address, _poseidon4Elements common.Address, _hermezGovernanceAddress common.Address, _withdrawalDelay uint64, _withdrawDelayerContract common.Address) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "initializeHermez", _verifiers, _verifiersParams, _withdrawVerifier, _hermezAuctionContract, _tokenHEZ, _forgeL1L2BatchTimeout, _feeAddToken, _poseidon2Elements, _poseidon3Elements, _poseidon4Elements, _hermezGovernanceAddress, _withdrawalDelay, _withdrawDelayerContract)
}

// InitializeHermez is a paid mutator transaction binding the contract method 0x599897e3.
//
// Solidity: function initializeHermez(address[] _verifiers, uint256[] _verifiersParams, address _withdrawVerifier, address _hermezAuctionContract, address _tokenHEZ, uint8 _forgeL1L2BatchTimeout, uint256 _feeAddToken, address _poseidon2Elements, address _poseidon3Elements, address _poseidon4Elements, address _hermezGovernanceAddress, uint64 _withdrawalDelay, address _withdrawDelayerContract) returns()
func (_Hermez *HermezSession) InitializeHermez(_verifiers []common.Address, _verifiersParams []*big.Int, _withdrawVerifier common.Address, _hermezAuctionContract common.Address, _tokenHEZ common.Address, _forgeL1L2BatchTimeout uint8, _feeAddToken *big.Int, _poseidon2Elements common.Address, _poseidon3Elements common.Address, _poseidon4Elements common.Address, _hermezGovernanceAddress common.Address, _withdrawalDelay uint64, _withdrawDelayerContract common.Address) (*types.Transaction, error) {
	return _Hermez.Contract.InitializeHermez(&_Hermez.TransactOpts, _verifiers, _verifiersParams, _withdrawVerifier, _hermezAuctionContract, _tokenHEZ, _forgeL1L2BatchTimeout, _feeAddToken, _poseidon2Elements, _poseidon3Elements, _poseidon4Elements, _hermezGovernanceAddress, _withdrawalDelay, _withdrawDelayerContract)
}

// InitializeHermez is a paid mutator transaction binding the contract method 0x599897e3.
//
// Solidity: function initializeHermez(address[] _verifiers, uint256[] _verifiersParams, address _withdrawVerifier, address _hermezAuctionContract, address _tokenHEZ, uint8 _forgeL1L2BatchTimeout, uint256 _feeAddToken, address _poseidon2Elements, address _poseidon3Elements, address _poseidon4Elements, address _hermezGovernanceAddress, uint64 _withdrawalDelay, address _withdrawDelayerContract) returns()
func (_Hermez *HermezTransactorSession) InitializeHermez(_verifiers []common.Address, _verifiersParams []*big.Int, _withdrawVerifier common.Address, _hermezAuctionContract common.Address, _tokenHEZ common.Address, _forgeL1L2BatchTimeout uint8, _feeAddToken *big.Int, _poseidon2Elements common.Address, _poseidon3Elements common.Address, _poseidon4Elements common.Address, _hermezGovernanceAddress common.Address, _withdrawalDelay uint64, _withdrawDelayerContract common.Address) (*types.Transaction, error) {
	return _Hermez.Contract.InitializeHermez(&_Hermez.TransactOpts, _verifiers, _verifiersParams, _withdrawVerifier, _hermezAuctionContract, _tokenHEZ, _forgeL1L2BatchTimeout, _feeAddToken, _poseidon2Elements, _poseidon3Elements, _poseidon4Elements, _hermezGovernanceAddress, _withdrawalDelay, _withdrawDelayerContract)
}

// SafeMode is a paid mutator transaction binding the contract method 0xabe3219c.
//
// Solidity: function safeMode() returns()
func (_Hermez *HermezTransactor) SafeMode(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "safeMode")
}

// SafeMode is a paid mutator transaction binding the contract method 0xabe3219c.
//
// Solidity: function safeMode() returns()
func (_Hermez *HermezSession) SafeMode() (*types.Transaction, error) {
	return _Hermez.Contract.SafeMode(&_Hermez.TransactOpts)
}

// SafeMode is a paid mutator transaction binding the contract method 0xabe3219c.
//
// Solidity: function safeMode() returns()
func (_Hermez *HermezTransactorSession) SafeMode() (*types.Transaction, error) {
	return _Hermez.Contract.SafeMode(&_Hermez.TransactOpts)
}

// UpdateBucketsParameters is a paid mutator transaction binding the contract method 0xac300ec9.
//
// Solidity: function updateBucketsParameters(uint256[] newBuckets) returns()
func (_Hermez *HermezTransactor) UpdateBucketsParameters(opts *bind.TransactOpts, newBuckets []*big.Int) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateBucketsParameters", newBuckets)
}

// UpdateBucketsParameters is a paid mutator transaction binding the contract method 0xac300ec9.
//
// Solidity: function updateBucketsParameters(uint256[] newBuckets) returns()
func (_Hermez *HermezSession) UpdateBucketsParameters(newBuckets []*big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateBucketsParameters(&_Hermez.TransactOpts, newBuckets)
}

// UpdateBucketsParameters is a paid mutator transaction binding the contract method 0xac300ec9.
//
// Solidity: function updateBucketsParameters(uint256[] newBuckets) returns()
func (_Hermez *HermezTransactorSession) UpdateBucketsParameters(newBuckets []*big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateBucketsParameters(&_Hermez.TransactOpts, newBuckets)
}

// UpdateFeeAddToken is a paid mutator transaction binding the contract method 0x314e5eda.
//
// Solidity: function updateFeeAddToken(uint256 newFeeAddToken) returns()
func (_Hermez *HermezTransactor) UpdateFeeAddToken(opts *bind.TransactOpts, newFeeAddToken *big.Int) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateFeeAddToken", newFeeAddToken)
}

// UpdateFeeAddToken is a paid mutator transaction binding the contract method 0x314e5eda.
//
// Solidity: function updateFeeAddToken(uint256 newFeeAddToken) returns()
func (_Hermez *HermezSession) UpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateFeeAddToken(&_Hermez.TransactOpts, newFeeAddToken)
}

// UpdateFeeAddToken is a paid mutator transaction binding the contract method 0x314e5eda.
//
// Solidity: function updateFeeAddToken(uint256 newFeeAddToken) returns()
func (_Hermez *HermezTransactorSession) UpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateFeeAddToken(&_Hermez.TransactOpts, newFeeAddToken)
}

// UpdateForgeL1L2BatchTimeout is a paid mutator transaction binding the contract method 0xcbd7b5fb.
//
// Solidity: function updateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout) returns()
func (_Hermez *HermezTransactor) UpdateForgeL1L2BatchTimeout(opts *bind.TransactOpts, newForgeL1L2BatchTimeout uint8) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateForgeL1L2BatchTimeout", newForgeL1L2BatchTimeout)
}

// UpdateForgeL1L2BatchTimeout is a paid mutator transaction binding the contract method 0xcbd7b5fb.
//
// Solidity: function updateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout) returns()
func (_Hermez *HermezSession) UpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout uint8) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateForgeL1L2BatchTimeout(&_Hermez.TransactOpts, newForgeL1L2BatchTimeout)
}

// UpdateForgeL1L2BatchTimeout is a paid mutator transaction binding the contract method 0xcbd7b5fb.
//
// Solidity: function updateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout) returns()
func (_Hermez *HermezTransactorSession) UpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout uint8) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateForgeL1L2BatchTimeout(&_Hermez.TransactOpts, newForgeL1L2BatchTimeout)
}

// UpdateTokenExchange is a paid mutator transaction binding the contract method 0x1a748c2d.
//
// Solidity: function updateTokenExchange(address[] addressArray, uint64[] valueArray) returns()
func (_Hermez *HermezTransactor) UpdateTokenExchange(opts *bind.TransactOpts, addressArray []common.Address, valueArray []uint64) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateTokenExchange", addressArray, valueArray)
}

// UpdateTokenExchange is a paid mutator transaction binding the contract method 0x1a748c2d.
//
// Solidity: function updateTokenExchange(address[] addressArray, uint64[] valueArray) returns()
func (_Hermez *HermezSession) UpdateTokenExchange(addressArray []common.Address, valueArray []uint64) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateTokenExchange(&_Hermez.TransactOpts, addressArray, valueArray)
}

// UpdateTokenExchange is a paid mutator transaction binding the contract method 0x1a748c2d.
//
// Solidity: function updateTokenExchange(address[] addressArray, uint64[] valueArray) returns()
func (_Hermez *HermezTransactorSession) UpdateTokenExchange(addressArray []common.Address, valueArray []uint64) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateTokenExchange(&_Hermez.TransactOpts, addressArray, valueArray)
}

// UpdateVerifiers is a paid mutator transaction binding the contract method 0x960207c0.
//
// Solidity: function updateVerifiers() returns()
func (_Hermez *HermezTransactor) UpdateVerifiers(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateVerifiers")
}

// UpdateVerifiers is a paid mutator transaction binding the contract method 0x960207c0.
//
// Solidity: function updateVerifiers() returns()
func (_Hermez *HermezSession) UpdateVerifiers() (*types.Transaction, error) {
	return _Hermez.Contract.UpdateVerifiers(&_Hermez.TransactOpts)
}

// UpdateVerifiers is a paid mutator transaction binding the contract method 0x960207c0.
//
// Solidity: function updateVerifiers() returns()
func (_Hermez *HermezTransactorSession) UpdateVerifiers() (*types.Transaction, error) {
	return _Hermez.Contract.UpdateVerifiers(&_Hermez.TransactOpts)
}

// UpdateWithdrawalDelay is a paid mutator transaction binding the contract method 0xef4a5c4a.
//
// Solidity: function updateWithdrawalDelay(uint64 newWithdrawalDelay) returns()
func (_Hermez *HermezTransactor) UpdateWithdrawalDelay(opts *bind.TransactOpts, newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "updateWithdrawalDelay", newWithdrawalDelay)
}

// UpdateWithdrawalDelay is a paid mutator transaction binding the contract method 0xef4a5c4a.
//
// Solidity: function updateWithdrawalDelay(uint64 newWithdrawalDelay) returns()
func (_Hermez *HermezSession) UpdateWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateWithdrawalDelay(&_Hermez.TransactOpts, newWithdrawalDelay)
}

// UpdateWithdrawalDelay is a paid mutator transaction binding the contract method 0xef4a5c4a.
//
// Solidity: function updateWithdrawalDelay(uint64 newWithdrawalDelay) returns()
func (_Hermez *HermezTransactorSession) UpdateWithdrawalDelay(newWithdrawalDelay uint64) (*types.Transaction, error) {
	return _Hermez.Contract.UpdateWithdrawalDelay(&_Hermez.TransactOpts, newWithdrawalDelay)
}

// WithdrawCircuit is a paid mutator transaction binding the contract method 0x9ce2ad42.
//
// Solidity: function withdrawCircuit(uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC, uint32 tokenID, uint192 amount, uint32 numExitRoot, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezTransactor) WithdrawCircuit(opts *bind.TransactOpts, proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int, tokenID uint32, amount *big.Int, numExitRoot uint32, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "withdrawCircuit", proofA, proofB, proofC, tokenID, amount, numExitRoot, idx, instantWithdraw)
}

// WithdrawCircuit is a paid mutator transaction binding the contract method 0x9ce2ad42.
//
// Solidity: function withdrawCircuit(uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC, uint32 tokenID, uint192 amount, uint32 numExitRoot, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezSession) WithdrawCircuit(proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int, tokenID uint32, amount *big.Int, numExitRoot uint32, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.Contract.WithdrawCircuit(&_Hermez.TransactOpts, proofA, proofB, proofC, tokenID, amount, numExitRoot, idx, instantWithdraw)
}

// WithdrawCircuit is a paid mutator transaction binding the contract method 0x9ce2ad42.
//
// Solidity: function withdrawCircuit(uint256[2] proofA, uint256[2][2] proofB, uint256[2] proofC, uint32 tokenID, uint192 amount, uint32 numExitRoot, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezTransactorSession) WithdrawCircuit(proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int, tokenID uint32, amount *big.Int, numExitRoot uint32, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.Contract.WithdrawCircuit(&_Hermez.TransactOpts, proofA, proofB, proofC, tokenID, amount, numExitRoot, idx, instantWithdraw)
}

// WithdrawMerkleProof is a paid mutator transaction binding the contract method 0xd9d4ca44.
//
// Solidity: function withdrawMerkleProof(uint32 tokenID, uint192 amount, uint256 babyPubKey, uint32 numExitRoot, uint256[] siblings, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezTransactor) WithdrawMerkleProof(opts *bind.TransactOpts, tokenID uint32, amount *big.Int, babyPubKey *big.Int, numExitRoot uint32, siblings []*big.Int, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.contract.Transact(opts, "withdrawMerkleProof", tokenID, amount, babyPubKey, numExitRoot, siblings, idx, instantWithdraw)
}

// WithdrawMerkleProof is a paid mutator transaction binding the contract method 0xd9d4ca44.
//
// Solidity: function withdrawMerkleProof(uint32 tokenID, uint192 amount, uint256 babyPubKey, uint32 numExitRoot, uint256[] siblings, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezSession) WithdrawMerkleProof(tokenID uint32, amount *big.Int, babyPubKey *big.Int, numExitRoot uint32, siblings []*big.Int, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.Contract.WithdrawMerkleProof(&_Hermez.TransactOpts, tokenID, amount, babyPubKey, numExitRoot, siblings, idx, instantWithdraw)
}

// WithdrawMerkleProof is a paid mutator transaction binding the contract method 0xd9d4ca44.
//
// Solidity: function withdrawMerkleProof(uint32 tokenID, uint192 amount, uint256 babyPubKey, uint32 numExitRoot, uint256[] siblings, uint48 idx, bool instantWithdraw) returns()
func (_Hermez *HermezTransactorSession) WithdrawMerkleProof(tokenID uint32, amount *big.Int, babyPubKey *big.Int, numExitRoot uint32, siblings []*big.Int, idx *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	return _Hermez.Contract.WithdrawMerkleProof(&_Hermez.TransactOpts, tokenID, amount, babyPubKey, numExitRoot, siblings, idx, instantWithdraw)
}

// HermezAddTokenIterator is returned from FilterAddToken and is used to iterate over the raw logs and unpacked data for AddToken events raised by the Hermez contract.
type HermezAddTokenIterator struct {
	Event *HermezAddToken // Event containing the contract specifics and raw log

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
func (it *HermezAddTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezAddToken)
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
		it.Event = new(HermezAddToken)
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
func (it *HermezAddTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezAddTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezAddToken represents a AddToken event raised by the Hermez contract.
type HermezAddToken struct {
	TokenAddress common.Address
	TokenID      uint32
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAddToken is a free log retrieval operation binding the contract event 0xcb73d161edb7cd4fb1d92fedfd2555384fd997fd44ab507656f8c81e15747dde.
//
// Solidity: event AddToken(address indexed tokenAddress, uint32 tokenID)
func (_Hermez *HermezFilterer) FilterAddToken(opts *bind.FilterOpts, tokenAddress []common.Address) (*HermezAddTokenIterator, error) {

	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "AddToken", tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return &HermezAddTokenIterator{contract: _Hermez.contract, event: "AddToken", logs: logs, sub: sub}, nil
}

// WatchAddToken is a free log subscription operation binding the contract event 0xcb73d161edb7cd4fb1d92fedfd2555384fd997fd44ab507656f8c81e15747dde.
//
// Solidity: event AddToken(address indexed tokenAddress, uint32 tokenID)
func (_Hermez *HermezFilterer) WatchAddToken(opts *bind.WatchOpts, sink chan<- *HermezAddToken, tokenAddress []common.Address) (event.Subscription, error) {

	var tokenAddressRule []interface{}
	for _, tokenAddressItem := range tokenAddress {
		tokenAddressRule = append(tokenAddressRule, tokenAddressItem)
	}

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "AddToken", tokenAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezAddToken)
				if err := _Hermez.contract.UnpackLog(event, "AddToken", log); err != nil {
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

// ParseAddToken is a log parse operation binding the contract event 0xcb73d161edb7cd4fb1d92fedfd2555384fd997fd44ab507656f8c81e15747dde.
//
// Solidity: event AddToken(address indexed tokenAddress, uint32 tokenID)
func (_Hermez *HermezFilterer) ParseAddToken(log types.Log) (*HermezAddToken, error) {
	event := new(HermezAddToken)
	if err := _Hermez.contract.UnpackLog(event, "AddToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezForgeBatchIterator is returned from FilterForgeBatch and is used to iterate over the raw logs and unpacked data for ForgeBatch events raised by the Hermez contract.
type HermezForgeBatchIterator struct {
	Event *HermezForgeBatch // Event containing the contract specifics and raw log

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
func (it *HermezForgeBatchIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezForgeBatch)
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
		it.Event = new(HermezForgeBatch)
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
func (it *HermezForgeBatchIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezForgeBatchIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezForgeBatch represents a ForgeBatch event raised by the Hermez contract.
type HermezForgeBatch struct {
	BatchNum     uint32
	L1UserTxsLen uint16
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterForgeBatch is a free log retrieval operation binding the contract event 0xe00040c8a3b0bf905636c26924e90520eafc5003324138236fddee2d34588618.
//
// Solidity: event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen)
func (_Hermez *HermezFilterer) FilterForgeBatch(opts *bind.FilterOpts, batchNum []uint32) (*HermezForgeBatchIterator, error) {

	var batchNumRule []interface{}
	for _, batchNumItem := range batchNum {
		batchNumRule = append(batchNumRule, batchNumItem)
	}

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "ForgeBatch", batchNumRule)
	if err != nil {
		return nil, err
	}
	return &HermezForgeBatchIterator{contract: _Hermez.contract, event: "ForgeBatch", logs: logs, sub: sub}, nil
}

// WatchForgeBatch is a free log subscription operation binding the contract event 0xe00040c8a3b0bf905636c26924e90520eafc5003324138236fddee2d34588618.
//
// Solidity: event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen)
func (_Hermez *HermezFilterer) WatchForgeBatch(opts *bind.WatchOpts, sink chan<- *HermezForgeBatch, batchNum []uint32) (event.Subscription, error) {

	var batchNumRule []interface{}
	for _, batchNumItem := range batchNum {
		batchNumRule = append(batchNumRule, batchNumItem)
	}

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "ForgeBatch", batchNumRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezForgeBatch)
				if err := _Hermez.contract.UnpackLog(event, "ForgeBatch", log); err != nil {
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

// ParseForgeBatch is a log parse operation binding the contract event 0xe00040c8a3b0bf905636c26924e90520eafc5003324138236fddee2d34588618.
//
// Solidity: event ForgeBatch(uint32 indexed batchNum, uint16 l1UserTxsLen)
func (_Hermez *HermezFilterer) ParseForgeBatch(log types.Log) (*HermezForgeBatch, error) {
	event := new(HermezForgeBatch)
	if err := _Hermez.contract.UnpackLog(event, "ForgeBatch", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezInitializeHermezEventIterator is returned from FilterInitializeHermezEvent and is used to iterate over the raw logs and unpacked data for InitializeHermezEvent events raised by the Hermez contract.
type HermezInitializeHermezEventIterator struct {
	Event *HermezInitializeHermezEvent // Event containing the contract specifics and raw log

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
func (it *HermezInitializeHermezEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezInitializeHermezEvent)
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
		it.Event = new(HermezInitializeHermezEvent)
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
func (it *HermezInitializeHermezEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezInitializeHermezEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezInitializeHermezEvent represents a InitializeHermezEvent event raised by the Hermez contract.
type HermezInitializeHermezEvent struct {
	ForgeL1L2BatchTimeout uint8
	FeeAddToken           *big.Int
	WithdrawalDelay       uint64
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterInitializeHermezEvent is a free log retrieval operation binding the contract event 0xc5272ad4c8d9f2e9af2f9555c11ead049be22b6e45c16975adc82371b7cd1040.
//
// Solidity: event InitializeHermezEvent(uint8 forgeL1L2BatchTimeout, uint256 feeAddToken, uint64 withdrawalDelay)
func (_Hermez *HermezFilterer) FilterInitializeHermezEvent(opts *bind.FilterOpts) (*HermezInitializeHermezEventIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "InitializeHermezEvent")
	if err != nil {
		return nil, err
	}
	return &HermezInitializeHermezEventIterator{contract: _Hermez.contract, event: "InitializeHermezEvent", logs: logs, sub: sub}, nil
}

// WatchInitializeHermezEvent is a free log subscription operation binding the contract event 0xc5272ad4c8d9f2e9af2f9555c11ead049be22b6e45c16975adc82371b7cd1040.
//
// Solidity: event InitializeHermezEvent(uint8 forgeL1L2BatchTimeout, uint256 feeAddToken, uint64 withdrawalDelay)
func (_Hermez *HermezFilterer) WatchInitializeHermezEvent(opts *bind.WatchOpts, sink chan<- *HermezInitializeHermezEvent) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "InitializeHermezEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezInitializeHermezEvent)
				if err := _Hermez.contract.UnpackLog(event, "InitializeHermezEvent", log); err != nil {
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

// ParseInitializeHermezEvent is a log parse operation binding the contract event 0xc5272ad4c8d9f2e9af2f9555c11ead049be22b6e45c16975adc82371b7cd1040.
//
// Solidity: event InitializeHermezEvent(uint8 forgeL1L2BatchTimeout, uint256 feeAddToken, uint64 withdrawalDelay)
func (_Hermez *HermezFilterer) ParseInitializeHermezEvent(log types.Log) (*HermezInitializeHermezEvent, error) {
	event := new(HermezInitializeHermezEvent)
	if err := _Hermez.contract.UnpackLog(event, "InitializeHermezEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezL1UserTxEventIterator is returned from FilterL1UserTxEvent and is used to iterate over the raw logs and unpacked data for L1UserTxEvent events raised by the Hermez contract.
type HermezL1UserTxEventIterator struct {
	Event *HermezL1UserTxEvent // Event containing the contract specifics and raw log

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
func (it *HermezL1UserTxEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezL1UserTxEvent)
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
		it.Event = new(HermezL1UserTxEvent)
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
func (it *HermezL1UserTxEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezL1UserTxEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezL1UserTxEvent represents a L1UserTxEvent event raised by the Hermez contract.
type HermezL1UserTxEvent struct {
	QueueIndex uint32
	Position   uint8
	L1UserTx   []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterL1UserTxEvent is a free log retrieval operation binding the contract event 0xdd5c7c5ea02d3c5d1621513faa6de53d474ee6f111eda6352a63e3dfe8c40119.
//
// Solidity: event L1UserTxEvent(uint32 indexed queueIndex, uint8 indexed position, bytes l1UserTx)
func (_Hermez *HermezFilterer) FilterL1UserTxEvent(opts *bind.FilterOpts, queueIndex []uint32, position []uint8) (*HermezL1UserTxEventIterator, error) {

	var queueIndexRule []interface{}
	for _, queueIndexItem := range queueIndex {
		queueIndexRule = append(queueIndexRule, queueIndexItem)
	}
	var positionRule []interface{}
	for _, positionItem := range position {
		positionRule = append(positionRule, positionItem)
	}

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "L1UserTxEvent", queueIndexRule, positionRule)
	if err != nil {
		return nil, err
	}
	return &HermezL1UserTxEventIterator{contract: _Hermez.contract, event: "L1UserTxEvent", logs: logs, sub: sub}, nil
}

// WatchL1UserTxEvent is a free log subscription operation binding the contract event 0xdd5c7c5ea02d3c5d1621513faa6de53d474ee6f111eda6352a63e3dfe8c40119.
//
// Solidity: event L1UserTxEvent(uint32 indexed queueIndex, uint8 indexed position, bytes l1UserTx)
func (_Hermez *HermezFilterer) WatchL1UserTxEvent(opts *bind.WatchOpts, sink chan<- *HermezL1UserTxEvent, queueIndex []uint32, position []uint8) (event.Subscription, error) {

	var queueIndexRule []interface{}
	for _, queueIndexItem := range queueIndex {
		queueIndexRule = append(queueIndexRule, queueIndexItem)
	}
	var positionRule []interface{}
	for _, positionItem := range position {
		positionRule = append(positionRule, positionItem)
	}

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "L1UserTxEvent", queueIndexRule, positionRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezL1UserTxEvent)
				if err := _Hermez.contract.UnpackLog(event, "L1UserTxEvent", log); err != nil {
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

// ParseL1UserTxEvent is a log parse operation binding the contract event 0xdd5c7c5ea02d3c5d1621513faa6de53d474ee6f111eda6352a63e3dfe8c40119.
//
// Solidity: event L1UserTxEvent(uint32 indexed queueIndex, uint8 indexed position, bytes l1UserTx)
func (_Hermez *HermezFilterer) ParseL1UserTxEvent(log types.Log) (*HermezL1UserTxEvent, error) {
	event := new(HermezL1UserTxEvent)
	if err := _Hermez.contract.UnpackLog(event, "L1UserTxEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezSafeModeIterator is returned from FilterSafeMode and is used to iterate over the raw logs and unpacked data for SafeMode events raised by the Hermez contract.
type HermezSafeModeIterator struct {
	Event *HermezSafeMode // Event containing the contract specifics and raw log

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
func (it *HermezSafeModeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezSafeMode)
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
		it.Event = new(HermezSafeMode)
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
func (it *HermezSafeModeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezSafeModeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezSafeMode represents a SafeMode event raised by the Hermez contract.
type HermezSafeMode struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterSafeMode is a free log retrieval operation binding the contract event 0x0410e6ef2bd89ecf5b2dc2f62157f9863e09e89cb7c7f1abb7d4ec43a6019d1e.
//
// Solidity: event SafeMode()
func (_Hermez *HermezFilterer) FilterSafeMode(opts *bind.FilterOpts) (*HermezSafeModeIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "SafeMode")
	if err != nil {
		return nil, err
	}
	return &HermezSafeModeIterator{contract: _Hermez.contract, event: "SafeMode", logs: logs, sub: sub}, nil
}

// WatchSafeMode is a free log subscription operation binding the contract event 0x0410e6ef2bd89ecf5b2dc2f62157f9863e09e89cb7c7f1abb7d4ec43a6019d1e.
//
// Solidity: event SafeMode()
func (_Hermez *HermezFilterer) WatchSafeMode(opts *bind.WatchOpts, sink chan<- *HermezSafeMode) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "SafeMode")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezSafeMode)
				if err := _Hermez.contract.UnpackLog(event, "SafeMode", log); err != nil {
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

// ParseSafeMode is a log parse operation binding the contract event 0x0410e6ef2bd89ecf5b2dc2f62157f9863e09e89cb7c7f1abb7d4ec43a6019d1e.
//
// Solidity: event SafeMode()
func (_Hermez *HermezFilterer) ParseSafeMode(log types.Log) (*HermezSafeMode, error) {
	event := new(HermezSafeMode)
	if err := _Hermez.contract.UnpackLog(event, "SafeMode", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateBucketWithdrawIterator is returned from FilterUpdateBucketWithdraw and is used to iterate over the raw logs and unpacked data for UpdateBucketWithdraw events raised by the Hermez contract.
type HermezUpdateBucketWithdrawIterator struct {
	Event *HermezUpdateBucketWithdraw // Event containing the contract specifics and raw log

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
func (it *HermezUpdateBucketWithdrawIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateBucketWithdraw)
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
		it.Event = new(HermezUpdateBucketWithdraw)
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
func (it *HermezUpdateBucketWithdrawIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateBucketWithdrawIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateBucketWithdraw represents a UpdateBucketWithdraw event raised by the Hermez contract.
type HermezUpdateBucketWithdraw struct {
	NumBucket   uint8
	BlockStamp  *big.Int
	Withdrawals *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUpdateBucketWithdraw is a free log retrieval operation binding the contract event 0xa35fe9a9e21cdbbc4774aa8a56e7b97ea9c06afc09ffb06af593d26951e350aa.
//
// Solidity: event UpdateBucketWithdraw(uint8 indexed numBucket, uint256 indexed blockStamp, uint256 withdrawals)
func (_Hermez *HermezFilterer) FilterUpdateBucketWithdraw(opts *bind.FilterOpts, numBucket []uint8, blockStamp []*big.Int) (*HermezUpdateBucketWithdrawIterator, error) {

	var numBucketRule []interface{}
	for _, numBucketItem := range numBucket {
		numBucketRule = append(numBucketRule, numBucketItem)
	}
	var blockStampRule []interface{}
	for _, blockStampItem := range blockStamp {
		blockStampRule = append(blockStampRule, blockStampItem)
	}

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateBucketWithdraw", numBucketRule, blockStampRule)
	if err != nil {
		return nil, err
	}
	return &HermezUpdateBucketWithdrawIterator{contract: _Hermez.contract, event: "UpdateBucketWithdraw", logs: logs, sub: sub}, nil
}

// WatchUpdateBucketWithdraw is a free log subscription operation binding the contract event 0xa35fe9a9e21cdbbc4774aa8a56e7b97ea9c06afc09ffb06af593d26951e350aa.
//
// Solidity: event UpdateBucketWithdraw(uint8 indexed numBucket, uint256 indexed blockStamp, uint256 withdrawals)
func (_Hermez *HermezFilterer) WatchUpdateBucketWithdraw(opts *bind.WatchOpts, sink chan<- *HermezUpdateBucketWithdraw, numBucket []uint8, blockStamp []*big.Int) (event.Subscription, error) {

	var numBucketRule []interface{}
	for _, numBucketItem := range numBucket {
		numBucketRule = append(numBucketRule, numBucketItem)
	}
	var blockStampRule []interface{}
	for _, blockStampItem := range blockStamp {
		blockStampRule = append(blockStampRule, blockStampItem)
	}

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateBucketWithdraw", numBucketRule, blockStampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateBucketWithdraw)
				if err := _Hermez.contract.UnpackLog(event, "UpdateBucketWithdraw", log); err != nil {
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

// ParseUpdateBucketWithdraw is a log parse operation binding the contract event 0xa35fe9a9e21cdbbc4774aa8a56e7b97ea9c06afc09ffb06af593d26951e350aa.
//
// Solidity: event UpdateBucketWithdraw(uint8 indexed numBucket, uint256 indexed blockStamp, uint256 withdrawals)
func (_Hermez *HermezFilterer) ParseUpdateBucketWithdraw(log types.Log) (*HermezUpdateBucketWithdraw, error) {
	event := new(HermezUpdateBucketWithdraw)
	if err := _Hermez.contract.UnpackLog(event, "UpdateBucketWithdraw", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateBucketsParametersIterator is returned from FilterUpdateBucketsParameters and is used to iterate over the raw logs and unpacked data for UpdateBucketsParameters events raised by the Hermez contract.
type HermezUpdateBucketsParametersIterator struct {
	Event *HermezUpdateBucketsParameters // Event containing the contract specifics and raw log

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
func (it *HermezUpdateBucketsParametersIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateBucketsParameters)
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
		it.Event = new(HermezUpdateBucketsParameters)
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
func (it *HermezUpdateBucketsParametersIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateBucketsParametersIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateBucketsParameters represents a UpdateBucketsParameters event raised by the Hermez contract.
type HermezUpdateBucketsParameters struct {
	ArrayBuckets []*big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterUpdateBucketsParameters is a free log retrieval operation binding the contract event 0xd4904145d7eae889c5493798579680417459783db0fa67398bea50e56859075f.
//
// Solidity: event UpdateBucketsParameters(uint256[] arrayBuckets)
func (_Hermez *HermezFilterer) FilterUpdateBucketsParameters(opts *bind.FilterOpts) (*HermezUpdateBucketsParametersIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateBucketsParameters")
	if err != nil {
		return nil, err
	}
	return &HermezUpdateBucketsParametersIterator{contract: _Hermez.contract, event: "UpdateBucketsParameters", logs: logs, sub: sub}, nil
}

// WatchUpdateBucketsParameters is a free log subscription operation binding the contract event 0xd4904145d7eae889c5493798579680417459783db0fa67398bea50e56859075f.
//
// Solidity: event UpdateBucketsParameters(uint256[] arrayBuckets)
func (_Hermez *HermezFilterer) WatchUpdateBucketsParameters(opts *bind.WatchOpts, sink chan<- *HermezUpdateBucketsParameters) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateBucketsParameters")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateBucketsParameters)
				if err := _Hermez.contract.UnpackLog(event, "UpdateBucketsParameters", log); err != nil {
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

// ParseUpdateBucketsParameters is a log parse operation binding the contract event 0xd4904145d7eae889c5493798579680417459783db0fa67398bea50e56859075f.
//
// Solidity: event UpdateBucketsParameters(uint256[] arrayBuckets)
func (_Hermez *HermezFilterer) ParseUpdateBucketsParameters(log types.Log) (*HermezUpdateBucketsParameters, error) {
	event := new(HermezUpdateBucketsParameters)
	if err := _Hermez.contract.UnpackLog(event, "UpdateBucketsParameters", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateFeeAddTokenIterator is returned from FilterUpdateFeeAddToken and is used to iterate over the raw logs and unpacked data for UpdateFeeAddToken events raised by the Hermez contract.
type HermezUpdateFeeAddTokenIterator struct {
	Event *HermezUpdateFeeAddToken // Event containing the contract specifics and raw log

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
func (it *HermezUpdateFeeAddTokenIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateFeeAddToken)
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
		it.Event = new(HermezUpdateFeeAddToken)
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
func (it *HermezUpdateFeeAddTokenIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateFeeAddTokenIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateFeeAddToken represents a UpdateFeeAddToken event raised by the Hermez contract.
type HermezUpdateFeeAddToken struct {
	NewFeeAddToken *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpdateFeeAddToken is a free log retrieval operation binding the contract event 0xd1c873cd16013f0dc5f37992c0d12794389698512895ec036a568e393b46e3c1.
//
// Solidity: event UpdateFeeAddToken(uint256 newFeeAddToken)
func (_Hermez *HermezFilterer) FilterUpdateFeeAddToken(opts *bind.FilterOpts) (*HermezUpdateFeeAddTokenIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateFeeAddToken")
	if err != nil {
		return nil, err
	}
	return &HermezUpdateFeeAddTokenIterator{contract: _Hermez.contract, event: "UpdateFeeAddToken", logs: logs, sub: sub}, nil
}

// WatchUpdateFeeAddToken is a free log subscription operation binding the contract event 0xd1c873cd16013f0dc5f37992c0d12794389698512895ec036a568e393b46e3c1.
//
// Solidity: event UpdateFeeAddToken(uint256 newFeeAddToken)
func (_Hermez *HermezFilterer) WatchUpdateFeeAddToken(opts *bind.WatchOpts, sink chan<- *HermezUpdateFeeAddToken) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateFeeAddToken")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateFeeAddToken)
				if err := _Hermez.contract.UnpackLog(event, "UpdateFeeAddToken", log); err != nil {
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

// ParseUpdateFeeAddToken is a log parse operation binding the contract event 0xd1c873cd16013f0dc5f37992c0d12794389698512895ec036a568e393b46e3c1.
//
// Solidity: event UpdateFeeAddToken(uint256 newFeeAddToken)
func (_Hermez *HermezFilterer) ParseUpdateFeeAddToken(log types.Log) (*HermezUpdateFeeAddToken, error) {
	event := new(HermezUpdateFeeAddToken)
	if err := _Hermez.contract.UnpackLog(event, "UpdateFeeAddToken", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateForgeL1L2BatchTimeoutIterator is returned from FilterUpdateForgeL1L2BatchTimeout and is used to iterate over the raw logs and unpacked data for UpdateForgeL1L2BatchTimeout events raised by the Hermez contract.
type HermezUpdateForgeL1L2BatchTimeoutIterator struct {
	Event *HermezUpdateForgeL1L2BatchTimeout // Event containing the contract specifics and raw log

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
func (it *HermezUpdateForgeL1L2BatchTimeoutIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateForgeL1L2BatchTimeout)
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
		it.Event = new(HermezUpdateForgeL1L2BatchTimeout)
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
func (it *HermezUpdateForgeL1L2BatchTimeoutIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateForgeL1L2BatchTimeoutIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateForgeL1L2BatchTimeout represents a UpdateForgeL1L2BatchTimeout event raised by the Hermez contract.
type HermezUpdateForgeL1L2BatchTimeout struct {
	NewForgeL1L2BatchTimeout uint8
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterUpdateForgeL1L2BatchTimeout is a free log retrieval operation binding the contract event 0xff6221781ac525b04585dbb55cd2ebd2a92c828ca3e42b23813a1137ac974431.
//
// Solidity: event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout)
func (_Hermez *HermezFilterer) FilterUpdateForgeL1L2BatchTimeout(opts *bind.FilterOpts) (*HermezUpdateForgeL1L2BatchTimeoutIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateForgeL1L2BatchTimeout")
	if err != nil {
		return nil, err
	}
	return &HermezUpdateForgeL1L2BatchTimeoutIterator{contract: _Hermez.contract, event: "UpdateForgeL1L2BatchTimeout", logs: logs, sub: sub}, nil
}

// WatchUpdateForgeL1L2BatchTimeout is a free log subscription operation binding the contract event 0xff6221781ac525b04585dbb55cd2ebd2a92c828ca3e42b23813a1137ac974431.
//
// Solidity: event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout)
func (_Hermez *HermezFilterer) WatchUpdateForgeL1L2BatchTimeout(opts *bind.WatchOpts, sink chan<- *HermezUpdateForgeL1L2BatchTimeout) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateForgeL1L2BatchTimeout")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateForgeL1L2BatchTimeout)
				if err := _Hermez.contract.UnpackLog(event, "UpdateForgeL1L2BatchTimeout", log); err != nil {
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

// ParseUpdateForgeL1L2BatchTimeout is a log parse operation binding the contract event 0xff6221781ac525b04585dbb55cd2ebd2a92c828ca3e42b23813a1137ac974431.
//
// Solidity: event UpdateForgeL1L2BatchTimeout(uint8 newForgeL1L2BatchTimeout)
func (_Hermez *HermezFilterer) ParseUpdateForgeL1L2BatchTimeout(log types.Log) (*HermezUpdateForgeL1L2BatchTimeout, error) {
	event := new(HermezUpdateForgeL1L2BatchTimeout)
	if err := _Hermez.contract.UnpackLog(event, "UpdateForgeL1L2BatchTimeout", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateTokenExchangeIterator is returned from FilterUpdateTokenExchange and is used to iterate over the raw logs and unpacked data for UpdateTokenExchange events raised by the Hermez contract.
type HermezUpdateTokenExchangeIterator struct {
	Event *HermezUpdateTokenExchange // Event containing the contract specifics and raw log

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
func (it *HermezUpdateTokenExchangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateTokenExchange)
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
		it.Event = new(HermezUpdateTokenExchange)
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
func (it *HermezUpdateTokenExchangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateTokenExchangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateTokenExchange represents a UpdateTokenExchange event raised by the Hermez contract.
type HermezUpdateTokenExchange struct {
	AddressArray []common.Address
	ValueArray   []uint64
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterUpdateTokenExchange is a free log retrieval operation binding the contract event 0x10ff643ebeca3e33002e61b76fa85e7e10091e30afa39295f91af9838b3033b3.
//
// Solidity: event UpdateTokenExchange(address[] addressArray, uint64[] valueArray)
func (_Hermez *HermezFilterer) FilterUpdateTokenExchange(opts *bind.FilterOpts) (*HermezUpdateTokenExchangeIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateTokenExchange")
	if err != nil {
		return nil, err
	}
	return &HermezUpdateTokenExchangeIterator{contract: _Hermez.contract, event: "UpdateTokenExchange", logs: logs, sub: sub}, nil
}

// WatchUpdateTokenExchange is a free log subscription operation binding the contract event 0x10ff643ebeca3e33002e61b76fa85e7e10091e30afa39295f91af9838b3033b3.
//
// Solidity: event UpdateTokenExchange(address[] addressArray, uint64[] valueArray)
func (_Hermez *HermezFilterer) WatchUpdateTokenExchange(opts *bind.WatchOpts, sink chan<- *HermezUpdateTokenExchange) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateTokenExchange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateTokenExchange)
				if err := _Hermez.contract.UnpackLog(event, "UpdateTokenExchange", log); err != nil {
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

// ParseUpdateTokenExchange is a log parse operation binding the contract event 0x10ff643ebeca3e33002e61b76fa85e7e10091e30afa39295f91af9838b3033b3.
//
// Solidity: event UpdateTokenExchange(address[] addressArray, uint64[] valueArray)
func (_Hermez *HermezFilterer) ParseUpdateTokenExchange(log types.Log) (*HermezUpdateTokenExchange, error) {
	event := new(HermezUpdateTokenExchange)
	if err := _Hermez.contract.UnpackLog(event, "UpdateTokenExchange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezUpdateWithdrawalDelayIterator is returned from FilterUpdateWithdrawalDelay and is used to iterate over the raw logs and unpacked data for UpdateWithdrawalDelay events raised by the Hermez contract.
type HermezUpdateWithdrawalDelayIterator struct {
	Event *HermezUpdateWithdrawalDelay // Event containing the contract specifics and raw log

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
func (it *HermezUpdateWithdrawalDelayIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezUpdateWithdrawalDelay)
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
		it.Event = new(HermezUpdateWithdrawalDelay)
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
func (it *HermezUpdateWithdrawalDelayIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezUpdateWithdrawalDelayIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezUpdateWithdrawalDelay represents a UpdateWithdrawalDelay event raised by the Hermez contract.
type HermezUpdateWithdrawalDelay struct {
	NewWithdrawalDelay uint64
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterUpdateWithdrawalDelay is a free log retrieval operation binding the contract event 0x9db302c4547a21fb20a3a794e5f63ee87eb6e4afc3325ebdadba2d1fb4a90737.
//
// Solidity: event UpdateWithdrawalDelay(uint64 newWithdrawalDelay)
func (_Hermez *HermezFilterer) FilterUpdateWithdrawalDelay(opts *bind.FilterOpts) (*HermezUpdateWithdrawalDelayIterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "UpdateWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return &HermezUpdateWithdrawalDelayIterator{contract: _Hermez.contract, event: "UpdateWithdrawalDelay", logs: logs, sub: sub}, nil
}

// WatchUpdateWithdrawalDelay is a free log subscription operation binding the contract event 0x9db302c4547a21fb20a3a794e5f63ee87eb6e4afc3325ebdadba2d1fb4a90737.
//
// Solidity: event UpdateWithdrawalDelay(uint64 newWithdrawalDelay)
func (_Hermez *HermezFilterer) WatchUpdateWithdrawalDelay(opts *bind.WatchOpts, sink chan<- *HermezUpdateWithdrawalDelay) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "UpdateWithdrawalDelay")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezUpdateWithdrawalDelay)
				if err := _Hermez.contract.UnpackLog(event, "UpdateWithdrawalDelay", log); err != nil {
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

// ParseUpdateWithdrawalDelay is a log parse operation binding the contract event 0x9db302c4547a21fb20a3a794e5f63ee87eb6e4afc3325ebdadba2d1fb4a90737.
//
// Solidity: event UpdateWithdrawalDelay(uint64 newWithdrawalDelay)
func (_Hermez *HermezFilterer) ParseUpdateWithdrawalDelay(log types.Log) (*HermezUpdateWithdrawalDelay, error) {
	event := new(HermezUpdateWithdrawalDelay)
	if err := _Hermez.contract.UnpackLog(event, "UpdateWithdrawalDelay", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezWithdrawEventIterator is returned from FilterWithdrawEvent and is used to iterate over the raw logs and unpacked data for WithdrawEvent events raised by the Hermez contract.
type HermezWithdrawEventIterator struct {
	Event *HermezWithdrawEvent // Event containing the contract specifics and raw log

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
func (it *HermezWithdrawEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezWithdrawEvent)
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
		it.Event = new(HermezWithdrawEvent)
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
func (it *HermezWithdrawEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezWithdrawEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezWithdrawEvent represents a WithdrawEvent event raised by the Hermez contract.
type HermezWithdrawEvent struct {
	Idx             *big.Int
	NumExitRoot     uint32
	InstantWithdraw bool
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterWithdrawEvent is a free log retrieval operation binding the contract event 0x69177d798b38e27bcc4e0338307e4f1490e12d1006729d0e6e9cc82a8732f415.
//
// Solidity: event WithdrawEvent(uint48 indexed idx, uint32 indexed numExitRoot, bool indexed instantWithdraw)
func (_Hermez *HermezFilterer) FilterWithdrawEvent(opts *bind.FilterOpts, idx []*big.Int, numExitRoot []uint32, instantWithdraw []bool) (*HermezWithdrawEventIterator, error) {

	var idxRule []interface{}
	for _, idxItem := range idx {
		idxRule = append(idxRule, idxItem)
	}
	var numExitRootRule []interface{}
	for _, numExitRootItem := range numExitRoot {
		numExitRootRule = append(numExitRootRule, numExitRootItem)
	}
	var instantWithdrawRule []interface{}
	for _, instantWithdrawItem := range instantWithdraw {
		instantWithdrawRule = append(instantWithdrawRule, instantWithdrawItem)
	}

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "WithdrawEvent", idxRule, numExitRootRule, instantWithdrawRule)
	if err != nil {
		return nil, err
	}
	return &HermezWithdrawEventIterator{contract: _Hermez.contract, event: "WithdrawEvent", logs: logs, sub: sub}, nil
}

// WatchWithdrawEvent is a free log subscription operation binding the contract event 0x69177d798b38e27bcc4e0338307e4f1490e12d1006729d0e6e9cc82a8732f415.
//
// Solidity: event WithdrawEvent(uint48 indexed idx, uint32 indexed numExitRoot, bool indexed instantWithdraw)
func (_Hermez *HermezFilterer) WatchWithdrawEvent(opts *bind.WatchOpts, sink chan<- *HermezWithdrawEvent, idx []*big.Int, numExitRoot []uint32, instantWithdraw []bool) (event.Subscription, error) {

	var idxRule []interface{}
	for _, idxItem := range idx {
		idxRule = append(idxRule, idxItem)
	}
	var numExitRootRule []interface{}
	for _, numExitRootItem := range numExitRoot {
		numExitRootRule = append(numExitRootRule, numExitRootItem)
	}
	var instantWithdrawRule []interface{}
	for _, instantWithdrawItem := range instantWithdraw {
		instantWithdrawRule = append(instantWithdrawRule, instantWithdrawItem)
	}

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "WithdrawEvent", idxRule, numExitRootRule, instantWithdrawRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezWithdrawEvent)
				if err := _Hermez.contract.UnpackLog(event, "WithdrawEvent", log); err != nil {
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

// ParseWithdrawEvent is a log parse operation binding the contract event 0x69177d798b38e27bcc4e0338307e4f1490e12d1006729d0e6e9cc82a8732f415.
//
// Solidity: event WithdrawEvent(uint48 indexed idx, uint32 indexed numExitRoot, bool indexed instantWithdraw)
func (_Hermez *HermezFilterer) ParseWithdrawEvent(log types.Log) (*HermezWithdrawEvent, error) {
	event := new(HermezWithdrawEvent)
	if err := _Hermez.contract.UnpackLog(event, "WithdrawEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// HermezHermezV2Iterator is returned from FilterHermezV2 and is used to iterate over the raw logs and unpacked data for HermezV2 events raised by the Hermez contract.
type HermezHermezV2Iterator struct {
	Event *HermezHermezV2 // Event containing the contract specifics and raw log

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
func (it *HermezHermezV2Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(HermezHermezV2)
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
		it.Event = new(HermezHermezV2)
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
func (it *HermezHermezV2Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *HermezHermezV2Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// HermezHermezV2 represents a HermezV2 event raised by the Hermez contract.
type HermezHermezV2 struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterHermezV2 is a free log retrieval operation binding the contract event 0xd5303fa2e7ece2a0fe77fbba1df5bb224b461198dd7bfd7fe0071f964c86c673.
//
// Solidity: event hermezV2()
func (_Hermez *HermezFilterer) FilterHermezV2(opts *bind.FilterOpts) (*HermezHermezV2Iterator, error) {

	logs, sub, err := _Hermez.contract.FilterLogs(opts, "hermezV2")
	if err != nil {
		return nil, err
	}
	return &HermezHermezV2Iterator{contract: _Hermez.contract, event: "hermezV2", logs: logs, sub: sub}, nil
}

// WatchHermezV2 is a free log subscription operation binding the contract event 0xd5303fa2e7ece2a0fe77fbba1df5bb224b461198dd7bfd7fe0071f964c86c673.
//
// Solidity: event hermezV2()
func (_Hermez *HermezFilterer) WatchHermezV2(opts *bind.WatchOpts, sink chan<- *HermezHermezV2) (event.Subscription, error) {

	logs, sub, err := _Hermez.contract.WatchLogs(opts, "hermezV2")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HermezHermezV2)
				if err := _Hermez.contract.UnpackLog(event, "hermezV2", log); err != nil {
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

// ParseHermezV2 is a log parse operation binding the contract event 0xd5303fa2e7ece2a0fe77fbba1df5bb224b461198dd7bfd7fe0071f964c86c673.
//
// Solidity: event hermezV2()
func (_Hermez *HermezFilterer) ParseHermezV2(log types.Log) (*HermezHermezV2, error) {
	event := new(HermezHermezV2)
	if err := _Hermez.contract.UnpackLog(event, "hermezV2", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
