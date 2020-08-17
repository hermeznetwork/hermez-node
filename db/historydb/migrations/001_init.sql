-- +migrate Up
CREATE TABLE block (
    eth_block_num BIGINT PRIMARY KEY,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    hash BYTEA NOT NULL
);

CREATE TABLE slot_min_prices (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    min_prices VARCHAR(200) NOT NULL
);

CREATE TABLE coordianator (
    forger_addr BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    beneficiary_addr BYTEA NOT NULL,
    withdraw_addr BYTEA NOT NULL,
    url VARCHAR(200) NOT NULL,
    PRIMARY KEY (forger_addr, eth_block_num)
);

CREATE TABLE batch (
    batch_num BIGINT PRIMARY KEY,
    eth_block_num BIGINT REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    fees_collected BYTEA NOT NULL,
    state_root BYTEA NOT NULL,
    num_accounts BIGINT NOT NULL,
    exit_root BYTEA NOT NULL,
    forge_l1_txs_num BIGINT,
    slot_num BIGINT NOT NULL
);

CREATE TABLE exit_tree (
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    account_idx BIGINT NOT NULL,
    merkle_proof BYTEA NOT NULL,
    amount NUMERIC NOT NULL,
    nullifier BYTEA NOT NULL,
    PRIMARY KEY (batch_num, account_idx)
);

CREATE TABLE bid (
    slot_num BIGINT NOT NULL,
    bid_value BYTEA NOT NULL, -- (check if we can do a max(), if not add float for order purposes)
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    PRIMARY KEY (slot_num, bid_value)
);

CREATE TABLE token (
    token_id INT PRIMARY KEY,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    eth_addr BYTEA UNIQUE NOT NULL,
    name VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimals INT NOT NULL
);

CREATE TABLE l1tx (
    tx_id BYTEA PRIMARY KEY,
    to_forge_l1_txs_num BIGINT NOT NULL,
    position INT NOT NULL,
    user_origin BOOLEAN NOT NULL,
    from_idx BIGINT NOT NULL,
    from_eth_addr BYTEA NOT NULL,
    from_bjj BYTEA NOT NULL,
    to_idx BIGINT NOT NULL,
    token_id INT NOT NULL REFERENCES token (token_id),
    amount NUMERIC NOT NULL,
    load_amount BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE
);

CREATE TABLE l2tx (
    tx_id BYTEA PRIMARY KEY,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    position INT NOT NULL,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT NOT NULL,
    amount NUMERIC NOT NULL,
    fee INT NOT NULL,
    nonce BIGINT NOT NULL
);

CREATE TABLE account (
    idx BIGINT PRIMARY KEY,
    token_id INT NOT NULL REFERENCES token (token_id),
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    bjj BYTEA NOT NULL,
    eth_addr BYTEA NOT NULL
);

-- +migrate Down
DROP TABLE account;
DROP TABLE l2tx;
DROP TABLE l1tx;
DROP TABLE token;
DROP TABLE bid;
DROP TABLE exit_tree;
DROP TABLE batch;
DROP TABLE coordianator;
DROP TABLE slot_min_prices;
DROP TABLE block;