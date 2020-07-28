package common

// Slot represents a slot of the Hermez network
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type Slot struct {
	SlotNum       SlotNum
	StartingBlock uint64      // Etherum block in which the slot starts
	Forger        Coordinator // Current Operaror winner information
}

// SlotNum identifies a slot
type SlotNum uint32
