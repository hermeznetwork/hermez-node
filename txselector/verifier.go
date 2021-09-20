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

type verifier struct {
	failedAGChan                                                 chan failedAtomicGroup
	errChan                                                      chan error
	unforjableL2TxsChan, nonSelectedL2TxsChan, selectedL2TxsChan chan common.PoolL2Tx
	txsL1ToBeProcessedChan                                       chan *common.L1Tx
	accAuthsChan                                                 chan []byte
	verificationEnded                                            chan int
	l1UserFutureTxs                                              []common.L1Tx
	selectionConfig                                              txprocessor.Config
	tp                                                           *txprocessor.TxProcessor
	txsel                                                        *TxSelector
	l1TxsWg                                                      *sync.WaitGroup
	txsWg                                                        *sync.WaitGroup
}

func newVerifier(
	l1UserFutureTxs []common.L1Tx,
	selectionConfig txprocessor.Config,
	tp *txprocessor.TxProcessor,
	txsel *TxSelector) *verifier {
	return &verifier{
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

func (v *verifier) transformTxs(l2txs []common.PoolL2Tx) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for _, tx := range l2txs {
			out <- tx
		}
		close(out)
	}()
	return out
}

func (v *verifier) checkBatchGreaterThanMaxNumBatch(l2txsChan <-chan common.PoolL2Tx, nextBatchNum uint32) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			if !v.isBatchGreaterThanMaxNumBatch(l2tx, nextBatchNum) {
				v.txsWg.Done()
				continue
			}
			out <- l2tx
		}
		close(out)
	}()
	return out
}

func (v *verifier) checkIsExitWithZeroAmount(l2txsChan <-chan common.PoolL2Tx) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			if !v.isExitWithZeroAmount(l2tx) {
				v.txsWg.Done()
				continue
			}
			out <- l2tx
		}
		close(out)
	}()
	return out
}

func (v *verifier) verifyTxs(l2txsChan <-chan common.PoolL2Tx, nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1 int) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			var isTxCorrect bool
			isTxCorrect = v.isEnoughSpace(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, l2tx)
			if !isTxCorrect {
				v.txsWg.Done()
				continue
			}

			// Check enough Balance on sender
			isTxCorrect = v.isEnoughBalanceOnSender(l2tx)
			if !isTxCorrect {
				v.txsWg.Done()
				continue
			}

			// get Nonce & TokenID from the Account by l2Tx.FromIdx
			accSender, err := v.tp.StateDB().GetAccount(l2tx.FromIdx)
			if err != nil {
				v.errChan <- tracerr.Wrap(err)
				return
			}
			l2tx.TokenID = accSender.TokenID

			// Check if Nonce is correct
			isTxCorrect = v.isNonceCorrect(l2tx, accSender)
			if !isTxCorrect {
				v.txsWg.Done()
				continue
			}

			// if TokenID does not exist yet, create new L1CoordinatorTx to
			// create the CoordinatorAccount for that TokenID, to receive
			// the fees. Only in the case that there does not exist yet a
			// pending L1CoordinatorTx to create the account for the
			// Coordinator for that TokenID
			var newL1CoordTx *common.L1Tx
			newL1CoordTx, positionL1, err = v.txsel.coordAccountForTokenID(accSender.TokenID, positionL1)
			if err != nil {
				v.errChan <- tracerr.Wrap(err)
				return
			}
			if newL1CoordTx != nil {
				// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
				// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
				isTxCorrect = v.isEnoughSpaceForL1CoordTx(l2tx, nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs)
				if !isTxCorrect {
					v.txsWg.Done()
					continue
				}

				// increase positionL1
				positionL1++
				//l1CoordinatorTxsChan <- *newL1CoordTx
				v.accAuthsChan <- v.txsel.coordAccount.AccountCreationAuth

				// process the L1CoordTx
				v.l1TxsWg.Add(1)
				v.txsL1ToBeProcessedChan <- newL1CoordTx
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
			var (
				validL2Tx       *common.PoolL2Tx
				l1CoordinatorTx *common.L1Tx
				accAuth         *common.AccountCreationAuth
			)
			if l2tx.ToIdx == 0 { // ToEthAddr/ToBJJ case
				validL2Tx, l1CoordinatorTx, accAuth, isTxCorrect = v.verifyAndBuildForTxToEthAddrBJJ(
					l2tx, nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1)

				if !isTxCorrect {
					v.txsWg.Done()
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
						v.accAuthsChan <- common.EmptyEthSignature
					} else if accAuth != nil {
						v.accAuthsChan <- accAuth.Signature
					}
					positionL1++
					// TODO: PROCESS TX
					// process the L1CoordTx
					v.l1TxsWg.Add(1)
					v.txsL1ToBeProcessedChan <- l1CoordinatorTx
					nAlreadyProcessedL1Txs++
				}
				if validL2Tx == nil {
					// Missing info on why this tx is not selected? Check l2Txs.Info at this point!
					// If tx is atomic, restart process without txs from the atomic group
					obj := common.TxSelectorError{
						Message: l2tx.Info,
						Code:    l2tx.ErrorCode,
						Type:    l2tx.ErrorType,
					}

					if v.isTxAtomic(l2tx, obj) {
						return
					}
					v.nonSelectedL2TxsChan <- l2tx
					v.txsWg.Done()
					continue
				}
			} else if l2tx.ToIdx >= common.IdxUserThreshold {
				isTxCorrect = v.isToIdxExists(l2tx)
				if !isTxCorrect {
					v.txsWg.Done()
					continue
				}
			}

			if newL1CoordTx != nil || l1CoordinatorTx != nil {
				v.l1TxsWg.Wait()
			}
			out <- l2tx
			nAlreadyProcessedL2Txs++
		}
		close(out)
	}()
	return out
}

func (v *verifier) processL2Txs(l2txsChan <-chan common.PoolL2Tx) {
	go func() {
		for l2tx := range l2txsChan {
			v.tryToProcessL2Tx(l2tx)
			v.txsWg.Done()
		}
	}()
}

func (v *verifier) tryToProcessL2Tx(
	l2tx common.PoolL2Tx) {
	// get CoordIdxsMap for the TokenID of the current l2Txs[i]
	// get TokenID from tx.Sender account
	tokenID := l2tx.TokenID
	coordIdx, err := v.txsel.getCoordIdx(tokenID)
	if err != nil {
		// if err is db.ErrNotFound, should not happen, as all
		// the selectedTxs.TokenID should have a CoordinatorIdx
		// created in the DB at this point
		v.errChan <- tracerr.Wrap(fmt.Errorf("could not get CoordIdx for TokenID=%d, "+
			"due: %s", tokenID, err))
		return
	}
	// prepare temp coordIdxsMap & AccumulatedFees for the call to
	// ProcessL2Tx
	coordIdxsMap := map[common.TokenID]common.Idx{tokenID: coordIdx}
	// tp.AccumulatedFees = make(map[common.Idx]*big.Int)
	if _, ok := v.tp.AccumulatedFees[coordIdx]; !ok {
		v.tp.AccumulatedFees[coordIdx] = big.NewInt(0)
	}

	// TODO: PROCESS TX
	_, _, _, err = v.tp.ProcessL2Tx(coordIdxsMap, nil, nil, &l2tx)
	if err != nil {
		obj := common.TxSelectorError{
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
				reason:     obj,
			}
			v.failedAGChan <- failedAG
			return
		}
		// Discard L2Tx, and update Info parameter of the tx,
		// and add it to the discardedTxs array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
		v.errChan <- tracerr.Wrap(err)
		return
	}
	v.selectedL2TxsChan <- l2tx
}

func (v *verifier) isTxAtomic(l2tx common.PoolL2Tx, reason common.TxSelectorError) bool {
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

		v.failedAGChan <- failedAG
		v.errChan <- err

		return true
	}
	return false
}

func (v *verifier) isEnoughSpace(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs int, l2tx common.PoolL2Tx) bool {
	if !canAddL2Tx(nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, v.selectionConfig) {
		// If tx is atomic, restart process without txs from the atomic group
		obj := common.TxSelectorError{
			Message: ErrNoAvailableSlots,
			Code:    ErrNoAvailableSlotsCode,
			Type:    ErrNoAvailableSlotsType,
		}
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		// no more available slots for L2Txs, so mark this tx
		// but also the rest of remaining txs as discarded
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) isBatchGreaterThanMaxNumBatch(
	l2tx common.PoolL2Tx, nextBatchNum uint32) bool {
	if l2tx.MaxNumBatch != 0 && nextBatchNum > l2tx.MaxNumBatch {
		obj := common.TxSelectorError{
			Message: ErrUnsupportedMaxNumBatch,
			Code:    ErrUnsupportedMaxNumBatchCode,
			Type:    ErrUnsupportedMaxNumBatchType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		// Tx won't be forjable since the current batch num won't go backwards
		v.unforjableL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) isExitWithZeroAmount(l2tx common.PoolL2Tx) bool {
	if l2tx.Type == common.TxTypeExit && l2tx.Amount.Cmp(big.NewInt(0)) <= 0 {
		// If tx is atomic, restart process without txs from the atomic group
		obj := common.TxSelectorError{
			Message: ErrExitAmount,
			Code:    ErrExitAmountCode,
			Type:    ErrExitAmountType,
		}
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		// Although tecnicaly forjable, it won't never get forged with current code
		v.unforjableL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) isNonceCorrect(l2tx common.PoolL2Tx, accSender *common.Account) bool {
	if l2tx.Nonce != accSender.Nonce {
		obj := common.TxSelectorError{
			Message: fmt.Sprintf(ErrNoCurrentNonce+"Tx.Nonce: %d, Account.Nonce: %d", l2tx.Nonce, accSender.Nonce),
			Code:    ErrNoCurrentNonceCode,
			Type:    ErrNoCurrentNonceType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		// not valid Nonce at tx. Discard L2Tx, and update Info
		// parameter of the tx, and add it to the discardedTxs
		// array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) isEnoughBalanceOnSender(l2tx common.PoolL2Tx) bool {
	// Check enough Balance on sender
	enoughBalance, balance, feeAndAmount := v.tp.CheckEnoughBalance(l2tx)
	if !enoughBalance {
		obj := common.TxSelectorError{
			Message: fmt.Sprintf(ErrSenderNotEnoughBalance+"Current sender account Balance: %s, Amount+Fee: %s",
				balance.String(), feeAndAmount.String()),
			Code: ErrSenderNotEnoughBalanceCode,
			Type: ErrSenderNotEnoughBalanceType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		// not valid Amount with current Balance. Discard L2Tx,
		// and update Info parameter of the tx, and add it to
		// the discardedTxs array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) isEnoughSpaceForL1CoordTx(
	l2tx common.PoolL2Tx, nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs int) bool {
	if !canAddL2TxThatNeedsNewCoordL1Tx(
		nAlreadyProcessedL1Txs,
		nAlreadyProcessedL2Txs,
		v.selectionConfig,
	) {
		obj := common.TxSelectorError{
			Message: ErrNotEnoughSpaceL1Coordinator,
			Code:    ErrNotEnoughSpaceL1CoordinatorCode,
			Type:    ErrNotEnoughSpaceL1CoordinatorType,
		}
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		// discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
		return false
	}
	return true
}

func (v *verifier) verifyAndBuildForTxToEthAddrBJJ(
	l2tx common.PoolL2Tx,
	nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1 int,
) (*common.PoolL2Tx, *common.L1Tx, *common.AccountCreationAuth, bool) {
	validL2Tx, l1CoordinatorTx, accAuth, err := v.txsel.processTxToEthAddrBJJ(
		v.selectionConfig,
		v.l1UserFutureTxs,
		nAlreadyProcessedL1Txs,
		nAlreadyProcessedL2Txs,
		positionL1,
		l2tx,
	)
	if err != nil {
		obj := common.TxSelectorError{
			Message: fmt.Sprintf(ErrTxDiscartedInProcessTxToEthAddrBJJ+" due to %s", err.Error()),
			Code:    ErrTxDiscartedInProcessTxToEthAddrBJJCode,
			Type:    ErrTxDiscartedInProcessTxToEthAddrBJJType,
		}
		log.Debugw("txsel.processTxToEthAddrBJJ", "err", err)
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return nil, nil, nil, false
		}
		// Discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.unforjableL2TxsChan <- l2tx
		return nil, nil, nil, false
	}
	return validL2Tx, l1CoordinatorTx, accAuth, true
}

func (v *verifier) isToIdxExists(l2tx common.PoolL2Tx) bool {
	_, err := v.txsel.localAccountsDB.GetAccount(l2tx.ToIdx)
	if err != nil {
		obj := common.TxSelectorError{
			Message: fmt.Sprintf(ErrToIdxNotFound+"ToIdx: %d", l2tx.ToIdx),
			Code:    ErrToIdxNotFoundCode,
			Type:    ErrToIdxNotFoundType,
		}
		// tx not valid
		log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
			"ToIdx", l2tx.ToIdx)
		// If tx is atomic, restart process without txs from the atomic group
		if v.isTxAtomic(l2tx, obj) {
			return false
		}
		// Discard L2Tx, and update Info parameter of
		// the tx, and add it to the discardedTxs array
		l2tx.Info = obj.Message
		l2tx.ErrorCode = obj.Code
		l2tx.ErrorType = obj.Type
		v.nonSelectedL2TxsChan <- l2tx
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
		nAlreadyProcessedL1Txs,
		nAlreadyProcessedL2Txs,
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
