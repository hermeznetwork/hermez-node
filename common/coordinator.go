package common

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Coordinator represents a Hermez network coordinator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Coordinator struct {
	Forger      ethCommon.Address `meddler:"forger_addr"`   // address of the forger
	EthBlockNum int64             `meddler:"eth_block_num"` // block in which the coordinator was registered
	Withdraw    ethCommon.Address `meddler:"withdraw_addr"` // address of the withdraw
	URL         string            `meddler:"url"`           // URL of the coordinators API
}
