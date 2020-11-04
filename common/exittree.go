package common

import (
	"math/big"

	"github.com/iden3/go-merkletree"
)

// ExitInfo represents the ExitTree Leaf data
type ExitInfo struct {
	BatchNum    BatchNum                        `meddler:"batch_num"`
	AccountIdx  Idx                             `meddler:"account_idx"`
	MerkleProof *merkletree.CircomVerifierProof `meddler:"merkle_proof,json"`
	Balance     *big.Int                        `meddler:"balance,bigint"`
	// InstantWithdrawn is the ethBlockNum in which the exit is withdrawn
	// instantly.  nil means this hasn't happened.
	InstantWithdrawn *int64 `meddler:"instant_withdrawn"`
	// DelayedWithdrawRequest is the ethBlockNum in which the exit is
	// requested to be withdrawn from the delayedWithdrawn smart contract.
	// nil means this hasn't happened.
	DelayedWithdrawRequest *int64 `meddler:"delayed_withdraw_request"`
	// DelayedWithdrawn is the ethBlockNum in which the exit is withdrawn
	// from the delayedWithdrawn smart contract.  nil means this hasn't
	// happened.
	DelayedWithdrawn *int64 `meddler:"delayed_withdrawn"`
}

// WithdrawInfo represents a withdraw action to the rollup
type WithdrawInfo struct {
	Idx             Idx
	NumExitRoot     BatchNum
	InstantWithdraw bool
}
