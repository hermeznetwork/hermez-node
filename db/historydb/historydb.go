/*
Package historydb is responsible for storing and retrieving the historic data of the Hermez network.
It's mostly but not exclusively used by the API and the synchronizer.

Apart from the logic defined in this package, it's important to notice that there are some triggers defined in the
migration files that have to be taken into consideration to understanding the results of some queries. This is especially true
for reorgs: all the data is directly or indirectly related to a block, this makes handling reorgs as easy as deleting the
reorged blocks from the block table, and all related items will be dropped in cascade. This is not the only case, in general
functions defined in this package that get affected somehow by the SQL level defined logic has a special mention on the function description.

Some of the database tooling used in this package such as meddler and migration tools is explained in the db package.

This package is spitted in different files following these ideas:
- historydb.go: constructor and functions used by packages other than the api.
- apiqueries.go: functions used by the API, the queries implemented in this functions use a semaphore
to restrict the maximum concurrent connections to the database.
- views.go: structs used to retrieve/store data from/to the database. When possible, the common structs are used, however
most of the time there is no 1:1 relation between the struct fields and the tables of the schema, especially when joining tables.
In some cases, some of the structs defined in this file also include custom Marshallers to easily match the expected api formats.
- nodeinfo.go: used to handle the interfaces and structs that allow communication across running in different machines/process but sharing the same database.
*/
package historydb

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
	"github.com/russross/meddler"
)

// HistoryDB persist the historic of the rollup
type HistoryDB struct {
	dbRead     *sqlx.DB
	dbWrite    *sqlx.DB
	apiConnCon *db.APIConnectionController
}

// NewHistoryDB initialize the DB
func NewHistoryDB(dbRead, dbWrite *sqlx.DB, apiConnCon *db.APIConnectionController) *HistoryDB {
	return &HistoryDB{
		dbRead:     dbRead,
		dbWrite:    dbWrite,
		apiConnCon: apiConnCon,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (hdb *HistoryDB) DB() *sqlx.DB {
	return hdb.dbWrite
}

// AddBlock insert a block into the DB
func (hdb *HistoryDB) AddBlock(block *common.Block) error { return hdb.addBlock(hdb.dbWrite, block) }
func (hdb *HistoryDB) addBlock(d meddler.DB, block *common.Block) error {
	return tracerr.Wrap(meddler.Insert(d, "block", block))
}

// AddBlocks inserts blocks into the DB
func (hdb *HistoryDB) AddBlocks(blocks []common.Block) error {
	return tracerr.Wrap(hdb.addBlocks(hdb.dbWrite, blocks))
}

func (hdb *HistoryDB) addBlocks(d meddler.DB, blocks []common.Block) error {
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO block (
			eth_block_num,
			timestamp,
			hash
		) VALUES %s;`,
		blocks,
	))
}

// GetBlock retrieve a block from the DB, given a block number
func (hdb *HistoryDB) GetBlock(blockNum int64) (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.dbRead, block,
		"SELECT * FROM block WHERE eth_block_num = $1;", blockNum,
	)
	return block, tracerr.Wrap(err)
}

// GetAllBlocks retrieve all blocks from the DB
func (hdb *HistoryDB) GetAllBlocks() ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.dbRead, &blocks,
		"SELECT * FROM block ORDER BY eth_block_num;",
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), tracerr.Wrap(err)
}

// getBlocks retrieve blocks from the DB, given a range of block numbers defined by from and to
func (hdb *HistoryDB) getBlocks(from, to int64) ([]common.Block, error) {
	var blocks []*common.Block
	err := meddler.QueryAll(
		hdb.dbRead, &blocks,
		"SELECT * FROM block WHERE $1 <= eth_block_num AND eth_block_num < $2 ORDER BY eth_block_num;",
		from, to,
	)
	return db.SlicePtrsToSlice(blocks).([]common.Block), tracerr.Wrap(err)
}

// GetLastBlock retrieve the block with the highest block number from the DB
func (hdb *HistoryDB) GetLastBlock() (*common.Block, error) {
	block := &common.Block{}
	err := meddler.QueryRow(
		hdb.dbRead, block, "SELECT * FROM block ORDER BY eth_block_num DESC LIMIT 1;",
	)
	return block, tracerr.Wrap(err)
}

// AddBatch insert a Batch into the DB
func (hdb *HistoryDB) AddBatch(batch *common.Batch) error { return hdb.addBatch(hdb.dbWrite, batch) }
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
		query = hdb.dbWrite.Rebind(query)
		if err := meddler.QueryAll(
			hdb.dbWrite, &tokenPrices, query, args...,
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
	// Check current ether price and insert it into batch table
	var ether TokenWithUSD
	err := meddler.QueryRow(
		hdb.dbRead, &ether,
		"SELECT * FROM token WHERE symbol = 'ETH';",
	)
	if err != nil {
		log.Warn("error getting ether price from db: ", err)
		batch.EtherPriceUSD = 0
	} else if ether.USD == nil {
		batch.EtherPriceUSD = 0
	} else {
		batch.EtherPriceUSD = *ether.USD
	}
	if batch.GasPrice == nil {
		batch.GasPrice = big.NewInt(0)
	}
	// Insert to DB
	return tracerr.Wrap(meddler.Insert(d, "batch", batch))
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
	return tracerr.Wrap(hdb.addBatches(hdb.dbWrite, batches))
}
func (hdb *HistoryDB) addBatches(d meddler.DB, batches []common.Batch) error {
	for i := 0; i < len(batches); i++ {
		if err := hdb.addBatch(d, &batches[i]); err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// GetBatch returns the batch with the given batchNum
func (hdb *HistoryDB) GetBatch(batchNum common.BatchNum) (*common.Batch, error) {
	var batch common.Batch
	err := meddler.QueryRow(
		hdb.dbRead, &batch, `SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr,
		batch.fees_collected, batch.fee_idxs_coordinator, batch.state_root,
		batch.num_accounts, batch.last_idx, batch.exit_root, batch.forge_l1_txs_num,
		batch.slot_num, batch.total_fees_usd, batch.gas_price, batch.gas_used, batch.ether_price_usd
		FROM batch WHERE batch_num = $1;`,
		batchNum,
	)
	return &batch, tracerr.Wrap(err)
}

// GetAllBatches retrieve all batches from the DB
func (hdb *HistoryDB) GetAllBatches() ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.dbRead, &batches,
		`SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr, batch.fees_collected,
		 batch.fee_idxs_coordinator, batch.state_root, batch.num_accounts, batch.last_idx, batch.exit_root,
		 batch.forge_l1_txs_num, batch.slot_num, batch.total_fees_usd, batch.eth_tx_hash FROM batch
		 ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(batches).([]common.Batch), tracerr.Wrap(err)
}

// GetBatches retrieve batches from the DB, given a range of batch numbers defined by from and to
func (hdb *HistoryDB) GetBatches(from, to common.BatchNum) ([]common.Batch, error) {
	var batches []*common.Batch
	err := meddler.QueryAll(
		hdb.dbRead, &batches,
		`SELECT batch_num, eth_block_num, forger_addr, fees_collected, fee_idxs_coordinator, 
		state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num, total_fees_usd, gas_price, gas_used, ether_price_usd 
		FROM batch WHERE $1 <= batch_num AND batch_num < $2 ORDER BY batch_num;`,
		from, to,
	)
	return db.SlicePtrsToSlice(batches).([]common.Batch), tracerr.Wrap(err)
}

// GetFirstBatchBlockNumBySlot returns the ethereum block number of the first
// batch within a slot
func (hdb *HistoryDB) GetFirstBatchBlockNumBySlot(slotNum int64) (int64, error) {
	row := hdb.dbRead.QueryRow(
		`SELECT eth_block_num FROM batch
		WHERE slot_num = $1 ORDER BY batch_num ASC LIMIT 1;`, slotNum,
	)
	var blockNum int64
	return blockNum, tracerr.Wrap(row.Scan(&blockNum))
}

// GetLastBatchNum returns the BatchNum of the latest forged batch
func (hdb *HistoryDB) GetLastBatchNum() (common.BatchNum, error) {
	row := hdb.dbRead.QueryRow("SELECT batch_num FROM batch ORDER BY batch_num DESC LIMIT 1;")
	var batchNum common.BatchNum
	return batchNum, tracerr.Wrap(row.Scan(&batchNum))
}

// GetLastBatch returns the last forged batch
func (hdb *HistoryDB) GetLastBatch() (*common.Batch, error) {
	var batch common.Batch
	err := meddler.QueryRow(
		hdb.dbRead, &batch, `SELECT batch.batch_num, batch.eth_block_num, batch.forger_addr,
		batch.fees_collected, batch.fee_idxs_coordinator, batch.state_root,
		batch.num_accounts, batch.last_idx, batch.exit_root, batch.forge_l1_txs_num,
		batch.slot_num, batch.total_fees_usd, batch.gas_price, batch.gas_used, batch.ether_price_usd
		FROM batch ORDER BY batch_num DESC LIMIT 1;`,
	)
	return &batch, tracerr.Wrap(err)
}

// GetLastL1BatchBlockNum returns the blockNum of the latest forged l1Batch
func (hdb *HistoryDB) GetLastL1BatchBlockNum() (int64, error) {
	row := hdb.dbRead.QueryRow(`SELECT eth_block_num FROM batch
		WHERE forge_l1_txs_num IS NOT NULL
		ORDER BY batch_num DESC LIMIT 1;`)
	var blockNum int64
	return blockNum, tracerr.Wrap(row.Scan(&blockNum))
}

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB from forged
// batches.  If there's no batch in the DB (nil, nil) is returned.
func (hdb *HistoryDB) GetLastL1TxsNum() (*int64, error) {
	row := hdb.dbRead.QueryRow("SELECT MAX(forge_l1_txs_num) FROM batch;")
	lastL1TxsNum := new(int64)
	return lastL1TxsNum, tracerr.Wrap(row.Scan(&lastL1TxsNum))
}

// Reorg deletes all the information that was added into the DB after the
// lastValidBlock.  If lastValidBlock is negative, all block information is
// deleted.
func (hdb *HistoryDB) Reorg(lastValidBlock int64) error {
	var err error
	if lastValidBlock < 0 {
		_, err = hdb.dbWrite.Exec("DELETE FROM block;")
	} else {
		_, err = hdb.dbWrite.Exec("DELETE FROM block WHERE eth_block_num > $1;", lastValidBlock)
	}
	return tracerr.Wrap(err)
}

// AddBids insert Bids into the DB
func (hdb *HistoryDB) AddBids(bids []common.Bid) error { return hdb.addBids(hdb.dbWrite, bids) }
func (hdb *HistoryDB) addBids(d meddler.DB, bids []common.Bid) error {
	if len(bids) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO bid (slot_num, bid_value, eth_block_num, bidder_addr) VALUES %s;",
		bids,
	))
}

// GetAllBids retrieve all bids from the DB
func (hdb *HistoryDB) GetAllBids() ([]common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.dbRead, &bids,
		`SELECT bid.slot_num, bid.bid_value, bid.eth_block_num, bid.bidder_addr FROM bid
		ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(bids).([]common.Bid), tracerr.Wrap(err)
}

// GetBestBidCoordinator returns the forger address of the highest bidder in a slot by slotNum
func (hdb *HistoryDB) GetBestBidCoordinator(slotNum int64) (*common.BidCoordinator, error) {
	bidCoord := &common.BidCoordinator{}
	err := meddler.QueryRow(
		hdb.dbRead, bidCoord,
		`SELECT (
			SELECT default_slot_set_bid
			FROM auction_vars
			WHERE default_slot_set_bid_slot_num <= $1
			ORDER BY eth_block_num DESC LIMIT 1
		),
		bid.slot_num, bid.bid_value, bid.bidder_addr,
		coordinator.forger_addr, coordinator.url
		FROM bid
		INNER JOIN (
			SELECT bidder_addr, MAX(item_id) AS item_id FROM coordinator
			GROUP BY bidder_addr
		) c ON bid.bidder_addr = c.bidder_addr 
		INNER JOIN coordinator ON c.item_id = coordinator.item_id
		WHERE bid.slot_num = $1 ORDER BY bid.item_id DESC LIMIT 1;`,
		slotNum)

	return bidCoord, tracerr.Wrap(err)
}

// AddCoordinators insert Coordinators into the DB
func (hdb *HistoryDB) AddCoordinators(coordinators []common.Coordinator) error {
	return tracerr.Wrap(hdb.addCoordinators(hdb.dbWrite, coordinators))
}
func (hdb *HistoryDB) addCoordinators(d meddler.DB, coordinators []common.Coordinator) error {
	if len(coordinators) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO coordinator (bidder_addr, forger_addr, eth_block_num, url) VALUES %s;",
		coordinators,
	))
}

// AddExitTree insert Exit tree into the DB
func (hdb *HistoryDB) AddExitTree(exitTree []common.ExitInfo) error {
	return tracerr.Wrap(hdb.addExitTree(hdb.dbWrite, exitTree))
}
func (hdb *HistoryDB) addExitTree(d meddler.DB, exitTree []common.ExitInfo) error {
	if len(exitTree) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		"INSERT INTO exit_tree (batch_num, account_idx, merkle_proof, balance, "+
			"instant_withdrawn, delayed_withdraw_request, delayed_withdrawn) VALUES %s;",
		exitTree,
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
	return tracerr.Wrap(meddler.Insert(hdb.dbWrite, "token", token))
}

// AddTokens insert tokens into the DB
func (hdb *HistoryDB) AddTokens(tokens []common.Token) error {
	return hdb.addTokens(hdb.dbWrite, tokens)
}
func (hdb *HistoryDB) addTokens(d meddler.DB, tokens []common.Token) error {
	if len(tokens) == 0 {
		return nil
	}
	// Sanitize name and symbol
	for i, token := range tokens {
		token.Name = strings.ToValidUTF8(token.Name, " ")
		token.Symbol = strings.ToValidUTF8(token.Symbol, " ")
		tokens[i] = token
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
		tokens,
	))
}

// UpdateTokenValue updates the USD value of a token.  Value is the price in
// USD of a normalized token (1 token = 10^decimals units)
func (hdb *HistoryDB) UpdateTokenValue(tokenAddr ethCommon.Address, value float64) error {
	_, err := hdb.dbWrite.Exec(
		"UPDATE token SET usd = $1 WHERE eth_addr = $2;",
		value, tokenAddr,
	)
	return tracerr.Wrap(err)
}

// UpdateTokenValueByTokenID updates the USD value of a token. Value is the price in
// USD of a normalized token (1 token = 10^decimals units)
func (hdb *HistoryDB) UpdateTokenValueByTokenID(tokenID uint, value float64) error {
	// usd_update field is gonna be updated automatically due to trigger trigger_token_usd_update
	_, err := hdb.dbWrite.Exec(
		"UPDATE token SET usd = $1 WHERE token_id = $2;",
		value, tokenID,
	)
	return tracerr.Wrap(err)
}

// GetFiatPrice recover the price for a currency
func (hdb *HistoryDB) GetFiatPrice(currency, baseCurrency string) (FiatCurrency, error) {
	var currencyPrice = &FiatCurrency{}
	err := meddler.QueryRow(
		hdb.dbRead, currencyPrice, `SELECT currency, base_currency, price, last_update FROM fiat WHERE currency = $1 AND base_currency = $2;`,
		currency, baseCurrency,
	)
	return *currencyPrice, tracerr.Wrap(err)
}

// GetAllFiatPrice recover the price for all currencies
func (hdb *HistoryDB) GetAllFiatPrice(baseCurrency string) ([]FiatCurrency, error) {
	var currencyPrices []*FiatCurrency
	err := meddler.QueryAll(
		hdb.dbRead, &currencyPrices, `SELECT currency, base_currency, price, last_update FROM fiat WHERE base_currency = $1;`,
		baseCurrency,
	)
	return db.SlicePtrsToSlice(currencyPrices).([]FiatCurrency), tracerr.Wrap(err)
}

// GetToken returns a token from the DB given a TokenID
func (hdb *HistoryDB) GetToken(tokenID common.TokenID) (*TokenWithUSD, error) {
	token := &TokenWithUSD{}
	err := meddler.QueryRow(
		hdb.dbRead, token, `SELECT * FROM token WHERE token_id = $1;`, tokenID,
	)
	return token, tracerr.Wrap(err)
}

// GetAllTokens returns all tokens from the DB
func (hdb *HistoryDB) GetAllTokens() ([]TokenWithUSD, error) {
	var tokens []*TokenWithUSD
	err := meddler.QueryAll(
		hdb.dbRead, &tokens,
		"SELECT * FROM token ORDER BY token_id;",
	)
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), tracerr.Wrap(err)
}

// GetTokenSymbolsAndAddrs returns all the token symbols and addresses from the DB
func (hdb *HistoryDB) GetTokenSymbolsAndAddrs() ([]TokenSymbolAndAddr, error) {
	var tokens []*TokenSymbolAndAddr
	err := meddler.QueryAll(
		hdb.dbRead, &tokens,
		"SELECT symbol, eth_addr, token_id FROM token;",
	)
	return db.SlicePtrsToSlice(tokens).([]TokenSymbolAndAddr), tracerr.Wrap(err)
}

// AddAccounts insert accounts into the DB
func (hdb *HistoryDB) AddAccounts(accounts []common.Account) error {
	return tracerr.Wrap(hdb.addAccounts(hdb.dbWrite, accounts))
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
		accounts,
	))
}

// GetAllAccounts returns a list of accounts from the DB
func (hdb *HistoryDB) GetAllAccounts() ([]common.Account, error) {
	var accs []*common.Account
	err := meddler.QueryAll(
		hdb.dbRead, &accs,
		"SELECT idx, token_id, batch_num, bjj, eth_addr FROM account ORDER BY idx;",
	)
	return db.SlicePtrsToSlice(accs).([]common.Account), tracerr.Wrap(err)
}

// AddAccountUpdates inserts accUpdates into the DB
func (hdb *HistoryDB) AddAccountUpdates(accUpdates []common.AccountUpdate) error {
	return tracerr.Wrap(hdb.addAccountUpdates(hdb.dbWrite, accUpdates))
}
func (hdb *HistoryDB) addAccountUpdates(d meddler.DB, accUpdates []common.AccountUpdate) error {
	if len(accUpdates) == 0 {
		return nil
	}
	return tracerr.Wrap(db.BulkInsert(
		d,
		`INSERT INTO account_update (
			eth_block_num,
			batch_num,
			idx,
			nonce,
			balance
		) VALUES %s;`,
		accUpdates,
	))
}

// GetAllAccountUpdates returns all the AccountUpdate from the DB
func (hdb *HistoryDB) GetAllAccountUpdates() ([]common.AccountUpdate, error) {
	var accUpdates []*common.AccountUpdate
	err := meddler.QueryAll(
		hdb.dbRead, &accUpdates,
		"SELECT eth_block_num, batch_num, idx, nonce, balance FROM account_update ORDER BY idx;",
	)
	return db.SlicePtrsToSlice(accUpdates).([]common.AccountUpdate), tracerr.Wrap(err)
}

// AddL1Txs inserts L1 txs to the DB. USD and DepositAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
// EffectiveAmount and EffectiveDepositAmount are seted with default values by the DB.
func (hdb *HistoryDB) AddL1Txs(l1txs []common.L1Tx) error {
	return tracerr.Wrap(hdb.addL1Txs(hdb.dbWrite, l1txs))
}

// addL1Txs inserts L1 txs to the DB. USD and DepositAmountUSD will be set automatically before storing the tx.
// If the tx is originated by a coordinator, BatchNum must be provided. If it's originated by a user,
// BatchNum should be null, and the value will be setted by a trigger when a batch forges the tx.
// EffectiveAmount and EffectiveDepositAmount are seted with default values by the DB.
func (hdb *HistoryDB) addL1Txs(d meddler.DB, l1txs []common.L1Tx) error {
	if len(l1txs) == 0 {
		return nil
	}
	txs := []txWrite{}
	for i := 0; i < len(l1txs); i++ {
		af := new(big.Float).SetInt(l1txs[i].Amount)
		amountFloat, _ := af.Float64()
		laf := new(big.Float).SetInt(l1txs[i].DepositAmount)
		depositAmountFloat, _ := laf.Float64()
		var effectiveFromIdx *common.Idx
		if l1txs[i].UserOrigin {
			if l1txs[i].Type != common.TxTypeCreateAccountDeposit &&
				l1txs[i].Type != common.TxTypeCreateAccountDepositTransfer {
				effectiveFromIdx = &l1txs[i].FromIdx
			}
		} else {
			effectiveFromIdx = &l1txs[i].EffectiveFromIdx
		}
		txs = append(txs, txWrite{
			// Generic
			IsL1:             true,
			TxID:             l1txs[i].TxID,
			Type:             l1txs[i].Type,
			Position:         l1txs[i].Position,
			FromIdx:          &l1txs[i].FromIdx,
			EffectiveFromIdx: effectiveFromIdx,
			ToIdx:            l1txs[i].ToIdx,
			Amount:           l1txs[i].Amount,
			AmountFloat:      amountFloat,
			TokenID:          l1txs[i].TokenID,
			BatchNum:         l1txs[i].BatchNum,
			EthBlockNum:      l1txs[i].EthBlockNum,
			// L1
			ToForgeL1TxsNum:    l1txs[i].ToForgeL1TxsNum,
			UserOrigin:         &l1txs[i].UserOrigin,
			FromEthAddr:        &l1txs[i].FromEthAddr,
			FromBJJ:            &l1txs[i].FromBJJ,
			DepositAmount:      l1txs[i].DepositAmount,
			DepositAmountFloat: &depositAmountFloat,
			EthTxHash:          &l1txs[i].EthTxHash,
			L1Fee:              l1txs[i].L1Fee,
		})
	}
	return tracerr.Wrap(hdb.addTxs(d, txs))
}

// AddL2Txs inserts L2 txs to the DB. TokenID, USD and FeeUSD will be set automatically before storing the tx.
func (hdb *HistoryDB) AddL2Txs(l2txs []common.L2Tx) error {
	return tracerr.Wrap(hdb.addL2Txs(hdb.dbWrite, l2txs))
}

// addL2Txs inserts L2 txs to the DB. TokenID, USD and FeeUSD will be set automatically before storing the tx.
func (hdb *HistoryDB) addL2Txs(d meddler.DB, l2txs []common.L2Tx) error {
	txs := []txWrite{}
	for i := 0; i < len(l2txs); i++ {
		f := new(big.Float).SetInt(l2txs[i].Amount)
		amountFloat, _ := f.Float64()
		txs = append(txs, txWrite{
			// Generic
			IsL1:             false,
			TxID:             l2txs[i].TxID,
			Type:             l2txs[i].Type,
			Position:         l2txs[i].Position,
			FromIdx:          &l2txs[i].FromIdx,
			EffectiveFromIdx: &l2txs[i].FromIdx,
			ToIdx:            l2txs[i].ToIdx,
			TokenID:          l2txs[i].TokenID,
			Amount:           l2txs[i].Amount,
			AmountFloat:      amountFloat,
			BatchNum:         &l2txs[i].BatchNum,
			EthBlockNum:      l2txs[i].EthBlockNum,
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
			effective_from_idx,
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
			eth_tx_hash,
			l1_fee,
			fee,
			nonce
		) VALUES %s;`,
		txs,
	))
}

// GetAllExits returns all exit from the DB
func (hdb *HistoryDB) GetAllExits() ([]common.ExitInfo, error) {
	var exits []*common.ExitInfo
	err := meddler.QueryAll(
		hdb.dbRead, &exits,
		`SELECT exit_tree.batch_num, exit_tree.account_idx, exit_tree.merkle_proof,
		exit_tree.balance, exit_tree.instant_withdrawn, exit_tree.delayed_withdraw_request,
		exit_tree.delayed_withdrawn FROM exit_tree ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(exits).([]common.ExitInfo), tracerr.Wrap(err)
}

// GetAllL1UserTxs returns all L1UserTxs from the DB
func (hdb *HistoryDB) GetAllL1UserTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.dbRead, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.effective_from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, (CASE WHEN tx.batch_num IS NULL THEN NULL WHEN tx.amount_success THEN tx.amount ELSE 0 END) AS effective_amount,
		tx.deposit_amount, (CASE WHEN tx.batch_num IS NULL THEN NULL WHEN tx.deposit_amount_success THEN tx.deposit_amount ELSE 0 END) AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = TRUE ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
}

// GetAllL1CoordinatorTxs returns all L1CoordinatorTxs from the DB
func (hdb *HistoryDB) GetAllL1CoordinatorTxs() ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	// Since the query specifies that only coordinator txs are returned, it's safe to assume
	// that returned txs will always have effective amounts
	err := meddler.QueryAll(
		hdb.dbRead, &txs,
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.effective_from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, tx.amount AS effective_amount,
		tx.deposit_amount, tx.deposit_amount AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE is_l1 = TRUE AND user_origin = FALSE ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
}

// GetAllL2Txs returns all L2Txs from the DB
func (hdb *HistoryDB) GetAllL2Txs() ([]common.L2Tx, error) {
	var txs []*common.L2Tx
	err := meddler.QueryAll(
		hdb.dbRead, &txs,
		`SELECT tx.id, tx.batch_num, tx.position,
		tx.from_idx, tx.to_idx, tx.amount, tx.token_id,
		tx.fee, tx.nonce, tx.type, tx.eth_block_num
		FROM tx WHERE is_l1 = FALSE ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(txs).([]common.L2Tx), tracerr.Wrap(err)
}

// GetUnforgedL1UserTxs gets L1 User Txs to be forged in the L1Batch with toForgeL1TxsNum.
func (hdb *HistoryDB) GetUnforgedL1UserTxs(toForgeL1TxsNum int64) ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.dbRead, &txs, // only L1 user txs can have batch_num set to null
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

// GetUnforgedL1UserFutureTxs gets L1 User Txs to be forged after the L1Batch
// with toForgeL1TxsNum (in one of the future batches, not in the next one).
func (hdb *HistoryDB) GetUnforgedL1UserFutureTxs(toForgeL1TxsNum int64) ([]common.L1Tx, error) {
	var txs []*common.L1Tx
	err := meddler.QueryAll(
		hdb.dbRead, &txs, // only L1 user txs can have batch_num set to null
		`SELECT tx.id, tx.to_forge_l1_txs_num, tx.position, tx.user_origin,
		tx.from_idx, tx.from_eth_addr, tx.from_bjj, tx.to_idx, tx.token_id,
		tx.amount, NULL AS effective_amount,
		tx.deposit_amount, NULL AS effective_deposit_amount,
		tx.eth_block_num, tx.type, tx.batch_num
		FROM tx WHERE batch_num IS NULL AND to_forge_l1_txs_num > $1
		ORDER BY position;`,
		toForgeL1TxsNum,
	)
	return db.SlicePtrsToSlice(txs).([]common.L1Tx), tracerr.Wrap(err)
}

// GetUnforgedL1UserTxsCount returns the count of unforged L1Txs (either in
// open or frozen queues that are not yet forged)
func (hdb *HistoryDB) GetUnforgedL1UserTxsCount() (int, error) {
	row := hdb.dbRead.QueryRow(
		`SELECT COUNT(*) FROM tx WHERE batch_num IS NULL;`,
	)
	var count int
	return count, tracerr.Wrap(row.Scan(&count))
}

// GetLastTxsPosition for a given to_forge_l1_txs_num
func (hdb *HistoryDB) GetLastTxsPosition(toForgeL1TxsNum int64) (int, error) {
	row := hdb.dbRead.QueryRow(
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
	if err := meddler.QueryRow(hdb.dbRead, &rollup,
		"SELECT * FROM rollup_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if err := meddler.QueryRow(hdb.dbRead, &auction,
		"SELECT * FROM auction_vars ORDER BY eth_block_num DESC LIMIT 1;"); err != nil {
		return nil, nil, nil, tracerr.Wrap(err)
	}
	if err := meddler.QueryRow(hdb.dbRead, &wDelayer,
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
		bucketUpdates,
	))
}

// AddBucketUpdatesTest allows call to unexported method
// only for internal testing purposes
func (hdb *HistoryDB) AddBucketUpdatesTest(d meddler.DB, bucketUpdates []common.BucketUpdate) error {
	return hdb.addBucketUpdates(d, bucketUpdates)
}

// GetAllBucketUpdates retrieves all the bucket updates
func (hdb *HistoryDB) GetAllBucketUpdates() ([]common.BucketUpdate, error) {
	var bucketUpdates []*common.BucketUpdate
	err := meddler.QueryAll(
		hdb.dbRead, &bucketUpdates,
		`SELECT eth_block_num, num_bucket, block_stamp, withdrawals  
		FROM bucket_update ORDER BY item_id;`,
	)
	return db.SlicePtrsToSlice(bucketUpdates).([]common.BucketUpdate), tracerr.Wrap(err)
}

func (hdb *HistoryDB) getMinBidInfo(d meddler.DB,
	currentSlot, lastClosedSlot int64) ([]MinBidInfo, error) {
	minBidInfo := []*MinBidInfo{}
	query := `
		SELECT DISTINCT default_slot_set_bid, default_slot_set_bid_slot_num FROM auction_vars
		WHERE default_slot_set_bid_slot_num < $1
		ORDER BY default_slot_set_bid_slot_num DESC
		LIMIT $2;`
	err := meddler.QueryAll(d, &minBidInfo, query, lastClosedSlot, int(lastClosedSlot-currentSlot)+1)
	return db.SlicePtrsToSlice(minBidInfo).([]MinBidInfo), tracerr.Wrap(err)
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
		tokenExchanges,
	))
}

// GetAllTokenExchanges retrieves all the token exchanges
func (hdb *HistoryDB) GetAllTokenExchanges() ([]common.TokenExchange, error) {
	var tokenExchanges []*common.TokenExchange
	err := meddler.QueryAll(
		hdb.dbRead, &tokenExchanges,
		"SELECT eth_block_num, eth_addr, value_usd FROM token_exchange ORDER BY item_id;",
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
		escapeHatchWithdrawals,
	))
}

// GetAllEscapeHatchWithdrawals retrieves all the escape hatch withdrawals
func (hdb *HistoryDB) GetAllEscapeHatchWithdrawals() ([]common.WDelayerEscapeHatchWithdrawal, error) {
	var escapeHatchWithdrawals []*common.WDelayerEscapeHatchWithdrawal
	err := meddler.QueryAll(
		hdb.dbRead, &escapeHatchWithdrawals,
		"SELECT eth_block_num, who_addr, to_addr, token_addr, amount FROM escape_hatch_withdrawal ORDER BY item_id;",
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
	txn, err := hdb.dbWrite.Beginx()
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

// setExtraInfoForgedL1UserTxs sets the EffectiveAmount, EffectiveDepositAmount
// and EffectiveFromIdx of the given l1UserTxs (with an UPDATE)
func (hdb *HistoryDB) setExtraInfoForgedL1UserTxs(d sqlx.Ext, txs []common.L1Tx) error {
	if len(txs) == 0 {
		return nil
	}
	// Effective amounts are stored as success flags in the DB, with true value by default
	// to reduce the amount of updates. Therefore, only amounts that became uneffective should be
	// updated to become false.  At the same time, all the txs that contain
	// accounts (FromIdx == 0) are updated to set the EffectiveFromIdx.
	type txUpdate struct {
		ID                   common.TxID `db:"id"`
		AmountSuccess        bool        `db:"amount_success"`
		DepositAmountSuccess bool        `db:"deposit_amount_success"`
		EffectiveFromIdx     common.Idx  `db:"effective_from_idx"`
	}
	txUpdates := []txUpdate{}
	equal := func(a *big.Int, b *big.Int) bool {
		return a.Cmp(b) == 0
	}
	for i := range txs {
		amountSuccess := equal(txs[i].Amount, txs[i].EffectiveAmount)
		depositAmountSuccess := equal(txs[i].DepositAmount, txs[i].EffectiveDepositAmount)
		if !amountSuccess || !depositAmountSuccess || txs[i].FromIdx == 0 {
			txUpdates = append(txUpdates, txUpdate{
				ID:                   txs[i].TxID,
				AmountSuccess:        amountSuccess,
				DepositAmountSuccess: depositAmountSuccess,
				EffectiveFromIdx:     txs[i].EffectiveFromIdx,
			})
		}
	}
	const query string = `
		UPDATE tx SET
			amount_success = tx_update.amount_success,
			deposit_amount_success = tx_update.deposit_amount_success,
			effective_from_idx = tx_update.effective_from_idx
		FROM (VALUES
			(NULL::::BYTEA, NULL::::BOOL, NULL::::BOOL, NULL::::BIGINT),
			(:id, :amount_success, :deposit_amount_success, :effective_from_idx)
		) as tx_update (id, amount_success, deposit_amount_success, effective_from_idx)
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
	txn, err := hdb.dbWrite.Beginx()
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

		// Add accounts
		if err := hdb.addAccounts(txn, batch.CreatedAccounts); err != nil {
			return tracerr.Wrap(err)
		}

		// Add accountBalances if it exists
		if err := hdb.addAccountUpdates(txn, batch.UpdatedAccounts); err != nil {
			return tracerr.Wrap(err)
		}

		// Set the EffectiveAmount and EffectiveDepositAmount of all the
		// L1UserTxs that have been forged in this batch
		if err = hdb.setExtraInfoForgedL1UserTxs(txn, batch.L1UserTxs); err != nil {
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

// AddAuctionVars insert auction vars into the DB
func (hdb *HistoryDB) AddAuctionVars(auctionVars *common.AuctionVariables) error {
	return tracerr.Wrap(meddler.Insert(hdb.dbWrite, "auction_vars", auctionVars))
}

// GetTokensTest used to get tokens in a testing context
func (hdb *HistoryDB) GetTokensTest() ([]TokenWithUSD, error) {
	tokens := []*TokenWithUSD{}
	if err := meddler.QueryAll(
		hdb.dbRead, &tokens,
		"SELECT * FROM token ORDER BY token_id ASC",
	); err != nil {
		return nil, tracerr.Wrap(err)
	}
	if len(tokens) == 0 {
		return []TokenWithUSD{}, nil
	}
	return db.SlicePtrsToSlice(tokens).([]TokenWithUSD), nil
}

const (
	// CreateAccountExtraFeePercentage is the multiplication factor over
	// the average fee for CreateAccount that is applied to obtain the
	// recommended fee for CreateAccount
	CreateAccountExtraFeePercentage float64 = 2.5
	// CreateAccountInternalExtraFeePercentage is the multiplication factor
	// over the average fee for CreateAccountInternal that is applied to
	// obtain the recommended fee for CreateAccountInternal
	CreateAccountInternalExtraFeePercentage float64 = 2.0
)

// GetRecommendedFee returns the RecommendedFee information
func (hdb *HistoryDB) GetRecommendedFee(minFeeUSD, maxFeeUSD float64) (*common.RecommendedFee, error) {
	var recommendedFee common.RecommendedFee
	// Get total txs and the batch of the first selected tx of the last hour
	type totalTxsSinceBatchNum struct {
		TotalTxs      int             `meddler:"total_txs"`
		FirstBatchNum common.BatchNum `meddler:"batch_num"`
	}
	ttsbn := &totalTxsSinceBatchNum{}
	if err := meddler.QueryRow(
		hdb.dbRead, ttsbn, `SELECT COUNT(tx.*) as total_txs, 
			COALESCE (MIN(tx.batch_num), 0) as batch_num 
			FROM tx INNER JOIN block ON tx.eth_block_num = block.eth_block_num
			WHERE block.timestamp >= NOW() - INTERVAL '1 HOURS';`,
	); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// Get the amount of batches and acumulated fees for the last hour
	type totalBatchesAndFee struct {
		TotalBatches int     `meddler:"total_batches"`
		TotalFees    float64 `meddler:"total_fees"`
	}
	tbf := &totalBatchesAndFee{}
	if err := meddler.QueryRow(
		hdb.dbRead, tbf, `SELECT COUNT(*) AS total_batches, 
			COALESCE (SUM(total_fees_usd), 0) AS total_fees FROM batch 
			WHERE batch_num > $1;`, ttsbn.FirstBatchNum,
	); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// Update NodeInfo struct
	var avgTransactionFee float64
	if ttsbn.TotalTxs > 0 {
		avgTransactionFee = tbf.TotalFees / float64(ttsbn.TotalTxs)
	} else {
		avgTransactionFee = 0
	}

	recommendedFee.ExistingAccount = math.Min(maxFeeUSD,
		math.Max(avgTransactionFee, minFeeUSD))
	recommendedFee.CreatesAccount = math.Min(maxFeeUSD,
		math.Max(CreateAccountExtraFeePercentage*avgTransactionFee, minFeeUSD))
	recommendedFee.CreatesAccountInternal = math.Min(maxFeeUSD,
		math.Max(CreateAccountInternalExtraFeePercentage*avgTransactionFee, minFeeUSD))
	return &recommendedFee, nil
}

// GetLatestBatches returns the latest forged batches
func (hdb *HistoryDB) GetLatestBatches(numElements int) ([]*common.Batch, error) {
	batchesInfo := []*common.Batch{}
	if err := meddler.QueryAll(
		hdb.dbRead, &batchesInfo, `SELECT batch_num, eth_block_num, forger_addr, fees_collected,
		fee_idxs_coordinator, state_root, num_accounts, last_idx, exit_root, forge_l1_txs_num, slot_num,
		total_fees_usd, eth_tx_hash, gas_price, gas_used, ether_price_usd FROM batch
		ORDER BY batch_num DESC LIMIT `+strconv.Itoa(numElements)+`;`,
	); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return batchesInfo, nil
}
