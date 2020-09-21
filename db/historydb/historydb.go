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

// TODO(Edu): Document here how HistoryDB is kept consistent

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

// AddBlocks inserts blocks into the DB
func (hdb *HistoryDB) AddBlocks(blocks []common.Block) error {
	return db.BulkInsert(
		hdb.db,
		`INSERT INTO block (
			eth_block_num,
			timestamp,
			hash
		) VALUES %s;`,
		blocks[:],
	)
}

// GetBlock retrieve a block from the DB, given a block number
func (hdb *HistoryDB) GetBlock(blockNum int64) (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.db, block,
		"SELECT * FROM block WHERE eth_block_num = $1;", blockNum,
	)
	return block, err
}

// GetBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) GetBlocks(from, to int64) ([]*common.Block, error) {
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

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
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

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB.  If there's no
// batch in the DB (nil, nil) is returned.
func (hdb *HistoryDB) GetLastL1TxsNum() (*int64, error) {
	row := hdb.db.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	lastL1TxsNum := new(int64)
	return lastL1TxsNum, row.Scan(&lastL1TxsNum)
}

// Reorg deletes all the information that was added into the DB after the lastValidBlock
func (hdb *HistoryDB) Reorg(lastValidBlock int64) error {
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
	if err := hdb.AddBatches(batches); err != nil {
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

// GetBids return the bids
func (hdb *HistoryDB) GetBids() ([]*common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		"SELECT * FROM bid;",
	)
	return bids, err
}

// AddToken insert a token into the DB
func (hdb *HistoryDB) AddToken(token *common.Token) error {
	return meddler.Insert(hdb.db, "token", token)
}

// AddTokens insert tokens into the DB
func (hdb *HistoryDB) AddTokens(tokens []common.Token) error {
	return db.BulkInsert(
		hdb.db,
		`INSERT INTO token (
			token_id,
			eth_block_num,
			eth_addr,
			name,
			symbol,
			decimals,
			usd,
			usd_update
		) VALUES %s;`,
		tokens[:],
	)
}

// UpdateTokenValue updates the USD value of a token
func (hdb *HistoryDB) UpdateTokenValue(tokenID common.TokenID, value float64) error {
	_, err := hdb.db.Exec(
		"UPDATE token SET usd = $1 WHERE token_id = $2;",
		value, tokenID,
	)
	return err
}

// GetTokens returns a list of tokens from the DB
func (hdb *HistoryDB) GetTokens() ([]*common.Token, error) {
	var tokens []*common.Token
	err := meddler.QueryAll(
		hdb.db, &tokens,
		"SELECT * FROM token ORDER BY token_id;",
	)
	return tokens, err
}

// AddAccounts insert accounts into the DB
func (hdb *HistoryDB) AddAccounts(accounts []common.Account) error {
	return db.BulkInsert(
		hdb.db,
		`INSERT INTO account (
			idx,
			token_id,
			batch_num,
			bjj,
			eth_addr
		) VALUES %s;`,
		accounts[:],
	)
}

// GetAccounts returns a list of accounts from the DB
func (hdb *HistoryDB) GetAccounts() ([]*common.Account, error) {
	var accs []*common.Account
	err := meddler.QueryAll(
		hdb.db, &accs,
		"SELECT * FROM account ORDER BY idx;",
	)
	return accs, err
}

// AddL1Txs inserts L1 txs to the DB
func (hdb *HistoryDB) AddL1Txs(l1txs []common.L1Tx) error {
	txs := []common.Tx{}
	for _, tx := range l1txs {
		txs = append(txs, *tx.Tx())
	}
	return hdb.AddTxs(txs)
}

// AddL2Txs inserts L2 txs to the DB
func (hdb *HistoryDB) AddL2Txs(l2txs []common.L2Tx) error {
	txs := []common.Tx{}
	for _, tx := range l2txs {
		txs = append(txs, *tx.Tx())
	}
	return hdb.AddTxs(txs)
}

// AddTxs insert L1 txs into the DB
func (hdb *HistoryDB) AddTxs(txs []common.Tx) error {
	return db.BulkInsert(
		hdb.db,
		`INSERT INTO tx (
			is_l1,
			id,
			type,
			position,
			from_idx,
			to_idx,
			amount,
			amount_f,
			token_id,
			amount_usd,
			batch_num,
			eth_block_num,
			to_forge_l1_txs_num,
			user_origin,
			from_eth_addr,
			from_bjj,
			load_amount,
			load_amount_f,
			load_amount_usd,
			fee,
			fee_usd,
			nonce
		) VALUES %s;`,
		txs[:],
	)
}

// GetTxs returns a list of txs from the DB
func (hdb *HistoryDB) GetTxs() ([]*common.Tx, error) {
	var txs []*common.Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT * FROM tx 
		ORDER BY (batch_num, position) ASC`,
	)
	return txs, err
}

// GetTx returns a tx from the DB
func (hdb *HistoryDB) GetTx(txID common.TxID) (*common.Tx, error) {
	tx := new(common.Tx)
	return tx, meddler.QueryRow(
		hdb.db, tx,
		"SELECT * FROM tx WHERE id = $1;",
		txID,
	)
}

// Close frees the resources used by HistoryDB
func (hdb *HistoryDB) Close() error {
	return hdb.db.Close()
}
