package common

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the coordinator hat is waiting to be forged
type PoolL2Tx struct {
	// Stored in DB: mandatory fileds

	// TxID (12 bytes) for L2Tx is:
	// bytes:  |  1   |    6    |   5   |
	// values: | type | FromIdx | Nonce |
	TxID      TxID                  `meddler:"tx_id"`
	FromIdx   Idx                   `meddler:"from_idx"`
	ToIdx     Idx                   `meddler:"to_idx,zeroisnull"`
	AuxToIdx  Idx                   `meddler:"-"` // AuxToIdx is only used internally at the StateDB to avoid repeated computation when processing transactions (from Synchronizer, TxSelector, BatchBuilder)
	ToEthAddr ethCommon.Address     `meddler:"to_eth_addr,zeroisnull"`
	ToBJJ     *babyjub.PublicKey    `meddler:"to_bjj"`
	TokenID   TokenID               `meddler:"token_id"`
	Amount    *big.Int              `meddler:"amount,bigint"` // TODO: change to float16
	Fee       FeeSelector           `meddler:"fee"`
	Nonce     Nonce                 `meddler:"nonce"` // effective 40 bits used
	State     PoolL2TxState         `meddler:"state"`
	Signature babyjub.SignatureComp `meddler:"signature"`         // tx signature
	Timestamp time.Time             `meddler:"timestamp,utctime"` // time when added to the tx pool
	// Stored in DB: optional fileds, may be uninitialized
	RqFromIdx         Idx                `meddler:"rq_from_idx,zeroisnull"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.DepositAmount (deposit)
	RqToIdx           Idx                `meddler:"rq_to_idx,zeroisnull"`   // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.DepositAmount (deposit)
	RqToEthAddr       ethCommon.Address  `meddler:"rq_to_eth_addr,zeroisnull"`
	RqToBJJ           *babyjub.PublicKey `meddler:"rq_to_bjj"` // TODO: stop using json, use scanner/valuer
	RqTokenID         TokenID            `meddler:"rq_token_id,zeroisnull"`
	RqAmount          *big.Int           `meddler:"rq_amount,bigintnull"` // TODO: change to float16
	RqFee             FeeSelector        `meddler:"rq_fee,zeroisnull"`
	RqNonce           Nonce              `meddler:"rq_nonce,zeroisnull"` // effective 48 bits used
	AbsoluteFee       float64            `meddler:"fee_usd,zeroisnull"`
	AbsoluteFeeUpdate time.Time          `meddler:"usd_update,utctimez"`
	Type              TxType             `meddler:"tx_type"`
	// Extra metadata, may be uninitialized
	RqTxCompressedData []byte `meddler:"-"` // 253 bits, optional for atomic txs
}

// NewPoolL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewPoolL2Tx(tx *PoolL2Tx) (*PoolL2Tx, error) {
	txTypeOld := tx.Type
	if err := tx.SetType(); err != nil {
		return nil, err
	}
	// If original Type doesn't match the correct one, return error
	if txTypeOld != "" && txTypeOld != tx.Type {
		return nil, tracerr.Wrap(fmt.Errorf("L2Tx.Type: %s, should be: %s",
			tx.Type, txTypeOld))
	}

	txIDOld := tx.TxID
	if err := tx.SetID(); err != nil {
		return nil, err
	}
	// If original TxID doesn't match the correct one, return error
	if txIDOld != (TxID{}) && txIDOld != tx.TxID {
		return tx, tracerr.Wrap(fmt.Errorf("PoolL2Tx.TxID: %s, should be: %s",
			tx.TxID.String(), txIDOld.String()))
	}

	return tx, nil
}

// SetType sets the type of the transaction
func (tx *PoolL2Tx) SetType() error {
	if tx.ToIdx >= IdxUserThreshold {
		tx.Type = TxTypeTransfer
	} else if tx.ToIdx == 1 {
		tx.Type = TxTypeExit
	} else if tx.ToIdx == 0 {
		if tx.ToBJJ != nil && tx.ToEthAddr == FFAddr {
			tx.Type = TxTypeTransferToBJJ
		} else if tx.ToEthAddr != FFAddr && tx.ToEthAddr != EmptyAddr {
			tx.Type = TxTypeTransferToEthAddr
		}
	} else {
		return tracerr.Wrap(errors.New("malformed transaction"))
	}
	return nil
}

// SetID sets the ID of the transaction.  Uses (FromIdx, Nonce).
func (tx *PoolL2Tx) SetID() error {
	tx.TxID[0] = TxIDPrefixL2Tx
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return tracerr.Wrap(err)
	}
	copy(tx.TxID[1:7], fromIdxBytes[:])
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return tracerr.Wrap(err)
	}
	copy(tx.TxID[7:12], nonceBytes[:])
	return nil
}

// TxCompressedData spec:
// [ 1 bits  ] toBJJSign // 1 byte
// [ 8 bits  ] userFee // 1 byte
// [ 40 bits ] nonce // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 16 bits ] amountFloat16 // 2 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// [ 16 bits ] chainId // 2 bytes
// [ 32 bits ] signatureConstant // 4 bytes
// Total bits compressed data:  241 bits // 31 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedData() (*big.Int, error) {
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var b [31]byte
	toBJJSign := byte(0)
	if tx.ToBJJ != nil && babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBJJSign = byte(1)
	}
	b[0] = toBJJSign
	b[1] = byte(tx.Fee)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.TokenID.Bytes())
	copy(b[11:13], amountFloat16.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[13:19], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[19:25], fromIdxBytes[:])
	copy(b[25:27], []byte{0, 0}) // TODO this will be generated by the ChainID config parameter
	copy(b[27:31], SignatureConstantBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// TxCompressedDataV2 spec:
// [ 1 bits  ] toBJJSign // 1 byte
// [ 8 bits  ] userFee // 1 byte
// [ 40 bits ] nonce // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 16 bits ] amountFloat16 // 2 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// Total bits compressed data:  193 bits // 25 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedDataV2() (*big.Int, error) {
	if tx.Amount == nil {
		tx.Amount = big.NewInt(0)
	}
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var b [25]byte
	toBJJSign := byte(0)
	if tx.ToBJJ != nil && babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBJJSign = byte(1)
	}
	b[0] = toBJJSign
	b[1] = byte(tx.Fee)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.TokenID.Bytes())
	copy(b[11:13], amountFloat16.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[13:19], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[19:25], fromIdxBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// RqTxCompressedDataV2 is like the TxCompressedDataV2 but using the 'Rq'
// parameters. In a future iteration of the hermez-node, the 'Rq' parameters
// can be inside a struct, which contains the 'Rq' transaction grouped inside,
// so then computing the 'RqTxCompressedDataV2' would be just calling
// 'tx.Rq.TxCompressedDataV2()'.
// RqTxCompressedDataV2 spec:
// [ 1 bits  ] rqToBJJSign // 1 byte
// [ 8 bits  ] rqUserFee // 1 byte
// [ 40 bits ] rqNonce // 5 bytes
// [ 32 bits ] rqTokenID // 4 bytes
// [ 16 bits ] rqAmountFloat16 // 2 bytes
// [ 48 bits ] rqToIdx // 6 bytes
// [ 48 bits ] rqFromIdx // 6 bytes
// Total bits compressed data:  193 bits // 25 bytes in *big.Int representation
func (tx *PoolL2Tx) RqTxCompressedDataV2() (*big.Int, error) {
	if tx.RqAmount == nil {
		tx.RqAmount = big.NewInt(0)
	}
	amountFloat16, err := NewFloat16(tx.RqAmount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	var b [25]byte
	toBJJSign := byte(0)
	if tx.RqToBJJ != nil && babyjub.PointCoordSign(tx.RqToBJJ.X) {
		toBJJSign = byte(1)
	}
	b[0] = toBJJSign
	b[1] = byte(tx.RqFee)
	nonceBytes, err := tx.RqNonce.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.RqTokenID.Bytes())
	copy(b[11:13], amountFloat16.Bytes())
	toIdxBytes, err := tx.RqToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[13:19], toIdxBytes[:])
	fromIdxBytes, err := tx.RqFromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[19:25], fromIdxBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// HashToSign returns the computed Poseidon hash from the *PoolL2Tx that will be signed by the sender.
func (tx *PoolL2Tx) HashToSign() (*big.Int, error) {
	toCompressedData, err := tx.TxCompressedData()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	toEthAddr := EthAddrToBigInt(tx.ToEthAddr)
	rqToEthAddr := EthAddrToBigInt(tx.RqToEthAddr)
	toBJJY := big.NewInt(0)
	if tx.ToBJJ != nil {
		toBJJY = tx.ToBJJ.Y
	}
	rqTxCompressedDataV2, err := tx.RqTxCompressedDataV2()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	rqToBJJY := big.NewInt(0)
	if tx.RqToBJJ != nil {
		rqToBJJY = tx.RqToBJJ.Y
	}

	return poseidon.Hash([]*big.Int{toCompressedData, toEthAddr, toBJJY, rqTxCompressedDataV2, rqToEthAddr, rqToBJJY})
}

// VerifySignature returns true if the signature verification is correct for the given PublicKey
func (tx *PoolL2Tx) VerifySignature(pk *babyjub.PublicKey) bool {
	h, err := tx.HashToSign()
	if err != nil {
		return false
	}
	s, err := tx.Signature.Decompress()
	if err != nil {
		return false
	}
	return pk.VerifyPoseidon(h, s)
}

// L2Tx returns a *L2Tx from the PoolL2Tx
func (tx PoolL2Tx) L2Tx() L2Tx {
	return L2Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		Fee:     tx.Fee,
		Nonce:   tx.Nonce,
		Type:    tx.Type,
	}
}

// Tx returns a *Tx from the PoolL2Tx
func (tx PoolL2Tx) Tx() Tx {
	return Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		TokenID: tx.TokenID,
		Nonce:   &tx.Nonce,
		Fee:     &tx.Fee,
		Type:    tx.Type,
	}
}

// PoolL2TxsToL2Txs returns an array of []L2Tx from an array of []PoolL2Tx
func PoolL2TxsToL2Txs(txs []PoolL2Tx) ([]L2Tx, error) {
	l2Txs := make([]L2Tx, len(txs))
	for i, poolTx := range txs {
		l2Txs[i] = poolTx.L2Tx()
	}
	return l2Txs, nil
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
