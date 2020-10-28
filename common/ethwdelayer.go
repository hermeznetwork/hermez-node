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
	HermezRollupAddress        ethCommon.Address `json:"hermezRollupAddress" meddler:"rollup_address"`
	HermezGovernanceDAOAddress ethCommon.Address `json:"hermezGovernanceDAOAddress" meddler:"govdao_address"`
	WhiteHackGroupAddress      ethCommon.Address `json:"whiteHackGroupAddress" meddler:"whg_address"`
	HermezKeeperAddress        ethCommon.Address `json:"hermezKeeperAddress" meddler:"keeper_address"`
	WithdrawalDelay            uint64            `json:"withdrawalDelay" meddler:"withdrawal_delay"`
	EmergencyModeStartingTime  uint64            `json:"emergencyModeStartingTime" meddler:"emergency_start_time"`
	EmergencyMode              bool              `json:"emergencyMode" meddler:"emergency_mode"`
}
