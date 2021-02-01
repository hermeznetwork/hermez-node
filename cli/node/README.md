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
   genbjj     Generate a new BabyJubJub key
   wipesql    Wipe the SQL DB (HistoryDB and L2DB), leaving the DB in a clean state
   run        Run the hermez-node in the indicated mode
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --mode MODE    Set node MODE (can be "sync" or "coord")
   --cfg FILE     Node configuration FILE
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

The node has two main modes of running:
- `sync`: Synchronizer mode.  In this mode the node will only synchronize the
  state of the hermez smart contracts, mainly processing the transactions in
  the batches.
- `coord`: Coordinator mode.  In this mode, apart from doing all the
  synchronization work, the node will also act as a coordinator, accepting L2
  transactions in the pool, and trying to forge batches when the proper
  conditions arise.

## Configuration

The node requires a single configuration file to run.

You can find a testing working configuration example at
[cfg.buidler.toml](./cfg.buidler.toml)

To read the documentation of each configuration parameter, please check the
`type Node` and `type Coordinator` at
[config/config.go](../../config/config.go).  All the sections that are prefixed
with `Coordinator` are only used in coord mode, and don't need to be defined
when running the coordinator in sync mode

### Notes

- The private key corresponding to the parameter `Coordinator.ForgerAddress` needs to be imported in the ethereum keystore
- The private key corresponding to the parameter `Coordinator.FeeAccount.Address` needs to be imported in the ethereum keystore
- The public key corresponding to the parameter `Coordinator.FeeAccount.BJJ` can be generated with the command `genbjj`
- There are two sets of debug parameters (`Debug` for all modes, and
  `Coordinator.Debug` for `coord` mode).  Some of these parameters may not be
  suitable for production.
- The parameter `Coordinator.Debug.BatchPath`, when set, causes the coordinator
  to store dumps of a lot of information related to batches in json files.
  This files can be around 2MB big.  If this parameter is set, be careful to
  monitor the size of the folder to avoid running out of space.
- The node requires a PostgreSQL database.  The parameters of the server and
  database must be set in the `PostgreSQL` section.

## Usage Examples

Run the node in mode synchronizer:
```
go run . --mode sync --cfg cfg.buidler.toml run
```

Run the node in mode coordinator:
```
go run . --mode coord --cfg cfg.buidler.toml run
```

Import an ethereum private key into the keystore:
```
go run . --mode coord --cfg cfg.buidler.toml importkey --privatekey  0x618b35096c477aab18b11a752be619f0023a539bb02dd6c813477a6211916cde
```

Generate a new BabyJubJub key pair:
```
go run . --mode coord --cfg cfg.buidler.toml genbjj
```

Wipe the entier SQL database (this will destroy all synchronized and pool data):
```
go run . --mode coord --cfg cfg.buidler.toml wipesql
```
