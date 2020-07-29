package priceupdater

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
)

const ApiUrl = "https://api-pub.bitfinex.com/v2/"

// ConfigPriceUpdater contains the configuration set by the coordinator
type ConfigPriceUpdater struct {
	RecommendedFee              uint64 // in dollars
	RecommendedCreateAccountFee uint64 // in dollars
	TokensList                  []string
}

// RecommendedFee is the struct that will be sent to FE by the API
type RecommendedFee struct {
	// all in $
	ExistingAccount float64
	CreatesAccount  float64
}

// TokenInfo contains the updated value for the token
type TokenInfo struct {
	Symbol      string
	Value       float64
	LastUpdated time.Time
}

// MemoryDB is a Key Value DB
type MemoryDB map[string]TokenInfo

// Get one token information
func (m *MemoryDB) Get(tokenSymbol string) (TokenInfo, error) {

	return (*m)[tokenSymbol], nil

}

// GetPrices gets all the prices contained in the DB
func (m *MemoryDB) GetPrices() (map[string]TokenInfo, error) {

	var info = make(map[string]TokenInfo)

	for key, value := range *m {
		info[key] = value
	}

	return info, nil
}

// UpdateTokenInfo updates one token info
func (m *MemoryDB) UpdateTokenInfo(tokenInfo TokenInfo) error {

	(*m)[tokenInfo.Symbol] = tokenInfo

	return nil
}

// PriceUpdater definition
type PriceUpdater struct {
	DB     *MemoryDB
	Config ConfigPriceUpdater
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(db *MemoryDB, config ConfigPriceUpdater) (PriceUpdater, error) {

	return PriceUpdater{
		DB:     db,
		Config: config,
	}, nil

}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the MemoryDB
func (p *PriceUpdater) UpdatePrices() error {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    10 * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(ApiUrl).Client(httpClient)

	state := [10]float64{}

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

		// (*p.DB)[p.Config.TokensList[ti]] = tinfo
		err = p.DB.UpdateTokenInfo(tinfo)

		if err != nil {
			return err
		}

	}

	return nil
}

// UpdateConfig allows to update the price-updater configuration
func (p *PriceUpdater) UpdateConfig(config ConfigPriceUpdater) error {

	p.Config = config

	return nil
}

// Get info for a token from the price-updatader database
func (p *PriceUpdater) Get(token string) (TokenInfo, error) {

	return p.DB.Get(token)

}
