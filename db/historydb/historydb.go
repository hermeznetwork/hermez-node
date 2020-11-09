package historydb

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
	"github.com/russross/meddler"
)

const (
	// OrderAsc indicates ascending order when using pagination
	OrderAsc = "ASC"
	// OrderDesc indicates descending order when using pagination
	OrderDesc = "DESC"
)

// TODO(Edu): Document here how HistoryDB is kept consistent

// HistoryDB persist the historic of the rollup
type HistoryDB struct {
	db *sqlx.DB
}

// NewHistoryDB initialize the DB
func NewHistoryDB(db *sqlx.DB) *HistoryDB {
	return &HistoryDB{db: db}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (hdb *HistoryDB) DB() *sqlx.DB {
	return hdb.db
}

// AddBlock insert a block into the DB
func (hdb *HistoryDB) AddBlock(block *common.Block) error { return hdb.addBlock(hdb.db, block) }
func (hdb *HistoryDB) addBlock(d meddler.DB, block *common.Block) error {
	return meddler.Insert(d, "block", block)
}

// AddBlocks inserts blocks into the DB
func (hdb *HistoryDB) AddBlocks(blocks []common.Block) error {
	return hdb.addBlocks(hdb.db, blocks)
}

func (hdb *HistoryDB) addBlocks(d meddler.DB, blocks []common.Block) error {
	return db.BulkInsert(
		d,
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

// GetAllBlocks retrieve all blocks from the DB
func (hdb *HistoryDB) GetAllBlocks() ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block;",
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), err
}

// GetBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) GetBlocks(from, to int64) ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2;",
		from, to,
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), err
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
func (hdb *HistoryDB) AddBatch(batch *common.Batch) error { return hdb.addBatch(hdb.db, batch) }
func (hdb *HistoryDB) addBatch(d meddler.DB, batch *common.Batch) error {
	// Calculate total collected fees in USD
	// Get IDs of collected tokens for fees
	tokenIDs := []common.TokenID{}
	for id := range batch.CollectedFees {
		tokenIDs = append(tokenIDs, id)
	}
	// Get USD value of the tokens
	type tokenPrice struct {
		ID       common.TokenID `meddler:"token_id"`
		USD      *float64       `meddler:"usd"`
		Decimals int            `meddler:"decimals"`
	}
	var tokenPrices []*tokenPrice
	if len(tokenIDs) > 0 {
		query, args, err := sqlx.In(
			"SELECT token_id, usd, decimals FROM token WHERE token_id IN (?)",
			tokenIDs,
		)
		if err != nil {
			return err
		}
		query = hdb.db.Rebind(query)
		if err := meddler.QueryAll(
			hdb.db, &tokenPrices, query, args...,
		); err != nil {
			return err
		}
	}
	// Calculate total collected
	var total float64
	for _, tokenPrice := range tokenPrices {
		if tokenPrice.USD == nil {
			continue
		}
		f := new(big.Float).SetInt(batch.CollectedFees[tokenPrice.ID])
		amount, _ := f.Float64()
		total += *tokenPrice.USD * (amount / math.Pow(10, float64(tokenPrice.Decimals))) //nolint decimals have to be ^10
	}
	batch.TotalFeesUSD = &total
	// Insert to DB
	return meddler.Insert(d, "batch", batch)
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
	return hdb.addBatches(hdb.db, batches)
}
func (hdb *HistoryDB) addBatches(d meddler.DB, batches []common.Batch) error {
	for i := 0; i < len(batches); i++ {
		if err := hdb.addBatch(d, &batches[i]); err != nil {
			return err
		}
	}
	return nil
}

// GetBatchAPI return the batch with the given batchNum
func (hdb *HistoryDB) GetBatchAPI(batchNum common.BatchNum) (*BatchAPI, error) {
	batch := &BatchAPI{}
	return batch, meddler.QueryRow(
		hdb.db, batch,
		`SELECT batch.*, block.timestamp, block.hash
	 	FROM batch INNER JOIN block ON batch.eth_block_num = block.eth_block_num
	 	WHERE batch_num = $1;`, batchNum,
	)
}

// GetBatchesAPI return the batches applying the given filters
func (hdb *HistoryDB) GetBatchesAPI(
	minBatchNum, maxBatchNum, slotNum *uint,
	forgerAddr *ethCommon.Address,
	fromItem, limit *uint, order string,
) ([]BatchAPI, *db.Pagination, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT batch.*, block.timestamp, block.hash,
	count(*) OVER() AS total_items, MIN(batch.item_id) OVER() AS first_item,
	MAX(batch.item_id) OVER() AS last_item 
	FROM batch INNER JOIN block ON batch.eth_block_num = block.eth_block_num `
	// Apply filters
	nextIsAnd := false
	// minBatchNum filter
	if minBatchNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "batch.batch_num > ? "
		args = append(args, minBatchNum)
		nextIsAnd = true
	}
	// maxBatchNum filter
	if maxBatchNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "batch.batch_num < ? "
		args = append(args, maxBatchNum)
		nextIsAnd = true
	}
	// slotNum filter
	if slotNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "batch.slot_num = ? "
		args = append(args, slotNum)
		nextIsAnd = true
	}
	// forgerAddr filter
	if forgerAddr != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "batch.forger_addr = ? "
		args = append(args, forgerAddr)
		nextIsAnd = true
	}
	// pagination
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "batch.item_id >= ? "
		} else {
			queryStr += "batch.item_id <= ? "
		}
		args = append(args, fromItem)
	}
	queryStr += "ORDER BY batch.item_id "
	if order == OrderAsc {
		queryStr += " ASC "
	} else {
		queryStr += " DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query = hdb.db.Rebind(queryStr)
	// log.Debug(query)
	batchPtrs := []*BatchAPI{}
	if err := meddler.QueryAll(hdb.db, &batchPtrs, query, args...); err != nil {
		return nil, nil, err
	}
	batches := db.SlicePtrsToSlice(batchPtrs).([]BatchAPI)
	if len(batches) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return batches, &db.Pagination{
		TotalItems: batches[0].TotalItems,
		FirstItem:  batches[0].FirstItem,
		LastItem:   batches[0].LastItem,
	}, nil
}

// GetAllBatches retrieve all batches from the DB
func (hdb *HistoryDB) GetAllBatches() ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.db, &batches,
		`SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr, batch.fees_collected,
		 batch.fee_idxs_coordinator, batch.state_root, batch.num_accounts, batch.last_idx, batch.exit_root,
		 batch.forge_l1_txs_num, batch.slot_num, batch.total_fees_usd FROM batch;`,
	)
	return db.SlicePtrsToSlice(batches).([]common.Batch), err
}

// GetBatches retrieve batches from the DB, given a range of batch numbers defined by from and to
func (hdb *HistoryDB) GetBatches(from, to common.BatchNum) ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.db, &batches,
		"SELECT * FROM batch WHERE $1 <= batch_num AND batch_num < $2;",
		from, to,
	)
	return db.SlicePtrsToSlice(batches).([]common.Batch), err
}

// GetLastBatchNum returns the BatchNum of the latest forged batch
func (hdb *HistoryDB) GetLastBatchNum() (common.BatchNum, error) {
	row := hdb.db.QueryRow("SELECT batch_num FROM batch ORDER BY batch_num DESC LIMIT 1;")
	var batchNum common.BatchNum
	return batchNum, row.Scan(&batchNum)
}

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB from forged
// batches.  If there's no batch in the DB (nil, nil) is returned.
func (hdb *HistoryDB) GetLastL1TxsNum() (*int64, error) {
	row := hdb.db.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	lastL1TxsNum := new(int64)
	return lastL1TxsNum, row.Scan(&lastL1TxsNum)
}

// Reorg deletes all the information that was added into the DB after the
// lastValidBlock.  If lastValidBlock is negative, all block information is
// deleted.
func (hdb *HistoryDB) Reorg(lastValidBlock int64) error {
	var err error
	if lastValidBlock < 0 {
		_, err = hdb.db.Exec("DELETE FROM block;")
	} else {
		_, err = hdb.db.Exec("DELETE FROM block WHERE eth_block_num > $1;", lastValidBlock)
	}
	return err
}

// SyncPoD stores all the data that can be changed / added on a block in the PoD SC
func (hdb *HistoryDB) SyncPoD(
	blockNum uint64,
	bids []common.Bid,
	coordinators []common.Coordinator,
	vars *common.AuctionVariables,
) error {
	return nil
}

// AddBids insert Bids into the DB
func (hdb *HistoryDB) AddBids(bids []common.Bid) error { return hdb.addBids(hdb.db, bids) }
func (hdb *HistoryDB) addBids(d meddler.DB, bids []common.Bid) error {
	// TODO: check the coordinator info
	return db.BulkInsert(
		d,
		"INSERT INTO bid (slot_num, bid_value, eth_block_num, bidder_addr) VALUES %s;",
		bids[:],
	)
}

// GetAllBids retrieve all bids from the DB
func (hdb *HistoryDB) GetAllBids() ([]common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		`SELECT bid.slot_num, bid.bid_value, bid.eth_block_num, bid.bidder_addr FROM bid;`,
	)
	return db.SlicePtrsToSlice(bids).([]common.Bid), err
}

// GetBestBidAPI returns the best bid in specific slot by slotNum
func (hdb *HistoryDB) GetBestBidAPI(slotNum *int64) (BidAPI, error) {
	bid := &BidAPI{}
	err := meddler.QueryRow(
		hdb.db, bid, `SELECT bid.*, block.timestamp, coordinator.forger_addr, coordinator.url 
		FROM bid INNER JOIN block ON bid.eth_block_num = block.eth_block_num 
		INNER JOIN coordinator ON bid.bidder_addr = coordinator.bidder_addr 
		WHERE slot_num = $1 ORDER BY item_id DESC LIMIT 1;`, slotNum,
	)
	return *bid, err
}

// GetBestBidsAPI returns the best bid in specific slot by slotNum
func (hdb *HistoryDB) GetBestBidsAPI(minSlotNum, maxSlotNum *int64, bidderAddr *ethCommon.Address, limit *uint, order string) ([]BidAPI, *db.Pagination, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT b.*, block.timestamp, coordinator.forger_addr, coordinator.url, 
	COUNT(*) OVER() AS total_items, MIN(b.slot_num) OVER() AS first_item, 
	MAX(b.slot_num) OVER() AS last_item FROM (
	   SELECT slot_num, MAX(item_id) as maxitem 
	   FROM bid GROUP BY slot_num
	   )
	AS x INNER JOIN bid AS b ON b.item_id = x.maxitem
	INNER JOIN block ON b.eth_block_num = block.eth_block_num
	INNER JOIN coordinator ON b.bidder_addr = coordinator.bidder_addr 
	WHERE (b.slot_num >= ? AND b.slot_num <= ?)`
	args = append(args, minSlotNum)
	args = append(args, maxSlotNum)
	// Apply filters
	if bidderAddr != nil {
		queryStr += " AND b.bidder_addr = ? "
		args = append(args, bidderAddr)
	}
	queryStr += " ORDER BY b.slot_num "
	if order == OrderAsc {
		queryStr += "ASC "
	} else {
		queryStr += "DESC "
	}
	if limit != nil {
		queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	}
	query = hdb.db.Rebind(queryStr)
	bidPtrs := []*BidAPI{}
	if err := meddler.QueryAll(hdb.db, &bidPtrs, query, args...); err != nil {
		return nil, nil, err
	}
	// log.Debug(query)
	bids := db.SlicePtrsToSlice(bidPtrs).([]BidAPI)
	if len(bids) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return bids, &db.Pagination{
		TotalItems: bids[0].TotalItems,
		FirstItem:  bids[0].FirstItem,
		LastItem:   bids[0].LastItem,
	}, nil
}

// GetBidsAPI return the bids applying the given filters
func (hdb *HistoryDB) GetBidsAPI(slotNum *int64, forgerAddr *ethCommon.Address, fromItem, limit *uint, order string) ([]BidAPI, *db.Pagination, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT bid.*, block.timestamp, coordinator.forger_addr, coordinator.url, 
	COUNT(*) OVER() AS total_items, MIN(bid.item_id) OVER() AS first_item, 
	MAX(bid.item_id) OVER() AS last_item FROM bid
	INNER JOIN block ON bid.eth_block_num = block.eth_block_num 
	INNER JOIN coordinator ON bid.bidder_addr = coordinator.bidder_addr `
	// Apply filters
	nextIsAnd := false
	// slotNum filter
	if slotNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "bid.slot_num = ? "
		args = append(args, slotNum)
		nextIsAnd = true
	}
	// slotNum filter
	if forgerAddr != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "bid.bidder_addr = ? "
		args = append(args, forgerAddr)
		nextIsAnd = true
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "bid.item_id >= ? "
		} else {
			queryStr += "bid.item_id <= ? "
		}
		args = append(args, fromItem)
	}
	// pagination
	queryStr += "ORDER BY bid.item_id "
	if order == OrderAsc {
		queryStr += "ASC "
	} else {
		queryStr += "DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query, argsQ, err := sqlx.In(queryStr, args...)
	if err != nil {
		return nil, nil, err
	}
	query = hdb.db.Rebind(query)
	bids := []*BidAPI{}
	if err := meddler.QueryAll(hdb.db, &bids, query, argsQ...); err != nil {
		return nil, nil, err
	}
	if len(bids) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(bids).([]BidAPI), &db.Pagination{
		TotalItems: bids[0].TotalItems,
		FirstItem:  bids[0].FirstItem,
		LastItem:   bids[0].LastItem,
	}, nil
}

// AddCoordinators insert Coordinators into the DB
func (hdb *HistoryDB) AddCoordinators(coordinators []common.Coordinator) error {
	return hdb.addCoordinators(hdb.db, coordinators)
}
func (hdb *HistoryDB) addCoordinators(d meddler.DB, coordinators []common.Coordinator) error {
	return db.BulkInsert(
		d,
		"INSERT INTO coordinator (bidder_addr, forger_addr, eth_block_num, url) VALUES %s;",
		coordinators[:],
	)
}

// AddExitTree insert Exit tree into the DB
func (hdb *HistoryDB) AddExitTree(exitTree []common.ExitInfo) error {
	return hdb.addExitTree(hdb.db, exitTree)
}
func (hdb *HistoryDB) addExitTree(d meddler.DB, exitTree []common.ExitInfo) error {
	return db.BulkInsert(
		d,
		"INSERT INTO exit_tree (batch_num, account_idx, merkle_proof, balance, "+
			"instant_withdrawn, delayed_withdraw_request, delayed_withdrawn) VALUES %s;",
		exitTree[:],
	)
}

type exitID struct {
	batchNum int64
	idx      int64
}

func (hdb *HistoryDB) updateExitTree(d meddler.DB, blockNum int64,
	instantWithdrawn []exitID, delayedWithdrawRequest []exitID) error {
	// helperQueryExitIDTuples is a helper function to build the query with
	// an array of tuples in the arguments side built from []exitID
	helperQueryExitIDTuples := func(queryTmpl string, blockNum int64, exits []exitID) (string, []interface{}) {
		args := make([]interface{}, len(exits)*2+1)
		holder := ""
		args[0] = blockNum
		for i, v := range exits {
			args[1+i*2+0] = v.batchNum
			args[1+i*2+1] = v.idx
			holder += "(?, ?),"
		}
		query := fmt.Sprintf(queryTmpl, holder[:len(holder)-1])
		return hdb.db.Rebind(query), args
	}

	if len(instantWithdrawn) > 0 {
		query, args := helperQueryExitIDTuples(
			`UPDATE exit_tree SET instant_withdrawn = ? WHERE (batch_num, account_idx) IN (%s);`,
			blockNum,
			instantWithdrawn,
		)
		if _, err := hdb.db.DB.Exec(query, args...); err != nil {
			return err
		}
	}
	if len(delayedWithdrawRequest) > 0 {
		query, args := helperQueryExitIDTuples(
			`UPDATE exit_tree SET delayed_withdraw_request = ? WHERE (batch_num, account_idx) IN (%s);`,
			blockNum,
			delayedWithdrawRequest,
		)
		if _, err := hdb.db.DB.Exec(query, args...); err != nil {
			return err
		}
	}
	return nil
}

// AddToken insert a token into the DB
func (hdb *HistoryDB) AddToken(token *common.Token) error {
	return meddler.Insert(hdb.db, "token", token)
}

// AddTokens insert tokens into the DB
func (hdb *HistoryDB) AddTokens(tokens []common.Token) error { return hdb.addTokens(hdb.db, tokens) }
func (hdb *HistoryDB) addTokens(d meddler.DB, tokens []common.Token) error {
	return db.BulkInsert(
		d,
		`INSERT INTO token (
			token_id,
			eth_block_num,
			eth_addr,
			name,
			symbol,
			decimals
		) VALUES %s;`,
		tokens[:],
	)
}

// UpdateTokenValue updates the USD value of a token
func (hdb *HistoryDB) UpdateTokenValue(tokenSymbol string, value float64) error {
	_, err := hdb.db.Exec(
		"UPDATE token SET usd = $1 WHERE symbol = $2;",
		value, tokenSymbol,
	)
	return err
}

// GetToken returns a token from the DB given a TokenID
func (hdb *HistoryDB) GetToken(tokenID common.TokenID) (*TokenWithUSD, error) {
	token := &TokenWithUSD{}
	err := meddler.QueryRow(
		hdb.db, token, `SELECT * FROM token WHERE token_id = $1;`, tokenID,
	)
	return token, err
}

// GetAllTokens returns all tokens from the DB
func (hdb *HistoryDB) GetAllTokens() ([]TokenWithUSD, error) {
	var tokens []*TokenWithUSD
	err := meddler.QueryAll(
		hdb.db, &tokens,
		"SELECT * FROM token ORDER BY token_id;",
	)
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), err
}

// GetTokens returns a list of tokens from the DB
func (hdb *HistoryDB) GetTokens(ids []common.TokenID, symbols []string, name string, fromItem, limit *uint, order string) ([]TokenWithUSD, *db.Pagination, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT * , COUNT(*) OVER() AS total_items, MIN(token.item_id) OVER() AS first_item, MAX(token.item_id) OVER() AS last_item FROM token `
	// Apply filters
	nextIsAnd := false
	if len(ids) > 0 {
		queryStr += "WHERE token_id IN (?) "
		nextIsAnd = true
		args = append(args, ids)
	}
	if len(symbols) > 0 {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "symbol IN (?) "
		args = append(args, symbols)
		nextIsAnd = true
	}
	if name != "" {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "name ~ ? "
		args = append(args, name)
		nextIsAnd = true
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "item_id >= ? "
		} else {
			queryStr += "item_id <= ? "
		}
		args = append(args, fromItem)
	}
	// pagination
	queryStr += "ORDER BY item_id "
	if order == OrderAsc {
		queryStr += "ASC "
	} else {
		queryStr += "DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query, argsQ, err := sqlx.In(queryStr, args...)
	if err != nil {
		return nil, nil, err
	}
	query = hdb.db.Rebind(query)
	tokens := []*TokenWithUSD{}
	if err := meddler.QueryAll(hdb.db, &tokens, query, argsQ...); err != nil {
		return nil, nil, err
	}
	if len(tokens) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), &db.Pagination{
		TotalItems: tokens[0].TotalItems,
		FirstItem:  tokens[0].FirstItem,
		LastItem:   tokens[0].LastItem,
	}, nil
}

// GetTokenSymbols returns all the token symbols from the DB
func (hdb *HistoryDB) GetTokenSymbols() ([]string, error) {
	var tokenSymbols []string
	rows, err := hdb.db.Query("SELECT symbol FROM token;")
	if err != nil {
		return nil, err
	}
	sym := new(string)
	for rows.Next() {
		err = rows.Scan(sym)
		if err != nil {
			return nil, err
		}
		tokenSymbols = append(tokenSymbols, *sym)
	}
	return tokenSymbols, nil
}

// AddAccounts insert accounts into the DB
func (hdb *HistoryDB) AddAccounts(accounts []common.Account) error {
	return hdb.addAccounts(hdb.db, accounts)
}
func (hdb *HistoryDB) addAccounts(d meddler.DB, accounts []common.Account) error {
	return db.BulkInsert(
		d,
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
func (hdb *HistoryDB) GetAccounts() ([]common.Account, error) {
	var accs []*common.Account
	err := meddler.QueryAll(
		hdb.db, &accs,
		"SELECT * FROM account ORDER BY idx;",
	)
	return db.SlicePtrsToSlice(accs).([]common.Account), err
}

// AddL1Txs inserts L1 txs to the DB. USD and LoadAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
func (hdb *HistoryDB) AddL1Txs(l1txs []common.L1Tx) error { return hdb.addL1Txs(hdb.db, l1txs) }

// addL1Txs inserts L1 txs to the DB. USD and LoadAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
func (hdb *HistoryDB) addL1Txs(d meddler.DB, l1txs []common.L1Tx) error {
	txs := []txWrite{}
	for i := 0; i < len(l1txs); i++ {
		af := new(big.Float).SetInt(l1txs[i].Amount)
		amountFloat, _ := af.Float64()
		laf := new(big.Float).SetInt(l1txs[i].LoadAmount)
		loadAmountFloat, _ := laf.Float64()
		txs = append(txs, txWrite{
			// Generic
			IsL1:        true,
			TxID:        l1txs[i].TxID,
			Type:        l1txs[i].Type,
			Position:    l1txs[i].Position,
			FromIdx:     &l1txs[i].FromIdx,
			ToIdx:       l1txs[i].ToIdx,
			Amount:      l1txs[i].Amount,
			AmountFloat: amountFloat,
			TokenID:     l1txs[i].TokenID,
			BatchNum:    l1txs[i].BatchNum,
			EthBlockNum: l1txs[i].EthBlockNum,
			// L1
			ToForgeL1TxsNum: l1txs[i].ToForgeL1TxsNum,
			UserOrigin:      &l1txs[i].UserOrigin,
			FromEthAddr:     &l1txs[i].FromEthAddr,
			FromBJJ:         l1txs[i].FromBJJ,
			LoadAmount:      l1txs[i].LoadAmount,
			LoadAmountFloat: &loadAmountFloat,
		})
	}
	return hdb.addTxs(d, txs)
}

// AddL2Txs inserts L2 txs to the DB. TokenID, USD and FeeUSD will be set automatically before storing the tx.
func (hdb *HistoryDB) AddL2Txs(l2txs []common.L2Tx) error { return hdb.addL2Txs(hdb.db, l2txs) }

// addL2Txs inserts L2 txs to the DB. TokenID, USD and FeeUSD will be set automatically before storing the tx.
func (hdb *HistoryDB) addL2Txs(d meddler.DB, l2txs []common.L2Tx) error {
	txs := []txWrite{}
	for i := 0; i < len(l2txs); i++ {
		f := new(big.Float).SetInt(l2txs[i].Amount)
		amountFloat, _ := f.Float64()
		txs = append(txs, txWrite{
			// Generic
			IsL1:        false,
			TxID:        l2txs[i].TxID,
			Type:        l2txs[i].Type,
			Position:    l2txs[i].Position,
			FromIdx:     &l2txs[i].FromIdx,
			ToIdx:       l2txs[i].ToIdx,
			Amount:      l2txs[i].Amount,
			AmountFloat: amountFloat,
			BatchNum:    &l2txs[i].BatchNum,
			EthBlockNum: l2txs[i].EthBlockNum,
			// L2
			Fee:   &l2txs[i].Fee,
			Nonce: &l2txs[i].Nonce,
		})
	}
	return hdb.addTxs(d, txs)
}

func (hdb *HistoryDB) addTxs(d meddler.DB, txs []txWrite) error {
	return db.BulkInsert(
		d,
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
			batch_num,
			eth_block_num,
			to_forge_l1_txs_num,
			user_origin,
			from_eth_addr,
			from_bjj,
			load_amount,
			load_amount_f,
			fee,
			nonce
		) VALUES %s;`,
		txs[:],
	)
}

// // GetTxs returns a list of txs from the DB
// func (hdb *HistoryDB) GetTxs() ([]common.Tx, error) {
// 	var txs []*common.Tx
// 	err := meddler.QueryAll(
// 		hdb.db, &txs,
// 		`SELECT * FROM tx
// 		ORDER BY (batch_num, position) ASC`,
// 	)
// 	return db.SlicePtrsToSlice(txs).([]common.Tx), err
// }

// GetHistoryTx returns a tx from the DB given a TxID
func (hdb *HistoryDB) GetHistoryTx(txID common.TxID) (*TxAPI, error) {
	tx := &TxAPI{}
	err := meddler.QueryRow(
		hdb.db, tx, `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
		hez_idx(tx.from_idx, token.symbol) AS from_idx, tx.from_eth_addr, tx.from_bjj,
		hez_idx(tx.to_idx, token.symbol) AS to_idx, tx.to_eth_addr, tx.to_bjj,
		tx.amount, tx.token_id, tx.amount_usd, 
		tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
		tx.load_amount, tx.load_amount_usd, tx.fee, tx.fee_usd, tx.nonce,
		token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
		token.eth_addr, token.name, token.symbol, token.decimals, token.usd,
		token.usd_update, block.timestamp
		FROM tx INNER JOIN token ON tx.token_id = token.token_id 
		INNER JOIN block ON tx.eth_block_num = block.eth_block_num 
		WHERE tx.id = $1;`, txID,
	)
	return tx, err
}

// GetHistoryTxs returns a list of txs from the DB using the HistoryTx struct
// and pagination info
func (hdb *HistoryDB) GetHistoryTxs(
	ethAddr *ethCommon.Address, bjj *babyjub.PublicKey,
	tokenID *common.TokenID, idx *common.Idx, batchNum *uint, txType *common.TxType,
	fromItem, limit *uint, order string,
) ([]TxAPI, *db.Pagination, error) {
	if ethAddr != nil && bjj != nil {
		return nil, nil, errors.New("ethAddr and bjj are incompatible")
	}
	var query string
	var args []interface{}
	queryStr := `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
	hez_idx(tx.from_idx, token.symbol) AS from_idx, tx.from_eth_addr, tx.from_bjj,
	hez_idx(tx.to_idx, token.symbol) AS to_idx, tx.to_eth_addr, tx.to_bjj,
	tx.amount, tx.token_id, tx.amount_usd, 
	tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
	tx.load_amount, tx.load_amount_usd, tx.fee, tx.fee_usd, tx.nonce,
	token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr, token.name, token.symbol, token.decimals, token.usd,
	token.usd_update, block.timestamp, count(*) OVER() AS total_items, 
	MIN(tx.item_id)  OVER() AS first_item, MAX(tx.item_id) OVER() AS last_item 
	FROM tx INNER JOIN token ON tx.token_id = token.token_id 
	INNER JOIN block ON tx.eth_block_num = block.eth_block_num `
	// Apply filters
	nextIsAnd := false
	// ethAddr filter
	if ethAddr != nil {
		queryStr = `WITH acc AS 
		(select idx from account where eth_addr = ?) ` + queryStr
		queryStr += ", acc WHERE (tx.from_idx IN(acc.idx) OR tx.to_idx IN(acc.idx)) "
		nextIsAnd = true
		args = append(args, ethAddr)
	} else if bjj != nil { // bjj filter
		queryStr = `WITH acc AS 
		(select idx from account where bjj = ?) ` + queryStr
		queryStr += ", acc WHERE (tx.from_idx IN(acc.idx) OR tx.to_idx IN(acc.idx)) "
		nextIsAnd = true
		args = append(args, bjj)
	}
	// tokenID filter
	if tokenID != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx.token_id = ? "
		args = append(args, tokenID)
		nextIsAnd = true
	}
	// idx filter
	if idx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "(tx.from_idx = ? OR tx.to_idx = ?) "
		args = append(args, idx, idx)
		nextIsAnd = true
	}
	// batchNum filter
	if batchNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx.batch_num = ? "
		args = append(args, batchNum)
		nextIsAnd = true
	}
	// txType filter
	if txType != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx.type = ? "
		args = append(args, txType)
		nextIsAnd = true
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "tx.item_id >= ? "
		} else {
			queryStr += "tx.item_id <= ? "
		}
		args = append(args, fromItem)
		nextIsAnd = true
	}
	if nextIsAnd {
		queryStr += "AND "
	} else {
		queryStr += "WHERE "
	}
	queryStr += "tx.batch_num IS NOT NULL "

	// pagination
	queryStr += "ORDER BY tx.item_id "
	if order == OrderAsc {
		queryStr += " ASC "
	} else {
		queryStr += " DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query = hdb.db.Rebind(queryStr)
	// log.Debug(query)
	txsPtrs := []*TxAPI{}
	if err := meddler.QueryAll(hdb.db, &txsPtrs, query, args...); err != nil {
		return nil, nil, err
	}
	txs := db.SlicePtrsToSlice(txsPtrs).([]TxAPI)
	if len(txs) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return txs, &db.Pagination{
		TotalItems: txs[0].TotalItems,
		FirstItem:  txs[0].FirstItem,
		LastItem:   txs[0].LastItem,
	}, nil
}

// GetAllExits returns all exit from the DB
func (hdb *HistoryDB) GetAllExits() ([]common.ExitInfo, error) {
	var exits []*common.ExitInfo
	err := meddler.QueryAll(
		hdb.db, &exits,
		`SELECT exit_tree.batch_num, exit_tree.account_idx, exit_tree.merkle_proof,
		exit_tree.balance, exit_tree.instant_withdrawn, exit_tree.delayed_withdraw_request,
		exit_tree.delayed_withdrawn FROM exit_tree;`,
	)
	return db.SlicePtrsToSlice(exits).([]common.ExitInfo), err
}

// GetExitAPI returns a exit from the DB
func (hdb *HistoryDB) GetExitAPI(batchNum *uint, idx *common.Idx) (*ExitAPI, error) {
	exit := &ExitAPI{}
	err := meddler.QueryRow(
		hdb.db, exit, `SELECT exit_tree.item_id, exit_tree.batch_num,
		hez_idx(exit_tree.account_idx, token.symbol) AS account_idx,
		exit_tree.merkle_proof, exit_tree.balance, exit_tree.instant_withdrawn,
		exit_tree.delayed_withdraw_request, exit_tree.delayed_withdrawn,
		token.token_id, token.item_id AS token_item_id, 
		token.eth_block_num AS token_block, token.eth_addr, token.name, token.symbol, 
		token.decimals, token.usd, token.usd_update
		FROM exit_tree INNER JOIN account ON exit_tree.account_idx = account.idx 
		INNER JOIN token ON account.token_id = token.token_id 
		WHERE exit_tree.batch_num = $1 AND exit_tree.account_idx = $2;`, batchNum, idx,
	)
	return exit, err
}

// GetExitsAPI returns a list of exits from the DB and pagination info
func (hdb *HistoryDB) GetExitsAPI(
	ethAddr *ethCommon.Address, bjj *babyjub.PublicKey, tokenID *common.TokenID,
	idx *common.Idx, batchNum *uint, onlyPendingWithdraws *bool,
	fromItem, limit *uint, order string,
) ([]ExitAPI, *db.Pagination, error) {
	if ethAddr != nil && bjj != nil {
		return nil, nil, errors.New("ethAddr and bjj are incompatible")
	}
	var query string
	var args []interface{}
	queryStr := `SELECT exit_tree.item_id, exit_tree.batch_num,
	hez_idx(exit_tree.account_idx, token.symbol) AS account_idx,
	exit_tree.merkle_proof, exit_tree.balance, exit_tree.instant_withdrawn,
	exit_tree.delayed_withdraw_request, exit_tree.delayed_withdrawn,
	token.token_id, token.item_id AS token_item_id,
	token.eth_block_num AS token_block, token.eth_addr, token.name, token.symbol,
	token.decimals, token.usd, token.usd_update, COUNT(*) OVER() AS total_items,
	MIN(exit_tree.item_id) OVER() AS first_item, MAX(exit_tree.item_id) OVER() AS last_item
	FROM exit_tree INNER JOIN account ON exit_tree.account_idx = account.idx 
	INNER JOIN token ON account.token_id = token.token_id `
	// Apply filters
	nextIsAnd := false
	// ethAddr filter
	if ethAddr != nil {
		queryStr += "WHERE account.eth_addr = ? "
		nextIsAnd = true
		args = append(args, ethAddr)
	} else if bjj != nil { // bjj filter
		queryStr += "WHERE account.bjj = ? "
		nextIsAnd = true
		args = append(args, bjj)
	}
	// tokenID filter
	if tokenID != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "account.token_id = ? "
		args = append(args, tokenID)
		nextIsAnd = true
	}
	// idx filter
	if idx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "exit_tree.account_idx = ? "
		args = append(args, idx)
		nextIsAnd = true
	}
	// batchNum filter
	if batchNum != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "exit_tree.batch_num = ? "
		args = append(args, batchNum)
		nextIsAnd = true
	}
	// onlyPendingWithdraws
	if onlyPendingWithdraws != nil {
		if *onlyPendingWithdraws {
			if nextIsAnd {
				queryStr += "AND "
			} else {
				queryStr += "WHERE "
			}
			queryStr += "(exit_tree.instant_withdrawn IS NULL AND exit_tree.delayed_withdrawn IS NULL) "
			nextIsAnd = true
		}
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "exit_tree.item_id >= ? "
		} else {
			queryStr += "exit_tree.item_id <= ? "
		}
		args = append(args, fromItem)
		// nextIsAnd = true
	}
	// pagination
	queryStr += "ORDER BY exit_tree.item_id "
	if order == OrderAsc {
		queryStr += " ASC "
	} else {
		queryStr += " DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query = hdb.db.Rebind(queryStr)
	// log.Debug(query)
	exits := []*ExitAPI{}
	if err := meddler.QueryAll(hdb.db, &exits, query, args...); err != nil {
		return nil, nil, err
	}
	if len(exits) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(exits).([]ExitAPI), &db.Pagination{
		TotalItems: exits[0].TotalItems,
		FirstItem:  exits[0].FirstItem,
		LastItem:   exits[0].LastItem,
	}, nil
}

// // GetTx returns a tx from the DB
// func (hdb *HistoryDB) GetTx(txID common.TxID) (*common.Tx, error) {
// 	tx := new(common.Tx)
// 	return tx, meddler.QueryRow(
// 		hdb.db, tx,
// 		"SELECT * FROM tx WHERE id = $1;",
// 		txID,
// 	)
// }

// GetAllL1UserTxs returns all L1UserTxs from the DB
func (hdb *HistoryDB) GetAllL1UserTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id, tx.amount,
		tx.load_amount, tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = TRUE;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), err
}

// GetAllL1CoordinatorTxs returns all L1CoordinatorTxs from the DB
func (hdb *HistoryDB) GetAllL1CoordinatorTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id, tx.amount,
		tx.load_amount, tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = FALSE;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), err
}

// GetAllL2Txs returns all L2Txs from the DB
func (hdb *HistoryDB) GetAllL2Txs() ([]common.L2Tx, error) {
	var txs []*common.L2Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT tx.id, tx.batch_num, tx.position,
		tx.from_idx, tx.to_idx, tx.amount, tx.fee, tx.nonce,
		tx.type, tx.eth_block_num
		FROM tx WHERE is_l1 = FALSE;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L2Tx), err
}

// GetL1UserTxs gets L1 User Txs to be forged in the L1Batch with toForgeL1TxsNum.
func (hdb *HistoryDB) GetL1UserTxs(toForgeL1TxsNum int64) ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id, tx.amount,
		tx.load_amount, tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE to_forge_l1_txs_num = $1 AND is_l1 = TRUE AND user_origin = TRUE;`,
		toForgeL1TxsNum,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), err
}

// TODO: Think about chaning all the queries that return a last value, to queries that return the next valid value.

// GetLastTxsPosition for a given to_forge_l1_txs_num
func (hdb *HistoryDB) GetLastTxsPosition(toForgeL1TxsNum int64) (int, error) {
	row := hdb.db.QueryRow("SELECT MAX(position) FROM tx WHERE to_forge_l1_txs_num = $1;", toForgeL1TxsNum)
	var lastL1TxsPosition int
	return lastL1TxsPosition, row.Scan(&lastL1TxsPosition)
}

// GetSCVars returns the rollup, auction and wdelayer smart contracts variables at their last update.
func (hdb *HistoryDB) GetSCVars() (*common.RollupVariables, *common.AuctionVariables,
	*common.WDelayerVariables, error) {
	var rollup common.RollupVariables
	var auction common.AuctionVariables
	var wDelayer common.WDelayerVariables
	if err := meddler.QueryRow(hdb.db, &rollup,
		"SELECT * FROM rollup_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, err
	}
	if err := meddler.QueryRow(hdb.db, &auction,
		"SELECT * FROM auction_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, err
	}
	if err := meddler.QueryRow(hdb.db, &wDelayer,
		"SELECT * FROM wdelayer_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, err
	}
	return &rollup, &auction, &wDelayer, nil
}

func (hdb *HistoryDB) setRollupVars(d meddler.DB, rollup *common.RollupVariables) error {
	return meddler.Insert(d, "rollup_vars", rollup)
}

func (hdb *HistoryDB) setAuctionVars(d meddler.DB, auction *common.AuctionVariables) error {
	return meddler.Insert(d, "auction_vars", auction)
}

func (hdb *HistoryDB) setWDelayerVars(d meddler.DB, wDelayer *common.WDelayerVariables) error {
	return meddler.Insert(d, "wdelayer_vars", wDelayer)
}

// SetInitialSCVars sets the initial state of rollup, auction, wdelayer smart
// contract variables.  This initial state is stored linked to block 0, which
// always exist in the DB and is used to store initialization data that always
// exist in the smart contracts.
func (hdb *HistoryDB) SetInitialSCVars(rollup *common.RollupVariables,
	auction *common.AuctionVariables, wDelayer *common.WDelayerVariables) error {
	txn, err := hdb.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			db.Rollback(txn)
		}
	}()
	// Force EthBlockNum to be 0 because it's the block used to link data
	// that belongs to the creation of the smart contracts
	rollup.EthBlockNum = 0
	auction.EthBlockNum = 0
	wDelayer.EthBlockNum = 0
	if err := hdb.setRollupVars(txn, rollup); err != nil {
		return err
	}
	if err := hdb.setAuctionVars(txn, auction); err != nil {
		return err
	}
	if err := hdb.setWDelayerVars(txn, wDelayer); err != nil {
		return err
	}

	return txn.Commit()
}

// AddBlockSCData stores all the information of a block retrieved by the
// Synchronizer.  Blocks should be inserted in order, leaving no gaps because
// the pagination system of the API/DB depends on this.  Within blocks, all
// items should also be in the correct order (Accounts, Tokens, Txs, etc.)
func (hdb *HistoryDB) AddBlockSCData(blockData *common.BlockData) (err error) {
	txn, err := hdb.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			db.Rollback(txn)
		}
	}()

	// Add block
	if err := hdb.addBlock(txn, &blockData.Block); err != nil {
		return err
	}

	// Add Coordinators
	if len(blockData.Auction.Coordinators) > 0 {
		if err := hdb.addCoordinators(txn, blockData.Auction.Coordinators); err != nil {
			return err
		}
	}

	// Add Bids
	if len(blockData.Auction.Bids) > 0 {
		if err := hdb.addBids(txn, blockData.Auction.Bids); err != nil {
			return err
		}
	}

	// Add Tokens
	if len(blockData.Rollup.AddedTokens) > 0 {
		if err := hdb.addTokens(txn, blockData.Rollup.AddedTokens); err != nil {
			return err
		}
	}

	// Add l1 Txs
	if len(blockData.Rollup.L1UserTxs) > 0 {
		if err := hdb.addL1Txs(txn, blockData.Rollup.L1UserTxs); err != nil {
			return err
		}
	}

	// Add Batches
	for i := range blockData.Rollup.Batches {
		batch := &blockData.Rollup.Batches[i]
		// Add Batch: this will trigger an update on the DB
		// that will set the batch num of forged L1 txs in this batch
		if err = hdb.addBatch(txn, &batch.Batch); err != nil {
			return err
		}

		// Add unforged l1 Txs
		if batch.L1Batch {
			if len(batch.L1CoordinatorTxs) > 0 {
				if err := hdb.addL1Txs(txn, batch.L1CoordinatorTxs); err != nil {
					return err
				}
			}
		}

		// Add l2 Txs
		if len(batch.L2Txs) > 0 {
			if err := hdb.addL2Txs(txn, batch.L2Txs); err != nil {
				return err
			}
		}

		// Add accounts
		if len(batch.CreatedAccounts) > 0 {
			if err := hdb.addAccounts(txn, batch.CreatedAccounts); err != nil {
				return err
			}
		}

		// Add exit tree
		if len(batch.ExitTree) > 0 {
			if err := hdb.addExitTree(txn, batch.ExitTree); err != nil {
				return err
			}
		}
	}
	if blockData.Rollup.Vars != nil {
		if err := hdb.setRollupVars(txn, blockData.Rollup.Vars); err != nil {
			return err
		}
	}
	if blockData.Auction.Vars != nil {
		if err := hdb.setAuctionVars(txn, blockData.Auction.Vars); err != nil {
			return err
		}
	}
	if blockData.WDelayer.Vars != nil {
		if err := hdb.setWDelayerVars(txn, blockData.WDelayer.Vars); err != nil {
			return err
		}
	}

	if len(blockData.Rollup.Withdrawals) > 0 {
		instantWithdrawn := []exitID{}
		delayedWithdrawRequest := []exitID{}
		for _, withdraw := range blockData.Rollup.Withdrawals {
			exitID := exitID{
				batchNum: int64(withdraw.NumExitRoot),
				idx:      int64(withdraw.Idx),
			}
			if withdraw.InstantWithdraw {
				instantWithdrawn = append(instantWithdrawn, exitID)
			} else {
				delayedWithdrawRequest = append(delayedWithdrawRequest, exitID)
			}
		}
		if err := hdb.updateExitTree(txn, blockData.Block.EthBlockNum,
			instantWithdrawn, delayedWithdrawRequest); err != nil {
			return err
		}
	}

	// TODO: Process WDelayer withdrawals

	return txn.Commit()
}

// GetCoordinatorAPI returns a coordinator by its bidderAddr
func (hdb *HistoryDB) GetCoordinatorAPI(bidderAddr ethCommon.Address) (*CoordinatorAPI, error) {
	coordinator := &CoordinatorAPI{}
	err := meddler.QueryRow(hdb.db, coordinator, "SELECT * FROM coordinator WHERE bidder_addr = $1;", bidderAddr)
	return coordinator, err
}

// GetCoordinatorsAPI returns a list of coordinators from the DB and pagination info
func (hdb *HistoryDB) GetCoordinatorsAPI(fromItem, limit *uint, order string) ([]CoordinatorAPI, *db.Pagination, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT coordinator.*, 
	COUNT(*) OVER() AS total_items, MIN(coordinator.item_id) OVER() AS first_item, MAX(coordinator.item_id) OVER() AS last_item
	FROM coordinator `
	// Apply filters
	if fromItem != nil {
		queryStr += "WHERE "
		if order == OrderAsc {
			queryStr += "coordinator.item_id >= ? "
		} else {
			queryStr += "coordinator.item_id <= ? "
		}
		args = append(args, fromItem)
	}
	// pagination
	queryStr += "ORDER BY coordinator.item_id "
	if order == OrderAsc {
		queryStr += " ASC "
	} else {
		queryStr += " DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query = hdb.db.Rebind(queryStr)

	coordinators := []*CoordinatorAPI{}
	if err := meddler.QueryAll(hdb.db, &coordinators, query, args...); err != nil {
		return nil, nil, err
	}
	if len(coordinators) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(coordinators).([]CoordinatorAPI), &db.Pagination{
		TotalItems: coordinators[0].TotalItems,
		FirstItem:  coordinators[0].FirstItem,
		LastItem:   coordinators[0].LastItem,
	}, nil
}

// AddAuctionVars insert auction vars into the DB
func (hdb *HistoryDB) AddAuctionVars(auctionVars *common.AuctionVariables) error {
	return meddler.Insert(hdb.db, "auction_vars", auctionVars)
}

// GetAuctionVars returns auction variables
func (hdb *HistoryDB) GetAuctionVars() (*common.AuctionVariables, error) {
	auctionVars := &common.AuctionVariables{}
	err := meddler.QueryRow(
		hdb.db, auctionVars, `SELECT * FROM auction_vars;`,
	)
	return auctionVars, err
}

// GetAccountAPI returns an account by its index
func (hdb *HistoryDB) GetAccountAPI(idx common.Idx) (*AccountAPI, error) {
	account := &AccountAPI{}
	err := meddler.QueryRow(hdb.db, account, `SELECT account.item_id, hez_idx(account.idx, token.symbol) as idx, account.batch_num, account.bjj, account.eth_addr,
	token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr as token_eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update 
	FROM account INNER JOIN token ON account.token_id = token.token_id WHERE idx = $1;`, idx)

	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetAccountsAPI returns a list of accounts from the DB and pagination info
func (hdb *HistoryDB) GetAccountsAPI(tokenIDs []common.TokenID, ethAddr *ethCommon.Address, bjj *babyjub.PublicKey, fromItem, limit *uint, order string) ([]AccountAPI, *db.Pagination, error) {
	if ethAddr != nil && bjj != nil {
		return nil, nil, errors.New("ethAddr and bjj are incompatible")
	}
	var query string
	var args []interface{}
	queryStr := `SELECT account.item_id, hez_idx(account.idx, token.symbol) as idx, account.batch_num, account.bjj, account.eth_addr,
	token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr as token_eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update, 
	COUNT(*) OVER() AS total_items, MIN(account.item_id) OVER() AS first_item, MAX(account.item_id) OVER() AS last_item  
	FROM account INNER JOIN token ON account.token_id = token.token_id `
	// Apply filters
	nextIsAnd := false
	// ethAddr filter
	if ethAddr != nil {
		queryStr += "WHERE account.eth_addr = ? "
		nextIsAnd = true
		args = append(args, ethAddr)
	} else if bjj != nil { // bjj filter
		queryStr += "WHERE account.bjj = ? "
		nextIsAnd = true
		args = append(args, bjj)
	}
	// tokenID filter
	if len(tokenIDs) > 0 {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "account.token_id IN (?) "
		args = append(args, tokenIDs)
		nextIsAnd = true
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "account.item_id >= ? "
		} else {
			queryStr += "account.item_id <= ? "
		}
		args = append(args, fromItem)
	}
	// pagination
	queryStr += "ORDER BY account.item_id "
	if order == OrderAsc {
		queryStr += " ASC "
	} else {
		queryStr += " DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)
	query, argsQ, err := sqlx.In(queryStr, args...)
	if err != nil {
		return nil, nil, err
	}
	query = hdb.db.Rebind(query)

	accounts := []*AccountAPI{}
	if err := meddler.QueryAll(hdb.db, &accounts, query, argsQ...); err != nil {
		return nil, nil, err
	}
	if len(accounts) == 0 {
		return nil, nil, sql.ErrNoRows
	}

	return db.SlicePtrsToSlice(accounts).([]AccountAPI), &db.Pagination{
		TotalItems: accounts[0].TotalItems,
		FirstItem:  accounts[0].FirstItem,
		LastItem:   accounts[0].LastItem,
	}, nil
}
