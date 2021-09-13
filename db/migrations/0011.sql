-- +migrate Up
ALTER TABLE batch ADD COLUMN gas_price DECIMAL(78,0) DEFAULT 0;
ALTER TABLE batch ADD COLUMN gas_used DECIMAL(78,0) DEFAULT 0;
ALTER TABLE batch ADD COLUMN ether_price_usd NUMERIC DEFAULT 0;


-- +migrate Down
ALTER TABLE batch DROP COLUMN gas_price;
ALTER TABLE batch DROP COLUMN gas_used;
ALTER TABLE batch DROP COLUMN ether_price_usd;