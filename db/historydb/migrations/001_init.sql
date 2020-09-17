-- +migrate Up
CREATE TABLE block (
    eth_block_num BIGINT PRIMARY KEY,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    hash BYTEA NOT NULL
);

CREATE TABLE coordianator (
    forger_addr BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    withdraw_addr BYTEA NOT NULL,
    url VARCHAR(200) NOT NULL,
    PRIMARY KEY (forger_addr, eth_block_num)
);

CREATE TABLE batch (
    batch_num BIGINT PRIMARY KEY,
    eth_block_num BIGINT REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    fees_collected BYTEA NOT NULL,
    state_root BYTEA NOT NULL,
    num_accounts BIGINT NOT NULL,
    exit_root BYTEA NOT NULL,
    forge_l1_txs_num BIGINT,
    slot_num BIGINT NOT NULL
);

CREATE TABLE exit_tree (
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE CASCADE,
    withdrawn BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL,
    account_idx BIGINT,
    merkle_proof BYTEA NOT NULL,
    balance NUMERIC NOT NULL,
    nullifier BYTEA NOT NULL,
    PRIMARY KEY (batch_num, account_idx)
);

CREATE TABLE bid (
    slot_num BIGINT NOT NULL,
    bid_value BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    PRIMARY KEY (slot_num, bid_value)
);

CREATE TABLE token (
    token_id INT PRIMARY KEY,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    eth_addr BYTEA UNIQUE NOT NULL,
    name VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimals INT NOT NULL,
    usd NUMERIC,
    usd_update TIMESTAMP
);

-- +migrate StatementBegin
CREATE FUNCTION set_token_usd_update() 
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
CREATE TRIGGER trigger_token_usd_update BEFORE UPDATE OR INSERT ON token
FOR EACH ROW EXECUTE PROCEDURE set_token_usd_update();

CREATE TABLE tx (
    -- Generic TX
    is_l1 BOOLEAN NOT NULL,
    id BYTEA PRIMARY KEY,
    type VARCHAR(40) NOT NULL,
    position INT NOT NULL,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT NOT NULL,
    amount BYTEA NOT NULL,
    amount_f NUMERIC NOT NULL,
    token_id INT NOT NULL REFERENCES token (token_id),
    amount_usd NUMERIC, -- Value of the amount in USD at the moment the tx was inserted in the DB
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL, -- Can be NULL in the case of L1 txs that are on the queue but not forged yet.
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    -- L1
    to_forge_l1_txs_num BIGINT,
    user_origin BOOLEAN,
    from_eth_addr BYTEA,
    from_bjj BYTEA,
    load_amount BYTEA,
    load_amount_f NUMERIC,
    load_amount_usd NUMERIC,
    -- L2
    fee INT,
    fee_usd NUMERIC,
    nonce BIGINT
);

-- +migrate StatementBegin
CREATE FUNCTION set_tx()
    RETURNS TRIGGER 
AS 
$BODY$
DECLARE token_value NUMERIC := (SELECT usd FROM token WHERE token_id = NEW.token_id);
BEGIN
    -- Validate L1/L2 constrains
    IF NEW.is_l1 AND (( -- L1 mandatory fields
        NEW.user_origin IS NULL OR
        NEW.from_eth_addr IS NULL OR
        NEW.from_bjj IS NULL OR
        NEW.load_amount IS NULL OR
        NEW.load_amount_f IS NULL
    ) OR (NOT NEW.user_origin AND NEW.batch_num IS NULL)) THEN -- If is Coordinator L1, must include batch_num
        RAISE EXCEPTION 'Invalid L1 tx.';
    ELSIF NOT NEW.is_l1 THEN
        IF NEW.fee IS NULL THEN
            NEW.fee = (SELECT 0);
        END IF;
        IF NEW.batch_num IS NULL OR NEW.nonce IS NULL THEN
            RAISE EXCEPTION 'Invalid L2 tx.';
        END IF;
    END IF;
    -- If is L2, add token_id
    IF NEW.token_id IS NULL THEN
        NEW."token_id" = (SELECT token_id FROM account WHERE idx = NEW."from_idx");
    END IF;
    -- Set value_usd
    NEW."amount_usd" = (SELECT token_value * NEW.amount_f);
    NEW."load_amount_usd" = (SELECT token_value * NEW.load_amount_f);
    IF NOT NEW.is_l1 THEN
        NEW."fee_usd" = (SELECT token_value * NEW.amount_f * CASE
            WHEN NEW.fee = 0 THEN 0	
            WHEN NEW.fee >= 1 AND NEW.fee <= 32 THEN POWER(10,-24+(NEW.fee::float/2))	
            WHEN NEW.fee >= 33 AND NEW.fee <= 223 THEN POWER(10,-8+(0.041666666666667*(NEW.fee::float-32)))	
            WHEN NEW.fee >= 224 AND NEW.fee <= 255 THEN POWER(10,NEW.fee-224) END);
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_set_tx BEFORE INSERT ON tx
FOR EACH ROW EXECUTE PROCEDURE set_tx();

-- +migrate StatementBegin
CREATE FUNCTION forge_l1_user_txs() 
    RETURNS TRIGGER 
AS 
$BODY$
BEGIN
    IF NEW.forge_l1_txs_num IS NOT NULL THEN
        UPDATE tx 
        SET batch_num = NEW.batch_num
        WHERE user_origin AND NEW.forge_l1_txs_num = to_forge_l1_txs_num;
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_forge_l1_txs AFTER INSERT ON batch
FOR EACH ROW EXECUTE PROCEDURE forge_l1_user_txs();

CREATE TABLE account (
    idx BIGINT PRIMARY KEY,
    token_id INT NOT NULL REFERENCES token (token_id) ON DELETE CASCADE,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    bjj BYTEA NOT NULL,
    eth_addr BYTEA NOT NULL
);

CREATE TABLE rollup_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forge_l1_timeout BYTEA NOT NULL,
    fee_l1_user_tx BYTEA NOT NULL,
    fee_add_token BYTEA NOT NULL,
    tokens_hez BYTEA NOT NULL,
    governance BYTEA NOT NULL
);

CREATE TABLE consensus_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    slot_deadline INT NOT NULL,
    close_auction_slots INT NOT NULL,
    open_auction_slots INT NOT NULL,
    min_bid_slots VARCHAR(200) NOT NULL,
    outbidding INT NOT NULL,
    donation_address BYTEA NOT NULL,
    governance_address BYTEA NOT NULL,
    allocation_ratio VARCHAR(200)
);

-- +migrate Down
DROP TABLE consensus_vars;
DROP TABLE rollup_vars;
DROP TABLE account;
DROP TABLE l2tx;
DROP TABLE l1tx;
DROP TABLE token;
DROP TABLE bid;
DROP TABLE exit_tree;
DROP TABLE batch;
DROP TABLE coordianator;
DROP TABLE block;