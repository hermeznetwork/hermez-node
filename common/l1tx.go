package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// L1TxBytesLen is the length of the byte array that represents the L1Tx
	L1TxBytesLen = 72
	// L1CoordinatorTxBytesLen is the length of the byte array that represents the L1CoordinatorTx
	L1CoordinatorTxBytesLen = 101
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	// Stored in DB: mandatory fileds

	// TxID (12 bytes) for L1Tx is:
	// bytes:  |  1   |        8        |    2     |      1      |
	// values: | type | ToForgeL1TxsNum | Position | 0 (padding) |
	// where type:
	// 	- L1UserTx: 0
	// 	- L1CoordinatorTx: 1
	TxID            TxID
	ToForgeL1TxsNum int64 // toForgeL1TxsNum in which the tx was forged / will be forged
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
	BatchNum        *BatchNum
	USD             *float64
	LoadAmountUSD   *float64
}

// NewL1Tx returns the given L1Tx with the TxId & Type parameters calculated
// from the L1Tx values
func NewL1Tx(l1Tx *L1Tx) (*L1Tx, error) {
	// calculate TxType
	var txType TxType
	if l1Tx.FromIdx == Idx(0) {
		if l1Tx.ToIdx == Idx(0) {
			txType = TxTypeCreateAccountDeposit
		} else if l1Tx.ToIdx >= IdxUserThreshold {
			txType = TxTypeCreateAccountDepositTransfer
		} else {
			return l1Tx, fmt.Errorf("Can not determine type of L1Tx, invalid ToIdx value: %d", l1Tx.ToIdx)
		}
	} else if l1Tx.FromIdx >= IdxUserThreshold {
		if l1Tx.ToIdx == Idx(0) {
			txType = TxTypeDeposit
		} else if l1Tx.ToIdx == Idx(1) {
			txType = TxTypeExit
		} else if l1Tx.ToIdx >= IdxUserThreshold {
			if l1Tx.LoadAmount.Int64() == int64(0) {
				txType = TxTypeForceTransfer
			} else {
				txType = TxTypeDepositTransfer
			}
		} else {
			return l1Tx, fmt.Errorf("Can not determine type of L1Tx, invalid ToIdx value: %d", l1Tx.ToIdx)
		}
	} else {
		return l1Tx, fmt.Errorf("Can not determine type of L1Tx, invalid FromIdx value: %d", l1Tx.FromIdx)
	}

	if l1Tx.Type != "" && l1Tx.Type != txType {
		return l1Tx, fmt.Errorf("L1Tx.Type: %s, should be: %s", l1Tx.Type, txType)
	}
	l1Tx.Type = txType

	var txid [TxIDLen]byte
	if !l1Tx.UserOrigin {
		txid[0] = TxIDPrefixL1CoordTx
	}
	var toForgeL1TxsNumBytes [8]byte
	binary.BigEndian.PutUint64(toForgeL1TxsNumBytes[:], uint64(l1Tx.ToForgeL1TxsNum))
	copy(txid[1:9], toForgeL1TxsNumBytes[:])

	var positionBytes [2]byte
	binary.BigEndian.PutUint16(positionBytes[:], uint16(l1Tx.Position))
	copy(txid[9:11], positionBytes[:])
	l1Tx.TxID = TxID(txid)

	return l1Tx, nil
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
		USD:             tx.USD,
		LoadAmountUSD:   tx.LoadAmountUSD,
	}
	if tx.LoadAmount != nil {
		lf := new(big.Float).SetInt(tx.LoadAmount)
		loadAmountFloat, _ := lf.Float64()
		genericTx.LoadAmountFloat = &loadAmountFloat
	}
	return genericTx
}

// Bytes encodes a L1Tx into []byte
func (tx *L1Tx) Bytes() ([]byte, error) {
	var b [L1TxBytesLen]byte
	copy(b[0:20], tx.FromEthAddr.Bytes())
	pkComp := tx.FromBJJ.Compress()
	copy(b[20:52], pkComp[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[52:58], fromIdxBytes[:])
	loadAmountFloat16, err := NewFloat16(tx.LoadAmount)
	if err != nil {
		return nil, err
	}
	copy(b[58:60], loadAmountFloat16.Bytes())
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	copy(b[60:62], amountFloat16.Bytes())
	copy(b[62:66], tx.TokenID.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[66:72], toIdxBytes[:])
	return b[:], nil
}

// BytesCoordinatorTx encodes a L1CoordinatorTx into []byte
func (tx *L1Tx) BytesCoordinatorTx(compressedSignatureBytes []byte) ([]byte, error) {
	var b [L1CoordinatorTxBytesLen]byte
	v := compressedSignatureBytes[64]
	s := compressedSignatureBytes[32:64]
	r := compressedSignatureBytes[0:32]
	b[0] = v
	copy(b[1:33], s)
	copy(b[33:65], r)
	pkComp := tx.FromBJJ.Compress()
	copy(b[65:97], pkComp[:])
	copy(b[97:101], tx.TokenID.Bytes())
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
	tx.FromIdx, err = IdxFromBytes(b[52:58])
	if err != nil {
		return nil, err
	}
	tx.LoadAmount = Float16FromBytes(b[58:60]).BigInt()
	tx.Amount = Float16FromBytes(b[60:62]).BigInt()
	tx.TokenID, err = TokenIDFromBytes(b[62:66])
	if err != nil {
		return nil, err
	}
	tx.ToIdx, err = IdxFromBytes(b[66:72])
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// L1TxFromCoordinatorBytes decodes a L1Tx from []byte
func L1TxFromCoordinatorBytes(b []byte) (*L1Tx, error) {
	if len(b) != L1CoordinatorTxBytesLen {
		return nil, fmt.Errorf("Can not parse L1CoordinatorTx bytes, expected length %d, current: %d", 101, len(b))
	}

	bytesMessage1 := []byte("\x19Ethereum Signed Message:\n98")
	bytesMessage2 := []byte("I authorize this babyjubjub key for hermez rollup account creation")

	tx := &L1Tx{}
	var err error
	// Ethereum adds 27 to v
	v := byte(int(b[0]) - 27) //nolint:gomnd
	s := b[1:33]
	r := b[33:65]

	pkCompB := b[65:97]
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkCompB)
	tx.FromBJJ, err = pkComp.Decompress()
	if err != nil {
		return nil, err
	}
	tx.TokenID, err = TokenIDFromBytes(b[97:101])
	if err != nil {
		return nil, err
	}

	var data []byte
	data = append(data, bytesMessage1...)
	data = append(data, bytesMessage2...)
	data = append(data, pkCompB...)
	var signature []byte
	signature = append(signature, r[:]...)
	signature = append(signature, s[:]...)
	signature = append(signature, v)
	hash := crypto.Keccak256(data)
	pubKeyBytes, err := crypto.Ecrecover(hash, signature)
	if err != nil {
		return nil, err
	}
	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, err
	}
	tx.FromEthAddr = crypto.PubkeyToAddress(*pubKey)
	return tx, nil
}
