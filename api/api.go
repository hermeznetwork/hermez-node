package api

import (
	"errors"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/tracerr"
)

// TODO: Add correct values to constants
const (
	createAccountExtraFeePercentage         float64 = 2
	createAccountInternalExtraFeePercentage float64 = 2.5
)

// Status define status of the network
type Status struct {
	sync.RWMutex
	Network           Network                  `json:"network"`
	Metrics           historydb.Metrics        `json:"metrics"`
	Rollup            common.RollupVariables   `json:"rollup"`
	Auction           common.AuctionVariables  `json:"auction"`
	WithdrawalDelayer common.WDelayerVariables `json:"withdrawalDelayer"`
	RecommendedFee    common.RecommendedFee    `json:"recommendedFee"`
}

// API serves HTTP requests to allow external interaction with the Hermez node
type API struct {
	h       *historydb.HistoryDB
	cg      *configAPI
	s       *statedb.StateDB
	l2      *l2db.L2DB
	status  Status
	chainID uint16
}

// NewAPI sets the endpoints and the appropriate handlers, but doesn't start the server
func NewAPI(
	coordinatorEndpoints, explorerEndpoints bool,
	server *gin.Engine,
	hdb *historydb.HistoryDB,
	sdb *statedb.StateDB,
	l2db *l2db.L2DB,
	config *Config,
	chainID uint16,
) (*API, error) {
	// Check input
	// TODO: is stateDB only needed for explorer endpoints or for both?
	if coordinatorEndpoints && l2db == nil {
		return nil, tracerr.Wrap(errors.New("cannot serve Coordinator endpoints without L2DB"))
	}
	if explorerEndpoints && hdb == nil {
		return nil, tracerr.Wrap(errors.New("cannot serve Explorer endpoints without HistoryDB"))
	}

	a := &API{
		h: hdb,
		cg: &configAPI{
			RollupConstants:   *newRollupConstants(config.RollupConstants),
			AuctionConstants:  config.AuctionConstants,
			WDelayerConstants: config.WDelayerConstants,
		},
		s:       sdb,
		l2:      l2db,
		status:  Status{},
		chainID: chainID,
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
