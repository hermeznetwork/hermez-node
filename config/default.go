package config

// DefaultValues is the default fonfigurations for the hermez node
const DefaultValues = `
[API]
Address = "0.0.0.0:8086"
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
Symbols = "2=UST,3=UDC,5=WBT,7=XAUT:,9=SUSHI:,10=COMP:,12=AAVE:,14=LINK:,24=GNT,27=ignore,28=ignore,29=ignore,30=ignore,31=ignore,32=ignore"

[[PriceUpdater.Provider]]
Provider = "CoinGeckoV3"
BASEURL = "https://api.coingecko.com/api/v3/"
URL = "simple/token_price/ethereum?contract_addresses="
URLExtraParams = "&vs_currencies=usd"
Addresses="6=0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2,17=0xde30da39c46104798bb5aa3fe8b9e0e1f348163f,18=0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0,21=0xc944e90c64b2c07662a292be6244bdf05cda44a7,26=0xD533a949740bb3306d119CC777fa900bA034cd52,27=ignore,28=ignore,29=ignore,30=ignore,31=ignore,32=ignore"

[Debug]
APIAddress = "0.0.0.0:12345"
MeddlerLogs = true
GinDebugMode = false

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
`
