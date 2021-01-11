package api

import (
	"database/sql"
	"math/big"
	"net/http"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

// Network define status of the network
type Network struct {
	LastEthBlock  int64               `json:"lastEthereumBlock"`
	LastSyncBlock int64               `json:"lastSynchedBlock"`
	LastBatch     *historydb.BatchAPI `json:"lastBatch"`
	CurrentSlot   int64               `json:"currentSlot"`
	NextForgers   []NextForger        `json:"nextForgers"`
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
	a.status.RLock()
	status := a.status //nolint
	a.status.RUnlock()
	c.JSON(http.StatusOK, status) //nolint
}

// SC Vars

// SetRollupVariables set Status.Rollup variables
func (a *API) SetRollupVariables(rollupVariables common.RollupVariables) {
	a.status.Lock()
	a.status.Rollup = rollupVariables
	a.status.Unlock()
}

// SetWDelayerVariables set Status.WithdrawalDelayer variables
func (a *API) SetWDelayerVariables(wDelayerVariables common.WDelayerVariables) {
	a.status.Lock()
	a.status.WithdrawalDelayer = wDelayerVariables
	a.status.Unlock()
}

// SetAuctionVariables set Status.Auction variables
func (a *API) SetAuctionVariables(auctionVariables common.AuctionVariables) {
	a.status.Lock()
	a.status.Auction = auctionVariables
	a.status.Unlock()
}

// Network

// UpdateNetworkInfoBlock update Status.Network block related information
func (a *API) UpdateNetworkInfoBlock(
	lastEthBlock, lastSyncBlock common.Block,
) {
	a.status.Network.LastSyncBlock = lastSyncBlock.Num
	a.status.Network.LastEthBlock = lastEthBlock.Num
}

// UpdateNetworkInfo update Status.Network information
func (a *API) UpdateNetworkInfo(
	lastEthBlock, lastSyncBlock common.Block,
	lastBatchNum common.BatchNum, currentSlot int64,
) error {
	lastBatch, err := a.h.GetBatchAPI(lastBatchNum)
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		lastBatch = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}
	lastClosedSlot := currentSlot + int64(a.status.Auction.ClosedAuctionSlots)
	nextForgers, err := a.getNextForgers(lastSyncBlock, currentSlot, lastClosedSlot)
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		nextForgers = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}
	a.status.Lock()
	a.status.Network.LastSyncBlock = lastSyncBlock.Num
	a.status.Network.LastEthBlock = lastEthBlock.Num
	a.status.Network.LastBatch = lastBatch
	a.status.Network.CurrentSlot = currentSlot
	a.status.Network.NextForgers = nextForgers

	// Update buckets withdrawals
	bucketsUpdate, err := a.h.GetBucketUpdates()
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		bucketsUpdate = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}

	for i, bucketParams := range a.status.Rollup.Buckets {
		for _, bucketUpdate := range bucketsUpdate {
			if bucketUpdate.NumBucket == i {
				bucketParams.Withdrawals = bucketUpdate.Withdrawals
				a.status.Rollup.Buckets[i] = bucketParams
				break
			}
		}
	}
	a.status.Unlock()
	return nil
}

// getNextForgers returns next forgers
func (a *API) getNextForgers(lastBlock common.Block, currentSlot, lastClosedSlot int64) ([]NextForger, error) {
	secondsPerBlock := int64(15) //nolint:gomnd
	// currentSlot and lastClosedSlot included
	limit := uint(lastClosedSlot - currentSlot + 1)
	bids, _, err := a.h.GetBestBidsAPI(&currentSlot, &lastClosedSlot, nil, &limit, "ASC")
	if err != nil && tracerr.Unwrap(err) != sql.ErrNoRows {
		return nil, tracerr.Wrap(err)
	}
	nextForgers := []NextForger{}
	// Get min bid info
	var minBidInfo []historydb.MinBidInfo
	if currentSlot >= a.status.Auction.DefaultSlotSetBidSlotNum {
		// All min bids can be calculated with the last update of AuctionVariables
		minBidInfo = []historydb.MinBidInfo{{
			DefaultSlotSetBid:        a.status.Auction.DefaultSlotSetBid,
			DefaultSlotSetBidSlotNum: a.status.Auction.DefaultSlotSetBidSlotNum,
		}}
	} else {
		// Get all the relevant updates from the DB
		minBidInfo, err = a.h.GetAuctionVarsUntilSetSlotNum(lastClosedSlot, int(lastClosedSlot-currentSlot)+1)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	// Create nextForger for each slot
	for i := currentSlot; i <= lastClosedSlot; i++ {
		fromBlock := i*int64(a.cg.AuctionConstants.BlocksPerSlot) + a.cg.AuctionConstants.GenesisBlockNum
		toBlock := (i+1)*int64(a.cg.AuctionConstants.BlocksPerSlot) + a.cg.AuctionConstants.GenesisBlockNum - 1
		nextForger := NextForger{
			Period: Period{
				SlotNum:       i,
				FromBlock:     fromBlock,
				ToBlock:       toBlock,
				FromTimestamp: lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(fromBlock-lastBlock.Num))),
				ToTimestamp:   lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(toBlock-lastBlock.Num))),
			},
		}
		foundForger := false
		// If there is a bid for a slot, get forger (coordinator)
		for j := range bids {
			slotNum := bids[j].SlotNum
			if slotNum == i {
				// There's a bid for the slot
				// Check if the bid is greater than the minimum required
				for i := 0; i < len(minBidInfo); i++ {
					// Find the most recent update
					if slotNum >= minBidInfo[i].DefaultSlotSetBidSlotNum {
						// Get min bid
						minBidSelector := slotNum % int64(len(a.status.Auction.DefaultSlotSetBid))
						minBid := minBidInfo[i].DefaultSlotSetBid[minBidSelector]
						// Check if the bid has beaten the minimum
						bid, ok := new(big.Int).SetString(string(bids[j].BidValue), 10)
						if !ok {
							return nil, tracerr.New("Wrong bid value, error parsing it as big.Int")
						}
						if minBid.Cmp(bid) == 1 {
							// Min bid is greater than bid, the slot will be forged by boot coordinator
							break
						}
						foundForger = true
						break
					}
				}
				if !foundForger { // There is no bid or it's smaller than the minimum
					break
				}
				coordinator, err := a.h.GetCoordinatorAPI(bids[j].Bidder)
				if err != nil {
					return nil, tracerr.Wrap(err)
				}
				nextForger.Coordinator = *coordinator
				break
			}
		}
		// If there is no bid, the coordinator that will forge is boot coordinator
		if !foundForger {
			nextForger.Coordinator = bootCoordinator
		}
		nextForgers = append(nextForgers, nextForger)
	}
	return nextForgers, nil
}

// Metrics

// UpdateMetrics update Status.Metrics information
func (a *API) UpdateMetrics() error {
	a.status.RLock()
	if a.status.Network.LastBatch == nil {
		a.status.RUnlock()
		return nil
	}
	batchNum := a.status.Network.LastBatch.BatchNum
	a.status.RUnlock()
	metrics, err := a.h.GetMetrics(batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	a.status.Lock()
	a.status.Metrics = *metrics
	a.status.Unlock()
	return nil
}

// Recommended fee

// UpdateRecommendedFee update Status.RecommendedFee information
func (a *API) UpdateRecommendedFee() error {
	feeExistingAccount, err := a.h.GetAvgTxFee()
	if err != nil {
		return tracerr.Wrap(err)
	}
	a.status.Lock()
	a.status.RecommendedFee.ExistingAccount = feeExistingAccount
	a.status.RecommendedFee.CreatesAccount = createAccountExtraFeePercentage * feeExistingAccount
	a.status.RecommendedFee.CreatesAccountAndRegister = createAccountInternalExtraFeePercentage * feeExistingAccount
	a.status.Unlock()
	return nil
}
