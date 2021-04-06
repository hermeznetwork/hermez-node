package requests

import ethCommon "github.com/ethereum/go-ethereum/common"

type GetCoordinatorsAPIRequest struct {
	BidderAddr *ethCommon.Address
	ForgerAddr *ethCommon.Address

	FromItem *uint
	Limit    *uint
	Order    string
}
