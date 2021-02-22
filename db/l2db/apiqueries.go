package l2db

import (
	"fmt"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
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
		l2db.db, auth,
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

	row := l2db.db.QueryRow(`SELECT
		($1::NUMERIC * token.usd * fee_percentage($2::NUMERIC)) /
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
		WHERE (SELECT COUNT(*) FROM tx_pool WHERE state = $%v) < $%v;`,
		namesPart, valuesPart,
		len(values)+1, len(values)+2) //nolint:gomnd
	values = append(values, common.PoolL2TxStatePending, l2db.maxTxs)
	res, err := l2db.db.Exec(q, values...)
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
const selectPoolTxAPI = `SELECT  tx_pool.tx_id, hez_idx(tx_pool.from_idx, token.symbol) AS from_idx, tx_pool.effective_from_eth_addr, 
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
		l2db.db, tx,
		selectPoolTxAPI+"WHERE tx_id = $1;",
		txID,
	))
}
