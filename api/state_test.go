package api

import (
	"math/big"
	"testing"

	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStatus struct {
	Network           testNetwork                   `json:"network"`
	Metrics           historydb.Metrics             `json:"metrics"`
	Rollup            historydb.RollupVariablesAPI  `json:"rollup"`
	Auction           historydb.AuctionVariablesAPI `json:"auction"`
	WithdrawalDelayer common.WDelayerVariables      `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee         `json:"recommendedFee"`
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
	assertEqualRollupVariables(t, *rollupVars, api.status.Rollup, true)
	api.SetRollupVariables(tc.rollupVars)
	assertEqualRollupVariables(t, tc.rollupVars, api.status.Rollup, true)
}

func assertEqualRollupVariables(t *testing.T, rollupVariables common.RollupVariables, apiVariables historydb.RollupVariablesAPI, checkBuckets bool) {
	assert.Equal(t, apitypes.NewBigIntStr(rollupVariables.FeeAddToken), apiVariables.FeeAddToken)
	assert.Equal(t, rollupVariables.ForgeL1L2BatchTimeout, apiVariables.ForgeL1L2BatchTimeout)
	assert.Equal(t, rollupVariables.WithdrawalDelay, apiVariables.WithdrawalDelay)
	assert.Equal(t, rollupVariables.SafeMode, apiVariables.SafeMode)
	if checkBuckets {
		for i, bucket := range rollupVariables.Buckets {
			assert.Equal(t, apitypes.NewBigIntStr(bucket.BlockWithdrawalRate), apiVariables.Buckets[i].BlockWithdrawalRate)
			assert.Equal(t, apitypes.NewBigIntStr(bucket.CeilUSD), apiVariables.Buckets[i].CeilUSD)
			assert.Equal(t, apitypes.NewBigIntStr(bucket.MaxWithdrawals), apiVariables.Buckets[i].MaxWithdrawals)
			assert.Equal(t, apitypes.NewBigIntStr(bucket.Withdrawals), apiVariables.Buckets[i].Withdrawals)
		}
	}
}

func TestSetWDelayerVariables(t *testing.T) {
	wdelayerVars := &common.WDelayerVariables{}
	assert.Equal(t, *wdelayerVars, api.status.WithdrawalDelayer)
	api.SetWDelayerVariables(tc.wdelayerVars)
	assert.Equal(t, tc.wdelayerVars, api.status.WithdrawalDelayer)
}

func TestSetAuctionVariables(t *testing.T) {
	auctionVars := &common.AuctionVariables{}
	assertEqualAuctionVariables(t, *auctionVars, api.status.Auction)
	api.SetAuctionVariables(tc.auctionVars)
	assertEqualAuctionVariables(t, tc.auctionVars, api.status.Auction)
}

func assertEqualAuctionVariables(t *testing.T, auctionVariables common.AuctionVariables, apiVariables historydb.AuctionVariablesAPI) {
	assert.Equal(t, auctionVariables.EthBlockNum, apiVariables.EthBlockNum)
	assert.Equal(t, auctionVariables.DonationAddress, apiVariables.DonationAddress)
	assert.Equal(t, auctionVariables.BootCoordinator, apiVariables.BootCoordinator)
	assert.Equal(t, auctionVariables.BootCoordinatorURL, apiVariables.BootCoordinatorURL)
	assert.Equal(t, auctionVariables.DefaultSlotSetBidSlotNum, apiVariables.DefaultSlotSetBidSlotNum)
	assert.Equal(t, auctionVariables.ClosedAuctionSlots, apiVariables.ClosedAuctionSlots)
	assert.Equal(t, auctionVariables.OpenAuctionSlots, apiVariables.OpenAuctionSlots)
	assert.Equal(t, auctionVariables.Outbidding, apiVariables.Outbidding)
	assert.Equal(t, auctionVariables.SlotDeadline, apiVariables.SlotDeadline)

	for i, slot := range auctionVariables.DefaultSlotSetBid {
		assert.Equal(t, apitypes.NewBigIntStr(slot), apiVariables.DefaultSlotSetBid[i])
	}

	for i, ratio := range auctionVariables.AllocationRatio {
		assert.Equal(t, ratio, apiVariables.AllocationRatio[i])
	}
}

func TestUpdateNetworkInfo(t *testing.T) {
	status := &Network{}
	assert.Equal(t, status.LastSyncBlock, api.status.Network.LastSyncBlock)
	assert.Equal(t, status.LastBatch, api.status.Network.LastBatch)
	assert.Equal(t, status.CurrentSlot, api.status.Network.CurrentSlot)
	assert.Equal(t, status.NextForgers, api.status.Network.NextForgers)
	lastBlock := tc.blocks[3]
	lastBatchNum := common.BatchNum(3)
	currentSlotNum := int64(1)

	// Generate some bucket_update data
	bucketUpdates := []common.BucketUpdate{
		{
			EthBlockNum: 4,
			NumBucket:   0,
			BlockStamp:  4,
			Withdrawals: big.NewInt(123),
		},
		{
			EthBlockNum: 5,
			NumBucket:   2,
			BlockStamp:  5,
			Withdrawals: big.NewInt(42),
		},
		{
			EthBlockNum: 5,
			NumBucket:   2, // Repeated bucket
			BlockStamp:  5,
			Withdrawals: big.NewInt(43),
		},
	}
	err := api.h.AddBucketUpdatesTest(api.h.DB(), bucketUpdates)
	require.NoError(t, err)

	err = api.UpdateNetworkInfo(lastBlock, lastBlock, lastBatchNum, currentSlotNum)
	assert.NoError(t, err)
	assert.Equal(t, lastBlock.Num, api.status.Network.LastSyncBlock)
	assert.Equal(t, lastBatchNum, api.status.Network.LastBatch.BatchNum)
	assert.Equal(t, currentSlotNum, api.status.Network.CurrentSlot)
	assert.Equal(t, int(api.status.Auction.ClosedAuctionSlots)+1, len(api.status.Network.NextForgers))
	assert.Equal(t, api.status.Rollup.Buckets[0].Withdrawals, apitypes.NewBigIntStr(big.NewInt(123)))
	assert.Equal(t, api.status.Rollup.Buckets[2].Withdrawals, apitypes.NewBigIntStr(big.NewInt(43)))
}

func TestUpdateMetrics(t *testing.T) {
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
	assert.Greater(t, api.status.Metrics.TransactionsPerSecond, float64(0))
	assert.Greater(t, api.status.Metrics.TotalAccounts, int64(0))
	assert.Greater(t, api.status.Metrics.TotalBJJs, int64(0))
	assert.Greater(t, api.status.Metrics.AvgTransactionFee, float64(0))
}

func TestUpdateRecommendedFee(t *testing.T) {
	err := api.UpdateRecommendedFee()
	assert.NoError(t, err)
	var minFeeUSD float64
	if api.l2 != nil {
		minFeeUSD = api.l2.MinFeeUSD()
	}
	assert.Greater(t, api.status.RecommendedFee.ExistingAccount, minFeeUSD)
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

	// SC vars
	// UpdateNetworkInfo will overwrite buckets withdrawal values
	// So they won't be checked here, they are checked at
	// TestUpdateNetworkInfo
	assertEqualRollupVariables(t, tc.rollupVars, status.Rollup, false)
	assertEqualAuctionVariables(t, tc.auctionVars, status.Auction)
	assert.Equal(t, tc.wdelayerVars, status.WithdrawalDelayer)
	// Network
	assert.Equal(t, lastBlock.Num, status.Network.LastEthBlock)
	assert.Equal(t, lastBlock.Num, status.Network.LastSyncBlock)
	// TODO: assert all the batch, not just the batch num
	assert.Equal(t, lastBatchNum, status.Network.LastBatch.BatchNum)
	assert.Equal(t, currentSlotNum, status.Network.CurrentSlot)
	assertNextForgers(t, tc.nextForgers, status.Network.NextForgers)
	// Metrics
	// TODO: perform real asserts (not just greater than 0)
	assert.Greater(t, status.Metrics.TransactionsPerBatch, float64(0))
	assert.Greater(t, status.Metrics.BatchFrequency, float64(0))
	assert.Greater(t, status.Metrics.TransactionsPerSecond, float64(0))
	assert.Greater(t, status.Metrics.TotalAccounts, int64(0))
	assert.Greater(t, status.Metrics.TotalBJJs, int64(0))
	assert.Greater(t, status.Metrics.AvgTransactionFee, float64(0))
	// Recommended fee
	// TODO: perform real asserts (not just greater than 0)
	assert.Greater(t, status.RecommendedFee.ExistingAccount, float64(0))
	assert.Equal(t, status.RecommendedFee.CreatesAccount,
		status.RecommendedFee.ExistingAccount*createAccountExtraFeePercentage)
	assert.Equal(t, status.RecommendedFee.CreatesAccountAndRegister,
		status.RecommendedFee.ExistingAccount*createAccountInternalExtraFeePercentage)
}

func assertNextForgers(t *testing.T, expected, actual []NextForger) {
	assert.Equal(t, len(expected), len(actual))
	for i := range expected {
		// ignore timestamps and other metadata
		actual[i].Period.FromTimestamp = expected[i].Period.FromTimestamp
		actual[i].Period.ToTimestamp = expected[i].Period.ToTimestamp
		actual[i].Coordinator.ItemID = expected[i].Coordinator.ItemID
		actual[i].Coordinator.EthBlockNum = expected[i].Coordinator.EthBlockNum
		assert.Equal(t, expected[i], actual[i])
	}
}
