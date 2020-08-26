package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

type RollupVars struct {
	EthBlockNum    uint64
	ForgeL1Timeout *big.Int
	FeeL1UserTx    *big.Int
	FeeAddToken    *big.Int
	TokensHEZ      eth.Address
	Governance     eth.Address
}

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

type MinBidSlots [6]uint

type AllocationRatio struct {
	Donation uint
	Burn     uint
	Forger   uint
}
