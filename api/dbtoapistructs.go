package api

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// Commons of the API

type pagination struct {
	TotalItems       int `json:"totalItems"`
	LastReturnedItem int `json:"lastReturnedItem"`
}

type paginationer interface {
	GetPagination() pagination
	Len() int
}

type errorMsg struct {
	Message string
}

// History Tx related

type historyTxsAPI struct {
	Txs        []historyTxAPI `json:"transactions"`
	Pagination pagination     `json:"pagination"`
}

func (htx *historyTxsAPI) GetPagination() pagination { return htx.Pagination }
func (htx *historyTxsAPI) Len() int                  { return len(htx.Txs) }

type l1Info struct {
	ToForgeL1TxsNum int64   `json:"toForgeL1TransactionsNum"`
	UserOrigin      bool    `json:"userOrigin"`
	FromEthAddr     string  `json:"fromEthereumAddress"`
	FromBJJ         string  `json:"fromBJJ"`
	LoadAmount      string  `json:"loadAmount"`
	LoadAmountUSD   float64 `json:"loadAmountUSD"`
	EthBlockNum     int64   `json:"ethereumBlockNum"`
}

type l2Info struct {
	Fee    common.FeeSelector `json:"fee"`
	FeeUSD float64            `json:"feeUSD"`
	Nonce  common.Nonce       `json:"nonce"`
}

type historyTxAPI struct {
	IsL1        string           `json:"L1orL2"`
	TxID        common.TxID      `json:"id"`
	Type        common.TxType    `json:"type"`
	Position    int              `json:"position"`
	FromIdx     string           `json:"fromAccountIndex"`
	ToIdx       string           `json:"toAccountIndex"`
	Amount      string           `json:"amount"`
	BatchNum    *common.BatchNum `json:"batchNum"`
	TokenID     common.TokenID   `json:"tokenId"`
	TokenSymbol string           `json:"tokenSymbol"`
	USD         float64          `json:"historicUSD"`
	Timestamp   time.Time        `json:"timestamp"`
	CurrentUSD  float64          `json:"currentUSD"`
	USDUpdate   time.Time        `json:"fiatUpdate"`
	L1Info      *l1Info          `json:"L1Info"`
	L2Info      *l2Info          `json:"L2Info"`
}

func historyTxsToAPI(dbTxs []*historydb.HistoryTx) []historyTxAPI {
	apiTxs := []historyTxAPI{}
	for i := 0; i < len(dbTxs); i++ {
		apiTx := historyTxAPI{
			TxID:        dbTxs[i].TxID,
			Type:        dbTxs[i].Type,
			Position:    dbTxs[i].Position,
			FromIdx:     "hez:" + dbTxs[i].TokenSymbol + ":" + strconv.Itoa(int(dbTxs[i].FromIdx)),
			ToIdx:       "hez:" + dbTxs[i].TokenSymbol + ":" + strconv.Itoa(int(dbTxs[i].ToIdx)),
			Amount:      dbTxs[i].Amount.String(),
			TokenID:     dbTxs[i].TokenID,
			USD:         dbTxs[i].USD,
			BatchNum:    nil,
			Timestamp:   dbTxs[i].Timestamp,
			TokenSymbol: dbTxs[i].TokenSymbol,
			CurrentUSD:  dbTxs[i].CurrentUSD,
			USDUpdate:   dbTxs[i].USDUpdate,
			L1Info:      nil,
			L2Info:      nil,
		}
		bn := dbTxs[i].BatchNum
		if dbTxs[i].BatchNum != 0 {
			apiTx.BatchNum = &bn
		}
		if dbTxs[i].IsL1 {
			apiTx.IsL1 = "L1"
			apiTx.L1Info = &l1Info{
				ToForgeL1TxsNum: dbTxs[i].ToForgeL1TxsNum,
				UserOrigin:      dbTxs[i].UserOrigin,
				FromEthAddr:     "hez:" + dbTxs[i].FromEthAddr.String(),
				FromBJJ:         bjjToString(dbTxs[i].FromBJJ),
				LoadAmount:      dbTxs[i].LoadAmount.String(),
				LoadAmountUSD:   dbTxs[i].LoadAmountUSD,
				EthBlockNum:     dbTxs[i].EthBlockNum,
			}
		} else {
			apiTx.IsL1 = "L2"
			apiTx.L2Info = &l2Info{
				Fee:    dbTxs[i].Fee,
				FeeUSD: dbTxs[i].FeeUSD,
				Nonce:  dbTxs[i].Nonce,
			}
		}
		apiTxs = append(apiTxs, apiTx)
	}
	return apiTxs
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
