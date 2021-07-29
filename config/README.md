# Configuration

This is an example of the env variable equivalations and the toml file:

```
[API]
Address = HEZNODE_API_ADDRESS
Explorer = HEZNODE_API_EXPLORER
UpdateMetricsInterval = HEZNODE_API_UPDATEMETRICSINTERVAL
UpdateRecommendedFeeInterval = HEZNODE_API_UPDATERECOMMENDEDFEEINTERVAL
MaxSQLConnections = HEZNODE_API_MAXSQLCONNECTIONS
SQLConnectionTimeout = HEZNODE_API_SQLCONNECTIONTIMEOUT

[PriceUpdater]
Interval = HEZNODE_PRICEUPDATER_INTERVAL
Priority = HEZNODE_PRICEUPDATER_PRIORITY
Statictokens = HEZNODE_PRICEUPDATER_STATICTOKENS

[PriceUpdater.Fiat]
APIKey = HEZNODE_PRICEUPDATER_FIAT_APIKEY
URL = HEZNODE_PRICEUPDATER_FIAT_URL
BaseCurrency = HEZNODE_PRICEUPDATER_FIAT_BASECURRENCY
Currencies = HEZNODE_PRICEUPDATER_FIAT_CURRENCIES

[[PriceUpdater.Provider]]
Provider = HEZNODE_PRICEUPDATER_PROVIDER_PROVIDER
BASEURL = HEZNODE_PRICEUPDATER_PROVIDER_BASEURL
URL = HEZNODE_PRICEUPDATER_PROVIDER_URL
URLExtraParams = HEZNODE_PRICEUPDATER_PROVIDER_URLEXTRAPARAMS
Symbols = HEZNODE_PRICEUPDATER_PROVIDER_SYMBOLS
Addresses = HEZNODE_PRICEUPDATER_PROVIDER_ADDRESSES

[Debug]
APIAddress = HEZNODE_DEBUG_APIADDRESS
MeddlerLogs = HEZNODE_DEBUG_MEDDLERLOGS
GinDebugMode = HEZNODE_DEBUG_GINDEBUGMODE

[StateDB]
Path = HEZNODE_STATEDB_PATH
Keep = HEZNODE_STATEDB_KEEP

[PostgreSQL]
PortWrite     = HEZNODE_POSTGRESQL_PORTWRITE
HostWrite     = HEZNODE_POSTGRESQL_HOSTWRITE
UserWrite     = HEZNODE_POSTGRESQL_USERWRITE
PasswordWrite = HEZNODE_POSTGRESQL_PASSWORDWRITE
NameWrite     = HEZNODE_POSTGRESQL_NAMEWRITE
PortRead     = HEZNODE_POSTGRESQL_PORTREAD
HostRead     = HEZNODE_POSTGRESQL_HOSTREAD
UserRead     = HEZNODE_POSTGRESQL_USERREAD
PasswordRead = HEZNODE_POSTGRESQL_PASSWORDREAD
NameRead     = HEZNODE_POSTGRESQL_NAMEREAD

[Web3]
URL = HEZNODE_WEB3_URL

[Synchronizer]
SyncLoopInterval = HEZNODE_SYNCHRONIZER_SYNCLOOPINTERVAL
StatsUpdateBlockNumDiffThreshold = HEZNODE_SYNCHRONIZER_STATSUPDATEBLOCKSNUMDIFFTHRESHOLD
StatsUpdateFrequencyDivider = HEZNODE_SYNCHRONIZER_STATSUPDATEFREQUENCYDIVIDER

[SmartContracts]
Rollup   = HEZNODE_SMARTCONTRACTS_ROLLUP

[Coordinator]
ForgerAddress = HEZNODE_COORDINATOR_FORGERADDRESS
MinimumForgeAddressBalance = HEZNODE_COORDINATOR_MINIMUMFORGEADDRESSBALANCE
ConfirmBlocks = HEZNODE_COORDINATOR_CONFIRMBLOCKS
L1BatchTimeoutPerc = HEZNODE_COORDINATOR_L1BATCHTIMEOUTPERC
StartSlotBlocksDelay = HEZNODE_COORDINATOR_STARTSLOTBLOCKSDELAY
ScheduleBatchBlocksAheadCheck = HEZNODE_COORDINATOR_SCHEDULEBATCHBLOCKSAHEADCHECK
SendBatchBlocksMarginCheck = HEZNODE_COORDINATOR_SENDBATCHBLOCKSMARGINCHECK
ProofServerPollInterval = HEZNODE_COORDINATOR_PROOFSERVERPOLLINTERVAL
ForgeRetryInterval = HEZNODE_COORDINATOR_FORGERETRYINTERVAL
SyncRetryInterval = HEZNODE_COORDINATOR_SYNCRETRYINTERVAL
ForgeDelay = HEZNODE_COORDINATOR_FORGEDELAY
ForgeNoTxsDelay = HEZNODE_COORDINATOR_FORGENOTXSDELAY
PurgeByExtDelInterval = HEZNODE_COORDINATOR_PURGEBYEXTDELINTERVAL
MustForgeAtSlotDeadline = HEZNODE_COORDINATOR_MUSTFORGEATSLOTDEADLINE
IgnoreSlotCommitment = HEZNODE_COORDINATOR_IGNORESLOTCOMMITMENT

[Coordinator.FeeAccount]
Address = HEZNODE_FEEACCOUNT_ADDRESS
BJJ = HEZNODE_FEEACCOUNT_BJJ


[Coordinator.L2DB]
SafetyPeriod = HEZNODE_L2DB_SAFETYPERIOD
MaxTxs       = HEZNODE_L2DB_MAXTXS
MinFeeUSD    = HEZNODE_L2DB_MINFEEUSD
MaxFeeUSD    = HEZNODE_L2DB_MAXFEEUSD
TTL          = HEZNODE_L2DB_TTL
PurgeBatchDelay = HEZNODE_L2DB_PURGEBATCHDELAY
InvalidateBatchDelay = HEZNODE_L2DB_INVALIDATEBATCHDELAY
PurgeBlockDelay = HEZNODE_L2DB_PURGEBLOCKDELAY
InvalidateBlockDelay = HEZNODE_L2DB_INVALIDATEBLOCKDELAY

[Coordinator.TxSelector]
Path = HEZNODE_TXSELECTOR_PATH

[Coordinator.BatchBuilder]
Path = HEZNODE_BATCHBUILDER_PATH

[[Coordinator.ServerProofs]]
URL = HEZNODE_SERVERPROOF_URL

[Coordinator.Circuit]
MaxTx = HEZNODE_CIRCUIT_MAXTX
NLevels = HEZNODE_CIRCUIT_NLEVELS

[Coordinator.EthClient]
CheckLoopInterval = HEZNODE_ETHCLIENT_CHECKLOOPINTERVAL
Attempts = HEZNODE_ETHCLIENT_ATTEMPTS
AttemptsDelay = HEZNODE_ETHCLIENT_ATTEMPTSDELAY
TxResendTimeout = HEZNODE_ETHCLIENT_TXRESENDTIMEOUT
NoReuseNonce = HEZNODE_ETHCLIENT_NOREUSENONCE
MaxGasPrice = HEZNODE_ETHCLIENT_MAXGASPRICE
MinGasPrice = HEZNODE_ETHCLIENT_MINGASPRICE
GasPriceIncPerc = HEZNODE_ETHCLIENT_GASPRICEINCPERC

[Coordinator.EthClient.Keystore]
Path = HEZNODE_KEYSTORE_PATH
Password = HEZNODE_KEYSTORE_PASSWORD

[Coordinator.EthClient.ForgeBatchGasCost]
Fixed = HEZNODE_FORGEBATCHGASCOST_FIXED
L1UserTx = HEZNODE_FORGEBATCHGASCOST_L1USERTX
L1CoordTx = HEZNODE_FORGEBATCHGASCOST_L1COORDTX
L2Tx = HEZNODE_FORGEBATCHGASCOST_L2TX

[Coordinator.API]
Coordinator = HEZNODE_COORDINATORAPI_COORDINATOR

[Coordinator.Debug]
BatchPath = HEZNODE_COORDINATORDEBUG_BATCHPATH
LightScrypt = HEZNODE_COORDINATORDEBUG_LIGHTSCRYPT
RollupVerifierIndex = HEZNODE_COORDINATORDEBUG_ROLLUPVERIFIERINDEX

[Coordinator.Etherscan]
URL = HEZNODE_ETHERSCAN_URL
APIKey = HEZNODE_ETHERSCAN_APIKEY

[RecommendedFeePolicy]
PolicyType = HEZNODE_RECOMMENDEDFEEPOLICY_POLICYTYPE
StaticValue = HEZNODE_RECOMMENDEDFEEPOLICY_STATICVALUE
```

## Table: 

|Section |Parameter Name |Env name  | Required/Optional|Default value |Description |
--- | --- | --- | --- | --- | --- |
|API|Address|HEZNODE_API_ADDRESS|Optional|"localhost:8086"|Url and port where the API will listen
|API|Explorer|HEZNODE_API_EXPLORER|Optional|true|Enables the Explorer API endpoints
|API|UpdateMetricsInterval|HEZNODE_API_UPDATEMETRICSINTERVAL|Optional|"10s"|Interval between updates of the API metrics
|API|UpdateRecommendedFeeInterval|HEZNODE_API_UPDATERECOMMENDEDFEEINTERVAL|Optional|"10s"|Interval between updates of the recommended fees
|API|MaxSQLConnections|HEZNODE_API_MAXSQLCONNECTIONS|Optional|100|Maximum concurrent connections allowed between API and SQL
|API|SQLConnectionTimeout|HEZNODE_API_SQLCONNECTIONTIMEOUT|Optional|"2s"|Maximum amount of time that an API request can wait to establish a SQL connection
|PriceUpdater|Interval|HEZNODE_PRICEUPDATER_INTERVAL|Optional|"60s"|Interval between price updater calls
|PriceUpdater|Priority|HEZNODE_PRICEUPDATER_PRIORITY|Optional|"bitfinexV2,CoinGeckoV3"|Priority order to use the different providers
|PriceUpdater|Statictokens|HEZNODE_PRICEUPDATER_STATICTOKENS|Optional|""|Static price for tokens
|PriceUpdater.Fiat|APIKey|HEZNODE_PRICEUPDATER_FIAT_APIKEY|Optional|""|This parameter allow access to the external service
|PriceUpdater.Fiat|URL|HEZNODE_PRICEUPDATER_FIAT_URL|Optional|"https://api.exchangeratesapi.io/v1/"|Url of the external service
|PriceUpdater.Fiat|BaseCurrency|HEZNODE_PRICEUPDATER_FIAT_BASECURRENCY|Optional|"USD"|Base currency used to get prices
|PriceUpdater.Fiat|Currencies|HEZNODE_PRICEUPDATER_FIAT_CURRENCIES|Optional|"CNY,EUR,JPY,GBP"|Currencies that the hermez node is going to store
|PriceUpdater.Provider|Provider|HEZNODE_PRICEUPDATER_PROVIDER_PROVIDER|Optional|"bitfinexV2"|Provider name used in "Priority" field
|PriceUpdater.Provider|BASEURL|HEZNODE_PRICEUPDATER_PROVIDER_BASEURL|Optional|"https://api-pub.bitfinex.com/v2/"|Provider api base url
|PriceUpdater.Provider|URL|HEZNODE_PRICEUPDATER_PROVIDER_URL|Optional|"ticker/t"| Provider url
|PriceUpdater.Provider|URLExtraParams|HEZNODE_PRICEUPDATER_PROVIDER_URLEXTRAPARAMS|Optional|"USD"|Second part of the url
|PriceUpdater.Provider|Symbols|HEZNODE_PRICEUPDATER_PROVIDER_SYMBOLS|Optional|"0=ETH,2=UST,8=ignore,9=SUSHI:,5=WBT,14=LINK:,12=AAVE:,7=XAUT:,10=COMP:,6=ETH"|Symbols that the price updater should use in order to get the prices successfully
|PriceUpdater.Provider|Addresses|HEZNODE_PRICEUPDATER_PROVIDER_ADDRESSES|Optional|Addresses="0=0x0000000000000000000000000000000000000000,2=0xdac17f958d2ee523a2206206994597c13d831ec7,8=ignore"|Addresses that the price updater should use in order to get the prices successfully
|Debug|APIAddress|HEZNODE_DEBUG_APIADDRESS|Optional|"0.0.0.0:12345"|If it is set, the debug api will listen in this address and port
|Debug|MeddlerLogs|HEZNODE_DEBUG_MEDDLERLOGS|Optional|true|Enables meddler debug mode, where unused columns and struct fields will be logged
|Debug|GinDebugMode|HEZNODE_DEBUG_GINDEBUGMODE|Optional|true|Sets the web framework Gin-Gonic to run in debug mode
|StateDB|Path|HEZNODE_STATEDB_PATH|Optional|"/var/hermez/statedb"|Path where the synchronizer StateDB is stored
|StateDB|Keep|HEZNODE_STATEDB_KEEP|Optional|256|Number of checkpoints to keep
|PostgreSQL|PortWrite|HEZNODE_POSTGRESQL_PORTWRITE|**Required**|5432|Port of the PostgreSQL write server
|PostgreSQL|HostWrite|HEZNODE_POSTGRESQL_HOSTWRITE|**Required**|"localhost"|Host of the PostgreSQL write server
|PostgreSQL|UserWrite|HEZNODE_POSTGRESQL_USERWRITE|**Required**|"hermez"|User of the PostgreSQL write server
|PostgreSQL|PasswordWrite|HEZNODE_POSTGRESQL_PASSWORDWRITE|**Required**|"yourpasswordhere"|Password of the PostgreSQL write server
|PostgreSQL|NameWrite|HEZNODE_POSTGRESQL_NAMEWRITE|**Required**|"hermez"|Name of the PostgreSQL write server database
|PostgreSQL|PortRead|HEZNODE_POSTGRESQL_PORTREAD|Optional|5432|Port of the PostgreSQL read server. If it is not set, the hermez node will use the postgresql write server configuration
|PostgreSQL|HostRead|HEZNODE_POSTGRESQL_HOSTREAD|Optional|"localhost"|Host of the PostgreSQL read server. If it is not set, the hermez node will use the postgresql write server configuration
|PostgreSQL|UserRead|HEZNODE_POSTGRESQL_USERREAD|Optional|"hermez"|User of the PostgreSQL read server. If it is not set, the hermez node will use the postgresql write server configuration
|PostgreSQL|PasswordRead|HEZNODE_POSTGRESQL_PASSWORDREAD|Optional|"yourpasswordhere"|Password of the PostgreSQL read server. If it is not set, the hermez node will use the postgresql write server configuration
|PostgreSQL|NameRead|HEZNODE_POSTGRESQL_NAMEREAD|Optional|"hermez"|Name of the PostgreSQL read server database. If it is not set, the hermez node will use the postgresql write server configuration
|Web3|URL|HEZNODE_WEB3_URL|**Required**|"http://localhost:8545"|Url of the web3 ethereum-node RPC server.  Only geth is officially supported
|Synchronizer|SyncLoopInterval|HEZNODE_SYNCHRONIZER_SYNCLOOPINTERVAL|Optional|"1s"|Interval between attempts to synchronize a new block from an ethereum node
|Synchronizer|StatsUpdateBlockNumDiffThreshold|HEZNODE_SYNCHRONIZER_STATSUPDATEBLOCKSNUMDIFFTHRESHOLD|Optional|100|Threshold of a number of Ethereum blocks left to synchronize, such that if there are more blocks to sync than the defined value synchronizer can aggressively skip calling UpdateEth to save network bandwidth and time. After reaching the threshold UpdateEth is called on each block. This value only affects the reported % of synchronization of blocks and batches, nothing else.
|Synchronizer|StatsUpdateFrequencyDivider|HEZNODE_SYNCHRONIZER_STATSUPDATEFREQUENCYDIVIDER|Optional|100|While having more blocks to sync than updateEthBlockNumThreshold, UpdateEth will be called once in a defined number of blocks. This value only affects the reported % of synchronization of blocks and batches, nothing else
|SmartContracts|Rollup|HEZNODE_SMARTCONTRACTS_ROLLUP|**Required**|"0x8EEaea23686c319133a7cC110b840d1591d9AeE0"|Smart contract address of the rollup Hermez.sol
|Coordinator|ForgerAddress|HEZNODE_COORDINATOR_FORGERADDRESS|**Required**|"0x05c23b938a85ab26A36E6314a0D02080E9ca6BeD"|Ethereum address that the coordinator is using to forge batches
|Coordinator|MinimumForgeAddressBalance|HEZNODE_COORDINATOR_MINIMUMFORGEADDRESSBALANCE|Optional|"0"|Minimum balance the forger address needs to start the coordinator in wei. If It is set to 0, the coordinator will not check the balance
|Coordinator|ConfirmBlocks|HEZNODE_COORDINATOR_CONFIRMBLOCKS|Optional|10|Number of confirmation blocks to be sure that the tx has been mined correctly
|Coordinator|L1BatchTimeoutPerc|HEZNODE_COORDINATOR_L1BATCHTIMEOUTPERC|Optional|0.6|Portion of the range before the L1Batch timeout that will trigger a schedule to forge an L1Batch
|Coordinator|StartSlotBlocksDelay|HEZNODE_COORDINATOR_STARTSLOTBLOCKSDELAY|Optional|2|Number of delay blocks to wait before starting the pipeline when a slot in which the coordinator can forge is reached
|Coordinator|ScheduleBatchBlocksAheadCheck|HEZNODE_COORDINATOR_SCHEDULEBATCHBLOCKSAHEADCHECK|Optional|3|Number of blocks ahead used to decide when to stop scheduling new batches
|Coordinator|SendBatchBlocksMarginCheck|HEZNODE_COORDINATOR_SENDBATCHBLOCKSMARGINCHECK|Optional|1|Number of marging blocks used to decide when to stop sending batches to the smart contract
|Coordinator|ProofServerPollInterval|HEZNODE_COORDINATOR_PROOFSERVERPOLLINTERVAL|Optional|"1s"|Interval between calls to the ProofServer to check the status
|Coordinator|ForgeRetryInterval|HEZNODE_COORDINATOR_FORGERETRYINTERVAL|Optional|"500ms"|Interval between forge retries after an error
|Coordinator|SyncRetryInterval|HEZNODE_COORDINATOR_SYNCRETRYINTERVAL|Optional|"1s"|Interval between calls to the main handler of a synced block after an error
|Coordinator|ForgeDelay|HEZNODE_COORDINATOR_FORGEDELAY|Optional|"10s"|Delay after which a batch is forged if the slot is already committed.  If It is set to "0s", the coordinator will continuously forge at the maximum rate
|Coordinator|ForgeNoTxsDelay|HEZNODE_COORDINATOR_FORGENOTXSDELAY|Optional|"0s"|Delay after a forged batch if there are no txs to forge. If It is set to 0s, the coordinator will continuously forge even if the batches are empty
|Coordinator|PurgeByExtDelInterval|HEZNODE_COORDINATOR_PURGEBYEXTDELINTERVAL|Optional|"1m"|Interval between calls to the PurgeByExternalDelete function of the l2db which deletes pending txs externally marked by the column `external_delete`
|Coordinator|MustForgeAtSlotDeadline|HEZNODE_COORDINATOR_MUSTFORGEATSLOTDEADLINE|Optional|true|Enables the coordinator to forge in slots if the empty slots reach the slot deadline.
|Coordinator|IgnoreSlotCommitment|HEZNODE_COORDINATOR_IGNORESLOTCOMMITMENT|Optional|false| It will make the coordinator forge at most one batch per slot, only if there are included txs in that batch, or pending l1UserTxs in the smart contract.  Setting this parameter overrides `ForgeDelay`, `ForgeNoTxsDelay`, `MustForgeAtSlotDeadline` and `IgnoreSlotCommitment`.
|Coordinator.FeeAccount|Address|HEZNODE_FEEACCOUNT_ADDRESS|**Required**|"0x56232B1c5B10038125Bc7345664B4AFD745bcF8E"|Ethereum address of the account that will receive the fees
|Coordinator.FeeAccount|BJJ|HEZNODE_FEEACCOUNT_BJJ|**Required**|"0x130c5c7f294792559f469220274f3d3b2dca6e89f4c5ec88d3a08bf73262171b"|BJJ is the baby jub jub public key of the account that will receive the fees
|Coordinator.L2DB|SafetyPeriod|HEZNODE_L2DB_SAFETYPERIOD|Optional|10|Number of batches after which non-pending L2Txs are deleted from the pool
|Coordinator.L2DB|MaxTxs|HEZNODE_L2DB_MAXTXS|Optional|512|Maximum number of pending L2Txs that can be stored in the pool
|Coordinator.L2DB|MinFeeUSD|HEZNODE_L2DB_MINFEEUSD|Optional|0.0|Minimum fee in USD that a tx must pay in order to be accepted into the pool
|Coordinator.L2DB|MaxFeeUSD|HEZNODE_L2DB_MAXFEEUSD|Optional|50.0|Maximum fee in USD that a tx must pay in order to be accepted into the pool
|Coordinator.L2DB|TTL|HEZNODE_L2DB_TTL|Optional|"24h"|Time To Live for L2Txs in the pool. L2Txs older than TTL will be deleted.
|Coordinator.L2DB|PurgeBatchDelay|HEZNODE_L2DB_PURGEBATCHDELAY|Optional|10|Delay between batches to purge outdated transactions. Outdated L2Txs are those that have been forged or marked as invalid for longer than the SafetyPeriod and pending L2Txs that have been in the pool for longer than TTL once there are MaxTxs
|Coordinator.L2DB|InvalidateBatchDelay|HEZNODE_L2DB_INVALIDATEBATCHDELAY|Optional|20|Delay between batches to mark invalid transactions due to nonce lower than the account nonce
|Coordinator.L2DB|PurgeBlockDelay|HEZNODE_L2DB_PURGEBLOCKDELAY|Optional|10|Delay between blocks to purge outdated transactions. Outdated L2Txs are those that have been forged or marked as invalid for longer than the SafetyPeriod and pending L2Txs that have been in the pool for longer than TTL once there are MaxTxs.
|Coordinator.L2DB|InvalidateBlockDelay|HEZNODE_L2DB_INVALIDATEBLOCKDELAY|Optional|20|Delay between blocks to mark invalid transactions due to nonce lower than the account nonce
|Coordinator.TxSelector|Path|HEZNODE_TXSELECTOR_PATH|Optional|"/var/hermez/txselector"|Path where the TxSelector StateDB is stored
|Coordinator.BatchBuilder|Path|HEZNODE_BATCHBUILDER_PATH|Optional|"/var/hermez/batchbuilder"|Path where the BatchBuilder StateDB is stored
|Coordinator.ServerProofs|URL|HEZNODE_SERVERPROOF_URL|**Required**|"http://localhost:3000/api"|Server proof API URL
|Coordinator.Circuit|MaxTx|HEZNODE_CIRCUIT_MAXTX|**Required**|512|Maximum number of txs supported by the circuit
|Coordinator.Circuit|NLevels|HEZNODE_CIRCUIT_NLEVELS|**Required**|32|Maximum number of merkle tree levels supported by the circuit
|Coordinator.EthClient|CheckLoopInterval|HEZNODE_ETHCLIENT_CHECKLOOPINTERVAL|Optional|"500ms"|Interval between receipt checks of ethereum transactions in the TxManager
|Coordinator.EthClient|Attempts|HEZNODE_ETHCLIENT_ATTEMPTS|Optional|4|Number of attempts to do an eth client RPC call before giving up
|Coordinator.EthClient|AttemptsDelay|HEZNODE_ETHCLIENT_ATTEMPTSDELAY|Optional|"500ms"|Delay between attempts do do an eth client RPC call
|Coordinator.EthClient|TxResendTimeout|HEZNODE_ETHCLIENT_TXRESENDTIMEOUT|Optional|"2m"|Timeout after which a non-mined ethereum transaction will be resent (reusing the nonce) with a newly calculated gas price
|Coordinator.EthClient|NoReuseNonce|HEZNODE_ETHCLIENT_NOREUSENONCE|Optional|false|Disables reusing nonces of pending transactions for new replacement transactions
|Coordinator.EthClient|MaxGasPrice|HEZNODE_ETHCLIENT_MAXGASPRICE|Optional|2000|Maximum gas price allowed for ethereum transactions in gwei
|Coordinator.EthClient|MinGasPrice|HEZNODE_ETHCLIENT_MINGASPRICE|Optional|5|Minimum gas price allowed for ethereum transactions in gwei
|Coordinator.EthClient|GasPriceIncPerc|HEZNODE_ETHCLIENT_GASPRICEINCPERC|Optional|10|Percentage increased of gas price set in an ethereum transaction from the suggested gas price by the ethereum node
|Coordinator.EthClient.Keystore|Path|HEZNODE_KEYSTORE_PATH|Optional|"/var/hermez/ethkeystore"|Path where the keystore is stored
|Coordinator.EthClient.Keystore|Password|HEZNODE_KEYSTORE_PASSWORD|**Required**|"yourpasswordhere"|Password used to decrypt the keys in the keystore
|Coordinator.EthClient.ForgeBatchGasCost|Fixed|HEZNODE_FORGEBATCHGASCOST_FIXED|Optional|600000|Gas needed to forge an empty batch
|Coordinator.EthClient.ForgeBatchGasCost|L1UserTx|HEZNODE_FORGEBATCHGASCOST_L1USERTX|Optional|15000|Gas needed per L1 tx
|Coordinator.EthClient.ForgeBatchGasCost|L1CoordTx|HEZNODE_FORGEBATCHGASCOST_L1COORDTX|Optional|8000|Gas needed for a coordinator L1 tx
|Coordinator.EthClient.ForgeBatchGasCost|L2Tx|HEZNODE_FORGEBATCHGASCOST_L2TX|Optional|250|Gas needed for an L2 tx
|Coordinator.API|Coordinator|HEZNODE_COORDINATORAPI_COORDINATOR|Optional|true|Enables coordinator API endpoints
|Coordinator.Debug|BatchPath|HEZNODE_COORDINATORDEBUG_BATCHPATH|Optional|"/var/hermez/batchesdebug"|If this parameter is set, specifies the path where batchInfo is stored in JSON in every step/update of the pipeline
|Coordinator.Debug|LightScrypt|HEZNODE_COORDINATORDEBUG_LIGHTSCRYPT|Optional|true|If lightScrypt is set, uses light parameters for the ethereum keystore encryption algorithm
|Coordinator.Debug|RollupVerifierIndex|HEZNODE_COORDINATORDEBUG_ROLLUPVERIFIERINDEX|Optional|0|RollupVerifierIndex is the index of the verifier to use in the Rollup smart contract. The verifier chosen by index must match with the Circuit parameters
|Coordinator.Etherscan|URL|HEZNODE_ETHERSCAN_URL|Optional|"https://api.etherscan.io"|If this parameter is set, specifies the etherscan endpoint to get the gas estimations for that moment
|Coordinator.Etherscan|APIKey|HEZNODE_ETHERSCAN_APIKEY|Optional|"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"|This parameter allow access to etherscan services
|RecommendedFeePolicy|PolicyType|HEZNODE_RECOMMENDEDFEEPOLICY_POLICYTYPE|Optional|"Static"|Selects the mode. "Static" or "AvgLastHour"
|RecommendedFeePolicy|StaticValue|HEZNODE_RECOMMENDEDFEEPOLICY_STATICVALUE|Optional|0.99|If PolicyType is "static" defines the recommended fee value
