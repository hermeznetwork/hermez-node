package historydb

import (
	"fmt"

	"github.com/gobuffalo/packr/v2"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver for postgres DB
	migrate "github.com/rubenv/sql-migrate"
	"github.com/russross/meddler"
)

// HistoryDB persist the historic of the rollup
type HistoryDB struct {
	db *sqlx.DB
}

// NewHistoryDB initialize the DB
func NewHistoryDB(port int, host, user, password, dbname string) (*HistoryDB, error) {
	// Connect to DB
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	hdb, err := sqlx.Connect("postgres", psqlconn)
	if err != nil {
		return nil, err
	}
	// Init meddler
	db.InitMeddler()
	meddler.Default = meddler.PostgreSQL

	// Run DB migrations
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("history-migrations", "./migrations"),
	}
	if _, err := migrate.Exec(hdb.DB, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}

	return &HistoryDB{hdb}, nil
}

// addBlocks insert blocks into the DB
func (hdb *HistoryDB) addBlocks(blocks []common.Block) error {
	return db.BulkInsert(
		hdb.db,
		"INSERT INTO block (eth_block_num, timestamp, hash) VALUES %s",
		blocks[:],
	)
}

// GetBlocks retrrieve blocks from the DB
func (hdb *HistoryDB) GetBlocks(from, to uint64) ([]*common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2",
		from, to,
	)
	return blocks, err
}

// reorg deletes all the information that was added into the DB after the lastValidBlock
// WARNING: this is a draaft of the function, useful at the moment for tests
func (hdb *HistoryDB) reorg(lastValidBlock uint64) error {
	_, err := hdb.db.Exec("DELETE FROM block WHERE eth_block_num > $1;", lastValidBlock)
	return err
}

// addBids insert Bids into the DB
func (hdb *HistoryDB) addBids(bids []common.Bid) error {
	// TODO: check the coordinator info
	return db.BulkInsert(
		hdb.db,
		"INSERT INTO bid (slot_num, forger_addr, bid_value, eth_block_num) VALUES %s",
		bids[:],
	)
}

// GetBidsByBlock return the bids done between the block from and to
func (hdb *HistoryDB) GetBidsByBlock(from, to uint64) ([]*common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		"SELECT * FROM bid WHERE $1 <= eth_block_num AND eth_block_num < $2",
		from, to,
	)
	return bids, err
}

// Close frees the resources used by HistoryDB
func (hdb *HistoryDB) Close() error {
	return hdb.db.Close()
}
