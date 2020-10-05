package eth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rollupClient *RollupClient

var absoluteMaxL1L2BatchTimeout = uint8(240)
var maxTx = big.NewInt(512)
var nLevels = big.NewInt(32)

func TestRollupConstants(t *testing.T) {
	rollupConstants, err := rollupClient.RollupConstants()
	require.Nil(t, err)
	assert.Equal(t, absoluteMaxL1L2BatchTimeout, rollupConstants.AbsoluteMaxL1L2BatchTimeout)
	assert.Equal(t, auctionAddressConst, rollupConstants.HermezAuctionContract)
	assert.Equal(t, tokenERC777AddressConst, rollupConstants.TokenHEZ)
	assert.Equal(t, maxTx, rollupConstants.Verifiers[0].MaxTx)
	assert.Equal(t, nLevels, rollupConstants.Verifiers[0].NLevels)
	assert.Equal(t, governanceAddressConst, rollupConstants.HermezGovernanceDAOAddress)
	assert.Equal(t, safetyAddressConst, rollupConstants.SafetyAddress)
	assert.Equal(t, wdelayerAddressConst, rollupConstants.WithdrawDelayerContract)
}
