package etherscan

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/hermeznetwork/tracerr"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

type etherscanResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Result  GasPriceEtherscan `json:"result"`
}

// GasPriceEtherscan definition
type GasPriceEtherscan struct {
	LastBlock       string `json:"LastBlock"`
	SafeGasPrice    string `json:"SafeGasPrice"`
	ProposeGasPrice string `json:"ProposeGasPrice"`
	FastGasPrice    string `json:"FastGasPrice"`
}

// Service definition
type Service struct {
	clientEtherscan *sling.Sling
	apiKey          string
}

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Blocking.  Returns the gas price.
	GetGasPrice(ctx context.Context) (*GasPriceEtherscan, error)
}

// NewEtherscanService is the constructor that creates an etherscanService
func NewEtherscanService(etherscanURL string, apikey string) (*Service, error) {
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &Service{
		clientEtherscan: sling.New().Base(etherscanURL).Client(httpClient),
		apiKey:          apikey,
	}, nil
}

// GetGasPrice retrieves the gas price estimation from etherscan
func (p *Service) GetGasPrice(ctx context.Context) (*GasPriceEtherscan, error) {
	var resBody etherscanResponse
	url := "/api?module=gastracker&action=gasoracle&apikey=" + p.apiKey
	req, err := p.clientEtherscan.New().Get(url).Request()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	res, err := p.clientEtherscan.Do(req.WithContext(ctx), &resBody, nil)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, tracerr.Wrap(fmt.Errorf("http response is not is %v", res.StatusCode))
	}
	return &resBody.Result, nil
}

// MockEtherscanClient is a mock EtherscanServer to be used in tests.  It doesn't calculate anything
type MockEtherscanClient struct {
}

// GetGasPrice retrieves the gas price estimation from etherscan
func (p *MockEtherscanClient) GetGasPrice(ctx context.Context) (*GasPriceEtherscan, error) {
	return &GasPriceEtherscan{
			LastBlock:       "0",
			SafeGasPrice:    "90",
			ProposeGasPrice: "100",
			FastGasPrice:    "110",
		},
		nil
}
