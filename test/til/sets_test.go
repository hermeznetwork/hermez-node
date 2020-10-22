package til

import (
	"strings"
	"testing"

	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/stretchr/testify/assert"
)

func TestCompileSets(t *testing.T) {
	parser := newParser(strings.NewReader(SetBlockchain0))
	_, err := parser.parse()
	assert.Nil(t, err)
	parser = newParser(strings.NewReader(SetPool0))
	_, err = parser.parse()
	assert.Nil(t, err)

	tc := NewContext(eth.RollupConstMaxL1UserTx)
	_, err = tc.GenerateBlocks(SetBlockchain0)
	assert.Nil(t, err)
	_, err = tc.GeneratePoolL2Txs(SetPool0)
	assert.Nil(t, err)
}
