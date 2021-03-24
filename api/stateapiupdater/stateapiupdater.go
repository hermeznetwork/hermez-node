package stateapiupdater

import (
	"database/sql"
	"sync"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/tracerr"
)

// Updater is an utility object to facilitate updating the StateAPI
type Updater struct {
	hdb    *historydb.HistoryDB
	state  historydb.StateAPI
	config historydb.NodeConfig
	vars   common.SCVariablesPtr
	consts historydb.Constants
	rw     sync.RWMutex
}

// NewUpdater creates a new Updater
func NewUpdater(hdb *historydb.HistoryDB, config *historydb.NodeConfig, vars *common.SCVariables,
	consts *historydb.Constants) *Updater {
	u := Updater{
		hdb:    hdb,
		config: *config,
		consts: *consts,
		state: historydb.StateAPI{
			NodePublicInfo: historydb.NodePublicInfo{
				ForgeDelay: config.ForgeDelay,
			},
		},
	}
	u.SetSCVars(vars.AsPtr())
	return &u
}

// Store the State in the HistoryDB
func (u *Updater) Store() error {
	u.rw.RLock()
	defer u.rw.RUnlock()
	return tracerr.Wrap(u.hdb.SetStateInternalAPI(&u.state))
}

// SetSCVars sets the smart contract vars (ony updates those that are not nil)
func (u *Updater) SetSCVars(vars *common.SCVariablesPtr) {
	u.rw.Lock()
	defer u.rw.Unlock()
	if vars.Rollup != nil {
		u.vars.Rollup = vars.Rollup
		rollupVars := historydb.NewRollupVariablesAPI(u.vars.Rollup)
		u.state.Rollup = *rollupVars
	}
	if vars.Auction != nil {
		u.vars.Auction = vars.Auction
		auctionVars := historydb.NewAuctionVariablesAPI(u.vars.Auction)
		u.state.Auction = *auctionVars
	}
	if vars.WDelayer != nil {
		u.vars.WDelayer = vars.WDelayer
		u.state.WithdrawalDelayer = *u.vars.WDelayer
	}
}

// UpdateRecommendedFee update Status.RecommendedFee information
func (u *Updater) UpdateRecommendedFee() error {
	// TODO: uncomment this once we have clear how to calculate recommended fee
	// recommendedFee, err := u.hdb.GetRecommendedFee(u.config.MinFeeUSD, u.config.MaxFeeUSD)
	// if err != nil {
	// 	return tracerr.Wrap(err)
	// }
	u.rw.Lock()
	hardcodedRecommendedFee := 0.99
	u.state.RecommendedFee = common.RecommendedFee{
		ExistingAccount:        hardcodedRecommendedFee,
		CreatesAccount:         hardcodedRecommendedFee,
		CreatesAccountInternal: hardcodedRecommendedFee,
	}
	u.rw.Unlock()
	return nil
}

// UpdateMetrics update Status.Metrics information
func (u *Updater) UpdateMetrics() error {
	u.rw.RLock()
	lastBatch := u.state.Network.LastBatch
	u.rw.RUnlock()
	if lastBatch == nil {
		return nil
	}
	lastBatchNum := lastBatch.BatchNum
	metrics, poolLoad, err := u.hdb.GetMetricsInternalAPI(lastBatchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	u.rw.Lock()
	u.state.Metrics = *metrics
	u.state.NodePublicInfo.PoolLoad = poolLoad
	u.rw.Unlock()
	return nil
}

// UpdateNetworkInfoBlock update Status.Network block related information
func (u *Updater) UpdateNetworkInfoBlock(lastEthBlock, lastSyncBlock common.Block) {
	u.rw.Lock()
	u.state.Network.LastSyncBlock = lastSyncBlock.Num
	u.state.Network.LastEthBlock = lastEthBlock.Num
	u.rw.Unlock()
}

// UpdateNetworkInfo update Status.Network information
func (u *Updater) UpdateNetworkInfo(
	lastEthBlock, lastSyncBlock common.Block,
	lastBatchNum common.BatchNum, currentSlot int64,
) error {
	// Get last batch in API format
	lastBatch, err := u.hdb.GetBatchInternalAPI(lastBatchNum)
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		lastBatch = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}
	u.rw.RLock()
	auctionVars := u.vars.Auction
	u.rw.RUnlock()
	// Get next forgers
	lastClosedSlot := currentSlot + int64(auctionVars.ClosedAuctionSlots)
	nextForgers, err := u.hdb.GetNextForgersInternalAPI(auctionVars, &u.consts.Auction,
		lastSyncBlock, currentSlot, lastClosedSlot)
	if tracerr.Unwrap(err) == sql.ErrNoRows {
		nextForgers = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}

	bucketUpdates, err := u.hdb.GetBucketUpdatesInternalAPI()
	if err == sql.ErrNoRows {
		bucketUpdates = nil
	} else if err != nil {
		return tracerr.Wrap(err)
	}

	u.rw.Lock()
	// Update NodeInfo struct
	for i, bucketParams := range u.state.Rollup.Buckets {
		for _, bucketUpdate := range bucketUpdates {
			if bucketUpdate.NumBucket == i {
				bucketParams.Withdrawals = bucketUpdate.Withdrawals
				u.state.Rollup.Buckets[i] = bucketParams
				break
			}
		}
	}
	// Update pending L1s
	pendingL1s, err := u.hdb.GetUnforgedL1UserTxsCount()
	if err != nil {
		return tracerr.Wrap(err)
	}
	u.state.Network.LastSyncBlock = lastSyncBlock.Num
	u.state.Network.LastEthBlock = lastEthBlock.Num
	u.state.Network.LastBatch = lastBatch
	u.state.Network.CurrentSlot = currentSlot
	u.state.Network.NextForgers = nextForgers
	u.state.Network.PendingL1Txs = pendingL1s
	u.rw.Unlock()
	return nil
}
