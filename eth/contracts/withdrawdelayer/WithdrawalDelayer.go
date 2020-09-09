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
const WithdrawalDelayerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EmergencyModeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"EscapeHatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezGovernanceDAOAddress\",\"type\":\"address\"}],\"name\":\"NewHermezGovernanceDAOAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezKeeperAddress\",\"type\":\"address\"}],\"name\":\"NewHermezKeeperAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newWhiteHackGroupAddress\",\"type\":\"address\"}],\"name\":\"NewWhiteHackGroupAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"NewWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_EMERGENCY_MODE_TIME\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_WITHDRAWAL_DELAY\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"changeWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"_amount\",\"type\":\"uint192\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"depositInfo\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enableEmergencyMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"escapeHatchWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyModeStartingTime\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezGovernanceDAOAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezKeeperAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWhiteHackGroupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWithdrawalDelay\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_initialWithdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_initialHermezRollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezKeeperAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezGovernanceDAOAddress\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_initialWhiteHackGroupAddress\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isEmergencyMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setHermezGovernanceDAOAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setHermezKeeperAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setWhiteHackGroupAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// WithdrawalDelayerBin is the compiled bytecode used for deploying new contracts.
var WithdrawalDelayerBin = "0x60806040526037805460ff60a01b1916905534801561001d57600080fd5b506001600055611776806100326000396000f3fe60806040526004361061011f5760003560e01c8063668cdd67116100a0578063c5b1c7d011610064578063c5b1c7d0146103d7578063cf3a25d9146103ec578063cfc0b64114610427578063d82b217c14610467578063de35f2821461049a5761011f565b8063668cdd6714610334578063a238f9df14610349578063acfd6ea81461037a578063ae7efbbd146103ad578063b4b8e39d146103c25761011f565b806320a194b8116100e757806320a194b814610251578063305887f91461027a5780633d4dff7b1461028f578063493b0170146102e4578063580fc6111461031f5761011f565b806303160940146101245780630a4db01b1461015e5780630e670af5146101935780630fd266d7146101c657806316b487ff146101f7575b600080fd5b34801561013057600080fd5b506101396104d5565b604080516fffffffffffffffffffffffffffffffff9092168252519081900360200190f35b34801561016a57600080fd5b506101916004803603602081101561018157600080fd5b50356001600160a01b03166104e4565b005b34801561019f57600080fd5b50610191600480360360208110156101b657600080fd5b50356001600160401b0316610590565b3480156101d257600080fd5b506101db6106c0565b604080516001600160a01b039092168252519081900360200190f35b34801561020357600080fd5b50610191600480360360a081101561021a57600080fd5b506001600160401b03813516906001600160a01b0360208201358116916040810135821691606082013581169160800135166106cf565b34801561025d57600080fd5b506102666107db565b604080519115158252519081900360200190f35b34801561028657600080fd5b506101db6107eb565b34801561029b57600080fd5b506102b9600480360360208110156102b257600080fd5b50356107fa565b604080516001600160c01b0390931683526001600160401b0390911660208301528051918290030190f35b3480156102f057600080fd5b506102b96004803603604081101561030757600080fd5b506001600160a01b0381358116916020013516610827565b34801561032b57600080fd5b506101db6108b8565b34801561034057600080fd5b506101396108c7565b34801561035557600080fd5b5061035e6108dd565b604080516001600160401b039092168252519081900360200190f35b34801561038657600080fd5b506101916004803603602081101561039d57600080fd5b50356001600160a01b03166108e4565b3480156103b957600080fd5b506101db61099d565b3480156103ce57600080fd5b5061035e6109ac565b3480156103e357600080fd5b506101916109b4565b3480156103f857600080fd5b506101916004803603604081101561040f57600080fd5b506001600160a01b0381358116916020013516610adc565b6101916004803603606081101561043d57600080fd5b5080356001600160a01b0390811691602081013590911690604001356001600160c01b0316610d72565b34801561047357600080fd5b506101916004803603602081101561048a57600080fd5b50356001600160a01b0316611191565b3480156104a657600080fd5b50610191600480360360408110156104bd57600080fd5b506001600160a01b038135811691602001351661124a565b6034546001600160401b031690565b6036546001600160a01b03163314610536576040805162461bcd60e51b815260206004820152601060248201526f4f6e6c7920574847206164647265737360801b604482015290519081900360640190fd5b603680546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517f284ca073b8bdde2195ae98779277678773a99d7739e5f0477dc19a03fc689011916020908290030190a150565b6037546001600160a01b03163314806105b357506038546001600160a01b031633145b610604576040805162461bcd60e51b815260206004820152601c60248201527f4f6e6c79206865726d657a206b6565706572206f7220726f6c6c757000000000604482015290519081900360640190fd5b621275006001600160401b0382161115610665576040805162461bcd60e51b815260206004820152601c60248201527f45786365656473204d41585f5749544844524157414c5f44454c415900000000604482015290519081900360640190fd5b6034805467ffffffffffffffff19166001600160401b03838116919091179182905560408051929091168252517f6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7916020908290030190a150565b6038546001600160a01b031681565b600154610100900460ff16806106e857506106e86114c2565b806106f6575060015460ff16155b6107315760405162461bcd60e51b815260040180806020018281038252602e815260200180611713602e913960400191505060405180910390fd5b600154610100900460ff1615801561075b576001805460ff1961ff00199091166101001716811790555b6034805467ffffffffffffffff19166001600160401b038816179055603880546001600160a01b03199081166001600160a01b03888116919091179092556037805482168784161790556035805482168684161790556036805490911691841691909117905580156107d3576001805461ff00191690555b505050505050565b603754600160a01b900460ff1690565b6037546001600160a01b031690565b6039602052600090815260409020546001600160c01b03811690600160c01b90046001600160401b031682565b6000806108326116fb565b505060408051606094851b6001600160601b03199081166020808401919091529490951b90941660348501528051808503602801815260488501808352815191850191909120600090815260399094529281902060888501909152546001600160c01b03811692839052600160c01b90046001600160401b031660689093018390525091565b6035546001600160a01b031690565b603454600160401b90046001600160401b031690565b6212750081565b6035546001600160a01b03163314610943576040805162461bcd60e51b815260206004820152601a60248201527f4f6e6c79204865726d657a20476f7665726e616e63652044414f000000000000604482015290519081900360640190fd5b603580546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517f03683be8debd93f8f5ff23dd03419bfcb9b8287a1868b0f130d858f03c3a08a1916020908290030190a150565b6036546001600160a01b031690565b6301dfe20081565b6037546001600160a01b03163314610a13576040805162461bcd60e51b815260206004820152601860248201527f4f6e6c79206865726d657a4b6565706572416464726573730000000000000000604482015290519081900360640190fd5b603754600160a01b900460ff1615610a72576040805162461bcd60e51b815260206004820152601e60248201527f456d657267656e6379206d6f646520616c726561647920656e61626c65640000604482015290519081900360640190fd5b6037805460ff60a01b1916600160a01b179055603480546001600160401b034216600160401b026fffffffffffffffff0000000000000000199091161790556040517f2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c590600090a1565b60026000541415610b34576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6002600055603754600160a01b900460ff16610b8d576040805162461bcd60e51b81526020600482015260136024820152724f6e6c7920456d657267656e6379204d6f646560681b604482015290519081900360640190fd5b6036546001600160a01b0316331480610bb057506035546001600160a01b031633145b610c01576040805162461bcd60e51b815260206004820152601960248201527f4f6e6c7920476f7665726e616e636544414f206f722057484700000000000000604482015290519081900360640190fd5b6036546001600160a01b0316331415610c88576034546001600160401b03600160401b90910481166301dfe200018116429091161015610c88576040805162461bcd60e51b815260206004820152601a60248201527f4e4f204d41585f454d455247454e43595f4d4f44455f54494d45000000000000604482015290519081900360640190fd5b60006001600160a01b038216610ca9575047610ca483826114c8565b610d2d565b604080516370a0823160e01b8152306004820152905183916001600160a01b038316916370a0823191602480820192602092909190829003018186803b158015610cf257600080fd5b505afa158015610d06573d6000803e3d6000fd5b505050506040513d6020811015610d1c57600080fd5b50519150610d2b838584611569565b505b6040516001600160a01b03808416919085169033907f065a030f4e05509e10831215a77cf703ff0d78a252b9fa008749d832eb1f61d990600090a45050600160005550565b6038546001600160a01b03163314610dd1576040805162461bcd60e51b815260206004820152601860248201527f4f6e6c79206865726d657a526f6c6c7570416464726573730000000000000000604482015290519081900360640190fd5b3415610e95576001600160a01b03821615610e33576040805162461bcd60e51b815260206004820152601d60248201527f4554482073686f756c6420626520746865203078302061646472657373000000604482015290519081900360640190fd5b34816001600160c01b031614610e90576040805162461bcd60e51b815260206004820152601e60248201527f446966666572656e7420616d6f756e7420616e64206d73672e76616c75650000604482015290519081900360640190fd5b611049565b60385460408051636eb1769f60e11b81526001600160a01b03928316600482015230602482015290516001600160c01b0384169285169163dd62ed3e916044808301926020929190829003018186803b158015610ef157600080fd5b505afa158015610f05573d6000803e3d6000fd5b505050506040513d6020811015610f1b57600080fd5b50511015610f70576040805162461bcd60e51b815260206004820152601d60248201527f446f65736e2774206861766520656e6f75676820616c6c6f77616e6365000000604482015290519081900360640190fd5b603854604080516323b872dd60e01b81526001600160a01b0392831660048201523060248201526001600160c01b03841660448201529051918416916323b872dd916064808201926020929091908290030181600087803b158015610fd457600080fd5b505af1158015610fe8573d6000803e3d6000fd5b505050506040513d6020811015610ffe57600080fd5b5051611049576040805162461bcd60e51b8152602060048201526015602482015274151bdad95b88151c985b9cd9995c8811985a5b1959605a1b604482015290519081900360640190fd5b60408051606085811b6001600160601b03199081166020808501919091529186901b166034830152825180830360280181526048909201835281519181019190912060008181526039909252919020546001600160c01b0390811683810191821610156110f0576040805162461bcd60e51b815260206004820152601060248201526f4465706f736974206f766572666c6f7760801b604482015290519081900360640190fd5b60008281526039602090815260409182902080546001600160401b03428116600160c01b9081026001600160c01b038089166001600160c01b03199095169490941784161793849055855192891683529092049091169181019190915281516001600160a01b0380881693908916927f41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd929081900390910190a35050505050565b6037546001600160a01b031633146111f0576040805162461bcd60e51b815260206004820152601a60248201527f4f6e6c79204865726d657a204b65657065722041646472657373000000000000604482015290519081900360640190fd5b603780546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517fc1e9be84fce652abec6a6944f7ec5bbb40de18caa44c285b05a0de7e3ad9d016916020908290030190a150565b600260005414156112a2576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6002600055603754600160a01b900460ff16156112f7576040805162461bcd60e51b815260206004820152600e60248201526d456d657267656e6379206d6f646560901b604482015290519081900360640190fd5b60408051606084811b6001600160601b03199081166020808501919091529185901b166034830152825180830360280181526048909201835281519181019190912060008181526039909252919020546001600160c01b031680611399576040805162461bcd60e51b81526020600482015260146024820152734e6f2066756e647320746f20776974686472617760601b604482015290519081900360640190fd5b6034546000838152603960205260409020546001600160401b03918216600160c01b909104821601811642909116101561141a576040805162461bcd60e51b815260206004820152601a60248201527f5769746864726177616c206e6f7420616c6c6f77656420796574000000000000604482015290519081900360640190fd5b6000828152603960205260408120556001600160a01b03831661144f5761144a84826001600160c01b03166114c8565b611463565b6114638385836001600160c01b0316611569565b836001600160a01b0316836001600160a01b03167f72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf32718360405180826001600160c01b0316815260200191505060405180910390a3505060016000555050565b303b1590565b6040516000906001600160a01b0384169083908381818185875af1925050503d8060008114611513576040519150601f19603f3d011682016040523d82523d6000602084013e611518565b606091505b5050905080611564576040805162461bcd60e51b8152602060048201526013602482015272115512081d1c985b9cd9995c8819985a5b1959606a1b604482015290519081900360640190fd5b505050565b604080518082018252601981527f7472616e7366657228616464726573732c75696e74323536290000000000000060209182015281516001600160a01b0385811660248301526044808301869052845180840390910181526064909201845291810180516001600160e01b031663a9059cbb60e01b1781529251815160009460609489169392918291908083835b602083106116165780518252601f1990920191602091820191016115f7565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d8060008114611678576040519150601f19603f3d011682016040523d82523d6000602084013e61167d565b606091505b50915091508180156116ab5750805115806116ab57508080602001905160208110156116a857600080fd5b50515b6116f4576040805162461bcd60e51b8152602060048201526015602482015274151bdad95b88151c985b9cd9995c8811985a5b1959605a1b604482015290519081900360640190fd5b5050505050565b60408051808201909152600080825260208201529056fe436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a6564a2646970667358221220c1c762163fd298f0328559fb5d7027caf2d51af3b3691a9b8808a2b55947492d64736f6c634300060c0033"

// DeployWithdrawalDelayer deploys a new Ethereum contract, binding an instance of WithdrawalDelayer to it.
func DeployWithdrawalDelayer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *WithdrawalDelayer, error) {
	parsed, err := abi.JSON(strings.NewReader(WithdrawalDelayerABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WithdrawalDelayerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &WithdrawalDelayer{WithdrawalDelayerCaller: WithdrawalDelayerCaller{contract: contract}, WithdrawalDelayerTransactor: WithdrawalDelayerTransactor{contract: contract}, WithdrawalDelayerFilterer: WithdrawalDelayerFilterer{contract: contract}}, nil
}

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
func (_WithdrawalDelayer *WithdrawalDelayerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
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
func (_WithdrawalDelayer *WithdrawalDelayerCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
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
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "MAX_EMERGENCY_MODE_TIME")
	return *ret0, err
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
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "MAX_WITHDRAWAL_DELAY")
	return *ret0, err
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
	var (
		ret0 = new(*big.Int)
		ret1 = new(uint64)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _WithdrawalDelayer.contract.Call(opts, out, "depositInfo", _owner, _token)
	return *ret0, *ret1, err
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
	ret := new(struct {
		Amount           *big.Int
		DepositTimestamp uint64
	})
	out := ret
	err := _WithdrawalDelayer.contract.Call(opts, out, "deposits", arg0)
	return *ret, err
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

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetEmergencyModeStartingTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getEmergencyModeStartingTime")
	return *ret0, err
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetEmergencyModeStartingTime() (*big.Int, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyModeStartingTime(&_WithdrawalDelayer.CallOpts)
}

// GetEmergencyModeStartingTime is a free data retrieval call binding the contract method 0x668cdd67.
//
// Solidity: function getEmergencyModeStartingTime() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetEmergencyModeStartingTime() (*big.Int, error) {
	return _WithdrawalDelayer.Contract.GetEmergencyModeStartingTime(&_WithdrawalDelayer.CallOpts)
}

// GetHermezGovernanceDAOAddress is a free data retrieval call binding the contract method 0x580fc611.
//
// Solidity: function getHermezGovernanceDAOAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetHermezGovernanceDAOAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getHermezGovernanceDAOAddress")
	return *ret0, err
}

// GetHermezGovernanceDAOAddress is a free data retrieval call binding the contract method 0x580fc611.
//
// Solidity: function getHermezGovernanceDAOAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetHermezGovernanceDAOAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezGovernanceDAOAddress(&_WithdrawalDelayer.CallOpts)
}

// GetHermezGovernanceDAOAddress is a free data retrieval call binding the contract method 0x580fc611.
//
// Solidity: function getHermezGovernanceDAOAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetHermezGovernanceDAOAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezGovernanceDAOAddress(&_WithdrawalDelayer.CallOpts)
}

// GetHermezKeeperAddress is a free data retrieval call binding the contract method 0x305887f9.
//
// Solidity: function getHermezKeeperAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetHermezKeeperAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getHermezKeeperAddress")
	return *ret0, err
}

// GetHermezKeeperAddress is a free data retrieval call binding the contract method 0x305887f9.
//
// Solidity: function getHermezKeeperAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetHermezKeeperAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezKeeperAddress(&_WithdrawalDelayer.CallOpts)
}

// GetHermezKeeperAddress is a free data retrieval call binding the contract method 0x305887f9.
//
// Solidity: function getHermezKeeperAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetHermezKeeperAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetHermezKeeperAddress(&_WithdrawalDelayer.CallOpts)
}

// GetWhiteHackGroupAddress is a free data retrieval call binding the contract method 0xae7efbbd.
//
// Solidity: function getWhiteHackGroupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetWhiteHackGroupAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getWhiteHackGroupAddress")
	return *ret0, err
}

// GetWhiteHackGroupAddress is a free data retrieval call binding the contract method 0xae7efbbd.
//
// Solidity: function getWhiteHackGroupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetWhiteHackGroupAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetWhiteHackGroupAddress(&_WithdrawalDelayer.CallOpts)
}

// GetWhiteHackGroupAddress is a free data retrieval call binding the contract method 0xae7efbbd.
//
// Solidity: function getWhiteHackGroupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetWhiteHackGroupAddress() (common.Address, error) {
	return _WithdrawalDelayer.Contract.GetWhiteHackGroupAddress(&_WithdrawalDelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetWithdrawalDelay(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getWithdrawalDelay")
	return *ret0, err
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerSession) GetWithdrawalDelay() (*big.Int, error) {
	return _WithdrawalDelayer.Contract.GetWithdrawalDelay(&_WithdrawalDelayer.CallOpts)
}

// GetWithdrawalDelay is a free data retrieval call binding the contract method 0x03160940.
//
// Solidity: function getWithdrawalDelay() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCallerSession) GetWithdrawalDelay() (*big.Int, error) {
	return _WithdrawalDelayer.Contract.GetWithdrawalDelay(&_WithdrawalDelayer.CallOpts)
}

// HermezRollupAddress is a free data retrieval call binding the contract method 0x0fd266d7.
//
// Solidity: function hermezRollupAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) HermezRollupAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "hermezRollupAddress")
	return *ret0, err
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
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "isEmergencyMode")
	return *ret0, err
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

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0xcf3a25d9.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) EscapeHatchWithdrawal(opts *bind.TransactOpts, _to common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "escapeHatchWithdrawal", _to, _token)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0xcf3a25d9.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EscapeHatchWithdrawal(&_WithdrawalDelayer.TransactOpts, _to, _token)
}

// EscapeHatchWithdrawal is a paid mutator transaction binding the contract method 0xcf3a25d9.
//
// Solidity: function escapeHatchWithdrawal(address _to, address _token) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) EscapeHatchWithdrawal(_to common.Address, _token common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.EscapeHatchWithdrawal(&_WithdrawalDelayer.TransactOpts, _to, _token)
}

// Initialize is a paid mutator transaction binding the contract method 0x16b487ff.
//
// Solidity: function initialize(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) Initialize(opts *bind.TransactOpts, _initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "initialize", _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x16b487ff.
//
// Solidity: function initialize(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) Initialize(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Initialize(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0x16b487ff.
//
// Solidity: function initialize(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) Initialize(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.Initialize(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
}

// SetHermezGovernanceDAOAddress is a paid mutator transaction binding the contract method 0xacfd6ea8.
//
// Solidity: function setHermezGovernanceDAOAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) SetHermezGovernanceDAOAddress(opts *bind.TransactOpts, newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "setHermezGovernanceDAOAddress", newAddress)
}

// SetHermezGovernanceDAOAddress is a paid mutator transaction binding the contract method 0xacfd6ea8.
//
// Solidity: function setHermezGovernanceDAOAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) SetHermezGovernanceDAOAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetHermezGovernanceDAOAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
}

// SetHermezGovernanceDAOAddress is a paid mutator transaction binding the contract method 0xacfd6ea8.
//
// Solidity: function setHermezGovernanceDAOAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) SetHermezGovernanceDAOAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetHermezGovernanceDAOAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
}

// SetHermezKeeperAddress is a paid mutator transaction binding the contract method 0xd82b217c.
//
// Solidity: function setHermezKeeperAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) SetHermezKeeperAddress(opts *bind.TransactOpts, newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "setHermezKeeperAddress", newAddress)
}

// SetHermezKeeperAddress is a paid mutator transaction binding the contract method 0xd82b217c.
//
// Solidity: function setHermezKeeperAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) SetHermezKeeperAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetHermezKeeperAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
}

// SetHermezKeeperAddress is a paid mutator transaction binding the contract method 0xd82b217c.
//
// Solidity: function setHermezKeeperAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) SetHermezKeeperAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetHermezKeeperAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
}

// SetWhiteHackGroupAddress is a paid mutator transaction binding the contract method 0x0a4db01b.
//
// Solidity: function setWhiteHackGroupAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) SetWhiteHackGroupAddress(opts *bind.TransactOpts, newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "setWhiteHackGroupAddress", newAddress)
}

// SetWhiteHackGroupAddress is a paid mutator transaction binding the contract method 0x0a4db01b.
//
// Solidity: function setWhiteHackGroupAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) SetWhiteHackGroupAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetWhiteHackGroupAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
}

// SetWhiteHackGroupAddress is a paid mutator transaction binding the contract method 0x0a4db01b.
//
// Solidity: function setWhiteHackGroupAddress(address newAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) SetWhiteHackGroupAddress(newAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.SetWhiteHackGroupAddress(&_WithdrawalDelayer.TransactOpts, newAddress)
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
	Who   common.Address
	To    common.Address
	Token common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterEscapeHatchWithdrawal is a free log retrieval operation binding the contract event 0x065a030f4e05509e10831215a77cf703ff0d78a252b9fa008749d832eb1f61d9.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token)
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

// WatchEscapeHatchWithdrawal is a free log subscription operation binding the contract event 0x065a030f4e05509e10831215a77cf703ff0d78a252b9fa008749d832eb1f61d9.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token)
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

// ParseEscapeHatchWithdrawal is a log parse operation binding the contract event 0x065a030f4e05509e10831215a77cf703ff0d78a252b9fa008749d832eb1f61d9.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseEscapeHatchWithdrawal(log types.Log) (*WithdrawalDelayerEscapeHatchWithdrawal, error) {
	event := new(WithdrawalDelayerEscapeHatchWithdrawal)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WithdrawalDelayerNewHermezGovernanceDAOAddressIterator is returned from FilterNewHermezGovernanceDAOAddress and is used to iterate over the raw logs and unpacked data for NewHermezGovernanceDAOAddress events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezGovernanceDAOAddressIterator struct {
	Event *WithdrawalDelayerNewHermezGovernanceDAOAddress // Event containing the contract specifics and raw log

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
func (it *WithdrawalDelayerNewHermezGovernanceDAOAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewHermezGovernanceDAOAddress)
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
		it.Event = new(WithdrawalDelayerNewHermezGovernanceDAOAddress)
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
func (it *WithdrawalDelayerNewHermezGovernanceDAOAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewHermezGovernanceDAOAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewHermezGovernanceDAOAddress represents a NewHermezGovernanceDAOAddress event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezGovernanceDAOAddress struct {
	NewHermezGovernanceDAOAddress common.Address
	Raw                           types.Log // Blockchain specific contextual infos
}

// FilterNewHermezGovernanceDAOAddress is a free log retrieval operation binding the contract event 0x03683be8debd93f8f5ff23dd03419bfcb9b8287a1868b0f130d858f03c3a08a1.
//
// Solidity: event NewHermezGovernanceDAOAddress(address newHermezGovernanceDAOAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewHermezGovernanceDAOAddress(opts *bind.FilterOpts) (*WithdrawalDelayerNewHermezGovernanceDAOAddressIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewHermezGovernanceDAOAddress")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewHermezGovernanceDAOAddressIterator{contract: _WithdrawalDelayer.contract, event: "NewHermezGovernanceDAOAddress", logs: logs, sub: sub}, nil
}

// WatchNewHermezGovernanceDAOAddress is a free log subscription operation binding the contract event 0x03683be8debd93f8f5ff23dd03419bfcb9b8287a1868b0f130d858f03c3a08a1.
//
// Solidity: event NewHermezGovernanceDAOAddress(address newHermezGovernanceDAOAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewHermezGovernanceDAOAddress(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewHermezGovernanceDAOAddress) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewHermezGovernanceDAOAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewHermezGovernanceDAOAddress)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceDAOAddress", log); err != nil {
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

// ParseNewHermezGovernanceDAOAddress is a log parse operation binding the contract event 0x03683be8debd93f8f5ff23dd03419bfcb9b8287a1868b0f130d858f03c3a08a1.
//
// Solidity: event NewHermezGovernanceDAOAddress(address newHermezGovernanceDAOAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewHermezGovernanceDAOAddress(log types.Log) (*WithdrawalDelayerNewHermezGovernanceDAOAddress, error) {
	event := new(WithdrawalDelayerNewHermezGovernanceDAOAddress)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceDAOAddress", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WithdrawalDelayerNewHermezKeeperAddressIterator is returned from FilterNewHermezKeeperAddress and is used to iterate over the raw logs and unpacked data for NewHermezKeeperAddress events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezKeeperAddressIterator struct {
	Event *WithdrawalDelayerNewHermezKeeperAddress // Event containing the contract specifics and raw log

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
func (it *WithdrawalDelayerNewHermezKeeperAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewHermezKeeperAddress)
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
		it.Event = new(WithdrawalDelayerNewHermezKeeperAddress)
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
func (it *WithdrawalDelayerNewHermezKeeperAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewHermezKeeperAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewHermezKeeperAddress represents a NewHermezKeeperAddress event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewHermezKeeperAddress struct {
	NewHermezKeeperAddress common.Address
	Raw                    types.Log // Blockchain specific contextual infos
}

// FilterNewHermezKeeperAddress is a free log retrieval operation binding the contract event 0xc1e9be84fce652abec6a6944f7ec5bbb40de18caa44c285b05a0de7e3ad9d016.
//
// Solidity: event NewHermezKeeperAddress(address newHermezKeeperAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewHermezKeeperAddress(opts *bind.FilterOpts) (*WithdrawalDelayerNewHermezKeeperAddressIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewHermezKeeperAddress")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewHermezKeeperAddressIterator{contract: _WithdrawalDelayer.contract, event: "NewHermezKeeperAddress", logs: logs, sub: sub}, nil
}

// WatchNewHermezKeeperAddress is a free log subscription operation binding the contract event 0xc1e9be84fce652abec6a6944f7ec5bbb40de18caa44c285b05a0de7e3ad9d016.
//
// Solidity: event NewHermezKeeperAddress(address newHermezKeeperAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewHermezKeeperAddress(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewHermezKeeperAddress) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewHermezKeeperAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewHermezKeeperAddress)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezKeeperAddress", log); err != nil {
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

// ParseNewHermezKeeperAddress is a log parse operation binding the contract event 0xc1e9be84fce652abec6a6944f7ec5bbb40de18caa44c285b05a0de7e3ad9d016.
//
// Solidity: event NewHermezKeeperAddress(address newHermezKeeperAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewHermezKeeperAddress(log types.Log) (*WithdrawalDelayerNewHermezKeeperAddress, error) {
	event := new(WithdrawalDelayerNewHermezKeeperAddress)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezKeeperAddress", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WithdrawalDelayerNewWhiteHackGroupAddressIterator is returned from FilterNewWhiteHackGroupAddress and is used to iterate over the raw logs and unpacked data for NewWhiteHackGroupAddress events raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewWhiteHackGroupAddressIterator struct {
	Event *WithdrawalDelayerNewWhiteHackGroupAddress // Event containing the contract specifics and raw log

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
func (it *WithdrawalDelayerNewWhiteHackGroupAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(WithdrawalDelayerNewWhiteHackGroupAddress)
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
		it.Event = new(WithdrawalDelayerNewWhiteHackGroupAddress)
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
func (it *WithdrawalDelayerNewWhiteHackGroupAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *WithdrawalDelayerNewWhiteHackGroupAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// WithdrawalDelayerNewWhiteHackGroupAddress represents a NewWhiteHackGroupAddress event raised by the WithdrawalDelayer contract.
type WithdrawalDelayerNewWhiteHackGroupAddress struct {
	NewWhiteHackGroupAddress common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterNewWhiteHackGroupAddress is a free log retrieval operation binding the contract event 0x284ca073b8bdde2195ae98779277678773a99d7739e5f0477dc19a03fc689011.
//
// Solidity: event NewWhiteHackGroupAddress(address newWhiteHackGroupAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) FilterNewWhiteHackGroupAddress(opts *bind.FilterOpts) (*WithdrawalDelayerNewWhiteHackGroupAddressIterator, error) {

	logs, sub, err := _WithdrawalDelayer.contract.FilterLogs(opts, "NewWhiteHackGroupAddress")
	if err != nil {
		return nil, err
	}
	return &WithdrawalDelayerNewWhiteHackGroupAddressIterator{contract: _WithdrawalDelayer.contract, event: "NewWhiteHackGroupAddress", logs: logs, sub: sub}, nil
}

// WatchNewWhiteHackGroupAddress is a free log subscription operation binding the contract event 0x284ca073b8bdde2195ae98779277678773a99d7739e5f0477dc19a03fc689011.
//
// Solidity: event NewWhiteHackGroupAddress(address newWhiteHackGroupAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewWhiteHackGroupAddress(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewWhiteHackGroupAddress) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewWhiteHackGroupAddress")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewWhiteHackGroupAddress)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWhiteHackGroupAddress", log); err != nil {
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

// ParseNewWhiteHackGroupAddress is a log parse operation binding the contract event 0x284ca073b8bdde2195ae98779277678773a99d7739e5f0477dc19a03fc689011.
//
// Solidity: event NewWhiteHackGroupAddress(address newWhiteHackGroupAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewWhiteHackGroupAddress(log types.Log) (*WithdrawalDelayerNewWhiteHackGroupAddress, error) {
	event := new(WithdrawalDelayerNewWhiteHackGroupAddress)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWhiteHackGroupAddress", log); err != nil {
		return nil, err
	}
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
	return event, nil
}
