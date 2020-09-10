package common

import (
	"math/big"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	// Stored in DB: mandatory fileds
	TxID     TxID        `meddler:"tx_id"`
	BatchNum BatchNum    `meddler:"batch_num"` // batchNum in which this tx was forged.
	Position int         `meddler:"position"`
	FromIdx  Idx         `meddler:"from_idx"`
	ToIdx    Idx         `meddler:"to_idx"`
	Amount   *big.Int    `meddler:"amount,bigint"`
	Fee      FeeSelector `meddler:"fee"`
	Nonce    Nonce       `meddler:"nonce"`
	Type     TxType      `meddler:"tx_type"`
}

// Tx returns a *Tx from the L2Tx
func (tx *L2Tx) Tx() *Tx {
	return &Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		Nonce:   tx.Nonce,
		Fee:     tx.Fee,
		Type:    tx.Type,
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
