package common

import (
	"database/sql/driver"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	// TXIDPrefixL1UserTx is the prefix that determines that the TxID is
	// for a L1UserTx
	//nolinter:gomnd
	TxIDPrefixL1UserTx = byte(0)

	// TXIDPrefixL1CoordTx is the prefix that determines that the TxID is
	// for a L1CoordinatorTx
	//nolinter:gomnd
	TxIDPrefixL1CoordTx = byte(1)

	// TxIDPrefixL2Tx is the prefix that determines that the TxID is for a
	// L2Tx (or PoolL2Tx)
	//nolinter:gomnd
	TxIDPrefixL2Tx = byte(2)

	// TxIDLen is the length of the TxID byte array
	TxIDLen = 12
)

// TxID is the identifier of a Hermez network transaction
type TxID [TxIDLen]byte

// Scan implements Scanner for database/sql.
func (txid *TxID) Scan(src interface{}) error {
	srcB, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("can't scan %T into TxID", src)
	}
	if len(srcB) != TxIDLen {
		return fmt.Errorf("can't scan []byte of len %d into TxID, need %d", len(srcB), TxIDLen)
	}
	copy(txid[:], srcB)
	return nil
}

// Value implements valuer for database/sql.
func (txid TxID) Value() (driver.Value, error) {
	return txid[:], nil
}

// String returns a string hexadecimal representation of the TxID
func (txid TxID) String() string {
	return "0x" + hex.EncodeToString(txid[:])
}

// NewTxIDFromString returns a string hexadecimal representation of the TxID
func NewTxIDFromString(idStr string) (TxID, error) {
	txid := TxID{}
	idStr = strings.TrimPrefix(idStr, "0x")
	decoded, err := hex.DecodeString(idStr)
	if err != nil {
		return TxID{}, err
	}
	if len(decoded) != TxIDLen {
		return txid, errors.New("Invalid idStr")
	}
	copy(txid[:], decoded)
	return txid, nil
}

// TxType is a string that represents the type of a Hermez network transaction
type TxType string

const (
	// TxTypeExit represents L2->L1 token transfer.  A leaf for this account appears in the exit tree of the block
	TxTypeExit TxType = "Exit"
	// TxTypeTransfer represents L2->L2 token transfer
	TxTypeTransfer TxType = "Transfer"
	// TxTypeDeposit represents L1->L2 transfer
	TxTypeDeposit TxType = "Deposit"
	// TxTypeCreateAccountDeposit represents creation of a new leaf in the state tree (newAcconut) + L1->L2 transfer
	TxTypeCreateAccountDeposit TxType = "CreateAccountDeposit"
	// TxTypeCreateAccountDepositTransfer represents L1->L2 transfer + L2->L2 transfer
	TxTypeCreateAccountDepositTransfer TxType = "CreateAccountDepositTransfer"
	// TxTypeDepositTransfer TBD
	TxTypeDepositTransfer TxType = "DepositTransfer"
	// TxTypeForceTransfer TBD
	TxTypeForceTransfer TxType = "ForceTransfer"
	// TxTypeForceExit TBD
	TxTypeForceExit TxType = "ForceExit"
	// TxTypeTransferToEthAddr TBD
	TxTypeTransferToEthAddr TxType = "TransferToEthAddr"
	// TxTypeTransferToBJJ TBD
	TxTypeTransferToBJJ TxType = "TransferToBJJ"
)

// Tx is a struct used by the TxSelector & BatchBuilder as a generic type generated from L1Tx & PoolL2Tx
// TODO: this should be changed for "mini Tx"
type Tx struct {
	// Generic
	IsL1        bool      `meddler:"is_l1"`
	TxID        TxID      `meddler:"id"`
	Type        TxType    `meddler:"type"`
	Position    int       `meddler:"position"`
	FromIdx     Idx       `meddler:"from_idx"`
	ToIdx       Idx       `meddler:"to_idx"`
	Amount      *big.Int  `meddler:"amount,bigint"`
	AmountFloat float64   `meddler:"amount_f"`
	TokenID     TokenID   `meddler:"token_id"`
	USD         *float64  `meddler:"amount_usd"`
	BatchNum    *BatchNum `meddler:"batch_num"`     // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum int64     `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum *int64             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin      *bool              `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr     ethCommon.Address  `meddler:"from_eth_addr"`
	FromBJJ         *babyjub.PublicKey `meddler:"from_bjj"`
	LoadAmount      *big.Int           `meddler:"load_amount,bigintnull"`
	LoadAmountFloat *float64           `meddler:"load_amount_f"`
	LoadAmountUSD   *float64           `meddler:"load_amount_usd"`
	// L2
	Fee    *FeeSelector `meddler:"fee"`
	FeeUSD *float64     `meddler:"fee_usd"`
	Nonce  *Nonce       `meddler:"nonce"`
}

// L1Tx returns a *L1Tx from the Tx
func (tx *Tx) L1Tx() (*L1Tx, error) {
	return &L1Tx{
		TxID:            tx.TxID,
		ToForgeL1TxsNum: tx.ToForgeL1TxsNum,
		Position:        tx.Position,
		UserOrigin:      *tx.UserOrigin,
		FromIdx:         tx.FromIdx,
		FromEthAddr:     tx.FromEthAddr,
		FromBJJ:         tx.FromBJJ,
		ToIdx:           tx.ToIdx,
		TokenID:         tx.TokenID,
		Amount:          tx.Amount,
		LoadAmount:      tx.LoadAmount,
		EthBlockNum:     tx.EthBlockNum,
		Type:            tx.Type,
		BatchNum:        tx.BatchNum,
	}, nil
}
