package statedb

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/kvdb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-merkletree"
	"github.com/iden3/go-merkletree/db"
)

var (
	// ErrStateDBWithoutMT is used when a method that requires a MerkleTree
	// is called in a StateDB that does not have a MerkleTree defined
	ErrStateDBWithoutMT = errors.New("Can not call method to use MerkleTree in a StateDB without MerkleTree")

	// ErrAccountAlreadyExists is used when CreateAccount is called and the
	// Account already exists
	ErrAccountAlreadyExists = errors.New("Can not CreateAccount because Account already exists")

	// ErrIdxNotFound is used when trying to get the Idx from EthAddr or
	// EthAddr&ToBJJ
	ErrIdxNotFound = errors.New("Idx can not be found")
	// ErrGetIdxNoCase is used when trying to get the Idx from EthAddr &
	// BJJ with not compatible combination
	ErrGetIdxNoCase = errors.New("Can not get Idx due unexpected combination of ethereum Address & BabyJubJub PublicKey")

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
	path string
	Typ  TypeStateDB
	db   *kvdb.KVDB
	MT   *merkletree.MerkleTree
	keep int
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage.  Checkpoints older than the value defined by `keep` will be
// deleted.
func NewStateDB(pathDB string, keep int, typ TypeStateDB, nLevels int) (*StateDB, error) {
	var kv *kvdb.KVDB
	var err error

	kv, err = kvdb.NewKVDB(pathDB, keep)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var mt *merkletree.MerkleTree = nil
	if typ == TypeSynchronizer || typ == TypeBatchBuilder {
		mt, err = merkletree.NewMerkleTree(kv.StorageWithPrefix(PrefixKeyMT), nLevels)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	if typ == TypeTxSelector && nLevels != 0 {
		return nil, tracerr.Wrap(fmt.Errorf("invalid StateDB parameters: StateDB type==TypeStateDB can not have nLevels!=0"))
	}

	return &StateDB{
		path: pathDB,
		db:   kv,
		MT:   mt,
		Typ:  typ,
		keep: keep,
	}, nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path.
// Internally this advances & stores the current BatchNum, and then stores a
// Checkpoint of the current state of the StateDB.
func (s *StateDB) MakeCheckpoint() error {
	log.Debugw("Making StateDB checkpoint", "batch", s.CurrentBatch()+1, "type", s.Typ)
	return s.db.MakeCheckpoint()
}

// CurrentBatch returns the current in-memory CurrentBatch of the StateDB.db
func (s *StateDB) CurrentBatch() common.BatchNum {
	return s.db.CurrentBatch
}

// CurrentIdx returns the current in-memory CurrentIdx of the StateDB.db
func (s *StateDB) CurrentIdx() common.Idx {
	return s.db.CurrentIdx
}

// GetCurrentBatch returns the current BatchNum stored in the StateDB.db
func (s *StateDB) GetCurrentBatch() (common.BatchNum, error) {
	return s.db.GetCurrentBatch()
}

// GetCurrentIdx returns the stored Idx from the localStateDB, which is the
// last Idx used for an Account in the localStateDB.
func (s *StateDB) GetCurrentIdx() (common.Idx, error) {
	return s.db.GetCurrentIdx()
}

// SetCurrentIdx stores Idx in the StateDB
func (s *StateDB) SetCurrentIdx(idx common.Idx) error {
	return s.db.SetCurrentIdx(idx)
}

// Reset resets the StateDB to the checkpoint at the given batchNum. Reset
// does not delete the checkpoints between old current and the new current,
// those checkpoints will remain in the storage, and eventually will be
// deleted when MakeCheckpoint overwrites them.
func (s *StateDB) Reset(batchNum common.BatchNum) error {
	err := s.db.Reset(batchNum)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if s.MT != nil {
		// open the MT for the current s.db
		mt, err := merkletree.NewMerkleTree(s.db.StorageWithPrefix(PrefixKeyMT), s.MT.MaxLevels())
		if err != nil {
			return tracerr.Wrap(err)
		}
		s.MT = mt
	}
	log.Debugw("Making StateDB Reset", "batch", batchNum)
	return nil
}

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	return GetAccountInTreeDB(s.db.DB(), idx)
}

// GetAccounts returns all the accounts in the db.  Use for debugging pruposes
// only.
func (s *StateDB) GetAccounts() ([]common.Account, error) {
	idxDB := s.db.StorageWithPrefix(PrefixKeyIdx)
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

// GetAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetAccount returns the account for the given Idx
func GetAccountInTreeDB(sto db.Storage, idx common.Idx) (*common.Account, error) {
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
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
	cpp, err := CreateAccountInTreeDB(s.db.DB(), s.MT, idx, account)
	if err != nil {
		return cpp, tracerr.Wrap(err)
	}
	// store idx by EthAddr & BJJ
	err = s.setIdxByEthAddrBJJ(idx, account.EthAddr, account.BJJ, account.TokenID)
	return cpp, tracerr.Wrap(err)
}

// CreateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Creates a new Account in the StateDB for the given Idx.  If
// StateDB.MT==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func CreateAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
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
	return UpdateAccountInTreeDB(s.db.DB(), s.MT, idx, account)
}

// UpdateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func UpdateAccountInTreeDB(sto db.Storage, mt *merkletree.MerkleTree, idx common.Idx, account *common.Account) (*merkletree.CircomProcessorProof, error) {
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
	if s.MT == nil {
		return nil, tracerr.Wrap(ErrStateDBWithoutMT)
	}
	p, err := s.MT.GenerateSCVerifierProof(idx.BigInt(), s.MT.Root())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return p, nil
}

// MTGetRoot returns the current root of the underlying Merkle Tree
func (s *StateDB) MTGetRoot() *big.Int {
	return s.MT.Root().BigInt()
}

// Close the StateDB
func (s *StateDB) Close() {
	s.db.Close()
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
	s, err := NewStateDB(path, keep, typ, nLevels)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &LocalStateDB{
		s,
		synchronizerDB,
	}, nil
}

// Reset performs a reset in the LocaStateDB. If fromSynchronizer is true, it
// gets the state from LocalStateDB.synchronizerStateDB for the given batchNum.
// If fromSynchronizer is false, get the state from LocalStateDB checkpoints.
func (l *LocalStateDB) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	if fromSynchronizer {
		err := l.db.ResetFromSynchronizer(batchNum, l.synchronizerStateDB.db)
		if err != nil {
			return tracerr.Wrap(err)
		}
		// open the MT for the current s.db
		if l.MT != nil {
			mt, err := merkletree.NewMerkleTree(l.db.StorageWithPrefix(PrefixKeyMT), l.MT.MaxLevels())
			if err != nil {
				return tracerr.Wrap(err)
			}
			l.MT = mt
		}
		return nil
	}
	// use checkpoint from LocalStateDB
	return l.StateDB.Reset(batchNum)
}
