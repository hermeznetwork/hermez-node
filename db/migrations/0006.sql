-- +migrate Up
ALTER TABLE tx_pool ADD COLUMN rq_offset BYTEA DEFAULT NULL;

-- +migrate Down
ALTER TABLE tx_pool DROP COLUMN rq_offset;