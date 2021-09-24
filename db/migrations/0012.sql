-- +migrate Up
CREATE TABLE mt_nodes (
    mt_id BIGINT,
    key BYTEA,
    type SMALLINT NOT NULL,
    child_l BYTEA,
    child_r BYTEA,
    entry BYTEA,
    created_at BIGINT,
    deleted_at BIGINT,
    PRIMARY KEY(mt_id, key)
);

CREATE TABLE mt_roots (
    mt_id BIGINT PRIMARY KEY,
    key BYTEA,
    created_at BIGINT,
    deleted_at BIGINT
);


-- +migrate Down
DROP TABLE mt_nodes;
DROP TABLE mt_roots;
