package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/config"
	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/node"
	"github.com/hermeznetwork/tracerr"
	"github.com/urfave/cli/v2"
)

const (
	flagCfg   = "cfg"
	flagMode  = "mode"
	flagSK    = "privatekey"
	flagYes   = "yes"
	modeSync  = "sync"
	modeCoord = "coord"
)

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
	go func() {
		for sig := range ossig {
			if sig == os.Interrupt {
				stopCh <- nil
			}
		}
	}()
	<-stopCh
	node.Stop()

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
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\nError: %v\n", tracerr.Sprint(err))
		os.Exit(1)
	}
}
