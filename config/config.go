package config

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"time"

	"github.com/BurntSushi/toml"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
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
	URL string `validate:"required"`
}

// Coordinator is the coordinator specific configuration.
type Coordinator struct {
	// ForgerAddress is the address under which this coordinator is forging
	ForgerAddress ethCommon.Address `validate:"required"`
	// FeeAccount is the Hermez account that the coordinator uses to receive fees
	FeeAccount struct {
		// Address is the ethereum address of the account to receive fees
		Address ethCommon.Address `validate:"required"`
		// BJJ is the baby jub jub public key of the account to receive fees
		BJJ babyjub.PublicKeyComp `validate:"required"`
	} `validate:"required"`
	// ConfirmBlocks is the number of confirmation blocks to wait for sent
	// ethereum transactions before forgetting about them
	ConfirmBlocks int64 `validate:"required"`
	// L1BatchTimeoutPerc is the portion of the range before the L1Batch
	// timeout that will trigger a schedule to forge an L1Batch
	L1BatchTimeoutPerc float64 `validate:"required"`
	// StartSlotBlocksDelay is the number of blocks of delay to wait before
	// starting the pipeline when we reach a slot in which we can forge.
	StartSlotBlocksDelay int64
	// ScheduleBatchBlocksAheadCheck is the number of blocks ahead in which
	// the forger address is checked to be allowed to forge (apart from
	// checking the next block), used to decide when to stop scheduling new
	// batches (by stopping the pipeline).
	// For example, if we are at block 10 and ScheduleBatchBlocksAheadCheck
	// is 5, eventhough at block 11 we canForge, the pipeline will be
	// stopped if we can't forge at block 15.
	// This value should be the expected number of blocks it takes between
	// scheduling a batch and having it mined.
	ScheduleBatchBlocksAheadCheck int64
	// SendBatchBlocksMarginCheck is the number of margin blocks ahead in
	// which the coordinator is also checked to be allowed to forge, apart
	// from the next block; used to decide when to stop sending batches to
	// the smart contract.
	// For example, if we are at block 10 and SendBatchBlocksMarginCheck is
	// 5, eventhough at block 11 we canForge, the batch will be discarded
	// if we can't forge at block 15.
	SendBatchBlocksMarginCheck int64
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
		MinFeeUSD float64
		// TTL is the Time To Live for L2Txs in the pool.  Once MaxTxs
		// L2Txs is reached, L2Txs older than TTL will be deleted.
		TTL Duration `validate:"required"`
		// PurgeBatchDelay is the delay between batches to purge
		// outdated transactions.  Oudated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBatchDelay int64 `validate:"required"`
		// InvalidateBatchDelay is the delay between batches to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBatchDelay int64 `validate:"required"`
		// PurgeBlockDelay is the delay between blocks to purge
		// outdated transactions.  Oudated L2Txs are those that have
		// been forged or marked as invalid for longer than the
		// SafetyPeriod and pending L2Txs that have been in the pool
		// for longer than TTL once there are MaxTxs.
		PurgeBlockDelay int64 `validate:"required"`
		// InvalidateBlockDelay is the delay between blocks to mark
		// invalid transactions due to nonce lower than the account
		// nonce.
		InvalidateBlockDelay int64 `validate:"required"`
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
		MaxTx int64 `validate:"required"`
		// NLevels is the maximum number of merkle tree levels
		// supported by the circuit
		NLevels int64 `validate:"required"`
	} `validate:"required"`
	EthClient struct {
		// MaxGasPrice is the maximum gas price allowed for ethereum
		// transactions
		MaxGasPrice *big.Int `validate:"required"`
		// GasPriceIncPerc is the percentage increase of gas price set
		// in an ethereum transaction from the suggested gas price by
		// the ehtereum node
		GasPriceIncPerc int64
		// CheckLoopInterval is the waiting interval between receipt
		// checks of ethereum transactions in the TxManager
		CheckLoopInterval Duration `validate:"required"`
		// Attempts is the number of attempts to do an eth client RPC
		// call before giving up
		Attempts int `validate:"required"`
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
	} `validate:"required"`
	API struct {
		// Coordinator enables the coordinator API endpoints
		Coordinator bool
	} `validate:"required"`
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
}

// Node is the hermez node configuration.
type Node struct {
	PriceUpdater struct {
		// Interval between price updater calls
		Interval Duration `valudate:"required"`
		// URL of the token prices provider
		URL string `valudate:"required"`
		// Type of the API of the token prices provider
		Type string `valudate:"required"`
	} `validate:"required"`
	StateDB struct {
		// Path where the synchronizer StateDB is stored
		Path string `validate:"required"`
		// Keep is the number of checkpoints to keep
		Keep int `validate:"required"`
	} `validate:"required"`
	PostgreSQL struct {
		// Port of the PostgreSQL server
		Port int `validate:"required"`
		// Host of the PostgreSQL server
		Host string `validate:"required"`
		// User of the PostgreSQL server
		User string `validate:"required"`
		// Password of the PostgreSQL server
		Password string `validate:"required"`
		// Name of the PostgreSQL server database
		Name string `validate:"required"`
	} `validate:"required"`
	Web3 struct {
		// URL is the URL of the web3 ethereum-node RPC server
		URL string `validate:"required"`
	} `validate:"required"`
	Synchronizer struct {
		// SyncLoopInterval is the interval between attempts to
		// synchronize a new block from an ethereum node
		SyncLoopInterval Duration `validate:"required"`
		// StatsRefreshPeriod is the interval between updates of the
		// synchronizer state Eth parameters (`Eth.LastBlock` and
		// `Eth.LastBatch`).  This value only affects the reported % of
		// synchronization of blocks and batches, nothing else.
		StatsRefreshPeriod Duration `validate:"required"`
		// StoreAccountUpdates when set to true makes the synchronizer
		// store every account update in the account_update SQL table.
		// This allows querying nonces and balances from the HistoryDB
		// via SQL.
		StoreAccountUpdates bool
	} `validate:"required"`
	SmartContracts struct {
		// Rollup is the address of the Hermez.sol smart contract
		Rollup ethCommon.Address `validate:"required"`
		// Rollup is the address of the HermezAuctionProtocol.sol smart
		// contract
		Auction ethCommon.Address `validate:"required"`
		// WDelayer is the address of the WithdrawalDelayer.sol smart
		// contract
		WDelayer ethCommon.Address `validate:"required"`
		// TokenHEZ is the address of the HEZTokenFull.sol smart
		// contract
		TokenHEZ ethCommon.Address `validate:"required"`
		// TokenHEZName is the name of the HEZ token deployed at
		// TokenHEZ address
		TokenHEZName string `validate:"required"`
	} `validate:"required"`
	API struct {
		// Address where the API will listen if set
		Address string
		// Explorer enables the Explorer API endpoints
		Explorer bool
		// UpdateMetricsInterval is the interval between updates of the
		// API metrics
		UpdateMetricsInterval Duration
		// UpdateRecommendedFeeInterval is the interval between updates of the
		// recommended fees
		UpdateRecommendedFeeInterval Duration
		// Maximum concurrent connections allowed between API and SQL
		MaxSQLConnections int `validate:"required"`
		// SQLConnectionTimeout is the maximum amount of time that an API request
		// can wait to stablish a SQL connection
		SQLConnectionTimeout Duration
	} `validate:"required"`
	Debug struct {
		// APIAddress is the address where the debugAPI will listen if
		// set
		APIAddress string
		// MeddlerLogs enables meddler debug mode, where unused columns and struct
		// fields will be logged
		MeddlerLogs bool
	}
	Coordinator Coordinator `validate:"-"`
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

// LoadCoordinator loads the Coordinator configuration from path.
func LoadCoordinator(path string) (*Node, error) {
	var cfg Node
	if err := Load(path, &cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error loading node configuration file: %w", err))
	}
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
	}
	if err := validate.Struct(cfg.Coordinator); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
	}
	return &cfg, nil
}

// LoadNode loads the Node configuration from path.
func LoadNode(path string) (*Node, error) {
	var cfg Node
	if err := Load(path, &cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error loading node configuration file: %w", err))
	}
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("error validating configuration file: %w", err))
	}
	return &cfg, nil
}
