package common

import (
	"encoding/binary"
	"fmt"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const (
	idxBytesLen = 4
	// maxIdxValue is the maximum value that Idx can have (32 bits: maxIdxValue=2**32-1)
	maxIdxValue = 0xffffffff
)

// Idx represents the account Index in the MerkleTree
type Idx uint32

// Bytes returns a byte array representing the Idx
func (idx Idx) Bytes() []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(idx))
	return b[:]
}

// BigInt returns a *big.Int representing the Idx
func (idx Idx) BigInt() *big.Int {
	return big.NewInt(int64(idx))
}

// IdxFromBytes returns Idx from a byte array
func IdxFromBytes(b []byte) (Idx, error) {
	if len(b) != idxBytesLen {
		return 0, fmt.Errorf("can not parse Idx, bytes len %d, expected 4", len(b))
	}
	idx := binary.LittleEndian.Uint32(b[:4])
	return Idx(idx), nil
}

// IdxFromBigInt converts a *big.Int to Idx type
func IdxFromBigInt(b *big.Int) (Idx, error) {
	if b.Int64() > maxIdxValue {
		return 0, ErrNumOverflow
	}
	return Idx(uint32(b.Int64())), nil
}

// TxID is the identifier of a Hermez network transaction
type TxID Hash // Hash is a guess

// TxType is a string that represents the type of a Hermez network transaction
type TxType string

const (
	// TxTypeExit represents L2->L1 token transfer.  A leaf for this account appears in the exit tree of the block
	TxTypeExit TxType = "Exit"
	// TxTypeWithdrawn represents the balance that was moved from L2->L1 has been widthrawn from the smart contract
	TxTypeWithdrawn TxType = "Withdrawn"
	// TxTypeTransfer represents L2->L2 token transfer
	TxTypeTransfer TxType = "Transfer"
	// TxTypeDeposit represents L1->L2 transfer
	TxTypeDeposit TxType = "Deposit"
	// TxTypeCreateAccountDeposit represents creation of a new leaf in the state tree (newAcconut) + L1->L2 transfer
	TxTypeCreateAccountDeposit TxType = "CreateAccountDeposit"
	// TxTypeCreateAccountDepositTransfer represents L1->L2 transfer + L2->L2 transfer
	TxTypeCreateAccountDepositTransfer TxType = "CreateAccountDepositTransfer"
	// TxTypeDepositTransfer TBD
	TxTypeDepositTransfer TxType = "TxTypeDepositTransfer"
	// TxTypeForceTransfer TBD
	TxTypeForceTransfer TxType = "TxTypeForceTransfer"
	// TxTypeForceExit TBD
	TxTypeForceExit TxType = "TxTypeForceExit"
	// TxTypeTransferToEthAddr TBD
	TxTypeTransferToEthAddr TxType = "TxTypeTransferToEthAddr"
	// TxTypeTransferToBJJ TBD
	TxTypeTransferToBJJ TxType = "TxTypeTransferToBJJ"
)

// Tx is a struct used by the TxSelector & BatchBuilder as a generic type generated from L1Tx & PoolL2Tx
type Tx struct {
	// Generic
	IsL1        bool     `meddler:"is_l1"`
	TxID        TxID     `meddler:"id"`
	Type        TxType   `meddler:"type"`
	Position    int      `meddler:"position"`
	FromIdx     Idx      `meddler:"from_idx"`
	ToIdx       Idx      `meddler:"to_idx"`
	Amount      *big.Int `meddler:"amount,bigint"`
	AmountF     float32  `meddler:"amount_f"`
	TokenID     TokenID  `meddler:"token_id"`
	USD         float32  `meddler:"value_usd,zeroisnull"`
	BatchNum    BatchNum `meddler:"batch_num,zeroisnull"` // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum int64    `meddler:"eth_block_num"`        // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum uint32             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin      bool               `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr     ethCommon.Address  `meddler:"from_eth_addr"`
	FromBJJ         *babyjub.PublicKey `meddler:"from_bjj"`
	LoadAmount      *big.Int           `meddler:"load_amount,bigintnull"`
	// L2
	Fee   FeeSelector `meddler:"fee,zeroisnull"`
	Nonce Nonce       `meddler:"nonce,zeroisnull"`
}
