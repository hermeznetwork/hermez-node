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
	AccountCreationAuth []byte // signature in byte array format
}

// SelectionConfig contains the parameters of configuration of the selection of
// transactions for the next batch
type SelectionConfig struct {
	// MaxL1UserTxs is the maximum L1-user-tx for a batch
	MaxL1UserTxs uint64

	// TxProcessorConfig contains the config for ProcessTxs
	TxProcessorConfig txprocessor.Config
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
	localAccountsDB, err := statedb.NewLocalStateDB(dbpath, 128,
		synchronizerStateDB, statedb.TypeTxSelector, 0) // without merkletree
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
func (txsel *TxSelector) Reset(batchNum common.BatchNum) error {
	err := txsel.localAccountsDB.Reset(batchNum, true)
	if err != nil {
		return tracerr.Wrap(err)
	}
	return nil
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
func (txsel *TxSelector) GetL2TxSelection(selectionConfig *SelectionConfig) ([]common.Idx,
	[][]byte, []common.L1Tx,
	[]common.PoolL2Tx, error) {
	coordIdxs, accCreationAuths, _, l1CoordinatorTxs, l2Txs, err :=
		txsel.GetL1L2TxSelection(selectionConfig, []common.L1Tx{})
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
	l1UserTxs []common.L1Tx) ([]common.Idx, [][]byte, []common.L1Tx,
	[]common.L1Tx, []common.PoolL2Tx, error) {
	// WIP.0: the TxSelector is not optimized and will need a redesign. The
	// current version is implemented in order to have a functional
	// implementation that can be used asap.
	//
	// WIP.1: this method uses a 'cherry-pick' of internal calls of the
	// StateDB, a refactor of the StateDB to reorganize it internally is
	// planned once the main functionallities are covered, with that
	// refactor the TxSelector will be updated also.

	// get pending l2-tx from tx-pool
	l2TxsRaw, err := txsel.l2db.GetPendingTxs()
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	txselStateDB := txsel.localAccountsDB.StateDB
	tp := txprocessor.NewTxProcessor(txselStateDB, selectionConfig.TxProcessorConfig)

	// Process L1UserTxs
	for i := 0; i < len(l1UserTxs); i++ {
		// assumption: l1usertx are sorted by L1Tx.Position
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1UserTxs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	var l1CoordinatorTxs []common.L1Tx
	positionL1 := len(l1UserTxs)
	var accAuths [][]byte

	// sort l2TxsRaw (cropping at MaxTx at this point)
	l2Txs0 := txsel.getL2Profitable(l2TxsRaw, selectionConfig.TxProcessorConfig.MaxTx)

	noncesMap := make(map[common.Idx]common.Nonce)
	var l2Txs []common.PoolL2Tx
	// iterate over l2Txs
	// - if tx.TokenID does not exist at CoordsIdxDB
	// 	- create new L1CoordinatorTx creating a CoordAccount, for
	// 	Coordinator to receive the fee of the new TokenID
	for i := 0; i < len(l2Txs0); i++ {
		accSender, err := tp.StateDB().GetAccount(l2Txs0[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		l2Txs0[i].TokenID = accSender.TokenID
		// populate the noncesMap used at the next iteration
		noncesMap[l2Txs0[i].FromIdx] = accSender.Nonce

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
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		if newL1CoordTx != nil {
			// if there is no space for the L1CoordinatorTx, discard the L2Tx
			if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1UserTxs)-len(l1UserTxs) {
				// discard L2Tx
				continue
			}
			// increase positionL1
			positionL1++
			l1CoordinatorTxs = append(l1CoordinatorTxs, *newL1CoordTx)
			accAuths = append(accAuths, txsel.coordAccount.AccountCreationAuth)
		}
		l2Txs = append(l2Txs, l2Txs0[i])
	}

	var validTxs []common.PoolL2Tx
	// iterate over l2TxsRaw
	// - check Nonces
	// - if needed, create new L1CoordinatorTxs for unexisting ToIdx
	// 	- keep used accAuths
	// - put the valid txs into validTxs array
	for i := 0; i < len(l2Txs); i++ {
		// check if Nonce is correct
		nonce := noncesMap[l2Txs[i].FromIdx]
		if l2Txs[i].Nonce == nonce {
			noncesMap[l2Txs[i].FromIdx]++
		} else {
			// not valid Nonce at tx
			continue
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
				log.Debug(err)
				continue
			}
			if accAuth != nil && l1CoordinatorTx != nil {
				accAuths = append(accAuths, accAuth.Signature)
				l1CoordinatorTxs = append(l1CoordinatorTxs, *l1CoordinatorTx)
				positionL1++
			}
			if validL2Tx != nil {
				validTxs = append(validTxs, *validL2Tx)
			}
		} else if l2Txs[i].ToIdx >= common.IdxUserThreshold {
			receiverAcc, err := txsel.localAccountsDB.GetAccount(l2Txs[i].ToIdx)
			if err != nil {
				// tx not valid
				log.Debugw("invalid L2Tx: ToIdx not found in StateDB",
					"ToIdx", l2Txs[i].ToIdx)
				continue
			}
			if l2Txs[i].ToEthAddr != common.EmptyAddr {
				if l2Txs[i].ToEthAddr != receiverAcc.EthAddr {
					log.Debugw("invalid L2Tx: ToEthAddr does not correspond to the Account.EthAddr",
						"ToIdx", l2Txs[i].ToIdx, "tx.ToEthAddr", l2Txs[i].ToEthAddr,
						"account.EthAddr", receiverAcc.EthAddr)
					continue
				}
			}
			if l2Txs[i].ToBJJ != common.EmptyBJJComp {
				if l2Txs[i].ToBJJ != receiverAcc.BJJ {
					log.Debugw("invalid L2Tx: ToBJJ does not correspond to the Account.BJJ",
						"ToIdx", l2Txs[i].ToIdx, "tx.ToEthAddr", l2Txs[i].ToBJJ,
						"account.BJJ", receiverAcc.BJJ)
					continue
				}
			}

			// Account found in the DB, include the l2Tx in the selection
			validTxs = append(validTxs, l2Txs[i])
		} else if l2Txs[i].ToIdx == common.Idx(1) {
			// valid txs (of Exit type)
			validTxs = append(validTxs, l2Txs[i])
		}
	}

	// Process L1CoordinatorTxs
	for i := 0; i < len(l1CoordinatorTxs); i++ {
		_, _, _, _, err := tp.ProcessL1Tx(nil, &l1CoordinatorTxs[i])
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
	}

	// get CoordIdxsMap for the TokenIDs
	coordIdxsMap := make(map[common.TokenID]common.Idx)
	for i := 0; i < len(l2Txs); i++ {
		// get TokenID from tx.Sender
		accSender, err := tp.StateDB().GetAccount(l2Txs[i].FromIdx)
		if err != nil {
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		tokenID := accSender.TokenID

		coordIdx, err := txsel.getCoordIdx(tokenID)
		if err != nil {
			// if err is db.ErrNotFound, should not happen, as all
			// the l2Txs.TokenID should have a CoordinatorIdx
			// created in the DB at this point
			return nil, nil, nil, nil, nil, tracerr.Wrap(err)
		}
		coordIdxsMap[tokenID] = coordIdx
	}

	var coordIdxs []common.Idx
	tp.AccumulatedFees = make(map[common.Idx]*big.Int)
	for _, idx := range coordIdxsMap {
		tp.AccumulatedFees[idx] = big.NewInt(0)
		coordIdxs = append(coordIdxs, idx)
	}
	// sort CoordIdxs
	sort.SliceStable(coordIdxs, func(i, j int) bool {
		return coordIdxs[i] < coordIdxs[j]
	})

	// get most profitable L2-tx
	maxL2Txs := int(selectionConfig.TxProcessorConfig.MaxTx) -
		len(l1UserTxs) - len(l1CoordinatorTxs)

	selectedL2Txs := l2Txs
	if len(l2Txs) > maxL2Txs {
		selectedL2Txs = selectedL2Txs[:maxL2Txs]
	}
	var finalL2Txs []common.PoolL2Tx
	for i := 0; i < len(selectedL2Txs); i++ {
		_, _, _, err = tp.ProcessL2Tx(coordIdxsMap, nil, nil, &selectedL2Txs[i])
		if err != nil {
			// the error can be due not valid tx data, or due other
			// cases (such as StateDB error). At this initial
			// version of the TxSelector, we discard the L2Tx and
			// log the error, assuming that this will be iterated
			// in a near future.
			log.Error(err)
			continue
		}
		finalL2Txs = append(finalL2Txs, selectedL2Txs[i])
	}

	// distribute the AccumulatedFees from the processed L2Txs into the
	// Coordinator Idxs
	for idx, accumulatedFee := range tp.AccumulatedFees {
		cmp := accumulatedFee.Cmp(big.NewInt(0))
		if cmp == 1 { // accumulatedFee>0
			// send the fee to the Idx of the Coordinator for the TokenID
			accCoord, err := txsel.localAccountsDB.GetAccount(idx)
			if err != nil {
				log.Errorw("Can not distribute accumulated fees to coordinator account: No coord Idx to receive fee", "idx", idx)
				return nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
			accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
			_, err = txsel.localAccountsDB.UpdateAccount(idx, accCoord)
			if err != nil {
				log.Error(err)
				return nil, nil, nil, nil, nil, tracerr.Wrap(err)
			}
		}
	}

	err = tp.StateDB().MakeCheckpoint()
	if err != nil {
		return nil, nil, nil, nil, nil, tracerr.Wrap(err)
	}

	return coordIdxs, accAuths, l1UserTxs, l1CoordinatorTxs, finalL2Txs, nil
}

// processTxsToEthAddrBJJ process the common.PoolL2Tx in the case where
// ToIdx==0, which can be the tx type of ToEthAddr or ToBJJ. If the receiver
// does not have an account yet, a new L1CoordinatorTx of type
// CreateAccountDeposit (with 0 as DepositAmount) is created and added to the
// l1CoordinatorTxs array, and then the PoolL2Tx is added into the validTxs
// array.
func (txsel *TxSelector) processTxToEthAddrBJJ(validTxs []common.PoolL2Tx,
	selectionConfig *SelectionConfig, nL1UserTxs int, l1CoordinatorTxs []common.L1Tx,
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
				return &l2Tx, nil, nil, nil
			}
			// if not, check if AccountCreationAuth exist for that
			// ToEthAddr
			accAuth, err = txsel.l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				// not found, l2Tx will not be added in the selection
				return nil, nil, nil, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
					l2Tx.ToIdx, l2Tx.ToEthAddr.Hex()))
			}
			if accAuth.BJJ != l2Tx.ToBJJ {
				// if AccountCreationAuth.BJJ is not the same
				// than in the tx, tx is not accepted
				return nil, nil, nil, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr & ToBJJ found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s, ToBJJ: %s",
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
				return nil, nil, nil, tracerr.Wrap(fmt.Errorf("invalid L2Tx: ToIdx not found in StateDB, neither ToEthAddr found in AccountCreationAuths L2DB. ToIdx: %d, ToEthAddr: %s",
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
	} else if bytes.Equal(l2Tx.ToEthAddr.Bytes(), common.FFAddr.Bytes()) &&
		l2Tx.ToBJJ != common.EmptyBJJComp {
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
	if len(l1CoordinatorTxs) >= int(selectionConfig.MaxL1UserTxs)-nL1UserTxs {
		// L2Tx discarded
		return nil, nil, nil, tracerr.Wrap(fmt.Errorf("L2Tx discarded due not slots for L1CoordinatorTx to create a new account for receiver of L2Tx"))
	}

	return &l2Tx, l1CoordinatorTx, accAuth, nil
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
func (txsel *TxSelector) getL2Profitable(l2Txs []common.PoolL2Tx, max uint32) []common.PoolL2Tx {
	sort.Sort(txs(l2Txs))
	if len(l2Txs) < int(max) {
		return l2Txs
	}
	l2Txs = l2Txs[:max]

	// sort l2Txs by Nonce. This can be done in many different ways, what
	// is needed is to output the l2Txs where the Nonce of l2Txs for each
	// Account is sorted, but the l2Txs can not be grouped by sender Account
	// neither by Fee. This is because later on the Nonces will need to be
	// sequential for the zkproof generation.
	sort.SliceStable(l2Txs, func(i, j int) bool {
		return l2Txs[i].Nonce < l2Txs[j].Nonce
	})

	return l2Txs
}
