package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strings"

	"github.com/arnaubennassar/eth2libp2p"
	"github.com/asaskevich/govalidator"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	auction "github.com/hermeznetwork/hermez-node/eth/contracts/auction"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"github.com/multiformats/go-multiaddr"
)

// SlotState is the state of a slot
type SlotState struct {
	Bidder           ethCommon.Address
	ForgerCommitment bool
	Fulfilled        bool
	BidAmount        *big.Int
	ClosedMinBid     *big.Int
}

// NewSlotState returns an empty SlotState
func NewSlotState() *SlotState {
	return &SlotState{
		Bidder:           ethCommon.Address{},
		Fulfilled:        false,
		ForgerCommitment: false,
		BidAmount:        big.NewInt(0),
		ClosedMinBid:     big.NewInt(0),
	}
}

// Coordinator is the details of the Coordinator identified by the forger address
type Coordinator struct {
	Forger ethCommon.Address
	URL    string
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

// AuctionEventInitialize is the InitializeHermezAuctionProtocolEvent event of
// the Smart Contract
type AuctionEventInitialize struct {
	DonationAddress        ethCommon.Address
	BootCoordinatorAddress ethCommon.Address
	BootCoordinatorURL     string
	Outbidding             uint16
	SlotDeadline           uint8
	ClosedAuctionSlots     uint16
	OpenAuctionSlots       uint16
	AllocationRatio        [3]uint16
}

// AuctionVariables returns the AuctionVariables from the initialize event
func (ei *AuctionEventInitialize) AuctionVariables(
	InitialMinimalBidding *big.Int) *common.AuctionVariables {
	return &common.AuctionVariables{
		EthBlockNum:        0,
		DonationAddress:    ei.DonationAddress,
		BootCoordinator:    ei.BootCoordinatorAddress,
		BootCoordinatorURL: ei.BootCoordinatorURL,
		DefaultSlotSetBid: [6]*big.Int{
			InitialMinimalBidding, InitialMinimalBidding, InitialMinimalBidding,
			InitialMinimalBidding, InitialMinimalBidding, InitialMinimalBidding,
		},
		DefaultSlotSetBidSlotNum: 0,
		ClosedAuctionSlots:       ei.ClosedAuctionSlots,
		OpenAuctionSlots:         ei.OpenAuctionSlots,
		AllocationRatio:          ei.AllocationRatio,
		Outbidding:               ei.Outbidding,
		SlotDeadline:             ei.SlotDeadline,
	}
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
	NewBootCoordinator    ethCommon.Address
	NewBootCoordinatorURL string
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
	AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address,
		newBootCoordinatorURL string) (*types.Transaction, error)
	AuctionGetBootCoordinator() (*ethCommon.Address, error)
	AuctionChangeDefaultSlotSetBid(slotSet int64,
		newInitialMinBid *big.Int) (*types.Transaction, error)

	// Coordinator Management
	AuctionSetCoordinator(forger ethCommon.Address, coordinatorURL string) (*types.Transaction,
		error)

	// Slot Info
	AuctionGetSlotNumber(blockNum int64) (int64, error)
	AuctionGetCurrentSlotNumber() (int64, error)
	AuctionGetMinBidBySlot(slot int64) (*big.Int, error)
	AuctionGetDefaultSlotSetBid(slotSet uint8) (*big.Int, error)
	AuctionGetSlotSet(slot int64) (*big.Int, error)

	// Bidding
	AuctionBid(amount *big.Int, slot int64, bidAmount *big.Int, deadline *big.Int) (
		tx *types.Transaction, err error)
	AuctionMultiBid(amount *big.Int, startingSlot, endingSlot int64, slotSets [6]bool,
		maxBid, minBid, deadline *big.Int) (tx *types.Transaction, err error)

	// Forge
	AuctionCanForge(forger ethCommon.Address, blockNum int64) (bool, error)
	AuctionForge(forger ethCommon.Address) (*types.Transaction, error)

	// Fees
	AuctionClaimHEZ() (*types.Transaction, error)
	AuctionGetClaimableHEZ(bidder ethCommon.Address) (*big.Int, error)

	// Smart Contract Status
	AuctionConstants() (*common.AuctionConstants, error)
	AuctionEventsByBlock(blockNum int64, blockHash *ethCommon.Hash) (*AuctionEvents, error)
	AuctionEventInit(genesisBlockNum int64) (*AuctionEventInitialize, int64, error)

	// Coordinators network
	GetCoordinatorsLibP2PAddrs() ([]multiaddr.Multiaddr, error)
}

//
// Implementation
//

// AuctionEthClient is the implementation of the interface to the Auction Smart Contract in ethereum.
type AuctionEthClient struct {
	client      *EthereumClient
	chainID     *big.Int
	address     ethCommon.Address
	auction     *auction.Auction
	token       *TokenClient
	contractAbi abi.ABI
	opts        *bind.CallOpts
}

// NewAuctionClient creates a new AuctionClient.  `tokenAddress` is the address of the HEZ tokens.
func NewAuctionClient(client *EthereumClient, address, tokenAddress ethCommon.Address) (*AuctionEthClient, error) {
	contractAbi, err :=
		abi.JSON(strings.NewReader(string(auction.AuctionABI)))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auction, err := auction.NewAuction(address, client.Client())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	token, err := NewTokenClient(client, tokenAddress)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	chainID, err := client.EthChainID()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &AuctionEthClient{
		client:      client,
		chainID:     chainID,
		address:     address,
		auction:     auction,
		token:       token,
		contractAbi: contractAbi,
		opts:        newCallOpts(),
	}, nil
}

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetSlotDeadline(auth, newDeadline)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting slotDeadline: %w", err))
	}
	return tx, nil
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetSlotDeadline() (slotDeadline uint8, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotDeadline, err = c.auction.GetSlotDeadline(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return slotDeadline, nil
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetOpenAuctionSlots(
	newOpenAuctionSlots uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetOpenAuctionSlots(auth, newOpenAuctionSlots)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting openAuctionSlots: %w", err))
	}
	return tx, nil
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetOpenAuctionSlots() (openAuctionSlots uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		openAuctionSlots, err = c.auction.GetOpenAuctionSlots(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return openAuctionSlots, nil
}

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetClosedAuctionSlots(
	newClosedAuctionSlots uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetClosedAuctionSlots(auth, newClosedAuctionSlots)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting closedAuctionSlots: %w", err))
	}
	return tx, nil
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetClosedAuctionSlots() (closedAuctionSlots uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		closedAuctionSlots, err = c.auction.GetClosedAuctionSlots(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return closedAuctionSlots, nil
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetOutbidding(newOutbidding uint16) (tx *types.Transaction,
	err error) {
	if tx, err = c.client.CallAuth(
		12500000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetOutbidding(auth, newOutbidding)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting setOutbidding: %w", err))
	}
	return tx, nil
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetOutbidding() (outbidding uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		outbidding, err = c.auction.GetOutbidding(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return outbidding, nil
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetAllocationRatio(
	newAllocationRatio [3]uint16) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetAllocationRatio(auth, newAllocationRatio)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting allocationRatio: %w", err))
	}
	return tx, nil
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetAllocationRatio() (allocationRation [3]uint16, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		allocationRation, err = c.auction.GetAllocationRatio(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return [3]uint16{}, tracerr.Wrap(err)
	}
	return allocationRation, nil
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetDonationAddress(
	newDonationAddress ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetDonationAddress(auth, newDonationAddress)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting donationAddress: %w", err))
	}
	return tx, nil
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetDonationAddress() (donationAddress *ethCommon.Address,
	err error) {
	var _donationAddress ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_donationAddress, err = c.auction.GetDonationAddress(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_donationAddress, nil
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address,
	newBootCoordinatorURL string) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetBootCoordinator(auth, newBootCoordinator,
				newBootCoordinatorURL)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed setting bootCoordinator: %w", err))
	}
	return tx, nil
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetBootCoordinator() (bootCoordinator *ethCommon.Address,
	err error) {
	var _bootCoordinator ethCommon.Address
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_bootCoordinator, err = c.auction.GetBootCoordinator(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &_bootCoordinator, nil
}

// AuctionChangeDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionChangeDefaultSlotSetBid(slotSet int64,
	newInitialMinBid *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			slotSetToSend := big.NewInt(slotSet)
			return c.auction.ChangeDefaultSlotSetBid(auth, slotSetToSend, newInitialMinBid)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed changing slotSet Bid: %w", err))
	}
	return tx, nil
}

// AuctionGetClaimableHEZ is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetClaimableHEZ(
	claimAddress ethCommon.Address) (claimableHEZ *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		claimableHEZ, err = c.auction.GetClaimableHEZ(c.opts, claimAddress)
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return claimableHEZ, nil
}

// AuctionSetCoordinator is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionSetCoordinator(forger ethCommon.Address,
	coordinatorURL string) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.SetCoordinator(auth, forger, coordinatorURL)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed set coordinator: %w", err))
	}
	return tx, nil
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetCurrentSlotNumber() (currentSlotNumber int64, err error) {
	var _currentSlotNumber *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_currentSlotNumber, err = c.auction.GetCurrentSlotNumber(c.opts)
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return _currentSlotNumber.Int64(), nil
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetMinBidBySlot(slot int64) (minBid *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotToSend := big.NewInt(slot)
		minBid, err = c.auction.GetMinBidBySlot(c.opts, slotToSend)
		return tracerr.Wrap(err)
	}); err != nil {
		return big.NewInt(0), tracerr.Wrap(err)
	}
	return minBid, nil
}

// AuctionGetSlotSet is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetSlotSet(slot int64) (slotSet *big.Int, err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		slotToSend := big.NewInt(slot)
		slotSet, err = c.auction.GetSlotSet(c.opts, slotToSend)
		return tracerr.Wrap(err)
	}); err != nil {
		return big.NewInt(0), tracerr.Wrap(err)
	}
	return slotSet, nil
}

// AuctionGetDefaultSlotSetBid is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetDefaultSlotSetBid(slotSet uint8) (minBidSlotSet *big.Int,
	err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		minBidSlotSet, err = c.auction.GetDefaultSlotSetBid(c.opts, slotSet)
		return tracerr.Wrap(err)
	}); err != nil {
		return big.NewInt(0), tracerr.Wrap(err)
	}
	return minBidSlotSet, nil
}

// AuctionGetSlotNumber is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionGetSlotNumber(blockNum int64) (slot int64, err error) {
	var _slot *big.Int
	if err := c.client.Call(func(ec *ethclient.Client) error {
		_slot, err = c.auction.GetSlotNumber(c.opts, big.NewInt(blockNum))
		return tracerr.Wrap(err)
	}); err != nil {
		return 0, tracerr.Wrap(err)
	}
	return _slot.Int64(), nil
}

// AuctionBid is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionBid(amount *big.Int, slot int64, bidAmount *big.Int,
	deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.token.hez.Nonces(c.opts, owner)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tokenName := c.token.name
			tokenAddr := c.token.address
			digest, _ := createPermitDigest(tokenAddr, owner, spender, c.chainID,
				amount, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			_slot := big.NewInt(slot)
			return c.auction.ProcessBid(auth, amount, _slot, bidAmount, permit)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed bid: %w", err))
	}
	return tx, nil
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionMultiBid(amount *big.Int, startingSlot, endingSlot int64,
	slotSets [6]bool, maxBid, minBid, deadline *big.Int) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		1000000, //nolint:gomnd
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			owner := c.client.account.Address
			spender := c.address
			nonce, err := c.token.hez.Nonces(c.opts, owner)
			if err != nil {
				return nil, tracerr.Wrap(err)
			}
			tokenName := c.token.name
			tokenAddr := c.token.address
			digest, _ := createPermitDigest(tokenAddr, owner, spender, c.chainID,
				amount, nonce, deadline, tokenName)
			signature, _ := c.client.ks.SignHash(*c.client.account, digest)
			permit := createPermit(owner, spender, amount, deadline, digest, signature)
			_startingSlot := big.NewInt(startingSlot)
			_endingSlot := big.NewInt(endingSlot)
			return c.auction.ProcessMultiBid(auth, amount, _startingSlot, _endingSlot,
				slotSets, maxBid, minBid, permit)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed multibid: %w", err))
	}
	return tx, nil
}

// AuctionCanForge is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionCanForge(forger ethCommon.Address, blockNum int64) (canForge bool,
	err error) {
	if err := c.client.Call(func(ec *ethclient.Client) error {
		canForge, err = c.auction.CanForge(c.opts, forger, big.NewInt(blockNum))
		return tracerr.Wrap(err)
	}); err != nil {
		return false, tracerr.Wrap(err)
	}
	return canForge, nil
}

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionClaimHEZ() (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.ClaimHEZ(auth)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed claim HEZ: %w", err))
	}
	return tx, nil
}

// AuctionForge is the interface to call the smart contract function
func (c *AuctionEthClient) AuctionForge(forger ethCommon.Address) (tx *types.Transaction, err error) {
	if tx, err = c.client.CallAuth(
		0,
		func(ec *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			return c.auction.Forge(auth, forger)
		},
	); err != nil {
		return nil, tracerr.Wrap(fmt.Errorf("Failed forge: %w", err))
	}
	return tx, nil
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *AuctionEthClient) AuctionConstants() (auctionConstants *common.AuctionConstants, err error) {
	auctionConstants = new(common.AuctionConstants)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auctionConstants.BlocksPerSlot, err = c.auction.BLOCKSPERSLOT(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		genesisBlock, err := c.auction.GenesisBlock(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionConstants.GenesisBlockNum = genesisBlock.Int64()
		auctionConstants.HermezRollup, err = c.auction.HermezRollup(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionConstants.InitialMinimalBidding, err =
			c.auction.INITIALMINIMALBIDDING(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionConstants.GovernanceAddress, err = c.auction.GovernanceAddress(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionConstants.TokenHEZ, err = c.auction.TokenHEZ(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		return nil
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return auctionConstants, nil
}

// AuctionVariables returns the variables of the Auction Smart Contract
func (c *AuctionEthClient) AuctionVariables() (auctionVariables *common.AuctionVariables, err error) {
	auctionVariables = new(common.AuctionVariables)
	if err := c.client.Call(func(ec *ethclient.Client) error {
		auctionVariables.AllocationRatio, err = c.AuctionGetAllocationRatio()
		if err != nil {
			return tracerr.Wrap(err)
		}
		bootCoordinator, err := c.AuctionGetBootCoordinator()
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionVariables.BootCoordinator = *bootCoordinator
		auctionVariables.BootCoordinatorURL, err = c.auction.BootCoordinatorURL(c.opts)
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionVariables.ClosedAuctionSlots, err = c.AuctionGetClosedAuctionSlots()
		if err != nil {
			return tracerr.Wrap(err)
		}
		var defaultSlotSetBid [6]*big.Int
		for i := uint8(0); i < 6; i++ {
			bid, err := c.AuctionGetDefaultSlotSetBid(i)
			if err != nil {
				return tracerr.Wrap(err)
			}
			defaultSlotSetBid[i] = bid
		}
		auctionVariables.DefaultSlotSetBid = defaultSlotSetBid
		donationAddress, err := c.AuctionGetDonationAddress()
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionVariables.DonationAddress = *donationAddress
		auctionVariables.OpenAuctionSlots, err = c.AuctionGetOpenAuctionSlots()
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionVariables.Outbidding, err = c.AuctionGetOutbidding()
		if err != nil {
			return tracerr.Wrap(err)
		}
		auctionVariables.SlotDeadline, err = c.AuctionGetSlotDeadline()
		return tracerr.Wrap(err)
	}); err != nil {
		return nil, tracerr.Wrap(err)
	}
	return auctionVariables, nil
}

var (
	logAuctionNewBid = ethCrypto.Keccak256Hash([]byte(
		"NewBid(uint128,uint128,address)"))
	logAuctionNewSlotDeadline = ethCrypto.Keccak256Hash([]byte(
		"NewSlotDeadline(uint8)"))
	logAuctionNewClosedAuctionSlots = ethCrypto.Keccak256Hash([]byte(
		"NewClosedAuctionSlots(uint16)"))
	logAuctionNewOutbidding = ethCrypto.Keccak256Hash([]byte(
		"NewOutbidding(uint16)"))
	logAuctionNewDonationAddress = ethCrypto.Keccak256Hash([]byte(
		"NewDonationAddress(address)"))
	logAuctionNewBootCoordinator = ethCrypto.Keccak256Hash([]byte(
		"NewBootCoordinator(address,string)"))
	logAuctionNewOpenAuctionSlots = ethCrypto.Keccak256Hash([]byte(
		"NewOpenAuctionSlots(uint16)"))
	logAuctionNewAllocationRatio = ethCrypto.Keccak256Hash([]byte(
		"NewAllocationRatio(uint16[3])"))
	logAuctionSetCoordinator = ethCrypto.Keccak256Hash([]byte(
		"SetCoordinator(address,address,string)"))
	logAuctionNewForgeAllocated = ethCrypto.Keccak256Hash([]byte(
		"NewForgeAllocated(address,address,uint128,uint128,uint128,uint128)"))
	logAuctionNewDefaultSlotSetBid = ethCrypto.Keccak256Hash([]byte(
		"NewDefaultSlotSetBid(uint128,uint128)"))
	logAuctionNewForge = ethCrypto.Keccak256Hash([]byte(
		"NewForge(address,uint128)"))
	logAuctionHEZClaimed = ethCrypto.Keccak256Hash([]byte(
		"HEZClaimed(address,uint128)"))
	logAuctionInitialize = ethCrypto.Keccak256Hash([]byte(
		"InitializeHermezAuctionProtocolEvent(address,address,string," +
			"uint16,uint8,uint16,uint16,uint16[3])"))
)

// AuctionEventInit returns the initialize event with its corresponding block number
func (c *AuctionEthClient) AuctionEventInit(genesisBlockNum int64) (*AuctionEventInitialize, int64, error) {
	query := ethereum.FilterQuery{
		Addresses: []ethCommon.Address{
			c.address,
		},
		FromBlock: big.NewInt(max(0, genesisBlockNum-blocksPerDay)),
		ToBlock:   big.NewInt(genesisBlockNum),
		Topics:    [][]ethCommon.Hash{{logAuctionInitialize}},
	}
	logs, err := c.client.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	if len(logs) != 1 {
		return nil, 0,
			tracerr.Wrap(fmt.Errorf("no event of type InitializeHermezAuctionProtocolEvent found"))
	}
	vLog := logs[0]
	if vLog.Topics[0] != logAuctionInitialize {
		return nil, 0, tracerr.Wrap(fmt.Errorf("event is not InitializeHermezAuctionProtocolEvent"))
	}

	var auctionInit AuctionEventInitialize
	if err := c.contractAbi.UnpackIntoInterface(&auctionInit,
		"InitializeHermezAuctionProtocolEvent", vLog.Data); err != nil {
		return nil, 0, tracerr.Wrap(err)
	}
	return &auctionInit, int64(vLog.BlockNumber), tracerr.Wrap(err)
}

// AuctionEventsByBlock returns the events in a block that happened in the
// Auction Smart Contract.
// To query by blockNum, set blockNum >= 0 and blockHash == nil.
// To query by blockHash set blockHash != nil, and blockNum will be ignored.
// If there are no events in that block the result is nil.
func (c *AuctionEthClient) AuctionEventsByBlock(blockNum int64,
	blockHash *ethCommon.Hash) (*AuctionEvents, error) {
	var auctionEvents AuctionEvents

	var blockNumBigInt *big.Int
	if blockHash == nil {
		blockNumBigInt = big.NewInt(blockNum)
	}
	query := ethereum.FilterQuery{
		BlockHash: blockHash,
		FromBlock: blockNumBigInt,
		ToBlock:   blockNumBigInt,
		Addresses: []ethCommon.Address{
			c.address,
		},
		Topics: [][]ethCommon.Hash{},
	}

	logs, err := c.client.client.FilterLogs(context.TODO(), query)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if len(logs) == 0 {
		return nil, nil
	}

	for _, vLog := range logs {
		if blockHash != nil && vLog.BlockHash != *blockHash {
			log.Errorw("Block hash mismatch", "expected", blockHash.String(), "got",
				vLog.BlockHash.String())
			return nil, tracerr.Wrap(ErrBlockHashMismatchEvent)
		}
		switch vLog.Topics[0] {
		case logAuctionNewBid:
			var auxNewBid struct {
				Slot      *big.Int
				BidAmount *big.Int
				Address   ethCommon.Address
			}
			var newBid AuctionEventNewBid
			if err := c.contractAbi.UnpackIntoInterface(&auxNewBid, "NewBid",
				vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			newBid.BidAmount = auxNewBid.BidAmount
			newBid.Slot = new(big.Int).SetBytes(vLog.Topics[1][:]).Int64()
			newBid.Bidder = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			auctionEvents.NewBid = append(auctionEvents.NewBid, newBid)
		case logAuctionNewSlotDeadline:
			var newSlotDeadline AuctionEventNewSlotDeadline
			if err := c.contractAbi.UnpackIntoInterface(&newSlotDeadline,
				"NewSlotDeadline", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			auctionEvents.NewSlotDeadline = append(auctionEvents.NewSlotDeadline, newSlotDeadline)
		case logAuctionNewClosedAuctionSlots:
			var newClosedAuctionSlots AuctionEventNewClosedAuctionSlots
			if err := c.contractAbi.UnpackIntoInterface(&newClosedAuctionSlots,
				"NewClosedAuctionSlots", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			auctionEvents.NewClosedAuctionSlots =
				append(auctionEvents.NewClosedAuctionSlots, newClosedAuctionSlots)
		case logAuctionNewOutbidding:
			var newOutbidding AuctionEventNewOutbidding
			if err := c.contractAbi.UnpackIntoInterface(&newOutbidding, "NewOutbidding",
				vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			auctionEvents.NewOutbidding = append(auctionEvents.NewOutbidding, newOutbidding)
		case logAuctionNewDonationAddress:
			var newDonationAddress AuctionEventNewDonationAddress
			newDonationAddress.NewDonationAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.NewDonationAddress = append(auctionEvents.NewDonationAddress,
				newDonationAddress)
		case logAuctionNewBootCoordinator:
			var newBootCoordinator AuctionEventNewBootCoordinator
			if err := c.contractAbi.UnpackIntoInterface(&newBootCoordinator,
				"NewBootCoordinator", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			newBootCoordinator.NewBootCoordinator = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.NewBootCoordinator = append(auctionEvents.NewBootCoordinator,
				newBootCoordinator)
		case logAuctionNewOpenAuctionSlots:
			var newOpenAuctionSlots AuctionEventNewOpenAuctionSlots
			if err := c.contractAbi.UnpackIntoInterface(&newOpenAuctionSlots,
				"NewOpenAuctionSlots", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			auctionEvents.NewOpenAuctionSlots =
				append(auctionEvents.NewOpenAuctionSlots, newOpenAuctionSlots)
		case logAuctionNewAllocationRatio:
			var newAllocationRatio AuctionEventNewAllocationRatio
			if err := c.contractAbi.UnpackIntoInterface(&newAllocationRatio,
				"NewAllocationRatio", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			auctionEvents.NewAllocationRatio = append(auctionEvents.NewAllocationRatio,
				newAllocationRatio)
		case logAuctionSetCoordinator:
			var setCoordinator AuctionEventSetCoordinator
			if err := c.contractAbi.UnpackIntoInterface(&setCoordinator,
				"SetCoordinator", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			setCoordinator.BidderAddress = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			setCoordinator.ForgerAddress = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			auctionEvents.SetCoordinator = append(auctionEvents.SetCoordinator, setCoordinator)
		case logAuctionNewForgeAllocated:
			var newForgeAllocated AuctionEventNewForgeAllocated
			if err := c.contractAbi.UnpackIntoInterface(&newForgeAllocated,
				"NewForgeAllocated", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			newForgeAllocated.Bidder = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForgeAllocated.Forger = ethCommon.BytesToAddress(vLog.Topics[2].Bytes())
			newForgeAllocated.SlotToForge = new(big.Int).SetBytes(vLog.Topics[3][:]).Int64()
			auctionEvents.NewForgeAllocated = append(auctionEvents.NewForgeAllocated,
				newForgeAllocated)
		case logAuctionNewDefaultSlotSetBid:
			var auxNewDefaultSlotSetBid struct {
				SlotSet          *big.Int
				NewInitialMinBid *big.Int
			}
			var newDefaultSlotSetBid AuctionEventNewDefaultSlotSetBid
			if err := c.contractAbi.UnpackIntoInterface(&auxNewDefaultSlotSetBid,
				"NewDefaultSlotSetBid", vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			newDefaultSlotSetBid.NewInitialMinBid = auxNewDefaultSlotSetBid.NewInitialMinBid
			newDefaultSlotSetBid.SlotSet = auxNewDefaultSlotSetBid.SlotSet.Int64()
			auctionEvents.NewDefaultSlotSetBid =
				append(auctionEvents.NewDefaultSlotSetBid, newDefaultSlotSetBid)
		case logAuctionNewForge:
			var newForge AuctionEventNewForge
			newForge.Forger = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			newForge.SlotToForge = new(big.Int).SetBytes(vLog.Topics[2][:]).Int64()
			auctionEvents.NewForge = append(auctionEvents.NewForge, newForge)
		case logAuctionHEZClaimed:
			var HEZClaimed AuctionEventHEZClaimed
			if err := c.contractAbi.UnpackIntoInterface(&HEZClaimed, "HEZClaimed",
				vLog.Data); err != nil {
				return nil, tracerr.Wrap(err)
			}
			HEZClaimed.Owner = ethCommon.BytesToAddress(vLog.Topics[1].Bytes())
			auctionEvents.HEZClaimed = append(auctionEvents.HEZClaimed, HEZClaimed)
		}
	}
	return &auctionEvents, nil
}

// GetCoordinatorsLibP2PAddrs return the libp2p addr associated to each coordinator
// that has been registered so far
func (c AuctionEthClient) GetCoordinatorsLibP2PAddrs() ([]multiaddr.Multiaddr, error) {
	// Get events
	query := ethereum.FilterQuery{
		Addresses: []ethCommon.Address{
			c.address,
		},
		Topics: [][]ethCommon.Hash{{logAuctionSetCoordinator}},
	}
	logs, err := c.client.client.FilterLogs(context.TODO(), query)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	libp2pAddrs := []multiaddr.Multiaddr{}
	for _, eventLog := range logs {
		// Get coordinator URL
		var setCoordinator AuctionEventSetCoordinator
		if err := c.contractAbi.UnpackIntoInterface(&setCoordinator,
			"SetCoordinator", eventLog.Data); err != nil {
			return nil, tracerr.Wrap(err)
		}
		url := setCoordinator.CoordinatorURL
		// Get coordinator public key
		tx, isPending, err := c.client.client.TransactionByHash(context.TODO(), eventLog.TxHash)
		if err != nil {
			log.Warn(err)
			continue
		}
		if isPending {
			continue
		}
		pubKey, err := pubKeyFromTx(tx)
		if err != nil {
			log.Warn(err)
			continue
		}

		// Generate libp2p address from URL and public key
		if addr, err := NewCoordinatorLibP2PAddr(url, pubKey); err == nil {
			libp2pAddrs = append(libp2pAddrs, addr)
		} else {
			log.Debug(err)
		}
	}
	if len(libp2pAddrs) == 0 {
		return nil, tracerr.New("Unable to generate any valid libp2p address for registered coordinators")
	}
	return libp2pAddrs, nil
}

func pubKeyFromTx(tx *types.Transaction) (*ecdsa.PublicKey, error) {
	// Get hash
	var hash ethCommon.Hash
	v, r, s := tx.RawSignatureValues()
	switch tx.Type() {
	case types.DynamicFeeTxType:
		signer := types.NewLondonSigner(tx.ChainId())
		hash = signer.Hash(tx)
	case types.LegacyTxType:
		if tx.Protected() {
			signer := types.NewEIP2930Signer(tx.ChainId())
			hash = signer.Hash(tx)
			// Special signature case
			v = new(big.Int).Sub(
				v,
				new(big.Int).Mul(tx.ChainId(), big.NewInt(2)), //nolint:gomnd
			)
			v = new(big.Int).Sub(v, big.NewInt(35)) //nolint:gomnd
		} else {
			signer := types.HomesteadSigner{}
			hash = signer.Hash(tx)
			// Special signature case
			v = new(big.Int).Sub(v, big.NewInt(27)) //nolint:gomnd
		}
	case types.AccessListTxType:
		signer := types.NewEIP2930Signer(tx.ChainId())
		hash = signer.Hash(tx)
	default:
		return nil, tracerr.New("Unexpected tx type")
	}

	// Get signature from V, R, S
	signature := make([]byte, 65) //nolint:gomnd
	rBytes, sBytes := r.Bytes(), s.Bytes()
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)
	signature[64] = byte(v.Uint64())

	// Generate public key from signature and hash to be signed
	return ethCrypto.SigToPub(hash.Bytes(), signature)
}

// NewCoordinatorLibP2PAddr returns the libp2p address associated to a coordinator
func NewCoordinatorLibP2PAddr(URL string, pubKey *ecdsa.PublicKey) (multiaddr.Multiaddr, error) {
	// Generate libp2p ID from Ethereum public key
	libp2pIDRaw, err := eth2libp2p.P2PIDFromEthPubKey(pubKey)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	libp2pID := libp2pIDRaw.Pretty()

	// Get rid of port
	u, _ := url.Parse(URL)
	host, _, err := net.SplitHostPort(u.Host)
	if err != nil && !strings.Contains(err.Error(), "missing port in address") {
		return nil, tracerr.Wrap(err)
	} else if err == nil {
		URL = host
	}
	// Get rid of / at the end of the string
	URL = strings.TrimSuffix(URL, "/")
	// IP4 case
	if ok := govalidator.IsIPv4(URL); ok {
		addr := fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", URL, common.CoordinatorsNetworkPort, libp2pID)
		return multiaddr.NewMultiaddr(addr)
	}
	// IP6 case
	if ok := govalidator.IsIPv6(URL); ok {
		addr := fmt.Sprintf("/ip6/%s/tcp/%s/p2p/%s", URL, common.CoordinatorsNetworkPort, libp2pID)
		return multiaddr.NewMultiaddr(addr)
	}
	// DNS case
	if ok := govalidator.IsURL(URL); ok {
		// Get rid of http(s)://
		URL = strings.TrimPrefix(URL, "https://")
		URL = strings.TrimPrefix(URL, "http://")
		addr := fmt.Sprintf("/dns/%s/tcp/%s/p2p/%s", URL, common.CoordinatorsNetworkPort, libp2pID)
		return multiaddr.NewMultiaddr(addr)
	}

	return nil, tracerr.New("Unexpected url format (won't be able to connect directly to this coordinator): " + URL)
}
