package requests

import "github.com/hermeznetwork/hermez-node/common"

type GetTokensAPIRequest struct {
	Ids     []common.TokenID
	Symbols []string
	Name    string

	FromItem *uint
	Limit    *uint
	Order    string
}
