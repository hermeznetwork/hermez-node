package l2db

import (
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/tracerr"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
	"github.com/russross/meddler"
)

// TODO(Edu): Check DB consistency while there's concurrent use from Coordinator/TxSelector & API

// L2DB stores L2 txs and authorization registers received by the coordinator and keeps them until they are no longer relevant
// due to them being forged or invalid after a safety period
type L2DB struct {
	db           *sqlx.DB
	safetyPeriod common.BatchNum
	ttl          time.Duration
	maxTxs       uint32 // limit of txs that are accepted in the pool
	minFeeUSD    float64
	apiConnCon   *db.APIConnectionController
}

// NewL2DB creates a L2DB.
// To create it, it's needed db connection, safety period expressed in batches,
// maxTxs that the DB should have and TTL (time to live) for pending txs.
func NewL2DB(
	db *sqlx.DB,
	safetyPeriod common.BatchNum,
	maxTxs uint32,
	minFeeUSD float64,
	TTL time.Duration,
	apiConnCon *db.APIConnectionController,
) *L2DB {
	return &L2DB{
		db:           db,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
		minFeeUSD:    minFeeUSD,
		apiConnCon:   apiConnCon,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (l2db *L2DB) DB() *sqlx.DB {
	return l2db.db
}

// AddAccountCreationAuth inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuth(auth *common.AccountCreationAuth) error {
	_, err := l2db.db.Exec(
		`INSERT INTO account_creation_auth (eth_addr, bjj, signature)
		VALUES ($1, $2, $3);`,
		auth.EthAddr, auth.BJJ, auth.Signature,
	)
	return tracerr.Wrap(err)
}

// GetAccountCreationAuth returns an account creation authorization from the DB
func (l2db *L2DB) GetAccountCreationAuth(addr ethCommon.Address) (*common.AccountCreationAuth, error) {
	auth := new(common.AccountCreationAuth)
	return auth, tracerr.Wrap(meddler.QueryRow(
		l2db.db, auth,
		"SELECT * FROM account_creation_auth WHERE eth_addr = $1;",
		addr,
	))
}

// UpdateTxsInfo updates the parameter Info of the pool transactions
func (l2db *L2DB) UpdateTxsInfo(txs []common.PoolL2Tx) error {
	if len(txs) == 0 {
		return nil
	}
	type txUpdate struct {
		ID   common.TxID `db:"id"`
		Info string      `db:"info"`
	}
	txUpdates := make([]txUpdate, len(txs))
	for i := range txs {
		txUpdates[i] = txUpdate{ID: txs[i].TxID, Info: txs[i].Info}
	}
	const query string = `
		UPDATE tx_pool SET
			info = tx_update.info
		FROM (VALUES
			(NULL::::BYTEA, NULL::::VARCHAR),
			(:id, :info)
		) as tx_update (id, info)
		WHERE tx_pool.tx_id = tx_update.id;
	`
	if len(txUpdates) > 0 {
		if _, err := sqlx.NamedExec(l2db.db, query, txUpdates); err != nil {
			return tracerr.Wrap(err)
		}
	}

	return nil
}

// NewPoolL2TxWriteFromPoolL2Tx creates a new PoolL2TxWrite from a PoolL2Tx
func NewPoolL2TxWriteFromPoolL2Tx(tx *common.PoolL2Tx) *PoolL2TxWrite {
	// transform tx from *common.PoolL2Tx to PoolL2TxWrite
	insertTx := &PoolL2TxWrite{
		TxID:      tx.TxID,
		FromIdx:   tx.FromIdx,
		TokenID:   tx.TokenID,
		Amount:    tx.Amount,
		Fee:       tx.Fee,
		Nonce:     tx.Nonce,
		State:     common.PoolL2TxStatePending,
		Signature: tx.Signature,
		RqAmount:  tx.RqAmount,
		Type:      tx.Type,
	}
	if tx.ToIdx != 0 {
		insertTx.ToIdx = &tx.ToIdx
	}
	nilAddr := ethCommon.BigToAddress(big.NewInt(0))
	if tx.ToEthAddr != nilAddr {
		insertTx.ToEthAddr = &tx.ToEthAddr
	}
	if tx.RqFromIdx != 0 {
		insertTx.RqFromIdx = &tx.RqFromIdx
	}
	if tx.RqToIdx != 0 { // if true, all Rq... fields must be different to nil
		insertTx.RqToIdx = &tx.RqToIdx
		insertTx.RqTokenID = &tx.RqTokenID
		insertTx.RqFee = &tx.RqFee
		insertTx.RqNonce = &tx.RqNonce
	}
	if tx.RqToEthAddr != nilAddr {
		insertTx.RqToEthAddr = &tx.RqToEthAddr
	}
	if tx.ToBJJ != common.EmptyBJJComp {
		insertTx.ToBJJ = &tx.ToBJJ
	}
	if tx.RqToBJJ != common.EmptyBJJComp {
		insertTx.RqToBJJ = &tx.RqToBJJ
	}
	f := new(big.Float).SetInt(tx.Amount)
	amountF, _ := f.Float64()
	insertTx.AmountFloat = amountF
	return insertTx
}

// AddTxTest inserts a tx into the L2DB. This is useful for test purposes,
// but in production txs will only be inserted through the API
func (l2db *L2DB) AddTxTest(tx *common.PoolL2Tx) error {
	insertTx := NewPoolL2TxWriteFromPoolL2Tx(tx)
	// insert tx
	return tracerr.Wrap(meddler.Insert(l2db.db, "tx_pool", insertTx))
}

// selectPoolTxCommon select part of queries to get common.PoolL2Tx
const selectPoolTxCommon = `SELECT  tx_pool.tx_id, from_idx, to_idx, tx_pool.to_eth_addr, 
tx_pool.to_bjj, tx_pool.token_id, tx_pool.amount, tx_pool.fee, tx_pool.nonce, 
tx_pool.state, tx_pool.info, tx_pool.signature, tx_pool.timestamp, rq_from_idx, 
rq_to_idx, tx_pool.rq_to_eth_addr, tx_pool.rq_to_bjj, tx_pool.rq_token_id, tx_pool.rq_amount, 
tx_pool.rq_fee, tx_pool.rq_nonce, tx_pool.tx_type, 
(fee_percentage(tx_pool.fee::NUMERIC) * token.usd * tx_pool.amount_f) /
	(10.0 ^ token.decimals::NUMERIC) AS fee_usd, token.usd_update
FROM tx_pool INNER JOIN token ON tx_pool.token_id = token.token_id `

// GetTx  return the specified Tx in common.PoolL2Tx format
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	tx := new(common.PoolL2Tx)
	return tx, tracerr.Wrap(meddler.QueryRow(
		l2db.db, tx,
		selectPoolTxCommon+"WHERE tx_id = $1;",
		txID,
	))
}

// GetPendingTxs return all the pending txs of the L2DB, that have a non NULL AbsoluteFee
func (l2db *L2DB) GetPendingTxs() ([]common.PoolL2Tx, error) {
	var txs []*common.PoolL2Tx
	err := meddler.QueryAll(
		l2db.db, &txs,
		selectPoolTxCommon+"WHERE state = $1",
		common.PoolL2TxStatePending,
	)
	return db.SlicePtrsToSlice(txs).([]common.PoolL2Tx), tracerr.Wrap(err)
}

// StartForging updates the state of the transactions that will begin the forging process.
// The state of the txs referenced by txIDs will be changed from Pending -> Forging
func (l2db *L2DB) StartForging(txIDs []common.TxID, batchNum common.BatchNum) error {
	if len(txIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In(
		`UPDATE tx_pool
		SET state = ?, batch_num = ?
		WHERE state = ? AND tx_id IN (?);`,
		common.PoolL2TxStateForging,
		batchNum,
		common.PoolL2TxStatePending,
		txIDs,
	)
	if err != nil {
		return tracerr.Wrap(err)
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return tracerr.Wrap(err)
}

// DoneForging updates the state of the transactions that have been forged
// so the state of the txs referenced by txIDs will be changed from Forging -> Forged
func (l2db *L2DB) DoneForging(txIDs []common.TxID, batchNum common.BatchNum) error {
	if len(txIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In(
		`UPDATE tx_pool
		SET state = ?, batch_num = ?
		WHERE state = ? AND tx_id IN (?);`,
		common.PoolL2TxStateForged,
		batchNum,
		common.PoolL2TxStateForging,
		txIDs,
	)
	if err != nil {
		return tracerr.Wrap(err)
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return tracerr.Wrap(err)
}

// InvalidateTxs updates the state of the transactions that are invalid.
// The state of the txs referenced by txIDs will be changed from * -> Invalid
func (l2db *L2DB) InvalidateTxs(txIDs []common.TxID, batchNum common.BatchNum) error {
	if len(txIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In(
		`UPDATE tx_pool
		SET state = ?, batch_num = ?
		WHERE tx_id IN (?);`,
		common.PoolL2TxStateInvalid,
		batchNum,
		txIDs,
	)
	if err != nil {
		return tracerr.Wrap(err)
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return tracerr.Wrap(err)
}

// GetPendingUniqueFromIdxs returns from all the pending transactions, the set
// of unique FromIdx
func (l2db *L2DB) GetPendingUniqueFromIdxs() ([]common.Idx, error) {
	var idxs []common.Idx
	rows, err := l2db.db.Query(`SELECT DISTINCT from_idx FROM tx_pool
		WHERE state = $1;`, common.PoolL2TxStatePending)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	defer db.RowsClose(rows)
	var idx common.Idx
	for rows.Next() {
		err = rows.Scan(&idx)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		idxs = append(idxs, idx)
	}
	return idxs, nil
}

var invalidateOldNoncesQuery = fmt.Sprintf(`
		UPDATE tx_pool SET
			state = '%s',
			batch_num = %%d
		FROM (VALUES
			(NULL::::BIGINT, NULL::::BIGINT),
			(:idx, :nonce)
		) as updated_acc (idx, nonce)
		WHERE tx_pool.state = '%s' AND
			tx_pool.from_idx = updated_acc.idx AND
			tx_pool.nonce < updated_acc.nonce;
	`, common.PoolL2TxStateInvalid, common.PoolL2TxStatePending)

// InvalidateOldNonces invalidate txs with nonces that are smaller or equal than their
// respective accounts nonces.  The state of the affected txs will be changed
// from Pending to Invalid
func (l2db *L2DB) InvalidateOldNonces(updatedAccounts []common.IdxNonce, batchNum common.BatchNum) (err error) {
	if len(updatedAccounts) == 0 {
		return nil
	}
	// Fill the batch_num in the query with Sprintf because we are using a
	// named query which works with slices, and doens't handle an extra
	// individual argument.
	query := fmt.Sprintf(invalidateOldNoncesQuery, batchNum)
	if _, err := sqlx.NamedExec(l2db.db, query, updatedAccounts); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// Reorg updates the state of txs that were updated in a batch that has been discarted due to a blockchain reorg.
// The state of the affected txs can change form Forged -> Pending or from Invalid -> Pending
func (l2db *L2DB) Reorg(lastValidBatch common.BatchNum) error {
	_, err := l2db.db.Exec(
		`UPDATE tx_pool SET batch_num = NULL, state = $1
		WHERE (state = $2 OR state = $3 OR state = $4) AND batch_num > $5`,
		common.PoolL2TxStatePending,
		common.PoolL2TxStateForging,
		common.PoolL2TxStateForged,
		common.PoolL2TxStateInvalid,
		lastValidBatch,
	)
	return tracerr.Wrap(err)
}

// Purge deletes transactions that have been forged or marked as invalid for longer than the safety period
// it also deletes pending txs that have been in the L2DB for longer than the ttl if maxTxs has been exceeded
func (l2db *L2DB) Purge(currentBatchNum common.BatchNum) (err error) {
	now := time.Now().UTC().Unix()
	_, err = l2db.db.Exec(
		`DELETE FROM tx_pool WHERE (
			batch_num < $1 AND (state = $2 OR state = $3)
		) OR (
			(SELECT count(*) FROM tx_pool WHERE state = $4) > $5 
			AND timestamp < $6 AND state = $4
		);`,
		currentBatchNum-l2db.safetyPeriod,
		common.PoolL2TxStateForged,
		common.PoolL2TxStateInvalid,
		common.PoolL2TxStatePending,
		l2db.maxTxs,
		time.Unix(now-int64(l2db.ttl.Seconds()), 0),
	)
	return tracerr.Wrap(err)
}

// PurgeByExternalDelete deletes all pending transactions marked with true in
// the `external_delete` column.  An external process can set this column to
// true to instruct the coordinator to delete the tx when possible.
func (l2db *L2DB) PurgeByExternalDelete() error {
	_, err := l2db.db.Exec(
		`DELETE from tx_pool WHERE (external_delete = true AND state = $1);`,
		common.PoolL2TxStatePending,
	)
	return tracerr.Wrap(err)
}
