-- +migrate Up
CREATE SEQUENCE atomic_group_id_seq;
CREATE TABLE atomic_group_index (
    atomic_group_index INTEGER NOT NULL DEFAULT nextval('atomic_group_id_seq')

);

-- +migrate Down
DROP TABLE atomic_group_index;
DROP SEQUENCE atomic_group_id_seq;