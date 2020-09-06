package test

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/log"
)

// Client implements the eth.IClient interface, allowing to manipulate the
// values for testing, working with deterministic results.
type Client struct {
	log      bool
	blockNum *big.Int
}

// NewTestEthClient returns a new test Client that implements the eth.IClient
// interface, at the given initialBlockNumber.
func NewTestEthClient(l bool, initialBlockNumber int64) *Client {
	return &Client{
		log:      l,
		blockNum: big.NewInt(initialBlockNumber),
	}
}

// Advance moves one block forward
func (c *Client) Advance() {
	c.blockNum = c.blockNum.Add(c.blockNum, big.NewInt(1))
	if c.log {
		log.Debugf("TestEthClient blockNum advanced: %d", c.blockNum)
	}
}

// SetBlockNum sets the Client.blockNum to the given blockNum
func (c *Client) SetBlockNum(blockNum *big.Int) {
	c.blockNum = blockNum
	if c.log {
		log.Debugf("TestEthClient blockNum set to: %d", c.blockNum)
	}
}

// CurrentBlock returns the current blockNum
func (c *Client) CurrentBlock() (*big.Int, error) {
	return c.blockNum, nil
}

func newHeader(number *big.Int) *types.Header {
	return &types.Header{
		Number: number,
		Time:   uint64(number.Int64()),
	}
}

// HeaderByNumber returns the *types.Header for the given block number in a
// deterministic way.
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return newHeader(number), nil
}

// BlockByNumber returns the *common.Block for the given block number in a
// deterministic way.
func (c *Client) BlockByNumber(ctx context.Context, number *big.Int) (*common.Block, error) {
	header := newHeader(number)

	return &common.Block{
		EthBlockNum: uint64(number.Int64()),
		Timestamp:   time.Unix(number.Int64(), 0),
		Hash:        header.Hash(),
	}, nil
}

// ForgeCall send the *common.CallDataForge to the Forge method of the mock
// smart contract
func (c *Client) ForgeCall(*common.CallDataForge) ([]byte, error) {
	return nil, nil
}
