package kvdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addTestKV(t *testing.T, db *KVDB, k, v []byte) {
	tx, err := db.db.NewTx()
	require.NoError(t, err)

	err = tx.Put(k, v)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)
}

func printCheckpoints(t *testing.T, path string) {
	files, err := ioutil.ReadDir(path)
	assert.NoError(t, err)

	fmt.Println(path)
	for _, f := range files {
		fmt.Println("	" + f.Name())
	}
}

func TestCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "sdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	db, err := NewKVDB(dir, 128)
	assert.NoError(t, err)

	// add test key-values
	for i := 0; i < 10; i++ {
		addTestKV(t, db, []byte{byte(i), byte(i)}, []byte{byte(i * 2), byte(i * 2)})
	}

	// do checkpoints and check that currentBatch is correct
	err = db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err := db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(1), cb)

	for i := 1; i < 10; i++ {
		err = db.MakeCheckpoint()
		assert.NoError(t, err)

		cb, err = db.GetCurrentBatch()
		assert.NoError(t, err)
		assert.Equal(t, common.BatchNum(i+1), cb)
	}

	// printCheckpoints(t, sdb.path)

	// reset checkpoint
	err = db.Reset(3)
	assert.NoError(t, err)

	// check that reset can be repeated (as there exist the 'current' and
	// 'BatchNum3', from where the 'current' is a copy)
	err = db.Reset(3)
	require.NoError(t, err)

	// check that currentBatch is as expected after Reset
	cb, err = db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(3), cb)

	// advance one checkpoint and check that currentBatch is fine
	err = db.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = db.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)

	err = db.DeleteCheckpoint(common.BatchNum(1))
	assert.NoError(t, err)
	err = db.DeleteCheckpoint(common.BatchNum(2))
	assert.NoError(t, err)
	err = db.DeleteCheckpoint(common.BatchNum(1)) // does not exist, should return err
	assert.NotNil(t, err)
	err = db.DeleteCheckpoint(common.BatchNum(2)) // does not exist, should return err
	assert.NotNil(t, err)

	// Create a new KVDB which will get Reset from the initial KVDB
	dirLocal, err := ioutil.TempDir("", "ldb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dirLocal))
	ldb, err := NewKVDB(dirLocal, 128)
	assert.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb.ResetFromSynchronizer(4, db)
	assert.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb
	err = ldb.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = ldb.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	// Create a 3rd KVDB which will get Reset from the initial KVDB
	dirLocal2, err := ioutil.TempDir("", "ldb2")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dirLocal2))
	ldb2, err := NewKVDB(dirLocal2, 128)
	assert.NoError(t, err)

	// get checkpoint 4 from sdb (StateDB) to ldb (LocalStateDB)
	err = ldb2.ResetFromSynchronizer(4, db)
	assert.NoError(t, err)
	// check that currentBatch is 4 after the Reset
	cb, err = ldb2.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(4), cb)
	// advance one checkpoint in ldb2
	err = ldb2.MakeCheckpoint()
	assert.NoError(t, err)
	cb, err = ldb2.GetCurrentBatch()
	assert.NoError(t, err)
	assert.Equal(t, common.BatchNum(5), cb)

	debug := false
	if debug {
		printCheckpoints(t, db.path)
		printCheckpoints(t, ldb.path)
		printCheckpoints(t, ldb2.path)
	}
}

func TestListCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	db, err := NewKVDB(dir, 128)
	require.NoError(t, err)

	numCheckpoints := 16
	// do checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = db.MakeCheckpoint()
		require.NoError(t, err)
	}
	list, err := db.ListCheckpoints()
	require.NoError(t, err)
	assert.Equal(t, numCheckpoints, len(list))
	assert.Equal(t, 1, list[0])
	assert.Equal(t, numCheckpoints, list[len(list)-1])

	numReset := 10
	err = db.Reset(common.BatchNum(numReset))
	require.NoError(t, err)
	list, err = db.ListCheckpoints()
	require.NoError(t, err)
	assert.Equal(t, numReset, len(list))
	assert.Equal(t, 1, list[0])
	assert.Equal(t, numReset, list[len(list)-1])
}

func TestDeleteOldCheckpoints(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.NoError(t, err)
	defer assert.NoError(t, os.RemoveAll(dir))

	keep := 16
	db, err := NewKVDB(dir, keep)
	require.NoError(t, err)

	numCheckpoints := 32
	// do checkpoints and check that we never have more than `keep`
	// checkpoints
	for i := 0; i < numCheckpoints; i++ {
		err = db.MakeCheckpoint()
		require.NoError(t, err)
		checkpoints, err := db.ListCheckpoints()
		require.NoError(t, err)
		assert.LessOrEqual(t, len(checkpoints), keep)
	}
}
