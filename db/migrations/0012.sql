-- +migrate Up
-- +migrate StatementEnd
CREATE TABLE IF NOT EXISTS provers (
    public_dns VARCHAR(100) NOT NULL,
    instance_id VARCHAR(30) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT timezone('utc', now()) NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT timezone('utc', now()) NOT NULL,
    PRIMARY KEY(instance_id)
);
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER
AS
$BODY$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_set_updated_at BEFORE UPDATE ON provers
FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- +migrate Down
DROP TRIGGER trigger_set_updated_at ON provers;
DROP FUNCTION set_updated_at;
DROP TABLE provers;
