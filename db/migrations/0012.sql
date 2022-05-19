-- +migrate Up
CREATE TABLE block_avail (
    avail_block_num BIGINT PRIMARY KEY,
    hash BYTEA NOT NULL,
    root BYTEA NOT NULL
);
-- +migrate Down
DROP TABLE IF EXISTS block_avail;