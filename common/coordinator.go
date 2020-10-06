package common

import (
	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Coordinator represents a Hermez network coordinator who wins an auction for an specific slot
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Coordinator struct {
	Bidder      ethCommon.Address `meddler:"bidder_addr"`   // address of the bidder
	Forger      ethCommon.Address `meddler:"forger_addr"`   // address of the forger
	EthBlockNum int64             `meddler:"eth_block_num"` // block in which the coordinator was registered
	URL         string            `meddler:"url"`           // URL of the coordinators API
}
