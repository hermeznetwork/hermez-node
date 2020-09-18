package common

import (
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// L1TxBytesLen is the length of the byte array that represents the L1Tx
	L1TxBytesLen = 68
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	// Stored in DB: mandatory fileds
	TxID            TxID
	ToForgeL1TxsNum uint32 // toForgeL1TxsNum in which the tx was forged / will be forged
	Position        int
	UserOrigin      bool // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromIdx         Idx  // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	FromEthAddr     ethCommon.Address
	FromBJJ         *babyjub.PublicKey
	ToIdx           Idx // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	TokenID         TokenID
	Amount          *big.Int
	LoadAmount      *big.Int
	EthBlockNum     int64 // Ethereum Block Number in which this L1Tx was added to the queue
	Type            TxType
	BatchNum        BatchNum
}

// Tx returns a *Tx from the L1Tx
func (tx *L1Tx) Tx() *Tx {
	f := new(big.Float).SetInt(tx.Amount)
	amountFloat, _ := f.Float64()
	genericTx := &Tx{
		IsL1:            true,
		TxID:            tx.TxID,
		Type:            tx.Type,
		Position:        tx.Position,
		FromIdx:         tx.FromIdx,
		ToIdx:           tx.ToIdx,
		Amount:          tx.Amount,
		AmountFloat:     amountFloat,
		TokenID:         tx.TokenID,
		ToForgeL1TxsNum: tx.ToForgeL1TxsNum,
		UserOrigin:      tx.UserOrigin,
		FromEthAddr:     tx.FromEthAddr,
		FromBJJ:         tx.FromBJJ,
		LoadAmount:      tx.LoadAmount,
		EthBlockNum:     tx.EthBlockNum,
	}
	if tx.LoadAmount != nil {
		lf := new(big.Float).SetInt(tx.LoadAmount)
		loadAmountFloat, _ := lf.Float64()
		genericTx.LoadAmountFloat = loadAmountFloat
	}
	return genericTx
}

// Bytes encodes a L1Tx into []byte
func (tx *L1Tx) Bytes(nLevels int) ([]byte, error) {
	var b [68]byte
	copy(b[0:20], tx.FromEthAddr.Bytes())
	pkComp := tx.FromBJJ.Compress()
	copy(b[20:52], pkComp[:])
	copy(b[52:56], tx.FromIdx.Bytes())
	loadAmountFloat16, err := NewFloat16(tx.LoadAmount)
	if err != nil {
		return nil, err
	}
	copy(b[56:58], loadAmountFloat16.Bytes())
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	copy(b[58:60], amountFloat16.Bytes())
	copy(b[60:64], tx.TokenID.Bytes())
	copy(b[64:68], tx.ToIdx.Bytes())
	return b[:], nil
}

// L1TxFromBytes decodes a L1Tx from []byte
func L1TxFromBytes(b []byte) (*L1Tx, error) {
	if len(b) != L1TxBytesLen {
		return nil, fmt.Errorf("Can not parse L1Tx bytes, expected length %d, current: %d", 68, len(b))
	}

	tx := &L1Tx{}
	var err error
	tx.FromEthAddr = ethCommon.BytesToAddress(b[0:20])
	pkCompB := b[20:52]
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkCompB)
	tx.FromBJJ, err = pkComp.Decompress()
	if err != nil {
		return nil, err
	}
	tx.FromIdx, err = IdxFromBytes(b[52:56])
	if err != nil {
		return nil, err
	}
	tx.LoadAmount = Float16FromBytes(b[56:58]).BigInt()
	tx.Amount = Float16FromBytes(b[58:60]).BigInt()
	tx.TokenID, err = TokenIDFromBytes(b[60:64])
	if err != nil {
		return nil, err
	}
	tx.ToIdx, err = IdxFromBytes(b[64:68])
	if err != nil {
		return nil, err
	}

	return tx, nil
}
