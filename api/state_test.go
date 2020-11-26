package api

import (
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/stretchr/testify/assert"
)

const secondsPerBlock = 15

type testStatus struct {
	Network           testNetwork              `json:"network"`
	Metrics           historydb.Metrics        `json:"metrics"`
	Rollup            common.RollupVariables   `json:"rollup"`
	Auction           common.AuctionVariables  `json:"auction"`
	WithdrawalDelayer common.WDelayerVariables `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee    `json:"recommendedFee"`
}

type testNetwork struct {
	LastEthBlock  int64        `json:"lastEthereumBlock"`
	LastSyncBlock int64        `json:"lastSynchedBlock"`
	LastBatch     testBatch    `json:"lastBatch"`
	CurrentSlot   int64        `json:"currentSlot"`
	NextForgers   []NextForger `json:"nextForgers"`
}

func TestSetRollupVariables(t *testing.T) {
	rollupVars := &common.RollupVariables{}
	assert.Equal(t, *rollupVars, api.status.Rollup)
	api.SetRollupVariables(tc.rollupVars)
	assert.Equal(t, tc.rollupVars, api.status.Rollup)
}

func TestSetWDelayerVariables(t *testing.T) {
	wdelayerVars := &common.WDelayerVariables{}
	assert.Equal(t, *wdelayerVars, api.status.WithdrawalDelayer)
	api.SetWDelayerVariables(tc.wdelayerVars)
	assert.Equal(t, tc.wdelayerVars, api.status.WithdrawalDelayer)
}

func TestSetAuctionVariables(t *testing.T) {
	auctionVars := &common.AuctionVariables{}
	assert.Equal(t, *auctionVars, api.status.Auction)
	api.SetAuctionVariables(tc.auctionVars)
	assert.Equal(t, tc.auctionVars, api.status.Auction)
}

func TestNextForgers(t *testing.T) {
	// It's assumed that bids for each slot will be received in increasing order
	bestBids := make(map[int64]testBid)
	for j := range tc.bids {
		bestBids[tc.bids[j].SlotNum] = tc.bids[j]
	}
	lastBlock := tc.blocks[len(tc.blocks)-1]
	for i := int64(0); i < tc.slots[len(tc.slots)-1].SlotNum; i++ {
		lastClosedSlot := i + int64(api.status.Auction.ClosedAuctionSlots)
		nextForgers, err := api.GetNextForgers(tc.blocks[len(tc.blocks)-1], i, lastClosedSlot)
		assert.NoError(t, err)
		for j := i; j <= lastClosedSlot; j++ {
			for q := range nextForgers {
				if nextForgers[q].Period.SlotNum == j {
					if nextForgers[q].Coordinator.ItemID != 0 {
						assert.Equal(t, bestBids[j].Bidder, nextForgers[q].Coordinator.Bidder)
					} else {
						assert.Equal(t, bootCoordinator.Bidder, nextForgers[q].Coordinator.Bidder)
					}
					firstBlockSlot, lastBlockSlot := api.getFirstLastBlock(j)
					fromTimestamp := lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(firstBlockSlot-lastBlock.Num)))
					toTimestamp := lastBlock.Timestamp.Add(time.Second * time.Duration(secondsPerBlock*(lastBlockSlot-lastBlock.Num)))
					assert.Equal(t, fromTimestamp.Unix(), nextForgers[q].Period.FromTimestamp.Unix())
					assert.Equal(t, toTimestamp.Unix(), nextForgers[q].Period.ToTimestamp.Unix())
				}
			}
		}
	}
}

func TestUpdateNetworkInfo(t *testing.T) {
	status := &Network{}
	assert.Equal(t, status.LastSyncBlock, api.status.Network.LastSyncBlock)
	assert.Equal(t, status.LastBatch.BatchNum, api.status.Network.LastBatch.BatchNum)
	assert.Equal(t, status.CurrentSlot, api.status.Network.CurrentSlot)
	assert.Equal(t, status.NextForgers, api.status.Network.NextForgers)
	lastBlock := tc.blocks[3]
	lastBatchNum := common.BatchNum(3)
	currentSlotNum := int64(1)
	err := api.UpdateNetworkInfo(lastBlock, lastBlock, lastBatchNum, currentSlotNum)
	assert.NoError(t, err)
	assert.Equal(t, lastBlock.Num, api.status.Network.LastSyncBlock)
	assert.Equal(t, lastBatchNum, api.status.Network.LastBatch.BatchNum)
	assert.Equal(t, currentSlotNum, api.status.Network.CurrentSlot)
	assert.Equal(t, int(api.status.Auction.ClosedAuctionSlots)+1, len(api.status.Network.NextForgers))
}

func TestUpdateMetrics(t *testing.T) {
	// TODO: Improve checks when til is integrated
	// Update Metrics needs api.status.Network.LastBatch.BatchNum to be updated
	lastBlock := tc.blocks[3]
	lastBatchNum := common.BatchNum(3)
	currentSlotNum := int64(1)
	err := api.UpdateNetworkInfo(lastBlock, lastBlock, lastBatchNum, currentSlotNum)
	assert.NoError(t, err)

	err = api.UpdateMetrics()
	assert.NoError(t, err)
	assert.Greater(t, api.status.Metrics.TransactionsPerBatch, float64(0))
	assert.Greater(t, api.status.Metrics.BatchFrequency, float64(0))
	assert.Greater(t, api.status.Metrics.TransactionsPerBatch, float64(0))
	assert.Greater(t, api.status.Metrics.TotalAccounts, int64(0))
	assert.Greater(t, api.status.Metrics.TotalBJJs, int64(0))
	assert.Greater(t, api.status.Metrics.AvgTransactionFee, float64(0))
}

func TestUpdateRecommendedFee(t *testing.T) {
	err := api.UpdateRecommendedFee()
	assert.NoError(t, err)
	assert.Greater(t, api.status.RecommendedFee.ExistingAccount, float64(0))
	assert.Equal(t, api.status.RecommendedFee.CreatesAccount,
		api.status.RecommendedFee.ExistingAccount*createAccountExtraFeePercentage)
	assert.Equal(t, api.status.RecommendedFee.CreatesAccountAndRegister,
		api.status.RecommendedFee.ExistingAccount*createAccountInternalExtraFeePercentage)
}

func TestGetState(t *testing.T) {
	lastBlock := tc.blocks[3]
	lastBatchNum := common.BatchNum(3)
	currentSlotNum := int64(1)
	api.SetRollupVariables(tc.rollupVars)
	api.SetWDelayerVariables(tc.wdelayerVars)
	api.SetAuctionVariables(tc.auctionVars)
	err := api.UpdateNetworkInfo(lastBlock, lastBlock, lastBatchNum, currentSlotNum)
	assert.NoError(t, err)
	err = api.UpdateMetrics()
	assert.NoError(t, err)
	err = api.UpdateRecommendedFee()
	assert.NoError(t, err)

	endpoint := apiURL + "state"
	var status testStatus

	assert.NoError(t, doGoodReq("GET", endpoint, nil, &status))
	assert.Equal(t, tc.rollupVars, status.Rollup)
	assert.Equal(t, tc.auctionVars, status.Auction)
	assert.Equal(t, tc.wdelayerVars, status.WithdrawalDelayer)
	assert.Equal(t, lastBlock.Num, status.Network.LastEthBlock)
	assert.Equal(t, lastBlock.Num, status.Network.LastSyncBlock)
	assert.Equal(t, lastBatchNum, status.Network.LastBatch.BatchNum)
	assert.Equal(t, currentSlotNum, status.Network.CurrentSlot)
	assert.Equal(t, int(api.status.Auction.ClosedAuctionSlots)+1, len(status.Network.NextForgers))
	assert.Greater(t, status.Metrics.TransactionsPerBatch, float64(0))
	assert.Greater(t, status.Metrics.BatchFrequency, float64(0))
	assert.Greater(t, status.Metrics.TransactionsPerBatch, float64(0))
	assert.Greater(t, status.Metrics.TotalAccounts, int64(0))
	assert.Greater(t, status.Metrics.TotalBJJs, int64(0))
	assert.Greater(t, status.Metrics.AvgTransactionFee, float64(0))
	assert.Greater(t, status.RecommendedFee.ExistingAccount, float64(0))
	assert.Equal(t, status.RecommendedFee.CreatesAccount,
		status.RecommendedFee.ExistingAccount*createAccountExtraFeePercentage)
	assert.Equal(t, status.RecommendedFee.CreatesAccountAndRegister,
		status.RecommendedFee.ExistingAccount*createAccountInternalExtraFeePercentage)
}
