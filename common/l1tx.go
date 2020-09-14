package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
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
	amountF, _ := f.Float32()
	return &Tx{
		IsL1:            true,
		TxID:            tx.TxID,
		Type:            tx.Type,
		Position:        tx.Position,
		FromIdx:         tx.FromIdx,
		ToIdx:           tx.ToIdx,
		Amount:          tx.Amount,
		AmountF:         amountF,
		TokenID:         tx.TokenID,
		ToForgeL1TxsNum: tx.ToForgeL1TxsNum,
		UserOrigin:      tx.UserOrigin,
		FromEthAddr:     tx.FromEthAddr,
		FromBJJ:         tx.FromBJJ,
		LoadAmount:      tx.LoadAmount,
		EthBlockNum:     tx.EthBlockNum,
	}
}
