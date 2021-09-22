package txselector

import (
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api"
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
	id         common.AtomicGroupID
	failedTxID common.TxID // ID of the tx that made the entire atomic group fail
	reason     common.TxSelectorError
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
	start := time.Now()
	defer func(start time.Time) {
		dur := time.Since(start)
		log.Infof("transaction selection took %d to execute", dur.Milliseconds())
	}(start)
	failedAtomicGroups := make(map[common.AtomicGroupID]failedAtomicGroup)
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
	selectableTxsTmp, discardedTxs := filterFailedAtomicGroups(l2TxsFromDB, failedAtomicGroups)
	// Filter invalid atomic groups
	selectableTxs, discardedTxsTmp := filterInvalidAtomicGroups(selectableTxsTmp)
	discardedTxs = append(discardedTxs, discardedTxsTmp...)

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

	// Tracking if any tx was selected in the round of tx verification
	var isAnyTxWasSelected bool
	// Used authorizations in the l1CoordinatorTxs
	var accAuths [][]byte
	// Processed txs for necessary account creation
	// (fees for coordinator or missing destinatary accounts)
	var l1CoordinatorTxs []common.L1Tx
	// Processed txs
	var selectedTxs []common.PoolL2Tx
	// txs that are not selected in the round of verification, but can be selected in the next round
	var nonSelectedTxs []common.PoolL2Tx
	// coordinator account idx to collect fees
	var coordIdxs []common.Idx

	// Start selection process

	// init orchestrator, that will send txs to the following channels
	txOrchestrator := newTxOrchestrator(
		l1UserFutureTxs,
		selectionConfig,
		tp,
		txsel,
	)
	// Order L2 txs. This has to be done just once,
	// as the array will get smaller over iterations, but the order won't be affected
	selectableTxs = sortL2Txs(selectableTxs, atomicFeeMap)
	// Iterate over l2Txs
	// - check Nonces
	// - check enough Balance for the Amount+Fee
	// - if needed, create new L1CoordinatorTxs for unexisting ToIdx
	// 	- keep used accAuths
	// - put the valid txs into selectedTxs array
	go txsel.verifyTxs(
		txOrchestrator,
		selectableTxs,                        // Txs that can be selected
		len(l1UserTxs)+len(l1CoordinatorTxs), // Already added L1 Txs
		len(selectedTxs),                     // Already added L2 Txs
	)

	// get CoordIdxsMap for the TokenIDs
	coordIdxsMap := make(map[common.TokenID]common.Idx)
	for {
		select {
		case tx := <-txOrchestrator.txsL1ToBeProcessedChan:
			_, _, _, _, err = tp.ProcessL1Tx(nil, tx)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			l1CoordinatorTxs = append(l1CoordinatorTxs, *tx)
			txOrchestrator.l1TxsWg.Done()
		case tx := <-txOrchestrator.nonSelectedL2TxsChan:
			nonSelectedTxs = append(nonSelectedTxs, tx)
		case tx := <-txOrchestrator.selectedL2TxsChan:
			isAnyTxWasSelected = true
			selectedTxs = append(selectedTxs, tx)
			accSender, err := tp.StateDB().GetAccount(tx.FromIdx)
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
			if _, ok := coordIdxsMap[tokenID]; !ok {
				coordIdxsMap[tokenID] = coordIdx
				coordIdxs = append(coordIdxs, coordIdx)
			}
		case tx := <-txOrchestrator.unforjableL2TxsChan:
			discardedTxs = append(discardedTxs, tx)
		case accAuth := <-txOrchestrator.accAuthsChan:
			accAuths = append(accAuths, accAuth)
		case failedAG := <-txOrchestrator.failedAGChan:
			failedAtomicGroups[failedAG.id] = failedAG
			if err := txsel.localAccountsDB.Reset(txsel.localAccountsDB.CurrentBatch(), false); err != nil {
				return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
			// go to the start of the selection process, bcs if one of transaction in atomic group failed, than every other transaction should be rejected
			goto START_SELECTION
		case err := <-txOrchestrator.errChan:
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		case <-txOrchestrator.verificationEnded:
			if !isAnyTxWasSelected {
				discardedTxs = append(discardedTxs, nonSelectedTxs...)
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
			// some txs was selected, so we can select one more time, bcs conditions, like account balance can be changed after previous round
			isAnyTxWasSelected = false
			selectableTxs = nonSelectedTxs
			nonSelectedTxs = []common.PoolL2Tx{}
			go txsel.verifyTxs(
				txOrchestrator,
				selectableTxs,                        // Txs that can be selected
				len(l1UserTxs)+len(l1CoordinatorTxs), // Already added L1 Txs
				len(selectedTxs),                     // Already added L2 Txs
			)
		}
	}
}

// verifyTxs is verifing txs and sent them to the according channels
func (txsel *TxSelector) verifyTxs(
	txOrchestrator *txOrchestrator,
	l2Txs []common.PoolL2Tx,
	alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount int,
) {
	positionL1 := alreadyProcessedL1TxsAmount
	nextBatchNum := uint32(txsel.localAccountsDB.CurrentBatch()) + 1
	// this tx group needed to track that orchestrator processed all the l2txs
	var txsWg sync.WaitGroup
	txsWg.Add(len(l2Txs))

	txOrchestrator.txsWg = &txsWg

	l2TxsChan := txOrchestrator.transformTxs(l2Txs)
	l2TxsChan = txOrchestrator.checkBatchGreaterThanMaxNumBatch(l2TxsChan, nextBatchNum)
	l2TxsChan = txOrchestrator.checkIsExitWithZeroAmount(l2TxsChan)
	l2TxsChan = txOrchestrator.verifyTxs(l2TxsChan, alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, positionL1)
	txOrchestrator.processL2Txs(l2TxsChan)

	go func() {
		// wait until all transactions are processed, so verification round can be ended
		txOrchestrator.txsWg.Wait()
		txOrchestrator.verificationEnded <- 1
	}()
}

// sortL2Txs sorts the PoolL2Txs by AverageFee if they are atomic and AbsoluteFee
// and then by Nonce if they aren't atomic txs that are within the same atomic group
// are guaranteed to manatin order and consecutiveness.
// Assumption: the order within the atomic groups is correct for l2Txs
func sortL2Txs(
	l2Txs []common.PoolL2Tx,
	atomicGroupsFee map[common.AtomicGroupID]float64,
) []common.PoolL2Tx {
	// Separate atomic txs
	atomicGroupsMap := make(map[common.AtomicGroupID][]common.PoolL2Tx)
	nonAtomicTxs := []common.PoolL2Tx{}
	for i := 0; i < len(l2Txs); i++ {
		groupID := l2Txs[i].AtomicGroupID
		if groupID != common.EmptyAtomicGroupID { // If it's an atomic tx
			// If it's the first tx of the group initialise slice
			if _, ok := atomicGroupsMap[groupID]; !ok {
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

	// Sort non atomic txs by AbsoluteFee DESC, then by FromIdx ASC and then
	// by Nonce ASC.
	//
	// This sorting sequence allows us to select firstly the most profitable
	// txs, even though this can mess with the Nonce sequence, but since the
	// Nonce sequence is only one of the rules in order to a txs be selected
	// and we have a txs reprocessing strategy, the tx selector will try to
	// select the `wrong nonce` txs in the next interation until it identifies
	// there is no more txs to be selected at this moment, when this situation
	// happens we can assume the tx selector selected all the most profitable
	// txs that can be processed at this moment.

	sort.Slice(nonAtomicTxs, func(i, j int) bool {
		if nonAtomicTxs[i].AbsoluteFee != nonAtomicTxs[j].AbsoluteFee {
			return nonAtomicTxs[i].AbsoluteFee > nonAtomicTxs[j].AbsoluteFee
		}

		if nonAtomicTxs[i].FromIdx != nonAtomicTxs[j].FromIdx {
			return nonAtomicTxs[i].FromIdx < nonAtomicTxs[j].FromIdx
		}

		return nonAtomicTxs[i].Nonce < nonAtomicTxs[j].Nonce
	})

	// Combine atomic and non atomic txs in a single slice, ordering them by AbsoluteFee vs AverageFee
	// and making sure that the atomic txs within same groups are consecutive
	// and preserve the original order (otherwise the RqOffsets will broke)
	sortedL2Txs := []common.PoolL2Tx{}
	var nextNonAtomicToAppend, nextAtomicGroupToAppend int
	// Iterate until all the non atoic txs has been appended
	// OR all the atomic txs inside atomic groups has been appended
	for nextNonAtomicToAppend != len(nonAtomicTxs) && nextAtomicGroupToAppend != len(atomicGroups) {
		if nonAtomicTxs[nextNonAtomicToAppend].AbsoluteFee >
			atomicGroupsFee[atomicGroups[nextAtomicGroupToAppend][0].AtomicGroupID] {
			// The fee of the next non atomic txs is greater
			// than the average fee of the next atomic group
			sortedL2Txs = append(sortedL2Txs, nonAtomicTxs[nextNonAtomicToAppend])
			nextNonAtomicToAppend++
		} else {
			// The fee of the next non atomic txs is smaller
			// than the average fee of the next atomic group
			// append all the txs of the group
			sortedL2Txs = append(sortedL2Txs, atomicGroups[nextAtomicGroupToAppend]...)
			nextAtomicGroupToAppend++
		}
	}
	// At this point one of the two slices (nonAtomicTxs and atomicGroups)
	//  is fully apended to sortedL2Txs while the other is not. Append remaining txs
	if nextNonAtomicToAppend == len(nonAtomicTxs) {
		// nonAtomicTxs is fully appended, append remaining txs in atomicGroups
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

func canAddL2TxThatNeedsNewCoordL1Tx(
	addedL1TxsAmount, addedL2txsAmount int,
	selectionConfig txprocessor.Config,
) bool {
	return addedL1TxsAmount < int(selectionConfig.MaxL1Tx) && // Capacity for L1s already reached
		addedL1TxsAmount+addedL2txsAmount+1 < int(selectionConfig.MaxTx)
}

func canAddL2Tx(addedL1TxsAmount, addedL2txsAmount int, selectionConfig txprocessor.Config) bool {
	return addedL1TxsAmount+addedL2txsAmount < int(selectionConfig.MaxTx)
}

// filterFailedAtomicGroups split the txs into the ones that can be porcessed
//  and the ones that can't because they belong to an AtomicGroupID that is part of failedGroups.
// The order of txsToProcess is consistent with the order of txs
func filterFailedAtomicGroups(
	txs []common.PoolL2Tx,
	failedGroups map[common.AtomicGroupID]failedAtomicGroup,
) (txsToProcess []common.PoolL2Tx, filteredTxs []common.PoolL2Tx) {
	// Filter failed atomic groups
	for i := 0; i < len(txs); i++ {
		if txs[i].AtomicGroupID == common.EmptyAtomicGroupID {
			// Tx is not atomic, not filtering
			txsToProcess = append(txsToProcess, txs[i])
			continue
		}
		txFailed := false
		if failedAtomicGroup, ok := failedGroups[txs[i].AtomicGroupID]; ok {
			obj := setInfoForFailedAtomicTx(
				txs[i].TxID == failedAtomicGroup.failedTxID,
				txs[i].AtomicGroupID,
				failedAtomicGroup.failedTxID,
				failedAtomicGroup.reason,
			)
			txs[i].Info = obj.Message
			txs[i].ErrorCode = obj.Code
			txs[i].ErrorType = obj.Type
			filteredTxs = append(filteredTxs, txs[i])
			txFailed = true
		}
		if !txFailed {
			txsToProcess = append(txsToProcess, txs[i])
		}
	}
	return txsToProcess, filteredTxs
}

// filterInvalidAtomicGroups split the txs into the ones that can be processed
// and the ones that can't because they belong to an AtomicGroup that is impossible to forge
// due to missing or bad ordered txs
func filterInvalidAtomicGroups(
	txs []common.PoolL2Tx,
) (txsToProcess []common.PoolL2Tx, filteredTxs []common.PoolL2Tx) {
	// Separate txs into atomic groups
	atomicGroups := make(map[common.AtomicGroupID]common.AtomicGroup)
	for i := 0; i < len(txs); i++ {
		atomicGroupID := txs[i].AtomicGroupID
		if atomicGroupID == common.EmptyAtomicGroupID {
			// Tx is not atomic, not filtering
			txsToProcess = append(txsToProcess, txs[i])
			continue
		}
		if atomicGroup, ok := atomicGroups[atomicGroupID]; !ok {
			atomicGroups[atomicGroupID] = common.AtomicGroup{
				Txs: []common.PoolL2Tx{txs[i]},
			}
		} else {
			atomicGroup.Txs = append(atomicGroup.Txs, txs[i])
			atomicGroups[atomicGroupID] = atomicGroup
		}
	}
	// Validate atomic groups
	for _, atomicGroup := range atomicGroups {
		if !isAtomicGroupValid(atomicGroup) {
			// Set Info message and add txs of the atomic group to filteredTxs
			for i := 0; i < len(atomicGroup.Txs); i++ {
				atomicGroup.Txs[i].Info = ErrInvalidAtomicGroup
				atomicGroup.Txs[i].ErrorType = ErrInvalidAtomicGroupType
				atomicGroup.Txs[i].ErrorCode = ErrInvalidAtomicGroupCode
				filteredTxs = append(filteredTxs, atomicGroup.Txs[i])
			}
		} else {
			// Atomic group is valid, add txs of the atomic group to txsToProcess
			for i := 0; i < len(atomicGroup.Txs); i++ {
				txsToProcess = append(txsToProcess, atomicGroup.Txs[i])
			}
		}
	}
	return txsToProcess, filteredTxs
}

func isAtomicGroupValid(atomicGroup common.AtomicGroup) bool {
	for i := 0; i < len(atomicGroup.Txs); i++ {
		// Find requested tx
		rqRelativePosition, err := api.RequestOffset2RelativePosition(atomicGroup.Txs[i].RqOffset)
		if err != nil {
			return false
		}
		requestedPosition := i + rqRelativePosition
		if requestedPosition > len(atomicGroup.Txs)-1 || requestedPosition < 0 {
			return false
		}
		requestedTx := atomicGroup.Txs[requestedPosition]
		// Check if the requested tx match the Rq fields
		if atomicGroup.Txs[i].RqFromIdx != requestedTx.FromIdx ||
			atomicGroup.Txs[i].RqToIdx != requestedTx.ToIdx ||
			atomicGroup.Txs[i].RqToEthAddr != requestedTx.ToEthAddr ||
			atomicGroup.Txs[i].RqToBJJ != requestedTx.ToBJJ ||
			atomicGroup.Txs[i].RqTokenID != requestedTx.TokenID ||
			atomicGroup.Txs[i].RqFee != requestedTx.Fee ||
			atomicGroup.Txs[i].RqNonce != requestedTx.Nonce {
			return false
		}
		// Check amount
		if atomicGroup.Txs[i].RqAmount != nil && requestedTx.Amount != nil {
			// If both are different to nil
			if atomicGroup.Txs[i].RqAmount.Cmp(requestedTx.Amount) != 0 {
				// They must have same value (to be valid)
				return false
			}
		} else if atomicGroup.Txs[i].RqAmount != requestedTx.Amount {
			// Else they must be both nil (to be valid)
			return false
		}
	}
	return true
}

// calculateAtomicGroupsAverageFee generates a map
// that represents AtomicGroupID => average absolute fee
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
	failMessage common.TxSelectorError,
) common.TxSelectorError {
	if isOriginOfFailure {
		obj := common.TxSelectorError{
			Message: failMessage.Message,
			Code:    failMessage.Code,
			Type:    failMessage.Type,
		}
		return obj
	}
	return common.TxSelectorError{
		Message: fmt.Sprintf("unselectable atomic group"+" %s, tx %s failed due to: %s",
			failedAtomicGroupID,
			failedTxID,
			failMessage.Message,
		),
		Code: ErrInvalidAtomicGroupCode,
		Type: ErrInvalidAtomicGroupType,
	}
}
