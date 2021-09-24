package statedb

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/hermeznetwork/hermez-node/common"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	merkletree "github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/sql"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
	"github.com/jmoiron/sqlx"
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
)

const (
	merkleTreeId = 1
	// TypeSynchronizer defines a StateDB used by the Synchronizer, that
	// generates the ExitTree when processing the txs
	TypeSynchronizer = "synchronizer"
	// TypeTxSelector defines a StateDB used by the TxSelector, without
	// computing ExitTree neither the ZKInputs
	TypeTxSelector = "txselector"
	// TypeBatchBuilder defines a StateDB used by the BatchBuilder, that
	// generates the ExitTree and the ZKInput when processing the txs
	TypeBatchBuilder = "batchbuilder"
	// MaxNLevels is the maximum value of NLevels for the merkle tree,
	// which comes from the fact that AccountIdx has 48 bits.
	MaxNLevels = 48
)

// TypeStateDB determines the type of StateDB
type TypeStateDB string

// Config of the StateDB
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	// Type of StateDB (
	Type TypeStateDB
	// NLevels is the number of merkle tree levels in case the Type uses a
	// merkle tree.  If the Type doesn't use a merkle tree, NLevels should
	// be 0.
	NLevels int
}

// StateDB represents the StateDB object
type StateDB struct {
	cfg Config
	db  *sqlx.DB
	sto merkletree.Storage
	mt  *merkletree.MerkleTree
}

// Last offers a subset of view methods of the StateDB that can be
// called via the LastRead method of StateDB in a thread-safe manner to obtain
// a consistent view to the last batch of the StateDB.
type Last struct {
	db merkletree.Storage
}

// GetAccount returns the account for the given Idx
func (s *Last) GetAccount(idx common.Idx) (*common.Account, error) {
	return GetAccountInTreeDB(s.db, idx)
}

// GetCurrentBatch returns the current BatchNum stored in Last.db
func (s *Last) GetCurrentBatch() (common.BatchNum, error) {
	cbBytes, err := s.dba.Get(kvdba.KeyCurrentBatch)
	if tracerr.Unwrap(err) == dba.ErrNotFound {
		return 0, nil
	} else if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return common.BatchNumFromBytes(cbBytes)
}

// DB returns the underlying storage of Last
func (s *Last) DB() merkletree.Storage {
	return s.db
}

// GetAccounts returns all the accounts in the db.  Use for debugging pruposes
// only.
func (s *Last) GetAccounts() ([]common.Account, error) {
	return getAccounts(s.db)
}

// NewStateDB creates a new StateDB, allowing to use an in-memory or in-disk
// storage.  Checkpoints older than the value defined by `keep` will be
// deleted.
// func NewStateDB(pathDB string, keep int, typ TypeStateDB, nLevels int) (*StateDB, error) {
func NewStateDB(cfg Config) (*StateDB, error) {
	var err error
	db, err := dbUtils.InitSQLDB(
		cfg.Port,
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Database,
	)
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
	}

	sqlStorage, err := sql.NewSqlStorage(db, merkleTreeId)
	if err != nil {
		return nil, err
	}

	var mt *merkletree.MerkleTree = nil
	if cfg.Type == TypeSynchronizer || cfg.Type == TypeBatchBuilder {
		mt, err = merkletree.NewMerkleTree(sqlStorage, cfg.NLevels)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	if cfg.Type == TypeTxSelector && cfg.NLevels != 0 {
		return nil, tracerr.Wrap(
			fmt.Errorf("invalid StateDB parameters: StateDB type==TypeStateDB can not have nLevels!=0"))
	}
	return &StateDB{
		cfg: cfg,
		db:  db,
		sto: sqlStorage,
		mt:  mt,
	}, nil
}

// Type returns the StateDB configured Type
func (s *StateDB) Type() TypeStateDB {
	return s.cfg.Type
}

// LastRead is a thread-safe method to query the last checkpoint of the StateDB
// via the Last type methods
func (s *StateDB) LastRead(fn func(sdbLast *Last) error) error {
	return s.db.LastRead(
		func(db *pebble.Storage) error {
			return fn(&Last{
				db: db,
			})
		},
	)
}

// LastGetAccount is a thread-safe method to query an account in the last
// checkpoint of the StateDB.
func (s *StateDB) LastGetAccount(idx common.Idx) (*common.Account, error) {
	var account *common.Account
	if err := s.LastRead(func(sdb *Last) error {
		var err error
		account, err = sdb.GetAccount(idx)
		return err
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return account, nil
}

// LastGetCurrentBatch is a thread-safe method to get the current BatchNum in
// the last checkpoint of the StateDB.
func (s *StateDB) LastGetCurrentBatch() (common.BatchNum, error) {
	var batchNum common.BatchNum
	if err := s.LastRead(func(sdb *Last) error {
		var err error
		batchNum, err = sdb.GetCurrentBatch()
		return err
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return batchNum, nil
}

// LastMTGetRoot returns the root of the underlying Merkle Tree in the last
// checkpoint of the StateDB.
func (s *StateDB) LastMTGetRoot() (*big.Int, error) {
	var root *big.Int
	if err := s.LastRead(func(sdb *Last) error {
		mt, err := merkletree.NewMerkleTree(s.sto, s.cfg.NLevels)
		if err != nil {
			return tracerr.Wrap(err)
		}
		root = mt.Root().BigInt()
		return nil
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return root, nil
}

// MakeCheckpoint does a checkpoint at the given batchNum in the defined path.
// Internally this advances & stores the current BatchNum, and then stores a
// Checkpoint of the current state of the StateDB.
func (s *StateDB) MakeCheckpoint() error {
	log.Debugw("Making StateDB checkpoint", "batch", s.CurrentBatch()+1, "type", s.cfg.Type)
	return s.db.MakeCheckpoint()
}

// DeleteOldCheckpoints deletes old checkpoints when there are more than
// `cfg.keep` checkpoints
func (s *StateDB) DeleteOldCheckpoints() error {
	return s.db.DeleteOldCheckpoints()
}

// CurrentBatch returns the current in-memory CurrentBatch of the StateDB.db
func (s *StateDB) CurrentBatch() common.BatchNum {
	return s.db.CurrentBatch
}

// CurrentIdx returns the current in-memory CurrentIdx of the StateDB.db
func (s *StateDB) CurrentIdx() common.Idx {
	return s.db.CurrentIdx
}

// getCurrentBatch returns the current BatchNum stored in the StateDB.db
func (s *StateDB) getCurrentBatch() (common.BatchNum, error) {
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
	log.Debugw("Making StateDB Reset", "batch", batchNum, "type", s.cfg.Type)
	if err := s.db.Reset(batchNum); err != nil {
		return tracerr.Wrap(err)
	}
	if s.mt != nil {
		// open the MT for the current s.db
		mt, err := merkletree.NewMerkleTree(s.db.StorageWithPrefix(PrefixKeyMT), s.mt.MaxLevels())
		if err != nil {
			return tracerr.Wrap(err)
		}
		s.mt = mt
	}
	return nil
}

// GetAccount returns the account for the given Idx
func (s *StateDB) GetAccount(idx common.Idx) (*common.Account, error) {
	return GetAccountInTreeDB(s.db.DB(), idx)
}

func (s *StateDB) accountsIter(db merkletree.Storage, fn func(a *common.Account) (bool, error)) error {
	idxDB := s.db
	if err := idxDB.Iterate(func(k []byte, v []byte) (bool, error) {
		idx, err := common.IdxFromBytes(k)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		acc, err := GetAccountInTreeDB(db, idx)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		ok, err := fn(acc)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		return ok, nil
	}); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func getAccounts(db merkletree.Storage) ([]common.Account, error) {
	accs := []common.Account{}
	if err := accountsIter(
		db,
		func(a *common.Account) (bool, error) {
			accs = append(accs, *a)
			return true, nil
		},
	); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return accs, nil
}

// TestGetAccounts returns all the accounts in the db.  Use only in tests.
// Outside tests getting all the accounts is discouraged because it's an
// expensive operation, but if you must do it, use `LastRead()` method to get a
// thread-safe and consistent view of the stateDB.
func (s *StateDB) TestGetAccounts() ([]common.Account, error) {
	return getAccounts(s.db.DB())
}

// GetAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  GetAccount returns the account for the given Idx
func GetAccountInTreeDB(sto merkletree.Storage, idx common.Idx) (*common.Account, error) {

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
func (s *StateDB) CreateAccount(idx common.Idx, account *common.Account) (
	*merkletree.CircomProcessorProof, error) {
	cpp, err := CreateAccountInTreeDB(s.sto, s.mt, idx, account)
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
func CreateAccountInTreeDB(sto merkletree.Storage, mt *merkletree.MerkleTree, idx common.Idx,
	account *common.Account) (*merkletree.CircomProcessorProof, error) {
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
func (s *StateDB) UpdateAccount(idx common.Idx, account *common.Account) (
	*merkletree.CircomProcessorProof, error) {
	return UpdateAccountInTreeDB(s.db.DB(), s.mt, idx, account)
}

// UpdateAccountInTreeDB is abstracted from StateDB to be used from StateDB and
// from ExitTree.  Updates the Account in the StateDB for the given Idx.  If
// StateDB.mt==nil, MerkleTree is not affected, otherwise updates the
// MerkleTree, returning a CircomProcessorProof.
func UpdateAccountInTreeDB(sto merkletree.Storage, mt *merkletree.MerkleTree, idx common.Idx,
	account *common.Account) (*merkletree.CircomProcessorProof, error) {
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
	p, err := s.mt.GenerateSCVerifierProof(idx.BigInt(), s.mt.Root())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return p, nil
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
func NewLocalStateDB(cfg Config, synchronizerDB *StateDB) (*LocalStateDB, error) {
	cfg.noGapsCheck = true
	cfg.NoLast = true
	s, err := NewStateDB(cfg)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &LocalStateDB{
		s,
		synchronizerDB,
	}, nil
}

// CheckpointExists returns true if the checkpoint exists
func (l *LocalStateDB) CheckpointExists(batchNum common.BatchNum) (bool, error) {
	return l.db.CheckpointExists(batchNum)
}

// Reset performs a reset in the LocalStateDB. If fromSynchronizer is true, it
// gets the state from LocalStateDB.synchronizerStateDB for the given batchNum.
// If fromSynchronizer is false, get the state from LocalStateDB checkpoints.
func (l *LocalStateDB) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	if fromSynchronizer {
		log.Debugw("Making StateDB ResetFromSynchronizer", "batch", batchNum, "type", l.cfg.Type)
		if err := l.db.ResetFromSynchronizer(batchNum, l.synchronizerStateDB.db); err != nil {
			return tracerr.Wrap(err)
		}
		// open the MT for the current s.db
		if l.MT != nil {
			mt, err := merkletree.NewMerkleTree(l.sto, l.mt.MaxLevels())
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
