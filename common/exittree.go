package common

import (
	"math/big"

	"github.com/iden3/go-merkletree"
)

// ExitInfo represents the ExitTree Leaf data
type ExitInfo struct {
	BatchNum    BatchNum                        `meddler:"batch_num"`
	Withdrawn   *big.Int                        `meddler:"withdrawn"`
	AccountIdx  Idx                             `meddler:"account_idx"`
	MerkleProof *merkletree.CircomVerifierProof `meddler:"merkle_proof"`
	Balance     *big.Int                        `meddler:"balance"`
	Nullifier   *big.Int                        `meddler:"nulifier"`
}
