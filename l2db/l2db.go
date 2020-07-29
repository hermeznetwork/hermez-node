package l2db

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/jinzhu/gorm"
)

// L2DB stores L2 txs and authorization registers received by the coordinator and keeps them until they are no longer relevant
// due to them being forged or invalid after a safety period
type L2DB struct {
	db           *gorm.DB
	safetyPeriod uint16
	ttl          time.Duration
	maxTxs       uint32
}

// NewL2DB creates a L2DB.
// More info on how to set dbDialect and dbArgs here: http://gorm.io/docs/connecting_to_the_database.html
// safetyPeriod is the ammount of blockchain blocks that must be waited before deleting anything (to avoid reorg problems).
// maxTxs indicates the desired maximum amount of txs stored on the L2DB.
// TTL indicates the maximum amount of time that a tx can be in the L2DB
// (to prevent tx that won't ever be forged to stay there, will be used if maxTxs is exceeded).
// autoPurgePeriod will be used as delay between calls to Purge. If the value is 0, it will be disabled.
func NewL2DB(
	dbDialect, dbArgs string,
	safetyPeriod uint16,
	maxTxs uint32,
	TTL time.Duration,
) (*L2DB, error) {
	// Stablish DB connection
	db, err := gorm.Open(dbDialect, dbArgs)
	if err != nil {
		return nil, err
	}

	// Create or update SQL schemas
	// WARNING: AutoMigrate will ONLY create tables, missing columns and missing indexes,
	// and WON’T change existing column’s type or delete unused columns to protect your data.
	// more info: http://gorm.io/docs/migration.html
	db.AutoMigrate(&common.PoolL2Tx{})
	// TODO: db.AutoMigrate(&common.RegisterAuthorization{})

	return &L2DB{
		db:           db,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
	}, nil
}

// AddTx inserts a tx into the L2DB
func (l2db *L2DB) AddTx(tx *common.PoolL2Tx) error {
	return nil
}

// AddRegisterAuthorization inserts a register authorization into the DB
func (l2db *L2DB) AddRegisterAuthorization() error { // TODO: AddRegisterAuthorization(auth &common.RegisterAuthorization)
	return nil
}

// GetTx return the specified Tx
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	return nil, nil
}

// GetPendingTxs return all the pending txs of the L2DB
func (l2db *L2DB) GetPendingTxs() ([]common.PoolL2Tx, error) {
	return nil, nil
}

// GetRegisterAuthorization return the authorization to make registers of an Etherum address
func (l2db *L2DB) GetRegisterAuthorization(ethAddr eth.Address) (int, error) { // TODO: int will be changed to *common.RegisterAuthorization
	return 0, nil
}

// StartForging updates the state of the transactions that will begin the forging process.
// The state of the txs referenced by txIDs will be changed from Pending -> Forging
func (l2db *L2DB) StartForging(txIDs []common.TxID) error {
	return nil
}

// DoneForging updates the state of the transactions that have been forged
// so the state of the txs referenced by txIDs will be changed from Forging -> Forged
func (l2db *L2DB) DoneForging(txIDs []common.TxID) error {
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
	return nil
}

// Reorg updates the state of txs that were updated in a batch that has been discarted due to a blockchian reorg.
// The state of the affected txs can change form Forged -> Pending or from Invalid -> Pending
func (l2db *L2DB) Reorg(lastValidBatch common.BatchNum) error {
	return nil
}

// Purge deletes transactions that have been forged or marked as invalid for longer than the safety period
// it also deletes txs that has been in the L2DB for longer than the ttl if maxTxs has been exceeded
func (l2db *L2DB) Purge() error {
	return nil
}

// Close frees the resources used by the L2DB
func (l2db *L2DB) Close() error {
	return l2db.db.Close()
}
