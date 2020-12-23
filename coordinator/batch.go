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

// TxStatus is used to mark the status of an ethereum transaction
type TxStatus string

const (
	// TxStatusPending marks the Tx as Pending
	TxStatusPending TxStatus = "pending"
	// TxStatusSent marks the Tx as Sent
	TxStatusSent TxStatus = "sent"
)

// BatchInfo contans the Batch information
type BatchInfo struct {
	BatchNum              common.BatchNum
	ServerProof           prover.Client
	ZKInputs              *common.ZKInputs
	Proof                 *prover.Proof
	PublicInputs          []*big.Int
	L1Batch               bool
	L1UserTxsExtra        []common.L1Tx
	L1CoordTxs            []common.L1Tx
	L1CoordinatorTxsAuths [][]byte
	L2Txs                 []common.L2Tx
	CoordIdxs             []common.Idx
	ForgeBatchArgs        *eth.RollupForgeBatchArgs
	// FeesInfo
	TxStatus TxStatus
	EthTx    *types.Transaction
	Receipt  *types.Receipt
}

// DebugStore is a debug function to store the BatchInfo as a json text file in
// storePath
func (b *BatchInfo) DebugStore(storePath string) error {
	batchJSON, err := json.Marshal(b)
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
