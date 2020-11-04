package api

import (
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
)

// Status define status of the network
type Status struct {
	Network           historydb.Network        `json:"network"`
	Metrics           historydb.Metrics        `json:"metrics"`
	Rollup            common.RollupVariables   `json:"rollup"`
	Auction           common.AuctionVariables  `json:"auction"`
	WithdrawalDelayer common.WDelayerVariables `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee    `json:"recommendedFee"`
}
