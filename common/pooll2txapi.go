package common

import (
	"encoding/json"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// PoolL2TxAPI represents a L2 Tx pool with extra metadata used by the API
type PoolL2TxAPI struct {
	ItemID               uint64
	TxID                 TxID
	Type                 TxType
	FromIdx              Idx
	EffectiveFromEthAddr *ethCommon.Address
	EffectiveFromBJJ     *babyjub.PublicKeyComp
	ToIdx                *Idx
	EffectiveToEthAddr   *ethCommon.Address
	EffectiveToBJJ       *babyjub.PublicKeyComp
	Amount               big.Int
	Fee                  FeeSelector
	Nonce                Nonce
	State                PoolL2TxState
	MaxNumBatch          uint32
	Info                 *string
	ErrorCode            *int
	ErrorType            *string
	Signature            babyjub.SignatureComp
	Timestamp            time.Time
	RqFromIdx            *Idx
	RqToIdx              *Idx
	RqToEthAddr          *ethCommon.Address
	RqToBJJ              *babyjub.PublicKeyComp
	RqTokenID            *TokenID
	RqAmount             big.Int
	RqFee                *FeeSelector
	RqNonce              *Nonce
	Token                struct {
		TokenID          TokenID
		TokenItemID      uint64
		TokenEthBlockNum int64
		TokenEthAddr     ethCommon.Address
		TokenName        string
		TokenSymbol      string
		TokenDecimals    uint64
		TokenUSD         *float64
		TokenUSDUpdate   *time.Time
	}
}

// MarshalJSON is used to convert PoolL2TxAPI in JSON
func (tx PoolL2TxAPI) MarshalJSON() ([]byte, error) {
	type jsonToken struct {
		TokenID          TokenID           `json:"id"`
		TokenItemID      uint64            `json:"itemId"`
		TokenEthBlockNum int64             `json:"ethereumBlockNum"`
		TokenEthAddr     ethCommon.Address `json:"ethereumAddress"`
		TokenName        string            `json:"name"`
		TokenSymbol      string            `json:"symbol"`
		TokenDecimals    uint64            `json:"decimals"`
		TokenUSD         *float64          `json:"USD"`
		TokenUSDUpdate   *time.Time        `json:"fiatUpdate"`
	}
	type jsonFormat struct {
		ItemID               uint64                 `json:"itemId"`
		TxID                 TxID                   `json:"id"`
		Type                 TxType                 `json:"type"`
		FromIdx              Idx                    `json:"fromAccountIndex"`
		EffectiveFromEthAddr *ethCommon.Address     `json:"fromHezEthereumAddress"`
		EffectiveFromBJJ     *babyjub.PublicKeyComp `json:"fromBJJ"`
		ToIdx                *Idx                   `json:"toAccountIndex"`
		EffectiveToEthAddr   *ethCommon.Address     `json:"toHezEthereumAddress"`
		EffectiveToBJJ       *babyjub.PublicKeyComp `json:"toBJJ"`
		Amount               big.Int                `json:"amount"`
		Fee                  FeeSelector            `json:"fee"`
		Nonce                Nonce                  `json:"nonce"`
		State                PoolL2TxState          `json:"state"`
		MaxNumBatch          uint32                 `json:"maxNumBatch"`
		Info                 *string                `json:"info"`
		ErrorCode            *int                   `json:"errorCode"`
		ErrorType            *string                `json:"errorType"`
		Signature            babyjub.SignatureComp  `json:"signature"`
		Timestamp            time.Time              `json:"timestamp"`
		RqFromIdx            *Idx                   `json:"requestFromAccountIndex"`
		RqToIdx              *Idx                   `json:"requestToAccountIndex"`
		RqToEthAddr          *ethCommon.Address     `json:"requestToHezEthereumAddress"`
		RqToBJJ              *babyjub.PublicKeyComp `json:"requestToBJJ"`
		RqTokenID            *TokenID               `json:"requestTokenId"`
		RqAmount             big.Int                `json:"requestAmount"`
		RqFee                *FeeSelector           `json:"requestFee"`
		RqNonce              *Nonce                 `json:"requestNonce"`
		Token                jsonToken              `json:"token"`
	}
	toMarshal := jsonFormat{
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
		Token: jsonToken{
			TokenID:          tx.Token.TokenID,
			TokenItemID:      tx.Token.TokenItemID,
			TokenEthBlockNum: tx.Token.TokenEthBlockNum,
			TokenEthAddr:     tx.Token.TokenEthAddr,
			TokenName:        tx.Token.TokenName,
			TokenSymbol:      tx.Token.TokenSymbol,
			TokenDecimals:    tx.Token.TokenDecimals,
			TokenUSD:         tx.Token.TokenUSD,
			TokenUSDUpdate:   tx.Token.TokenUSDUpdate,
		},
	}
	return json.Marshal(toMarshal)
}

// UnmarshalJSON is used to create a PoolL2TxAPI from JSON data
func (tx *PoolL2TxAPI) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, tx)
	if err != nil {
		return err
	}
	return nil
}
