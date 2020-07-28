package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

// PoHState give information about the forging mechanism of the Hermez network, and the synchronization status between the coordinator and the smart contract
// WARNING: this is strongly based on the previous implementation, once the new spec is done, this may change a lot.
type PoHState struct {
	IsSynched     bool        // true if the coordinator is fully synched with the ¿PoH? smart contract
	SyncProgress  float32     // percentage of synced progress with the ¿PoH? smart contract
	CurrentSlot   SlotNum     // slot in which batches are being forged at the current time
	ContractAddr  eth.Address // Etherum address of the ¿PoH? smart contract
	BlocksPerSlot uint16      // Slot duration measured in Etherum blocks
	SlotDeadline  uint16      // Time of the slot in which another coordinator can forge if the coordinator winner has not forge any block before
	GenesisBlock  uint64      // uint64 is a guess, Etherum block in which the ¿PoH? contract was deployed
	MinBid        *big.Int    // Minimum amount that an coordinator has to bid to participate in a slot auction
}
