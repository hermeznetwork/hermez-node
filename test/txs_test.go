package test

import (
	"strings"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTestL2Txs(t *testing.T) {
	s := `
		A (1): 10
		A (2): 20
		B (1): 5
		A-B (1): 6 1
		B-C (1): 3 1
		> advance batch
		C-A (1): 3 1
		A-B (1): 1 1
		A-B (2): 15 1
		User0   (1): 20
		User1 (3) : 20
		User0-User1 (1): 15 1
		User1-User0 (3): 15 1
		B-D (2): 3 1
	`
	parser := NewParser(strings.NewReader(s))
	instructions, err := parser.Parse()
	assert.Nil(t, err)

	l1txs, coordinatorL1txs, l2txs := GenerateTestTxs(t, instructions)
	assert.Equal(t, 2, len(l1txs))
	assert.Equal(t, 3, len(l1txs[0]))
	assert.Equal(t, 1, len(coordinatorL1txs[0]))
	assert.Equal(t, 2, len(l2txs[0]))
	assert.Equal(t, 2, len(l1txs[1]))
	assert.Equal(t, 4, len(coordinatorL1txs[1]))
	assert.Equal(t, 6, len(l2txs[1]))

	// l1txs
	assert.Equal(t, common.TxTypeCreateAccountDeposit, l1txs[0][0].Type)
	assert.Equal(t, "5bac784d938067d980a9d39bdd79bf84a0cbb296977c47cc30de2d5ce9229d2f", l1txs[0][0].FromBJJ.String())
	assert.Equal(t, "323ff10c28df37ecb787fe216e111db64aa7cfa2c517509fe0057ff08a10b30c", l1txs[0][1].FromBJJ.String())
	assert.Equal(t, "f3587ad5cc7414a47545770b6c75bc71930f63c491eb2294dde8b8a6670b8e96", l1txs[0][2].FromBJJ.String())
	assert.Equal(t, "750a24a874a81c6c6f8aaa168ff2ee88c58263fee9ddd96d9717bcffc809b027", l1txs[1][1].FromBJJ.String())

	// l2txs
	assert.Equal(t, common.TxTypeTransfer, l2txs[0][0].Type)
	assert.Equal(t, common.Idx(1), l2txs[0][0].FromIdx)
	assert.Equal(t, common.Idx(3), l2txs[0][0].ToIdx)
	assert.Equal(t, "f3587ad5cc7414a47545770b6c75bc71930f63c491eb2294dde8b8a6670b8e96", l2txs[0][0].ToBJJ.String())
	assert.Equal(t, "0x6813Eb9362372EEF6200f3b1dbC3f819671cBA69", l2txs[0][0].ToEthAddr.Hex())
	assert.Equal(t, common.Nonce(0), l2txs[0][0].Nonce)
	assert.Equal(t, common.Nonce(1), l2txs[1][1].Nonce)
	assert.Equal(t, common.FeeSelector(1), l2txs[0][0].Fee)
}
