package historydb

import (
	"github.com/hermeznetwork/tracerr"
	"github.com/russross/meddler"
)

type provers struct {
	PublicDns string
}

func (hdb *HistoryDB) GetProvers() ([]string, error) {
	var provers []*provers
	err := meddler.QueryAll(
		hdb.dbRead, &provers, "SELECT public_dns FROM provers",
	)
	var publicDns []string
	for _, prover := range provers {
		publicDns = append(publicDns, prover.PublicDns)
	}
	return publicDns, tracerr.Wrap(err)
}
