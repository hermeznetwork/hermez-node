package priceupdater

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"strconv"

	"github.com/dghubble/sling"
	ethCommon "github.com/ethereum/go-ethereum/common"
	// "github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/hermeznetwork/tracerr"
	"gopkg.in/go-playground/validator.v9"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

// Priority defines the priority provider
type Priority string

const (
	// UpdateMethodTypeBitFinexV2 is the http API used by bitfinex V2
	UpdateMethodTypeBitFinexV2 string = "bitfinexV2"
	// UpdateMethodTypeCoingeckoV3 is the http API used by copingecko V3
	UpdateMethodTypeCoingeckoV3 string = "CoinGeckoV3"
	// UpdateMethodTypeStatic is the value given by the configuration
	UpdateMethodTypeStatic string = "static"
	// UpdateMethodTypeIgnore indicates to not update the value, to set value 0
	// it's better to use UpdateMethodTypeStatic
	UpdateMethodTypeIgnore string = "ignore"
)

// Provider specifies how a provider get the price updated
type Provider struct {
	Provider string
	BASEURL  string
	URL  string
	URLExtraParams string
	SymbolsMap       SymbolsMap
	AddressesMap        AddressesMap
	Symbols       string
	Addresses         string
}
type SymbolsMap struct{
	Symbols map[int]string
}
// strToMapSymbol converts Symbols mapping from text.
func (d *SymbolsMap) strToMapSymbol(str string) error {
	mapping := make(map[int]string)
	if str != "" {
		elements := strings.Split(str,",")
		for i:=0;i<len(elements); i++ {
			values := strings.Split(elements[i],"=")
			num, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
			}
			mapping[num] = values[1]
		}
		d.Symbols = mapping
		log.Debug("Symbol mapping from config file: ",mapping)
	}
	return nil
}
type AddressesMap struct{
	Addresses map[int]ethCommon.Address
}
// strToMapAddress converts addresses mapping from text.
func (d *AddressesMap) strToMapAddress(str string) error {
	if str != "" {
		mapping := make(map[int]ethCommon.Address)
		elements := strings.Split(str,",")
		for i:=0;i<len(elements); i++ {
			values := strings.Split(elements[i],"=")
			num, err := strconv.Atoi(values[0])
			if err != nil {
				log.Error("Error converting string to int. Avoiding element: ", elements[i])
			}
			mapping[num] = ethCommon.HexToAddress(values[1])
		}
		d.Addresses = mapping
		log.Debug("Address mapping from config file: ",mapping)
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
	db                  *historydb.HistoryDB
	updateMethodsPriority []string
	tokensList          []historydb.TokenSymbolAndAddr
	providers        []Provider
	clientProviders   map[string]*sling.Sling
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(
	updateMethodTypesPriority Priority,
	providers []Provider,
	db *historydb.HistoryDB,
) (*PriceUpdater, error) {
	priorityArr := strings.Split(string(updateMethodTypesPriority),",")
	clientProviders := make(map[string]*sling.Sling)
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	for i:=0; i<len(providers); i++ {
		//create mappings
		providers[i].SymbolsMap.strToMapSymbol(providers[i].Symbols)
		providers[i].AddressesMap.strToMapAddress(providers[i].Addresses)
		//Create Client providers for each provider
		clientProviders[providers[i].Provider] = sling.New().Base(providers[i].BASEURL).Client(httpClient)
	}
	return &PriceUpdater{
		db:                  db,
		updateMethodsPriority: priorityArr,
		tokensList:          []historydb.TokenSymbolAndAddr{},
		providers:        providers,
		clientProviders:   clientProviders,
	}, nil
}
type coingecko map[ethCommon.Address]map[string]float64

func (p *PriceUpdater) getTokenPriceFromProvider(ctx context.Context, tokenId int) (float64, error) {
	for i:=0; i<len(p.updateMethodsPriority); i++ {
		for j:=0; j<len(p.providers); j++ {
			if p.updateMethodsPriority[i] == p.providers[j].Provider {
				var url string
				if _, ok := p.providers[j].AddressesMap.Addresses[tokenId]; ok {
					url = p.providers[j].URL + p.providers[j].AddressesMap.Addresses[tokenId].String() + p.providers[j].URLExtraParams
				} else {
					url = p.providers[j].URL + p.providers[j].SymbolsMap.Symbols[tokenId] + p.providers[j].URLExtraParams
				}
				req, err := p.clientProviders[p.providers[j].Provider].New().Get(url).Request()
				if err != nil {
					return 0, tracerr.Wrap(err)
				}
				var res *http.Response
				var result float64
				if p.providers[j].Provider == UpdateMethodTypeBitFinexV2 {
					var data interface{}
					res, err = p.clientProviders[p.providers[j].Provider].Do(req.WithContext(ctx), &data, nil)
					result = data.([]interface{})[6].(float64)
				} else if p.providers[j].Provider == UpdateMethodTypeCoingeckoV3 {
					var data coingecko
					res, err = p.clientProviders[p.providers[j].Provider].Do(req.WithContext(ctx), &data, nil)
					result = data[p.providers[j].AddressesMap.Addresses[tokenId]]["usd"]
				} else {
					log.Error("Unknown price provider: ", p.providers[j].Provider)
					return 0, tracerr.Wrap(fmt.Errorf("Error: Unknown price provider: "+ p.providers[j].Provider))
				}
				if err != nil {
					log.Warn("Trying another price provider: ", err)
					continue
				}else if res.StatusCode != http.StatusOK {
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

// func (p *PriceUpdater) getTokenPriceCoingecko(ctx context.Context, tokenAddr ethCommon.Address) (float64, error) {
// 	responseObject := make(map[string]map[string]float64)
// 	var url string
// 	var id string
// 	if tokenAddr == common.EmptyAddr { // Special case for Ether
// 		url = "simple/price?ids=ethereum&vs_currencies=usd"
// 		id = "ethereum"
// 	} else { // Common case (ERC20)
// 		id = strings.ToLower(tokenAddr.String())
// 		url = "simple/token_price/ethereum?contract_addresses=" +
// 			id + "&vs_currencies=usd"
// 	}
// 	req, err := p.clientCoingeckoV3.New().Get(url).Request()
// 	if err != nil {
// 		return 0, tracerr.Wrap(err)
// 	}
// 	res, err := p.clientCoingeckoV3.Do(req.WithContext(ctx), &responseObject, nil)
// 	if err != nil {
// 		return 0, tracerr.Wrap(err)
// 	}
// 	if res.StatusCode != http.StatusOK {
// 		return 0, tracerr.Wrap(fmt.Errorf("http response is not is %v", res.StatusCode))
// 	}
// 	price := responseObject[id]["usd"]
// 	if price <= 0 {
// 		return 0, tracerr.Wrap(fmt.Errorf("price not found for %v", id))
// 	}
// 	return price, nil
// }

// UpdatePrices is triggered by the Coordinator, and internally will update the
// token prices in the db
func (p *PriceUpdater) UpdatePrices(ctx context.Context) {
	tokenPrice, _ := p.getTokenPriceFromProvider(ctx, 2)
	log.Warn("tokenPrice: ",tokenPrice)
	// for _, token := range p.tokensConfig {
	// 	var tokenPrice float64
	// 	var err error
	// 	log.Warn("TOKEN TO CHECK PRICE: ",token.Symbol, token.StaticValue)
	// 	switch token.UpdateMethod {
	// 	case UpdateMethodTypeBitFinexV2:
	// 		log.Warn("Bitfinex method")
	// 		tokenPrice, err = p.getTokenPriceBitfinex(ctx, token.Symbol)
	// 	case UpdateMethodTypeCoingeckoV3:
	// 		log.Warn("coingecko method")
	// 		tokenPrice, err = p.getTokenPriceCoingecko(ctx, token.Addr)
	// 	case UpdateMethodTypeStatic:
	// 		log.Warn("static method")
	// 		tokenPrice = token.StaticValue
	// 		if tokenPrice == float64(0) {
	// 			log.Warn("token price is set to 0. Probably StaticValue is not put in the configuration file,",
	// 				"token", token.Symbol)
	// 		}
	// 	case UpdateMethodTypeIgnore:
	// 		log.Warn("ignore method")
	// 		continue
	// 	}
	// 	if ctx.Err() != nil {
	// 		return
	// 	}
	// 	if err != nil {
	// 		log.Warnw("token price not updated (get error)",
	// 			"err", err, "token", token.Symbol, "updateMethod", token.UpdateMethod)
	// 			return
	// 	}
	// 	log.Warn("token: ",token)
	// 	if err = p.db.UpdateTokenValue(token.Addr, tokenPrice); err != nil {
	// 		log.Errorw("token price not updated (db error)",
	// 			"err", err, "token", token.Symbol, "updateMethod", token.UpdateMethod)
	// 	}
	// }
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
		// add it with default update emthod
		for i:=0; i<len(p.providers); i++ {
			if len(p.providers[i].SymbolsMap.Symbols) != 0 {
				if _, ok := p.providers[i].SymbolsMap.Symbols[dbToken.TokenId]; !ok {
					p.providers[i].SymbolsMap.Symbols[dbToken.TokenId] = dbToken.Symbol
				}
			}
			if len(p.providers[i].AddressesMap.Addresses) != 0 {
				if _, ok := p.providers[i].AddressesMap.Addresses[dbToken.TokenId]; !ok {
					p.providers[i].AddressesMap.Addresses[dbToken.TokenId] = dbToken.Addr
				}
			}
		}
	}
	return nil
}
