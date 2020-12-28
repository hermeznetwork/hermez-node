package txselector

// current: very simple version of TxSelector

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree/db"
	"github.com/iden3/go-merkletree/db/pebble"
)

const (
	// PathCoordIdxsDB defines the path of the key-value db where the
	// CoordIdxs will be stored
	PathCoordIdxsDB = "/coordidxs"
)

// txs implements the interface Sort for an array of Tx
type txs []common.PoolL2Tx

func (t txs) Len() int {
	return len(t)
}
func (t txs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t txs) Less(i, j int) bool {
	return t[i].AbsoluteFee > t[j].AbsoluteFee
}

// CoordAccount contains the data of the Coordinator account, that will be used
// to create new transactions of CreateAccountDeposit type to add new TokenID
// accounts for the Coordinator to receive the fees.
type CoordAccount struct {
	Addr                ethCommon.Address
	BJJ                 babyjub.PublicKeyComp
	AccountCreationAuth []byte
}

// SelectionConfig contains the parameters of configuration of the selection of
// transactions for the next batch
type SelectionConfig struct {
	// MaxL1UserTxs is the maximum L1-user-tx for a batch
	MaxL1UserTxs uint64
	// MaxL1CoordinatorTxs is the maximum L1-coordinator-tx for a batch
	MaxL1CoordinatorTxs uint64

	// TxProcessorConfig contains the config for ProcessTxs
	TxProcessorConfig txprocessor.Config
}

// TxSelector implements all the functionalities to select the txs for the next
// batch
type TxSelector struct {
	l2db            *l2db.L2DB
	localAccountsDB *statedb.LocalStateDB

	coordAccount *CoordAccount
	coordIdxsDB  *pebble.Storage
}

// NewTxSelector returns a *TxSelector
func NewTxSelector(coordAccount *CoordAccount, dbpath string,
	synchronizerStateDB *statedb.StateDB, l2 *l2db.L2DB) (*TxSelector, error) {
	localAccountsDB, err := statedb.NewLocalStateDB(dbpath, 128,
		synchronizerStateDB, statedb.TypeTxSelector, 0) // without merkletree
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	coordIdxsDB, err := pebble.NewPebbleStorage(dbpath+PathCoordIdxsDB, false)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return &TxSelector{
		l2db:            l2,
		localAccountsDB: localAccountsDB,
		coordAccount:    coordAccount,
		coordIdxsDB:     coordIdxsDB,
	}, nil
}

// LocalAccountsDB returns the LocalStateDB of the TxSelector
func (txsel *TxSelector) LocalAccountsDB() *statedb.LocalStateDB {
	return txsel.localAccountsDB
}

// Reset tells the TxSelector to get it's internal AccountsDB
// from the required `batchNum`
func (txsel *TxSelector) Reset(batchNum common.BatchNum) error {
	err := txsel.localAccountsDB.Reset(batchNum, true)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// AddCoordIdxs stores the given TokenID with the correspondent Idx to the
// CoordIdxsDB
func (txsel *TxSelector) AddCoordIdxs(idxs map[common.TokenID]common.Idx) error {
	tx, err := txsel.coordIdxsDB.NewTx()
	if err != nil {
		return tracerr.Wrap(err)
	}
	for tokenID, idx := range idxs {
		idxBytes, err := idx.Bytes()
		if err != nil {
			return tracerr.Wrap(err)
		}
		err = tx.Put(tokenID.Bytes(), idxBytes[:])
		if err != nil {
			return tracerr.Wrap(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// GetCoordIdxs returns a map with the stored TokenID with the correspondent
// Coordinator Idx
func (txsel *TxSelector) GetCoordIdxs() (map[common.TokenID]common.Idx, error) {
	r := make(map[common.TokenID]common.Idx)
	err := txsel.coordIdxsDB.Iterate(func(tokenIDBytes []byte, idxBytes []byte) (bool, error) {
		idx, err := common.IdxFromBytes(idxBytes)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		tokenID, err := common.TokenIDFromBytes(tokenIDBytes)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		r[tokenID] = idx
		return true, nil
	})

	return r, tracerr.Wrap(err)
}

//nolint:unused
func (txsel *TxSelector) coordAccountForTokenID(l1CoordinatorTxs []common.L1Tx, tokenID common.TokenID, positionL1 int) (*common.L1Tx, int, error) {
	// check if CoordinatorAccount for TokenID is already pending to create
	if checkAlreadyPendingToCreate(l1CoordinatorTxs, tokenID, txsel.coordAccount.Addr, txsel.coordAccount.BJJ) {
		return nil, positionL1, nil
	}

	_, err := txsel.coordIdxsDB.Get(tokenID.Bytes())
	if tracerr.Unwrap(err) == db.ErrNotFound {
		// create L1CoordinatorTx to create new CoordAccount for TokenID
		l1CoordinatorTx := common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   txsel.coordAccount.Addr,
			FromBJJ:       txsel.coordAccount.BJJ,
			TokenID:       tokenID,
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}
		positionL1++

		return &l1CoordinatorTx, positionL1, nil
	}
	if err != nil {
		return nil, positionL1, tracerr.Wrap(err)
	}
	// CoordAccount for TokenID already exists
	return nil, positionL1, nil
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
func (txsel *TxSelector) GetL2TxSelection(selectionConfig *SelectionConfig,
	batchNum common.BatchNum) ([]common.Idx, [][]byte, []common.L1Tx, []common.PoolL2Tx, error) {
	coordIdxs, accCreationAuths, _, l1CoordinatorTxs, l2Txs, err := txsel.GetL1L2TxSelection(selectionConfig, batchNum,
		[]common.L1Tx{})
	return coordIdxs, accCreationAuths, l1CoordinatorTxs, l2Txs, tracerr.Wrap(err)
}

// GetL1L2TxSelection returns the selection of L1 + L2 txs.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (txsel *TxSelector) GetL1L2TxSelection(selectionConfig *SelectionConfig,
	batchNum common.BatchNum, l1Txs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx, []common.L1Tx,
	[]common.PoolL2Tx, error) {
	// TODO WIP this method uses a 'cherry-pick' of internal calls of the
	// StateDB, a refactor of the StateDB to reorganize it internally is
	// planned once the main functionallities are covered, with that
	// refactor the TxSelector will be updated also

	// apply l1-user-tx to localAccountDB
	//     create new leaves
	//     update balances
	//     update nonces

	// get existing CoordIdxs
	coordIdxsMap, err := txsel.GetCoordIdxs()
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}
	var coordIdxs []common.Idx
	for tokenID := range coordIdxsMap {
		coordIdxs = append(coordIdxs, coordIdxsMap[tokenID])
	}

	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs() // (batchID)
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	txselStateDB := txsel.localAccountsDB.StateDB
	tp := txprocessor.NewTxProcessor(txselStateDB, selectionConfig.TxProcessorConfig)

	var validTxs txs
	var l1CoordinatorTxs []common.L1Tx
	positionL1 := len(l1Txs)

	// Process L1UserTxs
	for i := 0; i < len(l1Txs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1Txs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	// get last idx from LocalStateDB
	// lastIdx := txsel.localStateDB.idx
	// update lastIdx with the L1UserTxs (of account creation)

	var accAuths [][]byte
	for i := 0; i < len(l2TxsRaw); i++ {
		// If tx.ToIdx>=256, tx.ToIdx should exist to localAccountsDB,
		// if so, tx is used.  If tx.ToIdx==0, for an L2Tx will be the
		// case of TxToEthAddr or TxToBJJ, check if
		// tx.ToEthAddr/tx.ToBJJ exist in localAccountsDB, if yes tx is
		// used; if not, check if tx.ToEthAddr is in
		// AccountCreationAuthDB, if so, tx is used and L1CoordinatorTx
		// of CreateAccountAndDeposit is created. If tx.ToIdx==1, is a
		// Exit type and is used.
		if l2TxsRaw[i].ToIdx == 0 { // ToEthAddr/ToBJJ case
			var accAuth *common.AccountCreationAuth
			validTxs, l1CoordinatorTxs, accAuth, positionL1, err =
				txsel.processTxToEthAddrBJJ(validTxs, l1CoordinatorTxs,
					positionL1, l2TxsRaw[i])
			if err != nil {
				log.Debug(err)
				continue
			}
			if accAuth != nil {
				accAuths = append(accAuths, accAuth.Signature)
			}
		} else if l2TxsRaw[i].ToIdx >= common.IdxUserThreshold {
			_, err = txsel.localAccountsDB.GetAccount(l2TxsRaw[i].ToIdx)
			if err != nil {
				// tx not valid
				log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
					"ToIdx", l2TxsRaw[i].ToIdx)
				continue
			}

			// TODO if EthAddr!=0 or BJJ!=0, check that ToIdxAccount.EthAddr or BJJ

			// Account found in the DB, include the l2Tx in the selection
			validTxs = append(validTxs, l2TxsRaw[i])
		} else if l2TxsRaw[i].ToIdx == common.Idx(1) {
			// valid txs (of Exit type)
			validTxs = append(validTxs, l2TxsRaw[i])
		}
		// TODO if needed add L1CoordinatorTx to create a Coordinator
		// account for the new TokenID
		// var newL1CoordTx *common.L1Tx
		// newL1CoordTx, positionL1, err = txsel.coordAccountForTokenID(l1CoordinatorTxs, l2TxsRaw[i].TokenID, positionL1)
		// if err != nil {
		//         return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		// }
		// if newL1CoordTx != nil {
		//         l1CoordinatorTxs = append(l1CoordinatorTxs, *newL1CoordTx)
		// }
	}

	// Process L1CoordinatorTxs
	for i := 0; i < len(l1CoordinatorTxs); i++ {
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1CoordinatorTxs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}
	tp.AccumulatedFees = make(map[common.Idx]*big.Int)
	for _, idx := range coordIdxs {
		tp.AccumulatedFees[idx] = big.NewInt(0)
	}

	// once L1UserTxs & L1CoordinatorTxs are processed, get TokenIDs of
	// coordIdxs. In this way, if a coordIdx uses an Idx that is being
	// created in the current batch, at this point the Idx will be created
	coordIdxsMap, err = txsel.localAccountsDB.GetTokenIDsFromIdxs(coordIdxs)
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	// get most profitable L2-tx
	maxL2Txs := selectionConfig.TxProcessorConfig.MaxTx - uint32(len(l1CoordinatorTxs)) // - len(l1UserTxs) // TODO if there are L1UserTxs take them in to account
	l2Txs := txsel.getL2Profitable(validTxs, maxL2Txs)

	// Process L2Txs
	for i := 0; i < len(l2Txs); i++ {
		_, _, _, err = tp.ProcessL2Tx(coordIdxsMap, nil, nil, &l2Txs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}
	err = txsel.AddCoordIdxs(coordIdxsMap)
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	err = txsel.localAccountsDB.MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	return nil, accAuths, l1Txs, l1CoordinatorTxs, l2Txs, nil
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created and added to the
// l1CoordinatorTxs array, and then the PoolL2Tx is added into the validTxs
// array.
func (txsel *TxSelector) processTxToEthAddrBJJ(validTxs txs, l1CoordinatorTxs []common.L1Tx,
	positionL1 int, l2Tx common.PoolL2Tx) (txs, []common.L1Tx, *common.AccountCreationAuth, int, error) {
	// if L2Tx needs a new L1CoordinatorTx of CreateAccount type, and a
	// previous L2Tx in the current process already created a
	// L1CoordinatorTx of this type, in the DB there still seem that needs
	// to create a new L1CoordinatorTx, but as is already created, the tx
	// is valid
	if checkAlreadyPendingToCreate(l1CoordinatorTxs, l2Tx.TokenID, l2Tx.ToEthAddr, l2Tx.ToBJJ) {
		validTxs = append(validTxs, l2Tx)
		return validTxs, l1CoordinatorTxs, nil, positionL1, nil
	}

	var accAuth *common.AccountCreationAuth
	if !bytes.Equal(l2Tx.ToEthAddr.Bytes(), common.EmptyAddr.Bytes()) &&
		!bytes.Equal(l2Tx.ToEthAddr.Bytes(), common.FFAddr.Bytes()) {
		// case: ToEthAddr != 0x00 neither 0xff
		if l2Tx.ToBJJ != common.EmptyBJJComp {
			// case: ToBJJ!=0:
			// if idx exist for EthAddr&BJJ use it
			_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2Tx.ToEthAddr,
				l2Tx.ToBJJ, l2Tx.TokenID)
			if err == nil {
				// account for ToEthAddr&ToBJJ already exist,
				// there is no need to create a new one.
				// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
				validTxs = append(validTxs, l2Tx)
				return validTxs, l1CoordinatorTxs, nil, positionL1, nil
			}
			// if not, check if AccountCreationAuth exist for that
			// ToEthAddr
			accAuth, err = txsel.l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				// not found, l2Tx will not be added in the selection
				return validTxs, l1CoordinatorTxs, nil, positionL1, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
					l2Tx.ToIdx, l2Tx.ToEthAddr.Hex()))
			}
			if accAuth.BJJ != l2Tx.ToBJJ {
				// if AccountCreationAuth.BJJ is not the same
				// than in the tx, tx is not accepted
				return validTxs, l1CoordinatorTxs, nil, positionL1, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr & ToBJJ found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s, ToBJJ: %s",
					l2Tx.ToIdx, l2Tx.ToEthAddr.Hex(), l2Tx.ToBJJ.String()))
			}
			validTxs = append(validTxs, l2Tx)
		} else {
			// case: ToBJJ==0:
			// if idx exist for EthAddr use it
			_, err := txsel.localAccountsDB.GetIdxByEthAddr(l2Tx.ToEthAddr, l2Tx.TokenID)
			if err == nil {
				// account for ToEthAddr already exist,
				// there is no need to create a new one.
				// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
				validTxs = append(validTxs, l2Tx)
				return validTxs, l1CoordinatorTxs, nil, positionL1, nil
			}
			// if not, check if AccountCreationAuth exist for that ToEthAddr
			accAuth, err = txsel.l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				// not found, l2Tx will not be added in the selection
				return validTxs, l1CoordinatorTxs, nil, positionL1, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
					l2Tx.ToIdx, l2Tx.ToEthAddr))
			}
			validTxs = append(validTxs, l2Tx)
		}
		// create L1CoordinatorTx for the accountCreation
		l1CoordinatorTx := common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   accAuth.EthAddr,
			FromBJJ:       accAuth.BJJ,
			TokenID:       l2Tx.TokenID,
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}
		positionL1++
		l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTx)
	} else if bytes.Equal(l2Tx.ToEthAddr.Bytes(), common.FFAddr.Bytes()) && l2Tx.ToBJJ != common.EmptyBJJComp {
		// if idx exist for EthAddr&BJJ use it
		_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2Tx.ToEthAddr, l2Tx.ToBJJ,
			l2Tx.TokenID)
		if err == nil {
			// account for ToEthAddr&ToBJJ already exist, (where ToEthAddr==0xff)
			// there is no need to create a new one.
			// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
			validTxs = append(validTxs, l2Tx)
			return validTxs, l1CoordinatorTxs, nil, positionL1, nil
		}
		// if idx don't exist for EthAddr&BJJ,
		// coordinator can create a new account without
		// L1Authorization, as ToEthAddr==0xff
		// create L1CoordinatorTx for the accountCreation
		l1CoordinatorTx := common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   l2Tx.ToEthAddr,
			FromBJJ:       l2Tx.ToBJJ,
			TokenID:       l2Tx.TokenID,
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}
		positionL1++
		l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTx)
	}

	return validTxs, l1CoordinatorTxs, accAuth, positionL1, nil
}

func checkAlreadyPendingToCreate(l1CoordinatorTxs []common.L1Tx, tokenID common.TokenID,
	addr ethCommon.Address, bjj babyjub.PublicKeyComp) bool {
	for i := 0; i < len(l1CoordinatorTxs); i++ {
		if bytes.Equal(l1CoordinatorTxs[i].FromEthAddr.Bytes(), addr.Bytes()) &&
			l1CoordinatorTxs[i].TokenID == tokenID &&
			l1CoordinatorTxs[i].FromBJJ == bjj {
			return true
		}
	}
	return false
}

// getL2Profitable returns the profitable selection of L2Txssorted by Nonce
func (txsel *TxSelector) getL2Profitable(txs txs, max uint32) txs {
	sort.Sort(txs)
	if len(txs) < int(max) {
		return txs
	}
	txs = txs[:max]

	// sort l2Txs by Nonce. This can be done in many different ways, what
	// is needed is to output the txs where the Nonce of txs for each
	// Account is sorted, but the txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.SliceStable(txs, func(i, j int) bool {
		return txs[i].Nonce < txs[j].Nonce
	})

	return txs
}
