package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Bid is a struct that represents one bid in the PoH
type Bid struct {
	SlotNum     SlotNum           `meddler:"slot_num"`
	ForgerAddr  ethCommon.Address `meddler:"forger_addr"` // Coordinator reference
	BidValue    *big.Int          `meddler:"bid_value,bigint"`
	EthBlockNum uint64            `meddler:"eth_block_num"`
}
