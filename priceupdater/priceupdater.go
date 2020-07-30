package priceupdater

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dghubble/sling"
)

var (
	// ErrSymbolDoesNotExistInDatabase is used when trying to get a token that is not in the DB
	ErrSymbolDoesNotExistInDatabase = errors.New("symbol does not exist in database")
)

// ConfigPriceUpdater contains the configuration set by the coordinator
type ConfigPriceUpdater struct {
	RecommendedFee              uint64 // in dollars
	RecommendedCreateAccountFee uint64 // in dollars
	TokensList                  []string
	ApiUrl                      string
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
	Config ConfigPriceUpdater
	mu     sync.RWMutex
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(config ConfigPriceUpdater) PriceUpdater {

	return PriceUpdater{
		DB:     make(map[string]TokenInfo),
		Config: config,
	}

}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the MemoryDB
func (p *PriceUpdater) UpdatePrices() error {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.Config.ApiUrl).Client(httpClient)

	state := [10]float64{}

	p.mu.Lock()
	defer p.mu.Unlock()

	for ti := range p.Config.TokensList {

		resp, err := client.New().Get("ticker/t" + p.Config.TokensList[ti] + "USD").ReceiveSuccess(&state)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("Unexpected response status code: %v", resp.StatusCode)
		}

		tinfo := TokenInfo{
			Symbol:      p.Config.TokensList[ti],
			Value:       state[6],
			LastUpdated: time.Now(),
		}

		(p.DB)[tinfo.Symbol] = tinfo

	}

	return nil
}

// UpdateConfig allows to update the price-updater configuration
func (p *PriceUpdater) UpdateConfig(config ConfigPriceUpdater) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Config = config

}

// Get one token information
func (p *PriceUpdater) Get(tokenSymbol string) (TokenInfo, error) {

	var info TokenInfo

	// Check if symbol exists in database
	p.mu.RLock()
	defer p.mu.RUnlock()

	if info, ok := p.DB[tokenSymbol]; ok {
		return info, nil
	}

	return info, ErrSymbolDoesNotExistInDatabase

}

// GetPrices gets all the prices contained in the DB
func (p *PriceUpdater) GetPrices() map[string]TokenInfo {

	var info = make(map[string]TokenInfo)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for key, value := range p.DB {
		info[key] = value
	}

	return info
}

// UpdateTokenInfo updates one token info
func (p *PriceUpdater) UpdateTokenInfo(tokenInfo TokenInfo) {

	p.mu.Lock()
	defer p.mu.Unlock()

	(p.DB)[tokenInfo.Symbol] = tokenInfo

}
