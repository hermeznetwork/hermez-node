package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration adds the column `rq_tx_id` on `tx_pool`

type migrationTest0007 struct{}

func (m migrationTest0007) InsertData(db *sqlx.DB) error {
	const queryInsertDefaults = `INSERT INTO atomic_group_index DEFAULT VALUES;`
	_, err := db.Exec(queryInsertDefaults)
	return err
}

func (m migrationTest0007) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	const queryQueryDefaults = `SELECT COUNT(*) FROM atomic_group_index WHERE atomic_group_index.atomic_group_index = 1;`
	row := db.QueryRow(queryQueryDefaults)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0007) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	const queryQueryDefaults = `SELECT COUNT(*) FROM atomic_group_index WHERE atomic_group_index.atomic_group_index = 1;`
	row := db.QueryRow(queryQueryDefaults)
	assert.Equal(t, `pq:  relation "atomic_group_index" does not exist`, row.Scan(&result).Error())
}

func TestMigration0007(t *testing.T) {
	runMigrationTest(t, 7, migrationTest0007{})
}
