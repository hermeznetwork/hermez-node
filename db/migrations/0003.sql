-- +migrate Up
ALTER TABLE batch ADD COLUMN eth_tx_hash BYTEA DEFAULT NULL;

-- +migrate Down
ALTER TABLE batch DROP COLUMN eth_tx_hash;