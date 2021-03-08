package priceupdater

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/sling"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
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
	// APITypeCoingeckoV3 is the http API used by copingecko V3
	APITypeCoingeckoV3 APIType = "coingeckoV3"
)

func (t *APIType) valid() bool {
	switch *t {
	case APITypeBitFinexV2:
		return true
	case APITypeCoingeckoV3:
		return true
	default:
		return false
	}
}

// PriceUpdater definition
type PriceUpdater struct {
	db      *historydb.HistoryDB
	apiURL  string
	apiType APIType
	tokens  []historydb.TokenSymbolAndAddr
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(apiURL string, apiType APIType, db *historydb.HistoryDB) (*PriceUpdater,
	error) {
	if !apiType.valid() {
		return nil, tracerr.Wrap(fmt.Errorf("Invalid apiType: %v", apiType))
	}
	return &PriceUpdater{
		db:      db,
		apiURL:  apiURL,
		apiType: apiType,
		tokens:  []historydb.TokenSymbolAndAddr{},
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

func getTokenPriceCoingecko(ctx context.Context, client *sling.Sling,
	tokenAddr ethCommon.Address) (float64, error) {
	responseObject := make(map[string]map[string]float64)
	var url string
	var id string
	if tokenAddr == common.EmptyAddr { // Special case for Ether
		url = "simple/price?ids=ethereum&vs_currencies=usd"
		id = "ethereum"
	} else { // Common case (ERC20)
		id = strings.ToLower(tokenAddr.String())
		url = "simple/token_price/ethereum?contract_addresses=" +
			id + "&vs_currencies=usd"
	}
	req, err := client.New().Get(url).Request()
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	res, err := client.Do(req.WithContext(ctx), &responseObject, nil)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	if res.StatusCode != http.StatusOK {
		return 0, tracerr.Wrap(fmt.Errorf("http response is not is %v", res.StatusCode))
	}
	price := responseObject[id]["usd"]
	if price <= 0 {
		return 0, tracerr.Wrap(fmt.Errorf("price not found for %v", id))
	}
	return price, nil
}

// UpdatePrices is triggered by the Coordinator, and internally will update the
// token prices in the db
func (p *PriceUpdater) UpdatePrices(ctx context.Context) {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.apiURL).Client(httpClient)

	for _, token := range p.tokens {
		var tokenPrice float64
		var err error
		switch p.apiType {
		case APITypeBitFinexV2:
			tokenPrice, err = getTokenPriceBitfinex(ctx, client, token.Symbol)
		case APITypeCoingeckoV3:
			tokenPrice, err = getTokenPriceCoingecko(ctx, client, token.Addr)
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Warnw("token price not updated (get error)",
				"err", err, "token", token.Symbol, "apiType", p.apiType)
		}
		if err = p.db.UpdateTokenValue(token.Symbol, tokenPrice); err != nil {
			log.Errorw("token price not updated (db error)",
				"err", err, "token", token.Symbol, "apiType", p.apiType)
		}
	}
}

// UpdateTokenList get the registered token symbols from HistoryDB
func (p *PriceUpdater) UpdateTokenList() error {
	tokens, err := p.db.GetTokenSymbolsAndAddrs()
	if err != nil {
		return tracerr.Wrap(err)
	}
	p.tokens = tokens
	return nil
}
