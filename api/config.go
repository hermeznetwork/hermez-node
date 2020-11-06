package api

import (
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
)

type rollupConstants struct {
	PublicConstants         common.RollupConstants `json:"publicConstants"`
	MaxFeeIdxCoordinator    int                    `json:"maxFeeIdxCoordinator"`
	ReservedIdx             int                    `json:"reservedIdx"`
	ExitIdx                 int                    `json:"exitIdx"`
	LimitLoadAmount         *big.Int               `json:"limitLoadAmount"`
	LimitL2TransferAmount   *big.Int               `json:"limitL2TransferAmount"`
	LimitTokens             int                    `json:"limitTokens"`
	L1CoordinatorTotalBytes int                    `json:"l1CoordinatorTotalBytes"`
	L1UserTotalBytes        int                    `json:"l1UserTotalBytes"`
	MaxL1UserTx             int                    `json:"maxL1UserTx"`
	MaxL1Tx                 int                    `json:"maxL1Tx"`
	InputSHAConstantBytes   int                    `json:"inputSHAConstantBytes"`
	NumBuckets              int                    `json:"numBuckets"`
	MaxWithdrawalDelay      int                    `json:"maxWithdrawalDelay"`
	ExchangeMultiplier      int                    `json:"exchangeMultiplier"`
}

type configAPI struct {
	RollupConstants   rollupConstants          `json:"hermez"`
	AuctionConstants  common.AuctionConstants  `json:"auction"`
	WDelayerConstants common.WDelayerConstants `json:"withdrawalDelayer"`
}

func (a *API) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, a.cg)
}
