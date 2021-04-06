package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

// API request struct for getting bids
type GetBidsAPIRequest struct {
	SlotNum    *int64
	BidderAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
