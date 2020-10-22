-- +migrate Up

-- History
CREATE TABLE block (
    eth_block_num BIGINT PRIMARY KEY,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    hash BYTEA NOT NULL
);

CREATE TABLE coordinator (
    item_id SERIAL PRIMARY KEY,
    bidder_addr BYTEA NOT NULL,
    forger_addr BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    url VARCHAR(200) NOT NULL
);

CREATE TABLE batch (
    item_id SERIAL PRIMARY KEY,
    batch_num BIGINT NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    fees_collected BYTEA NOT NULL,
    state_root BYTEA NOT NULL,
    num_accounts BIGINT NOT NULL,
    exit_root BYTEA NOT NULL,
    forge_l1_txs_num BIGINT,
    slot_num BIGINT NOT NULL,
    total_fees_usd NUMERIC
);

CREATE TABLE bid (
    slot_num BIGINT NOT NULL,
    bid_value BYTEA NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    bidder_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    PRIMARY KEY (slot_num, bid_value)
);

CREATE TABLE token (
    item_id SERIAL PRIMARY KEY,
    token_id INT UNIQUE NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    eth_addr BYTEA UNIQUE NOT NULL,
    name VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimals INT NOT NULL,
    usd NUMERIC,
    usd_update TIMESTAMP WITHOUT TIME ZONE
);

CREATE TABLE account (
    idx BIGINT PRIMARY KEY,
    token_id INT NOT NULL REFERENCES token (token_id) ON DELETE CASCADE,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    bjj BYTEA NOT NULL,
    eth_addr BYTEA NOT NULL
);

CREATE TABLE exit_tree (
    item_id SERIAL PRIMARY KEY,
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE CASCADE,
    account_idx BIGINT REFERENCES account (idx) ON DELETE CASCADE,
    merkle_proof BYTEA NOT NULL,
    balance BYTEA NOT NULL,
    instant_withdrawn BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL,
    delayed_withdraw_request BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL,
    delayed_withdrawn BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL
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

CREATE SEQUENCE tx_item_id;

CREATE TABLE tx (
    -- Generic TX
    item_id INTEGER PRIMARY KEY DEFAULT nextval('tx_item_id'),
    is_l1 BOOLEAN NOT NULL,
    id BYTEA,
    type VARCHAR(40) NOT NULL,
    position INT NOT NULL,
    from_idx BIGINT,
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
CREATE FUNCTION fee_percentage(NUMERIC)
    RETURNS NUMERIC 
AS 
$BODY$
DECLARE perc NUMERIC;
BEGIN
    SELECT CASE 
        WHEN $1 = 0 THEN 0
        WHEN $1 = 1 THEN 3.162278e-24
        WHEN $1 = 2 THEN 1.000000e-23
        WHEN $1 = 3 THEN 3.162278e-23
        WHEN $1 = 4 THEN 1.000000e-22
        WHEN $1 = 5 THEN 3.162278e-22
        WHEN $1 = 6 THEN 1.000000e-21
        WHEN $1 = 7 THEN 3.162278e-21
        WHEN $1 = 8 THEN 1.000000e-20
        WHEN $1 = 9 THEN 3.162278e-20
        WHEN $1 = 10 THEN 1.000000e-19
        WHEN $1 = 11 THEN 3.162278e-19
        WHEN $1 = 12 THEN 1.000000e-18
        WHEN $1 = 13 THEN 3.162278e-18
        WHEN $1 = 14 THEN 1.000000e-17
        WHEN $1 = 15 THEN 3.162278e-17
        WHEN $1 = 16 THEN 1.000000e-16
        WHEN $1 = 17 THEN 3.162278e-16
        WHEN $1 = 18 THEN 1.000000e-15
        WHEN $1 = 19 THEN 3.162278e-15
        WHEN $1 = 20 THEN 1.000000e-14
        WHEN $1 = 21 THEN 3.162278e-14
        WHEN $1 = 22 THEN 1.000000e-13
        WHEN $1 = 23 THEN 3.162278e-13
        WHEN $1 = 24 THEN 1.000000e-12
        WHEN $1 = 25 THEN 3.162278e-12
        WHEN $1 = 26 THEN 1.000000e-11
        WHEN $1 = 27 THEN 3.162278e-11
        WHEN $1 = 28 THEN 1.000000e-10
        WHEN $1 = 29 THEN 3.162278e-10
        WHEN $1 = 30 THEN 1.000000e-09
        WHEN $1 = 31 THEN 3.162278e-09
        WHEN $1 = 32 THEN 1.000000e-08
        WHEN $1 = 33 THEN 1.100694e-08
        WHEN $1 = 34 THEN 1.211528e-08
        WHEN $1 = 35 THEN 1.333521e-08
        WHEN $1 = 36 THEN 1.467799e-08
        WHEN $1 = 37 THEN 1.615598e-08
        WHEN $1 = 38 THEN 1.778279e-08
        WHEN $1 = 39 THEN 1.957342e-08
        WHEN $1 = 40 THEN 2.154435e-08
        WHEN $1 = 41 THEN 2.371374e-08
        WHEN $1 = 42 THEN 2.610157e-08
        WHEN $1 = 43 THEN 2.872985e-08
        WHEN $1 = 44 THEN 3.162278e-08
        WHEN $1 = 45 THEN 3.480701e-08
        WHEN $1 = 46 THEN 3.831187e-08
        WHEN $1 = 47 THEN 4.216965e-08
        WHEN $1 = 48 THEN 4.641589e-08
        WHEN $1 = 49 THEN 5.108970e-08
        WHEN $1 = 50 THEN 5.623413e-08
        WHEN $1 = 51 THEN 6.189658e-08
        WHEN $1 = 52 THEN 6.812921e-08
        WHEN $1 = 53 THEN 7.498942e-08
        WHEN $1 = 54 THEN 8.254042e-08
        WHEN $1 = 55 THEN 9.085176e-08
        WHEN $1 = 56 THEN 1.000000e-07
        WHEN $1 = 57 THEN 1.100694e-07
        WHEN $1 = 58 THEN 1.211528e-07
        WHEN $1 = 59 THEN 1.333521e-07
        WHEN $1 = 60 THEN 1.467799e-07
        WHEN $1 = 61 THEN 1.615598e-07
        WHEN $1 = 62 THEN 1.778279e-07
        WHEN $1 = 63 THEN 1.957342e-07
        WHEN $1 = 64 THEN 2.154435e-07
        WHEN $1 = 65 THEN 2.371374e-07
        WHEN $1 = 66 THEN 2.610157e-07
        WHEN $1 = 67 THEN 2.872985e-07
        WHEN $1 = 68 THEN 3.162278e-07
        WHEN $1 = 69 THEN 3.480701e-07
        WHEN $1 = 70 THEN 3.831187e-07
        WHEN $1 = 71 THEN 4.216965e-07
        WHEN $1 = 72 THEN 4.641589e-07
        WHEN $1 = 73 THEN 5.108970e-07
        WHEN $1 = 74 THEN 5.623413e-07
        WHEN $1 = 75 THEN 6.189658e-07
        WHEN $1 = 76 THEN 6.812921e-07
        WHEN $1 = 77 THEN 7.498942e-07
        WHEN $1 = 78 THEN 8.254042e-07
        WHEN $1 = 79 THEN 9.085176e-07
        WHEN $1 = 80 THEN 1.000000e-06
        WHEN $1 = 81 THEN 1.100694e-06
        WHEN $1 = 82 THEN 1.211528e-06
        WHEN $1 = 83 THEN 1.333521e-06
        WHEN $1 = 84 THEN 1.467799e-06
        WHEN $1 = 85 THEN 1.615598e-06
        WHEN $1 = 86 THEN 1.778279e-06
        WHEN $1 = 87 THEN 1.957342e-06
        WHEN $1 = 88 THEN 2.154435e-06
        WHEN $1 = 89 THEN 2.371374e-06
        WHEN $1 = 90 THEN 2.610157e-06
        WHEN $1 = 91 THEN 2.872985e-06
        WHEN $1 = 92 THEN 3.162278e-06
        WHEN $1 = 93 THEN 3.480701e-06
        WHEN $1 = 94 THEN 3.831187e-06
        WHEN $1 = 95 THEN 4.216965e-06
        WHEN $1 = 96 THEN 4.641589e-06
        WHEN $1 = 97 THEN 5.108970e-06
        WHEN $1 = 98 THEN 5.623413e-06
        WHEN $1 = 99 THEN 6.189658e-06
        WHEN $1 = 100 THEN 6.812921e-06
        WHEN $1 = 101 THEN 7.498942e-06
        WHEN $1 = 102 THEN 8.254042e-06
        WHEN $1 = 103 THEN 9.085176e-06
        WHEN $1 = 104 THEN 1.000000e-05
        WHEN $1 = 105 THEN 1.100694e-05
        WHEN $1 = 106 THEN 1.211528e-05
        WHEN $1 = 107 THEN 1.333521e-05
        WHEN $1 = 108 THEN 1.467799e-05
        WHEN $1 = 109 THEN 1.615598e-05
        WHEN $1 = 110 THEN 1.778279e-05
        WHEN $1 = 111 THEN 1.957342e-05
        WHEN $1 = 112 THEN 2.154435e-05
        WHEN $1 = 113 THEN 2.371374e-05
        WHEN $1 = 114 THEN 2.610157e-05
        WHEN $1 = 115 THEN 2.872985e-05
        WHEN $1 = 116 THEN 3.162278e-05
        WHEN $1 = 117 THEN 3.480701e-05
        WHEN $1 = 118 THEN 3.831187e-05
        WHEN $1 = 119 THEN 4.216965e-05
        WHEN $1 = 120 THEN 4.641589e-05
        WHEN $1 = 121 THEN 5.108970e-05
        WHEN $1 = 122 THEN 5.623413e-05
        WHEN $1 = 123 THEN 6.189658e-05
        WHEN $1 = 124 THEN 6.812921e-05
        WHEN $1 = 125 THEN 7.498942e-05
        WHEN $1 = 126 THEN 8.254042e-05
        WHEN $1 = 127 THEN 9.085176e-05
        WHEN $1 = 128 THEN 1.000000e-04
        WHEN $1 = 129 THEN 1.100694e-04
        WHEN $1 = 130 THEN 1.211528e-04
        WHEN $1 = 131 THEN 1.333521e-04
        WHEN $1 = 132 THEN 1.467799e-04
        WHEN $1 = 133 THEN 1.615598e-04
        WHEN $1 = 134 THEN 1.778279e-04
        WHEN $1 = 135 THEN 1.957342e-04
        WHEN $1 = 136 THEN 2.154435e-04
        WHEN $1 = 137 THEN 2.371374e-04
        WHEN $1 = 138 THEN 2.610157e-04
        WHEN $1 = 139 THEN 2.872985e-04
        WHEN $1 = 140 THEN 3.162278e-04
        WHEN $1 = 141 THEN 3.480701e-04
        WHEN $1 = 142 THEN 3.831187e-04
        WHEN $1 = 143 THEN 4.216965e-04
        WHEN $1 = 144 THEN 4.641589e-04
        WHEN $1 = 145 THEN 5.108970e-04
        WHEN $1 = 146 THEN 5.623413e-04
        WHEN $1 = 147 THEN 6.189658e-04
        WHEN $1 = 148 THEN 6.812921e-04
        WHEN $1 = 149 THEN 7.498942e-04
        WHEN $1 = 150 THEN 8.254042e-04
        WHEN $1 = 151 THEN 9.085176e-04
        WHEN $1 = 152 THEN 1.000000e-03
        WHEN $1 = 153 THEN 1.100694e-03
        WHEN $1 = 154 THEN 1.211528e-03
        WHEN $1 = 155 THEN 1.333521e-03
        WHEN $1 = 156 THEN 1.467799e-03
        WHEN $1 = 157 THEN 1.615598e-03
        WHEN $1 = 158 THEN 1.778279e-03
        WHEN $1 = 159 THEN 1.957342e-03
        WHEN $1 = 160 THEN 2.154435e-03
        WHEN $1 = 161 THEN 2.371374e-03
        WHEN $1 = 162 THEN 2.610157e-03
        WHEN $1 = 163 THEN 2.872985e-03
        WHEN $1 = 164 THEN 3.162278e-03
        WHEN $1 = 165 THEN 3.480701e-03
        WHEN $1 = 166 THEN 3.831187e-03
        WHEN $1 = 167 THEN 4.216965e-03
        WHEN $1 = 168 THEN 4.641589e-03
        WHEN $1 = 169 THEN 5.108970e-03
        WHEN $1 = 170 THEN 5.623413e-03
        WHEN $1 = 171 THEN 6.189658e-03
        WHEN $1 = 172 THEN 6.812921e-03
        WHEN $1 = 173 THEN 7.498942e-03
        WHEN $1 = 174 THEN 8.254042e-03
        WHEN $1 = 175 THEN 9.085176e-03
        WHEN $1 = 176 THEN 1.000000e-02
        WHEN $1 = 177 THEN 1.100694e-02
        WHEN $1 = 178 THEN 1.211528e-02
        WHEN $1 = 179 THEN 1.333521e-02
        WHEN $1 = 180 THEN 1.467799e-02
        WHEN $1 = 181 THEN 1.615598e-02
        WHEN $1 = 182 THEN 1.778279e-02
        WHEN $1 = 183 THEN 1.957342e-02
        WHEN $1 = 184 THEN 2.154435e-02
        WHEN $1 = 185 THEN 2.371374e-02
        WHEN $1 = 186 THEN 2.610157e-02
        WHEN $1 = 187 THEN 2.872985e-02
        WHEN $1 = 188 THEN 3.162278e-02
        WHEN $1 = 189 THEN 3.480701e-02
        WHEN $1 = 190 THEN 3.831187e-02
        WHEN $1 = 191 THEN 4.216965e-02
        WHEN $1 = 192 THEN 4.641589e-02
        WHEN $1 = 193 THEN 5.108970e-02
        WHEN $1 = 194 THEN 5.623413e-02
        WHEN $1 = 195 THEN 6.189658e-02
        WHEN $1 = 196 THEN 6.812921e-02
        WHEN $1 = 197 THEN 7.498942e-02
        WHEN $1 = 198 THEN 8.254042e-02
        WHEN $1 = 199 THEN 9.085176e-02
        WHEN $1 = 200 THEN 1.000000e-01
        WHEN $1 = 201 THEN 1.100694e-01
        WHEN $1 = 202 THEN 1.211528e-01
        WHEN $1 = 203 THEN 1.333521e-01
        WHEN $1 = 204 THEN 1.467799e-01
        WHEN $1 = 205 THEN 1.615598e-01
        WHEN $1 = 206 THEN 1.778279e-01
        WHEN $1 = 207 THEN 1.957342e-01
        WHEN $1 = 208 THEN 2.154435e-01
        WHEN $1 = 209 THEN 2.371374e-01
        WHEN $1 = 210 THEN 2.610157e-01
        WHEN $1 = 211 THEN 2.872985e-01
        WHEN $1 = 212 THEN 3.162278e-01
        WHEN $1 = 213 THEN 3.480701e-01
        WHEN $1 = 214 THEN 3.831187e-01
        WHEN $1 = 215 THEN 4.216965e-01
        WHEN $1 = 216 THEN 4.641589e-01
        WHEN $1 = 217 THEN 5.108970e-01
        WHEN $1 = 218 THEN 5.623413e-01
        WHEN $1 = 219 THEN 6.189658e-01
        WHEN $1 = 220 THEN 6.812921e-01
        WHEN $1 = 221 THEN 7.498942e-01
        WHEN $1 = 222 THEN 8.254042e-01
        WHEN $1 = 223 THEN 9.085176e-01
        WHEN $1 = 224 THEN 1.000000e+00
        WHEN $1 = 225 THEN 1.000000e+01
        WHEN $1 = 226 THEN 1.000000e+02
        WHEN $1 = 227 THEN 1.000000e+03
        WHEN $1 = 228 THEN 1.000000e+04
        WHEN $1 = 229 THEN 1.000000e+05
        WHEN $1 = 230 THEN 1.000000e+06
        WHEN $1 = 231 THEN 1.000000e+07
        WHEN $1 = 232 THEN 1.000000e+08
        WHEN $1 = 233 THEN 1.000000e+09
        WHEN $1 = 234 THEN 1.000000e+10
        WHEN $1 = 235 THEN 1.000000e+11
        WHEN $1 = 236 THEN 1.000000e+12
        WHEN $1 = 237 THEN 1.000000e+13
        WHEN $1 = 238 THEN 1.000000e+14
        WHEN $1 = 239 THEN 1.000000e+15
        WHEN $1 = 240 THEN 1.000000e+16
        WHEN $1 = 241 THEN 1.000000e+17
        WHEN $1 = 242 THEN 1.000000e+18
        WHEN $1 = 243 THEN 1.000000e+19
        WHEN $1 = 244 THEN 1.000000e+20
        WHEN $1 = 245 THEN 1.000000e+21
        WHEN $1 = 246 THEN 1.000000e+22
        WHEN $1 = 247 THEN 1.000000e+23
        WHEN $1 = 248 THEN 1.000000e+24
        WHEN $1 = 249 THEN 1.000000e+25
        WHEN $1 = 250 THEN 1.000000e+26
        WHEN $1 = 251 THEN 1.000000e+27
        WHEN $1 = 252 THEN 1.000000e+28
        WHEN $1 = 253 THEN 1.000000e+29
        WHEN $1 = 254 THEN 1.000000e+30
        WHEN $1 = 255 THEN 1.000000e+31
    END INTO perc;
    RETURN perc;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate StatementBegin
CREATE FUNCTION set_tx()
    RETURNS TRIGGER 
AS 
$BODY$
DECLARE
	_value NUMERIC;
	_usd_update TIMESTAMP;
    _tx_timestamp TIMESTAMP;
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
    IF NOT NEW.is_l1 THEN
        NEW."token_id" = (SELECT token_id FROM account WHERE idx = NEW."from_idx");
    END IF;
    -- Set value_usd
    SELECT INTO _value, _usd_update, _tx_timestamp 
        usd / POWER(10, decimals), usd_update, timestamp FROM token INNER JOIN block on token.eth_block_num = block.eth_block_num WHERE token_id = NEW.token_id;
    IF _tx_timestamp - interval '24 hours' < _usd_update AND _tx_timestamp + interval '24 hours' > _usd_update THEN
        NEW."amount_usd" = (SELECT _value * NEW.amount_f);
        IF NOT NEW.is_l1 THEN
            NEW."fee_usd" = (SELECT NEW."amount_usd" * fee_percentage(NEW.fee::NUMERIC));
        ELSE 
            NEW."load_amount_usd" = (SELECT _value * NEW.load_amount_f);
        END IF;
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
        SET item_id = nextval('tx_item_id'), batch_num = NEW.batch_num 
        WHERE id IN (
            SELECT id FROM tx 
            WHERE user_origin AND NEW.forge_l1_txs_num = to_forge_l1_txs_num 
            ORDER BY position
            FOR UPDATE
        ); 
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_forge_l1_txs AFTER INSERT ON batch
FOR EACH ROW EXECUTE PROCEDURE forge_l1_user_txs();

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

-- L2
CREATE TABLE tx_pool (
    tx_id BYTEA PRIMARY KEY,
    from_idx BIGINT NOT NULL,
    to_idx BIGINT,
    to_eth_addr BYTEA,
    to_bjj BYTEA,
    token_id INT NOT NULL REFERENCES token (token_id) ON DELETE CASCADE,
    amount BYTEA NOT NULL,
    amount_f NUMERIC NOT NULL,
    fee SMALLINT NOT NULL,
    nonce BIGINT NOT NULL,
    state CHAR(4) NOT NULL,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE DEFAULT timezone('utc', now()),
    batch_num BIGINT,
    rq_from_idx BIGINT,
    rq_to_idx BIGINT,
    rq_to_eth_addr BYTEA,
    rq_to_bjj BYTEA,
    rq_token_id INT,
    rq_amount BYTEA,
    rq_fee SMALLINT,
    rq_nonce BIGINT,
    tx_type VARCHAR(40) NOT NULL
);

CREATE TABLE account_creation_auth (
    eth_addr BYTEA PRIMARY KEY,
    bjj BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT timezone('utc', now())
);

-- +migrate Down
DROP TABLE account_creation_auth;
DROP TABLE tx_pool;
DROP TABLE consensus_vars;
DROP TABLE rollup_vars;
DROP TABLE account;
DROP TABLE tx;
DROP TABLE token;
DROP TABLE bid;
DROP TABLE exit_tree;
DROP TABLE batch;
DROP TABLE coordinator;
DROP TABLE block;
