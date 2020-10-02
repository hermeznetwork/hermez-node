package coordinator

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
)

// Proof TBD this type will be received from the proof server
type Proof struct {
}

// BatchInfo contans the Batch information
type BatchInfo struct {
	batchNum       common.BatchNum
	serverProof    ServerProofInterface
	zkInputs       *common.ZKInputs
	proof          *Proof
	L1UserTxsExtra []common.L1Tx
	L1OperatorTxs  []common.L1Tx
	L2Txs          []common.PoolL2Tx
	// FeesInfo
	ethTx *types.Transaction
}

// NewBatchInfo creates a new BatchInfo with the given batchNum &
// ServerProof
func NewBatchInfo(batchNum common.BatchNum, serverProof ServerProofInterface) BatchInfo {
	return BatchInfo{
		batchNum:    batchNum,
		serverProof: serverProof,
	}
}

// SetTxsInfo sets the l1UserTxs, l1OperatorTxs and l2Txs to the BatchInfo data
// structure
func (bi *BatchInfo) SetTxsInfo(l1UserTxsExtra, l1OperatorTxs []common.L1Tx, l2Txs []common.PoolL2Tx) {
	// TBD parameter: feesInfo
	bi.L1UserTxsExtra = l1UserTxsExtra
	bi.L1OperatorTxs = l1OperatorTxs
	bi.L2Txs = l2Txs
}

// SetZKInputs sets the ZKInputs to the BatchInfo data structure
func (bi *BatchInfo) SetZKInputs(zkInputs *common.ZKInputs) {
	bi.zkInputs = zkInputs
}

// SetServerProof sets the ServerProof to the BatchInfo data structure
func (bi *BatchInfo) SetServerProof(serverProof ServerProofInterface) {
	bi.serverProof = serverProof
}

// SetProof sets the Proof to the BatchInfo data structure
func (bi *BatchInfo) SetProof(proof *Proof) {
	bi.proof = proof
}

// SetEthTx sets the ethTx to the BatchInfo data structure
func (bi *BatchInfo) SetEthTx(ethTx *types.Transaction) {
	bi.ethTx = ethTx
}
