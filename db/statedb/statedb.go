package statedb

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

// ErrStateDBWithoutMT is used when a method that requires a MerkleTree is called in a StateDB that does not have a MerkleTree defined
var ErrStateDBWithoutMT = errors.New("Can not call method to use MerkleTree in a StateDB without MerkleTree")

// ErrAccountAlreadyExists is used when CreateAccount is called and the Account already exists
var ErrAccountAlreadyExists = errors.New("Can not CreateAccount because Account already exists")

// KEYCURRENTBATCH is used as key in the db to store the current BatchNum
var KEYCURRENTBATCH = []byte("currentbatch")

// PATHSTATEDB defines the subpath of the StateDB
const PATHSTATEDB = "/statedb"
const PATHBATCHNUM = "/BatchNum"
const PATHCURRENT = "/current"

// StateDB represents the StateDB object
type StateDB struct {
	path         string
	currentBatch common.BatchNum
	db           *pebble.PebbleStorage
	mt           *merkletree.MerkleTree
	// idx holds the current Idx that the BatchBuilder is using
	idx common.Idx
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage
func NewStateDB(path string, withMT bool, nLevels int) (*StateDB, error) {
	var sto *pebble.PebbleStorage
	var err error
	sto, err = pebble.NewPebbleStorage(path+PATHSTATEDB+PATHCURRENT, false)
	if err != nil {
		return nil, err
	}

	var mt *merkletree.MerkleTree = nil
	if withMT {
		mt, err = merkletree.NewMerkleTree(sto, nLevels)
		if err != nil {
			return nil, err
		}
	}

	sdb := &StateDB{
		path: path + PATHSTATEDB,
		db:   sto,
		mt:   mt,
	}

	// load currentBatch
	sdb.currentBatch, err = sdb.GetCurrentBatch()
	if err != nil {
		return nil, err
	}

	return sdb, nil
}

// DB returns the *pebble.PebbleStorage from the StateDB
func (s *StateDB) DB() *pebble.PebbleStorage {
	return s.db
}

// GetCurrentBatch returns the current BatchNum stored in the StateDB
func (s *StateDB) GetCurrentBatch() (common.BatchNum, error) {
	cbBytes, err := s.db.Get(KEYCURRENTBATCH)
	if err == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return common.BatchNumFromBytes(cbBytes)
}

// setCurrentBatch stores the current BatchNum in the StateDB
func (s *StateDB) setCurrentBatch() error {
	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}
	tx.Put(KEYCURRENTBATCH, s.currentBatch.Bytes())
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path. Internally this advances & stores the current BatchNum, and then stores a Checkpoint of the current state of the StateDB.
func (s *StateDB) MakeCheckpoint() error {
	// advance currentBatch
	s.currentBatch++

	checkpointPath := s.path + PATHBATCHNUM + strconv.Itoa(int(s.currentBatch))

	err := s.setCurrentBatch()
	if err != nil {
		return err
	}

	// if checkpoint BatchNum already exist in disk, delete it
	if _, err := os.Stat(checkpointPath); !os.IsNotExist(err) {
		err := os.RemoveAll(checkpointPath)
		if err != nil {
			return err
		}
	}

	// execute Checkpoint
	err = s.db.Pebble().Checkpoint(checkpointPath)
	if err != nil {
		return err
	}

	return nil
}

// DeleteCheckpoint removes if exist the checkpoint of the given batchNum
func (s *StateDB) DeleteCheckpoint(batchNum common.BatchNum) error {
	checkpointPath := s.path + PATHBATCHNUM + strconv.Itoa(int(batchNum))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return fmt.Errorf("Checkpoint with batchNum %d does not exist in DB", batchNum)
	}

	return os.RemoveAll(checkpointPath)
}

// Reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.
func (s *StateDB) Reset(batchNum common.BatchNum) error {
	if batchNum == 0 {
		s.idx = 0
		return nil
	}

	checkpointPath := s.path + PATHBATCHNUM + strconv.Itoa(int(batchNum))
	currentPath := s.path + PATHCURRENT

	// remove 'current'
	err := os.RemoveAll(currentPath)
	if err != nil {
		return err
	}
	// copy 'BatchNumX' to 'current'
	cmd := exec.Command("cp", "-r", checkpointPath, currentPath)
	err = cmd.Run()
	if err != nil {
		return err
	}

	// open the new 'current'
	sto, err := pebble.NewPebbleStorage(currentPath, false)
	if err != nil {
		return err
	}
	s.db = sto

	// get currentBatch num
	s.currentBatch, err = s.GetCurrentBatch()
	if err != nil {
		return err
	}
	// idx is obtained from the statedb reset
	s.idx, err = s.getIdx()
	if err != nil {
		return err
	}

	// open the MT for the current s.db
	mt, err := merkletree.NewMerkleTree(s.db, s.mt.MaxLevels())
	if err != nil {
		return err
	}
	s.mt = mt

	return nil
}

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	return getAccountInTreeDB(s.db, idx)
}

// getAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetAccount returns the account for the given Idx
func getAccountInTreeDB(sto db.Storage, idx common.Idx) (*common.Account, error) {
	vBytes, err := sto.Get(idx.Bytes())
	if err != nil {
		return nil, err
	}
	accBytes, err := sto.Get(vBytes)
	if err != nil {
		return nil, err
	}
	var b [32 * common.NLEAFELEMS]byte
	copy(b[:], accBytes)
	return common.AccountFromBytes(b)
}

// CreateAccount creates a new Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	return createAccountInTreeDB(s.db, s.mt, idx, account)
}

// createAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Creates a new Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func createAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return nil, err
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, err
	}

	// store the Leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, err
	}

	_, err = tx.Get(idx.Bytes())
	if err != db.ErrNotFound {
		return nil, ErrAccountAlreadyExists
	}

	tx.Put(v.Bytes(), accountBytes[:])
	tx.Put(idx.Bytes(), v.Bytes())

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if mt != nil {
		return mt.AddAndGetCircomProof(idx.BigInt(), v)
	}

	return nil, nil
}

// UpdateAccount updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) UpdateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	return updateAccountInTreeDB(s.db, s.mt, idx, account)
}

// updateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func updateAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: v, and value: account.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return nil, err
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, err
	}

	tx, err := sto.NewTx()
	if err != nil {
		return nil, err
	}
	tx.Put(v.Bytes(), accountBytes[:])
	tx.Put(idx.Bytes(), v.Bytes())

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	if mt != nil {
		return mt.Update(idx.BigInt(), v)
	}
	return nil, nil
}

// MTGetProof returns the CircomVerifierProof for a given Idx
func (s *StateDB) MTGetProof(idx common.Idx) (*merkletree.CircomVerifierProof, error) {
	if s.mt == nil {
		return nil, ErrStateDBWithoutMT
	}
	return s.mt.GenerateCircomVerifierProof(idx.BigInt(), s.mt.Root())
}

// LocalStateDB represents the local StateDB which allows to make copies from
// the synchronizer StateDB, and is used by the tx-selector and the
// batch-builder. LocalStateDB is an in-memory storage.
type LocalStateDB struct {
	*StateDB
	synchronizerStateDB *StateDB
}

// NewLocalStateDB returns a new LocalStateDB connected to the given
// synchronizerDB
func NewLocalStateDB(path string, synchronizerDB *StateDB, withMT bool, nLevels int) (*LocalStateDB, error) {
	s, err := NewStateDB(path, withMT, nLevels)
	if err != nil {
		return nil, err
	}
	return &LocalStateDB{
		s,
		synchronizerDB,
	}, nil
}

// Reset performs a reset in the LocaStateDB. If fromSynchronizer is true, it
// gets the state from LocalStateDB.synchronizerStateDB for the given batchNum. If fromSynchronizer is false, get the state from LocalStateDB checkpoints.
func (l *LocalStateDB) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	if batchNum == 0 {
		l.idx = 0
		return nil
	}

	synchronizerCheckpointPath := l.synchronizerStateDB.path + PATHBATCHNUM + strconv.Itoa(int(batchNum))
	checkpointPath := l.path + PATHBATCHNUM + strconv.Itoa(int(batchNum))
	currentPath := l.path + PATHCURRENT

	if fromSynchronizer {
		// use checkpoint from SynchronizerStateDB
		if _, err := os.Stat(synchronizerCheckpointPath); os.IsNotExist(err) {
			// if synchronizerStateDB does not have checkpoint at batchNum, return err
			return fmt.Errorf("Checkpoint not exist in Synchronizer")
		}

		// remove 'current'
		err := os.RemoveAll(currentPath)
		if err != nil {
			return err
		}
		// copy synchronizer'BatchNumX' to 'current'
		cmd := exec.Command("cp", "-r", synchronizerCheckpointPath, currentPath)
		err = cmd.Run()
		if err != nil {
			return err
		}
		// copy synchronizer-'BatchNumX' to 'BatchNumX'
		cmd = exec.Command("cp", "-r", synchronizerCheckpointPath, checkpointPath)
		err = cmd.Run()
		if err != nil {
			return err
		}

		// open the new 'current'
		sto, err := pebble.NewPebbleStorage(currentPath, false)
		if err != nil {
			return err
		}
		l.db = sto

		// get currentBatch num
		l.currentBatch, err = l.GetCurrentBatch()
		if err != nil {
			return err
		}
		// open the MT for the current s.db
		mt, err := merkletree.NewMerkleTree(l.db, l.mt.MaxLevels())
		if err != nil {
			return err
		}
		l.mt = mt

		return nil
	}
	// use checkpoint from LocalStateDB
	return l.StateDB.Reset(batchNum)
}
