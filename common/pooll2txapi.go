package common

import (
	"encoding/json"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2TxAPI represents a L2 Tx pool with extra metadata used by the API
type PoolL2TxAPI struct {
	ItemID               uint64                `json:"itemId"`
	TxID                 TxID                  `json:"id"`
	Type                 TxType                `json:"type"`
	FromIdx              string                `json:"fromAccountIndex"`
	EffectiveFromEthAddr *string               `json:"fromHezEthereumAddress"`
	EffectiveFromBJJ     *string               `json:"fromBJJ"`
	ToIdx                *string               `json:"toAccountIndex"`
	EffectiveToEthAddr   *string               `json:"toHezEthereumAddress"`
	EffectiveToBJJ       *string               `json:"toBJJ"`
	Amount               string                `json:"amount"`
	Fee                  FeeSelector           `json:"fee"`
	Nonce                Nonce                 `json:"nonce"`
	State                PoolL2TxState         `json:"state"`
	MaxNumBatch          uint32                `json:"maxNumBatch"`
	Info                 *string               `json:"info"`
	ErrorCode            *int                  `json:"errorCode"`
	ErrorType            *string               `json:"errorType"`
	Signature            babyjub.SignatureComp `json:"signature"`
	Timestamp            time.Time             `json:"timestamp"`
	RqFromIdx            *string               `json:"requestFromAccountIndex"`
	RqToIdx              *string               `json:"requestToAccountIndex"`
	RqToEthAddr          *string               `json:"requestToHezEthereumAddress"`
	RqToBJJ              *string               `json:"requestToBJJ"`
	RqTokenID            *TokenID              `json:"requestTokenId"`
	RqAmount             *string               `json:"requestAmount"`
	RqFee                *FeeSelector          `json:"requestFee"`
	RqNonce              *Nonce                `json:"requestNonce"`
	Token                struct {
		TokenID          TokenID           `json:"id"`
		TokenItemID      uint64            `json:"itemId"`
		TokenEthBlockNum int64             `json:"ethereumBlockNum"`
		TokenEthAddr     ethCommon.Address `json:"ethereumAddress"`
		TokenName        string            `json:"name"`
		TokenSymbol      string            `json:"symbol"`
		TokenDecimals    uint64            `json:"decimals"`
		TokenUSD         *float64          `json:"USD"`
		TokenUSDUpdate   *time.Time        `json:"fiatUpdate"`
	} `json:"token"`
}

// MarshalJSON is used to convert PoolL2TxAPI in JSON
func (tx PoolL2TxAPI) MarshalJSON() ([]byte, error) {
	return json.Marshal(tx)
}

// UnmarshalJSON is used to create a PoolL2TxAPI from JSON data
func (tx *PoolL2TxAPI) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, tx)
	if err != nil {
		return err
	}
	return nil
}
