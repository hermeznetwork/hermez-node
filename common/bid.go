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
