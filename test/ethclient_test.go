package test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientInterface(t *testing.T) {
	var c eth.ClientInterface
	client := NewTestEthClient(true, 1000)
	c = client
	require.NotNil(t, c)
}

func TestEthClient(t *testing.T) {
	c := NewTestEthClient(true, 1000)

	block, err := c.BlockByNumber(context.TODO(), big.NewInt(3))
	assert.Nil(t, err)
	assert.Equal(t, uint64(3), block.EthBlockNum)
	assert.Equal(t, time.Unix(3, 0), block.Timestamp)
	assert.Equal(t, "0x6b0ab5a7a0ebf5f05cef3b49bc7a9739de06469a4e05557d802ee828fdf5187e", block.Hash.Hex())

	header, err := c.HeaderByNumber(context.TODO(), big.NewInt(4))
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(4), header.Number)
	assert.Equal(t, uint64(4), header.Time)
	assert.Equal(t, "0x66cdb12322040a5a345ad29cea66ca97c14d6142b53987010947c8c008e26913", header.Hash().Hex())

	assert.Equal(t, big.NewInt(1000), c.blockNum)
	c.Advance()
	assert.Equal(t, big.NewInt(1001), c.blockNum)
	c.Advance()
	assert.Equal(t, big.NewInt(1002), c.blockNum)

	c.SetBlockNum(big.NewInt(5000))
	assert.Equal(t, big.NewInt(5000), c.blockNum)
	c.Advance()
	assert.Equal(t, big.NewInt(5001), c.blockNum)
}
