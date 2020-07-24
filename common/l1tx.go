package common

import "math/big"

// L1Tx is a struct that represents an already forged L1 tx
// WARNING: this struct is very unclear and a complete guess
type L1Tx struct {
	Tx
	Ax                 *big.Int // Ax is the x coordinate of the BabyJubJub curve point
	Ay                 *big.Int // Ay is the y coordinate of the BabyJubJub curve point
	LoadAmount         *big.Int // amount transfered from L1 -> L2
	Mined              bool
	BlockNum           uint64
	ToForgeL1TxsNumber uint32
}
