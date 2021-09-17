package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	hermez "github.com/hermeznetwork/hermez-node/eth/contracts/hermez"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// QueueStruct is the queue of L1Txs for a batch
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
type RollupState struct {
	StateRoot *big.Int
	ExitRoots []*big.Int
	// ExitNullifierMap       map[[256 / 8]byte]bool
	ExitNullifierMap       map[int64]map[int64]bool // batchNum -> idx -> bool
	TokenList              []ethCommon.Address
	TokenMap               map[ethCommon.Address]bool
	MapL1TxQueue           map[int64]*QueueStruct
	LastL1L2Batch          int64
	CurrentToForgeL1TxsNum int64
	LastToForgeL1TxsNum    int64
	CurrentIdx             int64
}

// RollupEventInitialize is the InitializeHermezEvent event of the
// Smart Contract
type RollupEventInitialize struct {
	ForgeL1L2BatchTimeout uint8
	FeeAddToken           *big.Int
	WithdrawalDelay       uint64
}

// RollupVariables returns the RollupVariables from the initialize event
func (ei *RollupEventInitialize) RollupVariables() *common.RollupVariables {
	return &common.RollupVariables{
		EthBlockNum:           0,
		FeeAddToken:           ei.FeeAddToken,
		ForgeL1L2BatchTimeout: int64(ei.ForgeL1L2BatchTimeout),
		WithdrawalDelay:       ei.WithdrawalDelay,
		Buckets:               []common.BucketParams{},
		SafeMode:              false,
	}
}

// RollupEventL1UserTx is an event of the Rollup Smart Contract
type RollupEventL1UserTx struct {
	// ToForgeL1TxsNum int64 // QueueIndex       *big.Int
	// Position        int   // TransactionIndex *big.Int
	L1UserTx common.L1Tx
}

// RollupEventL1UserTxAux is an event of the Rollup Smart Contract
type rollupEventL1UserTxAux struct {
	ToForgeL1TxsNum uint64 // QueueIndex       *big.Int
	Position        uint8  // TransactionIndex *big.Int
	L1UserTx        []byte
}

// RollupEventAddToken is an event of the Rollup Smart Contract
type RollupEventAddToken struct {
	TokenAddress ethCommon.Address
	TokenID      uint32
}

// RollupEventForgeBatch is an event of the Rollup Smart Contract
type RollupEventForgeBatch struct {
	BatchNum int64
	// Sender    ethCommon.Address
	EthTxHash    ethCommon.Hash
	L1UserTxsLen uint16
	GasUsed      uint64
	GasPrice     *big.Int
}

// RollupEventUpdateForgeL1L2BatchTimeout is an event of the Rollup Smart Contract
type RollupEventUpdateForgeL1L2BatchTimeout struct {
	NewForgeL1L2BatchTimeout int64
}

// RollupEventUpdateFeeAddToken is an event of the Rollup Smart Contract
type RollupEventUpdateFeeAddToken struct {
	NewFeeAddToken *big.Int
}

// RollupEventWithdraw is an event of the Rollup Smart Contract
type RollupEventWithdraw struct {
	Idx             uint64
	NumExitRoot     uint64
	InstantWithdraw bool
	TxHash          ethCommon.Hash // Hash of the transaction that generated this event
}

type rollupEventUpdateBucketWithdrawAux struct {
	NumBucket   uint8
	BlockStamp  *big.Int
	Withdrawals *big.Int
}

// RollupEventUpdateBucketWithdraw is an event of the Rollup Smart Contract
type RollupEventUpdateBucketWithdraw struct {
	NumBucket   int
	BlockStamp  int64 // blockNum
	Withdrawals *big.Int
}

// RollupEventUpdateWithdrawalDelay is an event of the Rollup Smart Contract
type RollupEventUpdateWithdrawalDelay struct {
	NewWithdrawalDelay uint64
}

// RollupUpdateBucketsParameters are the bucket parameters used in an update
type RollupUpdateBucketsParameters struct {
	CeilUSD         *big.Int
	BlockStamp      *big.Int
	Withdrawals     *big.Int
	RateBlocks      *big.Int
	RateWithdrawals *big.Int
	MaxWithdrawals  *big.Int
}

type rollupEventUpdateBucketsParametersAux struct {
	ArrayBuckets []*big.Int
}

// RollupEventUpdateBucketsParameters is an event of the Rollup Smart Contract
type RollupEventUpdateBucketsParameters struct {
	ArrayBuckets []RollupUpdateBucketsParameters
	SafeMode     bool
}

// RollupEventUpdateTokenExchange is an event of the Rollup Smart Contract
type RollupEventUpdateTokenExchange struct {
	AddressArray []ethCommon.Address
	ValueArray   []uint64
}

// RollupEventSafeMode is an event of the Rollup Smart Contract
type RollupEventSafeMode struct{}

// RollupEvents is the list of events in a block of the Rollup Smart Contract
type RollupEvents struct {
	L1UserTx                    []RollupEventL1UserTx
	AddToken                    []RollupEventAddToken
	ForgeBatch                  []RollupEventForgeBatch
	UpdateForgeL1L2BatchTimeout []RollupEventUpdateForgeL1L2BatchTimeout
	UpdateFeeAddToken           []RollupEventUpdateFeeAddToken
	Withdraw                    []RollupEventWithdraw
	UpdateWithdrawalDelay       []RollupEventUpdateWithdrawalDelay
	UpdateBucketWithdraw        []RollupEventUpdateBucketWithdraw
	UpdateBucketsParameters     []RollupEventUpdateBucketsParameters
	UpdateTokenExchange         []RollupEventUpdateTokenExchange
	SafeMode                    []RollupEventSafeMode
}

// NewRollupEvents creates an empty RollupEvents with the slices initialized.
func NewRollupEvents() RollupEvents {
	return RollupEvents{
		L1UserTx:                    make([]RollupEventL1UserTx, 0),
		AddToken:                    make([]RollupEventAddToken, 0),
		ForgeBatch:                  make([]RollupEventForgeBatch, 0),
		UpdateForgeL1L2BatchTimeout: make([]RollupEventUpdateForgeL1L2BatchTimeout, 0),
		UpdateFeeAddToken:           make([]RollupEventUpdateFeeAddToken, 0),
		Withdraw:                    make([]RollupEventWithdraw, 0),
	}
}

// RollupForgeBatchArgs are the arguments to the ForgeBatch function in the Rollup Smart Contract
type RollupForgeBatchArgs struct {
	NewLastIdx            int64
	NewStRoot             *big.Int
	NewExitRoot           *big.Int
	L1UserTxs             []common.L1Tx
	L1CoordinatorTxs      []common.L1Tx
	L1CoordinatorTxsAuths [][]byte // Authorization for accountCreations for each L1CoordinatorTx
	L2TxsData             []common.L2Tx
	FeeIdxCoordinator     []common.Idx
	// Circuit selector
	VerifierIdx uint8
	L1Batch     bool
	ProofA      [2]*big.Int
	ProofB      [2][2]*big.Int
	ProofC      [2]*big.Int
}

// RollupForgeBatchArgsAux are the arguments to the ForgeBatch function in the Rollup Smart Contract
type rollupForgeBatchArgsAux struct {
	NewLastIdx             *big.Int
	NewStRoot              *big.Int
	NewExitRoot            *big.Int
	EncodedL1CoordinatorTx []byte
	L1L2TxsData            []byte
	FeeIdxCoordinator      []byte
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

	RollupForgeBatch(*RollupForgeBatchArgs, *bind.TransactOpts) (*types.Transaction, error)
	RollupAddToken(tokenAddress ethCommon.Address, feeAddToken,
		deadline *big.Int) (*types.Transaction, error)

	RollupWithdrawMerkleProof(babyPubKey babyjub.PublicKeyComp, tokenID uint32, numExitRoot,
		idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (*types.Transaction,
		error)
	RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32,
		numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error)

	RollupL1UserTxERC20ETH(fromBJJ babyjub.PublicKeyComp, fromIdx int64, depositAmount *big.Int,
		amount *big.Int, tokenID uint32, toIdx int64) (*types.Transaction, error)
	RollupL1UserTxERC20Permit(fromBJJ babyjub.PublicKeyComp, fromIdx int64,
		depositAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64,
		deadline *big.Int) (tx *types.Transaction, err error)

	// Governance Public Functions
	RollupUpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout int64) (*types.Transaction, error)
	RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error)

	// Viewers
	RollupRegisterTokensCount() (*big.Int, error)
	RollupLastForgedBatch() (int64, error)

	//
	// Smart Contract Status
	//

	RollupConstants() (*common.RollupConstants, error)
	RollupEventsByBlock(blockNum int64, blockHash *ethCommon.Hash) (*RollupEvents, error)
	RollupForgeBatchArgs(ethCommon.Hash, uint16) (*RollupForgeBatchArgs, *ethCommon.Address, error)
	RollupEventInit(genesisBlockNum int64) (*RollupEventInitialize, int64, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
	client      *EthereumClient
	chainID     *big.Int
	address     ethCommon.Address
	hermez      *hermez.Hermez
	token       *TokenClient
	contractAbi abi.ABI
	opts        *bind.CallOpts
	consts      *common.RollupConstants
}

// NewRollupClient creates a new RollupClient
func NewRollupClient(client *EthereumClient, address ethCommon.Address) (*RollupClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(hermez.HermezABI)))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	hermez, err := hermez.NewHermez(address, client.Client())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	chainID, err := client.EthChainID()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	c := &RollupClient{
		client:      client,
		chainID:     chainID,
		address:     address,
		hermez:      hermez,
		contractAbi: contractAbi,
		opts:        newCallOpts(),
	}
	consts, err := c.RollupConstants()
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("RollupConstants at %v: %w", address, err))
	}
	c.consts = consts
	c.token, err = NewTokenClient(client, consts.TokenHEZ)
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("new token client at %v: %w", address, err))
	}
	return c, nil
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *RollupClient) RollupForgeBatch(args *RollupForgeBatchArgs, auth *bind.TransactOpts) (tx *types.Transaction, err error) {
	if auth == nil {
		auth, err = c.client.NewAuth()
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		auth.GasLimit = 1000000
	}

	nLevels := c.consts.Verifiers[args.VerifierIdx].NLevels
	lenBytes := nLevels / 8 //nolint:gomnd
	newLastIdx := big.NewInt(int64(args.NewLastIdx))
	// L1CoordinatorBytes
	var l1CoordinatorBytes []byte
	for i := 0; i < len(args.L1CoordinatorTxs); i++ {
		l1 := args.L1CoordinatorTxs[i]
		bytesl1, err := l1.BytesCoordinatorTx(args.L1CoordinatorTxsAuths[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1CoordinatorBytes = append(l1CoordinatorBytes, bytesl1[:]...)
	}
	// L1L2TxData
	var l1l2TxData []byte
	for i := 0; i < len(args.L1UserTxs); i++ {
		l1User := args.L1UserTxs[i]
		bytesl1User, err := l1User.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl1User[:]...)
	}
	for i := 0; i < len(args.L1CoordinatorTxs); i++ {
		l1Coord := args.L1CoordinatorTxs[i]
		bytesl1Coord, err := l1Coord.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl1Coord[:]...)
	}
	for i := 0; i < len(args.L2TxsData); i++ {
		l2 := args.L2TxsData[i]
		bytesl2, err := l2.BytesDataAvailability(uint32(nLevels))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1l2TxData = append(l1l2TxData, bytesl2[:]...)
	}
	// FeeIdxCoordinator
	var feeIdxCoordinator []byte
	if len(args.FeeIdxCoordinator) > common.RollupConstMaxFeeIdxCoordinator {
		return nil, tracerr.Wrap(fmt.Errorf("len(args.FeeIdxCoordinator) > %v",
			common.RollupConstMaxFeeIdxCoordinator))
	}
	for i := 0; i < common.RollupConstMaxFeeIdxCoordinator; i++ {
		feeIdx := common.Idx(0)
		if i < len(args.FeeIdxCoordinator) {
			feeIdx = args.FeeIdxCoordinator[i]
		}
		bytesFeeIdx, err := feeIdx.Bytes()
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		feeIdxCoordinator = append(feeIdxCoordinator, bytesFeeIdx[len(bytesFeeIdx)-int(lenBytes):]...)
	}
	tx, err = c.hermez.ForgeBatch(auth, newLastIdx, args.NewStRoot, args.NewExitRoot,
		l1CoordinatorBytes, l1l2TxData, feeIdxCoordinator, args.VerifierIdx, args.L1Batch,
		args.ProofA, args.ProofB, args.ProofC)
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Hermez.ForgeBatch: %w", err))
	}
	return tx, nil
}

// RollupAddToken is the interface to call the smart contract function.
// `feeAddToken` is the amount of HEZ tokens that will be paid to add the
// token.  `feeAddToken` must match the public value of the smart contract.
func (c *RollupClient) RollupAddToken(tokenAddress ethCommon.Address, feeAddToken,
	deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.token.hez.Nonces(c.opts, owner)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tokenName := c.token.name
			tokenAddr := c.token.address
			digest, _ := createPermitDigest(tokenAddr, owner, spender, c.chainID,
				feeAddToken, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, feeAddToken, deadline, digest,
				signature)

			return c.hermez.AddToken(auth, tokenAddress, permit)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed add Token %w", err))
	}
	return tx, nil
}

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *RollupClient) RollupWithdrawMerkleProof(fromBJJ babyjub.PublicKeyComp, tokenID uint32,
	numExitRoot, idx int64, amount *big.Int, siblings []*big.Int,
	instantWithdraw bool) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			pkCompB := common.SwapEndianness(fromBJJ[:])
			babyPubKey := new(big.Int).SetBytes(pkCompB)
			numExitRootB := uint32(numExitRoot)
			idxBig := big.NewInt(idx)
			return c.hermez.WithdrawMerkleProof(auth, tokenID, amount, babyPubKey,
				numExitRootB, siblings, idxBig, instantWithdraw)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update WithdrawMerkleProof: %w", err))
	}
	return tx, nil
}

// RollupWithdrawCircuit is the interface to call the smart contract function
func (c *RollupClient) RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int,
	tokenID uint32, numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction,
	error) {
	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupL1UserTxERC20ETH is the interface to call the smart contract function
func (c *RollupClient) RollupL1UserTxERC20ETH(fromBJJ babyjub.PublicKeyComp, fromIdx int64,
	depositAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64) (tx *types.Transaction,
	err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			var babyPubKey *big.Int
			if fromBJJ != common.EmptyBJJComp {
				pkCompB := common.SwapEndianness(fromBJJ[:])
				babyPubKey = new(big.Int).SetBytes(pkCompB)
			} else {
				babyPubKey = big.NewInt(0)
			}
			fromIdxBig := big.NewInt(fromIdx)
			toIdxBig := big.NewInt(toIdx)
			depositAmountF, err := common.NewFloat40(depositAmount)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			amountF, err := common.NewFloat40(amount)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			if tokenID == 0 {
				auth.Value = depositAmount
			}
			var permit []byte
			return c.hermez.AddL1Transaction(auth, babyPubKey, fromIdxBig, big.NewInt(int64(depositAmountF)),
				big.NewInt(int64(amountF)), tokenID, toIdxBig, permit)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed add L1 Tx ERC20/ETH: %w", err))
	}
	return tx, nil
}

// RollupL1UserTxERC20Permit is the interface to call the smart contract function
func (c *RollupClient) RollupL1UserTxERC20Permit(fromBJJ babyjub.PublicKeyComp, fromIdx int64,
	depositAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64,
	deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			var babyPubKey *big.Int
			if fromBJJ != common.EmptyBJJComp {
				pkCompB := common.SwapEndianness(fromBJJ[:])
				babyPubKey = new(big.Int).SetBytes(pkCompB)
			} else {
				babyPubKey = big.NewInt(0)
			}
			fromIdxBig := big.NewInt(fromIdx)
			toIdxBig := big.NewInt(toIdx)
			depositAmountF, err := common.NewFloat40(depositAmount)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			amountF, err := common.NewFloat40(amount)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			if tokenID == 0 {
				auth.Value = depositAmount
			}
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.token.hez.Nonces(c.opts, owner)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tokenName := c.token.name
			tokenAddr := c.token.address
			digest, _ := createPermitDigest(tokenAddr, owner, spender, c.chainID,
				amount, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			return c.hermez.AddL1Transaction(auth, babyPubKey, fromIdxBig,
				big.NewInt(int64(depositAmountF)), big.NewInt(int64(amountF)), tokenID, toIdxBig, permit)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed add L1 Tx ERC20Permit: %w", err))
	}
	return tx, nil
}

// RollupRegisterTokensCount is the interface to call the smart contract function
func (c *RollupClient) RollupRegisterTokensCount() (registerTokensCount *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		registerTokensCount, err = c.hermez.RegisterTokensCount(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return registerTokensCount, nil
}

// RollupLastForgedBatch is the interface to call the smart contract function
func (c *RollupClient) RollupLastForgedBatch() (lastForgedBatch int64, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_lastForgedBatch, err := c.hermez.LastForgedBatch(c.opts)
		lastForgedBatch = int64(_lastForgedBatch)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return lastForgedBatch, nil
}

// RollupUpdateForgeL1L2BatchTimeout is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateForgeL1L2BatchTimeout(
	newForgeL1L2BatchTimeout int64) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateForgeL1L2BatchTimeout(auth,
				uint8(newForgeL1L2BatchTimeout))
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update ForgeL1L2BatchTimeout: %w", err))
	}
	return tx, nil
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (tx *types.Transaction,
	err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateFeeAddToken(auth, newFeeAddToken)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update FeeAddToken: %w", err))
	}
	return tx, nil
}

// RollupUpdateBucketsParameters is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateBucketsParameters(
	arrayBuckets []RollupUpdateBucketsParameters,
) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		12500000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			params := make([]*big.Int, len(arrayBuckets))
			for i, bucket := range arrayBuckets {
				params[i], err = c.hermez.PackBucket(c.opts,
					bucket.CeilUSD, bucket.BlockStamp, bucket.Withdrawals,
					bucket.RateBlocks, bucket.RateWithdrawals, bucket.MaxWithdrawals)
				if err != nil {
					return nil, tracerr.Wrap(fmt.Errorf("failed to pack bucket: %w", err))
				}
			}
			return c.hermez.UpdateBucketsParameters(auth, params)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update Buckets Parameters: %w", err))
	}
	return tx, nil
}

// RollupUpdateTokenExchange is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateTokenExchange(addressArray []ethCommon.Address,
	valueArray []uint64) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateTokenExchange(auth, addressArray, valueArray)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update Token Exchange: %w", err))
	}
	return tx, nil
}

// RollupUpdateWithdrawalDelay is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateWithdrawalDelay(newWithdrawalDelay int64) (tx *types.Transaction,
	err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateWithdrawalDelay(auth, uint64(newWithdrawalDelay))
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update WithdrawalDelay: %w", err))
	}
	return tx, nil
}

// RollupSafeMode is the interface to call the smart contract function
func (c *RollupClient) RollupSafeMode() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.SafeMode(auth)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed update Safe Mode: %w", err))
	}
	return tx, nil
}

// RollupInstantWithdrawalViewer is the interface to call the smart contract function
func (c *RollupClient) RollupInstantWithdrawalViewer(tokenAddress ethCommon.Address,
	amount *big.Int) (instantAllowed bool, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		instantAllowed, err = c.hermez.InstantWithdrawalViewer(c.opts, tokenAddress, amount)
		return tracerr.Wrap(err)
	}); err != nil {
		return false, tracerr.Wrap(err)
	}
	return instantAllowed, nil
}

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *RollupClient) RollupConstants() (rollupConstants *common.RollupConstants, err error) {
	rollupConstants = new(common.RollupConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		absoluteMaxL1L2BatchTimeout, err := c.hermez.ABSOLUTEMAXL1L2BATCHTIMEOUT(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		rollupConstants.AbsoluteMaxL1L2BatchTimeout = int64(absoluteMaxL1L2BatchTimeout)
		rollupConstants.TokenHEZ, err = c.hermez.TokenHEZ(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		rollupVerifiersLength, err := c.hermez.RollupVerifiersLength(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		for i := int64(0); i < rollupVerifiersLength.Int64(); i++ {
			var newRollupVerifier common.RollupVerifierStruct
			rollupVerifier, err := c.hermez.RollupVerifiers(c.opts, big.NewInt(i))
			if err != nil {
				return tracerr.Wrap(err)
			}
			newRollupVerifier.MaxTx = rollupVerifier.MaxTx.Int64()
			newRollupVerifier.NLevels = rollupVerifier.NLevels.Int64()
			rollupConstants.Verifiers = append(rollupConstants.Verifiers,
				newRollupVerifier)
		}
		rollupConstants.HermezAuctionContract, err = c.hermez.HermezAuctionContract(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		rollupConstants.HermezGovernanceAddress, err = c.hermez.HermezGovernanceAddress(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		rollupConstants.WithdrawDelayerContract, err = c.hermez.WithdrawDelayerContract(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return rollupConstants, nil
}

var (
	logHermezL1UserTxEvent = crypto.Keccak256Hash([]byte(
		"L1UserTxEvent(uint32,uint8,bytes)"))
	logHermezAddToken = crypto.Keccak256Hash([]byte(
		"AddToken(address,uint32)"))
	logHermezForgeBatch = crypto.Keccak256Hash([]byte(
		"ForgeBatch(uint32,uint16)"))
	logHermezUpdateForgeL1L2BatchTimeout = crypto.Keccak256Hash([]byte(
		"UpdateForgeL1L2BatchTimeout(uint8)"))
	logHermezUpdateFeeAddToken = crypto.Keccak256Hash([]byte(
		"UpdateFeeAddToken(uint256)"))
	logHermezWithdrawEvent = crypto.Keccak256Hash([]byte(
		"WithdrawEvent(uint48,uint32,bool)"))
	logHermezUpdateBucketWithdraw = crypto.Keccak256Hash([]byte(
		"UpdateBucketWithdraw(uint8,uint256,uint256)"))
	logHermezUpdateWithdrawalDelay = crypto.Keccak256Hash([]byte(
		"UpdateWithdrawalDelay(uint64)"))
	logHermezUpdateBucketsParameters = crypto.Keccak256Hash([]byte(
		"UpdateBucketsParameters(uint256[])"))
	logHermezUpdateTokenExchange = crypto.Keccak256Hash([]byte(
		"UpdateTokenExchange(address[],uint64[])"))
	logHermezSafeMode = crypto.Keccak256Hash([]byte(
		"SafeMode()"))
	logHermezInitialize = crypto.Keccak256Hash([]byte(
		"InitializeHermezEvent(uint8,uint256,uint64)"))
)

// RollupEventInit returns the initialize event with its corresponding block number
func (c *RollupClient) RollupEventInit(genesisBlockNum int64) (*RollupEventInitialize, int64, error) {
	query := ethereum.FilterQuery{
		Addresses: []ethCommon.Address{
			c.address,
		},
		FromBlock: big.NewInt(max(0, genesisBlockNum-blocksPerDay)),
		ToBlock:   big.NewInt(genesisBlockNum),
		Topics:    [][]ethCommon.Hash{{logHermezInitialize}},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(logs) != 1 {
		return nil, 0, tracerr.Wrap(fmt.Errorf("no event of type InitializeHermezEvent found"))
	}
	vLog := logs[0]
	if vLog.Topics[0] != logHermezInitialize {
		return nil, 0, tracerr.Wrap(fmt.Errorf("event is not InitializeHermezEvent"))
	}

	var rollupInit RollupEventInitialize
	if err := c.contractAbi.UnpackIntoInterface(&rollupInit, "InitializeHermezEvent",
		vLog.Data); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	return &rollupInit, int64(vLog.BlockNumber), tracerr.Wrap(err)
}

// RollupEventsByBlock returns the events in a block that happened in the
// Rollup Smart Contract.
// To query by blockNum, set blockNum >= 0 and blockHash == nil.
// To query by blockHash set blockHash != nil, and blockNum will be ignored.
// If there are no events in that block the result is nil.
func (c *RollupClient) RollupEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*RollupEvents, error) {
	var rollupEvents RollupEvents

	var blockNumBigInt *big.Int
	if blockHash == nil {
		blockNumBigInt = big.NewInt(blockNum)
	}
	query := ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: blockNumBigInt,
		ToBlock:   blockNumBigInt,
		Addresses: []ethCommon.Address{
			c.address,
		},
		Topics: [][]ethCommon.Hash{},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if len(logs) == 0 {
		return nil, nil
	}

	for _, vLog := range logs {
		if blockHash != nil && vLog.BlockHash != *blockHash {
			log.Errorw("Block hash mismatch", "expected", blockHash.String(), "got", vLog.BlockHash.String())
			return nil, tracerr.Wrap(ErrBlockHashMismatchEvent)
		}
		switch vLog.Topics[0] {
		case logHermezL1UserTxEvent:
			var L1UserTxAux rollupEventL1UserTxAux
			var L1UserTx RollupEventL1UserTx
			err := c.contractAbi.UnpackIntoInterface(&L1UserTxAux, "L1UserTxEvent", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			L1Tx, err := common.L1UserTxFromBytes(L1UserTxAux.L1UserTx)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			toForgeL1TxsNum := new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			L1Tx.ToForgeL1TxsNum = &toForgeL1TxsNum
			L1Tx.Position = int(new(big.Int).SetBytes(vLog.Topics[2][:]).Int64())
			L1Tx.UserOrigin = true
			L1Tx.EthTxHash = vLog.TxHash
			//Get l1Fee in eth wei spent in the l1 tx
			tx, _, err := c.client.client.TransactionByHash(context.Background(), vLog.TxHash)
			if err != nil {
				return nil, tracerr.Wrap(fmt.Errorf("failed to get TransactionByHash, hash: %s, err: %w", vLog.TxHash.String(), err))
			}
			l1Fee := new(big.Int).Mul(tx.GasPrice(), new(big.Int).SetUint64(tx.Gas()))
			L1Tx.L1Fee = l1Fee
			L1UserTx.L1UserTx = *L1Tx
			rollupEvents.L1UserTx = append(rollupEvents.L1UserTx, L1UserTx)
		case logHermezAddToken:
			var addToken RollupEventAddToken
			err := c.contractAbi.UnpackIntoInterface(&addToken, "AddToken", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			addToken.TokenAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			rollupEvents.AddToken = append(rollupEvents.AddToken, addToken)
		case logHermezForgeBatch:
			var forgeBatch RollupEventForgeBatch
			err := c.contractAbi.UnpackIntoInterface(&forgeBatch, "ForgeBatch", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			forgeBatch.BatchNum = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			forgeBatch.EthTxHash = vLog.TxHash
			//Check tx info using EthTxHash to get gasprice and gas used
			tx, _, err := c.client.client.TransactionByHash(context.Background(), vLog.TxHash)
			if err != nil {
				return nil, tracerr.Wrap(fmt.Errorf("failed to get TransactionByHash, hash: %s, err: %w", vLog.TxHash.String(), err))
			}
			forgeBatch.GasPrice = tx.GasPrice()
			// Get gas used from TxReceipt
			txReceipt, err := c.client.client.TransactionReceipt(context.Background(), vLog.TxHash)
			if err != nil {
				return nil, tracerr.Wrap(fmt.Errorf("failed to get TransactionByHash, hash: %s, err: %w", vLog.TxHash.String(), err))
			}
			forgeBatch.GasUsed = txReceipt.GasUsed
			rollupEvents.ForgeBatch = append(rollupEvents.ForgeBatch, forgeBatch)
		case logHermezUpdateForgeL1L2BatchTimeout:
			var updateForgeL1L2BatchTimeout struct {
				NewForgeL1L2BatchTimeout uint8
			}
			err := c.contractAbi.UnpackIntoInterface(&updateForgeL1L2BatchTimeout,
				"UpdateForgeL1L2BatchTimeout", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			rollupEvents.UpdateForgeL1L2BatchTimeout = append(rollupEvents.UpdateForgeL1L2BatchTimeout,
				RollupEventUpdateForgeL1L2BatchTimeout{
					NewForgeL1L2BatchTimeout: int64(updateForgeL1L2BatchTimeout.NewForgeL1L2BatchTimeout),
				})
		case logHermezUpdateFeeAddToken:
			var updateFeeAddToken RollupEventUpdateFeeAddToken
			err := c.contractAbi.UnpackIntoInterface(&updateFeeAddToken, "UpdateFeeAddToken", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			rollupEvents.UpdateFeeAddToken = append(rollupEvents.UpdateFeeAddToken, updateFeeAddToken)
		case logHermezWithdrawEvent:
			var withdraw RollupEventWithdraw
			withdraw.Idx = new(big.Int).SetBytes(vLog.Topics[1][:]).Uint64()
			withdraw.NumExitRoot = new(big.Int).SetBytes(vLog.Topics[2][:]).Uint64()
			instantWithdraw := new(big.Int).SetBytes(vLog.Topics[3][:]).Uint64()
			if instantWithdraw == 1 {
				withdraw.InstantWithdraw = true
			}
			withdraw.TxHash = vLog.TxHash
			rollupEvents.Withdraw = append(rollupEvents.Withdraw, withdraw)
		case logHermezUpdateBucketWithdraw:
			var updateBucketWithdrawAux rollupEventUpdateBucketWithdrawAux
			var updateBucketWithdraw RollupEventUpdateBucketWithdraw
			err := c.contractAbi.UnpackIntoInterface(&updateBucketWithdrawAux,
				"UpdateBucketWithdraw", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			updateBucketWithdraw.Withdrawals = updateBucketWithdrawAux.Withdrawals
			updateBucketWithdraw.NumBucket = int(new(big.Int).SetBytes(vLog.Topics[1][:]).Int64())
			updateBucketWithdraw.BlockStamp = new(big.Int).SetBytes(vLog.Topics[2][:]).Int64()
			rollupEvents.UpdateBucketWithdraw =
				append(rollupEvents.UpdateBucketWithdraw, updateBucketWithdraw)

		case logHermezUpdateWithdrawalDelay:
			var withdrawalDelay RollupEventUpdateWithdrawalDelay
			err := c.contractAbi.UnpackIntoInterface(&withdrawalDelay, "UpdateWithdrawalDelay", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			rollupEvents.UpdateWithdrawalDelay = append(rollupEvents.UpdateWithdrawalDelay, withdrawalDelay)
		case logHermezUpdateBucketsParameters:
			var bucketsParametersAux rollupEventUpdateBucketsParametersAux
			var bucketsParameters RollupEventUpdateBucketsParameters
			err := c.contractAbi.UnpackIntoInterface(&bucketsParametersAux,
				"UpdateBucketsParameters", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			bucketsParameters.ArrayBuckets = make([]RollupUpdateBucketsParameters, len(bucketsParametersAux.ArrayBuckets))
			for i, bucket := range bucketsParametersAux.ArrayBuckets {
				bucket, err := c.hermez.UnpackBucket(c.opts, bucket)
				if err != nil {
					return nil, tracerr.Wrap(err)
				}
				bucketsParameters.ArrayBuckets[i].CeilUSD = bucket.CeilUSD
				bucketsParameters.ArrayBuckets[i].BlockStamp = bucket.BlockStamp
				bucketsParameters.ArrayBuckets[i].Withdrawals = bucket.Withdrawals
				bucketsParameters.ArrayBuckets[i].RateBlocks = bucket.RateBlocks
				bucketsParameters.ArrayBuckets[i].RateWithdrawals = bucket.RateWithdrawals
				bucketsParameters.ArrayBuckets[i].MaxWithdrawals = bucket.MaxWithdrawals
			}
			rollupEvents.UpdateBucketsParameters =
				append(rollupEvents.UpdateBucketsParameters, bucketsParameters)
		case logHermezUpdateTokenExchange:
			var tokensExchange RollupEventUpdateTokenExchange
			err := c.contractAbi.UnpackIntoInterface(&tokensExchange, "UpdateTokenExchange", vLog.Data)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			rollupEvents.UpdateTokenExchange = append(rollupEvents.UpdateTokenExchange, tokensExchange)
		case logHermezSafeMode:
			var safeMode RollupEventSafeMode
			rollupEvents.SafeMode = append(rollupEvents.SafeMode, safeMode)
			// Also add an UpdateBucketsParameter with
			// SafeMode=true to keep the order between `safeMode`
			// and `UpdateBucketsParameters`
			bucketsParameters := RollupEventUpdateBucketsParameters{
				SafeMode: true,
			}
			for i := range bucketsParameters.ArrayBuckets {
				bucketsParameters.ArrayBuckets[i].CeilUSD = big.NewInt(0)
				bucketsParameters.ArrayBuckets[i].BlockStamp = big.NewInt(0)
				bucketsParameters.ArrayBuckets[i].Withdrawals = big.NewInt(0)
				bucketsParameters.ArrayBuckets[i].RateBlocks = big.NewInt(0)
				bucketsParameters.ArrayBuckets[i].RateWithdrawals = big.NewInt(0)
				bucketsParameters.ArrayBuckets[i].MaxWithdrawals = big.NewInt(0)
			}
			rollupEvents.UpdateBucketsParameters = append(rollupEvents.UpdateBucketsParameters,
				bucketsParameters)
		}
	}
	return &rollupEvents, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the
// Rollup Smart Contract in the given transaction, and the sender address.
func (c *RollupClient) RollupForgeBatchArgs(ethTxHash ethCommon.Hash,
	l1UserTxsLen uint16) (*RollupForgeBatchArgs, *ethCommon.Address, error) {
	tx, _, err := c.client.client.TransactionByHash(context.Background(), ethTxHash)
	if err != nil {
		return nil, nil, tracerr.Wrap(fmt.Errorf("TransactionByHash: %w", err))
	}
	txData := tx.Data()

	method, err := c.contractAbi.MethodById(txData[:4])
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	receipt, err := c.client.client.TransactionReceipt(context.Background(), ethTxHash)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	sender, err := c.client.client.TransactionSender(context.Background(), tx,
		receipt.Logs[0].BlockHash, receipt.Logs[0].Index)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	var aux rollupForgeBatchArgsAux
	if values, err := method.Inputs.Unpack(txData[4:]); err != nil {
		return nil, nil, tracerr.Wrap(err)
	} else if err := method.Inputs.Copy(&aux, values); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	rollupForgeBatchArgs := RollupForgeBatchArgs{
		L1Batch:               aux.L1Batch,
		NewExitRoot:           aux.NewExitRoot,
		NewLastIdx:            aux.NewLastIdx.Int64(),
		NewStRoot:             aux.NewStRoot,
		ProofA:                aux.ProofA,
		ProofB:                aux.ProofB,
		ProofC:                aux.ProofC,
		VerifierIdx:           aux.VerifierIdx,
		L1CoordinatorTxs:      []common.L1Tx{},
		L1CoordinatorTxsAuths: [][]byte{},
		L2TxsData:             []common.L2Tx{},
		FeeIdxCoordinator:     []common.Idx{},
	}
	nLevels := c.consts.Verifiers[rollupForgeBatchArgs.VerifierIdx].NLevels
	lenL1L2TxsBytes := int((nLevels/8)*2 + common.Float40BytesLength + 1) //nolint:gomnd
	numBytesL1TxUser := int(l1UserTxsLen) * lenL1L2TxsBytes
	numTxsL1Coord := len(aux.EncodedL1CoordinatorTx) / common.RollupConstL1CoordinatorTotalBytes
	numBytesL1TxCoord := numTxsL1Coord * lenL1L2TxsBytes
	numBeginL2Tx := numBytesL1TxCoord + numBytesL1TxUser
	l1UserTxsData := []byte{}
	if l1UserTxsLen > 0 {
		l1UserTxsData = aux.L1L2TxsData[:numBytesL1TxUser]
	}
	for i := 0; i < int(l1UserTxsLen); i++ {
		l1Tx, err :=
			common.L1TxFromDataAvailability(l1UserTxsData[i*lenL1L2TxsBytes:(i+1)*lenL1L2TxsBytes],
				uint32(nLevels))
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		rollupForgeBatchArgs.L1UserTxs = append(rollupForgeBatchArgs.L1UserTxs, *l1Tx)
	}
	l2TxsData := []byte{}
	if numBeginL2Tx < len(aux.L1L2TxsData) {
		l2TxsData = aux.L1L2TxsData[numBeginL2Tx:]
	}
	numTxsL2 := len(l2TxsData) / lenL1L2TxsBytes
	for i := 0; i < numTxsL2; i++ {
		l2Tx, err :=
			common.L2TxFromBytesDataAvailability(l2TxsData[i*lenL1L2TxsBytes:(i+1)*lenL1L2TxsBytes],
				int(nLevels))
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		rollupForgeBatchArgs.L2TxsData = append(rollupForgeBatchArgs.L2TxsData, *l2Tx)
	}
	for i := 0; i < numTxsL1Coord; i++ {
		bytesL1Coordinator :=
			aux.EncodedL1CoordinatorTx[i*common.RollupConstL1CoordinatorTotalBytes : (i+1)*common.RollupConstL1CoordinatorTotalBytes] //nolint:lll
		var signature []byte
		v := bytesL1Coordinator[0]
		s := bytesL1Coordinator[1:33]
		r := bytesL1Coordinator[33:65]
		signature = append(signature, r[:]...)
		signature = append(signature, s[:]...)
		signature = append(signature, v)
		l1Tx, err := common.L1CoordinatorTxFromBytes(bytesL1Coordinator, c.chainID, c.address)
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		rollupForgeBatchArgs.L1CoordinatorTxs = append(rollupForgeBatchArgs.L1CoordinatorTxs, *l1Tx)
		rollupForgeBatchArgs.L1CoordinatorTxsAuths =
			append(rollupForgeBatchArgs.L1CoordinatorTxsAuths, signature)
	}
	lenFeeIdxCoordinatorBytes := int(nLevels / 8) //nolint:gomnd
	numFeeIdxCoordinator := len(aux.FeeIdxCoordinator) / lenFeeIdxCoordinatorBytes
	for i := 0; i < numFeeIdxCoordinator; i++ {
		var paddedFeeIdx [6]byte
		if lenFeeIdxCoordinatorBytes < common.IdxBytesLen {
			copy(paddedFeeIdx[6-lenFeeIdxCoordinatorBytes:],
				aux.FeeIdxCoordinator[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		} else {
			copy(paddedFeeIdx[:],
				aux.FeeIdxCoordinator[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		}
		feeIdxCoordinator, err := common.IdxFromBytes(paddedFeeIdx[:])
		if err != nil {
			return nil, nil, tracerr.Wrap(err)
		}
		if feeIdxCoordinator != common.Idx(0) {
			rollupForgeBatchArgs.FeeIdxCoordinator =
				append(rollupForgeBatchArgs.FeeIdxCoordinator, feeIdxCoordinator)
		}
	}
	return &rollupForgeBatchArgs, &sender, nil
}
