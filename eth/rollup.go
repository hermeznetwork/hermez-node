package eth

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/utils"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// FeeIdxCoordinatorLen is the number of tokens the coordinator can use
	// to collect fees (determines the number of tokens that the
	// coordinator can collect fees from).  This value is determined by the
	// circuit.
	FeeIdxCoordinatorLen = 64
)

// RollupConstants are the constants of the Rollup Smart Contract
type RollupConstants struct {
	// Maxim Deposit allowed
	MaxAmountDeposit *big.Int
	MaxAmountL2      *big.Int
	MaxTokens        uint32
	// maximum L1 transactions allowed to be queued for a batch
	MaxL1Tx int
	// maximum L1 user transactions allowed to be queued for a batch
	MaxL1UserTx        int
	Rfield             *big.Int
	L1CoordinatorBytes int
	L1UserBytes        int
	L2Bytes            int
}

// RollupVariables are the variables of the Rollup Smart Contract
type RollupVariables struct {
	MaxTxVerifiers     []int
	TokenHEZ           ethCommon.Address
	GovernanceAddress  ethCommon.Address
	SafetyBot          ethCommon.Address
	ConsensusContract  ethCommon.Address
	WithdrawalContract ethCommon.Address
	FeeAddToken        *big.Int
	ForgeL1Timeout     int64
	FeeL1UserTx        *big.Int
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
	L1Tx common.L1Tx
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

// RollupEventUpdateForgeL1Timeout is an event of the Rollup Smart Contract
type RollupEventUpdateForgeL1Timeout struct {
	ForgeL1Timeout int64
}

// RollupEventUpdateFeeL1UserTx is an event of the Rollup Smart Contract
type RollupEventUpdateFeeL1UserTx struct {
	FeeL1UserTx *big.Int
}

// RollupEventUpdateFeeAddToken is an event of the Rollup Smart Contract
type RollupEventUpdateFeeAddToken struct {
	FeeAddToken *big.Int
}

// RollupEventUpdateTokenHez is an event of the Rollup Smart Contract
type RollupEventUpdateTokenHez struct {
	TokenHEZ ethCommon.Address
}

// RollupEventWithdraw is an event of the Rollup Smart Contract
type RollupEventWithdraw struct {
	Idx         int64
	NumExitRoot int
}

// RollupEvents is the list of events in a block of the Rollup Smart Contract
type RollupEvents struct { //nolint:structcheck
	L1UserTx             []RollupEventL1UserTx
	AddToken             []RollupEventAddToken
	ForgeBatch           []RollupEventForgeBatch
	UpdateForgeL1Timeout []RollupEventUpdateForgeL1Timeout
	UpdateFeeL1UserTx    []RollupEventUpdateFeeL1UserTx
	UpdateFeeAddToken    []RollupEventUpdateFeeAddToken
	UpdateTokenHez       []RollupEventUpdateTokenHez
	Withdraw             []RollupEventWithdraw
}

// NewRollupEvents creates an empty RollupEvents with the slices initialized.
func NewRollupEvents() RollupEvents {
	return RollupEvents{
		L1UserTx:             make([]RollupEventL1UserTx, 0),
		AddToken:             make([]RollupEventAddToken, 0),
		ForgeBatch:           make([]RollupEventForgeBatch, 0),
		UpdateForgeL1Timeout: make([]RollupEventUpdateForgeL1Timeout, 0),
		UpdateFeeL1UserTx:    make([]RollupEventUpdateFeeL1UserTx, 0),
		UpdateFeeAddToken:    make([]RollupEventUpdateFeeAddToken, 0),
		UpdateTokenHez:       make([]RollupEventUpdateTokenHez, 0),
		Withdraw:             make([]RollupEventWithdraw, 0),
	}
}

// RollupForgeBatchArgs are the arguments to the ForgeBatch function in the Rollup Smart Contract
//nolint:structcheck,unused
type RollupForgeBatchArgs struct {
	ProofA                [2]*big.Int
	ProofB                [2][2]*big.Int
	ProofC                [2]*big.Int
	NewLastIdx            int64
	NewStRoot             *big.Int
	NewExitRoot           *big.Int
	L1CoordinatorTxs      []*common.L1Tx
	L1CoordinatorTxsAuths [][]byte // Authorization for accountCreations for each L1CoordinatorTxs
	L2Txs                 []*common.L2Tx
	FeeIdxCoordinator     []common.Idx
	// Circuit selector
	VerifierIdx int64
	L1Batch     bool
}

// RollupInterface is the inteface to to Rollup Smart Contract
type RollupInterface interface {
	//
	// Smart Contract Methods
	//

	// Public Functions

	RollupForgeBatch(*RollupForgeBatchArgs) (*types.Transaction, error)
	RollupAddToken(tokenAddress ethCommon.Address) (*types.Transaction, error)
	// RollupWithdrawSNARK() (*types.Transaction, error) // TODO (Not defined in Hermez.sol)
	RollupWithdrawMerkleProof(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey,
		numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error)
	RollupForceExit(fromIdx int64, amountF utils.Float16, tokenID int64) (*types.Transaction, error)
	RollupForceTransfer(fromIdx int64, amountF utils.Float16, tokenID, toIdx int64) (*types.Transaction, error)
	RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey,
		loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error)
	RollupDepositTransfer(fromIdx int64, loadAmountF, amountF utils.Float16,
		tokenID int64, toIdx int64) (*types.Transaction, error)
	RollupDeposit(fromIdx int64, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error)
	RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte,
		babyPubKey babyjub.PublicKey, loadAmountF utils.Float16) (*types.Transaction, error)
	RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF utils.Float16,
		tokenID int64) (*types.Transaction, error)

	RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error)
	RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error)
	RollupGetQueue(queue int64) ([]byte, error)

	// Governance Public Functions
	RollupUpdateForgeL1Timeout(newForgeL1Timeout int64) (*types.Transaction, error)
	RollupUpdateFeeL1UserTx(newFeeL1UserTx *big.Int) (*types.Transaction, error)
	RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error)
	RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (*types.Transaction, error)
	// RollupUpdateGovernance() (*types.Transaction, error) // TODO (Not defined in Hermez.sol)

	//
	// Smart Contract Status
	//

	RollupConstants() (*RollupConstants, error)
	RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error)
	RollupForgeBatchArgs(ethCommon.Hash) (*RollupForgeBatchArgs, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
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

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *RollupClient) RollupWithdrawMerkleProof(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey, numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupForceExit is the interface to call the smart contract function
func (c *RollupClient) RollupForceExit(fromIdx int64, amountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupForceTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupForceTransfer(fromIdx int64, amountF utils.Float16, tokenID, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDepositTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupDepositTransfer is the interface to call the smart contract function
func (c *RollupClient) RollupDepositTransfer(fromIdx int64, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupDeposit is the interface to call the smart contract function
func (c *RollupClient) RollupDeposit(fromIdx int64, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDepositFromRelayer is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte, babyPubKey babyjub.PublicKey, loadAmountF utils.Float16) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupCreateAccountDeposit is the interface to call the smart contract function
func (c *RollupClient) RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupGetTokenAddress is the interface to call the smart contract function
func (c *RollupClient) RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupGetL1TxFromQueue is the interface to call the smart contract function
func (c *RollupClient) RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupGetQueue is the interface to call the smart contract function
func (c *RollupClient) RollupGetQueue(queue int64) ([]byte, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateForgeL1Timeout is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateForgeL1Timeout(newForgeL1Timeout int64) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateFeeL1UserTx is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateFeeL1UserTx(newFeeL1UserTx *big.Int) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateTokensHEZ is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateGovernance is the interface to call the smart contract function
// func (c *RollupClient) RollupUpdateGovernance() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *RollupClient) RollupConstants() (*RollupConstants, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *RollupClient) RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error) {
	log.Error("TODO")
	return nil, nil, errTODO
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract in the given transaction
func (c *RollupClient) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*RollupForgeBatchArgs, error) {
	// tx := client.TransactionByHash(ethTxHash) -> types.Transaction
	// txData := types.Transaction -> Data()
	// m := abi.MethodById(txData) -> Method
	// m.Inputs.Unpack(txData) -> Args
	// client.TransactionReceipt()?
	log.Error("TODO")
	return nil, errTODO
}
