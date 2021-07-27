package config

import (
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/stateapiupdater"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/priceupdater"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"gopkg.in/go-playground/validator.v9"
	configLibrary "github.com/hermeznetwork/config-library"
)

// Duration is a wrapper type that parses time duration from text.
type Duration struct {
	time.Duration `validate:"required"`
}

// UnmarshalText unmarshalls time duration from text.
func (d *Duration) UnmarshalText(data []byte) error {
	duration, err := time.ParseDuration(string(data))
	if err != nil {
		return tracerr.Wrap(err)
	}
	d.Duration = duration
	return nil
}

// ServerProof is the server proof configuration data.
type ServerProof struct {
	// URL is the server proof API URL
	URL string `validate:"required,url" env:"SERVERPROOF_URL"`
}

// ForgeBatchGasCost is the costs associated to a ForgeBatch transaction, split
// into different parts to be used in a formula.
type ForgeBatchGasCost struct {
	Fixed     uint64 `validate:"required" env:"FORGEBATCHGASCOST_FIXED"`
	L1UserTx  uint64 `validate:"required" env:"FORGEBATCHGASCOST_L1USERTX"`
	L1CoordTx uint64 `validate:"required" env:"FORGEBATCHGASCOST_L1COORDTX"`
	L2Tx      uint64 `validate:"required" env:"FORGEBATCHGASCOST_L2TX"`
}

// CoordinatorAPI specifies the configuration parameters of the API in mode
// coordinator
type CoordinatorAPI struct {
	// Coordinator enables the coordinator API endpoints
	Coordinator bool `env:"COORDINATORAPI_COORDINATOR"`
}

// Coordinator is the coordinator specific configuration.
type Coordinator struct {
	// ForgerAddress is the address under which this coordinator is forging
	ForgerAddress ethCommon.Address `validate:"required" env:"COORDINATOR_FORGERADDRESS"`
	// MinimumForgeAddressBalance is the minimum balance the forger address
	// needs to start the coordinator in wei. Of set to 0, the coordinator
	// will not check the balance before starting.
	MinimumForgeAddressBalance *big.Int `env:"COORDINATOR_MINIMUMFORGEADDRESSBALANCE"`
	// FeeAccount is the Hermez account that the coordinator uses to receive fees
	FeeAccount struct {
		// Address is the ethereum address of the account to receive fees
		Address ethCommon.Address `validate:"required" env:"FEEACCOUNT_ADDRESS"`
		// BJJ is the baby jub jub public key of the account to receive fees
		BJJ babyjub.PublicKeyComp `validate:"required" env:"FEEACCOUNT_BJJ"`
	} `validate:"required"`
	// ConfirmBlocks is the number of confirmation blocks to wait for sent
	// ethereum transactions before forgetting about them
	ConfirmBlocks int64 `validate:"required,gte=0" env:"COORDINATOR_CONFIRMBLOCKS"`
	// L1BatchTimeoutPerc is the portion of the range before the L1Batch
	// timeout that will trigger a schedule to forge an L1Batch
	L1BatchTimeoutPerc float64 `validate:"required,lte=1.0,gte=0.0" env:"COORDINATOR_L1BATCHTIMEOUTPERC"`
	// StartSlotBlocksDelay is the number of blocks of delay to wait before
	// starting the pipeline when we reach a slot in which we can forge.
	StartSlotBlocksDelay int64 `validate:"gte=0" env:"COORDINATOR_STARTSLOTBLOCKSDELAY"`
	// ScheduleBatchBlocksAheadCheck is the number of blocks ahead in which
	// the forger address is checked to be allowed to forge (apart from
	// checking the next block), used to decide when to stop scheduling new
	// batches (by stopping the pipeline).
	// For example, if we are at block 10 and ScheduleBatchBlocksAheadCheck
	// is 5, even though at block 11 we canForge, the pipeline will be
	// stopped if we can't forge at block 15.
	// This value should be the expected number of blocks it takes between
	// scheduling a batch and having it mined.
	ScheduleBatchBlocksAheadCheck int64 `validate:"gte=0" env:"COORDINATOR_SCHEDULEBATCHBLOCKSAHEADCHECK"`
	// SendBatchBlocksMarginCheck is the number of margin blocks ahead in
	// which the coordinator is also checked to be allowed to forge, apart
	// from the next block; used to decide when to stop sending batches to
	// the smart contract.
	// For example, if we are at block 10 and SendBatchBlocksMarginCheck is
	// 5, even though at block 11 we canForge, the batch will be discarded
	// if we can't forge at block 15.
	SendBatchBlocksMarginCheck int64 `validate:"gte=0" env:"COORDINATOR_SENDBATCHBLOCKSMARGINCHECK"`
	// ProofServerPollInterval is the waiting interval between polling the
	// ProofServer while waiting for a particular status
	ProofServerPollInterval Duration `validate:"required" env:"COORDINATOR_PROOFSERVERPOLLINTERVAL"`
	// ForgeRetryInterval is the waiting interval between calls forge a
	// batch after an error
	ForgeRetryInterval Duration `validate:"required" env:"COORDINATOR_FORGERETRYINTERVAL"`
	// ForgeDelay is the delay after which a batch is forged if the slot is
	// already committed.  If set to 0s, the coordinator will continuously
	// forge at the maximum rate.
	ForgeDelay Duration `validate:"-" env:"COORDINATOR_FORGEDELAY"`
	// ForgeNoTxsDelay is the delay after which a batch is forged even if
	// there are no txs to forge if the slot is already committed.  If set
	// to 0s, the coordinator will continuously forge even if the batches
	// are empty.
	ForgeNoTxsDelay Duration `validate:"-" env:"COORDINATOR_FORGENOTXSDELAY"`
	// MustForgeAtSlotDeadline enables the coordinator to forge slots if
	// the empty slots reach the slot deadline.
	MustForgeAtSlotDeadline bool `env:"COORDINATOR_MUSTFORGEATSLOTDEADLINE"`
	// IgnoreSlotCommitment disables forcing the coordinator to forge a
	// slot immediately when the slot is not committed. If set to false,
	// the coordinator will immediately forge a batch at the beginning of a
	// slot if it's the slot winner.
	IgnoreSlotCommitment bool `env:"COORDINATOR_IGNORESLOTCOMMITMENT"`
	// ForgeOncePerSlotIfTxs will make the coordinator forge at most one
	// batch per slot, only if there are included txs in that batch, or
	// pending l1UserTxs in the smart contract.  Setting this parameter
	// overrides `ForgeDelay`, `ForgeNoTxsDelay`, `MustForgeAtSlotDeadline`
	// and `IgnoreSlotCommitment`.
	ForgeOncePerSlotIfTxs bool `env:"COORDINATOR_FORGEONCEPERSLOTIFTXS"`
	// SyncRetryInterval is the waiting interval between calls to the main
	// handler of a synced block after an error
	SyncRetryInterval Duration `validate:"required" env:"COORDINATOR_SYNCRETRYINTERVAL"`
	// PurgeByExtDelInterval is the waiting interval between calls
	// to the PurgeByExternalDelete function of the l2db which deletes
	// pending txs externally marked by the column `external_delete`
	PurgeByExtDelInterval Duration `validate:"required" env:"COORDINATOR_PURGEBYEXTDELINTERVAL"`
	// L2DB is the DB that holds the pool of L2Txs
	L2DB struct {
		// SafetyPeriod is the number of batches after which
		// non-pending L2Txs are deleted from the pool
		SafetyPeriod common.BatchNum `validate:"required" env:"L2DB_SAFETYPERIOD"`
		// MaxTxs is the maximum number of pending L2Txs that can be
		// stored in the pool.  Once this number of pending L2Txs is
		// reached, inserts to the pool will be denied until some of
		// the pending txs are forged.
		MaxTxs uint32 `validate:"required" env:"L2DB_MAXTXS"`
		// MinFeeUSD is the minimum fee in USD that a tx must pay in
		// order to be accepted into the pool.  Txs with lower than
		// minimum fee will be rejected at the API level.
		MinFeeUSD float64 `validate:"gte=0" env:"L2DB_MINFEEUSD"`
		// MaxFeeUSD is the maximum fee in USD that a tx must pay in
		// order to be accepted into the pool.  Txs with greater than
		// maximum fee will be rejected at the API level.
		MaxFeeUSD float64 `validate:"required,gte=0" env:"L2DB_MAXFEEUSD"`
		// TTL is the Time To Live for L2Txs in the pool. L2Txs older
		// than TTL will be deleted.
		TTL Duration `validate:"required" env:"L2DB_TTL"`
		// PurgeBatchDelay is the delay between batches to purge
		// outdated transactions. Outdated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBatchDelay int64 `validate:"required,gte=0" env:"L2DB_PURGEBATCHDELAY"`
		// InvalidateBatchDelay is the delay between batches to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBatchDelay int64 `validate:"required" env:"L2DB_INVALIDATEBATCHDELAY"`
		// PurgeBlockDelay is the delay between blocks to purge
		// outdated transactions. Outdated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBlockDelay int64 `validate:"required,gte=0" env:"L2DB_PURGEBLOCKDELAY"`
		// InvalidateBlockDelay is the delay between blocks to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBlockDelay int64 `validate:"required,gte=0" env:"L2DB_INVALIDATEBLOCKDELAY"`
	} `validate:"required"`
	TxSelector struct {
		// Path where the TxSelector StateDB is stored
		Path string `validate:"required" env:"TXSELECTOR_PATH"`
	} `validate:"required"`
	BatchBuilder struct {
		// Path where the BatchBuilder StateDB is stored
		Path string `validate:"required" env:"BATCHBUILDER_PATH"`
	} `validate:"required"`
	ServerProofs []ServerProof `validate:"required"`
	Circuit      struct {
		// MaxTx is the maximum number of txs supported by the circuit
		MaxTx int64 `validate:"required,gte=0" env:"CIRCUIT_MAXTX"`
		// NLevels is the maximum number of merkle tree levels
		// supported by the circuit
		NLevels int64 `validate:"required,gte=0" env:"CIRCUIT_NLEVELS"`
	} `validate:"required"`
	EthClient struct {
		// MaxGasPrice is the maximum gas price allowed for ethereum
		// transactions
		MaxGasPrice int64 `validate:"required" env:"ETHCLIENT_MAXGASPRICE"`
		// MinGasPrice is the minimum gas price in gwei allowed for ethereum
		MinGasPrice int64 `validate:"required" env:"ETHCLIENT_MINGASPRICE"`
		// GasPriceIncPerc is the percentage increase of gas price set
		// in an ethereum transaction from the suggested gas price by
		// the ethereum node
		GasPriceIncPerc int64 `validate:"gte=0" env:"ETHCLIENT_GASPRICEINCPERC"`
		// CheckLoopInterval is the waiting interval between receipt
		// checks of ethereum transactions in the TxManager
		CheckLoopInterval Duration `validate:"required" env:"ETHCLIENT_CHECKLOOPINTERVAL"`
		// Attempts is the number of attempts to do an eth client RPC
		// call before giving up
		Attempts int `validate:"required,gte=1" env:"ETHCLIENT_ATTEMPTS"`
		// AttemptsDelay is delay between attempts do do an eth client
		// RPC call
		AttemptsDelay Duration `validate:"required" env:"ETHCLIENT_ATTEMPTSDELAY"`
		// TxResendTimeout is the timeout after which a non-mined
		// ethereum transaction will be resent (reusing the nonce) with
		// a newly calculated gas price
		TxResendTimeout Duration `validate:"required" env:"ETHCLIENT_TXRESENDTIMEOUT"`
		// NoReuseNonce disables reusing nonces of pending transactions for
		// new replacement transactions
		NoReuseNonce bool `env:"ETHCLIENT_NOREUSENONCE"`
		// Keystore is the ethereum keystore where private keys are kept
		Keystore struct {
			// Path to the keystore
			Path string `validate:"required" env:"KEYSTORE_PATH"`
			// Password used to decrypt the keys in the keystore
			Password string `validate:"required" env:"KEYSTORE_PASSWORD"`
		} `validate:"required"`
		// ForgeBatchGasCost contains the cost of each action in the
		// ForgeBatch transaction.
		ForgeBatchGasCost ForgeBatchGasCost `validate:"required"`
	} `validate:"required"`
	API   CoordinatorAPI `validate:"required"`
	Debug struct {
		// BatchPath if set, specifies the path where batchInfo is stored
		// in JSON in every step/update of the pipeline
		BatchPath string `env:"COORDINATORDEBUG_BATCHPATH"`
		// LightScrypt if set, uses light parameters for the ethereum
		// keystore encryption algorithm.
		LightScrypt bool `env:"COORDINATORDEBUG_LIGHTSCRYPT"`
		// RollupVerifierIndex is the index of the verifier to use in
		// the Rollup smart contract.  The verifier chosen by index
		// must match with the Circuit parameters.
		RollupVerifierIndex *int `env:"COORDINATORDEBUG_ROLLUPVERIFIERINDEX"`
	}
	Etherscan struct {
		// URL if set, specifies the etherscan endpoint to get
		// the gas estimations for that moment.
		URL string `env:"ETHERSCAN_URL"`
		// APIKey allow access to etherscan services
		APIKey string `env:"ETHERSCAN_APIKEY"`
	}
}

// PostgreSQL is the postgreSQL configuration parameters.  It's possible to use
// differentiated SQL connections for read/write.  If the read configuration is
// not provided, the write one it's going to be used for both reads and writes
type PostgreSQL struct {
	// Port of the PostgreSQL write server
	PortWrite int `validate:"required" env:"POSTGRESQL_PORTWRITE"`
	// Host of the PostgreSQL write server
	HostWrite string `validate:"required" env:"POSTGRESQL_HOSTWRITE"`
	// User of the PostgreSQL write server
	UserWrite string `validate:"required" env:"POSTGRESQL_USERWRITE"`
	// Password of the PostgreSQL write server
	PasswordWrite string `validate:"required" env:"POSTGRESQL_PASSWORDWRITE"`
	// Name of the PostgreSQL write server database
	NameWrite string `validate:"required" env:"POSTGRESQL_NAMEWRITE"`
	// Port of the PostgreSQL read server
	PortRead int `env:"POSTGRESQL_PORTREAD"`
	// Host of the PostgreSQL read server
	HostRead string `validate:"nefield=HostWrite" env:"POSTGRESQL_HOSTREAD"`
	// User of the PostgreSQL read server
	UserRead string `env:"POSTGRESQL_USERREAD"`
	// Password of the PostgreSQL read server
	PasswordRead string `env:"POSTGRESQL_PASSWORDREAD"`
	// Name of the PostgreSQL read server database
	NameRead string `env:"POSTGRESQL_NAMEREAD"`
}

// NodeDebug specifies debug configuration parameters
type NodeDebug struct {
	// APIAddress is the address where the debugAPI will listen if
	// set
	APIAddress string `env:"DEBUG_APIADDRESS"`
	// MeddlerLogs enables meddler debug mode, where unused columns and struct
	// fields will be logged
	MeddlerLogs bool `env:"DEBUG_MEDDLERLOGS"`
	// GinDebugMode sets Gin-Gonic (the web framework) to run in
	// debug mode
	GinDebugMode bool `env:"DEBUG_GINDEBUGMODE"`
}

// Node is the hermez node configuration.
type Node struct {
	PriceUpdater struct {
		// Interval between price updater calls
		Interval Duration `validate:"required" env:"PRICEUPDATER_INTERVAL"`
		// Priority option defines the priority provider
		Priority string `validate:"required" env:"PRICEUPDATER_PRIORITY"`
		// TokensConfig to specify how each token get it's price updated
		Provider []priceupdater.Provider
		// Statictokens defines the static prices for tokens
		Statictokens string `env:"PRICEUPDATER_STATICTOKENS"`
		// Fiat defines the prices for fiat currencies
		Fiat priceupdater.Fiat
	} `validate:"required"`
	StateDB struct {
		// Path where the synchronizer StateDB is stored
		Path string `validate:"required" env:"STATEDB_PATH"`
		// Keep is the number of checkpoints to keep
		Keep int `validate:"required,gte=128" env:"STATEDB_KEEP"`
	} `validate:"required"`
	PostgreSQL PostgreSQL `validate:"required"`
	Web3       struct {
		// URL is the URL of the web3 ethereum-node RPC server.  Only
		// geth is officially supported.
		URL string `validate:"required,url" env:"WEB3_URL"`
	} `validate:"required"`
	Synchronizer struct {
		// SyncLoopInterval is the interval between attempts to
		// synchronize a new block from an ethereum node
		SyncLoopInterval Duration `validate:"required" env:"SYNCHRONIZER_SYNCLOOPINTERVAL"`
		// StatsUpdateBlockNumDiffThreshold is a threshold of a number of
		// Ethereum blocks left to synchronize, such that if there are more
		// blocks to sync than the defined value synchronizer can aggressively
		// skip calling UpdateEth to save network bandwidth and time.
		// After reaching the threshold UpdateEth is called on each block.
		// This value only affects the reported % of synchronization of
		// blocks and batches, nothing else.
		StatsUpdateBlockNumDiffThreshold uint16 `validate:"required,gt=32" env:"SYNCHRONIZER_STATSUPDATEBLOCKSNUMDIFFTHRESHOLD"`
		// StatsUpdateFrequencyDivider - While having more blocks to sync than
		// updateEthBlockNumThreshold, UpdateEth will be called once in a
		// defined number of blocks. This value only affects the reported % of
		// synchronization of blocks and batches, nothing else.
		StatsUpdateFrequencyDivider uint16 `validate:"required,gt=1" env:"SYNCHRONIZER_STATSUPDATEFREQUENCYDIVIDER"`
	} `validate:"required"`
	SmartContracts struct {
		// Rollup is the address of the Hermez.sol smart contract
		Rollup ethCommon.Address `validate:"required" env:"SMARTCONTRACTS_ROLLUP"`
	} `validate:"required"`
	// API specifies the configuration parameters of the API
	API struct {
		// Address where the API will listen if set
		Address string `env:"API_ADDRESS"`
		// Explorer enables the Explorer API endpoints
		Explorer bool `env:"API_EXPLORER"`
		// UpdateMetricsInterval is the interval between updates of the
		// API metrics
		UpdateMetricsInterval Duration `validate:"required" env:"API_UPDATEMETRICSINTERVAL"`
		// UpdateRecommendedFeeInterval is the interval between updates of the
		// recommended fees
		UpdateRecommendedFeeInterval Duration `validate:"required" env:"API_UPDATERECOMMENDEDFEEINTERVAL"`
		// Maximum concurrent connections allowed between API and SQL
		MaxSQLConnections int `validate:"required,gte=1" env:"API_MAXSQLCONNECTIONS"`
		// SQLConnectionTimeout is the maximum amount of time that an API request
		// can wait to establish a SQL connection
		SQLConnectionTimeout Duration `env:"API_SQLCONNECTIONTIMEOUT"`
	} `validate:"required"`
	RecommendedFeePolicy stateapiupdater.RecommendedFeePolicy `validate:"required"`
	Debug                NodeDebug                            `validate:"required"`
	Coordinator          Coordinator                          `validate:"-"`
}

// APIServer is the api server configuration parameters
type APIServer struct {
	// NodeAPI specifies the configuration parameters of the API
	API struct {
		// Address where the API will listen if set
		Address string `validate:"required" env:"API_ADDRESS"`
		// Explorer enables the Explorer API endpoints
		Explorer bool `env:"API_EXPLORER"`
		// Maximum concurrent connections allowed between API and SQL
		MaxSQLConnections int `validate:"required,gte=1" env:"API_MAXSQLCONNECTIONS"`
		// SQLConnectionTimeout is the maximum amount of time that an API request
		// can wait to establish a SQL connection
		SQLConnectionTimeout Duration `env:"API_SQLCONNECTIONTIMEOUT"`
	} `validate:"required"`
	PostgreSQL  PostgreSQL `validate:"required"`
	Coordinator struct {
		API struct {
			// Coordinator enables the coordinator API endpoints
			Coordinator bool `env:"COORDINATORAPI_COORDINATOR"`
		} `validate:"required"`
		L2DB struct {
			// MaxTxs is the maximum number of pending L2Txs that can be
			// stored in the pool.  Once this number of pending L2Txs is
			// reached, inserts to the pool will be denied until some of
			// the pending txs are forged.
			MaxTxs uint32 `validate:"required" env:"L2DB_MAXTXS"`
			// MinFeeUSD is the minimum fee in USD that a tx must pay in
			// order to be accepted into the pool.  Txs with lower than
			// minimum fee will be rejected at the API level.
			MinFeeUSD float64 `validate:"gte=0" env:"L2DB_MINFEEUSD"`
			// MaxFeeUSD is the maximum fee in USD that a tx must pay in
			// order to be accepted into the pool.  Txs with greater than
			// maximum fee will be rejected at the API level.
			MaxFeeUSD float64 `validate:"required,gte=0" env:"L2DB_MAXFEEUSD"`
		} `validate:"required"`
	}
	Debug NodeDebug `validate:"required"`
}

// LoadNode loads the Node configuration from path.
func LoadNode(path string, coordinator bool) (*Node, error) {
	var cfg Node
	err := configLibrary.LoadConfig(path, DefaultValues, &cfg)
	if err != nil {
		//Split errors depending on if there is a file error, a env error or a default error
		return nil, err
	}
	validate := validator.New()
	validate.RegisterStructValidation(priceupdater.ProviderValidation, priceupdater.Provider{})
	if err := validate.Struct(cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
	}
	if coordinator {
		if err := validate.Struct(cfg.Coordinator); err != nil {
			return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
		}
	}
	return &cfg, nil
}

// LoadAPIServer loads the APIServer configuration from path.
func LoadAPIServer(path string, coordinator bool) (*APIServer, error) {
	var cfg APIServer
	err := configLibrary.LoadConfig(path, DefaultValues, &cfg)
	if err != nil {
		//Split errors depending on if there is a file error, a env error or a default error
		return nil, err
	}
	validate := validator.New()
	validate.RegisterStructValidation(priceupdater.ProviderValidation, priceupdater.Provider{})
	if err := validate.Struct(cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
	}
	if coordinator {
		if err := validate.Struct(cfg.Coordinator); err != nil {
			return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
		}
	}
	return &cfg, nil
}
