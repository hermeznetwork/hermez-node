package test

import (
	"os"
	"strconv"
	"testing"

	dbUtils "github.com/hermeznetwork/hermez-node/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

// AssertUSD asserts pointers to float64, and checks that they are equal
// with a tolerance of 0.01%. After that, the actual value is setted to the expected value
// in order to be able to perform further assertions using the standar assert functions.
func AssertUSD(t *testing.T, expected, actual *float64) {
	if actual == nil {
		assert.Equal(t, expected, actual)
		return
	}
	if *expected < *actual {
		assert.InEpsilon(t, *actual, *expected, 0.0001)
	} else if *expected > *actual {
		assert.InEpsilon(t, *expected, *actual, 0.0001)
	}
	*expected = *actual
}

// WipeDB redo all the migrations of the SQL DB (HistoryDB and L2DB),
// efectively recreating the original state
func WipeDB(db *sqlx.DB) {
	if err := dbUtils.MigrationsDown(db.DB); err != nil {
		panic(err)
	}
	if err := dbUtils.MigrationsUp(db.DB); err != nil {
		panic(err)
	}
}

// InitTestSQLDB open test PostgreSQL database
func InitTestSQLDB() (*sqlx.DB, error) {
	host := os.Getenv("PGHOST")
	if host == "" {
		host = "localhost"
	}
	port, _ := strconv.Atoi(os.Getenv("PGPORT"))
	if port == 0 {
		port = 5432
	}
	user := os.Getenv("PGUSER")
	if user == "" {
		user = "hermez"
	}
	pass := os.Getenv("PGPASSWORD")
	if pass == "" {
		panic("No PGPASSWORD envvar specified")
	}
	dbname := os.Getenv("PGDATABASE")
	if dbname == "" {
		dbname = "hermez"
	}
	return dbUtils.InitSQLDB(port, host, user, pass, dbname)
}
