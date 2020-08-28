package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// SmartContractParameters describes the constant values of the parameters of the Hermez smart contracts
// WARNING: not stable at all
type SmartContractParameters struct {
	SlotDuration        uint64            // number of ethereum blocks in a slot
	Slot0BlockNum       uint64            // ethereum block number of the first slot (slot 0)
	MaxL1UserTxs        uint64            // maximum number of L1UserTxs that can be queued for a single batch
	FreeCoordinatorWait uint64            // if the winning coordinator doesn't forge during this number of blocks, anyone can forge
	ContractAddr        ethCommon.Address // Ethereum address of the rollup smart contract
	NLevels             uint16            // Heigth of the SMT. This will determine the maximum number of accounts that can coexist in the Hermez network by 2^nLevels
	MaxTxs              uint16            // Max amount of txs that can be added in a batch, either L1 or L2
	FeeL1Tx             *big.Int          // amount of eth (in wei) that has to be paid to do a L1 tx
	FeeDeposit          *big.Int          // amount of eth (in wei) that has to be paid to do a deposit
}
