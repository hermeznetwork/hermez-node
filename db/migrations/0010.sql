-- +migrate Up
ALTER TABLE tx ADD COLUMN eth_tx_hash BYTEA DEFAULT DECODE('0000000000000000000000000000000000000000000000000000000000000000', 'hex');

-- +migrate Down
ALTER TABLE tx DROP COLUMN eth_tx_hash;