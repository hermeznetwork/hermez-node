package common

import (
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the coordinator hat is waiting to be forged
type PoolL2Tx struct {
	Tx
	ToBJJ              babyjub.PublicKey
	Status             PoolL2TxStatus
	RqTxCompressedData []byte // 253 bits, optional for atomic txs
	RqTx               RqTx
	Timestamp          time.Time         // time when added to the tx pool
	Signature          babyjub.Signature // tx signature
	ToEthAddr          eth.Address
}

// RqTx Transaction Data used to indicate that a transaction depends on another transaction
type RqTx struct {
	FromEthAddr eth.Address
	ToEthAddr   eth.Address
	TokenID     TokenID
	Amount      *big.Int
	FeeSelector FeeSelector
	Nonce       uint64 // effective 48 bits used
}

// PoolL2TxStatus is a struct that represents the status of a L2 transaction
type PoolL2TxStatus string

const (
	// PoolL2TxStatusPending represents a valid L2Tx that hasn't started the forging process
	PoolL2TxStatusPending PoolL2TxStatus = "Pending"
	// PoolL2TxStatusForging represents a valid L2Tx that has started the forging process
	PoolL2TxStatusForging PoolL2TxStatus = "Forging"
	// PoolL2TxStatusForged represents a L2Tx that has already been forged
	PoolL2TxStatusForged PoolL2TxStatus = "Forged"
	// PoolL2TxStatusInvalid represents a L2Tx that has been invalidated
	PoolL2TxStatusInvalid PoolL2TxStatus = "Invalid"
)
