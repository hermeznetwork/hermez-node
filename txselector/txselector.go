package txselector

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/kvdb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type (
	// CoordAccount contains the data of the Coordinator account, that will be used
	// to create new transactions of CreateAccountDeposit type to add new TokenID
	// accounts for the Coordinator to receive the fees.
	CoordAccount struct {
		Addr                ethCommon.Address
		BJJ                 babyjub.PublicKeyComp
		AccountCreationAuth []byte // signature in byte array format
	}
	// TxSelector implements all the functionalities to select the txs for the next batch
	TxSelector struct {
		l2db            *l2db.L2DB
		localAccountsDB *statedb.LocalStateDB
		coordAccount    CoordAccount
	}
)

// NewTxSelector creates a new *TxSelector object
func NewTxSelector(coordAccount CoordAccount, dbpath string,
	synchronizerStateDB *statedb.StateDB, l2 *l2db.L2DB) (*TxSelector, error) {
	localAccountsDB, err := statedb.NewLocalStateDB(
		statedb.Config{
			Path:    dbpath,
			Keep:    kvdb.DefaultKeep,
			Type:    statedb.TypeTxSelector,
			NLevels: 0,
		},
		synchronizerStateDB) // without the merkle tree
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &TxSelector{
		l2db:            l2,
		localAccountsDB: localAccountsDB,
		coordAccount:    coordAccount,
	}, nil
}

// LocalAccountsDB returns the LocalStateDB of the TxSelector
func (s *TxSelector) LocalAccountsDB() *statedb.LocalStateDB {
	return s.localAccountsDB
}

// Reset tells the TxSelector to get it's internal AccountsDB
// from the required `batchNum`
func (s *TxSelector) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	return tracerr.Wrap(s.localAccountsDB.Reset(batchNum, fromSynchronizer))
}

// GetL2TxSelection returns the L1CoordinatorTxs and a selection of the L2Txs
// for the next batch, from the L2DB pool.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (s *TxSelector) GetL2TxSelection(selectionConfig txprocessor.Config, l1UserFutureTxs []common.L1Tx) ([]common.Idx,
	[][]byte, []common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metric.GetL2TxSelection.Inc()
	log.Debugw("TxSelector: GetL2TxSelection",
		"l1UserFutureTxs", len(l1UserFutureTxs),
	)
	coordIdxs, accCreationAuths, _, l1CoordinatorTxs, l2Txs, discardedL2Txs, err := s.getL1L2TxSelection(
		selectionConfig, []common.L1Tx{}, l1UserFutureTxs)
	return coordIdxs, accCreationAuths, l1CoordinatorTxs, l2Txs, discardedL2Txs, tracerr.Wrap(err)
}

// GetL1L2TxSelection returns the selection of L1 + L2 txs.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (s *TxSelector) GetL1L2TxSelection(selectionConfig txprocessor.Config,
	l1UserTxs, l1UserFutureTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metric.GetL1L2TxSelection.Inc()
	log.Debugw("TxSelector: GetL1L2TxSelection",
		"l1UserTxs", len(l1UserTxs),
		"l1UserFutureTxs", len(l1UserFutureTxs),
	)
	coordIdxs, accCreationAuths, l1UserTxs, l1CoordinatorTxs, l2Txs, discardedL2Txs, err := s.getL1L2TxSelection(
		selectionConfig, l1UserTxs, l1UserFutureTxs)
	return coordIdxs, accCreationAuths, l1UserTxs, l1CoordinatorTxs, l2Txs, discardedL2Txs, tracerr.Wrap(err)
}

// getL1L2TxSelection returns the selection of L1 + L2 txs.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (s *TxSelector) getL1L2TxSelection(selectionConfig txprocessor.Config, l1UserTxs, l1UserFutureTxs []common.L1Tx) (
	[]common.Idx, [][]byte, []common.L1Tx, []common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	processor := txprocessor.NewTxProcessor(s.localAccountsDB.StateDB, selectionConfig)
	processor.AccumulatedFees = make(map[common.Idx]*big.Int)

	// Process L1UserTxs
	for _, tx := range l1UserTxs {
		log.Debugw("TxSelector: processing L1 user tx", "TxID", tx.TxID.String())
		// assumption: l1usertx are sorted by L1Tx.Position
		_, _, _, _, err := processor.ProcessL1Tx(nil, &tx) //nolint:gosec
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	poolTxs, err := s.l2db.GetPendingTxs()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	log.Debugf("TxSelector: pending txs: %d", len(poolTxs))

	batch, err := NewTxBatch(selectionConfig, s.l2db, s.coordAccount, s.localAccountsDB, processor)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	err = batch.createTxGroups(poolTxs, l1UserTxs, l1UserFutureTxs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	err = batch.prune()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	err = batch.processor.StateDB().MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	return batch.getSelection()
}
