package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

type GetBatchesAPIRequest struct {
	MinBatchNum *uint
	MaxBatchNum *uint
	SlotNum     *uint
	ForgerAddr  *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
