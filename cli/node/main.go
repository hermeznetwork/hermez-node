package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/config"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/kvdb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/node"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v2"
)

const (
	flagCfg   = "cfg"
	flagMode  = "mode"
	flagSK    = "privatekey"
	flagYes   = "yes"
	flagBlock = "block"
	modeSync  = "sync"
	modeCoord = "coord"
)

var (
	// version represents the program based on the git tag
	version = "v0.1.0"
	// commit represents the program based on the git commit
	commit = "dev"
	// date represents the date of application was built
	date = ""
)

func cmdVersion(c *cli.Context) error {
	fmt.Printf("Version = \"%v\"\n", version)
	fmt.Printf("Build = \"%v\"\n", commit)
	fmt.Printf("Date = \"%v\"\n", date)
	return nil
}

func cmdGenBJJ(c *cli.Context) error {
	sk := babyjub.NewRandPrivKey()
	skBuf := [32]byte(sk)
	pk := sk.Public()
	fmt.Printf("BJJ = \"0x%s\"\n", pk.String())
	fmt.Printf("BJJPrivateKey = \"0x%s\"\n", hex.EncodeToString(skBuf[:]))
	return nil
}

func cmdImportKey(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	if _cfg.mode != node.ModeCoordinator {
		return tracerr.Wrap(fmt.Errorf("importkey must use mode coordinator"))
	}
	cfg := _cfg.node

	scryptN := ethKeystore.StandardScryptN
	scryptP := ethKeystore.StandardScryptP
	if cfg.Coordinator.Debug.LightScrypt {
		scryptN = ethKeystore.LightScryptN
		scryptP = ethKeystore.LightScryptP
	}
	keyStore := ethKeystore.NewKeyStore(cfg.Coordinator.EthClient.Keystore.Path,
		scryptN, scryptP)
	hexKey := c.String(flagSK)
	hexKey = strings.TrimPrefix(hexKey, "0x")
	sk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return tracerr.Wrap(err)
	}
	acc, err := keyStore.ImportECDSA(sk, cfg.Coordinator.EthClient.Keystore.Password)
	if err != nil {
		return tracerr.Wrap(err)
	}
	log.Infow("Imported private key", "addr", acc.Address.Hex())
	return nil
}

func resetStateDBs(cfg *Config, batchNum common.BatchNum) error {
	log.Infof("Reset Synchronizer StateDB to batchNum %v...", batchNum)

	// Manually make a checkpoint from batchNum to current to force current
	// to be a valid checkpoint.  This is useful because in case of a
	// crash, current can be corrupted and the first thing that
	// `kvdb.NewKVDB` does is read the current checkpoint, which wouldn't
	// succeed in case of corruption.
	dbPath := cfg.node.StateDB.Path
	source := path.Join(dbPath, fmt.Sprintf("%s%d", kvdb.PathBatchNum, batchNum))
	current := path.Join(dbPath, kvdb.PathCurrent)
	last := path.Join(dbPath, kvdb.PathLast)
	if err := os.RemoveAll(last); err != nil {
		return tracerr.Wrap(fmt.Errorf("os.RemoveAll: %w", err))
	}
	if batchNum == 0 {
		if err := os.RemoveAll(current); err != nil {
			return tracerr.Wrap(fmt.Errorf("os.RemoveAll: %w", err))
		}
	} else {
		if err := kvdb.PebbleMakeCheckpoint(source, current); err != nil {
			return tracerr.Wrap(fmt.Errorf("kvdb.PebbleMakeCheckpoint: %w", err))
		}
	}
	db, err := kvdb.NewKVDB(kvdb.Config{
		Path:        dbPath,
		NoGapsCheck: true,
		NoLast:      true,
	})
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("kvdb.NewKVDB: %w", err))
	}
	if err := db.Reset(batchNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("db.Reset: %w", err))
	}

	if cfg.mode == node.ModeCoordinator {
		log.Infof("Wipe Coordinator StateDBs...")

		// We wipe the Coordinator StateDBs entirely (by deleting
		// current and resetting to batchNum 0) because the Coordinator
		// StateDBs are always reset from Synchronizer when the
		// coordinator pipeline starts.
		dbPath := cfg.node.Coordinator.TxSelector.Path
		current := path.Join(dbPath, kvdb.PathCurrent)
		if err := os.RemoveAll(current); err != nil {
			return tracerr.Wrap(fmt.Errorf("os.RemoveAll: %w", err))
		}
		db, err := kvdb.NewKVDB(kvdb.Config{
			Path:        dbPath,
			NoGapsCheck: true,
			NoLast:      true,
		})
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("kvdb.NewKVDB: %w", err))
		}
		if err := db.Reset(0); err != nil {
			return tracerr.Wrap(fmt.Errorf("db.Reset: %w", err))
		}

		dbPath = cfg.node.Coordinator.BatchBuilder.Path
		current = path.Join(dbPath, kvdb.PathCurrent)
		if err := os.RemoveAll(current); err != nil {
			return tracerr.Wrap(fmt.Errorf("os.RemoveAll: %w", err))
		}
		db, err = kvdb.NewKVDB(kvdb.Config{
			Path:        dbPath,
			NoGapsCheck: true,
			NoLast:      true,
		})
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("statedb.NewKVDB: %w", err))
		}
		if err := db.Reset(0); err != nil {
			return tracerr.Wrap(fmt.Errorf("db.Reset: %w", err))
		}
	}
	return nil
}

func cmdWipeSQL(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	yes := c.Bool(flagYes)
	if !yes {
		fmt.Print("*WARNING* Are you sure you want to delete " +
			"the SQL DB and StateDBs? [y/N]: ")
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			return tracerr.Wrap(err)
		}
		input = strings.ToLower(input)
		if !(input == "y" || input == "yes") {
			return nil
		}
	}
	db, err := dbUtils.ConnectSQLDB(
		cfg.PostgreSQL.PortWrite,
		cfg.PostgreSQL.HostWrite,
		cfg.PostgreSQL.UserWrite,
		cfg.PostgreSQL.PasswordWrite,
		cfg.PostgreSQL.NameWrite,
	)
	if err != nil {
		return tracerr.Wrap(err)
	}
	log.Info("Wiping SQL DB...")
	if err := dbUtils.MigrationsDown(db.DB); err != nil {
		return tracerr.Wrap(fmt.Errorf("dbUtils.MigrationsDown: %w", err))
	}

	log.Info("Wiping StateDBs...")
	if err := resetStateDBs(_cfg, 0); err != nil {
		return tracerr.Wrap(fmt.Errorf("resetStateDBs: %w", err))
	}
	return nil
}

func waitSigInt() {
	stopCh := make(chan interface{})

	// catch ^C to send the stop signal
	ossig := make(chan os.Signal, 1)
	signal.Notify(ossig, os.Interrupt)
	const forceStopCount = 3
	go func() {
		n := 0
		for sig := range ossig {
			if sig == os.Interrupt {
				log.Info("Received Interrupt Signal")
				stopCh <- nil
				n++
				if n == forceStopCount {
					log.Fatalf("Received %v Interrupt Signals", forceStopCount)
				}
			}
		}
	}()
	<-stopCh
}

func cmdRun(c *cli.Context) error {
	cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	node, err := node.NewNode(cfg.mode, cfg.node)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error starting node: %w", err))
	}
	node.Start()
	waitSigInt()
	node.Stop()

	return nil
}

func cmdServeAPI(c *cli.Context) error {
	cfg, err := parseCliAPIServer(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	srv, err := node.NewAPIServer(cfg.mode, cfg.server)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error starting api server: %w", err))
	}
	srv.Start()
	waitSigInt()
	srv.Stop()

	return nil
}

func cmdDiscard(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	blockNum := c.Int64(flagBlock)
	log.Infof("Discarding all blocks up to block %v...", blockNum)

	dbWrite, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.PortWrite,
		cfg.PostgreSQL.HostWrite,
		cfg.PostgreSQL.UserWrite,
		cfg.PostgreSQL.PasswordWrite,
		cfg.PostgreSQL.NameWrite,
	)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
	}
	var dbRead *sqlx.DB
	if cfg.PostgreSQL.HostRead == "" {
		dbRead = dbWrite
	} else if cfg.PostgreSQL.HostRead == cfg.PostgreSQL.HostWrite {
		return tracerr.Wrap(fmt.Errorf(
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
			return tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
		}
	}
	historyDB := historydb.NewHistoryDB(dbRead, dbWrite, nil)
	if err := historyDB.Reorg(blockNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("historyDB.Reorg: %w", err))
	}
	batchNum, err := historyDB.GetLastBatchNum()
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("historyDB.GetLastBatchNum: %w", err))
	}
	l2DB := l2db.NewL2DB(
		dbRead, dbWrite,
		cfg.Coordinator.L2DB.SafetyPeriod,
		cfg.Coordinator.L2DB.MaxTxs,
		cfg.Coordinator.L2DB.MinFeeUSD,
		cfg.Coordinator.L2DB.MaxFeeUSD,
		cfg.Coordinator.L2DB.TTL.Duration,
		nil,
	)
	if err := l2DB.Reorg(batchNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("l2DB.Reorg: %w", err))
	}

	log.Info("Resetting StateDBs...")
	if err := resetStateDBs(_cfg, batchNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("resetStateDBs: %w", err))
	}

	return nil
}

// Config is the configuration of the hermez node execution
type Config struct {
	mode node.Mode
	node *config.Node
}

func parseCli(c *cli.Context) (*Config, error) {
	cfg, err := getConfig(c)
	if err != nil {
		if err := cli.ShowAppHelp(c); err != nil {
			panic(err)
		}
		return nil, tracerr.Wrap(err)
	}
	return cfg, nil
}

func getConfig(c *cli.Context) (*Config, error) {
	var cfg Config
	mode := c.String(flagMode)
	nodeCfgPath := c.String(flagCfg)
	var err error
	switch mode {
	case modeSync:
		cfg.mode = node.ModeSynchronizer
		cfg.node, err = config.LoadNode(nodeCfgPath, false)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	case modeCoord:
		cfg.mode = node.ModeCoordinator
		fmt.Println("LOADING CFG")
		cfg.node, err = config.LoadNode(nodeCfgPath, true)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	default:
		return nil, tracerr.Wrap(fmt.Errorf("invalid mode \"%v\"", mode))
	}

	return &cfg, nil
}

// ConfigAPIServer is the configuration of the api server execution
type ConfigAPIServer struct {
	mode   node.Mode
	server *config.APIServer
}

func parseCliAPIServer(c *cli.Context) (*ConfigAPIServer, error) {
	cfg, err := getConfigAPIServer(c)
	if err != nil {
		if err := cli.ShowAppHelp(c); err != nil {
			panic(err)
		}
		return nil, tracerr.Wrap(err)
	}
	return cfg, nil
}

func getConfigAPIServer(c *cli.Context) (*ConfigAPIServer, error) {
	var cfg ConfigAPIServer
	mode := c.String(flagMode)
	nodeCfgPath := c.String(flagCfg)
	var err error
	switch mode {
	case modeSync:
		cfg.mode = node.ModeSynchronizer
		cfg.server, err = config.LoadAPIServer(nodeCfgPath, false)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	case modeCoord:
		cfg.mode = node.ModeCoordinator
		cfg.server, err = config.LoadAPIServer(nodeCfgPath, true)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	default:
		return nil, tracerr.Wrap(fmt.Errorf("invalid mode \"%v\"", mode))
	}

	return &cfg, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "hermez-node"
	app.Version = version
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     flagMode,
			Usage:    fmt.Sprintf("Set node `MODE` (can be \"%v\" or \"%v\")", modeSync, modeCoord),
			Required: true,
		},
		&cli.StringFlag{
			Name:     flagCfg,
			Usage:    "Node configuration `FILE`",
			Required: true,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Show the application version and build",
			Action:  cmdVersion,
		},
		{
			Name:    "importkey",
			Aliases: []string{},
			Usage:   "Import ethereum private key",
			Action:  cmdImportKey,
			Flags: append(flags,
				&cli.StringFlag{
					Name:     flagSK,
					Usage:    "ethereum `PRIVATE_KEY` in hex",
					Required: true,
				}),
		},
		{
			Name:    "genbjj",
			Aliases: []string{},
			Usage:   "Generate a new BabyJubJub key",
			Action:  cmdGenBJJ,
		},
		{
			Name:    "wipesql",
			Aliases: []string{},
			Usage: "Wipe the SQL DB (HistoryDB and L2DB) and the StateDBs, " +
				"leaving the DB in a clean state",
			Action: cmdWipeSQL,
			Flags: append(flags,
				&cli.BoolFlag{
					Name:     flagYes,
					Usage:    "automatic yes to the prompt",
					Required: false,
				}),
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the hermez-node in the indicated mode",
			Action:  cmdRun,
			Flags:   flags,
		},
		{
			Name:    "serveapi",
			Aliases: []string{},
			Usage:   "Serve the API only",
			Action:  cmdServeAPI,
			Flags:   flags,
		},
		{
			Name:    "discard",
			Aliases: []string{},
			Usage:   "Discard blocks up to a specified block number",
			Action:  cmdDiscard,
			Flags: append(flags,
				&cli.Int64Flag{
					Name:     flagBlock,
					Usage:    "last block number to keep",
					Required: false,
				}),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", tracerr.Sprint(err))
		os.Exit(1)
	}
}
