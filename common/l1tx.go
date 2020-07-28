package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// L1Tx is a struct that represents an already forged L1 tx
// WARNING: this struct is very unclear and a complete guess
type L1Tx struct {
	Tx
	PublicKey          babyjub.PublicKey
	LoadAmount         *big.Int // amount transfered from L1 -> L2
	EthBlockNum        uint64
	EthTxHash          eth.Hash
	Position           int // Position among all the L1Txs in that batch
	ToForgeL1TxsNumber uint32
}
