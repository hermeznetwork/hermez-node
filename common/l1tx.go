package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	Tx
	UserOrigin      bool // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	PublicKey       babyjub.PublicKey
	LoadAmount      *big.Int // amount transfered from L1 -> L2
	EthBlockNum     uint64   // Ethereum Block Number in which this L1Tx was added to the queue
	EthTxHash       eth.Hash // TxHash that added this L1Tx to the queue
	Position        int      // Position among all the L1Txs in that batch
	ToForgeL1TxsNum uint32   // toForgeL1TxsNum in which the tx was forged / will be forged
	FromBJJ         babyjub.PublicKey
	CreateAccount   bool // "from" + token ID is a new account
	FromEthAddr     eth.Address
}
