-- +migrate Up
CREATE TABLE tx_pool (
    tx_id BYTEA PRIMARY KEY,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT NOT NULL,
    to_eth_addr BYTEA NOT NULL,
    to_bjj BYTEA NOT NULL,
    token_id INT NOT NULL,
    amount BYTEA NOT NULL,
    fee SMALLINT NOT NULL,
    nonce BIGINT NOT NULL,
    state CHAR(4) NOT NULL,
    batch_num BIGINT,
    rq_from_idx BIGINT,
    rq_to_idx BIGINT,
    rq_to_eth_addr BYTEA,
    rq_to_bjj BYTEA,
    rq_token_id INT,
    rq_amount BYTEA,
    rq_fee SMALLINT,
    rq_nonce BIGINT,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    absolute_fee NUMERIC,
    absolut_fee_update TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE account_creation_auth (
    eth_addr BYTEA PRIMARY KEY,
    bjj BYTEA NOT NULL,
    account_creation_auth_sig BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- +migrate Down
DROP TABLE account_creation_auth;
DROP TABLE tx_pool;