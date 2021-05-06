package txselector

// current: very simple version of TxSelector

import (
	"fmt"
	"math/big"
	"sort"

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

// CoordAccount contains the data of the Coordinator account, that will be used
// to create new transactions of CreateAccountDeposit type to add new TokenID
// accounts for the Coordinator to receive the fees.
type CoordAccount struct {
	Addr                ethCommon.Address
	BJJ                 babyjub.PublicKeyComp
	AccountCreationAuth []byte // signature in byte array format
}

// TxSelector implements all the functionalities to select the txs for the next
// batch
type TxSelector struct {
	l2db            *l2db.L2DB
	localAccountsDB *statedb.LocalStateDB

	coordAccount *CoordAccount
}

// NewTxSelector returns a *TxSelector
func NewTxSelector(coordAccount *CoordAccount, dbpath string,
	synchronizerStateDB *statedb.StateDB, l2 *l2db.L2DB) (*TxSelector, error) {
	localAccountsDB, err := statedb.NewLocalStateDB(
		statedb.Config{
			Path:    dbpath,
			Keep:    kvdb.DefaultKeep,
			Type:    statedb.TypeTxSelector,
			NLevels: 0,
		},
		synchronizerStateDB) // without merkletree
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
func (txsel *TxSelector) LocalAccountsDB() *statedb.LocalStateDB {
	return txsel.localAccountsDB
}

// Reset tells the TxSelector to get it's internal AccountsDB
// from the required `batchNum`
func (txsel *TxSelector) Reset(batchNum common.BatchNum, fromSynchronizer bool) error {
	return tracerr.Wrap(txsel.localAccountsDB.Reset(batchNum, fromSynchronizer))
}

func (txsel *TxSelector) getCoordIdx(tokenID common.TokenID) (common.Idx, error) {
	return txsel.localAccountsDB.GetIdxByEthAddrBJJ(txsel.coordAccount.Addr,
		txsel.coordAccount.BJJ, tokenID)
}

// coordAccountForTokenID creates a new L1CoordinatorTx to create a new
// Coordinator account for the given TokenID in the case that the account does
// not exist yet in the db, and does not exist a L1CoordinatorTx to creat that
// account in the given array of L1CoordinatorTxs. If a new Coordinator account
// needs to be created, a new L1CoordinatorTx will be returned from this
// function. After calling this method, if the l1CoordinatorTx is added to the
// selection, positionL1 must be increased 1.
func (txsel *TxSelector) coordAccountForTokenID(l1CoordinatorTxs []common.L1Tx,
	tokenID common.TokenID, positionL1 int) (*common.L1Tx, int, error) {
	// check if CoordinatorAccount for TokenID is already pending to create
	if checkPendingToCreateL1CoordTx(l1CoordinatorTxs, tokenID,
		txsel.coordAccount.Addr, txsel.coordAccount.BJJ) {
		return nil, positionL1, nil
	}
	_, err := txsel.getCoordIdx(tokenID)
	if tracerr.Unwrap(err) == statedb.ErrIdxNotFound {
		// create L1CoordinatorTx to create new CoordAccount for
		// TokenID
		l1CoordinatorTx := common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   txsel.coordAccount.Addr,
			FromBJJ:       txsel.coordAccount.BJJ,
			TokenID:       tokenID,
			Amount:        big.NewInt(0),
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}

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
func (txsel *TxSelector) GetL2TxSelection(selectionConfig txprocessor.Config, l1UserFutureTxs []common.L1Tx) ([]common.Idx,
	[][]byte, []common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metric.GetL2TxSelection.Inc()
	coordIdxs, accCreationAuths, _, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, err := txsel.getL1L2TxSelection(selectionConfig,
		[]common.L1Tx{}, l1UserFutureTxs)
	return coordIdxs, accCreationAuths, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, tracerr.Wrap(err)
}

// GetL1L2TxSelection returns the selection of L1 + L2 txs.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (txsel *TxSelector) GetL1L2TxSelection(selectionConfig txprocessor.Config,
	l1UserTxs, l1UserFutureTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metric.GetL1L2TxSelection.Inc()
	coordIdxs, accCreationAuths, l1UserTxs, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, err := txsel.getL1L2TxSelection(selectionConfig, l1UserTxs, l1UserFutureTxs)
	return coordIdxs, accCreationAuths, l1UserTxs, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, tracerr.Wrap(err)
}

// getL1L2TxSelection returns the selection of L1 + L2 txs.
// It returns: the CoordinatorIdxs used to receive the fees of the selected
// L2Txs. An array of bytearrays with the signatures of the
// AccountCreationAuthorization of the accounts of the users created by the
// Coordinator with L1CoordinatorTxs of those accounts that does not exist yet
// but there is a transactions to them and the authorization of account
// creation exists. The L1UserTxs, L1CoordinatorTxs, PoolL2Txs that will be
// included in the next batch.
func (txsel *TxSelector) getL1L2TxSelection(selectionConfig txprocessor.Config,
	l1UserTxs, l1UserFutureTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	// WIP.0: the TxSelector is not optimized and will need a redesign. The
	// current version is implemented in order to have a functional
	// implementation that can be used ASAP.

	// Steps of this method:
	// - ProcessL1Txs (User txs)
	// - getPendingTxs (forgable directly with current state & not forgable
	// yet)
	// - split between l2TxsForgable & l2TxsNonForgable, where:
	// 	- l2TxsForgable are the txs that are directly forgable with the
	// 	current state
	// 	- l2TxsNonForgable are the txs that are not directly forgable
	// 	with the current state, but that may be forgable once the
	// 	l2TxsForgable ones are processed
	// - for l2TxsForgable, and if needed, for l2TxsNonForgable:
	// 	- sort by Fee & Nonce
	// 	- loop over l2Txs (txsel.processL2Txs)
	// 	        - Fill tx.TokenID tx.Nonce
	// 	        - Check enough Balance on sender
	// 	        - Check Nonce
	// 	        - Create CoordAccount L1CoordTx for TokenID if needed
	// 	                - & ProcessL1Tx of L1CoordTx
	// 	        - Check validity of receiver Account for ToEthAddr / ToBJJ
	// 	        - Create UserAccount L1CoordTx if needed (and possible)
	// 	        - If everything is fine, store l2Tx to validTxs & update NoncesMap
	// - Prepare coordIdxsMap & AccumulatedFees
	// - Distribute AccumulatedFees to CoordIdxs
	// - MakeCheckpoint

	txselStateDB := txsel.localAccountsDB.StateDB
	tp := txprocessor.NewTxProcessor(txselStateDB, selectionConfig)
	tp.AccumulatedFees = make(map[common.Idx]*big.Int)

	// Process L1UserTxs
	for i := 0; i < len(l1UserTxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1UserTxs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	l2TxsFromDB, err := txsel.l2db.GetPendingTxs()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}
	l2TxsForgable, l2TxsNonForgable := splitL2ForgableAndNonForgable(tp, l2TxsFromDB)

	// in case that length of l2TxsForgable is 0, no need to continue, there
	// is no L2Txs to forge at all
	if len(l2TxsForgable) == 0 {
		var discardedL2Txs []common.PoolL2Tx
		for i := 0; i < len(l2TxsNonForgable); i++ {
			l2TxsNonForgable[i].Info =
				"Tx not selected due impossibility to be forged with the current state"
			discardedL2Txs = append(discardedL2Txs, l2TxsNonForgable[i])
		}
		err = tp.StateDB().MakeCheckpoint()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}

		metric.SelectedL1UserTxs.Set(float64(len(l1UserTxs)))
		metric.SelectedL1CoordinatorTxs.Set(0)
		metric.SelectedL2Txs.Set(0)
		metric.DiscardedL2Txs.Set(float64(len(discardedL2Txs)))

		return nil, nil, l1UserTxs, nil, nil, discardedL2Txs, nil
	}

	var accAuths [][]byte
	var l1CoordinatorTxs []common.L1Tx
	var validTxs, discardedL2Txs []common.PoolL2Tx
	l2TxsForgable = sortL2Txs(l2TxsForgable)
	accAuths, l1CoordinatorTxs, validTxs, discardedL2Txs, err =
		txsel.processL2Txs(tp, selectionConfig, len(l1UserTxs), l1UserFutureTxs,
			l2TxsForgable, validTxs, discardedL2Txs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	// if there is space for more txs get also the NonForgable txs, that may
	// be unblocked once the Forgable ones are processed
	if len(validTxs) < int(selectionConfig.MaxTx)-(len(l1UserTxs)+len(l1CoordinatorTxs)) {
		l2TxsNonForgable = sortL2Txs(l2TxsNonForgable)
		var accAuths2 [][]byte
		var l1CoordinatorTxs2 []common.L1Tx
		accAuths2, l1CoordinatorTxs2, validTxs, discardedL2Txs, err =
			txsel.processL2Txs(tp, selectionConfig,
				len(l1UserTxs)+len(l1CoordinatorTxs), l1UserFutureTxs,
				l2TxsNonForgable, validTxs, discardedL2Txs)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}

		accAuths = append(accAuths, accAuths2...)
		l1CoordinatorTxs = append(l1CoordinatorTxs, l1CoordinatorTxs2...)
	} else {
		// if there is no space for NonForgable txs, put them at the
		// discardedL2Txs array
		for i := 0; i < len(l2TxsNonForgable); i++ {
			l2TxsNonForgable[i].Info =
				"Tx not selected due not available slots for L2Txs"
			discardedL2Txs = append(discardedL2Txs, l2TxsNonForgable[i])
		}
	}

	// get CoordIdxsMap for the TokenIDs
	coordIdxsMap := make(map[common.TokenID]common.Idx)
	for i := 0; i < len(validTxs); i++ {
		// get TokenID from tx.Sender
		accSender, err := tp.StateDB().GetAccount(validTxs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		tokenID := accSender.TokenID

		coordIdx, err := txsel.getCoordIdx(tokenID)
		if err != nil {
			// if err is db.ErrNotFound, should not happen, as all
			// the validTxs.TokenID should have a CoordinatorIdx
			// created in the DB at this point
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		coordIdxsMap[tokenID] = coordIdx
	}

	var coordIdxs []common.Idx
	for _, idx := range coordIdxsMap {
		coordIdxs = append(coordIdxs, idx)
	}
	// sort CoordIdxs
	sort.SliceStable(coordIdxs, func(i, j int) bool {
		return coordIdxs[i] < coordIdxs[j]
	})

	// distribute the AccumulatedFees from the processed L2Txs into the
	// Coordinator Idxs
	for idx, accumulatedFee := range tp.AccumulatedFees {
		cmp := accumulatedFee.Cmp(big.NewInt(0))
		if cmp == 1 { // accumulatedFee>0
			// send the fee to the Idx of the Coordinator for the TokenID
			accCoord, err := txsel.localAccountsDB.GetAccount(idx)
			if err != nil {
				log.Errorw("Can not distribute accumulated fees to coordinator "+
					"account: No coord Idx to receive fee", "idx", idx)
				return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
			accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
			_, err = txsel.localAccountsDB.UpdateAccount(idx, accCoord)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
		}
	}

	err = tp.StateDB().MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	metric.SelectedL1CoordinatorTxs.Set(float64(len(l1CoordinatorTxs)))
	metric.SelectedL1UserTxs.Set(float64(len(l1UserTxs)))
	metric.SelectedL2Txs.Set(float64(len(validTxs)))
	metric.DiscardedL2Txs.Set(float64(len(discardedL2Txs)))

	return coordIdxs, accAuths, l1UserTxs, l1CoordinatorTxs, validTxs, discardedL2Txs, nil
}

func (txsel *TxSelector) processL2Txs(tp *txprocessor.TxProcessor,
	selectionConfig txprocessor.Config, nL1Txs int, l1UserFutureTxs []common.L1Tx,
	l2Txs, validTxs, discardedL2Txs []common.PoolL2Tx) ([][]byte, []common.L1Tx,
	[]common.PoolL2Tx, []common.PoolL2Tx, error) {
	var l1CoordinatorTxs []common.L1Tx
	positionL1 := nL1Txs
	var accAuths [][]byte
	// Iterate over l2Txs
	// - check Nonces
	// - check enough Balance for the Amount+Fee
	// - if needed, create new L1CoordinatorTxs for unexisting ToIdx
	// 	- keep used accAuths
	// - put the valid txs into validTxs array
	for i := 0; i < len(l2Txs); i++ {
		// Check if there is space for more L2Txs in the selection
		maxL2Txs := int(selectionConfig.MaxTx) - nL1Txs - len(l1CoordinatorTxs)
		if len(validTxs) >= maxL2Txs {
			// no more available slots for L2Txs, so mark this tx
			// but also the rest of remaining txs as discarded
			for j := i; j < len(l2Txs); j++ {
				l2Txs[j].Info =
					"Tx not selected due not available slots for L2Txs"
				discardedL2Txs = append(discardedL2Txs, l2Txs[j])
			}
			break
		}

		// get Nonce & TokenID from the Account by l2Tx.FromIdx
		accSender, err := tp.StateDB().GetAccount(l2Txs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, tracerr.Wrap(err)
		}
		l2Txs[i].TokenID = accSender.TokenID

		// Check enough Balance on sender
		enoughBalance, balance, feeAndAmount := tp.CheckEnoughBalance(l2Txs[i])
		if !enoughBalance {
			// not valid Amount with current Balance. Discard L2Tx,
			// and update Info parameter of the tx, and add it to
			// the discardedTxs array
			l2Txs[i].Info = fmt.Sprintf("Tx not selected due to not enough Balance at the sender. "+
				"Current sender account Balance: %s, Amount+Fee: %s",
				balance.String(), feeAndAmount.String())
			discardedL2Txs = append(discardedL2Txs, l2Txs[i])
			continue
		}

		// Check if Nonce is correct
		if l2Txs[i].Nonce != accSender.Nonce {
			// not valid Nonce at tx. Discard L2Tx, and update Info
			// parameter of the tx, and add it to the discardedTxs
			// array
			l2Txs[i].Info = fmt.Sprintf("Tx not selected due to not current Nonce. "+
				"Tx.Nonce: %d, Account.Nonce: %d", l2Txs[i].Nonce, accSender.Nonce)
			discardedL2Txs = append(discardedL2Txs, l2Txs[i])
			continue
		}

		// if TokenID does not exist yet, create new L1CoordinatorTx to
		// create the CoordinatorAccount for that TokenID, to receive
		// the fees. Only in the case that there does not exist yet a
		// pending L1CoordinatorTx to create the account for the
		// Coordinator for that TokenID
		var newL1CoordTx *common.L1Tx
		newL1CoordTx, positionL1, err =
			txsel.coordAccountForTokenID(l1CoordinatorTxs,
				accSender.TokenID, positionL1)
		if err != nil {
			return nil, nil, nil, nil, tracerr.Wrap(err)
		}
		if newL1CoordTx != nil {
			// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
			// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
			if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1Tx)-nL1Txs ||
				len(l1CoordinatorTxs)+1 >= int(selectionConfig.MaxTx)-nL1Txs {
				// discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = "Tx not selected because the L2Tx depends on a " +
					"L1CoordinatorTx and there is not enough space for L1Coordinator"
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}
			// increase positionL1
			positionL1++
			l1CoordinatorTxs = append(l1CoordinatorTxs, *newL1CoordTx)
			accAuths = append(accAuths, txsel.coordAccount.AccountCreationAuth)

			// process the L1CoordTx
			_, _, _, _, err := tp.ProcessL1Tx(nil, newL1CoordTx)
			if err != nil {
				return nil, nil, nil, nil, tracerr.Wrap(err)
			}
		}

		// If tx.ToIdx>=256, tx.ToIdx should exist to localAccountsDB,
		// if so, tx is used.  If tx.ToIdx==0, for an L2Tx will be the
		// case of TxToEthAddr or TxToBJJ, check if
		// tx.ToEthAddr/tx.ToBJJ exist in localAccountsDB, if yes tx is
		// used; if not, check if tx.ToEthAddr is in
		// AccountCreationAuthDB, if so, tx is used and L1CoordinatorTx
		// of CreateAccountAndDeposit is created. If tx.ToIdx==1, is a
		// Exit type and is used.
		if l2Txs[i].ToIdx == 0 { // ToEthAddr/ToBJJ case
			validL2Tx, l1CoordinatorTx, accAuth, err :=
				txsel.processTxToEthAddrBJJ(validTxs, selectionConfig,
					nL1Txs, l1UserFutureTxs, l1CoordinatorTxs,
					positionL1, l2Txs[i])
			if err != nil {
				log.Debugw("txsel.processTxToEthAddrBJJ", "err", err)
				// Discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = fmt.Sprintf("Tx not selected (in processTxToEthAddrBJJ) due to %s",
					err.Error())
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}
			// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
			// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
			if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1Tx)-nL1Txs ||
				len(l1CoordinatorTxs)+1 >= int(selectionConfig.MaxTx)-nL1Txs {
				// discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = "Tx not selected because the L2Tx depends on a " +
					"L1CoordinatorTx and there is not enough space for L1Coordinator"
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}

			if l1CoordinatorTx != nil && validL2Tx != nil {
				// If ToEthAddr == 0xff.. this means that we
				// are handling a TransferToBJJ, which doesn't
				// require an authorization because it doesn't
				// contain a valid ethereum address.
				// Otherwise only create the account if we have
				// the corresponding authorization
				if validL2Tx.ToEthAddr == common.FFAddr {
					accAuths = append(accAuths, common.EmptyEthSignature)
					l1CoordinatorTxs = append(l1CoordinatorTxs, *l1CoordinatorTx)
					positionL1++
				} else if accAuth != nil {
					accAuths = append(accAuths, accAuth.Signature)
					l1CoordinatorTxs = append(l1CoordinatorTxs, *l1CoordinatorTx)
					positionL1++
				}

				// process the L1CoordTx
				_, _, _, _, err := tp.ProcessL1Tx(nil, l1CoordinatorTx)
				if err != nil {
					return nil, nil, nil, nil, tracerr.Wrap(err)
				}
			}
			if validL2Tx == nil {
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}
		} else if l2Txs[i].ToIdx >= common.IdxUserThreshold {
			_, err := txsel.localAccountsDB.GetAccount(l2Txs[i].ToIdx)
			if err != nil {
				// tx not valid
				log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
					"ToIdx", l2Txs[i].ToIdx)
				// Discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = fmt.Sprintf("Tx not selected due to tx.ToIdx not found in StateDB. "+
					"ToIdx: %d", l2Txs[i].ToIdx)
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}
		}

		// get CoordIdxsMap for the TokenID of the current l2Txs[i]
		// get TokenID from tx.Sender account
		tokenID := accSender.TokenID
		coordIdx, err := txsel.getCoordIdx(tokenID)
		if err != nil {
			// if err is db.ErrNotFound, should not happen, as all
			// the validTxs.TokenID should have a CoordinatorIdx
			// created in the DB at this point
			return nil, nil, nil, nil,
				tracerr.Wrap(fmt.Errorf("Could not get CoordIdx for TokenID=%d, "+
					"due: %s", tokenID, err))
		}
		// prepare temp coordIdxsMap & AccumulatedFees for the call to
		// ProcessL2Tx
		coordIdxsMap := map[common.TokenID]common.Idx{tokenID: coordIdx}
		// tp.AccumulatedFees = make(map[common.Idx]*big.Int)
		if _, ok := tp.AccumulatedFees[coordIdx]; !ok {
			tp.AccumulatedFees[coordIdx] = big.NewInt(0)
		}

		_, _, _, err = tp.ProcessL2Tx(coordIdxsMap, nil, nil, &l2Txs[i])
		if err != nil {
			log.Debugw("txselector.getL1L2TxSelection at ProcessL2Tx", "err", err)
			// Discard L2Tx, and update Info parameter of the tx,
			// and add it to the discardedTxs array
			l2Txs[i].Info = fmt.Sprintf("Tx not selected (in ProcessL2Tx) due to %s",
				err.Error())
			discardedL2Txs = append(discardedL2Txs, l2Txs[i])
			continue
		}

		validTxs = append(validTxs, l2Txs[i])
	} // after this loop, no checks to discard txs should be done

	return accAuths, l1CoordinatorTxs, validTxs, discardedL2Txs, nil
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created and added to the
// l1CoordinatorTxs array, and then the PoolL2Tx is added into the validTxs
// array.
func (txsel *TxSelector) processTxToEthAddrBJJ(validTxs []common.PoolL2Tx,
	selectionConfig txprocessor.Config, nL1UserTxs int, l1UserFutureTxs,
	l1CoordinatorTxs []common.L1Tx, positionL1 int, l2Tx common.PoolL2Tx) (
	*common.PoolL2Tx, *common.L1Tx, *common.AccountCreationAuth, error) {
	// if L2Tx needs a new L1CoordinatorTx of CreateAccount type, and a
	// previous L2Tx in the current process already created a
	// L1CoordinatorTx of this type, in the DB there still seem that needs
	// to create a new L1CoordinatorTx, but as is already created, the tx
	// is valid
	if checkPendingToCreateL1CoordTx(l1CoordinatorTxs, l2Tx.TokenID, l2Tx.ToEthAddr, l2Tx.ToBJJ) {
		return &l2Tx, nil, nil, nil
	}

	// check if L2Tx receiver account will be created by a L1UserFutureTxs
	// (in the next batch, the current frozen queue). In that case, the L2Tx
	// will be discarded at the current batch, even if there is an
	// AccountCreationAuth for the account, as there is a L1UserTx in the
	// frozen queue that will create the receiver Account.  The L2Tx is
	// discarded to avoid the Coordinator creating a new L1CoordinatorTx to
	// create the receiver account, which will be also created in the next
	// batch from the L1UserFutureTx, ending with the user having 2
	// different accounts for the same TokenID. The double account creation
	// is supported by the Hermez zkRollup specification, but it was decided
	// to mitigate it at the TxSelector level for the explained cases.
	if checkPendingToCreateFutureTxs(l1UserFutureTxs, l2Tx.TokenID, l2Tx.ToEthAddr, l2Tx.ToBJJ) {
		return nil, nil, nil, fmt.Errorf("L2Tx discarded at the current batch, as the" +
			" receiver account does not exist yet, and there is a L1UserTx that" +
			" will create that account in a future batch.")
	}

	var l1CoordinatorTx *common.L1Tx
	var accAuth *common.AccountCreationAuth
	if l2Tx.ToEthAddr != common.EmptyAddr && l2Tx.ToEthAddr != common.FFAddr {
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
				return &l2Tx, nil, nil, nil
			}
			// if not, check if AccountCreationAuth exist for that
			// ToEthAddr
			accAuth, err = txsel.l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				// not found, l2Tx will not be added in the selection
				return nil, nil, nil,
					tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found "+
						"in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
						l2Tx.ToIdx, l2Tx.ToEthAddr.Hex()))
			}
			if accAuth.BJJ != l2Tx.ToBJJ {
				// if AccountCreationAuth.BJJ is not the same
				// than in the tx, tx is not accepted
				return nil, nil, nil,
					tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, "+
						"neither ToEthAddr & ToBJJ found in AccountCreationAuths L2DB. "+
						"ToIdx: %d, ToEthAddr: %s, ToBJJ: %s",
						l2Tx.ToIdx, l2Tx.ToEthAddr.Hex(), l2Tx.ToBJJ.String()))
			}
		} else {
			// case: ToBJJ==0:
			// if idx exist for EthAddr use it
			_, err := txsel.localAccountsDB.GetIdxByEthAddr(l2Tx.ToEthAddr, l2Tx.TokenID)
			if err == nil {
				// account for ToEthAddr already exist,
				// there is no need to create a new one.
				// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
				return &l2Tx, nil, nil, nil
			}
			// if not, check if AccountCreationAuth exist for that ToEthAddr
			accAuth, err = txsel.l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				// not found, l2Tx will not be added in the selection
				return nil, nil, nil,
					tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in "+
						"StateDB, neither ToEthAddr found in "+
						"AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
						l2Tx.ToIdx, l2Tx.ToEthAddr))
			}
		}
		// create L1CoordinatorTx for the accountCreation
		l1CoordinatorTx = &common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   accAuth.EthAddr,
			FromBJJ:       accAuth.BJJ,
			TokenID:       l2Tx.TokenID,
			Amount:        big.NewInt(0),
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}
	} else if l2Tx.ToEthAddr == common.FFAddr && l2Tx.ToBJJ != common.EmptyBJJComp {
		// if idx exist for EthAddr&BJJ use it
		_, err := txsel.localAccountsDB.GetIdxByEthAddrBJJ(l2Tx.ToEthAddr, l2Tx.ToBJJ,
			l2Tx.TokenID)
		if err == nil {
			// account for ToEthAddr&ToBJJ already exist, (where ToEthAddr==0xff)
			// there is no need to create a new one.
			// tx valid, StateDB will use the ToIdx==0 to define the AuxToIdx
			return &l2Tx, nil, nil, nil
		}
		// if idx don't exist for EthAddr&BJJ, coordinator can create a
		// new account without L1Authorization, as ToEthAddr==0xff
		// create L1CoordinatorTx for the accountCreation
		l1CoordinatorTx = &common.L1Tx{
			Position:      positionL1,
			UserOrigin:    false,
			FromEthAddr:   l2Tx.ToEthAddr,
			FromBJJ:       l2Tx.ToBJJ,
			TokenID:       l2Tx.TokenID,
			Amount:        big.NewInt(0),
			DepositAmount: big.NewInt(0),
			Type:          common.TxTypeCreateAccountDeposit,
		}
	}
	// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
	// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
	if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1Tx)-nL1UserTxs ||
		len(l1CoordinatorTxs)+1 >= int(selectionConfig.MaxTx)-nL1UserTxs {
		// L2Tx discarded
		return nil, nil, nil, tracerr.Wrap(fmt.Errorf("L2Tx discarded due to no available slots " +
			"for L1CoordinatorTx to create a new account for receiver of L2Tx"))
	}

	return &l2Tx, l1CoordinatorTx, accAuth, nil
}

func checkPendingToCreateL1CoordTx(l1CoordinatorTxs []common.L1Tx, tokenID common.TokenID,
	addr ethCommon.Address, bjj babyjub.PublicKeyComp) bool {
	for i := 0; i < len(l1CoordinatorTxs); i++ {
		if l1CoordinatorTxs[i].FromEthAddr == addr &&
			l1CoordinatorTxs[i].TokenID == tokenID &&
			l1CoordinatorTxs[i].FromBJJ == bjj {
			return true
		}
	}
	return false
}

func checkPendingToCreateFutureTxs(l1UserFutureTxs []common.L1Tx, tokenID common.TokenID,
	addr ethCommon.Address, bjj babyjub.PublicKeyComp) bool {
	for i := 0; i < len(l1UserFutureTxs); i++ {
		if l1UserFutureTxs[i].FromEthAddr == addr &&
			l1UserFutureTxs[i].TokenID == tokenID &&
			l1UserFutureTxs[i].FromBJJ == bjj {
			return true
		}
		if l1UserFutureTxs[i].FromEthAddr == addr &&
			l1UserFutureTxs[i].TokenID == tokenID &&
			common.EmptyBJJComp == bjj {
			return true
		}
	}
	return false
}

// sortL2Txs sorts the PoolL2Txs by AbsoluteFee and then by Nonce
func sortL2Txs(l2Txs []common.PoolL2Tx) []common.PoolL2Tx {
	// Sort by absolute fee with SliceStable, so that txs with same
	// AbsoluteFee are not rearranged and nonce order is kept in such case
	sort.SliceStable(l2Txs, func(i, j int) bool {
		return l2Txs[i].AbsoluteFee > l2Txs[j].AbsoluteFee
	})

	// sort l2Txs by Nonce. This can be done in many different ways, what
	// is needed is to output the l2Txs where the Nonce of l2Txs for each
	// Account is sorted, but the l2Txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.Slice(l2Txs, func(i, j int) bool {
		return l2Txs[i].Nonce < l2Txs[j].Nonce
	})

	return l2Txs
}

func splitL2ForgableAndNonForgable(tp *txprocessor.TxProcessor,
	l2Txs []common.PoolL2Tx) ([]common.PoolL2Tx, []common.PoolL2Tx) {
	var l2TxsForgable, l2TxsNonForgable []common.PoolL2Tx
	for i := 0; i < len(l2Txs); i++ {
		accSender, err := tp.StateDB().GetAccount(l2Txs[i].FromIdx)
		if err != nil {
			l2TxsNonForgable = append(l2TxsNonForgable, l2Txs[i])
			continue
		}

		if l2Txs[i].Nonce != accSender.Nonce {
			l2TxsNonForgable = append(l2TxsNonForgable, l2Txs[i])
			continue
		}
		enoughBalance, _, _ := tp.CheckEnoughBalance(l2Txs[i])
		if !enoughBalance {
			l2TxsNonForgable = append(l2TxsNonForgable, l2Txs[i])
			continue
		}
		l2TxsForgable = append(l2TxsForgable, l2Txs[i])
	}
	return l2TxsForgable, l2TxsNonForgable
}
