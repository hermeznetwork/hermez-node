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
	AuctionClient
	RollupClient
	WDelayerClient
}

// TokenConfig is used to define the information about token
type TokenConfig struct {
	Address ethCommon.Address
	Name    string
}

// RollupConfig is the configuration for the Rollup smart contract interface
type RollupConfig struct {
	Address ethCommon.Address
}

// AuctionConfig is the configuration for the Auction smart contract interface
type AuctionConfig struct {
	Address  ethCommon.Address
	TokenHEZ TokenConfig
}

// WDelayerConfig is the configuration for the WDelayer smart contract interface
type WDelayerConfig struct {
	Address ethCommon.Address
}

// ClientConfig is the configuration of the Client
type ClientConfig struct {
	Ethereum EthereumConfig
	Rollup   RollupConfig
	Auction  AuctionConfig
	WDelayer WDelayerConfig
}

// NewClient creates a new Client to interact with Ethereum and the Hermez smart contracts.
func NewClient(client *ethclient.Client, account *accounts.Account, ks *ethKeystore.KeyStore,
	cfg *ClientConfig) (*Client, error) {
	ethereumClient, err := NewEthereumClient(client, account, ks, &cfg.Ethereum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auctionClient, err := NewAuctionClient(ethereumClient, cfg.Auction.Address,
		cfg.Auction.TokenHEZ)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rollupClient, err := NewRollupClient(ethereumClient, cfg.Rollup.Address,
		cfg.Auction.TokenHEZ)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	wDelayerClient, err := NewWDelayerClient(ethereumClient, cfg.WDelayer.Address)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &Client{
		EthereumClient: *ethereumClient,
		AuctionClient:  *auctionClient,
		RollupClient:   *rollupClient,
		WDelayerClient: *wDelayerClient,
	}, nil
}
