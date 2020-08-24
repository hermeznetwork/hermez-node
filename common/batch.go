package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum         BatchNum
	SlotNum          SlotNum // Slot in which the batch is forged
	EthTxHash        ethCommon.Hash
	EthBlockNum      uint64 // Ethereum block in which the batch is forged
	ExitRoot         Hash
	OldStateRoot     Hash
	NewStateRoot     Hash
	OldNumAccounts   int
	NewNumAccounts   int
	ToForgeL1TxsNum  uint32         // optional, Only when the batch forges L1 txs. Identifier that corresponds to the group of L1 txs forged in the current batch.
	ToForgeL1TxsHash ethCommon.Hash // optional, Only when the batch forges L1 txs. Frozen from pendingL1TxsHash (which are the group of L1UserTxs), to be forged in ToForgeL1TxsNum + 1.
	ForgedL1TxsHash  ethCommon.Hash // optional, Only when the batch forges L1 txs. This will be the Hash of the group of L1 txs (L1UserTxs + L1CoordinatorTx) forged in the current batch.
	CollectedFees    map[TokenID]*big.Int
	ForgerAddr       ethCommon.Address // TODO: Should this be retrieved via slot reference?
}

// BatchNum identifies a batch
type BatchNum uint32

// Bytes returns a byte array of length 4 representing the BatchNum
func (bn BatchNum) Bytes() []byte {
	var batchNumBytes [4]byte
	binary.LittleEndian.PutUint32(batchNumBytes[:], uint32(bn))
	return batchNumBytes[:]
}

// BatchNumFromBytes returns BatchNum from a []byte
func BatchNumFromBytes(b []byte) (BatchNum, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("can not parse BatchNumFromBytes, bytes len %d, expected 4", len(b))
	}
	batchNum := binary.LittleEndian.Uint32(b[:4])
	return BatchNum(batchNum), nil
}
