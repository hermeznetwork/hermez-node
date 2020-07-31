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
	db     map[string]TokenInfo
	config ConfigPriceUpdater
	mu     sync.RWMutex
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(config ConfigPriceUpdater) PriceUpdater {

	return PriceUpdater{
		db:     make(map[string]TokenInfo),
		config: config,
	}

}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the db
func (p *PriceUpdater) UpdatePrices() error {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.config.APIURL).Client(httpClient)

	state := [10]float64{}

	for _, tokenSymbol := range p.config.TokensList {

		resp, err := client.New().Get("ticker/t" + tokenSymbol + "USD").ReceiveSuccess(&state)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
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
func (p *PriceUpdater) UpdateConfig(config ConfigPriceUpdater) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.config = config

}

// Get one token information
func (p *PriceUpdater) Get(tokenSymbol string) (TokenInfo, error) {

	var info TokenInfo

	// Check if symbol exists in database
	p.mu.RLock()
	defer p.mu.RUnlock()

	if info, ok := p.db[tokenSymbol]; ok {
		return info, nil
	}

	return info, ErrSymbolDoesNotExistInDatabase

}

// GetPrices gets all the prices contained in the db
func (p *PriceUpdater) GetPrices() map[string]TokenInfo {

	var info = make(map[string]TokenInfo)

	p.mu.RLock()
	defer p.mu.RUnlock()

	for key, value := range p.db {
		info[key] = value
	}

	return info
}

// UpdateTokenInfo updates one token info
func (p *PriceUpdater) UpdateTokenInfo(tokenInfo TokenInfo) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.db[tokenInfo.Symbol] = tokenInfo

}
