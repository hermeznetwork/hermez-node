package common

import "math/big"

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	TxID     TxID        `meddler:"tx_id"`
	FromIdx  Idx         `meddler:"from_idx"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx    Idx         `meddler:"to_idx"`   // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	TokenID  TokenID     `meddler:"token_id"`
	Amount   *big.Int    `meddler:"amount,bigint"` // TODO: change to float16
	Nonce    uint64      `meddler:"nonce"`         // effective 48 bits used
	Fee      FeeSelector `meddler:"fee"`
	Type     TxType      `meddler:"-"`         // optional, descrives which kind of tx it's
	BatchNum BatchNum    `meddler:"batch_num"` // batchNum in which this tx was forged. Presence indicates "forged" state.
	Position int         // Position among all the L1Txs in that batch
}

// Tx implements the Txer interface
func (l2tx *L2Tx) Tx() Tx {
	return Tx{
		TxID:     l2tx.TxID,
		FromIdx:  l2tx.FromIdx,
		ToIdx:    l2tx.ToIdx,
		TokenID:  l2tx.TokenID,
		Amount:   l2tx.Amount,
		Nonce:    l2tx.Nonce,
		Fee:      l2tx.Fee,
		Type:     l2tx.Type,
		BatchNum: l2tx.BatchNum,
	}
}
