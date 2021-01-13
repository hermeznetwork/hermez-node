package txsets

import (
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/test/til"
	"github.com/stretchr/testify/assert"
)

func TestCompileSetsBase(t *testing.T) {
	tc := til.NewContext(0, common.RollupConstMaxL1UserTx)
	_, err := tc.GenerateBlocks(SetBlockchain0)
	assert.NoError(t, err)
	_, err = tc.GeneratePoolL2Txs(SetPool0)
	assert.NoError(t, err)
}

func TestCompileSetsMinimumFlow(t *testing.T) {
	// minimum flow
	tc := til.NewContext(0, common.RollupConstMaxL1UserTx)
	_, err := tc.GenerateBlocks(SetBlockchainMinimumFlow0)
	assert.NoError(t, err)
	_, err = tc.GeneratePoolL2Txs(SetPoolL2MinimumFlow0)
	assert.NoError(t, err)
}
