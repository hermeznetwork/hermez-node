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
	// L1UserTxBytesLen is the length of the byte array that represents the L1Tx
	L1UserTxBytesLen = 72
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
	TxID            TxID               `meddler:"id"`
	ToForgeL1TxsNum *int64             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	Position        int                `meddler:"position"`
	UserOrigin      bool               `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromIdx         Idx                `meddler:"from_idx,zeroisnull"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	FromEthAddr     ethCommon.Address  `meddler:"from_eth_addr,zeroisnull"`
	FromBJJ         *babyjub.PublicKey `meddler:"from_bjj,zeroisnull"`
	ToIdx           Idx                `meddler:"to_idx"` // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	TokenID         TokenID            `meddler:"token_id"`
	Amount          *big.Int           `meddler:"amount,bigint"`
	// EffectiveAmount only applies to L1UserTx.
	EffectiveAmount *big.Int `meddler:"effective_amount,bigintnull"`
	LoadAmount      *big.Int `meddler:"load_amount,bigint"`
	// EffectiveLoadAmount only applies to L1UserTx.
	EffectiveLoadAmount *big.Int  `meddler:"effective_load_amount,bigintnull"`
	EthBlockNum         int64     `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	Type                TxType    `meddler:"type"`
	BatchNum            *BatchNum `meddler:"batch_num"`
}

// NewL1Tx returns the given L1Tx with the TxId & Type parameters calculated
// from the L1Tx values
func NewL1Tx(l1Tx *L1Tx) (*L1Tx, error) {
	// calculate TxType
	var txType TxType
	if l1Tx.FromIdx == 0 {
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
			txType = TxTypeForceExit
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

	txID, err := l1Tx.CalcTxID()
	if err != nil {
		return nil, err
	}
	l1Tx.TxID = *txID

	return l1Tx, nil
}

// CalcTxID calculates the TxId of the L1Tx
func (tx *L1Tx) CalcTxID() (*TxID, error) {
	var txID TxID
	if tx.UserOrigin {
		if tx.ToForgeL1TxsNum == nil {
			return nil, fmt.Errorf("L1Tx.UserOrigin == true && L1Tx.ToForgeL1TxsNum == nil")
		}
		txID[0] = TxIDPrefixL1UserTx
		var toForgeL1TxsNumBytes [8]byte
		binary.BigEndian.PutUint64(toForgeL1TxsNumBytes[:], uint64(*tx.ToForgeL1TxsNum))
		copy(txID[1:9], toForgeL1TxsNumBytes[:])
	} else {
		if tx.BatchNum == nil {
			return nil, fmt.Errorf("L1Tx.UserOrigin == false && L1Tx.BatchNum == nil")
		}
		txID[0] = TxIDPrefixL1CoordTx
		var batchNumBytes [8]byte
		binary.BigEndian.PutUint64(batchNumBytes[:], uint64(*tx.BatchNum))
		copy(txID[1:9], batchNumBytes[:])
	}
	var positionBytes [2]byte
	binary.BigEndian.PutUint16(positionBytes[:], uint16(tx.Position))
	copy(txID[9:11], positionBytes[:])

	return &txID, nil
}

// Tx returns a *Tx from the L1Tx
func (tx L1Tx) Tx() Tx {
	f := new(big.Float).SetInt(tx.EffectiveAmount)
	amountFloat, _ := f.Float64()
	userOrigin := new(bool)
	*userOrigin = tx.UserOrigin
	genericTx := Tx{
		IsL1:            true,
		TxID:            tx.TxID,
		Type:            tx.Type,
		Position:        tx.Position,
		FromIdx:         tx.FromIdx,
		ToIdx:           tx.ToIdx,
		Amount:          tx.EffectiveAmount,
		AmountFloat:     amountFloat,
		TokenID:         tx.TokenID,
		ToForgeL1TxsNum: tx.ToForgeL1TxsNum,
		UserOrigin:      userOrigin,
		FromEthAddr:     tx.FromEthAddr,
		FromBJJ:         tx.FromBJJ,
		LoadAmount:      tx.EffectiveLoadAmount,
		EthBlockNum:     tx.EthBlockNum,
	}
	if tx.LoadAmount != nil {
		lf := new(big.Float).SetInt(tx.LoadAmount)
		loadAmountFloat, _ := lf.Float64()
		genericTx.LoadAmountFloat = &loadAmountFloat
	}
	return genericTx
}

// TxCompressedData spec:
// [ 1 bits  ] empty (toBJJSign) // 1 byte
// [ 8 bits  ] empty (userFee) // 1 byte
// [ 40 bits ] empty (nonce) // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 16 bits ] amountFloat16 // 2 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// [ 16 bits ] chainId // 2 bytes
// [ 32 bits ] empty (signatureConstant) // 4 bytes
// Total bits compressed data:  241 bits // 31 bytes in *big.Int representation
func (tx L1Tx) TxCompressedData() (*big.Int, error) {
	amountFloat16, err := NewFloat16(tx.Amount)
	if err != nil {
		return nil, err
	}
	var b [31]byte
	// b[0:7] empty: no fee neither nonce
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
	copy(b[25:27], []byte{0, 1}) // TODO this will be generated by the ChainID config parameter
	// b[27:] empty: no signature

	bi := new(big.Int).SetBytes(b[:])
	return bi, nil
}

// BytesDataAvailability encodes a L1Tx into []byte for the Data Availability
func (tx *L1Tx) BytesDataAvailability(nLevels uint32) ([]byte, error) {
	idxLen := nLevels / 8 //nolint:gomnd

	b := make([]byte, ((nLevels*2)+16+8)/8) //nolint:gomnd

	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[0:idxLen], fromIdxBytes[6-idxLen:])
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, err
	}
	copy(b[idxLen:idxLen*2], toIdxBytes[6-idxLen:])

	if tx.EffectiveAmount != nil {
		amountFloat16, err := NewFloat16(tx.EffectiveAmount)
		if err != nil {
			return nil, err
		}
		copy(b[idxLen*2:idxLen*2+2], amountFloat16.Bytes())
	}
	// fee = 0 (as is L1Tx) b[10:11]
	return b[:], nil
}

// BytesGeneric returns the generic representation of a L1Tx. This method is
// used to compute the []byte representation of a L1UserTx, and also to compute
// the L1TxData for the ZKInputs (at the HashGlobalInputs), using this method
// for L1CoordinatorTxs & L1UserTxs (for the ZKInputs case).
func (tx *L1Tx) BytesGeneric() ([]byte, error) {
	var b [L1UserTxBytesLen]byte
	copy(b[0:20], tx.FromEthAddr.Bytes())
	if tx.FromBJJ != nil {
		pkCompL := tx.FromBJJ.Compress()
		pkCompB := SwapEndianness(pkCompL[:])
		copy(b[20:52], pkCompB[:])
	}
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

// BytesUser encodes a L1UserTx into []byte
func (tx *L1Tx) BytesUser() ([]byte, error) {
	if !tx.UserOrigin {
		return nil, fmt.Errorf("Can not calculate BytesUser() for a L1CoordinatorTx")
	}
	return tx.BytesGeneric()
}

// BytesCoordinatorTx encodes a L1CoordinatorTx into []byte
func (tx *L1Tx) BytesCoordinatorTx(compressedSignatureBytes []byte) ([]byte, error) {
	if tx.UserOrigin {
		return nil, fmt.Errorf("Can not calculate BytesCoordinatorTx() for a L1UserTx")
	}
	var b [L1CoordinatorTxBytesLen]byte
	v := compressedSignatureBytes[64]
	s := compressedSignatureBytes[32:64]
	r := compressedSignatureBytes[0:32]
	b[0] = v
	copy(b[1:33], s)
	copy(b[33:65], r)
	pkCompL := tx.FromBJJ.Compress()
	pkCompB := SwapEndianness(pkCompL[:])
	copy(b[65:97], pkCompB[:])
	copy(b[97:101], tx.TokenID.Bytes())
	return b[:], nil
}

// L1UserTxFromBytes decodes a L1Tx from []byte
func L1UserTxFromBytes(b []byte) (*L1Tx, error) {
	if len(b) != L1UserTxBytesLen {
		return nil, fmt.Errorf("Can not parse L1Tx bytes, expected length %d, current: %d", 68, len(b))
	}

	tx := &L1Tx{
		UserOrigin: true,
	}
	var err error
	tx.FromEthAddr = ethCommon.BytesToAddress(b[0:20])

	pkCompB := b[20:52]
	pkCompL := SwapEndianness(pkCompB)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkCompL)
	tx.FromBJJ, err = pkComp.Decompress()
	if err != nil {
		return nil, err
	}
	fromIdx, err := IdxFromBytes(b[52:58])
	if err != nil {
		return nil, err
	}
	tx.FromIdx = fromIdx
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

func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// L1CoordinatorTxFromBytes decodes a L1Tx from []byte
func L1CoordinatorTxFromBytes(b []byte, chainID *big.Int, hermezAddress ethCommon.Address) (*L1Tx, error) {
	if len(b) != L1CoordinatorTxBytesLen {
		return nil, fmt.Errorf("Can not parse L1CoordinatorTx bytes, expected length %d, current: %d", 101, len(b))
	}

	bytesMessage := []byte("I authorize this babyjubjub key for hermez rollup account creation")

	tx := &L1Tx{
		UserOrigin: false,
	}
	var err error
	v := b[0]
	s := b[1:33]
	r := b[33:65]
	pkCompB := b[65:97]
	pkCompL := SwapEndianness(pkCompB)
	var pkComp babyjub.PublicKeyComp
	copy(pkComp[:], pkCompL)
	tx.FromBJJ, err = pkComp.Decompress()
	if err != nil {
		return nil, err
	}
	tx.TokenID, err = TokenIDFromBytes(b[97:101])
	if err != nil {
		return nil, err
	}
	tx.Amount = big.NewInt(0)
	tx.LoadAmount = big.NewInt(0)
	if int(v) > 0 {
		// L1CoordinatorTX ETH
		// Ethereum adds 27 to v
		v = b[0] - byte(27) //nolint:gomnd
		chainIDBytes := ethCommon.LeftPadBytes(chainID.Bytes(), 2)
		var data []byte
		data = append(data, bytesMessage...)
		data = append(data, pkCompB...)
		data = append(data, chainIDBytes[:]...)
		data = append(data, hermezAddress.Bytes()...)
		var signature []byte
		signature = append(signature, r[:]...)
		signature = append(signature, s[:]...)
		signature = append(signature, v)
		hash := signHash(data)
		pubKeyBytes, err := crypto.Ecrecover(hash, signature)
		if err != nil {
			return nil, err
		}
		pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			return nil, err
		}
		tx.FromEthAddr = crypto.PubkeyToAddress(*pubKey)
	} else {
		// L1Coordinator Babyjub
		tx.FromEthAddr = RollupConstEthAddressInternalOnly
	}
	return tx, nil
}
