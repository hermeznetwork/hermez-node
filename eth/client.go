package eth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
)

var errTODO = fmt.Errorf("TODO: Not implemented yet")

// ClientInterface is the eth Client interface used by hermez-node modules to
// interact with Ethereum Blockchain and smart contracts.
type ClientInterface interface {
	EthereumInterface
	RollupInterface
	AuctionInterface
}

//
// Implementation
//

// Client is used to interact with Ethereum and the Hermez smart contracts.
type Client struct {
	EthereumClient
	AuctionClient
	RollupClient
}

// NewClient creates a new Client to interact with Ethereum and the Hermez smart contracts.
func NewClient(client *ethclient.Client, account *accounts.Account, ks *ethKeystore.KeyStore, config *EthereumConfig) *Client {
	ethereumClient := NewEthereumClient(client, account, ks, config)
	auctionClient := &AuctionClient{}
	rollupCient := &RollupClient{}
	return &Client{
		EthereumClient: *ethereumClient,
		AuctionClient:  *auctionClient,
		RollupClient:   *rollupCient,
	}
}
