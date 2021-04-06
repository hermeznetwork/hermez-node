package requests

import "github.com/hermeznetwork/hermez-node/common"

// GetTokensAPIRequest is an API request struct for getting tokens
type GetTokensAPIRequest struct {
	Ids     []common.TokenID
	Symbols []string
	Name    string

	FromItem *uint
	Limit    *uint
	Order    string
}
