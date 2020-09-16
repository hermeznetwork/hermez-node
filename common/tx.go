package common

import (
	"math/big"
)

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
	TxID     TxID
	FromIdx  Idx // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx    Idx // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositTransfer
	Amount   *big.Int
	Nonce    Nonce // effective 40 bits used
	Fee      FeeSelector
	Type     TxType   // optional, descrives which kind of tx it's
	BatchNum BatchNum // batchNum in which this tx was forged. Presence indicates "forged" state.
}
