package mock

import (
	"github.com/hermeznetwork/hermez-node/txselector/common"
)

type MockDB struct {
	Txs map[uint64][]common.Tx

	// AccountDB is the LocalAccountDB copy of the original AccountDB
	AccountDB map[[36]byte]common.Account // [36]byte is tx.ToEthAddr + tx.TokenID

	PendingRegistersDB map[[36]byte]common.Account // [36]byte is tx.ToEthAddr + tx.TokenID
}

func New() *MockDB {
	return &MockDB{
		Txs:                make(map[uint64][]common.Tx),
		AccountDB:          make(map[[36]byte]common.Account),
		PendingRegistersDB: make(map[[36]byte]common.Account),
	}
}

func (m *MockDB) AddTx(batchID uint64, tx common.Tx) {
	if _, ok := m.Txs[batchID]; !ok {
		m.Txs[batchID] = []common.Tx{}
	}
	m.Txs[batchID] = append(m.Txs[batchID], tx)
}

func (m *MockDB) GetTxs(batchID uint64) []common.Tx {
	return m.Txs[batchID]
}
