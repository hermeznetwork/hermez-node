package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/yourbasic/graph"
)

func (a *API) postPoolTx(c *gin.Context) {
	// Parse body
	var receivedTx receivedPoolTx
	if err := c.ShouldBindJSON(&receivedTx); err != nil {
		retBadReq(err, c)
		return
	}
	// Transform from received to insert format and validate
	writeTx := receivedTx.toPoolL2TxWrite()
	if err := a.verifyPoolL2TxWrite(writeTx); err != nil {
		retBadReq(err, c)
		return
	}
	writeTx.ClientIP = c.ClientIP()
	// Insert to DB
	// note that fee in USD validation happens in this function
	if err := a.l2.AddTxAPI(writeTx); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return TxID
	c.JSON(http.StatusOK, writeTx.TxID.String())
}

func (a *API) postAtomicPool(c *gin.Context) {
	// Parse body
	var receivedTxs []receivedPoolTx
	if err := c.ShouldBindJSON(&receivedTxs); err != nil {
		retBadReq(err, c)
		return
	}
	nTxs := len(receivedTxs)
	if nTxs <= 1 {
		retBadReq(errors.New(ErrSingleTxInAtomicEndpoint), c)
		return
	}
	// Transform from received to insert format and validate (individually)
	writeTxs := make([]l2db.PoolL2TxWrite, nTxs) // used for DB insert
	txIDStrings := make([]string, nTxs)          // used for successful response
	clientIP := c.ClientIP()
	for i, tx := range receivedTxs {
		writeTx := tx.toPoolL2TxWrite()
		if err := a.verifyPoolL2TxWrite(writeTx); err != nil {
			retBadReq(err, c)
			return
		}
		writeTx.ClientIP = clientIP
		writeTxs[i] = *writeTx
		txIDStrings[i] = writeTx.TxID.String()
	}
	// Validate that all txs in the payload represent an atomic group
	if !isAtomicGroup(writeTxs) {
		retBadReq(errors.New(ErrTxsNotAtomic), c)
		return
	}
	// Insert to DB
	if err := a.l2.AddAtomicTxsAPI(writeTxs); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return IDs of the added txs in the pool
	c.JSON(http.StatusOK, txIDStrings)
}

// isAtomicGroup returns true if all the txs are needed to be forged
// (all txs will be forged in the same batch or non of them will be forged)
func isAtomicGroup(txs []l2db.PoolL2TxWrite) bool {
	// Create a graph from the given txs to represent requests between transactions
	g := graph.New(len(txs))
	idToPos := make(map[common.TxID]int, len(txs))
	// Map tx ID to integers that will represent the nodes of the graph
	for i, tx := range txs {
		idToPos[tx.TxID] = i
	}
	// Create vertices that connect nodes of the graph (txs) using RqTxID
	for i, tx := range txs {
		if tx.RqTxID == nil {
			// if just one tx doesn't request any other tx, this tx could be forged alone
			// making the hole group not atomic
			return false
		}
		if rqTxPos, ok := idToPos[*tx.RqTxID]; ok {
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
	txID, err := parseParamTxID(c)
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
	txFilters, err := parseTxsFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// TxType
	txType, err := parseQueryTxType(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Pagination
	fromItem, order, limit, err := parsePagination(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Get state
	state, err := parseQueryPoolL2TxState(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch txs from l2DB
	txs, pendingItems, err := a.l2.GetPoolTxsAPI(l2db.GetPoolTxsAPIRequest{
		EthAddr:     txFilters.addr,
		FromEthAddr: txFilters.fromAddr,
		ToEthAddr:   txFilters.toAddr,
		Bjj:         txFilters.bjj,
		FromBjj:     txFilters.fromBjj,
		ToBjj:       txFilters.toBjj,
		TxType:      txType,
		TokenID:     txFilters.tokenID,
		Idx:         txFilters.idx,
		FromIdx:     txFilters.fromIdx,
		ToIdx:       txFilters.toIdx,
		State:       state,
		FromItem:    fromItem,
		Limit:       limit,
		Order:       order,
	})
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

type receivedPoolTx struct {
	TxID        common.TxID             `json:"id" binding:"required"`
	Type        common.TxType           `json:"type" binding:"required"`
	TokenID     common.TokenID          `json:"tokenId"`
	FromIdx     apitypes.StrHezIdx      `json:"fromAccountIndex" binding:"required"`
	ToIdx       *apitypes.StrHezIdx     `json:"toAccountIndex"`
	ToEthAddr   *apitypes.StrHezEthAddr `json:"toHezEthereumAddress"`
	ToBJJ       *apitypes.StrHezBJJ     `json:"toBjj"`
	Amount      apitypes.StrBigInt      `json:"amount" binding:"required"`
	Fee         common.FeeSelector      `json:"fee"`
	Nonce       common.Nonce            `json:"nonce"`
	Signature   babyjub.SignatureComp   `json:"signature" binding:"required"`
	RqTxID      *common.TxID            `json:"requestId"`
	RqFromIdx   *apitypes.StrHezIdx     `json:"requestFromAccountIndex"`
	RqToIdx     *apitypes.StrHezIdx     `json:"requestToAccountIndex"`
	RqToEthAddr *apitypes.StrHezEthAddr `json:"requestToHezEthereumAddress"`
	RqToBJJ     *apitypes.StrHezBJJ     `json:"requestToBjj"`
	RqTokenID   *common.TokenID         `json:"requestTokenId"`
	RqAmount    *apitypes.StrBigInt     `json:"requestAmount"`
	RqFee       *common.FeeSelector     `json:"requestFee"`
	RqNonce     *common.Nonce           `json:"requestNonce"`
}

func (tx *receivedPoolTx) toPoolL2TxWrite() *l2db.PoolL2TxWrite {
	f := new(big.Float).SetInt((*big.Int)(&tx.Amount))
	amountF, _ := f.Float64()
	return &l2db.PoolL2TxWrite{
		TxID:        tx.TxID,
		FromIdx:     common.Idx(tx.FromIdx),
		ToIdx:       (*common.Idx)(tx.ToIdx),
		ToEthAddr:   (*ethCommon.Address)(tx.ToEthAddr),
		ToBJJ:       (*babyjub.PublicKeyComp)(tx.ToBJJ),
		TokenID:     tx.TokenID,
		Amount:      (*big.Int)(&tx.Amount),
		AmountFloat: amountF,
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
		State:       common.PoolL2TxStatePending,
		Signature:   tx.Signature,
		RqTxID:      (*common.TxID)(tx.RqTxID),
		RqFromIdx:   (*common.Idx)(tx.RqFromIdx),
		RqToIdx:     (*common.Idx)(tx.RqToIdx),
		RqToEthAddr: (*ethCommon.Address)(tx.RqToEthAddr),
		RqToBJJ:     (*babyjub.PublicKeyComp)(tx.RqToBJJ),
		RqTokenID:   tx.RqTokenID,
		RqAmount:    (*big.Int)(tx.RqAmount),
		RqFee:       tx.RqFee,
		RqNonce:     tx.RqNonce,
		Type:        tx.Type,
	}
}

func (a *API) verifyPoolL2TxWrite(txw *l2db.PoolL2TxWrite) error {
	poolTx := common.PoolL2Tx{
		TxID:    txw.TxID,
		FromIdx: txw.FromIdx,
		TokenID: txw.TokenID,
		Amount:  txw.Amount,
		Fee:     txw.Fee,
		Nonce:   txw.Nonce,
		// State:     txw.State,
		Signature: txw.Signature,
		RqAmount:  txw.RqAmount,
		Type:      txw.Type,
	}
	// ToIdx
	if txw.ToIdx != nil {
		poolTx.ToIdx = *txw.ToIdx
	}
	// ToEthAddr
	if txw.ToEthAddr == nil {
		poolTx.ToEthAddr = common.EmptyAddr
	} else {
		poolTx.ToEthAddr = *txw.ToEthAddr
	}
	// ToBJJ
	if txw.ToBJJ == nil {
		poolTx.ToBJJ = common.EmptyBJJComp
	} else {
		poolTx.ToBJJ = *txw.ToBJJ
	}
	// Rq fields
	txHasRq := false
	// RqFromIdx
	if txw.RqFromIdx != nil {
		poolTx.RqFromIdx = *txw.RqFromIdx
		txHasRq = true
	}
	// RqToIdx
	if txw.RqToIdx != nil {
		poolTx.RqToIdx = *txw.RqToIdx
	}
	// RqToEthAddr
	if txw.RqToEthAddr == nil {
		poolTx.RqToEthAddr = common.EmptyAddr
	} else {
		poolTx.RqToEthAddr = *txw.RqToEthAddr
	}
	// RqToBJJ
	if txw.RqToBJJ == nil {
		poolTx.RqToBJJ = common.EmptyBJJComp
	} else {
		poolTx.RqToBJJ = *txw.RqToBJJ
	}
	// RqTokenID
	if txw.RqTokenID != nil {
		poolTx.RqTokenID = *txw.RqTokenID
	}
	// RqFee
	if txw.RqFee != nil {
		poolTx.RqFee = *txw.RqFee
	}
	// RqNonce
	if txw.RqNonce != nil {
		poolTx.RqNonce = *txw.RqNonce
	}
	// RqTxID
	if txw.RqTxID != nil {
		poolTx.RqTxID = *txw.RqTxID
	} else if txHasRq {
		// If any Rq field is set RqTxID shouldn't be nil
		return errors.New(ErrRqTxIDNotProvided)
	}
	// Check type, id and request id
	_, err := common.NewPoolL2Tx(&poolTx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Validate feeAmount
	_, err = common.CalcFeeAmount(poolTx.Amount, poolTx.Fee)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Get sender account information
	account, err := a.h.GetCommonAccountAPI(poolTx.FromIdx)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("Error getting from account: %w", err))
	}
	// Validate sender:
	// TokenID
	if poolTx.TokenID != account.TokenID {
		return tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != account.TokenID (%v)",
			poolTx.TokenID, account.TokenID))
	}
	// Nonce
	if poolTx.Nonce < account.Nonce {
		return tracerr.Wrap(fmt.Errorf("poolTx.Nonce (%v) < account.Nonce (%v)",
			poolTx.Nonce, account.Nonce))
	}
	// Check signature
	if !poolTx.VerifySignature(a.cg.ChainID, account.BJJ) {
		return tracerr.Wrap(errors.New("wrong signature"))
	}
	// Check destinatary, note that transactions that are not transfers
	// will always be valid in terms of destinatary (they use special ToIdx by protocol)
	switch poolTx.Type {
	case common.TxTypeTransfer:
		// ToIdx exists and match token
		toAccount, err := a.h.GetCommonAccountAPI(poolTx.ToIdx)
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("Error getting to account: %w", err))
		}
		if poolTx.TokenID != toAccount.TokenID {
			return tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != toAccount.TokenID (%v)",
				poolTx.TokenID, toAccount.TokenID))
		}
	case common.TxTypeTransferToEthAddr:
		// ToEthAddr has account created with matching token ID or authorization
		ok, err := a.h.CanSendToEthAddr(poolTx.ToEthAddr, poolTx.TokenID)
		if err != nil {
			return err
		}
		if !ok {
			return tracerr.Wrap(fmt.Errorf(
				"Destination eth addr (%v) has not a valid account created nor authorization",
				poolTx.ToEthAddr,
			))
		}
	}
	// Extra sanity checks: those checks are valid as per the protocol, but are very likely to
	// have unexpected side effects that could have a negative impact on users
	switch poolTx.Type {
	case common.TxTypeExit:
		if poolTx.Amount.Cmp(big.NewInt(0)) <= 0 {
			return tracerr.New(ErrExitAmount0)
		}
	}
	return nil
}
