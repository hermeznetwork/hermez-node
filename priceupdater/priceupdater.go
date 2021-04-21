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
	"gopkg.in/go-playground/validator.v9"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 2 * time.Second
)

// UpdateMethodType defines the token price update mechanism
type UpdateMethodType string

const (
	// UpdateMethodTypeBitFinexV2 is the http API used by bitfinex V2
	UpdateMethodTypeBitFinexV2 UpdateMethodType = "bitfinexV2"
	// UpdateMethodTypeCoingeckoV3 is the http API used by copingecko V3
	UpdateMethodTypeCoingeckoV3 UpdateMethodType = "coingeckoV3"
	// UpdateMethodTypeStatic is the value given by the configuration
	UpdateMethodTypeStatic UpdateMethodType = "static"
	// UpdateMethodTypeIgnore indicates to not update the value, to set value 0
	// it's better to use UpdateMethodTypeStatic
	UpdateMethodTypeIgnore UpdateMethodType = "ignore"
)

// ValidateUpdateMethodType method is for validation update method field in config
func ValidateUpdateMethodType(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(UpdateMethodType)
	switch field {
	case UpdateMethodTypeBitFinexV2:
		return true
	case UpdateMethodTypeCoingeckoV3:
		return true
	case UpdateMethodTypeStatic:
		return true
	case UpdateMethodTypeIgnore:
		return true
	default:
		return false
	}
}

// ValidateIsUpdateMethodTypeIsNotStatic method is for validation update method type field is not static
func ValidateIsUpdateMethodTypeIsNotStatic(fl validator.FieldLevel) bool {
	field := fl.Field().Interface().(UpdateMethodType)
	return field != UpdateMethodTypeStatic
}

// TokenConfig specifies how a single token get its price updated
type TokenConfig struct {
	UpdateMethod UpdateMethodType `validate:"is-valid-updatemethodtype"`
	StaticValue  float64          // required by UpdateMethodTypeStatic
	Symbol       string
	Addr         ethCommon.Address
}

// TokenConfigValidation method is for validation of tokenConfig struct
func TokenConfigValidation(sl validator.StructLevel) {
	tokenConfig := sl.Current().Interface().(TokenConfig)
	if tokenConfig.Addr == common.EmptyAddr && tokenConfig.Symbol != "ETH" {
		sl.ReportError(tokenConfig.Addr, "Addr", "Addr", "emptyaddrfornoteth", "")
		sl.ReportError(tokenConfig.Symbol, "Symbol", "Symbol", "emptyaddrfornoteth", "")
		return
	} else if tokenConfig.Symbol == "" && tokenConfig.UpdateMethod == UpdateMethodTypeBitFinexV2 {
		sl.ReportError(tokenConfig.Symbol, "Symbol", "Symbol", "emptysymbolforbitfinex", "")
		return
	}
}

// PriceUpdater definition
type PriceUpdater struct {
	db                  *historydb.HistoryDB
	defaultUpdateMethod UpdateMethodType
	tokensList          []historydb.TokenSymbolAndAddr
	tokensConfig        map[ethCommon.Address]TokenConfig
	clientCoingeckoV3   *sling.Sling
	clientBitfinexV2    *sling.Sling
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(
	defaultUpdateMethodType UpdateMethodType,
	tokensConfig []TokenConfig,
	db *historydb.HistoryDB,
	bitfinexV2URL, coingeckoV3URL string,
) (*PriceUpdater, error) {
	tokensConfigMap := make(map[ethCommon.Address]TokenConfig)
	for _, t := range tokensConfig {
		tokensConfigMap[t.Addr] = t
	}
	// Init
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	return &PriceUpdater{
		db:                  db,
		defaultUpdateMethod: defaultUpdateMethodType,
		tokensList:          []historydb.TokenSymbolAndAddr{},
		tokensConfig:        tokensConfigMap,
		clientCoingeckoV3:   sling.New().Base(coingeckoV3URL).Client(httpClient),
		clientBitfinexV2:    sling.New().Base(bitfinexV2URL).Client(httpClient),
	}, nil
}

func (p *PriceUpdater) getTokenPriceBitfinex(ctx context.Context, tokenSymbol string) (float64, error) {
	state := [10]float64{}
	url := "ticker/t" + tokenSymbol + "USD"
	req, err := p.clientBitfinexV2.New().Get(url).Request()
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	res, err := p.clientBitfinexV2.Do(req.WithContext(ctx), &state, nil)
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	if res.StatusCode != http.StatusOK {
		return 0, tracerr.Wrap(fmt.Errorf("http response is not is %v", res.StatusCode))
	}
	return state[6], nil
}

func (p *PriceUpdater) getTokenPriceCoingecko(ctx context.Context, tokenAddr ethCommon.Address) (float64, error) {
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
	req, err := p.clientCoingeckoV3.New().Get(url).Request()
	if err != nil {
		return 0, tracerr.Wrap(err)
	}
	res, err := p.clientCoingeckoV3.Do(req.WithContext(ctx), &responseObject, nil)
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
	for _, token := range p.tokensConfig {
		var tokenPrice float64
		var err error
		switch token.UpdateMethod {
		case UpdateMethodTypeBitFinexV2:
			tokenPrice, err = p.getTokenPriceBitfinex(ctx, token.Symbol)
		case UpdateMethodTypeCoingeckoV3:
			tokenPrice, err = p.getTokenPriceCoingecko(ctx, token.Addr)
		case UpdateMethodTypeStatic:
			tokenPrice = token.StaticValue
			if tokenPrice == float64(0) {
				log.Warn("token price is set to 0. Probably StaticValue is not put in the configuration file,",
					"token", token.Symbol)
			}
		case UpdateMethodTypeIgnore:
			continue
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Warnw("token price not updated (get error)",
				"err", err, "token", token.Symbol, "updateMethod", token.UpdateMethod)
		}
		if err = p.db.UpdateTokenValue(token.Addr, tokenPrice); err != nil {
			log.Errorw("token price not updated (db error)",
				"err", err, "token", token.Symbol, "updateMethod", token.UpdateMethod)
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
		// add it with default update emthod
		if _, ok := p.tokensConfig[dbToken.Addr]; !ok {
			p.tokensConfig[dbToken.Addr] = TokenConfig{
				UpdateMethod: p.defaultUpdateMethod,
				Symbol:       dbToken.Symbol,
				Addr:         dbToken.Addr,
			}
		}
	}
	return nil
}
