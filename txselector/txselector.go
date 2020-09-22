package txselector

// current: very simple version of TxSelector

import (
	"bytes"
	"math/big"
	"sort"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/iden3/go-iden3-crypto/babyjub"
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
				log.Debugw("invalid L2Tx: ToIdx not found in StateDB", "ToIdx", l2TxsRaw[i].ToIdx)
				continue
			}
			// Account found in the DB, include the l2Tx in the selection
			validTxs = append(validTxs, l2TxsRaw[i])
		} else if l2TxsRaw[i].ToIdx == common.Idx(0) {
			if checkAlreadyPendingToCreate(l1CoordinatorTxs, l2TxsRaw[i].ToEthAddr, l2TxsRaw[i].ToBJJ) {
				// if L2Tx needs a new L1CoordinatorTx of CreateAccount type,
				// and a previous L2Tx in the current process already created
				// a L1CoordinatorTx of this type, in the DB there still seem
				// that needs to create a new L1CoordinatorTx, but as is already
				// created, the tx is valid
				validTxs = append(validTxs, l2TxsRaw[i])
				continue
			}

			if !bytes.Equal(l2TxsRaw[i].ToEthAddr.Bytes(), common.EmptyAddr.Bytes()) &&
				!bytes.Equal(l2TxsRaw[i].ToEthAddr.Bytes(), common.FFAddr.Bytes()) {
				// case: ToEthAddr != 0x00 neither 0xff
				var accAuth *common.AccountCreationAuth
				if l2TxsRaw[i].ToBJJ != nil {
					// case: ToBJJ!=0:
					// if idx exist for EthAddr&BJJ use it
					_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2TxsRaw[i].ToEthAddr, l2TxsRaw[i].ToBJJ)
					if err == nil {
						// account for ToEthAddr&ToBJJ already exist,
						// there is no need to create a new one.
						// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
						validTxs = append(validTxs, l2TxsRaw[i])
						continue
					}
					// if not, check if AccountCreationAuth exist for that ToEthAddr&BJJ
					// accAuth, err = txsel.l2db.GetAccountCreationAuthBJJ(l2TxsRaw[i].ToEthAddr, l2TxsRaw[i].ToBJJ)
					accAuth, err = txsel.l2db.GetAccountCreationAuth(l2TxsRaw[i].ToEthAddr)
					if err != nil {
						// not found, l2Tx will not be added in the selection
						log.Debugw("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr & ToBJJ found in AccountCreationAuths L2DB", "ToIdx", l2TxsRaw[i].ToIdx, "ToEthAddr", l2TxsRaw[i].ToEthAddr)
						continue
					}
					if accAuth.BJJ != l2TxsRaw[i].ToBJJ {
						// if AccountCreationAuth.BJJ is not the same than in the tx, tx is not accepted
						log.Debugw("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr & ToBJJ found in AccountCreationAuths L2DB", "ToIdx", l2TxsRaw[i].ToIdx, "ToEthAddr", l2TxsRaw[i].ToEthAddr, "ToBJJ", l2TxsRaw[i].ToBJJ)
						continue
					}
					validTxs = append(validTxs, l2TxsRaw[i])
				} else {
					// case: ToBJJ==0:
					// if idx exist for EthAddr use it
					_, err := txsel.localAccountsDB.GetIdxByEthAddr(l2TxsRaw[i].ToEthAddr)
					if err == nil {
						// account for ToEthAddr already exist,
						// there is no need to create a new one.
						// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
						validTxs = append(validTxs, l2TxsRaw[i])
						continue
					}
					// if not, check if AccountCreationAuth exist for that ToEthAddr
					accAuth, err = txsel.l2db.GetAccountCreationAuth(l2TxsRaw[i].ToEthAddr)
					if err != nil {
						// not found, l2Tx will not be added in the selection
						log.Debugw("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB", "ToIdx", l2TxsRaw[i].ToIdx, "ToEthAddr", l2TxsRaw[i].ToEthAddr)
						continue
					}
					validTxs = append(validTxs, l2TxsRaw[i])
				}
				// create L1CoordinatorTx for the accountCreation
				l1CoordinatorTx := &common.L1Tx{
					UserOrigin:  false,
					FromEthAddr: accAuth.EthAddr,
					FromBJJ:     accAuth.BJJ,
					TokenID:     l2TxsRaw[i].TokenID,
					LoadAmount:  big.NewInt(0),
					Type:        common.TxTypeCreateAccountDeposit,
				}
				l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTx)
			} else if bytes.Equal(l2TxsRaw[i].ToEthAddr.Bytes(), common.FFAddr.Bytes()) && l2TxsRaw[i].ToBJJ != nil {
				// if idx exist for EthAddr&BJJ use it
				_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2TxsRaw[i].ToEthAddr, l2TxsRaw[i].ToBJJ)
				if err == nil {
					// account for ToEthAddr&ToBJJ already exist, (where ToEthAddr==0xff)
					// there is no need to create a new one.
					// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
					validTxs = append(validTxs, l2TxsRaw[i])
					continue
				}
				// if idx don't exist for EthAddr&BJJ,
				// coordinator can create a new account without
				// L1Authorization, as ToEthAddr==0xff
				// create L1CoordinatorTx for the accountCreation
				l1CoordinatorTx := &common.L1Tx{
					UserOrigin:  false,
					FromEthAddr: l2TxsRaw[i].ToEthAddr,
					FromBJJ:     l2TxsRaw[i].ToBJJ,
					TokenID:     l2TxsRaw[i].TokenID,
					LoadAmount:  big.NewInt(0),
					Type:        common.TxTypeCreateAccountDeposit,
				}
				l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTx)
			}
		} else if l2TxsRaw[i].ToIdx == common.Idx(1) {
			// valid txs (of Exit type)
			validTxs = append(validTxs, l2TxsRaw[i])
		}
	}

	// get most profitable L2-tx
	maxL2Txs := txsel.MaxTxs - uint64(len(l1CoordinatorTxs)) // - len(l1UserTxs)
	l2Txs := txsel.getL2Profitable(validTxs, maxL2Txs)

	// process the txs in the local AccountsDB
	_, _, err = txsel.localAccountsDB.ProcessTxs(false, false, l1Txs, l1CoordinatorTxs, l2Txs)
	if err != nil {
		return nil, nil, nil, err
	}
	err = txsel.localAccountsDB.MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, err
	}

	return l1Txs, l1CoordinatorTxs, l2Txs, nil
}

func checkAlreadyPendingToCreate(l1CoordinatorTxs []*common.L1Tx, addr ethCommon.Address, bjj *babyjub.PublicKey) bool {
	for i := 0; i < len(l1CoordinatorTxs); i++ {
		if bytes.Equal(l1CoordinatorTxs[i].FromEthAddr.Bytes(), addr.Bytes()) {
			if bjj == nil {
				return true
			}
			if l1CoordinatorTxs[i].FromBJJ == bjj {
				return true
			}
		}
	}
	return false
}

func (txsel *TxSelector) getL2Profitable(txs txs, max uint64) txs {
	sort.Sort(txs)
	if len(txs) < int(max) {
		return txs
	}
	return txs[:max]
}
