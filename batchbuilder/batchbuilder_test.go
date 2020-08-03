package batchbuilder

import (
	"fmt"
	"testing"

	"github.com/iden3/go-merkletree/db/memory"
	"github.com/stretchr/testify/assert"
)

func TestBatchBuilder(t *testing.T) {

	stateDB := memory.NewMemoryStorage()

	bb, err := NewBatchBuilder(stateDB, nil, 0, 0, 32)
	assert.Nil(t, err)
	fmt.Println(bb)
}
