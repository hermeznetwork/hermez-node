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

	v1 := server.Group("/v1")

	// Add coordinator endpoints
	if coordinatorEndpoints {
		// Account
		v1.POST("/account-creation-authorization", a.postAccountCreationAuth)
		v1.GET("/account-creation-authorization/:hezEthereumAddress", a.getAccountCreationAuth)
		// Transaction
		v1.POST("/transactions-pool", a.postPoolTx)
		v1.GET("/transactions-pool/:id", a.getPoolTx)
		v1.GET("/transactions-pool", a.getPoolTxs)
	}

	// Add explorer endpoints
	if explorerEndpoints {
		// Account
		v1.GET("/accounts", a.getAccounts)
		v1.GET("/accounts/:accountIndex", a.getAccount)
		v1.GET("/exits", a.getExits)
		v1.GET("/exits/:batchNum/:accountIndex", a.getExit)
		// Transaction
		v1.GET("/transactions-history", a.getHistoryTxs)
		v1.GET("/transactions-history/:id", a.getHistoryTx)
		// Status
		v1.GET("/batches", a.getBatches)
		v1.GET("/batches/:batchNum", a.getBatch)
		v1.GET("/full-batches/:batchNum", a.getFullBatch)
		v1.GET("/slots", a.getSlots)
		v1.GET("/slots/:slotNum", a.getSlot)
		v1.GET("/bids", a.getBids)
		v1.GET("/state", a.getState)
		v1.GET("/config", a.getConfig)
		v1.GET("/tokens", a.getTokens)
		v1.GET("/tokens/:id", a.getToken)
		v1.GET("/coordinators", a.getCoordinators)
	}

	return a, nil
}
