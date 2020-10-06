package test

import (
	"math/big"
	"strconv"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-node/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/jmoiron/sqlx"
)

// CleanL2DB deletes 'tx_pool' and 'account_creation_auth' from the given DB
func CleanL2DB(db *sqlx.DB) {
	if _, err := db.Exec("DELETE FROM tx_pool;"); err != nil {
		panic(err)
	}
	if _, err := db.Exec("DELETE FROM account_creation_auth;"); err != nil {
		panic(err)
	}
}

// GenPoolTxs generates L2 pool txs.
// WARNING: This tx doesn't follow the protocol (signature, txID, ...)
// it's just to test getting/setting from/to the DB.
func GenPoolTxs(n int, tokens []common.Token) []*common.PoolL2Tx {
	txs := make([]*common.PoolL2Tx, 0, n)
	privK := babyjub.NewRandPrivKey()
	for i := 256; i < 256+n; i++ {
		var state common.PoolL2TxState
		//nolint:gomnd
		if i%4 == 0 {
			state = common.PoolL2TxStatePending
			//nolint:gomnd
		} else if i%4 == 1 {
			state = common.PoolL2TxStateInvalid
			//nolint:gomnd
		} else if i%4 == 2 {
			state = common.PoolL2TxStateForging
			//nolint:gomnd
		} else if i%4 == 3 {
			state = common.PoolL2TxStateForged
		}
		f := new(big.Float).SetInt(big.NewInt(int64(i)))
		amountF, _ := f.Float64()
		var usd, absFee *float64
		fee := common.FeeSelector(i % 255) //nolint:gomnd
		token := tokens[i%len(tokens)]
		if token.USD != nil {
			usd = new(float64)
			absFee = new(float64)
			*usd = *token.USD * amountF
			*absFee = fee.Percentage() * *usd
		}
		toIdx := new(common.Idx)
		*toIdx = common.Idx(i + 1)
		toEthAddr := new(ethCommon.Address)
		*toEthAddr = ethCommon.BigToAddress(big.NewInt(int64(i)))
		tx := &common.PoolL2Tx{
			FromIdx:           common.Idx(i),
			ToIdx:             toIdx,
			ToEthAddr:         toEthAddr,
			ToBJJ:             privK.Public(),
			TokenID:           token.TokenID,
			Amount:            big.NewInt(int64(i)),
			AmountFloat:       amountF,
			USD:               usd,
			Fee:               fee,
			Nonce:             common.Nonce(i),
			State:             state,
			Signature:         privK.SignPoseidon(big.NewInt(int64(i))),
			Timestamp:         time.Now().UTC(),
			AbsoluteFee:       absFee,
			AbsoluteFeeUpdate: token.USDUpdate,
		}
		var err error
		tx, err = common.NewPoolL2Tx(tx)
		if err != nil {
			panic(err)
		}
		if i%2 == 0 { // Optional parameters: rq
			rqFromIdx := new(common.Idx)
			*rqFromIdx = common.Idx(i)
			tx.RqFromIdx = rqFromIdx
			rqToIdx := new(common.Idx)
			*rqToIdx = common.Idx(i + 1)
			tx.RqToIdx = rqToIdx
			rqToEthAddr := new(ethCommon.Address)
			*rqToEthAddr = ethCommon.BigToAddress(big.NewInt(int64(i)))
			tx.RqToEthAddr = rqToEthAddr
			tx.RqToBJJ = privK.Public()
			rqTokenID := new(common.TokenID)
			*rqTokenID = common.TokenID(i)
			tx.RqTokenID = rqTokenID
			tx.RqAmount = big.NewInt(int64(i))
			rqFee := new(common.FeeSelector)
			*rqFee = common.FeeSelector(i)
			tx.RqFee = rqFee
			rqNonce := new(uint64)
			*rqNonce = uint64(i)
			tx.RqNonce = rqNonce
		}
		if i%3 == 0 { // Optional parameters: things that get updated "a posteriori"
			batchNum := new(common.BatchNum)
			*batchNum = 489
			tx.BatchNum = batchNum
		}
		txs = append(txs, tx)
	}
	return txs
}

// GenAuths generates account creation authorizations
func GenAuths(nAuths int) []*common.AccountCreationAuth {
	auths := []*common.AccountCreationAuth{}
	for i := 0; i < nAuths; i++ {
		privK := babyjub.NewRandPrivKey()
		auths = append(auths, &common.AccountCreationAuth{
			EthAddr:   ethCommon.BigToAddress(big.NewInt(int64(i))),
			BJJ:       privK.Public(),
			Signature: []byte(strconv.Itoa(i)),
			Timestamp: time.Now(),
		})
	}
	return auths
}
