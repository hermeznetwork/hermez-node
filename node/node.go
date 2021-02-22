package node

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
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
	"github.com/hermeznetwork/hermez-node/priceupdater"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test/debugapi"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
	"github.com/jmoiron/sqlx"
	"github.com/russross/meddler"
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
	nodeAPI      *NodeAPI
	debugAPI     *debugapi.DebugAPI
	priceUpdater *priceupdater.PriceUpdater
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
	meddler.Debug = cfg.Debug.MeddlerLogs
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
	var apiConnCon *dbUtils.APIConnectionController
	if cfg.API.Explorer || mode == ModeCoordinator {
		apiConnCon = dbUtils.NewAPICnnectionController(
			cfg.API.MaxSQLConnections,
			cfg.API.SQLConnectionTimeout.Duration,
		)
	}

	historyDB := historydb.NewHistoryDB(db, apiConnCon)

	ethClient, err := ethclient.Dial(cfg.Web3.URL)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var ethCfg eth.EthereumConfig
	var account *accounts.Account
	var keyStore *ethKeystore.KeyStore
	if mode == ModeCoordinator {
		ethCfg = eth.EthereumConfig{
			CallGasLimit: 0, // cfg.Coordinator.EthClient.CallGasLimit,
			GasPriceDiv:  0, // cfg.Coordinator.EthClient.GasPriceDiv,
		}

		scryptN := ethKeystore.StandardScryptN
		scryptP := ethKeystore.StandardScryptP
		if cfg.Coordinator.Debug.LightScrypt {
			scryptN = ethKeystore.LightScryptN
			scryptP = ethKeystore.LightScryptP
		}
		keyStore = ethKeystore.NewKeyStore(cfg.Coordinator.EthClient.Keystore.Path,
			scryptN, scryptP)

		// Unlock Coordinator ForgerAddr in the keystore to make calls
		// to ForgeBatch in the smart contract
		if !keyStore.HasAddress(cfg.Coordinator.ForgerAddress) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"ethereum keystore doesn't have the key for address %v",
				cfg.Coordinator.ForgerAddress))
		}
		account = &accounts.Account{
			Address: cfg.Coordinator.ForgerAddress,
		}
		if err := keyStore.Unlock(*account,
			cfg.Coordinator.EthClient.Keystore.Password); err != nil {
			return nil, tracerr.Wrap(err)
		}
		log.Infow("Forger ethereum account unlocked in the keystore",
			"addr", cfg.Coordinator.ForgerAddress)
	}
	client, err := eth.NewClient(ethClient, account, keyStore, &eth.ClientConfig{
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

	chainID, err := client.EthChainID()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if !chainID.IsUint64() {
		return nil, tracerr.Wrap(fmt.Errorf("chainID cannot be represented as uint64"))
	}
	chainIDU64 := chainID.Uint64()
	const maxUint16 uint64 = 0xffff
	if chainIDU64 > maxUint16 {
		return nil, tracerr.Wrap(fmt.Errorf("chainID overflows uint16"))
	}
	chainIDU16 := uint16(chainIDU64)

	const safeStateDBKeep = 128
	if cfg.StateDB.Keep < safeStateDBKeep {
		return nil, tracerr.Wrap(fmt.Errorf("cfg.StateDB.Keep = %v < %v, which is unsafe",
			cfg.StateDB.Keep, safeStateDBKeep))
	}
	stateDB, err := statedb.NewStateDB(statedb.Config{
		Path:    cfg.StateDB.Path,
		Keep:    cfg.StateDB.Keep,
		Type:    statedb.TypeSynchronizer,
		NLevels: statedb.MaxNLevels,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	sync, err := synchronizer.NewSynchronizer(client, historyDB, stateDB, synchronizer.Config{
		StatsRefreshPeriod:  cfg.Synchronizer.StatsRefreshPeriod.Duration,
		StoreAccountUpdates: cfg.Synchronizer.StoreAccountUpdates,
		ChainID:             chainIDU16,
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
			cfg.Coordinator.L2DB.MinFeeUSD,
			cfg.Coordinator.L2DB.TTL.Duration,
			apiConnCon,
		)

		// Unlock FeeAccount EthAddr in the keystore to generate the
		// account creation authorization
		if !keyStore.HasAddress(cfg.Coordinator.FeeAccount.Address) {
			return nil, tracerr.Wrap(fmt.Errorf(
				"ethereum keystore doesn't have the key for address %v",
				cfg.Coordinator.FeeAccount.Address))
		}
		feeAccount := accounts.Account{
			Address: cfg.Coordinator.FeeAccount.Address,
		}
		if err := keyStore.Unlock(feeAccount,
			cfg.Coordinator.EthClient.Keystore.Password); err != nil {
			return nil, tracerr.Wrap(err)
		}
		auth := &common.AccountCreationAuth{
			EthAddr: cfg.Coordinator.FeeAccount.Address,
			BJJ:     cfg.Coordinator.FeeAccount.BJJ,
		}
		if err := auth.Sign(func(msg []byte) ([]byte, error) {
			return keyStore.SignHash(feeAccount, msg)
		}, chainIDU16, cfg.SmartContracts.Rollup); err != nil {
			return nil, err
		}
		coordAccount := &txselector.CoordAccount{
			Addr:                cfg.Coordinator.FeeAccount.Address,
			BJJ:                 cfg.Coordinator.FeeAccount.BJJ,
			AccountCreationAuth: auth.Signature,
		}
		txSelector, err := txselector.NewTxSelector(coordAccount, cfg.Coordinator.TxSelector.Path, stateDB, l2DB)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		batchBuilder, err := batchbuilder.NewBatchBuilder(cfg.Coordinator.BatchBuilder.Path,
			stateDB, 0, uint64(cfg.Coordinator.Circuit.NLevels))
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

		txProcessorCfg := txprocessor.Config{
			NLevels:  uint32(cfg.Coordinator.Circuit.NLevels),
			MaxTx:    uint32(cfg.Coordinator.Circuit.MaxTx),
			ChainID:  chainIDU16,
			MaxFeeTx: common.RollupConstMaxFeeIdxCoordinator,
			MaxL1Tx:  common.RollupConstMaxL1Tx,
		}
		var verifierIdx int
		if cfg.Coordinator.Debug.RollupVerifierIndex == nil {
			verifierIdx, err = scConsts.Rollup.FindVerifierIdx(
				cfg.Coordinator.Circuit.MaxTx,
				cfg.Coordinator.Circuit.NLevels,
			)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			log.Infow("Found verifier that matches circuit config", "verifierIdx", verifierIdx)
		} else {
			verifierIdx = *cfg.Coordinator.Debug.RollupVerifierIndex
			log.Infow("Using debug verifier index from config", "verifierIdx", verifierIdx)
			if verifierIdx >= len(scConsts.Rollup.Verifiers) {
				return nil, tracerr.Wrap(
					fmt.Errorf("verifierIdx (%v) >= "+
						"len(scConsts.Rollup.Verifiers) (%v)",
						verifierIdx, len(scConsts.Rollup.Verifiers)))
			}
			verifier := scConsts.Rollup.Verifiers[verifierIdx]
			if verifier.MaxTx != cfg.Coordinator.Circuit.MaxTx ||
				verifier.NLevels != cfg.Coordinator.Circuit.NLevels {
				return nil, tracerr.Wrap(
					fmt.Errorf("Circuit config and verifier params don't match.  "+
						"circuit.MaxTx = %v, circuit.NLevels = %v, "+
						"verifier.MaxTx = %v, verifier.NLevels = %v",
						cfg.Coordinator.Circuit.MaxTx, cfg.Coordinator.Circuit.NLevels,
						verifier.MaxTx, verifier.NLevels,
					))
			}
		}

		coord, err = coordinator.NewCoordinator(
			coordinator.Config{
				ForgerAddress:          cfg.Coordinator.ForgerAddress,
				ConfirmBlocks:          cfg.Coordinator.ConfirmBlocks,
				L1BatchTimeoutPerc:     cfg.Coordinator.L1BatchTimeoutPerc,
				ForgeRetryInterval:     cfg.Coordinator.ForgeRetryInterval.Duration,
				ForgeDelay:             cfg.Coordinator.ForgeDelay.Duration,
				ForgeNoTxsDelay:        cfg.Coordinator.ForgeNoTxsDelay.Duration,
				SyncRetryInterval:      cfg.Coordinator.SyncRetryInterval.Duration,
				PurgeByExtDelInterval:  cfg.Coordinator.PurgeByExtDelInterval.Duration,
				EthClientAttempts:      cfg.Coordinator.EthClient.Attempts,
				EthClientAttemptsDelay: cfg.Coordinator.EthClient.AttemptsDelay.Duration,
				EthNoReuseNonce:        cfg.Coordinator.EthClient.NoReuseNonce,
				EthTxResendTimeout:     cfg.Coordinator.EthClient.TxResendTimeout.Duration,
				MaxGasPrice:            cfg.Coordinator.EthClient.MaxGasPrice,
				GasPriceIncPerc:        cfg.Coordinator.EthClient.GasPriceIncPerc,
				TxManagerCheckInterval: cfg.Coordinator.EthClient.CheckLoopInterval.Duration,
				DebugBatchPath:         cfg.Coordinator.Debug.BatchPath,
				Purger: coordinator.PurgerCfg{
					PurgeBatchDelay:      cfg.Coordinator.L2DB.PurgeBatchDelay,
					InvalidateBatchDelay: cfg.Coordinator.L2DB.InvalidateBatchDelay,
					PurgeBlockDelay:      cfg.Coordinator.L2DB.PurgeBlockDelay,
					InvalidateBlockDelay: cfg.Coordinator.L2DB.InvalidateBlockDelay,
				},
				VerifierIdx:       uint8(verifierIdx),
				TxProcessorConfig: txProcessorCfg,
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
				ChainID:           chainIDU16,
				HermezAddress:     cfg.SmartContracts.Rollup,
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
	priceUpdater, err := priceupdater.NewPriceUpdater(cfg.PriceUpdater.URL,
		priceupdater.APIType(cfg.PriceUpdater.Type), historyDB)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		nodeAPI:      nodeAPI,
		debugAPI:     debugAPI,
		priceUpdater: priceUpdater,
		coord:        coord,
		sync:         sync,
		cfg:          cfg,
		mode:         mode,
		sqlConn:      db,
		ctx:          ctx,
		cancel:       cancel,
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

func (n *Node) handleNewBlock(ctx context.Context, stats *synchronizer.Stats, vars synchronizer.SCVariablesPtr,
	batches []common.BatchData) {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(ctx, coordinator.MsgSyncBlock{
			Stats:   *stats,
			Vars:    vars,
			Batches: batches,
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
				common.BatchNum(stats.Eth.LastBatchNum),
				stats.Sync.Auction.CurrentSlot.SlotNum,
			); err != nil {
				log.Errorw("API.UpdateNetworkInfo", "err", err)
			}
		} else {
			n.nodeAPI.api.UpdateNetworkInfoBlock(
				stats.Eth.LastBlock, stats.Sync.LastBlock,
			)
		}
	}
}

func (n *Node) handleReorg(ctx context.Context, stats *synchronizer.Stats, vars synchronizer.SCVariablesPtr) {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(ctx, coordinator.MsgSyncReorg{
			Stats: *stats,
			Vars:  vars,
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
func (n *Node) syncLoopFn(ctx context.Context, lastBlock *common.Block) (*common.Block, time.Duration, error) {
	blockData, discarded, err := n.sync.Sync2(ctx, lastBlock)
	stats := n.sync.Stats()
	if err != nil {
		// case: error
		return nil, n.cfg.Synchronizer.SyncLoopInterval.Duration, tracerr.Wrap(err)
	} else if discarded != nil {
		// case: reorg
		log.Infow("Synchronizer.Sync reorg", "discarded", *discarded)
		vars := n.sync.SCVars()
		n.handleReorg(ctx, stats, vars)
		return nil, time.Duration(0), nil
	} else if blockData != nil {
		// case: new block
		vars := synchronizer.SCVariablesPtr{
			Rollup:   blockData.Rollup.Vars,
			Auction:  blockData.Auction.Vars,
			WDelayer: blockData.WDelayer.Vars,
		}
		n.handleNewBlock(ctx, stats, vars, blockData.Rollup.Batches)
		return &blockData.Block, time.Duration(0), nil
	} else {
		// case: no block
		return lastBlock, n.cfg.Synchronizer.SyncLoopInterval.Duration, nil
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
	n.handleNewBlock(n.ctx, stats, vars, []common.BatchData{})

	n.wg.Add(1)
	go func() {
		var err error
		var lastBlock *common.Block
		waitDuration := time.Duration(0)
		for {
			select {
			case <-n.ctx.Done():
				log.Info("Synchronizer done")
				n.wg.Done()
				return
			case <-time.After(waitDuration):
				if lastBlock, waitDuration, err = n.syncLoopFn(n.ctx,
					lastBlock); err != nil {
					if n.ctx.Err() != nil {
						continue
					}
					if errors.Is(err, eth.ErrBlockHashMismatchEvent) {
						log.Warnw("Synchronizer.Sync", "err", err)
					} else if errors.Is(err, synchronizer.ErrUnknownBlock) {
						log.Warnw("Synchronizer.Sync", "err", err)
					} else {
						log.Errorw("Synchronizer.Sync", "err", err)
					}
				}
			}
		}
	}()

	n.wg.Add(1)
	go func() {
		for {
			select {
			case <-n.ctx.Done():
				log.Info("PriceUpdater done")
				n.wg.Done()
				return
			case <-time.After(n.cfg.PriceUpdater.Interval.Duration):
				if err := n.priceUpdater.UpdateTokenList(); err != nil {
					log.Errorw("PriceUpdater.UpdateTokenList()", "err", err)
				}
				n.priceUpdater.UpdatePrices(n.ctx)
			}
		}
	}()
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
			if n.ctx.Err() != nil {
				return
			}
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
			if n.ctx.Err() != nil {
				return
			}
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
	n.wg.Wait()
	if n.mode == ModeCoordinator {
		log.Info("Stopping Coordinator...")
		n.coord.Stop()
	}
	// Close kv DBs
	n.sync.StateDB().Close()
	if n.mode == ModeCoordinator {
		n.coord.TxSelector().LocalAccountsDB().Close()
		n.coord.BatchBuilder().LocalStateDB().Close()
	}
}
