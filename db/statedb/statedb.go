package statedb

import (
	"errors"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/leveldb"
	"github.com/iden3/go-merkletree/db/memory"
)

// ErrStateDBWithoutMT is used when a method that requires a MerkleTree is called in a StateDB that does not have a MerkleTree defined
var ErrStateDBWithoutMT = errors.New("Can not call method to use MerkleTree in a StateDB without MerkleTree")

// ErrAccountAlreadyExists is used when CreateAccount is called and the Account already exists
var ErrAccountAlreadyExists = errors.New("Can not CreateAccount because Account already exists")

// StateDB represents the StateDB object
type StateDB struct {
	db db.Storage
	mt *merkletree.MerkleTree
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage
func NewStateDB(path string, inDisk bool, withMT bool, nLevels int) (*StateDB, error) {
	var sto db.Storage
	var err error
	if inDisk {
		sto, err = leveldb.NewLevelDbStorage(path, false)
		if err != nil {
			return nil, err
		}
	} else {
		sto = memory.NewMemoryStorage()
	}
	var mt *merkletree.MerkleTree = nil
	if withMT {
		mt, err = merkletree.NewMerkleTree(sto, nLevels)
		if err != nil {
			return nil, err
		}
	}

	return &StateDB{
		db: sto,
		mt: mt,
	}, nil
}

// CheckPointAt does a checkpoint at the given batchNum in the defined path
func (s *StateDB) CheckPointAt(batchNum int, path string) error {
	// TODO

	return nil
}

// Reset resets the StateDB to the checkpoint at the given batchNum
func (s *StateDB) Reset(batchNum int) error {
	// TODO

	return nil
}

// Checkpoints returns a list of the checkpoints (batchNums)
func (s *StateDB) Checkpoints() ([]int, error) {
	// TODO

	//batchnums, err
	return nil, nil
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
func NewLocalStateDB(synchronizerDB *StateDB, withMT bool, nLevels int) (*LocalStateDB, error) {
	s, err := NewStateDB("", false, withMT, nLevels)
	if err != nil {
		return nil, err
	}
	return &LocalStateDB{
		s,
		synchronizerDB,
	}, nil
}

// Reset performs a reset, getting the state from
// LocalStateDB.synchronizerStateDB for the given batchNum
func (l *LocalStateDB) Reset(batchNum int, fromSynchronizer bool) error {
	// TODO

	return nil
}
