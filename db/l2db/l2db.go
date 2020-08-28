package l2db

import (
	"fmt"
	"strconv"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/gobuffalo/packr/v2"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

// L2DB stores L2 txs and authorization registers received by the coordinator and keeps them until they are no longer relevant
// due to them being forged or invalid after a safety period
type L2DB struct {
	db           *sqlx.DB
	safetyPeriod uint16
	ttl          time.Duration
	maxTxs       uint32
}

// NewL2DB creates a L2DB.
// More info on how to set dbDialect and dbArgs here: http://gorm.io/docs/connecting_to_the_database.html
// safetyPeriod is the amount of blockchain blocks that must be waited before deleting anything (to avoid reorg problems).
// maxTxs indicates the desired maximum amount of txs stored on the L2DB.
// TTL indicates the maximum amount of time that a tx can be in the L2DB
// (to prevent tx that won't ever be forged to stay there, will be used if maxTxs is exceeded).
// autoPurgePeriod will be used as delay between calls to Purge. If the value is 0, it will be disabled.
func NewL2DB(
	port int, host, user, password, dbname string,
	safetyPeriod uint16,
	maxTxs uint32,
	TTL time.Duration,
) (*L2DB, error) {
	// init meddler
	db.InitMeddler()
	meddler.Default = meddler.PostgreSQL
	// Stablish DB connection
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sqlx.Connect("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	// Run DB migrations
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("history-migrations", "./migrations"),
	}
	if _, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}

	return &L2DB{
		db:           db,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
	}, nil
}

// AddTx inserts a tx into the L2DB
func (l2db *L2DB) AddTx(tx *common.PoolL2Tx) error {
	return meddler.Insert(l2db.db, "tx_pool", tx)
}

// AddAccountCreationAuth inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuth(auth *common.AccountCreationAuth) error {
	// TODO: impl
	return nil
}

// GetTx return the specified Tx
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	tx := new(common.PoolL2Tx)
	return tx, meddler.QueryRow(
		l2db.db, tx,
		"SELECT * FROM tx_pool WHERE tx_id = $1;",
		txID,
	)
}

// GetPendingTxs return all the pending txs of the L2DB
func (l2db *L2DB) GetPendingTxs() ([]*common.PoolL2Tx, error) {
	var txs []*common.PoolL2Tx
	err := meddler.QueryAll(
		l2db.db, &txs,
		"SELECT * FROM tx_pool WHERE state = $1",
		common.PoolL2TxStatePending,
	)
	return txs, err
}

// GetAccountCreationAuth return the authorization to make registers of an Ethereum address
func (l2db *L2DB) GetAccountCreationAuth(ethAddr eth.Address) (*common.AccountCreationAuth, error) {
	// TODO: impl
	return nil, nil
}

// StartForging updates the state of the transactions that will begin the forging process.
// The state of the txs referenced by txIDs will be changed from Pending -> Forging
func (l2db *L2DB) StartForging(txIDs []common.TxID, batchNum common.BatchNum) error {
	query, args, err := sqlx.In(
		`UPDATE tx_pool
		SET state = ?, batch_num = ?
		WHERE state = ? AND tx_id IN (?);`,
		string(common.PoolL2TxStateForging),
		strconv.Itoa(int(batchNum)),
		string(common.PoolL2TxStatePending),
		txIDs,
	)
	if err != nil {
		return err
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return err
}

// DoneForging updates the state of the transactions that have been forged
// so the state of the txs referenced by txIDs will be changed from Forging -> Forged
func (l2db *L2DB) DoneForging(txIDs []common.TxID) error {
	// TODO: impl
	return nil
}

// InvalidateTxs updates the state of the transactions that are invalid.
// The state of the txs referenced by txIDs will be changed from * -> Invalid
func (l2db *L2DB) InvalidateTxs(txIDs []common.TxID) error {
	return nil
}

// CheckNonces invalidate txs with nonces that are smaller than their respective accounts nonces.
// The state of the affected txs will be changed from Pending -> Invalid
func (l2db *L2DB) CheckNonces(updatedAccounts []common.Account) error {
	// TODO: impl
	return nil
}

// GetTxsByAbsoluteFeeUpdate return the txs that have an AbsoluteFee updated before olderThan
func (l2db *L2DB) GetTxsByAbsoluteFeeUpdate(olderThan time.Time) ([]*common.PoolL2Tx, error) {
	// TODO: impl
	return nil, nil
}

// UpdateTxs update existing txs from the pool (TxID must exist)
func (l2db *L2DB) UpdateTxs(txs []*common.PoolL2Tx) error {
	// TODO: impl
	return nil
}

// Reorg updates the state of txs that were updated in a batch that has been discarted due to a blockchain reorg.
// The state of the affected txs can change form Forged -> Pending or from Invalid -> Pending
func (l2db *L2DB) Reorg(lastValidBatch common.BatchNum) error {
	// TODO: impl
	return nil
}

// Purge deletes transactions that have been forged or marked as invalid for longer than the safety period
// it also deletes txs that has been in the L2DB for longer than the ttl if maxTxs has been exceeded
func (l2db *L2DB) Purge() error {
	// TODO: impl
	return nil
}

// Close frees the resources used by the L2DB
func (l2db *L2DB) Close() error {
	return l2db.db.Close()
}
