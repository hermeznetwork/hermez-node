package common

import ethCommon "github.com/ethereum/go-ethereum/common"

// WDelayerConstants are the constants of the Withdrawal Delayer Smart Contract
type WDelayerConstants struct {
	// Max Withdrawal Delay
	MaxWithdrawalDelay uint64 `json:"maxWithdrawalDelay"`
	// Max Emergency mode time
	MaxEmergencyModeTime uint64 `json:"maxEmergencyModeTime"`
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address `json:"hermezRollup"`
}

// WDelayerVariables are the variables of the Withdrawal Delayer Smart Contract
type WDelayerVariables struct {
	EthBlockNum int64 `json:"ethereumBlockNum" meddler:"eth_block_num"`
	// HermezRollupAddress        ethCommon.Address `json:"hermezRollupAddress" meddler:"rollup_address"`
	HermezGovernanceAddress   ethCommon.Address `json:"hermezGovernanceAddress" meddler:"gov_address" validate:"required"`
	EmergencyCouncilAddress   ethCommon.Address `json:"emergencyCouncilAddress" meddler:"emg_address" validate:"required"`
	WithdrawalDelay           uint64            `json:"withdrawalDelay" meddler:"withdrawal_delay" validate:"required"`
	EmergencyModeStartingTime uint64            `json:"emergencyModeStartingTime" meddler:"emergency_start_time"`
	EmergencyMode             bool              `json:"emergencyMode" meddler:"emergency_mode"`
}

// Copy returns a deep copy of the Variables
func (v *WDelayerVariables) Copy() *WDelayerVariables {
	vCpy := *v
	return &vCpy
}
