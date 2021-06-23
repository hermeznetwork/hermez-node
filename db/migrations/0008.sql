-- +migrate Up
ALTER TABLE tx_pool ADD COLUMN atomic_group_id INTEGER;

-- +migrate Down
ALTER TABLE tx_pool DROP COLUMN atomic_group_id;