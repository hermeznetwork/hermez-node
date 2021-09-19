package txselector

import (
	"sync"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
)

type Verifier struct {
	failedAGChan                                                 chan failedAtomicGroup
	errChan                                                      chan error
	unforjableL2TxsChan, nonSelectedL2TxsChan, selectedL2TxsChan chan common.PoolL2Tx
	txsL1ToBeProcessedChan                                       chan *common.L1Tx
	accAuthsChan                                                 chan []byte
	l2Txs                                                        []common.PoolL2Tx
	l1UserFutureTxs                                              []common.L1Tx
	nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1   int
	selectionConfig                                              txprocessor.Config
	tp                                                           *txprocessor.TxProcessor
	txsel                                                        *TxSelector
	l1TxsWg                                                      *sync.WaitGroup
	txsWg                                                        *sync.WaitGroup
}

func NewVerifier(
	failedAGChan chan failedAtomicGroup,
	errChan chan error,
	unforjableL2TxsChan, nonSelectedL2TxsChan, selectedL2TxsChan chan common.PoolL2Tx,
	txsL1ToBeProcessedChan chan *common.L1Tx,
	accAuthsChan chan []byte,
	l2txs []common.PoolL2Tx,
	l1UserFutureTxs []common.L1Tx,
	nAlreadyProcessedL1Txs, nAlreadyProcessedL2Txs, positionL1 int,
	selectionConfig txprocessor.Config,
	tp *txprocessor.TxProcessor,
	txsel *TxSelector,
	l1TxsWg, txsWg *sync.WaitGroup) *Verifier {
	return &Verifier{
		failedAGChan:           failedAGChan,
		errChan:                errChan,
		unforjableL2TxsChan:    unforjableL2TxsChan,
		nonSelectedL2TxsChan:   nonSelectedL2TxsChan,
		selectedL2TxsChan:      selectedL2TxsChan,
		txsL1ToBeProcessedChan: txsL1ToBeProcessedChan,
		accAuthsChan:           accAuthsChan,
		l2Txs:                  l2txs,
		l1UserFutureTxs:        l1UserFutureTxs,
		nAlreadyProcessedL1Txs: nAlreadyProcessedL1Txs,
		nAlreadyProcessedL2Txs: nAlreadyProcessedL2Txs,
		positionL1:             positionL1,
		selectionConfig:        selectionConfig,
		tp:                     tp,
		txsel:                  txsel,
		l1TxsWg:                l1TxsWg,
		txsWg:                  txsWg,
	}
}

func (v *Verifier) checkBatchGreaterThanMaxNumBatch(l2txsChan <-chan common.PoolL2Tx, nextBatchNum uint32) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			if !isBatchGreaterThanMaxNumBatch(v.failedAGChan, v.errChan, v.unforjableL2TxsChan, l2tx, nextBatchNum) {
				v.txsWg.Done()
				continue
			}
			out <- l2tx
		}
		close(out)
	}()
	return out
}

func (v *Verifier) checkIsExitWithZeroAmount(l2txsChan <-chan common.PoolL2Tx) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			if !isExitWithZeroAmount(v.failedAGChan, v.errChan, v.unforjableL2TxsChan, l2tx) {
				v.txsWg.Done()
				continue
			}
			out <- l2tx
		}
		close(out)
	}()
	return out
}

func (v *Verifier) verifyTxs(l2txsChan <-chan common.PoolL2Tx) <-chan common.PoolL2Tx {
	out := make(chan common.PoolL2Tx)
	go func() {
		for l2tx := range l2txsChan {
			var isTxCorrect bool
			isTxCorrect = isEnoughSpace(v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan, v.nAlreadyProcessedL1Txs, v.nAlreadyProcessedL2Txs, v.selectionConfig, l2tx)
			if !isTxCorrect {
				v.txsWg.Done()
				continue
			}

			// Check enough Balance on sender
			isTxCorrect = isEnoughBalanceOnSender(v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan, l2tx, v.tp)
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
			isTxCorrect = isNonceCorrect(v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan, l2tx, accSender)
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
			newL1CoordTx, v.positionL1, err = v.txsel.coordAccountForTokenID(accSender.TokenID, v.positionL1)
			if err != nil {
				v.errChan <- tracerr.Wrap(err)
				return
			}
			if newL1CoordTx != nil {
				// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
				// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
				isTxCorrect = isEnoughSpaceForL1CoordTx(v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan, l2tx, v.nAlreadyProcessedL1Txs, v.nAlreadyProcessedL2Txs, v.selectionConfig)
				if !isTxCorrect {
					v.txsWg.Done()
					continue
				}

				// increase positionL1
				v.positionL1++
				//l1CoordinatorTxsChan <- *newL1CoordTx
				v.accAuthsChan <- v.txsel.coordAccount.AccountCreationAuth

				// process the L1CoordTx
				v.l1TxsWg.Add(1)
				v.txsL1ToBeProcessedChan <- newL1CoordTx
				v.nAlreadyProcessedL1Txs++
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
				validL2Tx, l1CoordinatorTx, accAuth, isTxCorrect = verifyAndBuildForTxToEthAddrBJJ(
					v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan,
					l2tx, v.txsel, v.selectionConfig, v.l1UserFutureTxs,
					v.nAlreadyProcessedL1Txs, v.nAlreadyProcessedL2Txs, v.positionL1)

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
					v.positionL1++
					// TODO: PROCESS TX
					// process the L1CoordTx
					v.l1TxsWg.Add(1)
					v.txsL1ToBeProcessedChan <- l1CoordinatorTx
					v.nAlreadyProcessedL1Txs++
				}
				if validL2Tx == nil {
					// Missing info on why this tx is not selected? Check l2Txs.Info at this point!
					// If tx is atomic, restart process without txs from the atomic group
					obj := common.TxSelectorError{
						Message: l2tx.Info,
						Code:    l2tx.ErrorCode,
						Type:    l2tx.ErrorType,
					}

					if isTxAtomic(v.failedAGChan, v.errChan, l2tx, obj) {
						return
					}
					v.nonSelectedL2TxsChan <- l2tx
					v.txsWg.Done()
					continue
				}
			} else if l2tx.ToIdx >= common.IdxUserThreshold {
				isTxCorrect = isToIdxExists(v.failedAGChan, v.errChan, v.nonSelectedL2TxsChan, l2tx, v.txsel)
				if !isTxCorrect {
					v.txsWg.Done()
					continue
				}
			}

			if newL1CoordTx != nil || l1CoordinatorTx != nil {
				v.l1TxsWg.Wait()
			}
			out <- l2tx
			v.nAlreadyProcessedL2Txs++
		}
		close(out)
	}()
	return out
}

func (v *Verifier) processL2Txs(l2txsChan <-chan common.PoolL2Tx) {
	go func() {
		for l2tx := range l2txsChan {
			go func(l2tx common.PoolL2Tx) {
				tryToProcessL2Tx(l2tx, v.selectedL2TxsChan, v.nonSelectedL2TxsChan, v.failedAGChan, v.errChan, v.txsel, v.tp)
				v.txsWg.Done()
			}(l2tx)
		}
	}()

}
