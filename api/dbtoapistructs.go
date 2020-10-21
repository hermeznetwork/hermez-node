package api

import (
	"encoding/base64"
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
)

type errorMsg struct {
	Message string
}

func bjjToString(bjj *babyjub.PublicKey) string {
	pkComp := [32]byte(bjj.Compress())
	sum := pkComp[0]
	for i := 1; i < len(pkComp); i++ {
		sum += pkComp[i]
	}
	bjjSum := append(pkComp[:], sum)
	return "hez:" + base64.RawURLEncoding.EncodeToString(bjjSum)
}

func ethAddrToHez(addr ethCommon.Address) string {
	return "hez:" + addr.String()
}

func idxToHez(idx common.Idx, tokenSymbol string) string {
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
}

// History Tx

type historyTxsAPI struct {
	Txs        []historyTxAPI `json:"transactions"`
	Pagination *db.Pagination `json:"pagination"`
}

func (htx *historyTxsAPI) GetPagination() *db.Pagination {
	if htx.Txs[0].ItemID < htx.Txs[len(htx.Txs)-1].ItemID {
		htx.Pagination.FirstReturnedItem = htx.Txs[0].ItemID
		htx.Pagination.LastReturnedItem = htx.Txs[len(htx.Txs)-1].ItemID
	} else {
		htx.Pagination.LastReturnedItem = htx.Txs[0].ItemID
		htx.Pagination.FirstReturnedItem = htx.Txs[len(htx.Txs)-1].ItemID
	}
	return htx.Pagination
}
func (htx *historyTxsAPI) Len() int { return len(htx.Txs) }

type l1Info struct {
	ToForgeL1TxsNum       *int64   `json:"toForgeL1TransactionsNum"`
	UserOrigin            bool     `json:"userOrigin"`
	FromEthAddr           string   `json:"fromHezEthereumAddress"`
	FromBJJ               string   `json:"fromBJJ"`
	LoadAmount            string   `json:"loadAmount"`
	HistoricLoadAmountUSD *float64 `json:"historicLoadAmountUSD"`
	EthBlockNum           int64    `json:"ethereumBlockNum"`
}

type l2Info struct {
	Fee            common.FeeSelector `json:"fee"`
	HistoricFeeUSD *float64           `json:"historicFeeUSD"`
	Nonce          common.Nonce       `json:"nonce"`
}

type historyTxAPI struct {
	IsL1        string           `json:"L1orL2"`
	TxID        string           `json:"id"`
	ItemID      int              `json:"itemId"`
	Type        common.TxType    `json:"type"`
	Position    int              `json:"position"`
	FromIdx     *string          `json:"fromAccountIndex"`
	ToIdx       string           `json:"toAccountIndex"`
	Amount      string           `json:"amount"`
	BatchNum    *common.BatchNum `json:"batchNum"`
	HistoricUSD *float64         `json:"historicUSD"`
	Timestamp   time.Time        `json:"timestamp"`
	L1Info      *l1Info          `json:"L1Info"`
	L2Info      *l2Info          `json:"L2Info"`
	Token       tokenAPI         `json:"token"`
}

func historyTxsToAPI(dbTxs []historydb.HistoryTx) []historyTxAPI {
	apiTxs := []historyTxAPI{}
	for i := 0; i < len(dbTxs); i++ {
		apiTx := historyTxAPI{
			TxID:        dbTxs[i].TxID.String(),
			ItemID:      dbTxs[i].ItemID,
			Type:        dbTxs[i].Type,
			Position:    dbTxs[i].Position,
			ToIdx:       idxToHez(dbTxs[i].ToIdx, dbTxs[i].TokenSymbol),
			Amount:      dbTxs[i].Amount.String(),
			HistoricUSD: dbTxs[i].HistoricUSD,
			BatchNum:    dbTxs[i].BatchNum,
			Timestamp:   dbTxs[i].Timestamp,
			Token: tokenAPI{
				TokenID:     dbTxs[i].TokenID,
				EthBlockNum: dbTxs[i].TokenEthBlockNum,
				EthAddr:     dbTxs[i].TokenEthAddr,
				Name:        dbTxs[i].TokenName,
				Symbol:      dbTxs[i].TokenSymbol,
				Decimals:    dbTxs[i].TokenDecimals,
				USD:         dbTxs[i].TokenUSD,
				USDUpdate:   dbTxs[i].TokenUSDUpdate,
			},
			L1Info: nil,
			L2Info: nil,
		}
		if dbTxs[i].FromIdx != nil {
			fromIdx := new(string)
			*fromIdx = idxToHez(*dbTxs[i].FromIdx, dbTxs[i].TokenSymbol)
			apiTx.FromIdx = fromIdx
		}
		if dbTxs[i].IsL1 {
			apiTx.IsL1 = "L1"
			apiTx.L1Info = &l1Info{
				ToForgeL1TxsNum:       dbTxs[i].ToForgeL1TxsNum,
				UserOrigin:            *dbTxs[i].UserOrigin,
				FromEthAddr:           ethAddrToHez(*dbTxs[i].FromEthAddr),
				FromBJJ:               bjjToString(dbTxs[i].FromBJJ),
				LoadAmount:            dbTxs[i].LoadAmount.String(),
				HistoricLoadAmountUSD: dbTxs[i].HistoricLoadAmountUSD,
				EthBlockNum:           dbTxs[i].EthBlockNum,
			}
		} else {
			apiTx.IsL1 = "L2"
			apiTx.L2Info = &l2Info{
				Fee:            *dbTxs[i].Fee,
				HistoricFeeUSD: dbTxs[i].HistoricFeeUSD,
				Nonce:          *dbTxs[i].Nonce,
			}
		}
		apiTxs = append(apiTxs, apiTx)
	}
	return apiTxs
}

// Exit

type exitsAPI struct {
	Exits      []exitAPI      `json:"exits"`
	Pagination *db.Pagination `json:"pagination"`
}

func (e *exitsAPI) GetPagination() *db.Pagination {
	if e.Exits[0].ItemID < e.Exits[len(e.Exits)-1].ItemID {
		e.Pagination.FirstReturnedItem = e.Exits[0].ItemID
		e.Pagination.LastReturnedItem = e.Exits[len(e.Exits)-1].ItemID
	} else {
		e.Pagination.LastReturnedItem = e.Exits[0].ItemID
		e.Pagination.FirstReturnedItem = e.Exits[len(e.Exits)-1].ItemID
	}
	return e.Pagination
}
func (e *exitsAPI) Len() int { return len(e.Exits) }

type exitAPI struct {
	ItemID                 int                             `json:"itemId"`
	BatchNum               common.BatchNum                 `json:"batchNum"`
	AccountIdx             string                          `json:"accountIndex"`
	MerkleProof            *merkletree.CircomVerifierProof `json:"merkleProof"`
	Balance                string                          `json:"balance"`
	InstantWithdrawn       *int64                          `json:"instantWithdrawn"`
	DelayedWithdrawRequest *int64                          `json:"delayedWithdrawRequest"`
	DelayedWithdrawn       *int64                          `json:"delayedWithdrawn"`
	Token                  tokenAPI                        `json:"token"`
}

func historyExitsToAPI(dbExits []historydb.HistoryExit) []exitAPI {
	apiExits := []exitAPI{}
	for i := 0; i < len(dbExits); i++ {
		apiExits = append(apiExits, exitAPI{
			ItemID:                 dbExits[i].ItemID,
			BatchNum:               dbExits[i].BatchNum,
			AccountIdx:             idxToHez(dbExits[i].AccountIdx, dbExits[i].TokenSymbol),
			MerkleProof:            dbExits[i].MerkleProof,
			Balance:                dbExits[i].Balance.String(),
			InstantWithdrawn:       dbExits[i].InstantWithdrawn,
			DelayedWithdrawRequest: dbExits[i].DelayedWithdrawRequest,
			DelayedWithdrawn:       dbExits[i].DelayedWithdrawn,
			Token: tokenAPI{
				TokenID:     dbExits[i].TokenID,
				EthBlockNum: dbExits[i].TokenEthBlockNum,
				EthAddr:     dbExits[i].TokenEthAddr,
				Name:        dbExits[i].TokenName,
				Symbol:      dbExits[i].TokenSymbol,
				Decimals:    dbExits[i].TokenDecimals,
				USD:         dbExits[i].TokenUSD,
				USDUpdate:   dbExits[i].TokenUSDUpdate,
			},
		})
	}
	return apiExits
}

// Tokens

type tokensAPI struct {
	Tokens     []tokenAPI     `json:"tokens"`
	Pagination *db.Pagination `json:"pagination"`
}

func (t *tokensAPI) GetPagination() *db.Pagination {
	if t.Tokens[0].ItemID < t.Tokens[len(t.Tokens)-1].ItemID {
		t.Pagination.FirstReturnedItem = t.Tokens[0].ItemID
		t.Pagination.LastReturnedItem = t.Tokens[len(t.Tokens)-1].ItemID
	} else {
		t.Pagination.LastReturnedItem = t.Tokens[0].ItemID
		t.Pagination.FirstReturnedItem = t.Tokens[len(t.Tokens)-1].ItemID
	}
	return t.Pagination
}
func (t *tokensAPI) Len() int { return len(t.Tokens) }

type tokenAPI struct {
	ItemID      int               `json:"itemId"`
	TokenID     common.TokenID    `json:"id"`
	EthBlockNum int64             `json:"ethereumBlockNum"` // Ethereum block number in which this token was registered
	EthAddr     ethCommon.Address `json:"ethereumAddress"`
	Name        string            `json:"name"`
	Symbol      string            `json:"symbol"`
	Decimals    uint64            `json:"decimals"`
	USD         *float64          `json:"USD"`
	USDUpdate   *time.Time        `json:"fiatUpdate"`
}

func tokensToAPI(dbTokens []historydb.TokenRead) []tokenAPI {
	apiTokens := []tokenAPI{}
	for i := 0; i < len(dbTokens); i++ {
		apiTokens = append(apiTokens, tokenAPI{
			ItemID:      dbTokens[i].ItemID,
			TokenID:     dbTokens[i].TokenID,
			EthBlockNum: dbTokens[i].EthBlockNum,
			EthAddr:     dbTokens[i].EthAddr,
			Name:        dbTokens[i].Name,
			Symbol:      dbTokens[i].Symbol,
			Decimals:    dbTokens[i].Decimals,
			USD:         dbTokens[i].USD,
			USDUpdate:   dbTokens[i].USDUpdate,
		})
	}
	return apiTokens
}

// Config

type rollupConstants struct {
	PublicConstants         eth.RollupPublicConstants `json:"publicConstants"`
	MaxFeeIdxCoordinator    int                       `json:"maxFeeIdxCoordinator"`
	ReservedIdx             int                       `json:"reservedIdx"`
	ExitIdx                 int                       `json:"exitIdx"`
	LimitLoadAmount         *big.Int                  `json:"limitLoadAmount"`
	LimitL2TransferAmount   *big.Int                  `json:"limitL2TransferAmount"`
	LimitTokens             int                       `json:"limitTokens"`
	L1CoordinatorTotalBytes int                       `json:"l1CoordinatorTotalBytes"`
	L1UserTotalBytes        int                       `json:"l1UserTotalBytes"`
	MaxL1UserTx             int                       `json:"maxL1UserTx"`
	MaxL1Tx                 int                       `json:"maxL1Tx"`
	InputSHAConstantBytes   int                       `json:"inputSHAConstantBytes"`
	NumBuckets              int                       `json:"numBuckets"`
	MaxWithdrawalDelay      int                       `json:"maxWithdrawalDelay"`
	ExchangeMultiplier      int                       `json:"exchangeMultiplier"`
}

type configAPI struct {
	RollupConstants   rollupConstants       `json:"hermez"`
	AuctionConstants  eth.AuctionConstants  `json:"auction"`
	WDelayerConstants eth.WDelayerConstants `json:"withdrawalDelayer"`
}
