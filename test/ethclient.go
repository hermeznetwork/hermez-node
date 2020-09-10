package test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sync"
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

func init() {
	copystructure.Copiers[reflect.TypeOf(big.Int{})] =
		func(raw interface{}) (interface{}, error) {
			in := raw.(big.Int)
			out := new(big.Int).Set(&in)
			return *out, nil
		}
}

// RollupBlock stores all the data related to the Rollup SC from an ethereum block
type RollupBlock struct {
	State     eth.RollupState
	Vars      eth.RollupVariables
	Events    eth.RollupEvents
	Constants *eth.RollupConstants
	Eth       *EthereumBlock
}

var (
	errBidClosed   = fmt.Errorf("Bid has already been closed")
	errBidNotOpen  = fmt.Errorf("Bid has not been opened yet")
	errBidBelowMin = fmt.Errorf("Bid below minimum")
	errCoordNotReg = fmt.Errorf("Coordinator not registered")
)

// AuctionBlock stores all the data related to the Auction SC from an ethereum block
type AuctionBlock struct {
	State     eth.AuctionState
	Vars      eth.AuctionVariables
	Events    eth.AuctionEvents
	Constants *eth.AuctionConstants
	Eth       *EthereumBlock
}

func (a *AuctionBlock) getSlotNumber(blockNumber int64) int64 {
	if a.Eth.BlockNum >= a.Constants.GenesisBlockNum {
		return (blockNumber - a.Constants.GenesisBlockNum) / int64(a.Constants.BlocksPerSlot)
	}
	return 0
}

func (a *AuctionBlock) getCurrentSlotNumber() int64 {
	return a.getSlotNumber(a.Eth.BlockNum)
}

func (a *AuctionBlock) getEpoch(slot int64) int64 {
	return slot % int64(len(a.Vars.MinBidEpoch))
}

func (a *AuctionBlock) getMinBidBySlot(slot int64) (*big.Int, error) {
	if slot < a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots) {
		return nil, errBidClosed
	}

	epoch := a.getEpoch(slot)
	var prevBid *big.Int
	slotState, ok := a.State.Slots[slot]
	// If the bidAmount for a slot is 0 it means that it has not yet been bid, so the midBid will be the minimum
	// bid for the slot time plus the outbidding set, otherwise it will be the bidAmount plus the outbidding
	if !ok || slotState.BidAmount.Cmp(big.NewInt(0)) == 0 {
		prevBid = a.Vars.MinBidEpoch[epoch]
	} else {
		prevBid = slotState.BidAmount
	}
	outBid := new(big.Int).Set(prevBid)
	outBid.Mul(outBid, big.NewInt(int64(a.Vars.Outbidding)))
	outBid.Div(outBid, big.NewInt(100)) //nolint:gomnd
	outBid.Add(prevBid, outBid)
	return outBid, nil
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

// Next prepares the successive block.
func (b *Block) Next() *Block {
	blockNextRaw, err := copystructure.Copy(b)
	if err != nil {
		panic(err)
	}
	blockNext := blockNextRaw.(*Block)
	blockNext.Rollup.Events = eth.NewRollupEvents()
	blockNext.Auction.Events = eth.NewAuctionEvents()
	blockNext.Eth = &EthereumBlock{
		BlockNum:   b.Eth.BlockNum + 1,
		ParentHash: b.Eth.Hash,
	}
	blockNext.Rollup.Constants = b.Rollup.Constants
	blockNext.Auction.Constants = b.Auction.Constants
	blockNext.Rollup.Eth = blockNext.Eth
	blockNext.Auction.Eth = blockNext.Eth
	return blockNext
}

// ClientSetup is used to initialize the constants of the Smart Contracts and
// other details of the test Client
type ClientSetup struct {
	RollupConstants  *eth.RollupConstants
	RollupVariables  *eth.RollupVariables
	AuctionConstants *eth.AuctionConstants
	AuctionVariables *eth.AuctionVariables
	VerifyProof      bool
}

// NewClientSetupExample returns a ClientSetup example with hardcoded realistic values.
// TODO: Fill all values that are currently default.
//nolint:gomnd
func NewClientSetupExample() *ClientSetup {
	rollupConstants := &eth.RollupConstants{}
	rollupVariables := &eth.RollupVariables{
		MaxTxVerifiers:     make([]int, 0),
		TokenHEZ:           ethCommon.Address{},
		GovernanceAddress:  ethCommon.Address{},
		SafetyBot:          ethCommon.Address{},
		ConsensusContract:  ethCommon.Address{},
		WithdrawalContract: ethCommon.Address{},
		FeeAddToken:        big.NewInt(1),
		ForgeL1Timeout:     16,
		FeeL1UserTx:        big.NewInt(2),
	}
	auctionConstants := &eth.AuctionConstants{
		BlocksPerSlot: 40,
	}
	auctionVariables := &eth.AuctionVariables{
		DonationAddress: ethCommon.Address{},
		BootCoordinator: ethCommon.Address{},
		MinBidEpoch: [6]*big.Int{
			big.NewInt(10), big.NewInt(11), big.NewInt(12),
			big.NewInt(13), big.NewInt(14), big.NewInt(15)},
		ClosedAuctionSlots: 2,
		OpenAuctionSlots:   100,
		AllocationRatio:    [3]uint8{},
		Outbidding:         10,
		SlotDeadline:       20,
	}
	return &ClientSetup{
		RollupConstants:  rollupConstants,
		RollupVariables:  rollupVariables,
		AuctionConstants: auctionConstants,
		AuctionVariables: auctionVariables,
	}
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
	rw               *sync.RWMutex
	log              bool
	addr             ethCommon.Address
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
func NewClient(l bool, timer Timer, addr ethCommon.Address, setup *ClientSetup) *Client {
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
			Vars:      *setup.RollupVariables,
			Events:    eth.NewRollupEvents(),
			Constants: setup.RollupConstants,
		},
		Auction: &AuctionBlock{
			State: eth.AuctionState{
				Slots:           make(map[int64]*eth.SlotState),
				PendingBalances: make(map[ethCommon.Address]*big.Int),
				Coordinators:    make(map[ethCommon.Address]*eth.Coordinator),
			},
			Vars:      *setup.AuctionVariables,
			Events:    eth.NewAuctionEvents(),
			Constants: setup.AuctionConstants,
		},
		Eth: &EthereumBlock{
			BlockNum:   blockNum,
			Time:       timer.Time(),
			Hash:       hasher.Next(),
			ParentHash: ethCommon.Hash{},
		},
	}
	blockCurrent.Rollup.Eth = blockCurrent.Eth
	blockCurrent.Auction.Eth = blockCurrent.Eth
	blocks[blockNum] = &blockCurrent
	blockNext := blockCurrent.Next()
	blocks[blockNum+1] = blockNext

	return &Client{
		rw:                    &sync.RWMutex{},
		log:                   l,
		addr:                  addr,
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

func (c *Client) currentBlock() *Block {
	return c.blocks[c.blockNum]
}

// CtlMineBlock moves one block forward
func (c *Client) CtlMineBlock() {
	c.rw.Lock()
	defer c.rw.Unlock()

	blockCurrent := c.nextBlock()
	c.blockNum++
	c.maxBlockNum = c.blockNum
	blockCurrent.Eth.Time = c.timer.Time()
	blockCurrent.Eth.Hash = c.hasher.Next()
	for ethTxHash, forgeBatchArgs := range c.forgeBatchArgsPending {
		c.forgeBatchArgs[ethTxHash] = forgeBatchArgs
	}
	c.forgeBatchArgsPending = make(map[ethCommon.Hash]*eth.RollupForgeBatchArgs)

	blockNext := blockCurrent.Next()
	c.blocks[c.blockNum+1] = blockNext
	c.Debugw("TestClient mined block", "blockNum", c.blockNum)
}

// CtlRollback discards the last mined block.  Use this to replace a mined
// block to simulate reorgs.
func (c *Client) CtlRollback() {
	c.rw.Lock()
	defer c.rw.Unlock()

	if c.blockNum == 0 {
		panic("Can't rollback at blockNum = 0")
	}
	delete(c.blocks, c.blockNum+1) // delete next block
	delete(c.blocks, c.blockNum)   // delete current block
	c.blockNum--
	blockCurrent := c.blocks[c.blockNum]
	blockNext := blockCurrent.Next()
	c.blocks[c.blockNum+1] = blockNext
}

//
// Ethereum
//

// EthCurrentBlock returns the current blockNum
func (c *Client) EthCurrentBlock() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

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
	c.rw.RLock()
	defer c.rw.RUnlock()

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
	c.rw.Lock()
	defer c.rw.Unlock()

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

type transactionData struct {
	Name  string
	Value interface{}
}

func (c *Client) newTransaction(name string, value interface{}) *types.Transaction {
	data, err := json.Marshal(transactionData{name, value})
	if err != nil {
		panic(err)
	}
	return types.NewTransaction(0, ethCommon.Address{}, nil, 0, nil,
		data)
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *Client) RollupForgeBatch(*eth.RollupForgeBatchArgs) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// CtlAddBatch adds forged batch to the Rollup, without checking any ZKProof
func (c *Client) CtlAddBatch(args *eth.RollupForgeBatchArgs) {
	c.rw.Lock()
	defer c.rw.Unlock()

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
	c.rw.Lock()
	defer c.rw.Unlock()

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
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupForceExit is the interface to call the smart contract function
func (c *Client) RollupForceExit(fromIdx int64, amountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupForceTransfer is the interface to call the smart contract function
func (c *Client) RollupForceTransfer(fromIdx int64, amountF utils.Float16, tokenID, toIdx int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupCreateAccountDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupDepositTransfer(fromIdx int64, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupDeposit is the interface to call the smart contract function
func (c *Client) RollupDeposit(fromIdx int64, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupCreateAccountDepositFromRelayer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte, babyPubKey babyjub.PublicKey, loadAmountF utils.Float16) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupCreateAccountDeposit is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupGetTokenAddress is the interface to call the smart contract function
func (c *Client) RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// RollupGetL1TxFromQueue is the interface to call the smart contract function
func (c *Client) RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// RollupGetQueue is the interface to call the smart contract function
func (c *Client) RollupGetQueue(queue int64) ([]byte, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// RollupUpdateForgeL1Timeout is the interface to call the smart contract function
func (c *Client) RollupUpdateForgeL1Timeout(newForgeL1Timeout int64) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupUpdateFeeL1UserTx is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeL1UserTx(newFeeL1UserTx *big.Int) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupUpdateTokensHEZ is the interface to call the smart contract function
func (c *Client) RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// RollupUpdateGovernance is the interface to call the smart contract function
// func (c *Client) RollupUpdateGovernance() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *Client) RollupConstants() (*eth.RollupConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *Client) RollupEventsByBlock(blockNum int64) (*eth.RollupEvents, *ethCommon.Hash, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, nil, fmt.Errorf("Block %v doesn't exist", blockNum)
	}
	return &block.Rollup.Events, &block.Eth.Hash, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract in the given transaction
func (c *Client) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*eth.RollupForgeBatchArgs, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

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
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionGetSlotDeadline() (uint8, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return 0, errTODO
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetOpenAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return 0, errTODO
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetClosedAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return 0, errTODO
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionGetOutbidding() (uint8, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionGetAllocationRatio() ([3]uint8, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return [3]uint8{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionGetBootCoordinator() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	currentBlock := c.currentBlock()
	a := currentBlock.Auction

	return &a.Vars.BootCoordinator, nil
}

// AuctionChangeEpochMinBid is the interface to call the smart contract function
func (c *Client) AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *Client) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	nextBlock := c.nextBlock()
	a := nextBlock.Auction

	if _, ok := a.State.Coordinators[forgerAddress]; ok {
		return nil, fmt.Errorf("Already registered")
	}
	a.State.Coordinators[forgerAddress] = &eth.Coordinator{
		WithdrawalAddress: c.addr,
		URL:               URL,
	}

	a.Events.NewCoordinator = append(a.Events.NewCoordinator,
		eth.AuctionEventNewCoordinator{
			ForgerAddress:     forgerAddress,
			WithdrawalAddress: c.addr,
			URL:               URL,
		})

	type data struct {
		ForgerAddress ethCommon.Address
		URL           string
	}
	return c.newTransaction("registercoordinator", data{forgerAddress, URL}), nil
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *Client) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return false, errTODO
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *Client) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *Client) AuctionGetCurrentSlotNumber() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return 0, errTODO
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// AuctionGetMinBidEpoch is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidEpoch(epoch uint8) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *Client) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBid is the interface to call the smart contract function
func (c *Client) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	nextBlock := c.nextBlock()
	a := nextBlock.Auction

	if slot < a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots) {
		return nil, errBidClosed
	}

	if slot >= a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots)+int64(a.Vars.OpenAuctionSlots) {
		return nil, errBidNotOpen
	}

	minBid, err := a.getMinBidBySlot(slot)
	if err != nil {
		return nil, err
	}
	if bidAmount.Cmp(minBid) == -1 {
		return nil, errBidBelowMin
	}

	if _, ok := a.State.Coordinators[forger]; !ok {
		return nil, errCoordNotReg
	}

	slotState, ok := a.State.Slots[slot]
	if !ok {
		slotState = &eth.SlotState{}
		a.State.Slots[slot] = slotState
	}
	slotState.Forger = forger
	slotState.BidAmount = bidAmount

	a.Events.NewBid = append(a.Events.NewBid,
		eth.AuctionEventNewBid{Slot: slot, BidAmount: bidAmount, CoordinatorForger: forger})

	type data struct {
		Slot      int64
		BidAmount *big.Int
		Forger    ethCommon.Address
	}
	return c.newTransaction("bid", data{slot, bidAmount, forger}), nil
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *Client) AuctionMultiBid(startingSlot int64, endingSlot int64, slotEpoch [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *Client) AuctionCanForge(forger ethCommon.Address) (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return false, errTODO
}

// AuctionForge is the interface to call the smart contract function
// func (c *Client) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *Client) AuctionClaimHEZ() (*types.Transaction, error) {
	c.rw.Lock()
	defer c.rw.Unlock()

	return nil, errTODO
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *Client) AuctionConstants() (*eth.AuctionConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.auctionConstants, nil
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *Client) AuctionEventsByBlock(blockNum int64) (*eth.AuctionEvents, *ethCommon.Hash, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, nil, fmt.Errorf("Block %v doesn't exist", blockNum)
	}
	return &block.Auction.Events, &block.Eth.Hash, nil
}
