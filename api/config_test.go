package api

import (
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func getConfigTest(chainID uint16) NetworkConfig {
	var config NetworkConfig

	var rollupPublicConstants common.RollupConstants
	rollupPublicConstants.AbsoluteMaxL1L2BatchTimeout = 240
	rollupPublicConstants.HermezAuctionContract = ethCommon.HexToAddress("0x500D1d6A4c7D8Ae28240b47c8FCde034D827fD5e")
	rollupPublicConstants.HermezGovernanceAddress = ethCommon.HexToAddress("0xeAD9C93b79Ae7C1591b1FB5323BD777E86e150d4")
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

	config.RollupConstants = rollupPublicConstants
	config.AuctionConstants = auctionConstants
	config.WDelayerConstants = wdelayerConstants

	config.ChainID = chainID
	config.HermezAddress = ethCommon.HexToAddress("0xc344E203a046Da13b0B4467EB7B3629D0C99F6E6")

	return config
}

func TestGetConfig(t *testing.T) {
	endpoint := apiURL + "config"
	var configTest configAPI
	assert.NoError(t, doGoodReq("GET", endpoint, nil, &configTest))
	assert.Equal(t, config, configTest)
	assert.Equal(t, api.config, &configTest)
}
