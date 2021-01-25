# hermez-node [![Go Report Card](https://goreportcard.com/badge/github.com/hermeznetwork/hermez-node)](https://goreportcard.com/report/github.com/hermeznetwork/hermez-node) [![Test Status](https://github.com/hermeznetwork/hermez-node/workflows/Test/badge.svg)](https://github.com/hermeznetwork/hermez-node/actions?query=workflow%3ATest) [![Lint Status](https://github.com/hermeznetwork/hermez-node/workflows/Lint/badge.svg)](https://github.com/hermeznetwork/hermez-node/actions?query=workflow%3ALint) [![GoDoc](https://godoc.org/github.com/hermeznetwork/hermez-node?status.svg)](https://godoc.org/github.com/hermeznetwork/hermez-node)

Go implementation of the Hermez node.

## Test

- First run a docker instance of the PostgresSQL (where `yourpasswordhere` should be your password)

```
POSTGRES_PASS=yourpasswordhere; sudo docker run --rm --name hermez-db-test -p 5432:5432 -e POSTGRES_DB=hermez -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD="$POSTGRES_PASS" -d postgres
```

- Then, run the tests with the password as env var

```
POSTGRES_PASS=yourpasswordhere go test -p 1 ./...
```

NOTE: `-p 1` forces execution of package test in serial.  Otherwise they may be
executed in paralel and the test may find unexpected entries in the SQL
databse because it's shared among all tests.

- There is an extra temporal option that allows to run the API server through the Go tests. This should be removed once the API can be properly initialized, with data from the synchronizer and so on. To use this, run `FAKE_SERVER=yes POSTGRES_PASS=yourpasswordhere go test -timeout 0  ./api -p 1 -count 1 -v`

## Lint

- Install [golangci-lint](https://golangci-lint.run)
- Once installed, to check the lints

```
golangci-lint run --timeout=5m -E whitespace -E gosec -E gci -E misspell -E gomnd -E gofmt -E goimports -E golint --exclude-use-default=false --max-same-issues 0
```

## Usage Details

### `/tmp` as tmpfs

For every processed batch, the node builds a temporary exit tree in a key-value
DB stored in `/tmp`.  It is highly recommended that `/tmp` is mounted as a RAM
file system in production to avoid unecessary reads an writes to disk.  This
can be done by mounting `/tmp` as tmpfs; for example, by having this line in
`/etc/fstab`:
```
tmpfs			/tmp		tmpfs		defaults,noatime,mode=1777	0 0
```


