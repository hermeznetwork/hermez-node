package eth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/tracerr"
)

var errTODO = fmt.Errorf("TODO: Not implemented yet")

const (
	blocksPerDay = (3600 * 24) / 15 //nolint:gomnd
)

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// ClientInterface is the eth Client interface used by hermez-node modules to
// interact with Ethereum Blockchain and smart contracts.
type ClientInterface interface {
	EthereumInterface
	RollupInterface
	AuctionInterface
	WDelayerInterface
}

//
// Implementation
//

// Client is used to interact with Ethereum and the Hermez smart contracts.
type Client struct {
	EthereumClient
	AuctionEthClient
	RollupClient
	WDelayerClient
}

// RollupConfig is the configuration for the Rollup smart contract interface
type RollupConfig struct {
	Address ethCommon.Address
}

// ClientConfig is the configuration of the Client
type ClientConfig struct {
	Ethereum EthereumConfig
	Rollup   RollupConfig
}

// NewClient creates a new Client to interact with Ethereum and the Hermez smart contracts.
func NewClient(client *ethclient.Client, account *accounts.Account, ks *ethKeystore.KeyStore,
	cfg *ClientConfig) (*Client, error) {
	ethereumClient, err := NewEthereumClient(client, account, ks, &cfg.Ethereum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rollupClient, err := NewRollupClient(ethereumClient, cfg.Rollup.Address)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auctionClient, err := NewAuctionClient(ethereumClient,
		rollupClient.consts.HermezAuctionContract,
		rollupClient.consts.TokenHEZ)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	wDelayerClient, err := NewWDelayerClient(ethereumClient,
		rollupClient.consts.WithdrawDelayerContract)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &Client{
		EthereumClient:   *ethereumClient,
		AuctionEthClient: *auctionClient,
		RollupClient:     *rollupClient,
		WDelayerClient:   *wDelayerClient,
	}, nil
}
