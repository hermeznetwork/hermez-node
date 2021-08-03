package eth

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	tokenhez "github.com/hermeznetwork/hermez-node/eth/contracts/tokenhez"
	"github.com/hermeznetwork/tracerr"
)

// TokenClient is the implementation of the interface to the Hez Token Smart Contract in ethereum.
type TokenClient struct {
	client  *EthereumClient
	hez     *tokenhez.Tokenhez
	address ethCommon.Address
	name    string
	opts    *bind.CallOpts
}

// NewTokenClient creates a new TokenClient
func NewTokenClient(client *EthereumClient, address ethCommon.Address) (*TokenClient, error) {
	hez, err := tokenhez.NewTokenhez(address, client.Client())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	opts := newCallOpts()
	name, err := hez.Name(opts)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &TokenClient{
		client:  client,
		hez:     hez,
		address: address,
		name:    name,
		opts:    opts,
	}, nil
}
