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
