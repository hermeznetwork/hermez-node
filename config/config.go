package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/tracerr"
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
	URL string `validate:"required"`
}

// Coordinator is the coordinator specific configuration.
type Coordinator struct {
	// ForgerAddress is the address under which this coordinator is forging
	ForgerAddress ethCommon.Address `validate:"required"`
	// ConfirmBlocks is the number of confirmation blocks to wait for sent
	// ethereum transactions before forgetting about them
	ConfirmBlocks int64 `validate:"required"`
	// L1BatchTimeoutPerc is the portion of the range before the L1Batch
	// timeout that will trigger a schedule to forge an L1Batch
	L1BatchTimeoutPerc float64 `validate:"required"`
	// ProofServerPollInterval is the waiting interval between polling the
	// ProofServer while waiting for a particular status
	ProofServerPollInterval Duration `validate:"required"`
	// SyncRetryInterval is the waiting interval between calls to the main
	// handler of a synced block after an error
	SyncRetryInterval Duration `validate:"required"`
	L2DB              struct {
		SafetyPeriod common.BatchNum `validate:"required"`
		MaxTxs       uint32          `validate:"required"`
		TTL          Duration        `validate:"required"`
		// PurgeBatchDelay is the delay between batches to purge outdated transactions
		PurgeBatchDelay int64 `validate:"required"`
		// InvalidateBatchDelay is the delay between batches to mark invalid transactions
		InvalidateBatchDelay int64 `validate:"required"`
		// PurgeBlockDelay is the delay between blocks to purge outdated transactions
		PurgeBlockDelay int64 `validate:"required"`
		// InvalidateBlockDelay is the delay between blocks to mark invalid transactions
		InvalidateBlockDelay int64 `validate:"required"`
	} `validate:"required"`
	TxSelector struct {
		Path string `validate:"required"`
	} `validate:"required"`
	BatchBuilder struct {
		Path string `validate:"required"`
	} `validate:"required"`
	ServerProofs []ServerProof `validate:"required"`
	EthClient    struct {
		CallGasLimit        uint64   `validate:"required"`
		DeployGasLimit      uint64   `validate:"required"`
		GasPriceDiv         uint64   `validate:"required"`
		ReceiptTimeout      Duration `validate:"required"`
		ReceiptLoopInterval Duration `validate:"required"`
		// CheckLoopInterval is the waiting interval between receipt
		// checks of ethereum transactions in the TxManager
		CheckLoopInterval Duration `validate:"required"`
		// Attempts is the number of attempts to do an eth client RPC
		// call before giving up
		Attempts int `validate:"required"`
		// AttemptsDelay is delay between attempts do do an eth client
		// RPC call
		AttemptsDelay Duration `validate:"required"`
	} `validate:"required"`
	API struct {
		Coordinator bool
	} `validate:"required"`
	Debug struct {
		// BatchPath if set, specifies the path where batchInfo is stored
		// in JSON in every step/update of the pipeline
		BatchPath string
	}
}

// Node is the hermez node configuration.
type Node struct {
	PriceUpdater struct {
		Interval Duration `valudate:"required"`
		URL      string   `valudate:"required"`
		Type     string   `valudate:"required"`
	} `validate:"required"`
	StateDB struct {
		Path string
	} `validate:"required"`
	PostgreSQL struct {
		Port     int    `validate:"required"`
		Host     string `validate:"required"`
		User     string `validate:"required"`
		Password string `validate:"required"`
		Name     string `validate:"required"`
	} `validate:"required"`
	Web3 struct {
		URL string `validate:"required"`
	} `validate:"required"`
	Synchronizer struct {
		SyncLoopInterval   Duration `validate:"required"`
		StatsRefreshPeriod Duration `validate:"required"`
	} `validate:"required"`
	SmartContracts struct {
		Rollup       ethCommon.Address `validate:"required"`
		Auction      ethCommon.Address `validate:"required"`
		WDelayer     ethCommon.Address `validate:"required"`
		TokenHEZ     ethCommon.Address `validate:"required"`
		TokenHEZName string            `validate:"required"`
	} `validate:"required"`
	API struct {
		Address                      string
		Explorer                     bool
		UpdateMetricsInterval        Duration
		UpdateRecommendedFeeInterval Duration
	} `validate:"required"`
	Debug struct {
		APIAddress string
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
