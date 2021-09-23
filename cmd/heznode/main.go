package main

import (
	"archive/zip"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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
	flagCfg     = "cfg"
	flagMode    = "mode"
	flagSK      = "privatekey"
	flagYes     = "yes"
	flagBlock   = "block"
	modeSync    = "sync"
	modeCoord   = "coord"
	nMigrations = "nMigrations"
	flagAccount = "account"
	flagPath    = "path"
)

var (
	// version represents the program based on the git tag
	version = "v0.1.0"
	// commit represents the program based on the git commit
	commit = "dev"
	// date represents the date of application was built
	date = ""
)

func cmdVersion(*cli.Context) error {
	fmt.Printf("Version = \"%v\"\n", version)
	fmt.Printf("Build = \"%v\"\n", commit)
	fmt.Printf("Date = \"%v\"\n", date)
	return nil
}

func cmdGenBJJ(*cli.Context) error {
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
	log.Init(cfg.Log.Level, cfg.Log.Out)

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

func cmdWipeDBs(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	log.Init(cfg.Log.Level, cfg.Log.Out)
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
	if err := dbUtils.MigrationsDown(db.DB, 0); err != nil {
		return tracerr.Wrap(fmt.Errorf("dbUtils.MigrationsDown: %w", err))
	}

	log.Info("Wiping StateDBs...")
	if err := resetStateDBs(_cfg, 0); err != nil {
		return tracerr.Wrap(fmt.Errorf("resetStateDBs: %w", err))
	}
	return nil
}

func cmdSQLMigrationDown(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	log.Init(cfg.Log.Level, cfg.Log.Out)
	yes := c.Bool(flagYes)
	migrationsToRun := c.Uint(nMigrations)
	if !yes {
		fmt.Printf("*WARNING* Are you sure you want to revert "+
			"%d the SQL migrations? [y/N]: ", migrationsToRun)
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			return tracerr.Wrap(err)
		}
		input = strings.ToLower(input)
		if !(input == "y" || input == "yes") {
			return nil
		}
	}
	if migrationsToRun == 0 {
		return errors.New(nMigrations + "is set to 0, this is equivalent to use wipedbs command. If this is your intention use the other command")
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
	log.Infof("Reverting %d SQL migrations...", migrationsToRun)
	if err := dbUtils.MigrationsDown(db.DB, migrationsToRun); err != nil {
		return tracerr.Wrap(fmt.Errorf("dbUtils.MigrationsDown: %w", err))
	}
	log.Info("SQL migrations down successfully")

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
	log.Init(cfg.node.Log.Level, cfg.node.Log.Out)
	innerNode, err := node.NewNode(cfg.mode, cfg.node, c.App.Version)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error starting node: %w", err))
	}
	innerNode.Start()
	waitSigInt()
	innerNode.Stop()

	return nil
}

func cmdServeAPI(c *cli.Context) error {
	cfg, err := parseCliAPIServer(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	log.Init(cfg.server.Log.Level, cfg.server.Log.Out)
	var ethClient *ethclient.Client
	if cfg.server.API.CoordinatorNetwork {
		if cfg.server.Web3.URL == "" {
			return tracerr.New("Web3.URL required when using CoordinatorNetwork")
		}
		ethClient, err = ethclient.Dial(cfg.server.Web3.URL)
		if err != nil {
			return tracerr.Wrap(err)
		}
	}
	srv, err := node.NewAPIServer(cfg.mode, cfg.server, c.App.Version, ethClient, &cfg.server.Coordinator.ForgerAddress)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error starting api server: %w", err))
	}
	srv.Start()
	waitSigInt()
	srv.Stop()

	return nil
}

func cmdGetAccountDetails(c *cli.Context) error {
	addr, bjj, accountIdxs, err := checkAccountParam(c)
	if err != nil {
		log.Error(err)
		return nil
	}
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	log.Init(cfg.Log.Level, cfg.Log.Out)

	historyDB, err := openDBConexion(cfg)
	if err != nil {
		log.Error(err)
		return nil
	}

	obj := historydb.GetAccountsAPIRequest{
		EthAddr: addr,
		Bjj:     bjj,
	}
	var apiAccounts []historydb.AccountAPI
	if len(accountIdxs) == 0 {
		apiAccounts, _, err = historyDB.GetAccountsAPI(obj)
		if err != nil {
			return tracerr.Wrap(fmt.Errorf("historyDB.GetAccountsAPI: %w", err))
		}
	} else {
		for i := 0; i < len(accountIdxs); i++ {
			apiAccount, err := historyDB.GetAccountAPI(accountIdxs[i])
			if err != nil {
				log.Debug(fmt.Errorf("historyDB.GetAccountAPI: %w", err))
			} else {
				apiAccounts = append(apiAccounts, *apiAccount)
			}
		}
	}
	log.Infof("Found %d account(s)", len(apiAccounts))
	for index, account := range apiAccounts {
		log.Infof("Details for account %d", index)
		log.Infof(" - Account index: %s", account.Idx)
		log.Infof(" - Account BJJ  : %s", account.PublicKey)
		log.Infof(" - Balance      : %s", *account.Balance)
		log.Infof(" - HEZ Address  : %s", account.EthAddr)
		log.Infof(" - Token name   : %s", account.TokenName)
		log.Infof(" - Token symbol : %s", account.TokenSymbol)
	}
	return nil
}

func cmdDiscard(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	log.Init(cfg.Log.Level, cfg.Log.Out)
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

func cmdMakeBackup(c *cli.Context) error {
	var wg sync.WaitGroup
	// two goroutines for state db and postgres db
	goroutinesAmount := 2
	wg.Add(goroutinesAmount)
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	log.Init(cfg.Log.Level, cfg.Log.Out)
	log.Info("starting to make a hermez-node backup...")
	zipPath := c.String(flagPath)
	if _, err = os.Stat(zipPath); os.IsNotExist(err) {
		log.Infof("creating directory %s for the backup", zipPath)
		err = os.MkdirAll(zipPath, os.ModePerm)
		if err != nil {
			log.Errorf("failed to create %s directory, err: %v", zipPath, err)
			return err
		}
	}

	backupPath := path.Join(zipPath, "backup")
	err = os.MkdirAll(backupPath, os.ModePerm)
	if err != nil {
		log.Errorf("failed to create %s directory, err: %v", backupPath, err)
		return err
	}

	today := time.Now().Format("20060102150405")
	go makeDBDump(&wg, cfg, backupPath, today)
	go makeStateDBDump(&wg, cfg, backupPath)

	wg.Wait()
	log.Info("finished with making state db dump")
	log.Info("finished with dumps. started to make zip file...")
	err = createZip(backupPath, path.Join(zipPath, fmt.Sprintf("hermez-%s.zip", today)))
	if err != nil {
		log.Errorf("failed to zip %s directory, err: %v", backupPath, err)
		return err
	}
	err = os.RemoveAll(backupPath)
	if err != nil {
		log.Errorf("failed to delete tmp folder %s", backupPath)
		return err
	}
	log.Infof("backup finished! You could find zip file there: %s", zipPath)
	return nil
}

func createZip(pathToZip, destPath string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	// nolint
	defer destFile.Close()

	w := zip.NewWriter(destFile)
	// nolint
	defer w.Close()

	return filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		log.Infof("crawling: %#v", filePath)
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
		zipFile, err := w.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filepath.Clean(filePath))
		if err != nil {
			return err
		}
		// nolint
		defer fsFile.Close()
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
}

func copyDirectory(wg *sync.WaitGroup, basePath, destPath string) {
	defer func() {
		log.Infof("made a copy of %s directory of the state db", basePath)
		wg.Done()
	}()
	cmdCp := "cp"
	args := []string{"-r", basePath, destPath}
	err := exec.Command(cmdCp, args...).Run() // #nosec G204
	if err != nil {
		log.Errorf("failed to copy folder %s, err: %v", basePath, err)
	}
}

func makeStateDBDump(wg *sync.WaitGroup, cfg *config.Node, destPath string) {
	defer wg.Done()
	log.Infof("started to make state db db dump...")

	dbWrite, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.PortWrite,
		cfg.PostgreSQL.HostWrite,
		cfg.PostgreSQL.UserWrite,
		cfg.PostgreSQL.PasswordWrite,
		cfg.PostgreSQL.NameWrite,
	)
	if err != nil {
		log.Errorf("failed to connect to historydb, err: %v", err)
		return
	}

	historyDB := historydb.NewHistoryDB(dbWrite, dbWrite, nil)
	batchNum, err := historyDB.GetLastBatchNum()
	if err != nil {
		log.Errorf("failed to get last batch num from historydb, err: %v", err)
		return
	}

	dbPath := cfg.StateDB.Path
	statedbBackupPath := path.Join(destPath, "statedb")
	err = os.MkdirAll(statedbBackupPath, os.ModePerm)
	if err != nil {
		log.Errorf("failed to create %s directory, err: %v", statedbBackupPath, err)
		return
	}
	var paths []string
	paths = append(paths, path.Join(dbPath, kvdb.PathCurrent), path.Join(dbPath, kvdb.PathLast))
	for i := 0; i < 10; i++ {
		paths = append(paths, path.Join(dbPath, fmt.Sprintf("%s%d", kvdb.PathBatchNum, int(batchNum)-i)))
	}
	wg.Add(len(paths))
	for _, p := range paths {
		go copyDirectory(wg, p, statedbBackupPath)
	}
}

func makeDBDump(wg *sync.WaitGroup, cfg *config.Node, destPath, today string) {
	defer func() {
		log.Infof("finished with making postgres db dump")
		wg.Done()
	}()
	log.Infof("started to make postgres db dump...")
	outfile, err := os.Create(path.Join(destPath, fmt.Sprintf("hermezdb-dump-%s.sql", today)))
	if err != nil {
		log.Errorf("failed to create sql dump file, err: %v", err)
		return
	}
	// nolint
	defer outfile.Close()
	cmdPgDump := "pg_dump"
	args := []string{
		"--dbname",
		fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
			cfg.PostgreSQL.UserWrite, cfg.PostgreSQL.PasswordWrite, cfg.PostgreSQL.HostWrite, cfg.PostgreSQL.PortWrite, cfg.PostgreSQL.NameWrite),
	}
	cmd := exec.Command(cmdPgDump, args...) // #nosec G204
	cmd.Stdout = outfile
	if err = cmd.Run(); err != nil {
		log.Errorf("failed to run pg_dump command, err: %v", err)
	}
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
			Required: false,
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
			Usage:   "Generate a new random BabyJubJub key",
			Action:  cmdGenBJJ,
		},
		{
			Name:    "wipedbs",
			Aliases: []string{},
			Usage: "Wipe the SQL DB (HistoryDB and L2DB) and the StateDBs, " +
				"leaving the DB in a clean state",
			Action: cmdWipeDBs,
			Flags: append(flags,
				&cli.BoolFlag{
					Name:     flagYes,
					Usage:    "automatic yes to the prompt",
					Required: false,
				}),
		},
		{
			Name:    "migratesqldown",
			Aliases: []string{},
			Usage: "Revert migrations of the SQL DB (HistoryDB and L2DB), " +
				"leaving the SQL schema as in previous versions",
			Action: cmdSQLMigrationDown,
			Flags: append(flags,
				&cli.BoolFlag{
					Name:     flagYes,
					Usage:    "automatic yes to the prompt",
					Required: false,
				},
				&cli.UintFlag{
					Name:     nMigrations,
					Usage:    "amount of migrations to be reverted",
					Required: true,
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
		{
			Name:    "accountInfo",
			Aliases: []string{},
			Usage:   "get information about the specified account",
			Action:  cmdGetAccountDetails,
			Flags: append(flags,
				&cli.StringFlag{
					Name:     flagAccount,
					Usage:    "account address in hex",
					Required: true,
				}),
		},
		{
			Name:    "backup",
			Aliases: []string{},
			Usage:   "Make the backup for postgres and statedb",
			Action:  cmdMakeBackup,
			Flags: append(flags,
				&cli.StringFlag{
					Name:     flagPath,
					Usage:    "path for saving backup",
					Required: true,
				}),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", tracerr.Sprint(err))
		os.Exit(1)
	}
}

func checkAccountParam(c *cli.Context) (*ethCommon.Address, *babyjub.PublicKeyComp, []common.Idx, error) {
	accountParam := c.String(flagAccount)
	const characters = 42
	var (
		addr        *ethCommon.Address
		accountIdxs []common.Idx
		bjj         *babyjub.PublicKeyComp
		err         error
	)
	matchIdx, _ := regexp.MatchString("^\\d+$", accountParam)
	if strings.HasPrefix(accountParam, "0x") { //Check ethereum address
		addr, err = common.HezStringToEthAddr("hez:"+accountParam, "hezEthereumAddress")
		if err != nil {
			return nil, nil, nil, err
		}
	} else if len(accountParam) > characters { //Check internal hermez account address
		bjj, err = common.HezStringToBJJ(accountParam, "BJJ")
		if err != nil {
			return nil, nil, nil, err
		}
	} else if matchIdx { //Check tokenID
		value, _ := strconv.Atoi(accountParam)
		accountIdxs = append(accountIdxs, (common.Idx)(value))
	} else {
		return nil, nil, nil, fmt.Errorf("invalid parameter. Only accepted ethereum address, bjj address or account index")
	}
	return addr, bjj, accountIdxs, nil
}

func openDBConexion(cfg *config.Node) (*historydb.HistoryDB, error) {
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
	apiConnCon := dbUtils.NewAPIConnectionController(
		cfg.API.MaxSQLConnections,
		cfg.API.SQLConnectionTimeout.Duration,
	)
	historyDB := historydb.NewHistoryDB(dbRead, dbWrite, apiConnCon)
	return historyDB, nil
}
