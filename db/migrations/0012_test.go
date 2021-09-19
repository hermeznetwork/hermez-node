package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `gas_price` and `gas_used` on `batch` table

type migrationTest0012 struct{}

func (m migrationTest0012) InsertData(db *sqlx.DB) error {
	// insert tx
	const queryInsert = `
    INSERT INTO provers (public_dns, instance_id) VALUES ('http://some-url-on-aws:9080', 'id-some-instance-id');
	`
	_, err := db.Exec(queryInsert)
	return err
}

func (m migrationTest0012) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	const queryGetProvers = `SELECT COUNT(*) FROM provers;`
	row := db.QueryRow(queryGetProvers)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)

	insert := `
    INSERT INTO provers (public_dns, instance_id) VALUES ('http://some-url-on-aws:9080', 'id-some-instance-id2');
    `
	_, err := db.Exec(insert)
	assert.NoError(t, err)
}

func (m migrationTest0012) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the batch inserted in previous step is persisted with same content
	var result int
	const queryGetProvers = `SELECT COUNT(*) FROM provers;`
	row := db.QueryRow(queryGetProvers)
	assert.Equal(t, `pq: relation "provers" does not exist`, row.Scan(&result).Error())
}

func TestMigration0012(t *testing.T) {
	runMigrationTest(t, 12, migrationTest0012{})
}
