package test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/utils"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/mitchellh/copystructure"
)

// RollupBlock stores all the data related to the Rollup SC from an ethereum block
type RollupBlock struct {
	State  eth.RollupState
	Vars   eth.RollupVariables
	Events eth.RollupEvents
}

// AuctionBlock stores all the data related to the Auction SC from an ethereum block
type AuctionBlock struct {
	State  eth.AuctionState
	Vars   eth.AuctionVariables
	Events eth.AuctionEvents
}

// EthereumBlock stores all the generic data related to the an ethereum block
type EthereumBlock struct {
	BlockNum   int64
	Time       int64
	Hash       ethCommon.Hash
	ParentHash ethCommon.Hash
	// state      ethState
}

// Block represents a ethereum block
type Block struct {
	Rollup  *RollupBlock
	Auction *AuctionBlock
	Eth     *EthereumBlock
}

// type ethState struct {
// 	blockNum int64
// }

// type state struct {
// 	rollupState  eth.RollupState
// 	rollupVars   eth.RollupVariables
// 	auctionState eth.AuctionState
// 	auctionVars  eth.AuctionVariables
// 	eth          ethState
// }

// ClientSetup is used to initialize the constants of the Smart Contracts and
// other details of the test Client
type ClientSetup struct {
	RollupConstants  *eth.RollupConstants
	RollupVariables  *eth.RollupVariables
	AuctionConstants *eth.AuctionConstants
	AuctionVariables *eth.AuctionVariables
	VerifyProof      bool
}

// Timer is an interface to simulate a source of time, useful to advance time
// virtually.
type Timer interface {
	Time() int64
}

// type forgeBatchArgs struct {
// 	ethTx     *types.Transaction
// 	blockNum  int64
// 	blockHash ethCommon.Hash
// }

// Client implements the eth.ClientInterface interface, allowing to manipulate the
// values for testing, working with deterministic results.
type Client struct {
	log              bool
	rollupConstants  *eth.RollupConstants
	auctionConstants *eth.AuctionConstants
	blocks           map[int64]*Block
	// state            state
	blockNum    int64 // last mined block num
	maxBlockNum int64 // highest block num calculated
	timer       Timer
	hasher      hasher

	forgeBatchArgsPending map[ethCommon.Hash]*eth.RollupForgeBatchArgs
	forgeBatchArgs        map[ethCommon.Hash]*eth.RollupForgeBatchArgs
}

// NewClient returns a new test Client that implements the eth.IClient
// interface, at the given initialBlockNumber.
func NewClient(l bool, timer Timer, setup *ClientSetup) *Client {
	blocks := make(map[int64]*Block)
	blockNum := int64(0)

	hasher := hasher{}
	// Add ethereum genesis block
	mapL1TxQueue := make(map[int64]*eth.QueueStruct)
	mapL1TxQueue[0] = eth.NewQueueStruct()
	mapL1TxQueue[1] = eth.NewQueueStruct()
	blockCurrent := Block{
		Rollup: &RollupBlock{
			State: eth.RollupState{
				StateRoot:              big.NewInt(0),
				ExitRoots:              make([]*big.Int, 0),
				ExitNullifierMap:       make(map[[256 / 8]byte]bool),
				TokenList:              make([]ethCommon.Address, 0),
				TokenMap:               make(map[ethCommon.Address]bool),
				MapL1TxQueue:           mapL1TxQueue,
				LastL1L2Batch:          0,
				CurrentToForgeL1TxsNum: 0,
				LastToForgeL1TxsNum:    1,
				CurrentIdx:             0,
			},
			Vars:   *setup.RollupVariables,
			Events: eth.NewRollupEvents(),
		},
		Auction: &AuctionBlock{
			State: eth.AuctionState{
				Slots:           make(map[int64]eth.SlotState),
				PendingBalances: make(map[ethCommon.Address]*big.Int),
				Coordinators:    make(map[ethCommon.Address]eth.Coordinator),
			},
			Vars:   *setup.AuctionVariables,
			Events: eth.NewAuctionEvents(),
		},
		Eth: &EthereumBlock{
			BlockNum:   blockNum,
			Time:       timer.Time(),
			Hash:       hasher.Next(),
			ParentHash: ethCommon.Hash{},
		},
	}
	blocks[blockNum] = &blockCurrent
	blockNextRaw, err := copystructure.Copy(&blockCurrent)
	if err != nil {
		panic(err)
	}
	blockNext := blockNextRaw.(*Block)
	blocks[blockNum+1] = blockNext

	return &Client{
		log:                   l,
		rollupConstants:       setup.RollupConstants,
		auctionConstants:      setup.AuctionConstants,
		blocks:                blocks,
		timer:                 timer,
		hasher:                hasher,
		forgeBatchArgsPending: make(map[ethCommon.Hash]*eth.RollupForgeBatchArgs),
		forgeBatchArgs:        make(map[ethCommon.Hash]*eth.RollupForgeBatchArgs),
	}
}

//
// Mock Control
//

// Debugf calls log.Debugf if c.log is true
func (c *Client) Debugf(template string, args ...interface{}) {
	if c.log {
		log.Debugf(template, args...)
	}
}

// Debugw calls log.Debugw if c.log is true
func (c *Client) Debugw(template string, kv ...interface{}) {
	if c.log {
		log.Debugw(template, kv...)
	}
}

type hasher struct {
	counter uint64
}

// Next returns the next hash
func (h *hasher) Next() ethCommon.Hash {
	var hash ethCommon.Hash
	binary.LittleEndian.PutUint64(hash[:], h.counter)
	h.counter++
	return hash
}

func (c *Client) nextBlock() *Block {
	return c.blocks[c.blockNum+1]
}

// CtlMineBlock moves one block forward
func (c *Client) CtlMineBlock() {
	blockCurrent := c.nextBlock()
	c.blockNum++
	c.maxBlockNum = c.blockNum
	blockCurrent.Eth = &EthereumBlock{
		BlockNum:   c.blockNum,
		Time:       c.timer.Time(),
		Hash:       c.hasher.Next(),
		ParentHash: blockCurrent.Eth.Hash,
	}
	for ethTxHash, forgeBatchArgs := range c.forgeBatchArgsPending {
		c.forgeBatchArgs[ethTxHash] = forgeBatchArgs
	}
	c.forgeBatchArgsPending = make(map[ethCommon.Hash]*eth.RollupForgeBatchArgs)

	blockNextRaw, err := copystructure.Copy(blockCurrent)
	if err != nil {
		panic(err)
	}
	blockNext := blockNextRaw.(*Block)
	blockNext.Rollup.Events = eth.NewRollupEvents()
	blockNext.Auction.Events = eth.NewAuctionEvents()
	c.blocks[c.blockNum+1] = blockNext
	c.Debugw("TestClient mined block", "blockNum", c.blockNum)
}

// CtlRollback discards the last mined block.  Use this to replace a mined
// block to simulate reorgs.
func (c *Client) CtlRollback() {
	if c.blockNum == 0 {
		panic("Can't rollback at blockNum = 0")
	}
	delete(c.blocks, c.blockNum+1) // delete next block
	delete(c.blocks, c.blockNum)   // delete current block
	c.blockNum--
	blockCurrent := c.blocks[c.blockNum]
	blockNextRaw, err := copystructure.Copy(blockCurrent)
	if err != nil {
		panic(err)
	}
	blockNext := blockNextRaw.(*Block)
	blockNext.Rollup.Events = eth.NewRollupEvents()
	blockNext.Auction.Events = eth.NewAuctionEvents()
	c.blocks[c.blockNum+1] = blockNext
}

//
// Ethereum
//

// EthCurrentBlock returns the current blockNum
func (c *Client) EthCurrentBlock() (int64, error) {
	if c.blockNum < c.maxBlockNum {
		panic("blockNum has decreased.  " +
			"After a rollback you must mine to reach the same or higher blockNum")
	}
	return c.blockNum, nil
}

// func newHeader(number *big.Int) *types.Header {
// 	return &types.Header{
// 		Number: number,
// 		Time:   uint64(number.Int64()),
// 	}
// }

// EthHeaderByNumber returns the *types.Header for the given block number in a
// deterministic way.
// func (c *Client) EthHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
// 	return newHeader(number), nil
// }

// EthBlockByNumber returns the *common.Block for the given block number in a
// deterministic way.
func (c *Client) EthBlockByNumber(ctx context.Context, blockNum int64) (*common.Block, error) {
	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, fmt.Errorf("block not found")
	}
	return &common.Block{
		EthBlockNum: blockNum,
		Timestamp:   time.Unix(block.Eth.Time, 0),
		Hash:        block.Eth.Hash,
		ParentHash:  block.Eth.ParentHash,
	}, nil
}

var errTODO = fmt.Errorf("TODO: Not implemented yet")

//
// Rollup
//

// CtlAddL1TxUser adds an L1TxUser to the L1UserTxs queue of the Rollup
func (c *Client) CtlAddL1TxUser(l1Tx *common.L1Tx) {
	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	queue := r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
	if len(queue.L1TxQueue) >= c.rollupConstants.MaxL1UserTx {
		r.State.LastToForgeL1TxsNum++
		r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum] = eth.NewQueueStruct()
		queue = r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
	}
	if int64(l1Tx.FromIdx) > r.State.CurrentIdx {
		panic("l1Tx.FromIdx > r.State.CurrentIdx")
	}
	if int(l1Tx.TokenID)+1 > len(r.State.TokenList) {
		panic("l1Tx.TokenID + 1 > len(r.State.TokenList)")
	}
	queue.L1TxQueue = append(queue.L1TxQueue, *l1Tx)
	r.Events.L1UserTx = append(r.Events.L1UserTx, eth.RollupEventL1UserTx{L1Tx: *l1Tx})
}

func (c *Client) newTransaction(name string, value interface{}) *types.Transaction {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return types.NewTransaction(0, ethCommon.Address{}, nil, 0, nil,
		data)
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *Client) RollupForgeBatch(*eth.RollupForgeBatchArgs) (*types.Transaction, error) {
	return nil, errTODO
}

// CtlAddBatch adds forged batch to the Rollup, without checking any ZKProof
func (c *Client) CtlAddBatch(args *eth.RollupForgeBatchArgs) {
	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	r.State.StateRoot = args.NewStRoot
	if args.NewLastIdx < r.State.CurrentIdx {
		panic("args.NewLastIdx < r.State.CurrentIdx")
	}
	r.State.CurrentIdx = args.NewLastIdx
	r.State.ExitRoots = append(r.State.ExitRoots, args.NewExitRoot)
	if args.L1Batch {
		r.State.CurrentToForgeL1TxsNum++
		if r.State.CurrentToForgeL1TxsNum == r.State.LastToForgeL1TxsNum {
			r.State.LastToForgeL1TxsNum++
			r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum] = eth.NewQueueStruct()
		}
	}
	ethTx := c.newTransaction("forgebatch", args)
	c.forgeBatchArgsPending[ethTx.Hash()] = args
	r.Events.ForgeBatch = append(r.Events.ForgeBatch, eth.RollupEventForgeBatch{
		BatchNum:  int64(len(r.State.ExitRoots)),
		EthTxHash: ethTx.Hash(),
	})
}

// RollupAddToken is the interface to call the smart contract function
func (c *Client) RollupAddToken(tokenAddress ethCommon.Address) (*types.Transaction, error) {
	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	if _, ok := r.State.TokenMap[tokenAddress]; ok {
		return nil, fmt.Errorf("Token %v already registered", tokenAddress)
	}

	r.State.TokenMap[tokenAddress] = true
	r.State.TokenList = append(r.State.TokenList, tokenAddress)
	r.Events.AddToken = append(r.Events.AddToken, eth.RollupEventAddToken{Address: tokenAddress,
		TokenID: uint32(len(r.State.TokenList) - 1)})
	return c.newTransaction("addtoken", tokenAddress), nil
}

// RollupWithdrawSNARK is the interface to call the smart contract function
// func (c *Client) RollupWithdrawSNARK() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *Client) RollupWithdrawMerkleProof(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey, numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupForceExit is the interface to call the smart contract function
func (c *Client) RollupForceExit(fromIdx int64, amountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupForceTransfer is the interface to call the smart contract function
func (c *Client) RollupForceTransfer(fromIdx int64, amountF utils.Float16, tokenID, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupDepositTransfer(fromIdx int64, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupDeposit is the interface to call the smart contract function
func (c *Client) RollupDeposit(fromIdx int64, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDepositFromRelayer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte, babyPubKey babyjub.PublicKey, loadAmountF utils.Float16) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDeposit is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupGetTokenAddress is the interface to call the smart contract function
func (c *Client) RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error) {
	return nil, errTODO
}

// RollupGetL1TxFromQueue is the interface to call the smart contract function
func (c *Client) RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error) {
	return nil, errTODO
}

// RollupGetQueue is the interface to call the smart contract function
func (c *Client) RollupGetQueue(queue int64) ([]byte, error) {
	return nil, errTODO
}

// RollupUpdateForgeL1Timeout is the interface to call the smart contract function
func (c *Client) RollupUpdateForgeL1Timeout(newForgeL1Timeout int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateFeeL1UserTx is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeL1UserTx(newFeeL1UserTx *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateTokensHEZ is the interface to call the smart contract function
func (c *Client) RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateGovernance is the interface to call the smart contract function
// func (c *Client) RollupUpdateGovernance() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *Client) RollupConstants() (*eth.RollupConstants, error) {
	return nil, errTODO
}

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *Client) RollupEventsByBlock(blockNum int64) (*eth.RollupEvents, *ethCommon.Hash, error) {
	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, nil, fmt.Errorf("Block %v doesn't exist", blockNum)
	}
	return &block.Rollup.Events, &block.Eth.Hash, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract in the given transaction
func (c *Client) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*eth.RollupForgeBatchArgs, error) {
	forgeBatchArgs, ok := c.forgeBatchArgs[ethTxHash]
	if !ok {
		return nil, fmt.Errorf("transaction not found")
	}
	return forgeBatchArgs, nil
}

//
// Auction
//

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionGetSlotDeadline() (uint8, error) {
	return 0, errTODO
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetOpenAuctionSlots() (uint16, error) { return 0, errTODO }

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetClosedAuctionSlots() (uint16, error) {
	return 0, errTODO
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionGetOutbidding() (uint8, error) {
	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionGetAllocationRatio() ([3]uint8, error) {
	return [3]uint8{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionGetBootCoordinator() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionChangeEpochMinBid is the interface to call the smart contract function
func (c *Client) AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *Client) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *Client) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *Client) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *Client) AuctionGetCurrentSlotNumber() (int64, error) {
	return 0, errTODO
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	return nil, errTODO
}

// AuctionGetMinBidEpoch is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidEpoch(epoch uint8) (*big.Int, error) {
	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *Client) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBid is the interface to call the smart contract function
func (c *Client) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *Client) AuctionMultiBid(startingSlot int64, endingSlot int64, slotEpoch [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *Client) AuctionCanForge(forger ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionForge is the interface to call the smart contract function
// func (c *Client) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *Client) AuctionClaimHEZ() (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *Client) AuctionConstants() (*eth.AuctionConstants, error) {
	return nil, errTODO
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *Client) AuctionEventsByBlock(blockNum int64) (*eth.AuctionEvents, *ethCommon.Hash, error) {
	return nil, nil, errTODO
}
