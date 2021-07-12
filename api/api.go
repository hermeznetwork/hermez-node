/*
Package api implements the public interface of the hermez-node using a HTTP REST API.
There are two subsets of endpoints:
- coordinatorEndpoints: used to receive L2 transactions and account creation authorizations. Targeted for wallets.
- explorerEndpoints: used to provide all sorts of information about the network. Targeted for explorers and similar services.

About the configuration of the API:
- The API is supposed to be launched using the cli found at the package cli/node, and configured through the configuration file.
- The mentioned configuration file allows exposing any combination of the endpoint subsets.
- Although the API can run in a "standalone" manner using the serveapi command, it won't work properly
unless another process acting as a coord or sync is filling the HistoryDB.

Design principles and considerations:
- In order to decouple the API process from the rest of the node, all the communication between this package and the rest of
the system is done through the SQL database. As a matter of fact, the only public function of the package is the constructor NewAPI.
All the information needed for the API to work should be obtained through the configuration file of the cli or the database.
- The format of the requests / responses doesn't match directly with the common types, and for this reason, the package api/apitypes is used
to facilitate the format conversion. Most of the time, this is done directly at the db level.
- The API endpoints are fully documented using OpenAPI aka Swagger. All the endpoints are tested against the spec to ensure consistency
between implementation and specification. To get a sense of which endpoints exist and how they work, it's strongly recommended to check this specification.
The specification can be found at api/swagger.yml.
- In general, all the API endpoints produce queries to the SQL database in order to retrieve / insert the requested information. The most notable exceptions to this are
the /config endpoint, which returns a static object generated at construction time, and the /state, which also is retrieved from the database, but it's generated by API/stateapiupdater package.
*/
package api

import (
	"errors"
	"reflect"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api/parsers"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/metric"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

// API serves HTTP requests to allow external interaction with the Hermez node
type API struct {
	h             *historydb.HistoryDB
	cg            *configAPI
	l2            *l2db.L2DB
	hermezAddress ethCommon.Address
	validate      *validator.Validate
}

// NewAPI sets the endpoints and the appropriate handlers, but doesn't start the server
func NewAPI(
	version string,
	coordinatorEndpoints, explorerEndpoints bool,
	server *gin.Engine,
	hdb *historydb.HistoryDB,
	l2db *l2db.L2DB,
	ethClient *ethclient.Client,
	forgerAddress *ethCommon.Address,
) (*API, error) {
	// Check input
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
			ChainID:           consts.ChainID,
		},
		l2:            l2db,
		hermezAddress: consts.HermezAddress,
		validate:      newValidate(),
	}

	middleware, err := metric.PrometheusMiddleware()
	if err != nil {
		return nil, err
	}
	server.Use(middleware)

	server.NoRoute(a.noRoute)

	v1 := server.Group("/v1")

	v1.GET("/health", gin.WrapH(a.healthRoute(version, ethClient, forgerAddress)))
	// Add coordinator endpoints
	if coordinatorEndpoints {
		// Account creation authorization
		v1.POST("/account-creation-authorization", a.postAccountCreationAuth)
		v1.GET("/account-creation-authorization/:hezEthereumAddress", a.getAccountCreationAuth)
		// Transaction
		v1.POST("/transactions-pool", a.postPoolTx)
		v1.POST("/atomic-pool", a.postAtomicPool)
		v1.GET("/transactions-pool/:id", a.getPoolTx)
		v1.GET("/transactions-pool", a.getPoolTxs)
		v1.GET("/atomic-pool/:id", a.getAtomicGroup)
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
		// Batches
		v1.GET("/batches", a.getBatches)
		v1.GET("/batches/:batchNum", a.getBatch)
		v1.GET("/full-batches/:batchNum", a.getFullBatch)
		// Slots
		v1.GET("/slots", a.getSlots)
		v1.GET("/slots/:slotNum", a.getSlot)
		// Bids
		v1.GET("/bids", a.getBids)
		// State
		v1.GET("/state", a.getState)
		// Config
		v1.GET("/config", a.getConfig)
		// Tokens
		v1.GET("/tokens", a.getTokens)
		v1.GET("/tokens/:id", a.getToken)
		// Fiat Currencies
		v1.GET("/currencies", a.getFiatCurrencies)
		v1.GET("/currencies/:symbol", a.getFiatCurrency)
		// Coordinators
		v1.GET("/coordinators", a.getCoordinators)
	}

	return a, nil
}

func newValidate() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	validate.RegisterStructValidation(parsers.ExitsFiltersStructValidation, parsers.ExitsFilters{})
	validate.RegisterStructValidation(parsers.BidsFiltersStructValidation, parsers.BidsFilters{})
	validate.RegisterStructValidation(parsers.AccountsFiltersStructValidation, parsers.AccountsFilters{})
	validate.RegisterStructValidation(parsers.HistoryTxsFiltersStructValidation, parsers.HistoryTxsFilters{})
	validate.RegisterStructValidation(parsers.PoolTxsTxsFiltersStructValidation, parsers.PoolTxsFilters{})
	validate.RegisterStructValidation(parsers.SlotsFiltersStructValidation, parsers.SlotsFilters{})

	return validate
}
