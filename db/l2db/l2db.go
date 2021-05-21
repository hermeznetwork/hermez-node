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
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"time"

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

// TODO(Edu): Check DB consistency while there's concurrent use from Coordinator/TxSelector & API

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
	type txUpdate struct {
		ID   common.TxID `db:"id"`
		Info string      `db:"info"`
	}
	txUpdates := make([]txUpdate, len(txs))
	batchN := strconv.FormatInt(int64(batchNum), 10)
	for i := range txs {
		txUpdates[i] = txUpdate{ID: txs[i].TxID, Info: "BatchNum: " + batchN + ". " + txs[i].Info}
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
		if _, err := sqlx.NamedExec(l2db.dbWrite, query, txUpdates); err != nil {
			return tracerr.Wrap(err)
		}
	}

	return nil
}

// AddTxTest inserts a tx into the L2DB, without security checks. This is useful for test purposes,
func (l2db *L2DB) AddTxTest(tx *common.PoolL2Tx) error {
	// Add tx without checking if pool is full
	_, err := l2db.addTx(tx, false)
	return err
}

func (l2db *L2DB) addTx(tx *common.PoolL2Tx, checkPoolIsFull bool) (sql.Result, error) {
	// Prepare extra DB fields and nullables
	var (
		toEthAddr *ethCommon.Address
		toBJJ     *babyjub.PublicKeyComp
		// Info (always nil)
		info *string
		// Rq fields, nil unless tx.RqFromIdx != 0
		rqFromIdx   *common.Idx
		rqToIdx     *common.Idx
		rqToEthAddr *ethCommon.Address
		rqToBJJ     *babyjub.PublicKeyComp
		rqTokenID   *common.TokenID
		rqAmount    *string
		rqFee       *common.FeeSelector
		rqNonce     *common.Nonce
	)
	// AmountFloat
	f := new(big.Float).SetInt((*big.Int)(tx.Amount))
	amountF, _ := f.Float64()
	// ToEthAddr
	if tx.ToEthAddr != common.EmptyAddr {
		toEthAddr = &tx.ToEthAddr
	}
	// ToBJJ
	if tx.ToBJJ != common.EmptyBJJComp {
		toBJJ = &tx.ToBJJ
	}
	// Rq fields
	if tx.RqFromIdx != 0 {
		// RqFromIdx
		rqFromIdx = &tx.RqFromIdx
		// RqToIdx
		if tx.RqToIdx != 0 {
			rqToIdx = &tx.RqToIdx
		}
		// RqToEthAddr
		if tx.RqToEthAddr != common.EmptyAddr {
			rqToEthAddr = &tx.RqToEthAddr
		}
		// RqToBJJ
		if tx.RqToBJJ != common.EmptyBJJComp {
			rqToBJJ = &tx.RqToBJJ
		}
		// RqTokenID
		rqTokenID = &tx.RqTokenID
		// RqAmount
		if tx.RqAmount != nil {
			rqAmountStr := tx.RqAmount.String()
			rqAmount = &rqAmountStr
		}
		// RqFee
		rqFee = &tx.RqFee
		// RqNonce
		rqNonce = &tx.RqNonce
	}
	const queryInsertPart = `INSERT INTO tx_pool (
		tx_id, from_idx, to_idx, to_eth_addr, to_bjj, token_id,
		amount, fee, nonce, state, info, signature, rq_from_idx, 
		rq_to_idx, rq_to_eth_addr, rq_to_bjj, rq_token_id, rq_amount, rq_fee, rq_nonce, 
		tx_type, amount_f, client_ip
	)`
	const queryVarsPart = `$1, $2, $3, $4, $5, $6, 
	$7, $8, $9, $10, $11, $12, $13,
	$14, $15, $16, $17, $18, $19, $20,
	$21, $22, $23`
	queryVars := []interface{}{tx.TxID, tx.FromIdx, tx.ToIdx, toEthAddr, toBJJ, tx.TokenID,
		tx.Amount.String(), tx.Fee, tx.Nonce, tx.State, info, tx.Signature, rqFromIdx,
		rqToIdx, rqToEthAddr, rqToBJJ, rqTokenID, rqAmount, rqFee, rqNonce,
		tx.Type, amountF, tx.ClientIP}
	var query string
	if checkPoolIsFull {
		query = queryInsertPart +
			"SELECT " +
			queryVarsPart +
			"WHERE (SELECT COUNT (*) FROM tx_pool WHERE state = $24 AND NOT external_delete) < $25;"
		queryVars = append(queryVars, common.PoolL2TxStatePending, l2db.maxTxs)
	} else {
		query = queryInsertPart + "VALUES(" + queryVarsPart + ");"
	}
	res, err := l2db.dbWrite.Exec(query, queryVars...)
	return res, tracerr.Wrap(err)
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
		selectPoolTxCommon+"WHERE state = $1 AND NOT external_delete;",
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
