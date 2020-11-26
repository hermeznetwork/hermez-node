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
	"github.com/ztrue/tracerr"
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
const HEZABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"initialHolder\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"authorizer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"}],\"name\":\"AuthorizationUsed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"EIP712DOMAIN_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NAME_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PERMIT_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TRANSFER_WITH_AUTHORIZATION_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"VERSION_HASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"authorizationState\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"chainId\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"nonces\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deadline\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"permit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validAfter\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"validBefore\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"nonce\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"transferWithAuthorization\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// HEZBin is the compiled bytecode used for deploying new contracts.
var HEZBin = "0x608060405234801561001057600080fd5b506040516111153803806111158339818101604052602081101561003357600080fd5b505161004a816a52b7d2dcc80cd2e4000000610050565b506101b1565b610069816000546100f260201b6109531790919060201c565b60009081556001600160a01b03831681526001602090815260409091205461009a9183906109536100f2821b17901c565b6001600160a01b03831660008181526001602090815260408083209490945583518581529351929391927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef9281900390910190a35050565b6040805180820190915260118152704d4154483a4144445f4f564552464c4f5760781b602082015281830190838210156101aa5760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b8381101561016f578181015183820152602001610157565b50505050905090810190601f16801561019c5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b5092915050565b610f55806101c06000396000f3fe608060405234801561001057600080fd5b50600436106101375760003560e01c806370a08231116100b8578063a9059cbb1161007c578063a9059cbb14610308578063c473af3314610334578063d505accf1461033c578063dd62ed3e1461038f578063e3ee160e146103bd578063e94a01021461041c57610137565b806370a08231146102a45780637ecebe00146102ca57806395d89b41146102f05780639e4e7318146102f8578063a0cc6a681461030057610137565b806323b872dd116100ff57806323b872dd1461022357806330adf81f14610259578063313ce567146102615780633408e4701461027f57806342966c681461028757610137565b806304622c2e1461013c57806306fdde0314610156578063095ea7b3146101d357806318160ddd1461021357806318369a2a1461021b575b600080fd5b610144610448565b60408051918252519081900360200190f35b61015e61046c565b6040805160208082528351818301528351919283929083019185019080838360005b83811015610198578181015183820152602001610180565b50505050905090810190601f1680156101c55780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6101ff600480360360408110156101e957600080fd5b506001600160a01b03813516906020013561049c565b604080519115158252519081900360200190f35b6101446104b2565b6101446104b8565b6101ff6004803603606081101561023957600080fd5b506001600160a01b038135811691602081013590911690604001356104c7565b610144610539565b61026961055d565b6040805160ff9092168252519081900360200190f35b610144610562565b6101ff6004803603602081101561029d57600080fd5b5035610566565b610144600480360360208110156102ba57600080fd5b50356001600160a01b031661057a565b610144600480360360208110156102e057600080fd5b50356001600160a01b031661058c565b61015e61059e565b6101446105bd565b6101446105e1565b6101ff6004803603604081101561031e57600080fd5b506001600160a01b038135169060200135610605565b610144610612565b61038d600480360360e081101561035257600080fd5b506001600160a01b03813581169160208101359091169060408101359060608101359060ff6080820135169060a08101359060c00135610636565b005b610144600480360360408110156103a557600080fd5b506001600160a01b0381358116916020013516610736565b61038d60048036036101208110156103d457600080fd5b506001600160a01b03813581169160208101359091169060408101359060608101359060808101359060a08101359060ff60c0820135169060e0810135906101000135610753565b6101ff6004803603604081101561043257600080fd5b506001600160a01b038135169060200135610933565b7f64c0a41a0260272b78f2a5bd50d5ff7c1779bc3bba16dcff4550c7c642b0e4b481565b604051806040016040528060148152602001732432b936b2bd102732ba3bb7b935902a37b5b2b760611b81525081565b60006104a9338484610a12565b50600192915050565b60005481565b6a52b7d2dcc80cd2e400000081565b6001600160a01b03831660009081526002602090815260408083203384529091528120546000198114610523576104fe8184610a74565b6001600160a01b03861660009081526002602090815260408083203384529091529020555b61052e858585610af0565b506001949350505050565b7f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c981565b601281565b4690565b60006105723383610bfa565b506001919050565b60016020526000908152604090205481565b60036020526000908152604090205481565b604051806040016040528060038152602001622422ad60e91b81525081565b7fc89efdaa54c0f20c7adf612882df0950f5a951637e0307cdcb4c672f298b8bc681565b7f7c7c6cdb67a18743f49ec6fa9b35f50d52ed05cbed4cc592e13b44501c1a226781565b60006104a9338484610af0565b7f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f81565b4284101561068b576040805162461bcd60e51b815260206004820152601960248201527f48455a3a3a7065726d69743a20415554485f4558504952454400000000000000604482015290519081900360640190fd5b6001600160a01b0380881660008181526003602090815260409182902080546001810190915582517f6e71edae12b1b97f4d1f60370fef10105fa2faae0126114a169c64845d6126c98184015280840194909452938a1660608401526080830189905260a083019390935260c08083018890528151808403909101815260e0909201905280519101206107218882868686610c8c565b61072c888888610a12565b5050505050505050565b600260209081526000928352604080842090915290825290205481565b8542116107915760405162461bcd60e51b8152600401808060200182810382526032815260200180610e926032913960400191505060405180910390fd5b8442106107cf5760405162461bcd60e51b815260040180806020018281038252602c815260200180610e44602c913960400191505060405180910390fd5b6001600160a01b038916600090815260046020908152604080832087845290915290205460ff16156108325760405162461bcd60e51b8152600401808060200182810382526031815260200180610eef6031913960400191505060405180910390fd5b604080517f7c7c6cdb67a18743f49ec6fa9b35f50d52ed05cbed4cc592e13b44501c1a22676020808301919091526001600160a01b03808d16838501528b166060830152608082018a905260a0820189905260c0820188905260e08083018890528351808403909101815261010090920190925280519101206108b88a82868686610c8c565b6001600160a01b038a1660009081526004602090815260408083208884529091529020805460ff191660011790556108f18a8a8a610af0565b60405185906001600160a01b038c16907f98de503528ee59b575ef0c0a2576a82497bfc029a5685b209e9ec333479b10a590600090a350505050505050505050565b600460209081526000928352604080842090915290825290205460ff1681565b6040805180820190915260118152704d4154483a4144445f4f564552464c4f5760781b60208201528183019083821015610a0b5760405162461bcd60e51b81526004018080602001828103825283818151815260200191508051906020019080838360005b838110156109d05781810151838201526020016109b8565b50505050905090810190601f1680156109fd5780820380516001836020036101000a031916815260200191505b509250505060405180910390fd5b5092915050565b6001600160a01b03808416600081815260026020908152604080832094871680845294825291829020859055815185815291517f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b9259281900390910190a3505050565b6040805180820190915260128152714d4154483a5355425f554e444552464c4f5760701b60208201528183039083821115610a0b5760405162461bcd60e51b81526020600482018181528351602484015283519092839260449091019190850190808383600083156109d05781810151838201526020016109b8565b6001600160a01b0382163014801590610b1157506001600160a01b03821615155b610b4c5760405162461bcd60e51b8152600401808060200182810382526022815260200180610e706022913960400191505060405180910390fd5b6001600160a01b038316600090815260016020526040902054610b6f9082610a74565b6001600160a01b038085166000908152600160205260408082209390935590841681522054610b9e9082610953565b6001600160a01b0380841660008181526001602090815260409182902094909455805185815290519193928716927fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef92918290030190a3505050565b6001600160a01b038216600090815260016020526040902054610c1d9082610a74565b6001600160a01b03831660009081526001602052604081209190915554610c449082610a74565b60009081556040805183815290516001600160a01b038516917fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef919081900360200190a35050565b60007f8b73c3c69bb8fe3d512ecc4cf759cc79239f7b179b0ffacaa9a75d522b39400f7f64c0a41a0260272b78f2a5bd50d5ff7c1779bc3bba16dcff4550c7c642b0e4b47fc89efdaa54c0f20c7adf612882df0950f5a951637e0307cdcb4c672f298b8bc6610cf9610562565b6040805160208082019690965280820194909452606084019290925260808301523060a0808401919091528151808403909101815260c08301825280519084012061190160f01b60e084015260e283018190526101028084018a9052825180850390910181526101228401808452815191860191909120600091829052610142850180855281905260ff8a1661016286015261018285018990526101a285018890529251919550919391926001926101c2808301939192601f198301929081900390910190855afa158015610dd2573d6000803e3d6000fd5b5050604051601f1901519150506001600160a01b03811615801590610e085750876001600160a01b0316816001600160a01b0316145b61072c5760405162461bcd60e51b815260040180806020018281038252602b815260200180610ec4602b913960400191505060405180910390fdfe48455a3a3a7472616e7366657257697468417574686f72697a6174696f6e3a20415554485f4558504952454448455a3a3a5f7472616e736665723a204e4f545f56414c49445f5452414e5346455248455a3a3a7472616e7366657257697468417574686f72697a6174696f6e3a20415554485f4e4f545f5945545f56414c494448455a3a3a5f76616c69646174655369676e6564446174613a20494e56414c49445f5349474e415455524548455a3a3a7472616e7366657257697468417574686f72697a6174696f6e3a20415554485f414c52454144595f55534544a264697066735822122016ca549a428c475103bb37fad40b1571ba0a864be3450231c0028d4f17385b3b64736f6c634300060c0033"

// DeployHEZ deploys a new Ethereum contract, binding an instance of HEZ to it.
func DeployHEZ(auth *bind.TransactOpts, backend bind.ContractBackend, initialHolder common.Address) (common.Address, *types.Transaction, *HEZ, error) {
	parsed, err := abi.JSON(strings.NewReader(HEZABI))
	if err != nil {
		return common.Address{}, nil, nil, tracerr.Wrap(err)
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HEZBin), backend, initialHolder)
	if err != nil {
		return common.Address{}, nil, nil, tracerr.Wrap(err)
	}
	return address, tx, &HEZ{HEZCaller: HEZCaller{contract: contract}, HEZTransactor: HEZTransactor{contract: contract}, HEZFilterer: HEZFilterer{contract: contract}}, nil
}

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
		return nil, tracerr.Wrap(err)
	}
	return &HEZ{HEZCaller: HEZCaller{contract: contract}, HEZTransactor: HEZTransactor{contract: contract}, HEZFilterer: HEZFilterer{contract: contract}}, nil
}

// NewHEZCaller creates a new read-only instance of HEZ, bound to a specific deployed contract.
func NewHEZCaller(address common.Address, caller bind.ContractCaller) (*HEZCaller, error) {
	contract, err := bindHEZ(address, caller, nil, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &HEZCaller{contract: contract}, nil
}

// NewHEZTransactor creates a new write-only instance of HEZ, bound to a specific deployed contract.
func NewHEZTransactor(address common.Address, transactor bind.ContractTransactor) (*HEZTransactor, error) {
	contract, err := bindHEZ(address, nil, transactor, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &HEZTransactor{contract: contract}, nil
}

// NewHEZFilterer creates a new log filterer instance of HEZ, bound to a specific deployed contract.
func NewHEZFilterer(address common.Address, filterer bind.ContractFilterer) (*HEZFilterer, error) {
	contract, err := bindHEZ(address, nil, nil, filterer)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &HEZFilterer{contract: contract}, nil
}

// bindHEZ binds a generic wrapper to an already deployed contract.
func bindHEZ(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HEZABI))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HEZ *HEZRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
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
func (_HEZ *HEZCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
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
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "EIP712DOMAIN_HASH")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "NAME_HASH")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "PERMIT_TYPEHASH")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "TRANSFER_WITH_AUTHORIZATION_TYPEHASH")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "VERSION_HASH")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "allowance", arg0, arg1)
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "authorizationState", arg0, arg1)
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "balanceOf", arg0)
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "decimals")
	return *ret0, tracerr.Wrap(err)
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

// GetChainId is a free data retrieval call binding the contract method 0x3408e470.
//
// Solidity: function getChainId() pure returns(uint256 chainId)
func (_HEZ *HEZCaller) GetChainId(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "getChainId")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "initialBalance")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "name")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "nonces", arg0)
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "symbol")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _HEZ.contract.Call(opts, out, "totalSupply")
	return *ret0, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZApproval)
				if err := _HEZ.contract.UnpackLog(event, "Approval", log); err != nil {
					return tracerr.Wrap(err)
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return tracerr.Wrap(err)
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZAuthorizationUsed)
				if err := _HEZ.contract.UnpackLog(event, "AuthorizationUsed", log); err != nil {
					return tracerr.Wrap(err)
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return tracerr.Wrap(err)
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(HEZTransfer)
				if err := _HEZ.contract.UnpackLog(event, "Transfer", log); err != nil {
					return tracerr.Wrap(err)
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return tracerr.Wrap(err)
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event, nil
}
