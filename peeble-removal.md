> We are reprioritizing this issue in order to improve the synchronizer performance and maintenance.

## Targets

- Move synchronizer to a separated binary
- Move from a multiple database to a single database solution
- Synchronize L1 data based on events instead block by block
- Synchronize only blocks with Hermez smart contract events
- Process things concurrently and in parallel when possible
- Have two different synchronization strategies: one for old data and another for fresh data in the blockchain
- Allow state data to be visualized/debugged in the database

## Current implementation

Today the node implementation depends on two different databases, it uses `PostgreSQL` and `Peeble`(https://github.com/iden3/go-merkletree/tree/master/db/pebble)

Internally the node has 4 database concepts:

- History: Stores information like `blocks`, `batches`, `tokens` and `tx` synchronized from L1
- L2: Stores information like `account creation authorization` and `tx pool`
- Merkle Tree: Stores information like tree leaves, indexes, data
- State: Stores information the merkle tree state in a certain point in time like current, last and by batch number.

The `History` and `L2` dbs are using `Postgres` to store its data, while `Merkle Tree` and `State` are using `Peeble`.

### Merkle Tree

@OBrezhniev created a fork from `Merkle Tree` repository providing a Merkle Tree `Postgres` storage driver, here is it: https://github.com/iden3/go-merkletree-sql

We are going to use this package to migrate the `Merkle Tree` from `Peeble` to `Postgres`. To keep things simples, we will use the same write configuration as the databases `History` and `L2` to connect to Postgres.

We also need to create two new tables:

```sql
-- +migrate Up
CREATE TABLE mt_nodes (
    mt_id BIGINT,
    key BYTEA,
    type SMALLINT NOT NULL,
    child_l BYTEA,
    child_r BYTEA,
    entry BYTEA,
    created_at BIGINT,
    deleted_at BIGINT,
    PRIMARY KEY(mt_id, key)
);

CREATE TABLE mt_roots (
    mt_id BIGINT PRIMARY KEY,
    key BYTEA,
    created_at BIGINT,
    deleted_at BIGINT
);


-- +migrate Down
DROP TABLE mt_nodes;
DROP TABLE mt_roots;
```

And we finally need to replace instantiation of the Merkle Tree to receive a `Postgres` storage instead of a `Peeble` storage, here is it: https://github.com/hermeznetwork/hermez-node/blob/develop/db/statedb/statedb.go#L143

### State

The current implementation of the `State` is a bit trick, inside of this abstraction we have things related to the synchronizer but also things related to the coordinator, lets try to separate things here to make it clear.

There are 3 different state concepts:

- Current: The current state of the Merkle Tree synchronized from the L1
- Last: It is created from the `Current` state and is changed by the `Coordinator` when forging a new `Batch`
- BatchNum: The state of the `Merkle Tree` for each `Batch` already forged or being forged by the `Coordinator`