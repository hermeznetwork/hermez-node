package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

const (
	// AuctionErrMsgCannotForge is the message returned in forge with the
	// address cannot forge
	AuctionErrMsgCannotForge = "HermezAuctionProtocol::forge: CANNOT_FORGE"
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
	GovernanceAddress ethCommon.Address `json:"governanceAddress"`
}

// SlotNum returns the slot number of a block number
func (c *AuctionConstants) SlotNum(blockNum int64) int64 {
	if blockNum >= c.GenesisBlockNum {
		return (blockNum - c.GenesisBlockNum) / int64(c.BlocksPerSlot)
	}
	// This result will be negative
	return (blockNum - c.GenesisBlockNum) / int64(c.BlocksPerSlot)
}

// SlotBlocks returns the first and the last block numbers included in that slot
func (c *AuctionConstants) SlotBlocks(slotNum int64) (int64, int64) {
	startBlock := c.GenesisBlockNum + slotNum*int64(c.BlocksPerSlot)
	endBlock := startBlock + int64(c.BlocksPerSlot) - 1
	return startBlock, endBlock
}

// RelativeBlock returns the relative block number within the slot where the
// block number belongs
func (c *AuctionConstants) RelativeBlock(blockNum int64) int64 {
	slotNum := c.SlotNum(blockNum)
	return blockNum - (c.GenesisBlockNum + (slotNum * int64(c.BlocksPerSlot)))
}

// AuctionVariables are the variables of the Auction Smart Contract
type AuctionVariables struct {
	EthBlockNum int64 `meddler:"eth_block_num"`
	// Donation Address
	DonationAddress ethCommon.Address `meddler:"donation_address" validate:"required"`
	// Boot Coordinator Address
	BootCoordinator ethCommon.Address `meddler:"boot_coordinator" validate:"required"`
	// Boot Coordinator URL
	BootCoordinatorURL string `meddler:"boot_coordinator_url" validate:"required"`
	// The minimum bid value in a series of 6 slots
	DefaultSlotSetBid [6]*big.Int `meddler:"default_slot_set_bid,json" validate:"required"`
	// SlotNum at which the new default_slot_set_bid applies
	DefaultSlotSetBidSlotNum int64 `meddler:"default_slot_set_bid_slot_num"`
	// Distance (#slots) to the closest slot to which you can bid ( 2 Slots = 2 * 40 Blocks = 20 min )
	ClosedAuctionSlots uint16 `meddler:"closed_auction_slots" validate:"required"`
	// Distance (#slots) to the farthest slot to which you can bid (30 days = 4320 slots )
	OpenAuctionSlots uint16 `meddler:"open_auction_slots" validate:"required"`
	// How the HEZ tokens deposited by the slot winner are distributed (Burn: 40% - Donation:
	// 40% - HGT: 20%)
	AllocationRatio [3]uint16 `meddler:"allocation_ratio,json" validate:"required"`
	// Minimum outbid (percentage) over the previous one to consider it valid
	Outbidding uint16 `meddler:"outbidding" validate:"required"`
	// Number of blocks at the end of a slot in which any coordinator can forge if the winner
	// has not forged one before
	SlotDeadline uint8 `meddler:"slot_deadline" validate:"required"`
}

// Copy returns a deep copy of the Variables
func (v *AuctionVariables) Copy() *AuctionVariables {
	vCpy := *v
	for i := range v.DefaultSlotSetBid {
		vCpy.DefaultSlotSetBid[i] = CopyBigInt(v.DefaultSlotSetBid[i])
	}
	return &vCpy
}
