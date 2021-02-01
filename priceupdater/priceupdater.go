package priceupdater

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

// APIType defines the token exchange API
type APIType string

const (
	// APITypeBitFinexV2 is the http API used by bitfinex V2
	APITypeBitFinexV2 APIType = "bitfinexV2"
)

func (t *APIType) valid() bool {
	switch *t {
	case APITypeBitFinexV2:
		return true
	default:
		return false
	}
}

// PriceUpdater definition
type PriceUpdater struct {
	db           *historydb.HistoryDB
	apiURL       string
	apiType      APIType
	tokenSymbols []string
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(apiURL string, apiType APIType, db *historydb.HistoryDB) (*PriceUpdater, error) {
	tokenSymbols := []string{}
	if !apiType.valid() {
		return nil, tracerr.Wrap(fmt.Errorf("Invalid apiType: %v", apiType))
	}
	return &PriceUpdater{
		db:           db,
		apiURL:       apiURL,
		apiType:      apiType,
		tokenSymbols: tokenSymbols,
	}, nil
}

func getTokenPriceBitfinex(ctx context.Context, client *sling.Sling,
	tokenSymbol string) (float64, error) {
	state := [10]float64{}
	req, err := client.New().Get("ticker/t" + tokenSymbol + "USD").Request()
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	res, err := client.Do(req.WithContext(ctx), &state, nil)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	if res.StatusCode != http.StatusOK {
		return 0, tracerr.Wrap(fmt.Errorf("http response is not is %v", res.StatusCode))
	}
	return state[6], nil
}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the db
func (p *PriceUpdater) UpdatePrices(ctx context.Context) {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.apiURL).Client(httpClient)

	for _, tokenSymbol := range p.tokenSymbols {
		var tokenPrice float64
		var err error
		switch p.apiType {
		case APITypeBitFinexV2:
			tokenPrice, err = getTokenPriceBitfinex(ctx, client, tokenSymbol)
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Warnw("token price not updated (get error)",
				"err", err, "token", tokenSymbol, "apiType", p.apiType)
		}
		if err = p.db.UpdateTokenValue(tokenSymbol, tokenPrice); err != nil {
			log.Errorw("token price not updated (db error)",
				"err", err, "token", tokenSymbol, "apiType", p.apiType)
		}
	}
}

// UpdateTokenList get the registered token symbols from HistoryDB
func (p *PriceUpdater) UpdateTokenList() error {
	tokenSymbols, err := p.db.GetTokenSymbols()
	if err != nil {
		return tracerr.Wrap(err)
	}
	p.tokenSymbols = tokenSymbols
	return nil
}
