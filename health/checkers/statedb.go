package checkers

import (
	"github.com/dimiro1/health"
	"github.com/hermeznetwork/hermez-node/db/statedb"
)

type StateDBChecker struct {
	stateDB *statedb.StateDB
}

func NewStateDBChecker(sdb *statedb.StateDB) StateDBChecker {
	return StateDBChecker{
		stateDB: sdb,
	}
}

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
		AddInfo("root", root)

	return h
}
