package historydb

import (
	"database/sql"
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

// BlockData contains the information of a Block
type BlockData struct {
	block *common.Block
	// Rollup
	L1Txs            []*common.L1Tx
	Batches          []*BatchData
	RegisteredTokens []*common.Token
	RollupVars       *common.RollupVars
	// Auction
	Bids         []*common.Bid
	Coordinators []*common.Coordinator
	AuctionVars  *common.AuctionVars
	// WithdrawalDelayer
	// TODO: enable when common.WithdrawalDelayerVars is Merged from Synchronizer PR
	// WithdrawalDelayerVars *common.WithdrawalDelayerVars
}

// BatchData contains the information of a Batch
type BatchData struct {
	L1UserTxs          []*common.L1Tx
	L1CoordinatorTxs   []*common.L1Tx
	L2Txs              []*common.L2Tx
	RegisteredAccounts []*common.Account
	ExitTree           []*common.ExitInfo
	Batch              *common.Batch
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
func (hdb *HistoryDB) AddBlock(txn *sql.Tx, block *common.Block) error {
	_, err := txn.Exec("INSERT INTO block (eth_block_num, timestamp, hash) VALUES ($1, $2, $3)", block.EthBlockNum, block.Timestamp, block.Hash)
	return err
}

// AddBlocks inserts blocks into the DB
func (hdb *HistoryDB) AddBlocks(txn *sql.Tx, blocks []*common.Block) error {
	return db.BulkInsert(
		txn,
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

// AddBatch insert a Batch into the DB
func (hdb *HistoryDB) AddBatch(txn *sql.Tx, batch *common.Batch) error {
	_, err := txn.Exec(`INSERT INTO batch (
						batch_num,
						eth_block_num,
						forger_addr,
						fees_collected,
						state_root,
						num_accounts,
						exit_root,
						forge_l1_txs_num,
						slot_num) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		batch.BatchNum, batch.EthBlockNum, batch.ForgerAddr, batch.CollectedFees, batch.StateRoot, batch.NumAccounts, batch.ExitRoot, batch.ForgeL1TxsNum, batch.SlotNum)
	return err
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(txn *sql.Tx, batches []*common.Batch) error {
	return db.BulkInsert(
		txn,
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
func (hdb *HistoryDB) Reorg(lastValidBlock int64) error {
	_, err := hdb.db.Exec("DELETE FROM block WHERE eth_block_num > $1;", lastValidBlock)
	return err
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

// AddBids insert Bids into the DB
func (hdb *HistoryDB) AddBids(txn *sql.Tx, bids []*common.Bid) error {
	// TODO: check the coordinator info
	return db.BulkInsert(
		txn,
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

// AddCoordinators insert Coordinators into the DB
func (hdb *HistoryDB) AddCoordinators(txn *sql.Tx, coordinators []*common.Coordinator) error {
	// TODO: check the coordinator info
	return db.BulkInsert(
		txn,
		"INSERT INTO coordianator (forger_addr, eth_block_num, withdraw_addr, url) VALUES %s", // TODO: Correct table name typo when merged
		coordinators[:],
	)
}

// AddExitTree insert Exit tree into the DB
func (hdb *HistoryDB) AddExitTree(txn *sql.Tx, exitTree []*common.ExitInfo) error {
	// TODO: check the coordinator info
	return db.BulkInsert(
		txn,
		"INSERT INTO exit_tree (batch_num, account_idx, withdrawn, merkle_proof, balance, nulifier) VALUES %s",
		exitTree[:],
	)
}

// AddToken insert a token into the DB
func (hdb *HistoryDB) AddToken(token *common.Token) error {
	return meddler.Insert(hdb.db, "token", token)
}

// AddTokens insert tokens into the DB
func (hdb *HistoryDB) AddTokens(txn *sql.Tx, tokens []*common.Token) error {
	return db.BulkInsert(
		txn,
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
func (hdb *HistoryDB) UpdateTokenValue(tokenID common.TokenID, value float32) error {
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
func (hdb *HistoryDB) AddAccounts(txn *sql.Tx, accounts []*common.Account) error {
	return db.BulkInsert(
		txn,
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
func (hdb *HistoryDB) AddL1Txs(txn *sql.Tx, l1txs []*common.L1Tx) error {
	txs := []*common.Tx{}
	for _, tx := range l1txs {
		txs = append(txs, tx.Tx())
	}
	return hdb.AddTxs(txn, txs)
}

// UpdateL1TxsBatchNum update L1 txs from the DB
func (hdb *HistoryDB) UpdateL1TxsBatchNum(txn *sql.Tx, l1txs []*common.L1Tx) error {
	txs := []*common.Tx{}
	for _, tx := range l1txs {
		txs = append(txs, tx.Tx())
	}
	return hdb.UpdateTxsBatchNum(txn, txs)
}

// AddL2Txs inserts L2 txs to the DB
func (hdb *HistoryDB) AddL2Txs(txn *sql.Tx, l2txs []*common.L2Tx) error {
	txs := []*common.Tx{}
	for _, tx := range l2txs {
		txs = append(txs, tx.Tx())
	}
	return hdb.AddTxs(txn, txs)
}

// AddTxs insert L1 txs into the DB
func (hdb *HistoryDB) AddTxs(txn *sql.Tx, txs []*common.Tx) error {
	return db.BulkInsert(
		txn,
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

// UpdateTxsBatchNum update L1 txs batch num in the DB
func (hdb *HistoryDB) UpdateTxsBatchNum(txn *sql.Tx, txs []*common.Tx) error {
	for _, tx := range txs {
		_, err := txn.Exec("UPDATE tx SET batch_num = $1 WHERE to_forge_l1_txs_num = $2 and from_idx = 0", tx.BatchNum, tx.ToForgeL1TxsNum)
		if err != nil {
			return err
		}
	}

	return nil
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

// GetUserTxsToAddAccount gets L1 User Txs to be forged in a batch that will create an account
func (hdb *HistoryDB) GetUserTxsToAddAccount(toForgeL1TxsNum uint32) ([]*common.Tx, error) {
	var txs []*common.Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		"SELECT * FROM tx WHERE to_forge_l1_txs_num = $1 AND from_idx = 0",
		toForgeL1TxsNum,
	)
	return txs, err
}

// GetLastTxsPosition for a given to_forge_l1_txs_num
func (hdb *HistoryDB) GetLastTxsPosition(toForgeL1TxsNum uint32) (int, error) {
	row := hdb.db.QueryRow("SELECT MAX(position) FROM tx WHERE to_forge_l1_txs_num = $1", toForgeL1TxsNum)
	var lastL1TxsPosition int
	return lastL1TxsPosition, row.Scan(&lastL1TxsPosition)
}

// AddBlockSCData stores all the information of a block retrieved by the Synchronizer
func (hdb *HistoryDB) AddBlockSCData(blockData *BlockData) error {
	txn, err := hdb.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// Rollback the transaction after the function returns.
		// If the transaction was already committed, this will do nothing.
		_ = txn.Rollback()
	}()

	// Add block
	err = hdb.AddBlock(txn, blockData.block)
	if err != nil {
		return err
	}

	// Add l1 Txs
	err = hdb.AddL1Txs(txn, blockData.L1Txs)
	if err != nil {
		return err
	}

	// Add Tokens
	err = hdb.AddTokens(txn, blockData.RegisteredTokens)
	if err != nil {
		return err
	}

	// Add Bids
	err = hdb.AddBids(txn, blockData.Bids)
	if err != nil {
		return err
	}

	// Add Coordinators
	err = hdb.AddCoordinators(txn, blockData.Coordinators)
	if err != nil {
		return err
	}

	// Add Batches
	for _, batch := range blockData.Batches {
		// Update l1 Txs
		err = hdb.UpdateL1TxsBatchNum(txn, batch.L1UserTxs)
		if err != nil {
			return err
		}
		err = hdb.AddL1Txs(txn, batch.L1CoordinatorTxs)
		if err != nil {
			return err
		}

		// Add l2 Txs
		err = hdb.AddL2Txs(txn, batch.L2Txs)
		if err != nil {
			return err
		}

		// Add accounts
		err = hdb.AddAccounts(txn, batch.RegisteredAccounts)
		if err != nil {
			return err
		}

		// Add exit tree
		err = hdb.AddExitTree(txn, batch.ExitTree)
		if err != nil {
			return err
		}

		// Add Batch
		err = hdb.AddBatch(txn, batch.Batch)
		if err != nil {
			return err
		}

		// TODO: INSERT CONTRACTS VARS
	}

	return txn.Commit()
}

// Close frees the resources used by HistoryDB
func (hdb *HistoryDB) Close() error {
	return hdb.db.Close()
}
