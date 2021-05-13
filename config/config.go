package config

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/BurntSushi/toml"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/stateapiupdater"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/priceupdater"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"gopkg.in/go-playground/validator.v9"
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
	URL string `validate:"required,url"`
}

// ForgeBatchGasCost is the costs associated to a ForgeBatch transaction, split
// into different parts to be used in a formula.
type ForgeBatchGasCost struct {
	Fixed     uint64 `validate:"required"`
	L1UserTx  uint64 `validate:"required"`
	L1CoordTx uint64 `validate:"required"`
	L2Tx      uint64 `validate:"required"`
}

// CoordinatorAPI specifies the configuration parameters of the API in mode
// coordinator
type CoordinatorAPI struct {
	// Coordinator enables the coordinator API endpoints
	Coordinator bool
}

// Coordinator is the coordinator specific configuration.
type Coordinator struct {
	// ForgerAddress is the address under which this coordinator is forging
	ForgerAddress ethCommon.Address `validate:"required"`
	// MinimumForgeAddressBalance is the minimum balance the forger address
	// needs to start the coordinator in wei. Of set to 0, the coordinator
	// will not check the balance before starting.
	MinimumForgeAddressBalance *big.Int
	// FeeAccount is the Hermez account that the coordinator uses to receive fees
	FeeAccount struct {
		// Address is the ethereum address of the account to receive fees
		Address ethCommon.Address `validate:"required"`
		// BJJ is the baby jub jub public key of the account to receive fees
		BJJ babyjub.PublicKeyComp `validate:"required"`
	} `validate:"required"`
	// ConfirmBlocks is the number of confirmation blocks to wait for sent
	// ethereum transactions before forgetting about them
	ConfirmBlocks int64 `validate:"required,gte=0"`
	// L1BatchTimeoutPerc is the portion of the range before the L1Batch
	// timeout that will trigger a schedule to forge an L1Batch
	L1BatchTimeoutPerc float64 `validate:"required,lte=1.0,gte=0.0"`
	// StartSlotBlocksDelay is the number of blocks of delay to wait before
	// starting the pipeline when we reach a slot in which we can forge.
	StartSlotBlocksDelay int64 `validate:"gte=0"`
	// ScheduleBatchBlocksAheadCheck is the number of blocks ahead in which
	// the forger address is checked to be allowed to forge (apart from
	// checking the next block), used to decide when to stop scheduling new
	// batches (by stopping the pipeline).
	// For example, if we are at block 10 and ScheduleBatchBlocksAheadCheck
	// is 5, even though at block 11 we canForge, the pipeline will be
	// stopped if we can't forge at block 15.
	// This value should be the expected number of blocks it takes between
	// scheduling a batch and having it mined.
	ScheduleBatchBlocksAheadCheck int64 `validate:"gte=0"`
	// SendBatchBlocksMarginCheck is the number of margin blocks ahead in
	// which the coordinator is also checked to be allowed to forge, apart
	// from the next block; used to decide when to stop sending batches to
	// the smart contract.
	// For example, if we are at block 10 and SendBatchBlocksMarginCheck is
	// 5, even though at block 11 we canForge, the batch will be discarded
	// if we can't forge at block 15.
	SendBatchBlocksMarginCheck int64 `validate:"gte=0"`
	// ProofServerPollInterval is the waiting interval between polling the
	// ProofServer while waiting for a particular status
	ProofServerPollInterval Duration `validate:"required"`
	// ForgeRetryInterval is the waiting interval between calls forge a
	// batch after an error
	ForgeRetryInterval Duration `validate:"required"`
	// ForgeDelay is the delay after which a batch is forged if the slot is
	// already committed.  If set to 0s, the coordinator will continuously
	// forge at the maximum rate.
	ForgeDelay Duration `validate:"-"`
	// ForgeNoTxsDelay is the delay after which a batch is forged even if
	// there are no txs to forge if the slot is already committed.  If set
	// to 0s, the coordinator will continuously forge even if the batches
	// are empty.
	ForgeNoTxsDelay Duration `validate:"-"`
	// MustForgeAtSlotDeadline enables the coordinator to forge slots if
	// the empty slots reach the slot deadline.
	MustForgeAtSlotDeadline bool
	// IgnoreSlotCommitment disables forcing the coordinator to forge a
	// slot immediately when the slot is not committed. If set to false,
	// the coordinator will immediately forge a batch at the beginning of a
	// slot if it's the slot winner.
	IgnoreSlotCommitment bool
	// ForgeOncePerSlotIfTxs will make the coordinator forge at most one
	// batch per slot, only if there are included txs in that batch, or
	// pending l1UserTxs in the smart contract.  Setting this parameter
	// overrides `ForgeDelay`, `ForgeNoTxsDelay`, `MustForgeAtSlotDeadline`
	// and `IgnoreSlotCommitment`.
	ForgeOncePerSlotIfTxs bool
	// SyncRetryInterval is the waiting interval between calls to the main
	// handler of a synced block after an error
	SyncRetryInterval Duration `validate:"required"`
	// PurgeByExtDelInterval is the waiting interval between calls
	// to the PurgeByExternalDelete function of the l2db which deletes
	// pending txs externally marked by the column `external_delete`
	PurgeByExtDelInterval Duration `validate:"required"`
	// L2DB is the DB that holds the pool of L2Txs
	L2DB struct {
		// SafetyPeriod is the number of batches after which
		// non-pending L2Txs are deleted from the pool
		SafetyPeriod common.BatchNum `validate:"required"`
		// MaxTxs is the maximum number of pending L2Txs that can be
		// stored in the pool.  Once this number of pending L2Txs is
		// reached, inserts to the pool will be denied until some of
		// the pending txs are forged.
		MaxTxs uint32 `validate:"required"`
		// MinFeeUSD is the minimum fee in USD that a tx must pay in
		// order to be accepted into the pool.  Txs with lower than
		// minimum fee will be rejected at the API level.
		MinFeeUSD float64 `validate:"gte=0"`
		// MaxFeeUSD is the maximum fee in USD that a tx must pay in
		// order to be accepted into the pool.  Txs with greater than
		// maximum fee will be rejected at the API level.
		MaxFeeUSD float64 `validate:"required,gte=0"`
		// TTL is the Time To Live for L2Txs in the pool. L2Txs older
		// than TTL will be deleted.
		TTL Duration `validate:"required"`
		// PurgeBatchDelay is the delay between batches to purge
		// outdated transactions. Outdated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBatchDelay int64 `validate:"required,gte=0"`
		// InvalidateBatchDelay is the delay between batches to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBatchDelay int64 `validate:"required"`
		// PurgeBlockDelay is the delay between blocks to purge
		// outdated transactions. Outdated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBlockDelay int64 `validate:"required,gte=0"`
		// InvalidateBlockDelay is the delay between blocks to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBlockDelay int64 `validate:"required,gte=0"`
	} `validate:"required"`
	TxSelector struct {
		// Path where the TxSelector StateDB is stored
		Path string `validate:"required"`
	} `validate:"required"`
	BatchBuilder struct {
		// Path where the BatchBuilder StateDB is stored
		Path string `validate:"required"`
	} `validate:"required"`
	ServerProofs []ServerProof `validate:"required"`
	Circuit      struct {
		// MaxTx is the maximum number of txs supported by the circuit
		MaxTx int64 `validate:"required,gte=0"`
		// NLevels is the maximum number of merkle tree levels
		// supported by the circuit
		NLevels int64 `validate:"required,gte=0"`
	} `validate:"required"`
	EthClient struct {
		// MaxGasPrice is the maximum gas price allowed for ethereum
		// transactions
		MaxGasPrice *big.Int `validate:"required"`
		// GasPriceIncPerc is the percentage increase of gas price set
		// in an ethereum transaction from the suggested gas price by
		// the ethereum node
		GasPriceIncPerc int64 `validate:"gte=0"`
		// CheckLoopInterval is the waiting interval between receipt
		// checks of ethereum transactions in the TxManager
		CheckLoopInterval Duration `validate:"required"`
		// Attempts is the number of attempts to do an eth client RPC
		// call before giving up
		Attempts int `validate:"required,gte=1"`
		// AttemptsDelay is delay between attempts do do an eth client
		// RPC call
		AttemptsDelay Duration `validate:"required"`
		// TxResendTimeout is the timeout after which a non-mined
		// ethereum transaction will be resent (reusing the nonce) with
		// a newly calculated gas price
		TxResendTimeout Duration `validate:"required"`
		// NoReuseNonce disables reusing nonces of pending transactions for
		// new replacement transactions
		NoReuseNonce bool
		// Keystore is the ethereum keystore where private keys are kept
		Keystore struct {
			// Path to the keystore
			Path string `validate:"required"`
			// Password used to decrypt the keys in the keystore
			Password string `validate:"required"`
		} `validate:"required"`
		// ForgeBatchGasCost contains the cost of each action in the
		// ForgeBatch transaction.
		ForgeBatchGasCost ForgeBatchGasCost `validate:"required"`
	} `validate:"required"`
	API   CoordinatorAPI `validate:"required"`
	Debug struct {
		// BatchPath if set, specifies the path where batchInfo is stored
		// in JSON in every step/update of the pipeline
		BatchPath string
		// LightScrypt if set, uses light parameters for the ethereum
		// keystore encryption algorithm.
		LightScrypt bool
		// RollupVerifierIndex is the index of the verifier to use in
		// the Rollup smart contract.  The verifier chosen by index
		// must match with the Circuit parameters.
		RollupVerifierIndex *int
	}
	Etherscan struct {
		// URL if set, specifies the etherscan endpoint to get
		// the gas estimations for that moment.
		URL string
		// APIKey allow access to etherscan services
		APIKey string
	}
}

// PostgreSQL is the postgreSQL configuration parameters.  It's possible to use
// differentiated SQL connections for read/write.  If the read configuration is
// not provided, the write one it's going to be used for both reads and writes
type PostgreSQL struct {
	// Port of the PostgreSQL write server
	PortWrite int `validate:"required"`
	// Host of the PostgreSQL write server
	HostWrite string `validate:"required"`
	// User of the PostgreSQL write server
	UserWrite string `validate:"required"`
	// Password of the PostgreSQL write server
	PasswordWrite string `validate:"required"`
	// Name of the PostgreSQL write server database
	NameWrite string `validate:"required"`
	// Port of the PostgreSQL read server
	PortRead int
	// Host of the PostgreSQL read server
	HostRead string `validate:"nefield=HostWrite"`
	// User of the PostgreSQL read server
	UserRead string
	// Password of the PostgreSQL read server
	PasswordRead string
	// Name of the PostgreSQL read server database
	NameRead string
}

// NodeDebug specifies debug configuration parameters
type NodeDebug struct {
	// APIAddress is the address where the debugAPI will listen if
	// set
	APIAddress string
	// MeddlerLogs enables meddler debug mode, where unused columns and struct
	// fields will be logged
	MeddlerLogs bool
	// GinDebugMode sets Gin-Gonic (the web framework) to run in
	// debug mode
	GinDebugMode bool
}

// Node is the hermez node configuration.
type Node struct {
	PriceUpdater struct {
		// Interval between price updater calls
		Interval Duration `validate:"required"`
		// Priority option defines the priority provider
		Priority priceupdater.Priority `validate:"required"`
		// TokensConfig to specify how each token get it's price updated
		Provider []priceupdater.Provider
	} `validate:"required"`
	StateDB struct {
		// Path where the synchronizer StateDB is stored
		Path string `validate:"required"`
		// Keep is the number of checkpoints to keep
		Keep int `validate:"required,gte=128"`
	} `validate:"required"`
	PostgreSQL PostgreSQL `validate:"required"`
	Web3       struct {
		// URL is the URL of the web3 ethereum-node RPC server.  Only
		// geth is officially supported.
		URL string `validate:"required,url"`
	} `validate:"required"`
	Synchronizer struct {
		// SyncLoopInterval is the interval between attempts to
		// synchronize a new block from an ethereum node
		SyncLoopInterval Duration `validate:"required"`
		// StatsUpdateBlockNumDiffThreshold is a threshold of a number of
		// Ethereum blocks left to synchronize, such that if there are more
		// blocks to sync than the defined value synchronizer can aggressively
		// skip calling UpdateEth to save network bandwidth and time.
		// After reaching the threshold UpdateEth is called on each block.
		// This value only affects the reported % of synchronization of
		// blocks and batches, nothing else.
		StatsUpdateBlockNumDiffThreshold uint16 `validate:"required,gt=32"`
		// StatsUpdateFrequencyDivider - While having more blocks to sync than
		// updateEthBlockNumThreshold, UpdateEth will be called once in a
		// defined number of blocks. This value only affects the reported % of
		// synchronization of blocks and batches, nothing else.
		StatsUpdateFrequencyDivider uint16 `validate:"required,gt=1"`
	} `validate:"required"`
	SmartContracts struct {
		// Rollup is the address of the Hermez.sol smart contract
		Rollup ethCommon.Address `validate:"required"`
	} `validate:"required"`
	// API specifies the configuration parameters of the API
	API struct {
		// Address where the API will listen if set
		Address string
		// Explorer enables the Explorer API endpoints
		Explorer bool
		// UpdateMetricsInterval is the interval between updates of the
		// API metrics
		UpdateMetricsInterval Duration `validate:"required"`
		// UpdateRecommendedFeeInterval is the interval between updates of the
		// recommended fees
		UpdateRecommendedFeeInterval Duration `validate:"required"`
		// Maximum concurrent connections allowed between API and SQL
		MaxSQLConnections int `validate:"required,gte=1"`
		// SQLConnectionTimeout is the maximum amount of time that an API request
		// can wait to establish a SQL connection
		SQLConnectionTimeout Duration
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
		Address string `validate:"required"`
		// Explorer enables the Explorer API endpoints
		Explorer bool
		// Maximum concurrent connections allowed between API and SQL
		MaxSQLConnections int `validate:"required,gte=1"`
		// SQLConnectionTimeout is the maximum amount of time that an API request
		// can wait to establish a SQL connection
		SQLConnectionTimeout Duration
	} `validate:"required"`
	PostgreSQL  PostgreSQL `validate:"required"`
	Coordinator struct {
		API struct {
			// Coordinator enables the coordinator API endpoints
			Coordinator bool
		} `validate:"required"`
		L2DB struct {
			// MaxTxs is the maximum number of pending L2Txs that can be
			// stored in the pool.  Once this number of pending L2Txs is
			// reached, inserts to the pool will be denied until some of
			// the pending txs are forged.
			MaxTxs uint32 `validate:"required"`
			// MinFeeUSD is the minimum fee in USD that a tx must pay in
			// order to be accepted into the pool.  Txs with lower than
			// minimum fee will be rejected at the API level.
			MinFeeUSD float64 `validate:"gte=0"`
			// MaxFeeUSD is the maximum fee in USD that a tx must pay in
			// order to be accepted into the pool.  Txs with greater than
			// maximum fee will be rejected at the API level.
			MaxFeeUSD float64 `validate:"required,gte=0"`
		} `validate:"required"`
	}
	Debug NodeDebug `validate:"required"`
}

// Load loads a generic config.
func Load(path string, cfg interface{}) error {
	bs, err := ioutil.ReadFile(path) //nolint:gosec
	if err != nil {
		return tracerr.Wrap(err)
	}
	cfgToml := string(bs)
	if _, err := toml.Decode(cfgToml, cfg); err != nil {
		return tracerr.Wrap(err)
	}
	return nil
}

// LoadNode loads the Node configuration from path.
func LoadNode(path string, coordinator bool) (*Node, error) {
	var cfg Node
	if err := Load(path, &cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error loading node configuration file: %w", err))
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
	if err := Load(path, &cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error loading apiServer configuration file: %w", err))
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
