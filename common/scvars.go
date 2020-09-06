package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

// RollupVars contain the Rollup smart contract variables
type RollupVars struct {
	EthBlockNum    uint64
	ForgeL1Timeout *big.Int
	FeeL1UserTx    *big.Int
	FeeAddToken    *big.Int
	TokensHEZ      eth.Address
	Governance     eth.Address
}

// AuctionVars contain the Auction smart contract variables
type AuctionVars struct {
	EthBlockNum       uint64
	SlotDeadline      uint
	CloseAuctionSlots uint
	OpenAuctionSlots  uint
	Governance        eth.Address
	MinBidSlots       MinBidSlots
	Outbidding        int
	DonationAddress   eth.Address
	GovernanceAddress eth.Address
	AllocationRatio   AllocationRatio
}

// MinBidSlots TODO
type MinBidSlots [6]uint

// AllocationRatio TODO
type AllocationRatio struct {
	Donation uint
	Burn     uint
	Forger   uint
}
