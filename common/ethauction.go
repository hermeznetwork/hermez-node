package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// AuctionConstants are the constants of the Rollup Smart Contract
type AuctionConstants struct {
	// Blocks per slot
	BlocksPerSlot uint8 `json:"blocksPerSlot"`
	// Minimum bid when no one has bid yet
	InitialMinimalBidding *big.Int `json:"initialMinimalBidding"`
	// First block where the first slot begins
	GenesisBlockNum int64 `json:"genesisBlockNum"`
	// ERC777 token with which the bids will be made
	TokenHEZ ethCommon.Address `json:"tokenHEZ"`
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address `json:"hermezRollup"`
	// Hermez Governanze Token smartcontract address who controls some parameters and collects HEZ fee
	// Only for test
	GovernanceAddress ethCommon.Address `json:"governanceAddress"`
}

// AuctionVariables are the variables of the Auction Smart Contract
type AuctionVariables struct {
	EthBlockNum int64 `json:"ethereumBlockNum" meddler:"eth_block_num"`
	// Boot Coordinator Address
	DonationAddress ethCommon.Address `json:"donationAddress" meddler:"donation_address" validate:"required"`
	// Boot Coordinator Address
	BootCoordinator ethCommon.Address `json:"bootCoordinator" meddler:"boot_coordinator" validate:"required"`
	// The minimum bid value in a series of 6 slots
	DefaultSlotSetBid [6]*big.Int `json:"defaultSlotSetBid" meddler:"default_slot_set_bid,json" validate:"required"`
	// Distance (#slots) to the closest slot to which you can bid ( 2 Slots = 2 * 40 Blocks = 20 min )
	ClosedAuctionSlots uint16 `json:"closedAuctionSlots" meddler:"closed_auction_slots" validate:"required"`
	// Distance (#slots) to the farthest slot to which you can bid (30 days = 4320 slots )
	OpenAuctionSlots uint16 `json:"openAuctionSlots" meddler:"open_auction_slots" validate:"required"`
	// How the HEZ tokens deposited by the slot winner are distributed (Burn: 40% - Donation: 40% - HGT: 20%)
	AllocationRatio [3]uint16 `json:"allocationRatio" meddler:"allocation_ratio,json" validate:"required"`
	// Minimum outbid (percentage) over the previous one to consider it valid
	Outbidding uint16 `json:"outbidding" meddler:"outbidding" validate:"required"`
	// Number of blocks at the end of a slot in which any coordinator can forge if the winner has not forged one before
	SlotDeadline uint8 `json:"slotDeadline" meddler:"slot_deadline" validate:"required"`
}

// Copy returns a deep copy of the Variables
func (v *AuctionVariables) Copy() *AuctionVariables {
	vCpy := *v
	for i := range v.DefaultSlotSetBid {
		vCpy.DefaultSlotSetBid[i] = new(big.Int).SetBytes(v.DefaultSlotSetBid[i].Bytes())
	}
	return &vCpy
}
