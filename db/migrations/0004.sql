-- +migrate Up
ALTER TABLE tx_pool ADD COLUMN rq_tx_id BYTEA DEFAULT NULL;

-- +migrate Down
ALTER TABLE tx_pool DROP COLUMN rq_tx_id;
