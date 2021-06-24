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

	if a.l2 != nil {
		l2DBChecker := checkers.NewCheckerWithDB(a.l2.DB().DB)
		healthHandler.AddChecker("l2DB", l2DBChecker)
	}
	if a.h != nil {
		historyDBChecker := checkers.NewCheckerWithDB(a.h.DB().DB)
		healthHandler.AddChecker("historyDB", historyDBChecker)
	}
	healthHandler.AddInfo("version", version)
	t := time.Now().UTC()
	healthHandler.AddInfo("timestamp", t)
	if ethClient != nil && forgerAddress != nil {
		balance, err := ethClient.BalanceAt(context.TODO(), *forgerAddress, nil)
		if err != nil {
			return healthHandler
		}
		healthHandler.AddInfo("coordinatorBalance", balance.String())
	}
	return healthHandler
}
