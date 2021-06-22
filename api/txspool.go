package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
	"github.com/yourbasic/graph"
)

func (a *API) postPoolTx(c *gin.Context) {
	// Parse body
	var receivedTx common.PoolL2Tx
	if err := c.ShouldBindJSON(&receivedTx); err != nil {
		retBadReq(err, c)
		return
	}
	// Transform from received to insert format and validate
	writeTx := receivedTx.toPoolL2TxWrite()
	// Reject atomic transactions
	if isAtomic(*writeTx) {
		retBadReq(errors.New(ErrIsAtomic), c)
		return
	}
	if err := a.verifyPoolL2TxWrite(writeTx); err != nil {
		retBadReq(err, c)
		return
	}
	receivedTx.ClientIP = c.ClientIP()
	// Insert to DB
	if err := a.l2.AddTxAPI(&receivedTx); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return TxID
	c.JSON(http.StatusOK, receivedTx.TxID.String())
}

func (a *API) postAtomicPool(c *gin.Context) {
	// Parse body
	var receivedTxs []common.PoolL2Tx
	if err := c.ShouldBindJSON(&receivedTxs); err != nil {
		retBadReq(err, c)
		return
	}
	nTxs := len(receivedTxs)
	if nTxs <= 1 {
		retBadReq(errors.New(ErrSingleTxInAtomicEndpoint), c)
		return
	}
	// set the Rq fields
	for _, tx1 := range receivedTxs {
		for i, tx2 := range receivedTxs {
			if tx1.RqTxID == tx2.TxID {
				tx1.RqFromIdx = tx2.FromIdx
				tx1.RqToIdx = tx2.ToIdx
				tx1.RqToEthAddr = tx2.ToEthAddr
				tx1.RqToBJJ = tx2.ToBJJ
				tx1.RqTokenID = tx2.TokenID
				tx1.RqAmount = tx2.Amount
				tx1.RqFee = tx2.Fee
				tx1.RqNonce = tx2.Nonce
				break
			}
			// check if was last and not set
			if i == (len(receivedTxs) + 1) {
				retBadReq(errors.New(ErrRqTxIDNotProvided), c)
				return
			}
		}
	}
	// Validate txs individually
	txIDStrings := make([]string, nTxs) // used for successful response
	clientIP := c.ClientIP()
	for i, tx := range receivedTxs {
		if err := a.verifyPoolL2Tx(tx); err != nil {
			retBadReq(err, c)
			return
		}
		receivedTxs[i].ClientIP = clientIP
		txIDStrings[i] = tx.TxID.String()
	}
	// Validate that all txs in the payload represent an atomic group
	if !isAtomicGroup(receivedTxs) {
		retBadReq(errors.New(ErrTxsNotAtomic), c)
		return
	}
	// Insert to DB
	if err := a.l2.AddAtomicTxsAPI(receivedTxs); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return IDs of the added txs in the pool
	c.JSON(http.StatusOK, txIDStrings)
}

// isAtomicGroup returns true if all the txs are needed to be forged
// (all txs will be forged in the same batch or non of them will be forged)
func isAtomicGroup(txs []common.PoolL2Tx) bool {
	// Create a graph from the given txs to represent requests between transactions
	g := graph.New(len(txs))
	idToPos := make(map[common.TxID]int, len(txs))
	// Map tx ID to integers that will represent the nodes of the graph
	for i, tx := range txs {
		idToPos[tx.TxID] = i
	}
	// Create vertices that connect nodes of the graph (txs) using RqTxID
	for i, tx := range txs {
		if tx.RqTxID == common.EmptyTxID {
			// if just one tx doesn't request any other tx, this tx could be forged alone
			// making the hole group not atomic
			return false
		}
		if rqTxPos, ok := idToPos[tx.RqTxID]; ok {
			g.Add(i, rqTxPos)
		} else {
			// tx is requesting a tx that is not provided in the payload
			return false
		}
	}
	// A graph with a single strongly connected component,
	// means that all the nodes can be reached from all the nodes.
	// If tx A "can reach" tx B it means that tx A requests tx B.
	// Therefore we can say that if there is a single strongly connected component in the graph,
	// all the transactions require all trnsactions to be forged, in other words: they are an atomic group
	strongComponents := graph.StrongComponents(g)
	return len(strongComponents) == 1
}

func (a *API) getPoolTx(c *gin.Context) {
	// Get TxID
	txID, err := parsers.ParsePoolTxFilter(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from l2DB
	tx, err := a.l2.GetTxAPI(txID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, tx)
}

func (a *API) getPoolTxs(c *gin.Context) {
	txAPIRequest, err := parsers.ParsePoolTxsFilters(c, a.validate)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch txs from l2DB
	txs, pendingItems, err := a.l2.GetPoolTxsAPI(txAPIRequest)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type txsResponse struct {
		Txs          []l2db.PoolTxAPI `json:"transactions"`
		PendingItems uint64           `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &txsResponse{
		Txs:          txs,
		PendingItems: pendingItems,
	})
}

func (a *API) verifyPoolL2Tx(tx common.PoolL2Tx) error {
	// Check type and id
	_, err := common.NewPoolL2Tx(&tx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Validate feeAmount
	_, err = common.CalcFeeAmount(tx.Amount, tx.Fee)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Get sender account information
	account, err := a.h.GetCommonAccountAPI(tx.FromIdx)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("Error getting from account: %w", err))
	}
	// Validate sender:
	// TokenID
	if tx.TokenID != account.TokenID {
		return tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != account.TokenID (%v)",
			tx.TokenID, account.TokenID))
	}
	// Nonce
	if tx.Nonce < account.Nonce {
		return tracerr.Wrap(fmt.Errorf("tx.Nonce (%v) < account.Nonce (%v)",
			tx.Nonce, account.Nonce))
	}
	// Check signature
	if !tx.VerifySignature(a.cg.ChainID, account.BJJ) {
		return tracerr.Wrap(errors.New("wrong signature"))
	}
	// Check destinatary, note that transactions that are not transfers
	// will always be valid in terms of destinatary (they use special ToIdx by protocol)
	switch tx.Type {
	case common.TxTypeTransfer:
		// ToIdx exists and match token
		toAccount, err := a.h.GetCommonAccountAPI(tx.ToIdx)
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("Error getting to account: %w", err))
		}
		if tx.TokenID != toAccount.TokenID {
			return tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != toAccount.TokenID (%v)",
				tx.TokenID, toAccount.TokenID))
		}
	case common.TxTypeTransferToEthAddr:
		// ToEthAddr has account created with matching token ID or authorization
		ok, err := a.h.CanSendToEthAddr(tx.ToEthAddr, tx.TokenID)
		if err != nil {
			return err
		}
		if !ok {
			return tracerr.Wrap(fmt.Errorf(
				"Destination eth addr (%v) has not a valid account created nor authorization",
				tx.ToEthAddr,
			))
		}
	}
	// Extra sanity checks: those checks are valid as per the protocol, but are very likely to
	// have unexpected side effects that could have a negative impact on users
	switch tx.Type {
	case common.TxTypeExit:
		if tx.Amount.Cmp(big.NewInt(0)) <= 0 {
			return tracerr.New(ErrExitAmount0)
		}
	}
	return nil
}

func isAtomic(tx l2db.PoolL2TxWrite) bool {
	// If a single "Rq" field is different from 0
	return (tx.RqFromIdx != nil && *tx.RqFromIdx != 0) ||
		(tx.RqToIdx != nil && *tx.RqToIdx != 0) ||
		(tx.RqToEthAddr != nil && *tx.RqToEthAddr != common.EmptyAddr) ||
		(tx.RqToBJJ != nil && *tx.RqToBJJ != common.EmptyBJJComp) ||
		(tx.RqAmount != nil && tx.RqAmount != big.NewInt(0)) ||
		(tx.RqFee != nil && *tx.RqFee != 0) ||
		(tx.RqNonce != nil && *tx.RqNonce != 0) ||
		(tx.RqTokenID != nil && *tx.RqTokenID != 0)
}
