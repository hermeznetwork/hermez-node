package priceupdater

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/sling"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

const (
	// UpdateMethodTypeBitFinexV2 is the http API used by bitfinex V2
	UpdateMethodTypeBitFinexV2 string = "bitfinexV2"
	// UpdateMethodTypeCoingeckoV3 is the http API used by copingecko V3
	UpdateMethodTypeCoingeckoV3 string = "CoinGeckoV3"
	// UpdateMethodTypeIgnore indicates to not update the value, to set value 0
	// it's better to use UpdateMethodTypeStatic
	UpdateMethodTypeIgnore string = "ignore"
)

// Fiat definition
type Fiat struct {
	APIKey       string
	URL          string
	BaseCurrency string
	Currencies   string
}

// Provider definition
type Provider struct {
	Provider       string
	BaseURL        string
	URL            string
	URLExtraParams string
	SymbolsMap     symbolsMap
	AddressesMap   addressesMap
	Symbols        string
	Addresses      string
}

type staticMap struct {
	Statictokens map[uint]float64
}

// strToStaticTokensMap converts Statictokens mapping from text.
func (d *staticMap) strToStaticTokensMap(str string) error {
	var lastErr error
	if str != "" {
		mapping := make(map[uint]float64)
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			tokenID, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
				lastErr = err
				continue
			}
			if price, err := strconv.ParseFloat(values[1], 64); err != nil {
				log.Error("function strToStaticTokensMap. Error converting string to float64. Avoiding element: ",
					elements[i], " Error: ", err)
				lastErr = err
				continue
			} else {
				mapping[uint(tokenID)] = price
			}
		}
		d.Statictokens = mapping
		log.Debug("StaticToken mapping from config file: ", mapping)
	}
	return lastErr
}

type symbolsMap struct {
	Symbols map[uint]string
}

// strToMapSymbol converts Symbols mapping from text.
func (d *symbolsMap) strToMapSymbol(str string) error {
	var lastErr error
	if str != "" {
		mapping := make(map[uint]string)
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			tokenID, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("function strToMapSymbol. Error converting string to int. Avoiding element: ", elements[i])
				lastErr = err
				continue
			}
			if values[1] == UpdateMethodTypeIgnore || values[1] == "" {
				mapping[uint(tokenID)] = UpdateMethodTypeIgnore
			} else {
				mapping[uint(tokenID)] = values[1]
			}
		}
		d.Symbols = mapping
		log.Debug("Symbol mapping from config file: ", mapping)
	} else {
		d.Symbols = make(map[uint]string)
	}
	return lastErr
}

type addressesMap struct {
	Addresses map[uint]ethCommon.Address
}

// strToMapAddress converts addresses mapping from text.
func (d *addressesMap) strToMapAddress(str string) error {
	var lastErr error
	if str != "" {
		mapping := make(map[uint]ethCommon.Address)
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			tokenID, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("function strToMapAddress. Error converting string to int. Avoiding element: ", elements[i])
				lastErr = err
				continue
			}
			if values[1] == UpdateMethodTypeIgnore || values[1] == "" {
				mapping[uint(tokenID)] = common.FFAddr
			} else {
				mapping[uint(tokenID)] = ethCommon.HexToAddress(values[1])
			}
		}
		d.Addresses = mapping
		log.Debug("Address mapping from config file: ", mapping)
	} else {
		d.Addresses = make(map[uint]ethCommon.Address)
	}
	return lastErr
}

// ProviderValidation method is for validation of Provider struct
func ProviderValidation(sl validator.StructLevel) {
	Provider := sl.Current().Interface().(Provider)
	if Provider.Symbols == "" && Provider.Addresses != "" {
		sl.ReportError(Provider.Addresses, "Addresses", "Addresses", "notokens", "")
		sl.ReportError(Provider.Symbols, "Symbols", "Symbols", "notokens", "")
		return
	}
}

// PriceUpdater definition
type PriceUpdater struct {
	db                    *historydb.HistoryDB
	updateMethodsPriority []string
	tokensList            map[uint]historydb.TokenSymbolAndAddr
	providers             map[string]Provider
	statictokensMap       staticMap
	fiat                  Fiat
	clientProviders       map[string]*sling.Sling
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(
	updateMethodTypesPriority string,
	providers []Provider,
	staticTokens string,
	fiat Fiat,
	db *historydb.HistoryDB,
) (*PriceUpdater, error) {
	priorityArr := strings.Split(string(updateMethodTypesPriority), ",")
	var staticTokensMap staticMap
	err := staticTokensMap.strToStaticTokensMap(staticTokens)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	clientProviders := make(map[string]*sling.Sling)
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	providersMap := make(map[string]Provider)
	for i := 0; i < len(providers); i++ {
		// create mappings
		err := providers[i].SymbolsMap.strToMapSymbol(providers[i].Symbols)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		err = providers[i].AddressesMap.strToMapAddress(providers[i].Addresses)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		// Create Client providers for each provider
		clientProviders[providers[i].Provider] = sling.New().Base(providers[i].BaseURL).Client(httpClient)
		clientProviders["fiat"] = sling.New().Base(fiat.URL).Client(httpClient)
		// Add provider to providersMap
		providersMap[providers[i].Provider] = providers[i]
	}
	return &PriceUpdater{
		db:                    db,
		updateMethodsPriority: priorityArr,
		tokensList:            map[uint]historydb.TokenSymbolAndAddr{},
		providers:             providersMap,
		statictokensMap:       staticTokensMap,
		fiat:                  fiat,
		clientProviders:       clientProviders,
	}, nil
}

func (p *PriceUpdater) getTokenPriceFromProvider(ctx context.Context, tokenID uint) (float64, error) {
	for i := 0; i < len(p.updateMethodsPriority); i++ {
		provider := p.providers[p.updateMethodsPriority[i]]
		var url string
		if _, ok := provider.AddressesMap.Addresses[tokenID]; ok {
			if provider.AddressesMap.Addresses[tokenID] == common.EmptyAddr {
				url = "simple/price?ids=ethereum" + provider.URLExtraParams
			} else {
				url = provider.URL + provider.AddressesMap.Addresses[tokenID].String() + provider.URLExtraParams
			}
		} else {
			url = provider.URL + provider.SymbolsMap.Symbols[tokenID] + provider.URLExtraParams
		}
		req, err := p.clientProviders[provider.Provider].New().Get(url).Request()
		if err != nil {
			return 0, tracerr.Wrap(err)
		}
		var (
			res           *http.Response
			result        float64
			isEmptyResult bool
		)
		switch provider.Provider {
		case UpdateMethodTypeBitFinexV2:
			var data interface{}
			res, err = p.clientProviders[provider.Provider].Do(req.WithContext(ctx), &data, nil)
			if data != nil {
				// The token price is received inside an array in the sixth position
				result = data.([]interface{})[6].(float64)
			} else {
				isEmptyResult = true
			}
		case UpdateMethodTypeCoingeckoV3:
			if provider.AddressesMap.Addresses[tokenID] == common.EmptyAddr {
				var data map[string]map[string]float64
				res, err = p.clientProviders[provider.Provider].Do(req.WithContext(ctx), &data, nil)
				result = data["ethereum"]["usd"]
				if len(data) == 0 {
					isEmptyResult = true
				}
			} else {
				var data map[ethCommon.Address]map[string]float64
				res, err = p.clientProviders[provider.Provider].Do(req.WithContext(ctx), &data, nil)
				result = data[provider.AddressesMap.Addresses[tokenID]]["usd"]
				if len(data) == 0 {
					isEmptyResult = true
				}
			}
		default:
			log.Error("Unknown price provider: ", provider.Provider)
			return 0, tracerr.Wrap(fmt.Errorf("Error: Unknown price provider: " + provider.Provider))
		}
		if err != nil || isEmptyResult || res.StatusCode != http.StatusOK {
			var errMsg strings.Builder
			errMsg.WriteString("Trying another price provider if it's possible.")
			if err != nil {
				errMsg.WriteString(" - Error: " + err.Error())
			}
			if res != nil {
				errMsg.WriteString(fmt.Sprintf(" - HTTP Error: %d %s", res.StatusCode, res.Status))
			}
			errMsg.WriteString(fmt.Sprintf(" - TokenID: %d - URL: %s", tokenID, url))
			log.Warn(errMsg.String())
			continue
		} else {
			return result, nil
		}
	}
	return 0, tracerr.Wrap(fmt.Errorf("Error getting price. All providers have failed"))
}

// UpdatePrices is triggered by the Coordinator, and internally will update the
// token prices in the db
func (p *PriceUpdater) UpdatePrices(ctx context.Context) {
	// Update static prices
	for tokenID, price := range p.statictokensMap.Statictokens {
		if err := p.db.UpdateTokenValueByTokenID(tokenID, price); err != nil {
			log.Errorw("token price not updated (db error)",
				"err", err)
		}
	}
	// Update token prices but ignore ones
	for _, token := range p.tokensList {
		if p.providers[p.updateMethodsPriority[0]].AddressesMap.Addresses[token.TokenID] != common.FFAddr ||
			p.providers[p.updateMethodsPriority[0]].SymbolsMap.Symbols[token.TokenID] == UpdateMethodTypeIgnore {
			tokenPrice, err := p.getTokenPriceFromProvider(ctx, token.TokenID)
			if err != nil {
				log.Errorw("token price from provider error", "err", err, "token", token.Symbol)
			} else if err := p.db.UpdateTokenValueByTokenID(token.TokenID, tokenPrice); err != nil {
				log.Errorw("token price not updated (db error)",
					"err", err, "token", token.Symbol)
			}
		}
	}
}

// UpdateTokenList get the registered token symbols from HistoryDB
func (p *PriceUpdater) UpdateTokenList() error {
	dbTokens, err := p.db.GetTokenSymbolsAndAddrs()
	if err != nil {
		return tracerr.Wrap(err)
	}
	// For each token from the DB
	for _, dbToken := range dbTokens {
		// If the token doesn't exists in the config list,
		// add it with default update method
		if _, ok := p.statictokensMap.Statictokens[dbToken.TokenID]; ok {
			continue
		} else {
			if !(p.providers[p.updateMethodsPriority[0]].SymbolsMap.Symbols[dbToken.TokenID] == UpdateMethodTypeIgnore ||
				p.providers[p.updateMethodsPriority[0]].AddressesMap.Addresses[dbToken.TokenID] == common.FFAddr) {
				p.tokensList[dbToken.TokenID] = dbToken
			}
		}
		for _, provider := range p.providers {
			switch provider.Provider {
			case UpdateMethodTypeBitFinexV2:
				if _, ok := provider.SymbolsMap.Symbols[dbToken.TokenID]; !ok {
					provider.SymbolsMap.Symbols[dbToken.TokenID] = dbToken.Symbol
				}
			case UpdateMethodTypeCoingeckoV3:
				if _, ok := provider.AddressesMap.Addresses[dbToken.TokenID]; !ok {
					provider.AddressesMap.Addresses[dbToken.TokenID] = dbToken.Addr
				}
			default:
				log.Error("Unknown provider detected: ", provider.Provider)
				return tracerr.Wrap(fmt.Errorf("Error: Unknown price provider: " + provider.Provider))
			}
		}
	}
	return nil
}

type fiatExchangeAPI struct {
	Base  string
	Rates interface{}
}

func (p *PriceUpdater) getFiatPrices(ctx context.Context) (map[string]interface{}, error) {
	url := "latest?base=" + p.fiat.BaseCurrency + "&symbols=" + p.fiat.Currencies + "&access_key=" + p.fiat.APIKey
	req, err := p.clientProviders["fiat"].New().Get(url).Request()
	if err != nil {
		return make(map[string]interface{}), tracerr.Wrap(err)
	}
	var (
		res    *http.Response
		result map[string]interface{}
		data   *fiatExchangeAPI
	)
	res, err = p.clientProviders["fiat"].Do(req.WithContext(ctx), &data, nil)
	if err != nil {
		return make(map[string]interface{}), tracerr.Wrap(err)
	}
	if data != nil {
		result = data.Rates.(map[string]interface{})
	} else {
		log.Error("Error: data got are empty. Http code: ", res.StatusCode, ". URL: ", url)
		return make(map[string]interface{}), tracerr.Wrap(fmt.Errorf("Empty data received from the fiat provider"))
	}
	return result, nil
}

// UpdateFiatPrices updates the fiat prices
func (p *PriceUpdater) UpdateFiatPrices(ctx context.Context) error {
	log.Debug("Updating fiat prices")
	// Retrieve fiat prices
	prices, err := p.getFiatPrices(ctx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	// Getting all price from database with baseCurrency USD
	currencies, err := p.db.GetAllFiatPrice("USD")
	if err != nil {
		return tracerr.Wrap(err)
	}
	for token, pr := range prices {
		price := pr.(float64)
		var exist bool
		for i := 0; i < len(currencies); i++ {
			if token == currencies[i].Currency {
				exist = true
			}
		}
		if exist {
			if err = p.db.UpdateFiatPrice(token, "USD", price); err != nil {
				log.Error("DB error updating fiat currency price: ", token, ", ", price, " Error: ", err)
			}
		} else {
			if err = p.db.CreateFiatPrice(token, "USD", price); err != nil {
				log.Error("DB error creating fiat currency price: ", token, ", ", price, " Error: ", err)
			}
		}
	}
	return err
}

// UpdateFiatPricesMock updates the fiat prices
func (p *PriceUpdater) UpdateFiatPricesMock(ctx context.Context) error {
	log.Debug("Updating fiat prices")
	// Retrieve fiat prices
	prices := make(map[string]interface{})
	prices["CNY"] = 6.4306
	prices["EUR"] = 0.817675
	prices["JPY"] = 108.709503
	prices["GBP"] = 0.70335

	// Getting all price from database with baseCurrency USD
	currencies, err := p.db.GetAllFiatPrice("USD")
	if err != nil {
		return tracerr.Wrap(err)
	}
	for token, pr := range prices {
		price := pr.(float64)
		var exist bool
		for i := 0; i < len(currencies); i++ {
			if token == currencies[i].Currency {
				exist = true
			}
		}
		if exist {
			if err = p.db.UpdateFiatPrice(token, "USD", price); err != nil {
				log.Error("DB error updating fiat currency price: ", token, ", ", price, " Error: ", err)
			}
		} else {
			if err = p.db.CreateFiatPrice(token, "USD", price); err != nil {
				log.Error("DB error creating fiat currency price: ", token, ", ", price, " Error: ", err)
			}
		}
	}
	return err
}
