package api

import (
	"encoding/base64"
	"math/big"
	"strconv"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const exitIdx = "hez:EXIT:1"

type errorMsg struct {
	Message string
}

func bjjToString(bjj *babyjub.PublicKey) string {
	pkComp := [32]byte(bjj.Compress())
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

func ethAddrToHez(addr ethCommon.Address) string {
	return "hez:" + addr.String()
}

func idxToHez(idx common.Idx, tokenSymbol string) string {
	if idx == 1 {
		return exitIdx
	}
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

// Config

type rollupConstants struct {
	PublicConstants         eth.RollupPublicConstants `json:"publicConstants"`
	MaxFeeIdxCoordinator    int                       `json:"maxFeeIdxCoordinator"`
	ReservedIdx             int                       `json:"reservedIdx"`
	ExitIdx                 int                       `json:"exitIdx"`
	LimitLoadAmount         *big.Int                  `json:"limitLoadAmount"`
	LimitL2TransferAmount   *big.Int                  `json:"limitL2TransferAmount"`
	LimitTokens             int                       `json:"limitTokens"`
	L1CoordinatorTotalBytes int                       `json:"l1CoordinatorTotalBytes"`
	L1UserTotalBytes        int                       `json:"l1UserTotalBytes"`
	MaxL1UserTx             int                       `json:"maxL1UserTx"`
	MaxL1Tx                 int                       `json:"maxL1Tx"`
	InputSHAConstantBytes   int                       `json:"inputSHAConstantBytes"`
	NumBuckets              int                       `json:"numBuckets"`
	MaxWithdrawalDelay      int                       `json:"maxWithdrawalDelay"`
	ExchangeMultiplier      int                       `json:"exchangeMultiplier"`
}

type configAPI struct {
	RollupConstants   rollupConstants       `json:"hermez"`
	AuctionConstants  eth.AuctionConstants  `json:"auction"`
	WDelayerConstants eth.WDelayerConstants `json:"withdrawalDelayer"`
}
