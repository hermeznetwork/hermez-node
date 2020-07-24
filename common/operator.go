package common

import (
	eth "github.com/ethereum/go-ethereum/common"
)

// Operator represents a Hermez network operator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Operator struct {
	Forger      eth.Address // address of the forger
	Beneficiary eth.Address // address of the beneficiary
	Withdraw    eth.Address // address of the withdraw
	URL         string      // URL of the operators API
}
