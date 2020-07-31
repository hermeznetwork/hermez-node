package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum         BatchNum
	SlotNum          SlotNum // Slot in which the batch is forged
	EthTxHash        eth.Hash
	EthBlockNum      uint64 // Ethereum block in which the batch is forged
	ExitRoot         Hash
	OldStateRoot     Hash
	NewStateRoot     Hash
	OldNumAccounts   int
	NewNumAccounts   int
	ToForgeL1TxsNum  uint32   // optional, Only when the batch forges L1 txs. Identifier that corresponds to the group of L1 txs forged in the current batch.
	ToForgeL1TxsHash eth.Hash // optional, Only when the batch forges L1 txs. Frozen from pendingL1TxsHash (which are the group of L1UserTxs), to be forged in ToForgeL1TxsNum + 1.
	ForgedL1TxsHash  eth.Hash // optional, Only when the batch forges L1 txs. This will be the Hash of the group of L1 txs (L1UserTxs + L1CoordinatorTx) forged in the current batch.
	CollectedFees    map[TokenID]*big.Int
	ForgerAddr       eth.Address // TODO: Should this be retrieved via slot reference?
}

// BatchNum identifies a batch
type BatchNum uint32
