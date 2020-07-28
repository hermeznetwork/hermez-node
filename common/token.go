package common

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
)

// Token is a struct that represents an Etherum token that is supported in Hermez network
type Token struct {
	ID       TokenID
	Addr     eth.Address
	Symbol   string
	Decimals uint64
}

// TokenInfo provides the price of the token in USD
type TokenInfo struct {
	TokenID     uint32
	Value       float64
	LastUpdated time.Time
}

// TokenID is the unique identifier of the token, as set in the smart contract
type TokenID uint32 // current implementation supports up to 2^32 tokens
