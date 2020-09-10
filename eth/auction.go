package eth

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// AuctionConstants are the constants of the Rollup Smart Contract
type AuctionConstants struct {
	// Blocks to wait before starting with the first slot
	DelayGenesis uint16
	// Blocks per slot
	BlocksPerSlot uint8
	// Minimum bid when no one has bid yet
	InitialMinimalBidding *big.Int
	// First block where the first slot begins
	GenesisBlockNum int64
	// Hermez Governanze Token smartcontract address who controls some parameters and collects HEZ fee
	GovernanceAddress ethCommon.Address
	// ERC777 token with which the bids will be made
	TokenHEZ ethCommon.Address
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address
}

// SlotState is the state of a slot
type SlotState struct {
	Forger       ethCommon.Address
	BidAmount    *big.Int
	ClosedMinBid *big.Int
	Fulfilled    bool
}

// Coordinator is the details of the Coordinator identified by the forger address
type Coordinator struct {
	WithdrawalAddress ethCommon.Address
	URL               string
}

// AuctionVariables are the variables of the Auction Smart Contract
type AuctionVariables struct {
	// Boot Coordinator Address
	DonationAddress ethCommon.Address
	// Boot Coordinator Address
	BootCoordinator ethCommon.Address
	// The minimum bid value in a series of 6 slots
	MinBidEpoch [6]*big.Int
	// Distance (#slots) to the closest slot to which you can bid ( 2 Slots = 2 * 40 Blocks = 20 min )
	ClosedAuctionSlots uint16
	// Distance (#slots) to the farthest slot to which you can bid (30 days = 4320 slots )
	OpenAuctionSlots uint16
	// How the HEZ tokens deposited by the slot winner are distributed (Burn: 40% - Donation: 40% - HGT: 20%)
	AllocationRatio [3]uint8
	// Minimum outbid (percentage) over the previous one to consider it valid
	Outbidding uint8
	// Number of blocks at the end of a slot in which any coordinator can forge if the winner has not forged one before
	SlotDeadline uint8
}

// AuctionState represents the state of the Rollup in the Smart Contract
type AuctionState struct {
	// Mapping to control slot state
	Slots map[int64]*SlotState
	// Mapping to control balances pending to claim
	PendingBalances map[ethCommon.Address]*big.Int
	// Mapping to register all the coordinators. The address used for the mapping is the forger address
	Coordinators map[ethCommon.Address]*Coordinator
}

// AuctionEventNewBid is an event of the Auction Smart Contract
type AuctionEventNewBid struct {
	Slot              int64
	BidAmount         *big.Int
	CoordinatorForger ethCommon.Address
}

// AuctionEventNewSlotDeadline is an event of the Auction Smart Contract
type AuctionEventNewSlotDeadline struct {
	NewSlotDeadline uint8
}

// AuctionEventNewClosedAuctionSlots is an event of the Auction Smart Contract
type AuctionEventNewClosedAuctionSlots struct {
	NewClosedAuctionSlots uint16
}

// AuctionEventNewOutbidding is an event of the Auction Smart Contract
type AuctionEventNewOutbidding struct {
	NewOutbidding uint8
}

// AuctionEventNewDonationAddress is an event of the Auction Smart Contract
type AuctionEventNewDonationAddress struct {
	NewDonationAddress ethCommon.Address
}

// AuctionEventNewBootCoordinator is an event of the Auction Smart Contract
type AuctionEventNewBootCoordinator struct {
	NewBootCoordinator ethCommon.Address
}

// AuctionEventNewOpenAuctionSlots is an event of the Auction Smart Contract
type AuctionEventNewOpenAuctionSlots struct {
	NewOpenAuctionSlots uint16
}

// AuctionEventNewAllocationRatio is an event of the Auction Smart Contract
type AuctionEventNewAllocationRatio struct {
	NewAllocationRatio [3]uint8
}

// AuctionEventNewCoordinator is an event of the Auction Smart Contract
type AuctionEventNewCoordinator struct {
	ForgerAddress     ethCommon.Address
	WithdrawalAddress ethCommon.Address
	URL               string
}

// AuctionEventCoordinatorUpdated is an event of the Auction Smart Contract
type AuctionEventCoordinatorUpdated struct {
	ForgerAddress     ethCommon.Address
	WithdrawalAddress ethCommon.Address
	URL               string
}

// AuctionEventNewForgeAllocated is an event of the Auction Smart Contract
type AuctionEventNewForgeAllocated struct {
	Forger           ethCommon.Address
	CurrentSlot      int64
	BurnAmount       *big.Int
	DonationAmount   *big.Int
	GovernanceAmount *big.Int
}

// AuctionEventNewMinBidEpoch is an event of the Auction Smart Contract
type AuctionEventNewMinBidEpoch struct {
	SlotEpoch        int64
	NewInitialMinBid *big.Int
}

// AuctionEventNewForge is an event of the Auction Smart Contract
type AuctionEventNewForge struct {
	Forger      ethCommon.Address
	CurrentSlot int64
}

// AuctionEventHEZClaimed is an event of the Auction Smart Contract
type AuctionEventHEZClaimed struct {
	Owner  ethCommon.Address
	Amount *big.Int
}

// AuctionEvents is the list of events in a block of the Auction Smart Contract
type AuctionEvents struct { //nolint:structcheck
	NewBid                []AuctionEventNewBid
	NewSlotDeadline       []AuctionEventNewSlotDeadline
	NewClosedAuctionSlots []AuctionEventNewClosedAuctionSlots
	NewOutbidding         []AuctionEventNewOutbidding
	NewDonationAddress    []AuctionEventNewDonationAddress
	NewBootCoordinator    []AuctionEventNewBootCoordinator
	NewOpenAuctionSlots   []AuctionEventNewOpenAuctionSlots
	NewAllocationRatio    []AuctionEventNewAllocationRatio
	NewCoordinator        []AuctionEventNewCoordinator
	CoordinatorUpdated    []AuctionEventCoordinatorUpdated
	NewForgeAllocated     []AuctionEventNewForgeAllocated
	NewMinBidEpoch        []AuctionEventNewMinBidEpoch
	NewForge              []AuctionEventNewForge
	HEZClaimed            []AuctionEventHEZClaimed
}

// NewAuctionEvents creates an empty AuctionEvents with the slices initialized.
func NewAuctionEvents() AuctionEvents {
	return AuctionEvents{
		NewBid:                make([]AuctionEventNewBid, 0),
		NewSlotDeadline:       make([]AuctionEventNewSlotDeadline, 0),
		NewClosedAuctionSlots: make([]AuctionEventNewClosedAuctionSlots, 0),
		NewOutbidding:         make([]AuctionEventNewOutbidding, 0),
		NewDonationAddress:    make([]AuctionEventNewDonationAddress, 0),
		NewBootCoordinator:    make([]AuctionEventNewBootCoordinator, 0),
		NewOpenAuctionSlots:   make([]AuctionEventNewOpenAuctionSlots, 0),
		NewAllocationRatio:    make([]AuctionEventNewAllocationRatio, 0),
		NewCoordinator:        make([]AuctionEventNewCoordinator, 0),
		CoordinatorUpdated:    make([]AuctionEventCoordinatorUpdated, 0),
		NewForgeAllocated:     make([]AuctionEventNewForgeAllocated, 0),
		NewMinBidEpoch:        make([]AuctionEventNewMinBidEpoch, 0),
		NewForge:              make([]AuctionEventNewForge, 0),
		HEZClaimed:            make([]AuctionEventHEZClaimed, 0),
	}
}

// AuctionInterface is the inteface to to Auction Smart Contract
type AuctionInterface interface {
	//
	// Smart Contract Methods
	//

	// Getter/Setter, where Setter is onlyOwner
	AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error)
	AuctionGetSlotDeadline() (uint8, error)
	AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error)
	AuctionGetOpenAuctionSlots() (uint16, error)
	AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error)
	AuctionGetClosedAuctionSlots() (uint16, error)
	AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error)
	AuctionGetOutbidding() (uint8, error)
	AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error)
	AuctionGetAllocationRatio() ([3]uint8, error)
	AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error)
	AuctionGetDonationAddress() (*ethCommon.Address, error)
	AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error)
	AuctionGetBootCoordinator() (*ethCommon.Address, error)
	AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error)

	// Coordinator Management
	AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error)
	AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error)
	AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error)

	// Slot Info
	AuctionGetCurrentSlotNumber() (int64, error)
	AuctionGetMinBidBySlot(slot int64) (*big.Int, error)
	AuctionGetMinBidEpoch(epoch uint8) (*big.Int, error)

	// Bidding
	// AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int,
	// 	userData, operatorData []byte) error // Only called from another smart contract
	AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error)
	AuctionMultiBid(startingSlot int64, endingSlot int64, slotEpoch [6]bool,
		maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error)

	// Forge
	AuctionCanForge(forger ethCommon.Address) (bool, error)
	// AuctionForge(forger ethCommon.Address) (bool, error) // Only called from another smart contract

	// Fees
	AuctionClaimHEZ() (*types.Transaction, error)

	//
	// Smart Contract Status
	//

	AuctionConstants() (*AuctionConstants, error)
	AuctionEventsByBlock(blockNum int64) (*AuctionEvents, *ethCommon.Hash, error)
}

//
// Implementation
//

// AuctionClient is the implementation of the interface to the Auction Smart Contract in ethereum.
type AuctionClient struct {
}

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotDeadline() (uint8, error) {
	return 0, errTODO
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOpenAuctionSlots() (uint16, error) {
	return 0, errTODO
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetClosedAuctionSlots() (uint16, error) {
	return 0, errTODO
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOutbidding() (uint8, error) {
	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetAllocationRatio() ([3]uint8, error) {
	return [3]uint8{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetBootCoordinator() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionChangeEpochMinBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *AuctionClient) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetCurrentSlotNumber() (int64, error) {
	return 0, errTODO
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	return nil, errTODO
}

// AuctionGetMinBidEpoch is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetMinBidEpoch(epoch uint8) (*big.Int, error) {
	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *AuctionClient) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionMultiBid(startingSlot int64, endingSlot int64, slotEpoch [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *AuctionClient) AuctionCanForge(forger ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionForge is the interface to call the smart contract function
// func (c *AuctionClient) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionClaimHEZ() (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *AuctionClient) AuctionConstants() (*AuctionConstants, error) {
	return nil, errTODO
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *AuctionClient) AuctionEventsByBlock(blockNum int64) (*AuctionEvents, *ethCommon.Hash, error) {
	return nil, nil, errTODO
}
