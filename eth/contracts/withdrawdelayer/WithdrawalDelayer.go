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
const WithdrawalDelayerABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"name\":\"Deposit\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[],\"name\":\"EmergencyModeEnabled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"EscapeHatchWithdrawal\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezGovernanceDAOAddress\",\"type\":\"address\"}],\"name\":\"NewHermezGovernanceDAOAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newHermezKeeperAddress\",\"type\":\"address\"}],\"name\":\"NewHermezKeeperAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newWhiteHackGroupAddress\",\"type\":\"address\"}],\"name\":\"NewWhiteHackGroupAddress\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"withdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"NewWithdrawalDelay\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"}],\"name\":\"Withdraw\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_EMERGENCY_MODE_TIME\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_WITHDRAWAL_DELAY\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_newWithdrawalDelay\",\"type\":\"uint64\"}],\"name\":\"changeWithdrawalDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint192\",\"name\":\"_amount\",\"type\":\"uint192\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"depositInfo\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"deposits\",\"outputs\":[{\"internalType\":\"uint192\",\"name\":\"amount\",\"type\":\"uint192\"},{\"internalType\":\"uint64\",\"name\":\"depositTimestamp\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"enableEmergencyMode\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"escapeHatchWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getEmergencyModeStartingTime\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezGovernanceDAOAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getHermezKeeperAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWhiteHackGroupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getWithdrawalDelay\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"hermezRollupAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isEmergencyMode\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setHermezGovernanceDAOAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setHermezKeeperAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"setWhiteHackGroupAddress\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"withdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_initialWithdrawalDelay\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"_initialHermezRollup\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezKeeperAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_initialHermezGovernanceDAOAddress\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_initialWhiteHackGroupAddress\",\"type\":\"address\"}],\"name\":\"withdrawalDelayerInitializer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// WithdrawalDelayerBin is the compiled bytecode used for deploying new contracts.
var WithdrawalDelayerBin = "0x608060405234801561001057600080fd5b50611be6806100206000396000f3fe60806040526004361061011f5760003560e01c8063668cdd67116100a0578063b4b8e39d11610064578063b4b8e39d14610405578063c5b1c7d01461041a578063cfc0b6411461042f578063d82b217c1461046f578063de35f282146104a25761011f565b8063668cdd67146103345780637fd6b10214610349578063a238f9df1461038c578063acfd6ea8146103bd578063ae7efbbd146103f05761011f565b8063305887f9116100e7578063305887f91461022057806336e566ed146102355780633d4dff7b1461028f578063493b0170146102e4578063580fc6111461031f5761011f565b806303160940146101245780630a4db01b1461015e5780630e670af5146101935780630fd266d7146101c657806320a194b8146101f7575b600080fd5b34801561013057600080fd5b506101396104dd565b604080516fffffffffffffffffffffffffffffffff9092168252519081900360200190f35b34801561016a57600080fd5b506101916004803603602081101561018157600080fd5b50356001600160a01b03166104ec565b005b34801561019f57600080fd5b50610191600480360360208110156101b657600080fd5b50356001600160401b031661058f565b3480156101d257600080fd5b506101db6106bf565b604080516001600160a01b039092168252519081900360200190f35b34801561020357600080fd5b5061020c6106ce565b604080519115158252519081900360200190f35b34801561022c57600080fd5b506101db6106de565b34801561024157600080fd5b50610191600480360360a081101561025857600080fd5b506001600160401b03813516906001600160a01b0360208201358116916040810135821691606082013581169160800135166106ed565b34801561029b57600080fd5b506102b9600480360360208110156102b257600080fd5b5035610809565b604080516001600160c01b0390931683526001600160401b0390911660208301528051918290030190f35b3480156102f057600080fd5b506102b96004803603604081101561030757600080fd5b506001600160a01b0381358116916020013516610836565b34801561032b57600080fd5b506101db6108c7565b34801561034057600080fd5b506101396108d6565b34801561035557600080fd5b506101916004803603606081101561036c57600080fd5b506001600160a01b038135811691602081013590911690604001356108ec565b34801561039857600080fd5b506103a1610af2565b604080516001600160401b039092168252519081900360200190f35b3480156103c957600080fd5b50610191600480360360208110156103e057600080fd5b50356001600160a01b0316610af9565b3480156103fc57600080fd5b506101db610b9c565b34801561041157600080fd5b506103a1610bab565b34801561042657600080fd5b50610191610bb2565b6101916004803603606081101561044557600080fd5b5080356001600160a01b0390811691602081013590911690604001356001600160c01b0316610cae565b34801561047b57600080fd5b506101916004803603602081101561049257600080fd5b50356001600160a01b0316611060565b3480156104ae57600080fd5b50610191600480360360408110156104c557600080fd5b506001600160a01b0381358116916020013516611103565b6065546001600160401b031690565b6067546001600160a01b031633146105355760405162461bcd60e51b815260040180806020018281038252603a815260200180611b20603a913960400191505060405180910390fd5b606780546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517f284ca073b8bdde2195ae98779277678773a99d7739e5f0477dc19a03fc689011916020908290030190a150565b6068546001600160a01b03163314806105b257506069546001600160a01b031633145b610603576040805162461bcd60e51b815260206004820152601c60248201527f4f6e6c79206865726d657a206b6565706572206f7220726f6c6c757000000000604482015290519081900360640190fd5b621275006001600160401b0382161115610664576040805162461bcd60e51b815260206004820152601c60248201527f45786365656473204d41585f5749544844524157414c5f44454c415900000000604482015290519081900360640190fd5b6065805467ffffffffffffffff19166001600160401b03838116919091179182905560408051929091168252517f6b3670ab51e04a9da086741e5fd1eb36ffaf1d661a15330c528e1f3e0c8722d7916020908290030190a150565b6069546001600160a01b031681565b606854600160a01b900460ff1690565b6068546001600160a01b031690565b600054610100900460ff1680610706575061070661135d565b80610714575060005460ff16155b61074f5760405162461bcd60e51b815260040180806020018281038252602e8152602001806119fb602e913960400191505060405180910390fd5b600054610100900460ff1615801561077a576000805460ff1961ff0019909116610100171660011790555b610782611363565b6065805467ffffffffffffffff19166001600160401b038816179055606980546001600160a01b03199081166001600160a01b0388811691909117909255606880546066805484168886161790556067805484168786161790559091169186169190911760ff60a01b191690558015610801576000805461ff00191690555b505050505050565b606a602052600090815260409020546001600160c01b03811690600160c01b90046001600160401b031682565b60008061084161176a565b505060408051606094851b6001600160601b03199081166020808401919091529490951b909416603485015280518085036028018152604885018083528151918501919091206000908152606a9094529281902060888501909152546001600160c01b03811692839052600160c01b90046001600160401b031660689093018390525091565b6066546001600160a01b031690565b606554600160401b90046001600160401b031690565b60335460ff16610943576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff19169055606854600160a01b900460ff166109955760405162461bcd60e51b81526004018080602001828103825260348152602001806118816034913960400191505060405180910390fd5b6067546001600160a01b03163314806109b857506066546001600160a01b031633145b6109f35760405162461bcd60e51b815260040180806020018281038252603d815260200180611997603d913960400191505060405180910390fd5b6067546001600160a01b0316331415610a63576065546001600160401b03600160401b909104811662eff100018116429091161015610a635760405162461bcd60e51b81526004018080602001828103825260448152602001806118e46044913960600191505060405180910390fd5b6001600160a01b038216610a8057610a7b8382611412565b610a8b565b610a8b8284836114a7565b816001600160a01b0316836001600160a01b0316336001600160a01b03167fde200220117ba95c9a6c4a1a13bb06b0b7be90faa85c8fb4576630119f891693846040518082815260200191505060405180910390a450506033805460ff1916600117905550565b6212750081565b6066546001600160a01b03163314610b425760405162461bcd60e51b81526004018080602001828103825260418152602001806117aa6041913960600191505060405180910390fd5b606680546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517f03683be8debd93f8f5ff23dd03419bfcb9b8287a1868b0f130d858f03c3a08a1916020908290030190a150565b6067546001600160a01b031690565b62eff10081565b6068546001600160a01b03163314610bfb5760405162461bcd60e51b8152600401808060200182810382526033815260200180611aed6033913960400191505060405180910390fd5b606854600160a01b900460ff1615610c445760405162461bcd60e51b815260040180806020018281038252603781526020018061184a6037913960400191505060405180910390fd5b6068805460ff60a01b1916600160a01b179055606580546001600160401b034216600160401b026fffffffffffffffff0000000000000000199091161790556040517f2064d51aa5a8bd67928c7675e267e05c67ad5adf7c9098d0a602d01f36fda9c590600090a1565b60335460ff16610d05576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff191690556069546001600160a01b03163314610d585760405162461bcd60e51b81526004018080602001828103825260278152602001806119d46027913960400191505060405180910390fd5b3415610df0576001600160a01b03821615610da45760405162461bcd60e51b815260040180806020018281038252602f8152602001806118b5602f913960400191505060405180910390fd5b34816001600160c01b031614610deb5760405162461bcd60e51b81526004018080602001828103825260288152602001806117826028913960400191505060405180910390fd5b611043565b60695460408051636eb1769f60e11b81526001600160a01b03928316600482015230602482015290516001600160c01b0384169285169163dd62ed3e916044808301926020929190829003018186803b158015610e4c57600080fd5b505afa158015610e60573d6000803e3d6000fd5b505050506040513d6020811015610e7657600080fd5b50511015610eb55760405162461bcd60e51b8152600401808060200182810382526030815260200180611b5a6030913960400191505060405180910390fd5b60006060836001600160a01b0316604051806060016040528060258152602001611825602591398051602091820120606954604080516001600160a01b0390921660248301523060448301526001600160c01b038816606480840191909152815180840390910181526084909201815292810180516001600160e01b03166001600160e01b031990931692909217825291518251909182918083835b60208310610f705780518252601f199092019160209182019101610f51565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d8060008114610fd2576040519150601f19603f3d011682016040523d82523d6000602084013e610fd7565b606091505b5091509150818015611005575080511580611005575080806020019051602081101561100257600080fd5b50515b6110405760405162461bcd60e51b8152600401808060200182810382526031815260200180611a296031913960400191505060405180910390fd5b50505b61104e83838361162b565b50506033805460ff1916600117905550565b6068546001600160a01b031633146110a95760405162461bcd60e51b815260040180806020018281038252603d815260200180611928603d913960400191505060405180910390fd5b606880546001600160a01b0319166001600160a01b03838116919091179182905560408051929091168252517fc1e9be84fce652abec6a6944f7ec5bbb40de18caa44c285b05a0de7e3ad9d016916020908290030190a150565b60335460ff1661115a576040805162461bcd60e51b815260206004820152601f60248201527f5265656e7472616e637947756172643a207265656e7472616e742063616c6c00604482015290519081900360640190fd5b6033805460ff19169055606854600160a01b900460ff16156111ad5760405162461bcd60e51b815260040180806020018281038252602a815260200180611a5a602a913960400191505060405180910390fd5b60408051606084811b6001600160601b03199081166020808501919091529185901b16603483015282518083036028018152604890920183528151918101919091206000818152606a909252919020546001600160c01b0316806112425760405162461bcd60e51b8152600401808060200182810382526027815260200180611b8a6027913960400191505060405180910390fd5b6065546000838152606a60205260409020546001600160401b03918216600160c01b90910482160181164290911610156112ad5760405162461bcd60e51b8152600401808060200182810382526035815260200180611ab86035913960400191505060405180910390fd5b6000828152606a60205260408120556001600160a01b0383166112e2576112dd84826001600160c01b0316611412565b6112f6565b6112f68385836001600160c01b03166114a7565b836001600160a01b0316836001600160a01b03167f72608e45b52a95a12c2ac7f15ff53f92fc9572c9d84b6e6b5d7f0f7826cf32718360405180826001600160c01b0316815260200191505060405180910390a350506033805460ff191660011790555050565b303b1590565b600054610100900460ff168061137c575061137c61135d565b8061138a575060005460ff16155b6113c55760405162461bcd60e51b815260040180806020018281038252602e8152602001806119fb602e913960400191505060405180910390fd5b600054610100900460ff161580156113f0576000805460ff1961ff0019909116610100171660011790555b6033805460ff19166001179055801561140f576000805461ff00191690555b50565b6040516000906001600160a01b0384169083908381818185875af1925050503d806000811461145d576040519150601f19603f3d011682016040523d82523d6000602084013e611462565b606091505b50509050806114a25760405162461bcd60e51b81526004018080602001828103825260328152602001806119656032913960400191505060405180910390fd5b505050565b604080518082018252601981527f7472616e7366657228616464726573732c75696e74323536290000000000000060209182015281516001600160a01b0385811660248301526044808301869052845180840390910181526064909201845291810180516001600160e01b031663a9059cbb60e01b1781529251815160009460609489169392918291908083835b602083106115545780518252601f199092019160209182019101611535565b6001836020036101000a0380198251168184511680821785525050505050509050019150506000604051808303816000865af19150503d80600081146115b6576040519150601f19603f3d011682016040523d82523d6000602084013e6115bb565b606091505b50915091508180156115e95750805115806115e957508080602001905160208110156115e657600080fd5b50515b6116245760405162461bcd60e51b815260040180806020018281038252603a8152602001806117eb603a913960400191505060405180910390fd5b5050505050565b60408051606085811b6001600160601b03199081166020808501919091529186901b16603483015282518083036028018152604890920183528151918101919091206000818152606a909252919020546001600160c01b0390811683810191821610156116c95760405162461bcd60e51b8152600401808060200182810382526034815260200180611a846034913960400191505060405180910390fd5b6000828152606a602090815260409182902080546001600160401b03428116600160c01b9081026001600160c01b038089166001600160c01b03199095169490941784161793849055855192891683529092049091169181019190915281516001600160a01b0380881693908916927f41219b99485f78192a5b9b1be28c7d53c3a2bdbe7900ae40c79fae8d9d6108fd929081900390910190a35050505050565b60408051808201909152600080825260208201529056fe5769746864726177616c44656c617965723a3a6465706f7369743a2057524f4e475f414d4f554e545769746864726177616c44656c617965723a3a7365744865726d657a476f7665726e616e636544414f416464726573733a204f4e4c595f474f5645524e414e43455769746864726177616c44656c617965723a3a5f746f6b656e5769746864726177616c3a20544f4b454e5f5452414e534645525f4641494c45447472616e7366657246726f6d28616464726573732c616464726573732c75696e74323536295769746864726177616c44656c617965723a3a656e61626c65456d657267656e63794d6f64653a20414c52454144595f454e41424c45445769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204f4e4c595f454d4f44455769746864726177616c44656c617965723a3a6465706f7369743a2057524f4e475f544f4b454e5f414444524553535769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204e4f5f4d41585f454d455247454e43595f4d4f44455f54494d455769746864726177616c44656c617965723a3a7365744865726d657a476f7665726e616e636544414f416464726573733a204f4e4c595f4b45455045525769746864726177616c44656c617965723a3a5f6574685769746864726177616c3a205452414e534645525f4641494c45445769746864726177616c44656c617965723a3a65736361706548617463685769746864726177616c3a204f4e4c595f474f5645524e414e43455f5748475769746864726177616c44656c617965723a3a6465706f7369743a204f4e4c595f524f4c4c5550436f6e747261637420696e7374616e63652068617320616c7265616479206265656e20696e697469616c697a65645769746864726177616c44656c617965723a3a6465706f7369743a20544f4b454e5f5452414e534645525f4641494c45445769746864726177616c44656c617965723a3a6465706f7369743a20454d455247454e43595f4d4f44455769746864726177616c44656c617965723a3a5f70726f636573734465706f7369743a204445504f5349545f4f564552464c4f575769746864726177616c44656c617965723a3a7769746864726177616c3a205749544844524157414c5f4e4f545f414c4c4f5745445769746864726177616c44656c617965723a3a656e61626c65456d657267656e63794d6f64653a204f4e4c595f4b45455045525769746864726177616c44656c617965723a3a7365744865726d657a476f7665726e616e636544414f416464726573733a204f4e4c595f5748475769746864726177616c44656c617965723a3a6465706f7369743a204e4f545f454e4f5547485f414c4c4f57414e43455769746864726177616c44656c617965723a3a7769746864726177616c3a204e4f5f46554e4453a2646970667358221220dc3cbf4fcef393039db1a6c32d7d5ec23383ea60dc16dd54e40defffc4d4d55564736f6c634300060c0033"

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

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x36e566ed.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactor) WithdrawalDelayerInitializer(opts *bind.TransactOpts, _initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.contract.Transact(opts, "withdrawalDelayerInitializer", _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
}

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x36e566ed.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerSession) WithdrawalDelayerInitializer(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerInitializer(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
}

// WithdrawalDelayerInitializer is a paid mutator transaction binding the contract method 0x36e566ed.
//
// Solidity: function withdrawalDelayerInitializer(uint64 _initialWithdrawalDelay, address _initialHermezRollup, address _initialHermezKeeperAddress, address _initialHermezGovernanceDAOAddress, address _initialWhiteHackGroupAddress) returns()
func (_WithdrawalDelayer *WithdrawalDelayerTransactorSession) WithdrawalDelayerInitializer(_initialWithdrawalDelay uint64, _initialHermezRollup common.Address, _initialHermezKeeperAddress common.Address, _initialHermezGovernanceDAOAddress common.Address, _initialWhiteHackGroupAddress common.Address) (*types.Transaction, error) {
	return _WithdrawalDelayer.Contract.WithdrawalDelayerInitializer(&_WithdrawalDelayer.TransactOpts, _initialWithdrawalDelay, _initialHermezRollup, _initialHermezKeeperAddress, _initialHermezGovernanceDAOAddress, _initialWhiteHackGroupAddress)
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
