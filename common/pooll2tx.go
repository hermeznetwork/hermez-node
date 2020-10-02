package common

import (
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the coordinator hat is waiting to be forged
type PoolL2Tx struct {
	// Stored in DB: mandatory fileds

	// TxID (12 bytes) for L2Tx is:
	// bytes:  |  1   |    6    |   5   |
	// values: | type | FromIdx | Nonce |
	TxID        TxID               `meddler:"tx_id"`
	FromIdx     Idx                `meddler:"from_idx"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx       Idx                `meddler:"to_idx"`   // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	AuxToIdx    Idx                `meddler:"-"`        // AuxToIdx is only used internally at the StateDB to avoid repeated computation when processing transactions
	ToEthAddr   ethCommon.Address  `meddler:"to_eth_addr"`
	ToBJJ       *babyjub.PublicKey `meddler:"to_bjj"` // TODO: stop using json, use scanner/valuer
	TokenID     TokenID            `meddler:"token_id"`
	Amount      *big.Int           `meddler:"amount,bigint"` // TODO: change to float16
	AmountFloat float64            `meddler:"amount_f"`      // TODO: change to float16
	USD         *float64           `meddler:"value_usd"`     // TODO: change to float16
	Fee         FeeSelector        `meddler:"fee"`
	Nonce       Nonce              `meddler:"nonce"` // effective 40 bits used
	State       PoolL2TxState      `meddler:"state"`
	Signature   *babyjub.Signature `meddler:"signature"`         // tx signature
	Timestamp   time.Time          `meddler:"timestamp,utctime"` // time when added to the tx pool
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
	AbsoluteFee       *float64           `meddler:"fee_usd"`
	AbsoluteFeeUpdate *time.Time         `meddler:"usd_update,utctime"`
	Type              TxType             `meddler:"tx_type"`
	// Extra metadata, may be uninitialized
	RqTxCompressedData []byte `meddler:"-"` // 253 bits, optional for atomic txs
	TokenSymbol        string `meddler:"token_symbol"`
}

// NewPoolL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewPoolL2Tx(poolL2Tx *PoolL2Tx) (*PoolL2Tx, error) {
	// calculate TxType
	var txType TxType
	if poolL2Tx.ToIdx == Idx(0) {
		txType = TxTypeTransfer
	} else if poolL2Tx.ToIdx == Idx(1) {
		txType = TxTypeExit
	} else if poolL2Tx.ToIdx >= IdxUserThreshold {
		txType = TxTypeTransfer
	} else {
		return poolL2Tx, fmt.Errorf("Can not determine type of PoolL2Tx, invalid ToIdx value: %d", poolL2Tx.ToIdx)
	}

	// if TxType!=poolL2Tx.TxType return error
	if poolL2Tx.Type != "" && poolL2Tx.Type != txType {
		return poolL2Tx, fmt.Errorf("PoolL2Tx.Type: %s, should be: %s", poolL2Tx.Type, txType)
	}
	poolL2Tx.Type = txType

	var txid [TxIDLen]byte
	txid[0] = TxIDPrefixL2Tx
	fromIdxBytes, err := poolL2Tx.FromIdx.Bytes()
	if err != nil {
		return poolL2Tx, err
	}
	copy(txid[1:7], fromIdxBytes[:])
	nonceBytes, err := poolL2Tx.Nonce.Bytes()
	if err != nil {
		return poolL2Tx, err
	}
	copy(txid[7:12], nonceBytes[:])
	poolL2Tx.TxID = TxID(txid)

	return poolL2Tx, nil
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
	// sigconstant
	sc, ok := new(big.Int).SetString("3322668559", 10)
	if !ok {
		return nil, fmt.Errorf("error parsing SignatureConstant")
	}

	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	var b [31]byte
	toBJJSign := byte(0)
	if babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBJJSign = byte(1)
	}
	b[0] = toBJJSign
	b[1] = byte(tx.Fee)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.TokenID.Bytes())
	copy(b[11:13], amountFloat16.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[13:19], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[19:25], fromIdxBytes[:])
	copy(b[25:27], []byte{0, 1, 0, 0}) // TODO check js implementation (unexpected behaviour from test vector generated from js)
	copy(b[27:31], sc.Bytes())

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
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	var b [25]byte
	toBJJSign := byte(0)
	if babyjub.PointCoordSign(tx.ToBJJ.X) {
		toBJJSign = byte(1)
	}
	b[0] = toBJJSign
	b[1] = byte(tx.Fee)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.TokenID.Bytes())
	copy(b[11:13], amountFloat16.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[13:19], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[19:25], fromIdxBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// HashToSign returns the computed Poseidon hash from the *PoolL2Tx that will be signed by the sender.
func (tx *PoolL2Tx) HashToSign() (*big.Int, error) {
	toCompressedData, err := tx.TxCompressedData()
	if err != nil {
		return nil, err
	}
	toEthAddr := EthAddrToBigInt(tx.ToEthAddr)
	toBJJAy := tx.ToBJJ.Y
	rqTxCompressedDataV2, err := tx.TxCompressedDataV2()
	if err != nil {
		return nil, err
	}

	return poseidon.Hash([]*big.Int{toCompressedData, toEthAddr, toBJJAy, rqTxCompressedDataV2, EthAddrToBigInt(tx.RqToEthAddr), tx.RqToBJJ.Y})
}

// VerifySignature returns true if the signature verification is correct for the given PublicKey
func (tx *PoolL2Tx) VerifySignature(pk *babyjub.PublicKey) bool {
	h, err := tx.HashToSign()
	if err != nil {
		return false
	}
	return pk.VerifyPoseidon(h, tx.Signature)
}

// L2Tx returns a *L2Tx from the PoolL2Tx
func (tx *PoolL2Tx) L2Tx() L2Tx {
	return L2Tx{
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

// Tx returns a *Tx from the PoolL2Tx
func (tx *PoolL2Tx) Tx() *Tx {
	return &Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		Amount:  tx.Amount,
		Nonce:   &tx.Nonce,
		Fee:     &tx.Fee,
		Type:    tx.Type,
	}
}

// PoolL2TxsToL2Txs returns an array of []*L2Tx from an array of []*PoolL2Tx
func PoolL2TxsToL2Txs(txs []PoolL2Tx) []L2Tx {
	var r []L2Tx
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
