package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// WIP this will be from hermeznetwork/common
type Tx struct {
	FromAx          [32]byte
	FromAy          [32]byte
	FromEthAddr     ethCommon.Address
	ToAx            [32]byte
	ToAy            [32]byte
	ToEthAddr       ethCommon.Address
	OnChain         bool
	RqOffset        []byte
	NewAccount      bool
	TokenID         uint32
	LoadAmount      [3]byte
	Amount          [3]byte
	Nonce           uint64
	UserFee         uint8
	UserFeeAbsolute uint64
	R8x             [32]byte
	R8y             [32]byte
	S               [32]byte
	RqTxData        [32]byte
}

// WIP this will be from hermeznetwork/common
type Account struct {
	EthAddr ethCommon.Address
	TokenID uint32
	Idx     uint32
	Nonce   uint64
	Balance *big.Int
	// Ax, Ay
}
