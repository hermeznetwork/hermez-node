package historydb

import (
	"github.com/hermeznetwork/tracerr"
	"github.com/russross/meddler"
)

type provers struct {
	PublicDNS string `meddler:"public_dns"`
}

// GetProvers get provers addresses saved on databases by hermez-proof-balancer tool
func (hdb *HistoryDB) GetProvers() ([]string, error) {
	var provers []*provers
	err := meddler.QueryAll(
		hdb.dbRead, &provers, "SELECT public_dns FROM provers WHERE status IN ('ready', 'success', 'failed', 'aborted', 'unverified');",
	)
	var publicDNS []string
	for _, prover := range provers {
		publicDNS = append(publicDNS, prover.PublicDNS)
	}
	return publicDNS, tracerr.Wrap(err)
}
