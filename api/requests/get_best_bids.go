package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

// API request struct for getting best bids
type GetBestBidsAPIRequest struct {
	MinSlotNum *int64
	MaxSlotNum *int64
	BidderAddr *ethCommon.Address

	Limit *uint
	Order string
}
