package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

// Tx is a struct that represents a Hermez network transaction
type Tx struct {
	ID          TxID
	FromEthAddr eth.Address
	ToEthAddr   eth.Address
	FromIdx     uint32
	ToIdx       uint32
	TokenID     TokenID
	Amount      *big.Int
	Nonce       uint64 // effective 48 bits used
	FeeSelector FeeSelector
	Type        TxType // optional, descrives which kind of tx it's
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
)
