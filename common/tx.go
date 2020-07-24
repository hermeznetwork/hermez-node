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
	// Exit represents L2->L1 token transfer.  A leaf for this account appears in the exit tree of the block
	Exit TxType = "Exit"
	// Withdrawn represents the balance that was moved from L2->L1 has been widthrawn from the smart contract
	Withdrawn TxType = "Withdrawn"
	// Transfer represents L2->L2 token transfer
	Transfer TxType = "Transfer"
	// Deposit represents L1->L2 transfer
	Deposit TxType = "Deposit"
	// CreateAccountDeposit represents creation of a new leaf in the state tree (newAcconut) + L1->L2 transfer
	CreateAccountDeposit TxType = "CreateAccountDeposit"
	// CreateAccountDepositTransfer represents L1->L2 transfer + L2->L2 transfer
	CreateAccountDepositTransfer TxType = "CreateAccountDepositTransfer"
)
