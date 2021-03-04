package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Bid is a struct that represents one bid in the PoH
type Bid struct {
	SlotNum     int64             `meddler:"slot_num"`
	BidValue    *big.Int          `meddler:"bid_value,bigint"`
	EthBlockNum int64             `meddler:"eth_block_num"`
	Bidder      ethCommon.Address `meddler:"bidder_addr"` // Coordinator reference
}

// BidCoordinator contains the coordinator info of a bid, along with the bid value
type BidCoordinator struct {
	SlotNum           int64             `meddler:"slot_num"`
	DefaultSlotSetBid [6]*big.Int       `meddler:"default_slot_set_bid,json"`
	BidValue          *big.Int          `meddler:"bid_value,bigint"`
	Bidder            ethCommon.Address `meddler:"bidder_addr"` // address of the bidder
	Forger            ethCommon.Address `meddler:"forger_addr"` // address of the forger
	URL               string            `meddler:"url"`         // URL of the coordinators API
}

// Slot contains relevant information of a slot
type Slot struct {
	SlotNum          int64
	DefaultSlotBid   *big.Int
	StartBlock       int64
	EndBlock         int64
	ForgerCommitment bool
	// BatchesLen       int
	BidValue  *big.Int
	BootCoord bool
	// Bidder, Forger and URL correspond to the winner of the slot (which is
	// not always the highest bidder).  These are the values of the
	// coordinator that is able to forge exclusively before the deadline.
	Bidder ethCommon.Address
	Forger ethCommon.Address
	URL    string
}
