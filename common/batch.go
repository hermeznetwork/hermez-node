package common

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
)

// Batch is a struct that represents Hermez network batch
type Batch struct {
	BatchNum      BatchNum
	SlotNum       SlotNum // Slot in which the batch is forged
	EthTxHash     eth.Hash
	BlockNum      uint64 // Etherum block in which the batch is forged
	Timestamp     time.Time
	Forger        eth.Address
	ExitRoot      Hash
	OldRoot       Hash
	NewRoot       Hash
	TotalAccounts uint64
}

// BatchNum identifies a batch
type BatchNum uint32
