package common

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Coordinator represents a Hermez network coordinator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Coordinator struct {
	EthBlockNum int64             // block in which the coordinator was registered
	Forger      ethCommon.Address // address of the forger
	Withdraw    ethCommon.Address // address of the withdraw
	URL         string            // URL of the coordinators API
}
