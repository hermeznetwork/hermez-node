package common

import (
	"encoding/binary"
	"math/big"
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

// IdxFromBigInt converts a *big.Int to Idx type
func IdxFromBigInt(b *big.Int) (Idx, error) {
	if b.Int64() > 0xffffffff { // 2**32-1
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
	// TxTypeCreateAccountDepositAndTransfer represents L1->L2 transfer + L2->L2 transfer
	TxTypeCreateAccountDepositAndTransfer TxType = "CreateAccountDepositAndTransfer"
	// TxTypeDepositAndTransfer TBD
	TxTypeDepositAndTransfer TxType = "TxTypeDepositAndTransfer"
	// TxTypeForceTransfer TBD
	TxTypeForceTransfer TxType = "TxTypeForceTransfer"
	// TxTypeForceExit TBD
	TxTypeForceExit TxType = "TxTypeForceExit"
	// TxTypeTransferToEthAddr TBD
	TxTypeTransferToEthAddr TxType = "TxTypeTransferToEthAddr"
	// TxTypeTransferToBJJ TBD
	TxTypeTransferToBJJ TxType = "TxTypeTransferToBJJ"
)
