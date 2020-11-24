package coordinator

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
)

// Proof TBD this type will be received from the proof server
type Proof struct {
}

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
	BatchNum       common.BatchNum
	ServerProof    ServerProofInterface
	ZKInputs       *common.ZKInputs
	Proof          *Proof
	L1UserTxsExtra []common.L1Tx
	L1OperatorTxs  []common.L1Tx
	L2Txs          []common.PoolL2Tx
	ForgeBatchArgs *eth.RollupForgeBatchArgs
	// FeesInfo
	TxStatus TxStatus
	EthTx    *types.Transaction
}
