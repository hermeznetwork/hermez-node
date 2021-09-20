package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type migrationTest0012 struct{}

func (m migrationTest0012) InsertData(db *sqlx.DB) error {
	return nil
}

func (m migrationTest0012) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	const queryGetProvers = `SELECT COUNT(*) FROM provers;`
	row := db.QueryRow(queryGetProvers)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 0, result)

	insert := `
    INSERT INTO provers (public_dns, instance_id) VALUES ('http://some-url-on-aws:9080', 'id-some-instance-id2');
    `
	_, err := db.Exec(insert)
	assert.NoError(t, err)

	row = db.QueryRow(queryGetProvers)
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0012) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	var result int
	const queryGetProvers = `SELECT COUNT(*) FROM provers;`
	row := db.QueryRow(queryGetProvers)
	err := row.Scan(&result)
	assert.Equal(t, `pq: relation "provers" does not exist`, err.Error())
}

func TestMigration0012(t *testing.T) {
	runMigrationTest(t, 12, migrationTest0012{})
}
