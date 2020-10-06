package eth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const slotDeadlineConst = uint8(20)
const openAuctionSlotsConst = uint16(4320)
const closedAuctionSlotsConst = uint16(2)
const outbiddingConst = uint16(1000)
const currentSlotConst = 0

var allocationRatioConst [3]uint16 = [3]uint16{4000, 4000, 2000}

var auctionClient *AuctionClient

//var genesisBlock = 93
var genesisBlock = 100

var minBidStr = "10000000000000000000"
var URL = "http://localhost:3000"
var newURL = "http://localhost:3002"
var BLOCKSPERSLOT = uint8(40)

func TestAuctionGetCurrentSlotNumber(t *testing.T) {
	currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	currentSlotInt := int(currentSlot)
	assert.Equal(t, currentSlotConst, currentSlotInt)
}

func TestAuctionConstants(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)

	auctionConstants, err := auctionClient.AuctionConstants()
	require.Nil(t, err)
	assert.Equal(t, auctionConstants.BlocksPerSlot, BLOCKSPERSLOT)
	assert.Equal(t, auctionConstants.GenesisBlockNum, int64(genesisBlock))
	assert.Equal(t, auctionConstants.HermezRollup, hermezRollupAddressTestConst)
	assert.Equal(t, auctionConstants.InitialMinimalBidding, INITMINBID)
	assert.Equal(t, auctionConstants.TokenHEZ.Hex(), tokenHezAddressConst.Hex())
}

func TestAuctionVariables(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)
	defaultSlotSetBid := [6]*big.Int{INITMINBID, INITMINBID, INITMINBID, INITMINBID, INITMINBID, INITMINBID}

	auctionVariables, err := auctionClient.AuctionVariables()
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
	slotDeadline, err := auctionClient.AuctionGetSlotDeadline()
	require.Nil(t, err)
	assert.Equal(t, slotDeadlineConst, slotDeadline)
}

func TestAuctionSetSlotDeadline(t *testing.T) {
	newSlotDeadline := uint8(25)

	_, err := auctionClient.AuctionSetSlotDeadline(newSlotDeadline)
	require.Nil(t, err)
	slotDeadline, err := auctionClient.AuctionGetSlotDeadline()
	require.Nil(t, err)
	assert.Equal(t, newSlotDeadline, slotDeadline)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newSlotDeadline, auctionEvents.NewSlotDeadline[0].NewSlotDeadline)
}

func TestAuctionGetOpenAuctionSlots(t *testing.T) {
	openAuctionSlots, err := auctionClient.AuctionGetOpenAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, openAuctionSlotsConst, openAuctionSlots)
}

func TestAuctionSetOpenAuctionSlots(t *testing.T) {
	newOpenAuctionSlots := uint16(4500)

	_, err := auctionClient.AuctionSetOpenAuctionSlots(newOpenAuctionSlots)
	require.Nil(t, err)
	openAuctionSlots, err := auctionClient.AuctionGetOpenAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, newOpenAuctionSlots, openAuctionSlots)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newOpenAuctionSlots, auctionEvents.NewOpenAuctionSlots[0].NewOpenAuctionSlots)
}

func TestAuctionGetClosedAuctionSlots(t *testing.T) {
	closedAuctionSlots, err := auctionClient.AuctionGetClosedAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, closedAuctionSlotsConst, closedAuctionSlots)
}

func TestAuctionSetClosedAuctionSlots(t *testing.T) {
	newClosedAuctionSlots := uint16(1)

	_, err := auctionClient.AuctionSetClosedAuctionSlots(newClosedAuctionSlots)
	require.Nil(t, err)
	closedAuctionSlots, err := auctionClient.AuctionGetClosedAuctionSlots()
	require.Nil(t, err)
	assert.Equal(t, newClosedAuctionSlots, closedAuctionSlots)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newClosedAuctionSlots, auctionEvents.NewClosedAuctionSlots[0].NewClosedAuctionSlots)
	_, err = auctionClient.AuctionSetClosedAuctionSlots(closedAuctionSlots)
	require.Nil(t, err)
}

func TestAuctionGetOutbidding(t *testing.T) {
	outbidding, err := auctionClient.AuctionGetOutbidding()
	require.Nil(t, err)
	assert.Equal(t, outbiddingConst, outbidding)
}

func TestAuctionSetOutbidding(t *testing.T) {
	newOutbidding := uint16(0xb)

	_, err := auctionClient.AuctionSetOutbidding(newOutbidding)
	require.Nil(t, err)
	outbidding, err := auctionClient.AuctionGetOutbidding()
	require.Nil(t, err)
	assert.Equal(t, newOutbidding, outbidding)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newOutbidding, auctionEvents.NewOutbidding[0].NewOutbidding)
	_, err = auctionClient.AuctionSetOutbidding(outbiddingConst)
	require.Nil(t, err)
}

func TestAuctionGetAllocationRatio(t *testing.T) {
	allocationRatio, err := auctionClient.AuctionGetAllocationRatio()
	require.Nil(t, err)
	assert.Equal(t, allocationRatioConst, allocationRatio)
}

func TestAuctionSetAllocationRatio(t *testing.T) {
	newAllocationRatio := [3]uint16{3000, 3000, 4000}

	_, err := auctionClient.AuctionSetAllocationRatio(newAllocationRatio)
	require.Nil(t, err)
	allocationRatio, err := auctionClient.AuctionGetAllocationRatio()
	require.Nil(t, err)
	assert.Equal(t, newAllocationRatio, allocationRatio)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newAllocationRatio, auctionEvents.NewAllocationRatio[0].NewAllocationRatio)
	_, err = auctionClient.AuctionSetAllocationRatio(allocationRatioConst)
	require.Nil(t, err)
}

func TestAuctionGetDonationAddress(t *testing.T) {
	donationAddress, err := auctionClient.AuctionGetDonationAddress()
	require.Nil(t, err)
	assert.Equal(t, &donationAddressConst, donationAddress)
}

func TestAuctionGetBootCoordinator(t *testing.T) {
	bootCoordinator, err := auctionClient.AuctionGetBootCoordinator()
	require.Nil(t, err)
	assert.Equal(t, &bootCoordinatorAddressConst, bootCoordinator)
}

func TestAuctionSetDonationAddress(t *testing.T) {
	newDonationAddress := governanceAddressConst

	_, err := auctionClient.AuctionSetDonationAddress(newDonationAddress)
	require.Nil(t, err)
	donationAddress, err := auctionClient.AuctionGetDonationAddress()
	require.Nil(t, err)
	assert.Equal(t, &newDonationAddress, donationAddress)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newDonationAddress, auctionEvents.NewDonationAddress[0].NewDonationAddress)
	_, err = auctionClient.AuctionSetDonationAddress(donationAddressConst)
	require.Nil(t, err)
}

func TestAuctionSetBootCoordinator(t *testing.T) {
	newBootCoordinator := governanceAddressConst

	_, err := auctionClient.AuctionSetBootCoordinator(newBootCoordinator)
	require.Nil(t, err)
	bootCoordinator, err := auctionClient.AuctionGetBootCoordinator()
	require.Nil(t, err)
	assert.Equal(t, &newBootCoordinator, bootCoordinator)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, newBootCoordinator, auctionEvents.NewBootCoordinator[0].NewBootCoordinator)
	_, err = auctionClient.AuctionSetBootCoordinator(bootCoordinatorAddressConst)
	require.Nil(t, err)
}

func TestAuctionGetSlotSet(t *testing.T) {
	slot := int64(10)

	slotSet, err := auctionClient.AuctionGetSlotSet(slot)
	require.Nil(t, err)
	assert.Equal(t, slotSet, big.NewInt(4))
}

func TestAuctionGetDefaultSlotSetBid(t *testing.T) {
	slotSet := uint8(3)

	minBid, err := auctionClient.AuctionGetDefaultSlotSetBid(slotSet)
	require.Nil(t, err)
	assert.Equal(t, minBid.String(), minBidStr)
}

func TestAuctionChangeDefaultSlotSetBid(t *testing.T) {
	slotSet := int64(3)
	set := uint8(3)
	newInitialMinBid := new(big.Int)
	newInitialMinBid.SetString("20000000000000000000", 10)

	_, err := auctionClient.AuctionChangeDefaultSlotSetBid(slotSet, newInitialMinBid)
	require.Nil(t, err)
	minBid, err := auctionClient.AuctionGetDefaultSlotSetBid(set)
	require.Nil(t, err)
	assert.Equal(t, minBid, newInitialMinBid)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, slotSet, auctionEvents.NewDefaultSlotSetBid[0].SlotSet)
	assert.Equal(t, newInitialMinBid, auctionEvents.NewDefaultSlotSetBid[0].NewInitialMinBid)
	newMinBid := new(big.Int)
	newMinBid.SetString("10000000000000000000", 10)
	_, err = auctionClient.AuctionChangeDefaultSlotSetBid(slotSet, newMinBid)
	require.Nil(t, err)
}

func TestAuctionGetClaimableHEZ(t *testing.T) {
	forgerAddress := governanceAddressConst

	claimableHEZ, err := auctionClient.AuctionGetClaimableHEZ(forgerAddress)
	require.Nil(t, err)
	assert.Equal(t, claimableHEZ.Int64(), int64(0))
}

func TestAuctionRegisterCoordinator(t *testing.T) {
	forgerAddress := governanceAddressConst

	_, err := auctionClient.AuctionSetCoordinator(forgerAddress, URL)
	require.Nil(t, err)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, forgerAddress, auctionEvents.SetCoordinator[0].ForgerAddress)
	assert.Equal(t, forgerAddress, auctionEvents.SetCoordinator[0].BidderAddress)
	assert.Equal(t, URL, auctionEvents.SetCoordinator[0].CoordinatorURL)
}

func TestAuctionBid(t *testing.T) {
	currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	bidAmount := new(big.Int)
	bidAmount.SetString("12000000000000000000", 10)
	forgerAddress := governanceAddressConst
	_, err = auctionClient.AuctionBid(currentSlot+4, bidAmount)
	require.Nil(t, err)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, bidAmount, auctionEvents.NewBid[0].BidAmount)
	assert.Equal(t, forgerAddress, auctionEvents.NewBid[0].Bidder)
	assert.Equal(t, currentSlot+4, auctionEvents.NewBid[0].Slot)
}

func TestAuctionGetSlotNumber(t *testing.T) {
	slotConst := 4
	blockNum := int(BLOCKSPERSLOT)*slotConst + genesisBlock

	slot, err := auctionClient.AuctionGetSlotNumber(int64(blockNum))
	require.Nil(t, err)
	assert.Equal(t, slot, int64(slotConst))
}

func TestAuctionCanForge(t *testing.T) {
	slotConst := 4
	blockNum := int(BLOCKSPERSLOT)*slotConst + genesisBlock

	canForge, err := auctionClient.AuctionCanForge(governanceAddressConst, int64(blockNum))
	require.Nil(t, err)
	assert.Equal(t, canForge, true)
}

func TestAuctionMultiBid(t *testing.T) {
	currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
	require.Nil(t, err)
	slotSet := [6]bool{true, false, true, false, true, false}
	maxBid := new(big.Int)
	maxBid.SetString("15000000000000000000", 10)
	minBid := new(big.Int)
	minBid.SetString("11000000000000000000", 10)
	budget := new(big.Int)
	budget.SetString("45200000000000000000", 10)
	forgerAddress := governanceAddressConst
	_, err = auctionClient.AuctionMultiBid(currentSlot+4, currentSlot+10, slotSet, maxBid, minBid, budget)
	require.Nil(t, err)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, forgerAddress, auctionEvents.NewBid[0].Bidder)
	assert.Equal(t, currentSlot+4, auctionEvents.NewBid[0].Slot)
	assert.Equal(t, forgerAddress, auctionEvents.NewBid[1].Bidder)
	assert.Equal(t, currentSlot+6, auctionEvents.NewBid[1].Slot)
	assert.Equal(t, forgerAddress, auctionEvents.NewBid[2].Bidder)
	assert.Equal(t, currentSlot+8, auctionEvents.NewBid[2].Slot)
	assert.Equal(t, forgerAddress, auctionEvents.NewBid[3].Bidder)
	assert.Equal(t, currentSlot+10, auctionEvents.NewBid[3].Slot)
}

func TestAuctionGetClaimableHEZ2(t *testing.T) {
	forgerAddress := governanceAddressConst
	amount := new(big.Int)
	amount.SetString("11000000000000000000", 10)

	claimableHEZ, err := auctionClient.AuctionGetClaimableHEZ(forgerAddress)
	require.Nil(t, err)
	assert.Equal(t, claimableHEZ, amount)
}

func TestAuctionClaimHEZ(t *testing.T) {
	amount := new(big.Int)
	amount.SetString("11000000000000000000", 10)

	_, err := auctionClient.AuctionClaimHEZ()
	require.Nil(t, err)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	auctionEvents, _, _ := auctionClient.AuctionEventsByBlock(currentBlockNum)
	assert.Equal(t, amount, auctionEvents.HEZClaimed[0].Amount)
	assert.Equal(t, governanceAddressConst, auctionEvents.HEZClaimed[0].Owner)
}

func TestAuctionForge(t *testing.T) {
	auctionClientHermez, err := NewAuctionClient(ethereumClientHermez, auctionAddressConst, tokenHezAddressConst)
	require.Nil(t, err)
	slotConst := 4
	blockNum := int64(int(BLOCKSPERSLOT)*slotConst + genesisBlock)
	currentBlockNum, _ := auctionClient.client.EthCurrentBlock()
	blocksToAdd := blockNum - currentBlockNum
	addBlocks(blocksToAdd, ethClientDialURL)
	currentBlockNum, _ = auctionClient.client.EthCurrentBlock()
	assert.Equal(t, currentBlockNum, blockNum)
	_, err = auctionClientHermez.AuctionForge(bootCoordinatorAddressConst)
	require.Contains(t, err.Error(), "Can't forge")
	_, err = auctionClientHermez.AuctionForge(governanceAddressConst)
	require.Nil(t, err)
}
