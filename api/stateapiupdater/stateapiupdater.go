/*
Package stateapiupdater is responsible for generating and storing the object response of the GET /state endpoint exposed through the api package.
This object is extensively defined at the OpenAPI spec located at api/swagger.yml.

Deployment considerations: in a setup where multiple processes are used (dedicated api process, separated coord / sync, ...), only one process should care
of using this package.
*/
package stateapiupdater

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"sync"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

// Updater is an utility object to facilitate updating the StateAPI
type Updater struct {
	hdb       *historydb.HistoryDB
	state     historydb.StateAPI
	config    historydb.NodeConfig
	vars      common.SCVariablesPtr
	consts    historydb.Constants
	rw        sync.RWMutex
	rfp       *RecommendedFeePolicy
	ethClient eth.ClientInterface
}

// RecommendedFeePolicy describes how the recommended fee is calculated
type RecommendedFeePolicy struct {
	PolicyType  RecommendedFeePolicyType `validate:"required" env:"HEZNODE_RECOMMENDEDFEEPOLICY_POLICYTYPE"`
	StaticValue float64                  `env:"HEZNODE_RECOMMENDEDFEEPOLICY_STATICVALUE"`
}

// RecommendedFeePolicyType describes the different available recommended fee strategies
type RecommendedFeePolicyType string

const (
	// RecommendedFeePolicyTypeStatic always give the same StaticValue as recommended fee
	RecommendedFeePolicyTypeStatic RecommendedFeePolicyType = "Static"
	// RecommendedFeePolicyTypeAvgLastHour set the recommended fee using the average fee of the last hour
	RecommendedFeePolicyTypeAvgLastHour RecommendedFeePolicyType = "AvgLastHour"
	// RecommendedFeePolicyTypeAvgLastHourResizable set the recommended fee using the average fee of the last hour and
	// taking in account the avg gas of the last ten ethereum blocks
	RecommendedFeePolicyTypeAvgLastHourResizable RecommendedFeePolicyType = "AvgLastHourResizable"
)

var gasPriceHistory []*big.Int

func (rfp *RecommendedFeePolicy) valid() bool {
	switch rfp.PolicyType {
	case RecommendedFeePolicyTypeStatic:
		if rfp.StaticValue == 0 {
			log.Warn("RecommendedFee is set to 0 USD, and the policy is static")
		}
		return true
	case RecommendedFeePolicyTypeAvgLastHour:
		return true
	case RecommendedFeePolicyTypeAvgLastHourResizable:
		return true
	default:
		return false
	}
}

// NewUpdater creates a new Updater
func NewUpdater(hdb *historydb.HistoryDB, config *historydb.NodeConfig, vars *common.SCVariables,
	consts *historydb.Constants, rfp *RecommendedFeePolicy, ethClient eth.ClientInterface) (*Updater, error) {
	if ok := rfp.valid(); !ok {
		return nil, tracerr.Wrap(fmt.Errorf("Invalid recommended fee policy: %v", rfp.PolicyType))
	}
	u := Updater{
		hdb:    hdb,
		config: *config,
		consts: *consts,
		state: historydb.StateAPI{
			NodePublicInfo: historydb.NodePublicInfo{
				ForgeDelay: config.ForgeDelay,
			},
		},
		rfp: rfp,
		ethClient: ethClient,
	}
	u.SetSCVars(vars.AsPtr())
	return &u, nil
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
	switch u.rfp.PolicyType {
	case RecommendedFeePolicyTypeStatic:
		u.rw.Lock()
		u.state.RecommendedFee = common.RecommendedFee{
			ExistingAccount:        u.rfp.StaticValue,
			CreatesAccount:         u.rfp.StaticValue,
			CreatesAccountInternal: u.rfp.StaticValue,
		}
		u.rw.Unlock()
	case RecommendedFeePolicyTypeAvgLastHour:
		recommendedFee, err := u.hdb.GetRecommendedFee(u.config.MinFeeUSD, u.config.MaxFeeUSD, nil)
		if err != nil {
			return tracerr.Wrap(err)
		}
		u.rw.Lock()
		u.state.RecommendedFee = *recommendedFee
		u.rw.Unlock()
	case RecommendedFeePolicyTypeAvgLastHourResizable:
		//First get the average gas price used from history array
		//Every time this function is executed, checks the gasPrice and stores it in the gasPriceHistory array. If gasPriceHistory has less than 10 elements include a new one
		//If gasPriceHistory has more than ten elements, then remove the position 0 without loosing the order and include the current one.
		ctx := context.Background() 
		gas, err := u.ethClient.EthSuggestGasPrice(ctx)
		if err != nil { //If err, gasPriceHistory is not modified to avoid deviations
			log.Error("error getting gas price for recommended fee: ", err)
		} else {
			log.Debug("gas Price History arr: ", gasPriceHistory)
			if len(gasPriceHistory) < 10 {
				gasPriceHistory = append(gasPriceHistory, gas)
			} else {
				//Remove old element and add the new one
				gasPriceHistory = removeGasPriceHistoryElement(gasPriceHistory, 0)
				gasPriceHistory = append(gasPriceHistory, gas)
			}
		}
		gasPriceAvg := big.NewInt(0)
		totSum := big.NewInt(0)
		if len(gasPriceHistory) != 0 {
			// Calculate gasPriceAvg
			for _,val := range gasPriceHistory {
				totSum.Add(totSum, val)
			}
			gasPriceAvg.Div(totSum, big.NewInt(int64(len(gasPriceHistory))))
		}
		gasPriceAvgf := new(big.Float).SetInt(gasPriceAvg)
		recommendedFee, err := u.hdb.GetRecommendedFee(u.config.MinFeeUSD, u.config.MaxFeeUSD, gasPriceAvgf)
		if err != nil {
			return tracerr.Wrap(err)
		}
		u.rw.Lock()
		u.state.RecommendedFee = *recommendedFee
		u.rw.Unlock()
	default:
		return tracerr.New("Invalid recommende fee policy: " + string(u.rfp.PolicyType))
	}

	return nil
}

func removeGasPriceHistoryElement(slice []*big.Int, s int) []*big.Int {
    return append(slice[:s], slice[s+1:]...)
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
