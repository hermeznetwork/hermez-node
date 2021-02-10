package l2db

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
	"github.com/russross/meddler"
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
	row := l2db.db.QueryRow(
		"SELECT COUNT(*) FROM tx_pool WHERE state = $1;",
		common.PoolL2TxStatePending,
	)
	var totalTxs uint32
	if err := row.Scan(&totalTxs); err != nil {
		return tracerr.Wrap(err)
	}
	if totalTxs >= l2db.maxTxs {
		return tracerr.New(
			"The pool is at full capacity. More transactions are not accepted currently",
		)
	}
	return tracerr.Wrap(meddler.Insert(l2db.db, "tx_pool", tx))
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
