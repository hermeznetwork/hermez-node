package api

import (
	"encoding/base64"
	"math/big"
	"strconv"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

const exitIdx = "hez:EXIT:1"

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
	if idx == 1 {
		return exitIdx
	}
	return "hez:" + tokenSymbol + ":" + strconv.Itoa(int(idx))
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

type merkleProofAPI struct {
	Root     string
	Siblings []string
	OldKey   string
	OldValue string
	IsOld0   bool
	Key      string
	Value    string
	Fnc      int
}

type exitAPI struct {
	ItemID                 int                    `json:"itemId"`
	BatchNum               common.BatchNum        `json:"batchNum"`
	AccountIdx             string                 `json:"accountIndex"`
	MerkleProof            merkleProofAPI         `json:"merkleProof"`
	Balance                string                 `json:"balance"`
	InstantWithdrawn       *int64                 `json:"instantWithdrawn"`
	DelayedWithdrawRequest *int64                 `json:"delayedWithdrawRequest"`
	DelayedWithdrawn       *int64                 `json:"delayedWithdrawn"`
	Token                  historydb.TokenWithUSD `json:"token"`
}

func historyExitsToAPI(dbExits []historydb.HistoryExit) []exitAPI {
	apiExits := []exitAPI{}
	for i := 0; i < len(dbExits); i++ {
		exit := exitAPI{
			ItemID:     dbExits[i].ItemID,
			BatchNum:   dbExits[i].BatchNum,
			AccountIdx: idxToHez(dbExits[i].AccountIdx, dbExits[i].TokenSymbol),
			MerkleProof: merkleProofAPI{
				Root:     dbExits[i].MerkleProof.Root.String(),
				OldKey:   dbExits[i].MerkleProof.OldKey.String(),
				OldValue: dbExits[i].MerkleProof.OldValue.String(),
				IsOld0:   dbExits[i].MerkleProof.IsOld0,
				Key:      dbExits[i].MerkleProof.Key.String(),
				Value:    dbExits[i].MerkleProof.Value.String(),
				Fnc:      dbExits[i].MerkleProof.Fnc,
			},
			Balance:                dbExits[i].Balance.String(),
			InstantWithdrawn:       dbExits[i].InstantWithdrawn,
			DelayedWithdrawRequest: dbExits[i].DelayedWithdrawRequest,
			DelayedWithdrawn:       dbExits[i].DelayedWithdrawn,
			Token: historydb.TokenWithUSD{
				TokenID:     dbExits[i].TokenID,
				EthBlockNum: dbExits[i].TokenEthBlockNum,
				EthAddr:     dbExits[i].TokenEthAddr,
				Name:        dbExits[i].TokenName,
				Symbol:      dbExits[i].TokenSymbol,
				Decimals:    dbExits[i].TokenDecimals,
				USD:         dbExits[i].TokenUSD,
				USDUpdate:   dbExits[i].TokenUSDUpdate,
			},
		}
		siblings := []string{}
		for j := 0; j < len(dbExits[i].MerkleProof.Siblings); j++ {
			siblings = append(siblings, dbExits[i].MerkleProof.Siblings[j].String())
		}
		exit.MerkleProof.Siblings = siblings
		apiExits = append(apiExits, exit)
	}
	return apiExits
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
