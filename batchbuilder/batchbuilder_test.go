package batchbuilder

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)

	synchDB, err := statedb.NewStateDB(dir, false, 0)
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	bb, err := NewBatchBuilder(bbDir, synchDB, nil, 0, 32)
	assert.Nil(t, err)
	fmt.Println(bb)
}
