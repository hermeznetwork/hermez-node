package synchronizer

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
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

// rollupData contains information returned by the Rollup SC
type rollupData struct {
	l1UserTxs []common.L1Tx
	batches   []common.BatchData
	// withdrawals      []*common.ExitInfo
	addTokens []common.Token
	vars      *common.RollupVars
}

// NewRollupData creates an empty rollupData with the slices initialized.
func newRollupData() rollupData {
	return rollupData{
		l1UserTxs: make([]common.L1Tx, 0),
		batches:   make([]common.BatchData, 0),
		// withdrawals:      make([]*common.ExitInfo, 0),
		addTokens: make([]common.Token, 0),
	}
}

// auctionData contains information returned by the Action SC
type auctionData struct {
	bids         []common.Bid
	coordinators []common.Coordinator
	vars         *common.AuctionVars
}

// newAuctionData creates an empty auctionData with the slices initialized.
func newAuctionData() *auctionData {
	return &auctionData{
		bids:         make([]common.Bid, 0),
		coordinators: make([]common.Coordinator, 0),
	}
}

type wdelayerData struct {
	vars *common.WithdrawDelayerVars
}

// Synchronizer implements the Synchronizer type
type Synchronizer struct {
	ethClient        eth.ClientInterface
	auctionConstants eth.AuctionConstants
	historyDB        *historydb.HistoryDB
	stateDB          *statedb.StateDB
	// firstSavedBlock  *common.Block
	// mux sync.Mutex
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(ethClient eth.ClientInterface, historyDB *historydb.HistoryDB, stateDB *statedb.StateDB) (*Synchronizer, error) {
	auctionConstants, err := ethClient.AuctionConstants()
	if err != nil {
		log.Errorw("NewSynchronizer", "err", err)
		return nil, err
	}
	return &Synchronizer{
		ethClient:        ethClient,
		auctionConstants: *auctionConstants,
		historyDB:        historyDB,
		stateDB:          stateDB,
	}, nil
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
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}
		// If we don't have any stored block, we must do a full sync starting from the rollup genesis block
		if err == sql.ErrNoRows {
			nextBlockNum = s.auctionConstants.GenesisBlockNum
		}
	}
	if lastSavedBlock != nil {
		nextBlockNum = lastSavedBlock.EthBlockNum + 1
	}

	ethBlock, err := s.ethClient.EthBlockByNumber(ctx, nextBlockNum)
	if err == ethereum.NotFound {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}

	log.Debugw("Syncing...", "block", nextBlockNum)

	// Check that the obtianed ethBlock.ParentHash == prevEthBlock.Hash; if not, reorg!
	if lastSavedBlock != nil {
		if lastSavedBlock.Hash != ethBlock.ParentHash {
			// Reorg detected
			log.Debugw("Reorg Detected",
				"blockNum", ethBlock.EthBlockNum,
				"block.parent", ethBlock.ParentHash, "parent.hash", lastSavedBlock.Hash)
			lastDBBlockNum, err := s.reorg(lastSavedBlock)
			if err != nil {
				return nil, nil, err
			}
			discarded := lastSavedBlock.EthBlockNum - lastDBBlockNum
			return nil, &discarded, nil
		}
	}

	// Get data from the rollup contract
	rollupData, err := s.rollupSync(ethBlock)
	if err != nil {
		return nil, nil, err
	}

	// Get data from the auction contract
	auctionData, err := s.auctionSync(ethBlock)
	if err != nil {
		return nil, nil, err
	}

	// Get data from the WithdrawalDelayer contract
	wdelayerData, err := s.wdelayerSync(ethBlock)
	if err != nil {
		return nil, nil, err
	}

	// Group all the block data into the structs to save into HistoryDB
	var blockData common.BlockData

	blockData.Block = *ethBlock

	blockData.L1UserTxs = rollupData.l1UserTxs
	blockData.Batches = rollupData.batches
	// blockData.withdrawals = rollupData.withdrawals // TODO
	blockData.AddedTokens = rollupData.addTokens
	blockData.RollupVars = rollupData.vars

	blockData.Bids = auctionData.bids
	blockData.Coordinators = auctionData.coordinators
	blockData.AuctionVars = auctionData.vars

	blockData.WithdrawDelayerVars = wdelayerData.vars

	// log.Debugw("Sync()", "block", blockData)
	// err = s.historyDB.AddBlock(blockData.Block)
	// if err != nil {
	// 	return err
	// }
	err = s.historyDB.AddBlockSCData(&blockData)
	if err != nil {
		return nil, nil, err
	}

	return &blockData, nil, nil
}

// reorg manages a reorg, updating History and State DB as needed.  Keeps
// checking previous blocks from the HistoryDB against the blockchain until a
// block hash match is found.  All future blocks in the HistoryDB and
// corresponding batches in StateBD are discarded.  Returns the last valid
// blockNum from the HistoryDB.
func (s *Synchronizer) reorg(uncleBlock *common.Block) (int64, error) {
	var block *common.Block
	blockNum := uncleBlock.EthBlockNum

	for blockNum >= s.auctionConstants.GenesisBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), blockNum)
		if err != nil {
			return 0, err
		}

		block, err = s.historyDB.GetBlock(blockNum)
		if err != nil {
			return 0, err
		}
		if block.Hash == ethBlock.Hash {
			log.Debugf("Found valid block: %v", blockNum)
			break
		}
		blockNum--
	}
	total := uncleBlock.EthBlockNum - blockNum
	log.Debugw("Discarding blocks", "total", total, "from", uncleBlock.EthBlockNum, "to", blockNum+1)

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

// TODO: Figure out who will use the Status output, and only return what's strictly need
/*
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
*/

// rollupSync retreives all the Rollup Smart Contract Data that happened at
// ethBlock.blockNum with ethBlock.Hash.
func (s *Synchronizer) rollupSync(ethBlock *common.Block) (*rollupData, error) {
	blockNum := ethBlock.EthBlockNum
	var rollupData = newRollupData()
	// var forgeL1TxsNum int64
	var numAccounts int

	// Get rollup events in the block, and make sure the block hash matches
	// the expected one.
	rollupEvents, blockHash, err := s.ethClient.RollupEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}
	if *blockHash != ethBlock.Hash {
		return nil, eth.ErrBlockHashMismatchEvent
	}

	var nextForgeL1TxsNum int64 // forgeL1TxsNum for the next L1Batch
	nextForgeL1TxsNumPtr, err := s.historyDB.GetLastL1TxsNum()
	if err != nil {
		return nil, err
	}
	if nextForgeL1TxsNumPtr != nil {
		nextForgeL1TxsNum = *nextForgeL1TxsNumPtr + 1
	} else {
		nextForgeL1TxsNum = 0
	}

	// Get newLastIdx that will be used to complete the accounts
	// idx, err := s.getIdx(rollupEvents)
	// if err != nil {
	// 	return nil, err
	// }

	// Get L1UserTX
	rollupData.l1UserTxs, err = getL1UserTx(rollupEvents.L1UserTx, blockNum)
	if err != nil {
		return nil, err
	}

	// Get ForgeBatch events to get the L1CoordinatorTxs
	for _, evtForgeBatch := range rollupEvents.ForgeBatch {
		batchData := common.NewBatchData()
		position := 0

		// Get the input for each Tx
		forgeBatchArgs, sender, err := s.ethClient.RollupForgeBatchArgs(evtForgeBatch.EthTxHash)
		if err != nil {
			return nil, err
		}

		batchNum := common.BatchNum(evtForgeBatch.BatchNum)
		nextForgeL1TxsNumCpy := nextForgeL1TxsNum
		var l1UserTxs []common.L1Tx
		// Check if this is a L1Batch to get L1 Tx from it
		if forgeBatchArgs.L1Batch {
			// Get L1UserTxs with toForgeL1TxsNum, which correspond
			// to the L1UserTxs that are forged in this batch, so
			// that stateDB can process them.

			// First try to find them in HistoryDB.
			l1UserTxs, err := s.historyDB.GetL1UserTxs(nextForgeL1TxsNumCpy)
			if len(l1UserTxs) == 0 {
				// If not found in the DB, try to find them in
				// this block.  This could happen because in a
				// block there could be multiple batches with
				// L1Batch = true (although it's a very rare
				// case).
				// If not found in the DB and the block doesn't
				// contain the l1UserTxs, it means that the
				// L1UserTxs queue with toForgeL1TxsNum was
				// closed empty, so we leave `l1UserTxs` as an
				// empty slice.
				for _, l1UserTx := range rollupData.l1UserTxs {
					if *l1UserTx.ToForgeL1TxsNum == nextForgeL1TxsNumCpy {
						l1UserTxs = append(l1UserTxs, l1UserTx)
					}
				}
			}
			if err != nil {
				return nil, err
			}
			nextForgeL1TxsNum++

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
				return nil, err
			}

			batchData.L1CoordinatorTxs = append(batchData.L1CoordinatorTxs, *l1Tx)
			position++
			fmt.Println("DGB l1coordtx")
		}

		// Insert all the txs forged in this batch (l1UserTxs,
		// L1CoordinatorTxs, PoolL2Txs) into stateDB so that they are
		// processed.
		poolL2Txs := common.L2TxsToPoolL2Txs(forgeBatchArgs.L2TxsData) // TODO: This is a big ugly, find a better way

		// TODO: Get createdAccounts from ProcessTxs()
		// TODO: Get CollectedFees from ProcessTxs()
		// TODO: Pass forgeBatchArgs.FeeIdxCoordinator to ProcessTxs()
		// ProcessTxs updates poolL2Txs adding: Nonce, TokenID
		_, exitInfo, err := s.stateDB.ProcessTxs(l1UserTxs, batchData.L1CoordinatorTxs, poolL2Txs)
		if err != nil {
			return nil, err
		}

		l2Txs, err := common.PoolL2TxsToL2Txs(poolL2Txs) // TODO: This is a big uggly, find a better way
		if err != nil {
			return nil, err
		}

		for i := range l2Txs {
			_l2Tx := l2Txs[i]
			_l2Tx.Position = position
			_l2Tx.EthBlockNum = blockNum
			_l2Tx.BatchNum = batchNum
			l2Tx, err := common.NewL2Tx(&_l2Tx)
			if err != nil {
				return nil, err
			}

			batchData.L2Txs = append(batchData.L2Txs, *l2Tx)
			position++
		}

		batchData.ExitTree = exitInfo

		slotNum := int64(0)
		if ethBlock.EthBlockNum >= s.auctionConstants.GenesisBlockNum {
			slotNum = (ethBlock.EthBlockNum - s.auctionConstants.GenesisBlockNum) /
				int64(s.auctionConstants.BlocksPerSlot)
		}

		// Get Batch information
		batch := common.Batch{
			BatchNum:    batchNum,
			EthBlockNum: blockNum,
			ForgerAddr:  *sender,
			// CollectedFees: , TODO: Clarify where to get them if they are still needed
			StateRoot:   forgeBatchArgs.NewStRoot,
			NumAccounts: numAccounts, // TODO: Calculate this value
			LastIdx:     forgeBatchArgs.NewLastIdx,
			ExitRoot:    forgeBatchArgs.NewExitRoot,
			SlotNum:     slotNum,
		}
		if forgeBatchArgs.L1Batch {
			batch.ForgeL1TxsNum = &nextForgeL1TxsNumCpy
			batchData.L1Batch = true
		}
		batchData.Batch = batch
		rollupData.batches = append(rollupData.batches, *batchData)
	}

	// Get Registered Tokens
	for _, evtAddToken := range rollupEvents.AddToken {
		var token common.Token

		token.TokenID = common.TokenID(evtAddToken.TokenID)
		token.EthAddr = evtAddToken.TokenAddress
		token.EthBlockNum = blockNum

		if consts, err := s.ethClient.EthERC20Consts(evtAddToken.TokenAddress); err != nil {
			log.Warnw("Error retreiving ERC20 token constants", "addr", evtAddToken.TokenAddress)
			// TODO: Add external information consulting SC about it using Address
			token.Name = "ERC20_ETH_ERROR"
			token.Symbol = "ERROR"
			token.Decimals = 1
		} else {
			token.Name = cutStringMax(consts.Name, 20)
			token.Symbol = cutStringMax(consts.Symbol, 10)
			token.Decimals = consts.Decimals
		}

		rollupData.addTokens = append(rollupData.addTokens, token)
	}

	// TODO: rollupEvents.UpdateForgeL1L2BatchTimeout
	// TODO: rollupEvents.UpdateFeeAddToken
	// TODO: rollupEvents.WithdrawEvent

	// TODO: Emergency Mechanism
	// TODO: Variables
	// TODO: Constants

	return &rollupData, nil
}

func cutStringMax(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// auctionSync gets information from the Auction Contract
func (s *Synchronizer) auctionSync(ethBlock *common.Block) (*auctionData, error) {
	blockNum := ethBlock.EthBlockNum
	var auctionData = newAuctionData()

	// Get auction events in the block
	auctionEvents, blockHash, err := s.ethClient.AuctionEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}
	if *blockHash != ethBlock.Hash {
		return nil, eth.ErrBlockHashMismatchEvent
	}

	// Get bids
	for _, evtNewBid := range auctionEvents.NewBid {
		bid := common.Bid{
			SlotNum:     evtNewBid.Slot,
			BidValue:    evtNewBid.BidAmount,
			Bidder:      evtNewBid.Bidder,
			EthBlockNum: blockNum,
		}
		auctionData.bids = append(auctionData.bids, bid)
	}

	// Get Coordinators
	for _, evtSetCoordinator := range auctionEvents.SetCoordinator {
		coordinator := common.Coordinator{
			Bidder: evtSetCoordinator.BidderAddress,
			Forger: evtSetCoordinator.ForgerAddress,
			URL:    evtSetCoordinator.CoordinatorURL,
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

	// TODO: VARS
	// TODO: CONSTANTS

	return auctionData, nil
}

// wdelayerSync gets information from the Withdrawal Delayer Contract
func (s *Synchronizer) wdelayerSync(ethBlock *common.Block) (*wdelayerData, error) {
	// blockNum := ethBlock.EthBlockNum
	// TODO: VARS
	// TODO: CONSTANTS

	return &wdelayerData{
		vars: nil,
	}, nil
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

func getL1UserTx(eventsL1UserTx []eth.RollupEventL1UserTx, blockNum int64) ([]common.L1Tx, error) {
	l1Txs := make([]common.L1Tx, len(eventsL1UserTx))
	for i := range eventsL1UserTx {
		eventsL1UserTx[i].L1UserTx.EthBlockNum = blockNum
		// Check validity of L1UserTx
		l1Tx, err := common.NewL1Tx(&eventsL1UserTx[i].L1UserTx)
		if err != nil {
			return nil, err
		}
		l1Txs[i] = *l1Tx
	}
	return l1Txs, nil
}
