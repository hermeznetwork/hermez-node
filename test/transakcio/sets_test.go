package transakcio

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

	tc := NewTestContext(eth.RollupConstMaxL1UserTx)
	_, err = tc.GenerateBlocks(SetBlockchain0)
	assert.Nil(t, err)
	_, err = tc.GenerateBlocks(SetPool0)
	assert.Nil(t, err)
}