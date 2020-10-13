package eth

import (
	"encoding/hex"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var rollupClient *RollupClient
var auctionClient *AuctionClient

var ethHashForge ethCommon.Hash
var argsForge *RollupForgeBatchArgs

var absoluteMaxL1L2BatchTimeout = int64(240)
var maxTx = int64(512)
var nLevels = int64(32)

func TestRollupConstants(t *testing.T) {
	rollupConstants, err := rollupClient.RollupConstants()
	require.Nil(t, err)
	assert.Equal(t, absoluteMaxL1L2BatchTimeout, rollupConstants.AbsoluteMaxL1L2BatchTimeout)
	assert.Equal(t, auctionAddressConst, rollupConstants.HermezAuctionContract)
	assert.Equal(t, tokenERC777AddressConst, rollupConstants.TokenHEZ)
	assert.Equal(t, maxTx, rollupConstants.Verifiers[0].MaxTx)
	assert.Equal(t, nLevels, rollupConstants.Verifiers[0].NLevels)
	assert.Equal(t, governanceAddressConst, rollupConstants.HermezGovernanceDAOAddress)
	assert.Equal(t, safetyAddressConst, rollupConstants.SafetyAddress)
	assert.Equal(t, wdelayerAddressConst, rollupConstants.WithdrawDelayerContract)
}

func TestAddToken(t *testing.T) {
	feeAddToken := big.NewInt(10)
	// Addtoken
	_, err := rollupClient.RollupAddToken(tokenERC777AddressConst, feeAddToken)
	require.Nil(t, err)
}

func TestRollupForgeBatch(t *testing.T) {
	// Register Coordinator
	forgerAddress := governanceAddressConst
	_, err := auctionClient.AuctionSetCoordinator(forgerAddress, URL)
	require.Nil(t, err)

	// MultiBid
	currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	slotSet := [6]bool{true, false, true, false, true, false}
	maxBid := new(big.Int)
	maxBid.SetString("15000000000000000000", 10)
	minBid := new(big.Int)
	minBid.SetString("11000000000000000000", 10)
	budget := new(big.Int)
	budget.SetString("45200000000000000000", 10)
	_, err = auctionClient.AuctionMultiBid(currentSlot+4, currentSlot+10, slotSet, maxBid, minBid, budget)
	require.Nil(t, err)

	// Add Blocks
	blockNum := int64(int(BLOCKSPERSLOT)*int(currentSlot+4) + genesisBlock)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	blocksToAdd := blockNum - currentBlockNum
	addBlocks(blocksToAdd, ethClientDialURL)

	// Forge
	args := new(RollupForgeBatchArgs)
	feeIdxCoordinatorBytes, err := hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.Nil(t, err)
	lenFeeIdxCoordinatorBytes := int(4)
	numFeeIdxCoordinator := len(feeIdxCoordinatorBytes) / lenFeeIdxCoordinatorBytes
	for i := 0; i < numFeeIdxCoordinator; i++ {
		var paddedFeeIdx [6]byte
		if lenFeeIdxCoordinatorBytes < common.IdxBytesLen {
			copy(paddedFeeIdx[6-lenFeeIdxCoordinatorBytes:], feeIdxCoordinatorBytes[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		} else {
			copy(paddedFeeIdx[:], feeIdxCoordinatorBytes[i*lenFeeIdxCoordinatorBytes:(i+1)*lenFeeIdxCoordinatorBytes])
		}
		FeeIdxCoordinator, err := common.IdxFromBytes(paddedFeeIdx[:])
		require.Nil(t, err)
		args.FeeIdxCoordinator = append(args.FeeIdxCoordinator, FeeIdxCoordinator)
	}
	l1CoordinatorBytes, err := hex.DecodeString("1c660323607bb113e586183609964a333d07ebe4bef3be82ec13af453bae9590bd7711cdb6abf42f176eadfbe5506fbef5e092e5543733f91b0061d9a7747fa10694a915a6470fa230de387b51e6f4db0b09787867778687b55197ad6d6a86eac000000001")
	require.Nil(t, err)
	numTxsL1 := len(l1CoordinatorBytes) / common.L1CoordinatorTxBytesLen
	for i := 0; i < numTxsL1; i++ {
		bytesL1Coordinator := l1CoordinatorBytes[i*common.L1CoordinatorTxBytesLen : (i+1)*common.L1CoordinatorTxBytesLen]
		var signature []byte
		v := bytesL1Coordinator[0]
		s := bytesL1Coordinator[1:33]
		r := bytesL1Coordinator[33:65]
		signature = append(signature, r[:]...)
		signature = append(signature, s[:]...)
		signature = append(signature, v)
		L1Tx, err := common.L1TxFromCoordinatorBytes(bytesL1Coordinator)
		require.Nil(t, err)
		args.L1CoordinatorTxs = append(args.L1CoordinatorTxs, *L1Tx)
		args.L1CoordinatorTxsAuths = append(args.L1CoordinatorTxsAuths, signature)
	}
	newStateRoot := new(big.Int)
	newStateRoot.SetString("18317824016047294649053625209337295956588174734569560016974612130063629505228", 10)
	newExitRoot := big.NewInt(0)
	args.NewLastIdx = int64(256)
	args.NewStRoot = newStateRoot
	args.NewExitRoot = newExitRoot
	args.L1Batch = true
	args.VerifierIdx = 0
	args.ProofA[0] = big.NewInt(0)
	args.ProofA[1] = big.NewInt(0)
	args.ProofB[0][0] = big.NewInt(0)
	args.ProofB[0][1] = big.NewInt(0)
	args.ProofB[1][0] = big.NewInt(0)
	args.ProofB[1][1] = big.NewInt(0)
	args.ProofC[0] = big.NewInt(0)
	args.ProofC[1] = big.NewInt(0)

	argsForge = args
	_, err = rollupClient.RollupForgeBatch(argsForge)
	require.Nil(t, err)

	currentBlockNum, _ = rollupClient.client.EthCurrentBlock()
	rollupEvents, _, _ := rollupClient.RollupEventsByBlock(currentBlockNum)

	assert.Equal(t, int64(1), rollupEvents.ForgeBatch[0].BatchNum)
	ethHashForge = rollupEvents.ForgeBatch[0].EthTxHash
}

func TestRollupForgeBatchArgs(t *testing.T) {
	args, err := rollupClient.RollupForgeBatchArgs(ethHashForge)
	require.Nil(t, err)
	assert.Equal(t, argsForge.FeeIdxCoordinator, args.FeeIdxCoordinator)
	assert.Equal(t, argsForge.L1Batch, args.L1Batch)
	assert.Equal(t, argsForge.L1CoordinatorTxs, args.L1CoordinatorTxs)
	assert.Equal(t, argsForge.L1CoordinatorTxsAuths, args.L1CoordinatorTxsAuths)
	assert.Equal(t, argsForge.L2TxsData, args.L2TxsData)
	assert.Equal(t, argsForge.NewLastIdx, args.NewLastIdx)
	assert.Equal(t, argsForge.NewStRoot, args.NewStRoot)
	assert.Equal(t, argsForge.VerifierIdx, args.VerifierIdx)
}
