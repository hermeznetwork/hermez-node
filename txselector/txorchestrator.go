package txselector

import (
	"fmt"
	"math/big"
	"sync"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const failedGroupErrMsg = "Failed forging atomic tx from Group %s." +
	" Restarting selection process without txs from this group"

// txOrchestrator this struct is for holding channels and propagate txs between channels
type txOrchestrator struct {
	failedAGChan                                                 chan failedAtomicGroup   // tracking failed atomic group
	errChan                                                      chan error               // tracking errors. If error is faced, return from the method
	unforjableL2TxsChan, nonSelectedL2TxsChan, selectedL2TxsChan chan common.PoolL2Tx     // tracking different types of txs
	txsL1ToBeProcessedChan                                       chan *common.L1Tx        // tracking l1 txs that have to be processed
	accAuthsChan                                                 chan []byte              // tracking account auths, that has to be created to process l2
	verificationEnded                                            chan int                 // tracking verification round ending
	l1UserFutureTxs                                              []common.L1Tx            // pending l1 tx
	selectionConfig                                              txprocessor.Config       // config of tx selection
	tp                                                           *txprocessor.TxProcessor // tx processor
	txsel                                                        *TxSelector              // tx selector
	l1TxsWg                                                      *sync.WaitGroup          // wait group to track proceeded l1 tx, so depended l2 tx can be processed
	txsWg                                                        *sync.WaitGroup          // wait group to track processed l2 tx for verification round ended
}

// newTxOrchestrator creates new tx orchestrator and init related channels
func newTxOrchestrator(
	l1UserFutureTxs []common.L1Tx,
	selectionConfig txprocessor.Config,
	tp *txprocessor.TxProcessor,
	txsel *TxSelector) *txOrchestrator {
	return &txOrchestrator{
		failedAGChan:           make(chan failedAtomicGroup),
		errChan:                make(chan error),
		unforjableL2TxsChan:    make(chan common.PoolL2Tx),
		nonSelectedL2TxsChan:   make(chan common.PoolL2Tx),
		selectedL2TxsChan:      make(chan common.PoolL2Tx),
		txsL1ToBeProcessedChan: make(chan *common.L1Tx),
		accAuthsChan:           make(chan []byte),
		verificationEnded:      make(chan int),
		l1UserFutureTxs:        l1UserFutureTxs, // Used to prevent the creation of unnecessary accounts
		selectionConfig:        selectionConfig,
		tp:                     tp,
		txsel:                  txsel,
		l1TxsWg:                &sync.WaitGroup{},
	}
}

// transformTxs transform txs array to the txs channel
func (txorches *txOrchestrator) transformTxs(l2txs []common.PoolL2Tx) <-chan common.PoolL2Tx {
	l2txsChanOut := make(chan common.PoolL2Tx)
	go func() {
		for _, tx := range l2txs {
			l2txsChanOut <- tx
		}
		close(l2txsChanOut)
	}()
	return l2txsChanOut
}

// checkBatchGreaterThanMaxNumBatch check if batch in tx is not greater than max num batch specified in config
// maxBatchNum - batch, until the transaction can be forged
func (txorches *txOrchestrator) checkBatchGreaterThanMaxNumBatch(l2txsChanInput <-chan common.PoolL2Tx, nextBatchNum uint32) <-chan common.PoolL2Tx {
	l2txsChanOut := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChanInput {
			if !txorches.isBatchGreaterThanMaxNumBatch(l2tx, nextBatchNum) {
				txorches.txsWg.Done()
				continue
			}
			l2txsChanOut <- l2tx
		}
		close(l2txsChanOut)
	}()
	return l2txsChanOut
}

// checkIsExitWithZeroAmount check if exit have zero amount. It's not needed to forge exit with zero amount
func (txorches *txOrchestrator) checkIsExitWithZeroAmount(l2txsChanInput <-chan common.PoolL2Tx) <-chan common.PoolL2Tx {
	l2txsChanOut := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChanInput {
			if !txorches.isExitWithZeroAmount(l2tx) {
				txorches.txsWg.Done()
				continue
			}
			l2txsChanOut <- l2tx
		}
		close(l2txsChanOut)
	}()
	return l2txsChanOut
}

// verifyTxs func to verify transactions, check for enough space to create l1/l2 tx, correct acc nonce, enough balance on sender
func (txorches *txOrchestrator) verifyTxs(l2txsChanInput <-chan common.PoolL2Tx, alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, positionL1 int) <-chan common.PoolL2Tx {
	l2txsChanOut := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChanInput {
			var isTxCorrect bool
			isTxCorrect = txorches.isEnoughSpace(alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, l2tx)
			if !isTxCorrect {
				txorches.txsWg.Done()
				continue
			}

			// Check enough Balance on sender
			isTxCorrect = txorches.isEnoughBalanceOnSender(l2tx)
			if !isTxCorrect {
				txorches.txsWg.Done()
				continue
			}

			// get Nonce & TokenID from the Account by l2Tx.FromIdx
			accSender, err := txorches.tp.StateDB().GetAccount(l2tx.FromIdx)
			if err != nil {
				txorches.errChan <- tracerr.Wrap(err)
				return
			}
			l2tx.TokenID = accSender.TokenID

			// Check if Nonce is correct
			isTxCorrect = txorches.isNonceCorrect(l2tx, accSender)
			if !isTxCorrect {
				txorches.txsWg.Done()
				continue
			}

			// if TokenID does not exist yet, create new L1CoordinatorTx to
			// create the CoordinatorAccount for that TokenID, to receive
			// the fees. Only in the case that there does not exist yet a
			// pending L1CoordinatorTx to create the account for the
			// Coordinator for that TokenID
			var newL1CoordTx *common.L1Tx
			newL1CoordTx, positionL1, err = txorches.txsel.coordAccountForTokenID(accSender.TokenID, positionL1)
			if err != nil {
				txorches.errChan <- tracerr.Wrap(err)
				return
			}
			if newL1CoordTx != nil {
				// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
				// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
				isTxCorrect = txorches.isEnoughSpaceForL1CoordTx(l2tx, alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount)
				if !isTxCorrect {
					txorches.txsWg.Done()
					continue
				}

				// increase positionL1
				positionL1++
				//l1CoordinatorTxsChan <- *newL1CoordTx
				txorches.accAuthsChan <- txorches.txsel.coordAccount.AccountCreationAuth

				// process the L1CoordTx
				txorches.l1TxsWg.Add(1)
				txorches.txsL1ToBeProcessedChan <- newL1CoordTx
				alreadyProcessedL1TxsAmount++
			}

			// If tx.ToIdx>=256, tx.ToIdx should exist to localAccountsDB,
			// if so, tx is used.  If tx.ToIdx==0, for an L2Tx will be the
			// case of TxToEthAddr or TxToBJJ, check if
			// tx.ToEthAddr/tx.ToBJJ exist in localAccountsDB, if yes tx is
			// used; if not, check if tx.ToEthAddr is in
			// AccountCreationAuthDB, if so, tx is used and L1CoordinatorTx
			// of CreateAccountAndDeposit is created. If tx.ToIdx==1, is a
			// Exit type and is used.
			var (
				validL2Tx       *common.PoolL2Tx
				l1CoordinatorTx *common.L1Tx
				accAuth         *common.AccountCreationAuth
			)
			if l2tx.ToIdx == 0 { // ToEthAddr/ToBJJ case
				validL2Tx, l1CoordinatorTx, accAuth, isTxCorrect = txorches.verifyAndBuildForTxToEthAddrBJJ(
					l2tx, alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, positionL1)

				if !isTxCorrect {
					txorches.txsWg.Done()
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
						txorches.accAuthsChan <- common.EmptyEthSignature
					} else if accAuth != nil {
						txorches.accAuthsChan <- accAuth.Signature
					}
					positionL1++
					// process the L1CoordTx
					txorches.l1TxsWg.Add(1)
					txorches.txsL1ToBeProcessedChan <- l1CoordinatorTx
					alreadyProcessedL1TxsAmount++
				}
				if validL2Tx == nil {
					// Missing info on why this tx is not selected? Check l2Txs.Info at this point!
					// If tx is atomic, restart process without txs from the atomic group
					txSelErr := common.TxSelectorError{
						Message: l2tx.Info,
						Code:    l2tx.ErrorCode,
						Type:    l2tx.ErrorType,
					}

					if txorches.isTxAtomic(l2tx, txSelErr) {
						return
					}
					txorches.nonSelectedL2TxsChan <- l2tx
					txorches.txsWg.Done()
					continue
				}
			} else if l2tx.ToIdx >= common.IdxUserThreshold {
				isTxCorrect = txorches.isToIdxExists(l2tx)
				if !isTxCorrect {
					txorches.txsWg.Done()
					continue
				}
			}

			// if l1tx is present, than we have to wait, until l1tx is processed, so l2tx can be safely processed too
			if newL1CoordTx != nil || l1CoordinatorTx != nil {
				txorches.l1TxsWg.Wait()
			}
			l2txsChanOut <- l2tx
			alreadyProcessedL2TxsAmount++
		}
		close(l2txsChanOut)
	}()
	return l2txsChanOut
}

// processL2Txs processing l2tx to see that every transaction is able to be processed
func (txorches *txOrchestrator) processL2Txs(l2txsChanInput <-chan common.PoolL2Tx) {
	go func() {
		for l2tx := range l2txsChanInput {
			txorches.tryToProcessL2Tx(l2tx)
			txorches.txsWg.Done()
		}
	}()
}

// tryToProcessL2Tx trying to process l2tx
func (txorches *txOrchestrator) tryToProcessL2Tx(
	l2tx common.PoolL2Tx) {
	// get CoordIdxsMap for the TokenID of the current l2Txs[i]
	// get TokenID from tx.Sender account
	tokenID := l2tx.TokenID
	coordIdx, err := txorches.txsel.getCoordIdx(tokenID)
	if err != nil {
		// if err is db.ErrNotFound, should not happen, as all
		// the selectedTxs.TokenID should have a CoordinatorIdx
		// created in the DB at this point
		txorches.errChan <- tracerr.Wrap(fmt.Errorf("could not get CoordIdx for TokenID=%d, "+
			"due: %s", tokenID, err))
		return
	}
	// prepare temp coordIdxsMap & AccumulatedFees for the call to
	// ProcessL2Tx
	coordIdxsMap := map[common.TokenID]common.Idx{tokenID: coordIdx}
	if _, ok := txorches.tp.AccumulatedFees[coordIdx]; !ok {
		txorches.tp.AccumulatedFees[coordIdx] = big.NewInt(0)
	}

	_, _, _, err = txorches.tp.ProcessL2Tx(coordIdxsMap, nil, nil, &l2tx)
	if err != nil {
		txSelErr := common.TxSelectorError{
			Message: fmt.Sprintf(ErrTxDiscartedInProcessL2Tx+" due to %s", err.Error()),
			Code:    ErrTxDiscartedInProcessL2TxCode,
			Type:    ErrTxDiscartedInProcessL2TxType,
		}
		log.Debugw("txselector.getL1L2TxSelection at ProcessL2Tx", "err", err)
		// If tx is atomic, restart process without txs from the atomic group
		if l2tx.AtomicGroupID != common.EmptyAtomicGroupID {
			failedAG := failedAtomicGroup{
				id:         l2tx.AtomicGroupID,
				failedTxID: l2tx.TxID,
				reason:     txSelErr,
			}
			txorches.failedAGChan <- failedAG
			return
		}
		// Discard L2Tx, and update Info parameter of the tx,
		// and add it to the discardedTxs array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return
	}
	txorches.selectedL2TxsChan <- l2tx
}

// isTxAtomic check if transaction atomic or not. If tx atomic, then add failedAtomicGroup to the channel and error to the error channel
func (txorches *txOrchestrator) isTxAtomic(l2tx common.PoolL2Tx, reason common.TxSelectorError) bool {
	if l2tx.AtomicGroupID != common.EmptyAtomicGroupID {
		failedAG := failedAtomicGroup{
			id:         l2tx.AtomicGroupID,
			failedTxID: l2tx.TxID,
			reason:     reason,
		}
		err := tracerr.Wrap(fmt.Errorf(
			failedGroupErrMsg,
			l2tx.AtomicGroupID.String(),
		))

		txorches.failedAGChan <- failedAG
		txorches.errChan <- err
		return true
	}
	return false
}

// isEnoughSpace checks for enough space for l2tx
// nAddedL1Txs+nAddedL2txs < int(selectionConfig.MaxTx)
func (txorches *txOrchestrator) isEnoughSpace(alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount int, l2tx common.PoolL2Tx) bool {
	if !canAddL2Tx(alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, txorches.selectionConfig) {
		// If tx is atomic, restart process without txs from the atomic group
		txSelErr := common.TxSelectorError{
			Message: ErrNoAvailableSlots,
			Code:    ErrNoAvailableSlotsCode,
			Type:    ErrNoAvailableSlotsType,
		}
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		// no more available slots for L2Txs, so mark this tx
		// but also the rest of remaining txs as discarded
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

// isBatchGreaterThanMaxNumBatch check if batch in tx is less than max batch num in config
func (txorches *txOrchestrator) isBatchGreaterThanMaxNumBatch(
	l2tx common.PoolL2Tx, nextBatchNum uint32) bool {
	if l2tx.MaxNumBatch != 0 && nextBatchNum > l2tx.MaxNumBatch {
		txSelErr := common.TxSelectorError{
			Message: ErrUnsupportedMaxNumBatch,
			Code:    ErrUnsupportedMaxNumBatchCode,
			Type:    ErrUnsupportedMaxNumBatchType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		// Tx won't be forjable since the current batch num won't go backwards
		txorches.unforjableL2TxsChan <- l2tx
		return false
	}
	return true
}

// isExitWithZeroAmount send to unforjed txs with exit type with zero amount
func (txorches *txOrchestrator) isExitWithZeroAmount(l2tx common.PoolL2Tx) bool {
	if l2tx.Type == common.TxTypeExit && l2tx.Amount.Cmp(big.NewInt(0)) <= 0 {
		// If tx is atomic, restart process without txs from the atomic group
		txSelErr := common.TxSelectorError{
			Message: ErrExitAmount,
			Code:    ErrExitAmountCode,
			Type:    ErrExitAmountType,
		}
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		// Although tecnicaly forjable, it won't never get forged with current code
		txorches.unforjableL2TxsChan <- l2tx
		return false
	}
	return true
}

// isNonceCorrect check that tx have the same nonce as accSender
func (txorches *txOrchestrator) isNonceCorrect(l2tx common.PoolL2Tx, accSender *common.Account) bool {
	if l2tx.Nonce != accSender.Nonce {
		txSelErr := common.TxSelectorError{
			Message: fmt.Sprintf(ErrNoCurrentNonce+"Tx.Nonce: %d, Account.Nonce: %d", l2tx.Nonce, accSender.Nonce),
			Code:    ErrNoCurrentNonceCode,
			Type:    ErrNoCurrentNonceType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		// not valid Nonce at tx. Discard L2Tx, and update Info
		// parameter of the tx, and add it to the discardedTxs
		// array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

// isEnoughBalanceOnSender checks for enough balance on sender account to process tx
func (txorches *txOrchestrator) isEnoughBalanceOnSender(l2tx common.PoolL2Tx) bool {
	// Check enough Balance on sender
	enoughBalance, balance, feeAndAmount := txorches.tp.CheckEnoughBalance(l2tx)
	if !enoughBalance {
		txSelErr := common.TxSelectorError{
			Message: fmt.Sprintf(ErrSenderNotEnoughBalance+"Current sender account Balance: %s, Amount+Fee: %s",
				balance.String(), feeAndAmount.String()),
			Code: ErrSenderNotEnoughBalanceCode,
			Type: ErrSenderNotEnoughBalanceType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		// not valid Amount with current Balance. Discard L2Tx,
		// and update Info parameter of the tx, and add it to
		// the discardedTxs array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

// isEnoughSpaceForL1CoordTx check is it enough space in tx-pool to process l1 tx
func (txorches *txOrchestrator) isEnoughSpaceForL1CoordTx(
	l2tx common.PoolL2Tx, alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount int) bool {
	if !canAddL2TxThatNeedsNewCoordL1Tx(
		alreadyProcessedL1TxsAmount,
		alreadyProcessedL2TxsAmount,
		txorches.selectionConfig,
	) {
		txSelErr := common.TxSelectorError{
			Message: ErrNotEnoughSpaceL1Coordinator,
			Code:    ErrNotEnoughSpaceL1CoordinatorCode,
			Type:    ErrNotEnoughSpaceL1CoordinatorType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		// discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

// verifyAndBuildForTxToEthAddrBJJ verifies and build transaction with specified toEthAddr or toBjj
func (txorches *txOrchestrator) verifyAndBuildForTxToEthAddrBJJ(
	l2tx common.PoolL2Tx,
	alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, positionL1 int,
) (*common.PoolL2Tx, *common.L1Tx, *common.AccountCreationAuth, bool) {
	validL2Tx, l1CoordinatorTx, accAuth, err := txorches.txsel.processTxToEthAddrBJJ(
		txorches.selectionConfig,
		txorches.l1UserFutureTxs,
		alreadyProcessedL1TxsAmount,
		alreadyProcessedL2TxsAmount,
		positionL1,
		l2tx,
	)
	if err != nil {
		txSelErr := common.TxSelectorError{
			Message: fmt.Sprintf(ErrTxDiscartedInProcessTxToEthAddrBJJ+" due to %s", err.Error()),
			Code:    ErrTxDiscartedInProcessTxToEthAddrBJJCode,
			Type:    ErrTxDiscartedInProcessTxToEthAddrBJJType,
		}
		log.Debugw("txsel.processTxToEthAddrBJJ", "err", err)
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return nil, nil, nil, false
		}
		// Discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.unforjableL2TxsChan <- l2tx
		return nil, nil, nil, false
	}
	return validL2Tx, l1CoordinatorTx, accAuth, true
}

// isToIdxExists checks if id of receiver account exists
func (txorches *txOrchestrator) isToIdxExists(l2tx common.PoolL2Tx) bool {
	_, err := txorches.txsel.localAccountsDB.GetAccount(l2tx.ToIdx)
	if err != nil {
		txSelErr := common.TxSelectorError{
			Message: fmt.Sprintf(ErrToIdxNotFound+"ToIdx: %d", l2tx.ToIdx),
			Code:    ErrToIdxNotFoundCode,
			Type:    ErrToIdxNotFoundType,
		}
		// tx not valid
		log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
			"ToIdx", l2tx.ToIdx)
		// If tx is atomic, restart process without txs from the atomic group
		if txorches.isTxAtomic(l2tx, txSelErr) {
			return false
		}
		// Discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = txSelErr.Message
		l2tx.ErrorCode = txSelErr.Code
		l2tx.ErrorType = txSelErr.Type
		txorches.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created
func (txsel *TxSelector) processTxToEthAddrBJJ(
	selectionConfig txprocessor.Config, l1UserFutureTxs []common.L1Tx,
	alreadyProcessedL1TxsAmount, alreadyProcessedL2TxsAmount, positionL1 int, l2Tx common.PoolL2Tx) (
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
					tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither "+
						"ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
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
	if !canAddL2TxThatNeedsNewCoordL1Tx(
		alreadyProcessedL1TxsAmount,
		alreadyProcessedL2TxsAmount,
		selectionConfig,
	) {
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
