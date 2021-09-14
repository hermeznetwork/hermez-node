package config

// DefaultValues is the default fonfigurations for the hermez node
const DefaultValues = `
[Log]
Level = "info"
Out = ["stdout"]

[API]
Address = "0.0.0.0:8086"
Explorer = true
UpdateMetricsInterval = "10s"
UpdateRecommendedFeeInterval = "15s"
MaxSQLConnections = 100
SQLConnectionTimeout = "2s"
ReadTimeout = "30s"
WriteTimeout = "30s"
CoordinatorNetwork = true
FindPeersCoordinatorNetworkInterval = "180s"

[StateDB]
Path = "/var/hermez/statedb"
Keep = 256

[PostgreSQL]
PortWrite     = 5432
HostWrite     = "localhost"
UserWrite     = "hermez"
PasswordWrite = "yourpasswordhere"
NameWrite     = "hermez"
# PortRead     = 5432
# HostRead     = "localhost"
# UserRead     = "hermez"
# PasswordRead = "yourpasswordhere"
# NameRead     = "hermez"

[Web3]
URL = "http://localhost:8545"

[Synchronizer]
SyncLoopInterval = "1s"
StatsUpdateBlockNumDiffThreshold = 100
StatsUpdateFrequencyDivider = 100

[SmartContracts]
Rollup   = "0xA68D85dF56E733A06443306A095646317B5Fa633"

[Coordinator]
ForgerAddress = "0x05c23b938a85ab26A36E6314a0D02080E9ca6BeD" # Non-Boot Coordinator
MinimumForgeAddressBalance = "0"
ConfirmBlocks = 5
L1BatchTimeoutPerc = 0.00001
StartSlotBlocksDelay = 0
ScheduleBatchBlocksAheadCheck = 0
SendBatchBlocksMarginCheck = 0
ProofServerPollInterval = "1s"
ForgeRetryInterval = "10s"
SyncRetryInterval = "1s"
ForgeDelay = "600s"
ForgeNoTxsDelay = "86400s"
PurgeByExtDelInterval = "1m"
MustForgeAtSlotDeadline = true
IgnoreSlotCommitment = true
ForgeOncePerSlotIfTxs = false
ProverWaitReadTimeout = "20s"

[Coordinator.FeeAccount]
Address = "0x56232B1c5B10038125Bc7345664B4AFD745bcF8E"
BJJ = "0x130c5c7f294792559f469220274f3d3b2dca6e89f4c5ec88d3a08bf73262171b"

[Coordinator.L2DB]
SafetyPeriod = 10
MaxTxs       = 1000000
MinFeeUSD    = 0.10
MaxFeeUSD    = 10.00
TTL          = "24h"
PurgeBatchDelay = 10
InvalidateBatchDelay = 20
PurgeBlockDelay = 10
InvalidateBlockDelay = 20

[Coordinator.TxSelector]
Path = "/var/hermez/txselector"

[Coordinator.BatchBuilder]
Path = "/var/hermez/batchbuilder"

[Coordinator.ServerProofs]
URLs = ["http://localhost:3000"]

[Coordinator.Circuit]
MaxTx = 2048
NLevels = 32

[Coordinator.EthClient]
CheckLoopInterval = "500ms"
Attempts = 4
AttemptsDelay = "500ms"
TxResendTimeout = "2m"
NoReuseNonce = false
MaxGasPrice = 500
MinGasPrice = 5
GasPriceIncPerc = 5

[Coordinator.EthClient.Keystore]
Path = "/var/hermez/ethkeystore"
Password = "yourpasswordhere"

[Coordinator.EthClient.ForgeBatchGasCost]
Fixed = 900000
L1UserTx = 15000
L1CoordTx = 7000
L2Tx = 600

[Coordinator.API]
Coordinator = true

[RecommendedFeePolicy]
PolicyType = "Static"
StaticValue = 0.10
BreakThreshold = 50
NumLastBatchAvg = 10
`
