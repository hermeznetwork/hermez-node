package common

import (
	"math/big"
)

// Slot represents a slot of the Hermez network
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Slot struct {
	SlotNum       SlotNum
	StartingBlock uint64      // Ethereum block in which the slot starts
	Forger        Coordinator // Current Operaror winner information
}

// SlotMinPrice is the policy of minimum prices for strt bidding in the slots
type SlotMinPrice struct {
	EthBlockNum uint64 // Etherum block in which the min price was updated
	MinPrices   [6]big.Int
}

// GetMinPrice returns the minimum bid to enter the auction for a specific slot
func (smp *SlotMinPrice) GetMinPrice(slotNum SlotNum) *big.Int {
	// TODO
	return nil
}

// SlotNum identifies a slot
type SlotNum uint32
