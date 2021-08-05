package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth/contracts/tokenhez"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

// ERC20Consts are the constants defined in a particular ERC20 Token instance
type ERC20Consts struct {
	Name     string
	Symbol   string
	Decimals uint64
}

// EthereumInterface is the interface to Ethereum
type EthereumInterface interface {
	EthLastBlock() (int64, error)
	// EthHeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	EthBlockByNumber(context.Context, int64) (*common.Block, error)
	EthAddress() (*ethCommon.Address, error)
	EthTransactionReceipt(context.Context, ethCommon.Hash) (*types.Receipt, error)

	EthERC20Consts(ethCommon.Address) (*ERC20Consts, error)
	EthChainID() (*big.Int, error)

	EthPendingNonceAt(ctx context.Context, account ethCommon.Address) (uint64, error)
	EthNonceAt(ctx context.Context, account ethCommon.Address, blockNumber *big.Int) (uint64, error)
	EthSuggestGasPrice(ctx context.Context) (*big.Int, error)
	EthKeyStore() *ethKeystore.KeyStore
	EthCall(ctx context.Context, tx *types.Transaction, blockNum *big.Int) ([]byte, error)
}

var (
	// ErrAccountNil is used when the calls can not be made because the account is nil
	ErrAccountNil = fmt.Errorf("Authorized calls can't be made when the account is nil")
	// ErrBlockHashMismatchEvent is used when there's a block hash mismatch
	// between different events of the same block
	ErrBlockHashMismatchEvent = fmt.Errorf("block hash mismatch in event log")
)

const (
	// default values
	defaultCallGasLimit = 300000
	defaultGasPriceDiv  = 100
)

// EthereumConfig defines the configuration parameters of the EthereumClient
type EthereumConfig struct {
	CallGasLimit uint64
	GasPriceDiv  uint64
}

// EthereumClient is an ethereum client to call Smart Contract methods and check blockchain
// information.
type EthereumClient struct {
	client  *ethclient.Client
	chainID *big.Int
	account *accounts.Account
	ks      *ethKeystore.KeyStore
	config  *EthereumConfig
	opts    *bind.CallOpts
}

// NewEthereumClient creates a EthereumClient instance.  The account is not mandatory (it can
// be nil).  If the account is nil, CallAuth will fail with ErrAccountNil.
func NewEthereumClient(client *ethclient.Client, account *accounts.Account,
	ks *ethKeystore.KeyStore, config *EthereumConfig) (*EthereumClient, error) {
	if config == nil {
		config = &EthereumConfig{
			CallGasLimit: defaultCallGasLimit,
			GasPriceDiv:  defaultGasPriceDiv,
		}
	}
	c := &EthereumClient{
		client:  client,
		account: account,
		ks:      ks,
		config:  config,
		opts:    newCallOpts(),
	}
	chainID, err := c.EthChainID()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	c.chainID = chainID
	return c, nil
}

// EthChainID returns the ChainID of the ethereum network
func (c *EthereumClient) EthChainID() (*big.Int, error) {
	chainID, err := c.client.ChainID(context.Background())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return chainID, nil
}

// BalanceAt retieves information about the default account
func (c *EthereumClient) BalanceAt(addr ethCommon.Address) (*big.Int, error) {
	return c.client.BalanceAt(context.TODO(), addr, nil)
}

// Account returns the underlying ethereum account
func (c *EthereumClient) Account() *accounts.Account {
	return c.account
}

// EthAddress returns the ethereum address of the account loaded into the EthereumClient
func (c *EthereumClient) EthAddress() (*ethCommon.Address, error) {
	if c.account == nil {
		return nil, tracerr.Wrap(ErrAccountNil)
	}
	return &c.account.Address, nil
}

// EthSuggestGasPrice retrieves the currently suggested gas price to allow a
// timely execution of a transaction.
func (c *EthereumClient) EthSuggestGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	var head *types.Header
	head, err = c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		err = fmt.Errorf("[EthSuggestGasPrice]. Error getting head: %s", err.Error())
		return
	}
	var tip *big.Int
	tip, err = c.client.SuggestGasTipCap(ctx)
	if err != nil {
		err = fmt.Errorf("[EthSuggestGasPrice]. Error getting tip: %s", err.Error())
		return
	}
	baseFee := head.BaseFee
	gasPrice = new(big.Int).Add(baseFee, tip)
	log.Debugw("Suggested Gas Price:", "tip", tip, "baseFee", baseFee, "gasPrice", gasPrice)
	return
}

// EthKeyStore returns the keystore in the EthereumClient
func (c *EthereumClient) EthKeyStore() *ethKeystore.KeyStore {
	return c.ks
}

// NewAuth builds a new auth object to make a transaction
func (c *EthereumClient) NewAuth() (*bind.TransactOpts, error) {
	if c.account == nil {
		return nil, tracerr.Wrap(ErrAccountNil)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	inc := new(big.Int).Set(gasPrice)
	inc.Div(inc, new(big.Int).SetUint64(c.config.GasPriceDiv))
	gasPrice.Add(gasPrice, inc)
	log.Debugw("Transaction metadata", "gasPrice", gasPrice)

	auth, err := bind.NewKeyStoreTransactorWithChainID(c.ks, *c.account, c.chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auth.Value = big.NewInt(0) // in wei
	auth.GasLimit = c.config.CallGasLimit
	auth.GasPrice = gasPrice
	auth.Context = context.Background()

	return auth, nil
}

// CallAuth performs a Smart Contract method call that requires authorization.
// This call requires a valid account with Ether that can be spend during the
// call.
func (c *EthereumClient) CallAuth(gasLimit uint64, fn func(*ethclient.Client, *bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction,
	error) {
	if c.account == nil {
		return nil, tracerr.Wrap(ErrAccountNil)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	inc := new(big.Int).Set(gasPrice)
	inc.Div(inc, new(big.Int).SetUint64(c.config.GasPriceDiv))
	gasPrice.Add(gasPrice, inc)
	log.Debugw("Transaction metadata", "gasPrice", gasPrice)

	auth, err := bind.NewKeyStoreTransactorWithChainID(c.ks, *c.account, c.chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	auth.Value = big.NewInt(0) // in wei
	if gasLimit == 0 {
		auth.GasLimit = c.config.CallGasLimit // in units
	} else {
		auth.GasLimit = gasLimit // in units
	}
	auth.GasPrice = gasPrice
	auth.Context = context.Background()

	tx, err := fn(c.client, auth)
	if tx != nil {
		log.Debugw("Transaction", "tx", tx.Hash().Hex(), "nonce", tx.Nonce())
	}
	return tx, tracerr.Wrap(err)
}

// ContractData contains the contract data
type ContractData struct {
	Address ethCommon.Address
	Tx      *types.Transaction
	Receipt *types.Receipt
}

// Call performs a read only Smart Contract method call.
func (c *EthereumClient) Call(fn func(*ethclient.Client) error) error {
	return fn(c.client)
}

// EthTransactionReceipt returns the transaction receipt of the given txHash
func (c *EthereumClient) EthTransactionReceipt(ctx context.Context,
	txHash ethCommon.Hash) (*types.Receipt, error) {
	return c.client.TransactionReceipt(ctx, txHash)
}

// EthLastBlock returns the last block number in the blockchain
func (c *EthereumClient) EthLastBlock() (int64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	return header.Number.Int64(), nil
}

// EthHeaderByNumber internally calls ethclient.Client HeaderByNumber
// func (c *EthereumClient) EthHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header,
// 	error) {
// 	return c.client.HeaderByNumber(ctx, number)
// }

// EthBlockByNumber internally calls ethclient.Client BlockByNumber and returns
// *common.Block.  If number == -1, the latests known block is returned.
func (c *EthereumClient) EthBlockByNumber(ctx context.Context, number int64) (*common.Block,
	error) {
	blockNum := big.NewInt(number)
	if number == -1 {
		blockNum = nil
	}
	header, err := c.client.HeaderByNumber(ctx, blockNum)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	b := &common.Block{
		Num:        header.Number.Int64(),
		Timestamp:  time.Unix(int64(header.Time), 0),
		ParentHash: header.ParentHash,
		Hash:       header.Hash(),
	}
	return b, nil
}

// EthERC20Consts returns the constants defined for a particular ERC20 Token instance.
func (c *EthereumClient) EthERC20Consts(tokenAddress ethCommon.Address) (*ERC20Consts, error) {
	// We use the HEZ token smart contract interfacehere because it's an
	// ERC20, which allows us to access the standard ERC20 constants.
	instance, err := tokenhez.NewTokenhez(tokenAddress, c.client)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	name, err := instance.Name(c.opts)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	symbol, err := instance.Symbol(c.opts)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	decimals, err := instance.Decimals(c.opts)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	return &ERC20Consts{
		Name:     name,
		Symbol:   symbol,
		Decimals: uint64(decimals),
	}, nil
}

// Client returns the internal ethclient.Client
func (c *EthereumClient) Client() *ethclient.Client {
	return c.client
}

// newCallOpts returns a CallOpts to be used in ethereum calls with a non-zero
// From address.  This is a workaround for a bug in ethereumjs-vm that shows up
// in ganache: https://github.com/hermeznetwork/hermez-node/issues/317
func newCallOpts() *bind.CallOpts {
	return &bind.CallOpts{
		From: ethCommon.HexToAddress("0x0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f0f"),
	}
}

// EthPendingNonceAt returns the account nonce of the given account in the pending
// state. This is the nonce that should be used for the next transaction.
func (c *EthereumClient) EthPendingNonceAt(ctx context.Context,
	account ethCommon.Address) (uint64, error) {
	return c.client.PendingNonceAt(ctx, account)
}

// EthNonceAt returns the account nonce of the given account. The block number can
// be nil, in which case the nonce is taken from the latest known block.
func (c *EthereumClient) EthNonceAt(ctx context.Context,
	account ethCommon.Address, blockNumber *big.Int) (uint64, error) {
	return c.client.NonceAt(ctx, account, blockNumber)
}

// EthCall runs the transaction as a call (without paying) in the local node at
// blockNum.
func (c *EthereumClient) EthCall(ctx context.Context, tx *types.Transaction,
	blockNum *big.Int) ([]byte, error) {
	if c.account == nil {
		return nil, tracerr.Wrap(ErrAccountNil)
	}
	msg := ethereum.CallMsg{
		From:     c.account.Address,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}
	result, err := c.client.CallContract(ctx, msg, blockNum)
	return result, tracerr.Wrap(err)
}
