package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"gopkg.in/go-playground/validator.v9"
)

// Duration is a wrapper type that parses time duration from text.
type Duration struct {
	time.Duration
}

// UnmarshalText unmarshalls time duration from text.
func (d *Duration) UnmarshalText(data []byte) error {
	duration, err := time.ParseDuration(string(data))
	if err != nil {
		return err
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
	ForgerAddress     ethCommon.Address `validate:"required"`
	ForgeLoopInterval Duration          `validate:"required"`
	L2DB              struct {
		SafetyPeriod common.BatchNum `validate:"required"`
		MaxTxs       uint32          `validate:"required"`
		TTL          Duration        `validate:"required"`
	} `validate:"required"`
	TxSelector struct {
		Path string `validate:"required"`
	} `validate:"required"`
	BatchBuilder struct {
		Path string `validate:"required"`
	} `validate:"required"`
	ServerProofs []ServerProof `validate:"required"`
}

// Node is the hermez node configuration.
type Node struct {
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
		SyncLoopInterval Duration `validate:"required"`
	} `validate:"required"`
	SmartContracts struct {
		Rollup   ethCommon.Address `validate:"required"`
		Auction  ethCommon.Address `validate:"required"`
		TokenHEZ ethCommon.Address `validate:"required"`
	} `validate:"required"`
	EthClient struct {
		CallGasLimit        uint64   `validate:"required"`
		DeployGasLimit      uint64   `validate:"required"`
		GasPriceDiv         uint64   `validate:"required"`
		ReceiptTimeout      Duration `validate:"required"`
		IntervalReceiptLoop Duration `validate:"required"`
	} `validate:"required"`
}

// Load loads a generic config.
func Load(path string, cfg interface{}) error {
	bs, err := ioutil.ReadFile(path) //nolint:gosec
	if err != nil {
		return err
	}
	cfgToml := string(bs)
	if _, err := toml.Decode(cfgToml, cfg); err != nil {
		return err
	}
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("error validating configuration file: %w", err)
	}
	return nil
}

// LoadCoordinator loads the Coordinator configuration from path.
func LoadCoordinator(path string) (*Coordinator, error) {
	var cfg Coordinator
	if err := Load(path, &cfg); err != nil {
		return nil, fmt.Errorf("error loading coordinator configuration file: %w", err)
	}
	return &cfg, nil
}

// LoadNode loads the Node configuration from path.
func LoadNode(path string) (*Node, error) {
	var cfg Node
	if err := Load(path, &cfg); err != nil {
		return nil, fmt.Errorf("error loading node configuration file: %w", err)
	}
	return &cfg, nil
}
