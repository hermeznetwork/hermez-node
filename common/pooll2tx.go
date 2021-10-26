package common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

// EmptyBJJComp contains the 32 byte array of a empty BabyJubJub PublicKey
// Compressed. It is a valid point in the BabyJubJub curve, so does not give
// errors when being decompressed.
var EmptyBJJComp = babyjub.PublicKeyComp([32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the
// coordinator that is waiting to be forged
type PoolL2Tx struct {
	// Stored in DB: mandatory fields

	// TxID (12 bytes) for L2Tx is:
	// bytes:  |  1   |    6    |   5   |
	// values: | type | FromIdx | Nonce |
	TxID    TxID `meddler:"tx_id"`
	FromIdx Idx  `meddler:"from_idx"`
	ToIdx   Idx  `meddler:"to_idx,zeroisnull"`
	// AuxToIdx is only used internally at the StateDB to avoid repeated
	// computation when processing transactions (from Synchronizer,
	// TxSelector, BatchBuilder)
	AuxToIdx    Idx                   `meddler:"-"`
	ToEthAddr   ethCommon.Address     `meddler:"to_eth_addr,zeroisnull"`
	ToBJJ       babyjub.PublicKeyComp `meddler:"to_bjj,zeroisnull"`
	TokenID     TokenID               `meddler:"token_id"`
	Amount      *big.Int              `meddler:"amount,bigint"`
	Fee         FeeSelector           `meddler:"fee"`
	Nonce       nonce.Nonce           `meddler:"nonce"` // effective 40 bits used
	State       PoolL2TxState         `meddler:"state"`
	MaxNumBatch uint32                `meddler:"max_num_batch,zeroisnull"`
	// Info contains information about the status & State of the
	// transaction. As for example, if the Tx has not been selected in the
	// last batch due not enough Balance at the Sender account, this reason
	// would appear at this parameter.
	Info      string                `meddler:"info,zeroisnull"`
	ErrorCode int                   `meddler:"error_code,zeroisnull"`
	ErrorType string                `meddler:"error_type,zeroisnull"`
	Signature babyjub.SignatureComp `meddler:"signature"`         // tx signature
	Timestamp time.Time             `meddler:"timestamp,utctime"` // time when added to the tx pool
	// Stored in DB: optional fields, may be uninitialized
	AtomicGroupID     AtomicGroupID         `meddler:"atomic_group_id,zeroisnull"`
	RqFromIdx         Idx                   `meddler:"rq_from_idx,zeroisnull"`
	RqToIdx           Idx                   `meddler:"rq_to_idx,zeroisnull"`
	RqToEthAddr       ethCommon.Address     `meddler:"rq_to_eth_addr,zeroisnull"`
	RqToBJJ           babyjub.PublicKeyComp `meddler:"rq_to_bjj,zeroisnull"`
	RqTokenID         TokenID               `meddler:"rq_token_id,zeroisnull"`
	RqAmount          *big.Int              `meddler:"rq_amount,bigintnull"`
	RqFee             FeeSelector           `meddler:"rq_fee,zeroisnull"`
	RqNonce           nonce.Nonce           `meddler:"rq_nonce,zeroisnull"` // effective 48 bits used
	AbsoluteFee       float64               `meddler:"fee_usd,zeroisnull"`
	AbsoluteFeeUpdate time.Time             `meddler:"usd_update,utctimez"`
	Type              TxType                `meddler:"tx_type"`
	RqOffset          uint8                 `meddler:"rq_offset,zeroisnull"` // (max 3 bits)
	// Extra DB write fields (not included in JSON)
	ClientIP string `meddler:"client_ip"`
	// Extra metadata, may be uninitialized
	RqTxCompressedData []byte `meddler:"-"` // 253 bits, optional for atomic txs
	TokenSymbol        string `meddler:"-"` // Used for JSON marshaling the ToIdx
}

// TxSelectorError struct that gives more details about the error
type TxSelectorError struct {
	Message string `meddler:"info,zeroisnull"`
	Code    int    `meddler:"error_code,zeroisnull"`
	Type    string `meddler:"error_type,zeroisnull"`
}

// NewPoolL2Tx returns the given L2Tx with the TxId & Type parameters calculated
// from the L2Tx values
func NewPoolL2Tx(tx *PoolL2Tx) (*PoolL2Tx, error) {
	txTypeOld := tx.Type
	if err := tx.SetType(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original Type doesn't match the correct one, return error
	if txTypeOld != "" && txTypeOld != tx.Type {
		return nil, tracerr.Wrap(fmt.Errorf("L2Tx.Type: %s, should be: %s",
			txTypeOld, tx.Type))
	}

	txIDOld := tx.TxID
	if err := tx.SetID(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original TxID doesn't match the correct one, return error
	if txIDOld != (TxID{}) && txIDOld != tx.TxID {
		return tx, tracerr.Wrap(fmt.Errorf("PoolL2Tx.TxID: %s, should be: %s",
			txIDOld.String(), tx.TxID.String()))
	}

	return tx, nil
}

// SetType sets the type of the transaction
func (tx *PoolL2Tx) SetType() error {
	isAddrEmpty := tx.ToBJJ == EmptyBJJComp && tx.ToEthAddr == EmptyAddr
	if tx.ToIdx >= IdxUserThreshold {
		tx.Type = TxTypeTransfer
	} else if isAddrEmpty && tx.ToIdx == 1 {
		tx.Type = TxTypeExit
	} else if tx.ToIdx == 0 {
		if tx.ToBJJ != EmptyBJJComp && tx.ToEthAddr == FFAddr {
			tx.Type = TxTypeTransferToBJJ
		} else if tx.ToEthAddr != FFAddr &&
			tx.ToEthAddr != EmptyAddr &&
			tx.ToBJJ == EmptyBJJComp {
			tx.Type = TxTypeTransferToEthAddr
		} else {
			return tracerr.Wrap(errors.New("malformed transaction"))
		}
	} else {
		return tracerr.Wrap(errors.New("malformed transaction"))
	}
	return nil
}

// SetID sets the ID of the transaction
func (tx *PoolL2Tx) SetID() error {
	txID, err := tx.L2Tx().CalculateTxID()
	if err != nil {
		return tracerr.Wrap(err)
	}
	tx.TxID = txID
	return nil
}

// TxCompressedData spec:
// [ 1 bits  ] toBJJSign // 1 byte
// [ 8 bits  ] userFee // 1 byte
// [ 40 bits ] nonce // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// [ 16 bits ] chainId // 2 bytes
// [ 32 bits ] signatureConstant // 4 bytes
// Total bits compressed data:  225 bits // 29 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedData(chainID uint16) (*big.Int, error) {
	var b [29]byte

	toBJJSign := byte(0)
	pkSign, _ := babyjub.UnpackSignY(tx.ToBJJ)
	if pkSign {
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
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[11:17], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[17:23], fromIdxBytes[:])
	binary.BigEndian.PutUint16(b[23:25], chainID)
	copy(b[25:29], SignatureConstantBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// TxCompressedDataEmpty calculates the TxCompressedData of an empty
// transaction
func TxCompressedDataEmpty(chainID uint16) *big.Int {
	var b [29]byte
	binary.BigEndian.PutUint16(b[23:25], chainID)
	copy(b[25:29], SignatureConstantBytes[:])
	bi := new(big.Int).SetBytes(b[:])
	return bi
}

// TxCompressedDataV2 spec:
// [ 1 bits  ] toBJJSign // 1 byte
// [ 8 bits  ] userFee // 1 byte
// [ 40 bits ] nonce // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 40 bits ] amountFloat40 // 5 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// Total bits compressed data:  217 bits // 28 bytes in *big.Int representation
func (tx *PoolL2Tx) TxCompressedDataV2() (*big.Int, error) {
	if tx.Amount == nil {
		tx.Amount = big.NewInt(0)
	}
	amountFloat40, err := NewFloat40(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var b [28]byte
	toBJJSign := byte(0)
	if tx.ToBJJ != EmptyBJJComp {
		sign, _ := babyjub.UnpackSignY(tx.ToBJJ)
		if sign {
			toBJJSign = byte(1)
		}
	}
	b[0] = toBJJSign
	b[1] = byte(tx.Fee)
	nonceBytes, err := tx.Nonce.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.TokenID.Bytes())
	copy(b[11:16], amountFloat40Bytes)
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[16:22], toIdxBytes[:])
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[22:28], fromIdxBytes[:])

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
// [ 40 bits ] rqAmountFloat40 // 5 bytes
// [ 48 bits ] rqToIdx // 6 bytes
// [ 48 bits ] rqFromIdx // 6 bytes
// Total bits compressed data:  217 bits // 28 bytes in *big.Int representation
func (tx *PoolL2Tx) RqTxCompressedDataV2() (*big.Int, error) {
	var amountFloat40 Float40
	if tx.RqAmount == nil {
		af40, err := NewFloat40(big.NewInt(0))
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		amountFloat40 = af40
	} else {
		af40, err := NewFloat40(tx.RqAmount)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		amountFloat40 = af40
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	var b [28]byte
	rqToBJJSign := byte(0)
	if tx.RqToBJJ != EmptyBJJComp {
		sign, _ := babyjub.UnpackSignY(tx.RqToBJJ)
		if sign {
			rqToBJJSign = byte(1)
		}
	}
	b[0] = rqToBJJSign
	b[1] = byte(tx.RqFee)
	nonceBytes, err := tx.RqNonce.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[2:7], nonceBytes[:])
	copy(b[7:11], tx.RqTokenID.Bytes())
	copy(b[11:16], amountFloat40Bytes)
	toIdxBytes, err := tx.RqToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[16:22], toIdxBytes[:])
	fromIdxBytes, err := tx.RqFromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[22:28], fromIdxBytes[:])

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// HashToSign returns the computed Poseidon hash from the *PoolL2Tx that will
// be signed by the sender.
func (tx *PoolL2Tx) HashToSign(chainID uint16) (*big.Int, error) {
	toCompressedData, err := tx.TxCompressedData(chainID)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	// e1: [4 bytes MaxBatchNum | 5 bytes AmountFloat40 | 20 bytes ToEthAddr]
	var e1B [29]byte
	maxNumBatchBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(maxNumBatchBytes, tx.MaxNumBatch)
	amountFloat40, err := NewFloat40(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(e1B[0:4], maxNumBatchBytes)
	copy(e1B[4:9], amountFloat40Bytes)
	copy(e1B[9:29], tx.ToEthAddr[:])
	e1 := new(big.Int).SetBytes(e1B[:])
	rqToEthAddr := EthAddrToBigInt(tx.RqToEthAddr)

	_, toBJJY := babyjub.UnpackSignY(tx.ToBJJ)

	rqTxCompressedDataV2, err := tx.RqTxCompressedDataV2()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	_, rqToBJJY := babyjub.UnpackSignY(tx.RqToBJJ)

	return poseidon.Hash([]*big.Int{toCompressedData, e1, toBJJY, rqTxCompressedDataV2,
		rqToEthAddr, rqToBJJY})
}

// VerifySignature returns true if the signature verification is correct for the given PublicKeyComp
func (tx *PoolL2Tx) VerifySignature(chainID uint16, pkComp babyjub.PublicKeyComp) bool {
	h, err := tx.HashToSign(chainID)
	if err != nil {
		return false
	}
	s, err := tx.Signature.Decompress()
	if err != nil {
		return false
	}
	pk, err := pkComp.Decompress()
	if err != nil {
		return false
	}
	return pk.VerifyPoseidon(h, s)
}

// L2Tx returns a *L2Tx from the PoolL2Tx
func (tx PoolL2Tx) L2Tx() L2Tx {
	var toIdx Idx
	if tx.ToIdx == Idx(0) {
		toIdx = tx.AuxToIdx
	} else {
		toIdx = tx.ToIdx
	}
	return L2Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   toIdx,
		TokenID: tx.TokenID,
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

// TxIDsFromPoolL2Txs returns an array of TxID from the []PoolL2Tx
func TxIDsFromPoolL2Txs(txs []PoolL2Tx) []TxID {
	txIDs := make([]TxID, len(txs))
	for i, tx := range txs {
		txIDs[i] = tx.TxID
	}
	return txIDs
}

// PoolL2TxState is a string that represents the status of a L2 transaction
type PoolL2TxState string

const (
	// PoolL2TxStatePending represents a valid L2Tx that hasn't started the
	// forging process
	PoolL2TxStatePending PoolL2TxState = "pend"
	// PoolL2TxStateForging represents a valid L2Tx that has started the
	// forging process
	PoolL2TxStateForging PoolL2TxState = "fing"
	// PoolL2TxStateForged represents a L2Tx that has already been forged
	PoolL2TxStateForged PoolL2TxState = "fged"
	// PoolL2TxStateInvalid represents a L2Tx that has been invalidated
	PoolL2TxStateInvalid PoolL2TxState = "invl"
)

// IsValid returns true if matches one of the supported PoolL2TxState
func (s PoolL2TxState) IsValid() bool {
	switch s {
	case PoolL2TxStatePending, PoolL2TxStateForging, PoolL2TxStateForged, PoolL2TxStateInvalid:
		return true
	default:
		return false
	}
}

// MarshalJSON formats a PoolL2Tx in the expected JSON defined by the API,
// this format is specified in `api/swagger.yml:schemas/PostPoolL2Transaction`
func (tx PoolL2Tx) MarshalJSON() ([]byte, error) {
	if tx.TokenSymbol == "" {
		return nil, errors.New("Invalid tx.TokenSymbol")
	}
	type jsonFormat struct {
		TxID        TxID                  `json:"id"`
		Type        TxType                `json:"type"`
		TokenID     TokenID               `json:"tokenId"`
		FromIdx     string                `json:"fromAccountIndex"`
		ToIdx       *string               `json:"toAccountIndex"`
		ToEthAddr   *string               `json:"toHezEthereumAddress"`
		ToBJJ       *string               `json:"toBjj"`
		Amount      string                `json:"amount"`
		Fee         FeeSelector           `json:"fee"`
		Nonce       nonce.Nonce           `json:"nonce"`
		MaxNumBatch uint32                `json:"maxNumBatch"`
		Signature   babyjub.SignatureComp `json:"signature"`
		RqOffset    *uint8                `json:"requestOffset"`
	}
	// Set fields that do not require extra logic
	toMarshal := jsonFormat{
		TxID:        tx.TxID,
		Type:        tx.Type,
		TokenID:     tx.TokenID,
		FromIdx:     IdxToHez(tx.FromIdx, tx.TokenSymbol),
		Amount:      tx.Amount.String(),
		Fee:         tx.Fee,
		Nonce:       tx.Nonce,
		MaxNumBatch: tx.MaxNumBatch,
		Signature:   tx.Signature,
	}
	// Set To fields
	if tx.ToIdx != 0 {
		toIdx := IdxToHez(tx.ToIdx, tx.TokenSymbol)
		toMarshal.ToIdx = &toIdx
	}
	if tx.ToEthAddr != EmptyAddr {
		toEth := EthAddrToHez(tx.ToEthAddr)
		toMarshal.ToEthAddr = &toEth
	}
	if tx.ToBJJ != EmptyBJJComp {
		toBJJ := BjjToString(tx.ToBJJ)
		toMarshal.ToBJJ = &toBJJ
	}
	// Check if is a atomic and require another tx
	if tx.RqFromIdx != 0 {
		toMarshal.RqOffset = &tx.RqOffset
	}
	return json.Marshal(toMarshal)
}

// UnmarshalJSON unmarshal a PoolL2Tx that has been formatted in json as specified
// in `api/swagger.yml:schemas/PostPoolL2Transaction`.
// Note that ClientIP is not included as part of the json and will be set to "".
// If State is not defined (State == ""), it will be set to PoolL2TxStatePending.
func (tx *PoolL2Tx) UnmarshalJSON(data []byte) error {
	receivedJSON := struct {
		TxID        TxID                  `json:"id" binding:"required"`
		Type        TxType                `json:"type" binding:"required"`
		TokenID     TokenID               `json:"tokenId"`
		FromIdx     StrHezIdx             `json:"fromAccountIndex" binding:"required"`
		ToIdx       StrHezIdx             `json:"toAccountIndex"`
		ToEthAddr   StrHezEthAddr         `json:"toHezEthereumAddress"`
		ToBJJ       StrHezBJJ             `json:"toBjj"`
		Amount      *StrBigInt            `json:"amount" binding:"required"`
		Fee         FeeSelector           `json:"fee"`
		Nonce       nonce.Nonce           `json:"nonce"`
		MaxNumBatch uint32                `json:"maxNumBatch"`
		Signature   babyjub.SignatureComp `json:"signature" binding:"required"`
		State       string                `json:"state"`
		RqOffset    uint8                 `json:"requestOffset"`
		Info        *string               `json:"info"`
	}{}
	if err := json.Unmarshal(data, &receivedJSON); err != nil {
		return err
	}
	// Check state
	state := PoolL2TxStatePending
	if receivedJSON.State != "" {
		state = PoolL2TxState(receivedJSON.State)
		if !state.IsValid() {
			return fmt.Errorf("Invalid state: %s", state)
		}
	}
	var info string
	if receivedJSON.Info != nil {
		info = *receivedJSON.Info
	}
	// Set values to destination struct
	*tx = PoolL2Tx{
		TxID:        receivedJSON.TxID,
		FromIdx:     Idx(receivedJSON.FromIdx.Idx),
		ToIdx:       Idx(receivedJSON.ToIdx.Idx),
		ToEthAddr:   ethCommon.Address(receivedJSON.ToEthAddr),
		ToBJJ:       babyjub.PublicKeyComp(receivedJSON.ToBJJ),
		TokenID:     receivedJSON.TokenID,
		Amount:      (*big.Int)(receivedJSON.Amount),
		Fee:         receivedJSON.Fee,
		Nonce:       receivedJSON.Nonce,
		MaxNumBatch: receivedJSON.MaxNumBatch,
		Signature:   receivedJSON.Signature,
		Type:        receivedJSON.Type,
		State:       state,
		TokenSymbol: receivedJSON.FromIdx.TokenSymbol,
		RqOffset:    receivedJSON.RqOffset,
		Info:        info,
	}
	return nil
}
