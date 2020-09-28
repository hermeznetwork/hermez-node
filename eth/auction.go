package eth

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	HermezAuctionProtocol "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
	ERC777 "github.com/hermeznetwork/hermez-node/eth/contracts/erc777"
	"golang.org/x/crypto/sha3"
)

// AuctionConstants are the constants of the Rollup Smart Contract
type AuctionConstants struct {
	// Blocks per slot
	BlocksPerSlot uint8
	// Minimum bid when no one has bid yet
	InitialMinimalBidding *big.Int
	// First block where the first slot begins
	GenesisBlockNum int64
	// Hermez Governanze Token smartcontract address who controls some parameters and collects HEZ fee
	// Only for test
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

// NewSlotState returns an empty SlotState
func NewSlotState() *SlotState {
	return &SlotState{
		Forger:       ethCommon.Address{},
		BidAmount:    big.NewInt(0),
		ClosedMinBid: big.NewInt(0),
		Fulfilled:    false,
	}
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
	DefaultSlotSetBid [6]*big.Int
	// Distance (#slots) to the closest slot to which you can bid ( 2 Slots = 2 * 40 Blocks = 20 min )
	ClosedAuctionSlots uint16
	// Distance (#slots) to the farthest slot to which you can bid (30 days = 4320 slots )
	OpenAuctionSlots uint16
	// How the HEZ tokens deposited by the slot winner are distributed (Burn: 40% - Donation: 40% - HGT: 20%)
	AllocationRatio [3]uint16
	// Minimum outbid (percentage) over the previous one to consider it valid
	Outbidding uint16
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
	NewOutbidding uint16
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
	NewAllocationRatio [3]uint16
}

// AuctionEventNewCoordinator is an event of the Auction Smart Contract
type AuctionEventNewCoordinator struct {
	ForgerAddress     ethCommon.Address
	WithdrawalAddress ethCommon.Address
	CoordinatorURL    string
}

// AuctionEventCoordinatorUpdated is an event of the Auction Smart Contract
type AuctionEventCoordinatorUpdated struct {
	ForgerAddress     ethCommon.Address
	WithdrawalAddress ethCommon.Address
	CoordinatorURL    string
}

// AuctionEventNewForgeAllocated is an event of the Auction Smart Contract
type AuctionEventNewForgeAllocated struct {
	Forger           ethCommon.Address
	CurrentSlot      int64
	BurnAmount       *big.Int
	DonationAmount   *big.Int
	GovernanceAmount *big.Int
}

// AuctionEventNewDefaultSlotSetBid is an event of the Auction Smart Contract
type AuctionEventNewDefaultSlotSetBid struct {
	SlotSet          int64
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
	NewDefaultSlotSetBid  []AuctionEventNewDefaultSlotSetBid
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
		NewDefaultSlotSetBid:  make([]AuctionEventNewDefaultSlotSetBid, 0),
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
	AuctionSetOutbidding(newOutbidding uint16) (*types.Transaction, error)
	AuctionGetOutbidding() (uint16, error)
	AuctionSetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error)
	AuctionGetAllocationRatio() ([3]uint16, error)
	AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error)
	AuctionGetDonationAddress() (*ethCommon.Address, error)
	AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error)
	AuctionGetBootCoordinator() (*ethCommon.Address, error)
	AuctionChangeDefaultSlotSetBid(slotSet int64, newInitialMinBid *big.Int) (*types.Transaction, error)

	// Coordinator Management
	AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error)
	AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error)
	AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error)

	// Slot Info
	AuctionGetCurrentSlotNumber() (int64, error)
	AuctionGetMinBidBySlot(slot int64) (*big.Int, error)
	AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error)
	AuctionGetSlotSet(slot int64) (*big.Int, error)
	AuctionGetSlotNumber(blockNum int64) (*big.Int, error)

	// Bidding
	// AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int,
	// 	userData, operatorData []byte) error // Only called from another smart contract
	AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error)
	AuctionMultiBid(startingSlot int64, endingSlot int64, slotSet [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error)

	// Forge
	AuctionCanForge(forger ethCommon.Address, blockNum int64) (bool, error)
	// AuctionForge(forger ethCommon.Address) (bool, error) // Only called from another smart contract

	// Fees
	AuctionClaimHEZ(claimAddress ethCommon.Address) (*types.Transaction, error)

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
	client       *EthereumClient
	address      ethCommon.Address
	tokenAddress ethCommon.Address
	gasLimit     uint64
	contractAbi  abi.ABI
}

// NewAuctionClient creates a new AuctionClient.  `tokenAddress` is the address of the HEZ tokens.
func NewAuctionClient(client *EthereumClient, address, tokenAddress ethCommon.Address) *AuctionClient {
	contractAbi, err := abi.JSON(strings.NewReader(string(HermezAuctionProtocol.HermezAuctionProtocolABI)))
	if err != nil {
		fmt.Println(err)
	}
	return &AuctionClient{
		client:       client,
		address:      address,
		tokenAddress: tokenAddress,
		gasLimit:     1000000, //nolint:gomnd
		contractAbi:  contractAbi,
	}
}

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetSlotDeadline(auth, newDeadline)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting slotDeadline: %w", err)
	}
	return tx, nil
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotDeadline() (uint8, error) {
	var slotDeadline uint8
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		slotDeadline, err = auction.GetSlotDeadline(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return slotDeadline, nil
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetOpenAuctionSlots(auth, newOpenAuctionSlots)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting openAuctionSlots: %w", err)
	}
	return tx, nil
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOpenAuctionSlots() (uint16, error) {
	var openAuctionSlots uint16
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		openAuctionSlots, err = auction.GetOpenAuctionSlots(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return openAuctionSlots, nil
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetClosedAuctionSlots(auth, newClosedAuctionSlots)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting closedAuctionSlots: %w", err)
	}
	return tx, nil
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetClosedAuctionSlots() (uint16, error) {
	var closedAuctionSlots uint16
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		closedAuctionSlots, err = auction.GetClosedAuctionSlots(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return closedAuctionSlots, nil
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOutbidding(newOutbidding uint16) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		12500000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetOutbidding(auth, newOutbidding)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting setOutbidding: %w", err)
	}
	return tx, nil
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOutbidding() (uint16, error) {
	var outbidding uint16
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		outbidding, err = auction.GetOutbidding(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return outbidding, nil
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetAllocationRatio(newAllocationRatio [3]uint16) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetAllocationRatio(auth, newAllocationRatio)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting allocationRatio: %w", err)
	}
	return tx, nil
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetAllocationRatio() ([3]uint16, error) {
	var allocationRation [3]uint16
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		allocationRation, err = auction.GetAllocationRatio(nil)
		return err
	}); err != nil {
		return [3]uint16{}, err
	}
	return allocationRation, nil
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetDonationAddress(auth, newDonationAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting donationAddress: %w", err)
	}
	return tx, nil
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	var donationAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		donationAddress, err = auction.GetDonationAddress(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &donationAddress, nil
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.SetBootCoordinator(auth, newBootCoordinator)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting bootCoordinator: %w", err)
	}
	return tx, nil
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetBootCoordinator() (*ethCommon.Address, error) {
	var bootCoordinator ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		bootCoordinator, err = auction.GetBootCoordinator(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &bootCoordinator, nil
}

// AuctionChangeDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionChangeDefaultSlotSetBid(slotSet int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			slotSetToSend := big.NewInt(slotSet)
			return auction.ChangeDefaultSlotSetBid(auth, slotSetToSend, newInitialMinBid)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed changing slotSet Bid: %w", err)
	}
	return tx, nil
}

// AuctionGetClaimableHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetClaimableHEZ(claimAddress ethCommon.Address) (*big.Int, error) {
	var claimableHEZ *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		claimableHEZ, err = auction.GetClaimableHEZ(nil, claimAddress)
		return err
	}); err != nil {
		return nil, err
	}
	return claimableHEZ, nil
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.RegisterCoordinator(auth, forgerAddress, URL)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed register coordinator: %w", err)
	}
	return tx, nil
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	var registered bool
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		registered, err = auction.IsRegisteredCoordinator(nil, forgerAddress)
		return err
	}); err != nil {
		return false, err
	}
	return registered, nil
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *AuctionClient) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.UpdateCoordinatorInfo(auth, forgerAddress, newWithdrawAddress, newURL)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed update coordinator info: %w", err)
	}
	return tx, nil
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetCurrentSlotNumber() (int64, error) {
	var _currentSlotNumber *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		_currentSlotNumber, err = auction.GetCurrentSlotNumber(nil)
		return err
	}); err != nil {
		return 0, err
	}
	currentSlotNumber := _currentSlotNumber.Int64()
	return currentSlotNumber, nil
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	var minBid *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		slotToSend := big.NewInt(slot)
		minBid, err = auction.GetMinBidBySlot(nil, slotToSend)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return minBid, nil
}

// AuctionGetSlotSet is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotSet(slot int64) (*big.Int, error) {
	var slotSet *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		slotToSend := big.NewInt(slot)
		slotSet, err = auction.GetSlotSet(nil, slotToSend)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return slotSet, nil
}

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	var minBidSlotSet *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		minBidSlotSet, err = auction.GetDefaultSlotSetBid(nil, slotSet)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return minBidSlotSet, nil
}

// AuctionGetSlotNumber is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotNumber(blockNum int64) (*big.Int, error) {
	var slot *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		blockNumBig := big.NewInt(blockNum)
		slot, err = auction.GetSlotNumber(nil, blockNumBig)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return slot, nil
}

// AuctionBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			tokens, err := ERC777.NewERC777(c.tokenAddress, ec)
			if err != nil {
				return nil, err
			}
			bidFnSignature := []byte("bid(uint128,uint128,address)")
			hash := sha3.NewLegacyKeccak256()
			_, err = hash.Write(bidFnSignature)
			if err != nil {
				return nil, err
			}
			methodID := hash.Sum(nil)[:4]
			slotBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(slotBytes, uint64(slot))
			paddedSlot := ethCommon.LeftPadBytes(slotBytes, 32)
			paddedAmount := ethCommon.LeftPadBytes(bidAmount.Bytes(), 32)
			paddedAddress := ethCommon.LeftPadBytes(forger.Bytes(), 32)
			var userData []byte
			userData = append(userData, methodID...)
			userData = append(userData, paddedSlot...)
			userData = append(userData, paddedAmount...)
			userData = append(userData, paddedAddress...)
			return tokens.Send(auth, c.address, bidAmount, userData)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed bid: %w", err)
	}
	return tx, nil
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionMultiBid(startingSlot int64, endingSlot int64, slotSet [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			tokens, err := ERC777.NewERC777(c.tokenAddress, ec)
			if err != nil {
				return nil, err
			}
			multiBidFnSignature := []byte("multiBid(uint128,uint128,bool[6],uint128,uint128,address)")
			hash := sha3.NewLegacyKeccak256()
			_, err = hash.Write(multiBidFnSignature)
			if err != nil {
				return nil, err
			}
			methodID := hash.Sum(nil)[:4]
			startingSlotBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(startingSlotBytes, uint64(startingSlot))
			paddedStartingSlot := ethCommon.LeftPadBytes(startingSlotBytes, 32)
			endingSlotBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(endingSlotBytes, uint64(endingSlot))
			paddedEndingSlot := ethCommon.LeftPadBytes(endingSlotBytes, 32)
			paddedMinBid := ethCommon.LeftPadBytes(closedMinBid.Bytes(), 32)
			paddedMaxBid := ethCommon.LeftPadBytes(maxBid.Bytes(), 32)
			paddedAddress := ethCommon.LeftPadBytes(forger.Bytes(), 32)
			var userData []byte
			userData = append(userData, methodID...)
			userData = append(userData, paddedStartingSlot...)
			userData = append(userData, paddedEndingSlot...)
			for i := 0; i < len(slotSet); i++ {
				if slotSet[i] {
					paddedSlotSet := ethCommon.LeftPadBytes([]byte{1}, 32)
					userData = append(userData, paddedSlotSet...)
				} else {
					paddedSlotSet := ethCommon.LeftPadBytes([]byte{0}, 32)
					userData = append(userData, paddedSlotSet...)
				}
			}
			userData = append(userData, paddedMaxBid...)
			userData = append(userData, paddedMinBid...)
			userData = append(userData, paddedAddress...)
			return tokens.Send(auth, c.address, budget, userData)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed multibid: %w", err)
	}
	return tx, nil
}

// AuctionCanForge is the interface to call the smart contract function
func (c *AuctionClient) AuctionCanForge(forger ethCommon.Address, blockNum int64) (bool, error) {
	var canForge bool
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		canForge, err = auction.CanForge(nil, forger, big.NewInt(blockNum))
		return err
	}); err != nil {
		return false, err
	}
	return canForge, nil
}

// AuctionForge is the interface to call the smart contract function
// func (c *AuctionClient) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionClaimHEZ(claimAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		c.gasLimit,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			return auction.ClaimHEZ(auth, claimAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed claim HEZ: %w", err)
	}
	return tx, nil
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *AuctionClient) AuctionConstants() (*AuctionConstants, error) {
	auctionConstants := new(AuctionConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
		if err != nil {
			return err
		}
		auctionConstants.BlocksPerSlot, err = auction.BLOCKSPERSLOT(nil)
		if err != nil {
			return err
		}
		genesisBlock, err := auction.GenesisBlock(nil)
		if err != nil {
			return err
		}
		auctionConstants.GenesisBlockNum = genesisBlock.Int64()
		auctionConstants.HermezRollup, err = auction.HermezRollup(nil)
		if err != nil {
			return err
		}
		auctionConstants.InitialMinimalBidding, err = auction.INITIALMINIMALBIDDING(nil)
		if err != nil {
			return err
		}
		auctionConstants.TokenHEZ, err = auction.TokenHEZ(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return auctionConstants, nil
}

// AuctionVariables returns the variables of the Auction Smart Contract
func (c *AuctionClient) AuctionVariables() (*AuctionVariables, error) {
	auctionVariables := new(AuctionVariables)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		var err error
		auctionVariables.AllocationRatio, err = c.AuctionGetAllocationRatio()
		if err != nil {
			return err
		}
		bootCoordinator, err := c.AuctionGetBootCoordinator()
		if err != nil {
			return err
		}
		auctionVariables.BootCoordinator = *bootCoordinator
		auctionVariables.ClosedAuctionSlots, err = c.AuctionGetClosedAuctionSlots()
		if err != nil {
			return err
		}
		var defaultSlotSetBid [6]*big.Int
		for i := uint8(0); i < 6; i++ {
			bid, err := c.AuctionGetDefaultSlotSetBid(i)
			if err != nil {
				return err
			}
			defaultSlotSetBid[i] = bid
		}
		auctionVariables.DefaultSlotSetBid = defaultSlotSetBid
		donationAddress, err := c.AuctionGetDonationAddress()
		if err != nil {
			return err
		}
		auctionVariables.DonationAddress = *donationAddress
		auctionVariables.OpenAuctionSlots, err = c.AuctionGetOpenAuctionSlots()
		if err != nil {
			return err
		}
		auctionVariables.Outbidding, err = c.AuctionGetOutbidding()
		if err != nil {
			return err
		}
		auctionVariables.SlotDeadline, err = c.AuctionGetSlotDeadline()
		return err
	}); err != nil {
		return nil, err
	}
	return auctionVariables, nil
}

var (
	logNewBid                = crypto.Keccak256Hash([]byte("NewBid(uint128,uint128,address)"))
	logNewSlotDeadline       = crypto.Keccak256Hash([]byte("NewSlotDeadline(uint8)"))
	logNewClosedAuctionSlots = crypto.Keccak256Hash([]byte("NewClosedAuctionSlots(uint16)"))
	logNewOutbidding         = crypto.Keccak256Hash([]byte("NewOutbidding(uint16)"))
	logNewDonationAddress    = crypto.Keccak256Hash([]byte("NewDonationAddress(address)"))
	logNewBootCoordinator    = crypto.Keccak256Hash([]byte("NewBootCoordinator(address)"))
	logNewOpenAuctionSlots   = crypto.Keccak256Hash([]byte("NewOpenAuctionSlots(uint16)"))
	logNewAllocationRatio    = crypto.Keccak256Hash([]byte("NewAllocationRatio(uint16[3])"))
	logNewCoordinator        = crypto.Keccak256Hash([]byte("NewCoordinator(address,address,string)"))
	logCoordinatorUpdated    = crypto.Keccak256Hash([]byte("CoordinatorUpdated(address,address,string)"))
	logNewForgeAllocated     = crypto.Keccak256Hash([]byte("NewForgeAllocated(address,uint128,uint128,uint128,uint128)"))
	logNewDefaultSlotSetBid  = crypto.Keccak256Hash([]byte("NewDefaultSlotSetBid(uint128,uint128)"))
	logNewForge              = crypto.Keccak256Hash([]byte("NewForge(address,uint128)"))
	logHEZClaimed            = crypto.Keccak256Hash([]byte("HEZClaimed(address,uint128)"))
)

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *AuctionClient) AuctionEventsByBlock(blockNum int64) (*AuctionEvents, *ethCommon.Hash, error) {
	var auctionEvents AuctionEvents

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []ethCommon.Address{
			c.address,
		},
		BlockHash: nil, // TODO: Maybe we can put the blockHash here to make sure we get the results from the known block.
		Topics:    [][]ethCommon.Hash{},
	}

	logs, err := c.client.client.FilterLogs(context.TODO(), query)
	if err != nil {
		fmt.Println(err)
	}
	for _, vLog := range logs {
		switch vLog.Topics[0] {
		case logNewBid:
			var auxNewBid struct {
				Slot      *big.Int
				BidAmount *big.Int
				Address   ethCommon.Address
			}
			var newBid AuctionEventNewBid
			if err := c.contractAbi.Unpack(&auxNewBid, "NewBid", vLog.Data); err != nil {
				return nil, nil, err
			}
			newBid.BidAmount = auxNewBid.BidAmount
			newBid.Slot = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			newBid.CoordinatorForger = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			auctionEvents.NewBid = append(auctionEvents.NewBid, newBid)
		case logNewSlotDeadline:
			var newSlotDeadline AuctionEventNewSlotDeadline
			if err := c.contractAbi.Unpack(&newSlotDeadline, "NewSlotDeadline", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewSlotDeadline = append(auctionEvents.NewSlotDeadline, newSlotDeadline)
		case logNewClosedAuctionSlots:
			var newClosedAuctionSlots AuctionEventNewClosedAuctionSlots
			if err := c.contractAbi.Unpack(&newClosedAuctionSlots, "NewClosedAuctionSlots", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewClosedAuctionSlots = append(auctionEvents.NewClosedAuctionSlots, newClosedAuctionSlots)
		case logNewOutbidding:
			var newOutbidding AuctionEventNewOutbidding
			if err := c.contractAbi.Unpack(&newOutbidding, "NewOutbidding", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewOutbidding = append(auctionEvents.NewOutbidding, newOutbidding)
		case logNewDonationAddress:
			var newDonationAddress AuctionEventNewDonationAddress
			if err := c.contractAbi.Unpack(&newDonationAddress, "NewDonationAddress", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewDonationAddress = append(auctionEvents.NewDonationAddress, newDonationAddress)
		case logNewBootCoordinator:
			var newBootCoordinator AuctionEventNewBootCoordinator
			if err := c.contractAbi.Unpack(&newBootCoordinator, "NewBootCoordinator", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewBootCoordinator = append(auctionEvents.NewBootCoordinator, newBootCoordinator)
		case logNewOpenAuctionSlots:
			var newOpenAuctionSlots AuctionEventNewOpenAuctionSlots
			if err := c.contractAbi.Unpack(&newOpenAuctionSlots, "NewOpenAuctionSlots", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewOpenAuctionSlots = append(auctionEvents.NewOpenAuctionSlots, newOpenAuctionSlots)
		case logNewAllocationRatio:
			var newAllocationRatio AuctionEventNewAllocationRatio
			if err := c.contractAbi.Unpack(&newAllocationRatio, "NewAllocationRatio", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewAllocationRatio = append(auctionEvents.NewAllocationRatio, newAllocationRatio)
		case logNewCoordinator:
			var newCoordinator AuctionEventNewCoordinator
			if err := c.contractAbi.Unpack(&newCoordinator, "NewCoordinator", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewCoordinator = append(auctionEvents.NewCoordinator, newCoordinator)
		case logCoordinatorUpdated:
			var coordinatorUpdated AuctionEventCoordinatorUpdated
			if err := c.contractAbi.Unpack(&coordinatorUpdated, "CoordinatorUpdated", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.CoordinatorUpdated = append(auctionEvents.CoordinatorUpdated, coordinatorUpdated)
		case logNewForgeAllocated:
			var newForgeAllocated AuctionEventNewForgeAllocated
			if err := c.contractAbi.Unpack(&newForgeAllocated, "NewForgeAllocated", vLog.Data); err != nil {
				return nil, nil, err
			}
			newForgeAllocated.Forger = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForgeAllocated.CurrentSlot = new(big.Int).SetBytes(vLog.Topics[2][:]).Int64()
			auctionEvents.NewForgeAllocated = append(auctionEvents.NewForgeAllocated, newForgeAllocated)
		case logNewDefaultSlotSetBid:
			var auxNewDefaultSlotSetBid struct {
				SlotSet          *big.Int
				NewInitialMinBid *big.Int
			}
			var newDefaultSlotSetBid AuctionEventNewDefaultSlotSetBid
			if err := c.contractAbi.Unpack(&auxNewDefaultSlotSetBid, "NewDefaultSlotSetBid", vLog.Data); err != nil {
				return nil, nil, err
			}
			newDefaultSlotSetBid.NewInitialMinBid = auxNewDefaultSlotSetBid.NewInitialMinBid
			newDefaultSlotSetBid.SlotSet = auxNewDefaultSlotSetBid.SlotSet.Int64()
			auctionEvents.NewDefaultSlotSetBid = append(auctionEvents.NewDefaultSlotSetBid, newDefaultSlotSetBid)
		case logNewForge:
			var newForge AuctionEventNewForge
			newForge.Forger = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForge.CurrentSlot = new(big.Int).SetBytes(vLog.Topics[2][:]).Int64()
			auctionEvents.NewForge = append(auctionEvents.NewForge, newForge)
		case logHEZClaimed:
			var HEZClaimed AuctionEventHEZClaimed
			if err := c.contractAbi.Unpack(&HEZClaimed, "HEZClaimed", vLog.Data); err != nil {
				return nil, nil, err
			}
			HEZClaimed.Owner = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.HEZClaimed = append(auctionEvents.HEZClaimed, HEZClaimed)
		}
	}
	return &auctionEvents, nil, nil
}
