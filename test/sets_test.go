package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileSets(t *testing.T) {
	parser := NewParser(strings.NewReader(SetTest0))
	_, err := parser.Parse()
	assert.Nil(t, err)
}
