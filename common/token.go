package common

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/tracerr"
)

// tokenIDBytesLen defines the length of the TokenID byte array representation
const tokenIDBytesLen = 4

// Token is a struct that represents an Ethereum token that is supported in Hermez network
type Token struct {
	TokenID TokenID `json:"id" meddler:"token_id"`
	// EthBlockNum indicates the Ethereum block number in which this token was registered
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthAddr     ethCommon.Address `json:"ethereumAddress" meddler:"eth_addr"`
	Name        string            `json:"name" meddler:"name"`
	Symbol      string            `json:"symbol" meddler:"symbol"`
	Decimals    uint64            `json:"decimals" meddler:"decimals"`
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
	binary.BigEndian.PutUint32(tokenIDBytes[:], uint32(t))
	return tokenIDBytes[:]
}

// BigInt returns the *big.Int representation of the TokenID
func (t TokenID) BigInt() *big.Int {
	return big.NewInt(int64(t))
}

// TokenIDFromBytes returns TokenID from a byte array
func TokenIDFromBytes(b []byte) (TokenID, error) {
	if len(b) != tokenIDBytesLen {
		return 0, tracerr.Wrap(fmt.Errorf("can not parse TokenID, bytes len %d, expected 4",
			len(b)))
	}
	tid := binary.BigEndian.Uint32(b[:4])
	return TokenID(tid), nil
}

// TokenIDFromBigInt returns a TokenID with the value of the given *big.Int
func TokenIDFromBigInt(b *big.Int) TokenID {
	return TokenID(b.Int64())
}
