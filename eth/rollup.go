package eth

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	Hermez "github.com/hermeznetwork/hermez-node/eth/contracts/hermez"
	"github.com/hermeznetwork/hermez-node/log"
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
}

// RollupVariables are the variables of the Rollup Smart Contract
type RollupVariables struct {
	FeeAddToken    *big.Int
	ForgeL1Timeout int64
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
	L1Tx             common.L1Tx
	QueueIndex       *big.Int
	TransactionIndex *big.Int
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
	ForgeL1Timeout *big.Int
}

// RollupEventUpdateFeeAddToken is an event of the Rollup Smart Contract
type RollupEventUpdateFeeAddToken struct {
	FeeAddToken *big.Int
}

// RollupEventWithdrawEvent is an event of the Rollup Smart Contract
type RollupEventWithdrawEvent struct {
	Idx             *big.Int
	NumExitRoot     *big.Int
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

	RollupConstants() (*RollupConstants, error)
	RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error)
	RollupForgeBatchArgs(ethCommon.Hash) (*RollupForgeBatchArgs, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
	client  *EthereumClient
	address ethCommon.Address
}

// NewRollupClient creates a new RollupClient
func NewRollupClient(client *EthereumClient, address ethCommon.Address) *RollupClient {
	return &RollupClient{
		client:  client,
		address: address,
	}
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
func (c *RollupClient) RollupConstants() (*RollupConstants, error) {
	rollupConstants := new(RollupConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		rollup, err := Hermez.NewHermez(c.address, ec)
		if err != nil {
			return err
		}
		// rollupConstants.GovernanceAddress :=
		l1CoordinatorBytes, err := rollup.L1COORDINATORBYTES(nil)
		if err != nil {
			return err
		}
		rollupConstants.L1CoordinatorBytes = int(l1CoordinatorBytes.Int64())
		l1UserBytes, err := rollup.L1USERBYTES(nil)
		if err != nil {
			return err
		}
		rollupConstants.L1UserBytes = int(l1UserBytes.Int64())
		l2Bytes, err := rollup.L2BYTES(nil)
		if err != nil {
			return err
		}
		rollupConstants.L2Bytes = int(l2Bytes.Int64())
		rollupConstants.LastIDx, err = rollup.LASTIDX(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxAmountDeposit, err = rollup.MAXLOADAMOUNT(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxAmountL2, err = rollup.MAXAMOUNT(nil)
		if err != nil {
			return err
		}
		maxL1Tx, err := rollup.MAXL1TX(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxL1Tx = int(maxL1Tx.Int64())
		maxL1UserTx, err := rollup.MAXL1USERTX(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxL1UserTx = int(maxL1UserTx.Int64())
		maxTokens, err := rollup.MAXTOKENS(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxTokens = maxTokens.Int64()
		maxWDelay, err := rollup.MAXWITHDRAWALDELAY(nil)
		if err != nil {
			return err
		}
		rollupConstants.MaxWDelay = maxWDelay.Int64()
		noLimitToken, err := rollup.NOLIMIT(nil)
		if err != nil {
			return err
		}
		rollupConstants.NoLimitToken = int(noLimitToken.Int64())
		numBuckets, err := rollup.NUMBUCKETS(nil)
		if err != nil {
			return err
		}
		rollupConstants.NumBuckets = int(numBuckets.Int64())
		// rollupConstants.ReservedIDx =
		rollupConstants.Rfield, err = rollup.RFIELD(nil)
		if err != nil {
			return err
		}
		// rollupConstants.SafetyBot =
		// rollupConstants.TokenHEZ =
		// rollupConstants.WithdrawalContract =
		return nil
	}); err != nil {
		return nil, err
	}
	return rollupConstants, nil
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
