package common

import (
	"math/big"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	// Stored in DB: mandatory fileds
	TxID        TxID
	BatchNum    BatchNum // batchNum in which this tx was forged.
	Position    int
	FromIdx     Idx
	ToIdx       Idx
	Amount      *big.Int
	Fee         FeeSelector
	Nonce       Nonce
	Type        TxType
	EthBlockNum int64 // Ethereum Block Number in which this L1Tx was added to the queue
}

// Tx returns a *Tx from the L2Tx
func (tx *L2Tx) Tx() *Tx {
	f := new(big.Float).SetInt(tx.Amount)
	amountF, _ := f.Float32()
	return &Tx{
		IsL1:        false,
		TxID:        tx.TxID,
		Type:        tx.Type,
		Position:    tx.Position,
		FromIdx:     tx.FromIdx,
		ToIdx:       tx.ToIdx,
		Amount:      tx.Amount,
		AmountF:     amountF,
		BatchNum:    tx.BatchNum,
		EthBlockNum: tx.EthBlockNum,
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
	}
}
