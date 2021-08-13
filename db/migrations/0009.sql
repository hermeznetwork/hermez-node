-- +migrate Up
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_pool_tx()
    RETURNS TRIGGER
AS
$BODY$
BEGIN
    NEW."effective_to_eth_addr" = NEW."to_eth_addr";
    NEW."effective_to_bjj" = NEW."to_bjj";
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_update_pool_tx BEFORE UPDATE ON tx_pool
FOR EACH ROW EXECUTE PROCEDURE update_pool_tx();

-- +migrate Down
DROP TRIGGER IF EXISTS trigger_update_pool_tx ON tx_pool;
DROP FUNCTION IF EXISTS update_pool_tx();