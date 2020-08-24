package common

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// L1Tx is a struct that represents a L1 tx
type L1Tx struct {
	// Stored in DB: mandatory fileds
	TxID            TxID               `meddler:"tx_id"`
	ToForgeL1TxsNum uint32             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	Position        int                `meddler:"position"`
	UserOrigin      bool               `meddler:"user_origin"` // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromIdx         Idx                `meddler:"from_idx"`    // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	FromEthAddr     ethCommon.Address  `meddler:"from_eth_addr"`
	FromBJJ         *babyjub.PublicKey `meddler:"from_bjj"`
	ToIdx           Idx                `meddler:"to_idx"` // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	TokenID         TokenID            `meddler:"token_id"`
	Amount          *big.Int           `meddler:"amount,bigint"`
	LoadAmount      *big.Int           `meddler:"load_amount,bigint"`
	EthBlockNum     uint64             `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// Extra metadata, may be uninitialized
	Type TxType `meddler:"-"` // optional, descrives which kind of tx it's
}

func (tx *L1Tx) Tx() *Tx {
	return &Tx{
		TxID:    tx.TxID,
		FromIdx: tx.FromIdx,
		ToIdx:   tx.ToIdx,
		TokenID: tx.TokenID,
		Amount:  tx.Amount,
		Nonce:   0,
		Fee:     0,
		Type:    tx.Type,
	}
}
