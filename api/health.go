package api

import (
	"context"
	"net/http"
	"time"

	"github.com/dimiro1/health"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/health/checkers"
)

func (a *API) healthRoute(version string, ethClient *ethclient.Client, forgerAddress *ethCommon.Address) http.Handler {
	// taking two checkers for one db in case that in
	// the future there will be two separated dbs
	healthHandler := health.NewHandler()

	if a.l2DB != nil {
		l2DBChecker := checkers.NewCheckerWithDB(a.l2DB.DB().DB)
		healthHandler.AddChecker("l2DB", l2DBChecker)
	}
	if a.historyDB != nil {
		historyDBChecker := checkers.NewCheckerWithDB(a.historyDB.DB().DB)
		healthHandler.AddChecker("historyDB", historyDBChecker)
	}
	if a.stateDB != nil {
		stateDBChecker := checkers.NewStateDBChecker(a.stateDB)
		healthHandler.AddChecker("stateDB", stateDBChecker)
	}
	healthHandler.AddInfo("version", version)
	t := time.Now().UTC()
	healthHandler.AddInfo("timestamp", t)
	if ethClient != nil && forgerAddress != nil {
		balance, err := ethClient.BalanceAt(context.TODO(), *forgerAddress, nil)
		if err != nil {
			return healthHandler
		}
		healthHandler.AddInfo("coordinatorForgerBalance", balance.String())
	}
	return healthHandler
}
