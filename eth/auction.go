package eth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	HermezAuctionProtocol "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
	"github.com/hermeznetwork/hermez-node/log"
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

	// Bidding
	// AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int,
	// 	userData, operatorData []byte) error // Only called from another smart contract
	AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error)
	AuctionMultiBid(startingSlot int64, endingSlot int64, slotSet [6]bool,
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
	client  *EthereumClient
	address ethCommon.Address
}

// NewAuctionClient creates a new AuctionClient
func NewAuctionClient(client *EthereumClient, address ethCommon.Address) *AuctionClient {
	return &AuctionClient{
		client:  client,
		address: address,
	}
}

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
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
		1000000,
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
		1000000,
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
func (c *AuctionClient) AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
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
	// TODO: Update
	// var outbidding uint8
	// if err := c.client.Call(func(ec *ethclient.Client) error {
	// 	auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	outbidding, err = auction.GetOutbidding(nil)
	// 	return err
	// }); err != nil {
	// 	return 0, err
	// }
	// return outbidding, nil
	log.Error("TODO")
	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
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
	// TODO: Update
	// var allocationRation [3]uint8
	// if err := c.client.Call(func(ec *ethclient.Client) error {
	// 	auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	allocationRation, err = auction.GetAllocationRatio(nil)
	// 	return err
	// }); err != nil {
	// 	return [3]uint8{}, err
	// }
	// return allocationRation, nil
	log.Error("TODO")
	return [3]uint16{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
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
		1000000,
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

// AuctionChangeEpochMinBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
			if err != nil {
				return nil, err
			}
			slotEpochToSend := big.NewInt(slotEpoch)
			fmt.Println(slotEpochToSend)
			fmt.Println(newInitialMinBid)
			return auction.ChangeEpochMinBid(auth, slotEpochToSend, newInitialMinBid)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed changing epoch minBid: %w", err)
	}
	fmt.Println(tx)
	return tx, nil
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		1000000,
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
		1000000,
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

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error) {
	// TODO: Update
	// 	var DefaultSlotSetBid *big.Int
	// 	if err := c.client.Call(func(ec *ethclient.Client) error {
	// 		auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(c.address, ec)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		defaultSlotSetBid, err = auction.GetDefaultSlotSetBid(nil, slotSet)
	// 		return err
	// 	}); err != nil {
	// 		return big.NewInt(0), err
	// 	}
	// 	return defaultSlotSetBid, nil

	log.Error("TODO")
	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *AuctionClient) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionMultiBid(startingSlot int64, endingSlot int64, slotSet [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *AuctionClient) AuctionCanForge(forger ethCommon.Address) (bool, error) {
	log.Error("TODO")
	return false, errTODO
}

// AuctionForge is the interface to call the smart contract function
// func (c *AuctionClient) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionClaimHEZ() (*types.Transaction, error) {
	log.Error("TODO")
	return nil, errTODO
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *AuctionClient) AuctionConstants() (*AuctionConstants, error) {
	log.Error("TODO")
	return nil, errTODO
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *AuctionClient) AuctionEventsByBlock(blockNum int64) (*AuctionEvents, *ethCommon.Hash, error) {
	log.Error("TODO")
	return nil, nil, errTODO
}
