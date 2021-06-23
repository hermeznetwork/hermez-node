-- +migrate Up
ALTER TABLE tx_pool ADD COLUMN rq_offset BYTEA;
ALTER TABLE tx_pool ADD COLUMN atomic_group_id INT;

-- +migrate Down
ALTER TABLE tx_pool DROP COLUMN rq_offset;
ALTER TABLE tx_pool DROP COLUMN atomic_group_id;