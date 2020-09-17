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
	EthBlockNum int64 // Ethereum Block Number in which this L2Tx was added to the queue
}

// Tx returns a *Tx from the L2Tx
func (tx *L2Tx) Tx() *Tx {
	f := new(big.Float).SetInt(tx.Amount)
	amountFloat, _ := f.Float64()
	return &Tx{
		IsL1:        false,
		TxID:        tx.TxID,
		Type:        tx.Type,
		Position:    tx.Position,
		FromIdx:     tx.FromIdx,
		ToIdx:       tx.ToIdx,
		Amount:      tx.Amount,
		AmountFloat: amountFloat,
		BatchNum:    tx.BatchNum,
		EthBlockNum: tx.EthBlockNum,
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
	}
}

// PoolL2Tx returns the data structure of PoolL2Tx with the parameters of a
// L2Tx filled
func (tx *L2Tx) PoolL2Tx() *PoolL2Tx {
	return &PoolL2Tx{
		TxID:     tx.TxID,
		BatchNum: tx.BatchNum,
		FromIdx:  tx.FromIdx,
		ToIdx:    tx.ToIdx,
		Amount:   tx.Amount,
		Fee:      tx.Fee,
		Nonce:    tx.Nonce,
		Type:     tx.Type,
	}
}

// L2TxsToPoolL2Txs returns an array of []*PoolL2Tx from an array of []*L2Tx,
// where the PoolL2Tx only have the parameters of a L2Tx filled.
func L2TxsToPoolL2Txs(txs []*L2Tx) []*PoolL2Tx {
	var r []*PoolL2Tx
	for _, tx := range txs {
		r = append(r, tx.PoolL2Tx())
	}
	return r
}
