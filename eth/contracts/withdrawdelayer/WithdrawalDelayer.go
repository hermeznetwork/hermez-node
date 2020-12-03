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
	"github.com/hermeznetwork/tracerr"
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
const WithdrawalDelayerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EmergencyModeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EscapeHatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"NewEmergencyCouncil\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezGovernanceAddress\",\"type\":\"address\"}],\"name\":\"NewHermezGovernanceAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"NewWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_EMERGENCY_MODE_TIME\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_WITHDRAWAL_DELAY\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"changeWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"claimGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"_amount\",\"type\":\"uint192\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"depositInfo\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enableEmergencyMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"escapeHatchWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyCouncil\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyModeStartingTime\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezGovernanceAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWithdrawalDelay\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isEmergencyMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingEmergencyCouncil\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pendingGovernance\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"transferEmergencyCouncil\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newGovernance\",\"type\":\"address\"}],\"name\":\"transferGovernance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_initialWithdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_initialHermezRollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezGovernanceAddress\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_initialEmergencyCouncil\",\"type\":\"address\"}],\"name\":\"withdrawalDelayerInitializer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// WithdrawalDelayerBin is the compiled bytecode used for deploying new contracts.
var WithdrawalDelayerBin = "0x608060405234801561001057600080fd5b50611cf5806100206000396000f3fe6080604052600436106101355760003560e01c80637fd6b102116100ab578063ca79033f1161006f578063ca79033f146103ee578063cfc0b64114610403578063d38bfff414610443578063db2a1a8114610476578063de35f282146104a9578063f39c38a0146104e457610135565b80637fd6b1021461033b57806399ef11c51461037e578063a238f9df14610393578063b4b8e39d146103c4578063c5b1c7d0146103d957610135565b80633d4dff7b116100fd5780633d4dff7b1461021857806342cb72161461026d578063493b0170146102c15780635d36b190146102fc578063668cdd671461031157806367fa24031461032657610135565b8063031609401461013a5780630b21d430146101745780630e670af5146101a55780630fd266d7146101da57806320a194b8146101ef575b600080fd5b34801561014657600080fd5b5061014f6104f9565b604080516fffffffffffffffffffffffffffffffff9092168252519081900360200190f35b34801561018057600080fd5b50610189610508565b604080516001600160a01b039092168252519081900360200190f35b3480156101b157600080fd5b506101d8600480360360208110156101c857600080fd5b50356001600160401b0316610517565b005b3480156101e657600080fd5b5061018961061b565b3480156101fb57600080fd5b5061020461062a565b604080519115158252519081900360200190f35b34801561022457600080fd5b506102426004803603602081101561023b57600080fd5b503561063a565b604080516001600160c01b0390931683526001600160401b0390911660208301528051918290030190f35b34801561027957600080fd5b506101d86004803603608081101561029057600080fd5b506001600160401b03813516906001600160a01b036020820135811691604081013582169160609091013516610667565b3480156102cd57600080fd5b50610242600480360360408110156102e457600080fd5b506001600160a01b0381358116916020013516610776565b34801561030857600080fd5b506101d8610807565b34801561031d57600080fd5b5061014f6108b1565b34801561033257600080fd5b506101896108c7565b34801561034757600080fd5b506101d86004803603606081101561035e57600080fd5b506001600160a01b038135811691602081013590911690604001356108d6565b34801561038a57600080fd5b50610189610af9565b34801561039f57600080fd5b506103a8610b08565b604080516001600160401b039092168252519081900360200190f35b3480156103d057600080fd5b506103a8610b0f565b3480156103e557600080fd5b506101d8610b16565b3480156103fa57600080fd5b506101d8610c12565b6101d86004803603606081101561041957600080fd5b5080356001600160a01b0390811691602081013590911690604001356001600160c01b0316610cbc565b34801561044f57600080fd5b506101d86004803603602081101561046657600080fd5b50356001600160a01b031661106e565b34801561048257600080fd5b506101d86004803603602081101561049957600080fd5b50356001600160a01b03166110d9565b3480156104b557600080fd5b506101d8600480360360408110156104cc57600080fd5b506001600160a01b0381358116916020013516611144565b3480156104f057600080fd5b5061018961139e565b6065546001600160401b031690565b6066546001600160a01b031690565b6066546001600160a01b031633148061053a5750606a546001600160a01b031633145b6105755760405162461bcd60e51b8152600401808060200182810382526043815260200180611a146043913960600191505060405180910390fd5b621275006001600160401b03821611156105c05760405162461bcd60e51b8152600401808060200182810382526046815260200180611bb76046913960600191505060405180910390fd5b6065805467ffffffffffffffff19166001600160401b03838116919091179182905560408051929091168252517f6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7916020908290030190a150565b606a546001600160a01b031681565b606954600160a01b900460ff1690565b606b602052600090815260409020546001600160c01b03811690600160c01b90046001600160401b031682565b600054610100900460ff168061068057506106806113ad565b8061068e575060005460ff16155b6106c95760405162461bcd60e51b815260040180806020018281038252602e815260200180611ac1602e913960400191505060405180910390fd5b600054610100900460ff161580156106f4576000805460ff1961ff0019909116610100171660011790555b6106fc6113b3565b6065805467ffffffffffffffff19166001600160401b038716179055606a80546001600160a01b03199081166001600160a01b0387811691909117909255606680548216868416179055606980549091169184169190911760ff60a01b19169055801561076f576000805461ff00191690555b5050505050565b6000806107816117b3565b505060408051606094851b6001600160601b03199081166020808401919091529490951b909416603485015280518085036028018152604885018083528151918501919091206000908152606b9094529281902060888501909152546001600160c01b03811692839052600160c01b90046001600160401b031660689093018390525091565b6067546001600160a01b031633146108505760405162461bcd60e51b815260040180806020018281038252603b8152602001806119a3603b913960400191505060405180910390fd5b60678054606680546001600160a01b038084166001600160a01b03199283161792839055921690925560408051929091168252517f3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf916020908290030190a1565b606554600160401b90046001600160401b031690565b6068546001600160a01b031681565b60335460ff1661092d576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff19169055606954600160a01b900460ff1661097f5760405162461bcd60e51b81526004018080602001828103825260348152602001806118ca6034913960400191505060405180910390fd5b6069546001600160a01b03163314806109a257506066546001600160a01b031633145b6109dd5760405162461bcd60e51b8152600401808060200182810382526039815260200180611b206039913960400191505060405180910390fd5b6069546001600160a01b031633148015610a0857506066546069546001600160a01b03908116911614155b15610a6a576065546001600160401b03600160401b909104811662eff100018116429091161015610a6a5760405162461bcd60e51b815260040180806020018281038252604481526020018061192d6044913960600191505060405180910390fd5b6001600160a01b038216610a8757610a828382611462565b610a92565b610a928284836114f7565b816001600160a01b0316836001600160a01b0316336001600160a01b03167fde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693846040518082815260200191505060405180910390a450506033805460ff1916600117905550565b6069546001600160a01b031690565b6212750081565b62eff10081565b6066546001600160a01b03163314610b5f5760405162461bcd60e51b8152600401808060200182810382526037815260200180611c326037913960400191505060405180910390fd5b606954600160a01b900460ff1615610ba85760405162461bcd60e51b81526004018080602001828103825260378152602001806118936037913960400191505060405180910390fd5b6069805460ff60a01b1916600160a01b179055606580546001600160401b034216600160401b026fffffffffffffffff0000000000000000199091161790556040517f2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c590600090a1565b6068546001600160a01b03163314610c5b5760405162461bcd60e51b81526004018080602001828103825260418152602001806117f36041913960600191505060405180910390fd5b60688054606980546001600160a01b038084166001600160a01b03199283161792839055921690925560408051929091168252517fcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396916020908290030190a1565b60335460ff16610d13576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff19169055606a546001600160a01b03163314610d665760405162461bcd60e51b8152600401808060200182810382526027815260200180611a9a6027913960400191505060405180910390fd5b3415610dfe576001600160a01b03821615610db25760405162461bcd60e51b815260040180806020018281038252602f8152602001806118fe602f913960400191505060405180910390fd5b34816001600160c01b031614610df95760405162461bcd60e51b81526004018080602001828103825260288152602001806117cb6028913960400191505060405180910390fd5b611051565b606a5460408051636eb1769f60e11b81526001600160a01b03928316600482015230602482015290516001600160c01b0384169285169163dd62ed3e916044808301926020929190829003018186803b158015610e5a57600080fd5b505afa158015610e6e573d6000803e3d6000fd5b505050506040513d6020811015610e8457600080fd5b50511015610ec35760405162461bcd60e51b8152600401808060200182810382526030815260200180611c696030913960400191505060405180910390fd5b60006060836001600160a01b031660405180606001604052806025815260200161186e602591398051602091820120606a54604080516001600160a01b0390921660248301523060448301526001600160c01b038816606480840191909152815180840390910181526084909201815292810180516001600160e01b03166001600160e01b031990931692909217825291518251909182918083835b60208310610f7e5780518252601f199092019160209182019101610f5f565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d8060008114610fe0576040519150601f19603f3d011682016040523d82523d6000602084013e610fe5565b606091505b5091509150818015611013575080511580611013575080806020019051602081101561101057600080fd5b50515b61104e5760405162461bcd60e51b8152600401808060200182810382526031815260200180611aef6031913960400191505060405180910390fd5b50505b61105c838383611674565b50506033805460ff1916600117905550565b6066546001600160a01b031633146110b75760405162461bcd60e51b81526004018080602001828103825260368152602001806119de6036913960400191505060405180910390fd5b606780546001600160a01b0319166001600160a01b0392909216919091179055565b6069546001600160a01b031633146111225760405162461bcd60e51b8152600401808060200182810382526043815260200180611a576043913960600191505060405180910390fd5b606880546001600160a01b0319166001600160a01b0392909216919091179055565b60335460ff1661119b576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff19169055606954600160a01b900460ff16156111ee5760405162461bcd60e51b815260040180806020018281038252602a815260200180611b59602a913960400191505060405180910390fd5b60408051606084811b6001600160601b03199081166020808501919091529185901b16603483015282518083036028018152604890920183528151918101919091206000818152606b909252919020546001600160c01b0316806112835760405162461bcd60e51b8152600401808060200182810382526027815260200180611c996027913960400191505060405180910390fd5b6065546000838152606b60205260409020546001600160401b03918216600160c01b90910482160181164290911610156112ee5760405162461bcd60e51b8152600401808060200182810382526035815260200180611bfd6035913960400191505060405180910390fd5b6000828152606b60205260408120556001600160a01b0383166113235761131e84826001600160c01b0316611462565b611337565b6113378385836001600160c01b03166114f7565b836001600160a01b0316836001600160a01b03167f72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf32718360405180826001600160c01b0316815260200191505060405180910390a350506033805460ff191660011790555050565b6067546001600160a01b031681565b303b1590565b600054610100900460ff16806113cc57506113cc6113ad565b806113da575060005460ff16155b6114155760405162461bcd60e51b815260040180806020018281038252602e815260200180611ac1602e913960400191505060405180910390fd5b600054610100900460ff16158015611440576000805460ff1961ff0019909116610100171660011790555b6033805460ff19166001179055801561145f576000805461ff00191690555b50565b6040516000906001600160a01b0384169083908381818185875af1925050503d80600081146114ad576040519150601f19603f3d011682016040523d82523d6000602084013e6114b2565b606091505b50509050806114f25760405162461bcd60e51b81526004018080602001828103825260328152602001806119716032913960400191505060405180910390fd5b505050565b604080518082018252601981527f7472616e7366657228616464726573732c75696e74323536290000000000000060209182015281516001600160a01b0385811660248301526044808301869052845180840390910181526064909201845291810180516001600160e01b031663a9059cbb60e01b1781529251815160009460609489169392918291908083835b602083106115a45780518252601f199092019160209182019101611585565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d8060008114611606576040519150601f19603f3d011682016040523d82523d6000602084013e61160b565b606091505b5091509150818015611639575080511580611639575080806020019051602081101561163657600080fd5b50515b61076f5760405162461bcd60e51b815260040180806020018281038252603a815260200180611834603a913960400191505060405180910390fd5b60408051606085811b6001600160601b03199081166020808501919091529186901b16603483015282518083036028018152604890920183528151918101919091206000818152606b909252919020546001600160c01b0390811683810191821610156117125760405162461bcd60e51b8152600401808060200182810382526034815260200180611b836034913960400191505060405180910390fd5b6000828152606b602090815260409182902080546001600160401b03428116600160c01b9081026001600160c01b038089166001600160c01b03199095169490941784161793849055855192891683529092049091169181019190915281516001600160a01b0380881693908916927f41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd929081900390910190a35050505050565b60408051808201909152600080825260208201529056fe5769746864726177616c44656c617965723a3a6465706f7369743a2057524f4e475f414d4f554e545769746864726177616c44656c617965723a3a636c61696d456d657267656e6379436f756e63696c3a204f4e4c595f50454e44494e475f474f5645524e414e43455769746864726177616c44656c617965723a3a5f746f6b656e5769746864726177616c3a20544f4b454e5f5452414e534645525f4641494c45447472616e7366657246726f6d28616464726573732c616464726573732c75696e74323536295769746864726177616c44656c617965723a3a656e61626c65456d657267656e63794d6f64653a20414c52454144595f454e41424c45445769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204f4e4c595f454d4f44455769746864726177616c44656c617965723a3a6465706f7369743a2057524f4e475f544f4b454e5f414444524553535769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204e4f5f4d41585f454d455247454e43595f4d4f44455f54494d455769746864726177616c44656c617965723a3a5f6574685769746864726177616c3a205452414e534645525f4641494c45445769746864726177616c44656c617965723a3a636c61696d476f7665726e616e63653a204f4e4c595f50454e44494e475f474f5645524e414e43455769746864726177616c44656c617965723a3a7472616e73666572476f7665726e616e63653a204f4e4c595f474f5645524e414e43455769746864726177616c44656c617965723a3a6368616e67655769746864726177616c44656c61793a204f4e4c595f524f4c4c55505f4f525f474f5645524e414e43455769746864726177616c44656c617965723a3a7472616e73666572456d657267656e6379436f756e63696c3a204f4e4c595f454d455247454e43595f434f554e43494c5769746864726177616c44656c617965723a3a6465706f7369743a204f4e4c595f524f4c4c5550436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65645769746864726177616c44656c617965723a3a6465706f7369743a20544f4b454e5f5452414e534645525f4641494c45445769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204f4e4c595f474f5645524e414e43455769746864726177616c44656c617965723a3a6465706f7369743a20454d455247454e43595f4d4f44455769746864726177616c44656c617965723a3a5f70726f636573734465706f7369743a204445504f5349545f4f564552464c4f575769746864726177616c44656c617965723a3a6368616e67655769746864726177616c44656c61793a20455843454544535f4d41585f5749544844524157414c5f44454c41595769746864726177616c44656c617965723a3a7769746864726177616c3a205749544844524157414c5f4e4f545f414c4c4f5745445769746864726177616c44656c617965723a3a656e61626c65456d657267656e63794d6f64653a204f4e4c595f474f5645524e414e43455769746864726177616c44656c617965723a3a6465706f7369743a204e4f545f454e4f5547485f414c4c4f57414e43455769746864726177616c44656c617965723a3a7769746864726177616c3a204e4f5f46554e4453a2646970667358221220e5a3370e58aedbb9299b84dab9f46ede78dea7760c7d54540ca28b4126cc4f0b64736f6c634300060c0033"

// DeployWithdrawalDelayer deploys a new Ethereum contract, binding an instance of WithdrawalDelayer to it.
func DeployWithdrawalDelayer(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *WithdrawalDelayer, error) {
	parsed, err := abi.JSON(strings.NewReader(WithdrawalDelayerABI))
	if err != nil {
		return common.Address{}, nil, nil, tracerr.Wrap(err)
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(WithdrawalDelayerBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayer{WithdrawalDelayerCaller: WithdrawalDelayerCaller{contract: contract}, WithdrawalDelayerTransactor: WithdrawalDelayerTransactor{contract: contract}, WithdrawalDelayerFilterer: WithdrawalDelayerFilterer{contract: contract}}, nil
}

// NewWithdrawalDelayerCaller creates a new read-only instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerCaller(address common.Address, caller bind.ContractCaller) (*WithdrawalDelayerCaller, error) {
	contract, err := bindWithdrawalDelayer(address, caller, nil, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerCaller{contract: contract}, nil
}

// NewWithdrawalDelayerTransactor creates a new write-only instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerTransactor(address common.Address, transactor bind.ContractTransactor) (*WithdrawalDelayerTransactor, error) {
	contract, err := bindWithdrawalDelayer(address, nil, transactor, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerTransactor{contract: contract}, nil
}

// NewWithdrawalDelayerFilterer creates a new log filterer instance of WithdrawalDelayer, bound to a specific deployed contract.
func NewWithdrawalDelayerFilterer(address common.Address, filterer bind.ContractFilterer) (*WithdrawalDelayerFilterer, error) {
	contract, err := bindWithdrawalDelayer(address, nil, nil, filterer)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerFilterer{contract: contract}, nil
}

// bindWithdrawalDelayer binds a generic wrapper to an already deployed contract.
func bindWithdrawalDelayer(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(WithdrawalDelayerABI))
	if err != nil {
		return nil, tracerr.Wrap(err)
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
	return *ret0, tracerr.Wrap(err)
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
	return *ret0, tracerr.Wrap(err)
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
	return *ret0, *ret1, tracerr.Wrap(err)
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
	return *ret, tracerr.Wrap(err)
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
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getEmergencyCouncil")
	return *ret0, tracerr.Wrap(err)
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
// Solidity: function getEmergencyModeStartingTime() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetEmergencyModeStartingTime(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getEmergencyModeStartingTime")
	return *ret0, tracerr.Wrap(err)
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

// GetHermezGovernanceAddress is a free data retrieval call binding the contract method 0x0b21d430.
//
// Solidity: function getHermezGovernanceAddress() view returns(address)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetHermezGovernanceAddress(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getHermezGovernanceAddress")
	return *ret0, tracerr.Wrap(err)
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
// Solidity: function getWithdrawalDelay() view returns(uint128)
func (_WithdrawalDelayer *WithdrawalDelayerCaller) GetWithdrawalDelay(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "getWithdrawalDelay")
	return *ret0, tracerr.Wrap(err)
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
	return *ret0, tracerr.Wrap(err)
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
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "pendingEmergencyCouncil")
	return *ret0, tracerr.Wrap(err)
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
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _WithdrawalDelayer.contract.Call(opts, out, "pendingGovernance")
	return *ret0, tracerr.Wrap(err)
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

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x42cb7216.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezGovernanceAddress, address _initialEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) WithdrawalDelayerInitializer(opts *bind.TransactOpts, _initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezGovernanceAddress common.Address, _initialEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "withdrawalDelayerInitializer", _initialWithdrawalDelay, _initialHermezRollup, _initialHermezGovernanceAddress, _initialEmergencyCouncil)
}

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x42cb7216.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezGovernanceAddress, address _initialEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) WithdrawalDelayerInitializer(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezGovernanceAddress common.Address, _initialEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerInitializer(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezGovernanceAddress, _initialEmergencyCouncil)
}

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x42cb7216.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezGovernanceAddress, address _initialEmergencyCouncil) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) WithdrawalDelayerInitializer(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezGovernanceAddress common.Address, _initialEmergencyCouncil common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerInitializer(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezGovernanceAddress, _initialEmergencyCouncil)
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerDeposit)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
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

// ParseDeposit is a log parse operation binding the contract event 0x41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd.
//
// Solidity: event Deposit(address indexed owner, address indexed token, uint192 amount, uint64 depositTimestamp)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseDeposit(log types.Log) (*WithdrawalDelayerDeposit, error) {
	event := new(WithdrawalDelayerDeposit)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "Deposit", log); err != nil {
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerEmergencyModeEnabledIterator{contract: _WithdrawalDelayer.contract, event: "EmergencyModeEnabled", logs: logs, sub: sub}, nil
}

// WatchEmergencyModeEnabled is a free log subscription operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchEmergencyModeEnabled(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerEmergencyModeEnabled) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "EmergencyModeEnabled")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerEmergencyModeEnabled)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
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

// ParseEmergencyModeEnabled is a log parse operation binding the contract event 0x2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c5.
//
// Solidity: event EmergencyModeEnabled()
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseEmergencyModeEnabled(log types.Log) (*WithdrawalDelayerEmergencyModeEnabled, error) {
	event := new(WithdrawalDelayerEmergencyModeEnabled)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "EmergencyModeEnabled", log); err != nil {
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerEscapeHatchWithdrawal)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
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

// ParseEscapeHatchWithdrawal is a log parse operation binding the contract event 0xde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693.
//
// Solidity: event EscapeHatchWithdrawal(address indexed who, address indexed to, address indexed token, uint256 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseEscapeHatchWithdrawal(log types.Log) (*WithdrawalDelayerEscapeHatchWithdrawal, error) {
	event := new(WithdrawalDelayerEscapeHatchWithdrawal)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "EscapeHatchWithdrawal", log); err != nil {
		return nil, tracerr.Wrap(err)
	}
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
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerNewEmergencyCouncilIterator{contract: _WithdrawalDelayer.contract, event: "NewEmergencyCouncil", logs: logs, sub: sub}, nil
}

// WatchNewEmergencyCouncil is a free log subscription operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewEmergencyCouncil(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewEmergencyCouncil) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewEmergencyCouncil")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewEmergencyCouncil)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
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

// ParseNewEmergencyCouncil is a log parse operation binding the contract event 0xcc267667d474ef34ee2de2d060e7c8b2c7295cefa22e57fd7049e22b5fdb5396.
//
// Solidity: event NewEmergencyCouncil(address newEmergencyCouncil)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewEmergencyCouncil(log types.Log) (*WithdrawalDelayerNewEmergencyCouncil, error) {
	event := new(WithdrawalDelayerNewEmergencyCouncil)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewEmergencyCouncil", log); err != nil {
		return nil, tracerr.Wrap(err)
	}
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
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerNewHermezGovernanceAddressIterator{contract: _WithdrawalDelayer.contract, event: "NewHermezGovernanceAddress", logs: logs, sub: sub}, nil
}

// WatchNewHermezGovernanceAddress is a free log subscription operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewHermezGovernanceAddress(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewHermezGovernanceAddress) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewHermezGovernanceAddress")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewHermezGovernanceAddress)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
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

// ParseNewHermezGovernanceAddress is a log parse operation binding the contract event 0x3bf02437d5cd40067085d9dac2c3cdcbef0a449d98a259a40d9c24380aca81bf.
//
// Solidity: event NewHermezGovernanceAddress(address newHermezGovernanceAddress)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewHermezGovernanceAddress(log types.Log) (*WithdrawalDelayerNewHermezGovernanceAddress, error) {
	event := new(WithdrawalDelayerNewHermezGovernanceAddress)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewHermezGovernanceAddress", log); err != nil {
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return &WithdrawalDelayerNewWithdrawalDelayIterator{contract: _WithdrawalDelayer.contract, event: "NewWithdrawalDelay", logs: logs, sub: sub}, nil
}

// WatchNewWithdrawalDelay is a free log subscription operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) WatchNewWithdrawalDelay(opts *bind.WatchOpts, sink chan<- *WithdrawalDelayerNewWithdrawalDelay) (event.Subscription, error) {

	logs, sub, err := _WithdrawalDelayer.contract.WatchLogs(opts, "NewWithdrawalDelay")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerNewWithdrawalDelay)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
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

// ParseNewWithdrawalDelay is a log parse operation binding the contract event 0x6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7.
//
// Solidity: event NewWithdrawalDelay(uint64 withdrawalDelay)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseNewWithdrawalDelay(log types.Log) (*WithdrawalDelayerNewWithdrawalDelay, error) {
	event := new(WithdrawalDelayerNewWithdrawalDelay)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "NewWithdrawalDelay", log); err != nil {
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(WithdrawalDelayerWithdraw)
				if err := _WithdrawalDelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
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

// ParseWithdraw is a log parse operation binding the contract event 0x72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf3271.
//
// Solidity: event Withdraw(address indexed token, address indexed owner, uint192 amount)
func (_WithdrawalDelayer *WithdrawalDelayerFilterer) ParseWithdraw(log types.Log) (*WithdrawalDelayerWithdraw, error) {
	event := new(WithdrawalDelayerWithdraw)
	if err := _WithdrawalDelayer.contract.UnpackLog(event, "Withdraw", log); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return event, nil
}
