package api

import (
	"net/http"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

// Network define status of the network
type Network struct {
	LastEthBlock  int64              `json:"lastEthereumBlock"`
	LastSyncBlock int64              `json:"lastSynchedBlock"`
	LastBatch     historydb.BatchAPI `json:"lastBatch"`
	CurrentSlot   int64              `json:"currentSlot"`
	NextForgers   []NextForger       `json:"nextForgers"`
}

// NextForger  is a representation of the information of a coordinator and the period will forge
type NextForger struct {
	Coordinator historydb.CoordinatorAPI `json:"coordinator"`
	Period      Period                   `json:"period"`
}

// Period is a representation of a period
type Period struct {
	SlotNum       int64     `json:"slotNum"`
	FromBlock     int64     `json:"fromBlock"`
	ToBlock       int64     `json:"toBlock"`
	FromTimestamp time.Time `json:"fromTimestamp"`
	ToTimestamp   time.Time `json:"toTimestamp"`
}

var bootCoordinator historydb.CoordinatorAPI = historydb.CoordinatorAPI{
	ItemID: 0,
	Bidder: ethCommon.HexToAddress("0x111111111111111111111111111111111111111"),
	Forger: ethCommon.HexToAddress("0x111111111111111111111111111111111111111"),
	URL:    "https://bootCoordinator",
}

func (a *API) getState(c *gin.Context) {
	// TODO: There are no events for the buckets information, so now this information will be 0
	c.JSON(http.StatusOK, a.status)
}

// SC Vars

// SetRollupVariables set Status.Rollup variables
func (a *API) SetRollupVariables(rollupVariables common.RollupVariables) {
	a.status.Rollup = rollupVariables
}

// SetWDelayerVariables set Status.WithdrawalDelayer variables
func (a *API) SetWDelayerVariables(wDelayerVariables common.WDelayerVariables) {
	a.status.WithdrawalDelayer = wDelayerVariables
}

// SetAuctionVariables set Status.Auction variables
func (a *API) SetAuctionVariables(auctionVariables common.AuctionVariables) {
	a.status.Auction = auctionVariables
}

// Network

// UpdateNetworkInfo update Status.Network information
func (a *API) UpdateNetworkInfo(
	lastEthBlock, lastSyncBlock common.Block,
	lastBatchNum common.BatchNum, currentSlot int64,
) error {
	a.status.Network.LastSyncBlock = lastSyncBlock.EthBlockNum
	a.status.Network.LastEthBlock = lastEthBlock.EthBlockNum
	lastBatch, err := a.h.GetBatchAPI(lastBatchNum)
	if err != nil {
		return err
	}
	a.status.Network.LastBatch = *lastBatch
	a.status.Network.CurrentSlot = currentSlot
	lastClosedSlot := currentSlot + int64(a.status.Auction.ClosedAuctionSlots)
	nextForgers, err := a.GetNextForgers(lastSyncBlock, currentSlot, lastClosedSlot)
	if err != nil {
		return err
	}
	a.status.Network.NextForgers = nextForgers
	return nil
}

// GetNextForgers returns next forgers
func (a *API) GetNextForgers(lastBlock common.Block, currentSlot, lastClosedSlot int64) ([]NextForger, error) {
	secondsPerBlock := int64(15) //nolint:gomnd
	// currentSlot and lastClosedSlot included
	limit := uint(lastClosedSlot - currentSlot + 1)
	bids, _, err := a.h.GetBestBidsAPI(&currentSlot, &lastClosedSlot, nil, &limit, "ASC")
	if err != nil {
		return nil, err
	}
	nextForgers := []NextForger{}
	// Create nextForger for each slot
	for i := currentSlot; i <= lastClosedSlot; i++ {
		fromBlock := i*int64(a.cg.AuctionConstants.BlocksPerSlot) + a.cg.AuctionConstants.GenesisBlockNum
		toBlock := (i+1)*int64(a.cg.AuctionConstants.BlocksPerSlot) + a.cg.AuctionConstants.GenesisBlockNum - 1
		nextForger := NextForger{
			Period: Period{
				SlotNum:       i,
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				FromTimestamp: lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(fromBlock-lastBlock.EthBlockNum))),
				ToTimestamp:   lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(toBlock-lastBlock.EthBlockNum))),
			},
		}
		foundBid := false
		// If there is a bid for a slot, get forger (coordinator)
		for j := range bids {
			if bids[j].SlotNum == i {
				foundBid = true
				coordinator, err := a.h.GetCoordinatorAPI(bids[j].Bidder)
				if err != nil {
					return nil, err
				}
				nextForger.Coordinator = *coordinator
				break
			}
		}
		// If there is no bid, the coordinator that will forge is boot coordinator
		if !foundBid {
			nextForger.Coordinator = bootCoordinator
		}
		nextForgers = append(nextForgers, nextForger)
	}
	return nextForgers, nil
}

// Metrics

// UpdateMetrics update Status.Metrics information
func (a *API) UpdateMetrics() error {
	metrics, err := a.h.GetMetrics(a.status.Network.LastBatch.BatchNum)
	if err != nil {
		return err
	}
	a.status.Metrics = *metrics
	return nil
}

// Recommended fee

// UpdateRecommendedFee update Status.RecommendedFee information
func (a *API) UpdateRecommendedFee() error {
	feeExistingAccount, err := a.h.GetAvgTxFee()
	if err != nil {
		return err
	}
	a.status.RecommendedFee.ExistingAccount = feeExistingAccount
	a.status.RecommendedFee.CreatesAccount = createAccountExtraFeePercentage * feeExistingAccount
	a.status.RecommendedFee.CreatesAccountAndRegister = createAccountInternalExtraFeePercentage * feeExistingAccount
	return nil
}
