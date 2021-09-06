package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/yourbasic/graph"
)

func (a *API) postAtomicPool(c *gin.Context) {
	// Parse body
	var receivedAtomicGroup common.AtomicGroup
	if err := c.ShouldBindJSON(&receivedAtomicGroup); err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Validate atomic group id
	if !receivedAtomicGroup.IsAtomicGroupIDValid() {
		retBadReq(&apiError{
			Err:  errors.New(ErrInvalidAtomicGroupID),
			Code: ErrInvalidAtomicGroupIDCode,
			Type: ErrInvalidAtomicGroupIDType,
		}, c)
		return
	}
	nTxs := len(receivedAtomicGroup.Txs)
	if nTxs <= 1 {
		retBadReq(&apiError{
			Err:  errors.New(ErrSingleTxInAtomicEndpoint),
			Code: ErrSingleTxInAtomicEndpointCode,
			Type: ErrSingleTxInAtomicEndpointType,
		}, c)
		return
	}
	// Validate txs
	txIDStrings := make([]string, nTxs) // used for successful response
	clientIP := c.ClientIP()
	for i, tx := range receivedAtomicGroup.Txs {
		// Find requested transaction
		relativePosition, err := RequestOffset2RelativePosition(tx.RqOffset)
		if err != nil {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrFailedToFindOffsetToRelativePositionCode,
				Type: ErrFailedToFindOffsetToRelativePositionType,
			}, c)
			return
		}
		requestedPosition := i + relativePosition
		if requestedPosition > len(receivedAtomicGroup.Txs)-1 || requestedPosition < 0 {
			retBadReq(&apiError{
				Err:  errors.New(ErrRqOffsetOutOfBounds),
				Code: ErrRqOffsetOutOfBoundsCode,
				Type: ErrRqOffsetOutOfBoundsType,
			}, c)
			return
		}
		// Set fields that are omitted in the JSON
		requestedTx := receivedAtomicGroup.Txs[requestedPosition]
		receivedAtomicGroup.Txs[i].RqFromIdx = requestedTx.FromIdx
		receivedAtomicGroup.Txs[i].RqToIdx = requestedTx.ToIdx
		receivedAtomicGroup.Txs[i].RqToEthAddr = requestedTx.ToEthAddr
		receivedAtomicGroup.Txs[i].RqToBJJ = requestedTx.ToBJJ
		receivedAtomicGroup.Txs[i].RqTokenID = requestedTx.TokenID
		receivedAtomicGroup.Txs[i].RqAmount = requestedTx.Amount
		receivedAtomicGroup.Txs[i].RqFee = requestedTx.Fee
		receivedAtomicGroup.Txs[i].RqNonce = requestedTx.Nonce
		receivedAtomicGroup.Txs[i].ClientIP = clientIP
		receivedAtomicGroup.Txs[i].AtomicGroupID = receivedAtomicGroup.ID
		receivedAtomicGroup.Txs[i].Info = ""

		// Validate transaction
		if err := a.verifyPoolL2Tx(receivedAtomicGroup.Txs[i]); err != nil {
			retBadReq(err, c)
			return
		}

		// Prepare response
		txIDStrings[i] = receivedAtomicGroup.Txs[i].TxID.String()
	}

	// Validate that all txs in the payload represent a single atomic group
	if !isSingleAtomicGroup(receivedAtomicGroup.Txs) {
		retBadReq(&apiError{
			Err:  errors.New(ErrTxsNotAtomic),
			Code: ErrTxsNotAtomicCode,
			Type: ErrTxsNotAtomicType,
		}, c)
		return
	}
	// Insert to DB
	if err := a.l2DB.AddAtomicTxsAPI(receivedAtomicGroup.Txs); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return IDs of the added txs in the pool
	c.JSON(http.StatusOK, txIDStrings)
}

// RequestOffset2RelativePosition translates from 0 to 7 to protocol position
func RequestOffset2RelativePosition(rqoffset uint8) (int, error) {
	const (
		rqOffsetZero       = 0
		rqOffsetOne        = 1
		rqOffsetTwo        = 2
		rqOffsetThree      = 3
		rqOffsetFour       = 4
		rqOffsetFive       = 5
		rqOffsetSix        = 6
		rqOffsetSeven      = 7
		rqOffsetMinusFour  = -4
		rqOffsetMinusThree = -3
		rqOffsetMinusTwo   = -2
		rqOffsetMinusOne   = -1
	)

	switch rqoffset {
	case rqOffsetZero:
		return rqOffsetZero, errors.New(ErrTxsNotAtomic)
	case rqOffsetOne:
		return rqOffsetOne, nil
	case rqOffsetTwo:
		return rqOffsetTwo, nil
	case rqOffsetThree:
		return rqOffsetThree, nil
	case rqOffsetFour:
		return rqOffsetMinusFour, nil
	case rqOffsetFive:
		return rqOffsetMinusThree, nil
	case rqOffsetSix:
		return rqOffsetMinusTwo, nil
	case rqOffsetSeven:
		return rqOffsetMinusOne, nil
	default:
		return rqOffsetZero, errors.New(ErrInvalidRqOffset)
	}
}

// isSingleAtomicGroup returns true if all the txs are needed to be forged
// (all txs will be forged in the same batch or non of them will be forged)
func isSingleAtomicGroup(txs []common.PoolL2Tx) bool {
	// Create a graph from the given txs to represent requests between transactions
	g := graph.New(len(txs))
	// Create vertices that connect nodes of the graph (txs) using RqOffset
	for i, tx := range txs {
		requestedRelativePosition, err := RequestOffset2RelativePosition(tx.RqOffset)
		if err != nil {
			return false
		}
		requestedPosition := i + requestedRelativePosition
		if requestedPosition < 0 || requestedPosition >= len(txs) {
			// Safety check: requested tx is not out of array bounds
			return false
		}
		g.Add(i, requestedPosition)
	}
	// A graph with a single strongly connected component,
	// means that all the nodes can be reached from all the nodes.
	// If tx A "can reach" tx B it means that tx A requests tx B.
	// Therefore we can say that if there is a single strongly connected component in the graph,
	// all the transactions require all trnsactions to be forged, in other words: they are an atomic group
	strongComponents := graph.StrongComponents(g)
	return len(strongComponents) == 1
}

func (a *API) getAtomicGroup(c *gin.Context) {
	// Get TxID
	atomicGroupID, err := parsers.ParseParamAtomicGroupID(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch tx from l2DB
	txs, err := a.l2DB.GetPoolTxsByAtomicGroupIDAPI(atomicGroupID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, txs)
}
