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
	Hermez "github.com/hermeznetwork/hermez-node/eth/contracts/hermez"
	HEZ "github.com/hermeznetwork/hermez-node/eth/contracts/tokenHEZ"
	"github.com/hermeznetwork/hermez-node/log"
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

// RollupEventL1UserTx is an event of the Rollup Smart Contract
type RollupEventL1UserTx struct {
	// ToForgeL1TxsNum int64 // QueueIndex       *big.Int
	// Position        int   // TransactionIndex *big.Int
	L1UserTx common.L1Tx
}

// RollupEventL1UserTxAux is an event of the Rollup Smart Contract
type RollupEventL1UserTxAux struct {
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
	EthTxHash ethCommon.Hash
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

// RollupEvents is the list of events in a block of the Rollup Smart Contract
type RollupEvents struct {
	L1UserTx                    []RollupEventL1UserTx
	AddToken                    []RollupEventAddToken
	ForgeBatch                  []RollupEventForgeBatch
	UpdateForgeL1L2BatchTimeout []RollupEventUpdateForgeL1L2BatchTimeout
	UpdateFeeAddToken           []RollupEventUpdateFeeAddToken
	Withdraw                    []RollupEventWithdraw
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
type RollupForgeBatchArgsAux struct {
	NewLastIdx             *big.Int
	NewStRoot              *big.Int
	NewExitRoot            *big.Int
	EncodedL1CoordinatorTx []byte
	L2TxsData              []byte
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

	RollupForgeBatch(*RollupForgeBatchArgs) (*types.Transaction, error)
	RollupAddToken(tokenAddress ethCommon.Address, feeAddToken, deadline *big.Int) (*types.Transaction, error)

	RollupWithdrawMerkleProof(babyPubKey *babyjub.PublicKey, tokenID uint32, numExitRoot, idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (*types.Transaction, error)
	RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32, numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error)

	RollupL1UserTxERC20ETH(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64) (*types.Transaction, error)
	RollupL1UserTxERC20Permit(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64, deadline *big.Int) (tx *types.Transaction, err error)

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
	RollupEventsByBlock(blockNum int64) (*RollupEvents, *ethCommon.Hash, error)
	RollupForgeBatchArgs(ethCommon.Hash) (*RollupForgeBatchArgs, *ethCommon.Address, error)
}

//
// Implementation
//

// RollupClient is the implementation of the interface to the Rollup Smart Contract in ethereum.
type RollupClient struct {
	client      *EthereumClient
	address     ethCommon.Address
	tokenHEZCfg TokenConfig
	hermez      *Hermez.Hermez
	tokenHEZ    *HEZ.HEZ
	contractAbi abi.ABI
}

// NewRollupClient creates a new RollupClient
func NewRollupClient(client *EthereumClient, address ethCommon.Address, tokenHEZCfg TokenConfig) (*RollupClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(Hermez.HermezABI)))
	if err != nil {
		return nil, err
	}
	hermez, err := Hermez.NewHermez(address, client.Client())
	if err != nil {
		return nil, err
	}
	tokenHEZ, err := HEZ.NewHEZ(tokenHEZCfg.Address, client.Client())
	if err != nil {
		return nil, err
	}
	return &RollupClient{
		client:      client,
		address:     address,
		tokenHEZCfg: tokenHEZCfg,
		hermez:      hermez,
		tokenHEZ:    tokenHEZ,
		contractAbi: contractAbi,
	}, nil
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *RollupClient) RollupForgeBatch(args *RollupForgeBatchArgs) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		1000000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			rollupConst, err := c.RollupConstants()
			if err != nil {
				return nil, err
			}
			nLevels := rollupConst.Verifiers[args.VerifierIdx].NLevels
			lenBytes := nLevels / 8 //nolint:gomnd
			newLastIdx := big.NewInt(int64(args.NewLastIdx))
			var l1CoordinatorBytes []byte
			for i := 0; i < len(args.L1CoordinatorTxs); i++ {
				l1 := args.L1CoordinatorTxs[i]
				bytesl1, err := l1.BytesCoordinatorTx(args.L1CoordinatorTxsAuths[i])
				if err != nil {
					return nil, err
				}
				l1CoordinatorBytes = append(l1CoordinatorBytes, bytesl1[:]...)
			}
			var l2DataBytes []byte
			for i := 0; i < len(args.L2TxsData); i++ {
				l2 := args.L2TxsData[i]
				bytesl2, err := l2.Bytes(int(nLevels))
				if err != nil {
					return nil, err
				}
				l2DataBytes = append(l2DataBytes, bytesl2[:]...)
			}
			var feeIdxCoordinator []byte
			if len(args.FeeIdxCoordinator) > common.RollupConstMaxFeeIdxCoordinator {
				return nil, fmt.Errorf("len(args.FeeIdxCoordinator) > %v",
					common.RollupConstMaxFeeIdxCoordinator)
			}
			for i := 0; i < common.RollupConstMaxFeeIdxCoordinator; i++ {
				feeIdx := common.Idx(0)
				if i < len(args.FeeIdxCoordinator) {
					feeIdx = args.FeeIdxCoordinator[i]
				}
				bytesFeeIdx, err := feeIdx.Bytes()
				if err != nil {
					return nil, err
				}
				feeIdxCoordinator = append(feeIdxCoordinator, bytesFeeIdx[len(bytesFeeIdx)-int(lenBytes):]...)
			}
			return c.hermez.ForgeBatch(auth, newLastIdx, args.NewStRoot, args.NewExitRoot, l1CoordinatorBytes, l2DataBytes, feeIdxCoordinator, args.VerifierIdx, args.L1Batch, args.ProofA, args.ProofB, args.ProofC)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed forge batch: %w", err)
	}
	return tx, nil
}

// RollupAddToken is the interface to call the smart contract function.
// `feeAddToken` is the amount of HEZ tokens that will be paid to add the
// token.  `feeAddToken` must match the public value of the smart contract.
func (c *RollupClient) RollupAddToken(tokenAddress ethCommon.Address, feeAddToken, deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.tokenHEZ.Nonces(nil, owner)
			if err != nil {
				return nil, err
			}
			tokenName := c.tokenHEZCfg.Name
			tokenAddr := c.tokenHEZCfg.Address
			chainid, _ := c.client.Client().ChainID(context.Background())
			digest, _ := createPermitDigest(tokenAddr, owner, spender, chainid, feeAddToken, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, feeAddToken, deadline, digest, signature)

			return c.hermez.AddToken(auth, tokenAddress, permit)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed add Token %w", err)
	}
	return tx, nil
}

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *RollupClient) RollupWithdrawMerkleProof(fromBJJ *babyjub.PublicKey, tokenID uint32, numExitRoot, idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			pkCompL := fromBJJ.Compress()
			pkCompB := common.SwapEndianness(pkCompL[:])
			babyPubKey := new(big.Int).SetBytes(pkCompB)
			numExitRootB := big.NewInt(numExitRoot)
			idxBig := big.NewInt(idx)
			return c.hermez.WithdrawMerkleProof(auth, tokenID, amount, babyPubKey, numExitRootB, siblings, idxBig, instantWithdraw)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed update WithdrawMerkleProof: %w", err)
	}
	return tx, nil
}

// RollupWithdrawCircuit is the interface to call the smart contract function
func (c *RollupClient) RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32, numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupL1UserTxERC20ETH is the interface to call the smart contract function
func (c *RollupClient) RollupL1UserTxERC20ETH(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			pkCompL := fromBJJ.Compress()
			pkCompB := common.SwapEndianness(pkCompL[:])
			babyPubKey := new(big.Int).SetBytes(pkCompB)
			fromIdxBig := big.NewInt(fromIdx)
			toIdxBig := big.NewInt(toIdx)
			loadAmountF, err := common.NewFloat16(loadAmount)
			if err != nil {
				return nil, err
			}
			amountF, err := common.NewFloat16(amount)
			if err != nil {
				return nil, err
			}
			if tokenID == 0 {
				auth.Value = loadAmount
			}
			var permit []byte
			return c.hermez.AddL1Transaction(auth, babyPubKey, fromIdxBig, uint16(loadAmountF),
				uint16(amountF), tokenID, toIdxBig, permit)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed add L1 Tx ERC20/ETH: %w", err)
	}
	return tx, nil
}

// RollupL1UserTxERC20Permit is the interface to call the smart contract function
func (c *RollupClient) RollupL1UserTxERC20Permit(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64, deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			pkCompL := fromBJJ.Compress()
			pkCompB := common.SwapEndianness(pkCompL[:])
			babyPubKey := new(big.Int).SetBytes(pkCompB)
			fromIdxBig := big.NewInt(fromIdx)
			toIdxBig := big.NewInt(toIdx)
			loadAmountF, err := common.NewFloat16(loadAmount)
			if err != nil {
				return nil, err
			}
			amountF, err := common.NewFloat16(amount)
			if err != nil {
				return nil, err
			}
			if tokenID == 0 {
				auth.Value = loadAmount
			}
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.tokenHEZ.Nonces(nil, owner)
			if err != nil {
				return nil, err
			}
			tokenName := c.tokenHEZCfg.Name
			tokenAddr := c.tokenHEZCfg.Address
			chainid, _ := c.client.Client().ChainID(context.Background())
			digest, _ := createPermitDigest(tokenAddr, owner, spender, chainid, amount, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			return c.hermez.AddL1Transaction(auth, babyPubKey, fromIdxBig, uint16(loadAmountF),
				uint16(amountF), tokenID, toIdxBig, permit)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed add L1 Tx ERC20Permit: %w", err)
	}
	return tx, nil
}

// RollupRegisterTokensCount is the interface to call the smart contract function
func (c *RollupClient) RollupRegisterTokensCount() (registerTokensCount *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		registerTokensCount, err = c.hermez.RegisterTokensCount(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return registerTokensCount, nil
}

// RollupLastForgedBatch is the interface to call the smart contract function
func (c *RollupClient) RollupLastForgedBatch() (lastForgedBatch int64, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_lastForgedBatch, err := c.hermez.LastForgedBatch(nil)
		lastForgedBatch = int64(_lastForgedBatch)
		return err
	}); err != nil {
		return 0, err
	}
	return lastForgedBatch, nil
}

// RollupUpdateForgeL1L2BatchTimeout is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout int64) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateForgeL1L2BatchTimeout(auth, uint8(newForgeL1L2BatchTimeout))
		},
	); err != nil {
		return nil, fmt.Errorf("Failed update ForgeL1L2BatchTimeout: %w", err)
	}
	return tx, nil
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *RollupClient) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.hermez.UpdateFeeAddToken(auth, newFeeAddToken)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed update FeeAddToken: %w", err)
	}
	return tx, nil
}

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *RollupClient) RollupConstants() (rollupConstants *common.RollupConstants, err error) {
	rollupConstants = new(common.RollupConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		absoluteMaxL1L2BatchTimeout, err := c.hermez.ABSOLUTEMAXL1L2BATCHTIMEOUT(nil)
		if err != nil {
			return err
		}
		rollupConstants.AbsoluteMaxL1L2BatchTimeout = int64(absoluteMaxL1L2BatchTimeout)
		rollupConstants.TokenHEZ, err = c.hermez.TokenHEZ(nil)
		if err != nil {
			return err
		}
		for i := int64(0); i < int64(common.LenVerifiers); i++ {
			var newRollupVerifier common.RollupVerifierStruct
			rollupVerifier, err := c.hermez.RollupVerifiers(nil, big.NewInt(i))
			if err != nil {
				return err
			}
			newRollupVerifier.MaxTx = rollupVerifier.MaxTx.Int64()
			newRollupVerifier.NLevels = rollupVerifier.NLevels.Int64()
			rollupConstants.Verifiers = append(rollupConstants.Verifiers, newRollupVerifier)
		}
		rollupConstants.HermezAuctionContract, err = c.hermez.HermezAuctionContract(nil)
		if err != nil {
			return err
		}
		rollupConstants.HermezGovernanceDAOAddress, err = c.hermez.HermezGovernanceDAOAddress(nil)
		if err != nil {
			return err
		}
		rollupConstants.SafetyAddress, err = c.hermez.SafetyAddress(nil)
		if err != nil {
			return err
		}
		rollupConstants.WithdrawDelayerContract, err = c.hermez.WithdrawDelayerContract(nil)
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
	var blockHash *ethCommon.Hash

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
		blockHash = &logs[0].BlockHash
	}
	for _, vLog := range logs {
		if vLog.BlockHash != *blockHash {
			log.Errorw("Block hash mismatch", "expected", blockHash.String(), "got", vLog.BlockHash.String())
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
			L1Tx, err := common.L1UserTxFromBytes(L1UserTxAux.L1UserTx)
			if err != nil {
				return nil, nil, err
			}
			toForgeL1TxsNum := new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			L1Tx.ToForgeL1TxsNum = &toForgeL1TxsNum
			L1Tx.Position = int(new(big.Int).SetBytes(vLog.Topics[2][:]).Int64())
			L1Tx.UserOrigin = true
			L1UserTx.L1UserTx = *L1Tx
			rollupEvents.L1UserTx = append(rollupEvents.L1UserTx, L1UserTx)
		case logHermezAddToken:
			var addToken RollupEventAddToken
			err := c.contractAbi.Unpack(&addToken, "AddToken", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			addToken.TokenAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			rollupEvents.AddToken = append(rollupEvents.AddToken, addToken)
		case logHermezForgeBatch:
			var forgeBatch RollupEventForgeBatch
			forgeBatch.BatchNum = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			forgeBatch.EthTxHash = vLog.TxHash
			// forgeBatch.Sender = vLog.Address
			rollupEvents.ForgeBatch = append(rollupEvents.ForgeBatch, forgeBatch)
		case logHermezUpdateForgeL1L2BatchTimeout:
			var updateForgeL1L2BatchTimeout struct {
				NewForgeL1L2BatchTimeout uint8
			}
			err := c.contractAbi.Unpack(&updateForgeL1L2BatchTimeout, "UpdateForgeL1L2BatchTimeout", vLog.Data)
			if err != nil {
				return nil, nil, err
			}
			rollupEvents.UpdateForgeL1L2BatchTimeout = append(rollupEvents.UpdateForgeL1L2BatchTimeout,
				RollupEventUpdateForgeL1L2BatchTimeout{
					NewForgeL1L2BatchTimeout: int64(updateForgeL1L2BatchTimeout.NewForgeL1L2BatchTimeout),
				})
		case logHermezUpdateFeeAddToken:
			var updateFeeAddToken RollupEventUpdateFeeAddToken
			err := c.contractAbi.Unpack(&updateFeeAddToken, "UpdateFeeAddToken", vLog.Data)
			if err != nil {
				return nil, nil, err
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
		}
	}
	return &rollupEvents, blockHash, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the
// Rollup Smart Contract in the given transaction, and the sender address.
func (c *RollupClient) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*RollupForgeBatchArgs, *ethCommon.Address, error) {
	tx, _, err := c.client.client.TransactionByHash(context.Background(), ethTxHash)
	if err != nil {
		return nil, nil, err
	}
	txData := tx.Data()
	method, err := c.contractAbi.MethodById(txData[:4])
	if err != nil {
		return nil, nil, err
	}
	receipt, err := c.client.client.TransactionReceipt(context.Background(), ethTxHash)
	if err != nil {
		return nil, nil, err
	}
	sender, err := c.client.client.TransactionSender(context.Background(), tx, receipt.Logs[0].BlockHash, receipt.Logs[0].Index)
	if err != nil {
		return nil, nil, err
	}
	var aux RollupForgeBatchArgsAux
	if err := method.Inputs.Unpack(&aux, txData[4:]); err != nil {
		return nil, nil, err
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
	numTxsL1 := len(aux.EncodedL1CoordinatorTx) / common.L1CoordinatorTxBytesLen
	for i := 0; i < numTxsL1; i++ {
		bytesL1Coordinator := aux.EncodedL1CoordinatorTx[i*common.L1CoordinatorTxBytesLen : (i+1)*common.L1CoordinatorTxBytesLen]
		var signature []byte
		v := bytesL1Coordinator[0]
		s := bytesL1Coordinator[1:33]
		r := bytesL1Coordinator[33:65]
		signature = append(signature, r[:]...)
		signature = append(signature, s[:]...)
		signature = append(signature, v)
		l1Tx, err := common.L1CoordinatorTxFromBytes(bytesL1Coordinator)
		if err != nil {
			return nil, nil, err
		}
		rollupForgeBatchArgs.L1CoordinatorTxs = append(rollupForgeBatchArgs.L1CoordinatorTxs, *l1Tx)
		rollupForgeBatchArgs.L1CoordinatorTxsAuths = append(rollupForgeBatchArgs.L1CoordinatorTxsAuths, signature)
	}
	rollupConsts, err := c.RollupConstants()
	if err != nil {
		return nil, nil, err
	}
	nLevels := rollupConsts.Verifiers[rollupForgeBatchArgs.VerifierIdx].NLevels
	lenL2TxsBytes := int((nLevels/8)*2 + 2 + 1)
	numTxsL2 := len(aux.L2TxsData) / lenL2TxsBytes
	for i := 0; i < numTxsL2; i++ {
		l2Tx, err := common.L2TxFromBytes(aux.L2TxsData[i*lenL2TxsBytes:(i+1)*lenL2TxsBytes], int(nLevels))
		if err != nil {
			return nil, nil, err
		}
		rollupForgeBatchArgs.L2TxsData = append(rollupForgeBatchArgs.L2TxsData, *l2Tx)
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
		feeIdxCoordinator, err := common.IdxFromBytes(paddedFeeIdx[:])
		if err != nil {
			return nil, nil, err
		}
		if feeIdxCoordinator != common.Idx(0) {
			rollupForgeBatchArgs.FeeIdxCoordinator = append(rollupForgeBatchArgs.FeeIdxCoordinator, feeIdxCoordinator)
		}
	}
	return &rollupForgeBatchArgs, &sender, nil
}
