package test

import (
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePoolL2Txs(t *testing.T) {
	set := `
		Transfer(1) A-B: 6 (1)
		Transfer(1) B-C: 3 (1)
		Transfer(1) C-A: 3 (1)
		Transfer(1) A-B: 1 (1)
		Transfer(2) A-B: 15 (1)
		Transfer(1) User0-User1: 15 (1)
		Transfer(3) User1-User0: 15 (1)
		Transfer(2) B-D: 3 (1)
		Exit(1) A: 3
	`
	tc := NewTestContext(t)
	poolL2Txs := tc.GeneratePoolL2Txs(set)
	assert.Equal(t, 9, len(poolL2Txs))
	assert.Equal(t, common.TxTypeTransfer, poolL2Txs[0].Type)
	assert.Equal(t, common.TxTypeExit, poolL2Txs[8].Type)
	assert.Equal(t, tc.accounts["B1"].Addr.Hex(), poolL2Txs[0].ToEthAddr.Hex())
	assert.Equal(t, tc.accounts["B1"].BJJ.Public().String(), poolL2Txs[0].ToBJJ.String())
	assert.Equal(t, tc.accounts["User11"].Addr.Hex(), poolL2Txs[5].ToEthAddr.Hex())
	assert.Equal(t, tc.accounts["User11"].BJJ.Public().String(), poolL2Txs[5].ToBJJ.String())

	assert.Equal(t, common.Nonce(1), poolL2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(2), poolL2Txs[3].Nonce)
	assert.Equal(t, common.Nonce(3), poolL2Txs[8].Nonce)

	// load another set in the same TestContext
	set = `
		Transfer(1) A-B: 6 (1)
		Transfer(1) B-C: 3 (1)
		Transfer(1) A-C: 3 (1)
	`
	poolL2Txs = tc.GeneratePoolL2Txs(set)
	assert.Equal(t, common.Nonce(4), poolL2Txs[0].Nonce)
	assert.Equal(t, common.Nonce(2), poolL2Txs[1].Nonce)
	assert.Equal(t, common.Nonce(5), poolL2Txs[2].Nonce)
}
