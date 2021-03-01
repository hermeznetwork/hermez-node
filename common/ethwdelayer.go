package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

// WDelayerConstants are the constants of the Withdrawal Delayer Smart Contract
type WDelayerConstants struct {
	// Max Withdrawal Delay
	MaxWithdrawalDelay uint64 `json:"maxWithdrawalDelay"`
	// Max Emergency mode time
	MaxEmergencyModeTime uint64 `json:"maxEmergencyModeTime"`
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address `json:"hermezRollup"`
}

// WDelayerEscapeHatchWithdrawal is an escape hatch withdrawal of the
// Withdrawal Delayer Smart Contract
type WDelayerEscapeHatchWithdrawal struct {
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"`
	Who         ethCommon.Address `json:"who" meddler:"who_addr"`
	To          ethCommon.Address `json:"to" meddler:"to_addr"`
	TokenAddr   ethCommon.Address `json:"tokenAddr" meddler:"token_addr"`
	Amount      *big.Int          `json:"amount" meddler:"amount,bigint"`
}

// WDelayerVariables are the variables of the Withdrawal Delayer Smart Contract
//nolint:lll
type WDelayerVariables struct {
	EthBlockNum int64 `json:"ethereumBlockNum" meddler:"eth_block_num"`
	// HermezRollupAddress        ethCommon.Address `json:"hermezRollupAddress" meddler:"rollup_address"`
	HermezGovernanceAddress    ethCommon.Address `json:"hermezGovernanceAddress" meddler:"gov_address" validate:"required"`
	EmergencyCouncilAddress    ethCommon.Address `json:"emergencyCouncilAddress" meddler:"emg_address" validate:"required"`
	WithdrawalDelay            uint64            `json:"withdrawalDelay" meddler:"withdrawal_delay" validate:"required"`
	EmergencyModeStartingBlock int64             `json:"emergencyModeStartingBlock" meddler:"emergency_start_block"`
	EmergencyMode              bool              `json:"emergencyMode" meddler:"emergency_mode"`
}

// Copy returns a deep copy of the Variables
func (v *WDelayerVariables) Copy() *WDelayerVariables {
	vCpy := *v
	return &vCpy
}
