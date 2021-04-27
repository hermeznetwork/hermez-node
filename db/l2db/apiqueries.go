package l2db

import (
	"fmt"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
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

// selectPoolTxsAPI select part of queries to get PoolL2TxRead transactions
const selectPoolTxsAPI = `SELECT tx_pool.item_id, tx_pool.tx_id, hez_idx(tx_pool.from_idx, token.symbol) AS from_idx, tx_pool.effective_from_eth_addr, 
tx_pool.effective_from_bjj, hez_idx(tx_pool.to_idx, token.symbol) AS to_idx, tx_pool.effective_to_eth_addr, 
tx_pool.effective_to_bjj, tx_pool.token_id, tx_pool.amount, tx_pool.fee, tx_pool.nonce, 
tx_pool.state, tx_pool.info, tx_pool.signature, tx_pool.timestamp, tx_pool.batch_num, hez_idx(tx_pool.rq_from_idx, token.symbol) AS rq_from_idx, 
hez_idx(tx_pool.rq_to_idx, token.symbol) AS rq_to_idx, tx_pool.rq_to_eth_addr, tx_pool.rq_to_bjj, tx_pool.rq_token_id, tx_pool.rq_amount, 
tx_pool.rq_fee, tx_pool.rq_nonce, tx_pool.tx_type, 
token.item_id AS token_item_id, token.eth_block_num, token.eth_addr, token.name, token.symbol, token.decimals, token.usd, token.usd_update, 
count(*) OVER() AS total_items 
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

// GetPoolTxsAPIRequest is an API request struct for getting txs from the pool
type GetPoolTxsAPIRequest struct {
	EthAddr     *ethCommon.Address
	FromEthAddr *ethCommon.Address
	ToEthAddr   *ethCommon.Address
	Bjj         *babyjub.PublicKeyComp
	FromBjj     *babyjub.PublicKeyComp
	ToBjj       *babyjub.PublicKeyComp
	TxType      *common.TxType
	Idx         *common.Idx
	FromIdx     *common.Idx
	ToIdx       *common.Idx
	State       *common.PoolL2TxState

	FromItem *uint
	Limit    *uint
	Order    string
}

// GetPoolTxsAPI return Txs from the pool
func (l2db *L2DB) GetPoolTxsAPI(request GetPoolTxsAPIRequest) ([]PoolTxAPI, uint64, error) {
	cancel, err := l2db.apiConnCon.Acquire()
	defer cancel()
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	defer l2db.apiConnCon.Release()
	// Apply filters
	nextIsAnd := false
	queryStr := selectPoolTxsAPI
	var args []interface{}
	// ethAddr filter
	if request.EthAddr != nil {
		queryStr += "WHERE (tx_pool.effective_from_eth_addr = ? OR tx_pool.effective_to_eth_addr = ?) "
		nextIsAnd = true
		args = append(args, request.EthAddr, request.EthAddr)
	} else if request.FromEthAddr != nil && request.ToEthAddr != nil {
		queryStr += "WHERE (tx_pool.effective_from_eth_addr = ? AND tx_pool.effective_to_eth_addr = ?) "
		nextIsAnd = true
		args = append(args, request.FromEthAddr, request.ToEthAddr)
	} else if request.FromEthAddr != nil {
		queryStr += "WHERE tx_pool.effective_from_eth_addr = ? "
		nextIsAnd = true
		args = append(args, request.FromEthAddr)
	} else if request.ToEthAddr != nil {
		queryStr += "WHERE tx_pool.effective_to_eth_addr = ? "
		nextIsAnd = true
		args = append(args, request.ToEthAddr)
	} else if request.Bjj != nil {
		queryStr += "WHERE (tx_pool.effective_from_bjj = ? OR tx_pool.effective_to_bjj = ?) "
		nextIsAnd = true
		args = append(args, request.Bjj, request.Bjj)
	} else if request.FromBjj != nil && request.ToBjj != nil {
		queryStr += "WHERE (tx_pool.effective_from_bjj = ? AND tx_pool.effective_to_bjj = ?) "
		nextIsAnd = true
		args = append(args, request.FromBjj, request.ToBjj)
	} else if request.FromBjj != nil {
		queryStr += "WHERE tx_pool.effective_from_bjj = ? "
		nextIsAnd = true
		args = append(args, request.FromBjj)
	} else if request.ToBjj != nil {
		queryStr += "WHERE tx_pool.effective_to_bjj = ? "
		nextIsAnd = true
		args = append(args, request.ToBjj)
	}

	// state filter
	if request.State != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.state = ? "
		args = append(args, request.State)
		nextIsAnd = true
	}

	// txType filter
	if request.TxType != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.tx_type = ? "
		args = append(args, request.TxType)
		nextIsAnd = true
	}

	// account index filter
	if request.Idx != nil {
		if nextIsAnd {
			queryStr += "AND ("
		} else {
			queryStr += "WHERE ("
		}
		queryStr += "tx_pool.from_idx = ? "
		queryStr += "OR tx_pool.to_idx = ?) "
		args = append(args, request.Idx, request.Idx)
		nextIsAnd = true
	} else if request.FromIdx != nil && request.ToIdx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.from_idx = ? AND tx_pool.to_idx = ? "
		args = append(args, request.FromIdx, request.ToIdx)
		nextIsAnd = true
	} else if request.FromIdx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.from_idx = ? "
		args = append(args, request.FromIdx)
		nextIsAnd = true
	} else if request.ToIdx != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		queryStr += "tx_pool.to_idx = ? "
		args = append(args, request.ToIdx)
		nextIsAnd = true
	}
	if request.FromItem != nil {
		if nextIsAnd {
			queryStr += "AND "
		} else {
			queryStr += "WHERE "
		}
		if request.Order == db.OrderAsc {
			queryStr += "tx_pool.item_id >= ? "
		} else {
			queryStr += "tx_pool.item_id <= ? "
		}
		args = append(args, request.FromItem)
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
	if request.Order == db.OrderAsc {
		queryStr += "ASC "
	} else {
		queryStr += "DESC "
	}
	queryStr += fmt.Sprintf("LIMIT %d;", *request.Limit)

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
