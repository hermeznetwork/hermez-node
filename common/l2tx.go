package common

import (
	"fmt"
	"math/big"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/tracerr"
)

// L2Tx is a struct that represents an already forged L2 tx
type L2Tx struct {
	// Stored in DB: mandatory fields
	TxID     TxID     `meddler:"id"`
	BatchNum BatchNum `meddler:"batch_num"` // batchNum in which this tx was forged.
	Position int      `meddler:"position"`
	FromIdx  Idx      `meddler:"from_idx"`
	ToIdx    Idx      `meddler:"to_idx"`
	// TokenID is filled by the TxProcessor
	TokenID TokenID     `meddler:"token_id"`
	Amount  *big.Int    `meddler:"amount,bigint"`
	Fee     FeeSelector `meddler:"fee"`
	// Nonce is filled by the TxProcessor
	Nonce nonce.Nonce `meddler:"nonce"`
	Type  TxType      `meddler:"type"`
	// EthBlockNum in which this L2Tx was added to the queue
	EthBlockNum int64 `meddler:"eth_block_num"`
}

// NewL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewL2Tx(tx *L2Tx) (*L2Tx, error) {
	txTypeOld := tx.Type
	if err := tx.SetType(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original Type doesn't match the correct one, return error
	if txTypeOld != "" && txTypeOld != tx.Type {
		return nil, tracerr.Wrap(fmt.Errorf("L2Tx.Type: %s, should be: %s",
			tx.Type, txTypeOld))
	}

	txIDOld := tx.TxID
	if err := tx.SetID(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original TxID doesn't match the correct one, return error
	if txIDOld != (TxID{}) && txIDOld != tx.TxID {
		return tx, tracerr.Wrap(fmt.Errorf("L2Tx.TxID: %s, should be: %s",
			tx.TxID.String(), txIDOld.String()))
	}

	return tx, nil
}

// SetType sets the type of the transaction.  Uses (FromIdx, Nonce).
func (tx *L2Tx) SetType() error {
	if tx.ToIdx == Idx(1) {
		tx.Type = TxTypeExit
	} else if tx.ToIdx >= IdxUserThreshold {
		tx.Type = TxTypeTransfer
	} else {
		return tracerr.Wrap(fmt.Errorf(
			"cannot determine type of L2Tx, invalid ToIdx value: %d", tx.ToIdx))
	}
	return nil
}

// SetID sets the ID of the transaction
func (tx *L2Tx) SetID() error {
	txID, err := tx.CalculateTxID()
	if err != nil {
		return err
	}
	tx.TxID = txID
	return nil
}

// CalculateTxID returns the TxID of the transaction. This method is used to
// set the TxID for L2Tx and for PoolL2Tx.
func (tx L2Tx) CalculateTxID() ([TxIDLen]byte, error) {
	var txID TxID
	var b []byte
	// FromIdx
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return txID, tracerr.Wrap(err)
	}
	b = append(b, fromIdxBytes[:]...)
	// TokenID
	b = append(b, tx.TokenID.Bytes()[:]...)
	// Amount
	amountFloat40, err := NewFloat40(tx.Amount)
	if err != nil {
		return txID, tracerr.Wrap(fmt.Errorf("%s: %d", err, tx.Amount))
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return txID, tracerr.Wrap(err)
	}
	b = append(b, amountFloat40Bytes...)
	// Nonce
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return txID, tracerr.Wrap(err)
	}
	b = append(b, nonceBytes[:]...)
	// Fee
	b = append(b, byte(tx.Fee))

	// calculate hash
	h := ethCrypto.Keccak256Hash(b).Bytes()

	txID[0] = TxIDPrefixL2Tx
	copy(txID[1:], h)
	return txID, nil
}

// Tx returns a *Tx from the L2Tx
func (tx *L2Tx) Tx() *Tx {
	batchNum := new(BatchNum)
	*batchNum = tx.BatchNum
	fee := new(FeeSelector)
	*fee = tx.Fee
	nonce := new(nonce.Nonce)
	*nonce = tx.Nonce
	return &Tx{
		IsL1:        false,
		TxID:        tx.TxID,
		Type:        tx.Type,
		Position:    tx.Position,
		FromIdx:     tx.FromIdx,
		ToIdx:       tx.ToIdx,
		TokenID:     tx.TokenID,
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
		TokenID: tx.TokenID,
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

// TxIDsFromL2Txs returns an array of TxID from the []L2Tx
func TxIDsFromL2Txs(txs []L2Tx) []TxID {
	txIDs := make([]TxID, len(txs))
	for i, tx := range txs {
		txIDs[i] = tx.TxID
	}
	return txIDs
}

// BytesDataAvailability encodes a L2Tx into []byte for the Data Availability
// [ fromIdx | toIdx | amountFloat40 | Fee ]
func (tx L2Tx) BytesDataAvailability(nLevels uint32) ([]byte, error) {
	idxLen := nLevels / 8 //nolint:gomnd

	b := make([]byte, ((nLevels*2)+40+8)/8) //nolint:gomnd

	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[0:idxLen], fromIdxBytes[6-idxLen:]) // [6-idxLen:] as is BigEndian

	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[idxLen:idxLen*2], toIdxBytes[6-idxLen:])

	amountFloat40, err := NewFloat40(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[idxLen*2:idxLen*2+Float40BytesLength], amountFloat40Bytes)
	b[idxLen*2+Float40BytesLength] = byte(tx.Fee)

	return b[:], nil
}

// L2TxFromBytesDataAvailability decodes a L2Tx from []byte (Data Availability)
func L2TxFromBytesDataAvailability(b []byte, nLevels int) (*L2Tx, error) {
	idxLen := nLevels / 8 //nolint:gomnd
	tx := &L2Tx{}
	var err error

	var paddedFromIdxBytes [6]byte
	copy(paddedFromIdxBytes[6-idxLen:], b[0:idxLen])
	tx.FromIdx, err = IdxFromBytes(paddedFromIdxBytes[:])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var paddedToIdxBytes [6]byte
	copy(paddedToIdxBytes[6-idxLen:6], b[idxLen:idxLen*2])
	tx.ToIdx, err = IdxFromBytes(paddedToIdxBytes[:])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	tx.Amount, err = Float40FromBytes(b[idxLen*2 : idxLen*2+Float40BytesLength]).BigInt()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.Fee = FeeSelector(b[idxLen*2+Float40BytesLength])
	return tx, nil
}
