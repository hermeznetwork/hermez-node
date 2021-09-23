/*
Package stateapiupdater is responsible for generating and storing the object response of the GET /state endpoint exposed through the api package.
This object is extensively defined at the OpenAPI spec located at api/swagger.yml.

Deployment considerations: in a setup where multiple processes are used (dedicated api process, separated coord / sync, ...), only one process should care
of using this package.
*/
package stateapiupdater

import (
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/params"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

// Updater is an utility object to facilitate updating the StateAPI
type Updater struct {
	hdb           *historydb.HistoryDB
	state         historydb.StateAPI
	config        historydb.NodeConfig
	vars          common.SCVariablesPtr
	consts        historydb.Constants
	rw            sync.RWMutex
	rfp           *RecommendedFeePolicy
	maxTxPerBatch int64
}

// RecommendedFeePolicy describes how the recommended fee is calculated
type RecommendedFeePolicy struct {
	PolicyType      RecommendedFeePolicyType `validate:"required" env:"HEZNODE_RECOMMENDEDFEEPOLICY_POLICYTYPE"`
	StaticValue     float64                  `env:"HEZNODE_RECOMMENDEDFEEPOLICY_STATICVALUE"`
	BreakThreshold  int                      `env:"HEZNODE_RECOMMENDEDFEEPOLICY_BREAKTHRESHOLD"`
	NumLastBatchAvg int                      `env:"HEZNODE_RECOMMENDEDFEEPOLICY_NUMLASTBATCHAVG"`
}

// RecommendedFeePolicyType describes the different available recommended fee strategies
type RecommendedFeePolicyType string

const (
	// RecommendedFeePolicyTypeStatic always give the same StaticValue as recommended fee
	RecommendedFeePolicyTypeStatic RecommendedFeePolicyType = "Static"
	// RecommendedFeePolicyTypeAvgLastHour set the recommended fee using the average fee of the last hour
	RecommendedFeePolicyTypeAvgLastHour RecommendedFeePolicyType = "AvgLastHour"
	// RecommendedFeePolicyTypeDynamicFee set the recommended fee taking in account the gas used in L1,
	// the gasPrice and the ether price in the last batches
	RecommendedFeePolicyTypeDynamicFee RecommendedFeePolicyType = "DynamicFee"
)

func (rfp *RecommendedFeePolicy) valid() bool {
	switch rfp.PolicyType {
	case RecommendedFeePolicyTypeStatic:
		if rfp.StaticValue == 0 {
			log.Warn("RecommendedFee is set to 0 USD, and the policy is static")
		}
		return true
	case RecommendedFeePolicyTypeAvgLastHour:
		return true
	case RecommendedFeePolicyTypeDynamicFee:
		return true
	default:
		return false
	}
}

// NewUpdater creates a new Updater
func NewUpdater(hdb *historydb.HistoryDB, config *historydb.NodeConfig, vars *common.SCVariables,
	consts *historydb.Constants, rfp *RecommendedFeePolicy, maxTxPerBatch int64) (*Updater, error) {
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
		rfp:           rfp,
		maxTxPerBatch: maxTxPerBatch,
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
		recommendedFee, err := u.hdb.GetRecommendedFee(u.config.MinFeeUSD, u.config.MaxFeeUSD)
		if err != nil {
			return tracerr.Wrap(err)
		}
		u.rw.Lock()
		u.state.RecommendedFee = *recommendedFee
		u.rw.Unlock()
	case RecommendedFeePolicyTypeDynamicFee:
		var recommendedFee common.RecommendedFee
		batchSize := u.maxTxPerBatch
		latestBatches, err := u.hdb.GetLatestBatches(u.rfp.NumLastBatchAvg)
		if err != nil {
			log.Error("error getting latest "+strconv.Itoa(u.rfp.NumLastBatchAvg)+" batches. Error: ", err)
		}
		breakThreshold := u.rfp.BreakThreshold

		//Calculate average batchCostUSD of the last x batches. batchCostUSD=batchGas*GasPrice(inGwei)*EtherPrice/10000000000
		avgBatchCostUSD := big.NewFloat(0)
		if len(latestBatches) != 0 {
			for _, batch := range latestBatches {
				gasUsedF := new(big.Float).SetUint64(batch.GasUsed)
				gasPriceF := new(big.Float).Quo(new(big.Float).SetInt(batch.GasPrice), new(big.Float).SetInt64(int64(params.GWei))) //In Gwei
				etherPriceF := new(big.Float).SetFloat64(batch.EtherPriceUSD)

				batchCostUSD := new(big.Float).Quo(new(big.Float).Mul(gasUsedF, new(big.Float).Mul(gasPriceF, etherPriceF)), new(big.Float).SetInt64(int64(params.GWei)))
				avgBatchCostUSD.Add(avgBatchCostUSD, batchCostUSD)
			}
			avgBatchCostUSD.Quo(avgBatchCostUSD, new(big.Float).SetInt64(int64(len(latestBatches))))
		}

		//breakEvenTxs=batchSize*breakEvenThreshold
		batchSizeF := new(big.Float).SetInt64(batchSize)
		if breakThreshold > 100 || breakThreshold < 0 {
			log.Error("invalid configuration parameter BreakThreshold. It is a percentage, must be a number between 0 and 100")
			return tracerr.New("invalid configuration parameter BreakThreshold. It is a percentage, must be a number between 0 and 100")
		}
		//nolint
		breakEvenThresholdF := new(big.Float).Quo(new(big.Float).SetInt64(int64(breakThreshold)), new(big.Float).SetInt64(100)) //Divided by 100 because it is a percentage
		breakEvenTxs := new(big.Float).Mul(batchSizeF, breakEvenThresholdF)

		//SuggestedPriceUSD=batchCostUSD/breakEvenTxs
		suggestedPriceUSD, _ := new(big.Float).Quo(avgBatchCostUSD, breakEvenTxs).Float64()
		log.Debug("suggestedPriceUSD: ", suggestedPriceUSD)

		//RecommendedFee
		recommendedFee.ExistingAccount = math.Min(u.config.MaxFeeUSD,
			math.Max(suggestedPriceUSD, u.config.MinFeeUSD))
		recommendedFee.CreatesAccount = math.Min(u.config.MaxFeeUSD,
			math.Max(historydb.CreateAccountExtraFeePercentage*suggestedPriceUSD, u.config.MinFeeUSD))
		recommendedFee.CreatesAccountInternal = math.Min(u.config.MaxFeeUSD,
			math.Max(historydb.CreateAccountInternalExtraFeePercentage*suggestedPriceUSD, u.config.MinFeeUSD))

		u.rw.Lock()
		u.state.RecommendedFee = recommendedFee
		u.rw.Unlock()
	default:
		return tracerr.New("Invalid recommende fee policy: " + string(u.rfp.PolicyType))
	}
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
