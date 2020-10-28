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
	ItemID      int                  `meddler:"item_id"`
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
	ToForgeL1TxsNum       *int64              `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin            *bool               `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	LoadAmount            *apitypes.BigIntStr `meddler:"load_amount"`
	HistoricLoadAmountUSD *float64            `meddler:"load_amount_usd"`
	// L2
	Fee            *common.FeeSelector `meddler:"fee"`
	HistoricFeeUSD *float64            `meddler:"fee_usd"`
	Nonce          *common.Nonce       `meddler:"nonce"`
	// API extras
	Timestamp        time.Time         `meddler:"timestamp,utctime"`
	TotalItems       int               `meddler:"total_items"`
	FirstItem        int               `meddler:"first_item"`
	LastItem         int               `meddler:"last_item"`
	TokenID          common.TokenID    `meddler:"token_id"`
	TokenItemID      int               `meddler:"token_item_id"`
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
		jsonTx["L1Info"] = map[string]interface{}{
			"toForgeL1TransactionsNum": tx.ToForgeL1TxsNum,
			"userOrigin":               tx.UserOrigin,
			"loadAmount":               tx.LoadAmount,
			"historicLoadAmountUSD":    tx.HistoricLoadAmountUSD,
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
	ToForgeL1TxsNum *int64             `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin      *bool              `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr     *ethCommon.Address `meddler:"from_eth_addr"`
	FromBJJ         *babyjub.PublicKey `meddler:"from_bjj"`
	LoadAmount      *big.Int           `meddler:"load_amount,bigintnull"`
	LoadAmountFloat *float64           `meddler:"load_amount_f"`
	// L2
	Fee   *common.FeeSelector `meddler:"fee"`
	Nonce *common.Nonce       `meddler:"nonce"`
}

// TokenWithUSD add USD info to common.Token
type TokenWithUSD struct {
	ItemID      int               `json:"itemId" meddler:"item_id"`
	TokenID     common.TokenID    `json:"id" meddler:"token_id"`
	EthBlockNum int64             `json:"ethereumBlockNum" meddler:"eth_block_num"` // Ethereum block number in which this token was registered
	EthAddr     ethCommon.Address `json:"ethereumAddress" meddler:"eth_addr"`
	Name        string            `json:"name" meddler:"name"`
	Symbol      string            `json:"symbol" meddler:"symbol"`
	Decimals    uint64            `json:"decimals" meddler:"decimals"`
	USD         *float64          `json:"USD" meddler:"usd"`
	USDUpdate   *time.Time        `json:"fiatUpdate" meddler:"usd_update,utctime"`
	TotalItems  int               `json:"-" meddler:"total_items"`
	FirstItem   int               `json:"-" meddler:"first_item"`
	LastItem    int               `json:"-" meddler:"last_item"`
}

// HistoryExit is a representation of a exit with additional information
// required by the API, and extracted by joining token table
type HistoryExit struct {
	ItemID                 int                             `meddler:"item_id"`
	BatchNum               common.BatchNum                 `meddler:"batch_num"`
	AccountIdx             common.Idx                      `meddler:"account_idx"`
	MerkleProof            *merkletree.CircomVerifierProof `meddler:"merkle_proof,json"`
	Balance                *big.Int                        `meddler:"balance,bigint"`
	InstantWithdrawn       *int64                          `meddler:"instant_withdrawn"`
	DelayedWithdrawRequest *int64                          `meddler:"delayed_withdraw_request"`
	DelayedWithdrawn       *int64                          `meddler:"delayed_withdrawn"`
	TotalItems             int                             `meddler:"total_items"`
	FirstItem              int                             `meddler:"first_item"`
	LastItem               int                             `meddler:"last_item"`
	TokenID                common.TokenID                  `meddler:"token_id"`
	TokenEthBlockNum       int64                           `meddler:"token_block"`
	TokenEthAddr           ethCommon.Address               `meddler:"eth_addr"`
	TokenName              string                          `meddler:"name"`
	TokenSymbol            string                          `meddler:"symbol"`
	TokenDecimals          uint64                          `meddler:"decimals"`
	TokenUSD               *float64                        `meddler:"usd"`
	TokenUSDUpdate         *time.Time                      `meddler:"usd_update"`
}

// CoordinatorAPI is a representation of a coordinator with additional information
// required by the API
type CoordinatorAPI struct {
	ItemID      int               `json:"itemId" meddler:"item_id"`
	Bidder      ethCommon.Address `json:"bidderAddr" meddler:"bidder_addr"`
	Forger      ethCommon.Address `json:"forgerAddr" meddler:"forger_addr"`
	EthBlockNum int64             `json:"ethereumBlock" meddler:"eth_block_num"`
	URL         string            `json:"URL" meddler:"url"`
	TotalItems  int               `json:"-" meddler:"total_items"`
	FirstItem   int               `json:"-" meddler:"first_item"`
	LastItem    int               `json:"-" meddler:"last_item"`
}

// BatchAPI is a representation of a batch with additional information
// required by the API, and extracted by joining block table
type BatchAPI struct {
	ItemID        int                    `json:"itemId" meddler:"item_id"`
	BatchNum      common.BatchNum        `json:"batchNum" meddler:"batch_num"`
	EthBlockNum   int64                  `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthBlockHash  ethCommon.Hash         `json:"ethereumBlockHash" meddler:"hash"`
	Timestamp     time.Time              `json:"timestamp" meddler:"timestamp,utctime"`
	ForgerAddr    ethCommon.Address      `json:"forgerAddr" meddler:"forger_addr"`
	CollectedFees apitypes.CollectedFees `json:"collectedFees" meddler:"fees_collected,json"`
	// CollectedFees map[common.TokenID]*big.Int `json:"collectedFees" meddler:"fees_collected,json"`
	TotalFeesUSD  *float64           `json:"historicTotalCollectedFeesUSD" meddler:"total_fees_usd"`
	StateRoot     apitypes.BigIntStr `json:"stateRoot" meddler:"state_root"`
	NumAccounts   int                `json:"numAccounts" meddler:"num_accounts"`
	ExitRoot      apitypes.BigIntStr `json:"exitRoot" meddler:"exit_root"`
	ForgeL1TxsNum *int64             `json:"forgeL1TransactionsNum" meddler:"forge_l1_txs_num"`
	SlotNum       int64              `json:"slotNum" meddler:"slot_num"`
	TotalItems    int                `json:"-" meddler:"total_items"`
	FirstItem     int                `json:"-" meddler:"first_item"`
	LastItem      int                `json:"-" meddler:"last_item"`
}

// Network define status of the network
type Network struct {
	LastBlock   int64            `json:"lastBlock"`
	LastBatch   BatchAPI         `json:"lastBatch"`
	CurrentSlot int64            `json:"currentSlot"`
	NextForgers []CoordinatorAPI `json:"nextForgers"`
}

// Metrics define metrics of the network
type Metrics struct {
	TransactionsPerBatch  float64 `json:"transactionsPerBatch"`
	BatchFrequency        float64 `json:"batchFrequency"`
	TransactionsPerSecond float64 `json:"transactionsPerSecond"`
	TotalAccounts         int64   `json:"totalAccounts"`
	TotalBJJs             int64   `json:"totalBJJs"`
	AvgTransactionFee     float64 `json:"avgTransactionFee"`
}

// BidAPI is a representation of a bid with additional information
// required by the API
type BidAPI struct {
	ItemID      int                `json:"itemId" meddler:"item_id"`
	SlotNum     int64              `json:"slotNum" meddler:"slot_num"`
	BidValue    apitypes.BigIntStr `json:"bidValue" meddler:"bid_value"`
	EthBlockNum int64              `json:"ethereumBlockNum" meddler:"eth_block_num"`
	Bidder      ethCommon.Address  `json:"bidderAddr" meddler:"bidder_addr"`
	Forger      ethCommon.Address  `json:"forgerAddr" meddler:"forger_addr"`
	URL         string             `json:"URL" meddler:"url"`
	Timestamp   time.Time          `json:"timestamp" meddler:"timestamp,utctime"`
	TotalItems  int                `json:"-" meddler:"total_items"`
	FirstItem   int                `json:"-" meddler:"first_item"`
	LastItem    int                `json:"-" meddler:"last_item"`
}
