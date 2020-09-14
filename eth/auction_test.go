package eth

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const slotDeadlineConst = uint8(20)
const openAuctionSlotsConst = 4320
const closedAuctionSlotsConst = 2
const outbiddingConst = 10
const currentSlotConst = 0

var allocationRatioConst [3]uint8 = [3]uint8{40, 40, 20}

var auctionClient *AuctionClient

var donationAddressConstStr = os.Getenv("DONATION_ADDRESS")
var bootCoordinatorConstStr = os.Getenv("BOOT_COORDINATOR_ADDRESS")
var integration = os.Getenv("INTEGRATION")
var ehtClientDialURL = os.Getenv("ETHCLIENT_DIAL_URL")
var auctionAddressStr = os.Getenv("AUCTION_ADDRESS")

func TestNewAction(t *testing.T) {
	if integration != "" {
		// Init eth client
		ethClient, err := ethclient.Dial(ehtClientDialURL)
		require.Nil(t, err)
		ethereumClient := NewEthereumClient(ethClient, nil, nil, nil)
		auctionAddress := common.HexToAddress(auctionAddressStr)
		auctionClient = NewAuctionClient(ethereumClient, auctionAddress)
	}
}

func TestAuctionGetSlotDeadline(t *testing.T) {
	if auctionClient != nil {
		slotDeadline, err := auctionClient.AuctionGetSlotDeadline()
		require.Nil(t, err)
		assert.Equal(t, slotDeadlineConst, slotDeadline)
	}
}

func TestAuctionGetOpenAuctionSlots(t *testing.T) {
	if auctionClient != nil {
		openAuctionSlots, err := auctionClient.AuctionGetOpenAuctionSlots()
		require.Nil(t, err)
		openAuctionSlotsInt := int(openAuctionSlots)
		assert.Equal(t, openAuctionSlotsConst, openAuctionSlotsInt)
	}
}

func TestAuctionGetClosedAuctionSlots(t *testing.T) {
	if auctionClient != nil {
		closedAuctionSlots, err := auctionClient.AuctionGetClosedAuctionSlots()
		require.Nil(t, err)
		closedAuctionSlotsInt := int(closedAuctionSlots)
		assert.Equal(t, closedAuctionSlotsConst, closedAuctionSlotsInt)
	}
}

func TestAuctionGetOutbidding(t *testing.T) {
	if auctionClient != nil {
		outbidding, err := auctionClient.AuctionGetOutbidding()
		require.Nil(t, err)
		outbiddingInt := int(outbidding)
		assert.Equal(t, outbiddingConst, outbiddingInt)
	}
}

func TestAuctionGetAllocationRatio(t *testing.T) {
	if auctionClient != nil {
		allocationRatio, err := auctionClient.AuctionGetAllocationRatio()
		require.Nil(t, err)
		assert.Equal(t, allocationRatioConst, allocationRatio)
	}
}

func TestAuctionGetDonationAddress(t *testing.T) {
	if auctionClient != nil {
		donationAddress, err := auctionClient.AuctionGetDonationAddress()
		require.Nil(t, err)
		donationAddressConst := common.HexToAddress(donationAddressConstStr)
		assert.Equal(t, &donationAddressConst, donationAddress)
	}
}

func TestAuctionGetBootCoordinator(t *testing.T) {
	if auctionClient != nil {
		bootCoordinator, err := auctionClient.AuctionGetBootCoordinator()
		require.Nil(t, err)
		bootCoordinatorConst := common.HexToAddress(bootCoordinatorConstStr)
		assert.Equal(t, &bootCoordinatorConst, bootCoordinator)
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
