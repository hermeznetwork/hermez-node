package coordinator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/prover"
	"github.com/hermeznetwork/tracerr"
)

// Status is used to mark the status of the batch
type Status string

const (
	// StatusPending marks the Tx as Pending
	StatusPending Status = "pending"
	// StatusForged marks the batch as forged internally
	StatusForged Status = "forged"
	// StatusProof marks the batch as proof calculated
	StatusProof Status = "proof"
	// StatusSent marks the EthTx as Sent
	StatusSent Status = "sent"
	// StatusMined marks the EthTx as Mined
	StatusMined Status = "mined"
	// StatusFailed marks the EthTx as Failed
	StatusFailed Status = "failed"
)

// BatchInfo contans the Batch information
type BatchInfo struct {
	BatchNum              common.BatchNum
	ServerProof           prover.Client
	ZKInputs              *common.ZKInputs
	Proof                 *prover.Proof
	PublicInputs          []*big.Int
	L1Batch               bool
	VerifierIdx           uint8
	L1UserTxsExtra        []common.L1Tx
	L1CoordTxs            []common.L1Tx
	L1CoordinatorTxsAuths [][]byte
	L2Txs                 []common.L2Tx
	CoordIdxs             []common.Idx
	ForgeBatchArgs        *eth.RollupForgeBatchArgs
	// FeesInfo
	Status  Status
	EthTx   *types.Transaction
	Receipt *types.Receipt
}

// DebugStore is a debug function to store the BatchInfo as a json text file in
// storePath
func (b *BatchInfo) DebugStore(storePath string) error {
	batchJSON, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	oldStateRoot := "null"
	if b.ZKInputs != nil && b.ZKInputs.OldStateRoot != nil {
		oldStateRoot = b.ZKInputs.OldStateRoot.String()
	}
	filename := fmt.Sprintf("%010d-%s.json", b.BatchNum, oldStateRoot)
	// nolint reason: 0640 allows rw to owner and r to group
	//nolint:gosec
	return ioutil.WriteFile(path.Join(storePath, filename), batchJSON, 0640)
}
