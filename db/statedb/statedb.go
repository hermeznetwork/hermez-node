package statedb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

// TODO(Edu): Document here how StateDB is kept consistent

var (
	// ErrStateDBWithoutMT is used when a method that requires a MerkleTree
	// is called in a StateDB that does not have a MerkleTree defined
	ErrStateDBWithoutMT = errors.New("Can not call method to use MerkleTree in a StateDB without MerkleTree")

	// ErrAccountAlreadyExists is used when CreateAccount is called and the
	// Account already exists
	ErrAccountAlreadyExists = errors.New("Can not CreateAccount because Account already exists")

	// ErrToIdxNotFound is used when trying to get the ToIdx from ToEthAddr
	// or ToEthAddr&ToBJJ
	ErrToIdxNotFound = errors.New("ToIdx can not be found")
	// ErrGetIdxNoCase is used when trying to get the Idx from EthAddr &
	// BJJ with not compatible combination
	ErrGetIdxNoCase = errors.New("Can not get Idx due unexpected combination of ethereum Address & BabyJubJub PublicKey")

	// KeyCurrentBatch is used as key in the db to store the current BatchNum
	KeyCurrentBatch = []byte("k:currentbatch")

	// PrefixKeyIdx is the key prefix for idx in the db
	PrefixKeyIdx = []byte("i:")
	// PrefixKeyAccHash is the key prefix for account hash in the db
	PrefixKeyAccHash = []byte("h:")
	// PrefixKeyMT is the key prefix for merkle tree in the db
	PrefixKeyMT = []byte("m:")
	// PrefixKeyAddr is the key prefix for address in the db
	PrefixKeyAddr = []byte("a:")
	// PrefixKeyAddrBJJ is the key prefix for address-babyjubjub in the db
	PrefixKeyAddrBJJ = []byte("ab:")
)

const (
	// PathBatchNum defines the subpath of the Batch Checkpoint in the
	// subpath of the StateDB
	PathBatchNum = "BatchNum"
	// PathCurrent defines the subpath of the current Batch in the subpath
	// of the StateDB
	PathCurrent = "current"
	// TypeSynchronizer defines a StateDB used by the Synchronizer, that
	// generates the ExitTree when processing the txs
	TypeSynchronizer = "synchronizer"
	// TypeTxSelector defines a StateDB used by the TxSelector, without
	// computing ExitTree neither the ZKInputs
	TypeTxSelector = "txselector"
	// TypeBatchBuilder defines a StateDB used by the BatchBuilder, that
	// generates the ExitTree and the ZKInput when processing the txs
	TypeBatchBuilder = "batchbuilder"
)

// TypeStateDB determines the type of StateDB
type TypeStateDB string

// StateDB represents the StateDB object
type StateDB struct {
	path         string
	currentBatch common.BatchNum
	db           *pebble.PebbleStorage
	mt           *merkletree.MerkleTree
	typ          TypeStateDB
	chainID      uint16
	// idx holds the current Idx that the BatchBuilder is using
	idx common.Idx
	zki *common.ZKInputs
	// i is the current transaction index in the ZKInputs generation (zki)
	i int
	// AccumulatedFees contains the accumulated fees for each token (Coord
	// Idx) in the processed batch
	AccumulatedFees map[common.Idx]*big.Int
	keep            int
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage.  Checkpoints older than the value defined by `keep` will be
// deleted.
func NewStateDB(pathDB string, keep int, typ TypeStateDB, nLevels int, chainID uint16) (*StateDB, error) {
	var sto *pebble.PebbleStorage
	var err error
	sto, err = pebble.NewPebbleStorage(path.Join(pathDB, PathCurrent), false)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var mt *merkletree.MerkleTree = nil
	if typ == TypeSynchronizer || typ == TypeBatchBuilder {
		mt, err = merkletree.NewMerkleTree(sto.WithPrefix(PrefixKeyMT), nLevels)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	if typ == TypeTxSelector && nLevels != 0 {
		return nil, tracerr.Wrap(fmt.Errorf("invalid StateDB parameters: StateDB type==TypeStateDB can not have nLevels!=0"))
	}

	sdb := &StateDB{
		path:    pathDB,
		db:      sto,
		mt:      mt,
		typ:     typ,
		chainID: chainID,
		keep:    keep,
	}

	// load currentBatch
	sdb.currentBatch, err = sdb.GetCurrentBatch()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// make reset (get checkpoint) at currentBatch
	err = sdb.reset(sdb.currentBatch, false)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return sdb, nil
}

// DB returns the *pebble.PebbleStorage from the StateDB
func (s *StateDB) DB() *pebble.PebbleStorage {
	return s.db
}

// GetCurrentBatch returns the current BatchNum stored in the StateDB
func (s *StateDB) GetCurrentBatch() (common.BatchNum, error) {
	cbBytes, err := s.db.Get(KeyCurrentBatch)
	if tracerr.Unwrap(err) == db.ErrNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return common.BatchNumFromBytes(cbBytes)
}

// setCurrentBatch stores the current BatchNum in the StateDB
func (s *StateDB) setCurrentBatch() error {
	tx, err := s.db.NewTx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	err = tx.Put(KeyCurrentBatch, s.currentBatch.Bytes())
	if err != nil {
		return tracerr.Wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path. Internally this advances & stores the current BatchNum, and then stores a Checkpoint of the current state of the StateDB.
func (s *StateDB) MakeCheckpoint() error {
	// advance currentBatch
	s.currentBatch++
	log.Debugw("Making StateDB checkpoint", "batch", s.currentBatch, "type", s.typ)

	checkpointPath := path.Join(s.path, fmt.Sprintf("%s%d", PathBatchNum, s.currentBatch))

	if err := s.setCurrentBatch(); err != nil {
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
	if err := s.db.Pebble().Checkpoint(checkpointPath); err != nil {
		return tracerr.Wrap(err)
	}
	// delete old checkpoints
	if err := s.deleteOldCheckpoints(); err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

// DeleteCheckpoint removes if exist the checkpoint of the given batchNum
func (s *StateDB) DeleteCheckpoint(batchNum common.BatchNum) error {
	checkpointPath := path.Join(s.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))

	if _, err := os.Stat(checkpointPath); os.IsNotExist(err) {
		return tracerr.Wrap(fmt.Errorf("Checkpoint with batchNum %d does not exist in DB", batchNum))
	}

	return os.RemoveAll(checkpointPath)
}

// listCheckpoints returns the list of batchNums of the checkpoints, sorted.
// If there's a gap between the list of checkpoints, an error is returned.
func (s *StateDB) listCheckpoints() ([]int, error) {
	files, err := ioutil.ReadDir(s.path)
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
				return nil, tracerr.Wrap(fmt.Errorf("checkpoint gap at %v", checkpoint))
			}
		}
	}
	return checkpoints, nil
}

// deleteOldCheckpoints deletes old checkpoints when there are more than
// `s.keep` checkpoints
func (s *StateDB) deleteOldCheckpoints() error {
	list, err := s.listCheckpoints()
	if err != nil {
		return tracerr.Wrap(err)
	}
	if len(list) > s.keep {
		for _, checkpoint := range list[:len(list)-s.keep] {
			if err := s.DeleteCheckpoint(common.BatchNum(checkpoint)); err != nil {
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

// Reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.
func (s *StateDB) Reset(batchNum common.BatchNum) error {
	return s.reset(batchNum, true)
}

// reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.  `closeCurrent` will close the
// currently opened db before doing the reset.
func (s *StateDB) reset(batchNum common.BatchNum, closeCurrent bool) error {
	currentPath := path.Join(s.path, PathCurrent)

	if closeCurrent {
		if err := s.db.Pebble().Close(); err != nil {
			return tracerr.Wrap(err)
		}
	}
	// remove 'current'
	err := os.RemoveAll(currentPath)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// remove all checkpoints > batchNum
	for i := batchNum + 1; i <= s.currentBatch; i++ {
		if err := s.DeleteCheckpoint(i); err != nil {
			return tracerr.Wrap(err)
		}
	}
	if batchNum == 0 {
		// if batchNum == 0, open the new fresh 'current'
		sto, err := pebble.NewPebbleStorage(currentPath, false)
		if err != nil {
			return tracerr.Wrap(err)
		}
		s.db = sto
		s.idx = 255
		s.currentBatch = batchNum

		if s.mt != nil {
			// open the MT for the current s.db
			mt, err := merkletree.NewMerkleTree(s.db.WithPrefix(PrefixKeyMT), s.mt.MaxLevels())
			if err != nil {
				return tracerr.Wrap(err)
			}
			s.mt = mt
		}
		return nil
	}

	checkpointPath := path.Join(s.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))
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
	s.db = sto

	// get currentBatch num
	s.currentBatch, err = s.GetCurrentBatch()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// idx is obtained from the statedb reset
	s.idx, err = s.GetIdx()
	if err != nil {
		return tracerr.Wrap(err)
	}

	if s.mt != nil {
		// open the MT for the current s.db
		mt, err := merkletree.NewMerkleTree(s.db.WithPrefix(PrefixKeyMT), s.mt.MaxLevels())
		if err != nil {
			return tracerr.Wrap(err)
		}
		s.mt = mt
	}

	return nil
}

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	return getAccountInTreeDB(s.db, idx)
}

// GetAccounts returns all the accounts in the db.  Use for debugging pruposes
// only.
func (s *StateDB) GetAccounts() ([]common.Account, error) {
	idxDB := s.db.WithPrefix(PrefixKeyIdx)
	idxs := []common.Idx{}
	// NOTE: Current implementation of Iterate in the pebble interface is
	// not efficient, as it iterates over all keys.  Improve it following
	// this example: https://github.com/cockroachdb/pebble/pull/923/files
	if err := idxDB.Iterate(func(k []byte, v []byte) (bool, error) {
		idx, err := common.IdxFromBytes(k)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		idxs = append(idxs, idx)
		return true, nil
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	accs := []common.Account{}
	for i := range idxs {
		acc, err := s.GetAccount(idxs[i])
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		accs = append(accs, *acc)
	}
	return accs, nil
}

// getAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetAccount returns the account for the given Idx
func getAccountInTreeDB(sto db.Storage, idx common.Idx) (*common.Account, error) {
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	vBytes, err := sto.Get(append(PrefixKeyIdx, idxBytes[:]...))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	accBytes, err := sto.Get(append(PrefixKeyAccHash, vBytes...))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var b [32 * common.NLeafElems]byte
	copy(b[:], accBytes)
	account, err := common.AccountFromBytes(b)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	account.Idx = idx
	return account, nil
}

// CreateAccount creates a new Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	cpp, err := createAccountInTreeDB(s.db, s.mt, idx, account)
	if err != nil {
		return cpp, tracerr.Wrap(err)
	}
	// store idx by EthAddr & BJJ
	err = s.setIdxByEthAddrBJJ(idx, account.EthAddr, account.PublicKey, account.TokenID)
	return cpp, tracerr.Wrap(err)
}

// createAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Creates a new Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func createAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	// store at the DB the key: v, and value: leaf.Bytes()
	v, err := account.HashValue()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// store the Leaf value
	tx, err := sto.NewTx()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	_, err = tx.Get(append(PrefixKeyIdx, idxBytes[:]...))
	if tracerr.Unwrap(err) != db.ErrNotFound {
		return nil, tracerr.Wrap(ErrAccountAlreadyExists)
	}

	err = tx.Put(append(PrefixKeyAccHash, v.Bytes()...), accountBytes[:])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyIdx, idxBytes[:]...), v.Bytes())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, tracerr.Wrap(err)
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
		return nil, tracerr.Wrap(err)
	}
	accountBytes, err := account.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	tx, err := sto.NewTx()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyAccHash, v.Bytes()...), accountBytes[:])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	idxBytes, err := idx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	err = tx.Put(append(PrefixKeyIdx, idxBytes[:]...), v.Bytes())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, tracerr.Wrap(err)
	}

	if mt != nil {
		proof, err := mt.Update(idx.BigInt(), v)
		return proof, tracerr.Wrap(err)
	}
	return nil, nil
}

// MTGetProof returns the CircomVerifierProof for a given Idx
func (s *StateDB) MTGetProof(idx common.Idx) (*merkletree.CircomVerifierProof, error) {
	if s.mt == nil {
		return nil, tracerr.Wrap(ErrStateDBWithoutMT)
	}
	return s.mt.GenerateCircomVerifierProof(idx.BigInt(), s.mt.Root())
}

// MTGetRoot returns the current root of the underlying Merkle Tree
func (s *StateDB) MTGetRoot() *big.Int {
	return s.mt.Root().BigInt()
}

// MerkleTree returns the underlying StateDB merkle tree.  It can be nil.
func (s *StateDB) MerkleTree() *merkletree.MerkleTree {
	return s.mt
}

// LocalStateDB represents the local StateDB which allows to make copies from
// the synchronizer StateDB, and is used by the tx-selector and the
// batch-builder. LocalStateDB is an in-memory storage.
type LocalStateDB struct {
	*StateDB
	synchronizerStateDB *StateDB
}

// NewLocalStateDB returns a new LocalStateDB connected to the given
// synchronizerDB.  Checkpoints older than the value defined by `keep` will be
// deleted.
func NewLocalStateDB(path string, keep int, synchronizerDB *StateDB, typ TypeStateDB,
	nLevels int) (*LocalStateDB, error) {
	s, err := NewStateDB(path, keep, typ, nLevels, synchronizerDB.chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
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

	synchronizerCheckpointPath := path.Join(l.synchronizerStateDB.path,
		fmt.Sprintf("%s%d", PathBatchNum, batchNum))
	checkpointPath := path.Join(l.path, fmt.Sprintf("%s%d", PathBatchNum, batchNum))
	currentPath := path.Join(l.path, PathCurrent)

	if fromSynchronizer {
		// use checkpoint from SynchronizerStateDB
		if _, err := os.Stat(synchronizerCheckpointPath); os.IsNotExist(err) {
			// if synchronizerStateDB does not have checkpoint at batchNum, return err
			return tracerr.Wrap(fmt.Errorf("Checkpoint \"%v\" not exist in Synchronizer",
				synchronizerCheckpointPath))
		}

		if err := l.db.Pebble().Close(); err != nil {
			return tracerr.Wrap(err)
		}
		// remove 'current'
		err := os.RemoveAll(currentPath)
		if err != nil {
			return tracerr.Wrap(err)
		}
		// copy synchronizer'BatchNumX' to 'current'
		err = pebbleMakeCheckpoint(synchronizerCheckpointPath, currentPath)
		if err != nil {
			return tracerr.Wrap(err)
		}
		// copy synchronizer'BatchNumX' to 'BatchNumX'
		err = pebbleMakeCheckpoint(synchronizerCheckpointPath, checkpointPath)
		if err != nil {
			return tracerr.Wrap(err)
		}

		// open the new 'current'
		sto, err := pebble.NewPebbleStorage(currentPath, false)
		if err != nil {
			return tracerr.Wrap(err)
		}
		l.db = sto

		// get currentBatch num
		l.currentBatch, err = l.GetCurrentBatch()
		if err != nil {
			return tracerr.Wrap(err)
		}
		// open the MT for the current s.db
		if l.mt != nil {
			mt, err := merkletree.NewMerkleTree(l.db.WithPrefix(PrefixKeyMT), l.mt.MaxLevels())
			if err != nil {
				return tracerr.Wrap(err)
			}
			l.mt = mt
		}

		return nil
	}
	// use checkpoint from LocalStateDB
	return l.StateDB.reset(batchNum, true)
}
