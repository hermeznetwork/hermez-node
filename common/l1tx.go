package common

import (
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	TxID            TxID        `meddler:"tx_id"`
	FromIdx         Idx         `meddler:"from_idx"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx           Idx         `meddler:"to_idx"`   // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	TokenID         TokenID     `meddler:"token_id"`
	Amount          *big.Int    `meddler:"amount,bigint"` // TODO: change to float16
	Nonce           uint64      `meddler:"nonce"`         // effective 48 bits used
	Fee             FeeSelector `meddler:"fee"`
	Type            TxType      `meddler:"-"`         // optional, descrives which kind of tx it's
	BatchNum        BatchNum    `meddler:"batch_num"` // batchNum in which this tx was forged. Presence indicates "forged" state.
	UserOrigin      bool        // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
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

// Tx implements the Txer interface
func (l1tx *L1Tx) Tx() Tx {
	return Tx{
		TxID:     l1tx.TxID,
		FromIdx:  l1tx.FromIdx,
		ToIdx:    l1tx.ToIdx,
		TokenID:  l1tx.TokenID,
		Amount:   l1tx.Amount,
		Nonce:    l1tx.Nonce,
		Fee:      l1tx.Fee,
		Type:     l1tx.Type,
		BatchNum: l1tx.BatchNum,
	}
}
