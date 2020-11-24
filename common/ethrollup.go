package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// RollupVars contain the Rollup smart contract variables
// type RollupVars struct {
// 	EthBlockNum    uint64
// 	ForgeL1Timeout *big.Int
// 	FeeL1UserTx    *big.Int
// 	FeeAddToken    *big.Int
// 	TokensHEZ      eth.Address
// 	Governance     eth.Address
// }

// AuctionVars contain the Auction smart contract variables
// type AuctionVars struct {
// 	EthBlockNum       uint64
// 	SlotDeadline      uint
// 	CloseAuctionSlots uint
// 	OpenAuctionSlots  uint
// 	Governance        eth.Address
// 	MinBidSlots       MinBidSlots
// 	Outbidding        int
// 	DonationAddress   eth.Address
// 	GovernanceAddress eth.Address
// 	AllocationRatio   AllocationRatio
// }

// WithdrawDelayerVars contains the Withdrawal Delayer smart contract variables
// type WithdrawDelayerVars struct {
// 	HermezRollupAddress        eth.Address
// 	HermezGovernanceDAOAddress eth.Address
// 	WhiteHackGroupAddress      eth.Address
// 	WithdrawalDelay            uint
// 	EmergencyModeStartingTime  time.Time
// 	EmergencyModeEnabled       bool
// }

// MinBidSlots TODO
// type MinBidSlots [6]uint
//
// // AllocationRatio TODO
// type AllocationRatio struct {
// 	Donation uint
// 	Burn     uint
// 	Forger   uint
// }

const (
	// RollupConstMaxFeeIdxCoordinator is the maximum number of tokens the
	// coordinator can use to collect fees (determines the number of tokens
	// that the coordinator can collect fees from).  This value is
	// determined by the circuit.
	RollupConstMaxFeeIdxCoordinator = 64
	// RollupConstReservedIDx First 256 indexes reserved, first user index will be the 256
	RollupConstReservedIDx = 255
	// RollupConstExitIDx IDX 1 is reserved for exits
	RollupConstExitIDx = 1
	// RollupConstLimitTokens Max number of tokens allowed to be registered inside the rollup
	RollupConstLimitTokens = (1 << 32)
	// RollupConstL1CoordinatorTotalBytes [4 bytes] token + [32 bytes] babyjub + [65 bytes] compressedSignature
	RollupConstL1CoordinatorTotalBytes = 101
	// RollupConstL1UserTotalBytes [20 bytes] fromEthAddr + [32 bytes] fromBjj-compressed + [6 bytes] fromIdx +
	// [2 bytes] loadAmountFloat16 + [2 bytes] amountFloat16 + [4 bytes] tokenId + [6 bytes] toIdx
	RollupConstL1UserTotalBytes = 72
	// RollupConstMaxL1UserTx Maximum L1-user transactions allowed to be queued in a batch
	RollupConstMaxL1UserTx = 128
	// RollupConstMaxL1Tx Maximum L1 transactions allowed to be queued in a batch
	RollupConstMaxL1Tx = 256
	// RollupConstInputSHAConstantBytes [6 bytes] lastIdx + [6 bytes] newLastIdx  + [32 bytes] stateRoot  + [32 bytes] newStRoot  + [32 bytes] newExitRoot +
	// [_MAX_L1_TX * _L1_USER_TOTALBYTES bytes] l1TxsData + totalL2TxsDataLength + feeIdxCoordinatorLength + [2 bytes] chainID =
	// 18542 bytes +  totalL2TxsDataLength + feeIdxCoordinatorLength
	RollupConstInputSHAConstantBytes = 18542
	// RollupConstNumBuckets Number of buckets
	RollupConstNumBuckets = 5
	// RollupConstMaxWithdrawalDelay max withdrawal delay in seconds
	RollupConstMaxWithdrawalDelay = 2 * 7 * 24 * 60 * 60
	// RollupConstExchangeMultiplier exchange multiplier
	RollupConstExchangeMultiplier = 1e14
	// LenVerifiers number of Rollup Smart Contract Verifiers
	LenVerifiers = 1
)

var (
	// RollupConstLimitLoadAmount Max load amount allowed (loadAmount: L1 --> L2)
	RollupConstLimitLoadAmount, _ = new(big.Int).SetString("340282366920938463463374607431768211456", 10)
	// RollupConstLimitL2TransferAmount Max amount allowed (amount L2 --> L2)
	RollupConstLimitL2TransferAmount, _ = new(big.Int).SetString("6277101735386680763835789423207666416102355444464034512896", 10)

	// RollupConstEthAddressInternalOnly This ethereum address is used internally for rollup accounts that don't have ethereum address, only Babyjubjub
	// This non-ethereum accounts can be created by the coordinator and allow users to have a rollup
	// account without needing an ethereum address
	RollupConstEthAddressInternalOnly = ethCommon.HexToAddress("0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF")
	// RollupConstRfield Modulus zkSNARK
	RollupConstRfield, _ = new(big.Int).SetString(
		"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	// RollupConstERC1820 ERC1820Registry address
	RollupConstERC1820 = ethCommon.HexToAddress("0x1820a4B7618BdE71Dce8cdc73aAB6C95905faD24")

	// ERC777 tokens signatures

	// RollupConstRecipientInterfaceHash ERC777 recipient interface hash
	RollupConstRecipientInterfaceHash = crypto.Keccak256([]byte("ERC777TokensRecipient"))
	// RollupConstPerformL1UserTxSignature the signature of the function that can be called thru an ERC777 `send`
	RollupConstPerformL1UserTxSignature = crypto.Keccak256([]byte("addL1Transaction(uint256,uint48,uint16,uint16,uint32,uint48)"))
	// RollupConstAddTokenSignature the signature of the function that can be called thru an ERC777 `send`
	RollupConstAddTokenSignature = crypto.Keccak256([]byte("addToken(address)"))
	// RollupConstSendSignature ERC777 Signature
	RollupConstSendSignature = crypto.Keccak256([]byte("send(address,uint256,bytes)"))
	// RollupConstERC777Granularity ERC777 Signature
	RollupConstERC777Granularity = crypto.Keccak256([]byte("granularity()"))
	// RollupConstWithdrawalDelayerDeposit  This constant are used to deposit tokens from ERC77 tokens into withdrawal delayer
	RollupConstWithdrawalDelayerDeposit = crypto.Keccak256([]byte("deposit(address,address,uint192)"))

	// ERC20 signature

	// RollupConstTransferSignature This constant is used in the _safeTransfer internal method in order to safe GAS.
	RollupConstTransferSignature = crypto.Keccak256([]byte("transfer(address,uint256)"))
	// RollupConstTransferFromSignature This constant is used in the _safeTransfer internal method in order to safe GAS.
	RollupConstTransferFromSignature = crypto.Keccak256([]byte("transferFrom(address,address,uint256)"))
	// RollupConstApproveSignature This constant is used in the _safeTransfer internal method in order to safe GAS.
	RollupConstApproveSignature = crypto.Keccak256([]byte("approve(address,uint256)"))
	// RollupConstERC20Signature ERC20 decimals signature
	RollupConstERC20Signature = crypto.Keccak256([]byte("decimals()"))
)

// RollupVerifierStruct is the information about verifiers of the Rollup Smart Contract
type RollupVerifierStruct struct {
	MaxTx   int64 `json:"maxTx"`
	NLevels int64 `json:"nlevels"`
}

// RollupConstants are the constants of the Rollup Smart Contract
type RollupConstants struct {
	AbsoluteMaxL1L2BatchTimeout int64                  `json:"absoluteMaxL1L2BatchTimeout"`
	TokenHEZ                    ethCommon.Address      `json:"tokenHEZ"`
	Verifiers                   []RollupVerifierStruct `json:"verifiers"`
	HermezAuctionContract       ethCommon.Address      `json:"hermezAuctionContract"`
	HermezGovernanceDAOAddress  ethCommon.Address      `json:"hermezGovernanceDAOAddress"`
	SafetyAddress               ethCommon.Address      `json:"safetyAddress"`
	WithdrawDelayerContract     ethCommon.Address      `json:"withdrawDelayerContract"`
}

// Bucket are the variables of each Bucket of Rollup Smart Contract
type Bucket struct {
	CeilUSD             uint64 `json:"ceilUSD"`
	BlockStamp          uint64 `json:"blockStamp"`
	Withdrawals         uint64 `json:"withdrawals"`
	BlockWithdrawalRate uint64 `json:"blockWithdrawalRate"`
	MaxWithdrawals      uint64 `json:"maxWithdrawals"`
}

// RollupVariables are the variables of the Rollup Smart Contract
type RollupVariables struct {
	EthBlockNum           int64                         `json:"ethereumBlockNum" meddler:"eth_block_num"`
	FeeAddToken           *big.Int                      `json:"feeAddToken" meddler:"fee_add_token,bigint" validate:"required"`
	ForgeL1L2BatchTimeout int64                         `json:"forgeL1L2BatchTimeout" meddler:"forge_l1_timeout" validate:"required"`
	WithdrawalDelay       uint64                        `json:"withdrawalDelay" meddler:"withdrawal_delay" validate:"required"`
	Buckets               [RollupConstNumBuckets]Bucket `json:"buckets" meddler:"buckets,json"`
}

// Copy returns a deep copy of the Variables
func (v *RollupVariables) Copy() *RollupVariables {
	vCpy := *v
	return &vCpy
}
