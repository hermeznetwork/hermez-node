package common

import (
	eth "github.com/ethereum/go-ethereum/common"
)

// Coordinator represents a Hermez network coordinator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Coordinator struct {
	CoordinatorID CoordinatorID
	Forger        eth.Address // address of the forger
	Beneficiary   eth.Address // address of the beneficiary
	Withdraw      eth.Address // address of the withdraw
	URL           string      // URL of the coordinators API
}

// CoordinatorID is use to identify a Hermez coordinator
type CoordinatorID uint64
