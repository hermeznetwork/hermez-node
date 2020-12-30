# node cli

This is the main cli for the node

## Usage

```
NAME:
   hermez-node - A new cli application

USAGE:
   node [global options] command [command options] [arguments...]

VERSION:
   0.1.0-alpha

COMMANDS:
   importkey  Import ethereum private key
   wipesql    Wipe the SQL DB (HistoryDB and L2DB), leaving the DB in a clean state
   run        Run the hermez-node in the indicated mode
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --mode MODE    Set node MODE (can be "sync" or "coord")
   --cfg FILE     Node configuration FILE
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Configuration

You can find a testing working configuration example at
[cfg.buidler.toml](./cfg.buidler.toml)

To read the documentation of each configuration parameter, please check the
`type Node` and `type Coordinator` at
[config/config.go](../../config/config.go)
