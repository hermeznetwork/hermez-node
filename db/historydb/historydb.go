package historydb

import (
	"fmt"

	"github.com/gobuffalo/packr/v2"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
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

// AddBlock insert a block into the DB
func (hdb *HistoryDB) AddBlock(block *common.Block) error {
	return meddler.Insert(hdb.db, "block", block)
}

// GetBlock retrieve a block from the DB, given a block number
func (hdb *HistoryDB) GetBlock(blockNum uint64) (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.db, block,
		"SELECT * FROM block WHERE eth_block_num = $1;", blockNum,
	)
	return block, err
}

// GetBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) GetBlocks(from, to uint64) ([]*common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2",
		from, to,
	)
	return blocks, err
}

// GetLastBlock retrieve the block with the highest block number from the DB
func (hdb *HistoryDB) GetLastBlock() (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.db, block, "SELECT * FROM block ORDER BY eth_block_num DESC LIMIT 1;",
	)
	return block, err
}

// addBatches insert Bids into the DB
func (hdb *HistoryDB) addBatches(batches []common.Batch) error {
	return db.BulkInsert(
		hdb.db,
		`INSERT INTO batch (
			batch_num,
			eth_block_num,
			forger_addr,
			fees_collected,
			state_root,
			num_accounts,
			exit_root,
			forge_l1_txs_num,
			slot_num
		) VALUES %s;`,
		batches[:],
	)
}

// GetBatches retrieve batches from the DB, given a range of batch numbers defined by from and to
func (hdb *HistoryDB) GetBatches(from, to common.BatchNum) ([]*common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.db, &batches,
		"SELECT * FROM batch WHERE $1 <= batch_num AND batch_num < $2",
		from, to,
	)
	return batches, err
}

// GetLastBatchNum returns the BatchNum of the latest forged batch
func (hdb *HistoryDB) GetLastBatchNum() (common.BatchNum, error) {
	row := hdb.db.QueryRow("SELECT batch_num FROM batch ORDER BY batch_num DESC LIMIT 1;")
	var batchNum common.BatchNum
	return batchNum, row.Scan(&batchNum)
}

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB
func (hdb *HistoryDB) GetLastL1TxsNum() (uint32, error) {
	row := hdb.db.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	var lastL1TxsNum uint32
	return lastL1TxsNum, row.Scan(&lastL1TxsNum)
}

// Reorg deletes all the information that was added into the DB after the lastValidBlock
func (hdb *HistoryDB) Reorg(lastValidBlock uint64) error {
	_, err := hdb.db.Exec("DELETE FROM block WHERE eth_block_num > $1;", lastValidBlock)
	return err
}

// SyncRollup stores all the data that can be changed / added on a block in the Rollup SC
func (hdb *HistoryDB) SyncRollup(
	blockNum uint64,
	l1txs []common.L1Tx,
	l2txs []common.L2Tx,
	registeredAccounts []common.Account,
	exitTree common.ExitInfo,
	withdrawals common.ExitInfo,
	registeredTokens []common.Token,
	batches []common.Batch,
	vars *common.RollupVars,
) error {
	// TODO: make all in a single DB commit
	if err := hdb.addBatches(batches); err != nil {
		return err
	}
	return nil
}

// SyncPoD stores all the data that can be changed / added on a block in the PoD SC
func (hdb *HistoryDB) SyncPoD(
	blockNum uint64,
	bids []common.Bid,
	coordinators []common.Coordinator,
	vars *common.AuctionVars,
) error {
	return nil
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

// GetBidsBySlot return the bids for a specific slot
func (hdb *HistoryDB) GetBidsBySlot(slotNum common.SlotNum) ([]*common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		"SELECT * FROM bid WHERE $1 = slot_num;",
		slotNum,
	)
	return bids, err
}

// Close frees the resources used by HistoryDB
func (hdb *HistoryDB) Close() error {
	return hdb.db.Close()
}
