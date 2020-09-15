package common

import ethCommon "github.com/ethereum/go-ethereum/common"

// SyncStatus is returned by the Status method of the Synchronizer
type SyncStatus struct {
	CurrentBlock      int64
	CurrentBatch      BatchNum
	CurrentForgerAddr ethCommon.Address
	NextForgerAddr    ethCommon.Address
	Synchronized      bool
}
