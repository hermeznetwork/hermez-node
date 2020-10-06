package l2db

import (
	"errors"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
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
	db           *sqlx.DB
	safetyPeriod common.BatchNum
	ttl          time.Duration
	maxTxs       uint32
}

// NewL2DB creates a L2DB.
// To create it, it's needed db connection, safety period expressed in batches,
// maxTxs that the DB should have and TTL (time to live) for pending txs.
func NewL2DB(db *sqlx.DB, safetyPeriod common.BatchNum, maxTxs uint32, TTL time.Duration) *L2DB {
	return &L2DB{
		db:           db,
		safetyPeriod: safetyPeriod,
		ttl:          TTL,
		maxTxs:       maxTxs,
	}
}

// DB returns a pointer to the L2DB.db. This method should be used only for
// internal testing purposes.
func (l2db *L2DB) DB() *sqlx.DB {
	return l2db.db
}

// AddAccountCreationAuth inserts an account creation authorization into the DB
func (l2db *L2DB) AddAccountCreationAuth(auth *common.AccountCreationAuth) error {
	return meddler.Insert(l2db.db, "account_creation_auth", auth)
}

// GetAccountCreationAuth returns an account creation authorization into the DB
func (l2db *L2DB) GetAccountCreationAuth(addr *ethCommon.Address) (*common.AccountCreationAuth, error) {
	if addr == nil {
		return nil, errors.New("addr cannot be nil")
	}
	auth := new(common.AccountCreationAuth)
	return auth, meddler.QueryRow(
		l2db.db, auth,
		"SELECT * FROM account_creation_auth WHERE eth_addr = $1;",
		addr,
	)
}

// AddTxTest inserts a tx into the L2DB. This is useful for test purposes,
// but in production txs will only be inserted through the API (method TBD)
func (l2db *L2DB) AddTxTest(tx *common.PoolL2Tx) error {
	type withouUSD struct {
		TxID        common.TxID          `meddler:"tx_id"`
		FromIdx     common.Idx           `meddler:"from_idx"`
		ToIdx       *common.Idx          `meddler:"to_idx"`
		ToEthAddr   *ethCommon.Address   `meddler:"to_eth_addr"`
		ToBJJ       *babyjub.PublicKey   `meddler:"to_bjj"`
		TokenID     common.TokenID       `meddler:"token_id"`
		Amount      *big.Int             `meddler:"amount,bigint"`
		AmountFloat float64              `meddler:"amount_f"`
		Fee         common.FeeSelector   `meddler:"fee"`
		Nonce       common.Nonce         `meddler:"nonce"`
		State       common.PoolL2TxState `meddler:"state"`
		Signature   *babyjub.Signature   `meddler:"signature"`
		Timestamp   time.Time            `meddler:"timestamp,utctime"`
		BatchNum    *common.BatchNum     `meddler:"batch_num"`
		RqFromIdx   *common.Idx          `meddler:"rq_from_idx"`
		RqToIdx     *common.Idx          `meddler:"rq_to_idx"`
		RqToEthAddr *ethCommon.Address   `meddler:"rq_to_eth_addr"`
		RqToBJJ     *babyjub.PublicKey   `meddler:"rq_to_bjj"`
		RqTokenID   *common.TokenID      `meddler:"rq_token_id"`
		RqAmount    *big.Int             `meddler:"rq_amount,bigintnull"`
		RqFee       *common.FeeSelector  `meddler:"rq_fee"`
		RqNonce     *uint64              `meddler:"rq_nonce"`
		Type        common.TxType        `meddler:"tx_type"`
	}
	return meddler.Insert(l2db.db, "tx_pool", &withouUSD{
		TxID:        tx.TxID,
		FromIdx:     tx.FromIdx,
		ToIdx:       tx.ToIdx,
		ToEthAddr:   tx.ToEthAddr,
		ToBJJ:       tx.ToBJJ,
		TokenID:     tx.TokenID,
		Amount:      tx.Amount,
		AmountFloat: tx.AmountFloat,
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
		State:       tx.State,
		Signature:   tx.Signature,
		Timestamp:   tx.Timestamp,
		BatchNum:    tx.BatchNum,
		RqFromIdx:   tx.RqFromIdx,
		RqToIdx:     tx.RqToIdx,
		RqToEthAddr: tx.RqToEthAddr,
		RqToBJJ:     tx.RqToBJJ,
		RqTokenID:   tx.RqTokenID,
		RqAmount:    tx.RqAmount,
		RqFee:       tx.RqFee,
		RqNonce:     tx.RqNonce,
		Type:        tx.Type,
	})
}

// selectPoolTx select part of queries to get common.PoolL2Tx
const selectPoolTx = `SELECT tx_pool.*, token.usd * tx_pool.amount_f AS value_usd, 
fee_percentage(tx_pool.fee::NUMERIC) * token.usd * tx_pool.amount_f AS fee_usd, token.usd_update  
FROM tx_pool INNER JOIN token ON tx_pool.token_id = token.token_id `

// GetTx return the specified Tx
func (l2db *L2DB) GetTx(txID common.TxID) (*common.PoolL2Tx, error) {
	tx := new(common.PoolL2Tx)
	return tx, meddler.QueryRow(
		l2db.db, tx,
		selectPoolTx+"WHERE tx_id = $1;",
		txID,
	)
}

// GetPendingTxs return all the pending txs of the L2DB, that have a non NULL AbsoluteFee
func (l2db *L2DB) GetPendingTxs() ([]*common.PoolL2Tx, error) {
	var txs []*common.PoolL2Tx
	err := meddler.QueryAll(
		l2db.db, &txs,
		selectPoolTx+"WHERE state = $1 AND token.usd IS NOT NULL",
		common.PoolL2TxStatePending,
	)
	return txs, err
}

// StartForging updates the state of the transactions that will begin the forging process.
// The state of the txs referenced by txIDs will be changed from Pending -> Forging
func (l2db *L2DB) StartForging(txIDs []common.TxID, batchNum common.BatchNum) error {
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
		return err
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return err
}

// DoneForging updates the state of the transactions that have been forged
// so the state of the txs referenced by txIDs will be changed from Forging -> Forged
func (l2db *L2DB) DoneForging(txIDs []common.TxID, batchNum common.BatchNum) error {
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
		return err
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return err
}

// InvalidateTxs updates the state of the transactions that are invalid.
// The state of the txs referenced by txIDs will be changed from * -> Invalid
func (l2db *L2DB) InvalidateTxs(txIDs []common.TxID, batchNum common.BatchNum) error {
	query, args, err := sqlx.In(
		`UPDATE tx_pool
		SET state = ?, batch_num = ?
		WHERE tx_id IN (?);`,
		common.PoolL2TxStateInvalid,
		batchNum,
		txIDs,
	)
	if err != nil {
		return err
	}
	query = l2db.db.Rebind(query)
	_, err = l2db.db.Exec(query, args...)
	return err
}

// CheckNonces invalidate txs with nonces that are smaller or equal than their respective accounts nonces.
// The state of the affected txs will be changed from Pending -> Invalid
func (l2db *L2DB) CheckNonces(updatedAccounts []common.Account, batchNum common.BatchNum) (err error) {
	txn, err := l2db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// Rollback the transaction if there was an error.
		if err != nil {
			err = txn.Rollback()
		}
	}()
	for i := 0; i < len(updatedAccounts); i++ {
		_, err = txn.Exec(
			`UPDATE tx_pool
			SET state = $1, batch_num = $2
			WHERE state = $3 AND from_idx = $4 AND nonce <= $5;`,
			common.PoolL2TxStateInvalid,
			batchNum,
			common.PoolL2TxStatePending,
			updatedAccounts[i].Idx,
			updatedAccounts[i].Nonce,
		)
		if err != nil {
			return err
		}
	}
	return txn.Commit()
}

// Reorg updates the state of txs that were updated in a batch that has been discarted due to a blockchain reorg.
// The state of the affected txs can change form Forged -> Pending or from Invalid -> Pending
func (l2db *L2DB) Reorg(lastValidBatch common.BatchNum) error {
	_, err := l2db.db.Exec(
		`UPDATE tx_pool SET batch_num = NULL, state = $1 
		WHERE (state = $2 OR state = $3) AND batch_num > $4`,
		common.PoolL2TxStatePending,
		common.PoolL2TxStateForged,
		common.PoolL2TxStateInvalid,
		lastValidBatch,
	)
	return err
}

// Purge deletes transactions that have been forged or marked as invalid for longer than the safety period
// it also deletes txs that has been in the L2DB for longer than the ttl if maxTxs has been exceeded
func (l2db *L2DB) Purge(currentBatchNum common.BatchNum) (err error) {
	txn, err := l2db.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		// Rollback the transaction if there was an error.
		if err != nil {
			err = txn.Rollback()
		}
	}()
	// Delete pending txs that have been in the pool after the TTL if maxTxs is reached
	now := time.Now().UTC().Unix()
	_, err = txn.Exec(
		`DELETE FROM tx_pool WHERE (SELECT count(*) FROM tx_pool) > $1 AND timestamp < $2`,
		l2db.maxTxs,
		time.Unix(now-int64(l2db.ttl.Seconds()), 0),
	)
	if err != nil {
		return err
	}
	// Delete txs that have been marked as forged / invalid after the safety period
	_, err = txn.Exec(
		`DELETE FROM tx_pool 
		WHERE batch_num < $1 AND (state = $2 OR state = $3)`,
		currentBatchNum-l2db.safetyPeriod,
		common.PoolL2TxStateForged,
		common.PoolL2TxStateInvalid,
	)
	if err != nil {
		return err
	}
	return txn.Commit()
}
