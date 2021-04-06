package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

// API request struct for getting batches
type GetBatchesAPIRequest struct {
	MinBatchNum *uint
	MaxBatchNum *uint
	SlotNum     *uint
	ForgerAddr  *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
