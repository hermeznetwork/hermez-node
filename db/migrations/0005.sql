-- +migrate Up
ALTER TABLE tx_pool
ADD COLUMN rq_tx_id BYTEA DEFAULT NULL,
ADD COLUMN rq_offset SMALLINT DEFAULT NULL,
ADD COLUMN atomic_group_id BIGINT DEFAULT NULL;

-- +migrate Down
ALTER TABLE tx_pool
DROP COLUMN rq_tx_id,
DROP COLUMN rq_offset,
DROP COLUMN atomic_group_id;