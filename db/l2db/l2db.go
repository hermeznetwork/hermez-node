/*
Package l2db is responsible for storing and retrieving the data received by the coordinator through the api.
Note that this data will be different for each coordinator in the network, as this represents the L2 information.

The data managed by this package is fundamentally PoolL2Tx and AccountCreationAuth. All this data come from
the API sent by clients and is used by the txselector to decide which transactions are selected to forge a batch.

Some of the database tooling used in this package such as meddler and migration tools is explained in the db package.

This package is spitted in different files following these ideas:
- l2db.go: constructor and functions used by packages other than the api.
- apiqueries.go: functions used by the API, the queries implemented in this functions use a semaphore
to restrict the maximum concurrent connections to the database.
- views.go: structs used to retrieve/store data from/to the database. When possible, the common structs are used, however
most of the time there is no 1:1 relation between the struct fields and the tables of the schema, especially when joining tables.
In some cases, some of the structs defined in this file also include custom Marshallers to easily match the expected api formats.
*/
package l2db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"

	//nolint:errcheck // driver for postgres DB
	_ "github.com/lib/pq"
	"github.com/russross/meddler"
)

// L2DB stores L2 txs and authorization registers received by the coordinator and keeps them until they are no longer relevant
// due to them being forged or invalid after a safety period
type L2DB struct {
	dbRead       *sqlx.DB
	dbWrite      *sqlx.DB
	safetyPeriod common.BatchNum
	ttl          time.Duration
	maxTxs       uint32 // limit of txs that are accepted in the pool
	minFeeUSD    float64
	maxFeeUSD    float64
	apiConnCon   *db.APIConnectionController
}

// NewL2DB creates a L2DB.
// To create it, it's needed db connection, safety period expressed in batches,
// maxTxs that the DB should have and TTL (time to live) for pending txs.
func NewL2DB(
	dbRead, dbWrite *sqlx.DB,
	safetyPeriod common.BatchNum,
	maxTxs uint32,
	minFeeUSD float64,
	maxFeeUSD float64,
	TTL time.Duration,
	apiConnCon *db.APIConnectionController,
) *L2DB {
	return &L2DB{
		dbRead:       dbRead,
		dbWrite:      dbWrite,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
		minFeeUSD:    minFeeUSD,
		maxFeeUSD:    maxFeeUSD,
		apiConnCon:   apiConnCon,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (l2db *L2DB) DB() *sqlx.DB {
	return l2db.dbWrite
}

// MinFeeUSD returns the minimum fee in USD that is required to accept txs into
// the pool
func (l2db *L2DB) MinFeeUSD() float64 {
	return l2db.minFeeUSD
}

// AddAccountCreationAuth inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuth(auth *common.AccountCreationAuth) error {
	_, err := l2db.dbWrite.Exec(
		`INSERT INTO account_creation_auth (eth_addr, bjj, signature)
		VALUES ($1, $2, $3);`,
		auth.EthAddr, auth.BJJ, auth.Signature,
	)
	return tracerr.Wrap(err)
}

// AddManyAccountCreationAuth inserts a batch of accounts creation authorization
// if not exist into the DB
func (l2db *L2DB) AddManyAccountCreationAuth(auths []common.AccountCreationAuth) error {
	_, err := sqlx.NamedExec(l2db.dbWrite,
		`INSERT INTO account_creation_auth (eth_addr, bjj, signature)
				VALUES (:ethaddr, :bjj, :signature) 
				ON CONFLICT (eth_addr) DO NOTHING`, auths)
	return tracerr.Wrap(err)
}

// GetAccountCreationAuth returns an account creation authorization from the DB
func (l2db *L2DB) GetAccountCreationAuth(addr ethCommon.Address) (*common.AccountCreationAuth, error) {
	auth := new(common.AccountCreationAuth)
	return auth, tracerr.Wrap(meddler.QueryRow(
		l2db.dbRead, auth,
		"SELECT * FROM account_creation_auth WHERE eth_addr = $1;",
		addr,
	))
}

// UpdateTxsInfo updates the parameter Info of the pool transactions
func (l2db *L2DB) UpdateTxsInfo(txs []common.PoolL2Tx, batchNum common.BatchNum) error {
	if len(txs) == 0 {
		return nil
	}

	const query string = `
		UPDATE tx_pool SET
			info = $2,
			error_code = $3,
			error_type = $4
		WHERE tx_pool.tx_id = $1;
	`

	batchN := strconv.FormatInt(int64(batchNum), 10)

	tx, err := l2db.dbWrite.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return tracerr.Wrap(err)
	}

	for i := range txs {
		info := "BatchNum: " + batchN + ". " + txs[i].Info

		if _, err := tx.Exec(query, txs[i].TxID, info, txs[i].ErrorCode, txs[i].ErrorType); err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				return tracerr.Wrap(fmt.Errorf("failed to rollback tx update: %v. error triggering rollback: %v", err, errRb))
			}
			return tracerr.Wrap(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}

	return nil
}

// AddTxTest inserts a tx into the L2DB, without security checks. This is useful for test purposes,
func (l2db *L2DB) AddTxTest(tx *common.PoolL2Tx) error {
	// Add tx without checking if pool is full
	return tracerr.Wrap(
		l2db.addTxs([]common.PoolL2Tx{*tx}, false),
	)
}

// Insert PoolL2Tx transactions into the pool. If checkPoolIsFull is set to true the insert will
// fail if the pool is fool and errPoolFull will be returned
func (l2db *L2DB) addTxs(txs []common.PoolL2Tx, checkPoolIsFull bool) error {
	// Set the columns that will be affected by the insert on the table
	const queryInsertPart = `INSERT INTO tx_pool (
		tx_id, from_idx, to_idx, to_eth_addr, to_bjj, token_id,
		amount, fee, nonce, state, info, signature, rq_from_idx, 
		rq_to_idx, rq_to_eth_addr, rq_to_bjj, rq_token_id, rq_amount, rq_fee, rq_nonce, 
		tx_type, amount_f, client_ip, rq_offset, atomic_group_id, max_num_batch
	)`
	var (
		queryVarsPart string
		queryVars     []interface{}
	)
	for i := range txs {
		// Format extra DB fields and nullables
		var (
			toEthAddr *ethCommon.Address
			toBJJ     *babyjub.PublicKeyComp
			// Info (always nil)
			info *string
			// Rq fields, nil unless tx.RqFromIdx != 0
			rqFromIdx     *common.Idx
			rqToIdx       *common.Idx
			rqToEthAddr   *ethCommon.Address
			rqToBJJ       *babyjub.PublicKeyComp
			rqTokenID     *common.TokenID
			rqAmount      *string
			rqFee         *common.FeeSelector
			rqNonce       *nonce.Nonce
			rqOffset      *uint8
			atomicGroupID *common.AtomicGroupID
			maxNumBatch   *uint32
		)
		// AmountFloat
		f := new(big.Float).SetInt((*big.Int)(txs[i].Amount))
		amountF, _ := f.Float64()
		// ToEthAddr
		if txs[i].ToEthAddr != common.EmptyAddr {
			toEthAddr = &txs[i].ToEthAddr
		}
		// ToBJJ
		if txs[i].ToBJJ != common.EmptyBJJComp {
			toBJJ = &txs[i].ToBJJ
		}
		// MAxNumBatch
		if txs[i].MaxNumBatch != 0 {
			maxNumBatch = &txs[i].MaxNumBatch
		}
		// Rq fields
		if txs[i].RqFromIdx != 0 {
			// RqFromIdx
			rqFromIdx = &txs[i].RqFromIdx
			// RqToIdx
			if txs[i].RqToIdx != 0 {
				rqToIdx = &txs[i].RqToIdx
			}
			// RqToEthAddr
			if txs[i].RqToEthAddr != common.EmptyAddr {
				rqToEthAddr = &txs[i].RqToEthAddr
			}
			// RqToBJJ
			if txs[i].RqToBJJ != common.EmptyBJJComp {
				rqToBJJ = &txs[i].RqToBJJ
			}
			// RqTokenID
			rqTokenID = &txs[i].RqTokenID
			// RqAmount
			if txs[i].RqAmount != nil {
				rqAmountStr := txs[i].RqAmount.String()
				rqAmount = &rqAmountStr
			}
			// RqFee
			rqFee = &txs[i].RqFee
			// RqNonce
			rqNonce = &txs[i].RqNonce
			// RqOffset
			rqOffset = &txs[i].RqOffset
			// AtomicGroupID
			atomicGroupID = &txs[i].AtomicGroupID
		}
		// Each ? match one of the columns to be inserted as defined in queryInsertPart
		const queryVarsPartPerTx = `(?::BYTEA, ?::BIGINT, ?::BIGINT, ?::BYTEA, ?::BYTEA, ?::INT, 
		?::NUMERIC, ?::SMALLINT, ?::BIGINT, ?::CHAR(4), ?::VARCHAR, ?::BYTEA, ?::BIGINT,
		?::BIGINT, ?::BYTEA, ?::BYTEA, ?::INT, ?::NUMERIC, ?::SMALLINT, ?::BIGINT,
		?::VARCHAR(40), ?::NUMERIC, ?::VARCHAR, ?::SMALLINT, ?::BYTEA, ?::BIGINT)`
		if i == 0 {
			queryVarsPart += queryVarsPartPerTx
		} else {
			// Add coma before next tx values.
			queryVarsPart += ", " + queryVarsPartPerTx
		}
		// Add values that will replace the ?
		queryVars = append(queryVars,
			txs[i].TxID, txs[i].FromIdx, txs[i].ToIdx, toEthAddr, toBJJ, txs[i].TokenID,
			txs[i].Amount.String(), txs[i].Fee, txs[i].Nonce, txs[i].State, info, txs[i].Signature, rqFromIdx,
			rqToIdx, rqToEthAddr, rqToBJJ, rqTokenID, rqAmount, rqFee, rqNonce,
			txs[i].Type, amountF, txs[i].ClientIP, rqOffset, atomicGroupID, maxNumBatch,
		)
	}
	// Query begins with the insert statement
	query := queryInsertPart
	if checkPoolIsFull {
		// This query creates a temporary table containing the values to insert
		// that will only get selected if the pool is not full
		query += " SELECT * FROM ( VALUES " + queryVarsPart + " ) as tmp " + // Temporary table with the values of the txs
			" WHERE (SELECT COUNT (*) FROM tx_pool WHERE state = ? AND NOT external_delete) < ?;" // Check if the pool is full
		queryVars = append(queryVars, common.PoolL2TxStatePending, l2db.maxTxs)
	} else {
		query += " VALUES " + queryVarsPart + ";"
	}
	// Replace "?, ?, ... ?" ==> "$1, $2, ..., $(len(queryVars))"
	query = l2db.dbRead.Rebind(query)
	// Execute query
	res, err := l2db.dbWrite.Exec(query, queryVars...)
	if err == nil && checkPoolIsFull {
		if rowsAffected, err := res.RowsAffected(); err != nil || rowsAffected == 0 {
			// If the query didn't affect any row, and there is no error in the query
			// it's safe to assume that the WERE clause wasn't true, and so the pool is full
			return tracerr.Wrap(errPoolFull)
		}
	}
	return tracerr.Wrap(err)
}

// Update PoolL2Tx transaction in the pool
func (l2db *L2DB) updateTx(tx common.PoolL2Tx) error {
	const queryUpdate = `UPDATE tx_pool SET to_idx = ?, to_eth_addr = ?, to_bjj = ?, max_num_batch = ?, 
	signature = ?, client_ip = ?, tx_type = ? WHERE tx_id = ? AND tx_pool.atomic_group_id IS NULL;`

	if tx.ToIdx == 0 && tx.ToEthAddr == common.EmptyAddr && tx.ToBJJ == common.EmptyBJJComp && tx.MaxNumBatch == 0 {
		return tracerr.Wrap(errors.New("nothing to update"))
	}

	queryVars := []interface{}{tx.ToIdx, tx.ToEthAddr, tx.ToBJJ, tx.MaxNumBatch, tx.Signature, tx.ClientIP, tx.Type, tx.TxID}

	query, args, err := sqlx.In(queryUpdate, queryVars...)
	if err != nil {
		return tracerr.Wrap(err)
	}

	query = l2db.dbWrite.Rebind(query)
	_, err = l2db.dbWrite.Exec(query, args...)
	return tracerr.Wrap(err)
}

func (l2db *L2DB) updateTxByIdxAndNonce(idx common.Idx, nonce nonce.Nonce, tx *common.PoolL2Tx) error {
	txn, err := l2db.dbWrite.Beginx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	defer func() {
		if err != nil {
			db.Rollback(txn)
		}
	}()
	var (
		res          sql.Result
		queryVars    []interface{}
		rowsAffected int64
	)

	const queryDelete = `DELETE FROM tx_pool WHERE from_idx = $1 AND nonce = $2 AND (state = $3 OR state = $4) AND atomic_group_id IS NULL AND NOT external_delete;`
	if res, err = txn.Exec(queryDelete, idx, nonce, common.PoolL2TxStatePending, common.PoolL2TxStateInvalid); err != nil {
		return tracerr.Wrap(err)
	}

	if rowsAffected, err = res.RowsAffected(); err != nil || rowsAffected == 0 {
		return tracerr.Wrap(sql.ErrNoRows)
	}

	const queryInsertPart = `INSERT INTO tx_pool (
		tx_id, from_idx, to_idx, to_eth_addr, to_bjj, token_id,
		amount, fee, nonce, state, info, signature, 
		tx_type, amount_f, client_ip, max_num_batch
	)`

	var (
		toEthAddr   *ethCommon.Address
		toBJJ       *babyjub.PublicKeyComp
		info        *string
		maxNumBatch *uint32
	)

	// AmountFloat
	f := new(big.Float).SetInt(tx.Amount)
	amountF, _ := f.Float64()
	// ToEthAddr
	if tx.ToEthAddr != common.EmptyAddr {
		toEthAddr = &tx.ToEthAddr
	}
	// ToBJJ
	if tx.ToBJJ != common.EmptyBJJComp {
		toBJJ = &tx.ToBJJ
	}
	// MAxNumBatch
	if tx.MaxNumBatch != 0 {
		maxNumBatch = &tx.MaxNumBatch
	}

	queryVarsPart := `(?::BYTEA, ?::BIGINT, ?::BIGINT, ?::BYTEA, ?::BYTEA, ?::INT,
	?::NUMERIC, ?::SMALLINT, ?::BIGINT, ?::CHAR(4), ?::VARCHAR, ?::BYTEA,
	?::VARCHAR(40), ?::NUMERIC, ?::VARCHAR, ?::BIGINT)`

	queryVars = append(queryVars,
		tx.TxID, tx.FromIdx, tx.ToIdx, toEthAddr, toBJJ, tx.TokenID,
		tx.Amount.String(), tx.Fee, tx.Nonce, tx.State, info, tx.Signature,
		tx.Type, amountF, tx.ClientIP, maxNumBatch)

	query := queryInsertPart + " VALUES " + queryVarsPart + ";"
	query = txn.Rebind(query)
	res, err = txn.Exec(query, queryVars...)
	if err != nil {
		return tracerr.Wrap(err)
	}
	if rowsAffected, err = res.RowsAffected(); err != nil || rowsAffected == 0 {
		return tracerr.Wrap(sql.ErrNoRows)
	}
	return tracerr.Wrap(txn.Commit())
}

// selectPoolTxCommon select part of queries to get common.PoolL2Tx
const selectPoolTxCommon = `SELECT  tx_pool.tx_id, from_idx, to_idx, tx_pool.to_eth_addr, 
tx_pool.to_bjj, tx_pool.token_id, tx_pool.amount, tx_pool.fee, tx_pool.nonce, 
tx_pool.state, tx_pool.info, tx_pool.signature, tx_pool.timestamp, rq_from_idx, 
rq_to_idx, tx_pool.rq_to_eth_addr, tx_pool.rq_to_bjj, tx_pool.rq_token_id, tx_pool.rq_amount, 
tx_pool.rq_fee, tx_pool.rq_nonce, tx_pool.tx_type, tx_pool.rq_offset, tx_pool.atomic_group_id, tx_pool.max_num_batch, 
(fee_percentage(tx_pool.fee::NUMERIC) * token.usd * tx_pool.amount_f) /
	(10.0 ^ token.decimals::NUMERIC) AS fee_usd, token.usd_update
FROM tx_pool INNER JOIN token ON tx_pool.token_id = token.token_id `

// GetTx  return the specified Tx in common.PoolL2Tx format
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	tx := new(common.PoolL2Tx)
	return tx, tracerr.Wrap(meddler.QueryRow(
		l2db.dbRead, tx,
		selectPoolTxCommon+"WHERE tx_id = $1;",
		txID,
	))
}

// GetPendingTxs return all the pending txs of the L2DB, that have a non NULL AbsoluteFee
func (l2db *L2DB) GetPendingTxs() ([]common.PoolL2Tx, error) {
	var txs []*common.PoolL2Tx
	err := meddler.QueryAll(
		l2db.dbRead, &txs,
		selectPoolTxCommon+"WHERE state = $1 AND NOT external_delete ORDER BY tx_pool.item_id ASC;",
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
	query = l2db.dbWrite.Rebind(query)
	_, err = l2db.dbWrite.Exec(query, args...)
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
	query = l2db.dbWrite.Rebind(query)
	_, err = l2db.dbWrite.Exec(query, args...)
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
	query = l2db.dbWrite.Rebind(query)
	_, err = l2db.dbWrite.Exec(query, args...)
	return tracerr.Wrap(err)
}

// GetPendingUniqueFromIdxs returns from all the pending transactions, the set
// of unique FromIdx
func (l2db *L2DB) GetPendingUniqueFromIdxs() ([]common.Idx, error) {
	var idxs []common.Idx
	rows, err := l2db.dbRead.Query(`SELECT DISTINCT from_idx FROM tx_pool
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

const invalidateOldNoncesInfo = `Nonce is smaller than account nonce`

var invalidateOldNoncesQuery = fmt.Sprintf(`
		UPDATE tx_pool SET
			state = '%s',
			info = '%s',
			batch_num = %%d
		FROM (VALUES
			(NULL::::BIGINT, NULL::::BIGINT),
			(:idx, :nonce)
		) as updated_acc (idx, nonce)
		WHERE tx_pool.state = '%s' AND
			tx_pool.from_idx = updated_acc.idx AND
			tx_pool.nonce < updated_acc.nonce;
	`, common.PoolL2TxStateInvalid, invalidateOldNoncesInfo, common.PoolL2TxStatePending)

// InvalidateOldNonces invalidate txs with nonces that are smaller or equal than their
// respective accounts nonces.  The state of the affected txs will be changed
// from Pending to Invalid
func (l2db *L2DB) InvalidateOldNonces(updatedAccounts []common.IdxNonce, batchNum common.BatchNum) (err error) {
	if len(updatedAccounts) == 0 {
		return nil
	}
	// Fill the batch_num in the query with Sprintf because we are using a
	// named query which works with slices, and doesn't handle an extra
	// individual argument.
	query := fmt.Sprintf(invalidateOldNoncesQuery, batchNum)
	if _, err := sqlx.NamedExec(l2db.dbWrite, query, updatedAccounts); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// Reorg updates the state of txs that were updated in a batch that has been discarted due to a blockchain reorg.
// The state of the affected txs can change form Forged -> Pending or from Invalid -> Pending
func (l2db *L2DB) Reorg(lastValidBatch common.BatchNum) error {
	_, err := l2db.dbWrite.Exec(
		`UPDATE tx_pool SET batch_num = NULL, state = $1, info = NULL
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
	_, err = l2db.dbWrite.Exec(
		`DELETE FROM tx_pool WHERE (
			batch_num < $1 AND (state = $2 OR state = $3)
		) OR (
			state = $4 AND timestamp < $5
		) OR (
			max_num_batch < $1
		);`,
		currentBatchNum-l2db.safetyPeriod,
		common.PoolL2TxStateForged,
		common.PoolL2TxStateInvalid,
		common.PoolL2TxStatePending,
		time.Unix(now-int64(l2db.ttl.Seconds()), 0),
	)
	return tracerr.Wrap(err)
}

// PurgeByExternalDelete deletes all pending transactions marked with true in
// the `external_delete` column.  An external process can set this column to
// true to instruct the coordinator to delete the tx when possible.
func (l2db *L2DB) PurgeByExternalDelete() error {
	_, err := l2db.dbWrite.Exec(
		`DELETE from tx_pool WHERE (external_delete = true AND state = $1);`,
		common.PoolL2TxStatePending,
	)
	return tracerr.Wrap(err)
}
