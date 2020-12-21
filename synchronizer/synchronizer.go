package synchronizer

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

var (
// ErrNotAbleToSync is used when there is not possible to find a valid block to sync
// ErrNotAbleToSync = errors.New("it has not been possible to synchronize any block")
)

// // SyncronizerState describes the synchronization progress of the smart contracts
// type SyncronizerState struct {
// 	LastUpdate                time.Time // last time this information was updated
// 	CurrentBatchNum           BatchNum  // Last batch that was forged on the blockchain
// 	CurrentBlockNum           uint64    // Last block that was mined on Ethereum
// 	CurrentToForgeL1TxsNum    uint32
// 	LastSyncedBatchNum        BatchNum // last batch synchronized by the coordinator
// 	LastSyncedBlockNum        uint64   // last Ethereum block synchronized by the coordinator
// 	LastSyncedToForgeL1TxsNum uint32
// }

// // SyncStatus is returned by the Status method of the Synchronizer
// type SyncStatus struct {
// 	CurrentBlock      int64
// 	CurrentBatch      BatchNum
// 	CurrentForgerAddr ethCommon.Address
// 	NextForgerAddr    ethCommon.Address
// 	Synchronized      bool
// }

// Stats of the syncrhonizer
type Stats struct {
	Eth struct {
		RefreshPeriod time.Duration
		Updated       time.Time
		FirstBlockNum int64
		LastBlock     common.Block
		LastBatch     int64
	}
	Sync struct {
		Updated   time.Time
		LastBlock common.Block
		LastBatch int64
		// LastL1BatchBlock is the last ethereum block in which an
		// l1Batch was forged
		LastL1BatchBlock  int64
		LastForgeL1TxsNum int64
		Auction           struct {
			CurrentSlot common.Slot
		}
	}
}

// Synced returns true if the Synchronizer is up to date with the last ethereum block
func (s *Stats) Synced() bool {
	return s.Eth.LastBlock.Num == s.Sync.LastBlock.Num
}

// TODO(Edu): Consider removing all the mutexes from StatsHolder, make
// Synchronizer.Stats not thread-safe, don't pass the synchronizer to the
// debugAPI, and have a copy of the Stats in the DebugAPI that the node passes
// when the Sync updates.

// StatsHolder stores stats and that allows reading and writing them
// concurrently
type StatsHolder struct {
	Stats
	rw sync.RWMutex
}

// NewStatsHolder creates a new StatsHolder
func NewStatsHolder(firstBlockNum int64, refreshPeriod time.Duration) *StatsHolder {
	stats := Stats{}
	stats.Eth.RefreshPeriod = refreshPeriod
	stats.Eth.FirstBlockNum = firstBlockNum
	return &StatsHolder{Stats: stats}
}

// UpdateCurrentSlot updates the auction stats
func (s *StatsHolder) UpdateCurrentSlot(slot common.Slot) {
	s.rw.Lock()
	s.Sync.Auction.CurrentSlot = slot
	s.rw.Unlock()
}

// UpdateSync updates the synchronizer stats
func (s *StatsHolder) UpdateSync(lastBlock *common.Block, lastBatch *common.BatchNum,
	lastL1BatchBlock *int64, lastForgeL1TxsNum *int64) {
	now := time.Now()
	s.rw.Lock()
	s.Sync.LastBlock = *lastBlock
	if lastBatch != nil {
		s.Sync.LastBatch = int64(*lastBatch)
	}
	if lastL1BatchBlock != nil {
		s.Sync.LastL1BatchBlock = *lastL1BatchBlock
		s.Sync.LastForgeL1TxsNum = *lastForgeL1TxsNum
	}
	s.Sync.Updated = now
	s.rw.Unlock()
}

// UpdateEth updates the ethereum stats, only if the previous stats expired
func (s *StatsHolder) UpdateEth(ethClient eth.ClientInterface) error {
	now := time.Now()
	s.rw.RLock()
	elapsed := now.Sub(s.Eth.Updated)
	s.rw.RUnlock()
	if elapsed < s.Eth.RefreshPeriod {
		return nil
	}

	lastBlock, err := ethClient.EthBlockByNumber(context.TODO(), -1)
	if err != nil {
		return tracerr.Wrap(err)
	}
	lastBatch, err := ethClient.RollupLastForgedBatch()
	if err != nil {
		return tracerr.Wrap(err)
	}
	s.rw.Lock()
	s.Eth.Updated = now
	s.Eth.LastBlock = *lastBlock
	s.Eth.LastBatch = lastBatch
	s.rw.Unlock()
	return nil
}

// CopyStats returns a copy of the inner Stats
func (s *StatsHolder) CopyStats() *Stats {
	s.rw.RLock()
	sCopy := s.Stats
	if s.Sync.Auction.CurrentSlot.BidValue != nil {
		sCopy.Sync.Auction.CurrentSlot.BidValue =
			common.CopyBigInt(s.Sync.Auction.CurrentSlot.BidValue)
	}
	if s.Sync.Auction.CurrentSlot.DefaultSlotBid != nil {
		sCopy.Sync.Auction.CurrentSlot.DefaultSlotBid =
			common.CopyBigInt(s.Sync.Auction.CurrentSlot.DefaultSlotBid)
	}
	s.rw.RUnlock()
	return &sCopy
}

func (s *StatsHolder) blocksPerc() float64 {
	syncLastBlockNum := s.Sync.LastBlock.Num
	if s.Sync.LastBlock.Num == 0 {
		syncLastBlockNum = s.Eth.FirstBlockNum - 1
	}
	return float64(syncLastBlockNum-(s.Eth.FirstBlockNum-1)) * 100.0 /
		float64(s.Eth.LastBlock.Num-(s.Eth.FirstBlockNum-1))
}

func (s *StatsHolder) batchesPerc(batchNum int64) float64 {
	return float64(batchNum) * 100.0 /
		float64(s.Eth.LastBatch)
}

// ConfigStartBlockNum sets the first block used to start tracking the smart
// contracts
type ConfigStartBlockNum struct {
	Rollup   int64 `validate:"required"`
	Auction  int64 `validate:"required"`
	WDelayer int64 `validate:"required"`
}

// SCVariables joins all the smart contract variables in a single struct
type SCVariables struct {
	Rollup   common.RollupVariables   `validate:"required"`
	Auction  common.AuctionVariables  `validate:"required"`
	WDelayer common.WDelayerVariables `validate:"required"`
}

// SCVariablesPtr joins all the smart contract variables as pointers in a single
// struct
type SCVariablesPtr struct {
	Rollup   *common.RollupVariables   `validate:"required"`
	Auction  *common.AuctionVariables  `validate:"required"`
	WDelayer *common.WDelayerVariables `validate:"required"`
}

// SCConsts joins all the smart contract constants in a single struct
type SCConsts struct {
	Rollup   common.RollupConstants
	Auction  common.AuctionConstants
	WDelayer common.WDelayerConstants
}

// Config is the Synchronizer configuration
type Config struct {
	StartBlockNum      ConfigStartBlockNum
	InitialVariables   SCVariables
	StatsRefreshPeriod time.Duration
}

// Synchronizer implements the Synchronizer type
type Synchronizer struct {
	ethClient eth.ClientInterface
	// auctionConstants  common.AuctionConstants
	// rollupConstants   common.RollupConstants
	// wDelayerConstants common.WDelayerConstants
	consts        SCConsts
	historyDB     *historydb.HistoryDB
	stateDB       *statedb.StateDB
	cfg           Config
	startBlockNum int64
	vars          SCVariables
	stats         *StatsHolder
	// firstSavedBlock  *common.Block
	// mux sync.Mutex
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(ethClient eth.ClientInterface, historyDB *historydb.HistoryDB,
	stateDB *statedb.StateDB, cfg Config) (*Synchronizer, error) {
	auctionConstants, err := ethClient.AuctionConstants()
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("NewSynchronizer ethClient.AuctionConstants(): %w",
			err))
	}
	rollupConstants, err := ethClient.RollupConstants()
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("NewSynchronizer ethClient.RollupConstants(): %w",
			err))
	}
	wDelayerConstants, err := ethClient.WDelayerConstants()
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("NewSynchronizer ethClient.WDelayerConstants(): %w",
			err))
	}

	// Set startBlockNum to the minimum between Auction, Rollup and
	// WDelayer StartBlockNum
	startBlockNum := cfg.StartBlockNum.Auction
	if cfg.StartBlockNum.Rollup < startBlockNum {
		startBlockNum = cfg.StartBlockNum.Rollup
	}
	if cfg.StartBlockNum.WDelayer < startBlockNum {
		startBlockNum = cfg.StartBlockNum.WDelayer
	}
	stats := NewStatsHolder(startBlockNum, cfg.StatsRefreshPeriod)
	s := &Synchronizer{
		ethClient: ethClient,
		consts: SCConsts{
			Rollup:   *rollupConstants,
			Auction:  *auctionConstants,
			WDelayer: *wDelayerConstants,
		},
		historyDB:     historyDB,
		stateDB:       stateDB,
		cfg:           cfg,
		startBlockNum: startBlockNum,
		stats:         stats,
	}
	return s, s.init()
}

// Stats returns a copy of the Synchronizer Stats.  It is safe to call Stats()
// during a Sync call
func (s *Synchronizer) Stats() *Stats {
	return s.stats.CopyStats()
}

// AuctionConstants returns the AuctionConstants read from the smart contract
func (s *Synchronizer) AuctionConstants() *common.AuctionConstants {
	return &s.consts.Auction
}

// RollupConstants returns the RollupConstants read from the smart contract
func (s *Synchronizer) RollupConstants() *common.RollupConstants {
	return &s.consts.Rollup
}

// WDelayerConstants returns the WDelayerConstants read from the smart contract
func (s *Synchronizer) WDelayerConstants() *common.WDelayerConstants {
	return &s.consts.WDelayer
}

// SCVars returns a copy of the Smart Contract Variables
func (s *Synchronizer) SCVars() SCVariablesPtr {
	return SCVariablesPtr{
		Rollup:   s.vars.Rollup.Copy(),
		Auction:  s.vars.Auction.Copy(),
		WDelayer: s.vars.WDelayer.Copy(),
	}
}

func (s *Synchronizer) updateCurrentSlotIfSync(batchesLen int) error {
	slot := common.Slot{
		SlotNum:    s.stats.Sync.Auction.CurrentSlot.SlotNum,
		BatchesLen: int(s.stats.Sync.Auction.CurrentSlot.BatchesLen),
	}
	// We want the next block because the current one is already mined
	blockNum := s.stats.Sync.LastBlock.Num + 1
	slotNum := s.consts.Auction.SlotNum(blockNum)
	if batchesLen == -1 {
		dbBatchesLen, err := s.historyDB.GetBatchesLen(slotNum)
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("historyDB.GetBatchesLen: %w", err))
		}
		slot.BatchesLen = dbBatchesLen
	} else if slotNum > slot.SlotNum {
		slot.BatchesLen = batchesLen
	} else {
		slot.BatchesLen += batchesLen
	}
	slot.SlotNum = slotNum
	slot.StartBlock, slot.EndBlock = s.consts.Auction.SlotBlocks(slot.SlotNum)
	// If Synced, update the current coordinator
	if s.stats.Synced() && blockNum >= s.consts.Auction.GenesisBlockNum {
		bidCoord, err := s.historyDB.GetBestBidCoordinator(slot.SlotNum)
		if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
			return tracerr.Wrap(err)
		}
		if tracerr.Unwrap(err) == sql.ErrNoRows {
			slot.BootCoord = true
			slot.Forger = s.vars.Auction.BootCoordinator
			slot.URL = s.vars.Auction.BootCoordinatorURL
		} else if err == nil {
			slot.BidValue = bidCoord.BidValue
			slot.DefaultSlotBid = bidCoord.DefaultSlotSetBid[slot.SlotNum%6]
			// Only if the highest bid value is higher than the
			// default slot bid, the bidder is the winner of the
			// slot.  Otherwise the boot coordinator is the winner.
			if slot.BidValue.Cmp(slot.DefaultSlotBid) >= 0 {
				slot.Bidder = bidCoord.Bidder
				slot.Forger = bidCoord.Forger
				slot.URL = bidCoord.URL
			} else {
				slot.BootCoord = true
				slot.Forger = s.vars.Auction.BootCoordinator
				slot.URL = s.vars.Auction.BootCoordinatorURL
			}
		}

		// TODO: Remove this SANITY CHECK once this code is tested enough
		// BEGIN SANITY CHECK
		canForge, err := s.ethClient.AuctionCanForge(slot.Forger, blockNum)
		if err != nil {
			return tracerr.Wrap(err)
		}
		if !canForge {
			return tracerr.Wrap(fmt.Errorf("Synchronized value of forger address for closed slot "+
				"differs from smart contract: %+v", slot))
		}
		// END SANITY CHECK
	}
	s.stats.UpdateCurrentSlot(slot)
	return nil
}

func (s *Synchronizer) init() error {
	// Update stats parameters so that they have valid values before the
	// first Sync call
	if err := s.stats.UpdateEth(s.ethClient); err != nil {
		return tracerr.Wrap(err)
	}
	lastBlock := &common.Block{}
	lastSavedBlock, err := s.historyDB.GetLastBlock()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// If we only have the default block 0,
	// make sure that the stateDB is clean
	if lastSavedBlock.Num == 0 {
		if err := s.stateDB.Reset(0); err != nil {
			return tracerr.Wrap(err)
		}
	} else {
		lastBlock = lastSavedBlock
	}
	if err := s.resetState(lastBlock); err != nil {
		return tracerr.Wrap(err)
	}

	log.Infow("Sync init block",
		"syncLastBlock", s.stats.Sync.LastBlock,
		"syncBlocksPerc", s.stats.blocksPerc(),
		"ethFirstBlockNum", s.stats.Eth.FirstBlockNum,
		"ethLastBlock", s.stats.Eth.LastBlock,
	)
	log.Infow("Sync init batch",
		"syncLastBatch", s.stats.Sync.LastBatch,
		"syncBatchesPerc", s.stats.batchesPerc(s.stats.Sync.LastBatch),
		"ethLastBatch", s.stats.Eth.LastBatch,
	)
	return nil
}

// Sync2 attems to synchronize an ethereum block starting from lastSavedBlock.
// If lastSavedBlock is nil, the lastSavedBlock value is obtained from de DB.
// If a block is synched, it will be returned and also stored in the DB.  If a
// reorg is detected, the number of discarded blocks will be returned and no
// synchronization will be made.
// TODO: Be smart about locking: only lock during the read/write operations
func (s *Synchronizer) Sync2(ctx context.Context, lastSavedBlock *common.Block) (*common.BlockData, *int64, error) {
	var nextBlockNum int64 // next block number to sync
	if lastSavedBlock == nil {
		var err error
		// Get lastSavedBlock from History DB
		lastSavedBlock, err = s.historyDB.GetLastBlock()
		if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
			return nil, nil, tracerr.Wrap(err)
		}
		// If we don't have any stored block, we must do a full sync
		// starting from the startBlockNum
		if tracerr.Unwrap(err) == sql.ErrNoRows || lastSavedBlock.Num == 0 {
			nextBlockNum = s.startBlockNum
			lastSavedBlock = nil
		}
	}
	if lastSavedBlock != nil {
		nextBlockNum = lastSavedBlock.Num + 1
		if lastSavedBlock.Num < s.startBlockNum {
			return nil, nil, tracerr.Wrap(
				fmt.Errorf("lastSavedBlock (%v) < startBlockNum (%v)",
					lastSavedBlock.Num, s.startBlockNum))
		}
	}

	ethBlock, err := s.ethClient.EthBlockByNumber(ctx, nextBlockNum)
	if tracerr.Unwrap(err) == ethereum.NotFound {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}
	log.Debugf("ethBlock: num: %v, parent: %v, hash: %v",
		ethBlock.Num, ethBlock.ParentHash.String(), ethBlock.Hash.String())

	if err := s.stats.UpdateEth(s.ethClient); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	log.Debugw("Syncing...",
		"block", nextBlockNum,
		"ethLastBlock", s.stats.Eth.LastBlock,
	)

	// Check that the obtianed ethBlock.ParentHash == prevEthBlock.Hash; if not, reorg!
	if lastSavedBlock != nil {
		if lastSavedBlock.Hash != ethBlock.ParentHash {
			// Reorg detected
			log.Debugw("Reorg Detected",
				"blockNum", ethBlock.Num,
				"block.parent(got)", ethBlock.ParentHash, "parent.hash(exp)", lastSavedBlock.Hash)
			lastDBBlockNum, err := s.reorg(lastSavedBlock)
			if err != nil {
				return nil, nil, tracerr.Wrap(err)
			}
			discarded := lastSavedBlock.Num - lastDBBlockNum
			return nil, &discarded, nil
		}
	}

	// Get data from the rollup contract
	rollupData, err := s.rollupSync(ethBlock)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	// Get data from the auction contract
	auctionData, err := s.auctionSync(ethBlock)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	// Get data from the WithdrawalDelayer contract
	wDelayerData, err := s.wdelayerSync(ethBlock)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	for i := range rollupData.Withdrawals {
		withdrawal := &rollupData.Withdrawals[i]
		if !withdrawal.InstantWithdraw {
			wDelayerTransfers := wDelayerData.DepositsByTxHash[withdrawal.TxHash]
			if len(wDelayerTransfers) == 0 {
				return nil, nil, tracerr.Wrap(fmt.Errorf("WDelayer deposit corresponding to " +
					"non-instant rollup withdrawal not found"))
			}
			// Pop the first wDelayerTransfer to consume them in chronological order
			wDelayerTransfer := wDelayerTransfers[0]
			wDelayerData.DepositsByTxHash[withdrawal.TxHash] =
				wDelayerData.DepositsByTxHash[withdrawal.TxHash][1:]

			withdrawal.Owner = wDelayerTransfer.Owner
			withdrawal.Token = wDelayerTransfer.Token
		}
	}

	// Group all the block data into the structs to save into HistoryDB
	blockData := common.BlockData{
		Block:    *ethBlock,
		Rollup:   *rollupData,
		Auction:  *auctionData,
		WDelayer: *wDelayerData,
	}

	// log.Debugw("Sync()", "block", blockData)
	// err = s.historyDB.AddBlock(blockData.Block)
	// if err != nil {
	// 	return err
	// }
	err = s.historyDB.AddBlockSCData(&blockData)
	if err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	batchesLen := len(rollupData.Batches)
	if batchesLen == 0 {
		s.stats.UpdateSync(ethBlock, nil, nil, nil)
	} else {
		var lastL1BatchBlock *int64
		var lastForgeL1TxsNum *int64
		for _, batchData := range rollupData.Batches {
			if batchData.L1Batch {
				lastL1BatchBlock = &batchData.Batch.EthBlockNum
				lastForgeL1TxsNum = batchData.Batch.ForgeL1TxsNum
			}
		}
		s.stats.UpdateSync(ethBlock,
			&rollupData.Batches[batchesLen-1].Batch.BatchNum, lastL1BatchBlock, lastForgeL1TxsNum)
	}
	if err := s.updateCurrentSlotIfSync(len(rollupData.Batches)); err != nil {
		return nil, nil, tracerr.Wrap(err)
	}

	log.Debugw("Synced block",
		"syncLastBlockNum", s.stats.Sync.LastBlock.Num,
		"syncBlocksPerc", s.stats.blocksPerc(),
		"ethLastBlockNum", s.stats.Eth.LastBlock.Num,
	)
	for _, batchData := range rollupData.Batches {
		log.Debugw("Synced batch",
			"syncLastBatch", batchData.Batch.BatchNum,
			"syncBatchesPerc", s.stats.batchesPerc(int64(batchData.Batch.BatchNum)),
			"ethLastBatch", s.stats.Eth.LastBatch,
		)
	}

	return &blockData, nil, nil
}

// reorg manages a reorg, updating History and State DB as needed.  Keeps
// checking previous blocks from the HistoryDB against the blockchain until a
// block hash match is found.  All future blocks in the HistoryDB and
// corresponding batches in StateBD are discarded.  Returns the last valid
// blockNum from the HistoryDB.
func (s *Synchronizer) reorg(uncleBlock *common.Block) (int64, error) {
	blockNum := uncleBlock.Num

	var block *common.Block
	for blockNum >= s.startBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), blockNum)
		if err != nil {
			return 0, tracerr.Wrap(fmt.Errorf("ethClient.EthBlockByNumber: %w", err))
		}

		block, err = s.historyDB.GetBlock(blockNum)
		if err != nil {
			return 0, tracerr.Wrap(fmt.Errorf("historyDB.GetBlock: %w", err))
		}
		if block.Hash == ethBlock.Hash {
			log.Debugf("Found valid block: %v", blockNum)
			break
		}
		blockNum--
	}
	total := uncleBlock.Num - block.Num
	log.Debugw("Discarding blocks", "total", total, "from", uncleBlock.Num, "to", block.Num+1)

	// Set History DB and State DB to the correct state
	err := s.historyDB.Reorg(block.Num)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}

	if err := s.resetState(block); err != nil {
		return 0, tracerr.Wrap(err)
	}

	return block.Num, nil
}

func (s *Synchronizer) resetState(block *common.Block) error {
	rollup, auction, wDelayer, err := s.historyDB.GetSCVars()
	// If SCVars are not in the HistoryDB, this is probably the first run
	// of the Synchronizer: store the initial vars taken from config
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		rollup = &s.cfg.InitialVariables.Rollup
		auction = &s.cfg.InitialVariables.Auction
		wDelayer = &s.cfg.InitialVariables.WDelayer
		log.Info("Setting initial SCVars in HistoryDB")
		if err = s.historyDB.SetInitialSCVars(rollup, auction, wDelayer); err != nil {
			return tracerr.Wrap(fmt.Errorf("historyDB.SetInitialSCVars: %w", err))
		}
		// Add initial boot coordinator to HistoryDB
		if err := s.historyDB.AddCoordinators([]common.Coordinator{{
			Forger:      auction.BootCoordinator,
			URL:         auction.BootCoordinatorURL,
			EthBlockNum: auction.EthBlockNum,
		}}); err != nil {
			return tracerr.Wrap(err)
		}
	}
	s.vars.Rollup = *rollup
	s.vars.Auction = *auction
	s.vars.WDelayer = *wDelayer

	batchNum, err := s.historyDB.GetLastBatchNum()
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		return tracerr.Wrap(fmt.Errorf("historyDB.GetLastBatchNum: %w", err))
	}
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		batchNum = 0
	}

	lastL1BatchBlockNum, err := s.historyDB.GetLastL1BatchBlockNum()
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		return tracerr.Wrap(fmt.Errorf("historyDB.GetLastL1BatchBlockNum: %w", err))
	}
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		lastL1BatchBlockNum = 0
	}

	lastForgeL1TxsNum, err := s.historyDB.GetLastL1TxsNum()
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		return tracerr.Wrap(fmt.Errorf("historyDB.GetLastL1BatchBlockNum: %w", err))
	}
	if tracerr.Unwrap(err) == sql.ErrNoRows || lastForgeL1TxsNum == nil {
		n := int64(-1)
		lastForgeL1TxsNum = &n
	}

	err = s.stateDB.Reset(batchNum)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("stateDB.Reset: %w", err))
	}

	s.stats.UpdateSync(block, &batchNum, &lastL1BatchBlockNum, lastForgeL1TxsNum)

	if err := s.updateCurrentSlotIfSync(-1); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// rollupSync retreives all the Rollup Smart Contract Data that happened at
// ethBlock.blockNum with ethBlock.Hash.
func (s *Synchronizer) rollupSync(ethBlock *common.Block) (*common.RollupData, error) {
	blockNum := ethBlock.Num
	var rollupData = common.NewRollupData()
	// var forgeL1TxsNum int64

	// Get rollup events in the block, and make sure the block hash matches
	// the expected one.
	rollupEvents, blockHash, err := s.ethClient.RollupEventsByBlock(blockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// No events in this block
	if blockHash == nil {
		return &rollupData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in Rollup events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
		return nil, tracerr.Wrap(eth.ErrBlockHashMismatchEvent)
	}

	var nextForgeL1TxsNum int64 // forgeL1TxsNum for the next L1Batch
	nextForgeL1TxsNumPtr, err := s.historyDB.GetLastL1TxsNum()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if nextForgeL1TxsNumPtr != nil {
		nextForgeL1TxsNum = *nextForgeL1TxsNumPtr + 1
	} else {
		nextForgeL1TxsNum = 0
	}

	// Get L1UserTX
	rollupData.L1UserTxs, err = getL1UserTx(rollupEvents.L1UserTx, blockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// Get ForgeBatch events to get the L1CoordinatorTxs
	for _, evtForgeBatch := range rollupEvents.ForgeBatch {
		batchData := common.NewBatchData()
		position := 0

		// Get the input for each Tx
		forgeBatchArgs, sender, err := s.ethClient.RollupForgeBatchArgs(evtForgeBatch.EthTxHash,
			evtForgeBatch.L1UserTxsLen)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		batchNum := common.BatchNum(evtForgeBatch.BatchNum)
		var l1UserTxs []common.L1Tx
		// Check if this is a L1Batch to get L1 Tx from it
		if forgeBatchArgs.L1Batch {
			// Get L1UserTxs with toForgeL1TxsNum, which correspond
			// to the L1UserTxs that are forged in this batch, so
			// that stateDB can process them.

			// First try to find them in HistoryDB.
			l1UserTxs, err = s.historyDB.GetUnforgedL1UserTxs(nextForgeL1TxsNum)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			// Apart from the DB, try to find them in this block.
			// This could happen because in a block there could be
			// multiple batches with L1Batch = true (although it's
			// a very rare case).  If not found in the DB and the
			// block doesn't contain the l1UserTxs, it means that
			// the L1UserTxs queue with toForgeL1TxsNum was closed
			// empty, so we leave `l1UserTxs` as an empty slice.
			for _, l1UserTx := range rollupData.L1UserTxs {
				if *l1UserTx.ToForgeL1TxsNum == nextForgeL1TxsNum {
					l1UserTxs = append(l1UserTxs, l1UserTx)
				}
			}

			position = len(l1UserTxs)
		}
		// Get L1 Coordinator Txs
		for i := range forgeBatchArgs.L1CoordinatorTxs {
			l1CoordinatorTx := forgeBatchArgs.L1CoordinatorTxs[i]
			l1CoordinatorTx.Position = position
			// l1CoordinatorTx.ToForgeL1TxsNum = &forgeL1TxsNum
			l1CoordinatorTx.UserOrigin = false
			l1CoordinatorTx.EthBlockNum = blockNum
			l1CoordinatorTx.BatchNum = &batchNum
			l1Tx, err := common.NewL1Tx(&l1CoordinatorTx)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}

			batchData.L1CoordinatorTxs = append(batchData.L1CoordinatorTxs, *l1Tx)
			position++
			// fmt.Println("DGB l1coordtx")
		}

		// Insert all the txs forged in this batch (l1UserTxs,
		// L1CoordinatorTxs, PoolL2Txs) into stateDB so that they are
		// processed.

		// Add TxID, TxType to L2 txs
		for i := range forgeBatchArgs.L2TxsData {
			nTx, err := common.NewL2Tx(&forgeBatchArgs.L2TxsData[i])
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			forgeBatchArgs.L2TxsData[i] = *nTx
		}

		// Transform L2 txs to PoolL2Txs
		poolL2Txs := common.L2TxsToPoolL2Txs(forgeBatchArgs.L2TxsData) // NOTE: This is a big ugly, find a better way

		// ProcessTxs updates poolL2Txs adding: Nonce (and also TokenID, but we don't use it).
		//nolint:gomnd
		ptc := statedb.ProcessTxsConfig{ // TODO TMP
			NLevels:  32,
			MaxFeeTx: 64,
			MaxTx:    512,
			MaxL1Tx:  64,
		}
		processTxsOut, err := s.stateDB.ProcessTxs(ptc, forgeBatchArgs.FeeIdxCoordinator,
			l1UserTxs, batchData.L1CoordinatorTxs, poolL2Txs)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		// Transform processed PoolL2 txs to L2 and store in BatchData
		if poolL2Txs != nil {
			l2Txs, err := common.PoolL2TxsToL2Txs(poolL2Txs) // NOTE: This is a big uggly, find a better way
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			for i := range l2Txs {
				l2Txs[i].Position = position
				l2Txs[i].EthBlockNum = blockNum
				l2Txs[i].BatchNum = batchNum
				position++
				// At this point TxID should be incorrect, since it was calculated before the nonce being setted
				// therefore TxID has to be calculated again. Furthermore the nonce should be -1 ed
				// TODO: StateDB should return the correct nonce
				// TODO: use set id method once the code is rebased with master
				// TODO: add unit test to check the TxID in historyDB is correct with synced L2Txs who's nonce > 0
				l2Txs[i].Nonce--
				txWithID, err := common.NewL2Tx(&l2Txs[i])
				if err != nil {
					return nil, tracerr.Wrap(err)
				}
				l2Txs[i] = *txWithID
			}
			batchData.L2Txs = l2Txs
		}

		// Set the BatchNum in the forged L1UserTxs
		for i := range l1UserTxs {
			l1UserTxs[i].BatchNum = &batchNum
		}
		batchData.L1UserTxs = l1UserTxs

		// Set batchNum in exits
		for i := range processTxsOut.ExitInfos {
			exit := &processTxsOut.ExitInfos[i]
			exit.BatchNum = batchNum
		}
		batchData.ExitTree = processTxsOut.ExitInfos

		for i := range processTxsOut.CreatedAccounts {
			createdAccount := &processTxsOut.CreatedAccounts[i]
			createdAccount.Nonce = 0
			createdAccount.Balance = big.NewInt(0)
			createdAccount.BatchNum = batchNum
		}
		batchData.CreatedAccounts = processTxsOut.CreatedAccounts

		slotNum := int64(0)
		if ethBlock.Num >= s.consts.Auction.GenesisBlockNum {
			slotNum = (ethBlock.Num - s.consts.Auction.GenesisBlockNum) /
				int64(s.consts.Auction.BlocksPerSlot)
		}

		// Get Batch information
		batch := common.Batch{
			BatchNum:           batchNum,
			EthBlockNum:        blockNum,
			ForgerAddr:         *sender,
			CollectedFees:      processTxsOut.CollectedFees,
			FeeIdxsCoordinator: forgeBatchArgs.FeeIdxCoordinator,
			StateRoot:          forgeBatchArgs.NewStRoot,
			NumAccounts:        len(batchData.CreatedAccounts),
			LastIdx:            forgeBatchArgs.NewLastIdx,
			ExitRoot:           forgeBatchArgs.NewExitRoot,
			SlotNum:            slotNum,
		}
		nextForgeL1TxsNumCpy := nextForgeL1TxsNum
		if forgeBatchArgs.L1Batch {
			batch.ForgeL1TxsNum = &nextForgeL1TxsNumCpy
			batchData.L1Batch = true
			nextForgeL1TxsNum++
		}
		batchData.Batch = batch
		rollupData.Batches = append(rollupData.Batches, *batchData)
	}

	// Get Registered Tokens
	for _, evtAddToken := range rollupEvents.AddToken {
		var token common.Token

		token.TokenID = common.TokenID(evtAddToken.TokenID)
		token.EthAddr = evtAddToken.TokenAddress
		token.EthBlockNum = blockNum

		if consts, err := s.ethClient.EthERC20Consts(evtAddToken.TokenAddress); err != nil {
			log.Warnw("Error retreiving ERC20 token constants", "addr", evtAddToken.TokenAddress)
			token.Name = "ERC20_ETH_ERROR"
			token.Symbol = "ERROR"
			token.Decimals = 1
		} else {
			token.Name = cutStringMax(consts.Name, 20)
			token.Symbol = cutStringMax(consts.Symbol, 10)
			token.Decimals = consts.Decimals
		}

		rollupData.AddedTokens = append(rollupData.AddedTokens, token)
	}

	for _, evt := range rollupEvents.UpdateBucketWithdraw {
		rollupData.UpdateBucketWithdraw = append(rollupData.UpdateBucketWithdraw,
			common.BucketUpdate{
				EthBlockNum: blockNum,
				NumBucket:   evt.NumBucket,
				BlockStamp:  evt.BlockStamp,
				Withdrawals: evt.Withdrawals,
			})
	}

	for _, evt := range rollupEvents.Withdraw {
		rollupData.Withdrawals = append(rollupData.Withdrawals, common.WithdrawInfo{
			Idx:             common.Idx(evt.Idx),
			NumExitRoot:     common.BatchNum(evt.NumExitRoot),
			InstantWithdraw: evt.InstantWithdraw,
			TxHash:          evt.TxHash,
		})
	}

	for _, evt := range rollupEvents.UpdateTokenExchange {
		if len(evt.AddressArray) != len(evt.ValueArray) {
			return nil, tracerr.Wrap(fmt.Errorf("in RollupEventUpdateTokenExchange "+
				"len(AddressArray) != len(ValueArray) (%v != %v)",
				len(evt.AddressArray), len(evt.ValueArray)))
		}
		for i := range evt.AddressArray {
			rollupData.TokenExchanges = append(rollupData.TokenExchanges,
				common.TokenExchange{
					EthBlockNum: blockNum,
					Address:     evt.AddressArray[i],
					ValueUSD:    int64(evt.ValueArray[i]),
				})
		}
	}

	varsUpdate := false

	for _, evt := range rollupEvents.UpdateForgeL1L2BatchTimeout {
		s.vars.Rollup.ForgeL1L2BatchTimeout = evt.NewForgeL1L2BatchTimeout
		varsUpdate = true
	}

	for _, evt := range rollupEvents.UpdateFeeAddToken {
		s.vars.Rollup.FeeAddToken = evt.NewFeeAddToken
		varsUpdate = true
	}

	for _, evt := range rollupEvents.UpdateWithdrawalDelay {
		s.vars.Rollup.WithdrawalDelay = evt.NewWithdrawalDelay
		varsUpdate = true
	}

	// NOTE: We skip the event rollupEvents.SafeMode because the
	// implementation RollupEventsByBlock already inserts a non-existing
	// RollupEventUpdateBucketsParameters into UpdateBucketsParameters with
	// all the bucket values at 0 and SafeMode = true

	for _, evt := range rollupEvents.UpdateBucketsParameters {
		for i, bucket := range evt.ArrayBuckets {
			s.vars.Rollup.Buckets[i] = common.BucketParams{
				CeilUSD:             bucket.CeilUSD,
				Withdrawals:         bucket.Withdrawals,
				BlockWithdrawalRate: bucket.BlockWithdrawalRate,
				MaxWithdrawals:      bucket.MaxWithdrawals,
			}
		}
		s.vars.Rollup.SafeMode = evt.SafeMode
		varsUpdate = true
	}

	if varsUpdate {
		s.vars.Rollup.EthBlockNum = blockNum
		rollupData.Vars = s.vars.Rollup.Copy()
	}

	return &rollupData, nil
}

func cutStringMax(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// auctionSync gets information from the Auction Contract
func (s *Synchronizer) auctionSync(ethBlock *common.Block) (*common.AuctionData, error) {
	blockNum := ethBlock.Num
	var auctionData = common.NewAuctionData()

	// Get auction events in the block
	auctionEvents, blockHash, err := s.ethClient.AuctionEventsByBlock(blockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// No events in this block
	if blockHash == nil {
		return &auctionData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in Auction events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
		return nil, tracerr.Wrap(eth.ErrBlockHashMismatchEvent)
	}

	// Get bids
	for _, evt := range auctionEvents.NewBid {
		bid := common.Bid{
			SlotNum:     evt.Slot,
			BidValue:    evt.BidAmount,
			Bidder:      evt.Bidder,
			EthBlockNum: blockNum,
		}
		auctionData.Bids = append(auctionData.Bids, bid)
	}

	// Get Coordinators
	for _, evt := range auctionEvents.SetCoordinator {
		coordinator := common.Coordinator{
			Bidder:      evt.BidderAddress,
			Forger:      evt.ForgerAddress,
			URL:         evt.CoordinatorURL,
			EthBlockNum: blockNum,
		}
		auctionData.Coordinators = append(auctionData.Coordinators, coordinator)
	}

	varsUpdate := false

	for _, evt := range auctionEvents.NewSlotDeadline {
		s.vars.Auction.SlotDeadline = evt.NewSlotDeadline
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewClosedAuctionSlots {
		s.vars.Auction.ClosedAuctionSlots = evt.NewClosedAuctionSlots
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewOutbidding {
		s.vars.Auction.Outbidding = evt.NewOutbidding
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewDonationAddress {
		s.vars.Auction.DonationAddress = evt.NewDonationAddress
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewBootCoordinator {
		s.vars.Auction.BootCoordinator = evt.NewBootCoordinator
		s.vars.Auction.BootCoordinatorURL = evt.NewBootCoordinatorURL
		varsUpdate = true
		// Add new boot coordinator
		auctionData.Coordinators = append(auctionData.Coordinators, common.Coordinator{
			Forger:      evt.NewBootCoordinator,
			URL:         evt.NewBootCoordinatorURL,
			EthBlockNum: blockNum,
		})
	}
	for _, evt := range auctionEvents.NewOpenAuctionSlots {
		s.vars.Auction.OpenAuctionSlots = evt.NewOpenAuctionSlots
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewAllocationRatio {
		s.vars.Auction.AllocationRatio = evt.NewAllocationRatio
		varsUpdate = true
	}
	for _, evt := range auctionEvents.NewDefaultSlotSetBid {
		if evt.SlotSet > 6 { //nolint:gomnd
			return nil, tracerr.Wrap(fmt.Errorf("unexpected SlotSet in "+
				"auctionEvents.NewDefaultSlotSetBid: %v", evt.SlotSet))
		}
		s.vars.Auction.DefaultSlotSetBid[evt.SlotSet] = evt.NewInitialMinBid
		s.vars.Auction.DefaultSlotSetBidSlotNum = s.consts.Auction.SlotNum(blockNum) +
			int64(s.vars.Auction.ClosedAuctionSlots) + 1
		varsUpdate = true
	}

	// NOTE: We ignore NewForgeAllocated
	// NOTE: We ignore NewForge because we're already tracking ForgeBatch event from Rollup
	// NOTE: We ignore HEZClaimed

	if varsUpdate {
		s.vars.Auction.EthBlockNum = blockNum
		auctionData.Vars = s.vars.Auction.Copy()
	}

	return &auctionData, nil
}

// wdelayerSync gets information from the Withdrawal Delayer Contract
func (s *Synchronizer) wdelayerSync(ethBlock *common.Block) (*common.WDelayerData, error) {
	blockNum := ethBlock.Num
	wDelayerData := common.NewWDelayerData()

	// Get wDelayer events in the block
	wDelayerEvents, blockHash, err := s.ethClient.WDelayerEventsByBlock(blockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	// No events in this block
	if blockHash == nil {
		return &wDelayerData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in WDelayer events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
		return nil, tracerr.Wrap(eth.ErrBlockHashMismatchEvent)
	}

	for _, evt := range wDelayerEvents.Deposit {
		wDelayerData.Deposits = append(wDelayerData.Deposits, common.WDelayerTransfer{
			Owner:  evt.Owner,
			Token:  evt.Token,
			Amount: evt.Amount,
		})
		wDelayerData.DepositsByTxHash[evt.TxHash] =
			append(wDelayerData.DepositsByTxHash[evt.TxHash],
				&wDelayerData.Deposits[len(wDelayerData.Deposits)-1])
	}
	for _, evt := range wDelayerEvents.Withdraw {
		wDelayerData.Withdrawals = append(wDelayerData.Withdrawals, common.WDelayerTransfer{
			Owner:  evt.Owner,
			Token:  evt.Token,
			Amount: evt.Amount,
		})
	}
	for _, evt := range wDelayerEvents.EscapeHatchWithdrawal {
		wDelayerData.EscapeHatchWithdrawals = append(wDelayerData.EscapeHatchWithdrawals,
			common.WDelayerEscapeHatchWithdrawal{
				EthBlockNum: blockNum,
				Who:         evt.Who,
				To:          evt.To,
				TokenAddr:   evt.Token,
				Amount:      evt.Amount,
			})
	}

	varsUpdate := false

	for range wDelayerEvents.EmergencyModeEnabled {
		s.vars.WDelayer.EmergencyMode = true
		s.vars.WDelayer.EmergencyModeStartingBlock = blockNum
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewWithdrawalDelay {
		s.vars.WDelayer.WithdrawalDelay = evt.WithdrawalDelay
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewEmergencyCouncil {
		s.vars.WDelayer.EmergencyCouncilAddress = evt.NewEmergencyCouncil
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewHermezGovernanceAddress {
		s.vars.WDelayer.HermezGovernanceAddress = evt.NewHermezGovernanceAddress
		varsUpdate = true
	}

	if varsUpdate {
		s.vars.WDelayer.EthBlockNum = blockNum
		wDelayerData.Vars = s.vars.WDelayer.Copy()
	}

	return &wDelayerData, nil
}

func getL1UserTx(eventsL1UserTx []eth.RollupEventL1UserTx, blockNum int64) ([]common.L1Tx, error) {
	l1Txs := make([]common.L1Tx, len(eventsL1UserTx))
	for i := range eventsL1UserTx {
		eventsL1UserTx[i].L1UserTx.EthBlockNum = blockNum
		// Check validity of L1UserTx
		l1Tx, err := common.NewL1Tx(&eventsL1UserTx[i].L1UserTx)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		l1Txs[i] = *l1Tx
	}
	return l1Txs, nil
}
