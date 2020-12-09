package historydb

import (
	"encoding/json"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/apitypes"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

// TxAPI is a representation of a generic Tx with additional information
// required by the API, and extracted by joining block and token tables
type TxAPI struct {
	// Generic
	IsL1        bool                 `meddler:"is_l1"`
	TxID        common.TxID          `meddler:"id"`
	ItemID      uint64               `meddler:"item_id"`
	Type        common.TxType        `meddler:"type"`
	Position    int                  `meddler:"position"`
	FromIdx     *apitypes.HezIdx     `meddler:"from_idx"`
	FromEthAddr *apitypes.HezEthAddr `meddler:"from_eth_addr"`
	FromBJJ     *apitypes.HezBJJ     `meddler:"from_bjj"`
	ToIdx       apitypes.HezIdx      `meddler:"to_idx"`
	ToEthAddr   *apitypes.HezEthAddr `meddler:"to_eth_addr"`
	ToBJJ       *apitypes.HezBJJ     `meddler:"to_bjj"`
	Amount      apitypes.BigIntStr   `meddler:"amount"`
	HistoricUSD *float64             `meddler:"amount_usd"`
	BatchNum    *common.BatchNum     `meddler:"batch_num"`     // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum int64                `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum          *int64              `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin               *bool               `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	DepositAmount            *apitypes.BigIntStr `meddler:"deposit_amount"`
	HistoricDepositAmountUSD *float64            `meddler:"deposit_amount_usd"`
	AmountSuccess            bool                `meddler:"amount_success"`
	DepositAmountSuccess     bool                `meddler:"deposit_amount_success"`
	// L2
	Fee            *common.FeeSelector `meddler:"fee"`
	HistoricFeeUSD *float64            `meddler:"fee_usd"`
	Nonce          *common.Nonce       `meddler:"nonce"`
	// API extras
	Timestamp        time.Time         `meddler:"timestamp,utctime"`
	TotalItems       uint64            `meddler:"total_items"`
	FirstItem        uint64            `meddler:"first_item"`
	LastItem         uint64            `meddler:"last_item"`
	TokenID          common.TokenID    `meddler:"token_id"`
	TokenItemID      uint64            `meddler:"token_item_id"`
	TokenEthBlockNum int64             `meddler:"token_block"`
	TokenEthAddr     ethCommon.Address `meddler:"eth_addr"`
	TokenName        string            `meddler:"name"`
	TokenSymbol      string            `meddler:"symbol"`
	TokenDecimals    uint64            `meddler:"decimals"`
	TokenUSD         *float64          `meddler:"usd"`
	TokenUSDUpdate   *time.Time        `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of TxAPI
// without the need of auxiliar structs
func (tx TxAPI) MarshalJSON() ([]byte, error) {
	jsonTx := map[string]interface{}{
		"id":                     tx.TxID,
		"itemId":                 tx.ItemID,
		"type":                   tx.Type,
		"position":               tx.Position,
		"fromAccountIndex":       tx.FromIdx,
		"fromHezEthereumAddress": tx.FromEthAddr,
		"fromBJJ":                tx.FromBJJ,
		"toAccountIndex":         tx.ToIdx,
		"toHezEthereumAddress":   tx.ToEthAddr,
		"toBJJ":                  tx.ToBJJ,
		"amount":                 tx.Amount,
		"batchNum":               tx.BatchNum,
		"historicUSD":            tx.HistoricUSD,
		"timestamp":              tx.Timestamp,
		"L1Info":                 nil,
		"L2Info":                 nil,
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
	}
	if tx.IsL1 {
		jsonTx["L1orL2"] = "L1"
		amountSuccess := tx.AmountSuccess
		depositAmountSuccess := tx.DepositAmountSuccess
		if tx.BatchNum == nil {
			amountSuccess = false
			depositAmountSuccess = false
		}
		jsonTx["L1Info"] = map[string]interface{}{
			"toForgeL1TransactionsNum": tx.ToForgeL1TxsNum,
			"userOrigin":               tx.UserOrigin,
			"depositAmount":            tx.DepositAmount,
			"amountSuccess":            amountSuccess,
			"depositAmountSuccess":     depositAmountSuccess,
			"historicDepositAmountUSD": tx.HistoricDepositAmountUSD,
			"ethereumBlockNum":         tx.EthBlockNum,
		}
	} else {
		jsonTx["L1orL2"] = "L2"
		jsonTx["L2Info"] = map[string]interface{}{
			"fee":            tx.Fee,
			"historicFeeUSD": tx.HistoricFeeUSD,
			"nonce":          tx.Nonce,
		}
	}
	return json.Marshal(jsonTx)
}

// txWrite is an representatiion that merges common.L1Tx and common.L2Tx
// in order to perform inserts into tx table
// EffectiveAmount and LoadEffectiveAmount are not set since they have default values in the DB
type txWrite struct {
	// Generic
	IsL1        bool             `meddler:"is_l1"`
	TxID        common.TxID      `meddler:"id"`
	Type        common.TxType    `meddler:"type"`
	Position    int              `meddler:"position"`
	FromIdx     *common.Idx      `meddler:"from_idx"`
	ToIdx       common.Idx       `meddler:"to_idx"`
	Amount      *big.Int         `meddler:"amount,bigint"`
	AmountFloat float64          `meddler:"amount_f"`
	TokenID     common.TokenID   `meddler:"token_id"`
	BatchNum    *common.BatchNum `meddler:"batch_num"`     // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum int64            `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum    *int64             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin         *bool              `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr        *ethCommon.Address `meddler:"from_eth_addr"`
	FromBJJ            *babyjub.PublicKey `meddler:"from_bjj"`
	DepositAmount      *big.Int           `meddler:"deposit_amount,bigintnull"`
	DepositAmountFloat *float64           `meddler:"deposit_amount_f"`
	// L2
	Fee   *common.FeeSelector `meddler:"fee"`
	Nonce *common.Nonce       `meddler:"nonce"`
}

// TokenWithUSD add USD info to common.Token
type TokenWithUSD struct {
	ItemID      uint64            `json:"itemId" meddler:"item_id"`
	TokenID     common.TokenID    `json:"id" meddler:"token_id"`
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"` // Ethereum block number in which this token was registered
	EthAddr     ethCommon.Address `json:"ethereumAddress" meddler:"eth_addr"`
	Name        string            `json:"name" meddler:"name"`
	Symbol      string            `json:"symbol" meddler:"symbol"`
	Decimals    uint64            `json:"decimals" meddler:"decimals"`
	USD         *float64          `json:"USD" meddler:"usd"`
	USDUpdate   *time.Time        `json:"fiatUpdate" meddler:"usd_update,utctime"`
	TotalItems  uint64            `json:"-" meddler:"total_items"`
	FirstItem   uint64            `json:"-" meddler:"first_item"`
	LastItem    uint64            `json:"-" meddler:"last_item"`
}

// ExitAPI is a representation of a exit with additional information
// required by the API, and extracted by joining token table
type ExitAPI struct {
	ItemID                 uint64                          `meddler:"item_id"`
	BatchNum               common.BatchNum                 `meddler:"batch_num"`
	AccountIdx             apitypes.HezIdx                 `meddler:"account_idx"`
	MerkleProof            *merkletree.CircomVerifierProof `meddler:"merkle_proof,json"`
	Balance                apitypes.BigIntStr              `meddler:"balance"`
	InstantWithdrawn       *int64                          `meddler:"instant_withdrawn"`
	DelayedWithdrawRequest *int64                          `meddler:"delayed_withdraw_request"`
	DelayedWithdrawn       *int64                          `meddler:"delayed_withdrawn"`
	TotalItems             uint64                          `meddler:"total_items"`
	FirstItem              uint64                          `meddler:"first_item"`
	LastItem               uint64                          `meddler:"last_item"`
	TokenID                common.TokenID                  `meddler:"token_id"`
	TokenItemID            uint64                          `meddler:"token_item_id"`
	TokenEthBlockNum       int64                           `meddler:"token_block"`
	TokenEthAddr           ethCommon.Address               `meddler:"eth_addr"`
	TokenName              string                          `meddler:"name"`
	TokenSymbol            string                          `meddler:"symbol"`
	TokenDecimals          uint64                          `meddler:"decimals"`
	TokenUSD               *float64                        `meddler:"usd"`
	TokenUSDUpdate         *time.Time                      `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of ExitAPI
// without the need of auxiliar structs
func (e ExitAPI) MarshalJSON() ([]byte, error) {
	siblings := []string{}
	for i := 0; i < len(e.MerkleProof.Siblings); i++ {
		siblings = append(siblings, e.MerkleProof.Siblings[i].String())
	}
	return json.Marshal(map[string]interface{}{
		"itemId":       e.ItemID,
		"batchNum":     e.BatchNum,
		"accountIndex": e.AccountIdx,
		"merkleProof": map[string]interface{}{
			"Root":     e.MerkleProof.Root.String(),
			"Siblings": siblings,
			"OldKey":   e.MerkleProof.OldKey.String(),
			"OldValue": e.MerkleProof.OldValue.String(),
			"IsOld0":   e.MerkleProof.IsOld0,
			"Key":      e.MerkleProof.Key.String(),
			"Value":    e.MerkleProof.Value.String(),
			"Fnc":      e.MerkleProof.Fnc,
		},
		"balance":                e.Balance,
		"instantWithdrawn":       e.InstantWithdrawn,
		"delayedWithdrawRequest": e.DelayedWithdrawRequest,
		"delayedWithdrawn":       e.DelayedWithdrawn,
		"token": map[string]interface{}{
			"id":               e.TokenID,
			"itemId":           e.TokenItemID,
			"ethereumBlockNum": e.TokenEthBlockNum,
			"ethereumAddress":  e.TokenEthAddr,
			"name":             e.TokenName,
			"symbol":           e.TokenSymbol,
			"decimals":         e.TokenDecimals,
			"USD":              e.TokenUSD,
			"fiatUpdate":       e.TokenUSDUpdate,
		},
	})
}

// CoordinatorAPI is a representation of a coordinator with additional information
// required by the API
type CoordinatorAPI struct {
	ItemID      uint64            `json:"itemId" meddler:"item_id"`
	Bidder      ethCommon.Address `json:"bidderAddr" meddler:"bidder_addr"`
	Forger      ethCommon.Address `json:"forgerAddr" meddler:"forger_addr"`
	EthBlockNum int64             `json:"ethereumBlock" meddler:"eth_block_num"`
	URL         string            `json:"URL" meddler:"url"`
	TotalItems  uint64            `json:"-" meddler:"total_items"`
	FirstItem   uint64            `json:"-" meddler:"first_item"`
	LastItem    uint64            `json:"-" meddler:"last_item"`
}

// AccountAPI is a representation of a account with additional information
// required by the API
type AccountAPI struct {
	ItemID           uint64              `meddler:"item_id"`
	Idx              apitypes.HezIdx     `meddler:"idx"`
	BatchNum         common.BatchNum     `meddler:"batch_num"`
	PublicKey        apitypes.HezBJJ     `meddler:"bjj"`
	EthAddr          apitypes.HezEthAddr `meddler:"eth_addr"`
	Nonce            common.Nonce        `meddler:"-"` // max of 40 bits used
	Balance          *apitypes.BigIntStr `meddler:"-"` // max of 192 bits used
	TotalItems       uint64              `meddler:"total_items"`
	FirstItem        uint64              `meddler:"first_item"`
	LastItem         uint64              `meddler:"last_item"`
	TokenID          common.TokenID      `meddler:"token_id"`
	TokenItemID      int                 `meddler:"token_item_id"`
	TokenEthBlockNum int64               `meddler:"token_block"`
	TokenEthAddr     ethCommon.Address   `meddler:"token_eth_addr"`
	TokenName        string              `meddler:"name"`
	TokenSymbol      string              `meddler:"symbol"`
	TokenDecimals    uint64              `meddler:"decimals"`
	TokenUSD         *float64            `meddler:"usd"`
	TokenUSDUpdate   *time.Time          `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of AccountAPI
// without the need of auxiliar structs
func (account AccountAPI) MarshalJSON() ([]byte, error) {
	jsonAccount := map[string]interface{}{
		"itemId":             account.ItemID,
		"accountIndex":       account.Idx,
		"nonce":              account.Nonce,
		"balance":            account.Balance,
		"bjj":                account.PublicKey,
		"hezEthereumAddress": account.EthAddr,
		"token": map[string]interface{}{
			"id":               account.TokenID,
			"itemId":           account.TokenItemID,
			"ethereumBlockNum": account.TokenEthBlockNum,
			"ethereumAddress":  account.TokenEthAddr,
			"name":             account.TokenName,
			"symbol":           account.TokenSymbol,
			"decimals":         account.TokenDecimals,
			"USD":              account.TokenUSD,
			"fiatUpdate":       account.TokenUSDUpdate,
		},
	}
	return json.Marshal(jsonAccount)
}

// BatchAPI is a representation of a batch with additional information
// required by the API, and extracted by joining block table
type BatchAPI struct {
	ItemID        uint64                 `json:"itemId" meddler:"item_id"`
	BatchNum      common.BatchNum        `json:"batchNum" meddler:"batch_num"`
	EthBlockNum   int64                  `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthBlockHash  ethCommon.Hash         `json:"ethereumBlockHash" meddler:"hash"`
	Timestamp     time.Time              `json:"timestamp" meddler:"timestamp,utctime"`
	ForgerAddr    ethCommon.Address      `json:"forgerAddr" meddler:"forger_addr"`
	CollectedFees apitypes.CollectedFees `json:"collectedFees" meddler:"fees_collected,json"`
	TotalFeesUSD  *float64               `json:"historicTotalCollectedFeesUSD" meddler:"total_fees_usd"`
	StateRoot     apitypes.BigIntStr     `json:"stateRoot" meddler:"state_root"`
	NumAccounts   int                    `json:"numAccounts" meddler:"num_accounts"`
	ExitRoot      apitypes.BigIntStr     `json:"exitRoot" meddler:"exit_root"`
	ForgeL1TxsNum *int64                 `json:"forgeL1TransactionsNum" meddler:"forge_l1_txs_num"`
	SlotNum       int64                  `json:"slotNum" meddler:"slot_num"`
	TotalItems    uint64                 `json:"-" meddler:"total_items"`
	FirstItem     uint64                 `json:"-" meddler:"first_item"`
	LastItem      uint64                 `json:"-" meddler:"last_item"`
}

// Metrics define metrics of the network
type Metrics struct {
	TransactionsPerBatch  float64 `json:"transactionsPerBatch"`
	BatchFrequency        float64 `json:"batchFrequency"`
	TransactionsPerSecond float64 `json:"transactionsPerSecond"`
	TotalAccounts         int64   `json:"totalAccounts" meddler:"total_accounts"`
	TotalBJJs             int64   `json:"totalBJJs" meddler:"total_bjjs"`
	AvgTransactionFee     float64 `json:"avgTransactionFee"`
}

// MetricsTotals is used to get temporal information from HistoryDB
// to calculate data to be stored into the Metrics struct
type MetricsTotals struct {
	TotalTransactions uint64          `meddler:"total_txs"`
	FirstBatchNum     common.BatchNum `meddler:"batch_num"`
	TotalBatches      int64           `meddler:"total_batches"`
	TotalFeesUSD      float64         `meddler:"total_fees"`
}

// BidAPI is a representation of a bid with additional information
// required by the API
type BidAPI struct {
	ItemID      uint64             `json:"itemId" meddler:"item_id"`
	SlotNum     int64              `json:"slotNum" meddler:"slot_num"`
	BidValue    apitypes.BigIntStr `json:"bidValue" meddler:"bid_value"`
	EthBlockNum int64              `json:"ethereumBlockNum" meddler:"eth_block_num"`
	Bidder      ethCommon.Address  `json:"bidderAddr" meddler:"bidder_addr"`
	Forger      ethCommon.Address  `json:"forgerAddr" meddler:"forger_addr"`
	URL         string             `json:"URL" meddler:"url"`
	Timestamp   time.Time          `json:"timestamp" meddler:"timestamp,utctime"`
	TotalItems  uint64             `json:"-" meddler:"total_items"`
	FirstItem   uint64             `json:"-" meddler:"first_item"`
	LastItem    uint64             `json:"-" meddler:"last_item"`
}
