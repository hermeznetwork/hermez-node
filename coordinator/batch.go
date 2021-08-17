package coordinator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	"github.com/hermeznetwork/hermez-node/eth"
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

// Debug information related to the Batch
type Debug struct {
	// StartTimestamp of is the time of batch start
	StartTimestamp time.Time
	// SendTimestamp  the time of batch sent to ethereum
	SendTimestamp time.Time
	// Status of the Batch
	Status Status
	// StartBlockNum is the blockNum when the Batch was started
	StartBlockNum int64
	// MineBlockNum is the blockNum in which the batch was mined
	MineBlockNum int64
	// SendBlockNum is the blockNum when the batch was sent to ethereum
	SendBlockNum int64
	// ResendNum is the number of times the tx has been resent
	ResendNum int
	// LastScheduledL1BatchBlockNum is the blockNum when the last L1Batch
	// was scheduled
	LastScheduledL1BatchBlockNum int64
	// LastL1BatchBlock is the blockNum in which the last L1Batch was
	// synced
	LastL1BatchBlock int64
	// LastL1BatchBlockDelta is the number of blocks after the last L1Batch
	LastL1BatchBlockDelta int64
	// L1BatchBlockScheduleDeadline is the number of blocks after the last
	// L1Batch after which an L1Batch will be scheduled
	L1BatchBlockScheduleDeadline int64
	// StartToMineBlocksDelay is the number of blocks that happen between
	// scheduling a batch and having it mined
	StartToMineBlocksDelay int64
	// StartToSendDelay is the delay between starting a batch and sending
	// it to ethereum, in seconds
	StartToSendDelay float64
	// StartToMineDelay is the delay between starting a batch and having
	// it mined in seconds
	StartToMineDelay float64
	// SendToMineDelay is the delay between sending a batch tx and having
	// it mined in seconds
	SendToMineDelay float64
}

// BatchInfo contans the Batch information
type BatchInfo struct {
	PipelineNum           int
	BatchNum              common.BatchNum
	ServerProof           prover.Client
	ProofStart            time.Time
	ZKInputs              *common.ZKInputs
	Proof                 *prover.Proof
	PublicInputs          []*big.Int
	L1Batch               bool
	VerifierIdx           uint8
	L1UserTxs             []common.L1Tx
	L1CoordTxs            []common.L1Tx
	L1CoordinatorTxsAuths [][]byte
	L2Txs                 []common.L2Tx
	CoordIdxs             []common.Idx
	ForgeBatchArgs        *eth.RollupForgeBatchArgs
	Auth                  *bind.TransactOpts `json:"-"`
	EthTxs                []*types.Transaction
	EthTxsErrs            []error
	// SendTimestamp  the time of batch sent to ethereum
	SendTimestamp time.Time
	Receipt       *types.Receipt
	// Fail is true if:
	// - The receipt status is failed
	// - A previous parent batch is failed
	Fail  bool
	Debug Debug
}

// DebugStore is a debug function to store the BatchInfo as a json text file in
// storePath.  The filename contains the batchNumber followed by a timestamp of
// batch start.
func (b *BatchInfo) DebugStore(storePath string) error {
	batchJSON, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return tracerr.Wrap(err)
	}
	// nolint reason: hardcoded 1_000_000 is the number of nanoseconds in a
	// millisecond
	//nolint:gomnd
	filename := fmt.Sprintf("%08d-%v.%03d.json", b.BatchNum,
		b.Debug.StartTimestamp.Unix(), b.Debug.StartTimestamp.Nanosecond()/1_000_000)
	// nolint reason: 0640 allows rw to owner and r to group
	//nolint:gosec
	return ioutil.WriteFile(path.Join(storePath, filename), batchJSON, 0640)
}
