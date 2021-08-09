package eth

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const slotDeadlineConst = uint8(20)
const openAuctionSlotsConst = uint16(4320)
const closedAuctionSlotsConst = uint16(2)
const outbiddingConst = uint16(1000)
const currentSlotConst = 0
const blocksPerSlot = uint8(40)
const minBidStr = "10000000000000000000"
const URL = "http://localhost:3000"

var allocationRatioConst [3]uint16 = [3]uint16{4000, 4000, 2000}
var auctionClientTest *AuctionEthClient

func TestAuctionGetCurrentSlotNumber(t *testing.T) {
	currentSlot, err := auctionClientTest.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	currentSlotInt := int(currentSlot)
	assert.Equal(t, currentSlotConst, currentSlotInt)
}

func TestAuctionEventInit(t *testing.T) {
	auctionInit, blockNum, err := auctionClientTest.AuctionEventInit(genesisBlock)
	require.NoError(t, err)
	assert.Equal(t, int64(18), blockNum)
	assert.Equal(t, donationAddressConst, auctionInit.DonationAddress)
	assert.Equal(t, bootCoordinatorAddressConst, auctionInit.BootCoordinatorAddress)
	assert.Equal(t, "https://boot.coordinator.io", auctionInit.BootCoordinatorURL)
	assert.Equal(t, uint16(1000), auctionInit.Outbidding)
	assert.Equal(t, uint8(20), auctionInit.SlotDeadline)
	assert.Equal(t, uint16(2), auctionInit.ClosedAuctionSlots)
	assert.Equal(t, uint16(4320), auctionInit.OpenAuctionSlots)
	assert.Equal(t, [3]uint16{4000, 4000, 2000}, auctionInit.AllocationRatio)
}

func TestAuctionConstants(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)

	auctionConstants, err := auctionClientTest.AuctionConstants()
	require.Nil(t, err)
	assert.Equal(t, auctionConstants.BlocksPerSlot, blocksPerSlot)
	assert.Equal(t, auctionConstants.GenesisBlockNum, genesisBlock)
	assert.Equal(t, auctionConstants.HermezRollup, hermezRollupTestAddressConst)
	assert.Equal(t, auctionConstants.InitialMinimalBidding, INITMINBID)
	assert.Equal(t, auctionConstants.TokenHEZ, tokenHEZAddressConst)
	assert.Equal(t, auctionConstants.GovernanceAddress, governanceAddressConst)
}

func TestAuctionVariables(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)
	defaultSlotSetBid := [6]*big.Int{INITMINBID, INITMINBID, INITMINBID, INITMINBID, INITMINBID,
		INITMINBID}

	auctionVariables, err := auctionClientTest.AuctionVariables()
	require.Nil(t, err)
	assert.Equal(t, auctionVariables.AllocationRatio, allocationRatioConst)
	assert.Equal(t, auctionVariables.BootCoordinator, bootCoordinatorAddressConst)
	assert.Equal(t, auctionVariables.ClosedAuctionSlots, closedAuctionSlotsConst)
	assert.Equal(t, auctionVariables.DefaultSlotSetBid, defaultSlotSetBid)
	assert.Equal(t, auctionVariables.DonationAddress, donationAddressConst)
	assert.Equal(t, auctionVariables.OpenAuctionSlots, openAuctionSlotsConst)
	assert.Equal(t, auctionVariables.Outbidding, outbiddingConst)
	assert.Equal(t, auctionVariables.SlotDeadline, slotDeadlineConst)
}

func TestAuctionGetSlotDeadline(t *testing.T) {
	slotDeadline, err := auctionClientTest.AuctionGetSlotDeadline()
	require.Nil(t, err)
	assert.Equal(t, slotDeadlineConst, slotDeadline)
}

func TestAuctionSetSlotDeadline(t *testing.T) {
	newSlotDeadline := uint8(25)

	_, err := auctionClientTest.AuctionSetSlotDeadline(newSlotDeadline)
	require.Nil(t, err)
	slotDeadline, err := auctionClientTest.AuctionGetSlotDeadline()
	require.Nil(t, err)
	assert.Equal(t, newSlotDeadline, slotDeadline)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newSlotDeadline, auctionEvents.NewSlotDeadline[0].NewSlotDeadline)
}

func TestAuctionGetOpenAuctionSlots(t *testing.T) {
	openAuctionSlots, err := auctionClientTest.AuctionGetOpenAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, openAuctionSlotsConst, openAuctionSlots)
}

func TestAuctionSetOpenAuctionSlots(t *testing.T) {
	newOpenAuctionSlots := uint16(4500)

	_, err := auctionClientTest.AuctionSetOpenAuctionSlots(newOpenAuctionSlots)
	require.Nil(t, err)
	openAuctionSlots, err := auctionClientTest.AuctionGetOpenAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, newOpenAuctionSlots, openAuctionSlots)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newOpenAuctionSlots, auctionEvents.NewOpenAuctionSlots[0].NewOpenAuctionSlots)
}

func TestAuctionGetClosedAuctionSlots(t *testing.T) {
	closedAuctionSlots, err := auctionClientTest.AuctionGetClosedAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, closedAuctionSlotsConst, closedAuctionSlots)
}

func TestAuctionSetClosedAuctionSlots(t *testing.T) {
	newClosedAuctionSlots := uint16(1)

	_, err := auctionClientTest.AuctionSetClosedAuctionSlots(newClosedAuctionSlots)
	require.Nil(t, err)
	closedAuctionSlots, err := auctionClientTest.AuctionGetClosedAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, newClosedAuctionSlots, closedAuctionSlots)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newClosedAuctionSlots,
		auctionEvents.NewClosedAuctionSlots[0].NewClosedAuctionSlots)
	_, err = auctionClientTest.AuctionSetClosedAuctionSlots(closedAuctionSlots)
	require.Nil(t, err)
}

func TestAuctionGetOutbidding(t *testing.T) {
	outbidding, err := auctionClientTest.AuctionGetOutbidding()
	require.Nil(t, err)
	assert.Equal(t, outbiddingConst, outbidding)
}

func TestAuctionSetOutbidding(t *testing.T) {
	newOutbidding := uint16(0xb)

	_, err := auctionClientTest.AuctionSetOutbidding(newOutbidding)
	require.Nil(t, err)
	outbidding, err := auctionClientTest.AuctionGetOutbidding()
	require.Nil(t, err)
	assert.Equal(t, newOutbidding, outbidding)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newOutbidding, auctionEvents.NewOutbidding[0].NewOutbidding)
	_, err = auctionClientTest.AuctionSetOutbidding(outbiddingConst)
	require.Nil(t, err)
}

func TestAuctionGetAllocationRatio(t *testing.T) {
	allocationRatio, err := auctionClientTest.AuctionGetAllocationRatio()
	require.Nil(t, err)
	assert.Equal(t, allocationRatioConst, allocationRatio)
}

func TestAuctionSetAllocationRatio(t *testing.T) {
	newAllocationRatio := [3]uint16{3000, 3000, 4000}

	_, err := auctionClientTest.AuctionSetAllocationRatio(newAllocationRatio)
	require.Nil(t, err)
	allocationRatio, err := auctionClientTest.AuctionGetAllocationRatio()
	require.Nil(t, err)
	assert.Equal(t, newAllocationRatio, allocationRatio)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newAllocationRatio, auctionEvents.NewAllocationRatio[0].NewAllocationRatio)
	_, err = auctionClientTest.AuctionSetAllocationRatio(allocationRatioConst)
	require.Nil(t, err)
}

func TestAuctionGetDonationAddress(t *testing.T) {
	donationAddress, err := auctionClientTest.AuctionGetDonationAddress()
	require.Nil(t, err)
	assert.Equal(t, &donationAddressConst, donationAddress)
}

func TestAuctionGetBootCoordinator(t *testing.T) {
	bootCoordinator, err := auctionClientTest.AuctionGetBootCoordinator()
	require.Nil(t, err)
	assert.Equal(t, &bootCoordinatorAddressConst, bootCoordinator)
}

func TestAuctionSetDonationAddress(t *testing.T) {
	newDonationAddress := governanceAddressConst

	_, err := auctionClientTest.AuctionSetDonationAddress(newDonationAddress)
	require.Nil(t, err)
	donationAddress, err := auctionClientTest.AuctionGetDonationAddress()
	require.Nil(t, err)
	assert.Equal(t, &newDonationAddress, donationAddress)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newDonationAddress, auctionEvents.NewDonationAddress[0].NewDonationAddress)
	_, err = auctionClientTest.AuctionSetDonationAddress(donationAddressConst)
	require.Nil(t, err)
}

func TestAuctionSetBootCoordinator(t *testing.T) {
	newBootCoordinator := governanceAddressConst
	bootCoordinatorURL := "https://boot.coordinator2.io"
	newBootCoordinatorURL := "https://boot.coordinator2.io"

	_, err := auctionClientTest.AuctionSetBootCoordinator(newBootCoordinator, newBootCoordinatorURL)
	require.Nil(t, err)
	bootCoordinator, err := auctionClientTest.AuctionGetBootCoordinator()
	require.Nil(t, err)
	assert.Equal(t, &newBootCoordinator, bootCoordinator)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, newBootCoordinator, auctionEvents.NewBootCoordinator[0].NewBootCoordinator)
	assert.Equal(t, newBootCoordinatorURL, auctionEvents.NewBootCoordinator[0].NewBootCoordinatorURL)
	_, err = auctionClientTest.AuctionSetBootCoordinator(bootCoordinatorAddressConst,
		bootCoordinatorURL)
	require.Nil(t, err)
}

func TestAuctionGetSlotSet(t *testing.T) {
	slot := int64(10)

	slotSet, err := auctionClientTest.AuctionGetSlotSet(slot)
	require.Nil(t, err)
	assert.Equal(t, slotSet, big.NewInt(4))
}

func TestAuctionGetDefaultSlotSetBid(t *testing.T) {
	slotSet := uint8(3)

	minBid, err := auctionClientTest.AuctionGetDefaultSlotSetBid(slotSet)
	require.Nil(t, err)
	assert.Equal(t, minBid.String(), minBidStr)
}

func TestAuctionChangeDefaultSlotSetBid(t *testing.T) {
	slotSet := int64(3)
	set := uint8(3)
	newInitialMinBid := new(big.Int)
	newInitialMinBid.SetString("20000000000000000000", 10)

	_, err := auctionClientTest.AuctionChangeDefaultSlotSetBid(slotSet, newInitialMinBid)
	require.Nil(t, err)
	minBid, err := auctionClientTest.AuctionGetDefaultSlotSetBid(set)
	require.Nil(t, err)
	assert.Equal(t, minBid, newInitialMinBid)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, slotSet, auctionEvents.NewDefaultSlotSetBid[0].SlotSet)
	assert.Equal(t, newInitialMinBid, auctionEvents.NewDefaultSlotSetBid[0].NewInitialMinBid)
	newMinBid := new(big.Int)
	newMinBid.SetString("10000000000000000000", 10)
	_, err = auctionClientTest.AuctionChangeDefaultSlotSetBid(slotSet, newMinBid)
	require.Nil(t, err)
}

func TestAuctionGetClaimableHEZ(t *testing.T) {
	bidderAddress := governanceAddressConst

	claimableHEZ, err := auctionClientTest.AuctionGetClaimableHEZ(bidderAddress)
	require.Nil(t, err)
	assert.Equal(t, claimableHEZ.Int64(), int64(0))
}

func TestAuctionRegisterCoordinator(t *testing.T) {
	forgerAddress := governanceAddressConst
	bidderAddress := governanceAddressConst

	_, err := auctionClientTest.AuctionSetCoordinator(forgerAddress, URL)
	require.Nil(t, err)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, forgerAddress, auctionEvents.SetCoordinator[0].ForgerAddress)
	assert.Equal(t, bidderAddress, auctionEvents.SetCoordinator[0].BidderAddress)
	assert.Equal(t, URL, auctionEvents.SetCoordinator[0].CoordinatorURL)
}

func TestAuctionBid(t *testing.T) {
	currentSlot, err := auctionClientTest.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	bidAmount := new(big.Int)
	bidAmount.SetString("12000000000000000000", 10)
	amount := new(big.Int)
	amount.SetString("12000000000000000000", 10)
	bidderAddress := governanceAddressConst
	_, err = auctionClientTest.AuctionBid(amount, currentSlot+4, bidAmount, deadline)
	require.Nil(t, err)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, bidAmount, auctionEvents.NewBid[0].BidAmount)
	assert.Equal(t, bidderAddress, auctionEvents.NewBid[0].Bidder)
	assert.Equal(t, currentSlot+4, auctionEvents.NewBid[0].Slot)
}

func TestAuctionGetSlotNumber(t *testing.T) {
	slotConst := 4
	blockNum := int(blocksPerSlot)*slotConst + int(genesisBlock)

	slot, err := auctionClientTest.AuctionGetSlotNumber(int64(blockNum))
	require.Nil(t, err)
	assert.Equal(t, slot, int64(slotConst))
}

func TestAuctionCanForge(t *testing.T) {
	slotConst := 4
	blockNum := int(blocksPerSlot)*slotConst + int(genesisBlock)

	canForge, err := auctionClientTest.AuctionCanForge(governanceAddressConst, int64(blockNum))
	require.Nil(t, err)
	assert.Equal(t, canForge, true)
}

func TestAuctionMultiBid(t *testing.T) {
	currentSlot, err := auctionClientTest.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	slotSet := [6]bool{true, false, true, false, true, false}
	maxBid := new(big.Int)
	maxBid.SetString("15000000000000000000", 10)
	minBid := new(big.Int)
	minBid.SetString("11000000000000000000", 10)
	budget := new(big.Int)
	budget.SetString("45200000000000000000", 10)
	bidderAddress := governanceAddressConst
	_, err = auctionClientTest.AuctionMultiBid(budget, currentSlot+4, currentSlot+10, slotSet,
		maxBid, minBid, deadline)
	require.Nil(t, err)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, bidderAddress, auctionEvents.NewBid[0].Bidder)
	assert.Equal(t, currentSlot+4, auctionEvents.NewBid[0].Slot)
	assert.Equal(t, bidderAddress, auctionEvents.NewBid[1].Bidder)
	assert.Equal(t, currentSlot+6, auctionEvents.NewBid[1].Slot)
	assert.Equal(t, bidderAddress, auctionEvents.NewBid[2].Bidder)
	assert.Equal(t, currentSlot+8, auctionEvents.NewBid[2].Slot)
	assert.Equal(t, bidderAddress, auctionEvents.NewBid[3].Bidder)
	assert.Equal(t, currentSlot+10, auctionEvents.NewBid[3].Slot)
}

func TestAuctionGetClaimableHEZ2(t *testing.T) {
	bidderAddress := governanceAddressConst
	amount := new(big.Int)
	amount.SetString("11000000000000000000", 10)

	claimableHEZ, err := auctionClientTest.AuctionGetClaimableHEZ(bidderAddress)
	require.Nil(t, err)
	assert.Equal(t, claimableHEZ, amount)
}

func TestAuctionClaimHEZ(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("11000000000000000000", 10)

	_, err := auctionClientTest.AuctionClaimHEZ()
	require.Nil(t, err)
	currentBlockNum, err := auctionClientTest.client.EthLastBlock()
	require.Nil(t, err)
	auctionEvents, err := auctionClientTest.AuctionEventsByBlock(currentBlockNum, nil)
	require.Nil(t, err)
	assert.Equal(t, amount, auctionEvents.HEZClaimed[0].Amount)
	assert.Equal(t, governanceAddressConst, auctionEvents.HEZClaimed[0].Owner)
}

func TestAuctionForge(t *testing.T) {
	auctionClientTestHermez, err := NewAuctionClient(ethereumClientHermez,
		auctionTestAddressConst, tokenHEZ)
	require.Nil(t, err)
	slotConst := 4
	blockNum := int64(int(blocksPerSlot)*slotConst + int(genesisBlock))
	currentBlockNum, err := auctionClientTestHermez.client.EthLastBlock()
	require.Nil(t, err)
	blocksToAdd := blockNum - currentBlockNum
	addBlocks(blocksToAdd, ethClientDialURL)
	currentBlockNum, err = auctionClientTestHermez.client.EthLastBlock()
	require.Nil(t, err)
	assert.Equal(t, currentBlockNum, blockNum)
	_, err = auctionClientTestHermez.AuctionForge(governanceAddressConst)
	require.Nil(t, err)
}

func TestGetCoordinatorsLibP2PAddrs(t *testing.T) {
	auctionClient, err := NewAuctionClient(ethereumClientHermez,
		auctionTestAddressConst, tokenHEZ)
	require.NoError(t, err)
	_, err = auctionClient.GetCoordinatorsLibP2PAddrs()
	require.NoError(t, err)
}

func TestPubKeyFromTx(t *testing.T) {
	auctionClient, err := NewAuctionClient(ethereumClientHermez,
		auctionTestAddressConst, tokenHEZ)
	require.NoError(t, err)
	ctx := context.Background()
	block, err := auctionClient.client.client.BlockByNumber(ctx, nil)
	require.NoError(t, err)
	txs := block.Transactions()
	for i := 0; i < txs.Len(); i++ {
		tx, err := auctionClient.client.client.TransactionInBlock(ctx, block.Hash(), uint(i))
		require.NoError(t, err)
		pubKey, err := pubKeyFromTx(tx)
		require.NoError(t, err)
		from, err := types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
		require.NoError(t, err)
		require.Equal(t, from, ethCrypto.PubkeyToAddress(*pubKey))
	}
}
