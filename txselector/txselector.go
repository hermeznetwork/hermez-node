package txselector

import (
	"sort"

	"github.com/hermeznetwork/hermez-node/txselector/common"
	"github.com/hermeznetwork/hermez-node/txselector/mock"
)

// txs implements the interface Sort for an array of Tx
type txs []common.Tx

func (t txs) Len() int {
	return len(t)
}
func (t txs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t txs) Less(i, j int) bool {
	return t[i].UserFeeAbsolute > t[j].UserFeeAbsolute
}

type TxSelector struct {
	// NMax is the maximum L1-user-tx for a batch
	NMax uint64
	// MMax is the maximum L1-operator-tx for a batch
	MMax uint64
	// PMax is the maximum L2-tx for a batch
	PMax uint64
	// DB is a pointer to the database interface
	DB *mock.MockDB
}

func NewTxSelector(db *mock.MockDB, nMax, mMax, pMax uint64) *TxSelector {
	return &TxSelector{
		NMax: nMax,
		MMax: mMax,
		PMax: pMax,
		DB:   db,
	}
}

func (txsel *TxSelector) updateLocalAccountDB(batchId uint64) error {
	// if batchID > max(localAccountDB.BatchID) + 1
	//     make a checkpoint of AccountDB at BatchID to a localAccountDB
	// use localAccountDB[inputBatchID-1]

	return nil
}

func (txsel *TxSelector) GetL2TxSelection(batchID uint64) ([]common.Tx, error) {
	err := txsel.updateLocalAccountDB(batchID)
	if err != nil {
		return nil, err
	}

	// get pending l2-tx from tx-pool
	txsRaw := txsel.DB.GetTxs(batchID)

	// discard the txs that don't have an Account in the AccountDB
	var validTxs txs
	for _, tx := range txsRaw {
		accountID := getAccountID(tx.ToEthAddr, tx.TokenID)
		if _, ok := txsel.DB.AccountDB[accountID]; ok {
			validTxs = append(validTxs, tx)
		}
	}

	// get most profitable L2-tx (len<NMax)
	txs := txsel.getL2Profitable(validTxs)

	// apply L2-tx to local AccountDB, make checkpoint tagged with BatchID
	//     update balances
	//     update nonces

	// return selected txs
	return txs, nil
}

func (txsel *TxSelector) GetL1L2TxSelection(batchID uint64, l1txs []common.Tx) ([]common.Tx, []common.Tx, []common.Tx, error) {
	err := txsel.updateLocalAccountDB(batchID)
	if err != nil {
		return nil, nil, nil, err
	}

	// apply l1-user-tx to localAccountDB
	//     create new leaves
	//     update balances
	//     update nonces

	// get pending l2-tx from tx-pool
	txsRaw := txsel.DB.GetTxs(batchID)

	// discard the txs that don't have an Account in the AccountDB
	// neither appear in the PendingRegistersDB
	var validTxs txs
	for _, tx := range txsRaw {
		accountID := getAccountID(tx.ToEthAddr, tx.TokenID)
		exist := txsel.checkIfAccountExist(accountID)
		if exist {
			validTxs = append(validTxs, tx)
		}
	}

	// get most profitable L2-tx (len<NMax)
	l2txs := txsel.getL2Profitable(validTxs)

	// prepare (from the selected l2txs) pending to register from the PendingRegistersDB
	var pendingRegisters []common.Account
	for _, tx := range l2txs {
		accountID := getAccountID(tx.ToEthAddr, tx.TokenID)
		if toRegister, ok := txsel.DB.PendingRegistersDB[accountID]; ok {
			pendingRegisters = append(pendingRegisters, toRegister)
		}
	}

	// create L1-operator-tx for each L2-tx selected in which the recipient does not have account
	l1OperatorTxs := txsel.createL1OperatorTxForL2Tx(pendingRegisters) // only with the L2-tx selected ones

	return l1txs, l2txs, l1OperatorTxs, nil
}

func (txsel *TxSelector) checkIfAccountExist(accountID [36]byte) bool {
	// check if account exist in AccountDB
	if _, ok := txsel.DB.AccountDB[accountID]; ok {
		return true
	}
	// check if account is pending to register
	if _, ok := txsel.DB.PendingRegistersDB[accountID]; ok {
		return true
	}
	return false
}

func (txsel *TxSelector) getL2Profitable(txs txs) txs {
	sort.Sort(txs)
	return txs[:txsel.PMax]
}
func (txsel *TxSelector) createL1OperatorTxForL2Tx(accounts []common.Account) txs {
	//
	return txs{}
}
