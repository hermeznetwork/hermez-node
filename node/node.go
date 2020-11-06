package node

import (
	"context"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
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
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test/debugapi"
	"github.com/hermeznetwork/hermez-node/txselector"
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
	debugAPI *debugapi.DebugAPI
	// Coordinator
	coord                    *coordinator.Coordinator
	coordCfg                 *config.Coordinator
	stopForge                chan bool
	stopGetProofCallForge    chan bool
	stopForgeCallConfirm     chan bool
	stoppedForge             chan bool
	stoppedGetProofCallForge chan bool
	stoppedForgeCallConfirm  chan bool

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
func NewNode(mode Mode, cfg *config.Node, coordCfg *config.Coordinator) (*Node, error) {
	// Stablish DB connection
	db, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Name,
	)
	if err != nil {
		return nil, err
	}

	historyDB := historydb.NewHistoryDB(db)

	stateDB, err := statedb.NewStateDB(cfg.StateDB.Path, statedb.TypeSynchronizer, 32)
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(cfg.Web3.URL)
	if err != nil {
		return nil, err
	}
	client, err := eth.NewClient(ethClient, nil, nil, &eth.ClientConfig{
		Ethereum: eth.EthereumConfig{
			CallGasLimit:        cfg.EthClient.CallGasLimit,
			DeployGasLimit:      cfg.EthClient.DeployGasLimit,
			GasPriceDiv:         cfg.EthClient.GasPriceDiv,
			ReceiptTimeout:      cfg.EthClient.ReceiptTimeout.Duration,
			IntervalReceiptLoop: cfg.EthClient.IntervalReceiptLoop.Duration,
		},
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
		return nil, err
	}

	sync, err := synchronizer.NewSynchronizer(client, historyDB, stateDB, synchronizer.Config{
		StartBlockNum:    cfg.Synchronizer.StartBlockNum,
		InitialVariables: cfg.Synchronizer.InitialVariables,
	})
	if err != nil {
		return nil, err
	}

	var coord *coordinator.Coordinator
	if mode == ModeCoordinator {
		l2DB := l2db.NewL2DB(
			db,
			coordCfg.L2DB.SafetyPeriod,
			coordCfg.L2DB.MaxTxs,
			coordCfg.L2DB.TTL.Duration,
		)
		// TODO: Get (maxL1UserTxs, maxL1OperatorTxs, maxTxs) from the smart contract
		txSelector, err := txselector.NewTxSelector(coordCfg.TxSelector.Path, stateDB, l2DB, 10, 10, 10)
		if err != nil {
			return nil, err
		}
		// TODO: Get (configCircuits []ConfigCircuit, batchNum common.BatchNum, nLevels uint64) from smart contract
		nLevels := uint64(32) //nolint:gomnd
		batchBuilder, err := batchbuilder.NewBatchBuilder(coordCfg.BatchBuilder.Path, stateDB, nil, 0, nLevels)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		serverProofs := make([]coordinator.ServerProofInterface, len(coordCfg.ServerProofs))
		for i, serverProofCfg := range coordCfg.ServerProofs {
			serverProofs[i] = coordinator.NewServerProof(serverProofCfg.URL)
		}
		coord = coordinator.NewCoordinator(
			coordinator.Config{
				ForgerAddress: coordCfg.ForgerAddress,
			},
			historyDB,
			txSelector,
			batchBuilder,
			serverProofs,
			client,
		)
	}
	var debugAPI *debugapi.DebugAPI
	println("apiaddr", cfg.Debug.APIAddress)
	if cfg.Debug.APIAddress != "" {
		debugAPI = debugapi.NewDebugAPI(cfg.Debug.APIAddress, stateDB)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		debugAPI: debugAPI,
		coord:    coord,
		coordCfg: coordCfg,
		sync:     sync,
		cfg:      cfg,
		mode:     mode,
		sqlConn:  db,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// StartCoordinator starts the coordinator
func (n *Node) StartCoordinator() {
	log.Info("Starting Coordinator...")

	// TODO: Replace stopXXX by context
	// TODO: Replace stoppedXXX by waitgroup

	n.stopForge = make(chan bool)
	n.stopGetProofCallForge = make(chan bool)
	n.stopForgeCallConfirm = make(chan bool)

	n.stoppedForge = make(chan bool, 1)
	n.stoppedGetProofCallForge = make(chan bool, 1)
	n.stoppedForgeCallConfirm = make(chan bool, 1)

	queueSize := 1
	batchCh0 := make(chan *coordinator.BatchInfo, queueSize)
	batchCh1 := make(chan *coordinator.BatchInfo, queueSize)

	go func() {
		defer func() { n.stoppedForge <- true }()
		for {
			select {
			case <-n.stopForge:
				return
			default:
				if forge, err := n.coord.ForgeLoopFn(batchCh0, n.stopForge); err == coordinator.ErrStop {
					return
				} else if err != nil {
					log.Errorw("Coordinator.ForgeLoopFn", "error", err)
				} else if !forge {
					time.Sleep(n.coordCfg.ForgeLoopInterval.Duration)
				}
			}
		}
	}()
	go func() {
		defer func() { n.stoppedGetProofCallForge <- true }()
		for {
			select {
			case <-n.stopGetProofCallForge:
				return
			default:
				if err := n.coord.GetProofCallForgeLoopFn(
					batchCh0, batchCh1, n.stopGetProofCallForge); err == coordinator.ErrStop {
					return
				} else if err != nil {
					log.Errorw("Coordinator.GetProofCallForgeLoopFn", "error", err)
				}
			}
		}
	}()
	go func() {
		defer func() { n.stoppedForgeCallConfirm <- true }()
		for {
			select {
			case <-n.stopForgeCallConfirm:
				return
			default:
				if err := n.coord.ForgeCallConfirmLoopFn(
					batchCh1, n.stopForgeCallConfirm); err == coordinator.ErrStop {
					return
				} else if err != nil {
					log.Errorw("Coordinator.ForgeCallConfirmLoopFn", "error", err)
				}
			}
		}
	}()
}

// StopCoordinator stops the coordinator
func (n *Node) StopCoordinator() {
	log.Info("Stopping Coordinator...")
	n.stopForge <- true
	n.stopGetProofCallForge <- true
	n.stopForgeCallConfirm <- true
	<-n.stoppedForge
	<-n.stoppedGetProofCallForge
	<-n.stoppedForgeCallConfirm
}

// StartSynchronizer starts the synchronizer
func (n *Node) StartSynchronizer() {
	log.Info("Starting Synchronizer...")
	n.wg.Add(1)
	go func() {
		defer func() {
			log.Info("Synchronizer routine stopped")
			n.wg.Done()
		}()
		var lastBlock *common.Block
		d := time.Duration(0)
		for {
			select {
			case <-n.ctx.Done():
				log.Info("Synchronizer done")
				return
			case <-time.After(d):
				if blockData, discarded, err := n.sync.Sync2(n.ctx, lastBlock); err != nil {
					log.Errorw("Synchronizer.Sync", "error", err)
					lastBlock = nil
					d = n.cfg.Synchronizer.SyncLoopInterval.Duration
				} else if discarded != nil {
					log.Infow("Synchronizer.Sync reorg", "discarded", *discarded)
					lastBlock = nil
					d = time.Duration(0)
				} else if blockData != nil {
					lastBlock = &blockData.Block
					d = time.Duration(0)
				} else {
					d = n.cfg.Synchronizer.SyncLoopInterval.Duration
				}
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

// Start the node
func (n *Node) Start() {
	log.Infow("Starting node...", "mode", n.mode)
	if n.debugAPI != nil {
		n.StartDebugAPI()
	}
	if n.mode == ModeCoordinator {
		n.StartCoordinator()
	}
	n.StartSynchronizer()
}

// Stop the node
func (n *Node) Stop() {
	log.Infow("Stopping node...")
	n.cancel()
	if n.mode == ModeCoordinator {
		n.StopCoordinator()
	}
	n.wg.Wait()
}
