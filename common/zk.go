package common

import (
	"math/big"
)

type ZKInputs struct {
	InitialIdx    uint64
	OldStRoot     Hash
	FeePlanCoins  *big.Int
	FeeTotals     *big.Int
	PubEthAddress *big.Int

	ImStateRoot []Hash
	ImExitRoot  []Hash

	ImOnChainHash []Hash
	ImOnChain     []*big.Int
	TxData        []*big.Int

	FromIdx     []uint64
	ToIdX       []uint64
	ToAx        []*big.Int
	ToAy        []*big.Int
	ToEthAddr   []*big.Int
	FromEthAddr []*big.Int
	FromAx      []*big.Int
	FromAy      []*big.Int

	RqTxData   []*big.Int
	LoadAmount []*big.Int

	S   []*big.Int
	R8x []*big.Int
	R8y []*big.Int

	Ax1       []*big.Int
	Ay1       []*big.Int
	Amount1   []*big.Int
	Nonce1    []*big.Int
	EthAddr1  []*big.Int
	Siblings1 [][]*big.Int
	IsOld01   []*big.Int `json:"isOld0_1"`
	OldKey1   []*big.Int
	OldValue1 []*big.Int

	Ax2       []*big.Int
	Ay2       []*big.Int
	Amount2   []*big.Int
	Nonce2    []*big.Int
	EthAddr2  []*big.Int
	Siblings2 [][]*big.Int
	IsOld02   []*big.Int `json:"isOld0_2"`
	OldKey2   []*big.Int
	OldValue2 []*big.Int
}
