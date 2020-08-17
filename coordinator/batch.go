package coordinator

import (
	"github.com/hermeznetwork/hermez-node/common"
)

type Proof struct {
	// TBD this type will be got from the proof server
}

// BatchInfo contans the Batch information
type BatchInfo struct {
	batchNum       uint64
	serverProof    *ServerProofInfo
	zkInputs       *common.ZKInputs
	proof          *Proof
	L1UserTxsExtra []*common.L1Tx
	L1OperatorTxs  []*common.L1Tx
	L2Txs          []*common.PoolL2Tx
	// FeesInfo
}

// NewBatchInfo creates a new BatchInfo with the given batchNum &
// ServerProofInfo
func NewBatchInfo(batchNum uint64, serverProof *ServerProofInfo) BatchInfo {
	return BatchInfo{
		batchNum:    batchNum,
		serverProof: serverProof,
	}
}

// SetTxsInfo sets the l1UserTxs, l1OperatorTxs and l2Txs to the BatchInfo data
// structure
func (bi *BatchInfo) SetTxsInfo(l1UserTxsExtra, l1OperatorTxs []*common.L1Tx, l2Txs []*common.PoolL2Tx) {
	// TBD parameter: feesInfo
	bi.L1UserTxsExtra = l1UserTxsExtra
	bi.L1OperatorTxs = l1OperatorTxs
	bi.L2Txs = l2Txs
}

// SetZKInputs sets the ZKInputs to the BatchInfo data structure
func (bi *BatchInfo) SetZKInputs(zkInputs *common.ZKInputs) {
	bi.zkInputs = zkInputs
}

// SetServerProof sets the ServerProofInfo to the BatchInfo data structure
func (bi *BatchInfo) SetServerProof(serverProof *ServerProofInfo) {
	bi.serverProof = serverProof
}

// SetProof sets the Proof to the BatchInfo data structure
func (bi *BatchInfo) SetProof(proof *Proof) {
	bi.proof = proof
}

// BatchQueue implements a FIFO queue of BatchInfo
type BatchQueue struct {
	queue []*BatchInfo
}

func NewBatchQueue() *BatchQueue {
	return &BatchQueue{
		queue: []*BatchInfo{},
	}
}

// Push adds the given BatchInfo to the BatchQueue
func (bq *BatchQueue) Push(b *BatchInfo) {
	bq.queue = append(bq.queue, b)
}

// Pop pops the first BatchInfo from the BatchQueue
func (bq *BatchQueue) Pop() *BatchInfo {
	if len(bq.queue) == 0 {
		return nil
	}
	b := bq.queue[0]
	bq.queue = bq.queue[1:]
	return b
}
