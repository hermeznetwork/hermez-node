package common

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/utils"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

// Nonce represents the nonce value in a uint64, which has the method Bytes that returns a byte array of length 5 (40 bits).
type Nonce uint64

// Bytes returns a byte array of length 5 representing the Nonce
func (n Nonce) Bytes() ([5]byte, error) {
	if n > maxNonceValue {
		return [5]byte{}, ErrNonceOverflow
	}
	var nonceBytes [8]byte
	binary.LittleEndian.PutUint64(nonceBytes[:], uint64(n))
	var b [5]byte
	copy(b[:], nonceBytes[:5])
	return b, nil
}

// NonceFromBytes returns Nonce from a [5]byte
func NonceFromBytes(b [5]byte) Nonce {
	var nonceBytes [8]byte
	copy(nonceBytes[:], b[:5])
	nonce := binary.LittleEndian.Uint64(nonceBytes[:])
	return Nonce(nonce)
}

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the coordinator hat is waiting to be forged
type PoolL2Tx struct {
	// Stored in DB: mandatory fileds
	TxID      TxID               `meddler:"tx_id"`
	FromIdx   Idx                `meddler:"from_idx"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx     Idx                `meddler:"to_idx"`   // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	ToEthAddr ethCommon.Address  `meddler:"to_eth_addr"`
	ToBJJ     *babyjub.PublicKey `meddler:"to_bjj"` // TODO: stop using json, use scanner/valuer
	TokenID   TokenID            `meddler:"token_id"`
	Amount    *big.Int           `meddler:"amount,bigint"` // TODO: change to float16
	Fee       FeeSelector        `meddler:"fee"`
	Nonce     Nonce              `meddler:"nonce"` // effective 40 bits used
	State     PoolL2TxState      `meddler:"state"`
	Signature *babyjub.Signature `meddler:"signature"`         // tx signature
	Timestamp time.Time          `meddler:"timestamp,utctime"` // time when added to the tx pool
	// Stored in DB: optional fileds, may be uninitialized
	BatchNum          BatchNum           `meddler:"batch_num,zeroisnull"`   // batchNum in which this tx was forged. Presence indicates "forged" state.
	RqFromIdx         Idx                `meddler:"rq_from_idx,zeroisnull"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	RqToIdx           Idx                `meddler:"rq_to_idx,zeroisnull"`   // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	RqToEthAddr       ethCommon.Address  `meddler:"rq_to_eth_addr"`
	RqToBJJ           *babyjub.PublicKey `meddler:"rq_to_bjj"` // TODO: stop using json, use scanner/valuer
	RqTokenID         TokenID            `meddler:"rq_token_id,zeroisnull"`
	RqAmount          *big.Int           `meddler:"rq_amount,bigintnull"` // TODO: change to float16
	RqFee             FeeSelector        `meddler:"rq_fee,zeroisnull"`
	RqNonce           uint64             `meddler:"rq_nonce,zeroisnull"` // effective 48 bits used
	AbsoluteFee       float64            `meddler:"absolute_fee,zeroisnull"`
	AbsoluteFeeUpdate time.Time          `meddler:"absolute_fee_update,utctimez"`
	Type              TxType             `meddler:"tx_type"`
	// Extra metadata, may be uninitialized
	RqTxCompressedData []byte `meddler:"-"` // 253 bits, optional for atomic txs
}

// TxCompressedData spec:
// [ 32 bits ] signatureConstant // 4 bytes: [0:4]
// [ 16 bits ] chainId // 2 bytes: [4:6]
// [ 48 bits ] fromIdx // 6 bytes: [6:12]
// [ 48 bits ] toIdx // 6 bytes: [12:18]
// [ 16 bits ] amountFloat16 // 2 bytes: [18:20]
// [ 32 bits ] tokenID // 4 bytes: [20:24]
// [ 40 bits ] nonce // 5 bytes: [24:29]
// [ 8 bits  ] userFee // 1 byte: [29:30]
// [ 1 bits  ] toBjjSign // 1 byte: [30:31]
// Total bits compressed data:  241 bits // 31 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedData() (*big.Int, error) {
	// sigconstant
	sc, ok := new(big.Int).SetString("3322668559", 10)
	if !ok {
		return nil, fmt.Errorf("error parsing SignatureConstant")
	}

	amountFloat16, err := utils.NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	var b [31]byte
	copy(b[:4], SwapEndianness(sc.Bytes()))
	copy(b[4:6], []byte{1, 0, 0, 0}) // LittleEndian representation of uint32(1) for Ethereum
	copy(b[6:12], tx.FromIdx.Bytes())
	copy(b[12:18], tx.ToIdx.Bytes())
	copy(b[18:20], amountFloat16.Bytes())
	copy(b[20:24], tx.TokenID.Bytes())
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[24:29], nonceBytes[:])
	b[29] = byte(tx.Fee)
	toBjjSign := byte(0)
	if babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBjjSign = byte(1)
	}
	b[30] = toBjjSign
	bi := new(big.Int).SetBytes(SwapEndianness(b[:]))

	return bi, nil
}

// TxCompressedDataV2 spec:
// [ 48 bits ] fromIdx // 6 bytes: [0:6]
// [ 48 bits ] toIdx // 6 bytes: [6:12]
// [ 16 bits ] amountFloat16 // 2 bytes: [12:14]
// [ 32 bits ] tokenID // 4 bytes: [14:18]
// [ 40 bits ] nonce // 5 bytes: [18:23]
// [ 8 bits  ] userFee // 1 byte: [23:24]
// [ 1 bits  ] toBjjSign // 1 byte: [24:25]
// Total bits compressed data:  193 bits // 25 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedDataV2() (*big.Int, error) {
	amountFloat16, err := utils.NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	var b [25]byte
	copy(b[0:6], tx.FromIdx.Bytes())
	copy(b[6:12], tx.ToIdx.Bytes())
	copy(b[12:14], amountFloat16.Bytes())
	copy(b[14:18], tx.TokenID.Bytes())
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[18:23], nonceBytes[:])
	b[23] = byte(tx.Fee)
	toBjjSign := byte(0)
	if babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBjjSign = byte(1)
	}
	b[24] = toBjjSign

	bi := new(big.Int).SetBytes(SwapEndianness(b[:]))
	return bi, nil
}

// HashToSign returns the computed Poseidon hash from the *PoolL2Tx that will be signed by the sender.
func (tx *PoolL2Tx) HashToSign() (*big.Int, error) {
	toCompressedData, err := tx.TxCompressedData()
	if err != nil {
		return nil, err
	}
	toEthAddr := EthAddrToBigInt(tx.ToEthAddr)
	toBjjAy := tx.ToBJJ.Y
	rqTxCompressedDataV2, err := tx.TxCompressedDataV2()
	if err != nil {
		return nil, err
	}

	return poseidon.Hash([]*big.Int{toCompressedData, toEthAddr, toBjjAy, rqTxCompressedDataV2, EthAddrToBigInt(tx.RqToEthAddr), tx.RqToBJJ.Y})
}

// VerifySignature returns true if the signature verification is correct for the given PublicKey
func (tx *PoolL2Tx) VerifySignature(pk *babyjub.PublicKey) bool {
	h, err := tx.HashToSign()
	if err != nil {
		return false
	}
	return pk.VerifyPoseidon(h, tx.Signature)
}

func (tx *PoolL2Tx) L2Tx() *L2Tx {
	return &L2Tx{
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

func (tx *PoolL2Tx) Tx() *Tx {
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

func PoolL2TxsToL2Txs(txs []*PoolL2Tx) []*L2Tx {
	var r []*L2Tx
	for _, tx := range txs {
		r = append(r, tx.L2Tx())
	}
	return r
}

// PoolL2TxState is a struct that represents the status of a L2 transaction
type PoolL2TxState string

const (
	// PoolL2TxStatePending represents a valid L2Tx that hasn't started the forging process
	PoolL2TxStatePending PoolL2TxState = "pend"
	// PoolL2TxStateForging represents a valid L2Tx that has started the forging process
	PoolL2TxStateForging PoolL2TxState = "fing"
	// PoolL2TxStateForged represents a L2Tx that has already been forged
	PoolL2TxStateForged PoolL2TxState = "fged"
	// PoolL2TxStateInvalid represents a L2Tx that has been invalidated
	PoolL2TxStateInvalid PoolL2TxState = "invl"
)
