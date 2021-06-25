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

func (m migrationTest0005) InsertDataWithLongName(db *sqlx.DB) error {
	// insert token
	/* #nosec */
	const queryInsertToken = `INSERT INTO token (
		token_id, eth_block_num, eth_addr, name, symbol, decimals, usd, usd_update
	) VALUES (
		16, 0, '0x0d8775f648430679a709e98d2b0cb6250d2887ef', 'Basic Attention Token', 'BAT', 18, 0.6933, '2021-06-10T09:27:40.071789Z'
	);`

	_, err := db.Exec(queryInsertToken)
	return err
}

func (m migrationTest0005) InsertData(db *sqlx.DB) error {
	// insert token
	/* #nosec */
	const queryInsertToken = `INSERT INTO token (
		token_id, eth_block_num, eth_addr, name, symbol, decimals, usd, usd_update
	) VALUES (
		17, 0, '0x514910771af9ca656af840dff83e8264ecf986ca', 'ChainLink Token', 'LINK', 18, 18.27, '2021-06-10T09:27:40.071789Z'
	);`

	_, err := db.Exec(queryInsertToken)
	return err
}

func (m migrationTest0005) RunAssertsAfterMigrationUpWithLongName(t *testing.T, db *sqlx.DB) {
	/* #nosec */
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 16 AND
		eth_block_num = 0 AND
		eth_addr = '0x0d8775f648430679a709e98d2b0cb6250d2887ef' AND
		name = 'Basic Attention Token' AND
        symbol = 'BAT' AND
		decimals = 18 AND
		usd = 0.6933 AND
		usd_update = '2021-06-10T09:27:40.071789Z';
	`
	row := db.QueryRow(queryGetToken)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0005) RunAssertsAfterMigrationUp(t *testing.T, db *sqlx.DB) {
	/* #nosec */
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 17 AND
		eth_block_num = 0 AND
		eth_addr = '0x514910771af9ca656af840dff83e8264ecf986ca' AND
		name = 'ChainLink Token' AND
        symbol = 'LINK' AND
		decimals = 18 AND
		usd = 18.27 AND
		usd_update = '2021-06-10T09:27:40.071789Z';
	`
	row := db.QueryRow(queryGetToken)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func (m migrationTest0005) RunAssertsAfterMigrationDownWithLongName(t *testing.T, db *sqlx.DB) {
	/* #nosec */
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 16 AND
		eth_block_num = 0 AND
		eth_addr = '0x0d8775f648430679a709e98d2b0cb6250d2887ef' AND
		name = 'Basic Attention Toke' AND
        symbol = 'BAT' AND
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
	/* #nosec */
	const queryGetToken = `SELECT COUNT(*) FROM token WHERE 
		token_id = 17 AND
		eth_block_num = 0 AND
		eth_addr = '0x514910771af9ca656af840dff83e8264ecf986ca' AND
		name = 'ChainLink Token' AND
        symbol = 'LINK' AND
		decimals = 18 AND
		usd = 18.27 AND
		usd_update = '2021-06-10T09:27:40.071789Z';
	`
	row := db.QueryRow(queryGetToken)
	var result int
	assert.NoError(t, row.Scan(&result))
	assert.Equal(t, 1, result)
}

func runMigration0005TestWithLongName(t *testing.T, migrationNumber int) {
	miter := migrationTest0005{}
	// Initialize an empty DB
	db, err := initCleanSQLDB()
	require.NoError(t, err)
	require.NoError(t, dbUtils.MigrationsDown(db.DB, 0))
	// Run migrations until migration to test
	require.NoError(t, runMigrationsUp(db, migrationNumber-1))
	// Insert data into table(s) before migration to check if error exists
	require.NotNil(t, miter.InsertDataWithLongName(db))
	// Run migration that is being tested
	require.NoError(t, runMigrationsUp(db, 1))
	// Insert data into table(s) affected by migration
	require.NoError(t, miter.InsertDataWithLongName(db))
	// Check that data is persisted properly after migration up
	miter.RunAssertsAfterMigrationUpWithLongName(t, db)
	// Revert migration to test
	require.NoError(t, dbUtils.MigrationsDown(db.DB, 1))
	// Check that data is persisted properly after migration down
	miter.RunAssertsAfterMigrationDownWithLongName(t, db)
}

func TestMigration0005(t *testing.T) {
	migrationNumber := 5
	runMigration0005TestWithLongName(t, migrationNumber)
	runMigrationTest(t, migrationNumber, migrationTest0005{})
}
