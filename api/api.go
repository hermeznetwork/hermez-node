package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
)

var h *historydb.HistoryDB
var cg *configAPI
var s *statedb.StateDB
var l2 *l2db.L2DB

// SetAPIEndpoints sets the endpoints and the appropriate handlers, but doesn't start the server
func SetAPIEndpoints(
	coordinatorEndpoints, explorerEndpoints bool,
	server *gin.Engine,
	hdb *historydb.HistoryDB,
	sdb *statedb.StateDB,
	l2db *l2db.L2DB,
	config *configAPI,
) error {
	// Check input
	// TODO: is stateDB only needed for explorer endpoints or for both?
	if coordinatorEndpoints && l2db == nil {
		return errors.New("cannot serve Coordinator endpoints without L2DB")
	}
	if explorerEndpoints && hdb == nil {
		return errors.New("cannot serve Explorer endpoints without HistoryDB")
	}

	h = hdb
	cg = config
	s = sdb
	l2 = l2db

	// Add coordinator endpoints
	if coordinatorEndpoints {
		// Account
		server.POST("/account-creation-authorization", postAccountCreationAuth)
		server.GET("/account-creation-authorization/:hermezEthereumAddress", getAccountCreationAuth)
		// Transaction
		server.POST("/transactions-pool", postPoolTx)
		server.GET("/transactions-pool/:id", getPoolTx)
	}

	// Add explorer endpoints
	if explorerEndpoints {
		// Account
		server.GET("/accounts", getAccounts)
		server.GET("/accounts/:hermezEthereumAddress/:accountIndex", getAccount)
		server.GET("/exits", getExits)
		server.GET("/exits/:batchNum/:accountIndex", getExit)
		// Transaction
		server.GET("/transactions-history", getHistoryTxs)
		server.GET("/transactions-history/:id", getHistoryTx)
		// Status
		server.GET("/batches", getBatches)
		server.GET("/batches/:batchNum", getBatch)
		server.GET("/full-batches/:batchNum", getFullBatch)
		server.GET("/slots", getSlots)
		server.GET("/slots/:slotNum", getSlot)
		server.GET("/bids", getBids)
		server.GET("/next-forgers", getNextForgers)
		server.GET("/state", getState)
		server.GET("/config", getConfig)
		server.GET("/tokens", getTokens)
		server.GET("/tokens/:id", getToken)
		server.GET("/recommendedFee", getRecommendedFee)
		server.GET("/coordinators", getCoordinators)
		server.GET("/coordinators/:bidderAddr", getCoordinator)
	}

	return nil
}
