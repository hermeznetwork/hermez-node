package migrations_test

import (
	"testing"

	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this migration changes length of the token name
type migrationTest0005 struct{}

func (m migrationTest0005) InsertData(db *sqlx.DB) error {
	// insert token
	const queryInsertToken = `INSERT INTO token (
		token_id, eth_block_num, eth_addr, name, symbol, decimals, usd, usd_update
	) VALUES (
		16, 0, '0x0d8775f648430679a709e98d2b0cb6250d2887ef', 'Basic Attention Token', 'BAT', 18, 0.6933, '2021-06-10T09:27:40.071789Z'
	);`

	_, err := db.Exec(queryInsertToken)
	return err
}

func (m migrationTest0005) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 16 AND
		eth_block_num = 0 AND
		eth_addr = '0x0d8775f648430679a709e98d2b0cb6250d2887ef' AND
		name = 'Basic Attention Token' AND
		decimals = 18 AND
		usd = 0.6933 AND
		usd_update = '2021-06-10T09:27:40.071789Z';
	`
	row := db.QueryRow(queryGetToken)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0005) RunAssertsAfterMigrationDown(t *testing.T, db *sqlx.DB) {
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 16 AND
		eth_block_num = 0 AND
		eth_addr = '0x0d8775f648430679a709e98d2b0cb6250d2887ef' AND
		name = 'Basic Attention Toke' AND
		decimals = 18 AND
		usd = 0.6933 AND
		usd_update = '2021-06-10T09:27:40.071789Z';
	`
	row := db.QueryRow(queryGetToken)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func TestMigration0005(t *testing.T) {
	migrationNumber := 5
	miter := migrationTest0005{}
	// Initialize an empty DB
	db, err := initCleanSQLDB()
	require.NoError(t, err)
	require.NoError(t, dbUtils.MigrationsDown(db.DB, 0))
	// Run migrations until migration to test
	require.NoError(t, runMigrationsUp(db, migrationNumber-1))
	// Insert data into table(s) before migration to check if error exists
	require.NotNil(t, miter.InsertData(db))
	// Run migration that is being tested
	require.NoError(t, runMigrationsUp(db, 1))
	// Insert data into table(s) affected by migration
	require.NoError(t, miter.InsertData(db))
	// Check that data is persisted properly after migration up
	miter.RunAssertsAfterMigrationUp(t, db)
	// Revert migration to test
	require.NoError(t, dbUtils.MigrationsDown(db.DB, 1))
	// Check that data is persisted properly after migration down
	miter.RunAssertsAfterMigrationDown(t, db)
}
