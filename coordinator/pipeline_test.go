package coordinator

import (
	"context"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/coordinator/prover"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/etherscan"
	"github.com/hermeznetwork/hermez-node/synchronizer"
	"github.com/hermeznetwork/hermez-node/test"
	"github.com/iden3/go-merkletree-sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineShouldL1L2Batch(t *testing.T) {
	ethClientSetup := test.NewClientSetupExample()
	ethClientSetup.ChainID = big.NewInt(int64(chainID))

	var timer timer
	ctx := context.Background()
	ethClient := test.NewClient(true, &timer, &bidder, ethClientSetup)
	etherScanService, _ := etherscan.NewEtherscanService("", "")
	modules := newTestModules(t)
	var stats synchronizer.Stats
	coord := newTestCoordinator(t, forger, ethClient, ethClientSetup, modules, etherScanService)
	pipeline, err := coord.newPipeline(ctx)
	require.NoError(t, err)
	pipeline.vars = coord.vars

	// Check that the parameters are the ones we expect and use in this test
	require.Equal(t, 0.5, pipeline.cfg.L1BatchTimeoutPerc)
	require.Equal(t, int64(10), ethClientSetup.RollupVariables.ForgeL1L2BatchTimeout)
	l1BatchTimeoutPerc := pipeline.cfg.L1BatchTimeoutPerc
	l1BatchTimeout := ethClientSetup.RollupVariables.ForgeL1L2BatchTimeout

	startBlock := int64(100)
	// Empty batchInfo to pass to shouldL1L2Batch() which sets debug information
	batchInfo := BatchInfo{}

	//
	// No scheduled L1Batch
	//

	// Last L1Batch was a long time ago
	stats.Eth.LastBlock.Num = startBlock
	stats.Sync.LastBlock = stats.Eth.LastBlock
	stats.Sync.LastL1BatchBlock = 0
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch(&batchInfo))

	stats.Sync.LastL1BatchBlock = startBlock

	// We are are one block before the timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock - 1 + int64(float64(l1BatchTimeout-1)*l1BatchTimeoutPerc) - 1
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, false, pipeline.shouldL1L2Batch(&batchInfo))

	// We are are at timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock - 1 + int64(float64(l1BatchTimeout-1)*l1BatchTimeoutPerc)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch(&batchInfo))

	//
	// Scheduled L1Batch
	//
	pipeline.state.lastScheduledL1BatchBlockNum = startBlock
	stats.Sync.LastL1BatchBlock = startBlock - 10

	// We are are one block before the timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock - 1 + int64(float64(l1BatchTimeout-1)*l1BatchTimeoutPerc) - 1
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, false, pipeline.shouldL1L2Batch(&batchInfo))

	// We are are at timeout range * 0.5
	stats.Eth.LastBlock.Num = startBlock - 1 + int64(float64(l1BatchTimeout-1)*l1BatchTimeoutPerc)
	stats.Sync.LastBlock = stats.Eth.LastBlock
	pipeline.stats = stats
	assert.Equal(t, true, pipeline.shouldL1L2Batch(&batchInfo))

	closeTestModules(t, modules)
}

func TestEthRollupForgeBatch(t *testing.T) {
	if os.Getenv("TEST_ROLLUP_FORGE_BATCH") == "" {
		return
	}
	const web3URL = "http://localhost:8545"
	const password = "test"
	addr := ethCommon.HexToAddress("0xb4124ceb3451635dacedd11767f004d8a28c6ee7")
	sk, err := crypto.HexToECDSA(
		"a8a54b2d8197bc0b19bb8a084031be71835580a01e70a45a13babd16c9bc1563")
	require.NoError(t, err)
	rollupAddr := ethCommon.HexToAddress("0x8EEaea23686c319133a7cC110b840d1591d9AeE0")
	pathKeystore, err := ioutil.TempDir("", "tmpKeystore")
	require.NoError(t, err)
	deleteme = append(deleteme, pathKeystore)
	ctx := context.Background()
	batchInfo := &BatchInfo{}
	proofClient := &prover.MockClient{}
	chainID := uint16(0)

	ethClient, err := ethclient.Dial(web3URL)
	require.NoError(t, err)
	ethCfg := eth.EthereumConfig{
		CallGasLimit: 300000,
		GasPriceDiv:  100,
	}
	scryptN := ethKeystore.LightScryptN
	scryptP := ethKeystore.LightScryptP
	keyStore := ethKeystore.NewKeyStore(pathKeystore,
		scryptN, scryptP)
	account, err := keyStore.ImportECDSA(sk, password)
	require.NoError(t, err)
	require.Equal(t, account.Address, addr)
	err = keyStore.Unlock(account, password)
	require.NoError(t, err)

	client, err := eth.NewClient(ethClient, &account, keyStore, &eth.ClientConfig{
		Ethereum: ethCfg,
		Rollup: eth.RollupConfig{
			Address: rollupAddr,
		},
	})
	require.NoError(t, err)

	zkInputs := common.NewZKInputs(chainID, 100, 24, 512, 32, big.NewInt(1))
	zkInputs.Metadata.NewStateRootRaw = &merkletree.Hash{1}
	zkInputs.Metadata.NewExitRootRaw = &merkletree.Hash{2}
	batchInfo.ZKInputs = zkInputs
	err = proofClient.CalculateProof(ctx, batchInfo.ZKInputs)
	require.NoError(t, err)

	proof, pubInputs, err := proofClient.GetProof(ctx)
	require.NoError(t, err)
	batchInfo.Proof = proof
	batchInfo.PublicInputs = pubInputs

	batchInfo.ForgeBatchArgs = prepareForgeBatchArgs(batchInfo)
	auth, err := client.NewAuth()
	require.NoError(t, err)
	auth.Context = context.Background()
	_, err = client.RollupForgeBatch(batchInfo.ForgeBatchArgs, auth)
	require.NoError(t, err)
	batchInfo.Proof = proof
}
