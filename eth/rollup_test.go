package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
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

var tokenIDERC777 uint32
var tokenHEZID uint32

type keys struct {
	BJJSecretKey *babyjub.PrivateKey
	BJJPublicKey *babyjub.PublicKey
	Addr         ethCommon.Address
}

func genKeysBjj(i int64) *keys {
	i++ // i = 0 doesn't work for the ecdsa key generation
	var sk babyjub.PrivateKey
	binary.LittleEndian.PutUint64(sk[:], uint64(i))

	// eth address
	var key ecdsa.PrivateKey
	key.D = big.NewInt(i) // only for testing
	key.PublicKey.X, key.PublicKey.Y = ethCrypto.S256().ScalarBaseMult(key.D.Bytes())
	key.Curve = ethCrypto.S256()

	return &keys{
		BJJSecretKey: &sk,
		BJJPublicKey: sk.Public(),
	}
}

func TestRollupConstants(t *testing.T) {
	rollupConstants, err := rollupClient.RollupConstants()
	require.Nil(t, err)
	assert.Equal(t, absoluteMaxL1L2BatchTimeout, rollupConstants.AbsoluteMaxL1L2BatchTimeout)
	assert.Equal(t, auctionAddressConst, rollupConstants.HermezAuctionContract)
	assert.Equal(t, tokenHEZAddressConst, rollupConstants.TokenHEZ)
	assert.Equal(t, maxTx, rollupConstants.Verifiers[0].MaxTx)
	assert.Equal(t, nLevels, rollupConstants.Verifiers[0].NLevels)
	assert.Equal(t, governanceAddressConst, rollupConstants.HermezGovernanceDAOAddress)
	assert.Equal(t, safetyAddressConst, rollupConstants.SafetyAddress)
	assert.Equal(t, wdelayerAddressConst, rollupConstants.WithdrawDelayerContract)
}

func TestRollupRegisterTokensCount(t *testing.T) {
	registerTokensCount, err := rollupClient.RollupRegisterTokensCount()
	require.Nil(t, err)
	assert.Equal(t, big.NewInt(1), registerTokensCount)
}

func TestRollupAddToken(t *testing.T) {
	feeAddToken := big.NewInt(10)
	// Addtoken ERC20Permit
	registerTokensCount, err := rollupClient.RollupRegisterTokensCount()
	require.Nil(t, err)
	_, err = rollupClient.RollupAddToken(tokenHEZAddressConst, feeAddToken, deadline)
	require.Nil(t, err)
	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, tokenHEZAddressConst, rollupEvents.AddToken[0].TokenAddress)
	assert.Equal(t, registerTokensCount, common.TokenID(rollupEvents.AddToken[0].TokenID).BigInt())
	tokenHEZID = rollupEvents.AddToken[0].TokenID
}

func TestRollupForgeBatch(t *testing.T) {
	chainid, _ := auctionClient.client.Client().ChainID(context.Background())
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
	_, err = auctionClient.AuctionMultiBid(budget, currentSlot+4, currentSlot+10, slotSet, maxBid, minBid, deadline)
	require.Nil(t, err)

	// Add Blocks
	blockNum := int64(int(blocksPerSlot)*int(currentSlot+4) + int(genesisBlock))
	currentBlockNum, err := auctionClient.client.EthLastBlock()
	require.Nil(t, err)
	blocksToAdd := blockNum - currentBlockNum
	addBlocks(blocksToAdd, ethClientDialURL)

	// Forge
	args := new(RollupForgeBatchArgs)
	args.FeeIdxCoordinator = []common.Idx{} // When encoded, 64 times the 0 idx means that no idx to collect fees is specified.
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
		l1Tx, err := common.L1CoordinatorTxFromBytes(bytesL1Coordinator, chainid, rollupClient.address)
		require.Nil(t, err)
		args.L1CoordinatorTxs = append(args.L1CoordinatorTxs, *l1Tx)
		args.L1CoordinatorTxsAuths = append(args.L1CoordinatorTxsAuths, signature)
	}
	args.L2TxsData = []common.L2Tx{}
	newStateRoot := new(big.Int)
	newStateRoot.SetString("18317824016047294649053625209337295956588174734569560016974612130063629505228", 10)
	newExitRoot := new(big.Int)
	bytesNumExitRoot, err := hex.DecodeString("10a89d5fe8d488eda1ba371d633515739933c706c210c604f5bd209180daa43b")
	require.Nil(t, err)
	newExitRoot.SetBytes(bytesNumExitRoot)
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

	currentBlockNum, err = rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, int64(1), rollupEvents.ForgeBatch[0].BatchNum)
	ethHashForge = rollupEvents.ForgeBatch[0].EthTxHash
}

func TestRollupForgeBatchArgs(t *testing.T) {
	args, sender, err := rollupClient.RollupForgeBatchArgs(ethHashForge)
	require.Nil(t, err)
	assert.Equal(t, *sender, rollupClient.client.account.Address)
	assert.Equal(t, argsForge.FeeIdxCoordinator, args.FeeIdxCoordinator)
	assert.Equal(t, argsForge.L1Batch, args.L1Batch)
	assert.Equal(t, argsForge.L1CoordinatorTxs, args.L1CoordinatorTxs)
	assert.Equal(t, argsForge.L1CoordinatorTxsAuths, args.L1CoordinatorTxsAuths)
	assert.Equal(t, argsForge.L2TxsData, args.L2TxsData)
	assert.Equal(t, argsForge.NewLastIdx, args.NewLastIdx)
	assert.Equal(t, argsForge.NewStRoot, args.NewStRoot)
	assert.Equal(t, argsForge.VerifierIdx, args.VerifierIdx)
}

func TestRollupUpdateForgeL1L2BatchTimeout(t *testing.T) {
	newForgeL1L2BatchTimeout := int64(222)
	_, err := rollupClient.RollupUpdateForgeL1L2BatchTimeout(newForgeL1L2BatchTimeout)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, newForgeL1L2BatchTimeout, rollupEvents.UpdateForgeL1L2BatchTimeout[0].NewForgeL1L2BatchTimeout)
}

func TestRollupUpdateFeeAddToken(t *testing.T) {
	newFeeAddToken := big.NewInt(12)
	_, err := rollupClient.RollupUpdateFeeAddToken(newFeeAddToken)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, newFeeAddToken, rollupEvents.UpdateFeeAddToken[0].NewFeeAddToken)
}

func TestRollupL1UserTxETHCreateAccountDeposit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20CreateAccountDeposit(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitCreateAccountDeposit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxETHDeposit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20Deposit(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitDeposit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxETHDepositTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20DepositTransfer(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitDepositTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxETHCreateAccountDepositTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20CreateAccountDepositTransfer(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitCreateAccountDepositTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxETHForceTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20ForceTransfer(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitForceTransfer(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxETHForceExit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(2)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	tokenIDUint32 := uint32(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDUint32),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDUint32, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20ForceExit(t *testing.T) {
	rollupClientAux2, err := NewRollupClient(ethereumClientAux2, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(1)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenHEZID),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux2.RollupL1UserTxERC20ETH(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenHEZID, toIdxInt64)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux2.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupL1UserTxERC20PermitForceExit(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)
	key := genKeysBjj(3)
	fromIdxInt64 := int64(0)
	toIdxInt64 := int64(0)
	fromIdx := new(common.Idx)
	*fromIdx = 0
	l1Tx := common.L1Tx{
		FromBJJ:    key.BJJPublicKey,
		FromIdx:    common.Idx(fromIdxInt64),
		ToIdx:      common.Idx(toIdxInt64),
		LoadAmount: big.NewInt(10),
		TokenID:    common.TokenID(tokenIDERC777),
		Amount:     big.NewInt(0),
	}

	_, err = rollupClientAux.RollupL1UserTxERC20Permit(l1Tx.FromBJJ, fromIdxInt64, l1Tx.LoadAmount, l1Tx.Amount, tokenIDERC777, toIdxInt64, deadline)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)
	assert.Equal(t, l1Tx.FromBJJ, rollupEvents.L1UserTx[0].L1UserTx.FromBJJ)
	assert.Equal(t, l1Tx.ToIdx, rollupEvents.L1UserTx[0].L1UserTx.ToIdx)
	assert.Equal(t, l1Tx.LoadAmount, rollupEvents.L1UserTx[0].L1UserTx.LoadAmount)
	assert.Equal(t, l1Tx.TokenID, rollupEvents.L1UserTx[0].L1UserTx.TokenID)
	assert.Equal(t, l1Tx.Amount, rollupEvents.L1UserTx[0].L1UserTx.Amount)
	assert.Equal(t, rollupClientAux.client.account.Address, rollupEvents.L1UserTx[0].L1UserTx.FromEthAddr)
}

func TestRollupForgeBatch2(t *testing.T) {
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
	newStateRoot := new(big.Int)
	newStateRoot.SetString("0", 10)
	newExitRoot := new(big.Int)
	newExitRoot.SetString("4694629460381124336935185586347620040847956843554725549791403956105308092690", 10)
	args.NewLastIdx = int64(1000)
	args.NewStRoot = newStateRoot
	args.NewExitRoot = newExitRoot
	args.L1Batch = true
	args.VerifierIdx = 0
	args.ProofA[0] = big.NewInt(0)
	args.ProofA[1] = big.NewInt(0)
	args.ProofB[0] = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	args.ProofB[1] = [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	args.ProofC[0] = big.NewInt(0)
	args.ProofC[1] = big.NewInt(0)

	argsForge = args
	_, err = rollupClient.RollupForgeBatch(argsForge)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, int64(2), rollupEvents.ForgeBatch[0].BatchNum)
	ethHashForge = rollupEvents.ForgeBatch[0].EthTxHash
}

func TestRollupWithdrawMerkleProof(t *testing.T) {
	rollupClientAux, err := NewRollupClient(ethereumClientAux, hermezRollupAddressConst, tokenHEZ)
	require.Nil(t, err)

	var pkComp babyjub.PublicKeyComp
	pkCompBE, err := hex.DecodeString("adc3b754f8da621967b073a787bef8eec7052f2ba712b23af57d98f65beea8b2")
	require.Nil(t, err)
	pkCompLE := common.SwapEndianness(pkCompBE)
	copy(pkComp[:], pkCompLE)
	// err = pkComp.UnmarshalText([]byte(hex.EncodeToString(pkCompL)))
	// require.Nil(t, err)

	pk, err := pkComp.Decompress()
	require.Nil(t, err)

	require.Nil(t, err)
	tokenID := uint32(1)
	numExitRoot := int64(2)
	fromIdx := int64(256)
	amount := big.NewInt(10)
	// siblingBytes0, err := new(big.Int).SetString("19508838618377323910556678335932426220272947530531646682154552299216398748115", 10)
	// require.Nil(t, err)
	// siblingBytes1, err := new(big.Int).SetString("15198806719713909654457742294233381653226080862567104272457668857208564789571", 10)
	// require.Nil(t, err)
	var siblings []*big.Int
	// siblings = append(siblings, siblingBytes0)
	// siblings = append(siblings, siblingBytes1)
	instantWithdraw := true

	_, err = rollupClientAux.RollupWithdrawMerkleProof(pk, tokenID, numExitRoot, fromIdx, amount, siblings, instantWithdraw)
	require.Nil(t, err)

	currentBlockNum, err := rollupClient.client.EthLastBlock()
	require.Nil(t, err)
	rollupEvents, _, err := rollupClient.RollupEventsByBlock(currentBlockNum)
	require.Nil(t, err)

	assert.Equal(t, uint64(fromIdx), rollupEvents.Withdraw[0].Idx)
	assert.Equal(t, instantWithdraw, rollupEvents.Withdraw[0].InstantWithdraw)
	assert.Equal(t, uint64(numExitRoot), rollupEvents.Withdraw[0].NumExitRoot)
}
