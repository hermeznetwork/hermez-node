package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

const batchNumBytesLen = 8

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum      BatchNum             `meddler:"batch_num"`
	EthBlockNum   int64                `meddler:"eth_block_num"` // Ethereum block in which the batch is forged
	ForgerAddr    ethCommon.Address    `meddler:"forger_addr"`
	CollectedFees map[TokenID]*big.Int `meddler:"fees_collected,json"`
	StateRoot     *big.Int             `meddler:"state_root,bigint"`
	NumAccounts   int                  `meddler:"num_accounts"`
	ExitRoot      *big.Int             `meddler:"exit_root,bigint"`
	ForgeL1TxsNum *int64               `meddler:"forge_l1_txs_num"` // optional, Only when the batch forges L1 txs. Identifier that corresponds to the group of L1 txs forged in the current batch.
	SlotNum       int64                `meddler:"slot_num"`         // Slot in which the batch is forged
	TotalFeesUSD  *float64             `meddler:"total_fees_usd"`
}

// BatchNum identifies a batch
type BatchNum int64

// Bytes returns a byte array of length 4 representing the BatchNum
func (bn BatchNum) Bytes() []byte {
	var batchNumBytes [batchNumBytesLen]byte
	binary.BigEndian.PutUint64(batchNumBytes[:], uint64(bn))
	return batchNumBytes[:]
}

// BatchNumFromBytes returns BatchNum from a []byte
func BatchNumFromBytes(b []byte) (BatchNum, error) {
	if len(b) != batchNumBytesLen {
		return 0, fmt.Errorf("can not parse BatchNumFromBytes, bytes len %d, expected %d", len(b), batchNumBytesLen)
	}
	batchNum := binary.BigEndian.Uint64(b[:batchNumBytesLen])
	return BatchNum(batchNum), nil
}
