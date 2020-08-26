package common

import (
	"math/big"
)

type ExitInfo struct {
	AccountIdx  Idx
	MerkleProof []byte
	Balance     *big.Int
	Nullifier   *big.Int
}
