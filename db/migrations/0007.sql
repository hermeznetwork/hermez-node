-- +migrate Up
ALTER TABLE tx_pool ADD COLUMN max_num_batch BIGINT DEFAULT NULL;

-- +migrate Down
ALTER TABLE tx_pool DROP COLUMN max_num_batch;