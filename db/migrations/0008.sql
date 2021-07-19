-- +migrate Up
ALTER TABLE tx_pool
ADD COLUMN error_code NUMERIC,
ADD COLUMN error_type VARCHAR;

-- +migrate Down
ALTER TABLE tx_pool
DROP COLUMN error_code,
DROP COLUMN error_type;
