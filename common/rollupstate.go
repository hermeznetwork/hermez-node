package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

// RollupState give information about the rollup, and the synchronization status between the operator and the smart contract
type RollupState struct {
	IsSynched        bool        // true if the operator is fully synched with the rollup smart contract
	SyncProgress     float32     // percentage of synced progress with the rollup smart contract
	LastBlockSynched uint64      // last Etherum block synchronized by the operator
	LastBatchSynched BatchNum    // last batch synchronized by the operator
	FeeDeposit       *big.Int    // amount of eth (in wei) that has to be payed to do a deposit
	FeeL1Tx          *big.Int    // amount of eth (in wei) that has to be payed to do a L1 tx
	ContractAddr     eth.Address // Etherum address of the rollup smart contract
	MaxTx            uint16      // Max amount of txs that can be added in a batch, either L1 or L2
	MaxL1Tx          uint16      // Max amount of L1 txs that can be added in a batch
	NLevels          uint16      // Heigth of the SMT. This will determine the maximum number of accounts that can coexist in the Hermez network by 2^nLevels
}
