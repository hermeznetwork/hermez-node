package common

import (
	"encoding/binary"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Token is a struct that represents an Ethereum token that is supported in Hermez network
type Token struct {
	TokenID     TokenID
	EthAddr     ethCommon.Address
	Name        string
	Symbol      string
	Decimals    uint64
	EthTxHash   ethCommon.Hash // Ethereum TxHash in which this token was registered
	EthBlockNum uint64         // Ethereum block number in which this token was registered
}

// TokenInfo provides the price of the token in USD
type TokenInfo struct {
	TokenID     uint32
	Value       float64
	LastUpdated time.Time
}

// TokenID is the unique identifier of the token, as set in the smart contract
type TokenID uint32 // current implementation supports up to 2^32 tokens

// Bytes returns a byte array of length 4 representing the TokenID
func (t TokenID) Bytes() []byte {
	var tokenIDBytes [4]byte
	binary.LittleEndian.PutUint32(tokenIDBytes[:], uint32(t))
	return tokenIDBytes[:]
}
