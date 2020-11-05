package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/hermeznetwork/hermez-node/config"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/node"
	"github.com/urfave/cli/v2"
)

const (
	flagCfg      = "cfg"
	flagCoordCfg = "coordcfg"
	flagMode     = "mode"
	modeSync     = "sync"
	modeCoord    = "coord"
)

func cmdInit(c *cli.Context) error {
	log.Info("Init")
	cfg, err := parseCli(c)
	if err != nil {
		return err
	}
	fmt.Println("TODO", cfg)
	return err
}

func cmdRun(c *cli.Context) error {
	cfg, err := parseCli(c)
	if err != nil {
		return fmt.Errorf("error parsing flags and config: %w", err)
	}
	node, err := node.NewNode(cfg.mode, cfg.node, cfg.coord)
	if err != nil {
		return fmt.Errorf("error starting node: %w", err)
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
	mode  node.Mode
	node  *config.Node
	coord *config.Coordinator
}

func parseCli(c *cli.Context) (*Config, error) {
	cfg, err := getConfig(c)
	if err != nil {
		if err := cli.ShowAppHelp(c); err != nil {
			panic(err)
		}
		return nil, err
	}
	return cfg, nil
}

func getConfig(c *cli.Context) (*Config, error) {
	var cfg Config
	mode := c.String(flagMode)
	switch mode {
	case modeSync:
		cfg.mode = node.ModeSynchronizer
	case modeCoord:
		cfg.mode = node.ModeCoordinator
	default:
		return nil, fmt.Errorf("invalid mode \"%v\"", mode)
	}

	if cfg.mode == node.ModeCoordinator {
		coordCfgPath := c.String(flagCoordCfg)
		if coordCfgPath == "" {
			return nil, fmt.Errorf("required flag \"%v\" not set", flagCoordCfg)
		}
		coordCfg, err := config.LoadCoordinator(coordCfgPath)
		if err != nil {
			return nil, err
		}
		cfg.coord = coordCfg
	}
	nodeCfgPath := c.String(flagCfg)
	if nodeCfgPath == "" {
		return nil, fmt.Errorf("required flag \"%v\" not set", flagCfg)
	}
	nodeCfg, err := config.LoadNode(nodeCfgPath)
	if err != nil {
		return nil, err
	}
	// nodeCfg.Synchronizer.InitialVariables.WDelayer.HermezRollupAddress = nodeCfg.SmartContracts.Rollup
	cfg.node = nodeCfg

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
		&cli.StringFlag{
			Name:  flagCoordCfg,
			Usage: "Coordinator configuration `FILE`",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Aliases: []string{},
			Usage:   "Initialize the hermez-node",
			Action:  cmdInit,
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
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
}
