package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
)

func (a *API) postPoolTx(c *gin.Context) {
	// Parse body
	var receivedTx common.PoolL2Tx
	if err := c.ShouldBindJSON(&receivedTx); err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	if isAtomic(receivedTx) {
		retBadReq(&apiError{
			Err:  errors.New(ErrNotAtomicTxsInPostPoolTx),
			Code: ErrNotAtomicTxsInPostPoolTxCode,
			Type: ErrNotAtomicTxsInPostPoolTxType,
		}, c)
		return
	}
	if receivedTx.MaxNumBatch != 0 {
		retBadReq(errors.New(ErrUnsupportedMaxNumBatch), c)
		return
	}
	// Check that tx is valid
	if err := a.verifyPoolL2Tx(receivedTx); err != nil {
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

func (a *API) getPoolTx(c *gin.Context) {
	// Get TxID
	txID, err := parsers.ParsePoolTxFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
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
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
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
		return &apiError{
			Err:  tracerr.Wrap(err),
			Code: ErrInvalidTxTypeOrTxIDCode,
			Type: ErrInvalidTxTypeOrTxIDType,
		}
	}
	// Validate feeAmount
	_, err = common.CalcFeeAmount(tx.Amount, tx.Fee)
	if err != nil {
		return &apiError{
			Err:  tracerr.Wrap(err),
			Code: ErrFeeOverflowCode,
			Type: ErrFeeOverflowType,
		}
	}
	// Get sender account information
	account, err := a.h.GetCommonAccountAPI(tx.FromIdx)
	if err != nil {
		return &apiError{
			Err:  tracerr.Wrap(fmt.Errorf("error getting sender account, idx %s, error: %w", tx.FromIdx, err)),
			Code: ErrGettingSenderAccountCode,
			Type: ErrGettingSenderAccountType,
		}
	}
	// Validate sender:
	// TokenID
	if tx.TokenID != account.TokenID {
		return &apiError{
			Err: tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != account.TokenID (%v)",
				tx.TokenID, account.TokenID)),
			Code: ErrAccountTokenNotEqualTxTokenCode,
			Type: ErrAccountTokenNotEqualTxTokenType,
		}
	}
	// Nonce
	if tx.Nonce < account.Nonce {
		return &apiError{
			Err: tracerr.Wrap(fmt.Errorf("tx.Nonce (%v) < account.Nonce (%v)",
				tx.Nonce, account.Nonce)),
			Code: ErrInvalidNonceCode,
			Type: ErrInvalidNonceType,
		}
	}
	// Check signature
	if !tx.VerifySignature(a.cg.ChainID, account.BJJ) {
		return &apiError{
			Err:  tracerr.Wrap(errors.New("wrong signature")),
			Code: ErrInvalidSignatureCode,
			Type: ErrInvalidSignatureType,
		}
	}
	// Check destination, note that transactions that are not transfers
	// will always be valid in terms of destination (they use special ToIdx by protocol)
	switch tx.Type {
	case common.TxTypeTransfer:
		// ToIdx exists and match token
		toAccount, err := a.h.GetCommonAccountAPI(tx.ToIdx)
		if err != nil {
			return &apiError{
				Err:  tracerr.Wrap(fmt.Errorf("error getting receiver account, idx %s, err: %w", tx.ToIdx, err)),
				Code: ErrGettingReceiverAccountCode,
				Type: ErrGettingReceiverAccountType,
			}
		}
		if tx.TokenID != toAccount.TokenID {
			return &apiError{
				Err: tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != toAccount.TokenID (%v)",
					tx.TokenID, toAccount.TokenID)),
				Code: ErrAccountTokenNotEqualTxTokenCode,
				Type: ErrAccountTokenNotEqualTxTokenType,
			}
		}
	case common.TxTypeTransferToEthAddr:
		// ToEthAddr has account created with matching token ID or authorization
		ok, err := a.h.CanSendToEthAddr(tx.ToEthAddr, tx.TokenID)
		if err != nil {
			return &apiError{
				Err:  tracerr.Wrap(err),
				Code: ErrCantSendToEthAddrCode,
				Type: ErrCantSendToEthAddrType,
			}
		}
		if !ok {
			return &apiError{
				Err: tracerr.Wrap(fmt.Errorf(
					"destination eth addr (%v) has not a valid account created nor authorization",
					tx.ToEthAddr)),
				Code: ErrCantSendToEthAddrCode,
				Type: ErrCantSendToEthAddrType,
			}
		}
	}
	// Extra sanity checks: those checks are valid as per the protocol, but are very likely to
	// have unexpected side effects that could have a negative impact on users
	switch tx.Type {
	case common.TxTypeExit:
		if tx.Amount.Cmp(big.NewInt(0)) <= 0 {
			return &apiError{
				Err:  tracerr.New(ErrExitAmount0),
				Code: ErrExitAmount0Code,
				Type: ErrExitAmount0Type,
			}
		}
	}
	return nil
}

func isAtomic(tx common.PoolL2Tx) bool {
	// If a single "Rq" field is different from 0
	return tx.RqFromIdx != 0 ||
		tx.RqToIdx != 0 ||
		tx.RqToEthAddr != common.EmptyAddr ||
		tx.RqToBJJ != common.EmptyBJJComp ||
		(tx.RqAmount != nil && tx.RqAmount.Cmp(big.NewInt(0)) != 0) ||
		tx.RqFee != 0 ||
		tx.RqNonce != 0 ||
		tx.RqTokenID != 0
}
