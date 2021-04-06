package api

import (
	"net/http"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

type GetTxsAPIRequest struct {
	EthAddr           *ethCommon.Address
	Bjj               *babyjub.PublicKeyComp
	TokenID           *common.TokenID
	Idx               *common.Idx
	BatchNum          *uint
	TxType            *common.TxType
	IncludePendingL1s *bool

	FromItem *uint
	Limit    *uint
	Order    string
}

func (a *API) getHistoryTxs(c *gin.Context) {
	// Get query parameters
	tokenID, addr, bjj, idx, err := parseExitFilters(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// BatchNum
	batchNum, err := parseQueryUint("batchNum", nil, 0, maxUint32, c)
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
	// IncludePendingL1s
	includePendingL1s := new(bool)
	*includePendingL1s = false
	includePendingL1s, err = parseQueryBool("includePendingL1s", includePendingL1s, c)
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
	request := GetTxsAPIRequest{
		EthAddr:           addr,
		Bjj:               bjj,
		TokenID:           tokenID,
		Idx:               idx,
		BatchNum:          batchNum,
		TxType:            txType,
		IncludePendingL1s: includePendingL1s,
		FromItem:          fromItem,
		Limit:             limit,
		Order:             order,
	}
	// Fetch txs from historyDB
	txs, pendingItems, err := a.h.GetTxsAPI(request)
	if err != nil {
		retSQLErr(err, c)
		return
	}

	// Build successful response
	type txsResponse struct {
		Txs          []historydb.TxAPI `json:"transactions"`
		PendingItems uint64            `json:"pendingItems"`
	}
	c.JSON(http.StatusOK, &txsResponse{
		Txs:          txs,
		PendingItems: pendingItems,
	})
}

func (a *API) getHistoryTx(c *gin.Context) {
	// Get TxID
	txID, err := parseParamTxID(c)
	if err != nil {
		retBadReq(err, c)
		return
	}
	// Fetch tx from historyDB
	tx, err := a.h.GetTxAPI(txID)
	if err != nil {
		retSQLErr(err, c)
		return
	}
	// Build successful response
	c.JSON(http.StatusOK, tx)
}
