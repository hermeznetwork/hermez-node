-- +migrate Up

-- NOTE: We use "DECIMAL(78,0)" to encode go *big.Int types.  All the *big.Int
-- that we deal with represent a value in the SNARK field, which is an integer
-- of 256 bits.  `log(2**256, 10) = 77.06`: that is, a 256 bit number can have
-- at most 78 digits, so we use this value to specify the precision in the
-- PostgreSQL DECIMAL guaranteeing that we will never lose precision.

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
    url BYTEA NOT NULL
);

CREATE TABLE batch (
    item_id SERIAL PRIMARY KEY,
    batch_num BIGINT UNIQUE NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    forger_addr BYTEA NOT NULL, -- fake foreign key for coordinator
    fees_collected BYTEA NOT NULL,
    fee_idxs_coordinator BYTEA NOT NULL,
    state_root DECIMAL(78,0) NOT NULL,
    num_accounts BIGINT NOT NULL,
    last_idx BIGINT NOT NULL,
    exit_root DECIMAL(78,0) NOT NULL,
    forge_l1_txs_num BIGINT,
    slot_num BIGINT NOT NULL,
    total_fees_usd NUMERIC
);

CREATE TABLE bid (
    item_id SERIAL PRIMARY KEY,
    slot_num BIGINT NOT NULL,
    bid_value DECIMAL(78,0) NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    bidder_addr BYTEA NOT NULL -- fake foreign key for coordinator
);

CREATE TABLE token (
    item_id SERIAL PRIMARY KEY,
    token_id INT UNIQUE NOT NULL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    eth_addr BYTEA UNIQUE NOT NULL,
    name VARCHAR(20) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimals INT NOT NULL,
    usd NUMERIC, -- value of a normalized token (1 token = 10^decimals units)
    usd_update TIMESTAMP WITHOUT TIME ZONE
);

-- Add ETH as TokenID 0
INSERT INTO block (
    eth_block_num,
    timestamp,
    hash
) VALUES (
    0,
    '2015-07-30 03:26:13',
    '\xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3'
); -- info from https://etherscan.io/block/0

INSERT INTO token (
    token_id,
    eth_block_num,
    eth_addr,
    name,
    symbol,
    decimals
) VALUES (
    0,
    0,
    '\x0000000000000000000000000000000000000000',
    'Ether',
    'ETH',
    18
);


-- +migrate StatementBegin
CREATE FUNCTION hez_idx(BIGINT, VARCHAR) 
    RETURNS VARCHAR 
AS 
$BODY$
BEGIN
    RETURN 'hez:' || $2 || ':' || $1;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd

CREATE TABLE account (
    item_id SERIAL,
    idx BIGINT PRIMARY KEY,
    token_id INT NOT NULL REFERENCES token (token_id) ON DELETE CASCADE,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    bjj BYTEA NOT NULL,
    eth_addr BYTEA NOT NULL
);

CREATE TABLE account_update (
    item_id SERIAL,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    batch_num BIGINT NOT NULL REFERENCES batch (batch_num) ON DELETE CASCADE,
    idx BIGINT NOT NULL REFERENCES account (idx) ON DELETE CASCADE,
    nonce BIGINT NOT NULL,
    balance DECIMAL(78,0) NOT NULL
);

CREATE TABLE exit_tree (
    item_id SERIAL PRIMARY KEY,
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE CASCADE,
    account_idx BIGINT REFERENCES account (idx) ON DELETE CASCADE,
    merkle_proof BYTEA NOT NULL,
    balance DECIMAL(78,0) NOT NULL,
    instant_withdrawn BIGINT REFERENCES block (eth_block_num) ON DELETE SET NULL,
    delayed_withdraw_request BIGINT REFERENCES block (eth_block_num) ON DELETE SET NULL,
    owner BYTEA,
    token BYTEA,
    delayed_withdrawn BIGINT REFERENCES block (eth_block_num) ON DELETE SET NULL
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

-- important note about "amount_success" and "deposit_amount_success" (only relevant to L1 user txs):
-- The behaviour should be:
-- When tx is not forged: amount_success = false, deposit_amount_success = false
-- When tx is forged: 
--      amount_success = false if the "effective amount" is 0, else true
--      deposit_amount_success = false if the "effective deposit amount" is 0, else true
--
-- However, in order to reduce the amount of updates, by default amount_success and deposit_amount_success will be set to true (when tx is unforged)
-- whne they should be false. This can be worked around at a query level by checking if "batch_num IS NULL" (which indicates that the tx is unforged).
CREATE TABLE tx (
    -- Generic TX
    item_id INTEGER PRIMARY KEY DEFAULT nextval('tx_item_id'),
    is_l1 BOOLEAN NOT NULL,
    id BYTEA,
    type VARCHAR(40) NOT NULL,
    position INT NOT NULL,
    from_idx BIGINT,
    effective_from_idx BIGINT REFERENCES account (idx) ON DELETE SET NULL,
    from_eth_addr BYTEA,
    from_bjj BYTEA,
    to_idx BIGINT NOT NULL,
    to_eth_addr BYTEA,
    to_bjj BYTEA,
    amount DECIMAL(78,0) NOT NULL,
    amount_success BOOLEAN NOT NULL DEFAULT true,
    amount_f NUMERIC NOT NULL,
    token_id INT NOT NULL REFERENCES token (token_id),
    amount_usd NUMERIC, -- Value of the amount in USD at the moment the tx was inserted in the DB
    batch_num BIGINT REFERENCES batch (batch_num) ON DELETE SET NULL, -- Can be NULL in the case of L1 txs that are on the queue but not forged yet.
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    -- L1
    to_forge_l1_txs_num BIGINT,
    user_origin BOOLEAN,
    deposit_amount DECIMAL(78,0),
    deposit_amount_success BOOLEAN NOT NULL DEFAULT true,
    deposit_amount_f NUMERIC,
    deposit_amount_usd NUMERIC,
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
        WHEN $1 = 000 THEN 0.000000e+00
        WHEN $1 = 001 THEN 2.675309e-18
        WHEN $1 = 002 THEN 8.251782e-18
        WHEN $1 = 003 THEN 2.545198e-17
        WHEN $1 = 004 THEN 7.850462e-17
        WHEN $1 = 005 THEN 2.421414e-16
        WHEN $1 = 006 THEN 7.468660e-16
        WHEN $1 = 007 THEN 2.303650e-15
        WHEN $1 = 008 THEN 7.105427e-15
        WHEN $1 = 009 THEN 2.191613e-14
        WHEN $1 = 010 THEN 6.759860e-14
        WHEN $1 = 011 THEN 2.085026e-13
        WHEN $1 = 012 THEN 6.431099e-13
        WHEN $1 = 013 THEN 1.983622e-12
        WHEN $1 = 014 THEN 6.118327e-12
        WHEN $1 = 015 THEN 1.887150e-11
        WHEN $1 = 016 THEN 5.820766e-11
        WHEN $1 = 017 THEN 1.795370e-10
        WHEN $1 = 018 THEN 5.537677e-10
        WHEN $1 = 019 THEN 1.708053e-09
        WHEN $1 = 020 THEN 5.268356e-09
        WHEN $1 = 021 THEN 1.624983e-08
        WHEN $1 = 022 THEN 5.012133e-08
        WHEN $1 = 023 THEN 1.545953e-07
        WHEN $1 = 024 THEN 4.768372e-07
        WHEN $1 = 025 THEN 1.470767e-06
        WHEN $1 = 026 THEN 4.536465e-06
        WHEN $1 = 027 THEN 1.399237e-05
        WHEN $1 = 028 THEN 4.315837e-05
        WHEN $1 = 029 THEN 1.331186e-04
        WHEN $1 = 030 THEN 4.105940e-04
        WHEN $1 = 031 THEN 1.266445e-03
        WHEN $1 = 032 THEN 3.906250e-03
        WHEN $1 = 033 THEN 4.044004e-03
        WHEN $1 = 034 THEN 4.186615e-03
        WHEN $1 = 035 THEN 4.334256e-03
        WHEN $1 = 036 THEN 4.487103e-03
        WHEN $1 = 037 THEN 4.645340e-03
        WHEN $1 = 038 THEN 4.809158e-03
        WHEN $1 = 039 THEN 4.978752e-03
        WHEN $1 = 040 THEN 5.154328e-03
        WHEN $1 = 041 THEN 5.336095e-03
        WHEN $1 = 042 THEN 5.524272e-03
        WHEN $1 = 043 THEN 5.719085e-03
        WHEN $1 = 044 THEN 5.920768e-03
        WHEN $1 = 045 THEN 6.129563e-03
        WHEN $1 = 046 THEN 6.345722e-03
        WHEN $1 = 047 THEN 6.569503e-03
        WHEN $1 = 048 THEN 6.801176e-03
        WHEN $1 = 049 THEN 7.041019e-03
        WHEN $1 = 050 THEN 7.289320e-03
        WHEN $1 = 051 THEN 7.546378e-03
        WHEN $1 = 052 THEN 7.812500e-03
        WHEN $1 = 053 THEN 8.088007e-03
        WHEN $1 = 054 THEN 8.373230e-03
        WHEN $1 = 055 THEN 8.668512e-03
        WHEN $1 = 056 THEN 8.974206e-03
        WHEN $1 = 057 THEN 9.290681e-03
        WHEN $1 = 058 THEN 9.618316e-03
        WHEN $1 = 059 THEN 9.957505e-03
        WHEN $1 = 060 THEN 1.030866e-02
        WHEN $1 = 061 THEN 1.067219e-02
        WHEN $1 = 062 THEN 1.104854e-02
        WHEN $1 = 063 THEN 1.143817e-02
        WHEN $1 = 064 THEN 1.184154e-02
        WHEN $1 = 065 THEN 1.225913e-02
        WHEN $1 = 066 THEN 1.269144e-02
        WHEN $1 = 067 THEN 1.313901e-02
        WHEN $1 = 068 THEN 1.360235e-02
        WHEN $1 = 069 THEN 1.408204e-02
        WHEN $1 = 070 THEN 1.457864e-02
        WHEN $1 = 071 THEN 1.509276e-02
        WHEN $1 = 072 THEN 1.562500e-02
        WHEN $1 = 073 THEN 1.617601e-02
        WHEN $1 = 074 THEN 1.674646e-02
        WHEN $1 = 075 THEN 1.733702e-02
        WHEN $1 = 076 THEN 1.794841e-02
        WHEN $1 = 077 THEN 1.858136e-02
        WHEN $1 = 078 THEN 1.923663e-02
        WHEN $1 = 079 THEN 1.991501e-02
        WHEN $1 = 080 THEN 2.061731e-02
        WHEN $1 = 081 THEN 2.134438e-02
        WHEN $1 = 082 THEN 2.209709e-02
        WHEN $1 = 083 THEN 2.287634e-02
        WHEN $1 = 084 THEN 2.368307e-02
        WHEN $1 = 085 THEN 2.451825e-02
        WHEN $1 = 086 THEN 2.538289e-02
        WHEN $1 = 087 THEN 2.627801e-02
        WHEN $1 = 088 THEN 2.720471e-02
        WHEN $1 = 089 THEN 2.816408e-02
        WHEN $1 = 090 THEN 2.915728e-02
        WHEN $1 = 091 THEN 3.018551e-02
        WHEN $1 = 092 THEN 3.125000e-02
        WHEN $1 = 093 THEN 3.235203e-02
        WHEN $1 = 094 THEN 3.349292e-02
        WHEN $1 = 095 THEN 3.467405e-02
        WHEN $1 = 096 THEN 3.589682e-02
        WHEN $1 = 097 THEN 3.716272e-02
        WHEN $1 = 098 THEN 3.847326e-02
        WHEN $1 = 099 THEN 3.983002e-02
        WHEN $1 = 100 THEN 4.123462e-02
        WHEN $1 = 101 THEN 4.268876e-02
        WHEN $1 = 102 THEN 4.419417e-02
        WHEN $1 = 103 THEN 4.575268e-02
        WHEN $1 = 104 THEN 4.736614e-02
        WHEN $1 = 105 THEN 4.903651e-02
        WHEN $1 = 106 THEN 5.076577e-02
        WHEN $1 = 107 THEN 5.255603e-02
        WHEN $1 = 108 THEN 5.440941e-02
        WHEN $1 = 109 THEN 5.632815e-02
        WHEN $1 = 110 THEN 5.831456e-02
        WHEN $1 = 111 THEN 6.037102e-02
        WHEN $1 = 112 THEN 6.250000e-02
        WHEN $1 = 113 THEN 6.470406e-02
        WHEN $1 = 114 THEN 6.698584e-02
        WHEN $1 = 115 THEN 6.934809e-02
        WHEN $1 = 116 THEN 7.179365e-02
        WHEN $1 = 117 THEN 7.432544e-02
        WHEN $1 = 118 THEN 7.694653e-02
        WHEN $1 = 119 THEN 7.966004e-02
        WHEN $1 = 120 THEN 8.246924e-02
        WHEN $1 = 121 THEN 8.537752e-02
        WHEN $1 = 122 THEN 8.838835e-02
        WHEN $1 = 123 THEN 9.150536e-02
        WHEN $1 = 124 THEN 9.473229e-02
        WHEN $1 = 125 THEN 9.807301e-02
        WHEN $1 = 126 THEN 1.015315e-01
        WHEN $1 = 127 THEN 1.051121e-01
        WHEN $1 = 128 THEN 1.088188e-01
        WHEN $1 = 129 THEN 1.126563e-01
        WHEN $1 = 130 THEN 1.166291e-01
        WHEN $1 = 131 THEN 1.207420e-01
        WHEN $1 = 132 THEN 1.250000e-01
        WHEN $1 = 133 THEN 1.294081e-01
        WHEN $1 = 134 THEN 1.339717e-01
        WHEN $1 = 135 THEN 1.386962e-01
        WHEN $1 = 136 THEN 1.435873e-01
        WHEN $1 = 137 THEN 1.486509e-01
        WHEN $1 = 138 THEN 1.538931e-01
        WHEN $1 = 139 THEN 1.593201e-01
        WHEN $1 = 140 THEN 1.649385e-01
        WHEN $1 = 141 THEN 1.707550e-01
        WHEN $1 = 142 THEN 1.767767e-01
        WHEN $1 = 143 THEN 1.830107e-01
        WHEN $1 = 144 THEN 1.894646e-01
        WHEN $1 = 145 THEN 1.961460e-01
        WHEN $1 = 146 THEN 2.030631e-01
        WHEN $1 = 147 THEN 2.102241e-01
        WHEN $1 = 148 THEN 2.176376e-01
        WHEN $1 = 149 THEN 2.253126e-01
        WHEN $1 = 150 THEN 2.332582e-01
        WHEN $1 = 151 THEN 2.414841e-01
        WHEN $1 = 152 THEN 2.500000e-01
        WHEN $1 = 153 THEN 2.588162e-01
        WHEN $1 = 154 THEN 2.679434e-01
        WHEN $1 = 155 THEN 2.773924e-01
        WHEN $1 = 156 THEN 2.871746e-01
        WHEN $1 = 157 THEN 2.973018e-01
        WHEN $1 = 158 THEN 3.077861e-01
        WHEN $1 = 159 THEN 3.186402e-01
        WHEN $1 = 160 THEN 3.298770e-01
        WHEN $1 = 161 THEN 3.415101e-01
        WHEN $1 = 162 THEN 3.535534e-01
        WHEN $1 = 163 THEN 3.660214e-01
        WHEN $1 = 164 THEN 3.789291e-01
        WHEN $1 = 165 THEN 3.922920e-01
        WHEN $1 = 166 THEN 4.061262e-01
        WHEN $1 = 167 THEN 4.204482e-01
        WHEN $1 = 168 THEN 4.352753e-01
        WHEN $1 = 169 THEN 4.506252e-01
        WHEN $1 = 170 THEN 4.665165e-01
        WHEN $1 = 171 THEN 4.829682e-01
        WHEN $1 = 172 THEN 5.000000e-01
        WHEN $1 = 173 THEN 5.176325e-01
        WHEN $1 = 174 THEN 5.358867e-01
        WHEN $1 = 175 THEN 5.547847e-01
        WHEN $1 = 176 THEN 5.743492e-01
        WHEN $1 = 177 THEN 5.946036e-01
        WHEN $1 = 178 THEN 6.155722e-01
        WHEN $1 = 179 THEN 6.372803e-01
        WHEN $1 = 180 THEN 6.597540e-01
        WHEN $1 = 181 THEN 6.830201e-01
        WHEN $1 = 182 THEN 7.071068e-01
        WHEN $1 = 183 THEN 7.320428e-01
        WHEN $1 = 184 THEN 7.578583e-01
        WHEN $1 = 185 THEN 7.845841e-01
        WHEN $1 = 186 THEN 8.122524e-01
        WHEN $1 = 187 THEN 8.408964e-01
        WHEN $1 = 188 THEN 8.705506e-01
        WHEN $1 = 189 THEN 9.012505e-01
        WHEN $1 = 190 THEN 9.330330e-01
        WHEN $1 = 191 THEN 9.659363e-01
        WHEN $1 = 192 THEN 1.000000e+00
        WHEN $1 = 193 THEN 2.000000e+00
        WHEN $1 = 194 THEN 4.000000e+00
        WHEN $1 = 195 THEN 8.000000e+00
        WHEN $1 = 196 THEN 1.600000e+01
        WHEN $1 = 197 THEN 3.200000e+01
        WHEN $1 = 198 THEN 6.400000e+01
        WHEN $1 = 199 THEN 1.280000e+02
        WHEN $1 = 200 THEN 2.560000e+02
        WHEN $1 = 201 THEN 5.120000e+02
        WHEN $1 = 202 THEN 1.024000e+03
        WHEN $1 = 203 THEN 2.048000e+03
        WHEN $1 = 204 THEN 4.096000e+03
        WHEN $1 = 205 THEN 8.192000e+03
        WHEN $1 = 206 THEN 1.638400e+04
        WHEN $1 = 207 THEN 3.276800e+04
        WHEN $1 = 208 THEN 6.553600e+04
        WHEN $1 = 209 THEN 1.310720e+05
        WHEN $1 = 210 THEN 2.621440e+05
        WHEN $1 = 211 THEN 5.242880e+05
        WHEN $1 = 212 THEN 1.048576e+06
        WHEN $1 = 213 THEN 2.097152e+06
        WHEN $1 = 214 THEN 4.194304e+06
        WHEN $1 = 215 THEN 8.388608e+06
        WHEN $1 = 216 THEN 1.677722e+07
        WHEN $1 = 217 THEN 3.355443e+07
        WHEN $1 = 218 THEN 6.710886e+07
        WHEN $1 = 219 THEN 1.342177e+08
        WHEN $1 = 220 THEN 2.684355e+08
        WHEN $1 = 221 THEN 5.368709e+08
        WHEN $1 = 222 THEN 1.073742e+09
        WHEN $1 = 223 THEN 2.147484e+09
        WHEN $1 = 224 THEN 4.294967e+09
        WHEN $1 = 225 THEN 8.589935e+09
        WHEN $1 = 226 THEN 1.717987e+10
        WHEN $1 = 227 THEN 3.435974e+10
        WHEN $1 = 228 THEN 6.871948e+10
        WHEN $1 = 229 THEN 1.374390e+11
        WHEN $1 = 230 THEN 2.748779e+11
        WHEN $1 = 231 THEN 5.497558e+11
        WHEN $1 = 232 THEN 1.099512e+12
        WHEN $1 = 233 THEN 2.199023e+12
        WHEN $1 = 234 THEN 4.398047e+12
        WHEN $1 = 235 THEN 8.796093e+12
        WHEN $1 = 236 THEN 1.759219e+13
        WHEN $1 = 237 THEN 3.518437e+13
        WHEN $1 = 238 THEN 7.036874e+13
        WHEN $1 = 239 THEN 1.407375e+14
        WHEN $1 = 240 THEN 2.814750e+14
        WHEN $1 = 241 THEN 5.629500e+14
        WHEN $1 = 242 THEN 1.125900e+15
        WHEN $1 = 243 THEN 2.251800e+15
        WHEN $1 = 244 THEN 4.503600e+15
        WHEN $1 = 245 THEN 9.007199e+15
        WHEN $1 = 246 THEN 1.801440e+16
        WHEN $1 = 247 THEN 3.602880e+16
        WHEN $1 = 248 THEN 7.205759e+16
        WHEN $1 = 249 THEN 1.441152e+17
        WHEN $1 = 250 THEN 2.882304e+17
        WHEN $1 = 251 THEN 5.764608e+17
        WHEN $1 = 252 THEN 1.152922e+18
        WHEN $1 = 253 THEN 2.305843e+18
        WHEN $1 = 254 THEN 4.611686e+18
        WHEN $1 = 255 THEN 9.223372e+18
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
    IF NEW.is_l1  THEN
        -- Validate L1 Tx
        IF NEW.user_origin IS NULL OR
        NEW.from_eth_addr IS NULL OR
        NEW.from_bjj IS NULL OR
        NEW.deposit_amount IS NULL OR
        NEW.deposit_amount_f IS NULL OR
        (NOT NEW.user_origin AND NEW.batch_num IS NULL)  THEN -- If is Coordinator L1, must include batch_num
            RAISE EXCEPTION 'Invalid L1 tx: %', NEW;
        END IF;
    ELSE
        -- Validate L2 Tx
        IF NEW.batch_num IS NULL OR NEW.nonce IS NULL THEN
            RAISE EXCEPTION 'Invalid L2 tx: %', NEW;
        END IF;
        -- Set fee if it's null
        IF NEW.fee IS NULL THEN
            NEW.fee = (SELECT 0);
        END IF;
        -- Set token_id
        NEW."token_id" = (SELECT token_id FROM account WHERE idx = NEW."from_idx");
        -- Set from_{eth_addr,bjj}
        SELECT INTO NEW."from_eth_addr", NEW."from_bjj" eth_addr, bjj FROM account WHERE idx = NEW.from_idx;
    END IF;
    -- Set USD related
    SELECT INTO _value, _usd_update, _tx_timestamp 
        usd / POWER(10, decimals), usd_update, timestamp FROM token INNER JOIN block on token.eth_block_num = block.eth_block_num WHERE token_id = NEW.token_id;
    IF _usd_update - interval '24 hours' < _usd_update AND _usd_update + interval '24 hours' > _usd_update THEN
        IF _value > 0.0 THEN
            IF NEW."amount_f" > 0.0 THEN
                NEW."amount_usd" = (SELECT _value * NEW."amount_f");
                IF NOT NEW."is_l1" AND NEW."fee" > 0 THEN
                    NEW."fee_usd" = (SELECT NEW."amount_usd" * fee_percentage(NEW.fee::NUMERIC));
                END IF;
            END IF;
            IF NEW."is_l1" AND NEW."deposit_amount_f" > 0.0 THEN
                NEW."deposit_amount_usd" = (SELECT _value * NEW.deposit_amount_f);
            END IF;
        END IF;
    END IF;
    -- Set to_{eth_addr,bjj}
    IF NEW."to_idx" > 255 THEN
        SELECT INTO NEW."to_eth_addr", NEW."to_bjj" eth_addr, bjj FROM account WHERE idx = NEW."to_idx";
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
        SET item_id = upd.item_id, batch_num = NEW.batch_num 
        FROM (
            SELECT id, nextval('tx_item_id') FROM tx 
            WHERE user_origin AND NEW.forge_l1_txs_num = to_forge_l1_txs_num 
            ORDER BY position
            FOR UPDATE
        ) as upd (id, item_id)
        WHERE tx.id = upd.id; 
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
    fee_add_token DECIMAL(78,0) NOT NULL,
    forge_l1_timeout BIGINT NOT NULL,
    withdrawal_delay BIGINT NOT NULL,
    buckets BYTEA NOT NULL,
    safe_mode BOOLEAN NOT NULL
);

CREATE TABLE bucket_update (
    item_id SERIAL PRIMARY KEY,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    num_bucket BIGINT NOT NULL,
    block_stamp BIGINT NOT NULL,
    withdrawals DECIMAL(78,0) NOT NULL
);

CREATE TABLE token_exchange (
    item_id SERIAL PRIMARY KEY,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    eth_addr BYTEA NOT NULL,
    value_usd BIGINT NOT NULL
);

CREATE TABLE escape_hatch_withdrawal (
    item_id SERIAL PRIMARY KEY,
    eth_block_num BIGINT NOT NULL REFERENCES block (eth_block_num) ON DELETE CASCADE,
    who_addr BYTEA NOT NULL,
    to_addr BYTEA NOT NULL,
    token_addr BYTEA NOT NULL,
    amount DECIMAL(78,0) NOT NULL
);

CREATE TABLE auction_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    donation_address BYTEA NOT NULL,
    boot_coordinator BYTEA NOT NULL,
    boot_coordinator_url BYTEA NOT NULL,
    default_slot_set_bid BYTEA NOT NULL,
    default_slot_set_bid_slot_num BIGINT NOT NULL, -- slot_num after which the new default_slot_set_bid applies
    closed_auction_slots INT NOT NULL,
    open_auction_slots INT NOT NULL,
    allocation_ratio VARCHAR(200),
    outbidding INT NOT NULL,
    slot_deadline INT NOT NULL
);

CREATE TABLE wdelayer_vars (
    eth_block_num BIGINT PRIMARY KEY REFERENCES block (eth_block_num) ON DELETE CASCADE,
    gov_address BYTEA NOT NULL,
    emg_address BYTEA NOT NULL,
    withdrawal_delay BIGINT NOT NULL,
    emergency_start_block BIGINT NOT NULL,
    emergency_mode BOOLEAN NOT NULL
);

-- L2
CREATE TABLE tx_pool (
    tx_id BYTEA PRIMARY KEY,
    from_idx BIGINT NOT NULL,
    effective_from_eth_addr BYTEA,
    effective_from_bjj BYTEA,
    to_idx BIGINT,
    to_eth_addr BYTEA,
    to_bjj BYTEA,
    effective_to_eth_addr BYTEA,
    effective_to_bjj BYTEA,
    token_id INT NOT NULL REFERENCES token (token_id) ON DELETE CASCADE,
    amount DECIMAL(78,0) NOT NULL,
    amount_f NUMERIC NOT NULL,
    fee SMALLINT NOT NULL,
    nonce BIGINT NOT NULL,
    state CHAR(4) NOT NULL,
    info VARCHAR,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE DEFAULT timezone('utc', now()),
    batch_num BIGINT,
    rq_from_idx BIGINT,
    rq_to_idx BIGINT,
    rq_to_eth_addr BYTEA,
    rq_to_bjj BYTEA,
    rq_token_id INT,
    rq_amount DECIMAL(78,0),
    rq_fee SMALLINT,
    rq_nonce BIGINT,
    tx_type VARCHAR(40) NOT NULL,
    client_ip VARCHAR,
    external_delete BOOLEAN NOT NULL DEFAULT false
);

-- +migrate StatementBegin
CREATE FUNCTION set_pool_tx()
    RETURNS TRIGGER 
AS 
$BODY$
BEGIN
    SELECT INTO NEW."effective_from_eth_addr", NEW."effective_from_bjj" eth_addr, bjj FROM account WHERE idx = NEW."from_idx";
     -- Set to_{eth_addr,bjj}
    IF NEW.to_idx > 255 THEN
        SELECT INTO NEW."effective_to_eth_addr", NEW."effective_to_bjj" eth_addr, bjj FROM account WHERE idx = NEW."to_idx";
    ELSE
        NEW."effective_to_eth_addr" = NEW."to_eth_addr";
        NEW."effective_to_bjj" = NEW."to_bjj";
    END IF;
    RETURN NEW;
END;
$BODY$
LANGUAGE plpgsql;
-- +migrate StatementEnd
CREATE TRIGGER trigger_set_pool_tx BEFORE INSERT ON tx_pool
FOR EACH ROW EXECUTE PROCEDURE set_pool_tx();

CREATE TABLE account_creation_auth (
    eth_addr BYTEA PRIMARY KEY,
    bjj BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT timezone('utc', now())
);

CREATE TABLE node_info (
    item_id SERIAL PRIMARY KEY,
    state BYTEA,            -- object returned by GET /state
    config BYTEA,           -- Node config
    -- max_pool_txs BIGINT,    -- L2DB config
    -- min_fee NUMERIC,        -- L2DB config
    constants BYTEA         -- info of the network that is constant
);
INSERT INTO node_info(item_id) VALUES (1); -- Always have a single row that we will update

CREATE VIEW account_state AS SELECT DISTINCT idx,
first_value(nonce) OVER w AS nonce,
first_value(balance) OVER w AS balance,
first_value(eth_block_num) OVER w AS eth_block_num,
first_value(batch_num) OVER w AS batch_num
FROM account_update
window w AS (partition by idx ORDER BY item_id desc);

-- +migrate Down
-- triggers
DROP TRIGGER IF EXISTS trigger_token_usd_update ON token;
DROP TRIGGER IF EXISTS trigger_set_tx ON tx;
DROP TRIGGER IF EXISTS trigger_forge_l1_txs ON batch;
DROP TRIGGER IF EXISTS trigger_set_pool_tx ON tx_pool;
-- drop views IF EXISTS
DROP VIEW IF EXISTS account_state;
-- functions
DROP FUNCTION IF EXISTS hez_idx;
DROP FUNCTION IF EXISTS set_token_usd_update;
DROP FUNCTION IF EXISTS fee_percentage;
DROP FUNCTION IF EXISTS set_tx;
DROP FUNCTION IF EXISTS forge_l1_user_txs;
DROP FUNCTION IF EXISTS set_pool_tx;
-- drop tables IF EXISTS
DROP TABLE IF EXISTS node_info;
DROP TABLE IF EXISTS account_creation_auth;
DROP TABLE IF EXISTS tx_pool;
DROP TABLE IF EXISTS auction_vars;
DROP TABLE IF EXISTS rollup_vars;
DROP TABLE IF EXISTS escape_hatch_withdrawal;
DROP TABLE IF EXISTS bucket_update;
DROP TABLE IF EXISTS token_exchange;
DROP TABLE IF EXISTS wdelayer_vars;
DROP TABLE IF EXISTS tx;
DROP TABLE IF EXISTS exit_tree;
DROP TABLE IF EXISTS account_update;
DROP TABLE IF EXISTS account;
DROP TABLE IF EXISTS token;
DROP TABLE IF EXISTS bid;
DROP TABLE IF EXISTS batch;
DROP TABLE IF EXISTS coordinator;
DROP TABLE IF EXISTS block;
-- sequences
DROP SEQUENCE IF EXISTS tx_item_id;
