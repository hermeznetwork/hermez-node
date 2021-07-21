# hermez-node [![Go Report Card](https://goreportcard.com/badge/github.com/hermeznetwork/hermez-node)](https://goreportcard.com/report/github.com/hermeznetwork/hermez-node) [![Test Status](https://github.com/hermeznetwork/hermez-node/workflows/Test/badge.svg)](https://github.com/hermeznetwork/hermez-node/actions?query=workflow%3ATest) [![Lint Status](https://github.com/hermeznetwork/hermez-node/workflows/Lint/badge.svg)](https://github.com/hermeznetwork/hermez-node/actions?query=workflow%3ALint) [![GoDoc](https://godoc.org/github.com/hermeznetwork/hermez-node?status.svg)](https://godoc.org/github.com/hermeznetwork/hermez-node)

Go implementation of the Hermez node.

## Developing

To contribute to this codebase you must follow the [branch model and development flow](council/gitflow.md).

### Go version

The `hermez-node` has been tested with go version 1.14

### Build

- Build the binary in the local environment for you current OS and check the current version:
```shell
$ make
$ ./dist/heznode version
```

- Build the binary in a docker container for all supported OS and check the current version (only docker needed):
```shell
$ make docker-build
$ ./dist/heznode_<LOCAL_OS>_amd64/heznode version
```

- Build the binary in the local environment for all supported OS using Goreleaser and check the current version:
```shell
$ make goreleaser
$ ./dist/heznode_<LOCAL_OS>_amd64/heznode version
```

### Run

First you must edit the default/template config file into [cmd/heznode/cfg.buidler.toml](cmd/heznode/cfg.builder.toml), 
there are more information about the config file into [cmd/heznode/README.md](cmd/heznode/README.md)

After setting the config, you can build and run the Hermez Node as a synchronizer:

```shell
$ make run-node
```

Or build and run as a coordinator, and also passing the config file from other location:

```shell
$ MODE=sync CONFIG=cmd/heznode/cfg.builder.toml make run-node
```

To check the useful make commands:

```shell
$ make help
```


### Run as a service

```shell
$ sudo make install
```

After, update the config file manually at `/etc/hermez/config.toml`

```shell
$ sudo service heznode start
```

To check status

```shell
$ sudo service heznode status
```

To stop

```shell
$ sudo service heznode stop
```

If you just kill the process systemd will restart without asking. 

### Unit testing

Running the unit tests requires a connection to a PostgreSQL database.  You can
run PostgreSQL with docker easily this way:

```shell
$ make run-database-container
```

Afterward, run the tests:
```shell
$ make test
```

There is an extra temporary option that allows you to run the API server using the 
Go tests. It will be removed once the API can be properly initialized with data 
from the synchronizer. To use this, run:

```shell
$ make test-api-server
```

It is also possible to run the tests with an existing PostgreSQL server, e.g.
```shell
$ PGHOST=someserver PGUSER=hermez2 PGPASSWORD=secret make test
```

### Lint

All Pull Requests need to pass the configured linter.

To run the linter locally, first, install [golangci-lint](https://golangci-lint.run).  
Afterward, you can check the lints with this command:

```shell
$ make gocilint
```

## Usage

### Node

See [cmd/heznode/README.md](cmd/heznode/README.md)

### Proof Server

The node in mode coordinator requires a proof server (a server capable of 
calculating proofs from the zkInputs). There is a mock proof server CLI 
at `test/proofserver/cmd` for testing purposes.

Usage of `test/proofserver/cmd`:

```shell
USAGE:
    go run ./test/proofserver/cmd OPTIONS

OPTIONS:
  -a string
        listen address (default "localhost:3000")
  -d duration
        proving time duration (default 2s)
```

Also, the Makefile commands can be used to run and stop the proof server 
in the background:

```shell
$ make run-proof-mock
$ make stop-proof-mock
```

### `/tmp` as tmpfs

For every processed batch, the node builds a temporary exit tree in a key-value
DB stored in `/tmp`.  It is highly recommended that `/tmp` is mounted as a RAM
file system in production to avoid unnecessary reads a writes to disk.  This
can be done by mounting `/tmp` as tmpfs; for example, by having this line in
`/etc/fstab`:
```
tmpfs			/tmp		tmpfs		defaults,noatime,mode=1777	0 0
```
