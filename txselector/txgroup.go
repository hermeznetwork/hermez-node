package txselector

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type (
	// TxGroup represents a group of transactions. It can be grouped by the
	// same destination address or by an atomic transaction group
	TxGroup struct {
		l1UserTxs     []common.L1Tx
		l1CoordTxs    []common.L1Tx
		l2Txs         []common.PoolL2Tx
		discardedTxs  []common.PoolL2Tx
		coordIdxsMap  map[common.TokenID]common.Idx
		accAuths      [][]byte
		feeAverage    *big.Float
		atomic        bool
		firstPosition int
		coordAccount  CoordAccount
	}
)

// NewTxGroup creates a new *TxGroup object
func NewTxGroup(atomic bool, poolTxs []common.PoolL2Tx, processor txProcessor, l2db l2DB, localAccountsDB stateDB,
	firstPosition int, coordAccount CoordAccount, l1txs []common.L1Tx) (*TxGroup, error) {
	txGroup := &TxGroup{
		l1UserTxs:     nil,
		l1CoordTxs:    nil,
		l2Txs:         nil,
		discardedTxs:  nil,
		coordIdxsMap:  map[common.TokenID]common.Idx{},
		accAuths:      nil,
		atomic:        false,
		firstPosition: firstPosition,
		coordAccount:  coordAccount,
	}
	return txGroup, txGroup.addPoolTxs(atomic, poolTxs, processor, l2db, localAccountsDB, l1txs)
}

// calcFeeAverage calculate the fee average of all L2 tx fees in USD (sum of all L2 fees / number of L2 txs)
func (g *TxGroup) calcFeeAverage() {
	txLength := float64(g.l2Length())
	feeSum := g.feeSum()
	f, _ := feeSum.Float64()
	if f == 0 || txLength == 0 {
		g.feeAverage = big.NewFloat(0)
		return
	}
	g.feeAverage = new(big.Float).Quo(feeSum, big.NewFloat(txLength))
}

// feeSum returns the sum of all L2 tx fees in USD
func (g *TxGroup) feeSum() *big.Float {
	feeSum := new(big.Float)
	for _, tx := range g.l2Txs {
		feeSum = feeSum.Add(feeSum, big.NewFloat(tx.AbsoluteFee))
	}
	return feeSum
}

// L2Length returns the L2 transaction count
func (g *TxGroup) l2Length() int {
	return len(g.l2Txs)
}

// l1Length returns the L1 transaction count
func (g *TxGroup) l1Length() int {
	return len(g.l1UserTxs) + len(g.l1CoordTxs)
}

// length returns all transactions count (L1/L2)
func (g *TxGroup) length() int {
	return g.l1Length() + g.l2Length()
}

// isEmpty check if don't have transactions
func (g *TxGroup) isEmpty() bool {
	return g.length() == 0
}

// sort sort all transactions by the most profitable
func (g *TxGroup) sort() {
	// Sort by absolute fee with SliceStable, so that txs with the same
	// AbsoluteFee are not rearranged, and nonce order is kept in such a case
	sort.SliceStable(g.l2Txs, func(i, j int) bool {
		return g.l2Txs[i].AbsoluteFee > g.l2Txs[j].AbsoluteFee
	})

	// atomic groups must be sorted only by fee
	if g.atomic {
		return
	}

	// sort l2Txs by Nonce. This is because later on, the nonces will need
	// to be sequential for the ZkProof generation.
	sort.Slice(g.l2Txs, func(i, j int) bool {
		return g.l2Txs[i].Nonce < g.l2Txs[j].Nonce
	})
}

// hashAccount creates a unique sha256 hash from the tokenID and Ethereum and Baby Jubjub addresses
func hashAccount(addr ethCommon.Address, bjj babyjub.PublicKeyComp, tokenID common.TokenID) string {
	h := sha256.New()
	_, _ = h.Write([]byte(bjj.String()))
	_, _ = h.Write(addr.Bytes())
	hash := h.Sum(tokenID.Bytes())
	return hex.EncodeToString(hash[:])
}

// createL1Txs build all L1 transactions it should be generated to fulfilling the protocol
// requirements for the L2 transactions
func (g *TxGroup) createL1Txs(processor txProcessor, l2db l2DB, localAccountsDB stateDB, l1Txs []common.L1Tx) error {
	// reset all L1 transactions
	g.l1UserTxs = make([]common.L1Tx, 0)
	g.l1CoordTxs = make([]common.L1Tx, 0)
	g.accAuths = make([][]byte, 0)
	l1TxsCheck := make(map[string]bool)

	// create a hash of each L1 transaction that already exist
	for _, l1Tx := range l1Txs {
		accHash := hashAccount(l1Tx.FromEthAddr, l1Tx.FromBJJ, l1Tx.TokenID)
		l1TxsCheck[accHash] = true
	}

	// create L1 transactions based in the L2
	for _, l2Tx := range g.l2Txs {
		// generate the unique hash from the coordinator account and check if this tx already exist
		accHash := hashAccount(g.coordAccount.Addr, g.coordAccount.BJJ, l2Tx.TokenID)
		_, ok := l1TxsCheck[accHash]
		_, err := localAccountsDB.GetIdxByEthAddrBJJ(g.coordAccount.Addr, g.coordAccount.BJJ, l2Tx.TokenID)
		// if the coordinator account is not found in the L1 map or in the database, create one
		unwrappedErr := tracerr.Unwrap(err)
		if !ok && unwrappedErr == statedb.ErrIdxNotFound {
			g.l1CoordTxs = append(g.l1CoordTxs, common.L1Tx{
				UserOrigin:    false,
				FromEthAddr:   g.coordAccount.Addr,
				FromBJJ:       g.coordAccount.BJJ,
				TokenID:       l2Tx.TokenID,
				Amount:        big.NewInt(0),
				DepositAmount: big.NewInt(0),
				Type:          common.TxTypeCreateAccountDeposit,
			})
			g.accAuths = append(g.accAuths, g.coordAccount.AccountCreationAuth)
			l1TxsCheck[accHash] = true
			log.Debugw("TxSelector: new coordinator L1 tx CreateAccountDeposit",
				"Addr", g.coordAccount.Addr.String(),
				"BJJ", g.coordAccount.BJJ.String(),
				"TokenID", l2Tx.TokenID,
			)
		} else if err != nil && unwrappedErr != statedb.ErrIdxNotFound {
			return tracerr.Wrap(err)
		}

		// generate the unique hash from the destination account and check if this tx already exist
		accHash = hashAccount(l2Tx.ToEthAddr, l2Tx.ToBJJ, l2Tx.TokenID)
		_, ok = l1TxsCheck[accHash]
		if ok {
			continue
		}
		l1TxsCheck[accHash] = true

		// create all L1 transactions for a TransferToEthAddr
		if l2Tx.ToIdx == 0 && l2Tx.ToEthAddr != common.EmptyAddr && l2Tx.ToEthAddr != common.FFAddr {
			if _, err := localAccountsDB.GetIdxByEthAddr(l2Tx.ToEthAddr, l2Tx.TokenID); err == nil {
				continue
			}

			// check if this account already exists in the node database
			accAuth, err := l2db.GetAccountCreationAuth(l2Tx.ToEthAddr)
			if err != nil {
				continue
			}
			// create L1CoordinatorTx for the accountCreation
			g.l1UserTxs = append(g.l1UserTxs, common.L1Tx{
				UserOrigin:    false,
				FromEthAddr:   accAuth.EthAddr,
				FromBJJ:       accAuth.BJJ,
				TokenID:       l2Tx.TokenID,
				Amount:        big.NewInt(0),
				DepositAmount: big.NewInt(0),
				Type:          common.TxTypeCreateAccountDeposit,
			})
			g.accAuths = append(g.accAuths, accAuth.Signature)
			log.Debugw("TxSelector: new user L1 TransferToEthAddr tx CreateAccountDeposit",
				"Addr", accAuth.EthAddr.String(),
				"BJJ", accAuth.BJJ.String(),
				"TokenID", l2Tx.TokenID,
			)
			continue
		}

		// create all L1 transactions for a TransferToBJJ
		if l2Tx.ToIdx == 0 && l2Tx.ToEthAddr == common.FFAddr && l2Tx.ToBJJ != common.EmptyBJJComp {
			if _, err := localAccountsDB.GetIdxByEthAddrBJJ(l2Tx.ToEthAddr, l2Tx.ToBJJ, l2Tx.TokenID); err == nil {
				continue
			}
			g.l1UserTxs = append(g.l1UserTxs, common.L1Tx{
				UserOrigin:    false,
				FromEthAddr:   l2Tx.ToEthAddr,
				FromBJJ:       l2Tx.ToBJJ,
				TokenID:       l2Tx.TokenID,
				Amount:        big.NewInt(0),
				DepositAmount: big.NewInt(0),
				Type:          common.TxTypeCreateAccountDeposit,
			})
			g.accAuths = append(g.accAuths, common.EmptyEthSignature)
			log.Debugw("TxSelector: new user L1 TransferToBJJ tx CreateAccountDeposit",
				"Addr", l2Tx.ToEthAddr.String(),
				"BJJ", l2Tx.ToBJJ.String(),
				"TokenID", l2Tx.TokenID,
			)
			continue
		}
	}
	g.addL1Positions()

	// process the L1 transactions
	allL1Txs := append(g.l1CoordTxs, g.l1UserTxs...)
	for _, tx := range allL1Txs {
		log.Debugw("TxSelector: processing new L1 tx", "TxID", tx.TxID.String())
		_, _, _, _, err := processor.ProcessL1Tx(nil, &tx) //nolint:gosec
		if err != nil {
			return tracerr.Wrap(err)
		}
	}
	return nil
}

// addPoolTxs add pool transactions to the tx group
func (g *TxGroup) addPoolTxs(atomic bool, poolTxs []common.PoolL2Tx, processor txProcessor, l2db l2DB,
	localAccountsDB stateDB, l1txs []common.L1Tx) error {
	g.atomic = atomic
	g.l2Txs = poolTxs
	if g.isEmpty() {
		return tracerr.Wrap(fmt.Errorf("empty txs"))
	}
	g.calcFeeAverage()
	g.sort()
	coordIdxsMap, err := g.validate(processor, l2db, localAccountsDB, l1txs)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// TODO: why is sort called twice?
	g.sort()
	// After sorting set RqOffset for atomic txs
	if g.atomic {
		for i := 0; i < len(g.l2Txs); i++ {
			for j := 0; j < len(g.l2Txs); j++ {
				if g.l2Txs[i].RqTxID == g.l2Txs[j].TxID {
					// Tx i is requesting tx j
					rqOffset, err := RelativePositionToRqOffset(j - i)
					if err != nil {
						return tracerr.Wrap(err)
					}
					g.l2Txs[i].RqOffset = rqOffset
					break
				}
			}
			if g.l2Txs[i].RqOffset == 0 {
				return tracerr.New(ErrUnexpectedRqOffset)
			}
		}
	}
	return g.distributeFee(coordIdxsMap, processor, localAccountsDB)
}

// RelativePositionToRqOffset transforms a natura relative position to the expected RqOffset format,
// as decribed here: https://docs.hermez.io/#/developers/protocol/hermez-protocol/circuits/circuits?id=rq-tx-verifier
func RelativePositionToRqOffset(relativePosition int) (uint8, error) {
	switch relativePosition {
	case -4:
		return 4, nil
	case -3:
		return 5, nil
	case -2:
		return 6, nil
	case -1:
		return 7, nil
	case 0:
		return 0, nil
	case 1:
		return 1, nil
	case 2:
		return 2, nil
	case 3:
		return 3, nil
	default:
		return 0, tracerr.New(txprocessor.ErrInvalidRqOffset)

	}
}

// addL1Positions add batch positions into the L1 transactions
func (g *TxGroup) addL1Positions() {
	if g.isEmpty() {
		return
	}
	nextPosition := g.firstPosition
	// add coordinator transactions first
	for i := range g.l1CoordTxs {
		g.l1CoordTxs[i].Position = nextPosition
		nextPosition++
	}
	for i := range g.l1UserTxs {
		g.l1UserTxs[i].Position = nextPosition
		nextPosition++
	}
}

// validate validates all transactions into a batch, checking balances, fees, nonce, token id, and signatures
func (g *TxGroup) validate(processor txProcessor, l2db l2DB, localAccountsDB stateDB,
	l1txs []common.L1Tx) (map[common.TokenID]common.Idx, error) {
	var (
		txs                     = g.l2Txs[:]
		insufficientFundsTxsLen = 0
	)

	// create all L1 transactions from valid L2 transactions
	if err := g.createL1Txs(processor, l2db, localAccountsDB, l1txs); err != nil {
		return nil, err
	}

	g.coordIdxsMap = make(map[common.TokenID]common.Idx)
	g.l2Txs = make([]common.PoolL2Tx, 0)
	g.discardedTxs = make([]common.PoolL2Tx, 0)

	// Use this GOTO statement for we can check if more transactions can fit into a block.
	// If a user without funds received the fund in the same batch, we could forge both transactions.
	// The revalidate works only for atomic transactions group.
REVALIDATE:
	insufficientFundsTxs := make([]common.PoolL2Tx, 0)
	i := 0
	for _, tx := range txs {
		switch tx.Type {
		case common.TxTypeExit:
			if tx.ToIdx != 1 {
				tx.Info = ErrInvalidExitToIdx
				log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID, "ToIdx", tx.ToIdx)
				g.discardedTxs = append(g.discardedTxs, tx)
				continue
			}
			if tx.Amount.Cmp(big.NewInt(0)) <= 0 {
				tx.Info = ErrExitZeroAmount
				log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
				g.discardedTxs = append(g.discardedTxs, tx)
				continue
			}
		case common.TxTypeTransfer:
			// check if the destination Idx exist
			_, err := localAccountsDB.GetAccount(tx.ToIdx)
			if err != nil {
				tx.Info = ErrRecipientNotFound
				log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID, "ToIdx", tx.ToIdx)
				g.discardedTxs = append(g.discardedTxs, tx)
				continue
			}
		case common.TxTypeTransferToEthAddr:
			if err := validateTransferToEthAddr(tx, l2db, localAccountsDB); err != nil {
				tx.Info = err.Error()
				log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
				g.discardedTxs = append(g.discardedTxs, tx)
				continue
			}
		case common.TxTypeTransferToBJJ:
			if err := validateTransferToBJJ(tx, localAccountsDB); err != nil {
				tx.Info = err.Error()
				log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
				g.discardedTxs = append(g.discardedTxs, tx)
				continue
			}
		default:
			return nil, tracerr.Wrap(fmt.Errorf("invalid tx (%s) type %s", tx.TxID.String(), tx.Type))
		}

		// check the balance and the nonce
		if err := checkBalanceAndNonce(tx, localAccountsDB); err != nil {
			tx.Info = err.Error()
			log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
			if tx.Info == ErrInsufficientFunds {
				insufficientFundsTxs = append(insufficientFundsTxs, tx)
				continue
			}
			g.discardedTxs = append(g.discardedTxs, tx)
			continue
		}

		// get coordinator idx to collect fees. If not exist, wait to next batch to forge,
		// after the L1 coordinators txs are forged
		coordIdx, err := localAccountsDB.GetIdxByEthAddrBJJ(g.coordAccount.Addr, g.coordAccount.BJJ, tx.TokenID)
		if err != nil {
			tx.Info = ErrCoordIdxNotFound
			log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
			g.discardedTxs = append(g.discardedTxs, tx)
			continue
		}
		g.coordIdxsMap[tx.TokenID] = coordIdx

		// process L2 transactions
		_, _, _, err = processor.ProcessL2Tx(g.coordIdxsMap, nil, nil, &tx) //nolint:gosec
		if err != nil {
			// Discard L2Tx, and update Info parameter of the tx,
			// and add it to the discardedTxs array
			tx.Info = fmt.Sprintf("Tx not selected (in ProcessL2Tx) due to %s", err.Error())
			log.Debugw("TxSelector: "+tx.Info, "TxID", tx.TxID)
			g.discardedTxs = append(g.discardedTxs, tx)
			continue
		}

		// add again into the slice only the valid transactions
		tx.Info = ""
		txs[i] = tx
		i++
	}
	// remove the garbage transactions
	g.l2Txs = append(g.l2Txs, txs[:i]...)

	// if there are some atomic transaction with insufficient funds try again after process
	// the all valid transactions.
	if g.atomic && len(insufficientFundsTxs) > 0 && len(insufficientFundsTxs) != insufficientFundsTxsLen {
		insufficientFundsTxsLen = len(insufficientFundsTxs)
		txs = insufficientFundsTxs[:]
		// if there are transaction if insufficient funds, try again to check if the wallet has a balance now
		log.Debugw("TxSelector: trying to validate insufficient funds transaction again",
			"insufficientFundsTxs", insufficientFundsTxsLen)
		goto REVALIDATE
	}
	// add the insufficient funds txs array to the discarded txs array
	g.discardedTxs = append(g.discardedTxs, insufficientFundsTxs...)
	insufficientFundsTxs = nil
	txs = nil

	// if the group is atomic, discard all transactions if one of them fails
	if g.atomic && len(g.discardedTxs) > 0 {
		for _, tx := range g.l2Txs {
			tx.Info = ErrAtomicGroupFail
			g.discardedTxs = append(g.discardedTxs, tx)
		}
		log.Debugw("TxSelector: falling all atomic transaction from the group",
			"insufficientFundsTxs", len(g.discardedTxs))
		g.l2Txs = make([]common.PoolL2Tx, 0)
	}
	return g.coordIdxsMap, nil
}

// distributeFee distribute fee to the coodinator idx
func (g *TxGroup) distributeFee(coordIdxsMap map[common.TokenID]common.Idx, processor txProcessor,
	localAccountsDB stateDB) error {
	// create coordinator Idx mapping
	coordIdxs := make([]common.Idx, 0)
	for _, idx := range coordIdxsMap {
		coordIdxs = append(coordIdxs, idx)
	}
	sort.SliceStable(coordIdxs, func(i, j int) bool {
		return coordIdxs[i] < coordIdxs[j]
	})

	// distribute the AccumulatedFees from the processed L2Txs into the Coordinator Idxs
	for idx, accumulatedFee := range processor.AccumulatedCoordFees() {
		// accumulatedFee > 0
		if accumulatedFee.Cmp(big.NewInt(0)) == 1 {
			// send the fee to the Idx of the Coordinator for the TokenID
			accCoord, err := localAccountsDB.GetAccount(idx)
			if err != nil {
				return tracerr.Wrap(err)
			}
			accCoord.Balance = new(big.Int).Add(accCoord.Balance, accumulatedFee)
			_, err = localAccountsDB.UpdateAccount(idx, accCoord)
			if err != nil {
				return tracerr.Wrap(err)
			}
		}
	}
	return nil
}

// validateTransferToEthAddr validates a TransferToEthAddr transaction, it's return a error if occurs
func validateTransferToEthAddr(tx common.PoolL2Tx, l2db l2DB, localAccountsDB stateDB) error {
	// Idx must to be 0
	if tx.ToIdx != 0 {
		return fmt.Errorf(ErrInvalidToIdx)
	}
	// the ToEthAddr must be a valid ethereum address
	if tx.ToEthAddr == common.EmptyAddr || tx.ToEthAddr == common.FFAddr {
		return fmt.Errorf(ErrInvalidToEthAddr)
	}
	// check if the idx already exist
	_, err := localAccountsDB.GetIdxByEthAddr(tx.ToEthAddr, tx.TokenID)
	if err != nil {
		return err
	}
	// check if AccountCreationAuth exist for that ToEthAddr
	_, err = l2db.GetAccountCreationAuth(tx.ToEthAddr)
	return err
}

// validateTransferToEthAddr validates a TransferToBJJ transaction, it's return a error if occurs
func validateTransferToBJJ(tx common.PoolL2Tx, localAccountsDB stateDB) error {
	// Idx must to be 0
	if tx.ToIdx != 0 {
		return fmt.Errorf(ErrInvalidToIdx)
	}
	// the ToBJJ must be a valid BJJ address
	if tx.ToBJJ == common.EmptyBJJComp {
		return fmt.Errorf(ErrInvalidToBjjAddr)
	}
	// the ToEthAddr must be different from the 0xFFF...
	if tx.ToEthAddr != common.FFAddr {
		return fmt.Errorf(ErrInvalidToFAddr)
	}
	// check if the idx already exist
	_, err := localAccountsDB.GetIdxByEthAddrBJJ(tx.ToEthAddr, tx.ToBJJ, tx.TokenID)
	return err
}

// checkBalanceAndNonce check if the balance and nonce are valid
func checkBalanceAndNonce(tx common.PoolL2Tx, localAccountsDB stateDB) error {
	fee, err := common.CalcFeeAmount(tx.Amount, tx.Fee)
	if err != nil {
		return err
	}
	feeAndAmount := new(big.Int).Add(tx.Amount, fee)
	// get the account from the future database to get the future balance
	acc, err := localAccountsDB.GetAccount(tx.FromIdx)
	if err != nil {
		return fmt.Errorf(ErrSenderNotFound)
	}

	// subtract amount and the fee from the sender
	acc.Balance = new(big.Int).Sub(acc.Balance, feeAndAmount)
	if acc.Balance.Cmp(big.NewInt(0)) < 0 { // balance < 0
		return fmt.Errorf(ErrInsufficientFunds)
	}

	// check the nonce only if tx is no processed by the TxProcessor
	if tx.Nonce != acc.Nonce {
		return fmt.Errorf(ErrInvalidNonce)
	}
	return nil
}

// prune prune the less profitable transactions from the group if it's possible
// based on the fee average. This method summits all transactions fee in USD and
// divides the number of transactions to calculate the fee average. After it, we
// compare this fee average from the N transaction to the N-1 transactions. It keeps
// the most profitable group, and after, compares the last result with the N-2 and
// consecutively others until the end to find the most profitable position group.
func (g *TxGroup) prune() bool {
	bestMatch := g.l2Txs
	if g.atomic || len(bestMatch) < 1 {
		return false
	}
	poolTxsLength := len(bestMatch)
	feeAverage := g.feeAverage
	nextIndex := poolTxsLength - 1
	for i := nextIndex; i > 0; i-- {
		auxTxs := g.l2Txs[:i]

		feeSum := new(big.Float)
		for _, auxTx := range auxTxs {
			feeSum = feeSum.Add(feeSum, big.NewFloat(auxTx.AbsoluteFee))
		}
		auxFeeAverage := new(big.Float).Quo(feeSum, big.NewFloat(float64(i)))
		if feeAverage.Cmp(auxFeeAverage) == -1 {
			bestMatch = auxTxs
			feeAverage = auxFeeAverage
		}
	}
	pruned := false
	for i := len(bestMatch); i < len(g.l2Txs); i++ {
		pruned = true
		tx := g.l2Txs[i]
		tx.Info = ErrMaxL2TxSlot
		g.discardedTxs = append(g.discardedTxs, tx)
	}
	log.Debugw("TxSelector: group pruned", "allL2Txs", len(g.l2Txs), "bestMatch", len(bestMatch))
	g.l2Txs = bestMatch
	g.calcFeeAverage()
	return pruned
}

// popAllTxs pop all txs and reset the group
func (g *TxGroup) popAllTxs() {
	g.l1UserTxs = make([]common.L1Tx, 0)
	g.l1CoordTxs = make([]common.L1Tx, 0)
	g.l2Txs = make([]common.PoolL2Tx, 0)
	g.discardedTxs = make([]common.PoolL2Tx, 0)
	g.coordIdxsMap = map[common.TokenID]common.Idx{}
	g.accAuths = make([][]byte, 0)
	g.feeAverage = new(big.Float)
	g.atomic = false
	g.firstPosition = 0
	g.coordAccount = CoordAccount{}
}

// popTx pop the last txs form the group
// It returns true if all txs are discarded
func (g *TxGroup) popTx(number int) bool {
	// If the number is greater then length, pop all transactions
	if g.atomic || number >= g.l2Length() {
		g.popAllTxs()
		return true
	}
	g.l2Txs = g.l2Txs[:len(g.l2Txs)-number]
	return false
}
