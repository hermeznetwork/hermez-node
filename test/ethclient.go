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

	"github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
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
	Txs       map[ethCommon.Hash]*types.Transaction
	Constants *eth.RollupPublicConstants
	Eth       *EthereumBlock
}

func (r *RollupBlock) addTransaction(tx *types.Transaction) *types.Transaction {
	txHash := tx.Hash()
	r.Txs[txHash] = tx
	return tx
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
	Txs       map[ethCommon.Hash]*types.Transaction
	Constants *eth.AuctionConstants
	Eth       *EthereumBlock
}

func (a *AuctionBlock) addTransaction(tx *types.Transaction) *types.Transaction {
	txHash := tx.Hash()
	a.Txs[txHash] = tx
	return tx
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

func (a *AuctionBlock) getSlotSet(slot int64) int64 {
	return slot % int64(len(a.Vars.DefaultSlotSetBid))
}

func (a *AuctionBlock) getMinBidBySlot(slot int64) (*big.Int, error) {
	if slot < a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots) {
		return nil, errBidClosed
	}

	slotSet := a.getSlotSet(slot)
	// fmt.Println("slot:", slot, "slotSet:", slotSet)
	var prevBid *big.Int
	slotState, ok := a.State.Slots[slot]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slot] = slotState
	}
	// If the bidAmount for a slot is 0 it means that it has not yet been bid, so the midBid will be the minimum
	// bid for the slot time plus the outbidding set, otherwise it will be the bidAmount plus the outbidding
	if slotState.BidAmount.Cmp(big.NewInt(0)) == 0 {
		prevBid = a.Vars.DefaultSlotSetBid[slotSet]
	} else {
		prevBid = slotState.BidAmount
	}
	outBid := new(big.Int).Set(prevBid)
	// fmt.Println("outBid:", outBid)
	outBid.Mul(outBid, big.NewInt(int64(a.Vars.Outbidding)))
	outBid.Div(outBid, big.NewInt(10000)) //nolint:gomnd
	outBid.Add(prevBid, outBid)
	// fmt.Println("minBid:", outBid)
	return outBid, nil
}

func (a *AuctionBlock) forge(forger ethCommon.Address) error {
	if ok, err := a.canForge(forger, a.Eth.BlockNum); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("Can't forge")
	}

	slotToForge := a.getSlotNumber(a.Eth.BlockNum)
	slotState, ok := a.State.Slots[slotToForge]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slotToForge] = slotState
	}
	slotState.Fulfilled = true

	a.Events.NewForge = append(a.Events.NewForge, eth.AuctionEventNewForge{
		Forger:      forger,
		SlotToForge: slotToForge,
	})
	return nil
}

func (a *AuctionBlock) canForge(forger ethCommon.Address, blockNum int64) (bool, error) {
	if blockNum < a.Constants.GenesisBlockNum {
		return false, fmt.Errorf("Auction has not started yet")
	}

	slotToForge := a.getSlotNumber(blockNum)
	// Get the relativeBlock to check if the slotDeadline has been exceeded
	relativeBlock := blockNum - (a.Constants.GenesisBlockNum + (slotToForge * int64(a.Constants.BlocksPerSlot)))

	// If the closedMinBid is 0 it means that we have to take as minBid the one that is set for this slot set,
	// otherwise the one that has been saved will be used
	var minBid *big.Int
	slotState, ok := a.State.Slots[slotToForge]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slotToForge] = slotState
	}
	if slotState.ClosedMinBid.Cmp(big.NewInt(0)) == 0 {
		minBid = a.Vars.DefaultSlotSetBid[a.getSlotSet(slotToForge)]
	} else {
		minBid = slotState.ClosedMinBid
	}

	if !slotState.Fulfilled && (relativeBlock >= int64(a.Vars.SlotDeadline)) {
		// if the relative block has exceeded the slotDeadline and no batch has been forged, anyone can forge
		return true, nil
		// TODO, find the forger set by the Bidder
	} else if coord, ok := a.State.Coordinators[slotState.Bidder]; ok &&
		coord.Forger == forger && slotState.BidAmount.Cmp(minBid) >= 0 {
		// if forger bidAmount has exceeded the minBid it can forge
		return true, nil
	} else if a.Vars.BootCoordinator == forger && slotState.BidAmount.Cmp(minBid) == -1 {
		// if it's the boot coordinator and it has not been bid or the bid is below the minimum it can forge
		return true, nil
	} else {
		return false, nil
	}
}

// EthereumBlock stores all the generic data related to the an ethereum block
type EthereumBlock struct {
	BlockNum   int64
	Time       int64
	Hash       ethCommon.Hash
	ParentHash ethCommon.Hash
	Tokens     map[ethCommon.Address]eth.ERC20Consts
	// state      ethState
}

// Block represents a ethereum block
type Block struct {
	Rollup  *RollupBlock
	Auction *AuctionBlock
	Eth     *EthereumBlock
}

func (b *Block) copy() *Block {
	bCopyRaw, err := copystructure.Copy(b)
	if err != nil {
		panic(err)
	}
	bCopy := bCopyRaw.(*Block)
	return bCopy
}

// Next prepares the successive block.
func (b *Block) Next() *Block {
	blockNext := b.copy()
	blockNext.Rollup.Events = eth.NewRollupEvents()
	blockNext.Auction.Events = eth.NewAuctionEvents()

	blockNext.Eth.BlockNum = b.Eth.BlockNum + 1
	blockNext.Eth.ParentHash = b.Eth.Hash

	blockNext.Rollup.Constants = b.Rollup.Constants
	blockNext.Auction.Constants = b.Auction.Constants
	blockNext.Rollup.Eth = blockNext.Eth
	blockNext.Auction.Eth = blockNext.Eth

	return blockNext
}

// ClientSetup is used to initialize the constants of the Smart Contracts and
// other details of the test Client
type ClientSetup struct {
	RollupConstants  *eth.RollupPublicConstants
	RollupVariables  *eth.RollupVariables
	AuctionConstants *eth.AuctionConstants
	AuctionVariables *eth.AuctionVariables
	VerifyProof      bool
}

// NewClientSetupExample returns a ClientSetup example with hardcoded realistic
// values.  With this setup, the rollup genesis will be block 1, and block 0
// and 1 will be premined.
//nolint:gomnd
func NewClientSetupExample() *ClientSetup {
	// rfield, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	// if !ok {
	// 	panic("bad rfield")
	// }
	initialMinimalBidding, ok := new(big.Int).SetString("10000000000000000000", 10) // 10 * (1e18)
	if !ok {
		panic("bad initialMinimalBidding")
	}
	tokenHEZ := ethCommon.HexToAddress("0x51D243D62852Bba334DD5cc33f242BAc8c698074")
	governanceAddress := ethCommon.HexToAddress("0x688EfD95BA4391f93717CF02A9aED9DBD2855cDd")
	rollupConstants := &eth.RollupPublicConstants{
		Verifiers: []eth.RollupVerifierStruct{
			{
				MaxTx:   2048,
				NLevels: 32,
			},
		},
		TokenHEZ:                   tokenHEZ,
		HermezGovernanceDAOAddress: governanceAddress,
		SafetyAddress:              ethCommon.HexToAddress("0x84d8B79E84fe87B14ad61A554e740f6736bF4c20"),
		HermezAuctionContract:      ethCommon.HexToAddress("0x8E442975805fb1908f43050c9C1A522cB0e28D7b"),
		WithdrawDelayerContract:    ethCommon.HexToAddress("0x5CB7979cBdbf65719BEE92e4D15b7b7Ed3D79114"),
	}
	rollupVariables := &eth.RollupVariables{
		FeeAddToken:           big.NewInt(11),
		ForgeL1L2BatchTimeout: 9,
		WithdrawalDelay:       80,
	}
	auctionConstants := &eth.AuctionConstants{
		BlocksPerSlot:         40,
		InitialMinimalBidding: initialMinimalBidding,
		GenesisBlockNum:       1,
		GovernanceAddress:     governanceAddress,
		TokenHEZ:              tokenHEZ,
		HermezRollup:          ethCommon.HexToAddress("0x474B6e29852257491cf283EfB1A9C61eBFe48369"),
	}
	auctionVariables := &eth.AuctionVariables{
		DonationAddress: ethCommon.HexToAddress("0x61Ed87CF0A1496b49A420DA6D84B58196b98f2e7"),
		BootCoordinator: ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		DefaultSlotSetBid: [6]*big.Int{
			big.NewInt(1000), big.NewInt(1100), big.NewInt(1200),
			big.NewInt(1300), big.NewInt(1400), big.NewInt(1500)},
		ClosedAuctionSlots: 2,
		OpenAuctionSlots:   4320,
		AllocationRatio:    [3]uint16{4000, 4000, 2000},
		Outbidding:         1000,
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

type batch struct {
	ForgeBatchArgs eth.RollupForgeBatchArgs
	Sender         ethCommon.Address
}

// Client implements the eth.ClientInterface interface, allowing to manipulate the
// values for testing, working with deterministic results.
type Client struct {
	rw               *sync.RWMutex
	log              bool
	addr             *ethCommon.Address
	rollupConstants  *eth.RollupPublicConstants
	auctionConstants *eth.AuctionConstants
	blocks           map[int64]*Block
	// state            state
	blockNum    int64 // last mined block num
	maxBlockNum int64 // highest block num calculated
	timer       Timer
	hasher      hasher

	forgeBatchArgsPending map[ethCommon.Hash]*batch
	forgeBatchArgs        map[ethCommon.Hash]*batch
}

// NewClient returns a new test Client that implements the eth.IClient
// interface, at the given initialBlockNumber.
func NewClient(l bool, timer Timer, addr *ethCommon.Address, setup *ClientSetup) *Client {
	blocks := make(map[int64]*Block)
	blockNum := int64(0)

	hasher := hasher{}
	// Add ethereum genesis block
	mapL1TxQueue := make(map[int64]*eth.QueueStruct)
	mapL1TxQueue[0] = eth.NewQueueStruct()
	mapL1TxQueue[1] = eth.NewQueueStruct()
	blockCurrent := &Block{
		Rollup: &RollupBlock{
			State: eth.RollupState{
				StateRoot:        big.NewInt(0),
				ExitRoots:        make([]*big.Int, 0),
				ExitNullifierMap: make(map[[256 / 8]byte]bool),
				// TokenID = 0 is ETH.  Set first entry in TokenList with 0x0 address for ETH.
				TokenList:              []ethCommon.Address{{}},
				TokenMap:               make(map[ethCommon.Address]bool),
				MapL1TxQueue:           mapL1TxQueue,
				LastL1L2Batch:          0,
				CurrentToForgeL1TxsNum: 0,
				LastToForgeL1TxsNum:    1,
				CurrentIdx:             0,
			},
			Vars:      *setup.RollupVariables,
			Txs:       make(map[ethCommon.Hash]*types.Transaction),
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
			Txs:       make(map[ethCommon.Hash]*types.Transaction),
			Events:    eth.NewAuctionEvents(),
			Constants: setup.AuctionConstants,
		},
		Eth: &EthereumBlock{
			BlockNum:   blockNum,
			Time:       timer.Time(),
			Hash:       hasher.Next(),
			ParentHash: ethCommon.Hash{},
			Tokens:     make(map[ethCommon.Address]eth.ERC20Consts),
		},
	}
	blockCurrent.Rollup.Eth = blockCurrent.Eth
	blockCurrent.Auction.Eth = blockCurrent.Eth
	blocks[blockNum] = blockCurrent
	blockNext := blockCurrent.Next()
	blocks[blockNum+1] = blockNext

	c := Client{
		rw:                    &sync.RWMutex{},
		log:                   l,
		addr:                  addr,
		rollupConstants:       setup.RollupConstants,
		auctionConstants:      setup.AuctionConstants,
		blocks:                blocks,
		timer:                 timer,
		hasher:                hasher,
		forgeBatchArgsPending: make(map[ethCommon.Hash]*batch),
		forgeBatchArgs:        make(map[ethCommon.Hash]*batch),
		blockNum:              blockNum,
		maxBlockNum:           blockNum,
	}

	for i := int64(1); i < setup.AuctionConstants.GenesisBlockNum+1; i++ {
		c.CtlMineBlock()
	}

	return &c
}

//
// Mock Control
//

func (c *Client) setNextBlock(block *Block) {
	c.blocks[c.blockNum+1] = block
}

func (c *Client) revertIfErr(err error, block *Block) {
	if err != nil {
		log.Infow("TestClient revert", "block", block.Eth.BlockNum, "err", err)
		c.setNextBlock(block)
	}
}

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

// CtlSetAddr sets the address of the client
func (c *Client) CtlSetAddr(addr ethCommon.Address) {
	c.addr = &addr
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
	c.forgeBatchArgsPending = make(map[ethCommon.Hash]*batch)

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

// EthTransactionReceipt returns the transaction receipt of the given txHash
func (c *Client) EthTransactionReceipt(ctx context.Context, txHash ethCommon.Hash) (*types.Receipt, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	for i := int64(0); i < c.blockNum; i++ {
		b := c.blocks[i]
		_, ok := b.Rollup.Txs[txHash]
		if !ok {
			_, ok = b.Auction.Txs[txHash]
		}
		if ok {
			return &types.Receipt{
				TxHash:      txHash,
				Status:      types.ReceiptStatusSuccessful,
				BlockHash:   b.Eth.Hash,
				BlockNumber: big.NewInt(b.Eth.BlockNum),
			}, nil
		}
	}

	return nil, nil
}

// CtlAddERC20 adds an ERC20 token to the blockchain.
func (c *Client) CtlAddERC20(tokenAddr ethCommon.Address, constants eth.ERC20Consts) {
	nextBlock := c.nextBlock()
	e := nextBlock.Eth
	e.Tokens[tokenAddr] = constants
}

// EthERC20Consts returns the constants defined for a particular ERC20 Token instance.
func (c *Client) EthERC20Consts(tokenAddr ethCommon.Address) (*eth.ERC20Consts, error) {
	currentBlock := c.currentBlock()
	e := currentBlock.Eth
	if constants, ok := e.Tokens[tokenAddr]; ok {
		return &constants, nil
	}
	return nil, fmt.Errorf("tokenAddr not found")
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

	if blockNum > c.blockNum {
		return nil, ethereum.NotFound
	}
	block := c.blocks[blockNum]
	return &common.Block{
		EthBlockNum: blockNum,
		Timestamp:   time.Unix(block.Eth.Time, 0),
		Hash:        block.Eth.Hash,
		ParentHash:  block.Eth.ParentHash,
	}, nil
}

// EthAddress returns the ethereum address of the account loaded into the Client
func (c *Client) EthAddress() (*ethCommon.Address, error) {
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}
	return c.addr, nil
}

var errTODO = fmt.Errorf("TODO: Not implemented yet")

//
// Rollup
//

// CtlAddL1TxUser adds an L1TxUser to the L1UserTxs queue of the Rollup
// func (c *Client) CtlAddL1TxUser(l1Tx *common.L1Tx) {
// 	c.rw.Lock()
// 	defer c.rw.Unlock()
//
// 	nextBlock := c.nextBlock()
// 	r := nextBlock.Rollup
// 	queue := r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
// 	if len(queue.L1TxQueue) >= eth.RollupConstMaxL1UserTx {
// 		r.State.LastToForgeL1TxsNum++
// 		r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum] = eth.NewQueueStruct()
// 		queue = r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
// 	}
// 	if int64(l1Tx.FromIdx) > r.State.CurrentIdx {
// 		panic("l1Tx.FromIdx > r.State.CurrentIdx")
// 	}
// 	if int(l1Tx.TokenID)+1 > len(r.State.TokenList) {
// 		panic("l1Tx.TokenID + 1 > len(r.State.TokenList)")
// 	}
// 	queue.L1TxQueue = append(queue.L1TxQueue, *l1Tx)
// 	r.Events.L1UserTx = append(r.Events.L1UserTx, eth.RollupEventL1UserTx{
// 		L1Tx:            *l1Tx,
// 		ToForgeL1TxsNum: r.State.LastToForgeL1TxsNum,
// 		Position:        len(queue.L1TxQueue) - 1,
// 	})
// }

// RollupL1UserTxERC20Permit is the interface to call the smart contract function
func (c *Client) RollupL1UserTxERC20Permit(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64, deadline *big.Int) (tx *types.Transaction, err error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupL1UserTxERC20ETH sends an L1UserTx to the Rollup.
func (c *Client) RollupL1UserTxERC20ETH(
	fromBJJ *babyjub.PublicKey,
	fromIdx int64,
	loadAmount *big.Int,
	amount *big.Int,
	tokenID uint32,
	toIdx int64,
) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()

	_, err = common.NewFloat16(amount)
	if err != nil {
		return nil, err
	}
	_, err = common.NewFloat16(loadAmount)
	if err != nil {
		return nil, err
	}

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	queue := r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
	if len(queue.L1TxQueue) >= eth.RollupConstMaxL1UserTx {
		r.State.LastToForgeL1TxsNum++
		r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum] = eth.NewQueueStruct()
		queue = r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
	}
	if fromIdx > r.State.CurrentIdx {
		panic("l1Tx.FromIdx > r.State.CurrentIdx")
	}
	if int(tokenID)+1 > len(r.State.TokenList) {
		panic("l1Tx.TokenID + 1 > len(r.State.TokenList)")
	}
	toForgeL1TxsNum := r.State.LastToForgeL1TxsNum
	l1Tx, err := common.NewL1Tx(&common.L1Tx{
		FromIdx:         common.Idx(fromIdx),
		FromEthAddr:     *c.addr,
		FromBJJ:         fromBJJ,
		Amount:          amount,
		LoadAmount:      loadAmount,
		TokenID:         common.TokenID(tokenID),
		ToIdx:           common.Idx(toIdx),
		ToForgeL1TxsNum: &toForgeL1TxsNum,
		Position:        len(queue.L1TxQueue) - 1,
		UserOrigin:      true,
	})
	if err != nil {
		return nil, err
	}

	queue.L1TxQueue = append(queue.L1TxQueue, *l1Tx)
	r.Events.L1UserTx = append(r.Events.L1UserTx, eth.RollupEventL1UserTx{
		L1UserTx: *l1Tx,
	})
	return r.addTransaction(newTransaction("l1UserTxERC20ETH", l1Tx)), nil
}

// RollupL1UserTxERC777 is the interface to call the smart contract function
// func (c *Client) RollupL1UserTxERC777(fromBJJ *babyjub.PublicKey, fromIdx int64, loadAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64) (*types.Transaction, error) {
// 	log.Error("TODO")
// 	return nil, errTODO
// }

// RollupRegisterTokensCount is the interface to call the smart contract function
func (c *Client) RollupRegisterTokensCount() (*big.Int, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupWithdrawCircuit is the interface to call the smart contract function
func (c *Client) RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int, tokenID uint32, numExitRoot, idx int64, amount *big.Int, instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *Client) RollupWithdrawMerkleProof(babyPubKey *babyjub.PublicKey, tokenID uint32, numExitRoot, idx int64, amount *big.Int, siblings []*big.Int, instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

type transactionData struct {
	Name  string
	Value interface{}
}

func newTransaction(name string, value interface{}) *types.Transaction {
	data, err := json.Marshal(transactionData{name, value})
	if err != nil {
		panic(err)
	}
	return types.NewTransaction(0, ethCommon.Address{}, nil, 0, nil,
		data)
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *Client) RollupForgeBatch(args *eth.RollupForgeBatchArgs) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	a := c.nextBlock().Auction
	ok, err := a.canForge(*c.addr, a.Eth.BlockNum)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("incorrect slot")
	}

	// TODO: Verify proof

	// Auction
	err = a.forge(*c.addr)
	if err != nil {
		return nil, err
	}

	// TODO: If successful, store the tx in a successful array.
	// TODO: If failed, store the tx in a failed array.
	// TODO: Add method to move the tx to another block, reapply it there, and possibly go from successful to failed.

	return c.addBatch(args)
}

// CtlAddBatch adds forged batch to the Rollup, without checking any ZKProof
func (c *Client) CtlAddBatch(args *eth.RollupForgeBatchArgs) {
	c.rw.Lock()
	defer c.rw.Unlock()

	if _, err := c.addBatch(args); err != nil {
		panic(err)
	}
}

func (c *Client) addBatch(args *eth.RollupForgeBatchArgs) (*types.Transaction, error) {
	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	r.State.StateRoot = args.NewStRoot
	if args.NewLastIdx < r.State.CurrentIdx {
		return nil, fmt.Errorf("args.NewLastIdx < r.State.CurrentIdx")
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
	ethTx := r.addTransaction(newTransaction("forgebatch", args))
	c.forgeBatchArgsPending[ethTx.Hash()] = &batch{*args, *c.addr}
	r.Events.ForgeBatch = append(r.Events.ForgeBatch, eth.RollupEventForgeBatch{
		BatchNum:  int64(len(r.State.ExitRoots)),
		EthTxHash: ethTx.Hash(),
	})

	return ethTx, nil
}

// RollupAddTokenSimple is a wrapper around RollupAddToken that automatically
// sets `deadlie`.
func (c *Client) RollupAddTokenSimple(tokenAddress ethCommon.Address, feeAddToken *big.Int) (tx *types.Transaction, err error) {
	return c.RollupAddToken(tokenAddress, feeAddToken, big.NewInt(9999)) //nolint:gomnd
}

// RollupAddToken is the interface to call the smart contract function
func (c *Client) RollupAddToken(tokenAddress ethCommon.Address, feeAddToken *big.Int,
	deadline *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	if _, ok := r.State.TokenMap[tokenAddress]; ok {
		return nil, fmt.Errorf("Token %v already registered", tokenAddress)
	}
	if feeAddToken.Cmp(r.Vars.FeeAddToken) != 0 {
		return nil, fmt.Errorf("Expected fee: %v but got: %v", r.Vars.FeeAddToken, feeAddToken)
	}

	r.State.TokenMap[tokenAddress] = true
	r.State.TokenList = append(r.State.TokenList, tokenAddress)
	r.Events.AddToken = append(r.Events.AddToken, eth.RollupEventAddToken{TokenAddress: tokenAddress,
		TokenID: uint32(len(r.State.TokenList) - 1)})
	return r.addTransaction(newTransaction("addtoken", tokenAddress)), nil
}

// RollupGetCurrentTokens is the interface to call the smart contract function
func (c *Client) RollupGetCurrentTokens() (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateForgeL1L2BatchTimeout is the interface to call the smart contract function
func (c *Client) RollupUpdateForgeL1L2BatchTimeout(newForgeL1Timeout int64) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// RollupUpdateTokensHEZ is the interface to call the smart contract function
// func (c *Client) RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (tx *types.Transaction, err error) {
// 	c.rw.Lock()
// 	defer c.rw.Unlock()
// 	cpy := c.nextBlock().copy()
// 	defer func() { c.revertIfErr(err, cpy) }()
//
// 	log.Error("TODO")
// 	return nil, errTODO
// }

// RollupUpdateGovernance is the interface to call the smart contract function
// func (c *Client) RollupUpdateGovernance() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *Client) RollupConstants() (*eth.RollupPublicConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
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
func (c *Client) RollupForgeBatchArgs(ethTxHash ethCommon.Hash) (*eth.RollupForgeBatchArgs, *ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	batch, ok := c.forgeBatchArgs[ethTxHash]
	if !ok {
		return nil, nil, fmt.Errorf("transaction not found")
	}
	return &batch.ForgeBatchArgs, &batch.Sender, nil
}

//
// Auction
//

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionSetSlotDeadline(newDeadline uint8) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionGetSlotDeadline() (uint8, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, errTODO
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetOpenAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, errTODO
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetClosedAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, errTODO
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionSetOutbidding(newOutbidding uint16) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionGetOutbidding() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionSetAllocationRatio(newAllocationRatio [3]uint16) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionGetAllocationRatio() ([3]uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return [3]uint16{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, errTODO
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
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

// AuctionChangeDefaultSlotSetBid is the interface to call the smart contract function
func (c *Client) AuctionChangeDefaultSlotSetBid(slotSet int64, newInitialMinBid *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionSetCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetCoordinator(forger ethCommon.Address, URL string) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	nextBlock := c.nextBlock()
	a := nextBlock.Auction

	a.State.Coordinators[*c.addr] = &eth.Coordinator{
		Forger: forger,
		URL:    URL,
	}

	a.Events.SetCoordinator = append(a.Events.SetCoordinator,
		eth.AuctionEventSetCoordinator{
			BidderAddress:  *c.addr,
			ForgerAddress:  forger,
			CoordinatorURL: URL,
		})

	type data struct {
		BidderAddress ethCommon.Address
		ForgerAddress ethCommon.Address
		URL           string
	}
	return a.addTransaction(newTransaction("registercoordinator", data{*c.addr, forger, URL})), nil
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *Client) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return false, errTODO
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *Client) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetSlotNumber is the interface to call the smart contract function
func (c *Client) AuctionGetSlotNumber(blockNum int64) (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	currentBlock := c.currentBlock()
	a := currentBlock.Auction
	return a.getSlotNumber(blockNum), nil
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *Client) AuctionGetCurrentSlotNumber() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, errTODO
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *Client) AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetSlotSet is the interface to call the smart contract function
func (c *Client) AuctionGetSlotSet(slot int64) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *Client) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBidSimple is a wrapper around AuctionBid that automatically sets `amount` and `deadline`.
func (c *Client) AuctionBidSimple(slot int64, bidAmount *big.Int) (tx *types.Transaction, err error) {
	return c.AuctionBid(bidAmount, slot, bidAmount, big.NewInt(99999)) //nolint:gomnd
}

// AuctionBid is the interface to call the smart contract function.  This
// implementation behaves as if any address has infinite tokens.
func (c *Client) AuctionBid(amount *big.Int, slot int64, bidAmount *big.Int,
	deadline *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { func() { c.revertIfErr(err, cpy) }() }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

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

	if _, ok := a.State.Coordinators[*c.addr]; !ok {
		return nil, errCoordNotReg
	}

	slotState, ok := a.State.Slots[slot]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slot] = slotState
	}
	slotState.Bidder = *c.addr
	slotState.BidAmount = bidAmount

	a.Events.NewBid = append(a.Events.NewBid,
		eth.AuctionEventNewBid{Slot: slot, BidAmount: bidAmount, Bidder: *c.addr})

	type data struct {
		Slot      int64
		BidAmount *big.Int
		Bidder    ethCommon.Address
	}
	return a.addTransaction(newTransaction("bid", data{slot, bidAmount, *c.addr})), nil
}

// AuctionMultiBid is the interface to call the smart contract function.  This
// implementation behaves as if any address has infinite tokens.
func (c *Client) AuctionMultiBid(amount *big.Int, startingSlot int64, endingSlot int64, slotSet [6]bool,
	maxBid, closedMinBid, deadline *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *Client) AuctionCanForge(forger ethCommon.Address, blockNum int64) (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	currentBlock := c.currentBlock()
	a := currentBlock.Auction
	return a.canForge(forger, blockNum)
}

// AuctionForge is the interface to call the smart contract function
func (c *Client) AuctionForge(forger ethCommon.Address) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *Client) AuctionClaimHEZ() (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, eth.ErrAccountNil
	}

	log.Error("TODO")
	return nil, errTODO
}

// AuctionGetClaimableHEZ is the interface to call the smart contract function
func (c *Client) AuctionGetClaimableHEZ(bidder ethCommon.Address) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
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
