package historydb

import (
	"encoding/json"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/common/apitypes"
	"github.com/hermeznetwork/hermez-node/common/nonce"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

// TokenJSON is a representation of the JSON structure returned
// by the serialization method
type TokenJSON struct {
	TokenID          common.TokenID    `json:"id"`
	TokenItemID      uint64            `json:"itemId"`
	TokenEthBlockNum int64             `json:"ethereumBlockNum"`
	TokenEthAddr     ethCommon.Address `json:"ethereumAddress"`
	TokenName        string            `json:"name"`
	TokenSymbol      string            `json:"symbol"`
	TokenDecimals    uint64            `json:"decimals"`
	TokenUSD         *float64          `json:"USD"`
	TokenUSDUpdate   *time.Time        `json:"fiatUpdate"`
}

// ExitAPIJSON is a representation of the JSON structure returned
// by the serialization method
type ExitAPIJSON struct {
	ItemID                 uint64                          `json:"itemId"`
	BatchNum               common.BatchNum                 `json:"batchNum"`
	AccountIdx             apitypes.HezIdx                 `json:"accountIndex"`
	Bjj                    *apitypes.HezBJJ                `json:"bjj"`
	EthAddr                *apitypes.HezEthAddr            `json:"hezEthereumAddress"`
	MerkleProof            *merkletree.CircomVerifierProof `json:"merkleProof"`
	Balance                apitypes.BigIntStr              `json:"balance"`
	InstantWithdrawn       *int64                          `json:"instantWithdraw"`
	DelayedWithdrawRequest *int64                          `json:"delayedWithdrawRequest"`
	DelayedWithdrawn       *int64                          `json:"delayedWithdraw"`
	TokenJSON              TokenJSON                       `json:"token"`
}

// AccountAPIJSON is a representation of the JSON structure returned
// by the serialization method
type AccountAPIJSON struct {
	ItemID             uint64              `json:"itemId"`
	AccountIndex       apitypes.HezIdx     `json:"accountIndex"`
	Nonce              nonce.Nonce         `json:"nonce"`
	Balance            *apitypes.BigIntStr `json:"balance"`
	Bjj                apitypes.HezBJJ     `json:"bjj"`
	HezEthereumAddress apitypes.HezEthAddr `json:"hezEthereumAddress"`
	TokenJSON          TokenJSON           `json:"token"`
}

// TxAPIJSON is a representation of the JSON structure returned
// by the serialization method
type TxAPIJSON struct {
	TxID        common.TxID          `json:"id"`
	ItemID      uint64               `json:"itemId"`
	Type        common.TxType        `json:"type"`
	Position    int                  `json:"position"`
	FromIdx     *apitypes.HezIdx     `json:"fromAccountIndex"`
	FromEthAddr *apitypes.HezEthAddr `json:"fromHezEthereumAddress"`
	FromBJJ     *apitypes.HezBJJ     `json:"fromBJJ"`
	ToIdx       apitypes.HezIdx      `json:"toAccountIndex"`
	ToEthAddr   *apitypes.HezEthAddr `json:"toHezEthereumAddress"`
	ToBJJ       *apitypes.HezBJJ     `json:"toBJJ"`
	Amount      apitypes.BigIntStr   `json:"amount"`
	BatchNum    *common.BatchNum     `json:"batchNum"`
	HistoricUSD *float64             `json:"historicUSD"`
	Timestamp   time.Time            `json:"timestamp"`
	L1Info      *L1infoJSON          `json:"L1Info"`
	L2Info      *L2infoJSON          `json:"L2Info"`
	TokenJSON   TokenJSON            `json:"token"`
	L1orL2      string               `json:"L1orL2"`
}

// L1infoJSON is a representation of the JSON structure returned
// by the serialization method
type L1infoJSON struct {
	ToForgeL1TxsNum          *int64              `json:"toForgeL1TransactionsNum"`
	UserOrigin               *bool               `json:"userOrigin"`
	DepositAmount            *apitypes.BigIntStr `json:"depositAmount"`
	AmountSuccess            bool                `json:"amountSuccess"`
	DepositAmountSuccess     bool                `json:"depositAmountSuccess"`
	HistoricDepositAmountUSD *float64            `json:"historicDepositAmountUSD"`
	EthereumBlockNum         int64               `json:"ethereumBlockNum"`
	EthereumTxHash           *ethCommon.Hash     `json:"ethereumTxHash"`
	L1Fee                    *apitypes.BigIntStr `json:"l1Fee"`
}

// L2infoJSON is a representation of the JSON structure returned
// by the serialization method
type L2infoJSON struct {
	Fee            *common.FeeSelector `json:"fee"`
	HistoricFeeUSD *float64            `json:"historicFeeUSD"`
	Nonce          *nonce.Nonce        `json:"nonce"`
}

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
	EthereumTxHash           ethCommon.Hash      `meddler:"eth_tx_hash,zeroisnull"`
	L1Fee                    *apitypes.BigIntStr `meddler:"l1_fee"`
	// L2
	Fee            *common.FeeSelector `meddler:"fee"`
	HistoricFeeUSD *float64            `meddler:"fee_usd"`
	Nonce          *nonce.Nonce        `meddler:"nonce"`
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
	txa := TxAPIJSON{
		TxID:        tx.TxID,
		ItemID:      tx.ItemID,
		Type:        tx.Type,
		Position:    tx.Position,
		FromIdx:     tx.FromIdx,
		FromEthAddr: tx.FromEthAddr,
		FromBJJ:     tx.FromBJJ,
		ToIdx:       tx.ToIdx,
		ToEthAddr:   tx.ToEthAddr,
		ToBJJ:       tx.ToBJJ,
		Amount:      tx.Amount,
		BatchNum:    tx.BatchNum,
		HistoricUSD: tx.HistoricUSD,
		Timestamp:   tx.Timestamp,
		L1Info:      nil,
		L2Info:      nil,
		TokenJSON: TokenJSON{
			TokenID:          tx.TokenID,
			TokenItemID:      tx.TokenItemID,
			TokenEthBlockNum: tx.TokenEthBlockNum,
			TokenEthAddr:     tx.TokenEthAddr,
			TokenName:        tx.TokenName,
			TokenSymbol:      tx.TokenSymbol,
			TokenDecimals:    tx.TokenDecimals,
			TokenUSD:         tx.TokenUSD,
			TokenUSDUpdate:   tx.TokenUSDUpdate,
		},
	}

	if tx.IsL1 {
		txa.L1orL2 = "L1"
		amountSuccess := tx.AmountSuccess
		depositAmountSuccess := tx.DepositAmountSuccess
		if tx.BatchNum == nil {
			amountSuccess = false
			depositAmountSuccess = false
		}
		txa.L1Info = &L1infoJSON{
			ToForgeL1TxsNum:          tx.ToForgeL1TxsNum,
			UserOrigin:               tx.UserOrigin,
			DepositAmount:            tx.DepositAmount,
			AmountSuccess:            amountSuccess,
			DepositAmountSuccess:     depositAmountSuccess,
			HistoricDepositAmountUSD: tx.HistoricDepositAmountUSD,
			EthereumBlockNum:         tx.EthBlockNum,
			EthereumTxHash:           &tx.EthereumTxHash,
			L1Fee:                    tx.L1Fee,
		}
	} else {
		txa.L1orL2 = "L2"
		txa.L2Info = &L2infoJSON{
			Fee:            tx.Fee,
			HistoricFeeUSD: tx.HistoricFeeUSD,
			Nonce:          tx.Nonce,
		}
	}
	return json.Marshal(txa)
}

// txWrite is an representatiion that merges common.L1Tx and common.L2Tx
// in order to perform inserts into tx table
// EffectiveAmount and EffectiveDepositAmount are not set since they have default values in the DB
type txWrite struct {
	// Generic
	IsL1             bool             `meddler:"is_l1"`
	TxID             common.TxID      `meddler:"id"`
	Type             common.TxType    `meddler:"type"`
	Position         int              `meddler:"position"`
	FromIdx          *common.Idx      `meddler:"from_idx"`
	EffectiveFromIdx *common.Idx      `meddler:"effective_from_idx"`
	ToIdx            common.Idx       `meddler:"to_idx"`
	Amount           *big.Int         `meddler:"amount,bigint"`
	AmountFloat      float64          `meddler:"amount_f"`
	TokenID          common.TokenID   `meddler:"token_id"`
	BatchNum         *common.BatchNum `meddler:"batch_num"`     // batchNum in which this tx was forged. If the tx is L2, this must be != 0
	EthBlockNum      int64            `meddler:"eth_block_num"` // Ethereum Block Number in which this L1Tx was added to the queue
	// L1
	ToForgeL1TxsNum    *int64                 `meddler:"to_forge_l1_txs_num"` // toForgeL1TxsNum in which the tx was forged / will be forged
	UserOrigin         *bool                  `meddler:"user_origin"`         // true if the tx was originated by a user, false if it was aoriginated by a coordinator. Note that this differ from the spec for implementation simplification purpposes
	FromEthAddr        *ethCommon.Address     `meddler:"from_eth_addr"`
	FromBJJ            *babyjub.PublicKeyComp `meddler:"from_bjj"`
	DepositAmount      *big.Int               `meddler:"deposit_amount,bigintnull"`
	DepositAmountFloat *float64               `meddler:"deposit_amount_f"`
	EthTxHash          *ethCommon.Hash        `meddler:"eth_tx_hash"`
	L1Fee              *big.Int               `meddler:"l1_fee,bigintnull"`
	// L2
	Fee   *common.FeeSelector `meddler:"fee"`
	Nonce *nonce.Nonce        `meddler:"nonce"`
}

// TokenSymbolAndAddr token representation with only Eth addr and symbol
type TokenSymbolAndAddr struct {
	TokenID uint              `meddler:"token_id"`
	Symbol  string            `meddler:"symbol"`
	Addr    ethCommon.Address `meddler:"eth_addr"`
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
	EthAddr                *apitypes.HezEthAddr            `meddler:"eth_addr"`
	BJJ                    *apitypes.HezBJJ                `meddler:"bjj"`
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
	TokenEthAddr           ethCommon.Address               `meddler:"token_eth_addr"`
	TokenName              string                          `meddler:"name"`
	TokenSymbol            string                          `meddler:"symbol"`
	TokenDecimals          uint64                          `meddler:"decimals"`
	TokenUSD               *float64                        `meddler:"usd"`
	TokenUSDUpdate         *time.Time                      `meddler:"usd_update"`
}

// MarshalJSON is used to neast some of the fields of ExitAPI
// without the need of auxiliar structs
func (e ExitAPI) MarshalJSON() ([]byte, error) {
	eaj := ExitAPIJSON{
		ItemID:                 e.ItemID,
		BatchNum:               e.BatchNum,
		AccountIdx:             e.AccountIdx,
		Bjj:                    e.BJJ,
		EthAddr:                e.EthAddr,
		MerkleProof:            e.MerkleProof,
		Balance:                e.Balance,
		InstantWithdrawn:       e.InstantWithdrawn,
		DelayedWithdrawRequest: e.DelayedWithdrawRequest,
		DelayedWithdrawn:       e.DelayedWithdrawn,
		TokenJSON: TokenJSON{
			TokenID:          e.TokenID,
			TokenItemID:      e.TokenItemID,
			TokenEthBlockNum: e.TokenEthBlockNum,
			TokenEthAddr:     e.TokenEthAddr,
			TokenName:        e.TokenName,
			TokenSymbol:      e.TokenSymbol,
			TokenDecimals:    e.TokenDecimals,
			TokenUSD:         e.TokenUSD,
			TokenUSDUpdate:   e.TokenUSDUpdate,
		},
	}

	return json.Marshal(eaj)
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

// FiatCurrency is a representation of a currency price object
type FiatCurrency struct {
	Currency     string    `json:"currency" meddler:"currency"`
	BaseCurrency string    `json:"baseCurrency" meddler:"base_currency"`
	Price        float64   `json:"price" meddler:"price"`
	LastUpdate   time.Time `json:"lastUpdate" meddler:"last_update"`
}

// AccountAPI is a representation of a account with additional information
// required by the API
type AccountAPI struct {
	ItemID           uint64              `meddler:"item_id"`
	Idx              apitypes.HezIdx     `meddler:"idx"`
	BatchNum         common.BatchNum     `meddler:"batch_num"`
	PublicKey        apitypes.HezBJJ     `meddler:"bjj"`
	EthAddr          apitypes.HezEthAddr `meddler:"eth_addr"`
	Nonce            nonce.Nonce         `meddler:"nonce"`   // max of 40 bits used
	Balance          *apitypes.BigIntStr `meddler:"balance"` // max of 192 bits used
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
	act := AccountAPIJSON{
		ItemID:             account.ItemID,
		AccountIndex:       account.Idx,
		Nonce:              account.Nonce,
		Balance:            account.Balance,
		Bjj:                account.PublicKey,
		HezEthereumAddress: account.EthAddr,
		TokenJSON: TokenJSON{
			TokenID:          account.TokenID,
			TokenItemID:      uint64(account.TokenItemID),
			TokenEthBlockNum: account.TokenEthBlockNum,
			TokenEthAddr:     account.TokenEthAddr,
			TokenName:        account.TokenName,
			TokenSymbol:      account.TokenSymbol,
			TokenDecimals:    account.TokenDecimals,
			TokenUSD:         account.TokenUSD,
			TokenUSDUpdate:   account.TokenUSDUpdate,
		},
	}
	return json.Marshal(act)
}

// BatchAPI is a representation of a batch with additional information
// required by the API, and extracted by joining block table
type BatchAPI struct {
	ItemID           uint64                      `json:"itemId" meddler:"item_id"`
	BatchNum         common.BatchNum             `json:"batchNum" meddler:"batch_num"`
	EthereumTxHash   ethCommon.Hash              `json:"ethereumTxHash" meddler:"eth_tx_hash"`
	EthBlockNum      int64                       `json:"ethereumBlockNum" meddler:"eth_block_num"`
	EthBlockHash     ethCommon.Hash              `json:"ethereumBlockHash" meddler:"hash"`
	Timestamp        time.Time                   `json:"timestamp" meddler:"timestamp,utctime"`
	ForgerAddr       ethCommon.Address           `json:"forgerAddr" meddler:"forger_addr"`
	CollectedFeesDB  map[common.TokenID]*big.Int `json:"-" meddler:"fees_collected,json"`
	CollectedFeesAPI apitypes.CollectedFeesAPI   `json:"collectedFees" meddler:"-"`
	TotalFeesUSD     *float64                    `json:"historicTotalCollectedFeesUSD" meddler:"total_fees_usd"`
	StateRoot        apitypes.BigIntStr          `json:"stateRoot" meddler:"state_root"`
	NumAccounts      int                         `json:"numAccounts" meddler:"num_accounts"`
	ExitRoot         apitypes.BigIntStr          `json:"exitRoot" meddler:"exit_root"`
	ForgeL1TxsNum    *int64                      `json:"forgeL1TransactionsNum" meddler:"forge_l1_txs_num"`
	SlotNum          int64                       `json:"slotNum" meddler:"slot_num"`
	ForgedTxs        int                         `json:"forgedTransactions" meddler:"forged_txs"`
	TotalItems       uint64                      `json:"-" meddler:"total_items"`
	FirstItem        uint64                      `json:"-" meddler:"first_item"`
	LastItem         uint64                      `json:"-" meddler:"last_item"`
}

// MetricsAPI define metrics of the network
type MetricsAPI struct {
	TransactionsPerBatch   float64 `json:"transactionsPerBatch"`
	BatchFrequency         float64 `json:"batchFrequency"`
	TransactionsPerSecond  float64 `json:"transactionsPerSecond"`
	TokenAccounts          int64   `json:"tokenAccounts"`
	Wallets                int64   `json:"wallets"`
	AvgTransactionFee      float64 `json:"avgTransactionFee"`
	EstimatedTimeToForgeL1 float64 `json:"estimatedTimeToForgeL1" meddler:"estimated_time_to_forge_l1"`
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

// MinBidInfo gives information of the minum bid for specific slot(s)
type MinBidInfo struct {
	DefaultSlotSetBid        [6]*big.Int `json:"defaultSlotSetBid" meddler:"default_slot_set_bid,json" validate:"required"`
	DefaultSlotSetBidSlotNum int64       `json:"-" meddler:"default_slot_set_bid_slot_num"`
}

// BucketUpdateAPI are the bucket updates (tracking the withdrawals value changes)
// in Rollup Smart Contract
type BucketUpdateAPI struct {
	EthBlockNum int64               `json:"ethereumBlockNum" meddler:"eth_block_num"`
	NumBucket   int                 `json:"numBucket" meddler:"num_bucket"`
	BlockStamp  int64               `json:"blockStamp" meddler:"block_stamp"`
	Withdrawals *apitypes.BigIntStr `json:"withdrawals" meddler:"withdrawals"`
}

// BucketParamsAPI are the parameter variables of each Bucket of Rollup Smart
// Contract
type BucketParamsAPI struct {
	CeilUSD         *apitypes.BigIntStr `json:"ceilUSD"`
	BlockStamp      *apitypes.BigIntStr `json:"blockStamp"`
	Withdrawals     *apitypes.BigIntStr `json:"withdrawals"`
	RateBlocks      *apitypes.BigIntStr `json:"rateBlocks"`
	RateWithdrawals *apitypes.BigIntStr `json:"rateWithdrawals"`
	MaxWithdrawals  *apitypes.BigIntStr `json:"maxWithdrawals"`
}

// RollupVariablesAPI are the variables of the Rollup Smart Contract
type RollupVariablesAPI struct {
	EthBlockNum           int64               `json:"ethereumBlockNum" meddler:"eth_block_num"`
	FeeAddToken           *apitypes.BigIntStr `json:"feeAddToken" meddler:"fee_add_token" validate:"required"`
	ForgeL1L2BatchTimeout int64               `json:"forgeL1L2BatchTimeout" meddler:"forge_l1_timeout" validate:"required"`
	WithdrawalDelay       uint64              `json:"withdrawalDelay" meddler:"withdrawal_delay" validate:"required"`
	Buckets               []BucketParamsAPI   `json:"buckets" meddler:"buckets,json"`
	SafeMode              bool                `json:"safeMode" meddler:"safe_mode"`
}

// NewRollupVariablesAPI creates a RollupVariablesAPI from common.RollupVariables
func NewRollupVariablesAPI(rollupVariables *common.RollupVariables) *RollupVariablesAPI {
	buckets := make([]BucketParamsAPI, len(rollupVariables.Buckets))
	rollupVars := RollupVariablesAPI{
		EthBlockNum:           rollupVariables.EthBlockNum,
		FeeAddToken:           apitypes.NewBigIntStr(rollupVariables.FeeAddToken),
		ForgeL1L2BatchTimeout: rollupVariables.ForgeL1L2BatchTimeout,
		WithdrawalDelay:       rollupVariables.WithdrawalDelay,
		SafeMode:              rollupVariables.SafeMode,
		Buckets:               buckets,
	}
	for i, bucket := range rollupVariables.Buckets {
		rollupVars.Buckets[i] = BucketParamsAPI{
			CeilUSD:         apitypes.NewBigIntStr(bucket.CeilUSD),
			BlockStamp:      apitypes.NewBigIntStr(bucket.BlockStamp),
			Withdrawals:     apitypes.NewBigIntStr(bucket.Withdrawals),
			RateBlocks:      apitypes.NewBigIntStr(bucket.RateBlocks),
			RateWithdrawals: apitypes.NewBigIntStr(bucket.RateWithdrawals),
			MaxWithdrawals:  apitypes.NewBigIntStr(bucket.MaxWithdrawals),
		}
	}
	return &rollupVars
}

// AuctionVariablesAPI are the variables of the Auction Smart Contract
type AuctionVariablesAPI struct {
	EthBlockNum int64 `json:"ethereumBlockNum" meddler:"eth_block_num"`
	// DonationAddress Address where the donations will be sent
	DonationAddress ethCommon.Address `json:"donationAddress" meddler:"donation_address" validate:"required"`
	// BootCoordinator Address of the boot coordinator
	BootCoordinator ethCommon.Address `json:"bootCoordinator" meddler:"boot_coordinator" validate:"required"`
	// BootCoordinatorURL URL of the boot coordinator
	BootCoordinatorURL string `json:"bootCoordinatorUrl" meddler:"boot_coordinator_url" validate:"required"`
	// DefaultSlotSetBid The minimum bid value in a series of 6 slots
	DefaultSlotSetBid [6]*apitypes.BigIntStr `json:"defaultSlotSetBid" meddler:"default_slot_set_bid,json" validate:"required"`
	// DefaultSlotSetBidSlotNum SlotNum at which the new default_slot_set_bid applies
	DefaultSlotSetBidSlotNum int64 `json:"defaultSlotSetBidSlotNum" meddler:"default_slot_set_bid_slot_num"`
	// ClosedAuctionSlots Distance (#slots) to the closest slot to which you can bid ( 2 Slots = 2 * 40 Blocks = 20 min )
	ClosedAuctionSlots uint16 `json:"closedAuctionSlots" meddler:"closed_auction_slots" validate:"required"`
	// OpenAuctionSlots Distance (#slots) to the farthest slot to which you can bid (30 days = 4320 slots )
	OpenAuctionSlots uint16 `json:"openAuctionSlots" meddler:"open_auction_slots" validate:"required"`
	// AllocationRatio How the HEZ tokens deposited by the slot winner are distributed (Burn: 40% - Donation: 40% - HGT: 20%)
	AllocationRatio [3]uint16 `json:"allocationRatio" meddler:"allocation_ratio,json" validate:"required"`
	// Outbidding Minimum outbid (percentage) over the previous one to consider it valid
	Outbidding uint16 `json:"outbidding" meddler:"outbidding" validate:"required"`
	// SlotDeadline Number of blocks at the end of a slot in which any coordinator can forge if the winner has not forged one before
	SlotDeadline uint8 `json:"slotDeadline" meddler:"slot_deadline" validate:"required"`
}

// NewAuctionVariablesAPI creates a AuctionVariablesAPI from common.AuctionVariables
func NewAuctionVariablesAPI(auctionVariables *common.AuctionVariables) *AuctionVariablesAPI {
	auctionVars := AuctionVariablesAPI{
		EthBlockNum:              auctionVariables.EthBlockNum,
		DonationAddress:          auctionVariables.DonationAddress,
		BootCoordinator:          auctionVariables.BootCoordinator,
		BootCoordinatorURL:       auctionVariables.BootCoordinatorURL,
		DefaultSlotSetBidSlotNum: auctionVariables.DefaultSlotSetBidSlotNum,
		ClosedAuctionSlots:       auctionVariables.ClosedAuctionSlots,
		OpenAuctionSlots:         auctionVariables.OpenAuctionSlots,
		Outbidding:               auctionVariables.Outbidding,
		SlotDeadline:             auctionVariables.SlotDeadline,
	}

	for i, slot := range auctionVariables.DefaultSlotSetBid {
		auctionVars.DefaultSlotSetBid[i] = apitypes.NewBigIntStr(slot)
	}

	for i, ratio := range auctionVariables.AllocationRatio {
		auctionVars.AllocationRatio[i] = ratio
	}

	return &auctionVars
}
