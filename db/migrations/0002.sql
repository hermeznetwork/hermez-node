-- +migrate Up
ALTER TABLE tx_pool DROP CONSTRAINT tx_pool_pkey;
ALTER TABLE tx_pool ADD COLUMN item_id SERIAL PRIMARY KEY;
ALTER TABLE tx_pool ADD CONSTRAINT tx_id_unique UNIQUE (tx_id);

-- +migrate Down
ALTER TABLE tx_pool DROP CONSTRAINT tx_id_unique;
ALTER TABLE tx_pool DROP CONSTRAINT tx_pool_pkey;
ALTER TABLE tx_pool ADD PRIMARY KEY (tx_id);
ALTER TABLE tx_pool DROP COLUMN item_id;