// Package kvdb provides a key-value database with Checkpoints & Resets system
package kvdb

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

const (
	// PathBatchNum defines the subpath of the Batch Checkpoint in the
	// subpath of the KVDB
	PathBatchNum = "BatchNum"
	// PathCurrent defines the subpath of the current Batch in the subpath
	// of the KVDB
	PathCurrent = "current"
)

var (
	// KeyCurrentBatch is used as key in the db to store the current BatchNum
	KeyCurrentBatch = []byte("k:currentbatch")
	// keyCurrentIdx is used as key in the db to store the CurrentIdx
	keyCurrentIdx = []byte("k:idx")
)

// KVDB represents the Key-Value DB object
type KVDB struct {
	path string
	db   *pebble.Storage
	// CurrentIdx holds the current Idx that the BatchBuilder is using
	CurrentIdx   common.Idx
	CurrentBatch common.BatchNum
	keep         int
}

// NewKVDB creates a new KVDB, allowing to use an in-memory or in-disk storage.
// Checkpoints older than the value defined by `keep` will be deleted.
func NewKVDB(pathDB string, keep int) (*KVDB, error) {
	var sto *pebble.Storage
	var err error
	sto, err = pebble.NewPebbleStorage(path.Join(pathDB, PathCurrent), false)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	kvdb := &KVDB{
		path: pathDB,
		db:   sto,
		keep: keep,
	}
	// load currentBatch
	kvdb.CurrentBatch, err = kvdb.GetCurrentBatch()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// make reset (get checkpoint) at currentBatch
	err = kvdb.reset(kvdb.CurrentBatch, true)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return kvdb, nil
}

// DB returns the *pebble.Storage from the KVDB
func (kvdb *KVDB) DB() *pebble.Storage {
	return kvdb.db
}

// StorageWithPrefix returns the db.Storage with the given prefix from the
// current KVDB
func (kvdb *KVDB) StorageWithPrefix(prefix []byte) db.Storage {
	return kvdb.db.WithPrefix(prefix)
}

// Reset resets the KVDB to the checkpoint at the given batchNum. Reset does
// not delete the checkpoints between old current and the new current, those
// checkpoints will remain in the storage, and eventually will be deleted when
// MakeCheckpoint overwrites them.
func (kvdb *KVDB) Reset(batchNum common.BatchNum) error {
	return kvdb.reset(batchNum, true)
}

// reset resets the KVDB to the checkpoint at the given batchNum. Reset does
// not delete the checkpoints between old current and the new current, those
// checkpoints will remain in the storage, and eventually will be deleted when
// MakeCheckpoint overwrites them.  `closeCurrent` will close the currently
// opened db before doing the reset.
func (kvdb *KVDB) reset(batchNum common.BatchNum, closeCurrent bool) error {
	currentPath := path.Join(kvdb.path, PathCurrent)

	if closeCurrent {
		if err := kvdb.db.Pebble().Close(); err != nil {
			return tracerr.Wrap(err)
		}
	}
	// remove 'current'
	err := os.RemoveAll(currentPath)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// remove all checkpoints > batchNum
	list, err := kvdb.ListCheckpoints()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Find first batch that is greater than batchNum, and delete
	// everything after that
	start := 0
	for ; start < len(list); start++ {
		if common.BatchNum(list[start]) > batchNum {
			break
		}
	}
	for _, bn := range list[start:] {
		if err := kvdb.DeleteCheckpoint(common.BatchNum(bn)); err != nil {
			return tracerr.Wrap(err)
		}
	}

	if batchNum == 0 {
		// if batchNum == 0, open the new fresh 'current'
		sto, err := pebble.NewPebbleStorage(currentPath, false)
		if err != nil {
			return tracerr.Wrap(err)
		}
		kvdb.db = sto
		kvdb.CurrentIdx = 255
		kvdb.CurrentBatch = 0

		return nil
	}

	checkpointPath := path.Join(kvdb.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))
	// copy 'BatchNumX' to 'current'
	err = pebbleMakeCheckpoint(checkpointPath, currentPath)
	if err != nil {
		return tracerr.Wrap(err)
	}

	// open the new 'current'
	sto, err := pebble.NewPebbleStorage(currentPath, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	kvdb.db = sto

	// get currentBatch num
	kvdb.CurrentBatch, err = kvdb.GetCurrentBatch()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// idx is obtained from the statedb reset
	kvdb.CurrentIdx, err = kvdb.GetCurrentIdx()
	if err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

// ResetFromSynchronizer performs a reset in the KVDB getting the state from
// synchronizerKVDB for the given batchNum.
func (kvdb *KVDB) ResetFromSynchronizer(batchNum common.BatchNum, synchronizerKVDB *KVDB) error {
	if synchronizerKVDB == nil {
		return tracerr.Wrap(fmt.Errorf("synchronizerKVDB can not be nil"))
	}

	currentPath := path.Join(kvdb.path, PathCurrent)
	if err := kvdb.db.Pebble().Close(); err != nil {
		return tracerr.Wrap(err)
	}

	// remove 'current'
	err := os.RemoveAll(currentPath)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// remove all checkpoints
	list, err := kvdb.ListCheckpoints()
	if err != nil {
		return tracerr.Wrap(err)
	}
	for _, bn := range list {
		if err := kvdb.DeleteCheckpoint(common.BatchNum(bn)); err != nil {
			return tracerr.Wrap(err)
		}
	}

	if batchNum == 0 {
		// if batchNum == 0, open the new fresh 'current'
		sto, err := pebble.NewPebbleStorage(currentPath, false)
		if err != nil {
			return tracerr.Wrap(err)
		}
		kvdb.db = sto
		kvdb.CurrentIdx = 255
		kvdb.CurrentBatch = 0

		return nil
	}

	checkpointPath := path.Join(kvdb.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))

	// use checkpoint from synchronizerKVDB
	synchronizerCheckpointPath := path.Join(synchronizerKVDB.path,
		fmt.Sprintf("%s%d", PathBatchNum, batchNum))
	if _, err := os.Stat(synchronizerCheckpointPath); os.IsNotExist(err) {
		// if synchronizerKVDB does not have checkpoint at batchNum, return err
		return tracerr.Wrap(fmt.Errorf("Checkpoint \"%v\" not exist in Synchronizer",
			synchronizerCheckpointPath))
	}
	// copy synchronizer'BatchNumX' to 'BatchNumX'
	err = pebbleMakeCheckpoint(synchronizerCheckpointPath, checkpointPath)
	if err != nil {
		return tracerr.Wrap(err)
	}

	// copy 'BatchNumX' to 'current'
	err = pebbleMakeCheckpoint(checkpointPath, currentPath)
	if err != nil {
		return tracerr.Wrap(err)
	}

	// open the new 'current'
	sto, err := pebble.NewPebbleStorage(currentPath, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	kvdb.db = sto

	// get currentBatch num
	kvdb.CurrentBatch, err = kvdb.GetCurrentBatch()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// get currentIdx
	kvdb.CurrentIdx, err = kvdb.GetCurrentIdx()
	if err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

// GetCurrentBatch returns the current BatchNum stored in the KVDB
func (kvdb *KVDB) GetCurrentBatch() (common.BatchNum, error) {
	cbBytes, err := kvdb.db.Get(KeyCurrentBatch)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return common.BatchNumFromBytes(cbBytes)
}

// setCurrentBatch stores the current BatchNum in the KVDB
func (kvdb *KVDB) setCurrentBatch() error {
	tx, err := kvdb.db.NewTx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = tx.Put(KeyCurrentBatch, kvdb.CurrentBatch.Bytes())
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// GetCurrentIdx returns the stored Idx from the KVDB, which is the last Idx
// used for an Account in the KVDB.
func (kvdb *KVDB) GetCurrentIdx() (common.Idx, error) {
	idxBytes, err := kvdb.db.Get(keyCurrentIdx)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return common.IdxFromBytes(idxBytes[:])
}

// SetCurrentIdx stores Idx in the KVDB
func (kvdb *KVDB) SetCurrentIdx(idx common.Idx) error {
	kvdb.CurrentIdx = idx

	tx, err := kvdb.db.NewTx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = tx.Put(keyCurrentIdx, idxBytes[:])
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path.
// Internally this advances & stores the current BatchNum, and then stores a
// Checkpoint of the current state of the KVDB.
func (kvdb *KVDB) MakeCheckpoint() error {
	// advance currentBatch
	kvdb.CurrentBatch++

	checkpointPath := path.Join(kvdb.path, fmt.Sprintf("%s%d", PathBatchNum, kvdb.CurrentBatch))

	if err := kvdb.setCurrentBatch(); err != nil {
		return tracerr.Wrap(err)
	}

	// if checkpoint BatchNum already exist in disk, delete it
	if _, err := os.Stat(checkpointPath); !os.IsNotExist(err) {
		err := os.RemoveAll(checkpointPath)
		if err != nil {
			return tracerr.Wrap(err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return tracerr.Wrap(err)
	}

	// execute Checkpoint
	if err := kvdb.db.Pebble().Checkpoint(checkpointPath); err != nil {
		return tracerr.Wrap(err)
	}
	// delete old checkpoints
	if err := kvdb.deleteOldCheckpoints(); err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

// DeleteCheckpoint removes if exist the checkpoint of the given batchNum
func (kvdb *KVDB) DeleteCheckpoint(batchNum common.BatchNum) error {
	checkpointPath := path.Join(kvdb.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return tracerr.Wrap(fmt.Errorf("Checkpoint with batchNum %d does not exist in DB", batchNum))
	}

	return os.RemoveAll(checkpointPath)
}

// ListCheckpoints returns the list of batchNums of the checkpoints, sorted.
// If there's a gap between the list of checkpoints, an error is returned.
func (kvdb *KVDB) ListCheckpoints() ([]int, error) {
	files, err := ioutil.ReadDir(kvdb.path)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	checkpoints := []int{}
	var checkpoint int
	pattern := fmt.Sprintf("%s%%d", PathBatchNum)
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() && strings.HasPrefix(fileName, PathBatchNum) {
			if _, err := fmt.Sscanf(fileName, pattern, &checkpoint); err != nil {
				return nil, tracerr.Wrap(err)
			}
			checkpoints = append(checkpoints, checkpoint)
		}
	}
	sort.Ints(checkpoints)
	if len(checkpoints) > 0 {
		first := checkpoints[0]
		for _, checkpoint := range checkpoints[1:] {
			first++
			if checkpoint != first {
				log.Errorw("GAP", "checkpoints", checkpoints)
				return nil, tracerr.Wrap(fmt.Errorf("checkpoint gap at %v", checkpoint))
			}
		}
	}
	return checkpoints, nil
}

// deleteOldCheckpoints deletes old checkpoints when there are more than
// `s.keep` checkpoints
func (kvdb *KVDB) deleteOldCheckpoints() error {
	list, err := kvdb.ListCheckpoints()
	if err != nil {
		return tracerr.Wrap(err)
	}
	if len(list) > kvdb.keep {
		for _, checkpoint := range list[:len(list)-kvdb.keep] {
			if err := kvdb.DeleteCheckpoint(common.BatchNum(checkpoint)); err != nil {
				return tracerr.Wrap(err)
			}
		}
	}
	return nil
}

func pebbleMakeCheckpoint(source, dest string) error {
	// Remove dest folder (if it exists) before doing the checkpoint
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		err := os.RemoveAll(dest)
		if err != nil {
			return tracerr.Wrap(err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return tracerr.Wrap(err)
	}

	sto, err := pebble.NewPebbleStorage(source, false)
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer func() {
		errClose := sto.Pebble().Close()
		if errClose != nil {
			log.Errorw("Pebble.Close", "err", errClose)
		}
	}()

	// execute Checkpoint
	err = sto.Pebble().Checkpoint(dest)
	if err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}
