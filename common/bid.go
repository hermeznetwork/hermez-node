package common

import (
	"math/big"
)

// Bid is a struct that represents one bid in the PoH
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Bid struct {
	SlotNum      SlotNum  // Slot in which the bid is done
	InfoOperator Operator // Operaror bidder information
	Amount       *big.Int
}
