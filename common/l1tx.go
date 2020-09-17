package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	fromBJJCompressedB = 256
	fromEthAddrB       = 160
	f16B               = 16
	tokenIDB           = 32
	cidXB              = 32
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
func (tx *L1Tx) Bytes(nLevels int) []byte {
	res := big.NewInt(0)
	res = res.Add(res, big.NewInt(0).Or(big.NewInt(0), tx.ToIdx.BigInt()))
	res = res.Add(res, big.NewInt(0).Lsh(big.NewInt(0).Or(big.NewInt(0), big.NewInt(int64(tx.TokenID))), uint(nLevels)))
	res = res.Add(res, big.NewInt(0).Lsh(big.NewInt(0).Or(big.NewInt(0), tx.Amount), uint(nLevels+tokenIDB)))
	res = res.Add(res, big.NewInt(0).Lsh(big.NewInt(0).Or(big.NewInt(0), tx.LoadAmount), uint(nLevels+tokenIDB+f16B)))
	res = res.Add(res, big.NewInt(0).Lsh(big.NewInt(0).Or(big.NewInt(0), tx.FromIdx.BigInt()), uint(nLevels+tokenIDB+2*f16B)))

	fromBJJ := big.NewInt(0)
	fromBJJ.SetString(tx.FromBJJ.String(), 16)
	fromBJJCompressed := big.NewInt(0).Or(big.NewInt(0), fromBJJ)
	res = res.Add(res, big.NewInt(0).Lsh(fromBJJCompressed, uint(2*nLevels+tokenIDB+2*f16B)))

	fromEthAddr := big.NewInt(0).Or(big.NewInt(0), tx.FromEthAddr.Hash().Big())
	res = res.Add(res, big.NewInt(0).Lsh(fromEthAddr, uint(fromBJJCompressedB+2*nLevels+tokenIDB+2*f16B)))

	return res.Bytes()
}

// L1TxFromBytes decodes a L1Tx from []byte
func L1TxFromBytes(l1TxEncoded []byte) (*L1Tx, error) {
	l1Tx := &L1Tx{}
	var idxB uint = cidXB

	l1TxEncodedBI := big.NewInt(0)
	l1TxEncodedBI.SetBytes(l1TxEncoded)

	toIdx, err := IdxFromBigInt(extract(l1TxEncodedBI, 0, idxB))

	if err != nil {
		return nil, err
	}

	l1Tx.ToIdx = toIdx

	l1Tx.TokenID = TokenID(extract(l1TxEncodedBI, idxB, tokenIDB).Uint64())
	l1Tx.Amount = extract(l1TxEncodedBI, idxB+tokenIDB, f16B)
	l1Tx.LoadAmount = extract(l1TxEncodedBI, idxB+tokenIDB+f16B, f16B)
	fromIdx, err := IdxFromBigInt(extract(l1TxEncodedBI, idxB+tokenIDB+2*f16B, f16B))

	if err != nil {
		return nil, err
	}

	l1Tx.FromIdx = fromIdx

	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], extract(l1TxEncodedBI, 2*idxB+tokenIDB+2*f16B, fromBJJCompressedB).Bytes())
	pk, err := pkComp.Decompress()

	if err != nil {
		return nil, err
	}

	l1Tx.FromBJJ = pk

	l1Tx.FromEthAddr = ethCommon.BigToAddress(extract(l1TxEncodedBI, fromBJJCompressedB+2*idxB+tokenIDB+2*f16B, fromEthAddrB))

	return l1Tx, nil
}

// extract masks and shifts a bigInt
func extract(num *big.Int, origin uint, len uint) *big.Int {
	mask := big.NewInt(0).Sub(big.NewInt(0).Lsh(big.NewInt(1), len), big.NewInt(1))
	return big.NewInt(0).And(big.NewInt(0).Rsh(num, origin), mask)
}
