-- +migrate Up
CREATE TABLE block (
    eth_block_num BIGINT PRIMARY KEY,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    hash BYTEA NOT NULL
);

CREATE TABLE coordianator (
    forger_addr BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
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
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE CASCADE,
    account_idx BIGINT,
    merkle_proof BYTEA NOT NULL,
    balance NUMERIC NOT NULL,
    nullifier BYTEA NOT NULL,
    PRIMARY KEY (batch_num, account_idx)
);

CREATE TABLE withdrawal (
    batch_num BIGINT,
    account_idx BIGINT,
    eth_block_num BIGINT REFERENCES block (eth_block_num) ON DELETE CASCADE,
    FOREIGN KEY (batch_num, account_idx) REFERENCES exit_tree (batch_num, account_idx) ON DELETE CASCADE,
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
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    tx_type VARCHAR(40) NOT NULL
);

CREATE TABLE l2tx (
    tx_id BYTEA PRIMARY KEY,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    position INT NOT NULL,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT NOT NULL,
    amount NUMERIC NOT NULL,
    fee INT NOT NULL,
    nonce BIGINT NOT NULL,
    tx_type VARCHAR(40) NOT NULL
);

CREATE TABLE account (
    idx BIGINT PRIMARY KEY,
    token_id INT NOT NULL REFERENCES token (token_id),
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    bjj BYTEA NOT NULL,
    eth_addr BYTEA NOT NULL
);

CREATE TABLE rollup_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forge_l1_timeout BYTEA NOT NULL,
    fee_l1_user_tx BYTEA NOT NULL,
    fee_add_token BYTEA NOT NULL,
    tokens_hez BYTEA NOT NULL,
    governance BYTEA NOT NULL
);

CREATE TABLE consensus_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    slot_deadline INT NOT NULL,
    close_auction_slots INT NOT NULL,
    open_auction_slots INT NOT NULL,
    min_bid_slots VARCHAR(200) NOT NULL,
    outbidding INT NOT NULL,
    donation_address BYTEA NOT NULL,
    governance_address BYTEA NOT NULL,
    allocation_ratio vARCHAR(200)
);

-- +migrate Down
DROP TABLE consensus_vars;
DROP TABLE rollup_vars;
DROP TABLE account;
DROP TABLE l2tx;
DROP TABLE l1tx;
DROP TABLE token;
DROP TABLE bid;
DROP TABLE exit_tree;
DROP TABLE batch;
DROP TABLE coordianator;
DROP TABLE block;