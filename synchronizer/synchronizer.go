package synchronizer

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"

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

// Config is the Synchronizer configuration
type Config struct {
	StartBlockNum    ConfigStartBlockNum
	InitialVariables SCVariables
}

// Synchronizer implements the Synchronizer type
type Synchronizer struct {
	ethClient         eth.ClientInterface
	auctionConstants  common.AuctionConstants
	rollupConstants   common.RollupConstants
	wDelayerConstants common.WDelayerConstants
	historyDB         *historydb.HistoryDB
	stateDB           *statedb.StateDB
	cfg               Config
	startBlockNum     int64
	vars              SCVariables
	// firstSavedBlock  *common.Block
	// mux sync.Mutex
}

// NewSynchronizer creates a new Synchronizer
func NewSynchronizer(ethClient eth.ClientInterface, historyDB *historydb.HistoryDB,
	stateDB *statedb.StateDB, cfg Config) (*Synchronizer, error) {
	auctionConstants, err := ethClient.AuctionConstants()
	if err != nil {
		log.Errorw("NewSynchronizer ethClient.AuctionConstants()", "err", err)
		return nil, err
	}
	rollupConstants, err := ethClient.RollupConstants()
	if err != nil {
		log.Errorw("NewSynchronizer ethClient.RollupConstants()", "err", err)
		return nil, err
	}
	wDelayerConstants, err := ethClient.WDelayerConstants()
	if err != nil {
		log.Errorw("NewSynchronizer ethClient.WDelayerConstants()", "err", err)
		return nil, err
	}

	// Set startBlockNum to the minimum between Auction, Rollup and
	// WDelayer StartBlockNum
	startBlockNum := cfg.StartBlockNum.Auction
	if startBlockNum < cfg.StartBlockNum.Rollup {
		startBlockNum = cfg.StartBlockNum.Rollup
	}
	if startBlockNum < cfg.StartBlockNum.WDelayer {
		startBlockNum = cfg.StartBlockNum.WDelayer
	}
	s := &Synchronizer{
		ethClient:         ethClient,
		auctionConstants:  *auctionConstants,
		rollupConstants:   *rollupConstants,
		wDelayerConstants: *wDelayerConstants,
		historyDB:         historyDB,
		stateDB:           stateDB,
		cfg:               cfg,
		startBlockNum:     startBlockNum,
	}
	return s, s.init()
}

// AuctionConstants returns the AuctionConstants read from the smart contract
func (s *Synchronizer) AuctionConstants() *common.AuctionConstants {
	return &s.auctionConstants
}

// RollupConstants returns the RollupConstants read from the smart contract
func (s *Synchronizer) RollupConstants() *common.RollupConstants {
	return &s.rollupConstants
}

// WDelayerConstants returns the WDelayerConstants read from the smart contract
func (s *Synchronizer) WDelayerConstants() *common.WDelayerConstants {
	return &s.wDelayerConstants
}

func (s *Synchronizer) init() error {
	rollup, auction, wDelayer, err := s.historyDB.GetSCVars()
	if err == sql.ErrNoRows {
		rollup = &s.cfg.InitialVariables.Rollup
		auction = &s.cfg.InitialVariables.Auction
		wDelayer = &s.cfg.InitialVariables.WDelayer
		log.Debug("Setting initial SCVars in HistoryDB")
		if err = s.historyDB.SetInitialSCVars(rollup, auction, wDelayer); err != nil {
			return err
		}
	}
	s.vars.Rollup = *rollup
	s.vars.Auction = *auction
	s.vars.WDelayer = *wDelayer
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
		if err != nil && err != sql.ErrNoRows {
			return nil, nil, err
		}
		// If we don't have any stored block, we must do a full sync
		// starting from the startBlockNum
		if err == sql.ErrNoRows || lastSavedBlock.EthBlockNum == 0 {
			nextBlockNum = s.startBlockNum
			lastSavedBlock = nil
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
	log.Debugf("ethBlock: num: %v, parent: %v, hash: %v", ethBlock.EthBlockNum, ethBlock.ParentHash.String(), ethBlock.Hash.String())

	log.Debugw("Syncing...", "block", nextBlockNum)

	// Check that the obtianed ethBlock.ParentHash == prevEthBlock.Hash; if not, reorg!
	if lastSavedBlock != nil {
		if lastSavedBlock.Hash != ethBlock.ParentHash {
			// Reorg detected
			log.Debugw("Reorg Detected",
				"blockNum", ethBlock.EthBlockNum,
				"block.parent(got)", ethBlock.ParentHash, "parent.hash(exp)", lastSavedBlock.Hash)
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
	wDelayerData, err := s.wdelayerSync(ethBlock)
	if err != nil {
		return nil, nil, err
	}

	for i := range rollupData.Withdrawals {
		withdrawal := &rollupData.Withdrawals[i]
		if !withdrawal.InstantWithdraw {
			wDelayerTransfer, ok := wDelayerData.DepositsByTxHash[withdrawal.TxHash]
			if !ok {
				return nil, nil, fmt.Errorf("WDelayer deposit corresponding to " +
					"non-instant rollup withdrawal not found")
			}
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
	blockNum := uncleBlock.EthBlockNum

	for blockNum >= s.startBlockNum {
		ethBlock, err := s.ethClient.EthBlockByNumber(context.Background(), blockNum)
		if err != nil {
			return 0, err
		}

		block, err := s.historyDB.GetBlock(blockNum)
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
	err := s.historyDB.Reorg(blockNum)
	if err != nil {
		return 0, err
	}

	batchNum, err := s.historyDB.GetLastBatchNum()
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		batchNum = 0
	}
	err = s.stateDB.Reset(batchNum)
	if err != nil {
		return 0, err
	}

	return blockNum, nil
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
func (s *Synchronizer) rollupSync(ethBlock *common.Block) (*common.RollupData, error) {
	blockNum := ethBlock.EthBlockNum
	var rollupData = common.NewRollupData()
	// var forgeL1TxsNum int64

	// Get rollup events in the block, and make sure the block hash matches
	// the expected one.
	rollupEvents, blockHash, err := s.ethClient.RollupEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}
	// No events in this block
	if blockHash == nil {
		return &rollupData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in Rollup events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
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

	// Get L1UserTX
	rollupData.L1UserTxs, err = getL1UserTx(rollupEvents.L1UserTx, blockNum)
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
		var l1UserTxs []common.L1Tx
		// Check if this is a L1Batch to get L1 Tx from it
		if forgeBatchArgs.L1Batch {
			// Get L1UserTxs with toForgeL1TxsNum, which correspond
			// to the L1UserTxs that are forged in this batch, so
			// that stateDB can process them.

			// First try to find them in HistoryDB.
			l1UserTxs, err = s.historyDB.GetL1UserTxs(nextForgeL1TxsNum)
			if err != nil {
				return nil, err
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
				return nil, err
			}

			batchData.L1CoordinatorTxs = append(batchData.L1CoordinatorTxs, *l1Tx)
			position++
			// fmt.Println("DGB l1coordtx")
		}

		// Insert all the txs forged in this batch (l1UserTxs,
		// L1CoordinatorTxs, PoolL2Txs) into stateDB so that they are
		// processed.
		poolL2Txs := common.L2TxsToPoolL2Txs(forgeBatchArgs.L2TxsData) // NOTE: This is a big ugly, find a better way

		// ProcessTxs updates poolL2Txs adding: Nonce (and also TokenID, but we don't use it).
		processTxsOut, err := s.stateDB.ProcessTxs(forgeBatchArgs.FeeIdxCoordinator, l1UserTxs,
			batchData.L1CoordinatorTxs, poolL2Txs)
		if err != nil {
			return nil, err
		}

		// Set batchNum in exits
		for i := range processTxsOut.ExitInfos {
			exit := &processTxsOut.ExitInfos[i]
			exit.BatchNum = batchNum
		}
		batchData.ExitTree = processTxsOut.ExitInfos

		l2Txs, err := common.PoolL2TxsToL2Txs(poolL2Txs) // NOTE: This is a big uggly, find a better way
		if err != nil {
			return nil, err
		}

		for i := range l2Txs {
			tx := &l2Txs[i]
			tx.Position = position
			tx.EthBlockNum = blockNum
			tx.BatchNum = batchNum
			nTx, err := common.NewL2Tx(tx)
			if err != nil {
				return nil, err
			}

			batchData.L2Txs = append(batchData.L2Txs, *nTx)
			position++
		}

		for i := range processTxsOut.CreatedAccounts {
			createdAccount := &processTxsOut.CreatedAccounts[i]
			createdAccount.Nonce = 0
			createdAccount.Balance = big.NewInt(0)
			createdAccount.BatchNum = batchNum
		}
		batchData.CreatedAccounts = processTxsOut.CreatedAccounts

		slotNum := int64(0)
		if ethBlock.EthBlockNum >= s.auctionConstants.GenesisBlockNum {
			slotNum = (ethBlock.EthBlockNum - s.auctionConstants.GenesisBlockNum) /
				int64(s.auctionConstants.BlocksPerSlot)
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

	varsUpdate := false

	for _, evtUpdateForgeL1L2BatchTimeout := range rollupEvents.UpdateForgeL1L2BatchTimeout {
		s.vars.Rollup.ForgeL1L2BatchTimeout = evtUpdateForgeL1L2BatchTimeout.NewForgeL1L2BatchTimeout
		varsUpdate = true
	}

	for _, evtUpdateFeeAddToken := range rollupEvents.UpdateFeeAddToken {
		s.vars.Rollup.FeeAddToken = evtUpdateFeeAddToken.NewFeeAddToken
		varsUpdate = true
	}

	// NOTE: WithdrawDelay update doesn't have event, so we can't track changes

	// NOTE: Buckets update dones't have event, so we can't track changes

	for _, evtWithdraw := range rollupEvents.Withdraw {
		rollupData.Withdrawals = append(rollupData.Withdrawals, common.WithdrawInfo{
			Idx:             common.Idx(evtWithdraw.Idx),
			NumExitRoot:     common.BatchNum(evtWithdraw.NumExitRoot),
			InstantWithdraw: evtWithdraw.InstantWithdraw,
			TxHash:          evtWithdraw.TxHash,
		})
	}

	if varsUpdate {
		s.vars.Rollup.EthBlockNum = blockNum
		rollupData.Vars = &s.vars.Rollup
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
	blockNum := ethBlock.EthBlockNum
	var auctionData = common.NewAuctionData()

	// Get auction events in the block
	auctionEvents, blockHash, err := s.ethClient.AuctionEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}
	// No events in this block
	if blockHash == nil {
		return &auctionData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in Auction events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
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
		auctionData.Bids = append(auctionData.Bids, bid)
	}

	// Get Coordinators
	for _, evtSetCoordinator := range auctionEvents.SetCoordinator {
		coordinator := common.Coordinator{
			Bidder: evtSetCoordinator.BidderAddress,
			Forger: evtSetCoordinator.ForgerAddress,
			URL:    evtSetCoordinator.CoordinatorURL,
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
		varsUpdate = true
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
			return nil, fmt.Errorf("unexpected SlotSet in "+
				"auctionEvents.NewDefaultSlotSetBid: %v", evt.SlotSet)
		}
		s.vars.Auction.DefaultSlotSetBid[evt.SlotSet] = evt.NewInitialMinBid
		varsUpdate = true
	}

	// NOTE: We ignore NewForgeAllocated
	// NOTE: We ignore NewForge because we're already tracking ForgeBatch event from Rollup
	// NOTE: We ignore HEZClaimed

	if varsUpdate {
		s.vars.Auction.EthBlockNum = blockNum
		auctionData.Vars = &s.vars.Auction
	}

	return &auctionData, nil
}

// wdelayerSync gets information from the Withdrawal Delayer Contract
func (s *Synchronizer) wdelayerSync(ethBlock *common.Block) (*common.WDelayerData, error) {
	blockNum := ethBlock.EthBlockNum
	wDelayerData := common.NewWDelayerData()

	// Get wDelayer events in the block
	wDelayerEvents, blockHash, err := s.ethClient.WDelayerEventsByBlock(blockNum)
	if err != nil {
		return nil, err
	}
	// No events in this block
	if blockHash == nil {
		return &wDelayerData, nil
	}
	if *blockHash != ethBlock.Hash {
		log.Errorw("Block hash mismatch in WDelayer events", "expected", ethBlock.Hash.String(),
			"got", blockHash.String())
		return nil, eth.ErrBlockHashMismatchEvent
	}

	for _, evt := range wDelayerEvents.Deposit {
		wDelayerData.Deposits = append(wDelayerData.Deposits, common.WDelayerTransfer{
			Owner:  evt.Owner,
			Token:  evt.Token,
			Amount: evt.Amount,
		})
		wDelayerData.DepositsByTxHash[evt.TxHash] =
			&wDelayerData.Deposits[len(wDelayerData.Deposits)-1]
	}
	for _, evt := range wDelayerEvents.Withdraw {
		wDelayerData.Withdrawals = append(wDelayerData.Withdrawals, common.WDelayerTransfer{
			Owner:  evt.Owner,
			Token:  evt.Token,
			Amount: evt.Amount,
		})
	}

	varsUpdate := false

	// TODO EscapeHatchWithdrawal
	for range wDelayerEvents.EmergencyModeEnabled {
		s.vars.WDelayer.EmergencyMode = true
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewWithdrawalDelay {
		s.vars.WDelayer.WithdrawalDelay = evt.WithdrawalDelay
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewHermezKeeperAddress {
		s.vars.WDelayer.HermezKeeperAddress = evt.NewHermezKeeperAddress
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewWhiteHackGroupAddress {
		s.vars.WDelayer.WhiteHackGroupAddress = evt.NewWhiteHackGroupAddress
		varsUpdate = true
	}
	for _, evt := range wDelayerEvents.NewHermezGovernanceDAOAddress {
		s.vars.WDelayer.HermezGovernanceDAOAddress = evt.NewHermezGovernanceDAOAddress
		varsUpdate = true
	}

	if varsUpdate {
		s.vars.WDelayer.EthBlockNum = blockNum
		wDelayerData.Vars = &s.vars.WDelayer
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
			return nil, err
		}
		l1Txs[i] = *l1Tx
	}
	return l1Txs, nil
}
