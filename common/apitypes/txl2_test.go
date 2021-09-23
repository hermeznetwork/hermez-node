package apitypes

import (
	"testing"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/stretchr/testify/assert"
)

func TestMarshalUnmarshal(t *testing.T) {
	txID, err := common.NewTxIDFromString("0x02c4ebd603b30d759a5a81622f7e4dee46bb1cb5cb756f45dcee9134893cd6a093")
	assert.NoError(t, err)
	var token = token{
		TokenName:        "Token name",
		TokenSymbol:      "TKSY",
		TokenEthBlockNum: 23,
	}
	var l2tx = TxL2{
		ItemID: 65,
		TxID:   txID,
		Token:  token,
	}
	bytes, err := l2tx.MarshalJSON()
	assert.NoError(t, err)

	var dest TxL2
	err = dest.UnmarshalJSON(bytes)
	assert.NoError(t, err)
	assert.Equal(t, l2tx.Token.TokenName, dest.Token.TokenName)
	assert.Equal(t, l2tx.Token.TokenSymbol, dest.Token.TokenSymbol)
	assert.Equal(t, l2tx.Token.TokenEthBlockNum, dest.Token.TokenEthBlockNum)
	assert.Equal(t, l2tx.ItemID, dest.ItemID)
	assert.Equal(t, l2tx.TxID, dest.TxID)
	assert.Equal(t, l2tx, dest)
}
