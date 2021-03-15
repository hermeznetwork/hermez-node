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
	if checkAlreadyPendingToCreate(l1CoordinatorTxs, tokenID,
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
func (txsel *TxSelector) GetL2TxSelection(selectionConfig txprocessor.Config) ([]common.Idx,
	[][]byte, []common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metricGetL2TxSelection.Inc()
	coordIdxs, accCreationAuths, _, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, err := txsel.getL1L2TxSelection(selectionConfig, []common.L1Tx{})
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
	l1UserTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	metricGetL1L2TxSelection.Inc()
	coordIdxs, accCreationAuths, l1UserTxs, l1CoordinatorTxs, l2Txs,
		discardedL2Txs, err := txsel.getL1L2TxSelection(selectionConfig, l1UserTxs)
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
	l1UserTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, []common.PoolL2Tx, error) {
	// WIP.0: the TxSelector is not optimized and will need a redesign. The
	// current version is implemented in order to have a functional
	// implementation that can be used ASAP.

	// Steps of this method:
	// - getPendingTxs
	// - ProcessL1Txs
	// - getProfitable (sort by fee & nonce)
	// - loop over l2Txs
	//         - Fill tx.TokenID tx.Nonce
	//         - Check enough Balance on sender
	//         - Check Nonce
	//         - Create CoordAccount L1CoordTx for TokenID if needed
	//                 - & ProcessL1Tx of L1CoordTx
	//         - Check validity of receiver Account for ToEthAddr / ToBJJ
	//         - Create UserAccount L1CoordTx if needed (and possible)
	//         - If everything is fine, store l2Tx to validTxs & update NoncesMap
	// - Prepare coordIdxsMap & AccumulatedFees
	// - Distribute AccumulatedFees to CoordIdxs
	// - MakeCheckpoint

	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs()
	if err != nil {
		return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	txselStateDB := txsel.localAccountsDB.StateDB
	tp := txprocessor.NewTxProcessor(txselStateDB, selectionConfig)

	// Process L1UserTxs
	for i := 0; i < len(l1UserTxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1UserTxs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	var l1CoordinatorTxs []common.L1Tx
	positionL1 := len(l1UserTxs)
	var accAuths [][]byte

	// Sort l2TxsRaw (cropping at MaxTx at this point).
	// discardedL2Txs contains an array of the L2Txs that have not been
	// selected in this Batch.
	l2Txs, discardedL2Txs := txsel.getL2Profitable(l2TxsRaw, selectionConfig.MaxTx-uint32(len(l1UserTxs)))
	for i := range discardedL2Txs {
		discardedL2Txs[i].Info =
			"Tx not selected due to low absolute fee (does not fit inside the profitable set)"
	}

	var validTxs []common.PoolL2Tx
	tp.AccumulatedFees = make(map[common.Idx]*big.Int)
	// Iterate over l2Txs
	// - check Nonces
	// - check enough Balance for the Amount+Fee
	// - if needed, create new L1CoordinatorTxs for unexisting ToIdx
	// 	- keep used accAuths
	// - put the valid txs into validTxs array
	for i := 0; i < len(l2Txs); i++ {
		// Check if there is space for more L2Txs in the selection
		maxL2Txs := int(selectionConfig.MaxTx) -
			len(l1UserTxs) - len(l1CoordinatorTxs)
		if len(validTxs) >= maxL2Txs {
			// no more available slots for L2Txs
			l2Txs[i].Info =
				"Tx not selected due not available slots for L2Txs"
			discardedL2Txs = append(discardedL2Txs, l2Txs[i])
			continue
		}

		// get Nonce & TokenID from the Account by l2Tx.FromIdx
		accSender, err := tp.StateDB().GetAccount(l2Txs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
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
			return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		if newL1CoordTx != nil {
			// if there is no space for the L1CoordinatorTx as MaxL1Tx, or no space
			// for L1CoordinatorTx + L2Tx as MaxTx, discard the L2Tx
			if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1Tx)-len(l1UserTxs) ||
				len(l1CoordinatorTxs)+1 >= int(selectionConfig.MaxTx)-len(l1UserTxs) {
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
				return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
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
					len(l1UserTxs), l1CoordinatorTxs, positionL1, l2Txs[i])
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
			if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1Tx)-len(l1UserTxs) ||
				len(l1CoordinatorTxs)+1 >= int(selectionConfig.MaxTx)-len(l1UserTxs) {
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
					return nil, nil, nil, nil, nil, nil, tracerr.Wrap(err)
				}
			}
			if validL2Tx == nil {
				discardedL2Txs = append(discardedL2Txs, l2Txs[i])
				continue
			}
		} else if l2Txs[i].ToIdx >= common.IdxUserThreshold {
			receiverAcc, err := txsel.localAccountsDB.GetAccount(l2Txs[i].ToIdx)
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
			if l2Txs[i].ToEthAddr != common.EmptyAddr {
				if l2Txs[i].ToEthAddr != receiverAcc.EthAddr {
					log.Debugw("invalid L2Tx: ToEthAddr does not correspond to the Account.EthAddr",
						"ToIdx", l2Txs[i].ToIdx, "tx.ToEthAddr",
						l2Txs[i].ToEthAddr, "account.EthAddr", receiverAcc.EthAddr)
					// Discard L2Tx, and update Info
					// parameter of the tx, and add it to
					// the discardedTxs array
					l2Txs[i].Info = fmt.Sprintf("Tx not selected because ToEthAddr "+
						"does not correspond to the Account.EthAddr. "+
						"tx.ToIdx: %d, tx.ToEthAddr: %s, account.EthAddr: %s",
						l2Txs[i].ToIdx, l2Txs[i].ToEthAddr, receiverAcc.EthAddr)
					discardedL2Txs = append(discardedL2Txs, l2Txs[i])
					continue
				}
			}
			if l2Txs[i].ToBJJ != common.EmptyBJJComp {
				if l2Txs[i].ToBJJ != receiverAcc.BJJ {
					log.Debugw("invalid L2Tx: ToBJJ does not correspond to the Account.BJJ",
						"ToIdx", l2Txs[i].ToIdx, "tx.ToEthAddr", l2Txs[i].ToBJJ,
						"account.BJJ", receiverAcc.BJJ)
					// Discard L2Tx, and update Info
					// parameter of the tx, and add it to
					// the discardedTxs array
					l2Txs[i].Info = fmt.Sprintf("Tx not selected because tx.ToBJJ "+
						"does not correspond to the Account.BJJ. "+
						"tx.ToIdx: %d, tx.ToEthAddr: %s, tx.ToBJJ: %s, account.BJJ: %s",
						l2Txs[i].ToIdx, l2Txs[i].ToEthAddr, l2Txs[i].ToBJJ, receiverAcc.BJJ)
					discardedL2Txs = append(discardedL2Txs, l2Txs[i])
					continue
				}
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
			return nil, nil, nil, nil, nil, nil,
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

	metricSelectedL1CoordinatorTxs.Set(float64(len(l1CoordinatorTxs)))
	metricSelectedL1UserTxs.Set(float64(len(l1UserTxs)))
	metricSelectedL2Txs.Set(float64(len(validTxs)))
	metricDiscardedL2Txs.Set(float64(len(discardedL2Txs)))

	// return coordIdxs, accAuths, l1UserTxs, l1CoordinatorTxs, validTxs, discardedL2Txs, nil
	return coordIdxs, accAuths, l1UserTxs, l1CoordinatorTxs, validTxs, discardedL2Txs, nil
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created and added to the
// l1CoordinatorTxs array, and then the PoolL2Tx is added into the validTxs
// array.
func (txsel *TxSelector) processTxToEthAddrBJJ(validTxs []common.PoolL2Tx,
	selectionConfig txprocessor.Config, nL1UserTxs int, l1CoordinatorTxs []common.L1Tx,
	positionL1 int, l2Tx common.PoolL2Tx) (*common.PoolL2Tx, *common.L1Tx,
	*common.AccountCreationAuth, error) {
	// if L2Tx needs a new L1CoordinatorTx of CreateAccount type, and a
	// previous L2Tx in the current process already created a
	// L1CoordinatorTx of this type, in the DB there still seem that needs
	// to create a new L1CoordinatorTx, but as is already created, the tx
	// is valid
	if checkAlreadyPendingToCreate(l1CoordinatorTxs, l2Tx.TokenID, l2Tx.ToEthAddr, l2Tx.ToBJJ) {
		return &l2Tx, nil, nil, nil
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

func checkAlreadyPendingToCreate(l1CoordinatorTxs []common.L1Tx, tokenID common.TokenID,
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

// getL2Profitable returns the profitable selection of L2Txssorted by Nonce
func (txsel *TxSelector) getL2Profitable(l2Txs []common.PoolL2Tx, max uint32) ([]common.PoolL2Tx,
	[]common.PoolL2Tx) {
	// First sort by nonce so that txs from the same account are sorted so
	// that they could be applied in succession.
	sort.Slice(l2Txs, func(i, j int) bool {
		return l2Txs[i].Nonce < l2Txs[j].Nonce
	})
	// Sort by absolute fee with SliceStable, so that txs with same
	// AbsoluteFee are not rearranged and nonce order is kept in such case
	sort.SliceStable(l2Txs, func(i, j int) bool {
		return l2Txs[i].AbsoluteFee > l2Txs[j].AbsoluteFee
	})

	discardedL2Txs := []common.PoolL2Tx{}
	if len(l2Txs) > int(max) {
		discardedL2Txs = l2Txs[max:]
		l2Txs = l2Txs[:max]
	}

	// sort l2Txs by Nonce. This can be done in many different ways, what
	// is needed is to output the l2Txs where the Nonce of l2Txs for each
	// Account is sorted, but the l2Txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.Slice(l2Txs, func(i, j int) bool {
		return l2Txs[i].Nonce < l2Txs[j].Nonce
	})

	return l2Txs, discardedL2Txs
}
