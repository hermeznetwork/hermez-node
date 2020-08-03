package common

import (
	"math/big"
)

// Bid is a struct that represents one bid in the PoH
type Bid struct {
	SlotNum     SlotNum // Slot in which the bid is done
	Coordinator         // Coordinator bidder information
	BidValue    *big.Int
	EthBlockNum uint64
	Won         bool // boolean flag that tells that this is the final winning bid of this SlotNum
}
