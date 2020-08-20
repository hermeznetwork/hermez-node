package common

import (
	"math/big"
)

type ExitTreeLeaf struct {
	AccountIdx  Idx
	MerkleProof []byte
	Amount      *big.Int
	Nullifier   *big.Int
}
