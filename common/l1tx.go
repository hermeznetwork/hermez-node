package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/hermeznetwork/tracerr"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	// Stored in DB: mandatory fileds

	// TxID (32 bytes) for L1Tx is the Keccak256 (ethereum) hash of:
	// bytes:  |  1   |        8        |    2     |      1      |
	// values: | type | ToForgeL1TxsNum | Position | 0 (padding) |
	// where type:
	// 	- L1UserTx: 0
	// 	- L1CoordinatorTx: 1
	TxID TxID `meddler:"id"`
	// ToForgeL1TxsNum indicates in which L1UserTx queue the tx was forged / will be forged
	ToForgeL1TxsNum *int64 `meddler:"to_forge_l1_txs_num"`
	Position        int    `meddler:"position"`
	// UserOrigin is set to true if the tx was originated by a user, false if it was
	// aoriginated by a coordinator. Note that this differ from the spec for implementation
	// simplification purpposes
	UserOrigin bool `meddler:"user_origin"`
	// FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.DepositAmount
	// (deposit)
	FromIdx          Idx                   `meddler:"from_idx,zeroisnull"`
	EffectiveFromIdx Idx                   `meddler:"effective_from_idx,zeroisnull"`
	FromEthAddr      ethCommon.Address     `meddler:"from_eth_addr,zeroisnull"`
	FromBJJ          babyjub.PublicKeyComp `meddler:"from_bjj,zeroisnull"`
	// ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	ToIdx   Idx      `meddler:"to_idx"`
	TokenID TokenID  `meddler:"token_id"`
	Amount  *big.Int `meddler:"amount,bigint"`
	// EffectiveAmount only applies to L1UserTx.
	EffectiveAmount *big.Int `meddler:"effective_amount,bigintnull"`
	DepositAmount   *big.Int `meddler:"deposit_amount,bigint"`
	// EffectiveDepositAmount only applies to L1UserTx.
	EffectiveDepositAmount *big.Int `meddler:"effective_deposit_amount,bigintnull"`
	// Ethereum Block Number in which this L1Tx was added to the queue
	EthBlockNum int64          `meddler:"eth_block_num"`
	EthTxHash   ethCommon.Hash `meddler:"eth_tx_hash,zeroisnull"`
	L1Fee       *big.Int       `meddler:"l1_fee,bigintnull"`
	Type        TxType         `meddler:"type"`
	BatchNum    *BatchNum      `meddler:"batch_num"`
}

// NewL1Tx returns the given L1Tx with the TxId & Type parameters calculated
// from the L1Tx values
func NewL1Tx(tx *L1Tx) (*L1Tx, error) {
	txTypeOld := tx.Type
	if err := tx.SetType(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original Type doesn't match the correct one, return error
	if txTypeOld != "" && txTypeOld != tx.Type {
		return nil, tracerr.Wrap(fmt.Errorf("L1Tx.Type: %s, should be: %s",
			tx.Type, txTypeOld))
	}

	txIDOld := tx.TxID
	if err := tx.SetID(); err != nil {
		return nil, tracerr.Wrap(err)
	}
	// If original TxID doesn't match the correct one, return error
	if txIDOld != (TxID{}) && txIDOld != tx.TxID {
		return tx, tracerr.Wrap(fmt.Errorf("L1Tx.TxID: %s, should be: %s",
			tx.TxID.String(), txIDOld.String()))
	}

	return tx, nil
}

// SetType sets the type of the transaction
func (tx *L1Tx) SetType() error {
	if tx.FromIdx == 0 {
		if tx.ToIdx == Idx(0) {
			tx.Type = TxTypeCreateAccountDeposit
		} else if tx.ToIdx >= IdxUserThreshold {
			tx.Type = TxTypeCreateAccountDepositTransfer
		} else {
			return tracerr.Wrap(fmt.Errorf(
				"Can not determine type of L1Tx, invalid ToIdx value: %d", tx.ToIdx))
		}
	} else if tx.FromIdx >= IdxUserThreshold {
		if tx.ToIdx == Idx(0) {
			tx.Type = TxTypeDeposit
		} else if tx.ToIdx == Idx(1) {
			tx.Type = TxTypeForceExit
		} else if tx.ToIdx >= IdxUserThreshold {
			if tx.DepositAmount.Int64() == int64(0) {
				tx.Type = TxTypeForceTransfer
			} else {
				tx.Type = TxTypeDepositTransfer
			}
		} else {
			return tracerr.Wrap(fmt.Errorf(
				"Can not determine type of L1Tx, invalid ToIdx value: %d", tx.ToIdx))
		}
	} else {
		return tracerr.Wrap(fmt.Errorf(
			"Can not determine type of L1Tx, invalid FromIdx value: %d", tx.FromIdx))
	}
	return nil
}

// SetID sets the ID of the transaction.  For L1UserTx uses (ToForgeL1TxsNum,
// Position), for L1CoordinatorTx uses (BatchNum, Position).
func (tx *L1Tx) SetID() error {
	var b []byte
	if tx.UserOrigin {
		if tx.ToForgeL1TxsNum == nil {
			return tracerr.Wrap(fmt.Errorf("L1Tx.UserOrigin == true && L1Tx.ToForgeL1TxsNum == nil"))
		}
		tx.TxID[0] = TxIDPrefixL1UserTx

		var toForgeL1TxsNumBytes [8]byte
		binary.BigEndian.PutUint64(toForgeL1TxsNumBytes[:], uint64(*tx.ToForgeL1TxsNum))
		b = append(b, toForgeL1TxsNumBytes[:]...)
	} else {
		if tx.BatchNum == nil {
			return tracerr.Wrap(fmt.Errorf("L1Tx.UserOrigin == false && L1Tx.BatchNum == nil"))
		}
		tx.TxID[0] = TxIDPrefixL1CoordTx

		var batchNumBytes [8]byte
		binary.BigEndian.PutUint64(batchNumBytes[:], uint64(*tx.BatchNum))
		b = append(b, batchNumBytes[:]...)
	}
	var positionBytes [2]byte
	binary.BigEndian.PutUint16(positionBytes[:], uint16(tx.Position))
	b = append(b, positionBytes[:]...)

	// calculate hash
	h := ethCrypto.Keccak256Hash(b).Bytes()

	copy(tx.TxID[1:], h)

	return nil
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
		DepositAmount:   tx.EffectiveDepositAmount,
		EthBlockNum:     tx.EthBlockNum,
	}
	if tx.DepositAmount != nil {
		lf := new(big.Float).SetInt(tx.DepositAmount)
		depositAmountFloat, _ := lf.Float64()
		genericTx.DepositAmountFloat = &depositAmountFloat
	}
	return genericTx
}

// TxCompressedData spec:
// [ 1 bits  ] empty (toBJJSign) // 1 byte
// [ 8 bits  ] empty (userFee) // 1 byte
// [ 40 bits ] empty (nonce) // 5 bytes
// [ 32 bits ] tokenID // 4 bytes
// [ 48 bits ] toIdx // 6 bytes
// [ 48 bits ] fromIdx // 6 bytes
// [ 16 bits ] chainId // 2 bytes
// [ 32 bits ] empty (signatureConstant) // 4 bytes
// Total bits compressed data:  225 bits // 29 bytes in *big.Int representation
func (tx L1Tx) TxCompressedData(chainID uint16) (*big.Int, error) {
	var b [29]byte
	// b[0:7] empty: no ToBJJSign, no fee, no nonce
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

// BytesDataAvailability encodes a L1Tx into []byte for the Data Availability
// [ fromIdx | toIdx | amountFloat40 | Fee ]
func (tx *L1Tx) BytesDataAvailability(nLevels uint32) ([]byte, error) {
	idxLen := nLevels / 8 //nolint:gomnd

	b := make([]byte, ((nLevels*2)+40+8)/8) //nolint:gomnd

	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[0:idxLen], fromIdxBytes[6-idxLen:])
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[idxLen:idxLen*2], toIdxBytes[6-idxLen:])

	if tx.EffectiveAmount != nil {
		amountFloat40, err := NewFloat40(tx.EffectiveAmount)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		amountFloat40Bytes, err := amountFloat40.Bytes()
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		copy(b[idxLen*2:idxLen*2+Float40BytesLength], amountFloat40Bytes)
	}
	// fee = 0 (as is L1Tx)
	return b[:], nil
}

// L1TxFromDataAvailability decodes a L1Tx from []byte (Data Availability)
func L1TxFromDataAvailability(b []byte, nLevels uint32) (*L1Tx, error) {
	idxLen := nLevels / 8 //nolint:gomnd

	fromIdxBytes := b[0:idxLen]
	toIdxBytes := b[idxLen : idxLen*2]
	amountBytes := b[idxLen*2 : idxLen*2+Float40BytesLength]

	l1tx := L1Tx{}
	fromIdx, err := IdxFromBytes(ethCommon.LeftPadBytes(fromIdxBytes, 6))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	l1tx.FromIdx = fromIdx
	toIdx, err := IdxFromBytes(ethCommon.LeftPadBytes(toIdxBytes, 6))
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	l1tx.ToIdx = toIdx
	l1tx.EffectiveAmount, err = Float40FromBytes(amountBytes).BigInt()
	return &l1tx, tracerr.Wrap(err)
}

// BytesGeneric returns the generic representation of a L1Tx. This method is
// used to compute the []byte representation of a L1UserTx, and also to compute
// the L1TxData for the ZKInputs (at the HashGlobalInputs), using this method
// for L1CoordinatorTxs & L1UserTxs (for the ZKInputs case).
func (tx *L1Tx) BytesGeneric() ([]byte, error) {
	var b [RollupConstL1UserTotalBytes]byte
	copy(b[0:20], tx.FromEthAddr.Bytes())
	if tx.FromBJJ != EmptyBJJComp {
		pkCompL := tx.FromBJJ
		pkCompB := SwapEndianness(pkCompL[:])
		copy(b[20:52], pkCompB[:])
	}
	fromIdxBytes, err := tx.FromIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[52:58], fromIdxBytes[:])

	depositAmountFloat40, err := NewFloat40(tx.DepositAmount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	depositAmountFloat40Bytes, err := depositAmountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[58:63], depositAmountFloat40Bytes)

	amountFloat40, err := NewFloat40(tx.Amount)
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	amountFloat40Bytes, err := amountFloat40.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[63:68], amountFloat40Bytes)

	copy(b[68:72], tx.TokenID.Bytes())
	toIdxBytes, err := tx.ToIdx.Bytes()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	copy(b[72:78], toIdxBytes[:])
	return b[:], nil
}

// BytesUser encodes a L1UserTx into []byte
func (tx *L1Tx) BytesUser() ([]byte, error) {
	if !tx.UserOrigin {
		return nil, tracerr.Wrap(fmt.Errorf("Can not calculate BytesUser() for a L1CoordinatorTx"))
	}
	return tx.BytesGeneric()
}

// BytesCoordinatorTx encodes a L1CoordinatorTx into []byte
func (tx *L1Tx) BytesCoordinatorTx(compressedSignatureBytes []byte) ([]byte, error) {
	if tx.UserOrigin {
		return nil, tracerr.Wrap(fmt.Errorf("Can not calculate BytesCoordinatorTx() for a L1UserTx"))
	}
	var b [RollupConstL1CoordinatorTotalBytes]byte
	v := compressedSignatureBytes[64]
	s := compressedSignatureBytes[32:64]
	r := compressedSignatureBytes[0:32]
	b[0] = v
	copy(b[1:33], s)
	copy(b[33:65], r)
	pkCompL := tx.FromBJJ
	pkCompB := SwapEndianness(pkCompL[:])
	copy(b[65:97], pkCompB[:])
	copy(b[97:101], tx.TokenID.Bytes())
	return b[:], nil
}

// L1UserTxFromBytes decodes a L1Tx from []byte
func L1UserTxFromBytes(b []byte) (*L1Tx, error) {
	if len(b) != RollupConstL1UserTotalBytes {
		return nil,
			tracerr.Wrap(fmt.Errorf("Can not parse L1Tx bytes, expected length %d, current: %d",
				68, len(b)))
	}

	tx := &L1Tx{
		UserOrigin: true,
	}
	var err error
	tx.FromEthAddr = ethCommon.BytesToAddress(b[0:20])

	pkCompB := b[20:52]
	pkCompL := SwapEndianness(pkCompB)
	copy(tx.FromBJJ[:], pkCompL)
	fromIdx, err := IdxFromBytes(b[52:58])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.FromIdx = fromIdx
	tx.DepositAmount, err = Float40FromBytes(b[58:63]).BigInt()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.Amount, err = Float40FromBytes(b[63:68]).BigInt()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.TokenID, err = TokenIDFromBytes(b[68:72])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.ToIdx, err = IdxFromBytes(b[72:78])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	return tx, nil
}

// L1CoordinatorTxFromBytes decodes a L1Tx from []byte
func L1CoordinatorTxFromBytes(b []byte, chainID *big.Int, hermezAddress ethCommon.Address) (*L1Tx,
	error) {
	if len(b) != RollupConstL1CoordinatorTotalBytes {
		return nil, tracerr.Wrap(
			fmt.Errorf("Can not parse L1CoordinatorTx bytes, expected length %d, current: %d",
				101, len(b)))
	}

	tx := &L1Tx{
		UserOrigin: false,
	}
	var err error
	v := b[0]
	s := b[1:33]
	r := b[33:65]
	pkCompB := b[65:97]
	pkCompL := SwapEndianness(pkCompB)
	copy(tx.FromBJJ[:], pkCompL)
	tx.TokenID, err = TokenIDFromBytes(b[97:101])
	if err != nil {
		return nil, tracerr.Wrap(err)
	}
	tx.Amount = big.NewInt(0)
	tx.DepositAmount = big.NewInt(0)
	if int(v) > 0 {
		// L1CoordinatorTX ETH
		// Ethereum adds 27 to v
		v = b[0] - byte(27) //nolint:gomnd
		var signature []byte
		signature = append(signature, r[:]...)
		signature = append(signature, s[:]...)
		signature = append(signature, v)

		accCreationAuth := AccountCreationAuth{
			BJJ: tx.FromBJJ,
		}
		h, err := accCreationAuth.HashToSign(uint16(chainID.Uint64()), hermezAddress)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		pubKeyBytes, err := ethCrypto.Ecrecover(h, signature)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		pubKey, err := ethCrypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}
		tx.FromEthAddr = ethCrypto.PubkeyToAddress(*pubKey)
	} else {
		// L1Coordinator Babyjub
		tx.FromEthAddr = RollupConstEthAddressInternalOnly
	}
	return tx, nil
}
