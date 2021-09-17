package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/tracerr"
)

const batchNumBytesLen = 8

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum  BatchNum       `meddler:"batch_num"`
	EthTxHash ethCommon.Hash `meddler:"eth_tx_hash"`
	// Ethereum block in which the batch is forged
	EthBlockNum        int64                `meddler:"eth_block_num"`
	ForgerAddr         ethCommon.Address    `meddler:"forger_addr"`
	CollectedFees      map[TokenID]*big.Int `meddler:"fees_collected,json"`
	FeeIdxsCoordinator []Idx                `meddler:"fee_idxs_coordinator,json"`
	StateRoot          *big.Int             `meddler:"state_root,bigint"`
	NumAccounts        int                  `meddler:"num_accounts"`
	LastIdx            int64                `meddler:"last_idx"`
	ExitRoot           *big.Int             `meddler:"exit_root,bigint"`
	GasUsed            uint64               `meddler:"gas_used"`
	GasPrice           *big.Int             `meddler:"gas_price,bigint"`
	EtherPriceUSD      float64              `meddler:"ether_price_usd"`
	// ForgeL1TxsNum is optional, Only when the batch forges L1 txs. Identifier that corresponds
	// to the group of L1 txs forged in the current batch.
	ForgeL1TxsNum *int64   `meddler:"forge_l1_txs_num"`
	SlotNum       int64    `meddler:"slot_num"` // Slot in which the batch is forged
	TotalFeesUSD  *float64 `meddler:"total_fees_usd"`
}

// NewEmptyBatch creates a new empty batch
func NewEmptyBatch() *Batch {
	return &Batch{
		BatchNum:           0,
		EthBlockNum:        0,
		ForgerAddr:         ethCommon.Address{},
		CollectedFees:      make(map[TokenID]*big.Int),
		FeeIdxsCoordinator: make([]Idx, 0),
		StateRoot:          big.NewInt(0),
		NumAccounts:        0,
		LastIdx:            0,
		ExitRoot:           big.NewInt(0),
		ForgeL1TxsNum:      nil,
		SlotNum:            0,
		TotalFeesUSD:       nil,
	}
}

// BatchNum identifies a batch
type BatchNum int64

// Bytes returns a byte array of length 4 representing the BatchNum
func (bn BatchNum) Bytes() []byte {
	var batchNumBytes [batchNumBytesLen]byte
	binary.BigEndian.PutUint64(batchNumBytes[:], uint64(bn))
	return batchNumBytes[:]
}

// BigInt returns a *big.Int representing the BatchNum
func (bn BatchNum) BigInt() *big.Int {
	return big.NewInt(int64(bn))
}

// BatchNumFromBytes returns BatchNum from a []byte
func BatchNumFromBytes(b []byte) (BatchNum, error) {
	if len(b) != batchNumBytesLen {
		return 0,
			tracerr.Wrap(fmt.Errorf("can not parse BatchNumFromBytes, bytes len %d, expected %d",
				len(b), batchNumBytesLen))
	}
	batchNum := binary.BigEndian.Uint64(b[:batchNumBytesLen])
	return BatchNum(batchNum), nil
}

// BatchData contains the information of a Batch
type BatchData struct {
	L1Batch bool
	// L1UserTxs that were forged in the batch
	L1UserTxs        []L1Tx
	L1CoordinatorTxs []L1Tx
	L2Txs            []L2Tx
	CreatedAccounts  []Account
	UpdatedAccounts  []AccountUpdate
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
