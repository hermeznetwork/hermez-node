package l2db

import (
	"encoding/json"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolTxAPIView represents a L2 Tx pool with extra metadata used by the API
type PoolTxAPIView struct {
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
	Nonce                common.Nonce          `meddler:"nonce"`
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
	RqNonce              *common.Nonce         `meddler:"rq_nonce"`
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

// MarshalJSON is used to neast some of the fields of PoolTxAPIView
// without the need of auxiliar structs
func (tx PoolTxAPIView) MarshalJSON() ([]byte, error) {
	toMarshal := common.PoolL2TxAPI{
		ItemID:               tx.ItemID,
		TxID:                 tx.TxID,
		Type:                 tx.Type,
		FromIdx:              string(tx.FromIdx),
		EffectiveFromEthAddr: (*string)(tx.EffectiveFromEthAddr),
		EffectiveFromBJJ:     (*string)(tx.EffectiveFromBJJ),
		ToIdx:                (*string)(tx.ToIdx),
		EffectiveToEthAddr:   (*string)(tx.EffectiveToEthAddr),
		EffectiveToBJJ:       (*string)(tx.EffectiveToBJJ),
		Amount:               string(tx.Amount),
		Fee:                  tx.Fee,
		Nonce:                tx.Nonce,
		State:                tx.State,
		MaxNumBatch:          tx.MaxNumBatch,
		Info:                 tx.Info,
		ErrorCode:            tx.ErrorCode,
		ErrorType:            tx.ErrorType,
		Signature:            tx.Signature,
		Timestamp:            tx.Timestamp,
		RqFromIdx:            (*string)(tx.RqFromIdx),
		RqToIdx:              (*string)(tx.RqToIdx),
		RqToEthAddr:          (*string)(tx.RqToEthAddr),
		RqToBJJ:              (*string)(tx.RqToBJJ),
		RqTokenID:            tx.RqTokenID,
		RqAmount:             (*string)(tx.RqAmount),
		RqFee:                tx.RqFee,
		RqNonce:              tx.RqNonce,
	}
	toMarshal.Token.TokenID = tx.TokenID
	toMarshal.Token.TokenItemID = tx.TokenItemID
	toMarshal.Token.TokenEthBlockNum = tx.TokenEthBlockNum
	toMarshal.Token.TokenEthAddr = tx.TokenEthAddr
	toMarshal.Token.TokenName = tx.TokenName
	toMarshal.Token.TokenSymbol = tx.TokenSymbol
	toMarshal.Token.TokenDecimals = tx.TokenDecimals
	toMarshal.Token.TokenUSD = tx.TokenUSD
	toMarshal.Token.TokenUSDUpdate = tx.TokenUSDUpdate
	return json.Marshal(toMarshal)
}

// AccountCreationAuthAPI represents an account creation auth in the expected format by the API
type AccountCreationAuthAPI struct {
	EthAddr   apitypes.HezEthAddr   `json:"hezEthereumAddress" meddler:"eth_addr" `
	BJJ       apitypes.HezBJJ       `json:"bjj"                meddler:"bjj" `
	Signature apitypes.EthSignature `json:"signature"          meddler:"signature" `
	Timestamp time.Time             `json:"timestamp"          meddler:"timestamp,utctime"`
}
