package common

import (
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/utils"
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
	copy(b[0:4], tx.ToIdx.Bytes())
	copy(b[4:8], tx.TokenID.Bytes())
	amountFloat16, err := utils.NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	copy(b[8:10], amountFloat16.Bytes())
	loadAmountFloat16, err := utils.NewFloat16(tx.LoadAmount)
	if err != nil {
		return nil, err
	}
	copy(b[10:12], loadAmountFloat16.Bytes())
	copy(b[12:16], tx.FromIdx.Bytes())
	pkComp := tx.FromBJJ.Compress()
	copy(b[16:48], SwapEndianness(pkComp[:]))
	copy(b[48:68], SwapEndianness(tx.FromEthAddr.Bytes()))
	return SwapEndianness(b[:]), nil
}

// L1TxFromBytes decodes a L1Tx from []byte
func L1TxFromBytes(bRaw []byte) (*L1Tx, error) {
	if len(bRaw) != L1TxBytesLen {
		return nil, fmt.Errorf("Can not parse L1Tx bytes, expected length %d, current: %d", 68, len(bRaw))
	}

	b := SwapEndianness(bRaw)
	tx := &L1Tx{}
	var err error
	tx.ToIdx, err = IdxFromBytes(b[0:4])
	if err != nil {
		return nil, err
	}
	tx.TokenID, err = TokenIDFromBytes(b[4:8])
	if err != nil {
		return nil, err
	}
	tx.Amount = new(big.Int).SetBytes(SwapEndianness(b[8:10]))
	tx.LoadAmount = new(big.Int).SetBytes(SwapEndianness(b[10:12]))
	tx.FromIdx, err = IdxFromBytes(b[12:16])
	if err != nil {
		return nil, err
	}
	pkCompB := SwapEndianness(b[16:48])
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkCompB)
	tx.FromBJJ, err = pkComp.Decompress()
	if err != nil {
		return nil, err
	}
	tx.FromEthAddr = ethCommon.BytesToAddress(SwapEndianness(b[48:68]))
	return tx, nil
}
