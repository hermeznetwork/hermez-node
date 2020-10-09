package eth

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	Hermez "github.com/hermeznetwork/hermez-node/eth/contracts/hermez"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// RollupConstFeeIdxCoordinatorLen is the number of tokens the coordinator can use
	// to collect fees (determines the number of tokens that the
	// coordinator can collect fees from).  This value is determined by the
	// circuit.
	RollupConstFeeIdxCoordinatorLen = 64
	// RollupConstReservedIDx First 256 indexes reserved, first user index will be the 256
	RollupConstReservedIDx = 255
	// RollupConstExitIDx IDX 1 is reserved for exits
	RollupConstExitIDx = 1
	// RollupConstLimitLoadAmount Max load amount allowed (loadAmount: L1 --> L2)
	RollupConstLimitLoadAmount = (1 << 128)
	// RollupConstLimitL2TransferAmount Max amount allowed (amount L2 --> L2)
	RollupConstLimitL2TransferAmount = (1 << 192)
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

// RollupConstants are the constants of the Rollup Smart Contract
/* type RollupConstants struct {
	// Maxim Deposit allowed
	MaxAmountDeposit *big.Int
	MaxAmountL2      *big.Int
	MaxTokens        int64
	// maximum L1 transactions allowed to be queued for a batch
	MaxL1Tx int
	// maximum L1 user transactions allowed to be queued for a batch
	MaxL1UserTx        int
	Rfield             *big.Int
	L1CoordinatorBytes int
	L1UserBytes        int
	L2Bytes            int
	MaxTxVerifiers     []int
	TokenHEZ           ethCommon.Address
	// Only test
	GovernanceAddress ethCommon.Address
	// Only test
	SafetyBot ethCommon.Address
	// Only test
	ConsensusContract ethCommon.Address
	// Only test
	WithdrawalContract ethCommon.Address
	ReservedIDx        uint32
	LastIDx            uint32
	ExitIDx            uint32
	NoLimitToken       int
	NumBuckets         int
	MaxWDelay          int64
}*/

// RollupPublicConstants are the constants of the Rollup Smart Contract
type RollupPublicConstants struct {
	AbsoluteMaxL1L2BatchTimeout int64
	TokenHEZ                    ethCommon.Address
	Verifiers                   []RollupVerifierStruct
	HermezAuctionContract       ethCommon.Address
	HermezGovernanceDAOAddress  ethCommon.Address
	SafetyAddress               ethCommon.Address
	WithdrawDelayerContract     ethCommon.Address
}

// RollupVariables are the variables of the Rollup Smart Contract
type RollupVariables struct {
	FeeAddToken           *big.Int
	ForgeL1L2BatchTimeout int64
	WithdrawalDelay       uint64
}

// QueueStruct is the queue of L1Txs for a batch
//nolint:structcheck
type QueueStruct struct {
	L1TxQueue    []common.L1Tx
	TotalL1TxFee *big.Int
}

// NewQueueStruct creates a new clear QueueStruct.
func NewQueueStruct() *QueueStruct {
	return &QueueStruct{
		L1TxQueue:    make([]common.L1Tx, 0),
		TotalL1TxFee: big.NewInt(0),
	}
}

// RollupVerifierStruct is the information about verifiers of the Rollup Smart Contract
type RollupVerifierStruct struct {
	MaxTx   int64
	NLevels int64
}

// RollupState represents the state of the Rollup in the Smart Contract
//nolint:structcheck,unused
type RollupState struct {
	StateRoot              *big.Int
	ExitRoots              []*big.Int
	ExitNullifierMap       map[[256 / 8]byte]bool
	TokenList              []ethCommon.Address
	TokenMap               map[ethCommon.Address]bool
	MapL1TxQueue           map[int64]*QueueStruct
	LastL1L2Batch          int64
	CurrentToForgeL1TxsNum int64
	LastToForgeL1TxsNum    int64
	CurrentIdx             int64
}

// RollupEventL1UserTx is an event of the Rollup Smart Contract
type RollupEventL1UserTx struct {
	ToForgeL1TxsNum int64 // QueueIndex       *big.Int
	Position        int   // TransactionIndex *big.Int
	L1Tx            common.L1Tx
}

// RollupEventL1UserTxAux is an event of the Rollup Smart Contract
type RollupEventL1UserTxAux struct {
	ToForgeL1TxsNum uint64 // QueueIndex       *big.Int
	Position        uint8  // TransactionIndex *big.Int
	L1Tx            []byte
}

// RollupEventAddToken is an event of the Rollup Smart Contract
type RollupEventAddToken struct {
	Address ethCommon.Address
	TokenID uint32
}

// RollupEventForgeBatch is an event of the Rollup Smart Contract
type RollupEventForgeBatch struct {
	BatchNum  int64
	EthTxHash ethCommon.Hash
}

// RollupEventUpdateForgeL1L2BatchTimeout is an event of the Rollup Smart Contract
type RollupEventUpdateForgeL1L2BatchTimeout struct {
	ForgeL1L2BatchTimeout uint8
}

// RollupEventUpdateFeeAddToken is an event of the Rollup Smart Contract
type RollupEventUpdateFeeAddToken struct {
	FeeAddToken *big.Int
}

// RollupEventWithdrawEvent is an event of the Rollup Smart Contract
type RollupEventWithdrawEvent struct {
	Idx             uint64
	NumExitRoot     uint64
	InstantWithdraw bool
}

// RollupEvents is the list of events in a block of the Rollup Smart Contract
type RollupEvents struct { //nolint:structcheck
	L1UserTx                    []RollupEventL1UserTx
	AddToken                    []RollupEventAddToken
	ForgeBatch                  []RollupEventForgeBatch
	UpdateForgeL1L2BatchTimeout []RollupEventUpdateForgeL1L2BatchTimeout
	UpdateFeeAddToken           []RollupEventUpdateFeeAddToken
	WithdrawEvent               []RollupEventWithdrawEvent
}

// NewRollupEvents creates an empty RollupEvents with the slices initialized.
func NewRollupEvents() RollupEvents {
	return RollupEvents{
		L1UserTx:                    make([]RollupEventL1UserTx, 0),
		AddToken:                    make([]RollupEventAddToken, 0),
		ForgeBatch:                  make([]RollupEventForgeBatch, 0),
		UpdateForgeL1L2BatchTimeout: make([]RollupEventUpdateForgeL1L2BatchTimeout, 0),
		UpdateFeeAddToken:           make([]RollupEventUpdateFeeAddToken, 0),
		WithdrawEvent:               make([]RollupEventWithdrawEvent, 0),
	}
}

// RollupForgeBatchArgs are the arguments to the ForgeBatch function in the Rollup Smart Contract
//nolint:structcheck,unused
type RollupForgeBatchArgs struct {
	NewLastIdx        int64
	NewStRoot         *big.Int
	NewExitRoot       *big.Int
	L1CoordinatorTxs  []*common.L1Tx
	L2TxsData         []*common.L2Tx
	FeeIdxCoordinator []common.Idx
	// Circuit selector
	VerifierIdx uint8
	L1Batch     bool
	ProofA      [2]*big.Int
	ProofB      [2][2]*big.Int
	ProofC      [2]*big.Int
}

// RollupForgeBatchArgsAux are the arguments to the ForgeBatch function in the Rollup Smart Contract
//nolint:structcheck,unused
type RollupForgeBatchArgsAux struct {
	NewLastIdx        uint64
	NewStRoot         *big.Int
	NewExitRoot       *big.Int
	L1CoordinatorTxs  []byte
	L2TxsData         []byte
	FeeIdxCoordinator []byte
	// Circuit selector
	VerifierIdx uint8
	L1Batch     bool
	ProofA      [2]*big.Int
	ProofB      [2][2]*big.Int
	ProofC      [2]*big.Int
}

// RollupInterface is the inteface to to Rollup Smart Contract
type RollupInterface interface {
	//
	// Smart Contract Methods
	//

	// Public Functions

	RollupForgeBatch(*RollupForgeBatchArgs) (*types.Transaction, error)
	RollupAddToken(tokenAddress ethCommon.Address) (*types.Transaction, error)
	RollupWithdraw(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey,
		numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error)
	RollupForceExit(fromIdx int64, amountF common.Float16, tokenID int64) (*types.Transaction, error)
	RollupForceTransfer(fromIdx int64, amountF common.Float16, tokenID, toIdx int64) (*types.Transaction, error)
	RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey,
		loadAmountF, amountF common.Float16, tokenID int64, toIdx int64) (*types.Transaction, error)
	RollupDepositTransfer(fromIdx int64, loadAmountF, amountF common.Float16,
		tokenID int64, toIdx int64) (*types.Transaction, error)
	RollupDeposit(fromIdx int64, loadAmountF common.Float16, tokenID int64) (*types.Transaction, error)
	RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte,
		babyPubKey babyjub.PublicKey, loadAmountF common.Float16) (*types.Transaction, error)
	RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF common.Float16,
		tokenID int64) (*types.Transaction, error)

	RollupGetCurrentTokens() (*big.Int, error)
	// RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error)
	// RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error)
	// RollupGetQueue(queue int64) ([]byte, error)

	// Governance Public Functions
	RollupUpdateForgeL1L2BatchTimeout(newForgeL1Timeout int64) (*types.Transaction, error)
	RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error)

	//
	// Smart Contract Status
	//

	RollupConstants() (*RollupPublicConstants, error)
	RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error)
	RollupForgeBatchArgs(ethCommon.Hash) (*RollupForgeBatchArgs, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
	client      *EthereumClient
	address     ethCommon.Address
	contractAbi abi.ABI
}

// NewRollupClient creates a new RollupClient
func NewRollupClient(client *EthereumClient, address ethCommon.Address) (*RollupClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(Hermez.HermezABI)))
	if err != nil {
		return nil, err
	}
	return &RollupClient{
		client:      client,
		address:     address,
		contractAbi: contractAbi,
	}, nil
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *RollupClient) RollupForgeBatch(args *RollupForgeBatchArgs) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupAddToken is the interface to call the smart contract function
func (c *RollupClient) RollupAddToken(tokenAddress ethCommon.Address) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupWithdrawSNARK is the interface to call the smart contract function
// func (c *RollupClient) RollupWithdrawSNARK() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupWithdraw is the interface to call the smart contract function
func (c *RollupClient) RollupWithdraw(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey, numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupForceExit is the interface to call the smart contract function
func (c *RollupClient) RollupForceExit(fromIdx int64, amountF common.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupForceTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupForceTransfer(fromIdx int64, amountF common.Float16, tokenID, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDepositTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey, loadAmountF, amountF common.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupDepositTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupDepositTransfer(fromIdx int64, loadAmountF, amountF common.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupDeposit is the interface to call the smart contract function
func (c *RollupClient) RollupDeposit(fromIdx int64, loadAmountF common.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDepositFromRelayer is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte, babyPubKey babyjub.PublicKey, loadAmountF common.Float16) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDeposit is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF common.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupGetTokenAddress is the interface to call the smart contract function
/* func (c *RollupClient) RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error) {
	return nil, errTODO
} */

// RollupGetCurrentTokens is the interface to call the smart contract function
func (c *RollupClient) RollupGetCurrentTokens() (*big.Int, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupGetL1TxFromQueue is the interface to call the smart contract function
/* func (c *RollupClient) RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error) {
	return nil, errTODO
} */

// RollupGetQueue is the interface to call the smart contract function
/* func (c *RollupClient) RollupGetQueue(queue int64) ([]byte, error) {
	return nil, errTODO
}*/

// RollupUpdateForgeL1L2BatchTimeout is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateForgeL1L2BatchTimeout(newForgeL1Timeout int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *RollupClient) RollupConstants() (*RollupPublicConstants, error) {
	rollupConstants := new(RollupPublicConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		hermez, err := Hermez.NewHermez(c.address, ec)
		if err != nil {
			return err
		}
		absoluteMaxL1L2BatchTimeout, err := hermez.ABSOLUTEMAXL1L2BATCHTIMEOUT(nil)
		if err != nil {
			return err
		}
		rollupConstants.AbsoluteMaxL1L2BatchTimeout = int64(absoluteMaxL1L2BatchTimeout)
		rollupConstants.TokenHEZ, err = hermez.TokenHEZ(nil)
		if err != nil {
			return err
		}
		for i := int64(0); i < int64(LenVerifiers); i++ {
			var newRollupVerifier RollupVerifierStruct
			rollupVerifier, err := hermez.RollupVerifiers(nil, big.NewInt(i))
			if err != nil {
				return err
			}
			newRollupVerifier.MaxTx = rollupVerifier.MaxTx.Int64()
			newRollupVerifier.NLevels = rollupVerifier.NLevels.Int64()
			rollupConstants.Verifiers = append(rollupConstants.Verifiers, newRollupVerifier)
		}
		rollupConstants.HermezAuctionContract, err = hermez.HermezAuctionContract(nil)
		if err != nil {
			return err
		}
		rollupConstants.HermezGovernanceDAOAddress, err = hermez.HermezGovernanceDAOAddress(nil)
		if err != nil {
			return err
		}
		rollupConstants.SafetyAddress, err = hermez.SafetyAddress(nil)
		if err != nil {
			return err
		}
		rollupConstants.WithdrawDelayerContract, err = hermez.WithdrawDelayerContract(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return rollupConstants, nil
}

var (
	logHermezL1UserTxEvent               = crypto.Keccak256Hash([]byte("L1UserTxEvent(uint64,uint8,bytes)"))
	logHermezAddToken                    = crypto.Keccak256Hash([]byte("AddToken(address,uint32)"))
	logHermezForgeBatch                  = crypto.Keccak256Hash([]byte("ForgeBatch(uint64)"))
	logHermezUpdateForgeL1L2BatchTimeout = crypto.Keccak256Hash([]byte("UpdateForgeL1L2BatchTimeout(uint8)"))
	logHermezUpdateFeeAddToken           = crypto.Keccak256Hash([]byte("UpdateFeeAddToken(uint256)"))
	logHermezWithdrawEvent               = crypto.Keccak256Hash([]byte("WithdrawEvent(uint48,uint48,bool)"))
)

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *RollupClient) RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error) {
	var rollupEvents RollupEvents
	var blockHash ethCommon.Hash

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []ethCommon.Address{
			c.address,
		},
		BlockHash: nil,
		Topics:    [][]ethCommon.Hash{},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, nil, err
	}
	if len(logs) > 0 {
		blockHash = logs[0].BlockHash
	}
	for _, vLog := range logs {
		if vLog.BlockHash != blockHash {
			return nil, nil, ErrBlockHashMismatchEvent
		}
		switch vLog.Topics[0] {
		case logHermezL1UserTxEvent:
			var L1UserTxAux RollupEventL1UserTxAux
			var L1UserTx RollupEventL1UserTx
			err := c.contractAbi.Unpack(&L1UserTxAux, "L1UserTxEvent", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			L1Tx, err := common.L1TxFromBytes(L1UserTxAux.L1Tx)
			if err != nil {
				return nil, nil, err
			}
			L1UserTx.ToForgeL1TxsNum = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			L1UserTx.Position = int(new(big.Int).SetBytes(vLog.Topics[2][:]).Int64())
			L1UserTx.L1Tx = *L1Tx
			rollupEvents.L1UserTx = append(rollupEvents.L1UserTx, L1UserTx)
		case logHermezAddToken:
			var addToken RollupEventAddToken
			err := c.contractAbi.Unpack(&addToken, "AddToken", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			addToken.Address = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			rollupEvents.AddToken = append(rollupEvents.AddToken, addToken)
		case logHermezForgeBatch:
			var forgeBatch RollupEventForgeBatch
			forgeBatch.BatchNum = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			forgeBatch.EthTxHash = vLog.TxHash
			rollupEvents.ForgeBatch = append(rollupEvents.ForgeBatch, forgeBatch)
		case logHermezUpdateForgeL1L2BatchTimeout:
			var updateForgeL1L2BatchTimeout RollupEventUpdateForgeL1L2BatchTimeout
			err := c.contractAbi.Unpack(&updateForgeL1L2BatchTimeout, "UpdateForgeL1L2BatchTimeout", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			rollupEvents.UpdateForgeL1L2BatchTimeout = append(rollupEvents.UpdateForgeL1L2BatchTimeout, updateForgeL1L2BatchTimeout)
		case logHermezUpdateFeeAddToken:
			var updateFeeAddToken RollupEventUpdateFeeAddToken
			err := c.contractAbi.Unpack(&updateFeeAddToken, "UpdateFeeAddToken", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			rollupEvents.UpdateFeeAddToken = append(rollupEvents.UpdateFeeAddToken, updateFeeAddToken)
		case logHermezWithdrawEvent:
			var withdraw RollupEventWithdrawEvent
			err := c.contractAbi.Unpack(&withdraw, "WithdrawEvent", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			withdraw.Idx = new(big.Int).SetBytes(vLog.Topics[1][:]).Uint64()
			withdraw.NumExitRoot = new(big.Int).SetBytes(vLog.Topics[2][:]).Uint64()
			rollupEvents.WithdrawEvent = append(rollupEvents.WithdrawEvent, withdraw)
		}
	}
	return &rollupEvents, &blockHash, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract in the given transaction
func (c *RollupClient) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*RollupForgeBatchArgs, error) {
	tx, _, err := c.client.client.TransactionByHash(context.Background(), ethTxHash)
	if err != nil {
		return nil, err
	}
	txData := tx.Data()
	method, err := c.contractAbi.MethodById(txData)
	if err != nil {
		return nil, err
	}
	var aux RollupForgeBatchArgsAux
	if err := method.Inputs.Unpack(&aux, txData); err != nil {
		return nil, err
	}
	var rollupForgeBatchArgs RollupForgeBatchArgs
	rollupForgeBatchArgs.L1Batch = aux.L1Batch
	rollupForgeBatchArgs.NewExitRoot = aux.NewExitRoot
	rollupForgeBatchArgs.NewLastIdx = int64(aux.NewLastIdx)
	rollupForgeBatchArgs.NewStRoot = aux.NewStRoot
	rollupForgeBatchArgs.ProofA = aux.ProofA
	rollupForgeBatchArgs.ProofB = aux.ProofB
	rollupForgeBatchArgs.ProofC = aux.ProofC
	rollupForgeBatchArgs.VerifierIdx = aux.VerifierIdx

	numTxsL1 := len(aux.L1CoordinatorTxs) / common.L1TxBytesLen
	for i := 0; i < numTxsL1; i++ {
		l1Tx, err := common.L1TxFromCoordinatorBytes(aux.L1CoordinatorTxs[i*common.L1CoordinatorTxBytesLen : (i+1)*common.L1CoordinatorTxBytesLen])
		if err != nil {
			return nil, err
		}
		rollupForgeBatchArgs.L1CoordinatorTxs = append(rollupForgeBatchArgs.L1CoordinatorTxs, l1Tx)
	}
	rollupConsts, err := c.RollupConstants()
	if err != nil {
		return nil, err
	}
	nLevels := rollupConsts.Verifiers[rollupForgeBatchArgs.VerifierIdx].NLevels
	lenL2TxsBytes := int((nLevels/8)*2 + 2 + 1)
	numTxsL2 := len(aux.L2TxsData) / lenL2TxsBytes
	for i := 0; i < numTxsL2; i++ {
		l2Tx, err := common.L2TxFromBytes(aux.L2TxsData[i*lenL2TxsBytes:(i+1)*lenL2TxsBytes], int(nLevels))
		if err != nil {
			return nil, err
		}
		rollupForgeBatchArgs.L2TxsData = append(rollupForgeBatchArgs.L2TxsData, l2Tx)
	}
	lenFeeIdxCoordinatorBytes := int(nLevels / 8) //nolint:gomnd
	numFeeIdxCoordinator := len(aux.FeeIdxCoordinator) / lenFeeIdxCoordinatorBytes
	for i := 0; i < numFeeIdxCoordinator; i++ {
		var paddedFeeIdx [6]byte
		// TODO: This check is not necessary: the first case will always work.  Test it before removing the if.
		if lenFeeIdxCoordinatorBytes < common.IdxBytesLen {
			copy(paddedFeeIdx[6-lenFeeIdxCoordinatorBytes:], aux.FeeIdxCoordinator[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		} else {
			copy(paddedFeeIdx[:], aux.FeeIdxCoordinator[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		}
		FeeIdxCoordinator, err := common.IdxFromBytes(paddedFeeIdx[:])
		if err != nil {
			return nil, err
		}
		rollupForgeBatchArgs.FeeIdxCoordinator = append(rollupForgeBatchArgs.FeeIdxCoordinator, FeeIdxCoordinator)
	}
	return &rollupForgeBatchArgs, nil
	// tx := client.TransactionByHash(ethTxHash) -> types.Transaction
	// txData := types.Transaction -> Data()
	// m := abi.MethodById(txData) -> Method
	// m.Inputs.Unpack(txData) -> Args
	// client.TransactionReceipt()?
}
