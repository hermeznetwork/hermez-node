package api

import (
	"errors"
	"net/http"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

func (a *API) postAccountCreationAuth(c *gin.Context) {
	// Parse body
	var apiAuth receivedAuth
	if err := c.ShouldBindJSON(&apiAuth); err != nil {
		retBadReq(err, c)
		return
	}
	// API to common + verify signature
	commonAuth := accountCreationAuthAPIToCommon(&apiAuth)
	if !commonAuth.VerifySignature(a.chainID, a.hermezAddress) {
		retBadReq(errors.New("invalid signature"), c)
		return
	}
	// Insert to DB
	if err := a.l2.AddAccountCreationAuthAPI(commonAuth); err != nil {
		retSQLErr(err, c)
		return
	}
	// Return OK
	c.Status(http.StatusOK)
}

func (a *API) getAccountCreationAuth(c *gin.Context) {
	// Get hezEthereumAddress
	addr, err := parseParamHezEthAddr(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch auth from l2DB
	auth, err := a.l2.GetAccountCreationAuthAPI(*addr)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build succesfull response
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
