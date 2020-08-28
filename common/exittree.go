package common

import (
	"math/big"

	"github.com/iden3/go-merkletree"
)

type ExitInfo struct {
	AccountIdx  Idx
	MerkleProof *merkletree.CircomVerifierProof
	Balance     *big.Int
	Nullifier   *big.Int
}
