-- +migrate Up
ALTER TABLE tx ADD COLUMN eth_tx_hash BYTEA;
ALTER TABLE tx ADD COLUMN l1_fee DECIMAL(78,0);

-- +migrate Down
ALTER TABLE tx DROP COLUMN eth_tx_hash;
ALTER TABLE tx DROP COLUMN l1_fee;