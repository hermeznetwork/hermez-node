package checkers

import (
	"database/sql"

	"github.com/dimiro1/health"
	dbHealth "github.com/dimiro1/health/db"
)

// PostgresChecker struct to check current status of the db
type PostgresChecker struct {
	postgresChecker dbHealth.Checker
}

// NewCheckerWithDB creates new instance of the PostgresChecker
func NewCheckerWithDB(db *sql.DB) PostgresChecker {
	return PostgresChecker{
		postgresChecker: dbHealth.NewPostgreSQLChecker(db),
	}
}

// Check function check is db is responding and returns status, version of db and id of the last migration
func (c PostgresChecker) Check() health.Health {
	h := c.postgresChecker.Check()

	q := `SELECT id FROM gorp_migrations ORDER BY id DESC LIMIT 1`
	row := c.postgresChecker.DB.QueryRow(q)
	var id string
	err := row.Scan(&id)
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	}

	h.Up().AddInfo("last_migration", id)

	return h
}
