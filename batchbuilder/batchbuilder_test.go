package batchbuilder

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deleteme []string

func init() {
	log.Init("debug", []string{"stdout"})
}
func TestMain(m *testing.M) {
	exitVal := m.Run()
	for _, dir := range deleteme {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
	os.Exit(exitVal)
}

func TestBatchBuilder(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	deleteme = append(deleteme, dir)

	synchDB, err := statedb.NewStateDB(statedb.Config{Path: dir, Keep: 128,
		Type: statedb.TypeBatchBuilder, NLevels: 0})
	assert.Nil(t, err)

	bbDir, err := ioutil.TempDir("", "tmpBatchBuilderDB")
	require.Nil(t, err)
	deleteme = append(deleteme, bbDir)
	bb, err := NewBatchBuilder(bbDir, synchDB, 0, 32)
	assert.Nil(t, err)

	bb.LocalStateDB().Close()
	synchDB.Close()
}
