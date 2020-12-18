package node

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/config"
	"github.com/hermeznetwork/hermez-node/coordinator"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test/debugapi"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
	"github.com/jmoiron/sqlx"
)

// Mode sets the working mode of the node (synchronizer or coordinator)
type Mode string

const (
	// ModeCoordinator defines the mode of the HermezNode as Coordinator, which
	// means that the node is set to forge (which also will be synchronizing with
	// the L1 blockchain state)
	ModeCoordinator Mode = "coordinator"

	// ModeSynchronizer defines the mode of the HermezNode as Synchronizer, which
	// means that the node is set to only synchronize with the L1 blockchain state
	// and will not forge
	ModeSynchronizer Mode = "synchronizer"
)

// Node is the Hermez Node
type Node struct {
	nodeAPI  *NodeAPI
	debugAPI *debugapi.DebugAPI
	// Coordinator
	coord *coordinator.Coordinator

	// Synchronizer
	sync *synchronizer.Synchronizer

	// General
	cfg     *config.Node
	mode    Mode
	sqlConn *sqlx.DB
	ctx     context.Context
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewNode creates a Node
func NewNode(mode Mode, cfg *config.Node) (*Node, error) {
	// Stablish DB connection
	db, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Name,
	)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	historyDB := historydb.NewHistoryDB(db)

	stateDB, err := statedb.NewStateDB(cfg.StateDB.Path, statedb.TypeSynchronizer, 32)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	ethClient, err := ethclient.Dial(cfg.Web3.URL)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var ethCfg eth.EthereumConfig
	if mode == ModeCoordinator {
		ethCfg = eth.EthereumConfig{
			CallGasLimit:        cfg.Coordinator.EthClient.CallGasLimit,
			DeployGasLimit:      cfg.Coordinator.EthClient.DeployGasLimit,
			GasPriceDiv:         cfg.Coordinator.EthClient.GasPriceDiv,
			ReceiptTimeout:      cfg.Coordinator.EthClient.ReceiptTimeout.Duration,
			IntervalReceiptLoop: cfg.Coordinator.EthClient.ReceiptLoopInterval.Duration,
		}
	}
	client, err := eth.NewClient(ethClient, nil, nil, &eth.ClientConfig{
		Ethereum: ethCfg,
		Rollup: eth.RollupConfig{
			Address: cfg.SmartContracts.Rollup,
		},
		Auction: eth.AuctionConfig{
			Address: cfg.SmartContracts.Auction,
			TokenHEZ: eth.TokenConfig{
				Address: cfg.SmartContracts.TokenHEZ,
				Name:    cfg.SmartContracts.TokenHEZName,
			},
		},
		WDelayer: eth.WDelayerConfig{
			Address: cfg.SmartContracts.WDelayer,
		},
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	sync, err := synchronizer.NewSynchronizer(client, historyDB, stateDB, synchronizer.Config{
		StatsRefreshPeriod: cfg.Synchronizer.StatsRefreshPeriod.Duration,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	initSCVars := sync.SCVars()

	scConsts := synchronizer.SCConsts{
		Rollup:   *sync.RollupConstants(),
		Auction:  *sync.AuctionConstants(),
		WDelayer: *sync.WDelayerConstants(),
	}

	var coord *coordinator.Coordinator
	var l2DB *l2db.L2DB
	if mode == ModeCoordinator {
		l2DB = l2db.NewL2DB(
			db,
			cfg.Coordinator.L2DB.SafetyPeriod,
			cfg.Coordinator.L2DB.MaxTxs,
			cfg.Coordinator.L2DB.TTL.Duration,
		)
		// TODO: Get (maxL1UserTxs, maxL1OperatorTxs, maxTxs) from the smart contract
		coordAccount := &txselector.CoordAccount{ // TODO TMP
			Addr:                ethCommon.HexToAddress("0xc58d29fA6e86E4FAe04DDcEd660d45BCf3Cb2370"),
			BJJ:                 nil,
			AccountCreationAuth: nil,
		}
		txSelector, err := txselector.NewTxSelector(coordAccount, cfg.Coordinator.TxSelector.Path, stateDB, l2DB)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		// TODO: Get (configCircuits []ConfigCircuit, batchNum common.BatchNum, nLevels uint64) from smart contract
		nLevels := uint64(32) //nolint:gomnd
		batchBuilder, err := batchbuilder.NewBatchBuilder(cfg.Coordinator.BatchBuilder.Path, stateDB, nil, 0, nLevels)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		serverProofs := make([]prover.Client, len(cfg.Coordinator.ServerProofs))
		for i, serverProofCfg := range cfg.Coordinator.ServerProofs {
			serverProofs[i] = prover.NewProofServerClient(serverProofCfg.URL,
				cfg.Coordinator.ProofServerPollInterval.Duration)
		}

		coord, err = coordinator.NewCoordinator(
			coordinator.Config{
				ForgerAddress:          cfg.Coordinator.ForgerAddress,
				ConfirmBlocks:          cfg.Coordinator.ConfirmBlocks,
				L1BatchTimeoutPerc:     cfg.Coordinator.L1BatchTimeoutPerc,
				SyncRetryInterval:      cfg.Coordinator.SyncRetryInterval.Duration,
				EthClientAttempts:      cfg.Coordinator.EthClient.Attempts,
				EthClientAttemptsDelay: cfg.Coordinator.EthClient.AttemptsDelay.Duration,
				TxManagerCheckInterval: cfg.Coordinator.EthClient.CheckLoopInterval.Duration,
				DebugBatchPath:         cfg.Coordinator.Debug.BatchPath,
				Purger: coordinator.PurgerCfg{
					PurgeBatchDelay:      cfg.Coordinator.L2DB.PurgeBatchDelay,
					InvalidateBatchDelay: cfg.Coordinator.L2DB.InvalidateBatchDelay,
					PurgeBlockDelay:      cfg.Coordinator.L2DB.PurgeBlockDelay,
					InvalidateBlockDelay: cfg.Coordinator.L2DB.InvalidateBlockDelay,
				},
			},
			historyDB,
			l2DB,
			txSelector,
			batchBuilder,
			serverProofs,
			client,
			&scConsts,
			&synchronizer.SCVariables{
				Rollup:   *initSCVars.Rollup,
				Auction:  *initSCVars.Auction,
				WDelayer: *initSCVars.WDelayer,
			},
		)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	var nodeAPI *NodeAPI
	if cfg.API.Address != "" {
		if cfg.API.UpdateMetricsInterval.Duration == 0 {
			return nil, tracerr.Wrap(fmt.Errorf("invalid cfg.API.UpdateMetricsInterval: %v",
				cfg.API.UpdateMetricsInterval.Duration))
		}
		if cfg.API.UpdateRecommendedFeeInterval.Duration == 0 {
			return nil, tracerr.Wrap(fmt.Errorf("invalid cfg.API.UpdateRecommendedFeeInterval: %v",
				cfg.API.UpdateRecommendedFeeInterval.Duration))
		}
		server := gin.Default()
		coord := false
		if mode == ModeCoordinator {
			coord = cfg.Coordinator.API.Coordinator
		}
		var err error
		nodeAPI, err = NewNodeAPI(
			cfg.API.Address,
			coord, cfg.API.Explorer,
			server,
			historyDB,
			stateDB,
			l2DB,
			&api.Config{
				RollupConstants:   scConsts.Rollup,
				AuctionConstants:  scConsts.Auction,
				WDelayerConstants: scConsts.WDelayer,
			},
		)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		nodeAPI.api.SetRollupVariables(*initSCVars.Rollup)
		nodeAPI.api.SetAuctionVariables(*initSCVars.Auction)
		nodeAPI.api.SetWDelayerVariables(*initSCVars.WDelayer)
	}
	var debugAPI *debugapi.DebugAPI
	if cfg.Debug.APIAddress != "" {
		debugAPI = debugapi.NewDebugAPI(cfg.Debug.APIAddress, stateDB, sync)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		nodeAPI:  nodeAPI,
		debugAPI: debugAPI,
		coord:    coord,
		sync:     sync,
		cfg:      cfg,
		mode:     mode,
		sqlConn:  db,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// NodeAPI holds the node http API
type NodeAPI struct { //nolint:golint
	api    *api.API
	engine *gin.Engine
	addr   string
}

func handleNoRoute(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "404 page not found",
	})
}

// NewNodeAPI creates a new NodeAPI (which internally calls api.NewAPI)
func NewNodeAPI(
	addr string,
	coordinatorEndpoints, explorerEndpoints bool,
	server *gin.Engine,
	hdb *historydb.HistoryDB,
	sdb *statedb.StateDB,
	l2db *l2db.L2DB,
	config *api.Config,
) (*NodeAPI, error) {
	engine := gin.Default()
	engine.NoRoute(handleNoRoute)
	engine.Use(cors.Default())
	_api, err := api.NewAPI(
		coordinatorEndpoints, explorerEndpoints,
		engine,
		hdb,
		sdb,
		l2db,
		config,
	)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &NodeAPI{
		addr:   addr,
		api:    _api,
		engine: engine,
	}, nil
}

// Run starts the http server of the NodeAPI.  To stop it, pass a context with
// cancelation.
func (a *NodeAPI) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    a.addr,
		Handler: a.engine,
		// TODO: Figure out best parameters for production
		ReadTimeout:    30 * time.Second, //nolint:gomnd
		WriteTimeout:   30 * time.Second, //nolint:gomnd
		MaxHeaderBytes: 1 << 20,          //nolint:gomnd
	}
	go func() {
		log.Infof("NodeAPI is ready at %v", a.addr)
		if err := server.ListenAndServe(); err != nil && tracerr.Unwrap(err) != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Info("Stopping NodeAPI...")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:gomnd
	defer cancel()
	if err := server.Shutdown(ctxTimeout); err != nil {
		return tracerr.Wrap(err)
	}
	log.Info("NodeAPI done")
	return nil
}

func (n *Node) handleNewBlock(stats *synchronizer.Stats, vars synchronizer.SCVariablesPtr,
	batches []common.BatchData) {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(coordinator.MsgSyncBlock{
			Stats:   *stats,
			Batches: batches,
			Vars: synchronizer.SCVariablesPtr{
				Rollup:   vars.Rollup,
				Auction:  vars.Auction,
				WDelayer: vars.WDelayer,
			},
		})
	}
	if n.nodeAPI != nil {
		if vars.Rollup != nil {
			n.nodeAPI.api.SetRollupVariables(*vars.Rollup)
		}
		if vars.Auction != nil {
			n.nodeAPI.api.SetAuctionVariables(*vars.Auction)
		}
		if vars.WDelayer != nil {
			n.nodeAPI.api.SetWDelayerVariables(*vars.WDelayer)
		}

		if stats.Synced() {
			if err := n.nodeAPI.api.UpdateNetworkInfo(
				stats.Eth.LastBlock, stats.Sync.LastBlock,
				common.BatchNum(stats.Eth.LastBatch),
				stats.Sync.Auction.CurrentSlot.SlotNum,
			); err != nil {
				log.Errorw("API.UpdateNetworkInfo", "err", err)
			}
		}
	}
}

func (n *Node) handleReorg(stats *synchronizer.Stats) {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(coordinator.MsgSyncReorg{
			Stats: *stats,
		})
	}
	if n.nodeAPI != nil {
		vars := n.sync.SCVars()
		n.nodeAPI.api.SetRollupVariables(*vars.Rollup)
		n.nodeAPI.api.SetAuctionVariables(*vars.Auction)
		n.nodeAPI.api.SetWDelayerVariables(*vars.WDelayer)
		n.nodeAPI.api.UpdateNetworkInfoBlock(
			stats.Eth.LastBlock, stats.Sync.LastBlock,
		)
	}
}

// TODO(Edu): Consider keeping the `lastBlock` inside synchronizer so that we
// don't have to pass it around.
func (n *Node) syncLoopFn(lastBlock *common.Block) (*common.Block, time.Duration) {
	blockData, discarded, err := n.sync.Sync2(n.ctx, lastBlock)
	stats := n.sync.Stats()
	if err != nil {
		// case: error
		if strings.Contains(err.Error(), "context canceled") {
			log.Warnw("Synchronizer.Sync", "err", err)
		} else {
			log.Errorw("Synchronizer.Sync", "err", err)
		}
		return nil, n.cfg.Synchronizer.SyncLoopInterval.Duration
	} else if discarded != nil {
		// case: reorg
		log.Infow("Synchronizer.Sync reorg", "discarded", *discarded)
		n.handleReorg(stats)
		return nil, time.Duration(0)
	} else if blockData != nil {
		// case: new block
		n.handleNewBlock(stats, synchronizer.SCVariablesPtr{
			Rollup:   blockData.Rollup.Vars,
			Auction:  blockData.Auction.Vars,
			WDelayer: blockData.WDelayer.Vars,
		}, blockData.Rollup.Batches)
		return &blockData.Block, time.Duration(0)
	} else {
		// case: no block
		return lastBlock, n.cfg.Synchronizer.SyncLoopInterval.Duration
	}
}

// StartSynchronizer starts the synchronizer
func (n *Node) StartSynchronizer() {
	log.Info("Starting Synchronizer...")

	// Trigger a manual call to handleNewBlock with the loaded state of the
	// synchronizer in order to quickly activate the API and Coordinator
	// and avoid waiting for the next block.  Without this, the API and
	// Coordinator will not react until the following block (starting from
	// the last synced one) is synchronized
	stats := n.sync.Stats()
	vars := n.sync.SCVars()
	n.handleNewBlock(stats, vars, []common.BatchData{})

	n.wg.Add(1)
	go func() {
		var lastBlock *common.Block
		waitDuration := time.Duration(0)
		for {
			select {
			case <-n.ctx.Done():
				log.Info("Synchronizer done")
				n.wg.Done()
				return
			case <-time.After(waitDuration):
				lastBlock, waitDuration = n.syncLoopFn(lastBlock)
			}
		}
	}()
	// TODO: Run price updater.  This is required by the API and the TxSelector
}

// StartDebugAPI starts the DebugAPI
func (n *Node) StartDebugAPI() {
	log.Info("Starting DebugAPI...")
	n.wg.Add(1)
	go func() {
		defer func() {
			log.Info("DebugAPI routine stopped")
			n.wg.Done()
		}()
		if err := n.debugAPI.Run(n.ctx); err != nil {
			log.Fatalw("DebugAPI.Run", "err", err)
		}
	}()
}

// StartNodeAPI starts the NodeAPI
func (n *Node) StartNodeAPI() {
	log.Info("Starting NodeAPI...")
	n.wg.Add(1)
	go func() {
		defer func() {
			log.Info("NodeAPI routine stopped")
			n.wg.Done()
		}()
		if err := n.nodeAPI.Run(n.ctx); err != nil {
			log.Fatalw("NodeAPI.Run", "err", err)
		}
	}()

	n.wg.Add(1)
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				log.Info("API.UpdateMetrics loop done")
				n.wg.Done()
				return
			case <-time.After(n.cfg.API.UpdateMetricsInterval.Duration):
				if err := n.nodeAPI.api.UpdateMetrics(); err != nil {
					log.Errorw("API.UpdateMetrics", "err", err)
				}
			}
		}
	}()

	n.wg.Add(1)
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				log.Info("API.UpdateRecommendedFee loop done")
				n.wg.Done()
				return
			case <-time.After(n.cfg.API.UpdateRecommendedFeeInterval.Duration):
				if err := n.nodeAPI.api.UpdateRecommendedFee(); err != nil {
					log.Errorw("API.UpdateRecommendedFee", "err", err)
				}
			}
		}
	}()
}

// Start the node
func (n *Node) Start() {
	log.Infow("Starting node...", "mode", n.mode)
	if n.debugAPI != nil {
		n.StartDebugAPI()
	}
	if n.nodeAPI != nil {
		n.StartNodeAPI()
	}
	if n.mode == ModeCoordinator {
		log.Info("Starting Coordinator...")
		n.coord.Start()
	}
	n.StartSynchronizer()
}

// Stop the node
func (n *Node) Stop() {
	log.Infow("Stopping node...")
	n.cancel()
	if n.mode == ModeCoordinator {
		log.Info("Stopping Coordinator...")
		n.coord.Stop()
	}
	n.wg.Wait()
}
