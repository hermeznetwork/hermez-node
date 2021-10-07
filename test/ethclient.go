package test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/mitchellh/copystructure"
	"github.com/multiformats/go-multiaddr"
)

func init() {
	log.Init("debug", []string{"stdout"})
	copystructure.Copiers[reflect.TypeOf(big.Int{})] =
		func(raw interface{}) (interface{}, error) {
			in := raw.(big.Int)
			out := new(big.Int).Set(&in)
			return *out, nil
		}
}

// WDelayerBlock stores all the data related to the WDelayer SC from an ethereum block
type WDelayerBlock struct {
	Vars      common.WDelayerVariables
	Events    eth.WDelayerEvents
	Txs       map[ethCommon.Hash]*types.Transaction
	Constants *common.WDelayerConstants
	Eth       *EthereumBlock
}

func (w *WDelayerBlock) addTransaction(tx *types.Transaction) *types.Transaction {
	txHash := tx.Hash()
	w.Txs[txHash] = tx
	return tx
}

func (w *WDelayerBlock) deposit(txHash ethCommon.Hash, owner, token ethCommon.Address,
	amount *big.Int) {
	w.Events.Deposit = append(w.Events.Deposit, eth.WDelayerEventDeposit{
		Owner:            owner,
		Token:            token,
		Amount:           amount,
		DepositTimestamp: uint64(w.Eth.Time),
		TxHash:           txHash,
	})
}

// RollupBlock stores all the data related to the Rollup SC from an ethereum block
type RollupBlock struct {
	State     eth.RollupState
	Vars      common.RollupVariables
	Events    eth.RollupEvents
	Txs       map[ethCommon.Hash]*types.Transaction
	Constants *common.RollupConstants
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
	Vars      common.AuctionVariables
	Events    eth.AuctionEvents
	Txs       map[ethCommon.Hash]*types.Transaction
	Constants *common.AuctionConstants
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
		return nil, tracerr.Wrap(errBidClosed)
	}

	slotSet := a.getSlotSet(slot)
	// fmt.Println("slot:", slot, "slotSet:", slotSet)
	var prevBid *big.Int
	slotState, ok := a.State.Slots[slot]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slot] = slotState
	}
	// If the bidAmount for a slot is 0 it means that it has not yet been
	// bid, so the midBid will be the minimum bid for the slot time plus
	// the outbidding set, otherwise it will be the bidAmount plus the
	// outbidding
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
		return tracerr.Wrap(err)
	} else if !ok {
		return tracerr.Wrap(fmt.Errorf("Can't forge"))
	}

	slotToForge := a.getSlotNumber(a.Eth.BlockNum)
	slotState, ok := a.State.Slots[slotToForge]
	if !ok {
		slotState = eth.NewSlotState()
		a.State.Slots[slotToForge] = slotState
	}

	if !slotState.ForgerCommitment {
		// Get the relativeBlock to check if the slotDeadline has been exceeded
		relativeBlock := a.Eth.BlockNum - (a.Constants.GenesisBlockNum +
			(slotToForge * int64(a.Constants.BlocksPerSlot)))
		if relativeBlock < int64(a.Vars.SlotDeadline) {
			slotState.ForgerCommitment = true
		}
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
		return false, tracerr.Wrap(fmt.Errorf("Auction has not started yet"))
	}

	slotToForge := a.getSlotNumber(blockNum)
	// Get the relativeBlock to check if the slotDeadline has been exceeded
	relativeBlock := blockNum - (a.Constants.GenesisBlockNum + (slotToForge *
		int64(a.Constants.BlocksPerSlot)))

	// If the closedMinBid is 0 it means that we have to take as minBid the
	// one that is set for this slot set, otherwise the one that has been
	// saved will be used
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

	if !slotState.ForgerCommitment && (relativeBlock >= int64(a.Vars.SlotDeadline)) {
		// if the relative block has exceeded the slotDeadline and no
		// batch has been forged, anyone can forge
		return true, nil
	} else if coord, ok := a.State.Coordinators[slotState.Bidder]; ok &&
		coord.Forger == forger && slotState.BidAmount.Cmp(minBid) >= 0 {
		// if forger bidAmount has exceeded the minBid it can forge
		return true, nil
	} else if a.Vars.BootCoordinator == forger && slotState.BidAmount.Cmp(minBid) == -1 {
		// if it's the boot coordinator and it has not been bid or the
		// bid is below the minimum it can forge
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
	Nonce      uint64
	// state      ethState
}

// Block represents a ethereum block
type Block struct {
	Rollup   *RollupBlock
	Auction  *AuctionBlock
	WDelayer *WDelayerBlock
	Eth      *EthereumBlock
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
	blockNext.WDelayer.Constants = b.WDelayer.Constants
	blockNext.Rollup.Eth = blockNext.Eth
	blockNext.Auction.Eth = blockNext.Eth
	blockNext.WDelayer.Eth = blockNext.Eth

	return blockNext
}

// ClientSetup is used to initialize the constants of the Smart Contracts and
// other details of the test Client
type ClientSetup struct {
	RollupConstants   *common.RollupConstants
	RollupVariables   *common.RollupVariables
	AuctionConstants  *common.AuctionConstants
	AuctionVariables  *common.AuctionVariables
	WDelayerConstants *common.WDelayerConstants
	WDelayerVariables *common.WDelayerVariables
	VerifyProof       bool
	ChainID           *big.Int
}

// NewClientSetupExample returns a ClientSetup example with hardcoded realistic
// values.  With this setup, the rollup genesis will be block 1, and block 0
// and 1 will be premined.
//nolint:gomnd
func NewClientSetupExample() *ClientSetup {
	initialMinimalBidding, ok := new(big.Int).SetString("10000000000000000000", 10) // 10 * (1e18)
	if !ok {
		panic("bad initialMinimalBidding")
	}
	tokenHEZ := ethCommon.HexToAddress("0x51D243D62852Bba334DD5cc33f242BAc8c698074")
	governanceAddress := ethCommon.HexToAddress("0x688EfD95BA4391f93717CF02A9aED9DBD2855cDd")
	rollupConstants := &common.RollupConstants{
		Verifiers: []common.RollupVerifierStruct{
			{
				MaxTx:   2048,
				NLevels: 32,
			},
		},
		TokenHEZ:                tokenHEZ,
		HermezGovernanceAddress: governanceAddress,
		HermezAuctionContract:   ethCommon.HexToAddress("0x8E442975805fb1908f43050c9C1A522cB0e28D7b"),
		WithdrawDelayerContract: ethCommon.HexToAddress("0x5CB7979cBdbf65719BEE92e4D15b7b7Ed3D79114"),
	}
	rollupVariables := &common.RollupVariables{
		FeeAddToken:           big.NewInt(11),
		ForgeL1L2BatchTimeout: 10,
		WithdrawalDelay:       80,
		Buckets:               []common.BucketParams{},
	}
	auctionConstants := &common.AuctionConstants{
		BlocksPerSlot:         40,
		InitialMinimalBidding: initialMinimalBidding,
		GenesisBlockNum:       1,
		GovernanceAddress:     governanceAddress,
		TokenHEZ:              tokenHEZ,
		HermezRollup:          ethCommon.HexToAddress("0x474B6e29852257491cf283EfB1A9C61eBFe48369"),
	}
	auctionVariables := &common.AuctionVariables{
		DonationAddress:    ethCommon.HexToAddress("0x61Ed87CF0A1496b49A420DA6D84B58196b98f2e7"),
		BootCoordinator:    ethCommon.HexToAddress("0xE39fEc6224708f0772D2A74fd3f9055A90E0A9f2"),
		BootCoordinatorURL: "https://boot.coordinator.com",
		DefaultSlotSetBid: [6]*big.Int{
			initialMinimalBidding, initialMinimalBidding, initialMinimalBidding,
			initialMinimalBidding, initialMinimalBidding, initialMinimalBidding,
		},
		ClosedAuctionSlots: 2,
		OpenAuctionSlots:   4320,
		AllocationRatio:    [3]uint16{4000, 4000, 2000},
		Outbidding:         1000,
		SlotDeadline:       20,
	}
	wDelayerConstants := &common.WDelayerConstants{
		MaxWithdrawalDelay:   60 * 60 * 24 * 7 * 2,  // 2 weeks
		MaxEmergencyModeTime: 60 * 60 * 24 * 7 * 26, // 26 weeks
		HermezRollup:         auctionConstants.HermezRollup,
	}
	wDelayerVariables := &common.WDelayerVariables{
		HermezGovernanceAddress:    ethCommon.HexToAddress("0xcfD0d163AE6432a72682323E2C3A5a69e6B37D12"),
		EmergencyCouncilAddress:    ethCommon.HexToAddress("0x2730700932a4FDB97B9268A3Ca29f97Ea5fd7EA0"),
		WithdrawalDelay:            60,
		EmergencyModeStartingBlock: 0,
		EmergencyMode:              false,
	}
	return &ClientSetup{
		RollupConstants:   rollupConstants,
		RollupVariables:   rollupVariables,
		AuctionConstants:  auctionConstants,
		AuctionVariables:  auctionVariables,
		WDelayerConstants: wDelayerConstants,
		WDelayerVariables: wDelayerVariables,
		VerifyProof:       false,
		ChainID:           big.NewInt(0),
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
	rw                *sync.RWMutex
	log               bool
	addr              *ethCommon.Address
	chainID           *big.Int
	rollupConstants   *common.RollupConstants
	auctionConstants  *common.AuctionConstants
	wDelayerConstants *common.WDelayerConstants
	blocks            map[int64]*Block
	// state            state
	blockNum    int64 // last mined block num
	maxBlockNum int64 // highest block num calculated
	timer       Timer
	hasher      hasher

	forgeBatchArgsPending map[ethCommon.Hash]*batch
	forgeBatchArgs        map[ethCommon.Hash]*batch

	startBlock int64
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
				ExitRoots:        make([]*big.Int, 1),
				ExitNullifierMap: make(map[int64]map[int64]bool),
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
		WDelayer: &WDelayerBlock{
			// State: TODO
			Vars:      *setup.WDelayerVariables,
			Txs:       make(map[ethCommon.Hash]*types.Transaction),
			Events:    eth.NewWDelayerEvents(),
			Constants: setup.WDelayerConstants,
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
		wDelayerConstants:     setup.WDelayerConstants,
		blocks:                blocks,
		timer:                 timer,
		hasher:                hasher,
		forgeBatchArgsPending: make(map[ethCommon.Hash]*batch),
		forgeBatchArgs:        make(map[ethCommon.Hash]*batch),
		blockNum:              blockNum,
		maxBlockNum:           blockNum,
	}

	if c.startBlock == 0 {
		c.startBlock = 2
	}
	for i := int64(1); i < c.startBlock; i++ {
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

// CtlLastBlock returns the last blockNum without checks
func (c *Client) CtlLastBlock() *common.Block {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block := c.blocks[c.blockNum]
	return &common.Block{
		Num:        c.blockNum,
		Timestamp:  time.Unix(block.Eth.Time, 0),
		Hash:       block.Eth.Hash,
		ParentHash: block.Eth.ParentHash,
	}
}

// CtlLastForgedBatch returns the last batchNum without checks
func (c *Client) CtlLastForgedBatch() int64 {
	c.rw.RLock()
	defer c.rw.RUnlock()

	currentBlock := c.currentBlock()
	e := currentBlock.Rollup
	return int64(len(e.State.ExitRoots)) - 1
}

// EthChainID returns the ChainID of the ethereum network
func (c *Client) EthChainID() (*big.Int, error) {
	return c.chainID, nil
}

// EthPendingNonceAt returns the account nonce of the given account in the pending
// state. This is the nonce that should be used for the next transaction.
func (c *Client) EthPendingNonceAt(ctx context.Context, account ethCommon.Address) (uint64, error) {
	// NOTE: For now Client doesn't simulate nonces
	return 0, nil
}

// EthNonceAt returns the account nonce of the given account. The block number can
// be nil, in which case the nonce is taken from the latest known block.
func (c *Client) EthNonceAt(ctx context.Context, account ethCommon.Address,
	blockNumber *big.Int) (uint64, error) {
	// NOTE: For now Client doesn't simulate nonces
	return 0, nil
}

// EthSuggestGasPrice retrieves the currently suggested gas price to allow a
// timely execution of a transaction.
func (c *Client) EthSuggestGasPrice(ctx context.Context) (*big.Int, error) {
	// NOTE: For now Client doesn't simulate gasPrice
	return big.NewInt(0), nil
}

// EthKeyStore returns the keystore in the Client
func (c *Client) EthKeyStore() *ethKeystore.KeyStore {
	return nil
}

// EthCall runs the transaction as a call (without paying) in the local node at
// blockNum.
func (c *Client) EthCall(ctx context.Context, tx *types.Transaction,
	blockNum *big.Int) ([]byte, error) {
	return nil, tracerr.Wrap(common.ErrTODO)
}

// EthLastBlock returns the last blockNum
func (c *Client) EthLastBlock() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if c.blockNum < c.maxBlockNum {
		panic("blockNum has decreased.  " +
			"After a rollback you must mine to reach the same or higher blockNum")
	}
	return c.blockNum, nil
}

// EthTransactionReceipt returns the transaction receipt of the given txHash
func (c *Client) EthTransactionReceipt(ctx context.Context,
	txHash ethCommon.Hash) (*types.Receipt, error) {
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
	return nil, tracerr.Wrap(fmt.Errorf("tokenAddr not found"))
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
// deterministic way.  If number == -1, the latests known block is returned.
func (c *Client) EthBlockByNumber(ctx context.Context, blockNum int64) (*common.Block, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	if blockNum > c.blockNum {
		return nil, ethereum.NotFound
	}
	if blockNum == -1 {
		blockNum = c.blockNum
	}
	block := c.blocks[blockNum]
	return &common.Block{
		Num:        blockNum,
		Timestamp:  time.Unix(block.Eth.Time, 0),
		Hash:       block.Eth.Hash,
		ParentHash: block.Eth.ParentHash,
	}, nil
}

// EthAddress returns the ethereum address of the account loaded into the Client
func (c *Client) EthAddress() (*ethCommon.Address, error) {
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
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
func (c *Client) RollupL1UserTxERC20Permit(fromBJJ babyjub.PublicKeyComp, fromIdx int64,
	depositAmount *big.Int, amount *big.Int, tokenID uint32, toIdx int64,
	deadline *big.Int) (tx *types.Transaction, err error) {
	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupL1UserTxERC20ETH sends an L1UserTx to the Rollup.
func (c *Client) RollupL1UserTxERC20ETH(
	fromBJJ babyjub.PublicKeyComp,
	fromIdx int64,
	depositAmount *big.Int,
	amount *big.Int,
	tokenID uint32,
	toIdx int64,
) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()

	_, err = common.NewFloat40(amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	_, err = common.NewFloat40(depositAmount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	queue := r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum]
	if len(queue.L1TxQueue) >= common.RollupConstMaxL1UserTx {
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
		DepositAmount:   depositAmount,
		TokenID:         common.TokenID(tokenID),
		ToIdx:           common.Idx(toIdx),
		ToForgeL1TxsNum: &toForgeL1TxsNum,
		Position:        len(queue.L1TxQueue),
		UserOrigin:      true,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	queue.L1TxQueue = append(queue.L1TxQueue, *l1Tx)
	r.Events.L1UserTx = append(r.Events.L1UserTx, eth.RollupEventL1UserTx{
		L1UserTx: *l1Tx,
	})
	return r.addTransaction(c.newTransaction("l1UserTxERC20ETH", l1Tx)), nil
}

// RollupL1UserTxERC777 is the interface to call the smart contract function
// func (c *Client) RollupL1UserTxERC777(fromBJJ *babyjub.PublicKey, fromIdx int64,
// 	depositAmount *big.Int, amount *big.Int, tokenID uint32,
//	toIdx int64) (*types.Transaction, error) {
// 	log.Error("TODO")
// 	return nil, errTODO
// }

// RollupRegisterTokensCount is the interface to call the smart contract function
func (c *Client) RollupRegisterTokensCount() (*big.Int, error) {
	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupLastForgedBatch is the interface to call the smart contract function
func (c *Client) RollupLastForgedBatch() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	currentBlock := c.currentBlock()
	e := currentBlock.Rollup
	return int64(len(e.State.ExitRoots)) - 1, nil
}

// RollupWithdrawCircuit is the interface to call the smart contract function
func (c *Client) RollupWithdrawCircuit(proofA, proofC [2]*big.Int, proofB [2][2]*big.Int,
	tokenID uint32, numExitRoot, idx int64, amount *big.Int,
	instantWithdraw bool) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *Client) RollupWithdrawMerkleProof(babyPubKey babyjub.PublicKeyComp,
	tokenID uint32, numExitRoot, idx int64, amount *big.Int, siblings []*big.Int,
	instantWithdraw bool) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup

	if int(numExitRoot) >= len(r.State.ExitRoots) {
		return nil, tracerr.Wrap(fmt.Errorf("numExitRoot >= len(r.State.ExitRoots)"))
	}
	if _, ok := r.State.ExitNullifierMap[numExitRoot][idx]; ok {
		return nil, tracerr.Wrap(fmt.Errorf("exit already withdrawn"))
	}
	r.State.ExitNullifierMap[numExitRoot][idx] = true

	babyPubKeyDecomp, err := babyPubKey.Decompress()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	type data struct {
		BabyPubKey      *babyjub.PublicKey
		TokenID         uint32
		NumExitRoot     int64
		Idx             int64
		Amount          *big.Int
		Siblings        []*big.Int
		InstantWithdraw bool
	}
	tx = r.addTransaction(c.newTransaction("withdrawMerkleProof", data{
		BabyPubKey:      babyPubKeyDecomp,
		TokenID:         tokenID,
		NumExitRoot:     numExitRoot,
		Idx:             idx,
		Amount:          amount,
		Siblings:        siblings,
		InstantWithdraw: instantWithdraw,
	}))
	r.Events.Withdraw = append(r.Events.Withdraw, eth.RollupEventWithdraw{
		Idx:             uint64(idx),
		NumExitRoot:     uint64(numExitRoot),
		InstantWithdraw: instantWithdraw,
		TxHash:          tx.Hash(),
	})

	if !instantWithdraw {
		w := nextBlock.WDelayer
		w.deposit(tx.Hash(), *c.addr, r.State.TokenList[int(tokenID)], amount)
	}
	return tx, nil
}

type transactionData struct {
	Name  string
	Value interface{}
}

func (c *Client) newTransaction(name string, value interface{}) *types.Transaction {
	eth := c.nextBlock().Eth
	nonce := eth.Nonce
	eth.Nonce++
	data, err := json.Marshal(transactionData{name, value})
	if err != nil {
		panic(err)
	}
	return types.NewTransaction(nonce, ethCommon.Address{}, nil, 0, nil,
		data)
}

// RollupForgeBatch is the interface to call the smart contract function
func (c *Client) RollupForgeBatch(args *eth.RollupForgeBatchArgs,
	auth *bind.TransactOpts) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	a := c.nextBlock().Auction
	ok, err := a.canForge(*c.addr, a.Eth.BlockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if !ok {
		return nil, tracerr.Wrap(fmt.Errorf(common.AuctionErrMsgCannotForge))
	}

	// Auction
	err = a.forge(*c.addr)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

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
		return nil, tracerr.Wrap(fmt.Errorf("args.NewLastIdx < r.State.CurrentIdx"))
	}
	r.State.CurrentIdx = args.NewLastIdx
	r.State.ExitNullifierMap[int64(len(r.State.ExitRoots))] = make(map[int64]bool)
	r.State.ExitRoots = append(r.State.ExitRoots, args.NewExitRoot)
	if args.L1Batch {
		r.State.CurrentToForgeL1TxsNum++
		if r.State.CurrentToForgeL1TxsNum == r.State.LastToForgeL1TxsNum {
			r.State.LastToForgeL1TxsNum++
			r.State.MapL1TxQueue[r.State.LastToForgeL1TxsNum] = eth.NewQueueStruct()
		}
	}
	ethTx := r.addTransaction(c.newTransaction("forgebatch", args))
	c.forgeBatchArgsPending[ethTx.Hash()] = &batch{*args, *c.addr}
	r.Events.ForgeBatch = append(r.Events.ForgeBatch, eth.RollupEventForgeBatch{
		BatchNum:     int64(len(r.State.ExitRoots)) - 1,
		EthTxHash:    ethTx.Hash(),
		L1UserTxsLen: uint16(len(args.L1UserTxs)),
	})

	return ethTx, nil
}

// RollupAddTokenSimple is a wrapper around RollupAddToken that automatically
// sets `deadlie`.
func (c *Client) RollupAddTokenSimple(tokenAddress ethCommon.Address,
	feeAddToken *big.Int) (tx *types.Transaction, err error) {
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
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	if _, ok := r.State.TokenMap[tokenAddress]; ok {
		return nil, tracerr.Wrap(fmt.Errorf("Token %v already registered", tokenAddress))
	}
	if feeAddToken.Cmp(r.Vars.FeeAddToken) != 0 {
		return nil,
			tracerr.Wrap(fmt.Errorf("Expected fee: %v but got: %v",
				r.Vars.FeeAddToken, feeAddToken))
	}

	r.State.TokenMap[tokenAddress] = true
	r.State.TokenList = append(r.State.TokenList, tokenAddress)
	r.Events.AddToken = append(r.Events.AddToken, eth.RollupEventAddToken{
		TokenAddress: tokenAddress,
		TokenID:      uint32(len(r.State.TokenList) - 1)})
	return r.addTransaction(c.newTransaction("addtoken", tokenAddress)), nil
}

// RollupGetCurrentTokens is the interface to call the smart contract function
func (c *Client) RollupGetCurrentTokens() (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupUpdateForgeL1L2BatchTimeout is the interface to call the smart contract function
func (c *Client) RollupUpdateForgeL1L2BatchTimeout(newForgeL1Timeout int64) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	nextBlock := c.nextBlock()
	r := nextBlock.Rollup
	r.Vars.ForgeL1L2BatchTimeout = newForgeL1Timeout
	r.Events.UpdateForgeL1L2BatchTimeout = append(r.Events.UpdateForgeL1L2BatchTimeout,
		eth.RollupEventUpdateForgeL1L2BatchTimeout{NewForgeL1L2BatchTimeout: newForgeL1Timeout})

	return r.addTransaction(c.newTransaction("updateForgeL1L2BatchTimeout", newForgeL1Timeout)), nil
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *Client) RollupConstants() (*common.RollupConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.rollupConstants, nil
}

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *Client) RollupEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*eth.RollupEvents, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, tracerr.Wrap(fmt.Errorf("Block %v doesn't exist", blockNum))
	}
	if blockHash != nil && *blockHash != block.Eth.Hash {
		return nil, tracerr.Wrap(fmt.Errorf("Hash mismatch, requested %v got %v",
			blockHash, block.Eth.Hash))
	}
	return &block.Rollup.Events, nil
}

// RollupEventInit returns the initialize event with its corresponding block number
func (c *Client) RollupEventInit(genesisBlockNum int64) (*eth.RollupEventInitialize, int64, error) {
	vars := c.blocks[0].Rollup.Vars
	return &eth.RollupEventInitialize{
		ForgeL1L2BatchTimeout: uint8(vars.ForgeL1L2BatchTimeout),
		FeeAddToken:           vars.FeeAddToken,
		WithdrawalDelay:       vars.WithdrawalDelay,
	}, 1, nil
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract
// in the given transaction
func (c *Client) RollupForgeBatchArgs(ethTxHash ethCommon.Hash,
	l1UserTxsLen uint16) (*eth.RollupForgeBatchArgs, *ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	batch, ok := c.forgeBatchArgs[ethTxHash]
	if !ok {
		return nil, nil, tracerr.Wrap(fmt.Errorf("transaction not found"))
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
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionGetSlotDeadline() (uint8, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	nextBlock := c.nextBlock()
	a := nextBlock.Auction
	a.Vars.OpenAuctionSlots = newOpenAuctionSlots
	a.Events.NewOpenAuctionSlots = append(a.Events.NewOpenAuctionSlots,
		eth.AuctionEventNewOpenAuctionSlots{NewOpenAuctionSlots: newOpenAuctionSlots})

	return a.addTransaction(c.newTransaction("setOpenAuctionSlots", newOpenAuctionSlots)), nil
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetOpenAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetClosedAuctionSlots() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionSetOutbidding(newOutbidding uint16) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionGetOutbidding() (uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionSetAllocationRatio(newAllocationRatio [3]uint16) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionGetAllocationRatio() ([3]uint16, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return [3]uint16{}, tracerr.Wrap(errTODO)
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionSetDonationAddress(
	newDonationAddress ethCommon.Address) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address,
	newBootCoordinatorURL string) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
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
func (c *Client) AuctionChangeDefaultSlotSetBid(slotSet int64,
	newInitialMinBid *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionSetCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetCoordinator(forger ethCommon.Address,
	URL string) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
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
	return a.addTransaction(c.newTransaction("registercoordinator", data{*c.addr, forger, URL})),
		nil
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *Client) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return false, tracerr.Wrap(errTODO)
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *Client) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address,
	newWithdrawAddress ethCommon.Address, newURL string) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
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
	return 0, tracerr.Wrap(errTODO)
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *Client) AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetSlotSet is the interface to call the smart contract function
func (c *Client) AuctionGetSlotSet(slot int64) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *Client) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int,
// 	userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBidSimple is a wrapper around AuctionBid that automatically sets `amount` and `deadline`.
func (c *Client) AuctionBidSimple(slot int64, bidAmount *big.Int) (tx *types.Transaction,
	err error) {
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
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	nextBlock := c.nextBlock()
	a := nextBlock.Auction

	if slot <= a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots) {
		return nil, tracerr.Wrap(errBidClosed)
	}

	if slot >
		a.getCurrentSlotNumber()+int64(a.Vars.ClosedAuctionSlots)+int64(a.Vars.OpenAuctionSlots) {
		return nil, tracerr.Wrap(errBidNotOpen)
	}

	minBid, err := a.getMinBidBySlot(slot)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if bidAmount.Cmp(minBid) == -1 {
		return nil, tracerr.Wrap(errBidBelowMin)
	}

	if _, ok := a.State.Coordinators[*c.addr]; !ok {
		return nil, tracerr.Wrap(errCoordNotReg)
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
	return a.addTransaction(c.newTransaction("bid", data{slot, bidAmount, *c.addr})), nil
}

// AuctionMultiBid is the interface to call the smart contract function.  This
// implementation behaves as if any address has infinite tokens.
func (c *Client) AuctionMultiBid(amount *big.Int, startingSlot int64, endingSlot int64,
	slotSet [6]bool, maxBid, closedMinBid, deadline *big.Int) (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
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
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *Client) AuctionClaimHEZ() (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionGetClaimableHEZ is the interface to call the smart contract function
func (c *Client) AuctionGetClaimableHEZ(bidder ethCommon.Address) (*big.Int, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *Client) AuctionConstants() (*common.AuctionConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.auctionConstants, nil
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *Client) AuctionEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*eth.AuctionEvents, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, tracerr.Wrap(fmt.Errorf("Block %v doesn't exist", blockNum))
	}
	if blockHash != nil && *blockHash != block.Eth.Hash {
		return nil, tracerr.Wrap(fmt.Errorf("Hash mismatch, requested %v got %v",
			blockHash, block.Eth.Hash))
	}
	return &block.Auction.Events, nil
}

// AuctionEventInit returns the initialize event with its corresponding block number
func (c *Client) AuctionEventInit(genesisBlockNum int64) (*eth.AuctionEventInitialize, int64, error) {
	vars := c.blocks[0].Auction.Vars
	return &eth.AuctionEventInitialize{
		DonationAddress:        vars.DonationAddress,
		BootCoordinatorAddress: vars.BootCoordinator,
		BootCoordinatorURL:     vars.BootCoordinatorURL,
		Outbidding:             vars.Outbidding,
		SlotDeadline:           vars.SlotDeadline,
		ClosedAuctionSlots:     vars.ClosedAuctionSlots,
		OpenAuctionSlots:       vars.OpenAuctionSlots,
		AllocationRatio:        vars.AllocationRatio,
	}, 1, nil
}

// GetCoordinatorsLibP2PAddrs required for the interface
func (c *Client) GetCoordinatorsLibP2PAddrs() ([]multiaddr.Multiaddr, error) {
	return nil, errors.New("TODO")
}

//
// WDelayer
//

// WDelayerGetHermezGovernanceAddress is the interface to call the smart contract function
func (c *Client) WDelayerGetHermezGovernanceAddress() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerTransferGovernance is the interface to call the smart contract function
func (c *Client) WDelayerTransferGovernance(newAddress ethCommon.Address) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerClaimGovernance is the interface to call the smart contract function
func (c *Client) WDelayerClaimGovernance() (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerGetEmergencyCouncil is the interface to call the smart contract function
func (c *Client) WDelayerGetEmergencyCouncil() (*ethCommon.Address, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerTransferEmergencyCouncil is the interface to call the smart contract function
func (c *Client) WDelayerTransferEmergencyCouncil(newAddress ethCommon.Address) (
	tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerClaimEmergencyCouncil is the interface to call the smart contract function
func (c *Client) WDelayerClaimEmergencyCouncil() (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerIsEmergencyMode is the interface to call the smart contract function
func (c *Client) WDelayerIsEmergencyMode() (bool, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return false, tracerr.Wrap(errTODO)
}

// WDelayerGetWithdrawalDelay is the interface to call the smart contract function
func (c *Client) WDelayerGetWithdrawalDelay() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// WDelayerGetEmergencyModeStartingTime is the interface to call the smart contract function
func (c *Client) WDelayerGetEmergencyModeStartingTime() (int64, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return 0, tracerr.Wrap(errTODO)
}

// WDelayerEnableEmergencyMode is the interface to call the smart contract function
func (c *Client) WDelayerEnableEmergencyMode() (tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerChangeWithdrawalDelay is the interface to call the smart contract function
func (c *Client) WDelayerChangeWithdrawalDelay(newWithdrawalDelay uint64) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	nextBlock := c.nextBlock()
	w := nextBlock.WDelayer
	w.Vars.WithdrawalDelay = newWithdrawalDelay
	w.Events.NewWithdrawalDelay = append(w.Events.NewWithdrawalDelay,
		eth.WDelayerEventNewWithdrawalDelay{WithdrawalDelay: newWithdrawalDelay})

	return w.addTransaction(c.newTransaction("changeWithdrawalDelay", newWithdrawalDelay)), nil
}

// WDelayerDepositInfo is the interface to call the smart contract function
func (c *Client) WDelayerDepositInfo(owner, token ethCommon.Address) (eth.DepositState, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	log.Error("TODO")
	return eth.DepositState{}, tracerr.Wrap(errTODO)
}

// WDelayerDeposit is the interface to call the smart contract function
func (c *Client) WDelayerDeposit(onwer, token ethCommon.Address, amount *big.Int) (
	tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerWithdrawal is the interface to call the smart contract function
func (c *Client) WDelayerWithdrawal(owner, token ethCommon.Address) (tx *types.Transaction,
	err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerEscapeHatchWithdrawal is the interface to call the smart contract function
func (c *Client) WDelayerEscapeHatchWithdrawal(to, token ethCommon.Address, amount *big.Int) (
	tx *types.Transaction, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	cpy := c.nextBlock().copy()
	defer func() { c.revertIfErr(err, cpy) }()
	if c.addr == nil {
		return nil, tracerr.Wrap(eth.ErrAccountNil)
	}

	log.Error("TODO")
	return nil, tracerr.Wrap(errTODO)
}

// WDelayerEventsByBlock returns the events in a block that happened in the WDelayer Contract
func (c *Client) WDelayerEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*eth.WDelayerEvents, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	block, ok := c.blocks[blockNum]
	if !ok {
		return nil, tracerr.Wrap(fmt.Errorf("Block %v doesn't exist", blockNum))
	}
	if blockHash != nil && *blockHash != block.Eth.Hash {
		return nil, tracerr.Wrap(fmt.Errorf("Hash mismatch, requested %v got %v",
			blockHash, block.Eth.Hash))
	}
	return &block.WDelayer.Events, nil
}

// WDelayerConstants returns the Constants of the WDelayer Contract
func (c *Client) WDelayerConstants() (*common.WDelayerConstants, error) {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.wDelayerConstants, nil
}

// WDelayerEventInit returns the initialize event with its corresponding block number
func (c *Client) WDelayerEventInit(genesisBlockNum int64) (*eth.WDelayerEventInitialize, int64, error) {
	vars := c.blocks[0].WDelayer.Vars
	return &eth.WDelayerEventInitialize{
		InitialWithdrawalDelay:         vars.WithdrawalDelay,
		InitialHermezGovernanceAddress: vars.HermezGovernanceAddress,
		InitialEmergencyCouncil:        vars.EmergencyCouncilAddress,
	}, 1, nil
}

// CtlAddBlocks adds block data to the smarts contracts.  The added blocks will
// appear as mined.  Not thread safe.
func (c *Client) CtlAddBlocks(blocks []common.BlockData) (err error) {
	// NOTE: We don't lock because internally we call public functions that
	// lock already.
	for _, block := range blocks {
		nextBlock := c.nextBlock()
		rollup := nextBlock.Rollup
		auction := nextBlock.Auction
		for _, token := range block.Rollup.AddedTokens {
			if _, err := c.RollupAddTokenSimple(token.EthAddr,
				rollup.Vars.FeeAddToken); err != nil {
				return tracerr.Wrap(err)
			}
		}
		for _, tx := range block.Rollup.L1UserTxs {
			c.CtlSetAddr(tx.FromEthAddr)
			if _, err := c.RollupL1UserTxERC20ETH(tx.FromBJJ, int64(tx.FromIdx),
				tx.DepositAmount, tx.Amount, uint32(tx.TokenID),
				int64(tx.ToIdx)); err != nil {
				return tracerr.Wrap(err)
			}
		}
		c.CtlSetAddr(auction.Vars.BootCoordinator)
		for _, batch := range block.Rollup.Batches {
			auths := make([][]byte, len(batch.L1CoordinatorTxs))
			for i := range auths {
				auths[i] = make([]byte, 65)
			}
			if _, err := c.RollupForgeBatch(&eth.RollupForgeBatchArgs{
				NewLastIdx:            batch.Batch.LastIdx,
				NewStRoot:             batch.Batch.StateRoot,
				NewExitRoot:           batch.Batch.ExitRoot,
				L1CoordinatorTxs:      batch.L1CoordinatorTxs,
				L1CoordinatorTxsAuths: auths,
				L2TxsData:             batch.L2Txs,
				FeeIdxCoordinator:     batch.Batch.FeeIdxsCoordinator,
				// Circuit selector
				VerifierIdx: 0, // Intentionally empty
				L1Batch:     batch.L1Batch,
				ProofA:      [2]*big.Int{},    // Intentionally empty
				ProofB:      [2][2]*big.Int{}, // Intentionally empty
				ProofC:      [2]*big.Int{},    // Intentionally empty
			}, nil); err != nil {
				return tracerr.Wrap(err)
			}
		}
		// Mine block and sync
		c.CtlMineBlock()
	}
	return nil
}
