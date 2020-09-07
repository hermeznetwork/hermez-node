package test

import (
	"context"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/hermez-node/utils"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// Client implements the eth.ClientInterface interface, allowing to manipulate the
// values for testing, working with deterministic results.
type Client struct {
	log      bool
	blockNum *big.Int
}

// NewClient returns a new test Client that implements the eth.IClient
// interface, at the given initialBlockNumber.
func NewClient(l bool, initialBlockNumber int64) *Client {
	return &Client{
		log:      l,
		blockNum: big.NewInt(initialBlockNumber),
	}
}

// Advance moves one block forward
func (c *Client) Advance() {
	c.blockNum = c.blockNum.Add(c.blockNum, big.NewInt(1))
	if c.log {
		log.Debugf("TestEthClient blockNum advanced: %d", c.blockNum)
	}
}

// SetBlockNum sets the Client.blockNum to the given blockNum
func (c *Client) SetBlockNum(blockNum *big.Int) {
	c.blockNum = blockNum
	if c.log {
		log.Debugf("TestEthClient blockNum set to: %d", c.blockNum)
	}
}

// EthCurrentBlock returns the current blockNum
func (c *Client) EthCurrentBlock() (*big.Int, error) {
	return c.blockNum, nil
}

func newHeader(number *big.Int) *types.Header {
	return &types.Header{
		Number: number,
		Time:   uint64(number.Int64()),
	}
}

// EthHeaderByNumber returns the *types.Header for the given block number in a
// deterministic way.
func (c *Client) EthHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return newHeader(number), nil
}

// EthBlockByNumber returns the *common.Block for the given block number in a
// deterministic way.
func (c *Client) EthBlockByNumber(ctx context.Context, number *big.Int) (*common.Block, error) {
	header := newHeader(number)

	return &common.Block{
		EthBlockNum: uint64(number.Int64()),
		Timestamp:   time.Unix(number.Int64(), 0),
		Hash:        header.Hash(),
	}, nil
}

var errTODO = fmt.Errorf("TODO: Not implemented yet")

//
// Rollup
//

// RollupForgeBatch is the interface to call the smart contract function
func (c *Client) RollupForgeBatch(*eth.RollupForgeBatchArgs) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupAddToken is the interface to call the smart contract function
func (c *Client) RollupAddToken(tokenAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupWithdrawSNARK is the interface to call the smart contract function
// func (c *Client) RollupWithdrawSNARK() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupWithdrawMerkleProof is the interface to call the smart contract function
func (c *Client) RollupWithdrawMerkleProof(tokenID int64, balance *big.Int, babyPubKey *babyjub.PublicKey, numExitRoot int64, siblings []*big.Int, idx int64, instantWithdraw bool) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupForceExit is the interface to call the smart contract function
func (c *Client) RollupForceExit(fromIdx int64, amountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupForceTransfer is the interface to call the smart contract function
func (c *Client) RollupForceTransfer(fromIdx int64, amountF utils.Float16, tokenID, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositTransfer(babyPubKey babyjub.PublicKey, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupDepositTransfer is the interface to call the smart contract function
func (c *Client) RollupDepositTransfer(fromIdx int64, loadAmountF, amountF utils.Float16, tokenID int64, toIdx int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupDeposit is the interface to call the smart contract function
func (c *Client) RollupDeposit(fromIdx int64, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDepositFromRelayer is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDepositFromRelayer(accountCreationAuthSig []byte, babyPubKey babyjub.PublicKey, loadAmountF utils.Float16) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupCreateAccountDeposit is the interface to call the smart contract function
func (c *Client) RollupCreateAccountDeposit(babyPubKey babyjub.PublicKey, loadAmountF utils.Float16, tokenID int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupGetTokenAddress is the interface to call the smart contract function
func (c *Client) RollupGetTokenAddress(tokenID int64) (*ethCommon.Address, error) {
	return nil, errTODO
}

// RollupGetL1TxFromQueue is the interface to call the smart contract function
func (c *Client) RollupGetL1TxFromQueue(queue int64, position int64) ([]byte, error) {
	return nil, errTODO
}

// RollupGetQueue is the interface to call the smart contract function
func (c *Client) RollupGetQueue(queue int64) ([]byte, error) {
	return nil, errTODO
}

// RollupUpdateForgeL1Timeout is the interface to call the smart contract function
func (c *Client) RollupUpdateForgeL1Timeout(newForgeL1Timeout int64) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateFeeL1UserTx is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeL1UserTx(newFeeL1UserTx *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateFeeAddToken is the interface to call the smart contract function
func (c *Client) RollupUpdateFeeAddToken(newFeeAddToken *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateTokensHEZ is the interface to call the smart contract function
func (c *Client) RollupUpdateTokensHEZ(newTokenHEZ ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// RollupUpdateGovernance is the interface to call the smart contract function
// func (c *Client) RollupUpdateGovernance() (*types.Transaction, error) { // TODO (Not defined in Hermez.sol)
// 	return nil, errTODO
// }

// RollupConstants returns the Constants of the Rollup Smart Contract
func (c *Client) RollupConstants() (*eth.RollupConstants, error) {
	return nil, errTODO
}

// RollupEventsByBlock returns the events in a block that happened in the Rollup Smart Contract
func (c *Client) RollupEventsByBlock(blockNum int64) (*eth.RollupEvents, *ethCommon.Hash, error) {
	return nil, nil, errTODO
}

// RollupForgeBatchArgs returns the arguments used in a ForgeBatch call in the Rollup Smart Contract in the given transaction
func (c *Client) RollupForgeBatchArgs(transaction *types.Transaction) (*eth.RollupForgeBatchArgs, error) {
	return nil, errTODO
}

//
// Auction
//

// AuctionSetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionSetSlotDeadline(newDeadline uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetSlotDeadline is the interface to call the smart contract function
func (c *Client) AuctionGetSlotDeadline() (uint8, error) {
	return 0, errTODO
}

// AuctionSetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetOpenAuctionSlots(newOpenAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOpenAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetOpenAuctionSlots() (uint16, error) { return 0, errTODO }

// AuctionSetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionSetClosedAuctionSlots(newClosedAuctionSlots uint16) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetClosedAuctionSlots is the interface to call the smart contract function
func (c *Client) AuctionGetClosedAuctionSlots() (uint16, error) {
	return 0, errTODO
}

// AuctionSetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionSetOutbidding(newOutbidding uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetOutbidding is the interface to call the smart contract function
func (c *Client) AuctionGetOutbidding() (uint8, error) {
	return 0, errTODO
}

// AuctionSetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionSetAllocationRatio(newAllocationRatio [3]uint8) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetAllocationRatio is the interface to call the smart contract function
func (c *Client) AuctionGetAllocationRatio() ([3]uint8, error) {
	return [3]uint8{}, errTODO
}

// AuctionSetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionSetDonationAddress(newDonationAddress ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetDonationAddress is the interface to call the smart contract function
func (c *Client) AuctionGetDonationAddress() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionSetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionSetBootCoordinator(newBootCoordinator ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetBootCoordinator is the interface to call the smart contract function
func (c *Client) AuctionGetBootCoordinator() (*ethCommon.Address, error) {
	return nil, errTODO
}

// AuctionChangeEpochMinBid is the interface to call the smart contract function
func (c *Client) AuctionChangeEpochMinBid(slotEpoch int64, newInitialMinBid *big.Int) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionRegisterCoordinator is the interface to call the smart contract function
func (c *Client) AuctionRegisterCoordinator(forgerAddress ethCommon.Address, URL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionIsRegisteredCoordinator is the interface to call the smart contract function
func (c *Client) AuctionIsRegisteredCoordinator(forgerAddress ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionUpdateCoordinatorInfo is the interface to call the smart contract function
func (c *Client) AuctionUpdateCoordinatorInfo(forgerAddress ethCommon.Address, newWithdrawAddress ethCommon.Address, newURL string) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionGetCurrentSlotNumber is the interface to call the smart contract function
func (c *Client) AuctionGetCurrentSlotNumber() (int64, error) {
	return 0, errTODO
}

// AuctionGetMinBidBySlot is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidBySlot(slot int64) (*big.Int, error) {
	return nil, errTODO
}

// AuctionGetMinBidEpoch is the interface to call the smart contract function
func (c *Client) AuctionGetMinBidEpoch(epoch uint8) (*big.Int, error) {
	return nil, errTODO
}

// AuctionTokensReceived is the interface to call the smart contract function
// func (c *Client) AuctionTokensReceived(operator, from, to ethCommon.Address, amount *big.Int, userData, operatorData []byte) error {
// 	return errTODO
// }

// AuctionBid is the interface to call the smart contract function
func (c *Client) AuctionBid(slot int64, bidAmount *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionMultiBid is the interface to call the smart contract function
func (c *Client) AuctionMultiBid(startingSlot int64, endingSlot int64, slotEpoch [6]bool, maxBid, closedMinBid, budget *big.Int, forger ethCommon.Address) (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionCanForge is the interface to call the smart contract function
func (c *Client) AuctionCanForge(forger ethCommon.Address) (bool, error) {
	return false, errTODO
}

// AuctionForge is the interface to call the smart contract function
// func (c *Client) AuctionForge(forger ethCommon.Address) (bool, error) {
// 	return false, errTODO
// }

// AuctionClaimHEZ is the interface to call the smart contract function
func (c *Client) AuctionClaimHEZ() (*types.Transaction, error) {
	return nil, errTODO
}

// AuctionConstants returns the Constants of the Auction Smart Contract
func (c *Client) AuctionConstants() (*eth.AuctionConstants, error) {
	return nil, errTODO
}

// AuctionEventsByBlock returns the events in a block that happened in the Auction Smart Contract
func (c *Client) AuctionEventsByBlock(blockNum int64) (*eth.AuctionEvents, *ethCommon.Hash, error) {
	return nil, nil, errTODO
}
