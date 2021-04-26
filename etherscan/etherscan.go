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

type EtherscanResponse struct {
	Status		string    				`json:"status"`
	Message		string 					`json:"message"`
	Result		GasPriceEtherscan   	`json:"result"`
}
type GasPriceEtherscan struct {
	LastBlock      		string    	`json:"LastBlock"`
	SafeGasPrice     	string 		`json:"SafeGasPrice"`
	ProposeGasPrice     string    	`json:"ProposeGasPrice"`
	FastGasPrice      	string    	`json:"FastGasPrice"`
}

// EtherScanService definition
type EtherScanService struct {
	clientEtherscan   *sling.Sling
}

// Client is the interface to a ServerProof that calculates zk proofs
type Client interface {
	// Blocking.  Returns the gas price.
	GetGasPrice(ctx context.Context, apiKey string) (*GasPriceEtherscan, error)
}

// NewEtherscanService is the constructor that creates an etherscanService 
func NewEtherscanService(etherscanURL string) (*EtherScanService, error) {
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &EtherScanService{
		clientEtherscan:   sling.New().Base(etherscanURL).Client(httpClient),
	}, nil
}
// GetgetGasPrice retrieves the gas price estimation from etherscan
func (p *EtherScanService) GetGasPrice(ctx context.Context, apiKey string) (*GasPriceEtherscan, error) {
	var resBody EtherscanResponse
	url := "/api?module=gastracker&action=gasoracle&apikey="+apiKey
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
func (p *MockEtherscanClient) GetGasPrice(ctx context.Context, apiKey string) (*GasPriceEtherscan, error) {
	return &GasPriceEtherscan{
			LastBlock: "0",
			SafeGasPrice: "90",
			ProposeGasPrice: "100",
			FastGasPrice: "110",
		},
		nil
}
