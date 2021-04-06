package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

type GetBidsAPIRequest struct {
	SlotNum    *int64
	BidderAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
