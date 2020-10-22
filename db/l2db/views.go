package l2db

import (
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2TxWrite holds the necessary data to perform inserts in tx_pool
type PoolL2TxWrite struct {
	TxID        common.TxID           `meddler:"tx_id"`
	FromIdx     common.Idx            `meddler:"from_idx"`
	ToIdx       *common.Idx           `meddler:"to_idx"`
	ToEthAddr   *ethCommon.Address    `meddler:"to_eth_addr"`
	ToBJJ       *babyjub.PublicKey    `meddler:"to_bjj"`
	TokenID     common.TokenID        `meddler:"token_id"`
	Amount      *big.Int              `meddler:"amount,bigint"`
	AmountFloat float64               `meddler:"amount_f"`
	Fee         common.FeeSelector    `meddler:"fee"`
	Nonce       common.Nonce          `meddler:"nonce"`
	State       common.PoolL2TxState  `meddler:"state"`
	Signature   babyjub.SignatureComp `meddler:"signature"`
	RqFromIdx   *common.Idx           `meddler:"rq_from_idx"`
	RqToIdx     *common.Idx           `meddler:"rq_to_idx"`
	RqToEthAddr *ethCommon.Address    `meddler:"rq_to_eth_addr"`
	RqToBJJ     *babyjub.PublicKey    `meddler:"rq_to_bjj"`
	RqTokenID   *common.TokenID       `meddler:"rq_token_id"`
	RqAmount    *big.Int              `meddler:"rq_amount,bigintnull"`
	RqFee       *common.FeeSelector   `meddler:"rq_fee"`
	RqNonce     *common.Nonce         `meddler:"rq_nonce"`
	Type        common.TxType         `meddler:"tx_type"`
}

// PoolL2TxRead represents a L2 Tx pool with extra metadata used by the API
type PoolL2TxRead struct {
	TxID        common.TxID           `meddler:"tx_id"`
	FromIdx     common.Idx            `meddler:"from_idx"`
	ToIdx       *common.Idx           `meddler:"to_idx"`
	ToEthAddr   *ethCommon.Address    `meddler:"to_eth_addr"`
	ToBJJ       *babyjub.PublicKey    `meddler:"to_bjj"`
	Amount      *big.Int              `meddler:"amount,bigint"`
	Fee         common.FeeSelector    `meddler:"fee"`
	Nonce       common.Nonce          `meddler:"nonce"`
	State       common.PoolL2TxState  `meddler:"state"`
	Signature   babyjub.SignatureComp `meddler:"signature"`
	RqFromIdx   *common.Idx           `meddler:"rq_from_idx"`
	RqToIdx     *common.Idx           `meddler:"rq_to_idx"`
	RqToEthAddr *ethCommon.Address    `meddler:"rq_to_eth_addr"`
	RqToBJJ     *babyjub.PublicKey    `meddler:"rq_to_bjj"`
	RqTokenID   *common.TokenID       `meddler:"rq_token_id"`
	RqAmount    *big.Int              `meddler:"rq_amount,bigintnull"`
	RqFee       *common.FeeSelector   `meddler:"rq_fee"`
	RqNonce     *common.Nonce         `meddler:"rq_nonce"`
	Type        common.TxType         `meddler:"tx_type"`
	// Extra read fileds
	BatchNum         *common.BatchNum  `meddler:"batch_num"`
	Timestamp        time.Time         `meddler:"timestamp,utctime"`
	TotalItems       int               `meddler:"total_items"`
	TokenID          common.TokenID    `meddler:"token_id"`
	TokenEthBlockNum int64             `meddler:"eth_block_num"`
	TokenEthAddr     ethCommon.Address `meddler:"eth_addr"`
	TokenName        string            `meddler:"name"`
	TokenSymbol      string            `meddler:"symbol"`
	TokenDecimals    uint64            `meddler:"decimals"`
	TokenUSD         *float64          `meddler:"usd"`
	TokenUSDUpdate   *time.Time        `meddler:"usd_update"`
}
