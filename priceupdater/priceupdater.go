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
	BASEURL        string
	URL            string
	URLExtraParams string
	SymbolsMap     symbolsMap
	AddressesMap   addressesMap
	Symbols        string
	Addresses      string
}
type staticMap struct {
	Statictokens map[int]float64
}

// strToMapStatic converts Statictokens mapping from text.
func (d *staticMap) strToMapStatic(str string) error {
	mapping := make(map[int]float64)
	if str != "" {
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			num, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
				continue
			}
			if price, err := strconv.ParseFloat(values[1], 64); err != nil {
				log.Error("Error converting string to float64. Avoiding element: ", elements[i], " Error: ", err)
				continue
			} else {
				mapping[num] = price
			}
		}
		d.Statictokens = mapping
		log.Debug("StaticToken mapping from config file: ", mapping)
	}
	return nil
}

type symbolsMap struct {
	Symbols map[int]string
}

// strToMapSymbol converts Symbols mapping from text.
func (d *symbolsMap) strToMapSymbol(str string) error {
	mapping := make(map[int]string)
	if str != "" {
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			num, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
				continue
			}
			if values[1] == UpdateMethodTypeIgnore || values[1] == "" {
				mapping[num] = UpdateMethodTypeIgnore
			} else {
				mapping[num] = values[1]
			}
		}
		d.Symbols = mapping
		log.Debug("Symbol mapping from config file: ", mapping)
	}
	return nil
}

type addressesMap struct {
	Addresses map[int]ethCommon.Address
}

// strToMapAddress converts addresses mapping from text.
func (d *addressesMap) strToMapAddress(str string) error {
	if str != "" {
		mapping := make(map[int]ethCommon.Address)
		elements := strings.Split(str, ",")
		for i := 0; i < len(elements); i++ {
			values := strings.Split(elements[i], "=")
			num, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
				continue
			}
			if values[1] == UpdateMethodTypeIgnore || values[1] == "" {
				mapping[num] = common.FFAddr
			} else {
				mapping[num] = ethCommon.HexToAddress(values[1])
			}
		}
		d.Addresses = mapping
		log.Debug("Address mapping from config file: ", mapping)
	}
	return nil
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
	tokensList            map[int]historydb.TokenSymbolAndAddr
	providers             []Provider
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
	err := staticTokensMap.strToMapStatic(staticTokens)
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
	for i := 0; i < len(providers); i++ {
		//create mappings
		err := providers[i].SymbolsMap.strToMapSymbol(providers[i].Symbols)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		err = providers[i].AddressesMap.strToMapAddress(providers[i].Addresses)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		//Create Client providers for each provider
		clientProviders[providers[i].Provider] = sling.New().Base(providers[i].BASEURL).Client(httpClient)
		clientProviders["fiat"] = sling.New().Base(fiat.URL).Client(httpClient)
	}
	return &PriceUpdater{
		db:                    db,
		updateMethodsPriority: priorityArr,
		tokensList:            map[int]historydb.TokenSymbolAndAddr{},
		providers:             providers,
		statictokensMap:       staticTokensMap,
		fiat:                  fiat,
		clientProviders:       clientProviders,
	}, nil
}

type coingecko map[ethCommon.Address]map[string]float64

func (p *PriceUpdater) getTokenPriceFromProvider(ctx context.Context, tokenID int) (float64, error) {
	for i := 0; i < len(p.updateMethodsPriority); i++ {
		for j := 0; j < len(p.providers); j++ {
			if p.updateMethodsPriority[i] == p.providers[j].Provider {
				var url string
				if _, ok := p.providers[j].AddressesMap.Addresses[tokenID]; ok {
					url = p.providers[j].URL + p.providers[j].AddressesMap.Addresses[tokenID].String() + p.providers[j].URLExtraParams
				} else {
					url = p.providers[j].URL + p.providers[j].SymbolsMap.Symbols[tokenID] + p.providers[j].URLExtraParams
				}
				req, err := p.clientProviders[p.providers[j].Provider].New().Get(url).Request()
				if err != nil {
					return 0, tracerr.Wrap(err)
				}
				var res *http.Response
				var result float64
				var errResult bool
				if p.providers[j].Provider == UpdateMethodTypeBitFinexV2 {
					var data interface{}
					res, err = p.clientProviders[p.providers[j].Provider].Do(req.WithContext(ctx), &data, nil)
					if data != nil {
						result = data.([]interface{})[6].(float64)
					} else {
						errResult = true
					}
				} else if p.providers[j].Provider == UpdateMethodTypeCoingeckoV3 {
					var data coingecko
					res, err = p.clientProviders[p.providers[j].Provider].Do(req.WithContext(ctx), &data, nil)
					result = data[p.providers[j].AddressesMap.Addresses[tokenID]]["usd"]
				} else {
					log.Error("Unknown price provider: ", p.providers[j].Provider)
					return 0, tracerr.Wrap(fmt.Errorf("Error: Unknown price provider: " + p.providers[j].Provider))
				}
				if err != nil || errResult {
					log.Warn("Trying another price provider: ", err, " http error code: ", res.StatusCode, " tokenId: ", tokenID, ". URL: ", url)
					continue
				} else if res.StatusCode != http.StatusOK {
					log.Warn("Trying another price provider. Http response code: ", res.StatusCode)
					continue
				} else {
					return result, nil
				}
			}
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
	//Update token prices but ignore ones
	for _, token := range p.tokensList {
		if p.providers[0].AddressesMap.Addresses[token.TokenID] != common.FFAddr ||
			p.providers[0].SymbolsMap.Symbols[token.TokenID] == UpdateMethodTypeIgnore {
			tokenPrice, _ := p.getTokenPriceFromProvider(ctx, token.TokenID)
			if err := p.db.UpdateTokenValueByTokenID(token.TokenID, tokenPrice); err != nil {
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
			if !(p.providers[0].SymbolsMap.Symbols[dbToken.TokenID] == UpdateMethodTypeIgnore ||
				p.providers[0].AddressesMap.Addresses[dbToken.TokenID] == common.FFAddr) {
				p.tokensList[dbToken.TokenID] = dbToken
			}
		}
		for i := 0; i < len(p.providers); i++ {
			if len(p.providers[i].SymbolsMap.Symbols) != 0 {
				if _, ok := p.providers[i].SymbolsMap.Symbols[dbToken.TokenID]; !ok {
					p.providers[i].SymbolsMap.Symbols[dbToken.TokenID] = dbToken.Symbol
				}
			}
			if len(p.providers[i].AddressesMap.Addresses) != 0 {
				if _, ok := p.providers[i].AddressesMap.Addresses[dbToken.TokenID]; !ok {
					p.providers[i].AddressesMap.Addresses[dbToken.TokenID] = dbToken.Addr
				}
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
	var url = "latest?base=" + p.fiat.BaseCurrency + "&symbols=" + p.fiat.Currencies + "&access_key=" + p.fiat.APIKey
	req, err := p.clientProviders["fiat"].New().Get(url).Request()
	if err != nil {
		return make(map[string]interface{}), tracerr.Wrap(err)
	}
	var res *http.Response
	var result map[string]interface{}
	var data *fiatExchangeAPI
	res, err = p.clientProviders["fiat"].Do(req.WithContext(ctx), &data, nil)
	if err != nil {
		return make(map[string]interface{}), tracerr.Wrap(err)
	}
	if data != nil {
		result = data.Rates.(map[string]interface{})
	} else {
		log.Error("Error: data got are empty. Http code: ", res.StatusCode, ". URL: ", url)
	}
	return result, nil
}

// UpdateFiatPrices updates the fiat prices
func (p *PriceUpdater) UpdateFiatPrices(ctx context.Context) error {
	log.Debug("Updating fiat prices")
	//Retrieve fiat prices
	prices, err := p.getFiatPrices(ctx)
	if err != nil {
		return tracerr.Wrap(err)
	}
	//Getting all price from database with baseCurrency USD
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
	//Retrieve fiat prices
	prices := make(map[string]interface{})
	prices["CNY"] = 6.4306
	prices["EUR"] = 0.817675
	prices["JPY"] = 108.709503
	prices["GBP"] = 0.70335

	//Getting all price from database with baseCurrency USD
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
