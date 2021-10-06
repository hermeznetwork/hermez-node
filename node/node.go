/*
Package node does the initialization of all the required objects to either run
as a synchronizer or as a coordinator.

The Node contains several goroutines that run in the background or that
periodically perform tasks.  One of this goroutines periodically calls the
`Synchronizer.Sync` function, allowing the synchronization of one block at a
time.  After every call to `Synchronizer.Sync`, the Node sends a message to the
Coordinator to notify it about the new synced block (and associated state) or
reorg (and resetted state) in case one happens.

Other goroutines perform tasks such as: updating the token prices, update
metrics stored in the historyDB, update recommended fee stored in the
historyDB, run the http API server, run the debug http API server, etc.
*/
package node

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hermeznetwork/hermez-node/api"
	"github.com/hermeznetwork/hermez-node/api/stateapiupdater"
	"github.com/hermeznetwork/hermez-node/batchbuilder"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/config"
	"github.com/hermeznetwork/hermez-node/coordinator"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/etherscan"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test/debugapi"
	"github.com/hermeznetwork/hermez-node/txprocessor"
	"github.com/hermeznetwork/hermez-node/txselector"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
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
	nodeAPI         *NodeAPI
	stateAPIUpdater *stateapiupdater.Updater
	debugAPI        *debugapi.DebugAPI
	// Coordinator
	coord *coordinator.Coordinator

	// Synchronizer
	sync *synchronizer.Synchronizer

	// General
	cfg          *config.Node
	mode         Mode
	sqlConnRead  *sqlx.DB
	sqlConnWrite *sqlx.DB
	historyDB    *historydb.HistoryDB
	ctx          context.Context
	wg           sync.WaitGroup
	cancel       context.CancelFunc
}

// NewNode creates a Node
func NewNode(mode Mode, cfg *config.Node, version string) (*Node, error) {
	meddler.Debug = cfg.Debug.MeddlerLogs
	// Stablish DB connection
	dbWrite, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.PortWrite,
		cfg.PostgreSQL.HostWrite,
		cfg.PostgreSQL.UserWrite,
		cfg.PostgreSQL.PasswordWrite,
		cfg.PostgreSQL.NameWrite,
	)
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
	}
	var dbRead *sqlx.DB
	if cfg.PostgreSQL.HostRead == "" {
		dbRead = dbWrite
	} else if cfg.PostgreSQL.HostRead == cfg.PostgreSQL.HostWrite {
		return nil, tracerr.Wrap(fmt.Errorf(
			"PostgreSQL.HostRead and PostgreSQL.HostWrite must be different",
		))
	} else {
		dbRead, err = dbUtils.InitSQLDB(
			cfg.PostgreSQL.PortRead,
			cfg.PostgreSQL.HostRead,
			cfg.PostgreSQL.UserRead,
			cfg.PostgreSQL.PasswordRead,
			cfg.PostgreSQL.NameRead,
		)
		if err != nil {
			return nil, tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
		}
	}
	var apiConnCon *dbUtils.APIConnectionController
	if cfg.API.Explorer || mode == ModeCoordinator {
		apiConnCon = dbUtils.NewAPIConnectionController(
			cfg.API.MaxSQLConnections,
			cfg.API.SQLConnectionTimeout.Duration,
		)
	}

	historyDB := historydb.NewHistoryDB(dbRead, dbWrite, apiConnCon)

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

		balance, err := ethClient.BalanceAt(context.TODO(), cfg.Coordinator.ForgerAddress, nil)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		minForgeBalance := cfg.Coordinator.MinimumForgeAddressBalance
		if minForgeBalance != nil && balance.Cmp(minForgeBalance) == -1 {
			return nil, tracerr.Wrap(fmt.Errorf(
				"forger account balance is less than cfg.Coordinator.MinimumForgeAddressBalance: %v < %v",
				balance, minForgeBalance))
		}
		log.Infow("forger ethereum account balance",
			"addr", cfg.Coordinator.ForgerAddress,
			"balance", balance,
			"minForgeBalance", minForgeBalance,
		)

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

	stateDB, err := statedb.NewStateDB(statedb.Config{
		Path:    cfg.StateDB.Path,
		Keep:    cfg.StateDB.Keep,
		Type:    statedb.TypeSynchronizer,
		NLevels: statedb.MaxNLevels,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var l2DB *l2db.L2DB
	if mode == ModeCoordinator {
		l2DB = l2db.NewL2DB(
			dbRead, dbWrite,
			cfg.Coordinator.L2DB.SafetyPeriod,
			cfg.Coordinator.L2DB.MaxTxs,
			cfg.Coordinator.L2DB.MinFeeUSD,
			cfg.Coordinator.L2DB.MaxFeeUSD,
			cfg.Coordinator.L2DB.TTL.Duration,
			apiConnCon,
		)
	}

	sync, err := synchronizer.NewSynchronizer(client, historyDB, l2DB, stateDB, synchronizer.Config{
		StatsUpdateBlockNumDiffThreshold: cfg.Synchronizer.StatsUpdateBlockNumDiffThreshold,
		StatsUpdateFrequencyDivider:      cfg.Synchronizer.StatsUpdateFrequencyDivider,
		ChainID:                          chainIDU16,
	})
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	initSCVars := sync.SCVars()

	scConsts := common.SCConsts{
		Rollup:   *sync.RollupConstants(),
		Auction:  *sync.AuctionConstants(),
		WDelayer: *sync.WDelayerConstants(),
	}

	hdbNodeCfg := historydb.NodeConfig{
		MaxPoolTxs: cfg.Coordinator.L2DB.MaxTxs,
		MinFeeUSD:  cfg.Coordinator.L2DB.MinFeeUSD,
		MaxFeeUSD:  cfg.Coordinator.L2DB.MaxFeeUSD,
		ForgeDelay: cfg.Coordinator.ForgeDelay.Duration.Seconds(),
	}
	if err := historyDB.SetNodeConfig(&hdbNodeCfg); err != nil {
		return nil, tracerr.Wrap(err)
	}
	hdbConsts := historydb.Constants{
		SCConsts: common.SCConsts{
			Rollup:   scConsts.Rollup,
			Auction:  scConsts.Auction,
			WDelayer: scConsts.WDelayer,
		},
		ChainID:       chainIDU16,
		HermezAddress: cfg.SmartContracts.Rollup,
	}
	if err := historyDB.SetConstants(&hdbConsts); err != nil {
		return nil, tracerr.Wrap(err)
	}
	var etherScanService *etherscan.Service
	if cfg.Coordinator.Etherscan.URL != "" && cfg.Coordinator.Etherscan.APIKey != "" {
		log.Info("EtherScan method detected in cofiguration file")
		etherScanService, _ = etherscan.NewEtherscanService(cfg.Coordinator.Etherscan.URL,
			cfg.Coordinator.Etherscan.APIKey)
	} else {
		log.Info("EtherScan method not configured in config file")
		etherScanService = nil
	}
	stateAPIUpdater, err := stateapiupdater.NewUpdater(
		historyDB,
		&hdbNodeCfg,
		initSCVars,
		&hdbConsts,
		&cfg.RecommendedFeePolicy,
		cfg.Coordinator.Circuit.MaxTx,
	)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var coord *coordinator.Coordinator
	if mode == ModeCoordinator {
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
		//Swap bjj endianness
		decodedBjjPubKey, err := hex.DecodeString(cfg.Coordinator.FeeAccount.BJJ.String())
		if err != nil {
			log.Error("Error decoding BJJ public key from config file. Error: ", err.Error())
			return nil, tracerr.Wrap(err)
		}
		bSwapped := common.SwapEndianness(decodedBjjPubKey)
		var bjj babyjub.PublicKeyComp
		copy(bjj[:], bSwapped[:])

		auth := &common.AccountCreationAuth{
			EthAddr: cfg.Coordinator.FeeAccount.Address,
			BJJ:     bjj,
		}
		if err := auth.Sign(func(msg []byte) ([]byte, error) {
			return keyStore.SignHash(feeAccount, msg)
		}, chainIDU16, cfg.SmartContracts.Rollup); err != nil {
			return nil, tracerr.Wrap(err)
		}
		coordAccount := txselector.CoordAccount{
			Addr:                cfg.Coordinator.FeeAccount.Address,
			BJJ:                 bjj,
			AccountCreationAuth: auth.Signature,
		}
		txSelector, err := txselector.NewTxSelector(&coordAccount,
			cfg.Coordinator.TxSelector.Path, stateDB, l2DB)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		batchBuilder, err := batchbuilder.NewBatchBuilder(cfg.Coordinator.BatchBuilder.Path,
			stateDB, 0, uint64(cfg.Coordinator.Circuit.NLevels))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		serverProofs := make([]prover.Client, len(cfg.Coordinator.ServerProofs.URLs))
		for i, serverProofCfg := range cfg.Coordinator.ServerProofs.URLs {
			serverProofs[i] = prover.NewProofServerClient(serverProofCfg,
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
				ForgerAddress:           cfg.Coordinator.ForgerAddress,
				ConfirmBlocks:           cfg.Coordinator.ConfirmBlocks,
				L1BatchTimeoutPerc:      cfg.Coordinator.L1BatchTimeoutPerc,
				ForgeRetryInterval:      cfg.Coordinator.ForgeRetryInterval.Duration,
				ForgeDelay:              cfg.Coordinator.ForgeDelay.Duration,
				MustForgeAtSlotDeadline: cfg.Coordinator.MustForgeAtSlotDeadline,
				IgnoreSlotCommitment:    cfg.Coordinator.IgnoreSlotCommitment,
				ForgeOncePerSlotIfTxs:   cfg.Coordinator.ForgeOncePerSlotIfTxs,
				ForgeNoTxsDelay:         cfg.Coordinator.ForgeNoTxsDelay.Duration,
				SyncRetryInterval:       cfg.Coordinator.SyncRetryInterval.Duration,
				PurgeByExtDelInterval:   cfg.Coordinator.PurgeByExtDelInterval.Duration,
				EthClientAttempts:       cfg.Coordinator.EthClient.Attempts,
				EthClientAttemptsDelay:  cfg.Coordinator.EthClient.AttemptsDelay.Duration,
				EthNoReuseNonce:         cfg.Coordinator.EthClient.NoReuseNonce,
				EthTxResendTimeout:      cfg.Coordinator.EthClient.TxResendTimeout.Duration,
				MaxGasPrice:             cfg.Coordinator.EthClient.MaxGasPrice,
				MinGasPrice:             cfg.Coordinator.EthClient.MinGasPrice,
				GasPriceIncPerc:         cfg.Coordinator.EthClient.GasPriceIncPerc,
				TxManagerCheckInterval:  cfg.Coordinator.EthClient.CheckLoopInterval.Duration,
				DebugBatchPath:          cfg.Coordinator.Debug.BatchPath,
				Purger: coordinator.PurgerCfg{
					PurgeBatchDelay:      cfg.Coordinator.L2DB.PurgeBatchDelay,
					InvalidateBatchDelay: cfg.Coordinator.L2DB.InvalidateBatchDelay,
					PurgeBlockDelay:      cfg.Coordinator.L2DB.PurgeBlockDelay,
					InvalidateBlockDelay: cfg.Coordinator.L2DB.InvalidateBlockDelay,
				},
				ForgeBatchGasCost: cfg.Coordinator.EthClient.ForgeBatchGasCost,
				VerifierIdx:       uint8(verifierIdx),
				TxProcessorConfig: txProcessorCfg,
				ProverReadTimeout: cfg.Coordinator.ProverWaitReadTimeout.Duration,
			},
			historyDB,
			l2DB,
			txSelector,
			batchBuilder,
			serverProofs,
			client,
			&scConsts,
			initSCVars,
			etherScanService,
		)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	var nodeAPI *NodeAPI
	if cfg.API.Address != "" {
		if cfg.Debug.GinDebugMode {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		server := gin.Default()
		server.Use(cors.Default())
		coord := false
		var coordnetConfig *api.CoordinatorNetworkConfig
		if mode == ModeCoordinator {
			coord = cfg.Coordinator.API.Coordinator
			if cfg.API.CoordinatorNetwork {
				// Setup coordinators network configuration
				// Get libp2p addresses of the registered coordinators
				// to be used as bootstrap nodes for the p2p network
				bootstrapAddrs, err := client.GetCoordinatorsLibP2PAddrs()
				if err != nil {
					log.Warn("error getting registered addresses from the SMC or no addresses registered. error:", err.Error())
				}
				// Get Ethereum private key of the coordinator
				keyJSON, err := keyStore.Export(*account, cfg.Coordinator.EthClient.Keystore.Password, cfg.Coordinator.EthClient.Keystore.Password)
				if err != nil {
					return nil, tracerr.Wrap(err)
				}
				key, err := ethKeystore.DecryptKey(keyJSON, cfg.Coordinator.EthClient.Keystore.Password)
				if err != nil {
					return nil, tracerr.Wrap(err)
				}
				coordnetConfig = &api.CoordinatorNetworkConfig{
					BootstrapPeers: bootstrapAddrs,
					EthPrivKey:     key.PrivateKey,
				}
			}
		}
		var err error
		nodeAPI, err = NewNodeAPI(cfg.API.Address, cfg.API, api.Config{
			Version:                  version,
			ExplorerEndpoints:        cfg.API.Explorer,
			CoordinatorEndpoints:     coord,
			Server:                   server,
			HistoryDB:                historyDB,
			L2DB:                     l2DB,
			StateDB:                  stateDB,
			EthClient:                ethClient,
			ForgerAddress:            &cfg.Coordinator.ForgerAddress,
			CoordinatorNetworkConfig: coordnetConfig,
		}, cfg.API.CoordinatorNetwork, cfg.API.FindPeersCoordinatorNetworkInterval.Duration)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	}
	var debugAPI *debugapi.DebugAPI
	if cfg.Debug.APIAddress != "" {
		debugAPI = debugapi.NewDebugAPI(cfg.Debug.APIAddress, stateDB, sync)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		stateAPIUpdater: stateAPIUpdater,
		nodeAPI:         nodeAPI,
		debugAPI:        debugAPI,
		coord:           coord,
		sync:            sync,
		cfg:             cfg,
		mode:            mode,
		sqlConnRead:     dbRead,
		sqlConnWrite:    dbWrite,
		historyDB:       historyDB,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// APIServer is a server that only runs the API
type APIServer struct {
	nodeAPI *NodeAPI
	mode    Mode
	ctx     context.Context
	wg      sync.WaitGroup
	cancel  context.CancelFunc
}

// NewAPIServer creates a new APIServer
func NewAPIServer(mode Mode, cfg *config.APIServer, version string, ethClient *ethclient.Client, forgerAddress *ethCommon.Address) (*APIServer, error) {
	meddler.Debug = cfg.Debug.MeddlerLogs
	// Stablish DB connection
	dbWrite, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.PortWrite,
		cfg.PostgreSQL.HostWrite,
		cfg.PostgreSQL.UserWrite,
		cfg.PostgreSQL.PasswordWrite,
		cfg.PostgreSQL.NameWrite,
	)
	if err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
	}
	var dbRead *sqlx.DB
	if cfg.PostgreSQL.HostRead == "" {
		dbRead = dbWrite
	} else {
		dbRead, err = dbUtils.InitSQLDB(
			cfg.PostgreSQL.PortRead,
			cfg.PostgreSQL.HostRead,
			cfg.PostgreSQL.UserRead,
			cfg.PostgreSQL.PasswordRead,
			cfg.PostgreSQL.NameRead,
		)
		if err != nil {
			return nil, tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
		}
	}
	apiConnCon := dbUtils.NewAPIConnectionController(
		cfg.API.MaxSQLConnections,
		cfg.API.SQLConnectionTimeout.Duration,
	)

	historyDB := historydb.NewHistoryDB(dbRead, dbWrite, apiConnCon)

	var l2DB *l2db.L2DB
	if mode == ModeCoordinator {
		l2DB = l2db.NewL2DB(
			dbRead, dbWrite,
			0,
			cfg.Coordinator.L2DB.MaxTxs,
			cfg.Coordinator.L2DB.MinFeeUSD,
			cfg.Coordinator.L2DB.MaxFeeUSD,
			0,
			apiConnCon,
		)
	}

	if cfg.Debug.GinDebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	server := gin.Default()
	server.Use(cors.Default())
	coord := false
	var coordnetConfig *api.CoordinatorNetworkConfig
	if mode == ModeCoordinator {
		coord = cfg.Coordinator.API.Coordinator
		if cfg.API.CoordinatorNetwork {
			// Prepare keystore
			scryptN := ethKeystore.StandardScryptN
			scryptP := ethKeystore.StandardScryptP
			if cfg.Coordinator.Keystore.LightScrypt {
				scryptN = ethKeystore.LightScryptN
				scryptP = ethKeystore.LightScryptP
			}
			keyStore := ethKeystore.NewKeyStore(cfg.Coordinator.Keystore.Path,
				scryptN, scryptP)
			// Unlock Coordinator ForgerAddr in the keystore to make calls
			// to ForgeBatch in the smart contract
			if forgerAddress == nil {
				return nil, errors.New("forgerAddress must be different than nil when using coordinators network")
			}
			if !keyStore.HasAddress(*forgerAddress) {
				return nil, tracerr.Wrap(fmt.Errorf(
					"ethereum keystore doesn't have the key for address %v",
					*forgerAddress),
				)
			}
			pass := cfg.Coordinator.Keystore.Password
			account := accounts.Account{
				Address: *forgerAddress,
			}
			if err := keyStore.Unlock(account, pass); err != nil {
				return nil, tracerr.Wrap(err)
			}
			// Setup eth client to read data from the blockchain
			client, err := eth.NewClient(ethClient, &account, keyStore, &eth.ClientConfig{
				Ethereum: eth.EthereumConfig{},
				Rollup: eth.RollupConfig{
					Address: cfg.Coordinator.Rollup,
				},
			})
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			// Setup coordinators network configuration
			// Get libp2p addresses of the registered coordinators
			// to be used as bootstrap nodes for the p2p network
			bootstrapAddrs, err := client.GetCoordinatorsLibP2PAddrs()
			if err != nil {
				log.Warn("error getting registered addresses from the SMC or no addresses registered. error:", err.Error())
			}
			// Get Ethereum private key of the coordinator
			keyJSON, err := keyStore.Export(account, pass, pass)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			key, err := ethKeystore.DecryptKey(keyJSON, pass)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			coordnetConfig = &api.CoordinatorNetworkConfig{
				BootstrapPeers: bootstrapAddrs,
				EthPrivKey:     key.PrivateKey,
			}
		}
	}
	nodeAPI, err := NewNodeAPI(cfg.API.Address, cfg.API, api.Config{
		Version:                  version,
		ExplorerEndpoints:        cfg.API.Explorer,
		CoordinatorEndpoints:     coord,
		Server:                   server,
		HistoryDB:                historyDB,
		L2DB:                     l2DB,
		StateDB:                  nil,
		EthClient:                ethClient,
		ForgerAddress:            forgerAddress,
		CoordinatorNetworkConfig: coordnetConfig,
	}, cfg.API.CoordinatorNetwork, cfg.API.FindPeersCoordinatorNetworkInterval.Duration)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &APIServer{
		nodeAPI: nodeAPI,
		mode:    mode,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Start the APIServer
func (s *APIServer) Start() {
	log.Infow("Starting api server...", "mode", s.mode)
	log.Info("Starting NodeAPI...")
	s.wg.Add(1)
	go func() {
		defer func() {
			log.Info("NodeAPI routine stopped")
			s.wg.Done()
		}()
		if err := s.nodeAPI.Run(s.ctx); err != nil {
			if s.ctx.Err() != nil {
				return
			}
			log.Fatalw("NodeAPI.Run", "err", err)
		}
	}()
}

// Stop the APIServer
func (s *APIServer) Stop() {
	log.Infow("Stopping NodeAPI...")
	s.cancel()
	s.wg.Wait()
}

// NodeAPI holds the node http API
type NodeAPI struct { //nolint:golint
	api                                     *api.API
	engine                                  *gin.Engine
	addr                                    string
	coordinatorNetwork                      bool
	coordinatorNetworkFindMorePeersInterval time.Duration
	readtimeout                             time.Duration
	writetimeout                            time.Duration
}

// NewNodeAPI creates a new NodeAPI (which internally calls api.NewAPI)
func NewNodeAPI(
	addr string,
	cfgAPI config.APIConfigParameters,
	apiConfig api.Config,
	coordinatorNetwork bool,
	coordinatorNetworkFindMorePeersInterval time.Duration,
) (*NodeAPI, error) {
	_api, err := api.NewAPI(apiConfig)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &NodeAPI{
		addr:                                    addr,
		api:                                     _api,
		engine:                                  apiConfig.Server,
		coordinatorNetwork:                      coordinatorNetwork,
		coordinatorNetworkFindMorePeersInterval: coordinatorNetworkFindMorePeersInterval,
		readtimeout:                             cfgAPI.Readtimeout.Duration,
		writetimeout:                            cfgAPI.Writetimeout.Duration,
	}, nil
}

// Run starts the http server of the NodeAPI.  To stop it, pass a context
// with cancellation.
func (a *NodeAPI) Run(ctx context.Context) error {
	server := &http.Server{
		Handler:        a.engine,
		ReadTimeout:    a.readtimeout,
		WriteTimeout:   a.writetimeout,
		MaxHeaderBytes: 1 << 20, //nolint:gomnd
	}
	listener, err := net.Listen("tcp", a.addr)
	if err != nil {
		return tracerr.Wrap(err)
	}
	log.Infof("NodeAPI is ready at %v", a.addr)
	go func() {
		if err := server.Serve(listener); err != nil &&
			tracerr.Unwrap(err) != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// Find more peers for coordinator network here
	if a.coordinatorNetwork && a.coordinatorNetworkFindMorePeersInterval > 0 {
		go func() {
			// Do an initial discovery on start up
			if err := a.api.FindMorePeersForCoordinatorsNetwork(); err != nil {
				log.Info("API.FindMorePeersForCoordinatorsNetwork. ", "err", err)
			}
			for {
				select {
				case <-ctx.Done():
					log.Info("API.FindMorePeersForCoordinatorsNetwork loop done")
					return
				case <-time.After(a.coordinatorNetworkFindMorePeersInterval):
					if err := a.api.FindMorePeersForCoordinatorsNetwork(); err != nil {
						log.Warnw("API.FindMorePeersForCoordinatorsNetwork. ", "err", err)
					}
				}
			}
		}()
	}

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

func (n *Node) handleNewBlock(ctx context.Context, stats *synchronizer.Stats,
	vars *common.SCVariablesPtr, batches []common.BatchData) error {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(ctx, coordinator.MsgSyncBlock{
			Stats:   *stats,
			Vars:    *vars,
			Batches: batches,
		})
	}
	n.stateAPIUpdater.SetSCVars(vars)

	/*
		When the state is out of sync, which means, the last block synchronized by the node is
		different/smaller from the last block provided by the ethereum, the network info in the state
		will not be updated. So, in order to get some information on the node state, we need
		to wait until the node finish the synchronization with the ethereum network.

		Side effects are information like lastBatch, nextForgers, metrics with zeros, defaults or null values
	*/
	if stats.Synced() {
		if err := n.stateAPIUpdater.UpdateNetworkInfo(
			stats.Eth.LastBlock, stats.Sync.LastBlock,
			common.BatchNum(stats.Eth.LastBatchNum),
			stats.Sync.Auction.CurrentSlot.SlotNum,
		); err != nil {
			log.Errorw("ApiStateUpdater.UpdateNetworkInfo", "err", err)
		}
	} else {
		n.stateAPIUpdater.UpdateNetworkInfoBlock(
			stats.Eth.LastBlock, stats.Sync.LastBlock,
		)
	}
	if err := n.stateAPIUpdater.Store(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (n *Node) handleReorg(ctx context.Context, stats *synchronizer.Stats,
	vars *common.SCVariables) error {
	if n.mode == ModeCoordinator {
		n.coord.SendMsg(ctx, coordinator.MsgSyncReorg{
			Stats: *stats,
			Vars:  *vars.AsPtr(),
		})
	}
	n.stateAPIUpdater.SetSCVars(vars.AsPtr())
	n.stateAPIUpdater.UpdateNetworkInfoBlock(
		stats.Eth.LastBlock, stats.Sync.LastBlock,
	)
	if err := n.stateAPIUpdater.Store(); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

func (n *Node) syncLoopFn(ctx context.Context, lastBlock *common.Block) (*common.Block,
	time.Duration, error) {
	blockData, discarded, err := n.sync.Sync(ctx, lastBlock)
	stats := n.sync.Stats()
	if err != nil {
		// case: error
		return nil, n.cfg.Synchronizer.SyncLoopInterval.Duration, tracerr.Wrap(err)
	} else if discarded != nil {
		// case: reorg
		log.Infow("Synchronizer.Sync reorg", "discarded", *discarded)
		vars := n.sync.SCVars()
		if err := n.handleReorg(ctx, stats, vars); err != nil {
			return nil, time.Duration(0), tracerr.Wrap(err)
		}
		return nil, time.Duration(0), nil
	} else if blockData != nil {
		// case: new block
		vars := common.SCVariablesPtr{
			Rollup:   blockData.Rollup.Vars,
			Auction:  blockData.Auction.Vars,
			WDelayer: blockData.WDelayer.Vars,
		}
		if err := n.handleNewBlock(ctx, stats, &vars, blockData.Rollup.Batches); err != nil {
			return nil, time.Duration(0), tracerr.Wrap(err)
		}
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
	if err := n.handleNewBlock(n.ctx, stats, vars.AsPtr(), []common.BatchData{}); err != nil {
		log.Fatalw("Node.handleNewBlock", "err", err)
	}

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
		// Do an initial update on startup
		if err := n.stateAPIUpdater.UpdateMetrics(); err != nil {
			log.Errorw("ApiStateUpdater.UpdateMetrics", "err", err)
		}
		if err := n.stateAPIUpdater.Store(); err != nil {
			log.Errorw("ApiStateUpdater.Store", "err", err)
		}
		for {
			select {
			case <-n.ctx.Done():
				log.Info("ApiStateUpdater.UpdateMetrics loop done")
				n.wg.Done()
				return
			case <-time.After(n.cfg.API.UpdateMetricsInterval.Duration):
				if err := n.stateAPIUpdater.UpdateMetrics(); err != nil {
					log.Errorw("ApiStateUpdater.UpdateMetrics", "err", err)
					continue
				}
				if err := n.stateAPIUpdater.Store(); err != nil {
					log.Errorw("ApiStateUpdater.Store", "err", err)
				}
			}
		}
	}()

	if n.cfg.API.CoordinatorNetwork && n.cfg.API.FindPeersCoordinatorNetworkInterval.Duration > 0 {
		n.wg.Add(1)
		go func() {
			// Do an initial discovery on start up
			if err := n.nodeAPI.api.FindMorePeersForCoordinatorsNetwork(); err != nil {
				log.Errorw("API.FindMorePeersForCoordinatorsNetwork. ", "err", err)
			}
			for {
				select {
				case <-n.ctx.Done():
					log.Info("API.FindMorePeersForCoordinatorsNetwork loop done")
					n.wg.Done()
					return
				case <-time.After(n.cfg.API.FindPeersCoordinatorNetworkInterval.Duration):
					if err := n.nodeAPI.api.FindMorePeersForCoordinatorsNetwork(); err != nil {
						log.Warnw("API.FindMorePeersForCoordinatorsNetwork. ", "err", err)
					}
				}
			}
		}()
	}

	n.wg.Add(1)
	go func() {
		// Do an initial update on startup
		if err := n.stateAPIUpdater.UpdateRecommendedFee(); err != nil {
			log.Errorw("ApiStateUpdater.UpdateRecommendedFee", "err", err)
		}
		if err := n.stateAPIUpdater.Store(); err != nil {
			log.Errorw("ApiStateUpdater.Store", "err", err)
		}
		for {
			select {
			case <-n.ctx.Done():
				log.Info("ApiStateUpdaterAPI.UpdateRecommendedFee loop done")
				n.wg.Done()
				return
			case <-time.After(n.cfg.API.UpdateRecommendedFeeInterval.Duration):
				if err := n.stateAPIUpdater.UpdateRecommendedFee(); err != nil {
					log.Errorw("ApiStateUpdaterAPI.UpdateRecommendedFee", "err", err)
					continue
				}
				if err := n.stateAPIUpdater.Store(); err != nil {
					log.Errorw("ApiStateUpdater.Store", "err", err)
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
