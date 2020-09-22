package txselector

// current: very simple version of TxSelector

import (
	"math/big"
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
func (txsel *TxSelector) Reset(batchNum common.BatchNum) error {
	err := txsel.localAccountsDB.Reset(batchNum, true)
	if err != nil {
		return err
	}
	return nil
}

// GetL2TxSelection returns a selection of the L2Txs for the next batch, from the L2DB pool
func (txsel *TxSelector) GetL2TxSelection(batchNum common.BatchNum) ([]*common.PoolL2Tx, error) {
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

	// process the txs in the local AccountsDB
	_, _, err = txsel.localAccountsDB.ProcessTxs(false, false, nil, nil, txs)
	if err != nil {
		return nil, err
	}
	err = txsel.localAccountsDB.MakeCheckpoint()
	return txs, err
}

// GetL1L2TxSelection returns the selection of L1 + L2 txs
func (txsel *TxSelector) GetL1L2TxSelection(batchNum common.BatchNum, l1Txs []*common.L1Tx) ([]*common.L1Tx, []*common.L1Tx, []*common.PoolL2Tx, error) {
	// apply l1-user-tx to localAccountDB
	//     create new leaves
	//     update balances
	//     update nonces

	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs() // (batchID)
	if err != nil {
		return nil, nil, nil, err
	}

	var validTxs txs
	var l1CoordinatorTxs []*common.L1Tx

	// if tx.ToIdx>=256, tx.ToIdx should exist to localAccountsDB, if so,
	// tx is used.  if tx.ToIdx==0, check if tx.ToEthAddr/tx.ToBJJ exist in
	// localAccountsDB, if yes tx is used; if not, check if tx.ToEthAddr is
	// in AccountCreationAuthDB, if so, tx is used and L1CoordinatorTx of
	// CreateAccountAndDeposit is created.
	for i := 0; i < len(l2TxsRaw); i++ {
		if l2TxsRaw[i].ToIdx >= common.IdxUserThreshold {
			_, err = txsel.localAccountsDB.GetAccount(l2TxsRaw[i].ToIdx)
			if err != nil {
				// tx not valid
				continue
			}
			// Account found in the DB, include the l2Tx in the selection
			validTxs = append(validTxs, l2TxsRaw[i])
		} else if l2TxsRaw[i].ToIdx == common.Idx(0) {
			_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2TxsRaw[i].ToEthAddr, l2TxsRaw[i].ToBJJ)
			if err == nil {
				// account for ToEthAddr&ToBJJ already exist,
				// there is no need to create a new one.
				// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
				validTxs = append(validTxs, l2TxsRaw[i])
				continue
			}
			// check if ToEthAddr is in AccountCreationAuths
			_, err = txsel.l2db.GetAccountCreationAuth(l2TxsRaw[i].ToEthAddr) // TODO once l2db.GetAccountCreationAuth is ready, use the value returned as 'accAuth'
			if err != nil {
				// not found, l2Tx will not be added in the selection
				continue
			}
			validTxs = append(validTxs, l2TxsRaw[i])

			// create L1CoordinatorTx for the accountCreation
			l1CoordinatorTx := &common.L1Tx{
				UserOrigin: false,
				// FromEthAddr: accAuth.EthAddr, // TODO This 2 lines will panic, as l2db.GetAccountCreationAuth is not implemented yet and returns nil. Uncomment this 2 lines once l2db method is done.
				// FromBJJ:     accAuth.BJJ,
				TokenID:    l2TxsRaw[i].TokenID,
				LoadAmount: big.NewInt(0),
				Type:       common.TxTypeCreateAccountDeposit,
			}
			l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTx)
		} else if l2TxsRaw[i].ToIdx == common.Idx(1) {
			// valid txs (of Exit type)
			validTxs = append(validTxs, l2TxsRaw[i])
		}
	}

	// get most profitable L2-tx
	maxL2Txs := txsel.MaxTxs - uint64(len(l1CoordinatorTxs)) // - len(l1UserTxs)
	l2Txs := txsel.getL2Profitable(validTxs, maxL2Txs)

	// TODO This 3 lines will panic, as l2db.GetAccountCreationAuth is not implemented yet and returns nil. Uncomment this lines once l2db method is done.
	// process the txs in the local AccountsDB
	// _, _, err = txsel.localAccountsDB.ProcessTxs(false, false, l1Txs, l1CoordinatorTxs, l2Txs)
	// if err != nil {
	//         return nil, nil, nil, err
	// }
	err = txsel.localAccountsDB.MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, err
	}

	return l1Txs, l1CoordinatorTxs, l2Txs, nil
}

func (txsel *TxSelector) getL2Profitable(txs txs, max uint64) txs {
	sort.Sort(txs)
	if len(txs) < int(max) {
		return txs
	}
	return txs[:max]
}
