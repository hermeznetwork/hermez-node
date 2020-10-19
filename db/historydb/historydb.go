package historydb

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
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

// BlockData contains the information of a Block
type BlockData struct {
	Block *common.Block
	// Rollup
	// L1UserTxs that were submitted in the block
	L1UserTxs        []common.L1Tx
	Batches          []BatchData
	RegisteredTokens []common.Token
	RollupVars       *common.RollupVars
	// Auction
	Bids                []common.Bid
	Coordinators        []common.Coordinator
	AuctionVars         *common.AuctionVars
	WithdrawDelayerVars *common.WithdrawDelayerVars
	// TODO: enable when common.WithdrawalDelayerVars is Merged from Synchronizer PR
	// WithdrawalDelayerVars *common.WithdrawalDelayerVars
}

// BatchData contains the information of a Batch
type BatchData struct {
	// L1UserTxs that were forged in the batch
	L1Batch          bool // TODO: Remove once Batch.ForgeL1TxsNum is a pointer
	L1UserTxs        []common.L1Tx
	L1CoordinatorTxs []common.L1Tx
	L2Txs            []common.L2Tx
	CreatedAccounts  []common.Account
	ExitTree         []common.ExitInfo
	Batch            *common.Batch
}

// NewBatchData creates an empty BatchData with the slices initialized.
func NewBatchData() *BatchData {
	return &BatchData{
		L1Batch:          false,
		L1UserTxs:        make([]common.L1Tx, 0),
		L1CoordinatorTxs: make([]common.L1Tx, 0),
		L2Txs:            make([]common.L2Tx, 0),
		CreatedAccounts:  make([]common.Account, 0),
		ExitTree:         make([]common.ExitInfo, 0),
		Batch:            &common.Batch{},
	}
}

// NewHistoryDB initialize the DB
func NewHistoryDB(db *sqlx.DB) *HistoryDB {
	return &HistoryDB{db: db}
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
	return meddler.Insert(d, "batch", batch)
}

// AddBatches insert Bids into the DB
func (hdb *HistoryDB) AddBatches(batches []common.Batch) error {
	return hdb.addBatches(hdb.db, batches)
}
func (hdb *HistoryDB) addBatches(d meddler.DB, batches []common.Batch) error {
	// TODO: Calculate and insert total_fees_usd
	return db.BulkInsert(
		d,
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

// GetLastL1TxsNum returns the greatest ForgeL1TxsNum in the DB.  If there's no
// batch in the DB (nil, nil) is returned.
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
	vars *common.AuctionVars,
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

// GetBids return the bids
func (hdb *HistoryDB) GetBids() ([]common.Bid, error) {
	var bids []*common.Bid
	err := meddler.QueryAll(
		hdb.db, &bids,
		"SELECT * FROM bid;",
	)
	return db.SlicePtrsToSlice(bids).([]common.Bid), err
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
func (hdb *HistoryDB) GetToken(tokenID common.TokenID) (*TokenRead, error) {
	token := &TokenRead{}
	err := meddler.QueryRow(
		hdb.db, token, `SELECT * FROM token WHERE token_id = $1;`, tokenID,
	)
	return token, err
}

// GetTokens returns a list of tokens from the DB
func (hdb *HistoryDB) GetTokens(ids []common.TokenID, symbols []string, name string, fromItem, limit *uint, order string) ([]TokenRead, *db.Pagination, error) {
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
	tokens := []*TokenRead{}
	if err := meddler.QueryAll(hdb.db, &tokens, query, argsQ...); err != nil {
		return nil, nil, err
	}
	if len(tokens) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(tokens).([]TokenRead), &db.Pagination{
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
func (hdb *HistoryDB) GetHistoryTx(txID common.TxID) (*HistoryTx, error) {
	tx := &HistoryTx{}
	err := meddler.QueryRow(
		hdb.db, tx, `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
		tx.from_idx, tx.to_idx, tx.amount, tx.token_id, tx.amount_usd, 
		tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
		tx.from_eth_addr, tx.from_bjj, tx.load_amount, 
		tx.load_amount_usd, tx.fee, tx.fee_usd, tx.nonce,
		token.token_id, token.eth_block_num AS token_block,
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
) ([]HistoryTx, *db.Pagination, error) {
	if ethAddr != nil && bjj != nil {
		return nil, nil, errors.New("ethAddr and bjj are incompatible")
	}
	var query string
	var args []interface{}
	queryStr := `SELECT tx.item_id, tx.is_l1, tx.id, tx.type, tx.position, 
	tx.from_idx, tx.to_idx, tx.amount, tx.token_id, tx.amount_usd, 
	tx.batch_num, tx.eth_block_num, tx.to_forge_l1_txs_num, tx.user_origin, 
	tx.from_eth_addr, tx.from_bjj, tx.load_amount, 
	tx.load_amount_usd, tx.fee, tx.fee_usd, tx.nonce,
	token.token_id, token.eth_block_num AS token_block,
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
	log.Debug(query)
	txsPtrs := []*HistoryTx{}
	if err := meddler.QueryAll(hdb.db, &txsPtrs, query, args...); err != nil {
		return nil, nil, err
	}
	txs := db.SlicePtrsToSlice(txsPtrs).([]HistoryTx)
	if len(txs) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return txs, &db.Pagination{
		TotalItems: txs[0].TotalItems,
		FirstItem:  txs[0].FirstItem,
		LastItem:   txs[0].LastItem,
	}, nil
}

// GetExit returns a exit from the DB
func (hdb *HistoryDB) GetExit(batchNum *uint, idx *common.Idx) (*HistoryExit, error) {
	exit := &HistoryExit{}
	err := meddler.QueryRow(
		hdb.db, exit, `SELECT exit_tree.*, token.token_id, token.eth_block_num AS token_block,
		token.eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update
		FROM exit_tree INNER JOIN account ON exit_tree.account_idx = account.idx 
		INNER JOIN token ON account.token_id = token.token_id 
		WHERE exit_tree.batch_num = $1 AND exit_tree.account_idx = $2;`, batchNum, idx,
	)
	return exit, err
}

// GetExits returns a list of exits from the DB and pagination info
func (hdb *HistoryDB) GetExits(
	ethAddr *ethCommon.Address, bjj *babyjub.PublicKey,
	tokenID *common.TokenID, idx *common.Idx, batchNum *uint,
	fromItem, limit *uint, order string,
) ([]HistoryExit, *db.Pagination, error) {
	if ethAddr != nil && bjj != nil {
		return nil, nil, errors.New("ethAddr and bjj are incompatible")
	}
	var query string
	var args []interface{}
	queryStr := `SELECT exit_tree.*, token.token_id, token.eth_block_num AS token_block,
	token.eth_addr, token.name, token.symbol, token.decimals, token.usd,
	token.usd_update, COUNT(*) OVER() AS total_items, MIN(exit_tree.item_id) OVER() AS first_item, MAX(exit_tree.item_id) OVER() AS last_item
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
	exits := []*HistoryExit{}
	if err := meddler.QueryAll(hdb.db, &exits, query, args...); err != nil {
		return nil, nil, err
	}
	if len(exits) == 0 {
		return nil, nil, sql.ErrNoRows
	}
	return db.SlicePtrsToSlice(exits).([]HistoryExit), &db.Pagination{
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

// // GetL1UserTxs gets L1 User Txs to be forged in a batch that will create an account
// // TODO: This is currently not used.  Figure out if it should be used somewhere or removed.
// func (hdb *HistoryDB) GetL1UserTxs(toForgeL1TxsNum int64) ([]*common.Tx, error) {
// 	var txs []*common.Tx
// 	err := meddler.QueryAll(
// 		hdb.db, &txs,
// 		"SELECT * FROM tx WHERE to_forge_l1_txs_num = $1 AND is_l1 = TRUE AND user_origin = TRUE;",
// 		toForgeL1TxsNum,
// 	)
// 	return txs, err
// }

// TODO: Think about chaning all the queries that return a last value, to queries that return the next valid value.

// GetLastTxsPosition for a given to_forge_l1_txs_num
func (hdb *HistoryDB) GetLastTxsPosition(toForgeL1TxsNum int64) (int, error) {
	row := hdb.db.QueryRow("SELECT MAX(position) FROM tx WHERE to_forge_l1_txs_num = $1;", toForgeL1TxsNum)
	var lastL1TxsPosition int
	return lastL1TxsPosition, row.Scan(&lastL1TxsPosition)
}

// AddBlockSCData stores all the information of a block retrieved by the Synchronizer
func (hdb *HistoryDB) AddBlockSCData(blockData *BlockData) (err error) {
	txn, err := hdb.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			errRollback := txn.Rollback()
			if errRollback != nil {
				log.Errorw("Rollback", "err", errRollback)
			}
		}
	}()

	// Add block
	err = hdb.addBlock(txn, blockData.Block)
	if err != nil {
		return err
	}

	// Add Coordinators
	if len(blockData.Coordinators) > 0 {
		err = hdb.addCoordinators(txn, blockData.Coordinators)
		if err != nil {
			return err
		}
	}

	// Add Bids
	if len(blockData.Bids) > 0 {
		err = hdb.addBids(txn, blockData.Bids)
		if err != nil {
			return err
		}
	}

	// Add Tokens
	if len(blockData.RegisteredTokens) > 0 {
		err = hdb.addTokens(txn, blockData.RegisteredTokens)
		if err != nil {
			return err
		}
	}

	// Add l1 Txs
	if len(blockData.L1UserTxs) > 0 {
		err = hdb.addL1Txs(txn, blockData.L1UserTxs)
		if err != nil {
			return err
		}
	}

	// Add Batches
	for _, batch := range blockData.Batches {
		// Add Batch: this will trigger an update on the DB
		// that will set the batch num of forged L1 txs in this batch
		err = hdb.addBatch(txn, batch.Batch)
		if err != nil {
			return err
		}

		// Add unforged l1 Txs
		if batch.L1Batch {
			if len(batch.L1CoordinatorTxs) > 0 {
				err = hdb.addL1Txs(txn, batch.L1CoordinatorTxs)
				if err != nil {
					return err
				}
			}
		}

		// Add l2 Txs
		if len(batch.L2Txs) > 0 {
			err = hdb.addL2Txs(txn, batch.L2Txs)
			if err != nil {
				return err
			}
		}

		// Add accounts
		if len(batch.CreatedAccounts) > 0 {
			err = hdb.addAccounts(txn, batch.CreatedAccounts)
			if err != nil {
				return err
			}
		}

		// Add exit tree
		if len(batch.ExitTree) > 0 {
			err = hdb.addExitTree(txn, batch.ExitTree)
			if err != nil {
				return err
			}
		}

		// TODO: INSERT CONTRACTS VARS
	}

	return txn.Commit()
}
