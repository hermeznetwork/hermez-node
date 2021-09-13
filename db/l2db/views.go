package l2db

import (
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// poolTxAPIView represents a L2 Tx pool with extra metadata used by the API
type poolTxAPIView struct {
	ItemID               uint64                `meddler:"item_id"`
	TxID                 common.TxID           `meddler:"tx_id"`
	FromIdx              apitypes.HezIdx       `meddler:"from_idx"`
	EffectiveFromEthAddr *apitypes.HezEthAddr  `meddler:"effective_from_eth_addr"`
	EffectiveFromBJJ     *apitypes.HezBJJ      `meddler:"effective_from_bjj"`
	ToIdx                *apitypes.HezIdx      `meddler:"to_idx"`
	EffectiveToEthAddr   *apitypes.HezEthAddr  `meddler:"effective_to_eth_addr"`
	EffectiveToBJJ       *apitypes.HezBJJ      `meddler:"effective_to_bjj"`
	Amount               apitypes.BigIntStr    `meddler:"amount"`
	Fee                  common.FeeSelector    `meddler:"fee"`
	Nonce                nonce.Nonce           `meddler:"nonce"`
	State                common.PoolL2TxState  `meddler:"state"`
	MaxNumBatch          uint32                `meddler:"max_num_batch,zeroisnull"`
	Info                 *string               `meddler:"info"`
	ErrorCode            *int                  `meddler:"error_code"`
	ErrorType            *string               `meddler:"error_type"`
	Signature            babyjub.SignatureComp `meddler:"signature"`
	RqFromIdx            *apitypes.HezIdx      `meddler:"rq_from_idx"`
	RqToIdx              *apitypes.HezIdx      `meddler:"rq_to_idx"`
	RqToEthAddr          *apitypes.HezEthAddr  `meddler:"rq_to_eth_addr"`
	RqToBJJ              *apitypes.HezBJJ      `meddler:"rq_to_bjj"`
	RqTokenID            *common.TokenID       `meddler:"rq_token_id"`
	RqAmount             *apitypes.BigIntStr   `meddler:"rq_amount"`
	RqFee                *common.FeeSelector   `meddler:"rq_fee"`
	RqNonce              *nonce.Nonce          `meddler:"rq_nonce"`
	Type                 common.TxType         `meddler:"tx_type"`
	BatchNum             *common.BatchNum      `meddler:"batch_num"`
	Timestamp            time.Time             `meddler:"timestamp,utctime"`
	TotalItems           uint64                `meddler:"total_items"`
	TokenID              common.TokenID        `meddler:"token_id"`
	TokenItemID          uint64                `meddler:"token_item_id"`
	TokenEthBlockNum     int64                 `meddler:"eth_block_num"`
	TokenEthAddr         ethCommon.Address     `meddler:"eth_addr"`
	TokenName            string                `meddler:"name"`
	TokenSymbol          string                `meddler:"symbol"`
	TokenDecimals        uint64                `meddler:"decimals"`
	TokenUSD             *float64              `meddler:"usd"`
	TokenUSDUpdate       *time.Time            `meddler:"usd_update"`
}

// AccountCreationAuthAPI represents an account creation auth in the expected format by the API
type AccountCreationAuthAPI struct {
	EthAddr   apitypes.HezEthAddr   `json:"hezEthereumAddress" meddler:"eth_addr" `
	BJJ       apitypes.HezBJJ       `json:"bjj"                meddler:"bjj" `
	Signature apitypes.EthSignature `json:"signature"          meddler:"signature" `
	Timestamp time.Time             `json:"timestamp"          meddler:"timestamp,utctime"`
}

// ToAPI converts a poolTxAPIView into TxL2
func (tx *poolTxAPIView) ToAPI() apitypes.TxL2 {
	pooll2apilocal := apitypes.TxL2{
		ItemID:               tx.ItemID,
		TxID:                 tx.TxID,
		Type:                 tx.Type,
		FromIdx:              tx.FromIdx,
		EffectiveFromEthAddr: tx.EffectiveFromEthAddr,
		EffectiveFromBJJ:     tx.EffectiveFromBJJ,
		ToIdx:                tx.ToIdx,
		EffectiveToEthAddr:   tx.EffectiveToEthAddr,
		EffectiveToBJJ:       tx.EffectiveToBJJ,
		Amount:               tx.Amount,
		Fee:                  tx.Fee,
		Nonce:                tx.Nonce,
		State:                tx.State,
		MaxNumBatch:          tx.MaxNumBatch,
		BatchNum:             tx.BatchNum,
		Info:                 tx.Info,
		ErrorCode:            tx.ErrorCode,
		ErrorType:            tx.ErrorType,
		Signature:            tx.Signature,
		Timestamp:            tx.Timestamp,
		RqFromIdx:            tx.RqFromIdx,
		RqToIdx:              tx.RqToIdx,
		RqToEthAddr:          tx.RqToEthAddr,
		RqToBJJ:              tx.RqToBJJ,
		RqTokenID:            tx.RqTokenID,
		RqAmount:             tx.RqAmount,
		RqFee:                tx.RqFee,
		RqNonce:              tx.RqNonce,
	}
	pooll2apilocal.Token.TokenID = tx.TokenID
	pooll2apilocal.Token.TokenItemID = tx.TokenItemID
	pooll2apilocal.Token.TokenEthBlockNum = tx.TokenEthBlockNum
	pooll2apilocal.Token.TokenEthAddr = tx.TokenEthAddr
	pooll2apilocal.Token.TokenName = tx.TokenName
	pooll2apilocal.Token.TokenSymbol = tx.TokenSymbol
	pooll2apilocal.Token.TokenDecimals = tx.TokenDecimals
	pooll2apilocal.Token.TokenUSD = tx.TokenUSD
	pooll2apilocal.Token.TokenUSDUpdate = tx.TokenUSDUpdate

	return pooll2apilocal
}
