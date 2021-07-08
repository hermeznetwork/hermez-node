package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

func (a *API) postPoolTx(c *gin.Context) {
	// Parse body
	var receivedTx receivedPoolTx
	if err := c.ShouldBindJSON(&receivedTx); err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Transform from received to insert format and validate
	writeTx := receivedTx.toPoolL2TxWrite()
	// Reject atomic transactions
	if isAtomic(*writeTx) {
		retBadReq(&apiError{
			Err:  errors.New(ErrIsAtomic),
			Code: ErrIsAtomicCode,
			Type: ErrIsAtomicType,
		}, c)
		return
	}
	if err := a.verifyPoolL2TxWrite(writeTx); err != nil {
		retBadReq(err, c)
		return
	}
	writeTx.ClientIP = c.ClientIP()
	// Insert to DB
	if err := a.l2.AddTxAPI(writeTx); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return TxID
	c.JSON(http.StatusOK, writeTx.TxID.String())
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

func (a *API) verifyPoolL2TxWrite(txw *l2db.PoolL2TxWrite) *apiError {
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
	// RqFromIdx
	if txw.RqFromIdx != nil {
		poolTx.RqFromIdx = *txw.RqFromIdx
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
	// Check type and id
	_, err := common.NewPoolL2Tx(&poolTx)
	if err != nil {
		return &apiError{
			Err:  tracerr.Wrap(err),
			Code: ErrInvalidTxTypeOrTxIDCode,
			Type: ErrInvalidTxTypeOrTxIDType,
		}
	}
	// Validate feeAmount
	_, err = common.CalcFeeAmount(poolTx.Amount, poolTx.Fee)
	if err != nil {
		return &apiError{
			Err:  tracerr.Wrap(err),
			Code: ErrFeeOverflowCode,
			Type: ErrFeeOverflowType,
		}
	}
	// Get sender account information
	account, err := a.h.GetCommonAccountAPI(poolTx.FromIdx)
	if err != nil {
		return &apiError{
			Err:  tracerr.Wrap(fmt.Errorf("error getting sender account, idx %s, error: %w", poolTx.FromIdx, err)),
			Code: ErrGettingSenderAccountCode,
			Type: ErrGettingSenderAccountType,
		}
	}
	// Validate sender:
	// TokenID
	if poolTx.TokenID != account.TokenID {
		return &apiError{
			Err: tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != account.TokenID (%v)",
				poolTx.TokenID, account.TokenID)),
			Code: ErrAccountTokenNotEqualTxTokenCode,
			Type: ErrAccountTokenNotEqualTxTokenType,
		}
	}
	// Nonce
	if poolTx.Nonce < account.Nonce {
		return &apiError{
			Err: tracerr.Wrap(fmt.Errorf("poolTx.Nonce (%v) < account.Nonce (%v)",
				poolTx.Nonce, account.Nonce)),
			Code: ErrInvalidNonceCode,
			Type: ErrInvalidNonceType,
		}
	}
	// Check signature
	if !poolTx.VerifySignature(a.cg.ChainID, account.BJJ) {
		return &apiError{
			Err:  tracerr.Wrap(errors.New("wrong signature")),
			Code: ErrInvalidSignatureCode,
			Type: ErrInvalidSignatureType,
		}
	}
	// Check destination, note that transactions that are not transfers
	// will always be valid in terms of destination (they use special ToIdx by protocol)
	switch poolTx.Type {
	case common.TxTypeTransfer:
		// ToIdx exists and match token
		toAccount, err := a.h.GetCommonAccountAPI(poolTx.ToIdx)
		if err != nil {
			return &apiError{
				Err:  tracerr.Wrap(fmt.Errorf("error getting receiver account, idx %s, err: %w", poolTx.ToIdx, err)),
				Code: ErrGettingReceiverAccountCode,
				Type: ErrGettingReceiverAccountType,
			}
		}
		if poolTx.TokenID != toAccount.TokenID {
			return &apiError{
				Err: tracerr.Wrap(fmt.Errorf("tx.TokenID (%v) != toAccount.TokenID (%v)",
					poolTx.TokenID, toAccount.TokenID)),
				Code: ErrAccountTokenNotEqualTxTokenCode,
				Type: ErrAccountTokenNotEqualTxTokenType,
			}
		}
	case common.TxTypeTransferToEthAddr:
		// ToEthAddr has account created with matching token ID or authorization
		ok, err := a.h.CanSendToEthAddr(poolTx.ToEthAddr, poolTx.TokenID)
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
					poolTx.ToEthAddr)),
				Code: ErrCantSendToEthAddrCode,
				Type: ErrCantSendToEthAddrType,
			}
		}
	}
	// Extra sanity checks: those checks are valid as per the protocol, but are very likely to
	// have unexpected side effects that could have a negative impact on users
	switch poolTx.Type {
	case common.TxTypeExit:
		if poolTx.Amount.Cmp(big.NewInt(0)) <= 0 {
			return &apiError{
				Err:  tracerr.New(ErrExitAmount0),
				Code: ErrExitAmount0Code,
				Type: ErrExitAmount0Type,
			}
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
