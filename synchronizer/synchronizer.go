package synchronizer

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"sync"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
)

const (
	blocksToSync = 20 // TODO: This will be deleted once we can get the firstSavedBlock from the ethClient
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
	l1UserTxs          []*common.L1Tx
	l1CoordinatorTxs   []*common.L1Tx
	l2Txs              []*common.L2Tx
	registeredAccounts []*common.Account
	exitTree           []*common.ExitInfo
	batch              *common.Batch
}

// NewBatchData creates an empty BatchData with the slices initialized.
func NewBatchData() *BatchData {
	return &BatchData{
		l1UserTxs:          make([]*common.L1Tx, 0),
		l1CoordinatorTxs:   make([]*common.L1Tx, 0),
		l2Txs:              make([]*common.L2Tx, 0),
		registeredAccounts: make([]*common.Account, 0),
		exitTree:           make([]*common.ExitInfo, 0),
	}
}

// BlockData contains information about Blocks from the contracts
type BlockData struct {
	block *common.Block
	// Rollup
	l1Txs   []*common.L1Tx
	batches []*BatchData
	// withdrawals      []*common.ExitInfo
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

// Sync updates History and State DB with information from the blockchain
func (s *Synchronizer) Sync() error {
	// Avoid new sync while performing one
	s.mux.Lock()
	defer s.mux.Unlock()

	// TODO: Get this information from ethClient once it's implemented
	// for the moment we will get the latestblock - 20 as firstSavedBlock
	latestBlock, err := s.ethClient.EthBlockByNumber(context.Background(), 0)
	if err != nil {
		return err
	}
	s.firstSavedBlock, err = s.ethClient.EthBlockByNumber(context.Background(), latestBlock.EthBlockNum-blocksToSync)
	if err != nil {
		return err
	}

	// Get lastSavedBlock from History DB
	lastSavedBlock, err := s.historyDB.GetLastBlock()
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Check if we got a block or nil
	// In case of nil we must do a full sync
	if lastSavedBlock == nil || lastSavedBlock.EthBlockNum == 0 {
		lastSavedBlock = s.firstSavedBlock
	} else {
		// Get the latest block we have in History DB from blockchain to detect a reorg
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), lastSavedBlock.EthBlockNum)
		if err != nil {
			return err
		}

		if ethBlock.Hash != lastSavedBlock.Hash {
			// Reorg detected
			log.Debugf("Reorg Detected...")
			err := s.reorg(lastSavedBlock)
			if err != nil {
				return err
			}

			lastSavedBlock, err = s.historyDB.GetLastBlock()
			if err != nil {
				return err
			}
		}
	}

	log.Debugf("Syncing...")

	// Get latest blockNum in blockchain
	latestBlockNum, err := s.ethClient.EthCurrentBlock()
	if err != nil {
		return err
	}

	log.Debugf("Blocks to sync: %v (lastSavedBlock: %v, latestBlock: %v)", latestBlockNum-lastSavedBlock.EthBlockNum, lastSavedBlock.EthBlockNum, latestBlockNum)

	for lastSavedBlock.EthBlockNum < latestBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), lastSavedBlock.EthBlockNum+1)
		if err != nil {
			return err
		}

		// Get data from the rollup contract
		rollupData, err := s.rollupSync(ethBlock)
		if err != nil {
			return err
		}

		// Get data from the auction contract
		auctionData, err := s.auctionSync(ethBlock)
		if err != nil {
			return err
		}

		// Get data from the WithdrawalDelayer contract
		wdelayerData, err := s.wdelayerSync(ethBlock)
		if err != nil {
			return err
		}

		// Group all the block data into the structs to save into HistoryDB
		var blockData BlockData

		blockData.block = ethBlock

		if rollupData != nil {
			blockData.l1Txs = rollupData.l1Txs
			blockData.batches = rollupData.batches
			// blockData.withdrawals = rollupData.withdrawals
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

		// We get the block on every iteration
		lastSavedBlock, err = s.historyDB.GetLastBlock()
		if err != nil {
			return err
		}
	}

	return nil
}

// reorg manages a reorg, updating History and State DB as needed
func (s *Synchronizer) reorg(uncleBlock *common.Block) error {
	var block *common.Block
	blockNum := uncleBlock.EthBlockNum
	found := false

	log.Debugf("Reorg first uncle block: %v", blockNum)

	// Iterate History DB and the blokchain looking for the latest valid block
	for !found && blockNum > s.firstSavedBlock.EthBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), blockNum)
		if err != nil {
			return err
		}

		block, err = s.historyDB.GetBlock(blockNum)
		if err != nil {
			return err
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
			return err
		}

		batchNum, err := s.historyDB.GetLastBatchNum()
		if err != nil && err != sql.ErrNoRows {
			return err
		}
		if batchNum != 0 {
			err = s.stateDB.Reset(batchNum)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return ErrNotAbleToSync
}

// Status returns current status values from the Synchronizer
func (s *Synchronizer) Status() (*common.SyncStatus, error) {
	// Avoid possible inconsistencies
	s.mux.Lock()
	defer s.mux.Unlock()

	var status *common.SyncStatus

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

	// TODO: Get CurrentForgerAddr & NextForgerAddr from the Auction SC

	// Check if Synchronizer is synchronized
	status.Synchronized = status.CurrentBlock == latestBlockNum
	return status, nil
}

// rollupSync gets information from the Rollup Contract
func (s *Synchronizer) rollupSync(block *common.Block) (*rollupData, error) {
	var rollupData = newRollupData()
	var forgeL1TxsNum uint32
	var numAccounts int

	// using GetLastL1TxsNum as GetNextL1TxsNum
	lastStoredForgeL1TxsNum := uint32(0)
	lastStoredForgeL1TxsNumPtr, err := s.historyDB.GetLastL1TxsNum()
	if err != nil {
		return nil, err
	}
	if lastStoredForgeL1TxsNumPtr != nil {
		lastStoredForgeL1TxsNum = *lastStoredForgeL1TxsNumPtr + 1
	}
	// }

	// Get rollup events in the block
	rollupEvents, _, err := s.ethClient.RollupEventsByBlock(block.EthBlockNum)

	if err != nil {
		return nil, err
	}

	// Get newLastIdx that will be used to complete the accounts
	idx, err := s.getIdx(rollupEvents)

	if err != nil {
		return nil, err
	}

	// Get L1UserTX
	rollupData.l1Txs = s.getL1UserTx(rollupEvents.L1UserTx, block)

	// Get ForgeBatch events to get the L1CoordinatorTxs
	for _, fbEvent := range rollupEvents.ForgeBatch {
		batchData := NewBatchData()

		// TODO: Get position from HistoryDB filtering by
		// to_forge_l1_txs_num and batch_num and latest position, then add 1
		position := 1

		// Get the input for each Tx
		forgeBatchArgs, err := s.ethClient.RollupForgeBatchArgs(fbEvent.EthTxHash)

		if err != nil {
			return nil, err
		}

		// Check if this is a L1Bath to get L1 Tx from it
		if forgeBatchArgs.L1Batch {
			// Get L1 User Txs from History DB
			// TODO: Get L1TX from HistoryDB filtered by toforgeL1txNum & fromidx = 0 and
			// update batch number and add accounts to registeredAccounts updating idx

			// l1UserTxs, err := s.historyDB.GetL1UserTxs(lastStoredForgeL1TxsNum)

			// Get L1 Coordinator Txs
			for _, l1CoordinatorTx := range forgeBatchArgs.L1CoordinatorTxs {
				l1CoordinatorTx.Position = position
				l1CoordinatorTx.ToForgeL1TxsNum = uint32(lastStoredForgeL1TxsNum)
				l1CoordinatorTx.TxID = common.TxID(common.Hash([]byte("0x01" + strconv.FormatInt(int64(lastStoredForgeL1TxsNum), 10) + strconv.FormatInt(int64(l1CoordinatorTx.Position), 10) + "00")))
				l1CoordinatorTx.UserOrigin = false
				l1CoordinatorTx.EthBlockNum = block.EthBlockNum
				l1CoordinatorTx.BatchNum = common.BatchNum(fbEvent.BatchNum)

				batchData.l1CoordinatorTxs = append(batchData.l1CoordinatorTxs, l1CoordinatorTx)

				forgeL1TxsNum++

				// Check if we have to register an account
				if l1CoordinatorTx.FromIdx == 0 {
					account := common.Account{
						// TODO: Uncommnent when common.account has IDx
						// IDx:       common.Idx(idx),
						TokenID:   l1CoordinatorTx.TokenID,
						Nonce:     0,
						Balance:   l1CoordinatorTx.LoadAmount,
						PublicKey: l1CoordinatorTx.FromBJJ,
						EthAddr:   l1CoordinatorTx.FromEthAddr,
					}

					idx++

					batchData.registeredAccounts = append(batchData.registeredAccounts, &account)

					numAccounts++
				}

				position++
			}

			lastStoredForgeL1TxsNum++
		}

		// Get L2Txs
		poolL2Txs := common.L2TxsToPoolL2Txs(forgeBatchArgs.L2Txs) // TODO: This is a big uggly, find a better way

		// Get exitTree
		_, exitInfo, err := s.stateDB.ProcessTxs(true, false, batchData.l1UserTxs, batchData.l1CoordinatorTxs, poolL2Txs)
		if err != nil {
			return nil, err
		}

		l2Txs := common.PoolL2TxsToL2Txs(poolL2Txs) // TODO: This is a big uggly, find a better way
		batchData.l2Txs = append(batchData.l2Txs, l2Txs...)

		batchData.exitTree = exitInfo

		// Get Batch information
		batch := &common.Batch{
			BatchNum:    common.BatchNum(fbEvent.BatchNum),
			EthBlockNum: block.EthBlockNum,
			// ForgerAddr: , TODO: Get it from ethClient
			// CollectedFees: , TODO: Clarify where to get them if they are still needed
			StateRoot:     common.Hash(forgeBatchArgs.NewStRoot.Bytes()),
			NumAccounts:   numAccounts,
			ExitRoot:      common.Hash(forgeBatchArgs.NewExitRoot.Bytes()),
			ForgeL1TxsNum: forgeL1TxsNum,
			// SlotNum: TODO: Calculate once ethClient provides the info
		}

		batchData.batch = batch

		rollupData.batches = append(rollupData.batches, batchData)
	}

	// Get Registered Tokens
	for _, eAddToken := range rollupEvents.AddToken {
		var token *common.Token

		token.TokenID = common.TokenID(eAddToken.TokenID)
		token.EthAddr = eAddToken.Address
		token.EthBlockNum = block.EthBlockNum

		// TODO: Add external information consulting SC about it using Address
		rollupData.registeredTokens = append(rollupData.registeredTokens, token)
	}

	// TODO: Emergency Mechanism
	// TODO: Variables
	// TODO: Constants

	return &rollupData, nil
}

// auctionSync gets information from the Auction Contract
func (s *Synchronizer) auctionSync(block *common.Block) (*auctionData, error) {
	var auctionData = newAuctionData()

	// Get auction events in the block
	auctionEvents, _, err := s.ethClient.AuctionEventsByBlock(block.EthBlockNum)

	if err != nil {
		return nil, err
	}

	// Get bids
	for _, eNewBid := range auctionEvents.NewBid {
		bid := &common.Bid{

			SlotNum:     common.SlotNum(eNewBid.Slot),
			BidValue:    eNewBid.BidAmount,
			ForgerAddr:  eNewBid.CoordinatorForger,
			EthBlockNum: block.EthBlockNum,
		}
		auctionData.bids = append(auctionData.bids, bid)
	}

	// Get Coordinators
	for _, eNewCoordinator := range auctionEvents.NewCoordinator {
		coordinator := &common.Coordinator{
			Forger:   eNewCoordinator.ForgerAddress,
			Withdraw: eNewCoordinator.WithdrawalAddress,
			URL:      eNewCoordinator.URL,
		}
		auctionData.coordinators = append(auctionData.coordinators, coordinator)
	}

	// Get Coordinators from updates
	for _, eCoordinatorUpdated := range auctionEvents.CoordinatorUpdated {
		coordinator := &common.Coordinator{
			Forger:   eCoordinatorUpdated.ForgerAddress,
			Withdraw: eCoordinatorUpdated.WithdrawalAddress,
			URL:      eCoordinatorUpdated.URL,
		}
		auctionData.coordinators = append(auctionData.coordinators, coordinator)
	}

	// TODO: VARS
	// TODO: CONSTANTS

	return auctionData, nil
}

// wdelayerSync gets information from the Withdrawal Delayer Contract
func (s *Synchronizer) wdelayerSync(block *common.Block) (*common.WithdrawalDelayerVars, error) {
	// TODO: VARS
	// TODO: CONSTANTS

	return nil, nil
}

func (s *Synchronizer) getIdx(rollupEvents *eth.RollupEvents) (int64, error) {
	lastForgeBatch := rollupEvents.ForgeBatch[len(rollupEvents.ForgeBatch)-1]

	// Get the input for forgeBatch
	forgeBatchArgs, err := s.ethClient.RollupForgeBatchArgs(lastForgeBatch.EthTxHash)

	if err != nil {
		return 0, err
	}

	return forgeBatchArgs.NewLastIdx + 1, nil
}

func (s *Synchronizer) getL1UserTx(l1UserTxEvents []eth.RollupEventL1UserTx, block *common.Block) []*common.L1Tx {
	l1Txs := make([]*common.L1Tx, 0)

	for _, eL1UserTx := range l1UserTxEvents {
		// Fill aditional Tx fields
		eL1UserTx.L1Tx.TxID = common.TxID(common.Hash([]byte("0x00" + strconv.FormatInt(int64(eL1UserTx.ToForgeL1TxsNum), 10) + strconv.FormatInt(int64(eL1UserTx.Position), 10) + "00")))
		eL1UserTx.L1Tx.ToForgeL1TxsNum = uint32(eL1UserTx.ToForgeL1TxsNum)
		eL1UserTx.L1Tx.Position = eL1UserTx.Position
		eL1UserTx.L1Tx.UserOrigin = true
		eL1UserTx.L1Tx.EthBlockNum = block.EthBlockNum
		eL1UserTx.L1Tx.BatchNum = 0

		l1Txs = append(l1Txs, &eL1UserTx.L1Tx)
	}
	return l1Txs
}
