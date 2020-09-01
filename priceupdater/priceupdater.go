package priceupdater

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dghubble/sling"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 10
)

var (
	// ErrSymbolDoesNotExistInDatabase is used when trying to get a token that is not in the DB
	ErrSymbolDoesNotExistInDatabase = errors.New("symbol does not exist in database")
)

// Config contains the configuration set by the coordinator
type Config struct {
	RecommendedFee              uint64 // in dollars
	RecommendedCreateAccountFee uint64 // in dollars
	TokensList                  []string
	APIURL                      string
}

// TokenInfo contains the updated value for the token
type TokenInfo struct {
	Symbol      string
	Value       float64
	LastUpdated time.Time
}

// PriceUpdater definition
type PriceUpdater struct {
	DB     map[string]TokenInfo
	Config Config
	Mu     sync.RWMutex
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(config Config) *PriceUpdater {
	return &PriceUpdater{
		DB:     make(map[string]TokenInfo),
		Config: config,
	}
}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the db
func (p *PriceUpdater) UpdatePrices() error {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.Config.APIURL).Client(httpClient)

	state := [10]float64{}

	for _, tokenSymbol := range p.Config.TokensList {
		resp, err := client.New().Get("ticker/t" + tokenSymbol + "USD").ReceiveSuccess(&state)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Unexpected response status code: %v", resp.StatusCode)
		}

		tinfo := TokenInfo{
			Symbol:      tokenSymbol,
			Value:       state[6],
			LastUpdated: time.Now(),
		}

		p.UpdateTokenInfo(tinfo)
	}

	return nil
}

// UpdateConfig allows to update the price-updater configuration
func (p *PriceUpdater) UpdateConfig(config Config) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	p.Config = config
}

// Get one token information
func (p *PriceUpdater) Get(tokenSymbol string) (TokenInfo, error) {
	var info TokenInfo

	// Check if symbol exists in database
	p.Mu.RLock()
	defer p.Mu.RUnlock()

	if info, ok := p.DB[tokenSymbol]; ok {
		return info, nil
	}

	return info, ErrSymbolDoesNotExistInDatabase
}

// GetPrices gets all the prices contained in the db
func (p *PriceUpdater) GetPrices() map[string]TokenInfo {
	var info = make(map[string]TokenInfo)

	p.Mu.RLock()
	defer p.Mu.RUnlock()

	for key, value := range p.DB {
		info[key] = value
	}

	return info
}

// UpdateTokenInfo updates one token info
func (p *PriceUpdater) UpdateTokenInfo(tokenInfo TokenInfo) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	p.DB[tokenInfo.Symbol] = tokenInfo
}
