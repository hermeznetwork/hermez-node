-- +migrate Up
ALTER TABLE tx_pool
ADD COLUMN rq_offset SMALLINT DEFAULT NULL,
ADD COLUMN atomic_group_id BYTEA DEFAULT NULL;

-- +migrate Down
ALTER TABLE tx_pool
DROP COLUMN rq_offset,
DROP COLUMN atomic_group_id;