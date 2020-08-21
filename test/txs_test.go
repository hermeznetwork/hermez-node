package test

import (
	"strings"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTestL2Txs(t *testing.T) {
	s := `
		A (1): 10
		A (2): 20
		B (1): 5
		A-B (1): 6 1
		B-C (1): 3 1
		C-A (1): 3 1
		A-B (1): 1 1
		A-B (2): 15 1
		User0   (1): 20
		User1 (3) : 20
		User0-User1 (1): 15 1
		User1-User0 (3): 15 1
	`
	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1txs, l2txs := GenerateTestTxs(t, instructions)
	require.Equal(t, 5, len(l1txs))
	require.Equal(t, 7, len(l2txs))

	// l1txs
	assert.Equal(t, common.TxTypeCreateAccountDeposit, l1txs[0].Type)
	assert.Equal(t, "5bac784d938067d980a9d39bdd79bf84a0cbb296977c47cc30de2d5ce9229d2f", l1txs[0].FromBJJ.String())
	assert.Equal(t, "323ff10c28df37ecb787fe216e111db64aa7cfa2c517509fe0057ff08a10b30c", l1txs[1].FromBJJ.String())
	assert.Equal(t, "f3587ad5cc7414a47545770b6c75bc71930f63c491eb2294dde8b8a6670b8e96", l1txs[2].FromBJJ.String())
	assert.Equal(t, "b6856a87832b182e5a9a1e738dbcd1f3c728bbc67ea1010aaff563eb5316131b", l1txs[4].FromBJJ.String())

	// l2txs
	assert.Equal(t, common.TxTypeTransfer, l2txs[0].Type)
	assert.Equal(t, common.Idx(1), l2txs[0].FromIdx)
	assert.Equal(t, common.Idx(3), l2txs[0].ToIdx)
	assert.Equal(t, "f3587ad5cc7414a47545770b6c75bc71930f63c491eb2294dde8b8a6670b8e96", l2txs[0].ToBJJ.String())
	assert.Equal(t, "0x6813Eb9362372EEF6200f3b1dbC3f819671cBA69", l2txs[0].ToEthAddr.Hex())
	assert.Equal(t, common.Nonce(0), l2txs[0].Nonce)
	assert.Equal(t, common.Nonce(1), l2txs[3].Nonce)
	assert.Equal(t, common.FeeSelector(1), l2txs[0].Fee)
}
