package test

import "github.com/jmoiron/sqlx"

// CleanL2DB deletes 'tx_pool' and 'account_creation_auth' from the given DB
func CleanL2DB(db *sqlx.DB) {
	if _, err := db.Exec("DELETE FROM tx_pool"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM account_creation_auth"); err != nil {
		panic(err)
	}
}
