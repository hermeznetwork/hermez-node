package l2db

import (
	"fmt"
	"github.com/hermeznetwork/hermez-node/db"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/russross/meddler"
)

var (
	errPoolFull = fmt.Errorf("the pool is at full capacity. More transactions are not accepted currently")
)

// AddAccountCreationAuthAPI inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuthAPI(auth *common.AccountCreationAuth) error {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	return l2db.AddAccountCreationAuth(auth)
}

// GetAccountCreationAuthAPI returns an account creation authorization from the DB
func (l2db *L2DB) GetAccountCreationAuthAPI(addr ethCommon.Address) (*AccountCreationAuthAPI, error) {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	auth := new(AccountCreationAuthAPI)
	return auth, tracerr.Wrap(meddler.QueryRow(
		l2db.dbRead, auth,
		"SELECT * FROM account_creation_auth WHERE eth_addr = $1;",
		addr,
	))
}

// AddTxAPI inserts a tx to the pool
func (l2db *L2DB) AddTxAPI(tx *PoolL2TxWrite) error {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()

	row := l2db.dbRead.QueryRow(`SELECT
		($1::NUMERIC * COALESCE(token.usd, 0) * fee_percentage($2::NUMERIC)) /
			(10.0 ^ token.decimals::NUMERIC)
		FROM token WHERE token.token_id = $3;`,
		tx.AmountFloat, tx.Fee, tx.TokenID)
	var feeUSD float64
	if err := row.Scan(&feeUSD); err != nil {
		return tracerr.Wrap(err)
	}
	if feeUSD < l2db.minFeeUSD {
		return tracerr.Wrap(fmt.Errorf("tx.feeUSD (%v) < minFeeUSD (%v)",
			feeUSD, l2db.minFeeUSD))
	}
	if feeUSD > l2db.maxFeeUSD {
		return tracerr.Wrap(fmt.Errorf("tx.feeUSD (%v) > maxFeeUSD (%v)",
			feeUSD, l2db.maxFeeUSD))
	}

	// Prepare insert SQL query argument parameters
	namesPart, err := meddler.Default.ColumnsQuoted(tx, false)
	if err != nil {
		return err
	}
	valuesPart, err := meddler.Default.PlaceholdersString(tx, false)
	if err != nil {
		return err
	}
	values, err := meddler.Default.Values(tx, false)
	if err != nil {
		return err
	}

	q := fmt.Sprintf(
		`INSERT INTO tx_pool (%s)
		SELECT %s
		WHERE (SELECT COUNT(*) FROM tx_pool WHERE state = $%v AND NOT external_delete) < $%v;`,
		namesPart, valuesPart,
		len(values)+1, len(values)+2) //nolint:gomnd
	values = append(values, common.PoolL2TxStatePending, l2db.maxTxs)
	res, err := l2db.dbWrite.Exec(q, values...)
	if err != nil {
		return tracerr.Wrap(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return tracerr.Wrap(err)
	}
	if rowsAffected == 0 {
		return tracerr.Wrap(errPoolFull)
	}
	return nil
}

// selectPoolTxAPI select part of queries to get PoolL2TxRead
const selectPoolTxAPI = `SELECT tx_pool.item_id, tx_pool.tx_id, hez_idx(tx_pool.from_idx, token.symbol) AS from_idx, tx_pool.effective_from_eth_addr, 
tx_pool.effective_from_bjj, hez_idx(tx_pool.to_idx, token.symbol) AS to_idx, tx_pool.effective_to_eth_addr, 
tx_pool.effective_to_bjj, tx_pool.token_id, tx_pool.amount, tx_pool.fee, tx_pool.nonce, 
tx_pool.state, tx_pool.info, tx_pool.signature, tx_pool.timestamp, tx_pool.batch_num, hez_idx(tx_pool.rq_from_idx, token.symbol) AS rq_from_idx, 
hez_idx(tx_pool.rq_to_idx, token.symbol) AS rq_to_idx, tx_pool.rq_to_eth_addr, tx_pool.rq_to_bjj, tx_pool.rq_token_id, tx_pool.rq_amount, 
tx_pool.rq_fee, tx_pool.rq_nonce, tx_pool.tx_type, 
token.item_id AS token_item_id, token.eth_block_num, token.eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update 
FROM tx_pool INNER JOIN token ON tx_pool.token_id = token.token_id `

// GetTxAPI return the specified Tx in PoolTxAPI format
func (l2db *L2DB) GetTxAPI(txID common.TxID) (*PoolTxAPI, error) {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	tx := new(PoolTxAPI)
	return tx, tracerr.Wrap(meddler.QueryRow(
		l2db.dbRead, tx,
		selectPoolTxAPI+"WHERE tx_id = $1;",
		txID,
	))
}

// GetPoolTxs return Txs from the pool
func (l2db *L2DB) GetPoolTxs(ethAddr, fromEthAddr, toEthAddr *ethCommon.Address,
	bjj, fromBjj, toBjj *babyjub.PublicKeyComp,
	txType *common.TxType, idx, fromIdx, toIdx *common.Idx, state *common.PoolL2TxState,
	fromItem, limit *uint, order string) ([]PoolTxAPI, uint64, error) {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	// Apply filters
	nextIsAnd := false
	queryStr := selectPoolTxAPI
	var args []interface{}
	if state != nil {
		queryStr += "WHERE state = ? "
		args = append(args, state)
		nextIsAnd = true
	}
	// ethAddr filter
	if ethAddr != nil {
		queryStr += "WHERE (tx_pool.effective_from_eth_addr = ? OR tx_pool.effective_to_eth_addr = ?) "
		nextIsAnd = true
		args = append(args, ethAddr, ethAddr)
	} else if fromEthAddr != nil {
		queryStr += "WHERE tx_pool.effective_from_eth_addr = ? "
		nextIsAnd = true
		args = append(args, fromEthAddr)
	} else if toEthAddr != nil {
		queryStr += "WHERE tx_pool.effective_to_eth_addr = ? "
		nextIsAnd = true
		args = append(args, toEthAddr)
	} else if bjj != nil {
		queryStr += "WHERE (tx_pool.effective_from_bjj = ? OR tx_pool.effective_to_bjj = ?) "
		nextIsAnd = true
		args = append(args, bjj, bjj)
	} else if fromBjj != nil {
		queryStr += "WHERE tx_pool.effective_from_bjj = ? "
		nextIsAnd = true
		args = append(args, fromBjj)
	} else if toBjj != nil {
		queryStr += "WHERE tx_pool.effective_to_bjj = ? "
		nextIsAnd = true
		args = append(args, toBjj)
	}

	// txType filter
	if txType != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.tx_type = ? "
		args = append(args, txType)
		nextIsAnd = true
	}

	// account index filter
	if idx != nil {
		if nextIsAnd {
			queryStr += "AND ("
		} else {
			queryStr += "WHERE ("
		}
		queryStr += "tx_pool.from_idx = ? "
		queryStr += "OR tx_pool.to_idx = ?) "
		args = append(args, idx, idx)
		nextIsAnd = true
	} else if fromIdx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.from_idx = ? "
		args = append(args, fromIdx)
		nextIsAnd = true
	} else if toIdx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.to_idx = ? "
		args = append(args, toIdx)
		nextIsAnd = true
	}
	if fromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if order == OrderAsc {
			queryStr += "tx_pool.item_id >= ? "
		} else {
			queryStr += "tx_pool.item_id <= ? "
		}
		args = append(args, fromItem)
		nextIsAnd = true
	}
	if nextIsAnd {
		queryStr += "AND "
	} else {
		queryStr += "WHERE "
	}
	queryStr += "NOT external_delete "

	// pagination
	queryStr += "ORDER BY tx_pool.item_id "
	if order == OrderAsc {
		queryStr += "ASC "
	} else {
		queryStr += "DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *limit)

	query := l2db.dbRead.Rebind(queryStr)
	txsPtrs := []*PoolTxAPI{}
	if err = meddler.QueryAll(
		l2db.dbRead, &txsPtrs,
		query,
		args...); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	txs := db.SlicePtrsToSlice(txsPtrs).([]PoolTxAPI)
	if len(txs) == 0 {
		return txs, 0, nil
	}
	return txs, txs[0].TotalItems - uint64(len(txs)), tracerr.Wrap(err)
}
