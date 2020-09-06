package common

import (
	"math/big"

	"github.com/iden3/go-merkletree"
)

// ExitInfo represents the ExitTree Leaf data
type ExitInfo struct {
	AccountIdx  Idx
	MerkleProof *merkletree.CircomVerifierProof
	Balance     *big.Int
	Nullifier   *big.Int
}
