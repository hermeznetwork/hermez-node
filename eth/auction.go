package eth

import (
	"context"
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
	HEZ "github.com/hermeznetwork/hermez-node/eth/contracts/tokenHEZ"
)

// AuctionConstants are the constants of the Rollup Smart Contract
type AuctionConstants struct {
	// Blocks per slot
	BlocksPerSlot uint8 `json:"blocksPerSlot"`
	// Minimum bid when no one has bid yet
	InitialMinimalBidding *big.Int `json:"initialMinimalBidding"`
	// First block where the first slot begins
	GenesisBlockNum int64 `json:"genesisBlockNum"`
	// ERC777 token with which the bids will be made
	TokenHEZ ethCommon.Address `json:"tokenHEZ"`
	// HermezRollup smartcontract address
	HermezRollup ethCommon.Address `json:"hermezRollup"`
	// Hermez Governanze Token smartcontract address who controls some parameters and collects HEZ fee
	// Only for test
	GovernanceAddress ethCommon.Address `json:"governanceAddress"`
}

// SlotState is the state of a slot
type SlotState struct {
	Bidder       ethCommon.Address
	Fulfilled    bool
	BidAmount    *big.Int
	ClosedMinBid *big.Int
}

// NewSlotState returns an empty SlotState
func NewSlotState() *SlotState {
	return &SlotState{
		Bidder:       ethCommon.Address{},
		Fulfilled:    false,
		BidAmount:    big.NewInt(0),
		ClosedMinBid: big.NewInt(0),
	}
}

// Coordinator is the details of the Coordinator identified by the forger address
type Coordinator struct {
	Forger ethCommon.Address
	URL    string
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
	Slot      int64
	BidAmount *big.Int
	Bidder    ethCommon.Address
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

// AuctionEventSetCoordinator is an event of the Auction Smart Contract
type AuctionEventSetCoordinator struct {
	BidderAddress  ethCommon.Address
	ForgerAddress  ethCommon.Address
	CoordinatorURL string
}

// AuctionEventNewForgeAllocated is an event of the Auction Smart Contract
type AuctionEventNewForgeAllocated struct {
	Bidder           ethCommon.Address
	Forger           ethCommon.Address
	SlotToForge      int64
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
	SlotToForge int64
}

// AuctionEventHEZClaimed is an event of the Auction Smart Contract
type AuctionEventHEZClaimed struct {
	Owner  ethCommon.Address
	Amount *big.Int
}

// AuctionEvents is the list of events in a block of the Auction Smart Contract
type AuctionEvents struct {
	NewBid                []AuctionEventNewBid
	NewSlotDeadline       []AuctionEventNewSlotDeadline
	NewClosedAuctionSlots []AuctionEventNewClosedAuctionSlots
	NewOutbidding         []AuctionEventNewOutbidding
	NewDonationAddress    []AuctionEventNewDonationAddress
	NewBootCoordinator    []AuctionEventNewBootCoordinator
	NewOpenAuctionSlots   []AuctionEventNewOpenAuctionSlots
	NewAllocationRatio    []AuctionEventNewAllocationRatio
	SetCoordinator        []AuctionEventSetCoordinator
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
		SetCoordinator:        make([]AuctionEventSetCoordinator, 0),
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
	AuctionSetCoordinator(forger ethCommon.Address, coordinatorURL string) (*types.Transaction, error)

	// Slot Info
	AuctionGetSlotNumber(blockNum int64) (int64, error)
	AuctionGetCurrentSlotNumber() (int64, error)
	AuctionGetMinBidBySlot(slot int64) (*big.Int, error)
	AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error)
	AuctionGetSlotSet(slot int64) (*big.Int, error)

	// Bidding
	AuctionBid(amount *big.Int, slot int64, bidAmount *big.Int, deadline *big.Int) (tx *types.Transaction, err error)
	AuctionMultiBid(amount *big.Int, startingSlot, endingSlot int64, slotSets [6]bool,
		maxBid, minBid, deadline *big.Int) (tx *types.Transaction, err error)

	// Forge
	AuctionCanForge(forger ethCommon.Address, blockNum int64) (bool, error)
	AuctionForge(forger ethCommon.Address) (*types.Transaction, error)

	// Fees
	AuctionClaimHEZ() (*types.Transaction, error)
	AuctionGetClaimableHEZ(bidder ethCommon.Address) (*big.Int, error)

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
	client      *EthereumClient
	address     ethCommon.Address
	tokenHEZ    TokenConfig
	auction     *HermezAuctionProtocol.HermezAuctionProtocol
	contractAbi abi.ABI
}

// NewAuctionClient creates a new AuctionClient.  `tokenAddress` is the address of the HEZ tokens.
func NewAuctionClient(client *EthereumClient, address ethCommon.Address, tokenHEZ TokenConfig) (*AuctionClient, error) {
	contractAbi, err := abi.JSON(strings.NewReader(string(HermezAuctionProtocol.HermezAuctionProtocolABI)))
	if err != nil {
		return nil, err
	}
	auction, err := HermezAuctionProtocol.NewHermezAuctionProtocol(address, client.Client())
	if err != nil {
		return nil, err
	}
	return &AuctionClient{
		client:      client,
		address:     address,
		tokenHEZ:    tokenHEZ,
		auction:     auction,
		contractAbi: contractAbi,
	}, nil
}

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetSlotDeadline(auth, newDeadline)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting slotDeadline: %w", err)
	}
	return tx, nil
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotDeadline() (slotDeadline uint8, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotDeadline, err = c.auction.GetSlotDeadline(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return slotDeadline, nil
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetOpenAuctionSlots(auth, newOpenAuctionSlots)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting openAuctionSlots: %w", err)
	}
	return tx, nil
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOpenAuctionSlots() (openAuctionSlots uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		openAuctionSlots, err = c.auction.GetOpenAuctionSlots(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return openAuctionSlots, nil
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetClosedAuctionSlots(auth, newClosedAuctionSlots)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting closedAuctionSlots: %w", err)
	}
	return tx, nil
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetClosedAuctionSlots() (closedAuctionSlots uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		closedAuctionSlots, err = c.auction.GetClosedAuctionSlots(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return closedAuctionSlots, nil
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetOutbidding(newOutbidding uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		12500000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetOutbidding(auth, newOutbidding)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting setOutbidding: %w", err)
	}
	return tx, nil
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetOutbidding() (outbidding uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		outbidding, err = c.auction.GetOutbidding(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return outbidding, nil
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetAllocationRatio(newAllocationRatio [3]uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetAllocationRatio(auth, newAllocationRatio)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting allocationRatio: %w", err)
	}
	return tx, nil
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetAllocationRatio() (allocationRation [3]uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		allocationRation, err = c.auction.GetAllocationRatio(nil)
		return err
	}); err != nil {
		return [3]uint16{}, err
	}
	return allocationRation, nil
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetDonationAddress(auth, newDonationAddress)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting donationAddress: %w", err)
	}
	return tx, nil
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDonationAddress() (donationAddress *ethCommon.Address, err error) {
	var _donationAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_donationAddress, err = c.auction.GetDonationAddress(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &_donationAddress, nil
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetBootCoordinator(auth, newBootCoordinator)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed setting bootCoordinator: %w", err)
	}
	return tx, nil
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetBootCoordinator() (bootCoordinator *ethCommon.Address, err error) {
	var _bootCoordinator ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_bootCoordinator, err = c.auction.GetBootCoordinator(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return &_bootCoordinator, nil
}

// AuctionChangeDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionChangeDefaultSlotSetBid(slotSet int64, newInitialMinBid *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			slotSetToSend := big.NewInt(slotSet)
			return c.auction.ChangeDefaultSlotSetBid(auth, slotSetToSend, newInitialMinBid)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed changing slotSet Bid: %w", err)
	}
	return tx, nil
}

// AuctionGetClaimableHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetClaimableHEZ(claimAddress ethCommon.Address) (claimableHEZ *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		claimableHEZ, err = c.auction.GetClaimableHEZ(nil, claimAddress)
		return err
	}); err != nil {
		return nil, err
	}
	return claimableHEZ, nil
}

// AuctionSetCoordinator is the interface to call the smart contract function
func (c *AuctionClient) AuctionSetCoordinator(forger ethCommon.Address, coordinatorURL string) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetCoordinator(auth, forger, coordinatorURL)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed set coordinator: %w", err)
	}
	return tx, nil
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetCurrentSlotNumber() (currentSlotNumber int64, err error) {
	var _currentSlotNumber *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_currentSlotNumber, err = c.auction.GetCurrentSlotNumber(nil)
		return err
	}); err != nil {
		return 0, err
	}
	return _currentSlotNumber.Int64(), nil
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetMinBidBySlot(slot int64) (minBid *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotToSend := big.NewInt(slot)
		minBid, err = c.auction.GetMinBidBySlot(nil, slotToSend)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return minBid, nil
}

// AuctionGetSlotSet is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotSet(slot int64) (slotSet *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotToSend := big.NewInt(slot)
		slotSet, err = c.auction.GetSlotSet(nil, slotToSend)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return slotSet, nil
}

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetDefaultSlotSetBid(slotSet uint8) (minBidSlotSet *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		minBidSlotSet, err = c.auction.GetDefaultSlotSetBid(nil, slotSet)
		return err
	}); err != nil {
		return big.NewInt(0), err
	}
	return minBidSlotSet, nil
}

// AuctionGetSlotNumber is the interface to call the smart contract function
func (c *AuctionClient) AuctionGetSlotNumber(blockNum int64) (slot int64, err error) {
	var _slot *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_slot, err = c.auction.GetSlotNumber(nil, big.NewInt(blockNum))
		return err
	}); err != nil {
		return 0, err
	}
	return _slot.Int64(), nil
}

// AuctionBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionBid(amount *big.Int, slot int64, bidAmount *big.Int, deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			tokenHEZcontract, err := HEZ.NewHEZ(c.tokenHEZ.Address, ec)
			if err != nil {
				return nil, err
			}
			owner := c.client.account.Address
			spender := c.address
			nonce, err := tokenHEZcontract.Nonces(nil, owner)
			tokenname := c.tokenHEZ.Name
			tokenAddr := c.tokenHEZ.Address
			chainid, _ := c.client.client.ChainID(context.Background())
			digest, _ := createPermitDigest(tokenAddr, owner, spender, chainid, amount, nonce, deadline, tokenname)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			_slot := big.NewInt(slot)
			return c.auction.ProcessBid(auth, amount, _slot, bidAmount, permit)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed bid: %w", err)
	}
	return tx, nil

}

// AuctionMultiBid is the interface to call the smart contract function
func (c *AuctionClient) AuctionMultiBid(amount *big.Int, startingSlot, endingSlot int64, slotSets [6]bool,
	maxBid, minBid, deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		1000000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			tokenHEZcontract, err := HEZ.NewHEZ(c.tokenHEZ.Address, ec)
			if err != nil {
				return nil, err
			}
			owner := c.client.account.Address
			spender := c.address
			nonce, err := tokenHEZcontract.Nonces(nil, owner)
			tokenname := c.tokenHEZ.Name
			tokenAddr := c.tokenHEZ.Address
			chainid, _ := c.client.client.ChainID(context.Background())

			digest, _ := createPermitDigest(tokenAddr, owner, spender, chainid, amount, nonce, deadline, tokenname)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			_startingSlot := big.NewInt(startingSlot)
			_endingSlot := big.NewInt(endingSlot)
			return c.auction.ProcessMultiBid(auth, amount, _startingSlot, _endingSlot, slotSets, maxBid, minBid, permit)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed multibid: %w", err)
	}
	return tx, nil
}

// AuctionCanForge is the interface to call the smart contract function
func (c *AuctionClient) AuctionCanForge(forger ethCommon.Address, blockNum int64) (canForge bool, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		canForge, err = c.auction.CanForge(nil, forger, big.NewInt(blockNum))
		return err
	}); err != nil {
		return false, err
	}
	return canForge, nil
}

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *AuctionClient) AuctionClaimHEZ() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.ClaimHEZ(auth)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed claim HEZ: %w", err)
	}
	return tx, nil
}

// AuctionForge is the interface to call the smart contract function
func (c *AuctionClient) AuctionForge(forger ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.Forge(auth, forger)
		},
	); err != nil {
		return nil, fmt.Errorf("Failed forge: %w", err)
	}
	return tx, nil
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *AuctionClient) AuctionConstants() (auctionConstants *AuctionConstants, err error) {
	auctionConstants = new(AuctionConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auctionConstants.BlocksPerSlot, err = c.auction.BLOCKSPERSLOT(nil)
		if err != nil {
			return err
		}
		genesisBlock, err := c.auction.GenesisBlock(nil)
		if err != nil {
			return err
		}
		auctionConstants.GenesisBlockNum = genesisBlock.Int64()
		auctionConstants.HermezRollup, err = c.auction.HermezRollup(nil)
		if err != nil {
			return err
		}
		auctionConstants.InitialMinimalBidding, err = c.auction.INITIALMINIMALBIDDING(nil)
		if err != nil {
			return err
		}
		auctionConstants.TokenHEZ, err = c.auction.TokenHEZ(nil)
		return err
	}); err != nil {
		return nil, err
	}
	return auctionConstants, nil
}

// AuctionVariables returns the variables of the Auction Smart Contract
func (c *AuctionClient) AuctionVariables() (auctionVariables *AuctionVariables, err error) {
	auctionVariables = new(AuctionVariables)
	if err := c.client.Call(func(ec *ethclient.Client) error {
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
	logAuctionNewBid                = crypto.Keccak256Hash([]byte("NewBid(uint128,uint128,address)"))
	logAuctionNewSlotDeadline       = crypto.Keccak256Hash([]byte("NewSlotDeadline(uint8)"))
	logAuctionNewClosedAuctionSlots = crypto.Keccak256Hash([]byte("NewClosedAuctionSlots(uint16)"))
	logAuctionNewOutbidding         = crypto.Keccak256Hash([]byte("NewOutbidding(uint16)"))
	logAuctionNewDonationAddress    = crypto.Keccak256Hash([]byte("NewDonationAddress(address)"))
	logAuctionNewBootCoordinator    = crypto.Keccak256Hash([]byte("NewBootCoordinator(address)"))
	logAuctionNewOpenAuctionSlots   = crypto.Keccak256Hash([]byte("NewOpenAuctionSlots(uint16)"))
	logAuctionNewAllocationRatio    = crypto.Keccak256Hash([]byte("NewAllocationRatio(uint16[3])"))
	logAuctionSetCoordinator        = crypto.Keccak256Hash([]byte("SetCoordinator(address,address,string)"))
	logAuctionNewForgeAllocated     = crypto.Keccak256Hash([]byte("NewForgeAllocated(address,address,uint128,uint128,uint128,uint128)"))
	logAuctionNewDefaultSlotSetBid  = crypto.Keccak256Hash([]byte("NewDefaultSlotSetBid(uint128,uint128)"))
	logAuctionNewForge              = crypto.Keccak256Hash([]byte("NewForge(address,uint128)"))
	logAuctionHEZClaimed            = crypto.Keccak256Hash([]byte("HEZClaimed(address,uint128)"))
)

// AuctionEventsByBlock returns the events in a block that happened in the
// Auction Smart Contract and the blockHash where the eents happened.  If there
// are no events in that block, blockHash is nil.
func (c *AuctionClient) AuctionEventsByBlock(blockNum int64) (*AuctionEvents, *ethCommon.Hash, error) {
	var auctionEvents AuctionEvents
	var blockHash ethCommon.Hash

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(blockNum),
		ToBlock:   big.NewInt(blockNum),
		Addresses: []ethCommon.Address{
			c.address,
		},
		Topics: [][]ethCommon.Hash{},
	}

	logs, err := c.client.client.FilterLogs(context.TODO(), query)
	if err != nil {
		return nil, nil, err
	}
	if len(logs) > 0 {
		blockHash = logs[0].BlockHash
	}
	for _, vLog := range logs {
		if vLog.BlockHash != blockHash {
			return nil, nil, ErrBlockHashMismatchEvent
		}
		switch vLog.Topics[0] {
		case logAuctionNewBid:
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
			newBid.Bidder = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			auctionEvents.NewBid = append(auctionEvents.NewBid, newBid)
		case logAuctionNewSlotDeadline:
			var newSlotDeadline AuctionEventNewSlotDeadline
			if err := c.contractAbi.Unpack(&newSlotDeadline, "NewSlotDeadline", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewSlotDeadline = append(auctionEvents.NewSlotDeadline, newSlotDeadline)
		case logAuctionNewClosedAuctionSlots:
			var newClosedAuctionSlots AuctionEventNewClosedAuctionSlots
			if err := c.contractAbi.Unpack(&newClosedAuctionSlots, "NewClosedAuctionSlots", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewClosedAuctionSlots = append(auctionEvents.NewClosedAuctionSlots, newClosedAuctionSlots)
		case logAuctionNewOutbidding:
			var newOutbidding AuctionEventNewOutbidding
			if err := c.contractAbi.Unpack(&newOutbidding, "NewOutbidding", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewOutbidding = append(auctionEvents.NewOutbidding, newOutbidding)
		case logAuctionNewDonationAddress:
			var newDonationAddress AuctionEventNewDonationAddress
			newDonationAddress.NewDonationAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.NewDonationAddress = append(auctionEvents.NewDonationAddress, newDonationAddress)
		case logAuctionNewBootCoordinator:
			var newBootCoordinator AuctionEventNewBootCoordinator
			newBootCoordinator.NewBootCoordinator = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.NewBootCoordinator = append(auctionEvents.NewBootCoordinator, newBootCoordinator)
		case logAuctionNewOpenAuctionSlots:
			var newOpenAuctionSlots AuctionEventNewOpenAuctionSlots
			if err := c.contractAbi.Unpack(&newOpenAuctionSlots, "NewOpenAuctionSlots", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewOpenAuctionSlots = append(auctionEvents.NewOpenAuctionSlots, newOpenAuctionSlots)
		case logAuctionNewAllocationRatio:
			var newAllocationRatio AuctionEventNewAllocationRatio
			if err := c.contractAbi.Unpack(&newAllocationRatio, "NewAllocationRatio", vLog.Data); err != nil {
				return nil, nil, err
			}
			auctionEvents.NewAllocationRatio = append(auctionEvents.NewAllocationRatio, newAllocationRatio)
		case logAuctionSetCoordinator:
			var setCoordinator AuctionEventSetCoordinator
			if err := c.contractAbi.Unpack(&setCoordinator, "SetCoordinator", vLog.Data); err != nil {
				return nil, nil, err
			}
			setCoordinator.BidderAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			setCoordinator.ForgerAddress = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			auctionEvents.SetCoordinator = append(auctionEvents.SetCoordinator, setCoordinator)
		case logAuctionNewForgeAllocated:
			var newForgeAllocated AuctionEventNewForgeAllocated
			if err := c.contractAbi.Unpack(&newForgeAllocated, "NewForgeAllocated", vLog.Data); err != nil {
				return nil, nil, err
			}
			newForgeAllocated.Bidder = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForgeAllocated.Forger = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			newForgeAllocated.SlotToForge = new(big.Int).SetBytes(vLog.Topics[3][:]).Int64()
			auctionEvents.NewForgeAllocated = append(auctionEvents.NewForgeAllocated, newForgeAllocated)
		case logAuctionNewDefaultSlotSetBid:
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
		case logAuctionNewForge:
			var newForge AuctionEventNewForge
			newForge.Forger = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForge.SlotToForge = new(big.Int).SetBytes(vLog.Topics[2][:]).Int64()
			auctionEvents.NewForge = append(auctionEvents.NewForge, newForge)
		case logAuctionHEZClaimed:
			var HEZClaimed AuctionEventHEZClaimed
			if err := c.contractAbi.Unpack(&HEZClaimed, "HEZClaimed", vLog.Data); err != nil {
				return nil, nil, err
			}
			HEZClaimed.Owner = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.HEZClaimed = append(auctionEvents.HEZClaimed, HEZClaimed)
		}
	}
	return &auctionEvents, &blockHash, nil
}
