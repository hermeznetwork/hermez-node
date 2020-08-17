package statedb

import (
	"encoding/binary"
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

// STATEDBPATH defines the subpath of the StateDB
const STATEDBPATH = "/statedb"

// StateDB represents the StateDB object
type StateDB struct {
	path         string
	currentBatch uint64
	db           *pebble.PebbleStorage
	mt           *merkletree.MerkleTree
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage
func NewStateDB(path string, withMT bool, nLevels int) (*StateDB, error) {
	var sto *pebble.PebbleStorage
	var err error
	sto, err = pebble.NewPebbleStorage(path+STATEDBPATH+"/current", false)
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
		path: path + STATEDBPATH,
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
func (s *StateDB) GetCurrentBatch() (uint64, error) {
	cbBytes, err := s.db.Get(KEYCURRENTBATCH)
	if err == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	cb := binary.LittleEndian.Uint64(cbBytes[:8])
	return cb, nil
}

// setCurrentBatch stores the current BatchNum in the StateDB
func (s *StateDB) setCurrentBatch() error {
	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}
	var cbBytes [8]byte
	binary.LittleEndian.PutUint64(cbBytes[:], s.currentBatch)
	tx.Put(KEYCURRENTBATCH, cbBytes[:])
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path
func (s *StateDB) MakeCheckpoint() error {
	// advance currentBatch
	s.currentBatch++

	checkpointPath := s.path + "/BatchNum" + strconv.Itoa(int(s.currentBatch))

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
func (s *StateDB) DeleteCheckpoint(batchNum uint64) error {
	checkpointPath := s.path + "/BatchNum" + strconv.Itoa(int(batchNum))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return fmt.Errorf("Checkpoint with batchNum %d does not exist in DB", batchNum)
	}

	return os.RemoveAll(checkpointPath)
}

// Reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.
func (s *StateDB) Reset(batchNum uint64) error {
	checkpointPath := s.path + "/BatchNum" + strconv.Itoa(int(batchNum))
	currentPath := s.path + "/current"

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
	return nil
}

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	vBytes, err := s.db.Get(idx.Bytes())
	if err != nil {
		return nil, err
	}
	accBytes, err := s.db.Get(vBytes)
	if err != nil {
		return nil, err
	}
	var b [32 * common.NLEAFELEMS]byte
	copy(b[:], accBytes)
	return common.AccountFromBytes(b)
}

// CreateAccount creates a new Account in the StateDB for the given Idx.
// MerkleTree is not affected.
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) error {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return err
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return err
	}

	// store the Leaf value
	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}

	_, err = tx.Get(idx.Bytes())
	if err != db.ErrNotFound {
		return ErrAccountAlreadyExists
	}

	tx.Put(v.Bytes(), accountBytes[:])
	tx.Put(idx.Bytes(), v.Bytes())

	return tx.Commit()
}

// UpdateAccount updates the Account in the StateDB for the given Idx.
// MerkleTree is not affected.
func (s *StateDB) UpdateAccount(idx common.Idx, account *common.Account) error {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return err
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return err
	}

	tx, err := s.db.NewTx()
	if err != nil {
		return err
	}
	tx.Put(v.Bytes(), accountBytes[:])
	tx.Put(idx.Bytes(), v.Bytes())

	return tx.Commit()
}

// MTCreateAccount creates a new Account in the StateDB for the given Idx,
// and updates the MerkleTree, returning a CircomProcessorProof
func (s *StateDB) MTCreateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	if s.mt == nil {
		return nil, ErrStateDBWithoutMT
	}
	err := s.CreateAccount(idx, account)
	if err != nil {
		return nil, err
	}

	v, err := account.HashValue() // already computed in s.CreateAccount, next iteration reuse first computation
	if err != nil {
		return nil, err
	}
	// Add k & v into the MT
	return s.mt.AddAndGetCircomProof(idx.BigInt(), v)
}

// MTUpdateAccount updates the Account in the StateDB for the given Idx, and
// updates the MerkleTree, returning a CircomProcessorProof
func (s *StateDB) MTUpdateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	if s.mt == nil {
		return nil, ErrStateDBWithoutMT
	}
	err := s.UpdateAccount(idx, account)
	if err != nil {
		return nil, err
	}

	v, err := account.HashValue() // already computed in s.CreateAccount, next iteration reuse first computation
	if err != nil {
		return nil, err
	}
	// Add k & v into the MT
	return s.mt.Update(idx.BigInt(), v)
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
func (l *LocalStateDB) Reset(batchNum uint64, fromSynchronizer bool) error {

	synchronizerCheckpointPath := l.synchronizerStateDB.path + "/BatchNum" + strconv.Itoa(int(batchNum))
	checkpointPath := l.path + "/BatchNum" + strconv.Itoa(int(batchNum))
	currentPath := l.path + "/current"

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
		return nil
	}
	// use checkpoint from LocalStateDB
	return l.StateDB.Reset(batchNum)
}
