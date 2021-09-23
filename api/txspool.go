package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

const sqlError = "500 - sql error"

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
	receivedTx.ClientIP = c.ClientIP()
	receivedTx.Info = ""
	if err := a.validateAndStorePoolTx(receivedTx); err != nil {
		if err.Type == sqlError {
			retSQLErr(err.Err, c)
			return
		}
		retBadReq(err, c)
		return
	}
	// Return TxID
	c.JSON(http.StatusOK, receivedTx.TxID.String())
	// Publish tx on coordinator network
	if a.coordnet != nil {
		if err := a.coordnet.PublishTx(receivedTx); err != nil {
			log.Warn(err)
		}
	}
}

func (a *API) coordnetPoolTxHandler(tx common.PoolL2Tx) error {
	if err := a.validateAndStorePoolTx(tx); err != nil {
		return err.Err
	}
	return nil
}

func (a *API) validateAndStorePoolTx(tx common.PoolL2Tx) *apiError {
	// Check if tx is atomic (has any non 0ed Rq* field)
	if isAtomic(tx) {
		return &apiError{
			Err:  errors.New(ErrNotAtomicTxsInPostPoolTx),
			Code: ErrNotAtomicTxsInPostPoolTxCode,
			Type: ErrNotAtomicTxsInPostPoolTxType,
		}
	}
	// Check that tx is valid
	if err := a.verifyPoolL2Tx(tx); err != nil {
		return err
	}
	// Insert to DB
	if err := a.l2DB.AddTxAPI(&tx); err != nil {
		if strings.Contains(err.Error(), "< minFeeUSD") {
			return &apiError{
				Err:  err,
				Code: ErrFeeTooLowCode,
				Type: ErrFeeTooLowType,
			}
		} else if strings.Contains(err.Error(), "> maxFeeUSD") {
			return &apiError{
				Err:  err,
				Code: ErrFeeTooBigCode,
				Type: ErrFeeTooBigType,
			}
		}
		// SQL error
		return &apiError{
			Err:  err,
			Type: sqlError,
		}
	}
	// Tx inserted
	return nil
}

func (a *API) putPoolTxByIdxAndNonce(c *gin.Context) {
	idx, nonce, err := parsers.ParsePoolTxUpdateByIdxAndNonceFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	var receivedTx common.PoolL2Tx
	if err = c.ShouldBindJSON(&receivedTx); err != nil {
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
	if receivedTx.State != common.PoolL2TxStatePending || receivedTx.FromIdx != idx || receivedTx.Nonce != nonce {
		retBadReq(&apiError{
			Err:  errors.New("tx state is not pend or invl or fromIdx or nonce in request body not equal request uri params"),
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	if apiErr := a.verifyPoolL2Tx(receivedTx); apiErr != nil {
		retBadReq(apiErr, c)
		return
	}

	receivedTx.ClientIP = c.ClientIP()
	receivedTx.Info = ""

	if err := a.l2DB.UpdateTxByIdxAndNonceAPI(idx, nonce, &receivedTx); err != nil {
		if strings.Contains(err.Error(), "< minFeeUSD") {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrFeeTooLowCode,
				Type: ErrFeeTooLowType,
			}, c)
			return
		} else if strings.Contains(err.Error(), "> maxFeeUSD") {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrFeeTooBigCode,
				Type: ErrFeeTooBigType,
			}, c)
			return
		}
		retSQLErr(err, c)
		return
	}

	c.JSON(http.StatusOK, receivedTx.TxID.String())
}

func (a *API) putPoolTx(c *gin.Context) {
	txID, err := parsers.ParsePoolTxFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	var receivedTx common.PoolL2Tx
	if err := c.ShouldBindJSON(&receivedTx); err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}

	receivedTx.TxID = txID

	if isAtomic(receivedTx) {
		retBadReq(&apiError{
			Err:  errors.New(ErrNotAtomicTxsInPostPoolTx),
			Code: ErrNotAtomicTxsInPostPoolTxCode,
			Type: ErrNotAtomicTxsInPostPoolTxType,
		}, c)
		return
	}
	if apiErr := a.verifyPoolL2Tx(receivedTx); apiErr != nil {
		retBadReq(apiErr, c)
		return
	}
	receivedTx.ClientIP = c.ClientIP()
	receivedTx.Info = ""

	if err := a.l2DB.UpdateTxAPI(&receivedTx); err != nil {
		if strings.Contains(err.Error(), "< minFeeUSD") {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrFeeTooLowCode,
				Type: ErrFeeTooLowType,
			}, c)
			return
		} else if strings.Contains(err.Error(), "> maxFeeUSD") {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrFeeTooBigCode,
				Type: ErrFeeTooBigType,
			}, c)
			return
		} else if strings.Contains(err.Error(), "nothing to update") {
			retBadReq(&apiError{
				Err:  err,
				Code: ErrNothingToUpdateCode,
				Type: ErrNothingToUpdateType,
			}, c)
		}
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
	tx, err := a.l2DB.GetTxAPI(txID)
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
	txs, pendingItems, err := a.l2DB.GetPoolTxsAPI(txAPIRequest)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type txsResponse struct {
		Txs          []apitypes.TxL2 `json:"transactions"`
		PendingItems uint64          `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &txsResponse{
		Txs:          txs,
		PendingItems: pendingItems,
	})
}

func (a *API) verifyPoolL2Tx(tx common.PoolL2Tx) *apiError {
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
	account, err := a.historyDB.GetCommonAccountAPI(tx.FromIdx)
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
	if !tx.VerifySignature(a.config.ChainID, account.BJJ) {
		return &apiError{
			Err:  tracerr.Wrap(errors.New("wrong signature")),
			Code: ErrInvalidSignatureCode,
			Type: ErrInvalidSignatureType,
		}
	}
	switch tx.Type {
	// Check destination, note that transactions that are not transfers
	// will always be valid in terms of destination (they use special ToIdx by protocol)
	case common.TxTypeTransfer:
		// ToIdx exists and match token
		toAccount, err := a.historyDB.GetCommonAccountAPI(tx.ToIdx)
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
	// Extra sanity checks: those checks are valid as per the protocol, but are very likely to
	// have unexpected side effects that could have a negative impact on users
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
