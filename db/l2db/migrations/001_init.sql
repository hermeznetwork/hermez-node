-- +migrate Up
CREATE TABLE tx_pool (
    tx_id BYTEA PRIMARY KEY,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT NOT NULL,
    to_eth_addr BYTEA NOT NULL,
    to_bjj BYTEA NOT NULL,
    token_id INT NOT NULL,
    amount BYTEA NOT NULL,
    amount_f NUMERIC NOT NULL,
    value_usd NUMERIC,
    fee SMALLINT NOT NULL,
    nonce BIGINT NOT NULL,
    state CHAR(4) NOT NULL,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    batch_num BIGINT,
    rq_from_idx BIGINT,
    rq_to_idx BIGINT,
    rq_to_eth_addr BYTEA,
    rq_to_bjj BYTEA,
    rq_token_id INT,
    rq_amount BYTEA,
    rq_fee SMALLINT,
    rq_nonce BIGINT,
    fee_usd NUMERIC,
    usd_update TIMESTAMP WITHOUT TIME ZONE,
    tx_type VARCHAR(40) NOT NULL
);

CREATE TABLE account_creation_auth (
    eth_addr BYTEA PRIMARY KEY,
    bjj BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- +migrate Down
DROP TABLE account_creation_auth;
DROP TABLE tx_pool;