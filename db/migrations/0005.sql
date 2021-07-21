-- +migrate Up
ALTER TABLE token ALTER COLUMN name TYPE varchar;

-- +migrate Down
ALTER TABLE token ALTER COLUMN name TYPE varchar(20) USING SUBSTR(name, 1, 20);
