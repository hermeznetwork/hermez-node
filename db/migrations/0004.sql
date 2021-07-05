-- +migrate Up
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_token_usd_update() 
    RETURNS TRIGGER 
AS 
$BODY$
BEGIN
	if tg_op = 'INSERT' THEN 
    IF NEW."usd" IS NOT NULL AND NEW."usd_update" IS NULL THEN
        NEW."usd_update" = timezone('utc', now());
    END IF;
    elsif tg_op = 'UPDATE' then
        NEW."usd_update" = timezone('utc', now());
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TABLE IF NOT EXISTS fiat (
    item_id SERIAL NOT NULL,
    currency VARCHAR(10) NOT NULL,
    base_currency VARCHAR(10) NOT NULL,
    price NUMERIC NOT NULL,
    last_update TIMESTAMP WITHOUT TIME ZONE DEFAULT timezone('utc', now()) NOT NULL,
    PRIMARY KEY(currency, base_currency)
);
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_fiat_last_update() 
    RETURNS TRIGGER 
AS 
$BODY$
BEGIN
	IF tg_op = 'INSERT' THEN 
        NEW."last_update" = timezone('utc', now());
    ELSIF tg_op = 'UPDATE' then
        NEW."last_update" = timezone('utc', now());
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_fiat_price_update BEFORE UPDATE OR INSERT ON fiat
FOR EACH ROW EXECUTE PROCEDURE set_fiat_last_update();

-- +migrate Down
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_token_usd_update() 
    RETURNS TRIGGER 
AS 
$BODY$
BEGIN
    IF NEW."usd" IS NOT NULL AND NEW."usd_update" IS NULL THEN
        NEW."usd_update" = timezone('utc', now());
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
DROP TRIGGER trigger_fiat_price_update ON fiat;
DROP FUNCTION set_fiat_last_update;
DROP TABLE fiat;
