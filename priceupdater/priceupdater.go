package priceupdater

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dghubble/sling"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/log"
	"github.com/ztrue/tracerr"
)

const (
	defaultMaxIdleConns    = 10
	defaultIdleConnTimeout = 10
)

// PriceUpdater definition
type PriceUpdater struct {
	db           *historydb.HistoryDB
	apiURL       string
	tokenSymbols []string
}

// NewPriceUpdater is the constructor for the updater
func NewPriceUpdater(apiURL string, db *historydb.HistoryDB) PriceUpdater {
	tokenSymbols := []string{}
	return PriceUpdater{
		db:           db,
		apiURL:       apiURL,
		tokenSymbols: tokenSymbols,
	}
}

// UpdatePrices is triggered by the Coordinator, and internally will update the token prices in the db
func (p *PriceUpdater) UpdatePrices() {
	tr := &http.Transport{
		MaxIdleConns:       defaultMaxIdleConns,
		IdleConnTimeout:    defaultIdleConnTimeout * time.Second,
		DisableCompression: true,
	}
	httpClient := &http.Client{Transport: tr}
	client := sling.New().Base(p.apiURL).Client(httpClient)

	state := [10]float64{}

	for _, tokenSymbol := range p.tokenSymbols {
		resp, err := client.New().Get("ticker/t" + tokenSymbol + "USD").ReceiveSuccess(&state)
		errString := tokenSymbol + " not updated, error: "
		if err != nil {
			log.Error(errString + err.Error())
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Error(errString + "response is not 200, is " + strconv.Itoa(resp.StatusCode))
			continue
		}
		err = p.db.UpdateTokenValue(tokenSymbol, state[6])
		if err != nil {
			log.Error(errString + err.Error())
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
