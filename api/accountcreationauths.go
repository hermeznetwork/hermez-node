package api

import (
	"errors"
	"net/http"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

func (a *API) postAccountCreationAuth(c *gin.Context) {
	// Parse body
	var apiAuth receivedAuth
	if err := c.ShouldBindJSON(&apiAuth); err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType}, c)
		return
	}
	// API to common + verify signature
	commonAuth := accountCreationAuthAPIToCommon(&apiAuth)
	isValid, err := commonAuth.VerifySignature(a.config.ChainID, a.hermezAddress)
	if !isValid && err != nil {
		retBadReq(&apiError{
			Err:  errors.New("invalid signature: " + err.Error()),
			Code: ErrInvalidSignatureCode,
			Type: ErrInvalidSignatureType,
		}, c)
		return
	}
	// Insert to DB
	if err := a.l2DB.AddAccountCreationAuthAPI(commonAuth); err != nil {
		retSQLErr(err, c)
		return
	}
	type okResponse struct {
		Success string `json:"success"`
	}
	// Return OK
	c.JSON(http.StatusOK, &okResponse{
		Success: "OK",
	})
}

func (a *API) getAccountCreationAuth(c *gin.Context) {
	// Get hezEthereumAddress
	addr, err := parsers.ParseGetAccountCreationAuthFilter(c)
	if err != nil {
		retBadReq(&apiError{
			Err:  err,
			Code: ErrParamValidationFailedCode,
			Type: ErrParamValidationFailedType,
		}, c)
		return
	}
	// Fetch auth from l2DB
	auth, err := a.l2DB.GetAccountCreationAuthAPI(*addr)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, auth)
}

type receivedAuth struct {
	EthAddr   apitypes.StrHezEthAddr `json:"hezEthereumAddress" binding:"required"`
	BJJ       apitypes.StrHezBJJ     `json:"bjj" binding:"required"`
	Signature apitypes.EthSignature  `json:"signature" binding:"required"`
	Timestamp time.Time              `json:"timestamp"`
}

func accountCreationAuthAPIToCommon(apiAuth *receivedAuth) *common.AccountCreationAuth {
	return &common.AccountCreationAuth{
		EthAddr:   ethCommon.Address(apiAuth.EthAddr),
		BJJ:       (babyjub.PublicKeyComp)(apiAuth.BJJ),
		Signature: []byte(apiAuth.Signature),
		Timestamp: apiAuth.Timestamp,
	}
}
