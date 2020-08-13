package common

import (
	"math/big"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2Tx is a struct that represents a L2Tx sent by an account to the coordinator hat is waiting to be forged
type PoolL2Tx struct {
<<<<<<< HEAD
	Tx
	ToBJJ              babyjub.PublicKey
	Status             PoolL2TxStatus
	RqTxCompressedData []byte // 253 bits, optional for atomic txs
	RqTx               RqTx
	Timestamp          time.Time         // time when added to the tx pool
	Signature          babyjub.Signature // tx signature
	ToEthAddr          eth.Address
	AbsoluteFee        float64 // TODO add methods to calculate this value from Tx.Fee tables + priceupdater tables
=======
	TxID               TxID              `meddler:"tx_id"`
	FromIdx            Idx               `meddler:"from_idx"` // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	ToIdx              Idx               `meddler:"to_idx"`   // ToIdx is ignored in L1Tx/Deposit, but used in the L1Tx/DepositAndTransfer
	ToEthAddr          eth.Address       `meddler:"to_eth_addr"`
	ToBJJ              babyjub.PublicKey `meddler:"to_bjj,json"` // TODO: stop using json, use scanner/valuer
	TokenID            TokenID           `meddler:"token_id"`
	Amount             *big.Int          `meddler:"amount,bigint"` // TODO: change to float16
	Fee                FeeSelector       `meddler:"fee"`
	Nonce              uint64            `meddler:"nonce"` // effective 48 bits used
	State              PoolL2TxState     `meddler:"state"`
	BatchNum           BatchNum          `meddler:"batch_num,zeroisnull"` // batchNum in which this tx was forged. Presence indicates "forged" state.
	RqFromIdx          Idx               `meddler:"rq_from_idx"`          // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	RqToIdx            Idx               `meddler:"rq_to_idx"`            // FromIdx is used by L1Tx/Deposit to indicate the Idx receiver of the L1Tx.LoadAmount (deposit)
	RqToEthAddr        eth.Address       `meddler:"rq_to_eth_addr"`
	RqToBJJ            babyjub.PublicKey `meddler:"rq_to_bjj,json"` // TODO: stop using json, use scanner/valuer
	RqTokenID          TokenID           `meddler:"rq_token_id"`
	RqAmount           *big.Int          `meddler:"rq_amount,bigintnull"` // TODO: change to float16
	RqFee              FeeSelector       `meddler:"rq_fee"`
	RqNonce            uint64            `meddler:"rq_nonce,zeroisnull"` // effective 48 bits used
	Signature          babyjub.Signature `meddler:"-"`                   // tx signature
	Timestamp          time.Time         `meddler:"timestamp"`           // time when added to the tx pool
	Type               TxType            `meddler:"-"`                   // optional, descrives which kind of tx it's
	RqTxCompressedData []byte            `meddler:"-"`                   // 253 bits, optional for atomic txs
>>>>>>> WIP
}

// PoolL2TxState is a struct that represents the status of a L2 transaction
type PoolL2TxState string

const (
	// PoolL2TxStatePending represents a valid L2Tx that hasn't started the forging process
	PoolL2TxStatePending PoolL2TxState = "pend"
	// PoolL2TxStateForging represents a valid L2Tx that has started the forging process
	PoolL2TxStateForging PoolL2TxState = "fing"
	// PoolL2TxStateForged represents a L2Tx that has already been forged
	PoolL2TxStateForged PoolL2TxState = "fged"
	// PoolL2TxStateInvalid represents a L2Tx that has been invalidated
	PoolL2TxStateInvalid PoolL2TxState = "invl"
)

// Tx implements the Txer interface
func (pooltx *PoolL2Tx) Tx() Tx {
	return Tx{
		TxID:     pooltx.TxID,
		FromIdx:  pooltx.FromIdx,
		ToIdx:    pooltx.ToIdx,
		TokenID:  pooltx.TokenID,
		Amount:   pooltx.Amount,
		Nonce:    pooltx.Nonce,
		Fee:      pooltx.Fee,
		Type:     pooltx.Type,
		BatchNum: pooltx.BatchNum,
	}
}
