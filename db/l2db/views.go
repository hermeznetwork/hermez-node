package l2db

import (
	"encoding/json"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/api/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolTxAPI represents a L2 Tx pool with extra metadata used by the API
type PoolTxAPI struct {
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
	Info                 *string               `meddler:"info"`
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
	// Extra read fileds
	BatchNum         *common.BatchNum  `meddler:"batch_num"`
	Timestamp        time.Time         `meddler:"timestamp,utctime"`
	TotalItems       uint64            `meddler:"total_items"`
	TokenID          common.TokenID    `meddler:"token_id"`
	TokenItemID      uint64            `meddler:"token_item_id"`
	TokenEthBlockNum int64             `meddler:"eth_block_num"`
	TokenEthAddr     ethCommon.Address `meddler:"eth_addr"`
	TokenName        string            `meddler:"name"`
	TokenSymbol      string            `meddler:"symbol"`
	TokenDecimals    uint64            `meddler:"decimals"`
	TokenUSD         *float64          `meddler:"usd"`
	TokenUSDUpdate   *time.Time        `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of PoolTxAPI
// without the need of auxiliar structs
func (tx PoolTxAPI) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"itemId":                      tx.ItemID,
		"id":                          tx.TxID,
		"type":                        tx.Type,
		"fromAccountIndex":            tx.FromIdx,
		"fromHezEthereumAddress":      tx.EffectiveFromEthAddr,
		"fromBJJ":                     tx.EffectiveFromBJJ,
		"toAccountIndex":              tx.ToIdx,
		"toHezEthereumAddress":        tx.EffectiveToEthAddr,
		"toBJJ":                       tx.EffectiveToBJJ,
		"amount":                      tx.Amount,
		"fee":                         tx.Fee,
		"nonce":                       tx.Nonce,
		"state":                       tx.State,
		"info":                        tx.Info,
		"signature":                   tx.Signature,
		"timestamp":                   tx.Timestamp,
		"requestFromAccountIndex":     tx.RqFromIdx,
		"requestToAccountIndex":       tx.RqToIdx,
		"requestToHezEthereumAddress": tx.RqToEthAddr,
		"requestToBJJ":                tx.RqToBJJ,
		"requestTokenId":              tx.RqTokenID,
		"requestAmount":               tx.RqAmount,
		"requestFee":                  tx.RqFee,
		"requestNonce":                tx.RqNonce,
		"token": map[string]interface{}{
			"id":               tx.TokenID,
			"itemId":           tx.TokenItemID,
			"ethereumBlockNum": tx.TokenEthBlockNum,
			"ethereumAddress":  tx.TokenEthAddr,
			"name":             tx.TokenName,
			"symbol":           tx.TokenSymbol,
			"decimals":         tx.TokenDecimals,
			"USD":              tx.TokenUSD,
			"fiatUpdate":       tx.TokenUSDUpdate,
		},
	})
}

// AccountCreationAuthAPI represents an account creation auth in the expected format by the API
type AccountCreationAuthAPI struct {
	EthAddr   apitypes.HezEthAddr   `json:"hezEthereumAddress" meddler:"eth_addr" `
	BJJ       apitypes.HezBJJ       `json:"bjj"                meddler:"bjj" `
	Signature apitypes.EthSignature `json:"signature"          meddler:"signature" `
	Timestamp time.Time             `json:"timestamp"          meddler:"timestamp,utctime"`
}
