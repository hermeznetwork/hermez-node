package common

import (
	"time"
)

// SyncronizerState describes the syncronization progress of the smart contracts
type SyncronizerState struct {
	LastUpdate                time.Time // last time this information was updated
	CurrentBatchNum           BatchNum  // Last batch that was forged on the blockchain
	CurrentBlockNum           uint64    // Last block that was mined on Ethereum
	CurrentToForgeL1TxsNum    uint32
	LastSyncedBatchNum        BatchNum // last batch synchronized by the coordinator
	LastSyncedBlockNum        uint64   // last Ethereum block synchronized by the coordinator
	LastSyncedToForgeL1TxsNum uint32
}
