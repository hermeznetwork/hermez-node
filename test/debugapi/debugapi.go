package debugapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/ztrue/tracerr"
)

func handleNoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "404 page not found",
	})
}

type errorMsg struct {
	Message string
}

func badReq(err error, c *gin.Context) {
	log.Errorw("Bad request", "err", err)
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}

// DebugAPI is an http API with debugging endpoints
type DebugAPI struct {
	addr    string
	stateDB *statedb.StateDB // synchronizer statedb
	sync    *synchronizer.Synchronizer
}

// NewDebugAPI creates a new DebugAPI
func NewDebugAPI(addr string, stateDB *statedb.StateDB, sync *synchronizer.Synchronizer) *DebugAPI {
	return &DebugAPI{
		addr:    addr,
		stateDB: stateDB,
		sync:    sync,
	}
}

func (a *DebugAPI) handleAccount(c *gin.Context) {
	uri := struct {
		Idx uint32
	}{}
	if err := c.ShouldBindUri(&uri); err != nil {
		badReq(err, c)
		return
	}
	account, err := a.stateDB.GetAccount(common.Idx(uri.Idx))
	if err != nil {
		badReq(err, c)
		return
	}
	c.JSON(http.StatusOK, account)
}

func (a *DebugAPI) handleAccounts(c *gin.Context) {
	accounts, err := a.stateDB.GetAccounts()
	if err != nil {
		badReq(err, c)
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (a *DebugAPI) handleCurrentBatch(c *gin.Context) {
	batchNum, err := a.stateDB.GetCurrentBatch()
	if err != nil {
		badReq(err, c)
		return
	}
	c.JSON(http.StatusOK, batchNum)
}

func (a *DebugAPI) handleMTRoot(c *gin.Context) {
	root := a.stateDB.MTGetRoot()
	c.JSON(http.StatusOK, root)
}

func (a *DebugAPI) handleSyncStats(c *gin.Context) {
	stats := a.sync.Stats()
	c.JSON(http.StatusOK, stats)
}

// Run starts the http server of the DebugAPI.  To stop it, pass a context with
// cancelation (see `debugapi_test.go` for an example).
func (a *DebugAPI) Run(ctx context.Context) error {
	api := gin.Default()
	api.NoRoute(handleNoRoute)
	api.Use(cors.Default())
	debugAPI := api.Group("/debug")

	debugAPI.GET("sdb/batchnum", a.handleCurrentBatch)
	debugAPI.GET("sdb/mtroot", a.handleMTRoot)
	// Accounts returned by these endpoints will always have BatchNum = 0,
	// because the stateDB doesn't store the BatchNum in which an account
	// is created.
	debugAPI.GET("sdb/accounts", a.handleAccounts)
	debugAPI.GET("sdb/accounts/:Idx", a.handleAccount)

	debugAPI.GET("sync/stats", a.handleSyncStats)

	debugAPIServer := &http.Server{
		Addr:    a.addr,
		Handler: api,
		// Use some hardcoded numberes that are suitable for testing
		ReadTimeout:    30 * time.Second, //nolint:gomnd
		WriteTimeout:   30 * time.Second, //nolint:gomnd
		MaxHeaderBytes: 1 << 20,          //nolint:gomnd
	}
	go func() {
		log.Infof("DebugAPI is ready at %v", a.addr)
		if err := debugAPIServer.ListenAndServe(); err != nil && tracerr.Unwrap(err) != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Info("Stopping DebugAPI...")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:gomnd
	defer cancel()
	if err := debugAPIServer.Shutdown(ctxTimeout); err != nil {
		return tracerr.Wrap(err)
	}
	log.Info("DebugAPI done")
	return nil
}
