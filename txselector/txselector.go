package txselector

import (
	"sort"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
)

// txs implements the interface Sort for an array of Tx
type txs []*common.PoolL2Tx

func (t txs) Len() int {
	return len(t)
}
func (t txs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t txs) Less(i, j int) bool {
	return t[i].AbsoluteFee > t[j].AbsoluteFee
}

// TxSelector implements all the functionalities to select the txs for the next batch
type TxSelector struct {
	// MaxL1UserTxs is the maximum L1-user-tx for a batch
	MaxL1UserTxs uint64
	// MaxL1OperatorTxs is the maximum L1-operator-tx for a batch
	MaxL1OperatorTxs uint64
	// MaxTxs is the maximum txs for a batch
	MaxTxs uint64

	l2db            *l2db.L2DB
	localAccountsDB *statedb.LocalStateDB
}

// NewTxSelector returns a *TxSelector
func NewTxSelector(dbpath string, synchronizerStateDB *statedb.StateDB, l2 *l2db.L2DB, maxL1UserTxs, maxL1OperatorTxs, maxTxs uint64) (*TxSelector, error) {
	localAccountsDB, err := statedb.NewLocalStateDB(dbpath, synchronizerStateDB, false, 0) // without merkletree
	if err != nil {
		return nil, err
	}

	return &TxSelector{
		MaxL1UserTxs:     maxL1UserTxs,
		MaxL1OperatorTxs: maxL1OperatorTxs,
		MaxTxs:           maxTxs,
		l2db:             l2,
		localAccountsDB:  localAccountsDB,
	}, nil
}

// Reset tells the TxSelector to get it's internal AccountsDB
// from the required `batchNum`
func (txsel *TxSelector) Reset(batchNum uint64) error {
	err := txsel.localAccountsDB.Reset(batchNum, true)
	if err != nil {
		return err
	}
	return nil
}

// GetL2TxSelection returns a selection of the L2Txs for the next batch, from the L2DB pool
func (txsel *TxSelector) GetL2TxSelection(batchNum uint64) ([]*common.PoolL2Tx, error) {
	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs() // once l2db ready, maybe use parameter 'batchNum'
	if err != nil {
		return nil, err
	}

	// discard the txs that don't have an Account in the AccountDB
	var validTxs txs
	for _, tx := range l2TxsRaw {
		_, err = txsel.localAccountsDB.GetAccount(tx.FromIdx)
		if err == nil {
			// if FromIdx has an account into the AccountsDB
			validTxs = append(validTxs, tx)
		}
	}

	// get most profitable L2-tx
	txs := txsel.getL2Profitable(validTxs, txsel.MaxTxs)

	// apply L2-tx to local AccountDB, make checkpoint tagged with BatchID
	//     update balances
	//     update nonces

	// return selected txs
	return txs, nil
}

// GetL1L2TxSelection returns the selection of L1 + L2 txs
func (txsel *TxSelector) GetL1L2TxSelection(batchNum uint64, l1txs []*common.L1Tx) ([]*common.L1Tx, []*common.L1Tx, []*common.PoolL2Tx, error) {
	// apply l1-user-tx to localAccountDB
	//     create new leaves
	//     update balances
	//     update nonces

	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs() // (batchID)
	if err != nil {
		return nil, nil, nil, err
	}

	// discard the txs that don't have an Account in the AccountDB
	// neither appear in the AccountCreationAuthsDB
	var validTxs txs
	for _, tx := range l2TxsRaw {
		if txsel.checkIfAccountExistOrPending(tx.FromIdx) {
			// if FromIdx has an account into the AccountsDB
			validTxs = append(validTxs, tx)
		}
	}

	// prepare (from the selected l2txs) pending to create from the AccountCreationAuthsDB
	var accountCreationAuths []*common.Account
	// TODO once DB ready:
	// if tx.ToIdx is in AccountCreationAuthsDB, take it and add it to
	// the array 'accountCreationAuths'
	// for _, tx := range l2txs {
	//         account, err := txsel.localAccountsDB.GetAccount(tx.ToIdx)
	//         if err != nil {
	//                 return nil, nil, nil, err
	//         }
	//         if accountToCreate, ok := txsel.DB.AccountCreationAuthsDB[accountID]; ok {
	//                 accountCreationAuths = append(accountCreationAuths, accountToCreate)
	//         }
	// }

	// create L1-operator-tx for each L2-tx selected in which the recipient does not have account
	l1OperatorTxs := txsel.createL1OperatorTxForL2Tx(accountCreationAuths) // only with the L2-tx selected ones

	// get most profitable L2-tx
	maxL2Txs := txsel.MaxTxs - uint64(len(l1OperatorTxs)) // - len(l1UserTxs)
	l2txs := txsel.getL2Profitable(validTxs, maxL2Txs)

	return l1txs, l1OperatorTxs, l2txs, nil
}

func (txsel *TxSelector) checkIfAccountExistOrPending(idx common.Idx) bool {
	// check if account exist in AccountDB
	_, err := txsel.localAccountsDB.GetAccount(idx)
	if err != nil {
		return false
	}
	// check if account is pending to create
	// TODO need a method in the DB to get the PendingRegisters
	// if _, ok := txsel.DB.AccountCreationAuthsDB[accountID]; ok {
	//         return true
	// }
	return false
}

func (txsel *TxSelector) getL2Profitable(txs txs, max uint64) txs {
	sort.Sort(txs)
	return txs[:max]
}
func (txsel *TxSelector) createL1OperatorTxForL2Tx(accounts []*common.Account) []*common.L1Tx {
	//
	return nil
}
