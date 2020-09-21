package eth

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
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

/*var donationAddressStr = os.Getenv("DONATION_ADDRESS")
var bootCoordinatorStr = os.Getenv("BOOT_COORDINATOR_ADDRESS")
var auctionAddressStr = os.Getenv("AUCTION_ADDRESS")
var tokenHezStr = os.Getenv("TOKEN_ADDRESS")
var hermezStr = os.Getenv("HERMEZ_ADDRESS")
var governanceAddressStr = os.Getenv("GOV_ADDRESS")
var governancePrivateKey = os.Getenv("GOV_PK")
var ehtClientDialURL = os.Getenv("ETHCLIENT_DIAL_URL")*/
var integration = os.Getenv("INTEGRATION")

var donationAddressStr = "0x6c365935CA8710200C7595F0a72EB6023A7706Cd"
var bootCoordinatorStr = "0xc783df8a850f42e7f7e57013759c285caa701eb6"
var auctionAddressStr = "0x3619DbE27d7c1e7E91aA738697Ae7Bc5FC3eACA5"
var tokenHezStr = "0xf4e77E5Da47AC3125140c470c71cBca77B5c638c" //nolint:gosec
var hermezStr = "0xc4905364b78a742ccce7B890A89514061E47068D"
var governanceAddressStr = "0xead9c93b79ae7c1591b1fb5323bd777e86e150d4"
var governancePrivateKey = "d49743deccbccc5dc7baa8e69e5be03298da8688a15dd202e20f15d5e0e9a9fb"
var ehtClientDialURL = "http://localhost:8545"
var DONATION = common.HexToAddress(donationAddressStr)
var BOOTCOORDINATOR = common.HexToAddress(bootCoordinatorStr)
var TOKENHEZ = common.HexToAddress(tokenHezStr)
var HERMEZROLLUP = common.HexToAddress(hermezStr)

var minBidStr = "10000000000000000000"
var URL = "http://localhost:3000"
var newURL = "http://localhost:3002"
var BLOCKSPERSLOT = uint8(40)
var password = "pass"

func TestNewAction(t *testing.T) {
	key, err := crypto.HexToECDSA(governancePrivateKey)
	require.Nil(t, err)
	dir, err := ioutil.TempDir("", "tmpks")
	require.Nil(t, err)
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	account, err := ks.ImportECDSA(key, password)
	require.Nil(t, err)
	err = ks.Unlock(account, password)
	require.Nil(t, err)
	// Init eth client
	ethClient, err := ethclient.Dial(ehtClientDialURL)
	require.Nil(t, err)
	ethereumClient := NewEthereumClient(ethClient, &account, ks, nil)
	auctionAddress := common.HexToAddress(auctionAddressStr)
	if integration != "" {
		auctionClient = NewAuctionClient(ethereumClient, auctionAddress)
	}
}

func TestAuctionGetCurrentSlotNumber(t *testing.T) {
	if auctionClient != nil {
		currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
		require.Nil(t, err)
		currentSlotInt := int(currentSlot)
		assert.Equal(t, currentSlotConst, currentSlotInt)
	}
}

func TestAuctionConstants(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)
	if auctionClient != nil {
		auctionConstants, err := auctionClient.AuctionConstants()
		require.Nil(t, err)
		assert.Equal(t, auctionConstants.BlocksPerSlot, BLOCKSPERSLOT)
		// assert.Equal(t, auctionConstants.GenesisBlockNum, GENESISBLOCKNUM)
		assert.Equal(t, auctionConstants.HermezRollup, HERMEZROLLUP)
		assert.Equal(t, auctionConstants.InitialMinimalBidding, INITMINBID)
		assert.Equal(t, auctionConstants.TokenHEZ, TOKENHEZ)
	}
}

func TestAuctionVariables(t *testing.T) {
	INITMINBID := new(big.Int)
	INITMINBID.SetString(minBidStr, 10)
	defaultSlotSetBid := [6]*big.Int{INITMINBID, INITMINBID, INITMINBID, INITMINBID, INITMINBID, INITMINBID}
	if auctionClient != nil {
		auctionVariables, err := auctionClient.AuctionVariables()
		require.Nil(t, err)
		assert.Equal(t, auctionVariables.AllocationRatio, allocationRatioConst)
		assert.Equal(t, auctionVariables.BootCoordinator, BOOTCOORDINATOR)
		assert.Equal(t, auctionVariables.ClosedAuctionSlots, closedAuctionSlotsConst)
		assert.Equal(t, auctionVariables.DefaultSlotSetBid, defaultSlotSetBid)
		assert.Equal(t, auctionVariables.DonationAddress, DONATION)
		assert.Equal(t, auctionVariables.OpenAuctionSlots, openAuctionSlotsConst)
		assert.Equal(t, auctionVariables.Outbidding, outbiddingConst)
		assert.Equal(t, auctionVariables.SlotDeadline, slotDeadlineConst)
	}
}

func TestAuctionGetSlotDeadline(t *testing.T) {
	if auctionClient != nil {
		slotDeadline, err := auctionClient.AuctionGetSlotDeadline()
		require.Nil(t, err)
		assert.Equal(t, slotDeadlineConst, slotDeadline)
	}
}

func TestAuctionSetSlotDeadline(t *testing.T) {
	newSlotDeadline := uint8(25)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetSlotDeadline(newSlotDeadline)
		require.Nil(t, err)
		slotDeadline, err := auctionClient.AuctionGetSlotDeadline()
		require.Nil(t, err)
		assert.Equal(t, newSlotDeadline, slotDeadline)
		_, err = auctionClient.AuctionSetSlotDeadline(slotDeadlineConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetOpenAuctionSlots(t *testing.T) {
	if auctionClient != nil {
		openAuctionSlots, err := auctionClient.AuctionGetOpenAuctionSlots()
		require.Nil(t, err)
		assert.Equal(t, openAuctionSlotsConst, openAuctionSlots)
	}
}

func TestAuctionSetOpenAuctionSlots(t *testing.T) {
	newOpenAuctionSlots := uint16(4500)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetOpenAuctionSlots(newOpenAuctionSlots)
		require.Nil(t, err)
		openAuctionSlots, err := auctionClient.AuctionGetOpenAuctionSlots()
		require.Nil(t, err)
		assert.Equal(t, newOpenAuctionSlots, openAuctionSlots)
		_, err = auctionClient.AuctionSetOpenAuctionSlots(openAuctionSlotsConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetClosedAuctionSlots(t *testing.T) {
	if auctionClient != nil {
		closedAuctionSlots, err := auctionClient.AuctionGetClosedAuctionSlots()
		require.Nil(t, err)
		assert.Equal(t, closedAuctionSlotsConst, closedAuctionSlots)
	}
}

func TestAuctionSetClosedAuctionSlots(t *testing.T) {
	newClosedAuctionSlots := uint16(5)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetClosedAuctionSlots(newClosedAuctionSlots)
		require.Nil(t, err)
		closedAuctionSlots, err := auctionClient.AuctionGetClosedAuctionSlots()
		require.Nil(t, err)
		assert.Equal(t, newClosedAuctionSlots, closedAuctionSlots)
		_, err = auctionClient.AuctionSetClosedAuctionSlots(closedAuctionSlotsConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetOutbidding(t *testing.T) {
	if auctionClient != nil {
		outbidding, err := auctionClient.AuctionGetOutbidding()
		require.Nil(t, err)
		assert.Equal(t, outbiddingConst, outbidding)
	}
}

func TestAuctionSetOutbidding(t *testing.T) {
	newOutbidding := uint16(0xb)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetOutbidding(newOutbidding)
		require.Nil(t, err)
		outbidding, err := auctionClient.AuctionGetOutbidding()
		require.Nil(t, err)
		assert.Equal(t, newOutbidding, outbidding)
		_, err = auctionClient.AuctionSetOutbidding(outbiddingConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetAllocationRatio(t *testing.T) {
	if auctionClient != nil {
		allocationRatio, err := auctionClient.AuctionGetAllocationRatio()
		require.Nil(t, err)
		assert.Equal(t, allocationRatioConst, allocationRatio)
	}
}

func TestAuctionSetAllocationRatio(t *testing.T) {
	newAllocationRatio := [3]uint16{3000, 3000, 4000}
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetAllocationRatio(newAllocationRatio)
		require.Nil(t, err)
		allocationRatio, err := auctionClient.AuctionGetAllocationRatio()
		require.Nil(t, err)
		assert.Equal(t, newAllocationRatio, allocationRatio)
		_, err = auctionClient.AuctionSetAllocationRatio(allocationRatioConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetDonationAddress(t *testing.T) {
	if auctionClient != nil {
		donationAddress, err := auctionClient.AuctionGetDonationAddress()
		require.Nil(t, err)
		donationAddressConst := common.HexToAddress(donationAddressStr)
		assert.Equal(t, &donationAddressConst, donationAddress)
	}
}

func TestAuctionGetBootCoordinator(t *testing.T) {
	if auctionClient != nil {
		bootCoordinator, err := auctionClient.AuctionGetBootCoordinator()
		require.Nil(t, err)
		bootCoordinatorConst := common.HexToAddress(bootCoordinatorStr)
		assert.Equal(t, &bootCoordinatorConst, bootCoordinator)
	}
}

func TestAuctionSetDonationAddress(t *testing.T) {
	newDonationAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetDonationAddress(newDonationAddress)
		require.Nil(t, err)
		donationAddress, err := auctionClient.AuctionGetDonationAddress()
		require.Nil(t, err)
		assert.Equal(t, &newDonationAddress, donationAddress)
		donationAddressConst := common.HexToAddress(donationAddressStr)
		_, err = auctionClient.AuctionSetDonationAddress(donationAddressConst)
		require.Nil(t, err)
	}
}

func TestAuctionSetBootCoordinator(t *testing.T) {
	newBootCoordinator := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		_, err := auctionClient.AuctionSetBootCoordinator(newBootCoordinator)
		require.Nil(t, err)
		bootCoordinator, err := auctionClient.AuctionGetBootCoordinator()
		require.Nil(t, err)
		assert.Equal(t, &newBootCoordinator, bootCoordinator)
		bootCoordinatorConst := common.HexToAddress(bootCoordinatorStr)
		_, err = auctionClient.AuctionSetBootCoordinator(bootCoordinatorConst)
		require.Nil(t, err)
	}
}

func TestAuctionGetSlotSet(t *testing.T) {
	slot := int64(10)
	if auctionClient != nil {
		slotSet, err := auctionClient.AuctionGetSlotSet(slot)
		require.Nil(t, err)
		assert.Equal(t, slotSet, big.NewInt(4))
	}
}

func TestAuctionGetDefaultSlotSetBid(t *testing.T) {
	slotSet := uint8(3)
	if auctionClient != nil {
		minBid, err := auctionClient.AuctionGetDefaultSlotSetBid(slotSet)
		require.Nil(t, err)
		assert.Equal(t, minBid.String(), minBidStr)
	}
}

func TestAuctionChangeDefaultSlotSetBid(t *testing.T) {
	slotSet := int64(3)
	set := uint8(3)
	newInitialMinBid := new(big.Int)
	newInitialMinBid.SetString("20000000000000000000", 10)
	if auctionClient != nil {
		_, err := auctionClient.AuctionChangeDefaultSlotSetBid(slotSet, newInitialMinBid)
		require.Nil(t, err)
		minBid, err := auctionClient.AuctionGetDefaultSlotSetBid(set)
		require.Nil(t, err)
		assert.Equal(t, minBid, newInitialMinBid)
		newMinBid := new(big.Int)
		newMinBid.SetString("10000000000000000000", 10)
		_, err = auctionClient.AuctionChangeDefaultSlotSetBid(slotSet, newMinBid)
		require.Nil(t, err)
	}
}

func TestAuctionGetClaimableHEZ(t *testing.T) {
	forgerAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		claimableHEZ, err := auctionClient.AuctionGetClaimableHEZ(forgerAddress)
		require.Nil(t, err)
		assert.Equal(t, claimableHEZ.Int64(), int64(0))
	}
}

func TestAuctionIsRegisteredCoordinator(t *testing.T) {
	forgerAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		registered, err := auctionClient.AuctionIsRegisteredCoordinator(forgerAddress)
		require.Nil(t, err)
		assert.Equal(t, registered, false)
	}
}

func TestAuctionRegisterCoordinator(t *testing.T) {
	forgerAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		_, err := auctionClient.AuctionRegisterCoordinator(forgerAddress, URL)
		require.Nil(t, err)
	}
}

func TestAuctionIsRegisteredCoordinatorTrue(t *testing.T) {
	forgerAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		registered, err := auctionClient.AuctionIsRegisteredCoordinator(forgerAddress)
		require.Nil(t, err)
		assert.Equal(t, registered, true)
	}
}

func TestAuctionUpdateCoordinatorInfo(t *testing.T) {
	forgerAddress := common.HexToAddress(governanceAddressStr)
	if auctionClient != nil {
		_, err := auctionClient.AuctionUpdateCoordinatorInfo(forgerAddress, forgerAddress, newURL)
		require.Nil(t, err)
	}
}

func TestAuctionBid(t *testing.T) {
	if auctionClient != nil {
		currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
		require.Nil(t, err)
		bidAmount := new(big.Int)
		bidAmount.SetString("11000000000000000000", 10)
		forgerAddress := common.HexToAddress(governanceAddressStr)
		_, err = auctionClient.AuctionBid(currentSlot+4, bidAmount, forgerAddress, TOKENHEZ)
		require.Nil(t, err)
	}
}

func TestAuctionMultiBid(t *testing.T) {
	if auctionClient != nil {
		currentSlot, err := auctionClient.AuctionGetCurrentSlotNumber()
		require.Nil(t, err)
		slotSet := [6]bool{false, true, false, true, false, true}
		maxBid := new(big.Int)
		maxBid.SetString("11000000000000000000", 10)
		minBid := new(big.Int)
		minBid.SetString("11000000000000000000", 10)
		budget := new(big.Int)
		budget.SetString("110000000000000000000", 10)
		forgerAddress := common.HexToAddress(governanceAddressStr)
		_, err = auctionClient.AuctionMultiBid(currentSlot+5, currentSlot+10, slotSet, maxBid, minBid, budget, forgerAddress, TOKENHEZ)
		require.Nil(t, err)
	}
}
