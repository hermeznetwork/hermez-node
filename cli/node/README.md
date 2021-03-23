# node cli

This is the main cli for the node

## Go version

The `hermez-node` has been tested with go version 1.14

## Usage

```shell
NAME:
   hermez-node - A new cli application

USAGE:
   node [global options] command [command options] [arguments...]

VERSION:
   v0.1.0-6-gd8a50c5

COMMANDS:
   version    Show the application version
   importkey  Import ethereum private key
   genbjj     Generate a new BabyJubJub key
   wipesql    Wipe the SQL DB (HistoryDB and L2DB) and the StateDBs, leaving the DB in a clean state
   run        Run the hermez-node in the indicated mode
   serveapi   Serve the API only
   discard    Discard blocks up to a specified block number
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
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

When running the API in standalone mode, the required configuration is a subset
of the node configuration.  Please, check the `type APIServer` at
[config/config.go](../../config/config.go) to learn about all the parametes.

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
- The node requires a web3 RPC server to work.  The node has only been tested
  with geth and may not work correctly with other ethereum nodes
  implementations.

## Building

*All commands assume you are at the `cli/node` directory.*

Building the node requires using the packr utility to bundle the database
migrations inside the resulting binary.  Install the packr utility with:
```shell
cd /tmp && go get -u github.com/gobuffalo/packr/v2/packr2 && cd -
```

Make sure your `$PATH` contains `$GOPATH/bin`, otherwise the packr utility will
not be found.

Now build the node executable:
```shell
cd ../../db && packr2 && cd -
go build .
cd ../../db && packr2 clean && cd -
```

The executable is `node`.

## Usage Examples

The following commands assume you have built the node previously.  You can also
run the following examples by replacing `./node` with `go run .` and executing
them in the `cli/node` directory to build from source and run at the same time.

Run the node in mode synchronizer:
```shell
./node run --mode sync --cfg cfg.buidler.toml
```

Run the node in mode coordinator:
```shell
./node run --mode coord --cfg cfg.buidler.toml
```

Serve the API in standalone mode.  This command allows serving the API just
with access to the PostgreSQL database that a node is using.  Several instances
of `serveapi` can be running at the same time with a single PostgreSQL
database:
```shell
./node serveapi --mode coord --cfg cfg.buidler.toml
```

Import an ethereum private key into the keystore:
```shell
./node importkey --mode coord --cfg cfg.buidler.toml --privatekey  0x618b35096c477aab18b11a752be619f0023a539bb02dd6c813477a6211916cde
```

Generate a new BabyJubJub key pair:
```shell
./node genbjj
```

Check the binary version:
```shell
./node version
```

Wipe the entier SQL database (this will destroy all synchronized and pool
data):
```shell
./node wipesql --mode coord --cfg cfg.buidler.toml 
```

Discard all synchronized blocks and associated state up to a given block
number.  This command is useful in case the synchronizer reaches an invalid
state and you want to roll back a few blocks and try again (maybe with some
fixes in the code).
```shell
./node discard --mode coord --cfg cfg.buidler.toml --block 8061330
```
