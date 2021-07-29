package config

// DefaultValues is the default fonfigurations for the hermez node
const DefaultValues = `
[API]
Address = "localhost:8086"
Explorer = true
UpdateMetricsInterval = "10s"
UpdateRecommendedFeeInterval = "10s"
MaxSQLConnections = 100
SQLConnectionTimeout = "2s"

[PriceUpdater]
Interval = "60s"
Priority = "bitfinexV2,CoinGeckoV3"
Statictokens=""  # <tokenId>=<forced_price>,<tokenId>=<forced_price>

[PriceUpdater.Fiat]
APIKey=""
URL="https://api.exchangeratesapi.io/v1/"
BaseCurrency="USD"
Currencies="CNY,EUR,JPY,GBP"

[[PriceUpdater.Provider]]
Provider = "bitfinexV2"
BASEURL = "https://api-pub.bitfinex.com/v2/"
URL = "ticker/t"
URLExtraParams = "USD"
Symbols = "0=ETH,2=UST,8=ignore,9=SUSHI:,5=WBT,14=LINK:,12=AAVE:,7=XAUT:,10=COMP:,6=ETH"

[[PriceUpdater.Provider]]
Provider = "CoinGeckoV3"
BASEURL = "https://api.coingecko.com/api/v3/"
URL = "simple/token_price/ethereum?contract_addresses="
URLExtraParams = "&vs_currencies=usd"
Addresses="0=0x0000000000000000000000000000000000000000,2=0xdac17f958d2ee523a2206206994597c13d831ec7,8=ignore"

[Debug]
APIAddress = "0.0.0.0:12345"
MeddlerLogs = true
GinDebugMode = true

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
Rollup   = "0x8EEaea23686c319133a7cC110b840d1591d9AeE0"

[Coordinator]
ForgerAddress = "0x05c23b938a85ab26A36E6314a0D02080E9ca6BeD" # Non-Boot Coordinator
MinimumForgeAddressBalance = "0"
ConfirmBlocks = 10
L1BatchTimeoutPerc = 0.6
StartSlotBlocksDelay = 2
ScheduleBatchBlocksAheadCheck = 3
SendBatchBlocksMarginCheck = 1
ProofServerPollInterval = "1s"
ForgeRetryInterval = "500ms"
SyncRetryInterval = "1s"
ForgeDelay = "10s"
ForgeNoTxsDelay = "0s"
PurgeByExtDelInterval = "1m"
MustForgeAtSlotDeadline = true
IgnoreSlotCommitment = false

[Coordinator.FeeAccount]
Address = "0x56232B1c5B10038125Bc7345664B4AFD745bcF8E"
BJJ = "0x130c5c7f294792559f469220274f3d3b2dca6e89f4c5ec88d3a08bf73262171b"

[Coordinator.L2DB]
SafetyPeriod = 10
MaxTxs       = 512
MinFeeUSD    = 0.0
MaxFeeUSD    = 50.0
TTL          = "24h"
PurgeBatchDelay = 10
InvalidateBatchDelay = 20
PurgeBlockDelay = 10
InvalidateBlockDelay = 20

[Coordinator.TxSelector]
Path = "/var/hermez/txselector"

[Coordinator.BatchBuilder]
Path = "/var/hermez/batchbuilder"

[[Coordinator.ServerProofs]]
URL = "http://localhost:3000/api"

[Coordinator.Circuit]
MaxTx = 512
NLevels = 32

[Coordinator.EthClient]
CheckLoopInterval = "500ms"
Attempts = 4
AttemptsDelay = "500ms"
TxResendTimeout = "2m"
NoReuseNonce = false
MaxGasPrice = 2000
MinGasPrice = 5
GasPriceIncPerc = 10

[Coordinator.EthClient.Keystore]
Path = "/var/hermez/ethkeystore"
Password = "yourpasswordhere"

[Coordinator.EthClient.ForgeBatchGasCost]
Fixed = 600000
L1UserTx = 15000
L1CoordTx = 8000
L2Tx = 250

[Coordinator.API]
Coordinator = true

[Coordinator.Debug]
BatchPath = "/var/hermez/batchesdebug"
LightScrypt = true
# RollupVerifierIndex = 0

[Coordinator.Etherscan]
URL = "https://api.etherscan.io"
APIKey = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"

[RecommendedFeePolicy]
PolicyType = "Static"
StaticValue = 0.99
`
