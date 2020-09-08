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

var clientSetup *ClientSetup

func init() {
	rollupConstants := &eth.RollupConstants{}
	rollupVariables := &eth.RollupVariables{
		MaxTxVerifiers:     make([]int, 0),
		TokenHEZ:           ethCommon.Address{},
		GovernanceAddress:  ethCommon.Address{},
		SafetyBot:          ethCommon.Address{},
		ConsensusContract:  ethCommon.Address{},
		WithdrawalContract: ethCommon.Address{},
		FeeAddToken:        big.NewInt(1),
		ForgeL1Timeout:     16,
		FeeL1UserTx:        big.NewInt(2),
	}
	auctionConstants := &eth.AuctionConstants{}
	auctionVariables := &eth.AuctionVariables{
		DonationAddress: ethCommon.Address{},
		BootCoordinator: ethCommon.Address{},
		MinBidEpoch: [6]*big.Int{
			big.NewInt(10), big.NewInt(11), big.NewInt(12),
			big.NewInt(13), big.NewInt(14), big.NewInt(15)},
		ClosedAuctionSlots: 0,
		OpenAuctionSlots:   0,
		AllocationRatio:    [3]uint8{},
		Outbidding:         0,
		SlotDeadline:       0,
	}
	clientSetup = &ClientSetup{
		RollupConstants:  rollupConstants,
		RollupVariables:  rollupVariables,
		AuctionConstants: auctionConstants,
		AuctionVariables: auctionVariables,
	}
}

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
	client := NewClient(true, &timer, clientSetup)
	c = client
	require.NotNil(t, c)
}

func TestEthClient(t *testing.T) {
	token1Addr := ethCommon.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")

	var timer timer
	c := NewClient(true, &timer, clientSetup)
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
			FromIdx:     common.Idx(0),
			FromEthAddr: keys[i].Addr,
			FromBJJ:     keys[i].BJJPublicKey,
			TokenID:     common.TokenID(0),
			LoadAmount:  big.NewInt(10 + int64(i)),
		}
		c.CtlAddL1TxUser(&l1UserTx)
	}
	c.CtlMineBlock()

	blockNum, err = c.EthCurrentBlock()
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
		L2Txs:             []*common.L2Tx{},
		FeeIdxCoordinator: make([]common.Idx, eth.FeeIdxCoordinatorLen),
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
		L2Txs:             []*common.L2Tx{},
		FeeIdxCoordinator: make([]common.Idx, eth.FeeIdxCoordinatorLen),
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
