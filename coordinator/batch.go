package coordinator

import (
	"github.com/hermeznetwork/hermez-node/common"
)

// BatchInfo contans the Batch information
type BatchInfo struct {
	batchNum       uint64
	serverProof    *ServerProofInfo
	zkInputs       *common.ZKInputs
	L1UserTxsExtra []common.L1Tx
	L1OperatorTxs  []common.L1Tx
	L2Txs          []common.PoolL2Tx
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

// AddTxsInfo adds the l1UserTxs, l1OperatorTxs and l2Txs to the BatchInfo data
// structure
func (bi *BatchInfo) AddTxsInfo(l1UserTxsExtra, l1OperatorTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// TBD parameter: feesInfo
	bi.L1UserTxsExtra = l1UserTxsExtra
	bi.L1OperatorTxs = l1OperatorTxs
	bi.L2Txs = l2Txs
}

// AddTxsInfo adds the ZKInputs to the BatchInfo data structure
func (bi *BatchInfo) AddZKInputs(zkInputs *common.ZKInputs) {
	bi.zkInputs = zkInputs
}

// AddTxsInfo adds the ServerProofInfo to the BatchInfo data structure
func (bi *BatchInfo) AddServerProof(serverProof *ServerProofInfo) {
	bi.serverProof = serverProof
}

// BatchQueue implements a FIFO queue of BatchInfo
type BatchQueue struct {
	queue []*BatchInfo
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
