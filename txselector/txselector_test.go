package txselector

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/db/statedb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
func initTestDB(l2 *l2db.L2DB, sdb *statedb.StateDB) *mock.MockDB {
	txs := []common.Tx{
		{
			FromIdx:     common.Idx(0),
			ToIdx:       common.Idx(1),
			TokenID:     1,
			Nonce:       1,
			AbsoluteFee: 1,
		},
		{
			FromIdx:     common.Idx(0),
			ToIdx:       common.Idx(1),
			TokenID:     1,
			Nonce:       2,
			AbsoluteFee: 3,
		},
		{
			FromIdx:     common.Idx(0),
			ToIdx:       common.Idx(1),
			TokenID:     1,
			Nonce:       4,
			AbsoluteFee: 6,
		},
		{
			FromIdx:     common.Idx(0),
			ToIdx:       common.Idx(1),
			TokenID:     1,
			Nonce:       4,
			AbsoluteFee: 4,
		},
		{
			ToIdx:       common.Idx(1),
			FromIdx:     common.Idx(0),
			TokenID:     1,
			Nonce:       1,
			AbsoluteFee: 4,
		},
		{
			ToIdx:       common.Idx(1),
			FromIdx:     common.Idx(0),
			TokenID:     1,
			Nonce:       2,
			AbsoluteFee: 3,
		},
		{
			ToIdx:       common.Idx(1),
			FromIdx:     common.Idx(0),
			TokenID:     1,
			Nonce:       3,
			AbsoluteFee: 5,
		},
		{
			// this tx will not be selected, as the ToEthAddr does not have an account
			FromIdx:     common.Idx(1),
			ToIdx:       common.Idx(2),
			TokenID:     1,
			Nonce:       4,
			AbsoluteFee: 5,
		},
	}

	// n := 0
	nBatch := 0
	for i := 0; i < len(txs); i++ {
		// for i := 0; i < nBatch; i++ {
		//         for j := 0; j < len(txs)/nBatch; j++ {
		// store tx
		l2db.AddTx(uint64(nBatch), txs[i])

		// store account if not yet
		accountID := getAccountID(txs[i].FromEthAddr, txs[i].TokenID)
		if _, ok := m.AccountDB[accountID]; !ok {
			account := common.Account{
				// EthAddr: txs[i].FromEthAddr,
				TokenID: txs[i].TokenID,
				Nonce:   0,
				Balance: big.NewInt(0),
			}
			m.AccountDB[accountID] = account
		}
		// n++
		// }
	}

	return m
}
*/

func TestGetL2TxSelection(t *testing.T) {
	dir, err := ioutil.TempDir("", "tmpdb")
	require.Nil(t, err)
	sdb, err := statedb.NewStateDB(dir, false, false, 0)
	assert.Nil(t, err)
	testL2DB := &l2db.L2DB{}
	// initTestDB(testL2DB, sdb)

	txsel, err := NewTxSelector(sdb, testL2DB, 3, 3, 3)
	assert.Nil(t, err)
	fmt.Println(txsel)

	// txs, err := txsel.GetL2TxSelection(0)
	// assert.Nil(t, err)
	// for _, tx := range txs {
	//         fmt.Println(tx.FromIdx, tx.ToIdx, tx.AbsoluteFee)
	// }
	// assert.Equal(t, 3, len(txs))
	// assert.Equal(t, uint64(6), txs[0].AbsoluteFee)
	// assert.Equal(t, uint64(5), txs[1].AbsoluteFee)
	// assert.Equal(t, uint64(4), txs[2].AbsoluteFee)
}
