package batchbuilder

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(dir))

	chainID := uint16(0)
	synchDB, err := statedb.NewStateDB(dir, 128, statedb.TypeBatchBuilder, 0, chainID)
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	defer assert.Nil(t, os.RemoveAll(bbDir))
	_, err = NewBatchBuilder(bbDir, synchDB, nil, 0, 32)
	assert.Nil(t, err)
}
