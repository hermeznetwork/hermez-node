package api

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/hermeznetwork/hermez-node/db"
	"github.com/hermeznetwork/hermez-node/db/historydb"
	"github.com/hermeznetwork/hermez-node/db/l2db"
	"github.com/hermeznetwork/hermez-node/eth"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree"
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
	TxID        common.TxID      `json:"id"`
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
			TxID:        dbTxs[i].TxID,
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

// PoolL2Tx

type receivedPoolTx struct {
	TxID        common.TxID           `json:"id" binding:"required"`
	Type        common.TxType         `json:"type" binding:"required"`
	TokenID     common.TokenID        `json:"tokenId"`
	FromIdx     string                `json:"fromAccountIndex" binding:"required"`
	ToIdx       *string               `json:"toAccountIndex"`
	ToEthAddr   *string               `json:"toHezEthereumAddress"`
	ToBJJ       *string               `json:"toBjj"`
	Amount      string                `json:"amount" binding:"required"`
	Fee         common.FeeSelector    `json:"fee"`
	Nonce       common.Nonce          `json:"nonce"`
	Signature   babyjub.SignatureComp `json:"signature" binding:"required"`
	RqFromIdx   *string               `json:"requestFromAccountIndex"`
	RqToIdx     *string               `json:"requestToAccountIndex"`
	RqToEthAddr *string               `json:"requestToHezEthereumAddress"`
	RqToBJJ     *string               `json:"requestToBjj"`
	RqTokenID   *common.TokenID       `json:"requestTokenId"`
	RqAmount    *string               `json:"requestAmount"`
	RqFee       *common.FeeSelector   `json:"requestFee"`
	RqNonce     *common.Nonce         `json:"requestNonce"`
}

func (tx *receivedPoolTx) toDBWritePoolL2Tx() (*l2db.PoolL2TxWrite, error) {
	amount := new(big.Int)
	amount.SetString(tx.Amount, 10)
	txw := &l2db.PoolL2TxWrite{
		TxID:      tx.TxID,
		TokenID:   tx.TokenID,
		Amount:    amount,
		Fee:       tx.Fee,
		Nonce:     tx.Nonce,
		State:     common.PoolL2TxStatePending,
		Signature: tx.Signature,
		RqTokenID: tx.RqTokenID,
		RqFee:     tx.RqFee,
		RqNonce:   tx.RqNonce,
		Type:      tx.Type,
	}
	// Check FromIdx (required)
	fidx, err := stringToIdx(tx.FromIdx, "fromAccountIndex")
	if err != nil {
		return nil, err
	}
	if fidx == nil {
		return nil, errors.New("invalid fromAccountIndex")
	}
	// Set FromIdx
	txw.FromIdx = common.Idx(*fidx)
	// Set AmountFloat
	f := new(big.Float).SetInt(amount)
	amountF, _ := f.Float64()
	txw.AmountFloat = amountF
	if amountF < 0 {
		return nil, errors.New("amount must be positive")
	}
	// Check "to" fields, only one of: ToIdx, ToEthAddr, ToBJJ
	if tx.ToIdx != nil { // Case: Tx with ToIdx setted
		// Set ToIdx
		tidxUint, err := stringToIdx(*tx.ToIdx, "toAccountIndex")
		if err != nil || tidxUint == nil {
			return nil, errors.New("invalid toAccountIndex")
		}
		tidx := common.Idx(*tidxUint)
		txw.ToIdx = &tidx
	} else if tx.ToBJJ != nil { // Case: Tx with ToBJJ setted
		// tx.ToEthAddr must be equal to ethAddrWhenBJJLower or ethAddrWhenBJJUpper
		if tx.ToEthAddr != nil {
			toEthAddr, err := hezStringToEthAddr(*tx.ToEthAddr, "toHezEthereumAddress")
			if err != nil || *toEthAddr != common.FFAddr {
				return nil, fmt.Errorf("if toBjj is setted, toHezEthereumAddress must be hez:%s", common.FFAddr.Hex())
			}
		} else {
			return nil, fmt.Errorf("if toBjj is setted, toHezEthereumAddress must be hez:%s and toAccountIndex must be null", common.FFAddr.Hex())
		}
		// Set ToEthAddr and ToBJJ
		toBJJ, err := hezStringToBJJ(*tx.ToBJJ, "toBjj")
		if err != nil || toBJJ == nil {
			return nil, errors.New("invalid toBjj")
		}
		txw.ToBJJ = toBJJ
		txw.ToEthAddr = &common.FFAddr
	} else if tx.ToEthAddr != nil { // Case: Tx with ToEthAddr setted
		// Set ToEthAddr
		toEthAddr, err := hezStringToEthAddr(*tx.ToEthAddr, "toHezEthereumAddress")
		if err != nil || toEthAddr == nil {
			return nil, errors.New("invalid toHezEthereumAddress")
		}
		txw.ToEthAddr = toEthAddr
	} else {
		return nil, errors.New("one of toAccountIndex, toHezEthereumAddress or toBjj must be setted")
	}
	// Check "rq" fields
	if tx.RqFromIdx != nil {
		// check and set RqFromIdx
		rqfidxUint, err := stringToIdx(tx.FromIdx, "requestFromAccountIndex")
		if err != nil || rqfidxUint == nil {
			return nil, errors.New("invalid requestFromAccountIndex")
		}
		// Set RqFromIdx
		rqfidx := common.Idx(*rqfidxUint)
		txw.RqFromIdx = &rqfidx
		// Case: RqTx with RqToIdx setted
		if tx.RqToIdx != nil {
			// Set ToIdx
			tidxUint, err := stringToIdx(*tx.RqToIdx, "requestToAccountIndex")
			if err != nil || tidxUint == nil {
				return nil, errors.New("invalid requestToAccountIndex")
			}
			tidx := common.Idx(*tidxUint)
			txw.ToIdx = &tidx
		} else if tx.RqToBJJ != nil { // Case: Tx with ToBJJ setted
			// tx.ToEthAddr must be equal to ethAddrWhenBJJLower or ethAddrWhenBJJUpper
			if tx.RqToEthAddr != nil {
				rqEthAddr, err := hezStringToEthAddr(*tx.RqToEthAddr, "")
				if err != nil || *rqEthAddr != common.FFAddr {
					return nil, fmt.Errorf("if requestToBjj is setted, requestToHezEthereumAddress must be hez:%s", common.FFAddr.Hex())
				}
			} else {
				return nil, fmt.Errorf("if requestToBjj is setted, toHezEthereumAddress must be hez:%s and requestToAccountIndex must be null", common.FFAddr.Hex())
			}
			// Set ToEthAddr and ToBJJ
			rqToBJJ, err := hezStringToBJJ(*tx.RqToBJJ, "requestToBjj")
			if err != nil || rqToBJJ == nil {
				return nil, errors.New("invalid requestToBjj")
			}
			txw.RqToBJJ = rqToBJJ
			txw.RqToEthAddr = &common.FFAddr
		} else if tx.RqToEthAddr != nil { // Case: Tx with ToEthAddr setted
			// Set ToEthAddr
			rqToEthAddr, err := hezStringToEthAddr(*tx.ToEthAddr, "requestToHezEthereumAddress")
			if err != nil || rqToEthAddr == nil {
				return nil, errors.New("invalid requestToHezEthereumAddress")
			}
			txw.RqToEthAddr = rqToEthAddr
		} else {
			return nil, errors.New("one of requestToAccountIndex, requestToHezEthereumAddress or requestToBjj must be setted")
		}
		if tx.RqAmount == nil {
			return nil, errors.New("requestAmount must be provided if other request fields are setted")
		}
		rqAmount := new(big.Int)
		rqAmount.SetString(*tx.RqAmount, 10)
		txw.RqAmount = rqAmount
	} else if tx.RqToIdx != nil && tx.RqToEthAddr != nil && tx.RqToBJJ != nil &&
		tx.RqTokenID != nil && tx.RqAmount != nil && tx.RqNonce != nil && tx.RqFee != nil {
		// if tx.RqToIdx is not setted, tx.Rq* must be null as well
		return nil, errors.New("if requestFromAccountIndex is setted, the rest of request fields must be null as well")
	}

	return txw, validatePoolL2TxWrite(txw)
}

func validatePoolL2TxWrite(txw *l2db.PoolL2TxWrite) error {
	poolTx := common.PoolL2Tx{
		TxID:      txw.TxID,
		FromIdx:   txw.FromIdx,
		ToBJJ:     txw.ToBJJ,
		TokenID:   txw.TokenID,
		Amount:    txw.Amount,
		Fee:       txw.Fee,
		Nonce:     txw.Nonce,
		State:     txw.State,
		Signature: txw.Signature,
		RqToBJJ:   txw.RqToBJJ,
		RqAmount:  txw.RqAmount,
		Type:      txw.Type,
	}
	// ToIdx
	if txw.ToIdx != nil {
		poolTx.ToIdx = *txw.ToIdx
	}
	// ToEthAddr
	if txw.ToEthAddr == nil {
		poolTx.ToEthAddr = common.EmptyAddr
	} else {
		poolTx.ToEthAddr = *txw.ToEthAddr
	}
	// RqFromIdx
	if txw.RqFromIdx != nil {
		poolTx.RqFromIdx = *txw.RqFromIdx
	}
	// RqToIdx
	if txw.RqToIdx != nil {
		poolTx.RqToIdx = *txw.RqToIdx
	}
	// RqToEthAddr
	if txw.RqToEthAddr == nil {
		poolTx.RqToEthAddr = common.EmptyAddr
	} else {
		poolTx.RqToEthAddr = *txw.RqToEthAddr
	}
	// RqTokenID
	if txw.RqTokenID != nil {
		poolTx.RqTokenID = *txw.RqTokenID
	}
	// RqFee
	if txw.RqFee != nil {
		poolTx.RqFee = *txw.RqFee
	}
	// RqNonce
	if txw.RqNonce != nil {
		poolTx.RqNonce = *txw.RqNonce
	}
	// Check type and id
	_, err := common.NewPoolL2Tx(&poolTx)
	if err != nil {
		return err
	}
	// Check signature
	// Get public key
	account, err := s.GetAccount(poolTx.FromIdx)
	if err != nil {
		return err
	}
	if !poolTx.VerifySignature(account.PublicKey) {
		return errors.New("wrong signature")
	}
	return nil
}

type sendPoolTx struct {
	TxID        common.TxID           `json:"id"`
	Type        common.TxType         `json:"type"`
	FromIdx     string                `json:"fromAccountIndex"`
	ToIdx       *string               `json:"toAccountIndex"`
	ToEthAddr   *string               `json:"toHezEthereumAddress"`
	ToBJJ       *string               `json:"toBjj"`
	Amount      string                `json:"amount"`
	Fee         common.FeeSelector    `json:"fee"`
	Nonce       common.Nonce          `json:"nonce"`
	State       common.PoolL2TxState  `json:"state"`
	Signature   babyjub.SignatureComp `json:"signature"`
	Timestamp   time.Time             `json:"timestamp"`
	BatchNum    *common.BatchNum      `json:"batchNum"`
	RqFromIdx   *string               `json:"requestFromAccountIndex"`
	RqToIdx     *string               `json:"requestToAccountIndex"`
	RqToEthAddr *string               `json:"requestToHezEthereumAddress"`
	RqToBJJ     *string               `json:"requestToBJJ"`
	RqTokenID   *common.TokenID       `json:"requestTokenId"`
	RqAmount    *string               `json:"requestAmount"`
	RqFee       *common.FeeSelector   `json:"requestFee"`
	RqNonce     *common.Nonce         `json:"requestNonce"`
	Token       tokenAPI              `json:"token"`
}

func poolL2TxReadToSend(dbTx *l2db.PoolL2TxRead) *sendPoolTx {
	tx := &sendPoolTx{
		TxID:      dbTx.TxID,
		Type:      dbTx.Type,
		FromIdx:   idxToHez(dbTx.FromIdx, dbTx.TokenSymbol),
		Amount:    dbTx.Amount.String(),
		Fee:       dbTx.Fee,
		Nonce:     dbTx.Nonce,
		State:     dbTx.State,
		Signature: dbTx.Signature,
		Timestamp: dbTx.Timestamp,
		BatchNum:  dbTx.BatchNum,
		RqTokenID: dbTx.RqTokenID,
		RqFee:     dbTx.RqFee,
		RqNonce:   dbTx.RqNonce,
		Token: tokenAPI{
			TokenID:     dbTx.TokenID,
			EthBlockNum: dbTx.TokenEthBlockNum,
			EthAddr:     dbTx.TokenEthAddr,
			Name:        dbTx.TokenName,
			Symbol:      dbTx.TokenSymbol,
			Decimals:    dbTx.TokenDecimals,
			USD:         dbTx.TokenUSD,
			USDUpdate:   dbTx.TokenUSDUpdate,
		},
	}
	// ToIdx
	if dbTx.ToIdx != nil {
		toIdx := idxToHez(*dbTx.ToIdx, dbTx.TokenSymbol)
		tx.ToIdx = &toIdx
	}
	// ToEthAddr
	if dbTx.ToEthAddr != nil {
		toEth := ethAddrToHez(*dbTx.ToEthAddr)
		tx.ToEthAddr = &toEth
	}
	// ToBJJ
	if dbTx.ToBJJ != nil {
		toBJJ := bjjToString(dbTx.ToBJJ)
		tx.ToBJJ = &toBJJ
	}
	// RqFromIdx
	if dbTx.RqFromIdx != nil {
		rqFromIdx := idxToHez(*dbTx.RqFromIdx, dbTx.TokenSymbol)
		tx.RqFromIdx = &rqFromIdx
	}
	// RqToIdx
	if dbTx.RqToIdx != nil {
		rqToIdx := idxToHez(*dbTx.RqToIdx, dbTx.TokenSymbol)
		tx.RqToIdx = &rqToIdx
	}
	// RqToEthAddr
	if dbTx.RqToEthAddr != nil {
		rqToEth := ethAddrToHez(*dbTx.RqToEthAddr)
		tx.RqToEthAddr = &rqToEth
	}
	// RqToBJJ
	if dbTx.RqToBJJ != nil {
		rqToBJJ := bjjToString(dbTx.RqToBJJ)
		tx.RqToBJJ = &rqToBJJ
	}
	// RqAmount
	if dbTx.RqAmount != nil {
		rqAmount := dbTx.RqAmount.String()
		tx.RqAmount = &rqAmount
	}
	return tx
}
