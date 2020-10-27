package api

import (
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
)

// Status define status of the network
type Status struct {
	Network           historydb.Network     `json:"network"`
	Metrics           historydb.Metrics     `json:"metrics"`
	Rollup            eth.RollupVariables   `json:"rollup"`
	Auction           eth.AuctionVariables  `json:"auction"`
	WithdrawalDelayer eth.WDelayerVariables `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee `json:"recommendedFee"`
}
