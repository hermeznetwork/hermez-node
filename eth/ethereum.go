package eth

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethKeystore "github.com/ethereum/go-ethereum/accounts/keystore"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/eth/contracts/erc20"
	"github.com/hermeznetwork/hermez-node/log"
)

// ERC20Consts are the constants defined in a particular ERC20 Token instance
type ERC20Consts struct {
	Name     string
	Symbol   string
	Decimals uint64
}

// EthereumInterface is the interface to Ethereum
type EthereumInterface interface {
	EthCurrentBlock() (int64, error)
	// EthHeaderByNumber(context.Context, *big.Int) (*types.Header, error)
	EthBlockByNumber(context.Context, int64) (*common.Block, error)
	EthAddress() (*ethCommon.Address, error)
	EthTransactionReceipt(context.Context, ethCommon.Hash) (*types.Receipt, error)

	EthERC20Consts(ethCommon.Address) (*ERC20Consts, error)
}

var (
	// ErrAccountNil is used when the calls can not be made because the account is nil
	ErrAccountNil = fmt.Errorf("Authorized calls can't be made when the account is nil")
	// ErrReceiptStatusFailed is used when receiving a failed transaction
	ErrReceiptStatusFailed = fmt.Errorf("receipt status is failed")
	// ErrReceiptNotReceived is used when unable to retrieve a transaction
	ErrReceiptNotReceived = fmt.Errorf("receipt not available")
	// ErrBlockHashMismatchEvent is used when there's a block hash mismatch
	// beetween different events of the same block
	ErrBlockHashMismatchEvent = fmt.Errorf("block hash mismatch in event log")
)

const (
	errStrDeploy      = "deployment of %s failed: %w"
	errStrWaitReceipt = "wait receipt of %s deploy failed: %w"

	// default values
	defaultCallGasLimit        = 300000
	defaultDeployGasLimit      = 1000000
	defaultGasPriceDiv         = 100
	defaultReceiptTimeout      = 60
	defaultIntervalReceiptLoop = 200
)

// EthereumConfig defines the configuration parameters of the EthereumClient
type EthereumConfig struct {
	CallGasLimit        uint64
	DeployGasLimit      uint64
	GasPriceDiv         uint64
	ReceiptTimeout      time.Duration // in seconds
	IntervalReceiptLoop time.Duration // in milliseconds
}

// EthereumClient is an ethereum client to call Smart Contract methods and check blockchain information.
type EthereumClient struct {
	client         *ethclient.Client
	account        *accounts.Account
	ks             *ethKeystore.KeyStore
	ReceiptTimeout time.Duration
	config         *EthereumConfig
}

// NewEthereumClient creates a EthereumClient instance.  The account is not mandatory (it can
// be nil).  If the account is nil, CallAuth will fail with ErrAccountNil.
func NewEthereumClient(client *ethclient.Client, account *accounts.Account, ks *ethKeystore.KeyStore, config *EthereumConfig) *EthereumClient {
	if config == nil {
		config = &EthereumConfig{
			CallGasLimit:        defaultCallGasLimit,
			DeployGasLimit:      defaultDeployGasLimit,
			GasPriceDiv:         defaultGasPriceDiv,
			ReceiptTimeout:      defaultReceiptTimeout,
			IntervalReceiptLoop: defaultIntervalReceiptLoop,
		}
	}
	return &EthereumClient{client: client, account: account, ks: ks, ReceiptTimeout: config.ReceiptTimeout * time.Second, config: config}
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
		return nil, ErrAccountNil
	}
	return &c.account.Address, nil
}

// CallAuth performs a Smart Contract method call that requires authorization.
// This call requires a valid account with Ether that can be spend during the
// call.
func (c *EthereumClient) CallAuth(gasLimit uint64,
	fn func(*ethclient.Client, *bind.TransactOpts) (*types.Transaction, error)) (*types.Transaction, error) {
	if c.account == nil {
		return nil, ErrAccountNil
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	inc := new(big.Int).Set(gasPrice)
	inc.Div(inc, new(big.Int).SetUint64(c.config.GasPriceDiv))
	gasPrice.Add(gasPrice, inc)
	log.Debugw("Transaction metadata", "gasPrice", gasPrice)

	auth, err := bind.NewKeyStoreTransactor(c.ks, *c.account)
	if err != nil {
		return nil, err
	}
	auth.Value = big.NewInt(0) // in wei
	if gasLimit == 0 {
		auth.GasLimit = c.config.CallGasLimit // in units
	} else {
		auth.GasLimit = gasLimit // in units
	}
	auth.GasPrice = gasPrice

	tx, err := fn(c.client, auth)
	if tx != nil {
		log.Debugw("Transaction", "tx", tx.Hash().Hex(), "nonce", tx.Nonce())
	}
	return tx, err
}

// ContractData contains the contract data
type ContractData struct {
	Address ethCommon.Address
	Tx      *types.Transaction
	Receipt *types.Receipt
}

// Deploy a smart contract.  `name` is used to log deployment information.  fn
// is a wrapper to the deploy function generated by abigen.  In case of error,
// the returned `ContractData` may have some parameters filled depending on the
// kind of error that occurred.
func (c *EthereumClient) Deploy(name string,
	fn func(c *ethclient.Client, auth *bind.TransactOpts) (ethCommon.Address, *types.Transaction, interface{}, error)) (ContractData, error) {
	var contractData ContractData
	log.Infow("Deploying", "contract", name)
	tx, err := c.CallAuth(
		c.config.DeployGasLimit,
		func(client *ethclient.Client, auth *bind.TransactOpts) (*types.Transaction, error) {
			addr, tx, _, err := fn(client, auth)
			if err != nil {
				return nil, err
			}
			contractData.Address = addr
			return tx, nil
		},
	)
	if err != nil {
		return contractData, fmt.Errorf(errStrDeploy, name, err)
	}
	log.Infow("Waiting receipt", "tx", tx.Hash().Hex(), "contract", name)
	contractData.Tx = tx
	receipt, err := c.WaitReceipt(tx)
	if err != nil {
		return contractData, fmt.Errorf(errStrWaitReceipt, name, err)
	}
	contractData.Receipt = receipt
	return contractData, nil
}

// Call performs a read only Smart Contract method call.
func (c *EthereumClient) Call(fn func(*ethclient.Client) error) error {
	return fn(c.client)
}

// WaitReceipt will block until a transaction is confirmed.  Internally it
// polls the state every 200 milliseconds.
func (c *EthereumClient) WaitReceipt(tx *types.Transaction) (*types.Receipt, error) {
	return c.waitReceipt(context.TODO(), tx, c.ReceiptTimeout)
}

// GetReceipt will check if a transaction is confirmed and return
// immediately, waiting at most 1 second and returning error if the transaction
// is still pending.
func (c *EthereumClient) GetReceipt(tx *types.Transaction) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	return c.waitReceipt(ctx, tx, 0)
}

// EthTransactionReceipt returns the transaction receipt of the given txHash
func (c *EthereumClient) EthTransactionReceipt(ctx context.Context, txHash ethCommon.Hash) (*types.Receipt, error) {
	return c.client.TransactionReceipt(ctx, txHash)
}

func (c *EthereumClient) waitReceipt(ctx context.Context, tx *types.Transaction, timeout time.Duration) (*types.Receipt, error) {
	var err error
	var receipt *types.Receipt

	txHash := tx.Hash()
	log.Debugw("Waiting for receipt", "tx", txHash.Hex())

	start := time.Now()
	for {
		receipt, err = c.client.TransactionReceipt(ctx, txHash)
		if receipt != nil || time.Since(start) >= timeout {
			break
		}
		time.Sleep(c.config.IntervalReceiptLoop * time.Millisecond)
	}

	if receipt != nil && receipt.Status == types.ReceiptStatusFailed {
		log.Errorw("Failed transaction", "tx", txHash.Hex())
		return receipt, ErrReceiptStatusFailed
	}

	if receipt == nil {
		log.Debugw("Pendingtransaction / Wait receipt timeout", "tx", txHash.Hex(), "lasterr", err)
		return receipt, ErrReceiptNotReceived
	}
	log.Debugw("Successful transaction", "tx", txHash.Hex())

	return receipt, err
}

// EthCurrentBlock returns the current block number in the blockchain
func (c *EthereumClient) EthCurrentBlock() (int64, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	header, err := c.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	return header.Number.Int64(), nil
}

// EthHeaderByNumber internally calls ethclient.Client HeaderByNumber
// func (c *EthereumClient) EthHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
// 	return c.client.HeaderByNumber(ctx, number)
// }

// EthBlockByNumber internally calls ethclient.Client BlockByNumber and returns *common.Block
func (c *EthereumClient) EthBlockByNumber(ctx context.Context, number int64) (*common.Block, error) {
	blockNum := big.NewInt(number)
	if number == 0 {
		blockNum = nil
	}
	block, err := c.client.BlockByNumber(ctx, blockNum)
	if err != nil {
		return nil, err
	}
	b := &common.Block{
		EthBlockNum: block.Number().Int64(),
		Timestamp:   time.Unix(int64(block.Time()), 0),
		Hash:        block.Hash(),
	}
	return b, nil
}

// EthERC20Consts returns the constants defined for a particular ERC20 Token instance.
func (c *EthereumClient) EthERC20Consts(tokenAddress ethCommon.Address) (*ERC20Consts, error) {
	instance, err := erc20.NewERC20(tokenAddress, c.client)
	if err != nil {
		return nil, err
	}
	name, err := instance.Name(nil)
	if err != nil {
		return nil, err
	}

	symbol, err := instance.Symbol(nil)
	if err != nil {
		return nil, err
	}

	decimals, err := instance.Decimals(nil)
	if err != nil {
		return nil, err
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
