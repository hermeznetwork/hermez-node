package synchronizer

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"sync"

	ethCommon "github.com/ethereum/go-ethereum/common"
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

// BatchData contains information about Batches from the contracts
type BatchData struct {
	// l1txs              []common.L1Tx
	// l2txs              []common.L2Tx
	// registeredAccounts []common.Account
	// exitTree           []common.ExitTreeLeaf
}

// BlockData contains information about Blocks from the contracts
type BlockData struct {
	// block *common.Block
	// Rollup
	// batches          []BatchData
	// withdrawals      []common.ExitTreeLeaf
	// registeredTokens []common.Token
	// rollupVars       *common.RollupVars
	// Auction
	// bids         []common.Bid
	// coordinators []common.Coordinator
	// auctionVars  *common.AuctionVars
}

// Status is returned by the Status method
type Status struct {
	CurrentBlock      uint64
	CurrentBatch      common.BatchNum
	CurrentForgerAddr ethCommon.Address
	NextForgerAddr    ethCommon.Address
	Synchronized      bool
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

	var lastStoredForgeL1TxsNum uint64

	// TODO: Get this information from ethClient once it's implemented
	// for the moment we will get the latestblock - 20 as firstSavedBlock
	latestBlock, err := s.ethClient.BlockByNumber(context.Background(), nil)
	if err != nil {
		return err
	}
	s.firstSavedBlock, err = s.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(latestBlock.EthBlockNum-blocksToSync)))
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

		ethBlock, err := s.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(lastSavedBlock.EthBlockNum)))
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
	latestBlockNum, err := s.ethClient.CurrentBlock()

	if err != nil {
		return err
	}

	log.Debugf("lastSavedBlock: %v", latestBlockNum)
	log.Debugf("latestBlock: %v", latestBlockNum)
	log.Debugf("Blocks to sync: %v", latestBlockNum.Uint64()-lastSavedBlock.EthBlockNum)

	for lastSavedBlock.EthBlockNum < latestBlockNum.Uint64() {
		ethBlock, err := s.ethClient.BlockByNumber(context.Background(), big.NewInt(int64(lastSavedBlock.EthBlockNum+1)))
		if err != nil {
			return err
		}

		// Get data from the rollup contract
		blockData, batchData, err := s.rollupSync(ethBlock, lastStoredForgeL1TxsNum)

		if err != nil {
			return err
		}

		// Get data from the auction contract
		err = s.auctionSync(blockData, batchData)

		if err != nil {
			return err
		}

		// Add rollupData and auctionData once the method is updated
		err = s.historyDB.AddBlock(ethBlock)

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
	// Iterate History DB and the blokchain looking for the latest valid block
	var block *common.Block
	blockNum := uncleBlock.EthBlockNum
	found := false

	log.Debugf("Reorg first uncle block: %v", blockNum)

	for !found && blockNum > s.firstSavedBlock.EthBlockNum {
		header, err := s.ethClient.HeaderByNumber(context.Background(), big.NewInt(int64(blockNum)))
		if err != nil {
			return err
		}

		block, err = s.historyDB.GetBlock(blockNum)

		if err != nil {
			return err
		}

		if block.Hash == header.Hash() {
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
func (s *Synchronizer) Status() (*Status, error) {
	// Avoid possible incosistencies
	s.mux.Lock()
	defer s.mux.Unlock()

	var status *Status

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
	latestBlockNum, err := s.ethClient.CurrentBlock()

	if err != nil {
		return nil, err
	}

	// TODO: Get CurrentForgerAddr & NextForgerAddr

	// Check if Synchronizer is synchronized
	status.Synchronized = status.CurrentBlock == latestBlockNum.Uint64()

	return status, nil
}

// rollupSync gets information from the Rollup Contract
func (s *Synchronizer) rollupSync(block *common.Block, lastStoredForgeL1TxsNum uint64) (*BlockData, []*BatchData, error) {
	// To be implemented
	return nil, nil, nil
}

// auctionSync gets information from the Auction Contract
func (s *Synchronizer) auctionSync(blockData *BlockData, batchData []*BatchData) error {
	// To be implemented
	return nil
}
