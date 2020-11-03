package api

import (
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func getConfigTest() configAPI {
	var config configAPI

	config.RollupConstants.ExchangeMultiplier = common.RollupConstExchangeMultiplier
	config.RollupConstants.ExitIdx = common.RollupConstExitIDx
	config.RollupConstants.ReservedIdx = common.RollupConstReservedIDx
	config.RollupConstants.LimitLoadAmount, _ = new(big.Int).SetString("340282366920938463463374607431768211456", 10)
	config.RollupConstants.LimitL2TransferAmount, _ = new(big.Int).SetString("6277101735386680763835789423207666416102355444464034512896", 10)
	config.RollupConstants.LimitTokens = common.RollupConstLimitTokens
	config.RollupConstants.L1CoordinatorTotalBytes = common.RollupConstL1CoordinatorTotalBytes
	config.RollupConstants.L1UserTotalBytes = common.RollupConstL1UserTotalBytes
	config.RollupConstants.MaxL1UserTx = common.RollupConstMaxL1UserTx
	config.RollupConstants.MaxL1Tx = common.RollupConstMaxL1Tx
	config.RollupConstants.InputSHAConstantBytes = common.RollupConstInputSHAConstantBytes
	config.RollupConstants.NumBuckets = common.RollupConstNumBuckets
	config.RollupConstants.MaxWithdrawalDelay = common.RollupConstMaxWithdrawalDelay
	var rollupPublicConstants common.RollupConstants
	rollupPublicConstants.AbsoluteMaxL1L2BatchTimeout = 240
	rollupPublicConstants.HermezAuctionContract = ethCommon.HexToAddress("0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e")
	rollupPublicConstants.HermezGovernanceDAOAddress = ethCommon.HexToAddress("0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4")
	rollupPublicConstants.SafetyAddress = ethCommon.HexToAddress("0xE5904695748fe4A84b40b3fc79De2277660BD1D3")
	rollupPublicConstants.TokenHEZ = ethCommon.HexToAddress("0xf784709d2317D872237C4bC22f867d1BAe2913AB")
	rollupPublicConstants.WithdrawDelayerContract = ethCommon.HexToAddress("0xD6C850aeBFDC46D7F4c207e445cC0d6B0919BDBe")
	var verifier common.RollupVerifierStruct
	verifier.MaxTx = 512
	verifier.NLevels = 32
	rollupPublicConstants.Verifiers = append(rollupPublicConstants.Verifiers, verifier)

	var auctionConstants common.AuctionConstants
	auctionConstants.BlocksPerSlot = 40
	auctionConstants.GenesisBlockNum = 100
	auctionConstants.GovernanceAddress = ethCommon.HexToAddress("0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4")
	auctionConstants.InitialMinimalBidding, _ = new(big.Int).SetString("10000000000000000000", 10)
	auctionConstants.HermezRollup = ethCommon.HexToAddress("0xEa960515F8b4C237730F028cBAcF0a28E7F45dE0")
	auctionConstants.TokenHEZ = ethCommon.HexToAddress("0xf784709d2317D872237C4bC22f867d1BAe2913AB")

	var wdelayerConstants common.WDelayerConstants
	wdelayerConstants.HermezRollup = ethCommon.HexToAddress("0xEa960515F8b4C237730F028cBAcF0a28E7F45dE0")
	wdelayerConstants.MaxEmergencyModeTime = uint64(1000000)
	wdelayerConstants.MaxWithdrawalDelay = uint64(10000000)

	config.RollupConstants.PublicConstants = rollupPublicConstants
	config.AuctionConstants = auctionConstants
	config.WDelayerConstants = wdelayerConstants

	return config
}

func TestGetConfig(t *testing.T) {
	endpoint := apiURL + "config"
	var configTest configAPI
	assert.NoError(t, doGoodReq("GET", endpoint, nil, &configTest))
	assert.Equal(t, config, configTest)
	assert.Equal(t, cg, &configTest)
}
