package test

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"math/big"
	"testing"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timer struct {
	time int64
}

func (t *timer) Time() int64 {
	currentTime := t.time
	t.time++
	return currentTime
}

func TestClientInterface(t *testing.T) {
	var c eth.ClientInterface
	var timer timer
	clientSetup := NewClientSetupExample()
	client := NewClient(true, &timer, &ethCommon.Address{}, clientSetup)
	c = client
	require.NotNil(t, c)
}

func TestClientEth(t *testing.T) {
	var timer timer
	clientSetup := NewClientSetupExample()
	c := NewClient(true, &timer, &ethCommon.Address{}, clientSetup)
	blockNum, err := c.EthCurrentBlock()
	require.Nil(t, err)
	assert.Equal(t, int64(0), blockNum)

	block, err := c.EthBlockByNumber(context.TODO(), 0)
	require.Nil(t, err)
	assert.Equal(t, int64(0), block.EthBlockNum)
	assert.Equal(t, time.Unix(0, 0), block.Timestamp)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", block.Hash.Hex())
	assert.Equal(t, int64(0), c.blockNum)

	// Mine some empty blocks

	c.CtlMineBlock()
	assert.Equal(t, int64(1), c.blockNum)
	c.CtlMineBlock()
	assert.Equal(t, int64(2), c.blockNum)

	block, err = c.EthBlockByNumber(context.TODO(), 2)
	require.Nil(t, err)
	assert.Equal(t, int64(2), block.EthBlockNum)
	assert.Equal(t, time.Unix(2, 0), block.Timestamp)

	// Add a token
	tokenAddr := ethCommon.HexToAddress("0x44021007485550008e0f9f1f7b506c7d970ad8ce")
	constants := eth.ERC20Consts{
		Name:     "FooBar",
		Symbol:   "FOO",
		Decimals: 4,
	}
	c.CtlAddERC20(tokenAddr, constants)
	c.CtlMineBlock()
	tokenConstants, err := c.EthERC20Consts(tokenAddr)
	require.Nil(t, err)
	assert.Equal(t, constants, *tokenConstants)
}

func TestClientAuction(t *testing.T) {
	addrBidder1 := ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	addrBidder2 := ethCommon.HexToAddress("0xc27cadc437d067a6ec869502cc9f7F834cFc087a")
	addrForge := ethCommon.HexToAddress("0xCfAA413eEb796f328620a3630Ae39124cabcEa92")
	addrForge2 := ethCommon.HexToAddress("0x1fCb4ac309428feCc61B1C8cA5823C15A5e1a800")

	var timer timer
	clientSetup := NewClientSetupExample()
	clientSetup.AuctionVariables.ClosedAuctionSlots = 2
	clientSetup.AuctionVariables.OpenAuctionSlots = 4320
	clientSetup.AuctionVariables.DefaultSlotSetBid = [6]*big.Int{
		big.NewInt(1000), big.NewInt(1100), big.NewInt(1200),
		big.NewInt(1300), big.NewInt(1400), big.NewInt(1500)}
	c := NewClient(true, &timer, &addrBidder1, clientSetup)

	// Check several cases in which bid doesn't succed, and also do 2 successful bids.

	_, err := c.AuctionBid(0, big.NewInt(1))
	assert.Equal(t, errBidClosed, err)

	_, err = c.AuctionBid(4322, big.NewInt(1))
	assert.Equal(t, errBidNotOpen, err)

	// 101 % 6 = 5;  defaultSlotSetBid[5] = 1500;  1500 + 10% = 1650
	_, err = c.AuctionBid(101, big.NewInt(1650))
	assert.Equal(t, errCoordNotReg, err)

	_, err = c.AuctionSetCoordinator(addrForge, "https://foo.bar")
	assert.Nil(t, err)

	_, err = c.AuctionBid(3, big.NewInt(1))
	assert.Equal(t, errBidBelowMin, err)

	_, err = c.AuctionBid(3, big.NewInt(1650))
	assert.Nil(t, err)

	c.CtlSetAddr(addrBidder2)
	_, err = c.AuctionSetCoordinator(addrForge2, "https://foo2.bar")
	assert.Nil(t, err)

	_, err = c.AuctionBid(3, big.NewInt(16))
	assert.Equal(t, errBidBelowMin, err)

	// 1650 + 10% = 1815
	_, err = c.AuctionBid(3, big.NewInt(1815))
	assert.Nil(t, err)

	c.CtlMineBlock()

	blockNum, err := c.EthCurrentBlock()
	require.Nil(t, err)

	auctionEvents, _, err := c.AuctionEventsByBlock(blockNum)
	require.Nil(t, err)
	assert.Equal(t, 2, len(auctionEvents.NewBid))
}

func TestClientRollup(t *testing.T) {
	token1Addr := ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")

	var timer timer
	clientSetup := NewClientSetupExample()
	c := NewClient(true, &timer, &ethCommon.Address{}, clientSetup)

	// Add a token

	tx, err := c.RollupAddToken(token1Addr)
	require.Nil(t, err)
	assert.NotNil(t, tx)

	// Add some L1UserTxs
	// Create Accounts

	const N = 16
	var keys [N]*keys
	for i := 0; i < N; i++ {
		keys[i] = genKeys(int64(i))
		l1UserTx := common.L1Tx{
			FromIdx:     0,
			FromEthAddr: keys[i].Addr,
			FromBJJ:     keys[i].BJJPublicKey,
			TokenID:     common.TokenID(0),
			LoadAmount:  big.NewInt(10 + int64(i)),
		}
		c.CtlAddL1TxUser(&l1UserTx)
	}
	c.CtlMineBlock()

	blockNum, err := c.EthCurrentBlock()
	require.Nil(t, err)
	rollupEvents, _, err := c.RollupEventsByBlock(blockNum)
	require.Nil(t, err)
	assert.Equal(t, N, len(rollupEvents.L1UserTx))
	assert.Equal(t, 1, len(rollupEvents.AddToken))

	// Forge a batch

	c.CtlAddBatch(&eth.RollupForgeBatchArgs{
		NewLastIdx:        0,
		NewStRoot:         big.NewInt(1),
		NewExitRoot:       big.NewInt(100),
		L1CoordinatorTxs:  []*common.L1Tx{},
		L2TxsData:         []*common.L2Tx{},
		FeeIdxCoordinator: make([]common.Idx, eth.RollupConstFeeIdxCoordinatorLen),
		VerifierIdx:       0,
		L1Batch:           true,
	})
	c.CtlMineBlock()

	blockNumA, err := c.EthCurrentBlock()
	require.Nil(t, err)
	rollupEvents, hashA, err := c.RollupEventsByBlock(blockNumA)
	require.Nil(t, err)
	assert.Equal(t, 0, len(rollupEvents.L1UserTx))
	assert.Equal(t, 0, len(rollupEvents.AddToken))
	assert.Equal(t, 1, len(rollupEvents.ForgeBatch))

	// Simulate reorg discarding last mined block

	c.CtlRollback()
	c.CtlMineBlock()

	blockNumB, err := c.EthCurrentBlock()
	require.Nil(t, err)
	rollupEvents, hashB, err := c.RollupEventsByBlock(blockNumA)
	require.Nil(t, err)
	assert.Equal(t, 0, len(rollupEvents.L1UserTx))
	assert.Equal(t, 0, len(rollupEvents.AddToken))
	assert.Equal(t, 0, len(rollupEvents.ForgeBatch))

	assert.Equal(t, blockNumA, blockNumB)
	assert.NotEqual(t, hashA, hashB)

	// Forge again
	rollupForgeBatchArgs0 := &eth.RollupForgeBatchArgs{
		NewLastIdx:        0,
		NewStRoot:         big.NewInt(1),
		NewExitRoot:       big.NewInt(100),
		L1CoordinatorTxs:  []*common.L1Tx{},
		L2TxsData:         []*common.L2Tx{},
		FeeIdxCoordinator: make([]common.Idx, eth.RollupConstFeeIdxCoordinatorLen),
		VerifierIdx:       0,
		L1Batch:           true,
	}
	c.CtlAddBatch(rollupForgeBatchArgs0)
	c.CtlMineBlock()

	// Retrieve ForgeBatchArguments starting from the events

	blockNum, err = c.EthCurrentBlock()
	require.Nil(t, err)
	rollupEvents, _, err = c.RollupEventsByBlock(blockNum)
	require.Nil(t, err)

	rollupForgeBatchArgs1, err := c.RollupForgeBatchArgs(rollupEvents.ForgeBatch[0].EthTxHash)
	require.Nil(t, err)
	assert.Equal(t, rollupForgeBatchArgs0, rollupForgeBatchArgs1)
}

type keys struct {
	BJJSecretKey *babyjub.PrivateKey
	BJJPublicKey *babyjub.PublicKey
	Addr         ethCommon.Address
}

func genKeys(i int64) *keys {
	i++ // i = 0 doesn't work for the ecdsa key generation
	var sk babyjub.PrivateKey
	binary.LittleEndian.PutUint64(sk[:], uint64(i))

	// eth address
	var key ecdsa.PrivateKey
	key.D = big.NewInt(i) // only for testing
	key.PublicKey.X, key.PublicKey.Y = ethCrypto.S256().ScalarBaseMult(key.D.Bytes())
	key.Curve = ethCrypto.S256()
	addr := ethCrypto.PubkeyToAddress(key.PublicKey)

	return &keys{
		BJJSecretKey: &sk,
		BJJPublicKey: sk.Public(),
		Addr:         addr,
	}
}
