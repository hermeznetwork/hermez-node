package l2db

import (
	"encoding/json"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolTxAPIView represents a L2 Tx pool with extra metadata used by the API
type PoolTxAPIView struct {
	ItemID               uint64                 `meddler:"item_id"`
	TxID                 common.TxID            `meddler:"tx_id"`
	FromIdx              common.Idx             `meddler:"from_idx"`
	EffectiveFromEthAddr *ethCommon.Address     `meddler:"effective_from_eth_addr"`
	EffectiveFromBJJ     *babyjub.PublicKeyComp `meddler:"effective_from_bjj"`
	ToIdx                *common.Idx            `meddler:"to_idx"`
	EffectiveToEthAddr   *ethCommon.Address     `meddler:"effective_to_eth_addr"`
	EffectiveToBJJ       *babyjub.PublicKeyComp `meddler:"effective_to_bjj"`
	Amount               big.Int                `meddler:"amount"`
	Fee                  common.FeeSelector     `meddler:"fee"`
	Nonce                common.Nonce           `meddler:"nonce"`
	State                common.PoolL2TxState   `meddler:"state"`
	MaxNumBatch          uint32                 `meddler:"max_num_batch,zeroisnull"`
	Info                 *string                `meddler:"info"`
	ErrorCode            *int                   `meddler:"error_code"`
	ErrorType            *string                `meddler:"error_type"`
	Signature            babyjub.SignatureComp  `meddler:"signature"`
	RqFromIdx            *common.Idx            `meddler:"rq_from_idx"`
	RqToIdx              *common.Idx            `meddler:"rq_to_idx"`
	RqToEthAddr          *ethCommon.Address     `meddler:"rq_to_eth_addr"`
	RqToBJJ              *babyjub.PublicKeyComp `meddler:"rq_to_bjj"`
	RqTokenID            *common.TokenID        `meddler:"rq_token_id"`
	RqAmount             big.Int                `meddler:"rq_amount"`
	RqFee                *common.FeeSelector    `meddler:"rq_fee"`
	RqNonce              *common.Nonce          `meddler:"rq_nonce"`
	Type                 common.TxType          `meddler:"tx_type"`
	BatchNum             *common.BatchNum       `meddler:"batch_num"`
	Timestamp            time.Time              `meddler:"timestamp,utctime"`
	TotalItems           uint64                 `meddler:"total_items"`
	TokenID              common.TokenID         `meddler:"token_id"`
	TokenItemID          uint64                 `meddler:"token_item_id"`
	TokenEthBlockNum     int64                  `meddler:"eth_block_num"`
	TokenEthAddr         ethCommon.Address      `meddler:"eth_addr"`
	TokenName            string                 `meddler:"name"`
	TokenSymbol          string                 `meddler:"symbol"`
	TokenDecimals        uint64                 `meddler:"decimals"`
	TokenUSD             *float64               `meddler:"usd"`
	TokenUSDUpdate       *time.Time             `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of PoolTxAPIView
// without the need of auxiliar structs
func (tx PoolTxAPIView) MarshalJSON() ([]byte, error) {
	toMarshal := tx.ToAPI()
	return json.Marshal(toMarshal)
}

// AccountCreationAuthAPI represents an account creation auth in the expected format by the API
type AccountCreationAuthAPI struct {
	EthAddr   ethCommon.Address     `json:"hezEthereumAddress" meddler:"eth_addr" `
	BJJ       babyjub.PublicKeyComp `json:"bjj"                meddler:"bjj" `
	Signature apitypes.EthSignature `json:"signature"          meddler:"signature" `
	Timestamp time.Time             `json:"timestamp"          meddler:"timestamp,utctime"`
}

// ToAPI converts a PoolTxAPIView into PoolL2TxAPI
func (tx *PoolTxAPIView) ToAPI() common.PoolL2TxAPI {
	pooll2apilocal := common.PoolL2TxAPI{
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
