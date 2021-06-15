package migrations_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// This migration creates the fiat table

type migrationTest0004 struct{}

func (m migrationTest0004) InsertData(db *sqlx.DB) error {
	return nil
}

func (m migrationTest0004) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	//Insert data in the fiat table
	const queryInsertFiat = `INSERT INTO fiat(currency, base_currency, price) VALUES ('EUR','USD',0.82);`
	_, err := db.Exec(queryInsertFiat)
	assert.NoError(t, err)
	const queryGetNumberItemsFiat = `select count(*) from fiat;`
	row := db.QueryRow(queryGetNumberItemsFiat)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0004) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	// check that the fiat table is not created and I can't insert data
	const queryInsertFiat = `INSERT INTO fiat(currency, base_currency, price) VALUES ('CNY','USD',6.4306);`
	_, err := db.Exec(queryInsertFiat)
	if assert.Error(t, err) {
		assert.Equal(t, `pq: relation "fiat" does not exist`, err.Error())
	}
}

func TestMigration0004(t *testing.T) {
	runMigrationTest(t, 4, migrationTest0004{})
}
