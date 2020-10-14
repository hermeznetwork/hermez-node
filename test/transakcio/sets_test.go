package transakcio

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileSets(t *testing.T) {
	parser := newParser(strings.NewReader(SetBlockchain0))
	_, err := parser.parse()
	assert.Nil(t, err)
	parser = newParser(strings.NewReader(SetPool0))
	_, err = parser.parse()
	assert.Nil(t, err)

	tc := NewTestContext()
	_ = tc.GenerateBlocks(SetBlockchain0)
	_ = tc.GenerateBlocks(SetPool0)
}
