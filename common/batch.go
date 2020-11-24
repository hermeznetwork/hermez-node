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
	BatchNum           BatchNum             `meddler:"batch_num"`
	EthBlockNum        int64                `meddler:"eth_block_num"` // Ethereum block in which the batch is forged
	ForgerAddr         ethCommon.Address    `meddler:"forger_addr"`
	CollectedFees      map[TokenID]*big.Int `meddler:"fees_collected,json"`
	FeeIdxsCoordinator []Idx                `meddler:"fee_idxs_coordinator,json"`
	StateRoot          *big.Int             `meddler:"state_root,bigint"`
	NumAccounts        int                  `meddler:"num_accounts"`
	LastIdx            int64                `meddler:"last_idx"`
	ExitRoot           *big.Int             `meddler:"exit_root,bigint"`
	ForgeL1TxsNum      *int64               `meddler:"forge_l1_txs_num"` // optional, Only when the batch forges L1 txs. Identifier that corresponds to the group of L1 txs forged in the current batch.
	SlotNum            int64                `meddler:"slot_num"`         // Slot in which the batch is forged
	TotalFeesUSD       *float64             `meddler:"total_fees_usd"`
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

// BatchData contains the information of a Batch
type BatchData struct {
	// L1UserTxs that were forged in the batch
	L1Batch bool // TODO: Remove once Batch.ForgeL1TxsNum is a pointer
	// L1UserTxs        []common.L1Tx
	L1CoordinatorTxs []L1Tx
	L2Txs            []L2Tx
	CreatedAccounts  []Account
	ExitTree         []ExitInfo
	Batch            Batch
}

// NewBatchData creates an empty BatchData with the slices initialized.
func NewBatchData() *BatchData {
	return &BatchData{
		L1Batch: false,
		// L1UserTxs:        make([]common.L1Tx, 0),
		L1CoordinatorTxs: make([]L1Tx, 0),
		L2Txs:            make([]L2Tx, 0),
		CreatedAccounts:  make([]Account, 0),
		ExitTree:         make([]ExitInfo, 0),
		Batch:            Batch{},
	}
}
