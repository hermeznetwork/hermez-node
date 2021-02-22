package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/config"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/node"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
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

func cmdWipeSQL(c *cli.Context) error {
	_cfg, err := parseCli(c)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("error parsing flags and config: %w", err))
	}
	cfg := _cfg.node
	yes := c.Bool(flagYes)
	if !yes {
		fmt.Print("*WARNING* Are you sure you want to delete the SQL DB? [y/N]: ")
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
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Name,
	)
	if err != nil {
		return tracerr.Wrap(err)
	}
	log.Info("Wiping SQL DB...")
	if err := dbUtils.MigrationsDown(db.DB); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
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
	node.Stop()

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

	db, err := dbUtils.InitSQLDB(
		cfg.PostgreSQL.Port,
		cfg.PostgreSQL.Host,
		cfg.PostgreSQL.User,
		cfg.PostgreSQL.Password,
		cfg.PostgreSQL.Name,
	)
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("dbUtils.InitSQLDB: %w", err))
	}
	historyDB := historydb.NewHistoryDB(db, nil)
	if err := historyDB.Reorg(blockNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("historyDB.Reorg: %w", err))
	}
	batchNum, err := historyDB.GetLastBatchNum()
	if err != nil {
		return tracerr.Wrap(fmt.Errorf("historyDB.GetLastBatchNum: %w", err))
	}
	l2DB := l2db.NewL2DB(
		db,
		cfg.Coordinator.L2DB.SafetyPeriod,
		cfg.Coordinator.L2DB.MaxTxs,
		cfg.Coordinator.L2DB.MinFeeUSD,
		cfg.Coordinator.L2DB.TTL.Duration,
		nil,
	)
	if err := l2DB.Reorg(batchNum); err != nil {
		return tracerr.Wrap(fmt.Errorf("l2DB.Reorg: %w", err))
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
	if nodeCfgPath == "" {
		return nil, tracerr.Wrap(fmt.Errorf("required flag \"%v\" not set", flagCfg))
	}
	var err error
	switch mode {
	case modeSync:
		cfg.mode = node.ModeSynchronizer
		cfg.node, err = config.LoadNode(nodeCfgPath)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
	case modeCoord:
		cfg.mode = node.ModeCoordinator
		cfg.node, err = config.LoadCoordinator(nodeCfgPath)
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
	app.Version = "0.1.0-alpha"
	app.Flags = []cli.Flag{
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
			Name:    "importkey",
			Aliases: []string{},
			Usage:   "Import ethereum private key",
			Action:  cmdImportKey,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagSK,
					Usage:    "ethereum `PRIVATE_KEY` in hex",
					Required: true,
				}},
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
			Usage: "Wipe the SQL DB (HistoryDB and L2DB), " +
				"leaving the DB in a clean state",
			Action: cmdWipeSQL,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:     flagYes,
					Usage:    "automatic yes to the prompt",
					Required: false,
				}},
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the hermez-node in the indicated mode",
			Action:  cmdRun,
		},
		{
			Name:    "discard",
			Aliases: []string{},
			Usage:   "Discard blocks up to a specified block number",
			Action:  cmdDiscard,
			Flags: []cli.Flag{
				&cli.Int64Flag{
					Name:     flagBlock,
					Usage:    "last block number to keep",
					Required: false,
				}},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", tracerr.Sprint(err))
		os.Exit(1)
	}
}
