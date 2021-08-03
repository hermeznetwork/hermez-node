package checkers

import (
	"github.com/dimiro1/health"
	"github.com/hermeznetwork/hermez-node/db/statedb"
)

// StateDBChecker struct for state db connection checker
type StateDBChecker struct {
	stateDB *statedb.StateDB
}

// NewStateDBChecker init state db connection checker
func NewStateDBChecker(sdb *statedb.StateDB) StateDBChecker {
	return StateDBChecker{
		stateDB: sdb,
	}
}

// Check state db health
func (sdb StateDBChecker) Check() health.Health {
	h := health.NewHealth()

	batchNum, err := sdb.stateDB.LastGetCurrentBatch()
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	}

	root, err := sdb.stateDB.LastMTGetRoot()
	if err != nil {
		h.Down().AddInfo("error", err.Error())
		return h
	}

	h.Up().
		AddInfo("batchNum", batchNum).
		AddInfo("root", root.String())

	return h
}
