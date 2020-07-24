package txselector

import (
	"fmt"
	"math/big"
	"testing"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/tx-selector/common"
	"github.com/hermeznetwork/hermez-node/tx-selector/mock"
	"github.com/stretchr/testify/assert"
)

func initMockDB() *mock.MockDB {
	m := mock.New()

	txs := []common.Tx{
		{
			FromEthAddr:     ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			ToEthAddr:       ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			TokenID:         1,
			Nonce:           1,
			UserFeeAbsolute: 1,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			ToEthAddr:       ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			TokenID:         1,
			Nonce:           2,
			UserFeeAbsolute: 3,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			ToEthAddr:       ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			TokenID:         1,
			Nonce:           4,
			UserFeeAbsolute: 6,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			ToEthAddr:       ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			TokenID:         1,
			Nonce:           4,
			UserFeeAbsolute: 4,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			ToEthAddr:       ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			TokenID:         1,
			Nonce:           1,
			UserFeeAbsolute: 4,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			ToEthAddr:       ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			TokenID:         1,
			Nonce:           2,
			UserFeeAbsolute: 3,
		},
		{
			FromEthAddr:     ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			ToEthAddr:       ethCommon.HexToAddress("0x859c3d0d5aD917F146fF6654A4C676f1ddeCE26a"),
			TokenID:         1,
			Nonce:           3,
			UserFeeAbsolute: 5,
		},
		{
			// this tx will not be selected, as the ToEthAddr does not have an account
			FromEthAddr:     ethCommon.HexToAddress("0x6950E814B82d276DB5Fa7f34253CfeE1387fe03E"),
			ToEthAddr:       ethCommon.HexToAddress("0x4a2CFDF534725D8D6e07Af97B237Fff19BDb3c93"),
			TokenID:         1,
			Nonce:           4,
			UserFeeAbsolute: 5,
		},
	}

	// n := 0
	nBatch := 0
	for i := 0; i < len(txs); i++ {
		// for i := 0; i < nBatch; i++ {
		//         for j := 0; j < len(txs)/nBatch; j++ {
		// store tx
		m.AddTx(uint64(nBatch), txs[i])

		// store account if not yet
		accountID := getAccountID(txs[i].FromEthAddr, txs[i].TokenID)
		if _, ok := m.AccountDB[accountID]; !ok {
			account := common.Account{
				EthAddr: txs[i].FromEthAddr,
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

func TestGetL2TxSelection(t *testing.T) {
	mockDB := initMockDB()
	txsel := NewTxSelector(mockDB, 3, 3, 3)

	txs, err := txsel.GetL2TxSelection(0)
	assert.Nil(t, err)
	for _, tx := range txs {
		fmt.Println(tx.FromEthAddr.String(), tx.ToEthAddr.String(), tx.UserFeeAbsolute)
	}
	assert.Equal(t, 3, len(txs))
	assert.Equal(t, uint64(6), txs[0].UserFeeAbsolute)
	assert.Equal(t, uint64(5), txs[1].UserFeeAbsolute)
	assert.Equal(t, uint64(4), txs[2].UserFeeAbsolute)
}
