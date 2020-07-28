package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// Account is a struct that gives information of the holdings of an address for a specific token
type Account struct {
	EthAddr   eth.Address
	TokenID   TokenID  // effective 32 bits
	Idx       uint32   // bits = SMT levels (SMT levels needs to be decided)
	Nonce     uint64   // effective 48 bits
	Balance   *big.Int // Up to 192 bits
	PublicKey babyjub.PublicKey
}
