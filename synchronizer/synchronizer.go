package synchronizer

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
)

var (
	// ErrNotAbleToSync is used when there is not possible to find a valid block to sync
	ErrNotAbleToSync = errors.New("it has not been possible to synchronize any block")
)

// rollupData contains information returned by the Rollup SC
type rollupData struct {
	l1Txs   []*common.L1Tx
	batches []*BatchData
	// withdrawals      []*common.ExitInfo
	registeredTokens []*common.Token
	rollupVars       *common.RollupVars
}

// NewRollupData creates an empty rollupData with the slices initialized.
func newRollupData() rollupData {
	return rollupData{
		l1Txs:   make([]*common.L1Tx, 0),
		batches: make([]*BatchData, 0),
		// withdrawals:      make([]*common.ExitInfo, 0),
		registeredTokens: make([]*common.Token, 0),
	}
}

// auctionData contains information returned by the Action SC
type auctionData struct {
	bids         []*common.Bid
	coordinators []*common.Coordinator
	auctionVars  *common.AuctionVars
}

// newAuctionData creates an empty auctionData with the slices initialized.
func newAuctionData() *auctionData {
	return &auctionData{
		bids:         make([]*common.Bid, 0),
		coordinators: make([]*common.Coordinator, 0),
	}
}

// BatchData contains information about Batches from the contracts
type BatchData struct {
	l1UserTxs        []*common.L1Tx
	l1CoordinatorTxs []*common.L1Tx
	l2Txs            []*common.L2Tx
	createdAccounts  []*common.Account
	exitTree         []common.ExitInfo
	batch            *common.Batch
}

// NewBatchData creates an empty BatchData with the slices initialized.
func NewBatchData() *BatchData {
	return &BatchData{
		l1UserTxs:        make([]*common.L1Tx, 0),
		l1CoordinatorTxs: make([]*common.L1Tx, 0),
		l2Txs:            make([]*common.L2Tx, 0),
		createdAccounts:  make([]*common.Account, 0),
		exitTree:         make([]common.ExitInfo, 0),
	}
}

// BlockData contains information about Blocks from the contracts
type BlockData struct {
	block *common.Block
	// Rollup
	l1Txs   []*common.L1Tx // TODO: Answer: User? Coordinator? Both?
	batches []*BatchData   // TODO: Also contains L1Txs!
	// withdrawals      []*common.ExitInfo // TODO
	registeredTokens []*common.Token
	rollupVars       *common.RollupVars
	// Auction
	bids         []*common.Bid
	coordinators []*common.Coordinator
	auctionVars  *common.AuctionVars
	// WithdrawalDelayer
	withdrawalDelayerVars *common.WithdrawalDelayerVars
}

// Synchronizer implements the Synchronizer type
type Synchronizer struct {
	ethClient       *eth.Client
	historyDB       *historydb.HistoryDB
	stateDB         *statedb.StateDB
	firstSavedBlock *common.Block
	mux             sync.Mutex
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(ethClient *eth.Client, historyDB *historydb.HistoryDB, stateDB *statedb.StateDB) *Synchronizer {
	s := &Synchronizer{
		ethClient: ethClient,
		historyDB: historyDB,
		stateDB:   stateDB,
	}
	return s
}

// TODO: Be smart about locking: only lock during the read/write operations

// Sync updates History and State DB with information from the blockchain
// TODO: Return true if a new block was processed
// TODO: Add argument: maximum number of blocks to process
// TODO: Check reorgs in the middle of syncing a block.  Probably make
// rollupSync, auctionSync and withdrawalSync return the block hash.
func (s *Synchronizer) Sync() error {
	// Avoid new sync while performing one
	s.mux.Lock()
	defer s.mux.Unlock()

	var nextBlockNum int64 // next block number to sync

	// Get lastSavedBlock from History DB
	lastSavedBlock, err := s.historyDB.GetLastBlock()
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	// If we don't have any stored block, we must do a full sync starting from the rollup genesis block
	if err == sql.ErrNoRows {
		// TODO: Query rollup constants and genesis information, store them
		nextBlockNum = 1234 // TODO: Replace this with genesisBlockNum
	} else {
		// Get the latest block we have in History DB from blockchain to detect a reorg
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), lastSavedBlock.EthBlockNum)
		if err != nil {
			return err
		}

		if ethBlock.Hash != lastSavedBlock.Hash {
			// Reorg detected
			log.Debugf("Reorg Detected...")
			_, err := s.reorg(lastSavedBlock)
			if err != nil {
				return err
			}

			lastSavedBlock, err = s.historyDB.GetLastBlock()
			if err != nil {
				return err
			}
		}
		nextBlockNum = lastSavedBlock.EthBlockNum + 1
	}

	log.Debugf("Syncing...")

	// Get latest blockNum in blockchain
	latestBlockNum, err := s.ethClient.EthCurrentBlock()
	if err != nil {
		return err
	}

	log.Debugf("Blocks to sync: %v (firstBlockToSync: %v, latestBlock: %v)", latestBlockNum-nextBlockNum+1, nextBlockNum, latestBlockNum)

	for nextBlockNum < latestBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), nextBlockNum)
		if err != nil {
			return err
		}
		// TODO: Check that the obtianed ethBlock.ParentHash == prevEthBlock.Hash; if not, reorg!

		// TODO: Send the ethHash in rollupSync(), auctionSync() and
		// wdelayerSync() and make sure they all use the same block
		// hash.

		// Get data from the rollup contract
		rollupData, err := s.rollupSync(nextBlockNum)
		if err != nil {
			return err
		}

		// Get data from the auction contract
		auctionData, err := s.auctionSync(nextBlockNum)
		if err != nil {
			return err
		}

		// Get data from the WithdrawalDelayer contract
		wdelayerData, err := s.wdelayerSync(nextBlockNum)
		if err != nil {
			return err
		}

		// Group all the block data into the structs to save into HistoryDB
		var blockData BlockData

		blockData.block = ethBlock

		if rollupData != nil {
			blockData.l1Txs = rollupData.l1Txs
			blockData.batches = rollupData.batches
			// blockData.withdrawals = rollupData.withdrawals // TODO
			blockData.registeredTokens = rollupData.registeredTokens
			blockData.rollupVars = rollupData.rollupVars
		}

		if auctionData != nil {
			blockData.bids = auctionData.bids
			blockData.coordinators = auctionData.coordinators
			blockData.auctionVars = auctionData.auctionVars
		}

		if wdelayerData != nil {
			blockData.withdrawalDelayerVars = wdelayerData
		}

		// Add rollupData and auctionData once the method is updated
		// TODO: Save Whole Struct -> AddBlockSCData(blockData)
		err = s.historyDB.AddBlock(blockData.block)
		if err != nil {
			return err
		}
	}

	return nil
}

// reorg manages a reorg, updating History and State DB as needed.  Keeps
// checking previous blocks from the HistoryDB against the blockchain until a
// block hash match is found.  All future blocks in the HistoryDB and
// corresponding batches in StateBD are discarded.  Returns the last valid
// blockNum from the HistoryDB.
func (s *Synchronizer) reorg(uncleBlock *common.Block) (int64, error) {
	var block *common.Block
	blockNum := uncleBlock.EthBlockNum
	found := false

	log.Debugf("Reorg first uncle block: %v", blockNum)

	// Iterate History DB and the blokchain looking for the latest valid block
	for !found && blockNum > s.firstSavedBlock.EthBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), blockNum)
		if err != nil {
			return 0, err
		}

		block, err = s.historyDB.GetBlock(blockNum)
		if err != nil {
			return 0, err
		}
		if block.Hash == ethBlock.Hash {
			found = true
			log.Debugf("Found valid block: %v", blockNum)
		} else {
			log.Debugf("Discarding block: %v", blockNum)
		}

		blockNum--
	}

	if found {
		// Set History DB and State DB to the correct state
		err := s.historyDB.Reorg(block.EthBlockNum)
		if err != nil {
			return 0, err
		}

		batchNum, err := s.historyDB.GetLastBatchNum()
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}
		if batchNum != 0 {
			err = s.stateDB.Reset(batchNum)
			if err != nil {
				return 0, err
			}
		}

		return block.EthBlockNum, nil
	}

	return 0, ErrNotAbleToSync
}

// Status returns current status values from the Synchronizer
func (s *Synchronizer) Status() (*common.SyncStatus, error) {
	// Avoid possible inconsistencies
	s.mux.Lock()
	defer s.mux.Unlock()

	var status *common.SyncStatus

	// TODO: Join all queries to the DB into a single transaction so that
	// we can remove the mutex locking here:
	// - HistoryDB.GetLastBlock
	// - HistoryDB.GetLastBatchNum
	// - HistoryDB.GetCurrentForgerAddr
	// - HistoryDB.GetNextForgerAddr

	// Get latest block in History DB
	lastSavedBlock, err := s.historyDB.GetLastBlock()
	if err != nil {
		return nil, err
	}
	status.CurrentBlock = lastSavedBlock.EthBlockNum

	// Get latest batch in History DB
	lastSavedBatch, err := s.historyDB.GetLastBatchNum()
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	status.CurrentBatch = lastSavedBatch

	// Get latest blockNum in blockchain
	latestBlockNum, err := s.ethClient.EthCurrentBlock()
	if err != nil {
		return nil, err
	}

	// TODO: Get CurrentForgerAddr & NextForgerAddr from the Auction SC / Or from the HistoryDB

	// Check if Synchronizer is synchronized
	status.Synchronized = status.CurrentBlock == latestBlockNum
	return status, nil
}

// rollupSync gets information from the Rollup Contract
func (s *Synchronizer) rollupSync(blockNum int64) (*rollupData, error) {
	var rollupData = newRollupData()
	// var forgeL1TxsNum int64
	var numAccounts int

	// Get rollup events in the block
	rollupEvents, _, err := s.ethClient.RollupEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}

	// TODO: Replace GetLastL1TxsNum by GetNextL1TxsNum
	nextForgeL1TxsNum := int64(0)
	nextForgeL1TxsNumPtr, err := s.historyDB.GetLastL1TxsNum()
	if err != nil {
		return nil, err
	}
	if nextForgeL1TxsNumPtr != nil {
		nextForgeL1TxsNum = *nextForgeL1TxsNumPtr + 1
	}

	// Get newLastIdx that will be used to complete the accounts
	// idx, err := s.getIdx(rollupEvents)
	// if err != nil {
	// 	return nil, err
	// }

	// Get L1UserTX
	rollupData.l1Txs, err = getL1UserTx(rollupEvents.L1UserTx, blockNum)
	if err != nil {
		return nil, err
	}

	// Get ForgeBatch events to get the L1CoordinatorTxs
	for _, fbEvent := range rollupEvents.ForgeBatch {
		batchData := NewBatchData()
		position := 0

		// Get the input for each Tx
		forgeBatchArgs, err := s.ethClient.RollupForgeBatchArgs(fbEvent.EthTxHash)
		if err != nil {
			return nil, err
		}
		forgeL1TxsNum := int64(0)
		// Check if this is a L1Batch to get L1 Tx from it
		if forgeBatchArgs.L1Batch {
			forgeL1TxsNum = nextForgeL1TxsNum

			// Get L1 User Txs from History DB
			// TODO: Get L1TX from HistoryDB filtered by toforgeL1txNum & fromidx = 0 and
			// update batch number and add accounts to createdAccounts updating idx

			// l1UserTxs, err := s.historyDB.GetL1UserTxs(nextForgeL1TxsNum)
			// If HistoryDB doesn't have L1UserTxs at
			// nextForgeL1TxsNum, check if they exist in
			// rollupData.l1Txs.  This could happen because in a
			// block there could be multiple batches with L1Batch =
			// true (although it's a very rare case).  If the
			// L1UserTxs are not in rollupData.l1Txs, use an empty
			// array (this happens when the L1UserTxs queue is
			// frozen but didn't store any tx).
			l1UserTxs := []common.L1Tx{}
			position = len(l1UserTxs)

			// Get L1 Coordinator Txs
			for _, l1CoordinatorTx := range forgeBatchArgs.L1CoordinatorTxs {
				l1CoordinatorTx.Position = position
				l1CoordinatorTx.ToForgeL1TxsNum = nextForgeL1TxsNum
				l1CoordinatorTx.UserOrigin = false
				l1CoordinatorTx.EthBlockNum = blockNum
				bn := new(common.BatchNum)
				*bn = common.BatchNum(fbEvent.BatchNum)
				l1CoordinatorTx.BatchNum = bn
				l1CoordinatorTx, err = common.NewL1Tx(l1CoordinatorTx)
				if err != nil {
					return nil, err
				}

				batchData.l1CoordinatorTxs = append(batchData.l1CoordinatorTxs, l1CoordinatorTx)

				// Check if we have to register an account
				// if l1CoordinatorTx.FromIdx == 0 {
				// 	account := common.Account{
				// 		// TODO: Uncommnent when common.account has IDx
				// 		// IDx:       common.Idx(idx),
				// 		TokenID:   l1CoordinatorTx.TokenID,
				// 		Nonce:     0,
				// 		Balance:   l1CoordinatorTx.LoadAmount,
				// 		PublicKey: l1CoordinatorTx.FromBJJ,
				// 		EthAddr:   l1CoordinatorTx.FromEthAddr,
				// 	}
				// 	idx++
				// 	batchData.createdAccounts = append(batchData.createdAccounts, &account)
				// 	numAccounts++
				// }
				position++
			}
			nextForgeL1TxsNum++
		}

		// Get L2Txs
		poolL2Txs := common.L2TxsToPoolL2Txs(forgeBatchArgs.L2Txs) // TODO: This is a big uggly, find a better way

		// Get exitTree
		// TODO: Get createdAccounts from ProcessTxs()
		// TODO: Get CollectedFees from ProcessTxs()
		// TODO: Pass forgeBatchArgs.FeeIdxCoordinator to ProcessTxs()
		_, exitInfo, err := s.stateDB.ProcessTxs(batchData.l1UserTxs, batchData.l1CoordinatorTxs, poolL2Txs)
		if err != nil {
			return nil, err
		}

		l2Txs := common.PoolL2TxsToL2Txs(poolL2Txs) // TODO: This is a big uggly, find a better way
		batchData.l2Txs = append(batchData.l2Txs, l2Txs...)

		batchData.exitTree = exitInfo

		// Get Batch information
		batch := &common.Batch{
			BatchNum:    common.BatchNum(fbEvent.BatchNum),
			EthBlockNum: blockNum,
			// ForgerAddr: , TODO: Get it from ethClient -> Add ForgerAddr to RollupEventForgeBatch
			// CollectedFees: , TODO: Clarify where to get them if they are still needed
			StateRoot:     common.Hash(forgeBatchArgs.NewStRoot.Bytes()),
			NumAccounts:   numAccounts,
			ExitRoot:      common.Hash(forgeBatchArgs.NewExitRoot.Bytes()),
			ForgeL1TxsNum: forgeL1TxsNum,
			// SlotNum: TODO: Calculate once ethClient provides the info // calculate from blockNum + ethClient Constants
		}
		batchData.batch = batch
		rollupData.batches = append(rollupData.batches, batchData)
	}

	// Get Registered Tokens
	for _, eAddToken := range rollupEvents.AddToken {
		var token *common.Token

		token.TokenID = common.TokenID(eAddToken.TokenID)
		token.EthAddr = eAddToken.Address
		token.EthBlockNum = blockNum

		// TODO: Add external information consulting SC about it using Address
		rollupData.registeredTokens = append(rollupData.registeredTokens, token)
	}

	// TODO: rollupEvents.UpdateForgeL1L2BatchTimeout
	// TODO: rollupEvents.UpdateFeeAddToken
	// TODO: rollupEvents.WithdrawEvent

	// TODO: Emergency Mechanism
	// TODO: Variables
	// TODO: Constants

	return &rollupData, nil
}

// auctionSync gets information from the Auction Contract
func (s *Synchronizer) auctionSync(blockNum int64) (*auctionData, error) {
	var auctionData = newAuctionData()

	// Get auction events in the block
	auctionEvents, _, err := s.ethClient.AuctionEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}

	// Get bids
	for _, eNewBid := range auctionEvents.NewBid {
		bid := &common.Bid{
			SlotNum:     common.SlotNum(eNewBid.Slot),
			BidValue:    eNewBid.BidAmount,
			ForgerAddr:  eNewBid.CoordinatorForger,
			EthBlockNum: blockNum,
		}
		auctionData.bids = append(auctionData.bids, bid)
	}

	// Get Coordinators
	for _, eNewCoordinator := range auctionEvents.NewCoordinator {
		coordinator := &common.Coordinator{
			Forger:       eNewCoordinator.ForgerAddress,
			WithdrawAddr: eNewCoordinator.WithdrawalAddress,
			URL:          eNewCoordinator.CoordinatorURL,
		}
		auctionData.coordinators = append(auctionData.coordinators, coordinator)
	}

	// TODO: NewSlotDeadline
	// TODO: NewClosedAuctionSlots
	// TODO: NewOutbidding
	// TODO: NewDonationAddress
	// TODO: NewBootCoordinator
	// TODO: NewOpenAuctionSlots
	// TODO: NewAllocationRatio
	// TODO: NewForgeAllocated
	// TODO: NewDefaultSlotSetBid
	// TODO: NewForge
	// TODO: HEZClaimed

	// TODO: Think about separating new coordinaors from coordinator updated

	// Get Coordinators from updates
	for _, eCoordinatorUpdated := range auctionEvents.CoordinatorUpdated {
		coordinator := &common.Coordinator{
			Forger:       eCoordinatorUpdated.ForgerAddress,
			WithdrawAddr: eCoordinatorUpdated.WithdrawalAddress,
			URL:          eCoordinatorUpdated.CoordinatorURL,
		}
		auctionData.coordinators = append(auctionData.coordinators, coordinator)
	}

	// TODO: VARS
	// TODO: CONSTANTS

	return auctionData, nil
}

// wdelayerSync gets information from the Withdrawal Delayer Contract
func (s *Synchronizer) wdelayerSync(blockNum int64) (*common.WithdrawalDelayerVars, error) {
	// TODO: VARS
	// TODO: CONSTANTS

	return nil, nil
}

// func (s *Synchronizer) getIdx(rollupEvents *eth.RollupEvents) (int64, error) {
// 	// TODO: FIXME: There will be an error here when `len(rollupEvents.ForgeBatch) == 0`
// 	lastForgeBatch := rollupEvents.ForgeBatch[len(rollupEvents.ForgeBatch)-1]
//
// 	// TODO: RollupForgeBatchArgs is already called in `rollupSync`.
// 	// Ideally it should not need to be called twice for the same batch.
// 	// Get the input for forgeBatch
// 	forgeBatchArgs, err := s.ethClient.RollupForgeBatchArgs(lastForgeBatch.EthTxHash)
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	return forgeBatchArgs.NewLastIdx + 1, nil
// }

func getL1UserTx(l1UserTxEvents []eth.RollupEventL1UserTx, blockNum int64) ([]*common.L1Tx, error) {
	l1Txs := make([]*common.L1Tx, 0)

	for _, eL1UserTx := range l1UserTxEvents {
		// Fill aditional Tx fields
		eL1UserTx.L1Tx.ToForgeL1TxsNum = eL1UserTx.ToForgeL1TxsNum
		eL1UserTx.L1Tx.Position = eL1UserTx.Position
		eL1UserTx.L1Tx.UserOrigin = true
		eL1UserTx.L1Tx.EthBlockNum = blockNum
		nL1Tx, err := common.NewL1Tx(&eL1UserTx.L1Tx)
		if err != nil {
			return nil, err
		}
		eL1UserTx.L1Tx = *nL1Tx

		l1Txs = append(l1Txs, &eL1UserTx.L1Tx)
	}
	return l1Txs, nil
}
