package common

import (
	"encoding/binary"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Token is a struct that represents an Ethereum token that is supported in Hermez network
type Token struct {
	TokenID     TokenID           `meddler:"token_id"`
	EthBlockNum int64             `meddler:"eth_block_num"` // Ethereum block number in which this token was registered
	EthAddr     ethCommon.Address `meddler:"eth_addr"`
	Name        string            `meddler:"name"`
	Symbol      string            `meddler:"symbol"`
	Decimals    uint64            `meddler:"decimals"`
	USD         float64           `meddler:"usd,zeroisnull"`
	USDUpdate   time.Time         `meddler:"usd_update,utctimez"`
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

// BigInt returns the *big.Int representation of the TokenID
func (t TokenID) BigInt() *big.Int {
	return big.NewInt(int64(t))
}
