package common

import (
	"fmt"
	"math/big"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	// Stored in DB: mandatory fileds
	TxID        TxID        `meddler:"id"`
	BatchNum    BatchNum    `meddler:"batch_num"` // batchNum in which this tx was forged.
	Position    int         `meddler:"position"`
	FromIdx     Idx         `meddler:"from_idx"`
	ToIdx       Idx         `meddler:"to_idx"`
	Amount      *big.Int    `meddler:"amount,bigint"`
	Fee         FeeSelector `meddler:"fee"`
	Nonce       Nonce       `meddler:"nonce"`
	Type        TxType      `meddler:"type"`
	EthBlockNum int64       `meddler:"eth_block_num"` // Ethereum Block Number in which this L2Tx was added to the queue
}

// NewL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewL2Tx(l2Tx *L2Tx) (*L2Tx, error) {
	// calculate TxType
	var txType TxType
	if l2Tx.ToIdx == Idx(1) {
		txType = TxTypeExit
	} else if l2Tx.ToIdx >= IdxUserThreshold {
		txType = TxTypeTransfer
	} else {
		return l2Tx, fmt.Errorf("Can not determine type of L2Tx, invalid ToIdx value: %d", l2Tx.ToIdx)
	}

	// if TxType!=l2Tx.TxType return error
	if l2Tx.Type != "" && l2Tx.Type != txType {
		return l2Tx, fmt.Errorf("L2Tx.Type: %s, should be: %s", l2Tx.Type, txType)
	}
	l2Tx.Type = txType

	var txid [TxIDLen]byte
	txid[0] = TxIDPrefixL2Tx
	fromIdxBytes, err := l2Tx.FromIdx.Bytes()
	if err != nil {
		return l2Tx, err
	}
	copy(txid[1:7], fromIdxBytes[:])
	nonceBytes, err := l2Tx.Nonce.Bytes()
	if err != nil {
		return l2Tx, err
	}
	copy(txid[7:12], nonceBytes[:])
	l2Tx.TxID = TxID(txid)

	return l2Tx, nil
}

// Tx returns a *Tx from the L2Tx
func (tx *L2Tx) Tx() *Tx {
	batchNum := new(BatchNum)
	*batchNum = tx.BatchNum
	fee := new(FeeSelector)
	*fee = tx.Fee
	nonce := new(Nonce)
	*nonce = tx.Nonce
	return &Tx{
		IsL1:        false,
		TxID:        tx.TxID,
		Type:        tx.Type,
		Position:    tx.Position,
		FromIdx:     tx.FromIdx,
		ToIdx:       tx.ToIdx,
		Amount:      tx.Amount,
		BatchNum:    batchNum,
		EthBlockNum: tx.EthBlockNum,
		Fee:         fee,
		Nonce:       nonce,
	}
}

// PoolL2Tx returns the data structure of PoolL2Tx with the parameters of a
// L2Tx filled
func (tx L2Tx) PoolL2Tx() *PoolL2Tx {
	return &PoolL2Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		Fee:     tx.Fee,
		Nonce:   tx.Nonce,
		Type:    tx.Type,
	}
}

// L2TxsToPoolL2Txs returns an array of []*PoolL2Tx from an array of []*L2Tx,
// where the PoolL2Tx only have the parameters of a L2Tx filled.
func L2TxsToPoolL2Txs(txs []L2Tx) []PoolL2Tx {
	var r []PoolL2Tx
	for _, tx := range txs {
		r = append(r, *tx.PoolL2Tx())
	}
	return r
}

// Bytes encodes a L2Tx into []byte
func (tx L2Tx) Bytes(nLevels uint32) ([]byte, error) {
	idxLen := nLevels / 8 //nolint:gomnd

	b := make([]byte, ((nLevels*2)+16+8)/8) //nolint:gomnd

	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[0:idxLen], fromIdxBytes[6-idxLen:]) // [6-idxLen:] as is BigEndian

	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[idxLen:idxLen*2], toIdxBytes[6-idxLen:])

	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}

	copy(b[idxLen*2:idxLen*2+2], amountFloat16.Bytes())
	b[idxLen*2+2] = byte(tx.Fee)

	return b[:], nil
}

// L2TxFromBytes decodes a L1Tx from []byte
func L2TxFromBytes(b []byte, nLevels int) (*L2Tx, error) {
	idxLen := nLevels / 8 //nolint:gomnd
	tx := &L2Tx{}
	var err error

	var paddedFromIdxBytes [6]byte
	copy(paddedFromIdxBytes[6-idxLen:], b[0:idxLen])
	tx.FromIdx, err = IdxFromBytes(paddedFromIdxBytes[:])
	if err != nil {
		return nil, err
	}

	var paddedToIdxBytes [6]byte
	copy(paddedToIdxBytes[6-idxLen:6], b[idxLen:idxLen*2])
	tx.ToIdx, err = IdxFromBytes(paddedToIdxBytes[:])
	if err != nil {
		return nil, err
	}

	tx.Amount = Float16FromBytes(b[idxLen*2 : idxLen*2+2]).BigInt()
	tx.Fee = FeeSelector(b[idxLen*2+2])
	return tx, nil
}
