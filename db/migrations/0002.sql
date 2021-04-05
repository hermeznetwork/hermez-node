-- +migrate Up
ALTER TABLE tx_pool DROP CONSTRAINT tx_pool_pkey;
ALTER TABLE tx_pool ADD COLUMN item_id SERIAL PRIMARY KEY;

-- +migrate Down
ALTER TABLE tx_pool DROP CONSTRAINT tx_pool_pkey;
ALTER TABLE tx_pool ADD PRIMARY KEY (tx_id);
ALTER TABLE tx_pool DROP COLUMN item_id;