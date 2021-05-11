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
