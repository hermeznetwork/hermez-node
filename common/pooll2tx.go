package common

import (
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the operator hat is waiting to be forged
type PoolL2Tx struct {
	Tx
	Status             PoolL2TxStatus
	RqTxCompressedData []byte            // 253 bits, optional for atomic txs
	RqToEthAddr        eth.Address       // optional for atomic txs
	RqToBjj            babyjub.PublicKey // optional for atomic txs
	RqFromeEthAddr     eth.Address       // optional for atomic txs
	Received           time.Time         // time when added to the tx pool
	Signature          babyjub.Signature // tx signature
}

// PoolL2TxStatus is a struct that represents the status of a L2 transaction
type PoolL2TxStatus string

const (
	// Pending represents a valid L2Tx that hasn't started the forging process
	Pending PoolL2TxStatus = "Pending"
	// Forging represents a valid L2Tx that has started the forging process
	Forging PoolL2TxStatus = "Forging"
	// Forged represents a L2Tx that has already been forged
	Forged PoolL2TxStatus = "Forged"
	// Invalid represents a L2Tx that has been invalidated
	Invalid PoolL2TxStatus = "Invalid"
)
