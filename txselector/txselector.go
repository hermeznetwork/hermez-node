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
func (txsel *TxSelector) coordAccountForTokenID(tokenID common.TokenID, positionL1 int) (*common.L1Tx, int, error) {
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

type failedAtomicGroup struct {
	id     common.AtomicGroupID
	txID   common.TxID
	reason string
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
	// 	        - If everything is fine, store l2Tx to selectedTxs & update NoncesMap
	// - Prepare coordIdxsMap & AccumulatedFees
	// - Distribute AccumulatedFees to CoordIdxs
	// - MakeCheckpoint
	failedAtomicGroups := []failedAtomicGroup{}
START_SELECTION:
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

	// Get pending txs from the pool
	l2TxsFromDB, err := txsel.l2db.GetPendingTxs()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}
	// Filter transactions belonging to failed atomic groups
	selectableTxs, discardedTxs := filterFailedAtomicGroups(l2TxsFromDB, failedAtomicGroups)

	// in case that length of l2TxsForgable is 0, no need to continue, there
	// is no L2Txs to forge at all
	if len(selectableTxs) == 0 {
		err = tp.StateDB().MakeCheckpoint()
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}

		metric.SelectedL1UserTxs.Set(float64(len(l1UserTxs)))
		metric.SelectedL1CoordinatorTxs.Set(0)
		metric.SelectedL2Txs.Set(0)
		metric.DiscardedL2Txs.Set(float64(len(discardedTxs)))

		return nil, nil, l1UserTxs, nil, nil, discardedTxs, nil
	}

	// Calculate average fee for atomic groups
	atomicFeeMap := calculateAtomicGroupsAverageFee(selectableTxs)

	// Initialize selection arrays
	var accAuths [][]byte              // Used authorizations in the l1CoordinatorTxs
	var l1CoordinatorTxs []common.L1Tx // Processed txs for necessary account creation (fees for coordinator or missing destinatary accounts)
	var selectedTxs []common.PoolL2Tx  // Processed txs
	// Start selection process
	shouldKeepSelectionProcess := true
	// Order L2 txs. This has to be done just once, as the array will get smaller over iterations, but the order won't be affected
	selectableTxs = sortL2Txs(selectableTxs, atomicFeeMap)
	for shouldKeepSelectionProcess {
		// Process txs and get selection
		iteAccAuths, iteL1CoordinatorTxs, iteSelectedTxs, nonSelectedTxs, invalidTxs, failedAtomicGroup, err := txsel.processL2Txs(
			tp,
			selectionConfig,
			len(l1UserTxs)+len(l1CoordinatorTxs), // Already added L1 Txs
			len(selectedTxs),                     // Already added L2 Txs
			l1UserFutureTxs,                      // L1Txs that will be added in the future, used to prevent the creation of unnecessary accounts
			selectableTxs,                        // Txs that can be selected
		)
		if failedAtomicGroup.id != common.EmptyAtomicGroupID {
			// An atomic group failed to be processed after at least one tx from the group already alteredthe state.
			// Revert state to current batch and start the selection process again, ignoring the txs from the group that failed
			log.Info(err)
			failedAtomicGroups = append(failedAtomicGroups, failedAtomicGroup)
			if err := txsel.localAccountsDB.Reset(txsel.localAccountsDB.CurrentBatch(), false); err != nil {
				return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
			goto START_SELECTION
		}
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		// Add iteration results to selection arrays
		accAuths = append(accAuths, iteAccAuths...)
		l1CoordinatorTxs = append(l1CoordinatorTxs, iteL1CoordinatorTxs...)
		selectedTxs = append(selectedTxs, iteSelectedTxs...)
		discardedTxs = append(discardedTxs, invalidTxs...)
		// Prepare for next iteration
		if len(iteSelectedTxs) == 0 { // Stop iterating
			// If in this iteration no txs got selected, stop selection process
			shouldKeepSelectionProcess = false
			// Add non selected txs to the discarded array as at this point they won't get selected
			for i := 0; i < len(nonSelectedTxs); i++ {
				discardedTxs = append(discardedTxs, nonSelectedTxs[i])
			}
		} else { // Keep iterating
			// Try to select nonSelected txs in next iteration
			selectableTxs = nonSelectedTxs
		}
	}

	// get CoordIdxsMap for the TokenIDs
	coordIdxsMap := make(map[common.TokenID]common.Idx)
	for i := 0; i < len(selectedTxs); i++ {
		// get TokenID from tx.Sender
		accSender, err := tp.StateDB().GetAccount(selectedTxs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		tokenID := accSender.TokenID

		coordIdx, err := txsel.getCoordIdx(tokenID)
		if err != nil {
			// if err is db.ErrNotFound, should not happen, as all
			// the selectedTxs.TokenID should have a CoordinatorIdx
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
	metric.SelectedL2Txs.Set(float64(len(selectedTxs)))
	metric.DiscardedL2Txs.Set(float64(len(discardedTxs)))

	return coordIdxs, accAuths, l1UserTxs, l1CoordinatorTxs, selectedTxs, discardedTxs, nil
}

func (txsel *TxSelector) processL2Txs(
	tp *txprocessor.TxProcessor,
	selectionConfig txprocessor.Config,
	nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs int,
	l1UserFutureTxs []common.L1Tx,
	l2Txs []common.PoolL2Tx,
) (
	accAuths [][]byte,
	l1CoordinatorTxs []common.L1Tx, // Processed txs for creating accounts (for coordinator fees or destinatary accounts)
	selectedL2Txs []common.PoolL2Tx, // Processed L2Txs
	nonSelectedL2Txs []common.PoolL2Tx, // L2Txs that are not selected but could get selected in future iterations
	unforjableL2Txs []common.PoolL2Tx, // Discarded txs that are impossible to forge (nonce too small or impossible to create destinatary account)
	failedAG failedAtomicGroup, // Info for the atomic group that failed
	err error,
) {
	const failedGroupErrMsg = "Failed forging atomic tx from Group %d. Restarting selection process without txs from this group"
	// TODO: differentiate between nonSelectedL2Txs and unforjableL2Txs (right now all fall into nonSelectedL2Txs, which is safe but non optimal)
	positionL1 := nAlreadyProcessedL1Txs
	// Iterate over l2Txs
	// - check Nonces
	// - check enough Balance for the Amount+Fee
	// - if needed, create new L1CoordinatorTxs for unexisting ToIdx
	// 	- keep used accAuths
	// - put the valid txs into selectedTxs array
	for i := 0; i < len(l2Txs); i++ {
		// Check if there is space for more L2Txs in the selection
		if !canAddL2Tx(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, selectionConfig) {
			// If tx is atomic, restart process without txs from the atomic group
			const failingMsg = "Tx not selected due not available slots for L2Txs"
			if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
				failedAG = failedAtomicGroup{
					id:     l2Txs[i].AtomicGroupID,
					txID:   l2Txs[i].TxID,
					reason: failingMsg,
				}
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
					failedGroupErrMsg,
					l2Txs[i].AtomicGroupID,
				))
			}
			// no more available slots for L2Txs, so mark this tx
			// but also the rest of remaining txs as discarded
			for j := i; j < len(l2Txs); j++ {
				l2Txs[j].Info = failingMsg
				nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[j])
			}
			break
		}

		// Discard exits with amount 0
		if l2Txs[i].Type == common.TxTypeExit && l2Txs[i].Amount.Cmp(big.NewInt(0)) <= 0 {
			// If tx is atomic, restart process without txs from the atomic group
			if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
				failedAG = failedAtomicGroup{
					id:     l2Txs[i].AtomicGroupID,
					txID:   l2Txs[i].TxID,
					reason: "Exits with amount 0 have no sense, not accepting to prevent unintended transactions",
				}
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
					failedGroupErrMsg,
					l2Txs[i].AtomicGroupID,
				))
			}
			l2Txs[i].Info = "Exits with amount 0 have no sense, not accepting to prevent unintended transactions"
			unforjableL2Txs = append(unforjableL2Txs, l2Txs[i]) // Although tecnicaly forjable, it won't never get forged with current code
			continue
		}

		// get Nonce & TokenID from the Account by l2Tx.FromIdx
		accSender, err := tp.StateDB().GetAccount(l2Txs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(err)
		}
		l2Txs[i].TokenID = accSender.TokenID

		// Check enough Balance on sender
		enoughBalance, balance, feeAndAmount := tp.CheckEnoughBalance(l2Txs[i])
		if !enoughBalance {
			failingMsg := fmt.Sprintf("Tx not selected due to not enough Balance at the sender. "+
				"Current sender account Balance: %s, Amount+Fee: %s",
				balance.String(), feeAndAmount.String())
			// If tx is atomic, restart process without txs from the atomic group
			if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
				failedAG = failedAtomicGroup{
					id:     l2Txs[i].AtomicGroupID,
					txID:   l2Txs[i].TxID,
					reason: failingMsg,
				}
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
					failedGroupErrMsg,
					l2Txs[i].AtomicGroupID,
				))
			}
			// not valid Amount with current Balance. Discard L2Tx,
			// and update Info parameter of the tx, and add it to
			// the discardedTxs array
			l2Txs[i].Info = failingMsg
			nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
			continue
		}

		// Check if Nonce is correct
		if l2Txs[i].Nonce != accSender.Nonce {
			failingMsg := fmt.Sprintf("Tx not selected due to not current Nonce. "+
				"Tx.Nonce: %d, Account.Nonce: %d", l2Txs[i].Nonce, accSender.Nonce)
			// If tx is atomic, restart process without txs from the atomic group
			if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
				failedAG = failedAtomicGroup{
					id:     l2Txs[i].AtomicGroupID,
					txID:   l2Txs[i].TxID,
					reason: failingMsg,
				}
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
					failedGroupErrMsg,
					l2Txs[i].AtomicGroupID,
				))
			}
			// not valid Nonce at tx. Discard L2Tx, and update Info
			// parameter of the tx, and add it to the discardedTxs
			// array
			l2Txs[i].Info = failingMsg
			nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
			continue
		}

		// if TokenID does not exist yet, create new L1CoordinatorTx to
		// create the CoordinatorAccount for that TokenID, to receive
		// the fees. Only in the case that there does not exist yet a
		// pending L1CoordinatorTx to create the account for the
		// Coordinator for that TokenID
		var newL1CoordTx *common.L1Tx
		newL1CoordTx, positionL1, err = txsel.coordAccountForTokenID(accSender.TokenID, positionL1)
		if err != nil {
			return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(err)
		}
		if newL1CoordTx != nil {
			const failingMsg = "Tx not selected because the L2Tx depends on a L1CoordinatorTx and there is not enough space for L1Coordinator"
			// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
			// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
			if !canAddL2TxThatNeedsNewCoordL1Tx(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, selectionConfig) {
				// If tx is atomic, restart process without txs from the atomic group
				if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
					failedAG = failedAtomicGroup{
						id:     l2Txs[i].AtomicGroupID,
						txID:   l2Txs[i].TxID,
						reason: failingMsg,
					}
					return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
						failedGroupErrMsg,
						l2Txs[i].AtomicGroupID,
					))
				}
				// discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = failingMsg
				nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
				continue
			}
			// increase positionL1
			positionL1++
			l1CoordinatorTxs = append(l1CoordinatorTxs, *newL1CoordTx)
			accAuths = append(accAuths, txsel.coordAccount.AccountCreationAuth)

			// process the L1CoordTx
			_, _, _, _, err := tp.ProcessL1Tx(nil, newL1CoordTx)
			if err != nil {
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(err)
			}
			nAlreadyProcessedL1Txs++
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
			validL2Tx, l1CoordinatorTx, accAuth, err := txsel.processTxToEthAddrBJJ(
				selectionConfig,
				l1UserFutureTxs,
				nAlreadyProcessedL1Txs,
				nAlreadyProcessedL2Txs,
				positionL1,
				l2Txs[i],
			)
			if err != nil {
				failingMsg := fmt.Sprintf("Tx not selected (in processTxToEthAddrBJJ) due to %s",
					err.Error())
				log.Debugw("txsel.processTxToEthAddrBJJ", "err", err)
				// If tx is atomic, restart process without txs from the atomic group
				if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
					failedAG = failedAtomicGroup{
						id:     l2Txs[i].AtomicGroupID,
						txID:   l2Txs[i].TxID,
						reason: failingMsg,
					}
					return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
						failedGroupErrMsg,
						l2Txs[i].AtomicGroupID,
					))
				}
				// Discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = failingMsg
				nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
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
					return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(err)
				}
				nAlreadyProcessedL1Txs++
			}
			if validL2Tx == nil {
				// TODO: Missing info on why this tx is not selected? Check l2Txs.Info at this point!
				// If tx is atomic, restart process without txs from the atomic group
				if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
					failedAG = failedAtomicGroup{
						id:     l2Txs[i].AtomicGroupID,
						txID:   l2Txs[i].TxID,
						reason: l2Txs[i].Info,
					}
					return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
						failedGroupErrMsg,
						l2Txs[i].AtomicGroupID,
					))
				}
				nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
				continue
			}
		} else if l2Txs[i].ToIdx >= common.IdxUserThreshold {
			_, err := txsel.localAccountsDB.GetAccount(l2Txs[i].ToIdx)
			if err != nil {
				failingMsg := fmt.Sprintf("Tx not selected due to tx.ToIdx not found in StateDB. "+
					"ToIdx: %d", l2Txs[i].ToIdx)
				// tx not valid
				log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
					"ToIdx", l2Txs[i].ToIdx)
				// If tx is atomic, restart process without txs from the atomic group
				if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
					failedAG = failedAtomicGroup{
						id:     l2Txs[i].AtomicGroupID,
						txID:   l2Txs[i].TxID,
						reason: failingMsg,
					}
					return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
						failedGroupErrMsg,
						l2Txs[i].AtomicGroupID,
					))
				}
				// Discard L2Tx, and update Info parameter of
				// the tx, and add it to the discardedTxs array
				l2Txs[i].Info = failingMsg
				nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
				continue
			}
		}

		// get CoordIdxsMap for the TokenID of the current l2Txs[i]
		// get TokenID from tx.Sender account
		tokenID := accSender.TokenID
		coordIdx, err := txsel.getCoordIdx(tokenID)
		if err != nil {
			// if err is db.ErrNotFound, should not happen, as all
			// the selectedTxs.TokenID should have a CoordinatorIdx
			// created in the DB at this point
			return nil, nil, nil, nil, nil, failedAG,
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
			failingMsg := fmt.Sprintf("Tx not selected (in ProcessL2Tx) due to %s",
				err.Error())
			log.Debugw("txselector.getL1L2TxSelection at ProcessL2Tx", "err", err)
			// If tx is atomic, restart process without txs from the atomic group
			if l2Txs[i].AtomicGroupID != common.EmptyAtomicGroupID {
				failedAG = failedAtomicGroup{
					id:     l2Txs[i].AtomicGroupID,
					txID:   l2Txs[i].TxID,
					reason: failingMsg,
				}
				return nil, nil, nil, nil, nil, failedAG, tracerr.Wrap(fmt.Errorf(
					failedGroupErrMsg,
					l2Txs[i].AtomicGroupID,
				))
			}
			// Discard L2Tx, and update Info parameter of the tx,
			// and add it to the discardedTxs array
			l2Txs[i].Info = failingMsg
			nonSelectedL2Txs = append(nonSelectedL2Txs, l2Txs[i])
			continue
		}
		nAlreadyProcessedL2Txs++

		selectedL2Txs = append(selectedL2Txs, l2Txs[i])
	} // after this loop, no checks to discard txs should be done

	return accAuths, l1CoordinatorTxs, selectedL2Txs, nonSelectedL2Txs, unforjableL2Txs, failedAG, nil
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created
func (txsel *TxSelector) processTxToEthAddrBJJ(
	selectionConfig txprocessor.Config, l1UserFutureTxs []common.L1Tx,
	nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1 int, l2Tx common.PoolL2Tx) (
	*common.PoolL2Tx, *common.L1Tx, *common.AccountCreationAuth, error) {
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
	if !canAddL2TxThatNeedsNewCoordL1Tx(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, selectionConfig) {
		// L2Tx discarded
		return nil, nil, nil, tracerr.Wrap(fmt.Errorf("L2Tx discarded due to no available slots " +
			"for L1CoordinatorTx to create a new account for receiver of L2Tx"))
	}

	return &l2Tx, l1CoordinatorTx, accAuth, nil
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

// sortL2Txs sorts the PoolL2Txs by AverageFee if they are atomic and AbsoluteFee and then by Nonce if they aren't
// atomic txs that are within the same atomic group are guaranteed to manatin order and consecutiveness
// Assumption the order within the atomic groups is correct for l2Txs
func sortL2Txs(l2Txs []common.PoolL2Tx, atomicGroupsFee map[common.AtomicGroupID]float64) []common.PoolL2Tx {
	// Separate atomic txs
	atomicGroupsMap := make(map[common.AtomicGroupID][]common.PoolL2Tx)
	nonAtomicTxs := []common.PoolL2Tx{}
	for i := 0; i < len(l2Txs); i++ {
		groupID := l2Txs[i].AtomicGroupID
		if groupID != common.EmptyAtomicGroupID { // If it's an atomic tx
			if _, ok := atomicGroupsMap[groupID]; !ok { // If it's the first tx of the group initialise slice
				atomicGroupsMap[groupID] = []common.PoolL2Tx{}
			}
			atomicGroupsMap[groupID] = append(atomicGroupsMap[groupID], l2Txs[i])
		} else { // If it's a non atomic tx
			nonAtomicTxs = append(nonAtomicTxs, l2Txs[i])
		}
	}
	// Sort atomic groups by average fee
	// First, convert map to slice
	atomicGroups := [][]common.PoolL2Tx{}
	for groupID := range atomicGroupsMap {
		atomicGroups = append(atomicGroups, atomicGroupsMap[groupID])
	}
	sort.SliceStable(atomicGroups, func(i, j int) bool {
		// Sort by the average fee of each tx group
		// assumption: each atomic group has at least one tx, and they all share the same groupID
		return atomicGroupsFee[atomicGroups[i][0].AtomicGroupID] >
			atomicGroupsFee[atomicGroups[j][0].AtomicGroupID]
	})

	// Sort non atomic txs by absolute fee with SliceStable, so that txs with same
	// AbsoluteFee are not rearranged and nonce order is kept in such case
	sort.SliceStable(nonAtomicTxs, func(i, j int) bool {
		return nonAtomicTxs[i].AbsoluteFee > nonAtomicTxs[j].AbsoluteFee
	})

	// sort non atomic txs by Nonce. This can be done in many different ways, what
	// is needed is to output the l2Txs where the Nonce of l2Txs for each
	// Account is sorted, but the l2Txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.Slice(nonAtomicTxs, func(i, j int) bool {
		return nonAtomicTxs[i].Nonce < nonAtomicTxs[j].Nonce
	})

	// Combine atomic and non atomic txs in a single slice, ordering them by AbsoluteFee vs AverageFee
	// and making sure that the atomic txs within same groups are consecutive and preserve the original order (otherwise the RqOffsets will broke)
	sortedL2Txs := []common.PoolL2Tx{}
	var nextNonAtomicToAppend, nextAtomicGroupToAppend int
	// Iterate until all the non atoic txs has been appended OR all the atomic txs inside atomic groups has been appended
	for nextNonAtomicToAppend != len(nonAtomicTxs) && nextAtomicGroupToAppend != len(atomicGroups) {
		if nonAtomicTxs[nextNonAtomicToAppend].AbsoluteFee >
			atomicGroupsFee[atomicGroups[nextAtomicGroupToAppend][0].AtomicGroupID] {
			// The fee of the next non atomic txs is greater than the average fee of the next atomic group
			sortedL2Txs = append(sortedL2Txs, nonAtomicTxs[nextNonAtomicToAppend])
			nextNonAtomicToAppend++
		} else {
			// The fee of the next non atomic txs is smaller than the average fee of the next atomic group
			// append all the txs of the group
			sortedL2Txs = append(sortedL2Txs, atomicGroups[nextAtomicGroupToAppend]...)
			nextAtomicGroupToAppend++
		}
	}
	// At this point one of the two slices (nonAtomicTxs and atomicGroups) is fully apended to sortedL2Txs
	// while the other is not. Append remaining txs
	if nextNonAtomicToAppend == len(nonAtomicTxs) { // nonAtomicTxs is fully appended, append remaining txs in atomicGroups
		for i := nextAtomicGroupToAppend; i < len(atomicGroups); i++ {
			sortedL2Txs = append(sortedL2Txs, atomicGroups[i]...)
		}
	} else { // all txs in atomicGroups appended, append remaining nonAtomicTxs
		for i := nextNonAtomicToAppend; i < len(nonAtomicTxs); i++ {
			sortedL2Txs = append(sortedL2Txs, nonAtomicTxs[i])
		}
	}
	return sortedL2Txs
}

func canAddL2TxThatNeedsNewCoordL1Tx(nAddedL1Txs, nAddedL2txs int, selectionConfig txprocessor.Config) bool {
	return nAddedL1Txs < int(selectionConfig.MaxL1Tx) && // Capacity for L1s already reached
		nAddedL1Txs+nAddedL2txs+1 < int(selectionConfig.MaxTx)
}

func canAddL2Tx(nAddedL1Txs, nAddedL2txs int, selectionConfig txprocessor.Config) bool {
	return nAddedL1Txs+nAddedL2txs < int(selectionConfig.MaxTx)
}

// filterFailedAtomicGroups split the txs into the ones that can be porcessed and the ones that can't because they belong
// to an AtomicGroupID that is part of failedGroups. The order of txsToProcess is consistent with the order of txs
func filterFailedAtomicGroups(txs []common.PoolL2Tx, faildeGroups []failedAtomicGroup) (txsToProcess []common.PoolL2Tx, filteredTxs []common.PoolL2Tx) {
	// Filter failed atomic groups
	for i := 0; i < len(txs); i++ {
		if txs[i].AtomicGroupID == common.EmptyAtomicGroupID {
			// Tx is not atomic, not filtering
			txsToProcess = append(txsToProcess, txs[i])
			continue
		}
		txFailed := false
		for _, failedAtomicGroup := range faildeGroups {
			if failedAtomicGroup.id == txs[i].AtomicGroupID {
				txs[i].Info = setInfoForFailedAtomicTx(
					txs[i].TxID == failedAtomicGroup.txID,
					txs[i].AtomicGroupID,
					failedAtomicGroup.txID,
					failedAtomicGroup.reason,
				)
				filteredTxs = append(filteredTxs, txs[i])
				txFailed = true
				break
			}
		}
		if !txFailed {
			txsToProcess = append(txsToProcess, txs[i])
		}
	}
	return txsToProcess, filteredTxs
}

// calculateAtomicGroupsAverageFee generates a map that represents AtomicGroupID => average absolute fee of the transactions of the AtomicGroupID
func calculateAtomicGroupsAverageFee(txs []common.PoolL2Tx) map[common.AtomicGroupID]float64 {
	txsPerGroup := make(map[common.AtomicGroupID]int)
	groupAverageFee := make(map[common.AtomicGroupID]float64)
	// Set sum of absolute fee per group
	for i := 0; i < len(txs); i++ {
		groupID := txs[i].AtomicGroupID
		if groupID == common.EmptyAtomicGroupID {
			// Not an atomic tx
			continue
		}
		// Add the absolute fee to the relevant group and increase the txs counted on the group
		if _, ok := groupAverageFee[groupID]; !ok {
			groupAverageFee[groupID] = txs[i].AbsoluteFee
			txsPerGroup[groupID] = 1
		} else {
			groupAverageFee[groupID] += txs[i].AbsoluteFee
			txsPerGroup[groupID]++
		}
	}
	// Calculate the average fee based on how many txs have each group
	for groupID := range groupAverageFee {
		groupAverageFee[groupID] = groupAverageFee[groupID] / float64(txsPerGroup[groupID])
	}
	return groupAverageFee
}

func setInfoForFailedAtomicTx(
	isOriginOfFailure bool,
	failedAtomicGroupID common.AtomicGroupID,
	failedTxID common.TxID,
	failMessage string,
) string {
	if isOriginOfFailure {
		return fmt.Sprintf(
			"Unselectable atomic group %s, tx %s failed due to: %s",
			failedAtomicGroupID,
			failedTxID,
			failMessage,
		)
	}
	return failMessage
}
