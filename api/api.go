package api

import (
	"errors"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/tracerr"
)

// API serves HTTP requests to allow external interaction with the Hermez node
type API struct {
	h             *historydb.HistoryDB
	cg            *configAPI
	l2            *l2db.L2DB
	chainID       uint16
	hermezAddress ethCommon.Address
}

// NewAPI sets the endpoints and the appropriate handlers, but doesn't start the server
func NewAPI(
	coordinatorEndpoints, explorerEndpoints bool,
	server *gin.Engine,
	hdb *historydb.HistoryDB,
	l2db *l2db.L2DB,
) (*API, error) {
	// Check input
	// TODO: is stateDB only needed for explorer endpoints or for both?
	if coordinatorEndpoints && l2db == nil {
		return nil, tracerr.Wrap(errors.New("cannot serve Coordinator endpoints without L2DB"))
	}
	if explorerEndpoints && hdb == nil {
		return nil, tracerr.Wrap(errors.New("cannot serve Explorer endpoints without HistoryDB"))
	}
	consts, err := hdb.GetConstants()
	if err != nil {
		return nil, err
	}
	a := &API{
		h: hdb,
		cg: &configAPI{
			RollupConstants:   *newRollupConstants(consts.Rollup),
			AuctionConstants:  consts.Auction,
			WDelayerConstants: consts.WDelayer,
		},
		l2:            l2db,
		chainID:       consts.ChainID,
		hermezAddress: consts.HermezAddress,
	}

	// Add coordinator endpoints
	if coordinatorEndpoints {
		// Account
		server.POST("/account-creation-authorization", a.postAccountCreationAuth)
		server.GET("/account-creation-authorization/:hezEthereumAddress", a.getAccountCreationAuth)
		// Transaction
		server.POST("/transactions-pool", a.postPoolTx)
		server.GET("/transactions-pool/:id", a.getPoolTx)
	}

	// Add explorer endpoints
	if explorerEndpoints {
		// Account
		server.GET("/accounts", a.getAccounts)
		server.GET("/accounts/:accountIndex", a.getAccount)
		server.GET("/exits", a.getExits)
		server.GET("/exits/:batchNum/:accountIndex", a.getExit)
		// Transaction
		server.GET("/transactions-history", a.getHistoryTxs)
		server.GET("/transactions-history/:id", a.getHistoryTx)
		// Status
		server.GET("/batches", a.getBatches)
		server.GET("/batches/:batchNum", a.getBatch)
		server.GET("/full-batches/:batchNum", a.getFullBatch)
		server.GET("/slots", a.getSlots)
		server.GET("/slots/:slotNum", a.getSlot)
		server.GET("/bids", a.getBids)
		server.GET("/state", a.getState)
		server.GET("/config", a.getConfig)
		server.GET("/tokens", a.getTokens)
		server.GET("/tokens/:id", a.getToken)
		server.GET("/coordinators", a.getCoordinators)
	}

	return a, nil
}
