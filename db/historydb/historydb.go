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
	"github.com/hermeznetwork/tracerr"
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
	return tracerr.Wrap(meddler.Insert(d, "block", block))
}

// AddBlocks inserts blocks into the DB
func (hdb *HistoryDB) AddBlocks(blocks []common.Block) error {
	return tracerr.Wrap(hdb.addBlocks(hdb.db, blocks))
}

func (hdb *HistoryDB) addBlocks(d meddler.DB, blocks []common.Block) error {
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO block (
			eth_block_num,
			timestamp,
			hash
		) VALUES %s;`,
		blocks[:],
	))
}

// GetBlock retrieve a block from the DB, given a block number
func (hdb *HistoryDB) GetBlock(blockNum int64) (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.db, block,
		"SELECT * FROM block WHERE eth_block_num = $1;", blockNum,
	)
	return block, tracerr.Wrap(err)
}

// GetAllBlocks retrieve all blocks from the DB
func (hdb *HistoryDB) GetAllBlocks() ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block;",
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), tracerr.Wrap(err)
}

// GetBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) GetBlocks(from, to int64) ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.db, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2;",
		from, to,
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), tracerr.Wrap(err)
}

// GetLastBlock retrieve the block with the highest block number from the DB
func (hdb *HistoryDB) GetLastBlock() (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.db, block, "SELECT * FROM block ORDER BY eth_block_num DESC LIMIT 1;",
	)
	return block, tracerr.Wrap(err)
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
			"SELECT token_id, usd, decimals FROM token WHERE token_id IN (?);",
			tokenIDs,
		)
		if err != nil {
			return tracerr.Wrap(err)
		}
		query = hdb.db.Rebind(query)
		if err := meddler.QueryAll(
			hdb.db, &tokenPrices, query, args...,
		); err != nil {
			return tracerr.Wrap(err)
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
	return tracerr.Wrap(meddler.Insert(d, "batch", batch))
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
	return tracerr.Wrap(hdb.addBatches(hdb.db, batches))
}
func (hdb *HistoryDB) addBatches(d meddler.DB, batches []common.Batch) error {
	for i := 0; i < len(batches); i++ {
		if err := hdb.addBatch(d, &batches[i]); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// GetBatchAPI return the batch with the given batchNum
func (hdb *HistoryDB) GetBatchAPI(batchNum common.BatchNum) (*BatchAPI, error) {
	batch := &BatchAPI{}
	return batch, tracerr.Wrap(meddler.QueryRow(
		hdb.db, batch,
		`SELECT batch.*, block.timestamp, block.hash
	 	FROM batch INNER JOIN block ON batch.eth_block_num = block.eth_block_num
	 	WHERE batch_num = $1;`, batchNum,
	))
}

// GetBatchesAPI return the batches applying the given filters
func (hdb *HistoryDB) GetBatchesAPI(
	minBatchNum, maxBatchNum, slotNum *uint,
	forgerAddr *ethCommon.Address,
	fromItem, limit *uint, order string,
) ([]BatchAPI, uint64, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT batch.*, block.timestamp, block.hash,
	count(*) OVER() AS total_items
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
		return nil, 0, tracerr.Wrap(err)
	}
	batches := db.SlicePtrsToSlice(batchPtrs).([]BatchAPI)
	if len(batches) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return batches, batches[0].TotalItems - uint64(len(batches)), nil
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
	return db.SlicePtrsToSlice(batches).([]common.Batch), tracerr.Wrap(err)
}

// GetBatches retrieve batches from the DB, given a range of batch numbers defined by from and to
func (hdb *HistoryDB) GetBatches(from, to common.BatchNum) ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.db, &batches,
		"SELECT * FROM batch WHERE $1 <= batch_num AND batch_num < $2;",
		from, to,
	)
	return db.SlicePtrsToSlice(batches).([]common.Batch), tracerr.Wrap(err)
}

// GetBatchesLen retrieve number of batches from the DB, given a slotNum
func (hdb *HistoryDB) GetBatchesLen(slotNum int64) (int, error) {
	row := hdb.db.QueryRow("SELECT COUNT(*) FROM batch WHERE slot_num = $1;", slotNum)
	var batchesLen int
	return batchesLen, tracerr.Wrap(row.Scan(&batchesLen))
}

// GetLastBatchNum returns the BatchNum of the latest forged batch
func (hdb *HistoryDB) GetLastBatchNum() (common.BatchNum, error) {
	row := hdb.db.QueryRow("SELECT batch_num FROM batch ORDER BY batch_num DESC LIMIT 1;")
	var batchNum common.BatchNum
	return batchNum, tracerr.Wrap(row.Scan(&batchNum))
}

// GetLastL1BatchBlockNum returns the blockNum of the latest forged l1Batch
func (hdb *HistoryDB) GetLastL1BatchBlockNum() (int64, error) {
	row := hdb.db.QueryRow(`SELECT eth_block_num FROM batch
		WHERE forge_l1_txs_num IS NOT NULL
		ORDER BY batch_num DESC LIMIT 1;`)
	var blockNum int64
	return blockNum, tracerr.Wrap(row.Scan(&blockNum))
}

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB from forged
// batches.  If there's no batch in the DB (nil, nil) is returned.
func (hdb *HistoryDB) GetLastL1TxsNum() (*int64, error) {
	row := hdb.db.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	lastL1TxsNum := new(int64)
	return lastL1TxsNum, tracerr.Wrap(row.Scan(&lastL1TxsNum))
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
	return tracerr.Wrap(err)
}

// AddBids insert Bids into the DB
func (hdb *HistoryDB) AddBids(bids []common.Bid) error { return hdb.addBids(hdb.db, bids) }
func (hdb *HistoryDB) addBids(d meddler.DB, bids []common.Bid) error {
	if len(bids) == 0 {
		return nil
	}
	// TODO: check the coordinator info
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO bid (slot_num, bid_value, eth_block_num, bidder_addr) VALUES %s;",
		bids[:],
	))
}

// GetAllBids retrieve all bids from the DB
func (hdb *HistoryDB) GetAllBids() ([]common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		`SELECT bid.slot_num, bid.bid_value, bid.eth_block_num, bid.bidder_addr FROM bid;`,
	)
	return db.SlicePtrsToSlice(bids).([]common.Bid), tracerr.Wrap(err)
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
	return *bid, tracerr.Wrap(err)
}

// GetBestBidCoordinator returns the forger address of the highest bidder in a slot by slotNum
func (hdb *HistoryDB) GetBestBidCoordinator(slotNum int64) (*common.BidCoordinator, error) {
	bidCoord := &common.BidCoordinator{}
	err := meddler.QueryRow(
		hdb.db, bidCoord,
		`SELECT (
			SELECT default_slot_set_bid
			FROM auction_vars
			WHERE default_slot_set_bid_slot_num <= $1
			ORDER BY eth_block_num DESC LIMIT 1
			),
		bid.slot_num, bid.bid_value, bid.bidder_addr,
		coordinator.forger_addr, coordinator.url
		FROM bid
		INNER JOIN coordinator ON bid.bidder_addr = coordinator.bidder_addr
		WHERE bid.slot_num = $1 ORDER BY bid.item_id DESC LIMIT 1;`,
		slotNum)

	return bidCoord, tracerr.Wrap(err)
}

// GetBestBidsAPI returns the best bid in specific slot by slotNum
func (hdb *HistoryDB) GetBestBidsAPI(
	minSlotNum, maxSlotNum *int64,
	bidderAddr *ethCommon.Address,
	limit *uint, order string,
) ([]BidAPI, uint64, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT b.*, block.timestamp, coordinator.forger_addr, coordinator.url, 
	COUNT(*) OVER() AS total_items FROM (
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
		return nil, 0, tracerr.Wrap(err)
	}
	// log.Debug(query)
	bids := db.SlicePtrsToSlice(bidPtrs).([]BidAPI)
	if len(bids) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return bids, bids[0].TotalItems - uint64(len(bids)), nil
}

// GetBidsAPI return the bids applying the given filters
func (hdb *HistoryDB) GetBidsAPI(
	slotNum *int64, forgerAddr *ethCommon.Address,
	fromItem, limit *uint, order string,
) ([]BidAPI, uint64, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT bid.*, block.timestamp, coordinator.forger_addr, coordinator.url, 
	COUNT(*) OVER() AS total_items
	FROM bid INNER JOIN block ON bid.eth_block_num = block.eth_block_num 
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
		return nil, 0, tracerr.Wrap(err)
	}
	query = hdb.db.Rebind(query)
	bids := []*BidAPI{}
	if err := meddler.QueryAll(hdb.db, &bids, query, argsQ...); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(bids) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return db.SlicePtrsToSlice(bids).([]BidAPI), bids[0].TotalItems - uint64(len(bids)), nil
}

// AddCoordinators insert Coordinators into the DB
func (hdb *HistoryDB) AddCoordinators(coordinators []common.Coordinator) error {
	return tracerr.Wrap(hdb.addCoordinators(hdb.db, coordinators))
}
func (hdb *HistoryDB) addCoordinators(d meddler.DB, coordinators []common.Coordinator) error {
	if len(coordinators) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO coordinator (bidder_addr, forger_addr, eth_block_num, url) VALUES %s;",
		coordinators[:],
	))
}

// AddExitTree insert Exit tree into the DB
func (hdb *HistoryDB) AddExitTree(exitTree []common.ExitInfo) error {
	return tracerr.Wrap(hdb.addExitTree(hdb.db, exitTree))
}
func (hdb *HistoryDB) addExitTree(d meddler.DB, exitTree []common.ExitInfo) error {
	if len(exitTree) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO exit_tree (batch_num, account_idx, merkle_proof, balance, "+
			"instant_withdrawn, delayed_withdraw_request, delayed_withdrawn) VALUES %s;",
		exitTree[:],
	))
}

func (hdb *HistoryDB) updateExitTree(d sqlx.Ext, blockNum int64,
	rollupWithdrawals []common.WithdrawInfo, wDelayerWithdrawals []common.WDelayerTransfer) error {
	if len(rollupWithdrawals) == 0 && len(wDelayerWithdrawals) == 0 {
		return nil
	}
	type withdrawal struct {
		BatchNum               int64              `db:"batch_num"`
		AccountIdx             int64              `db:"account_idx"`
		InstantWithdrawn       *int64             `db:"instant_withdrawn"`
		DelayedWithdrawRequest *int64             `db:"delayed_withdraw_request"`
		DelayedWithdrawn       *int64             `db:"delayed_withdrawn"`
		Owner                  *ethCommon.Address `db:"owner"`
		Token                  *ethCommon.Address `db:"token"`
	}
	withdrawals := make([]withdrawal, len(rollupWithdrawals)+len(wDelayerWithdrawals))
	for i := range rollupWithdrawals {
		info := &rollupWithdrawals[i]
		withdrawals[i] = withdrawal{
			BatchNum:   int64(info.NumExitRoot),
			AccountIdx: int64(info.Idx),
		}
		if info.InstantWithdraw {
			withdrawals[i].InstantWithdrawn = &blockNum
		} else {
			withdrawals[i].DelayedWithdrawRequest = &blockNum
			withdrawals[i].Owner = &info.Owner
			withdrawals[i].Token = &info.Token
		}
	}
	for i := range wDelayerWithdrawals {
		info := &wDelayerWithdrawals[i]
		withdrawals[len(rollupWithdrawals)+i] = withdrawal{
			DelayedWithdrawn: &blockNum,
			Owner:            &info.Owner,
			Token:            &info.Token,
		}
	}
	// In VALUES we set an initial row of NULLs to set the types of each
	// variable passed as argument
	const query string = `
		UPDATE exit_tree e SET
			instant_withdrawn = d.instant_withdrawn,
			delayed_withdraw_request = CASE
				WHEN e.delayed_withdraw_request IS NOT NULL THEN e.delayed_withdraw_request
				ELSE d.delayed_withdraw_request
			END,
			delayed_withdrawn = d.delayed_withdrawn,
			owner = d.owner,
			token = d.token
		FROM (VALUES
			(NULL::::BIGINT, NULL::::BIGINT, NULL::::BIGINT, NULL::::BIGINT, NULL::::BIGINT, NULL::::BYTEA, NULL::::BYTEA),
			(:batch_num,
			 :account_idx,
			 :instant_withdrawn,
			 :delayed_withdraw_request,
			 :delayed_withdrawn,
			 :owner,
			 :token)
		) as d (batch_num, account_idx, instant_withdrawn, delayed_withdraw_request, delayed_withdrawn, owner, token)
		WHERE
			(d.batch_num IS NOT NULL AND e.batch_num = d.batch_num AND e.account_idx = d.account_idx) OR
			(d.delayed_withdrawn IS NOT NULL AND e.delayed_withdrawn IS NULL AND e.owner = d.owner AND e.token = d.token);
		`
	if len(withdrawals) > 0 {
		if _, err := sqlx.NamedExec(d, query, withdrawals); err != nil {
			return tracerr.Wrap(err)
		}
	}

	return nil
}

// AddToken insert a token into the DB
func (hdb *HistoryDB) AddToken(token *common.Token) error {
	return tracerr.Wrap(meddler.Insert(hdb.db, "token", token))
}

// AddTokens insert tokens into the DB
func (hdb *HistoryDB) AddTokens(tokens []common.Token) error { return hdb.addTokens(hdb.db, tokens) }
func (hdb *HistoryDB) addTokens(d meddler.DB, tokens []common.Token) error {
	if len(tokens) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
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
	))
}

// UpdateTokenValue updates the USD value of a token
func (hdb *HistoryDB) UpdateTokenValue(tokenSymbol string, value float64) error {
	_, err := hdb.db.Exec(
		"UPDATE token SET usd = $1 WHERE symbol = $2;",
		value, tokenSymbol,
	)
	return tracerr.Wrap(err)
}

// GetToken returns a token from the DB given a TokenID
func (hdb *HistoryDB) GetToken(tokenID common.TokenID) (*TokenWithUSD, error) {
	token := &TokenWithUSD{}
	err := meddler.QueryRow(
		hdb.db, token, `SELECT * FROM token WHERE token_id = $1;`, tokenID,
	)
	return token, tracerr.Wrap(err)
}

// GetAllTokens returns all tokens from the DB
func (hdb *HistoryDB) GetAllTokens() ([]TokenWithUSD, error) {
	var tokens []*TokenWithUSD
	err := meddler.QueryAll(
		hdb.db, &tokens,
		"SELECT * FROM token ORDER BY token_id;",
	)
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), tracerr.Wrap(err)
}

// GetTokens returns a list of tokens from the DB
func (hdb *HistoryDB) GetTokens(
	ids []common.TokenID, symbols []string, name string, fromItem,
	limit *uint, order string,
) ([]TokenWithUSD, uint64, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT * , COUNT(*) OVER() AS total_items FROM token `
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
		return nil, 0, tracerr.Wrap(err)
	}
	query = hdb.db.Rebind(query)
	tokens := []*TokenWithUSD{}
	if err := meddler.QueryAll(hdb.db, &tokens, query, argsQ...); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(tokens) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), uint64(len(tokens)) - tokens[0].TotalItems, nil
}

// GetTokenSymbols returns all the token symbols from the DB
func (hdb *HistoryDB) GetTokenSymbols() ([]string, error) {
	var tokenSymbols []string
	rows, err := hdb.db.Query("SELECT symbol FROM token;")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer db.RowsClose(rows)
	sym := new(string)
	for rows.Next() {
		err = rows.Scan(sym)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		tokenSymbols = append(tokenSymbols, *sym)
	}
	return tokenSymbols, nil
}

// AddAccounts insert accounts into the DB
func (hdb *HistoryDB) AddAccounts(accounts []common.Account) error {
	return tracerr.Wrap(hdb.addAccounts(hdb.db, accounts))
}
func (hdb *HistoryDB) addAccounts(d meddler.DB, accounts []common.Account) error {
	if len(accounts) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO account (
			idx,
			token_id,
			batch_num,
			bjj,
			eth_addr
		) VALUES %s;`,
		accounts[:],
	))
}

// GetAllAccounts returns a list of accounts from the DB
func (hdb *HistoryDB) GetAllAccounts() ([]common.Account, error) {
	var accs []*common.Account
	err := meddler.QueryAll(
		hdb.db, &accs,
		"SELECT * FROM account ORDER BY idx;",
	)
	return db.SlicePtrsToSlice(accs).([]common.Account), tracerr.Wrap(err)
}

// AddL1Txs inserts L1 txs to the DB. USD and DepositAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
// EffectiveAmount and EffectiveDepositAmount are seted with default values by the DB.
func (hdb *HistoryDB) AddL1Txs(l1txs []common.L1Tx) error {
	return tracerr.Wrap(hdb.addL1Txs(hdb.db, l1txs))
}

// addL1Txs inserts L1 txs to the DB. USD and DepositAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
// EffectiveAmount and EffectiveDepositAmount are seted with default values by the DB.
func (hdb *HistoryDB) addL1Txs(d meddler.DB, l1txs []common.L1Tx) error {
	txs := []txWrite{}
	for i := 0; i < len(l1txs); i++ {
		af := new(big.Float).SetInt(l1txs[i].Amount)
		amountFloat, _ := af.Float64()
		laf := new(big.Float).SetInt(l1txs[i].DepositAmount)
		depositAmountFloat, _ := laf.Float64()
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
			ToForgeL1TxsNum:    l1txs[i].ToForgeL1TxsNum,
			UserOrigin:         &l1txs[i].UserOrigin,
			FromEthAddr:        &l1txs[i].FromEthAddr,
			FromBJJ:            l1txs[i].FromBJJ,
			DepositAmount:      l1txs[i].DepositAmount,
			DepositAmountFloat: &depositAmountFloat,
		})
	}
	return tracerr.Wrap(hdb.addTxs(d, txs))
}

// AddL2Txs inserts L2 txs to the DB. TokenID, USD and FeeUSD will be set automatically before storing the tx.
func (hdb *HistoryDB) AddL2Txs(l2txs []common.L2Tx) error {
	return tracerr.Wrap(hdb.addL2Txs(hdb.db, l2txs))
}

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
	return tracerr.Wrap(hdb.addTxs(d, txs))
}

func (hdb *HistoryDB) addTxs(d meddler.DB, txs []txWrite) error {
	if len(txs) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
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
			deposit_amount,
			deposit_amount_f,
			fee,
			nonce
		) VALUES %s;`,
		txs[:],
	))
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
	// Warning: amount_success and deposit_amount_success have true as default for
	// performance reasons. The expected default value is false (when txs are unforged)
	// this case is handled at the function func (tx TxAPI) MarshalJSON() ([]byte, error)
	tx := &TxAPI{}
	err := meddler.QueryRow(
		hdb.db, tx, `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
		hez_idx(tx.from_idx, token.symbol) AS from_idx, tx.from_eth_addr, tx.from_bjj,
		hez_idx(tx.to_idx, token.symbol) AS to_idx, tx.to_eth_addr, tx.to_bjj,
		tx.amount, tx.amount_success, tx.token_id, tx.amount_usd, 
		tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
		tx.deposit_amount, tx.deposit_amount_usd, tx.deposit_amount_success, tx.fee, tx.fee_usd, tx.nonce,
		token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
		token.eth_addr, token.name, token.symbol, token.decimals, token.usd,
		token.usd_update, block.timestamp
		FROM tx INNER JOIN token ON tx.token_id = token.token_id 
		INNER JOIN block ON tx.eth_block_num = block.eth_block_num 
		WHERE tx.id = $1;`, txID,
	)
	return tx, tracerr.Wrap(err)
}

// GetHistoryTxs returns a list of txs from the DB using the HistoryTx struct
// and pagination info
func (hdb *HistoryDB) GetHistoryTxs(
	ethAddr *ethCommon.Address, bjj *babyjub.PublicKey,
	tokenID *common.TokenID, idx *common.Idx, batchNum *uint, txType *common.TxType,
	fromItem, limit *uint, order string,
) ([]TxAPI, uint64, error) {
	// Warning: amount_success and deposit_amount_success have true as default for
	// performance reasons. The expected default value is false (when txs are unforged)
	// this case is handled at the function func (tx TxAPI) MarshalJSON() ([]byte, error)
	if ethAddr != nil && bjj != nil {
		return nil, 0, tracerr.Wrap(errors.New("ethAddr and bjj are incompatible"))
	}
	var query string
	var args []interface{}
	queryStr := `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
	hez_idx(tx.from_idx, token.symbol) AS from_idx, tx.from_eth_addr, tx.from_bjj,
	hez_idx(tx.to_idx, token.symbol) AS to_idx, tx.to_eth_addr, tx.to_bjj,
	tx.amount, tx.amount_success, tx.token_id, tx.amount_usd, 
	tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
	tx.deposit_amount, tx.deposit_amount_usd, tx.deposit_amount_success, tx.fee, tx.fee_usd, tx.nonce,
	token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr, token.name, token.symbol, token.decimals, token.usd,
	token.usd_update, block.timestamp, count(*) OVER() AS total_items 
	FROM tx INNER JOIN token ON tx.token_id = token.token_id 
	INNER JOIN block ON tx.eth_block_num = block.eth_block_num `
	// Apply filters
	nextIsAnd := false
	// ethAddr filter
	if ethAddr != nil {
		queryStr += "WHERE (tx.from_eth_addr = ? OR tx.to_eth_addr = ?) "
		nextIsAnd = true
		args = append(args, ethAddr, ethAddr)
	} else if bjj != nil { // bjj filter
		queryStr += "WHERE (tx.from_bjj = ? OR tx.to_bjj = ?) "
		nextIsAnd = true
		args = append(args, bjj, bjj)
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
		return nil, 0, tracerr.Wrap(err)
	}
	txs := db.SlicePtrsToSlice(txsPtrs).([]TxAPI)
	if len(txs) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return txs, txs[0].TotalItems - uint64(len(txs)), nil
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
	return db.SlicePtrsToSlice(exits).([]common.ExitInfo), tracerr.Wrap(err)
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
	return exit, tracerr.Wrap(err)
}

// GetExitsAPI returns a list of exits from the DB and pagination info
func (hdb *HistoryDB) GetExitsAPI(
	ethAddr *ethCommon.Address, bjj *babyjub.PublicKey, tokenID *common.TokenID,
	idx *common.Idx, batchNum *uint, onlyPendingWithdraws *bool,
	fromItem, limit *uint, order string,
) ([]ExitAPI, uint64, error) {
	if ethAddr != nil && bjj != nil {
		return nil, 0, tracerr.Wrap(errors.New("ethAddr and bjj are incompatible"))
	}
	var query string
	var args []interface{}
	queryStr := `SELECT exit_tree.item_id, exit_tree.batch_num,
	hez_idx(exit_tree.account_idx, token.symbol) AS account_idx,
	exit_tree.merkle_proof, exit_tree.balance, exit_tree.instant_withdrawn,
	exit_tree.delayed_withdraw_request, exit_tree.delayed_withdrawn,
	token.token_id, token.item_id AS token_item_id,
	token.eth_block_num AS token_block, token.eth_addr, token.name, token.symbol,
	token.decimals, token.usd, token.usd_update, COUNT(*) OVER() AS total_items
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
		return nil, 0, tracerr.Wrap(err)
	}
	if len(exits) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return db.SlicePtrsToSlice(exits).([]ExitAPI), exits[0].TotalItems - uint64(len(exits)), nil
}

// GetAllL1UserTxs returns all L1UserTxs from the DB
func (hdb *HistoryDB) GetAllL1UserTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.db, &txs, // Note that '\x' gets parsed as a big.Int with value = 0
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, (CASE WHEN tx.batch_num IS NULL THEN NULL WHEN tx.amount_success THEN tx.amount ELSE '\x' END) AS effective_amount,
		tx.deposit_amount, (CASE WHEN tx.batch_num IS NULL THEN NULL WHEN tx.deposit_amount_success THEN tx.deposit_amount ELSE '\x' END) AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = TRUE;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
}

// GetAllL1CoordinatorTxs returns all L1CoordinatorTxs from the DB
func (hdb *HistoryDB) GetAllL1CoordinatorTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	// Since the query specifies that only coordinator txs are returned, it's safe to assume
	// that returned txs will always have effective amounts
	err := meddler.QueryAll(
		hdb.db, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, tx.amount AS effective_amount,
		tx.deposit_amount, tx.deposit_amount AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = FALSE;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
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
	return db.SlicePtrsToSlice(txs).([]common.L2Tx), tracerr.Wrap(err)
}

// GetUnforgedL1UserTxs gets L1 User Txs to be forged in the L1Batch with toForgeL1TxsNum.
func (hdb *HistoryDB) GetUnforgedL1UserTxs(toForgeL1TxsNum int64) ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.db, &txs, // only L1 user txs can have batch_num set to null
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, NULL AS effective_amount,
		tx.deposit_amount, NULL AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE batch_num IS NULL AND to_forge_l1_txs_num = $1
		ORDER BY position;`,
		toForgeL1TxsNum,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
}

// TODO: Think about chaning all the queries that return a last value, to queries that return the next valid value.

// GetLastTxsPosition for a given to_forge_l1_txs_num
func (hdb *HistoryDB) GetLastTxsPosition(toForgeL1TxsNum int64) (int, error) {
	row := hdb.db.QueryRow(
		"SELECT position FROM tx WHERE to_forge_l1_txs_num = $1 ORDER BY position DESC;",
		toForgeL1TxsNum,
	)
	var lastL1TxsPosition int
	return lastL1TxsPosition, tracerr.Wrap(row.Scan(&lastL1TxsPosition))
}

// GetSCVars returns the rollup, auction and wdelayer smart contracts variables at their last update.
func (hdb *HistoryDB) GetSCVars() (*common.RollupVariables, *common.AuctionVariables,
	*common.WDelayerVariables, error) {
	var rollup common.RollupVariables
	var auction common.AuctionVariables
	var wDelayer common.WDelayerVariables
	if err := meddler.QueryRow(hdb.db, &rollup,
		"SELECT * FROM rollup_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if err := meddler.QueryRow(hdb.db, &auction,
		"SELECT * FROM auction_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if err := meddler.QueryRow(hdb.db, &wDelayer,
		"SELECT * FROM wdelayer_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	return &rollup, &auction, &wDelayer, nil
}

func (hdb *HistoryDB) setRollupVars(d meddler.DB, rollup *common.RollupVariables) error {
	return tracerr.Wrap(meddler.Insert(d, "rollup_vars", rollup))
}

func (hdb *HistoryDB) setAuctionVars(d meddler.DB, auction *common.AuctionVariables) error {
	return tracerr.Wrap(meddler.Insert(d, "auction_vars", auction))
}

func (hdb *HistoryDB) setWDelayerVars(d meddler.DB, wDelayer *common.WDelayerVariables) error {
	return tracerr.Wrap(meddler.Insert(d, "wdelayer_vars", wDelayer))
}

func (hdb *HistoryDB) addBucketUpdates(d meddler.DB, bucketUpdates []common.BucketUpdate) error {
	if len(bucketUpdates) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO bucket_update (
		 	eth_block_num,
		 	num_bucket,
		 	block_stamp,
		 	withdrawals
		) VALUES %s;`,
		bucketUpdates[:],
	))
}

// GetAllBucketUpdates retrieves all the bucket updates
func (hdb *HistoryDB) GetAllBucketUpdates() ([]common.BucketUpdate, error) {
	var bucketUpdates []*common.BucketUpdate
	err := meddler.QueryAll(
		hdb.db, &bucketUpdates,
		"SELECT * FROM bucket_update;",
	)
	return db.SlicePtrsToSlice(bucketUpdates).([]common.BucketUpdate), tracerr.Wrap(err)
}

func (hdb *HistoryDB) addTokenExchanges(d meddler.DB, tokenExchanges []common.TokenExchange) error {
	if len(tokenExchanges) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO token_exchange (
			eth_block_num,
    			eth_addr,
    			value_usd
		) VALUES %s;`,
		tokenExchanges[:],
	))
}

// GetAllTokenExchanges retrieves all the token exchanges
func (hdb *HistoryDB) GetAllTokenExchanges() ([]common.TokenExchange, error) {
	var tokenExchanges []*common.TokenExchange
	err := meddler.QueryAll(
		hdb.db, &tokenExchanges,
		"SELECT * FROM token_exchange;",
	)
	return db.SlicePtrsToSlice(tokenExchanges).([]common.TokenExchange), tracerr.Wrap(err)
}

func (hdb *HistoryDB) addEscapeHatchWithdrawals(d meddler.DB,
	escapeHatchWithdrawals []common.WDelayerEscapeHatchWithdrawal) error {
	if len(escapeHatchWithdrawals) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO escape_hatch_withdrawal (
			eth_block_num,
			who_addr,
			to_addr,
			token_addr,
			amount
		) VALUES %s;`,
		escapeHatchWithdrawals[:],
	))
}

// GetAllEscapeHatchWithdrawals retrieves all the escape hatch withdrawals
func (hdb *HistoryDB) GetAllEscapeHatchWithdrawals() ([]common.WDelayerEscapeHatchWithdrawal, error) {
	var escapeHatchWithdrawals []*common.WDelayerEscapeHatchWithdrawal
	err := meddler.QueryAll(
		hdb.db, &escapeHatchWithdrawals,
		"SELECT * FROM escape_hatch_withdrawal;",
	)
	return db.SlicePtrsToSlice(escapeHatchWithdrawals).([]common.WDelayerEscapeHatchWithdrawal),
		tracerr.Wrap(err)
}

// SetInitialSCVars sets the initial state of rollup, auction, wdelayer smart
// contract variables.  This initial state is stored linked to block 0, which
// always exist in the DB and is used to store initialization data that always
// exist in the smart contracts.
func (hdb *HistoryDB) SetInitialSCVars(rollup *common.RollupVariables,
	auction *common.AuctionVariables, wDelayer *common.WDelayerVariables) error {
	txn, err := hdb.db.Beginx()
	if err != nil {
		return tracerr.Wrap(err)
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
	auction.DefaultSlotSetBidSlotNum = 0
	if err := hdb.setRollupVars(txn, rollup); err != nil {
		return tracerr.Wrap(err)
	}
	if err := hdb.setAuctionVars(txn, auction); err != nil {
		return tracerr.Wrap(err)
	}
	if err := hdb.setWDelayerVars(txn, wDelayer); err != nil {
		return tracerr.Wrap(err)
	}

	return tracerr.Wrap(txn.Commit())
}

// setL1UserTxEffectiveAmounts sets the EffectiveAmount and EffectiveDepositAmount
// of the given l1UserTxs (with an UPDATE)
func (hdb *HistoryDB) setL1UserTxEffectiveAmounts(d sqlx.Ext, txs []common.L1Tx) error {
	if len(txs) == 0 {
		return nil
	}
	// Effective amounts are stored as success flags in the DB, with true value by default
	// to reduce the amount of updates. Therefore, only amounts that became uneffective should be
	// updated to become false
	type txUpdate struct {
		ID                   common.TxID `db:"id"`
		AmountSuccess        bool        `db:"amount_success"`
		DepositAmountSuccess bool        `db:"deposit_amount_success"`
	}
	txUpdates := []txUpdate{}
	equal := func(a *big.Int, b *big.Int) bool {
		return a.Cmp(b) == 0
	}
	for i := range txs {
		amountSuccess := equal(txs[i].Amount, txs[i].EffectiveAmount)
		depositAmountSuccess := equal(txs[i].DepositAmount, txs[i].EffectiveDepositAmount)
		if !amountSuccess || !depositAmountSuccess {
			txUpdates = append(txUpdates, txUpdate{
				ID:                   txs[i].TxID,
				AmountSuccess:        amountSuccess,
				DepositAmountSuccess: depositAmountSuccess,
			})
		}
	}
	const query string = `
		UPDATE tx SET
			amount_success = tx_update.amount_success,
			deposit_amount_success = tx_update.deposit_amount_success
		FROM (VALUES
			(NULL::::BYTEA, NULL::::BOOL, NULL::::BOOL),
			(:id, :amount_success, :deposit_amount_success)
		) as tx_update (id, amount_success, deposit_amount_success)
		WHERE tx.id = tx_update.id;
	`
	if len(txUpdates) > 0 {
		if _, err := sqlx.NamedExec(d, query, txUpdates); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// AddBlockSCData stores all the information of a block retrieved by the
// Synchronizer.  Blocks should be inserted in order, leaving no gaps because
// the pagination system of the API/DB depends on this.  Within blocks, all
// items should also be in the correct order (Accounts, Tokens, Txs, etc.)
func (hdb *HistoryDB) AddBlockSCData(blockData *common.BlockData) (err error) {
	txn, err := hdb.db.Beginx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer func() {
		if err != nil {
			db.Rollback(txn)
		}
	}()

	// Add block
	if err := hdb.addBlock(txn, &blockData.Block); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Coordinators
	if err := hdb.addCoordinators(txn, blockData.Auction.Coordinators); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Bids
	if err := hdb.addBids(txn, blockData.Auction.Bids); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Tokens
	if err := hdb.addTokens(txn, blockData.Rollup.AddedTokens); err != nil {
		return tracerr.Wrap(err)
	}

	// Prepare user L1 txs to be added.
	// They must be added before the batch that will forge them (which can be in the same block)
	// and after the account that will be sent to (also can be in the same block).
	// Note: insert order is not relevant since item_id will be updated by a DB trigger when
	// the batch that forges those txs is inserted
	userL1s := make(map[common.BatchNum][]common.L1Tx)
	for i := range blockData.Rollup.L1UserTxs {
		batchThatForgesIsInTheBlock := false
		for _, batch := range blockData.Rollup.Batches {
			if batch.Batch.ForgeL1TxsNum != nil &&
				*batch.Batch.ForgeL1TxsNum == *blockData.Rollup.L1UserTxs[i].ToForgeL1TxsNum {
				// Tx is forged in this block. It's guaranteed that:
				// * the first batch of the block won't forge user L1 txs that have been added in this block
				// * batch nums are sequential therefore it's safe to add the tx at batch.BatchNum -1
				batchThatForgesIsInTheBlock = true
				addAtBatchNum := batch.Batch.BatchNum - 1
				userL1s[addAtBatchNum] = append(userL1s[addAtBatchNum], blockData.Rollup.L1UserTxs[i])
				break
			}
		}
		if !batchThatForgesIsInTheBlock {
			// User artificial batchNum 0 to add txs that are not forge in this block
			// after all the accounts of the block have been added
			userL1s[0] = append(userL1s[0], blockData.Rollup.L1UserTxs[i])
		}
	}

	// Add Batches
	for i := range blockData.Rollup.Batches {
		batch := &blockData.Rollup.Batches[i]

		// Add Batch: this will trigger an update on the DB
		// that will set the batch num of forged L1 txs in this batch
		if err = hdb.addBatch(txn, &batch.Batch); err != nil {
			return tracerr.Wrap(err)
		}

		// Set the EffectiveAmount and EffectiveDepositAmount of all the
		// L1UserTxs that have been forged in this batch
		if err = hdb.setL1UserTxEffectiveAmounts(txn, batch.L1UserTxs); err != nil {
			return tracerr.Wrap(err)
		}

		// Add accounts
		if err := hdb.addAccounts(txn, batch.CreatedAccounts); err != nil {
			return tracerr.Wrap(err)
		}

		// Add forged l1 coordinator Txs
		if err := hdb.addL1Txs(txn, batch.L1CoordinatorTxs); err != nil {
			return tracerr.Wrap(err)
		}

		// Add l2 Txs
		if err := hdb.addL2Txs(txn, batch.L2Txs); err != nil {
			return tracerr.Wrap(err)
		}

		// Add user L1 txs that will be forged in next batch
		if userlL1s, ok := userL1s[batch.Batch.BatchNum]; ok {
			if err := hdb.addL1Txs(txn, userlL1s); err != nil {
				return tracerr.Wrap(err)
			}
		}

		// Add exit tree
		if err := hdb.addExitTree(txn, batch.ExitTree); err != nil {
			return tracerr.Wrap(err)
		}
	}
	// Add user L1 txs that won't be forged in this block
	if userL1sNotForgedInThisBlock, ok := userL1s[0]; ok {
		if err := hdb.addL1Txs(txn, userL1sNotForgedInThisBlock); err != nil {
			return tracerr.Wrap(err)
		}
	}

	// Set SC Vars if there was an update
	if blockData.Rollup.Vars != nil {
		if err := hdb.setRollupVars(txn, blockData.Rollup.Vars); err != nil {
			return tracerr.Wrap(err)
		}
	}
	if blockData.Auction.Vars != nil {
		if err := hdb.setAuctionVars(txn, blockData.Auction.Vars); err != nil {
			return tracerr.Wrap(err)
		}
	}
	if blockData.WDelayer.Vars != nil {
		if err := hdb.setWDelayerVars(txn, blockData.WDelayer.Vars); err != nil {
			return tracerr.Wrap(err)
		}
	}

	// Update withdrawals in exit tree table
	if err := hdb.updateExitTree(txn, blockData.Block.Num,
		blockData.Rollup.Withdrawals, blockData.WDelayer.Withdrawals); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Escape Hatch Withdrawals
	if err := hdb.addEscapeHatchWithdrawals(txn,
		blockData.WDelayer.EscapeHatchWithdrawals); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Buckets withdrawals updates
	if err := hdb.addBucketUpdates(txn, blockData.Rollup.UpdateBucketWithdraw); err != nil {
		return tracerr.Wrap(err)
	}

	// Add Token exchange updates
	if err := hdb.addTokenExchanges(txn, blockData.Rollup.TokenExchanges); err != nil {
		return tracerr.Wrap(err)
	}

	return tracerr.Wrap(txn.Commit())
}

// GetCoordinatorAPI returns a coordinator by its bidderAddr
func (hdb *HistoryDB) GetCoordinatorAPI(bidderAddr ethCommon.Address) (*CoordinatorAPI, error) {
	coordinator := &CoordinatorAPI{}
	err := meddler.QueryRow(hdb.db, coordinator, "SELECT * FROM coordinator WHERE bidder_addr = $1;", bidderAddr)
	return coordinator, tracerr.Wrap(err)
}

// GetCoordinatorsAPI returns a list of coordinators from the DB and pagination info
func (hdb *HistoryDB) GetCoordinatorsAPI(fromItem, limit *uint, order string) ([]CoordinatorAPI, uint64, error) {
	var query string
	var args []interface{}
	queryStr := `SELECT coordinator.*, 
	COUNT(*) OVER() AS total_items
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
		return nil, 0, tracerr.Wrap(err)
	}
	if len(coordinators) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}
	return db.SlicePtrsToSlice(coordinators).([]CoordinatorAPI),
		coordinators[0].TotalItems - uint64(len(coordinators)), nil
}

// AddAuctionVars insert auction vars into the DB
func (hdb *HistoryDB) AddAuctionVars(auctionVars *common.AuctionVariables) error {
	return tracerr.Wrap(meddler.Insert(hdb.db, "auction_vars", auctionVars))
}

// GetAuctionVars returns auction variables
func (hdb *HistoryDB) GetAuctionVars() (*common.AuctionVariables, error) {
	auctionVars := &common.AuctionVariables{}
	err := meddler.QueryRow(
		hdb.db, auctionVars, `SELECT * FROM auction_vars;`,
	)
	return auctionVars, tracerr.Wrap(err)
}

// GetAccountAPI returns an account by its index
func (hdb *HistoryDB) GetAccountAPI(idx common.Idx) (*AccountAPI, error) {
	account := &AccountAPI{}
	err := meddler.QueryRow(hdb.db, account, `SELECT account.item_id, hez_idx(account.idx, 
	token.symbol) as idx, account.batch_num, account.bjj, account.eth_addr,
	token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr as token_eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update 
	FROM account INNER JOIN token ON account.token_id = token.token_id WHERE idx = $1;`, idx)

	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return account, nil
}

// GetAccountsAPI returns a list of accounts from the DB and pagination info
func (hdb *HistoryDB) GetAccountsAPI(
	tokenIDs []common.TokenID, ethAddr *ethCommon.Address,
	bjj *babyjub.PublicKey, fromItem, limit *uint, order string,
) ([]AccountAPI, uint64, error) {
	if ethAddr != nil && bjj != nil {
		return nil, 0, tracerr.Wrap(errors.New("ethAddr and bjj are incompatible"))
	}
	var query string
	var args []interface{}
	queryStr := `SELECT account.item_id, hez_idx(account.idx, token.symbol) as idx, account.batch_num, 
	account.bjj, account.eth_addr, token.token_id, token.item_id AS token_item_id, token.eth_block_num AS token_block,
	token.eth_addr as token_eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update, 
	COUNT(*) OVER() AS total_items
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
		return nil, 0, tracerr.Wrap(err)
	}
	query = hdb.db.Rebind(query)

	accounts := []*AccountAPI{}
	if err := meddler.QueryAll(hdb.db, &accounts, query, argsQ...); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(accounts) == 0 {
		return nil, 0, tracerr.Wrap(sql.ErrNoRows)
	}

	return db.SlicePtrsToSlice(accounts).([]AccountAPI),
		accounts[0].TotalItems - uint64(len(accounts)), nil
}

// GetMetrics returns metrics
func (hdb *HistoryDB) GetMetrics(lastBatchNum common.BatchNum) (*Metrics, error) {
	metricsTotals := &MetricsTotals{}
	metrics := &Metrics{}
	err := meddler.QueryRow(
		hdb.db, metricsTotals, `SELECT COUNT(tx.*) as total_txs,
		COALESCE (MIN(tx.batch_num), 0) as batch_num 
		FROM tx INNER JOIN block ON tx.eth_block_num = block.eth_block_num
		WHERE block.timestamp >= NOW() - INTERVAL '24 HOURS';`)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	metrics.TransactionsPerSecond = float64(metricsTotals.TotalTransactions / (24 * 60 * 60))
	if (lastBatchNum - metricsTotals.FirstBatchNum) > 0 {
		metrics.TransactionsPerBatch = float64(int64(metricsTotals.TotalTransactions) /
			int64(lastBatchNum-metricsTotals.FirstBatchNum))
	} else {
		metrics.TransactionsPerBatch = float64(0)
	}

	err = meddler.QueryRow(
		hdb.db, metricsTotals, `SELECT COUNT(*) AS total_batches, 
		COALESCE (SUM(total_fees_usd), 0) AS total_fees FROM batch 
		WHERE batch_num > $1;`, metricsTotals.FirstBatchNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if metricsTotals.TotalBatches > 0 {
		metrics.BatchFrequency = float64((24 * 60 * 60) / metricsTotals.TotalBatches)
	} else {
		metrics.BatchFrequency = 0
	}
	if metricsTotals.TotalTransactions > 0 {
		metrics.AvgTransactionFee = metricsTotals.TotalFeesUSD / float64(metricsTotals.TotalTransactions)
	} else {
		metrics.AvgTransactionFee = 0
	}
	err = meddler.QueryRow(
		hdb.db, metrics,
		`SELECT COUNT(*) AS total_bjjs, COUNT(DISTINCT(bjj)) AS total_accounts FROM account;`)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return metrics, nil
}

// GetAvgTxFee returns average transaction fee of the last 1h
func (hdb *HistoryDB) GetAvgTxFee() (float64, error) {
	metricsTotals := &MetricsTotals{}
	err := meddler.QueryRow(
		hdb.db, metricsTotals, `SELECT COUNT(tx.*) as total_txs, 
		COALESCE (MIN(tx.batch_num), 0) as batch_num 
		FROM tx INNER JOIN block ON tx.eth_block_num = block.eth_block_num
		WHERE block.timestamp >= NOW() - INTERVAL '1 HOURS';`)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	err = meddler.QueryRow(
		hdb.db, metricsTotals, `SELECT COUNT(*) AS total_batches, 
		COALESCE (SUM(total_fees_usd), 0) AS total_fees FROM batch 
		WHERE batch_num > $1;`, metricsTotals.FirstBatchNum)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}

	var avgTransactionFee float64
	if metricsTotals.TotalTransactions > 0 {
		avgTransactionFee = metricsTotals.TotalFeesUSD / float64(metricsTotals.TotalTransactions)
	} else {
		avgTransactionFee = 0
	}

	return avgTransactionFee, nil
}
